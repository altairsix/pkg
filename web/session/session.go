package session

import (
	gocontext "context"
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/altairsix/pkg/web"
	gorilla "github.com/gorilla/sessions"
)

const (
	key = "session"
)

func Value(req *http.Request) (interface{}, bool) {
	v := req.Context().Value(key)
	return v, v != nil
}

func New(req *http.Request, w http.ResponseWriter, store gorilla.Store, cookieName string, v interface{}) error {
	s, err := store.New(req, cookieName)
	if err != nil {
		return err
	}

	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	s.Values[key] = data
	return s.Save(req, w)
}

func Filter(prototype interface{}, cookieName string, store gorilla.Store) web.Filter {
	t := reflect.TypeOf(prototype)

	return func(h web.HandlerFunc) web.HandlerFunc {
		return func(c *web.Context) error {
			if _, err := c.Request.Cookie(cookieName); err == nil {
				s, err := store.Get(c.Request, cookieName)
				if err != nil {
					return err
				}
				if v, ok := s.Values[key]; ok {
					if data, ok := v.([]byte); ok {
						obj := reflect.New(t).Interface()
						err := json.Unmarshal(data, obj)
						if err != nil {
							return err
						}

						ctx := c.Request.Context()
						ctx = gocontext.WithValue(ctx, key, obj)
						c.Request = c.Request.WithContext(ctx)
					}
				}

				s.Save(c.Request, c.Response)
			}

			return h(c)
		}
	}
}
