// zap_logger.go
package logger

import (
	"go.uber.org/zap"
)

type ZapLogger struct {
	logger *zap.Logger
}

func NewZapLogger() Logger {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	return &ZapLogger{logger: logger}
}

func (z *ZapLogger) Info(msg string, fields ...Field) {
	z.logger.Info(msg, z.convertFields(fields)...)
}

func (z *ZapLogger) Error(msg string, fields ...Field) {
	z.logger.Error(msg, z.convertFields(fields)...)
}

func (z *ZapLogger) Warn(msg string, fields ...Field) {
	z.logger.Warn(msg, z.convertFields(fields)...)
}

func (z *ZapLogger) convertFields(fields []Field) []zap.Field {
	zapFields := make([]zap.Field, len(fields))
	for i, f := range fields {
		zapFields[i] = zap.Any(f.Key, f.Value)
	}
	return zapFields
}
