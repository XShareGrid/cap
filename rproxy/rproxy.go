package rproxy

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/XShareGrid/cap/debug"
	"github.com/soheilhy/cmux"
)

func prettyMap(m map[int64]string) string {
	var s []string
	now := time.Now().UnixNano()
	for k, v := range m {
		du := time.Duration(now-k) * time.Nanosecond
		s = append(s, v+fmt.Sprintf("[%.2f]S", du.Seconds()))
	}
	sort.Strings(s)
	for i, _ := range s {
		s[i] = fmt.Sprintf("[%02d]%s", i, s[i])
	}
	return strings.Join(s, "\n")
}

// debug
func init() {
	debug.Register("pconn", func(stdin io.ReadCloser, stdout io.WriteCloser, stderr io.WriteCloser, args ...string) error {
		fmt.Fprintln(stdout, fmt.Sprintf("active connections: %d", len(activeConnections)), prettyMap(activeConnections))
		if len(args) > 0 {
			seconds, _ := strconv.Atoi(args[0])
			if seconds > 0 {
				ticker := time.Tick(1 * time.Second)
				for i := 0; i < seconds; i++ {
					<-ticker
					fmt.Fprintln(stdout, fmt.Sprintf("active connections: %d", len(activeConnections)), prettyMap(activeConnections))
					fmt.Fprintln(stdout, "=======================================================")
				}
			}
		}
		return nil
	})
}

var (
	activeConnections map[int64]string
	mu                sync.Mutex
)

func incrementConnections(reqID int64, reqDetail string) {
	if activeConnections == nil {
		activeConnections = make(map[int64]string)
	}
	mu.Lock()
	activeConnections[reqID] = reqDetail
	mu.Unlock()
}

func decrementConnections(reqID int64) {
	mu.Lock()
	// 从activeConnections删除reqID
	delete(activeConnections, reqID)
	mu.Unlock()
}

// StartProxy start reverse proxy
func StartProxy(addr string, opts ...ServerOption) {
	for _, o := range opts {
		o.apply()
	}
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	cl := cmux.New(l)
	registerHTTPServer(cl)
	registerGRPCServer(cl)
	cl.Serve()
}

// ServerOption ...
type ServerOption interface {
	apply()
}

// StopProxy start reverse proxy
func StopProxy() {
	fmt.Println("stopping http server...active connections:", activeConnections)
	if httpServer != nil {
		httpServer.Shutdown(context.Background())
		fmt.Println("http server stopped...")
	}
	fmt.Println("stopping gRPC server...")
	if grpcServer != nil {
		grpcServer.GracefulStop()
		fmt.Println("gRPC server stopped...")
	}
}
