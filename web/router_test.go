package web_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/altairsix/pkg/web"

	"path/filepath"

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

func TestFilter_Apply(t *testing.T) {
	calls := 0
	r := web.NewRouter()
	group := r.Group("", func(h web.HandlerFunc) web.HandlerFunc {
		return func(c *web.Context) error {
			calls++
			return h(c)
		}
	})
	group.GET("/", func(c *web.Context) error { return c.Text(http.StatusOK, "ok") })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "http://localhost/", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 1, calls)
}

func TestFilter_Prefix(t *testing.T) {
	r := web.NewRouter()
	group := r.Group("a").Group("b").Group("c")

	calls := 0
	group.GET("/", func(c *web.Context) error {
		calls++
		return c.Text(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "http://localhost/a/b/c", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 1, calls)
}

func TestFilePath(t *testing.T) {
	assert.Equal(t, "", filepath.Join("", ""))
}
