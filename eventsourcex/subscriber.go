package eventsourcex

import (
	"context"

	"github.com/altairsix/pkg/action"
	"github.com/altairsix/pkg/action/ticker"
	"github.com/nats-io/go-nats"
)

// Subscribe listens for update events
func Subscribe(ctx context.Context, nc *nats.Conn, env, boundedContext string, fn func(id string)) error {
	subject := AggregateSubject(env, boundedContext)
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

// SubscribeSingleton start a single singleton across the cluster
func SubscribeSingleton(ctx context.Context, nc *nats.Conn, env, boundedContext, key string, fn func(id string)) error {
	subject := makeTickerSubject(env, boundedContext, key)
	t := ticker.Nats(nc, subject)
	singleton := action.Singleton(t, func(ctx context.Context) error {
		return Subscribe(ctx, nc, env, boundedContext, fn)
	})

	return singleton(ctx)
}
