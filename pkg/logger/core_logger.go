package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type zapLogger struct {
	logger *zap.Logger
}

func (l *zapLogger) Debug(msg string, fields ...zapcore.Field) {
	l.logger.Debug(msg, fields...)
}

func (l *zapLogger) Info(msg string, fields ...zapcore.Field) {
	l.logger.Info(msg, fields...)
}

func (l *zapLogger) Warn(msg string, fields ...zapcore.Field) {
	l.logger.Warn(msg, fields...)
}

func (l *zapLogger) Error(msg string, fields ...zapcore.Field) {
	l.logger.Error(msg, fields...)
}

func (l *zapLogger) Fatal(msg string, fields ...zapcore.Field) {
	l.logger.Fatal(msg, fields...)
}

func (l *zapLogger) Panic(msg string, fields ...zapcore.Field) {
	l.logger.Panic(msg, fields...)
}
