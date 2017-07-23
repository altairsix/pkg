package eventsourcex

import (
	"context"
	"reflect"
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
	DefaultTimeout = time.Second * 5
)

// StreamSubject returns the streaming subject for a specific bounded context
func StreamSubject(env, boundedContext string) string {
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
		t := reflect.TypeOf(cmd)
		for t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		name := t.Name()
		segment, ctx := tracer.NewSegment(ctx, "eventsource.apply_command",
			log.String("cmd", name),
			log.String("id", cmd.AggregateID()),
		)
		defer segment.Finish()

		return repo.Apply(ctx, cmd)
	})
}

// WithConsistentRead provides a faux consistent read.  Should wrap WithNotifier to ensure that
// the NoticesSubject.{ID} is subscribed to prior to the command being executed.
func WithConsistentRead(repo Repository, nc *nats.Conn, env, boundedContext string) Repository {
	subject := NoticesSubject(env, boundedContext)

	return RepositoryFunc(func(ctx context.Context, cmd eventsource.Command) (int, error) {
		version, err := repo.Apply(ctx, cmd)
		if err != nil {
			return 0, err
		}

		child, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()

		segment := tracer.SegmentFromContext(ctx)
		segment.Info("eventsource.publish_notice",
			log.String("subject", subject),
		)
		nc.RequestWithContext(child, subject, []byte(cmd.AggregateID()))

		return version, nil
	})
}
