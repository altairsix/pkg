package natz

import (
	"context"
	"time"

	"github.com/altairsix/pkg/types"
	"github.com/nats-io/go-nats"
)

const (
	// aggregateQueue holds name of queue used by SubscribeAggregate
	aggregateQueue = "aggregate"
)

func makeAggregateSubject(env, boundedContext string) string {
	return env + ".aggregate." + boundedContext
}

// SyncAggregate provides a blocking function that publishes the update of a specific aggregate id
func SyncAggregate(nc *nats.Conn, env, boundedContext string, timeout time.Duration) func(context.Context, types.Key) {
	subject := makeAggregateSubject(env, boundedContext)

	return func(ctx context.Context, id types.Key) {
		child, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		nc.RequestWithContext(child, subject, []byte(id.String()))
	}
}

// SubscribeAggregate subscribes to the nats subject that contains updates to a specific bounded context
func SubscribeAggregate(ctx context.Context, nc *nats.Conn, env, boundedContext string, fn func(id types.Key)) {
	subject := makeAggregateSubject(env, boundedContext)

	nc.QueueSubscribe(subject, aggregateQueue, func(msg *nats.Msg) {
		fn(types.Key(msg.Data))

		if msg.Reply == "" {
			return
		}

		nc.Publish(msg.Reply, []byte("ok"))
	})
}
