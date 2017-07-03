package domainx

import (
	"context"

	"github.com/altairsix/eventsource"
	"github.com/altairsix/eventsource-connector/publisher"
	"github.com/altairsix/eventsource-connector/publisher/nats"
	gonats "github.com/nats-io/go-nats"
	"github.com/nats-io/go-nats-streaming"
	"github.com/segmentio/ksuid"
)

const (
	// ClusterID contains the nats streaming cluster that holds our events
	ClusterID = "events"
)

func Start(ctx context.Context, nc *gonats.Conn, cp publisher.Checkpointer, env, bc string, r eventsource.StreamReader) (<-chan struct{}, error) {
	st, err := stan.Connect(ClusterID, ksuid.New().String(), stan.NatsConn(nc))
	if err != nil {
		return nil, err
	}

	subject := AggregateSubject(env, bc)
	key := "stan:" + subject
	h := nats.Handler(st, subject)
	return publisher.Start(ctx, r, h, key, cp)
}
