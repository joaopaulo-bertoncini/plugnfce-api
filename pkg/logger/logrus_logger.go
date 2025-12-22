// logrus_logger.go
package logger

import (
	"github.com/sirupsen/logrus"
)

type LogrusLogger struct {
	logger *logrus.Logger
}

func NewLogrusLogger() Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	return &LogrusLogger{logger: logger}
}

func (l *LogrusLogger) Info(msg string, fields ...Field) {
	entry := l.logger.WithFields(l.convertFields(fields))
	entry.Info(msg)
}

func (l *LogrusLogger) Error(msg string, fields ...Field) {
	entry := l.logger.WithFields(l.convertFields(fields))
	entry.Error(msg)
}

func (l *LogrusLogger) Warn(msg string, fields ...Field) {
	entry := l.logger.WithFields(l.convertFields(fields))
	entry.Warn(msg)
}

func (l *LogrusLogger) convertFields(fields []Field) logrus.Fields {
	logrusFields := logrus.Fields{}
	for _, f := range fields {
		logrusFields[f.Key] = f.Value
	}
	return logrusFields
}
