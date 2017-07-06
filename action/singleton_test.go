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

	invocations := int32(0)
	a := Run(&invocations)
	singleton := action.Singleton(mock,
		action.WithRestarts(-1), // repeat forever
		action.WithInterval(interval),
		action.WithElections(interval*3),
		action.WithLease(interval*10),
	)
	err := a.Use(singleton).Do(ctx)
	assert.Nil(t, err)
	assert.True(t, invocations > 5, "expected at least 5 invocations, got %v", invocations)
}
