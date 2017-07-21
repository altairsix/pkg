package notice

import (
	"context"
	"time"

	"github.com/altairsix/eventsource"
	"github.com/altairsix/pkg/eventsourcex"
	"github.com/nats-io/go-nats"
)

// WithConsistentRead provides a faux consistent read.  Should wrap WithNotifier to ensure that
// the NoticesSubject.{ID} is subscribed to prior to the command being executed.
func WithConsistentRead(repo eventsourcex.Repository, nc *nats.Conn, subject string, timeout time.Duration) eventsourcex.Repository {
	return eventsourcex.RepositoryFunc(func(ctx context.Context, cmd eventsource.Command) (int, error) {
		version, err := repo.Apply(ctx, cmd)
		if err != nil {
			return 0, err
		}

		child, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		nc.RequestWithContext(child, subject, []byte(cmd.AggregateID()))

		return version, nil
	})
}
