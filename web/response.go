package web

import "net/http"

type Response struct {
	Status int `json:"status"`
}

var (
	Ok        = Response{Status: http.StatusOK}
	NotFound  = Response{Status: http.StatusNotFound}
	Forbidden = Response{Status: http.StatusForbidden}
)
