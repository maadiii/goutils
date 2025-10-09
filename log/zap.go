package log

import (
	"context"
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// ZapLogger implements Logger interface
type ZapLogger struct {
	logger *zap.Logger
	async  bool
	ch     chan func()
	done   chan struct{}
}

// New creates a new ZapLogger with multi-writer and optional Lumberjack
func New(cfg *Config) Logger {
	var writers []io.Writer

	if len(cfg.Writers) == 0 && !cfg.UseLumberjack {
		writers = []io.Writer{os.Stdout}
	} else {
		writers = cfg.Writers
	}

	if cfg.UseLumberjack {
		lumberjackWriter := &lumberjack.Logger{
			Filename:   cfg.LumberjackConfig.Filename,
			MaxSize:    cfg.LumberjackConfig.MaxSize,
			MaxBackups: cfg.LumberjackConfig.MaxBackups,
			MaxAge:     cfg.LumberjackConfig.MaxAge,
			Compress:   cfg.LumberjackConfig.Compress,
		}
		writers = append(writers, lumberjackWriter)
	}

	if cfg.QueueSize <= 0 {
		cfg.QueueSize = 1024
	}

	multi := io.MultiWriter(writers...)

	encoderCfg := zapcore.EncoderConfig{
		TimeKey:      "time",
		LevelKey:     "level",
		NameKey:      "logger",
		MessageKey:   "msg",
		CallerKey:    "caller",
		EncodeTime:   zapcore.ISO8601TimeEncoder,
		EncodeLevel:  zapcore.CapitalColorLevelEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
	}

	level, err := zapcore.ParseLevel(cfg.Level.String())
	if err != nil {
		panic(err)
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.AddSync(multi),
		level,
	)

	zl := &ZapLogger{
		logger: zap.New(core),
		async:  cfg.Async,
	}

	if cfg.Async {
		zl.ch = make(chan func(), cfg.QueueSize)
		zl.done = make(chan struct{})
		go zl.worker()
	}

	return zl
}

func (z *ZapLogger) Debug(ctx context.Context, msg string, fields ...any) {
	z.log(func() {
		z.logger.Sugar().Debugw(msg, fields...)
	})
}

func (z *ZapLogger) Info(ctx context.Context, msg string, fields ...any) {
	z.log(func() {
		z.logger.Sugar().Infow(msg, fields...)
	})
}

func (z *ZapLogger) Warn(ctx context.Context, msg string, fields ...any) {
	z.log(func() {
		z.logger.Sugar().Warnw(msg, fields...)
	})
}

func (z *ZapLogger) Error(ctx context.Context, msg string, fields ...any) {
	z.log(func() {
		z.logger.Sugar().Errorw(msg, fields...)
	})
}

func (z *ZapLogger) Sync() error {
	if z.async {
		close(z.ch)
		<-z.done
	}
	return z.logger.Sync()
}

func (z *ZapLogger) log(fn func()) {
	if z.async {
		z.enqueue(fn)
	} else {
		fn()
	}
}

func (z *ZapLogger) worker() {
	for fn := range z.ch {
		fn()
	}
	close(z.done)
}

func (z *ZapLogger) enqueue(fn func()) {
	select {
	case z.ch <- fn:
	default:
		// Drop log if queue full
	}
}
