package swagx_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/altairsix/pkg/web/swagx"
	"github.com/altairsix/pkg/web/webmock"
	"github.com/stretchr/testify/assert"
)

type Input struct {
	In string
}

type Output struct {
	Out string
}

type Foo struct {
}

func (f *Foo) Public(in Input) Output {
	return Output{Out: "hello " + in.In}
}

func (f *Foo) private(Input) Output {
	return Output{Out: "hello world"}
}

func TestBind(t *testing.T) {
	e, err := swagx.Endpoints("/api/", &Foo{})
	assert.Nil(t, err)
	assert.Len(t, e, 1)
	assert.Equal(t, "/api/Public", e[0].Path)
}

func TestHandler(t *testing.T) {
	testCases := map[string]struct {
		Receiver interface{}
		Method   string
		In       string
		Out      string
	}{
		"simple": {
			Receiver: &Foo{},
			Method:   "Public",
			In:       `{"In":"world"}`,
			Out:      `{"Out":"hello world"}`,
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			method, ok := reflect.TypeOf(tc.Receiver).MethodByName(tc.Method)
			assert.True(t, ok)
			h := swagx.Handler(reflect.ValueOf(tc.Receiver), method)

			c := webmock.NewContext(webmock.WithBodyString(tc.In))
			err := h(c)
			assert.Nil(t, err)
			assert.Equal(t, tc.Out, strings.TrimSpace(c.Recorder().Body.String()))
		})
	}
}
