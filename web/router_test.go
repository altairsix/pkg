package web_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/altairsix/pkg/web"

	"github.com/stretchr/testify/assert"
)

func TestRouter_Use(t *testing.T) {
	applied := []string{}

	r := web.NewRouter()
	r.Use(func(h web.HandlerFunc) web.HandlerFunc {
		return func(c *web.Context) error {
			applied = append(applied, "a")
			return h(c)
		}
	})
	r.Use(func(h web.HandlerFunc) web.HandlerFunc {
		return func(c *web.Context) error {
			applied = append(applied, "b")
			return h(c)
		}
	})
	r.GET("/", func(c *web.Context) error {
		applied = append(applied, "c")
		return nil
	})

	req, err := http.NewRequest("GET", "http://localhost/", nil)
	assert.Nil(t, err)

	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	assert.Equal(t, []string{"a", "b", "c"}, applied)
}
