package eventsourcex

import (
	"context"
	"strings"
	"time"

	"github.com/altairsix/eventsource"
	"github.com/altairsix/pkg/tracer"
	"github.com/nats-io/go-nats"
	"github.com/opentracing/opentracing-go/log"
)

const (
	// ClusterID specifies the cluster-id of the nats streaming cluster that hosts events
	ClusterID = "events"

	// DefaultTimeout specifies how much time to wait for the eventual consistency
	DefaultTimeout = time.Second * 3
)

// AggregateSubject returns the streaming subject for a specific bounded context
func AggregateSubject(env, boundedContext string) string {
	return env + ".streams.aggregate." + boundedContext
}

// NoticesSubject returns the subject for a specific bounded context
func NoticesSubject(env, boundedContext string, args ...string) string {
	subject := env + ".aggregate." + boundedContext
	if len(args) > 0 {
		subject += "." + strings.Join(args, ".")
	}
	return subject
}

// Repository provides an abstraction over *eventsource.Repository over the
// mutator function
type Repository interface {
	// Apply executes the command against the current version of the aggregate and returns the updated version
	// of the aggregate (or the current version if no updates were made)
	Apply(ctx context.Context, cmd eventsource.Command) (int, error)
}

// RepositoryFunc provides a func helper around Repository
type RepositoryFunc func(ctx context.Context, cmd eventsource.Command) (int, error)

// Apply implements Repository's Apply method
func (fn RepositoryFunc) Apply(ctx context.Context, cmd eventsource.Command) (int, error) {
	return fn(ctx, cmd)
}

// WithTrace logs published events to the tracer
func WithTrace(repo Repository) Repository {
	return RepositoryFunc(func(ctx context.Context, cmd eventsource.Command) (int, error) {
		segment, ctx := tracer.NewSegment(ctx, "repository:trace", log.String("id", cmd.AggregateID()))
		defer segment.Finish()
		return repo.Apply(ctx, cmd)
	})
}

// WithNotifier publishes an event to the NoticesSubject upon the successful completion of an event.
// This enables any listeners to the NoticesSubject to immediately update their stores, providing,
// hopefully a more consistent experience for the user
func WithNotifier(repo Repository, nc *nats.Conn, env, boundedContext string) Repository {
	subject := NoticesSubject(env, boundedContext)

	return RepositoryFunc(func(ctx context.Context, cmd eventsource.Command) (int, error) {
		version, err := repo.Apply(ctx, cmd)
		if err != nil {
			return 0, err
		}

		if env == "local" {
			go func() {
				segment := tracer.SegmentFromContext(ctx)
				segment.Info("repository:notifier:publish", log.String("subject", subject), log.String("id", cmd.AggregateID()))
			}()
		}

		go func() {
			nc.Publish(subject, []byte(cmd.AggregateID()))
		}()

		return version, nil
	})
}

// subscribeForUpdates provides a utility function that waits for any message to be received on
// the specified subject or timeout if no message was received in time
func subscribeForUpdates(nc *nats.Conn, subject string, timeout time.Duration) <-chan struct{} {
	updated := make(chan struct{}, 1)

	sub, err := nc.Subscribe(subject, func(m *nats.Msg) {
		select {
		case updated <- struct{}{}:
		default:
		}
	})
	if err != nil {
		close(updated)
		return updated
	}

	go func() {
		defer close(updated)
		defer sub.Unsubscribe()

		select {
		case <-time.After(timeout):
		case <-updated:
		}
	}()

	return updated
}

// WithConsistentRead provides a faux consistent read.  Should wrap WithNotifier to ensure that
// the NoticesSubject.{ID} is subscribed to prior to the command being executed.
func WithConsistentRead(repo Repository, nc *nats.Conn, env, boundedContext string) Repository {
	rootSubject := NoticesSubject(env, boundedContext) + "."

	return RepositoryFunc(func(ctx context.Context, cmd eventsource.Command) (int, error) {
		subject := rootSubject + cmd.AggregateID()
		updated := subscribeForUpdates(nc, subject, DefaultTimeout)

		version, err := repo.Apply(ctx, cmd)
		if err != nil {
			return 0, err
		}

		<-updated

		return version, nil
	})
}
