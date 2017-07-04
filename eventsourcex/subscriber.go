package eventsourcex

import (
	"context"

	"github.com/nats-io/go-nats"
	"github.com/nats-io/go-nats-streaming"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
)

// SubscribeNotices listens for notices for a specific bounded context.  Notices are published when the
// caller of a command would like a consistent read after writer.  The notice provides eventually
// consistent services an opportunity to update the read model immediately.
func SubscribeNotices(ctx context.Context, nc *nats.Conn, env, boundedContext string, fn func(id string)) error {
	subject := NoticesSubject(env, boundedContext)
	sub, err := nc.Subscribe(subject, func(m *nats.Msg) {
		fn(string(m.Data))
	})
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()

	<-ctx.Done()
	return nil
}

// SubscribeStream subscribes to a nats stream for the specified bounded context
func SubscribeStream(ctx context.Context, nc *nats.Conn, cp Checkpointer, env, boundedContext string, fn func(*stan.Msg)) (<-chan struct{}, error) {
	st, err := stan.Connect(ClusterID, ksuid.New().String(), stan.NatsConn(nc))
	if err != nil {
		return nil, errors.Wrapf(err, "unable to connect to cluster, %v", ClusterID)
	}

	subject := StreamSubject(env, boundedContext)
	cpKey := makeCheckpointKey(env, boundedContext)

	sequence, err := cp.Load(ctx, cpKey)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to load checkpoint, %v", cpKey)
	}

	sub, err := st.Subscribe(subject, fn, stan.StartAtSequence(sequence))
	if err != nil {
		return nil, errors.Wrapf(err, "unable to scribe to stan subject, %v", subject)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		defer sub.Unsubscribe()
		<-ctx.Done()
	}()

	return done, nil
}
