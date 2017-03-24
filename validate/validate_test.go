package validate_test

import (
	"fmt"
	"testing"

	"github.com/altairsix/pkg/validate"

	"github.com/stretchr/testify/assert"
)

type Sample struct {
	Name string `json:"name" validate:"required,label"`
}

func TestStruct(t *testing.T) {
	testCases := map[string]struct {
		In      string
		IsValid bool
	}{
		"empty": {
			In:      "",
			IsValid: false,
		},
		"filled": {
			In:      "1234",
			IsValid: true,
		},
		"too-long": {
			In:      "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890",
			IsValid: false,
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			v := Sample{Name: tc.In}
			err := validate.Struct(v)
			fmt.Printf("%12s -> %#v\n", label, err)
			assert.True(t, (err == nil) == tc.IsValid)
		})
	}
}
