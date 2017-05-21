package web

import "net/http"

type Response struct {
	Status  int    `json:"status"`
	Message string `json:"message,omitempty"`
}

var (
	Ok         = Response{Status: http.StatusOK}
	NotFound   = Response{Status: http.StatusNotFound}
	Forbidden  = Response{Status: http.StatusForbidden}
	BadRequest = Response{Status: http.StatusBadRequest}
)
