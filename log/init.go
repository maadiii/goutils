package log

import (
	"context"
	"time"

	"github.com/maadiii/utils"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Config struct {
	ServiceName string
	Env         string
	Level       Level
	Writer      WriterConfig
	QueueSize   int
	BatchSize   int
	BatchDur    time.Duration
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

func WithContext(ctx context.Context, logger utils.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

func FromContext(ctx context.Context) utils.Logger {
	return ctx.Value(loggerKey{}).(utils.Logger)
}
