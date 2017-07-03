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

func TestWithNotifier(t *testing.T) {
	nc, err := nats.Connect(nats.DefaultURL)
	assert.Nil(t, err)
	defer nc.Close()

	id := randx.AlphaN(12)
	env := "local"
	bc := randx.AlphaN(12)

	received := make(chan struct{}, 1)
	sub, err := nc.Subscribe(eventsourcex.NoticesSubject(env, bc), func(m *nats.Msg) {
		select {
		case received <- struct{}{}:
		default:
		}
	})
	assert.Nil(t, err)
	defer sub.Unsubscribe()

	m := Mock{}
	repo := eventsourcex.WithNotifier(m, nc, env, bc)
	_, err = repo.Apply(context.Background(), &Command{CommandModel: eventsource.CommandModel{ID: id}})
	assert.Nil(t, err)

	<-received // expect a message to be received
}

func TestWithConsistenRead(t *testing.T) {
	nc, err := nats.Connect(nats.DefaultURL)
	assert.Nil(t, err)
	defer nc.Close()

	id := randx.AlphaN(12)
	env := "local"
	bc := randx.AlphaN(12)
	subject := eventsourcex.NoticesSubject(env, bc) + "." + id

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// continually publish an update that WithConsistentRead can listen to
	go func() {
		timer := time.NewTimer(time.Millisecond * 25)
		defer timer.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				nc.Publish(subject, []byte(id))
			}
		}
	}()

	startedAt := time.Now()
	repo := eventsourcex.WithConsistentRead(Mock{}, nc, env, bc)
	_, err = repo.Apply(context.Background(), &Command{CommandModel: eventsource.CommandModel{ID: id}})
	assert.Nil(t, err)
	assert.True(t, time.Now().Sub(startedAt) < time.Millisecond*250)
}
