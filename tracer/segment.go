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

func SetDebug(enabled bool) {
	if enabled {
		atomic.StoreInt64(&DebugEnabled, 1)
	} else {
		atomic.StoreInt64(&DebugEnabled, 0)
	}
}

type Segment struct {
	span    opentracing.Span
	records []opentracing.LogRecord
}

func (s *Segment) Finish() {
	s.span.FinishWithOptions(opentracing.FinishOptions{
		FinishTime: time.Now(),
		LogRecords: s.records,
	})
}

func (s *Segment) LogFields(fields ...log.Field) {
	s.span.LogFields(fields...)
}

func (s *Segment) SetBaggageItem(key, value string) {
	s.span.SetBaggageItem(key, value)
}

func (s *Segment) Info(msg string, fields ...log.Field) {
	if s.records == nil {
		s.records = make([]opentracing.LogRecord, 0, 8)
	}

	s.records = append(s.records, opentracing.LogRecord{
		Timestamp: time.Now(),
		Fields:    append(fields, log.String(MessageKey, msg), Caller(CallerKey, 2)),
	})
}

func (s *Segment) Debug(msg string, fields ...log.Field) {
	if DebugEnabled != 0 {
		return
	}

	if s.records == nil {
		s.records = make([]opentracing.LogRecord, 0, 8)
	}

	s.records = append(s.records, opentracing.LogRecord{
		Timestamp: time.Now(),
		Fields:    append(fields, log.String(MessageKey, msg), Caller(CallerKey, 2)),
	})
}

func Caller(key string, skip int) log.Field {
	_, file, line, _ := runtime.Caller(skip)
	return log.String(key, filepath.Base(filepath.Dir(file))+"/"+filepath.Base(file)+":"+strconv.Itoa(line))
}

func StartSpan(ctx context.Context, operationName string, fields ...log.Field) (*Segment, context.Context) {
	span, ctx := opentracing.StartSpanFromContext(ctx, operationName)
	span.LogFields(fields...)
	return &Segment{span: span}, ctx
}
