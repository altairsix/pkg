package eventsourcex

import (
	"context"
	"time"

	"github.com/altairsix/eventsource"
	nats "github.com/nats-io/go-nats"
)

const (
	// DefaultTimeout specifies how much time to wait for the eventual consistency
	DefaultTimeout = time.Second * 3
)

// AggregateSubject returns the subject for a specific bounded context
func AggregateSubject(env, boundedContext string) string {
	return env + ".aggregate." + boundedContext
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

// WithNotifier publishes an event to the AggregateSubject upon the successful completion of an event.
// This enables any listeners to the AggregateSubject to immediately update their stores, providing,
// hopefully a more consistent experience for the user
func WithNotifier(repo Repository, nc *nats.Conn, env, boundedContext string) Repository {
	subject := AggregateSubject(env, boundedContext)

	return RepositoryFunc(func(ctx context.Context, cmd eventsource.Command) (int, error) {
		version, err := repo.Apply(ctx, cmd)
		if err != nil {
			return 0, err
		}

		go func() {
			nc.Publish(subject, nil)
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
// the AggregateSubject.{ID} is subscribed to prior to the command being executed.
func WithConsistentRead(repo Repository, nc *nats.Conn, env, boundedContext string) Repository {
	rootSubject := AggregateSubject(env, boundedContext) + "."

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
