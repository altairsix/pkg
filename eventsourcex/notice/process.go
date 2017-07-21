package notice

import (
	"context"
	"sync/atomic"
)

// Handler accepts a notice and performs the necessary operations to update the read model
type Handler interface {
	// Process will be called as notices are received
	Process(ctx context.Context, notice Message)
}

// HandlerFunc provides a convenient func wrapper around Handler
type HandlerFunc func(ctx context.Context, notice Message)

// Process implements the Handler interface
func (fn HandlerFunc) Process(ctx context.Context, notice Message) {
	fn(ctx, notice)
}

// Process reads from the channel and invokes the Handler.  Process exits when either the context is canceled or
// the channel is closed.
//
// Process guarantees that only one invocation of a Handler will operate upon a given aggregateID at a time
func Process(ctx context.Context, ch <-chan Message, handler Handler) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	inFlight := map[string]struct{}{} // keeps track of work to prevent duplication of effort
	inFlightCount := int32(0)

	done := make(chan struct{})
	finished := make(chan string, 128)

	// Wait for in flight tasks to complete
	go func() {
		defer close(done)
		defer close(finished)
		<-ctx.Done()

		for atomic.LoadInt32(&inFlightCount) > 0 {
			<-finished
		}
	}()

	for {
		select {
		case <-ctx.Done():
			<-done
			return

		case id := <-finished:
			delete(inFlight, id)

		case m, ok := <-ch:
			if !ok {
				cancel()
				<-done
				return
			}

			id := m.AggregateID()
			if _, found := inFlight[id]; !found {
				inFlight[id] = struct{}{}
				atomic.AddInt32(&inFlightCount, 1)

				go func(m Message) {
					defer func() { atomic.AddInt32(&inFlightCount, -1) }()
					defer func() { finished <- m.AggregateID() }()
					defer m.Close()
					handler.Process(ctx, m)
				}(m)
			}
		}
	}
}

// ProcessFunc provides a convenience func wrapper around Process
func ProcessFunc(ctx context.Context, ch <-chan Message, h HandlerFunc) {
	Process(ctx, ch, h)
}
