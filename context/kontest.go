package context

import (
	gocontext "context"
	"log"
	"net/http"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	defaultLogger *zap.Logger
)

func init() {
	fields := zap.Fields(
		zap.Stringer("timestamp", &timestamp{}),
	)
	logger, err := zap.NewProduction(fields)
	if err != nil {
		log.Fatalln(err)
	}

	defaultLogger = logger
}

type Kontext struct {
	Context gocontext.Context
	Logger  *zap.Logger
}

func (k Kontext) Value(key interface{}) interface{} {
	return k.Context.Value(key)
}

func WithCancel(k Kontext) (Kontext, func()) {
	ctx, cancel := gocontext.WithCancel(k.Context)
	return Kontext{
		Context: ctx,
		Logger:  k.Logger,
	}, cancel
}

func WithTimeout(k Kontext, d time.Duration) (Kontext, func()) {
	ctx, cancel := gocontext.WithCancel(k.Context)
	return Kontext{
		Context: ctx,
		Logger:  k.Logger,
	}, cancel
}

func WithValue(k Kontext, key, val interface{}) Kontext {
	return Kontext{
		Context: gocontext.WithValue(k.Context, key, val),
		Logger:  k.Logger,
	}
}

func Request(req *http.Request) Kontext {
	return Kontext{
		Context: req.Context(),
		Logger:  defaultLogger,
	}
}

func (k Kontext) Info(msg string, fields ...zapcore.Field) {
	k.Logger.Info(msg, fields...)
}

func (k Kontext) Debug(msg string, fields ...zapcore.Field) {
	k.Logger.Debug(msg, fields...)
}

type timestamp struct {
}

func (t *timestamp) String() string {
	return time.Now().Format(time.RFC3339)
}

func Background(env string) Kontext {
	fields := zap.Fields(
		zap.Stringer("timestamp", &timestamp{}),
		zap.String("env", env),
	)

	config := zap.NewProductionConfig()
	if env == "local" {
		config = zap.NewDevelopmentConfig()
	}

	logger, err := config.Build(fields)
	if err != nil {
		log.Fatalln(err)
	}

	return Kontext{
		Context: gocontext.Background(),
		Logger:  logger,
	}
}
