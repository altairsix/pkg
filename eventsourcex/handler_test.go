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

type MockCP struct {
	saveCalled   int
	saveKey      string
	saveSequence uint64

	loadCalled   int
	loadKey      string
	loadSequence uint64
}

func (cp *MockCP) Save(ctx context.Context, key string, sequence uint64) error {
	cp.saveKey = key
	cp.saveSequence = sequence
	cp.saveCalled++
	return nil
}

func (cp *MockCP) Load(ctx context.Context, key string) (uint64, error) {
	cp.loadCalled++
	cp.loadKey = key
	return cp.loadSequence, nil
}

func TestSlices(t *testing.T) {
	arr := []string{"a", "b", "c"}
	assert.Equal(t, []string{}, arr[0:0])
	assert.Equal(t, arr, arr[0:3])
}

func TestMessageHandlerInFlight(t *testing.T) {
	t.Run("process in flight message on close", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		processorCalled := 0
		p := func(ctx context.Context, events ...eventsource.Event) error {
			processorCalled++
			return nil
		}

		unmarshalCalled := 0
		u := func(data []byte) (eventsource.Event, error) {
			unmarshalCalled++
			return eventsource.Model{}, nil
		}
		cp := &MockCP{}
		h := eventsourcex.NewMessageHandler(ctx, p, u, cp, randx.AlphaN(12))
		defer h.Close()

		// When
		h.Handle(0, nil)
		time.Sleep(time.Millisecond)

		// Then
		h.Close()
		<-h.Done()

		assert.Equal(t, 1, processorCalled)
		assert.Equal(t, 1, unmarshalCalled)
	})
}

func TestMessageHandlerInterval(t *testing.T) {
	t.Run("publish messages after interval", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		received := make(chan struct{}, 1)
		processorCalled := 0
		p := func(ctx context.Context, events ...eventsource.Event) error {
			processorCalled++
			received <- struct{}{}
			return nil
		}

		unmarshalCalled := 0
		u := func(data []byte) (eventsource.Event, error) {
			unmarshalCalled++
			return eventsource.Model{}, nil
		}
		cp := &MockCP{}
		interval := time.Millisecond * 25
		h := eventsourcex.NewMessageHandler(ctx, p, u, cp, randx.AlphaN(12),
			eventsourcex.WithInterval(interval),
		)
		defer h.Close()

		// When
		startedAt := time.Now()
		h.Handle(0, nil)
		<-received
		assert.True(t, time.Now().Sub(startedAt) < time.Second)
	})
}

func TestMessageHandler(t *testing.T) {
	t.Run("publish messages when buffer size is reached", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		received := make(chan struct{}, 1)
		processorCalled := 0
		p := func(ctx context.Context, events ...eventsource.Event) error {
			processorCalled++
			received <- struct{}{}
			return nil
		}

		unmarshalCalled := 0
		u := func(data []byte) (eventsource.Event, error) {
			unmarshalCalled++
			return eventsource.Model{}, nil
		}
		cp := &MockCP{}
		h := eventsourcex.NewMessageHandler(ctx, p, u, cp, randx.AlphaN(12),
			eventsourcex.WithInterval(time.Second),
			eventsourcex.WithBufferSize(1),
		)
		defer h.Close()

		// When
		startedAt := time.Now()
		h.Handle(0, nil)
		<-received
		assert.True(t, time.Now().Sub(startedAt) < time.Millisecond*50)
	})
}

func TestWithSendNotices(t *testing.T) {
	nc, err := nats.Connect(nats.DefaultURL)
	assert.Nil(t, err)
	defer nc.Close()

	env := "local"
	boundedContext := randx.AlphaN(15)
	id := randx.AlphaN(15)
	source := randx.AlphaN(12)

	received := make(chan string, 1)
	fn := func(m *nats.Msg) {
		select {
		case received <- string(m.Data):
		default:
		}
	}

	subject := eventsourcex.NoticesSubject(env, boundedContext, id)
	sub, err := nc.Subscribe(subject, fn)
	assert.Nil(t, err)
	defer sub.Unsubscribe()

	p := eventsourcex.Processor(func(ctx context.Context, events ...eventsource.Event) error { return nil })
	p = eventsourcex.WithSendNotices(p, nc, env, boundedContext, source)

	event := eventsource.Model{ID: id}
	err = p.Do(context.Background(), event)
	assert.Nil(t, err)

	<-received
}
