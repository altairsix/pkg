package web

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func WrapHandler(h http.Handler) HandlerFunc {
	return func(c *Context) error {
		h.ServeHTTP(c.Response, c.Request)
		return nil
	}
}

func Wrap(h HandlerFunc) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
		c := &Context{
			Request:  req,
			Response: w,
			params:   p,
		}

		err := h(c)
		handleErr(w, err)
	}
}

func handleErr(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]string{
		"err": err.Error(),
	})
}
