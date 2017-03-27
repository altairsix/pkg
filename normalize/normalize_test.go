package normalize_test

import (
	"testing"

	"github.com/altairsix/pkg/normalize"
	"github.com/stretchr/testify/assert"
)

func TestEmail(t *testing.T) {
	testCases := map[string]struct {
		In  string
		Out string
	}{
		"simple": {
			In:  " Foo.Bar@Example.Com ",
			Out: "foo.bar@example.com",
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			email := normalize.Email(tc.In)
			assert.Equal(t, tc.Out, email)
		})
	}
}
