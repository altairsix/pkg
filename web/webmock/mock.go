package webmock

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/textproto"
	"net/url"
	"regexp"
	"time"

	"github.com/altairsix/pkg/web"
	"github.com/savaki/swag/swagger"
)

type Client struct {
	codebase string
	authFunc func(*http.Request) error
	client   *http.Client
	writer   io.Writer
	observer func(code int, method, endpoint string, elapsed time.Duration)
}

func (c *Client) Do(method, path string, values url.Values, body interface{}, opts ...func(r *http.Request)) (*http.Response, error) {
	// -- Create the Request ------------------------------------------------
	//
	urlStr := c.urlStr(path, values)
	r := newReader(body)
	req, err := http.NewRequest(method, urlStr, r)
	if err != nil {
		return nil, err
	}

	// -- Configure Request -------------------------------------------------
	//
	if err = c.authFunc(req); err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(req)
	}

	// -- Print Request -----------------------------------------------------
	//
	buf := bytes.NewBuffer([]byte{})
	fmt.Fprintln(buf, "\n#-- Request ------------------------------------------")
	fmt.Fprintf(buf, "%v %v\n", method, urlStr)
	for key, values := range req.Header {
		for _, value := range values {
			fmt.Fprintf(buf, "%v: %v\n", key, value)
		}
	}
	if body != nil {
		io.WriteString(buf, "\n")
		io.Copy(buf, prettyReader(body))
	}

	io.Copy(c.writer, buf)

	// -- Execute Request ---------------------------------------------------
	//
	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	req = req.WithContext(ctx)
	since := time.Now()
	resp, err := c.client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
		c.observer(resp.StatusCode, req.Method, req.URL.Path, time.Now().Sub(since))
	}
	if err != nil {
		return nil, err
	}

	// -- Print Response ----------------------------------------------------
	//
	buf.Reset()
	if resp.Body != nil {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		resp.Body = ioutil.NopCloser(bytes.NewReader(data))

		fmt.Fprintln(buf, "\n\n#-- Response -----------------------------------------")
		fmt.Fprintf(buf, "%v\n", resp.Status)
		for key, values := range resp.Header {
			for _, value := range values {
				fmt.Fprintf(buf, "%v: %v\n", key, value)
			}
		}
		buf.Write(data)
	}

	fmt.Fprintln(buf, "\n\n#-- End ----------------------------------------------")
	io.Copy(c.writer, buf)

	return resp, nil
}

func (c *Client) Get(path string, values url.Values) (*http.Response, error) {
	return c.Do("GET", path, values, nil)
}

func (c *Client) Post(path string, values url.Values, body interface{}) (*http.Response, error) {
	return c.Do("POST", path, values, body)
}

func (c *Client) Put(path string, values url.Values, body interface{}) (*http.Response, error) {
	return c.Do("PUT", path, values, body)
}

func (c *Client) Patch(path string, values url.Values, body interface{}) (*http.Response, error) {
	return c.Do("PATCH", path, values, body)
}

func (c *Client) Delete(path string, values url.Values) (*http.Response, error) {
	return c.Do("DELETE", path, values, nil)
}

func (c *Client) Upload(path string, values url.Values, r io.Reader, field, filename, contentType string) (*http.Response, error) {
	buf := bytes.NewBuffer([]byte{})

	mp := multipart.NewWriter(buf)

	h := textproto.MIMEHeader{}
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%v"; filename="%v"`, field, filename))
	h.Set("Content-Type", contentType)
	w, _ := mp.CreatePart(h)
	io.Copy(w, r)

	mp.Close()

	req, err := http.NewRequest("POST", path, bytes.NewReader(buf.Bytes()))
	if err != nil {
		return nil, fmt.Errorf("unable to generate request - %v", err)
	}

	req.Header.Set("Content-Type", mp.FormDataContentType())

	return c.client.Do(req)
}

func (c *Client) Cookie(name string) (string, bool) {
	u, err := url.Parse(c.codebase)
	if err != nil {
		return "", false
	}

	if cookies := c.client.Jar.Cookies(u); cookies != nil {
		for _, cookie := range cookies {
			if cookie.Name == name {
				return cookie.Value, true
			}
		}
	}

	return "", false
}

func New(handler http.Handler, opts ...Option) *Client {
	cookieJar, _ := cookiejar.New(nil)
	httpClient := &http.Client{
		Jar: cookieJar,
	}

	if handler != nil {
		httpClient.Transport = &roundTripper{handler: handler}
	}

	c := &Client{
		client:   httpClient,
		observer: func(code int, method, endpoint string, elapsed time.Duration) {},
		authFunc: func(req *http.Request) error { return nil },
		writer:   ioutil.Discard,
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.codebase == "" {
		c.codebase = "http://localhost"
	}

	// strip trailing slashes from codebase
	pat := regexp.MustCompile(`/+$`)
	c.codebase = pat.ReplaceAllString(c.codebase, "")

	return c
}

// Endpoint constructs a client directly from a swagger endpoint
func Endpoints(endpoints ...*swagger.Endpoint) http.Handler {
	router := web.NewRouter()
	for _, endpoint := range endpoints {
		h := endpoint.Handler.(web.HandlerFunc)
		router.Handle(endpoint.Method, endpoint.Path, h)
	}

	return router
}
