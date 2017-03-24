package av

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitize(t *testing.T) {
	testCases := map[string]struct {
		In    string
		Key   string
		Value string
	}{
		"simple": {
			In:    "a",
			Key:   "#a",
			Value: ":a",
		},
		"dash": {
			In:    "a-b",
			Key:   "#aB",
			Value: ":aB",
		},
		"under": {
			In:    "a_b",
			Key:   "#aB",
			Value: ":aB",
		},
		"multi-under": {
			In:    "a__b",
			Key:   "#aB",
			Value: ":aB",
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			key, value := sanitize(tc.In)
			assert.Equal(t, tc.Key, key)
			assert.Equal(t, tc.Value, value)
		})
	}
}
