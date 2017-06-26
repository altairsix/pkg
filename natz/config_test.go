package natz_test

import (
	"testing"

	"github.com/altairsix/pkg/local"
	"github.com/altairsix/pkg/natz"
	"github.com/nats-io/go-nats"
	"github.com/savaki/randx"
	"github.com/stretchr/testify/assert"
)

func TestSubject(t *testing.T) {
	sample := randx.AlphaN(20)
	service := randx.AlphaN(20)
	fqSubject := natz.Subject(local.Env, service, sample)
	assert.Equal(t, local.Env+"."+service+"."+sample, fqSubject)
}

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
			assert.Equal(t, tc.Expected, natz.Url(natz.Config{
				Url:      tc.Url,
				Username: tc.Username,
				Password: tc.Password,
			}))
		})
	}
}
