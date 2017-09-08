package tracer_test

import (
	"context"
	"testing"

	"github.com/altairsix/pkg/tracer"
	"github.com/opentracing/opentracing-go"
)

func TestStderr(t *testing.T) {
	opentracing.InitGlobalTracer(tracer.StderrTracer)

	segment, _ := tracer.NewSegment(context.Background(), "sample")
	defer segment.Finish()
}
