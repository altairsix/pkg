package domainx

import (
	"context"
	"time"

	"github.com/altairsix/eventsource"
	"github.com/nats-io/go-nats"
)

const (
	// DefaultTimeout specifies how long to wait for the record to be written
	DefaultTimeout = time.Second * 3
)

type Repository interface {
	Apply(ctx context.Context, command eventsource.Command) (int, error)
}

// CommandHandler wraps the request and listens for a response
type repository struct {
	repo             Repository
	nc               *nats.Conn
	timeout          time.Duration
	aggregateSubject string
}

// Apply executes the command and returns the new aggregate version
func (r *repository) Apply(ctx context.Context, command eventsource.Command) (int, error) {
	subject := r.aggregateSubject + "." + command.AggregateID()
	done := subscribeForUpdates(r.nc, subject, r.timeout)

	version, err := r.repo.Apply(ctx, command)
	if err != nil {
		return 0, err
	}

	r.nc.Publish(r.aggregateSubject, nil)
	<-done

	return version, nil
}

// Wrap takes a repo and wraps it with a CommandHandler to create a semi-consist request
func Wrap(repo Repository, nc *nats.Conn, env, bc string) Repository {
	return &repository{
		repo:             repo,
		nc:               nc,
		timeout:          DefaultTimeout,
		aggregateSubject: AggregateSubject(env, bc),
	}
}
