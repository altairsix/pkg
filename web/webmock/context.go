package webmock

import (
	"net/http"
	"net/http/httptest"

	"github.com/altairsix/pkg/types"
	"github.com/altairsix/pkg/web"
)

type mock struct {
	request  *http.Request
	response *httptest.ResponseRecorder
	route    map[string]string
	query    map[string]string
	form     map[string]string
	data     map[string]interface{}
}

type Context interface {
	web.Context
	Recorder() *httptest.ResponseRecorder
}

type ContextOption func(m *mock)

func WithRequest(req *http.Request) ContextOption {
	return func(m *mock) {
		m.request = req
	}
}

func WithRoute(key, value string) ContextOption {
	return func(m *mock) {
		if m.route == nil {
			m.route = map[string]string{}
		}

		m.route[key] = value
	}
}

func WithQuery(key, value string) ContextOption {
	return func(m *mock) {
		if m.query == nil {
			m.query = map[string]string{}
		}

		m.query[key] = value
	}
}

func WithForm(key, value string) ContextOption {
	return func(m *mock) {
		if m.form == nil {
			m.form = map[string]string{}
		}

		m.form[key] = value
	}
}

func WithValue(key string, value interface{}) ContextOption {
	return func(m *mock) {
		if m.data == nil {
			m.data = map[string]interface{}{}
		}

		m.data[key] = value
	}
}

func NewContext(opts ...ContextOption) Context {
	req, _ := http.NewRequest("GET", "/", nil)
	m := &mock{
		request:  req,
		response: httptest.NewRecorder(),
		route:    map[string]string{},
		query:    map[string]string{},
		form:     map[string]string{},
		data:     map[string]interface{}{},
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func (m *mock) Recorder() *httptest.ResponseRecorder {
	return m.response
}

func (m *mock) Response() http.ResponseWriter {
	return m.response
}

func (m *mock) Request() *http.Request {
	return m.request
}

func (m *mock) WithRequest(req *http.Request) web.Context {
	return &mock{
		request:  req,
		response: m.response,
		route:    m.route,
		query:    m.query,
		form:     m.form,
		data:     m.data,
	}
}

func (m *mock) RouteValue(name string) string {
	return m.route[name]
}

func (m *mock) RouteKey(name string) types.Key {
	return types.Key(m.RouteValue(name))
}

func (m *mock) RouteID(name string) types.ID {
	v, err := types.NewID(m.RouteValue(name))
	if err != nil {
		return types.ZeroID
	}
	return v
}

func (m *mock) Query(name string) string {
	return m.query[name]
}

func (m *mock) FormValue(name string) string {
	return m.form[name]
}

func (m *mock) Set(k string, v interface{}) {
	m.data[k] = v
}

func (m *mock) Get(k string) interface{} {
	return m.data[k]
}

func (m *mock) JSON(status int, in interface{}) error {
	return nil
}

func (m *mock) XMLBlob(status int, in []byte) error {
	return nil
}

func (m *mock) Text(status int, in string) error {
	return nil
}

func (m *mock) HTML(status int, in string) error {
	return nil
}

func (m *mock) Redirect(status int, location string) error {
	return nil
}
