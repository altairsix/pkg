package notice_test

import (
	"context"
	"testing"

	"github.com/altairsix/pkg/eventsourcex/notice"
	"github.com/nats-io/go-nats"
	"github.com/savaki/randx"
	"github.com/stretchr/testify/assert"
)

func TestSubscribe(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nc, err := nats.Connect(nats.DefaultURL)
	assert.Nil(t, err)
	defer nc.Close()

	subject := randx.AlphaN(12)
	in, err := notice.Subscribe(ctx, nc, subject, 128)
	assert.Nil(t, err)

	nc.Publish(subject, []byte("a"))
	nc.Publish(subject, []byte("b"))

	assert.Equal(t, "a", (<-in).AggregateID())
	assert.Equal(t, "b", (<-in).AggregateID())
}
