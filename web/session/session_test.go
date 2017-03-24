package session_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/altairsix/pkg/web"
	"github.com/altairsix/pkg/web/session"
	"github.com/gorilla/sessions"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type Session struct {
	Name string
}

func TestSession(t *testing.T) {
	pairs := session.GenerateKeyPairs()
	cookieName := "blah"
	store := sessions.NewCookieStore(pairs...)
	filter := session.Filter(Session{}, cookieName, store)
	tracer := errors.New("marker")
	expected := Session{Name: "Joe"}

	// -- verify no cookies set on load

	fn := func(c *web.Context) error { return tracer }
	fn = filter(fn)

	req, _ := http.NewRequest("GET", "http://localhost", nil)
	w := httptest.NewRecorder()
	err := fn(&web.Context{Request: req, Response: w})
	assert.Equal(t, tracer, err)
	assert.Equal(t, 0, len(w.HeaderMap))

	// -- verify cookies set on save

	fn = func(c *web.Context) error {
		err := session.New(c.Request, c.Response, store, cookieName, expected)
		assert.Nil(t, err)
		return tracer
	}
	fn = filter(fn)

	req, _ = http.NewRequest("GET", "http://localhost", nil)
	w = httptest.NewRecorder()
	err = fn(&web.Context{Request: req, Response: w})
	assert.Equal(t, tracer, err)
	assert.Equal(t, 1, len(w.HeaderMap))
	assert.Contains(t, w.HeaderMap, "Set-Cookie")

	assert.True(t, strings.HasPrefix(w.HeaderMap.Get("Set-Cookie"), cookieName+"="))
	cookieValue := w.HeaderMap.Get("Set-Cookie")[len(cookieName)+1:]
	cookieValue = strings.Split(cookieValue, ";")[0]
	cookieValue = strings.TrimSpace(cookieValue)

	// -- verify can read cookies

	fn = func(c *web.Context) error {
		v, ok := session.Value(c.Request)
		assert.True(t, ok)
		assert.Equal(t, &expected, v)
		return tracer
	}
	fn = filter(fn)

	req, _ = http.NewRequest("GET", "http://localhost", nil)
	req.AddCookie(&http.Cookie{
		Name:  cookieName,
		Value: cookieValue,
	})
	w = httptest.NewRecorder()
	err = fn(&web.Context{Request: req, Response: w})
	assert.Equal(t, tracer, err)
}
