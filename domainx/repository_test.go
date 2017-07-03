package domainx

import (
	"testing"
	"time"

	"github.com/nats-io/go-nats"
	"github.com/savaki/randx"
	"github.com/stretchr/testify/assert"
)

func TestSubscribeForUpdatesTimeout(t *testing.T) {
	nc, err := nats.Connect(nats.DefaultURL)
	assert.Nil(t, err)

	subject := randx.AlphaN(20)
	timeout := time.Millisecond * 100

	started := time.Now()
	done := subscribeForUpdates(nc, subject, timeout)
	<-done
	assert.True(t, time.Now().Sub(started) >= timeout)
}

func TestSubscribeForUpdates(t *testing.T) {
	nc, err := nats.Connect(nats.DefaultURL)
	assert.Nil(t, err)

	subject := randx.AlphaN(20)
	timeout := time.Second * 3

	started := time.Now()
	done := subscribeForUpdates(nc, subject, timeout)
	err = nc.Publish(subject, nil)
	assert.Nil(t, err)

	<-done
	assert.True(t, time.Now().Sub(started) < timeout)
}
