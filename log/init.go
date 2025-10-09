package log

import (
	"context"
	"io"

	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger interface {
	Debug(ctx context.Context, msg string, fields ...any)
	Info(ctx context.Context, msg string, fields ...any)
	Warn(ctx context.Context, msg string, fields ...any)
	Error(ctx context.Context, msg string, fields ...any)
	Sync() error
}

// Config holds logger configuration
type Config struct {
	Level            Level
	Writers          []io.Writer // multi-writer
	Async            bool
	QueueSize        int
	UseLumberjack    bool              // enable lumberjack rotation
	LumberjackConfig lumberjack.Logger // pass lumberjack config if needed
}

// Level defines log severity levels.
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
