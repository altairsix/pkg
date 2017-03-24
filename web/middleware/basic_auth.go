package middleware

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func BasicAuth(h http.Handler, username, password string, ignores ...string) http.Handler {
	if username == "" || password == "" {
		return h
	}

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		for _, ignore := range ignores {
			if strings.HasPrefix(req.URL.Path, ignore) {
				h.ServeHTTP(w, req)
				return
			}
		}

		u, p, _ := req.BasicAuth()
		if username != u || password != p {
			w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm="%v"`, "Altair Six authentication required"))
			w.WriteHeader(http.StatusUnauthorized)
			io.WriteString(w, "unauthorized")
			return
		}

		h.ServeHTTP(w, req)
	})

}
