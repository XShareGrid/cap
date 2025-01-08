package rproxy

type funcOption struct {
	f func()
}

func (fo *funcOption) apply() {
	fo.f()
}

func newFuncOption(f func()) ServerOption {
	return &funcOption{f: f}
}

// NewHTTPMiddleware ...
func NewHTTPMiddleware(mds ...HTTPMiddleware) ServerOption {
	return newFuncOption(func() {
		registerHTTPMiddlewares(mds...)
	})
}
