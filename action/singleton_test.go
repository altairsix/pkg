package action_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/altairsix/pkg/action"
	"github.com/altairsix/pkg/tracer"
	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
)

func init() {
	opentracing.InitGlobalTracer(tracer.DefaultTracer)
}

type MockTicker struct {
	ch chan action.Tick
}

func (m *MockTicker) Publish(tick action.Tick) error {
	return nil
}

func (m *MockTicker) Receive(ctx context.Context) (<-chan action.Tick, error) {
	return m.ch, nil
}

func Run(counter *int32) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		atomic.AddInt32(counter, 1)
		return nil
	}
}

func TestSingleton(t *testing.T) {
	ch := make(chan action.Tick)
	defer close(ch)
	mock := &MockTicker{
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
	fn1 := action.Singleton(mock, Run(&invocations),
		action.WithInterval(interval),
		action.WithElections(interval*3),
		action.WithLease(interval*10),
	)
	err := fn1(ctx)
	assert.Nil(t, err)
	assert.True(t, invocations > 5)
}
