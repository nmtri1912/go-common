package logger

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

var globalLogger *zapLogger

func init() {
	//default logger
	logger := NewZapLogger(false)
	globalLogger = &zapLogger{
		logger: logger,
	}
}

func InitLogger(production bool) {
	if production {
		logger := NewZapLogger(true)
		globalLogger.logger = logger
	}
}

func L() ILogger {
	return globalLogger
}

func Ctx(ctx context.Context) ILogger {
	span := trace.SpanFromContext(ctx)
	return &spanLogger{
		logger: globalLogger.logger,
		span:   span,
	}
}

func Sync() error {
	return globalLogger.logger.Sync()
}
