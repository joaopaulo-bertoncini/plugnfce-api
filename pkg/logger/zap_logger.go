package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapLogger implements Logger interface using zap
type ZapLogger struct {
	logger *zap.Logger
}

// NewZapLogger creates a new zap logger instance
func NewZapLogger() Logger {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, _ := config.Build()
	return &ZapLogger{logger: logger}
}

// Debug logs a debug message
func (l *ZapLogger) Debug(msg string, fields ...Field) {
	zapFields := make([]zap.Field, len(fields))
	for i, field := range fields {
		zapFields[i] = zap.Any(field.Key, field.Value)
	}
	l.logger.Debug(msg, zapFields...)
}

// Info logs an info message
func (l *ZapLogger) Info(msg string, fields ...Field) {
	zapFields := make([]zap.Field, len(fields))
	for i, field := range fields {
		zapFields[i] = zap.Any(field.Key, field.Value)
	}
	l.logger.Info(msg, zapFields...)
}

// Warn logs a warning message
func (l *ZapLogger) Warn(msg string, fields ...Field) {
	zapFields := make([]zap.Field, len(fields))
	for i, field := range fields {
		zapFields[i] = zap.Any(field.Key, field.Value)
	}
	l.logger.Warn(msg, zapFields...)
}

// Error logs an error message
func (l *ZapLogger) Error(msg string, fields ...Field) {
	zapFields := make([]zap.Field, len(fields))
	for i, field := range fields {
		zapFields[i] = zap.Any(field.Key, field.Value)
	}
	l.logger.Error(msg, zapFields...)
}

// Fatal logs a fatal message and exits
func (l *ZapLogger) Fatal(msg string, fields ...Field) {
	zapFields := make([]zap.Field, len(fields))
	for i, field := range fields {
		zapFields[i] = zap.Any(field.Key, field.Value)
	}
	l.logger.Fatal(msg, zapFields...)
}
