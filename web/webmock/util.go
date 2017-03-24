package webmock

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
)

func newReader(body interface{}) io.Reader {
	if body == nil {
		return nil
	}

	var r io.Reader
	switch v := body.(type) {
	case []byte:
		r = bytes.NewReader(v)
	case io.Reader:
		r = v
	default:
		data, _ := json.Marshal(body)
		r = bytes.NewReader(data)
	}

	return r
}

type roundTripper struct {
	handler http.Handler
}

func (r *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	w := httptest.NewRecorder()
	r.handler.ServeHTTP(w, req)

	resp := &http.Response{
		StatusCode: w.Code,
		Request:    req,
		Header:     w.HeaderMap,
	}

	if w.Body != nil {
		resp.Body = ioutil.NopCloser(bytes.NewReader(w.Body.Bytes()))
	}

	return resp, nil
}

func prettyReader(body interface{}) io.Reader {
	if body == nil {
		return nil
	}

	var r io.Reader
	switch body.(type) {
	case []byte:
		r = strings.NewReader("[binary content]")
	case io.Reader:
		r = strings.NewReader("[binary content]")
	default:
		data, _ := json.MarshalIndent(body, "", "  ")
		r = bytes.NewReader(data)
	}

	return r
}

func (c *Client) urlStr(path string, values url.Values) string {
	var urlStr string
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		urlStr = path
	} else {
		urlStr = c.codebase + path
	}
	if len(values) > 0 {
		urlStr = urlStr + "?" + values.Encode()
	}

	return urlStr
}
