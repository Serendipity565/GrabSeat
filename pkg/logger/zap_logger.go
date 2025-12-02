package logger

import (
	"go.uber.org/zap"
)

type ZapLogger struct {
	L  *zap.Logger
	Al *zap.AtomicLevel
}

func (l *ZapLogger) Debug(msg string, fields ...Field) {
	l.L.Debug(msg, fields...)
}

func (l *ZapLogger) Info(msg string, fields ...Field) {
	l.L.Info(msg, fields...)
}

func (l *ZapLogger) Warn(msg string, fields ...Field) {
	l.L.Warn(msg, fields...)
}

func (l *ZapLogger) Error(msg string, fields ...Field) {
	l.L.Error(msg, fields...)
}

func (l *ZapLogger) Panic(msg string, fields ...Field) {
	l.L.Panic(msg, fields...)
}

func (l *ZapLogger) Fatal(msg string, fields ...Field) {
	l.L.Fatal(msg, fields...)
}

func (l *ZapLogger) Sync() error {
	return l.L.Sync()
}
