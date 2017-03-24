package webmock

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Unmarshal(t *testing.T, resp *http.Response, v interface{}) {
	if !assert.NotNil(t, resp) {
		return
	}
	defer resp.Body.Close()

	if v == nil {
		io.Copy(ioutil.Discard, resp.Body)
		return
	}

	err := json.NewDecoder(resp.Body).Decode(v)
	assert.Nil(t, err)
}
