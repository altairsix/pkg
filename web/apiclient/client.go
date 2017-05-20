package apiclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"

	"github.com/altairsix/pkg/types"
	"github.com/pkg/errors"
)

type roundTripper struct {
	handler http.Handler
}

func (r *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Body == nil {
		req.Body = ioutil.NopCloser(bytes.NewReader([]byte{}))
	}

	w := httptest.NewRecorder()
	r.handler.ServeHTTP(w, req)

	return w.Result(), nil
}

type Client struct {
	output         io.Writer
	handler        http.Handler
	httpClient     *http.Client
	codebase       string
	filters        []func(*http.Request)
	ignoredHeaders types.StringSet
	expects        []int
}

type Option func(c *Client)

func WithBasicAuth(username, password string) Option {
	return func(c *Client) {
		if username == "" || password == "" {
			return
		}

		c.filters = append(c.filters, func(req *http.Request) {
			req.SetBasicAuth(username, password)
		})
	}
}

func WithHandler(handler http.Handler) Option {
	return func(c *Client) {
		if handler != nil {
			c.httpClient.Transport = &roundTripper{handler: handler}
		}
	}
}

func WithCodebase(codebase string) Option {
	return func(c *Client) {
		c.codebase = codebase
	}
}

func WithOutput(w io.Writer) Option {
	return func(c *Client) {
		c.output = w
	}
}

func New(opts ...Option) (*Client, error) {
	cookieJar, _ := cookiejar.New(nil)
	httpClient := &http.Client{
		Jar: cookieJar,
	}

	ignoredHeaders := types.StringSet{}
	ignoredHeaders.Add(
		"Pragma",
		"Expires",
		"Cache-Control",
		"Content-Length",
		"Date",
		"X-Content-Type-Options",
	)

	client := &Client{
		output:         ioutil.Discard,
		httpClient:     httpClient,
		filters:        []func(*http.Request){},
		ignoredHeaders: ignoredHeaders,
		expects:        []int{http.StatusOK, http.StatusCreated, http.StatusTemporaryRedirect},
	}

	for _, opt := range opts {
		opt(client)
	}

	if client.codebase == "" {
		client.codebase = "http://localhost"
	}
	if client.output == nil {
		client.output = ioutil.Discard
	}

	return client, nil
}

type P struct {
	K string
	V string
}

func (c *Client) Expects(statusCodes ...int) *Client {
	return &Client{
		output:         c.output,
		handler:        c.handler,
		httpClient:     c.httpClient,
		codebase:       c.codebase,
		filters:        c.filters,
		ignoredHeaders: c.ignoredHeaders,
		expects:        statusCodes,
	}
}

func (c *Client) Url(path string) string {
	return c.codebase + path
}

func (c *Client) GET(path string, out interface{}, params ...P) error {
	return c.Do("GET", path, nil, out, params...)
}

func (c *Client) POST(path string, in, out interface{}, params ...P) error {
	return c.Do("POST", path, in, out, params...)
}

func (c *Client) Do(method, path string, in, out interface{}, params ...P) error {
	var body io.Reader
	if in != nil {
		data, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewReader(data)
	}

	if len(params) > 0 {
		values := url.Values{}
		for _, p := range params {
			values.Add(p.K, p.V)
		}
		path += "?" + values.Encode()
	}
	u := c.Url(path)

	req, err := http.NewRequest(method, u, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	for _, opt := range c.filters {
		opt(req)
	}

	fmt.Fprintln(c.output, "")
	fmt.Fprintf(c.output, "%v %v\n", method, path)

	if in != nil {
		data, _ := json.MarshalIndent(in, "", "  ")
		fmt.Fprintln(c.output, "")
		fmt.Fprintln(c.output, string(data))
	}

	fmt.Fprintln(c.output, "")
	fmt.Fprintln(c.output, "--")
	fmt.Fprintln(c.output, "")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "api call failed")
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "unable to response body")
	}

	fmt.Fprintln(c.output, resp.Status)
	for k, values := range resp.Header {
		if c.ignoredHeaders.Contains(k) {
			continue
		}
		for _, v := range values {
			fmt.Fprintf(c.output, "%v: %v\n", k, v)
		}
	}

	if len(data) > 0 {
		v := map[string]interface{}{}
		err := json.Unmarshal(data, &v)
		if err != nil {
			fmt.Fprintln(c.output, string(data))
			return errors.Wrap(err, "unable to unmarshal content")
		}
		fmt.Fprintln(c.output, "")
		encoder := json.NewEncoder(c.output)
		encoder.SetIndent("", "  ")
		encoder.Encode(v)
		fmt.Fprintln(c.output, "")
	}

	if out != nil {
		err = json.Unmarshal(data, out)
		if err != nil {
			return err
		}
	}

	if len(c.expects) > 0 {
		for _, sc := range c.expects {
			if sc == resp.StatusCode {
				return nil
			}
		}
	} else {
		return nil
	}

	return fmt.Errorf("%v %v - got status %v, expected one of %v", method, path, resp.StatusCode, c.expects)
}
