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
		segment, _ := tracer.NewSegment(context.Background(), "StartLogger")
		defer segment.Finish()

		segment.Info("hello", log.String("key", "value"))
	}()

	time.Sleep(time.Millisecond * 150)
}
