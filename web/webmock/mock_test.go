package webmock_test

import (
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/altairsix/pkg/web"
	"github.com/altairsix/pkg/web/webmock"
	"github.com/savaki/swag/endpoint"
	"github.com/stretchr/testify/assert"
)

func SetHeader(k, v string) web.Filter {
	return func(h web.HandlerFunc) web.HandlerFunc {
		return func(c *web.Context) error {
			c.Response.Header().Set(k, v)
			return h(c)
		}
	}
}

func TestNew(t *testing.T) {
	msg := "hello world"

	t.Run("http.Handler", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			io.WriteString(w, msg)
		})

		client := webmock.New(
			webmock.Handler(handler),
		)
		resp, err := client.Get("/", nil)
		assert.Nil(t, err)
		defer resp.Body.Close()

		data, err := ioutil.ReadAll(resp.Body)
		assert.Nil(t, err)
		assert.Equal(t, msg, string(data))
	})

	t.Run("web.HandlerFunc", func(t *testing.T) {
		handler := func(c *web.Context) error {
			return c.Text(http.StatusOK, msg)
		}

		client := webmock.New(
			webmock.Filters(SetHeader("k1", "v1"), SetHeader("k2", "v2")),
			webmock.HandlerFunc(handler),
		)
		resp, err := client.Get("/", nil)
		assert.Nil(t, err)
		defer resp.Body.Close()
		assert.Equal(t, "v1", resp.Header.Get("k1"))
		assert.Equal(t, "v2", resp.Header.Get("k2"))

		data, err := ioutil.ReadAll(resp.Body)
		assert.Nil(t, err)
		assert.Equal(t, msg, string(data))
	})

	t.Run("*swagger.Endpoint", func(t *testing.T) {
		handler := func(c *web.Context) error {
			return c.Text(http.StatusOK, msg)
		}
		e := endpoint.New("get", "/blah", "swagger defined endpoint",
			endpoint.Handler(handler),
		)

		client := webmock.New(
			webmock.Filters(SetHeader("k1", "v1"), SetHeader("k2", "v2")),
			webmock.Endpoints(e),
		)
		resp, err := client.Get(e.Path, nil)
		assert.Nil(t, err)
		defer resp.Body.Close()
		assert.Equal(t, "v1", resp.Header.Get("k1"))
		assert.Equal(t, "v2", resp.Header.Get("k2"))

		data, err := ioutil.ReadAll(resp.Body)
		assert.Nil(t, err)
		assert.Equal(t, msg, string(data))
	})
}
