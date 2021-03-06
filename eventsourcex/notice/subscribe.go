package notice

import (
	"context"
	"io"

	"github.com/altairsix/pkg/tracer"
	"github.com/nats-io/go-nats"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
)

const (
	// Group provides the name of the nats group
	Group = "vavende-identities-dbase"
)

// Notice contains a request by the submitted of a command to update a specific read model
type Message interface {
	AggregateID() string
}

type MessageCloser interface {
	Message
	io.Closer
}

type message struct {
	nc  *nats.Conn
	msg *nats.Msg
}

// Close sends the response back to the original requester.  This should be call prior to disposing the Notice
func (m *message) Close() error {
	if m.msg.Reply == "" {
		return nil
	}

	return m.nc.Publish(m.msg.Reply, m.msg.Data)
}

// AggregateID refers the aggregate that was recently updated
func (m *message) AggregateID() string {
	return string(m.msg.Data)
}

// Subscribe listens for notices on the nats subject provided
func Subscribe(ctx context.Context, nc *nats.Conn, subject string, bufferSize int) (<-chan MessageCloser, error) {
	segment, _ := tracer.NewSegment(ctx, "nats.notice_listener")

	ch := make(chan MessageCloser, bufferSize)
	sub, err := nc.QueueSubscribe(subject, Group, func(msg *nats.Msg) {
		select {
		case ch <- &message{nc: nc, msg: msg}:
			segment.Info("nats.notice_received",
				log.String("subject", msg.Subject),
				log.String("id", string(msg.Data)),
			)
		default:
		}
	})
	if err != nil {
		defer close(ch)
		return nil, errors.Wrapf(err, "unable to subscribe to subject, %v", subject)
	}

	go func() {
		defer segment.Finish()
		defer close(ch)
		defer sub.Unsubscribe()
		<-ctx.Done()
	}()

	return ch, nil
}
