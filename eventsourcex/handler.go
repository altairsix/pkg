package eventsourcex

import (
	"context"
	"time"

	"github.com/altairsix/eventsource"
	nats "github.com/nats-io/go-nats"
	"github.com/nats-io/go-nats-streaming"
	"github.com/pkg/errors"
)

// Processor performs operations on a stream of events, usually in conjunction with SubscribeStream
type Processor func(ctx context.Context, events ...eventsource.Event) error

// Do is a helper method to invoke the Processor
func (fn Processor) Do(ctx context.Context, events ...eventsource.Event) error {
	return fn(ctx, events...)
}

// WithSendNotices publishes an event to the notices subject if the processor executes successfully
func WithSendNotices(p Processor, nc *nats.Conn, env, boundedContext, source string) Processor {
	return func(ctx context.Context, events ...eventsource.Event) error {
		if err := p.Do(ctx, events...); err != nil {
			return err
		}

		go func() {
			for _, event := range events {
				subject := NoticesSubject(env, boundedContext, event.AggregateID())
				nc.Publish(subject, []byte(source))
			}
		}()

		return nil
	}
}

// Unmarshaler accepts a []byte encoded event and returns an event
type Unmarshaler func([]byte) (eventsource.Event, error)

// Checkpointer persists and retrieves nats streaming offsets
type Checkpointer interface {
	// Load retrieves the specified nats streaming offset
	Load(ctx context.Context, key string) (uint64, error)

	// Save persists the specified nats streaming offset
	Save(ctx context.Context, key string, offset uint64) error
}

// MessageHandler encapsulates a nats streaming processor that performs buffered processing
type MessageHandler struct {
	ctx        context.Context
	cancel     func()
	done       chan struct{}
	processor  Processor
	unmarshal  Unmarshaler
	cp         Checkpointer
	cpKey      string
	ch         chan *stan.Msg
	interval   time.Duration
	bufferSize int
}

// MessageHandlerOption allows options to be specified for NewMessageHandler
type MessageHandlerOption func(m *MessageHandler)

// WithBufferSize specifies the max number of messages to be called before callng the Processor
func WithBufferSize(in int) MessageHandlerOption {
	return func(m *MessageHandler) {
		m.bufferSize = in
	}
}

// WithInterval specifies the maximum amount of time that can elapse until we try to flush
// any received events
func WithInterval(d time.Duration) MessageHandlerOption {
	return func(m *MessageHandler) {
		m.interval = d
	}
}

// Done returns a chan that signals when all the resources used by MessageHandler have been released
func (m *MessageHandler) Done() <-chan struct{} {
	return m.done
}

func (m *MessageHandler) flush(data ...*stan.Msg) error {
	if len(data) == 0 {
		return nil
	}

	events, sequence, err := bytesToEvents(m.unmarshal, data...)
	if err != nil {
		return errors.Wrap(err, "unable to convert []byte to events")
	}

	if err := m.processor.Do(m.ctx, events...); err != nil {
		return errors.Wrap(err, "unable to process events")
	}

	if err := m.cp.Save(m.ctx, m.cpKey, sequence); err != nil {
		return errors.Wrapf(err, "unable to save checkpoint, %v %v", m.cpKey, sequence)
	}

	return nil
}

func (m *MessageHandler) start() {
	defer close(m.done)
	defer close(m.ch)
	defer m.cancel()

	t := time.NewTicker(m.interval)
	defer t.Stop()

	size := m.bufferSize
	buffer := make([]*stan.Msg, size)
	offset := 0

	for {
		select {
		case v := <-m.ch:
			buffer[offset] = v
			offset++
			if offset == size {
				m.flush(buffer[0:offset]...)
				offset = 0
			}

		case <-t.C:
			m.flush(buffer[0:offset]...)
			offset = 0

		case <-m.ctx.Done():
			m.flush(buffer[0:offset]...)
			offset = 0
			return
		}
	}
}

// Handle provides the nats streaming message handler
func (m *MessageHandler) Handle(msg *stan.Msg) {
	m.ch <- msg
}

// Close releases resources associated with the MessageHandler
func (m *MessageHandler) Close() error {
	m.cancel()
	<-m.done
	return nil
}

// NewMessageHandler constructs a new MessageHandler with the arguments provided
func NewMessageHandler(ctx context.Context, p Processor, u Unmarshaler, cp Checkpointer, cpKey string, opts ...MessageHandlerOption) *MessageHandler {
	child, cancel := context.WithCancel(ctx)

	h := &MessageHandler{
		ctx:        child,
		cancel:     cancel,
		done:       make(chan struct{}),
		processor:  p,
		unmarshal:  u,
		cp:         cp,
		ch:         make(chan *stan.Msg, 256),
		interval:   time.Millisecond * 250,
		bufferSize: 100,
	}

	for _, opt := range opts {
		opt(h)
	}

	go h.start()
	return h
}

func bytesToEvents(unmarshal Unmarshaler, messages ...*stan.Msg) ([]eventsource.Event, uint64, error) {
	events := make([]eventsource.Event, 0, len(messages))

	sequence := uint64(0)
	for _, m := range messages {
		event, err := unmarshal(m.Data)
		if err != nil {
			return nil, 0, errors.Wrapf(err, "unable to unmarshal event at sequence, %v", m.Sequence)
		}

		events = append(events, event)
		sequence = m.Sequence
	}

	return events, sequence, nil
}
