package session

import (
	gocontext "context"
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/altairsix/pkg/web"
	gorilla "github.com/gorilla/sessions"
)

const (
	key = "session"
)

func Value(req *http.Request) (interface{}, bool) {
	v := req.Context().Value(key)
	return v, v != nil
}

func New(req *http.Request, w http.ResponseWriter, store gorilla.Store, cookieName string, v interface{}) error {
	s, err := store.New(req, cookieName)
	if err != nil {
		return err
	}

	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	s.Values[key] = data
	return s.Save(req, w)
}

type manager struct {
	prototype  reflect.Type
	cookieName string
	store      gorilla.Store
	errHandler func(w http.ResponseWriter, req *http.Request, err error)
}

func (m *manager) attachSession(req *http.Request, w http.ResponseWriter) *http.Request {
	if _, err := req.Cookie(m.cookieName); err != nil {
		return req
	}

	s, err := m.store.Get(req, m.cookieName)
	if err != nil {
		m.errHandler(w, req, err)
		return req
	}

	v, ok := s.Values[key]
	if !ok {
		return req
	}

	data, ok := v.([]byte)
	if !ok {
		return req
	}

	obj := reflect.New(m.prototype).Interface()
	err = json.Unmarshal(data, obj)
	if err != nil {
		m.errHandler(w, req, err)
		return req
	}

	ctx := req.Context()
	ctx = gocontext.WithValue(ctx, key, obj)
	req = req.WithContext(ctx)

	s.Save(req, w)
	return req
}

func newManager(prototype interface{}, cookieName string, store gorilla.Store, opts ...Option) *manager {
	m := &manager{
		prototype:  reflect.TypeOf(prototype),
		cookieName: cookieName,
		store:      store,
		errHandler: func(w http.ResponseWriter, req *http.Request, err error) {},
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func WrapHandler(prototype interface{}, cookieName string, store gorilla.Store, h http.Handler, opts ...Option) http.Handler {
	m := newManager(prototype, cookieName, store, opts...)

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		req = m.attachSession(req, w)
		h.ServeHTTP(w, req)
	})
}

func Filter(prototype interface{}, cookieName string, store gorilla.Store, opts ...Option) web.Filter {
	m := newManager(prototype, cookieName, store, opts...)

	return func(h web.HandlerFunc) web.HandlerFunc {
		return func(c *web.Context) error {
			c.Request = m.attachSession(c.Request, c.Response)
			return h(c)
		}
	}
}
