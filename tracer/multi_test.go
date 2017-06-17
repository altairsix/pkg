package tracer_test

import (
	"testing"
	"time"

	"github.com/altairsix/pkg/tracer"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

func TestMulti(t *testing.T) {
	func() {
		multi := tracer.Multi(tracer.DefaultTracer)

		// 2) Demonstrate simple OpenTracing instrumentation
		parent := multi.StartSpan("Parent")
		for i := 0; i < 20; i++ {
			parent.LogFields(log.Int("starting_child", i))
			child := multi.StartSpan("Child", opentracing.ChildOf(parent.Context()))
			time.Sleep(10 * time.Millisecond)
			child.Finish()
		}
		parent.Finish()
	}()
	time.Sleep(time.Millisecond * 150)
}
