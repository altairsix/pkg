package tracer

import (
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

type multiSpan struct {
	tracer opentracing.Tracer
	spans  []opentracing.Span
}

func (m *multiSpan) Map(fn func(span opentracing.Span) opentracing.Span) opentracing.Span {
	spans := make([]opentracing.Span, 0, len(m.spans))
	for _, span := range spans {
		spans = append(spans, span)
	}

	return &multiSpan{
		tracer: m.tracer,
		spans:  spans,
	}
}

func (m *multiSpan) ChildSpans() []opentracing.Span {
	return m.spans
}

func (m *multiSpan) ForeachBaggageItem(handler func(k, v string) bool) {
	if len(m.spans) == 0 {
		return
	}

	m.spans[0].Context().ForeachBaggageItem(handler)
}

func (m *multiSpan) Finish() {
	for _, span := range m.spans {
		span.Finish()
	}
}

func (m *multiSpan) FinishWithOptions(opts opentracing.FinishOptions) {
	for _, span := range m.spans {
		span.FinishWithOptions(opts)
	}
}

func (m *multiSpan) Context() opentracing.SpanContext {
	return m
}

func (m *multiSpan) SetOperationName(operationName string) opentracing.Span {
	return m.Map(func(span opentracing.Span) opentracing.Span {
		return span.SetOperationName(operationName)
	})
}

func (m *multiSpan) SetTag(key string, value interface{}) opentracing.Span {
	return m.Map(func(span opentracing.Span) opentracing.Span {
		return span.SetTag(key, value)
	})
}

func (m *multiSpan) LogFields(fields ...log.Field) {
	for _, span := range m.spans {
		span.LogFields(fields...)
	}
}

func (m *multiSpan) LogKV(alternatingKeyValues ...interface{}) {
	for _, span := range m.spans {
		span.LogKV(alternatingKeyValues...)
	}
}

func (m *multiSpan) SetBaggageItem(restrictedKey, value string) opentracing.Span {
	return m.Map(func(span opentracing.Span) opentracing.Span {
		return span.SetBaggageItem(restrictedKey, value)
	})
}

func (m *multiSpan) BaggageItem(restrictedKey string) string {
	for _, span := range m.spans {
		return span.BaggageItem(restrictedKey)
	}
	return ""
}

func (m *multiSpan) Tracer() opentracing.Tracer {
	return m.tracer
}

// Deprecated: use LogFields or LogKV
func (m *multiSpan) LogEvent(event string) {
	for _, span := range m.spans {
		span.LogEvent(event)
	}
}

// Deprecated: use LogFields or LogKV
func (m *multiSpan) LogEventWithPayload(event string, payload interface{}) {
	for _, span := range m.spans {
		span.LogEventWithPayload(event, payload)
	}
}

// Deprecated: use LogFields or LogKV
func (m *multiSpan) Log(data opentracing.LogData) {
	for _, span := range m.spans {
		span.Log(data)
	}
}

func (m *multiSpan) Info(msg string, fields ...log.Field) {
	for _, span := range m.spans {
		if v, ok := span.(logger); ok {
			v.Info(msg, fields...)
		}
	}
}

func (m *multiSpan) Debug(msg string, fields ...log.Field) {
	for _, span := range m.spans {
		if v, ok := span.(logger); ok {
			v.Debug(msg, fields...)
		}
	}
}

type multiTracer struct {
	tracers []opentracing.Tracer
}

func (m *multiTracer) StartSpan(operationName string, opts ...opentracing.StartSpanOption) opentracing.Span {
	spans := make([]opentracing.Span, len(m.tracers))

	for index, t := range m.tracers {
		span := t.StartSpan(operationName, opts...)
		spans[index] = span
	}

	return &multiSpan{
		tracer: m,
		spans:  spans,
	}
}

func (m *multiTracer) Inject(sm opentracing.SpanContext, format interface{}, carrier interface{}) error {
	panic("not implemented")
}

func (m *multiTracer) Extract(format interface{}, carrier interface{}) (opentracing.SpanContext, error) {
	panic("not implemented")
}

func Multi(tracers ...opentracing.Tracer) opentracing.Tracer {
	return &multiTracer{tracers: tracers}
}
