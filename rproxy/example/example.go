package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/XShareGrid/cap/rproxy"
	"github.com/XShareGrid/cap/rproxy/example/proto"
	"google.golang.org/grpc"
)

type handler struct {
	name string
}

type TestServiceServer struct {
}

func (s TestServiceServer) TestRequest(ctx context.Context, req *proto.Req) (*proto.Rsp, error) {
	return &proto.Rsp{Rsp: req.Req + time.Now().String()}, nil
}
func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("[%s] request received: \n", h.name)
	b, _ := ioutil.ReadAll(r.Body)
	fmt.Println(r.Method, r.RequestURI, r.UserAgent(), b)
	http.NotFound(w, r)
}

func startHTTPService(name, addr string) {
	go http.ListenAndServe(addr, &handler{name})
	fmt.Println("started service: " + name)
}

func startGRPCService(addr string) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	gs := grpc.NewServer()
	proto.RegisterTestServiceServer(gs, &TestServiceServer{})
	go gs.Serve(lis)

}

func main() {
	startHTTPService("service-0", ":8000")
	startHTTPService("service-1", ":8001")
	startGRPCService(":8002")
	rproxy.HTTP.CreateOrUpdateRule("/fas", "http://localhost:8000")
	rproxy.HTTP.CreateOrUpdateRule("/", "http://localhost:8001")
	rproxy.GRPC.CreateOrUpdateRule("/proto.TestService", "localhost:8002", true)
	rproxy.StartProxy(":8088")
}
