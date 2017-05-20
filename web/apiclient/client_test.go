package apiclient_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/altairsix/pkg/web/apiclient"
	"github.com/stretchr/testify/assert"
)

type Content struct {
	Name string
}

func NewHandler(content interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(content)
	}
}

func TestNew(t *testing.T) {
	expected := Content{Name: "value"}
	h := NewHandler(expected)

	buf := bytes.NewBuffer(nil)
	client, err := apiclient.New(
		apiclient.WithHandler(h),
		apiclient.WithOutput(buf),
	)
	assert.Nil(t, err)

	actual := &Content{}
	err = client.GET("http://localhost", actual)
	assert.Nil(t, err)
	assert.Equal(t, expected.Name, actual.Name)
	assert.Contains(t, buf.String(), "Content-Type: application/json")
}
