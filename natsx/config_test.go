package natsx_test

import (
	"testing"

	"github.com/altairsix/pkg/natsx"
	"github.com/nats-io/go-nats"
	"github.com/stretchr/testify/assert"
)

func TestUrl(t *testing.T) {
	testCases := map[string]struct {
		Url      string
		Username string
		Password string
		Expected string
	}{
		"default": {
			Url:      nats.DefaultURL,
			Expected: nats.DefaultURL,
		},
		"multiple": {
			Url:      "nats://localhost:4222,nats://localhost:4223",
			Expected: "nats://localhost:4222, nats://localhost:4223",
		},
		"url with inline password": {
			Url:      "nats://username:password@localhost:4222",
			Expected: "nats://username:password@localhost:4222",
		},
		"url with external password": {
			Url:      "nats://localhost:4222",
			Username: "username",
			Password: "password",
			Expected: "nats://username:password@localhost:4222",
		},
		"multiple urls with external password": {
			Url:      "nats://localhost:4222,nats://localhost:4223",
			Username: "username",
			Password: "password",
			Expected: "nats://username:password@localhost:4222, nats://username:password@localhost:4223",
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			assert.Equal(t, tc.Expected, natsx.Url(natsx.Config{
				Url:      tc.Url,
				Username: tc.Username,
				Password: tc.Password,
			}))
		})
	}
}
