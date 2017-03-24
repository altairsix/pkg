package web

import "net/http"

type Response struct {
	Status int
}

var Ok = Response{Status: http.StatusOK}
