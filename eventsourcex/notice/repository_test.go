package notice_test

import (
	"context"
	"testing"
	"time"

	"github.com/altairsix/eventsource"
	"github.com/altairsix/pkg/eventsourcex"
	"github.com/altairsix/pkg/eventsourcex/notice"
	"github.com/nats-io/go-nats"
	"github.com/savaki/randx"
	"github.com/stretchr/testify/assert"
)

func TestWithConsistentRead(t *testing.T) {
	nc, err := nats.Connect(nats.DefaultURL)
	assert.Nil(t, err)
	defer nc.Close()

	fn := eventsourcex.RepositoryFunc(func(ctx context.Context, cmd eventsource.Command) (int, error) {
		return 1, nil
	})

	subject := randx.AlphaN(12)
	sub, err := nc.Subscribe(subject, func(msg *nats.Msg) {
		if subject := msg.Reply; subject != "" {
			nc.Publish(subject, msg.Data)
		}
	})
	assert.Nil(t, err)
	sub.AutoUnsubscribe(1)

	timeout := time.Second
	startedAt := time.Now()
	repo := notice.WithConsistentRead(fn, nc, subject, timeout)
	repo.Apply(context.Background(), eventsource.CommandModel{ID: "abc"})
	assert.True(t, time.Now().Sub(startedAt) < timeout)
}
