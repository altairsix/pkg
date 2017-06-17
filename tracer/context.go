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

type Logger struct {
	span    opentracing.Span
	records []opentracing.LogRecord
}

func (l *Logger) Finish() {
	l.span.FinishWithOptions(opentracing.FinishOptions{
		FinishTime: time.Now(),
		LogRecords: l.records,
	})
}

func (l *Logger) LogFields(fields ...log.Field) {
	l.span.LogFields(fields...)
}

func (l *Logger) SetBaggageItem(key, value string) {
	l.span.SetBaggageItem(key, value)
}

func (l *Logger) Info(msg string, fields ...log.Field) {
	if l.records == nil {
		l.records = make([]opentracing.LogRecord, 0, 8)
	}

	l.records = append(l.records, opentracing.LogRecord{
		Timestamp: time.Now(),
		Fields:    append(fields, log.String(MessageKey, msg), Caller(CallerKey, 2)),
	})
}

func (l *Logger) Debug(msg string, fields ...log.Field) {
	if DebugEnabled != 0 {
		return
	}

	if l.records == nil {
		l.records = make([]opentracing.LogRecord, 0, 8)
	}

	l.records = append(l.records, opentracing.LogRecord{
		Timestamp: time.Now(),
		Fields:    append(fields, log.String(MessageKey, msg), Caller(CallerKey, 2)),
	})
}

func Caller(key string, skip int) log.Field {
	_, file, line, _ := runtime.Caller(skip)
	return log.String(key, filepath.Base(filepath.Dir(file))+"/"+filepath.Base(file)+":"+strconv.Itoa(line))
}

func StartSpan(ctx context.Context, operationName string, fields ...log.Field) (*Logger, context.Context) {
	span, ctx := opentracing.StartSpanFromContext(ctx, operationName)
	span.LogFields(fields...)
	return &Logger{span: span}, ctx
}
