package eventsourcex_test

import (
	"testing"

	"github.com/nats-io/go-nats"
	"github.com/stretchr/testify/assert"
)

func TestSubscribeNotices(t *testing.T) {
	nc, err := nats.Connect(nats.DefaultURL)
	assert.Nil(t, err)
	defer nc.Close()
}
