package session

import "net/http"

type Option func(*manager)

func ErrHandler(fn func(w http.ResponseWriter, req *http.Request, err error)) Option {
	return func(c *manager) {
		c.errHandler = fn
	}
}
