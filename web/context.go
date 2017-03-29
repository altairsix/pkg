package web

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/altairsix/pkg/types"

	"github.com/julienschmidt/httprouter"
)

type HandlerFunc func(c *Context) error

type Context struct {
	Request    *http.Request
	Response   http.ResponseWriter
	params     httprouter.Params
	data       map[string]interface{}
	formParsed bool
}

func (c *Context) RouteValue(name string) string {
	return c.params.ByName(name)
}

func (c *Context) RouteKey(name string) types.Key {
	return types.Key(c.RouteValue(name))
}

func (c *Context) RouteID(name string) types.ID {
	id, err := types.NewID(c.RouteValue(name))
	if err != nil {
		return types.ZeroID
	}

	return id
}

func (c *Context) Query(name string) string {
	if c.formParsed {
		c.Request.ParseForm()
	}

	return c.Request.FormValue(name)
}

func (c *Context) FormValue(name string) string {
	if c.formParsed {
		c.Request.ParseForm()
	}

	return c.Request.PostFormValue(name)
}

func (c *Context) Set(k string, v interface{}) {
	if c.data == nil {
		c.data = map[string]interface{}{}
	}
	c.data[k] = v
}

func (c *Context) Get(k string) interface{} {
	if c.data == nil {
		c.data = map[string]interface{}{}
	}

	return c.data[k]
}

func (c *Context) JSON(status int, in interface{}) error {
	c.Response.Header().Set("Content-Type", "application/json")
	c.Response.WriteHeader(status)
	return json.NewEncoder(c.Response).Encode(in)
}

func (c *Context) XMLBlob(status int, in []byte) error {
	c.Response.Header().Set("Content-Type", "text/xml")
	c.Response.WriteHeader(status)
	_, err := c.Response.Write(in)
	return err
}

func (c *Context) Text(status int, in string) error {
	c.Response.Header().Set("Content-Type", "text/plain")
	c.Response.WriteHeader(status)
	_, err := io.WriteString(c.Response, in)
	return err
}

func (c *Context) HTML(status int, in string) error {
	c.Response.Header().Set("Content-Type", "text/html")
	c.Response.WriteHeader(status)
	_, err := io.WriteString(c.Response, in)
	return err
}
