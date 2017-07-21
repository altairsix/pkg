package eventsourcex_test

import (
	"context"
	"testing"
	"time"

	"github.com/altairsix/eventsource"
	"github.com/altairsix/pkg/eventsourcex"
	"github.com/nats-io/go-nats"
	"github.com/savaki/randx"
	"github.com/stretchr/testify/assert"
)

type Mock struct {
}

type Command struct {
	eventsource.CommandModel
}

func (m Mock) Apply(ctx context.Context, cmd eventsource.Command) (int, error) {
	return 0, nil
}

func TestWithConsistentRead(t *testing.T) {
	nc, err := nats.Connect(nats.DefaultURL)
	assert.Nil(t, err)
	defer nc.Close()

	id := randx.AlphaN(12)
	env := randx.AlphaN(12)
	bc := randx.AlphaN(12)
	subject := eventsourcex.NoticesSubject(env, bc)

	sub, err := nc.Subscribe(subject, func(msg *nats.Msg) {
		if subject := msg.Reply; subject != "" {
			nc.Publish(msg.Reply, msg.Data)
		}
	})
	assert.Nil(t, err)
	sub.AutoUnsubscribe(1)

	startedAt := time.Now()
	repo := eventsourcex.WithConsistentRead(Mock{}, nc, env, bc)
	_, err = repo.Apply(context.Background(), &Command{CommandModel: eventsource.CommandModel{ID: id}})
	assert.Nil(t, err)
	assert.True(t, time.Now().Sub(startedAt) < time.Millisecond*250)
}
