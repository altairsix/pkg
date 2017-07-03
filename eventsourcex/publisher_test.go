package eventsourcex_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/altairsix/eventsource"
	"github.com/altairsix/pkg/checkpoint"
	"github.com/altairsix/pkg/eventsourcex"
	"github.com/altairsix/pkg/local"
	"github.com/nats-io/go-nats"
	"github.com/savaki/randx"
	"github.com/stretchr/testify/assert"
)

func TestPublisher(t *testing.T) {
	received := []eventsource.StreamRecord{}
	h := func(record eventsource.StreamRecord) error {
		received = append(received, record)
		return nil
	}

	records := []eventsource.StreamRecord{
		{
			Record: eventsource.Record{
				Version: 1,
				Data:    []byte("hello world"),
			},
			Offset:      0,
			AggregateID: "abc",
		},
	}
	r := eventsource.StreamReaderFunc(func(ctx context.Context, startingOffset int64, recordCount int) ([]eventsource.StreamRecord, error) {
		return records, nil
	})
	cp := checkpoint.New(local.Env, local.DynamoDB)

	publisher := eventsourcex.PublishStream(context.Background(), h, r, cp, "local", randx.AlphaN(20))
	defer publisher.Close()

	publisher.Check()
	time.Sleep(time.Millisecond * 50)
	assert.True(t, len(received) > 0, "expected at least 1 record to be received")
	assert.Equal(t, records[0], received[0])
}

type mockPublisher struct {
	checkCalled int32
	closeCalled int32
	done        chan struct{}
}

func (m *mockPublisher) Check() {
	m.checkCalled++
}

func (m *mockPublisher) Close() error {
	m.closeCalled++
	return nil
}

func (m *mockPublisher) Done() <-chan struct{} {
	return m.done
}

func TestWithReceiveNotifications(t *testing.T) {
	p := &mockPublisher{
		done: make(chan struct{}),
	}

	nc, err := nats.Connect(nats.DefaultURL)
	assert.Nil(t, err)
	defer nc.Close()

	env := "local"
	bc := randx.AlphaN(12)

	eventsourcex.WithReceiveNotifications(p, nc, env, bc)

	subject := eventsourcex.NoticesSubject(env, bc)

	time.Sleep(time.Millisecond * 25) // give the receiver a moment to setup the subscription
	nc.Publish(subject, []byte("hello"))
	time.Sleep(time.Millisecond * 25) // wait for the message to be received
	assert.Equal(t, int32(1), p.checkCalled)
}

func TestPublishEvents(t *testing.T) {
	nc, err := nats.Connect(nats.DefaultURL)
	assert.Nil(t, err)
	defer nc.Close()

	env := "local"
	bc := randx.AlphaN(12)
	id := randx.AlphaN(18)
	version := 123
	subject := eventsourcex.AggregateSubject(env, bc) + "." + id

	received := int32(0)
	sub, err := nc.Subscribe(subject, func(m *nats.Msg) {
		atomic.AddInt32(&received, 1)
	})
	assert.Nil(t, err)
	defer sub.Unsubscribe()

	h := eventsourcex.WithPublishEvents(func(eventsource.StreamRecord) error { return nil }, nc, env, bc)
	err = h(eventsource.StreamRecord{
		AggregateID: id,
		Record: eventsource.Record{
			Version: version,
		},
	})
	assert.Nil(t, err)

	// give the message a moment to propagate
	time.Sleep(time.Millisecond * 50)
	assert.Equal(t, int32(1), received, "expected message to have been propagated")
}
