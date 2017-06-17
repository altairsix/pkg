package tracer_test

import (
	"context"
	"testing"
	"time"

	"github.com/altairsix/pkg/tracer"
	"github.com/opentracing/opentracing-go/log"
)

func TestStartLogger(t *testing.T) {
	func() {
		span, _ := tracer.StartSpan(context.Background(), "StartLogger")
		defer span.Finish()

		span.Info("hello", log.String("key", "value"))
	}()

	time.Sleep(time.Millisecond * 150)
}
