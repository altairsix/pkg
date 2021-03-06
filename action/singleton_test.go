package action_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/altairsix/pkg/action"
	"github.com/stretchr/testify/assert"
)

type MockHeartbeat struct {
	ch chan action.Tick
}

func (m *MockHeartbeat) Publish(tick action.Tick) error {
	return nil
}

func (m *MockHeartbeat) Receive(ctx context.Context) (<-chan action.Tick, error) {
	return m.ch, nil
}

func Run(counter *int32) action.Action {
	return func(ctx context.Context) error {
		atomic.AddInt32(counter, 1)
		return nil
	}
}

func TestSingleton(t *testing.T) {
	ch := make(chan action.Tick)
	defer close(ch)
	mock := &MockHeartbeat{
		ch: ch,
	}

	interval := time.Millisecond * 50

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after 25 ticks
	go func() {
		defer cancel()
		time.Sleep(interval * 25)
	}()

	// New leader on the scene after 6 ticks
	go func() {
		time.Sleep(interval * 6)
		mock.ch <- action.Tick{StartedAt: time.Now().Add(-time.Hour)}
	}()

	calls := int32(0)
	a := func(ctx context.Context) error {
		atomic.AddInt32(&calls, 1)
		select {
		case <-ctx.Done():
		case <-time.After(time.Minute):
		}
		return nil
	}
	singleton := action.Singleton(mock,
		action.WithInterval(interval),
		action.WithElections(interval*3),
		action.WithLease(interval*10),
	)
	err := singleton.AndThen(a).Do(ctx)
	assert.Nil(t, err)
	assert.Equal(t, int32(1), calls)
}
