package web

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/altairsix/pkg/types"
	"github.com/julienschmidt/httprouter"
)

type HandlerFunc func(c Context) error

type Context interface {
	Response() http.ResponseWriter
	Request() *http.Request
	WithRequest(req *http.Request) Context
	RouteValue(name string) string
	RouteKey(name string) types.Key
	RouteID(name string) types.ID
	Query(name string) string
	FormValue(name string) string
	Set(k string, v interface{})
	Get(k string) interface{}
	JSON(status int, in interface{}) error
	XMLBlob(status int, in []byte) error
	Text(status int, in string) error
	HTML(status int, in string) error
	Redirect(status int, location string) error
}

type rawContext struct {
	request    *http.Request
	response   http.ResponseWriter
	params     httprouter.Params
	data       map[string]interface{}
	formParsed bool
}

func (r *rawContext) Response() http.ResponseWriter {
	return r.response
}

func (r *rawContext) Request() *http.Request {
	return r.request
}

func (r *rawContext) WithRequest(req *http.Request) Context {
	return &rawContext{
		request:    req,
		response:   r.response,
		params:     r.params,
		data:       r.data,
		formParsed: r.formParsed,
	}
}

func (r *rawContext) RouteValue(name string) string {
	return r.params.ByName(name)
}

func (r *rawContext) RouteKey(name string) types.Key {
	return types.Key(r.RouteValue(name))
}

func (r *rawContext) RouteID(name string) types.ID {
	id, err := types.NewID(r.RouteValue(name))
	if err != nil {
		return types.ZeroID
	}

	return id
}

func (r *rawContext) Query(name string) string {
	if r.formParsed {
		r.request.ParseForm()
	}

	return r.request.FormValue(name)
}

func (r *rawContext) FormValue(name string) string {
	if r.formParsed {
		r.request.ParseForm()
	}

	return r.request.PostFormValue(name)
}

func (r *rawContext) Set(k string, v interface{}) {
	if r.data == nil {
		r.data = map[string]interface{}{}
	}
	r.data[k] = v
}

func (r *rawContext) Get(k string) interface{} {
	if r.data == nil {
		r.data = map[string]interface{}{}
	}

	return r.data[k]
}

func (r *rawContext) JSON(status int, in interface{}) error {
	r.response.Header().Set("Content-Type", "application/json")
	r.response.WriteHeader(status)
	return json.NewEncoder(r.response).Encode(in)
}

func (r *rawContext) XMLBlob(status int, in []byte) error {
	r.response.Header().Set("Content-Type", "text/xml")
	r.response.WriteHeader(status)
	_, err := r.response.Write(in)
	return err
}

func (r *rawContext) Text(status int, in string) error {
	r.response.Header().Set("Content-Type", "text/plain")
	r.response.WriteHeader(status)
	_, err := io.WriteString(r.response, in)
	return err
}

func (r *rawContext) HTML(status int, in string) error {
	r.response.Header().Set("Content-Type", "text/html")
	r.response.WriteHeader(status)
	_, err := io.WriteString(r.response, in)
	return err
}

// Redirect the browser to a new location; status is typically http.StatusTemporaryRedirect
func (r *rawContext) Redirect(status int, location string) error {
	r.response.Header().Set("Location", location)
	r.response.WriteHeader(status)
	return nil
}
