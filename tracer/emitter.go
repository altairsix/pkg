package tracer

import (
	"github.com/opentracing/opentracing-go/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Emitter interface {
	Emit(span *Span, event string, fields ...log.Field)
}

type ZapEmitter struct {
	Logger      *zap.Logger
	SkipCallers int
}

func (z *ZapEmitter) Emit(span *Span, msg string, fields ...log.Field) {
	encoder := &Encoder{}

	span.ForeachBaggageItem(func(k, v string) bool {
		encoder.EmitString(k, v)
		return true
	})
	span.ForeachField(func(k string, f log.Field) bool {
		f.Marshal(encoder)
		return true
	})
	for _, f := range fields {
		f.Marshal(encoder)
	}

	Caller("caller", z.SkipCallers).Marshal(encoder)

	z.Logger.Info(msg, encoder.Fields...)
}

type Encoder struct {
	Fields []zapcore.Field
}

func (e *Encoder) EmitString(key, value string) {
	e.Fields = append(e.Fields, zap.String(key, value))
}

func (e *Encoder) EmitBool(key string, value bool) {
	e.Fields = append(e.Fields, zap.Bool(key, value))
}

func (e *Encoder) EmitInt(key string, value int) {
	e.Fields = append(e.Fields, zap.Int(key, value))
}

func (e *Encoder) EmitInt32(key string, value int32) {
	e.Fields = append(e.Fields, zap.Int32(key, value))
}

func (e *Encoder) EmitInt64(key string, value int64) {
	e.Fields = append(e.Fields, zap.Int64(key, value))
}

func (e *Encoder) EmitUint32(key string, value uint32) {
	e.Fields = append(e.Fields, zap.Uint32(key, value))
}

func (e *Encoder) EmitUint64(key string, value uint64) {
	e.Fields = append(e.Fields, zap.Uint64(key, value))
}

func (e *Encoder) EmitFloat32(key string, value float32) {
	e.Fields = append(e.Fields, zap.Float32(key, value))
}

func (e *Encoder) EmitFloat64(key string, value float64) {
	e.Fields = append(e.Fields, zap.Float64(key, value))
}

func (e *Encoder) EmitObject(key string, value interface{}) {
	panic("log.Object not supported")
}

func (e *Encoder) EmitLazyLogger(value log.LazyLogger) {
	panic("log.EmitLazyLogger not supported")
}
