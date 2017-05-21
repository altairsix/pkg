package webmock

import (
	"io"
	"log"
	"net/http"
	"time"

	"github.com/altairsix/pkg/web"
	"github.com/savaki/swag/swagger"
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

func Filters(filters ...web.Filter) Option {
	return func(c *Client) {
		c.filters = filters
	}
}

func HandlerFunc(h web.HandlerFunc) Option {
	return func(c *Client) {
		c.factory = func(filters web.Filters) http.Handler {
			router := web.NewRouter()
			h = filters.Apply(h)
			router.GET("/", h)
			router.POST("/", h)
			return router
		}
	}
}

func Handler(h http.Handler) Option {
	return func(c *Client) {
		c.factory = func(filters web.Filters) http.Handler {
			if len(filters) > 0 {
				log.Fatalln("webmock.Handler cannot be used in conjunction with webmock.Filters")
			}
			return h
		}
	}
}

// Endpoint constructs a client directly from a swagger endpoint
func Endpoints(endpoints ...*swagger.Endpoint) Option {
	return func(c *Client) {
		c.factory = func(filters web.Filters) http.Handler {
			router := web.NewRouter()
			for _, endpoint := range endpoints {
				switch h := endpoint.Handler.(type) {
				case web.HandlerFunc:
					h = filters.Apply(h)
					router.Handle(endpoint.Method, endpoint.Path, h)

				case func(c web.Context) error:
					h = filters.Apply(h)
					router.Handle(endpoint.Method, endpoint.Path, h)

				default:
					log.Fatalf("*swagger.Endpoint contained unhandled handler, %#v", endpoint.Handler)
				}
			}
			return router
		}
	}
}
