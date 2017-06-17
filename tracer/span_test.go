package tracer_test

import (
	"context"
	"testing"

	"github.com/altairsix/pkg/tracer"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

func init() {
	opentracing.InitGlobalTracer(tracer.DefaultTracer)
}

func TestTracer(t *testing.T) {
	root := context.Background()

	parentSpan, parentCtx := opentracing.StartSpanFromContext(root, "op")
	defer parentSpan.Finish()

	parentSpan.SetBaggageItem("argle", "bargle")
	parentSpan.LogFields(log.String("a", "b"))

	childSpan, _ := opentracing.StartSpanFromContext(parentCtx, "op")
	defer childSpan.Finish()

	childSpan.LogFields(log.String("hello", "world"))
}
