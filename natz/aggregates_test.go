package natz_test

import (
	"context"
	"testing"
	"time"

	"github.com/altairsix/pkg/natz"
	"github.com/altairsix/pkg/types"
	"github.com/nats-io/go-nats"
	"github.com/stretchr/testify/assert"
)

func TestAggregates(t *testing.T) {
	nc, err := nats.Connect(nats.DefaultURL)
	assert.Nil(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	env := "local"
	bc := "bc"
	key := types.Key("abc")

	received := make(chan types.Key, 8)
	defer close(received)

	// Given a subscriber to the aggregate
	natz.SubscribeAggregate(ctx, nc, env, bc, func(id types.Key) { received <- id })

	// When aggregate sync requested
	fn := natz.SyncAggregate(nc, "local", bc, time.Second)
	fn(ctx, key)

	// Then subscriber processed the request
	assert.Equal(t, key, <-received)
}
