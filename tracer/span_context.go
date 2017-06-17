package tracer

type SpanContext struct {
	// A probabilistically unique identifier for a [multi-span] trace.
	TraceID uint64

	// A probabilistically unique identifier for a span.
	SpanID uint64

	// Whether the trace is sampled.
	Sampled bool

	// The span's associated baggage.
	Baggage map[string]string // initialized on first use
}

// ForeachBaggageItem belongs to the opentracing.SpanContext interface
func (c SpanContext) ForeachBaggageItem(handler func(k, v string) bool) {
	for k, v := range c.Baggage {
		if !handler(k, v) {
			break
		}
	}
}

func (c SpanContext) WithBaggageItem(key, val string) SpanContext {
	baggage := make(map[string]string, len(c.Baggage)+1)
	if c.Baggage != nil {
		for k, v := range c.Baggage {
			baggage[k] = v
		}
	}

	baggage[key] = val

	// Use positional parameters so the compiler will help catch new fields.
	return SpanContext{
		c.TraceID,
		c.SpanID,
		c.Sampled,
		baggage,
	}
}
