package logger

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	With(args ...interface{}) Logger
	WithContext(ctx context.Context) Logger
}

type ZapLogger struct {
	logger *zap.Logger
}

func NewLogger() (Logger, error) {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.MessageKey = "message"
	config.EncoderConfig.CallerKey = "caller"
	config.EncoderConfig.StacktraceKey = "stacktrace"
	config.EncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return &ZapLogger{logger: logger}, nil
}

func NewStdLogger() Logger {
	logger, err := NewLogger()
	if err != nil {
		zapLogger, _ := zap.NewProduction()
		return &ZapLogger{logger: zapLogger}
	}
	return logger
}

func (l *ZapLogger) Debug(msg string, args ...interface{}) {
	fields := l.convertArgs(args...)
	l.logger.Debug(msg, fields...)
}

func (l *ZapLogger) Info(msg string, args ...interface{}) {
	fields := l.convertArgs(args...)
	l.logger.Info(msg, fields...)
}

func (l *ZapLogger) Warn(msg string, args ...interface{}) {
	fields := l.convertArgs(args...)
	l.logger.Warn(msg, fields...)
}

func (l *ZapLogger) Error(msg string, args ...interface{}) {
	fields := l.convertArgs(args...)
	l.logger.Error(msg, fields...)
}

func (l *ZapLogger) With(args ...interface{}) Logger {
	fields := l.convertArgs(args...)
	return &ZapLogger{logger: l.logger.With(fields...)}
}

func (l *ZapLogger) WithContext(ctx context.Context) Logger {
	return l
}

func (l *ZapLogger) convertArgs(args ...interface{}) []zap.Field {
	if len(args) == 0 {
		return nil
	}

	fields := make([]zap.Field, 0, len(args)/2)
	for i := 0; i < len(args)-1; i += 2 {
		key, ok := args[i].(string)
		if !ok {
			continue
		}
		value := args[i+1]
		fields = append(fields, zap.Any(key, value))
	}

	return fields
}
