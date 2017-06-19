package tracer

import (
	"sync"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

type Span struct {
	sync.Mutex
	emitter       Emitter
	operationName string
	tracer        *Tracer
	fields        map[string]log.Field
	startedAt     time.Time
	baggage       map[string]string
	tags          map[string]interface{}
}

func (s *Span) clone() *Span {
	var fields map[string]log.Field
	if s.fields != nil {
		fields = make(map[string]log.Field, len(s.fields))
		for k, v := range s.fields {
			fields[k] = v
		}
	}

	var baggage map[string]string
	if s.baggage != nil {
		baggage = make(map[string]string, len(s.baggage))
		for k, v := range baggage {
			baggage[k] = v
		}
	}

	return &Span{
		emitter:       s.emitter,
		operationName: s.operationName,
		tracer:        s.tracer,
		fields:        fields,
		startedAt:     time.Now(),
		baggage:       baggage,
	}
}

// Sets the end timestamp and finalizes Span state.
//
// With the exception of calls to Context() (which are always allowed),
// Finish() must be the last call made to any span instance, and to do
// otherwise leads to undefined behavior.
func (s *Span) Finish() {
	s.FinishWithOptions(opentracing.FinishOptions{
		FinishTime: time.Now(),
	})
}

// FinishWithOptions is like Finish() but with explicit control over
// timestamps and log data.
func (s *Span) FinishWithOptions(opts opentracing.FinishOptions) {
	if opts.LogRecords != nil {
		for _, record := range opts.LogRecords {
			s.emitter.Emit(s, "", record.Fields...)
		}
	}

	elapsed := time.Now().Sub(s.startedAt) / time.Millisecond
	s.emitter.Emit(s, s.operationName, log.Int64("elapsed", int64(elapsed)))
}

// Context() yields the SpanContext for this Span. Note that the return
// value of Context() is still valid after a call to Span.Finish(), as is
// a call to Span.Context() after a call to Span.Finish().
func (s *Span) Context() opentracing.SpanContext {
	return s
}

// Sets or changes the operation name.
func (s *Span) SetOperationName(operationName string) opentracing.Span {
	dupe := s.clone()
	dupe.operationName = operationName
	return dupe
}

// Adds a tag to the span.
//
// If there is a pre-existing tag set for `key`, it is overwritten.
//
// Tag values can be numeric types, strings, or bools. The behavior of
// other tag value types is undefined at the OpenTracing level. If a
// tracing system does not know how to handle a particular value type, it
// may ignore the tag, but shall not panic.
func (s *Span) SetTag(key string, value interface{}) opentracing.Span {
	s.Lock()
	defer s.Unlock()

	if s.tags == nil {
		s.tags = make(map[string]interface{}, 1)
	}

	return nil
}

// LogFields is an efficient and type-checked way to record key:value
// logging data about a Span, though the programming interface is a little
// more verbose than LogKV(). Here's an example:
//
//    span.LogFields(
//        log.String("event", "soft error"),
//        log.String("type", "cache timeout"),
//        log.Int("waited.millis", 1500))
//
// Also see Span.FinishWithOptions() and FinishOptions.BulkLogData.
func (s *Span) LogFields(fields ...log.Field) {
	s.Lock()
	defer s.Unlock()

	if s.fields == nil {
		s.fields = map[string]log.Field{}
	}

	for _, field := range fields {
		s.fields[field.Key()] = field
	}
}

// LogKV is a concise, readable way to record key:value logging data about
// a Span, though unfortunately this also makes it less efficient and less
// type-safe than LogFields(). Here's an example:
//
//    span.LogKV(
//        "event", "soft error",
//        "type", "cache timeout",
//        "waited.millis", 1500)
//
// For LogKV (as opposed to LogFields()), the parameters must appear as
// key-value pairs, like
//
//    span.LogKV(key1, val1, key2, val2, key3, val3, ...)
//
// The keys must all be strings. The values may be strings, numeric types,
// bools, Go error instances, or arbitrary structs.
//
// (Note to implementors: consider the log.InterleavedKVToFields() helper)
func (s *Span) LogKV(alternatingKeyValues ...interface{}) {
	fields, err := log.InterleavedKVToFields(alternatingKeyValues...)
	if err != nil {
		panic("LogKV requires an even number of parameters")
	}
	s.LogFields(fields...)
}

// SetBaggageItem sets a key:value pair on this Span and its SpanContext
// that also propagates to descendants of this Span.
//
// SetBaggageItem() enables powerful functionality given a full-stack
// opentracing integration (e.g., arbitrary application data from a mobile
// app can make it, transparently, all the way into the depths of a storage
// system), and with it some powerful costs: use this feature with care.
//
// IMPORTANT NOTE #1: SetBaggageItem() will only propagate baggage items to
// *future* causal descendants of the associated Span.
//
// IMPORTANT NOTE #2: Use this thoughtfully and with care. Every key and
// value is copied into every local *and remote* child of the associated
// Span, and that can add up to a lot of network and cpu overhead.
//
// Returns a reference to this Span for chaining.
func (s *Span) SetBaggageItem(restrictedKey, value string) opentracing.Span {
	s.Lock()
	defer s.Unlock()

	if s.baggage == nil {
		s.baggage = map[string]string{}
	}

	s.baggage[restrictedKey] = value
	return s
}

// Gets the value for a baggage item given its key. Returns the empty string
// if the value isn't found in this Span.
func (s *Span) BaggageItem(restrictedKey string) string {
	s.Lock()
	defer s.Unlock()

	if s.baggage == nil {
		return ""
	}

	return s.baggage[restrictedKey]
}

// Provides access to the Tracer that created this Span.
func (s *Span) Tracer() opentracing.Tracer {
	return s.tracer
}

// Deprecated: use LogFields or LogKV
func (s *Span) LogEvent(event string) {
	panic("LogEvent: Deprecated: use LogFields or LogKV")
}

// Deprecated: use LogFields or LogKV
func (s *Span) LogEventWithPayload(event string, payload interface{}) {
	panic("LogEventWithPayload: Deprecated: use LogFields or LogKV")
}

// Deprecated: use LogFields or LogKV
func (s *Span) Log(data opentracing.LogData) {
	panic("Log: Deprecated: use LogFields or LogKV")
}

// ForeachBaggageItem grants access to all baggage items stored in the
// SpanContext.
// The handler function will be called for each baggage key/value pair.
// The ordering of items is not guaranteed.
//
// The bool return value indicates if the handler wants to continue iterating
// through the rest of the baggage items; for example if the handler is trying to
// find some baggage item by pattern matching the name, it can return false
// as soon as the item is found to stop further iterations.
func (s *Span) ForeachBaggageItem(handler func(k, v string) bool) {
	if s.baggage == nil {
		return
	}

	for k, v := range s.baggage {
		if ok := handler(k, v); !ok {
			return
		}
	}
}

func (s *Span) ForeachField(handler func(k string, f log.Field) bool) {
	if s.fields == nil {
		return
	}

	for k, v := range s.fields {
		if ok := handler(k, v); !ok {
			return
		}
	}
}

// Info allows the Span to log at arbitrary times
func (s *Span) Info(msg string, fields ...log.Field) {
	s.emitter.Emit(s, msg, fields...)
}

// Debug allows the Span to log at arbitrary times
func (s *Span) Debug(msg string, fields ...log.Field) {
	s.emitter.Emit(s, msg, fields...)
}
