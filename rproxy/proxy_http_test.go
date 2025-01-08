package rproxy

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

func Test_proxyHandler_ServeHTTP(t *testing.T) {
	u := url.PathEscape("abc/FA%23abc")
	fmt.Println(u)
}

func TestRegisterHTTPMiddlewares(t *testing.T) {
	m1 := func(h http.Handler) http.Handler {
		fmt.Println("m1")
		return h
	}
	m2 := func(h http.Handler) http.Handler {
		fmt.Println("m2")
		return h
	}
	registerHTTPMiddlewares(m1, m2)
	defaultHTTPMiddleware(nil)
}
