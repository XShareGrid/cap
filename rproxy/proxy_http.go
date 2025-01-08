package rproxy

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/XShareGrid/cap/logger"
	"github.com/astaxie/beego"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/soheilhy/cmux"
	"golang.org/x/net/http2"
)

var webRequestTotal *prometheus.CounterVec
var webRequestDuration *prometheus.HistogramVec
var ph http.Handler

func init() {
	webRequestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "web_request_total",
			Help: "Number of requests in total",
		},
		[]string{"method", "endpoint"},
	)

	webRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "web_request_duration_seconds",
			Help:    "web request duration distribution",
			Buckets: []float64{0.1, 0.3, 0.5, 0.7, 0.9, 1},
		},
		[]string{"method", "endpoint"},
	)
	prometheus.MustRegister(webRequestTotal)
	prometheus.MustRegister(webRequestDuration)
	ph = promhttp.Handler()

}

func newReqID() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

type proxyHandler struct {
}

func matchPrefix(prefix string, uri string) (relativePath string, match bool) {
	if prefix == "/" {
		relativePath = uri
		return relativePath, true
	}
	if strings.HasPrefix(uri, prefix) {
		relativePath = strings.TrimPrefix(uri, prefix)
		if strings.HasPrefix(relativePath, "/") || relativePath == "" {
			return relativePath, true
		}
	}
	return "", false

}

func urlEscape(uri string) string {
	uri = SlashEscape(uri)
	// 这里做特殊处理，后期做成通用的函数
	return strings.Replace(uri, "#", "%23", -1)
}

// 静态资源代理
func staticProxy(w http.ResponseWriter, r *http.Request) {
	pURL := r.URL.Query().Get("url")

	// forward
	remote, err := url.Parse(pURL)
	if err != nil {
		logger.CAP.Info("error parsing host: " + r.URL.Host)
		return
	}
	r.Header.Set("X-Real-IP", r.RemoteAddr)
	r.URL = remote
	forward, err := url.Parse(fmt.Sprintf("%s://%s", r.URL.Scheme, r.URL.Host))
	if err != nil {
		logger.CAP.Info("error parsing host: " + r.URL.Host)
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(forward)
	proxy.ServeHTTP(w, r)
}

func (p *proxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	reqID := start.UnixNano()
	if beego.BConfig.RunMode == "dev" {
		incrementConnections(reqID, fmt.Sprintf("[%s] %s from [%s]", r.Method, r.URL.Path, start))
		defer func() {
			end := time.Now()
			logger.CAP.Info("ServeHTTP[%s] %s from [%s] to [%s], duration[%d ms]", r.Method, r.URL.Path, start, end, end.Sub(start).Milliseconds())
			decrementConnections(reqID)
		}()
	}

	// grpc-web会以http1.1的形式下发请求，使用包装器做协议转换，serve为grpc
	if serveGRPCWeb(w, r) {
		return
	}

	uri := r.URL.EscapedPath()
	uri = urlEscape(uri)
	// hard coded，cap用header或referer来转发
	if strings.HasPrefix(uri, "/cap") {
		subsystemHeader := r.Header.Get("Subsystem")
		if subsystemHeader != "" {
			uri = subsystemHeader
		} else {
			referer := r.Header.Get("Referer")
			u, err := url.Parse(referer)
			if err == nil {
				segments := strings.Split(u.Path, "/")
				if len(segments) > 0 {
					uri = segments[0]
				}
			}
		}
	}
	if strings.HasPrefix(uri, "/metrics") {
		ph.ServeHTTP(w, r)
		return
	}

	if strings.HasPrefix(uri, "/staticproxy") {
		staticProxy(w, r)
		return
	}

	matchedPrefix := ""
	HTTP.rangeRule(func(prefix, dst string) bool {
		if _, ok := matchPrefix(prefix, uri); ok {
			if len(matchedPrefix) < len(prefix) {
				matchedPrefix = prefix
			}
		}
		return true
	})
	if matchedPrefix != "" {
		dst, ok := HTTP.Load(matchedPrefix)
		if !ok {
			logger.CAP.Info("error loading prefix: " + matchedPrefix)
			return
		}
		rPath, ok := matchPrefix(matchedPrefix, uri)
		if !ok {
			logger.CAP.Info("error matching prefix: " + matchedPrefix)
			return
		}
		u := fmt.Sprintf("%s%s", dst, rPath)
		if r.URL.RawQuery != "" {
			u += fmt.Sprintf("?%s", r.URL.RawQuery)
		}
		// rewrite request
		r.RequestURI = ""
		fullURL, err := url.Parse(u)
		if err != nil {
			logger.CAP.Info("error parsing url: " + dst.(string))
			return
		}
		r.URL = fullURL
		// forward
		remote, err := url.Parse(fmt.Sprintf("%s://%s", r.URL.Scheme, r.URL.Host))
		if err != nil {
			logger.CAP.Info("error parsing host: " + r.URL.Host)
			return
		}
		r.Header.Set("X-Real-IP", r.RemoteAddr)
		proxy := httputil.NewSingleHostReverseProxy(remote)
		reqID := newReqID()
		logger.CAP.Info("request[%s] received %s, headers: [%s] from [%s] matched route %s, forward to %s",
			reqID, r.URL.RequestURI(), r.Header, r.RemoteAddr, matchedPrefix, dst.(string))
		r.Header.Set("psi-req-id", reqID)
		proxy.ServeHTTP(w, r)
		duration := time.Since(start)
		endpoint := matchedPrefix
		if route := w.Header().Get("__route"); route != "" {
			endpoint = route
		}
		if setCookie := w.Header().Get("Set-Cookie"); setCookie != "" {
			logger.CAP.Info("request[%s] set-cookie [%s]",
				reqID, setCookie)
		}
		w.Header().Set("psi-req-id", reqID)
		webRequestTotal.With(prometheus.Labels{"method": r.Method, "endpoint": endpoint}).Inc()
		webRequestDuration.With(prometheus.Labels{"method": r.Method, "endpoint": endpoint}).Observe(duration.Seconds())
	} else {
		logger.CAP.Info("no router is matched")
		http.NotFound(w, r)
	}
}

// HTTPMiddleware ...
type HTTPMiddleware func(http.Handler) http.Handler

var defaultHTTPMiddleware = func(h http.Handler) http.Handler {
	fmt.Println("default")
	return h
}

// RegisterHTTPMiddlewares MUST be called before StartProxy
func registerHTTPMiddlewares(ms ...HTTPMiddleware) {
	for _, m := range ms {
		tmp := defaultHTTPMiddleware
		mid := m
		defaultHTTPMiddleware = func(h http.Handler) http.Handler {
			return mid(tmp(h))
		}
	}
}

var httpServer *http.Server

func registerHTTPServer(listener cmux.CMux) {
	httpListener := listener.Match(cmux.HTTP1Fast(http.MethodPatch))
	h := defaultHTTPMiddleware(&proxyHandler{})
	httpServer = &http.Server{
		Handler:     h,
		IdleTimeout: 30 * time.Second,
	}
	http2.ConfigureServer(httpServer, &http2.Server{})
	go func() {
		err := httpServer.Serve(httpListener)
		if err != nil {
			log.Println("httpServer.Serve:", err)
		}
	}()
}
