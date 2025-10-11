package log

import (
	"context"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger interface {
	Debug(ctx context.Context, msg string, fields ...any)
	Info(ctx context.Context, msg string, fields ...any)
	Warn(ctx context.Context, msg string, fields ...any)
	Error(ctx context.Context, msg string, fields ...any)
	Sync() error
}

type Config struct {
	Level     Level
	Writer    WriterConfig
	QueueSize int
	BatchSize int
	BatchDur  time.Duration
}

type WriterConfig struct {
	Stdout     bool
	FileConfig *lumberjack.Logger
}

type Level string

func (l Level) String() string {
	return string(l)
}

const (
	DebugLevel Level = "debug"
	InfoLevel  Level = "info"
	WarnLevel  Level = "warn"
	ErrorLevel Level = "error"
)

type loggerKey struct{}

func WithContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

func FromContext(ctx context.Context) Logger {
	return ctx.Value(loggerKey{}).(Logger)
}
