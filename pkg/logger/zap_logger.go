package logger

import (
	"fmt"

	"github.com/Serendipity565/GrabSeat/config"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewZapLogger(conf *config.LogConfig) *ZapLogger {
	level := InfoLevel

	al := zap.NewAtomicLevelAt(level)
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = zapcore.RFC3339TimeEncoder

	lumberJackLogger := &lumberjack.Logger{
		Filename:   conf.File,
		MaxSize:    conf.MaxSize,    // 在进行切割之前，日志文件的最大大小（以MB为单位）
		MaxBackups: conf.MaxBackups, // 保留旧文件的最大个数
		MaxAge:     conf.MaxAge,     // 保留旧文件的最大天数
		Compress:   conf.Compress,   // 是否压缩/归档旧文件
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(cfg),
		zapcore.AddSync(lumberJackLogger),
		al,
	)
	return &ZapLogger{l: zap.New(core), al: &al}
}

type ZapLogger struct {
	l  *zap.Logger
	al *zap.AtomicLevel
}

func (l *ZapLogger) SetLevel(level Level) {
	if l.al != nil {
		l.al.SetLevel(level)
	}
}

func (l *ZapLogger) Debug(msg string, fields ...Field) {
	l.l.Debug(msg, fields...)
}

func (l *ZapLogger) Debugf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.l.Debug(msg, nil...)
}

func (l *ZapLogger) Info(msg string, fields ...Field) {
	l.l.Info(msg, fields...)
}

func (l *ZapLogger) Infof(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.l.Info(msg, nil...)
}

func (l *ZapLogger) Warn(msg string, fields ...Field) {
	l.l.Warn(msg, fields...)
}

func (l *ZapLogger) Warnf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.l.Warn(msg, nil...)
}

func (l *ZapLogger) Error(msg string, fields ...Field) {
	l.l.Error(msg, fields...)
}

func (l *ZapLogger) Errorf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.l.Error(msg, nil...)
}

func (l *ZapLogger) Panic(msg string, fields ...Field) {
	l.l.Panic(msg, fields...)
}

func (l *ZapLogger) Panicf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.l.Panic(msg, nil...)
}

func (l *ZapLogger) Fatal(msg string, fields ...Field) {
	l.l.Fatal(msg, fields...)
}

func (l *ZapLogger) Fatalf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.l.Fatal(msg, nil...)
}

func (l *ZapLogger) Sync() error {
	return l.l.Sync()
}
