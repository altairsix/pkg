package middleware

import (
	"io"
	"net/http"
	"strings"
)

func Swagger(h http.Handler, prefix string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if !strings.HasPrefix(req.URL.Path, prefix) {
			h.ServeHTTP(w, req)
			return
		}

		path := strings.Replace(req.URL.String(), prefix, "", 1)
		switch len(path) {
		case 1:
			http.Redirect(w, req, req.URL.Path+"?url=/api", http.StatusTemporaryRedirect)
			return
		case 0:
			http.Redirect(w, req, req.URL.Path+"/?url=/api", http.StatusTemporaryRedirect)
			return
		}

		resp, err := http.Get("http://petstore.swagger.io" + path)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, err.Error())
			return
		}
		defer resp.Body.Close()

		w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	})
}
