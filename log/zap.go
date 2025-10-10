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
	ch        chan func()
	done      chan struct{}
	batchSize int
	batchDur  time.Duration
}

func Zap(cfg Config) Logger {
	var ws []zapcore.WriteSyncer
	if cfg.Writer.Stdout {
		ws = append(ws, zapcore.AddSync(os.Stdout))
	}
	if cfg.Writer.FileConfig != nil {
		lw := &lumberjack.Logger{
			Filename:   cfg.Writer.FileConfig.Filename,
			MaxSize:    cfg.Writer.FileConfig.MaxSize,
			MaxBackups: cfg.Writer.FileConfig.MaxBackups,
			MaxAge:     cfg.Writer.FileConfig.MaxAge,
			Compress:   cfg.Writer.FileConfig.Compress,
		}
		ws = append(ws, zapcore.AddSync(lw))
	}

	encCfg := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		MessageKey:     "message",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.EpochTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
	}

	level, err := zapcore.ParseLevel(cfg.Level.String())
	if err != nil {
		panic(err)
	}

	multi := zapcore.NewMultiWriteSyncer(ws...)
	encoder := zapcore.NewJSONEncoder(encCfg)
	core := zapcore.NewCore(encoder, multi, level)
	finalCore := zapcore.NewTee(core)

	zl := &zapLogger{
		logger:    zap.New(finalCore),
		ch:        make(chan func(), cfg.QueueSize),
		done:      make(chan struct{}),
		batchSize: cfg.BatchSize,
		batchDur:  cfg.BatchDur,
	}

	go zl.worker()

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
	close(z.ch)
	<-z.done

	return z.logger.Sync()
}

func (z *zapLogger) log(fn func()) {
	select {
	case z.ch <- fn:
	default:
		// Drop log if full
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
