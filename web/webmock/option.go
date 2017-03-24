package webmock

import (
	"io"
	"net/http"
	"time"
)

type Option func(*Client)

// BasicAuth provides basic auth
func BasicAuth(username, password string) Option {
	return func(c *Client) {
		c.authFunc = func(req *http.Request) error {
			req.SetBasicAuth(username, password)
			return nil
		}
	}
}

// AuthFunc allows for an arbitrary authentication function
func AuthFunc(authFunc func(*http.Request) error) Option {
	return func(c *Client) {
		if authFunc == nil {
			c.authFunc = func(*http.Request) error { return nil }
		} else {
			c.authFunc = authFunc
		}
	}
}

// Codebase allows one to specify a remote codebase (defaults to http://localhost)
func Codebase(codebase string) Option {
	return func(c *Client) {
		c.codebase = codebase
	}
}

// Output enables debug output to be written to the specified writer
func Output(w io.Writer) Option {
	return func(c *Client) {
		c.writer = w
	}
}

func Observer(fn func(code int, method, endpoint string, elapsed time.Duration)) Option {
	return func(c *Client) {
		c.observer = fn
	}
}
