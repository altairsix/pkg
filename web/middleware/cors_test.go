package middleware_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/altairsix/pkg/web/middleware"
	"github.com/stretchr/testify/assert"
)

func TestBasicAuth(t *testing.T) {
	username := "username"
	password := "password"

	t.Run("unauthorized", func(t *testing.T) {
		fn := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			io.WriteString(w, "ok")
		})

		h := middleware.BasicAuth(fn, username, password)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "http://localhost", nil)
		h.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("ok", func(t *testing.T) {
		fn := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			io.WriteString(w, "ok")
		})

		h := middleware.BasicAuth(fn, username, password)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "http://localhost", nil)
		req.SetBasicAuth(username, password)
		h.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
