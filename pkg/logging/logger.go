package logging

import (
	"context"
	"log"

	"go.uber.org/zap"
)

const loggerKey = "logger"

var (
	logger *zap.Logger
	err    error
)

// init initializes logger
func init() {
	logger, err = zap.NewProduction()
	if err != nil {
		log.Fatalf("could not init logger: %s", err)
	}
}

// NewContext returns a context has a zap logger with the extra fields added
func NewContext(ctx context.Context, fields ...zap.Field) context.Context {
	return context.WithValue(ctx, loggerKey, WithContext(ctx).With(fields...))
}

// WithContext returns a zap logger with as much context as possible
func WithContext(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return logger
	}

	if ctxLogger, ok := ctx.Value(loggerKey).(*zap.Logger); ok {
		return ctxLogger
	}

	return logger
}
