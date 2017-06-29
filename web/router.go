package web

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/savaki/swag"
	"github.com/savaki/swag/swagger"
)

type Filter func(h HandlerFunc) HandlerFunc

func (f Filter) Apply(h HandlerFunc) HandlerFunc {
	return f(h)
}

type Filters []Filter

func (f Filters) Apply(h HandlerFunc) HandlerFunc {
	if f != nil {
		for i := len(f) - 1; i >= 0; i-- {
			filter := f[i]
			h = filter.Apply(h)
		}
	}

	return h
}

type Option interface{}

type Observer interface {
	On(method, path string, opts ...Option)
}

type nopObserver struct{}

func (n nopObserver) On(method, path string, opts ...Option) {}

type Router struct {
	target   *httprouter.Router
	prefix   string
	filters  Filters
	observer Observer
}

func NewRouter() *Router {
	return &Router{
		target:   httprouter.New(),
		observer: nopObserver{},
	}
}

func (r *Router) WithObserver(observer Observer) *Router {
	if observer != nil {
		r.observer = observer
	}
	return r
}

func (r *Router) Use(filters ...Filter) {
	if r.filters == nil {
		r.filters = Filters{}
	}

	r.filters = append(r.filters, filters...)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.target.ServeHTTP(w, req)
}

func (r *Router) Handle(method, path string, h HandlerFunc, opts ...Option) {
	path = filepath.Join(r.prefix, swag.ColonPath(path))
	method = strings.ToUpper(method)
	h = r.filters.Apply(h)

	r.observer.On(method, path, opts...)
	r.target.Handle(method, path, Wrap(h))
}

func (r *Router) GET(path string, h HandlerFunc, opts ...Option) {
	r.Handle("GET", path, h, opts...)
}

func (r *Router) POST(path string, h HandlerFunc, opts ...Option) {
	r.Handle("POST", path, h, opts...)
}

func (r *Router) DELETE(path string, h HandlerFunc, opts ...Option) {
	r.Handle("DELETE", path, h, opts...)
}

func (r *Router) PUT(path string, h HandlerFunc, opts ...Option) {
	r.Handle("PUT", path, h, opts...)
}

func (r *Router) OPTION(path string, h HandlerFunc, opts ...Option) {
	r.Handle("OPTION", path, h, opts...)
}

func (r *Router) HEAD(path string, h HandlerFunc, opts ...Option) {
	r.Handle("HEAD", path, h, opts...)
}

func (r *Router) Group(prefix string, filters ...Filter) *Router {
	joined := filepath.Join(r.prefix, prefix)
	if joined != "" && !strings.HasPrefix(joined, "/") {
		joined = "/" + joined
	}

	return &Router{
		target:   r.target,
		prefix:   joined,
		filters:  append(r.filters, filters...),
		observer: r.observer,
	}
}

func (r *Router) Bind(endpoints ...*swagger.Endpoint) error {
	for _, endpoint := range endpoints {
		h := endpoint.Handler.(HandlerFunc)
		r.Handle(endpoint.Method, endpoint.Path, h)
	}

	return nil
}
