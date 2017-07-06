package heartbeat_test

import (
	"context"
	"testing"
	"time"

	"github.com/altairsix/pkg/action"
	"github.com/altairsix/pkg/action/heartbeat"
	"github.com/nats-io/go-nats"
	"github.com/savaki/randx"
	"github.com/stretchr/testify/assert"
)

func TestNats(t *testing.T) {
	nc, err := nats.Connect(nats.DefaultURL)
	assert.Nil(t, err)

	subject := randx.AlphaN(12)
	tk := heartbeat.Nats(nc, subject)

	tick := action.Tick{
		ID:        randx.AlphaN(12),
		StartedAt: time.Now(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := tk.Receive(ctx)
	assert.Nil(t, err)

	err = tk.Publish(tick)
	assert.Nil(t, err)

	actual := <-ch
	assert.Equal(t, tick, actual)
}
