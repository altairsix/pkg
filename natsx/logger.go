package natsx

import (
	"context"

	"github.com/altairsix/pkg/tracer"
	"github.com/altairsix/pkg/tracer/k"
	"github.com/savaki/nats-protobuf"
)

// Log provides the standardized logging service for nats queries
func Log(bc string) nats_protobuf.Filter {
	return func(fn nats_protobuf.HandlerFunc) nats_protobuf.HandlerFunc {
		return func(ctx context.Context, subject string, m *nats_protobuf.Message) (*nats_protobuf.Message, error) {
			segment, ctx := tracer.NewSegment(ctx, "nats.api",
				k.String("subject", subject),
				k.String("bc", bc),
				k.String("method", m.Method),
			)
			defer segment.Finish()

			out, err := fn(ctx, subject, m)
			if err != nil {
				segment.LogFields(k.Err(err))
				return nil, err
			}

			return out, nil
		}
	}
}
