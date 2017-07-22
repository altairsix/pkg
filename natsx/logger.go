package natsx

import (
	"context"

	"github.com/altairsix/pkg/tracer"
	"github.com/altairsix/pkg/tracer/k"
	"github.com/savaki/nats-protobuf"
)

// Log provides the standardized logging service for nats queries
func Log(service string) nats_protobuf.Filter {
	return func(fn nats_protobuf.HandlerFunc) nats_protobuf.HandlerFunc {
		return func(ctx context.Context, m *nats_protobuf.Message) (*nats_protobuf.Message, error) {
			segment, ctx := tracer.NewSegment(ctx, m.Method,
				k.String("service", service),
				k.String("method", m.Method),
			)
			defer segment.Finish()

			out, err := fn(ctx, m)
			if err != nil {
				segment.LogFields(k.Err(err))
				return nil, err
			}

			return out, nil
		}
	}
}
