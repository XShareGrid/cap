package rproxy

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"

	//"cloud.google.com/go/functions/metadata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/peer"

	//"github.com/mwitkow/grpc-proxy/proxy"
	"github.com/XShareGrid/cap/logger"
	"github.com/mwitkow/grpc-proxy/proxy"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/soheilhy/cmux"
	"github.com/zhlicen/grpc-web/go/grpcweb"
	"google.golang.org/grpc/metadata"
)

var grpcRequestTotal *prometheus.CounterVec
var grpcServer *grpc.Server

func init() {
	grpcRequestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_request_total",
			Help: "Number of grpc requests in total",
		},
		[]string{"function", "endpoint"},
	)
	prometheus.MustRegister(grpcRequestTotal)
}

const grpcWebContentType = "application/grpc-web"

// grpcWebServerWrap 对于grpc-web的封装，专用于grpc-web的协议转换
var grpcWebServerWrap *grpcweb.WrappedGrpcServer

func serveGRPCWeb(w http.ResponseWriter, r *http.Request) bool {
	accessControlHeaders := strings.ToLower(r.Header.Get("Access-Control-Request-Headers"))
	setupCORSHeaders := func(w http.ResponseWriter) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Subsystem, Content-Type, Referer, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-User-Agent, X-Grpc-Web")
		w.Header().Set("grpc-status", "")
		w.Header().Set("grpc-message", "")
	}
	if r.Method == http.MethodOptions && strings.Contains(accessControlHeaders, "x-grpc-web") {
		setupCORSHeaders(w)
		return true
	}

	if grpcWebServerWrap != nil && grpcWebServerWrap.IsGrpcWebRequest(r) {
		setupCORSHeaders(w)
		grpcWebServerWrap.HandleGrpcWebRequest(w, r)
		return true
	}

	return false
}

func registerGRPCServer(listener cmux.CMux) {
	// grpcListener := listener.Match(cmux.HTTP2HeaderField("content-type", "application/grpc"))
	grpcListener := listener.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
	grpcServer = grpc.NewServer(getGRPCProxyServerOptions()...)
	grpcWebServerWrap = grpcweb.WrapServer(grpcServer)

	go func() {
		err := grpcServer.Serve(grpcListener)
		if err != nil {
			log.Println("grpcServer.Serve:", err)
		}
	}()

}

var proxyDialOptions []grpc.DialOption
var proxyDialOptionsInsecure []grpc.DialOption
var grpcConn sync.Map

func init() {
	proxyDialOptions = []grpc.DialOption{grpc.WithCodec(proxy.Codec())}
	proxyDialOptionsInsecure = []grpc.DialOption{grpc.WithCodec(proxy.Codec()), grpc.WithInsecure()}
}

func getGRPCClietIP(ctx context.Context) (string, error) {
	pr, ok := peer.FromContext(ctx)
	if !ok {
		return "", fmt.Errorf("[GetGRPCClietIP] invoke FromContext() failed")
	}
	if pr.Addr == net.Addr(nil) {
		return "", fmt.Errorf("[GetGRPCClietIP] IP Address is nil")
	}
	addSlice := strings.Split(pr.Addr.String(), ":")
	return addSlice[0], nil
}

func getGRPCProxyServerOptions() []grpc.ServerOption {
	director := func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
		// Make sure we never forward internal services.
		if strings.HasPrefix(fullMethodName, "/com.example.internal.") {
			return nil, nil, grpc.Errorf(codes.Unimplemented, "Unknown method")
		}
		md, ok := metadata.FromIncomingContext(ctx)
		if strings.HasPrefix(fullMethodName, "/cap.") {
			subsystem := md.Get("Subsystem")
			if len(subsystem) > 0 {
				subsystem[0] = strings.ReplaceAll(subsystem[0], "/", "")
				if subsystem[0] != "" {
					fullMethodName = fmt.Sprintf("/%s%s", subsystem[0], fullMethodName)
				}
			} else {
				referer := md.Get("Referer")
				if len(referer) > 0 {
					u, err := url.Parse(referer[0])
					if err == nil {
						path := u.Path
						if len(path) > 1 {
							path = path[1:]
						}
						segments := strings.Split(path, "/")
						if len(segments) > 0 && segments[0] != "" {
							fullMethodName = fmt.Sprintf("/%s%s", segments[0], fullMethodName)
						}
					}
				}
			}
		}
		// Copy the inbound metadata explicitly.
		outCtx, _ := context.WithCancel(ctx)
		outCtx = metadata.NewOutgoingContext(outCtx, md.Copy())
		if ok {
			var dst string
			insecure := false
			// match fullmethodname with service prefix
			GRPC.rangeRule(func(servicePrefix string, rule *GRPCRule) bool {
				if strings.HasPrefix(fullMethodName, servicePrefix) {
					dst = rule.ForwardDst
					insecure = rule.Insecure
					return false
				}
				return true
			})
			clientIP, err := getGRPCClietIP(ctx)
			if err != nil {
				fmt.Println(err.Error())
			}
			if dst != "" {
				grpcRequestTotal.With(prometheus.Labels{"function": fullMethodName, "endpoint": dst}).Inc()
				logger.CAP.Info("ClientIP: %s GRPC call[%s] from[%v] matched -> [%s]", clientIP, md, fullMethodName, dst)
				dialOpts := proxyDialOptions
				if insecure {
					dialOpts = proxyDialOptionsInsecure
				}
				if storeConn, ok := grpcConn.Load(dst); ok {
					conn := storeConn.(*grpc.ClientConn)
					if conn.GetState() == connectivity.TransientFailure {
						fmt.Println("gRPC TransientFailure, retry connection")
						conn.ResetConnectBackoff()
					}
					return outCtx, conn, nil
				}
				dialOpts = append(dialOpts, grpc.WithMaxMsgSize(1024*1024*100))
				conn, err := grpc.DialContext(outCtx, dst, dialOpts...)
				if err == nil {
					grpcConn.Store(dst, conn)
				}
				go func() {
					for {
						conn.WaitForStateChange(context.Background(), conn.GetState())
						fmt.Println("conn state change to:", conn.GetState().String())
						if conn.GetState() == connectivity.Shutdown {
							grpcConn.Delete(dst)
							break
						}
					}

				}()
				return outCtx, conn, err
			} else {
				logger.CAP.Error("ClientIP: %s GRPC call[%s] no rules found", fullMethodName)
			}
		}
		return nil, nil, grpc.Errorf(codes.Unimplemented, "Unknown method")
	}

	return []grpc.ServerOption{
		grpc.CustomCodec(proxy.Codec()),
		grpc.UnknownServiceHandler(proxy.TransparentHandler(director)),
		grpc.MaxRecvMsgSize(1024 * 1024 * 100), grpc.MaxSendMsgSize(1024 * 1024 * 100),
	}

}
