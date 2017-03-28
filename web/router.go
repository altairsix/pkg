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

type Router struct {
	target  *httprouter.Router
	prefix  string
	filters []Filter
}

func NewRouter() *Router {
	return &Router{
		target: httprouter.New(),
	}
}

func (r *Router) Use(filters ...Filter) {
	if r.filters == nil {
		r.filters = []Filter{}
	}

	r.filters = append(r.filters, filters...)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.target.ServeHTTP(w, req)
}

func (r *Router) applyFilters(h HandlerFunc) HandlerFunc {
	if r.filters != nil {
		for i := len(r.filters) - 1; i >= 0; i-- {
			h = r.filters[i](h)
		}
	}

	return h
}

func (r *Router) Handle(method, path string, h HandlerFunc) {
	path = filepath.Join(r.prefix, swag.ColonPath(path))
	h = r.applyFilters(h)
	r.target.Handle(strings.ToUpper(method), path, Wrap(h))
}

func (r *Router) GET(path string, h HandlerFunc) {
	r.Handle("GET", path, h)
}

func (r *Router) POST(path string, h HandlerFunc) {
	r.Handle("POST", path, h)
}

func (r *Router) DELETE(path string, h HandlerFunc) {
	r.Handle("DELETE", path, h)
}

func (r *Router) PUT(path string, h HandlerFunc) {
	r.Handle("PUT", path, h)
}

func (r *Router) OPTION(path string, h HandlerFunc) {
	r.Handle("OPTION", path, h)
}

func (r *Router) HEAD(path string, h HandlerFunc) {
	r.Handle("HEAD", path, h)
}

func (r *Router) Group(prefix string, filters ...Filter) *Router {
	return &Router{
		target:  r.target,
		prefix:  prefix,
		filters: r.filters,
	}
}

func (r *Router) Bind(endpoints ...*swagger.Endpoint) error {
	for _, endpoint := range endpoints {
		h := endpoint.Handler.(HandlerFunc)
		r.Handle(endpoint.Method, endpoint.Path, h)
	}

	return nil
}
