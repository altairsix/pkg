package tracer

import (
	"context"
	"path/filepath"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

var DebugEnabled = int64(0)

var (
	NopSegment Segment = nopSegment{}
)

func SetDebug(enabled bool) {
	if enabled {
		atomic.StoreInt64(&DebugEnabled, 1)
	} else {
		atomic.StoreInt64(&DebugEnabled, 0)
	}
}

type logger interface {
	Info(msg string, fields ...log.Field)
	Debug(msg string, fields ...log.Field)
}

type Segment interface {
	Finish()
	LogFields(fields ...log.Field)
	SetBaggageItem(key, value string)
	Info(msg string, fields ...log.Field)
	Debug(msg string, fields ...log.Field)
}

type segment struct {
	span    opentracing.Span
	records []opentracing.LogRecord
}

func (s *segment) Finish() {
	if s == nil {
		return
	}

	s.span.FinishWithOptions(opentracing.FinishOptions{
		FinishTime: time.Now(),
		LogRecords: s.records,
	})
}

func (s *segment) LogFields(fields ...log.Field) {
	if s == nil {
		return
	}

	s.span.LogFields(fields...)
}

func (s *segment) SetBaggageItem(key, value string) {
	if s == nil {
		return
	}

	s.span.SetBaggageItem(key, value)
}

func (s *segment) Info(msg string, fields ...log.Field) {
	if s == nil {
		return
	}

	if v, ok := s.span.(logger); ok {
		v.Info(msg, fields...)
	}
}

func (s *segment) Debug(msg string, fields ...log.Field) {
	if s == nil {
		return
	}

	if v, ok := s.span.(logger); ok {
		v.Debug(msg, fields...)
	}
}

func Caller(key string, skip int) log.Field {
	_, file, line, _ := runtime.Caller(skip)
	return log.String(key, filepath.Base(filepath.Dir(file))+"/"+filepath.Base(file)+":"+strconv.Itoa(line))
}

func NewSegment(ctx context.Context, operationName string, fields ...log.Field) (Segment, context.Context) {
	span, ctx := opentracing.StartSpanFromContext(ctx, operationName)
	span.LogFields(fields...)
	return &segment{span: span}, ctx
}

// SegmentFromContext returns an existing segment from the context
func SegmentFromContext(ctx context.Context) Segment {
	span := opentracing.SpanFromContext(ctx)
	return &segment{span: span}
}

type nopSegment struct{}

func (nopSegment) Finish()                               {}
func (nopSegment) LogFields(fields ...log.Field)         {}
func (nopSegment) SetBaggageItem(key, value string)      {}
func (nopSegment) Info(msg string, fields ...log.Field)  {}
func (nopSegment) Debug(msg string, fields ...log.Field) {}
