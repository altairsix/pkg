package webmock_test

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/altairsix/pkg/web/webmock"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	msg := "hello world"
	username := "username"
	password := "password"

	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		u, p, _ := req.BasicAuth()
		assert.Equal(t, username, u)
		assert.Equal(t, password, p)
		io.WriteString(w, msg)
	})

	client := webmock.New(handler, webmock.BasicAuth(username, password), webmock.Output(os.Stdout))
	resp, err := client.Get("/", nil)
	assert.Nil(t, err)
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	assert.Equal(t, msg, string(data))
}
