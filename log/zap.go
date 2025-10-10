package log

import (
	"context"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type zapLogger struct {
	logger    *zap.Logger
	async     bool
	ch        chan func()
	done      chan struct{}
	batchSize int
	batchDur  time.Duration
}

func Zap(cfg Config) Logger {
	cores := make([]zapcore.Core, 0)

	for i := range cfg.Writers {
		var ws []zapcore.WriteSyncer
		if cfg.Writers[i].Stdout {
			ws = append(ws, zapcore.AddSync(os.Stdout))
		}
		if cfg.Writers[i].File {
			lw := &lumberjack.Logger{
				Filename:   cfg.Writers[i].FileConfig.Filename,
				MaxSize:    cfg.Writers[i].FileConfig.MaxSize,
				MaxBackups: cfg.Writers[i].FileConfig.MaxBackups,
				MaxAge:     cfg.Writers[i].FileConfig.MaxAge,
				Compress:   cfg.Writers[i].FileConfig.Compress,
			}
			ws = append(ws, zapcore.AddSync(lw))
		}

		multi := zapcore.NewMultiWriteSyncer(ws...)

		encCfg := zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "lvl",
			NameKey:        "logger",
			CallerKey:      "",
			MessageKey:     "msg",
			StacktraceKey:  "",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.EpochTimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
		}

		level, err := zapcore.ParseLevel(cfg.Level.String())
		if err != nil {
			panic(err)
		}

		encoder := zapcore.NewJSONEncoder(encCfg)
		core := zapcore.NewCore(encoder, multi, level)
		cores = append(cores, core)
	}

	finalCore := zapcore.NewTee(cores...)
	zl := &zapLogger{
		logger:    zap.New(finalCore),
		async:     cfg.Async,
		ch:        make(chan func(), cfg.QueueSize),
		done:      make(chan struct{}),
		batchSize: cfg.BatchSize,
		batchDur:  cfg.BatchDur,
	}

	if cfg.Async {
		go zl.worker()
	}

	return zl
}

func (z *zapLogger) Debug(ctx context.Context, msg string, fields ...any) {
	z.log(func() { z.logger.Sugar().Debugw(msg, fields...) })
}

func (z *zapLogger) Info(ctx context.Context, msg string, fields ...any) {
	z.log(func() { z.logger.Sugar().Infow(msg, fields...) })
}

func (z *zapLogger) Warn(ctx context.Context, msg string, fields ...any) {
	z.log(func() { z.logger.Sugar().Warnw(msg, fields...) })
}

func (z *zapLogger) Error(ctx context.Context, msg string, fields ...any) {
	z.log(func() { z.logger.Sugar().Errorw(msg, fields...) })
}

func (z *zapLogger) Sync() error {
	if z.async {
		close(z.ch)
		<-z.done
	}

	return z.logger.Sync()
}

func (z *zapLogger) log(fn func()) {
	if z.async {
		select {
		case z.ch <- fn:
		default:
			// Drop log if full
		}
	} else {
		fn()
	}
}

func (z *zapLogger) worker() {
	batch := make([]func(), 0, z.batchSize)
	timer := time.NewTimer(z.batchDur)
	defer timer.Stop()

	for {
		select {
		case fn, ok := <-z.ch:
			if !ok {
				// flush remaining
				for _, f := range batch {
					f()
				}
				close(z.done)

				return
			}
			batch = append(batch, fn)
			if len(batch) >= z.batchSize {
				for _, f := range batch {
					f()
				}
				batch = batch[:0]
				timer.Reset(z.batchDur)
			}
		case <-timer.C:
			if len(batch) > 0 {
				for _, f := range batch {
					f()
				}
				batch = batch[:0]
			}
			timer.Reset(z.batchDur)
		}
	}
}
