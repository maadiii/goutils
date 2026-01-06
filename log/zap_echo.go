package log

import (
	"io"
	"os"
	"time"

	"github.com/labstack/gommon/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type zapEcho struct {
	logger    *zap.Logger
	output    io.Writer
	prefix    string
	level     log.Lvl
	ch        chan func()
	done      chan struct{}
	batchSize int
	batchDur  time.Duration
}

func ZapEcho(cfg Config) *zapEcho {
	var ws []zapcore.WriteSyncer
	if cfg.Writer.Stdout {
		ws = append(ws, zapcore.AddSync(os.Stdout))
	}
	if cfg.Writer.FileConfig != nil {
		ws = append(ws, zapcore.AddSync(
			&lumberjack.Logger{
				Filename:   cfg.Writer.FileConfig.Filename,
				MaxSize:    cfg.Writer.FileConfig.MaxSize,
				MaxBackups: cfg.Writer.FileConfig.MaxBackups,
				MaxAge:     cfg.Writer.FileConfig.MaxAge,
				Compress:   cfg.Writer.FileConfig.Compress,
			},
		))
	}

	encCfg := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		MessageKey:     "message",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.RFC3339TimeEncoder,
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

	ze := &zapEcho{
		logger:    zap.New(finalCore),
		output:    multi,
		prefix:    cfg.ServiceName,
		level:     toGommonLevel(cfg.Level),
		ch:        make(chan func(), cfg.QueueSize),
		done:      make(chan struct{}),
		batchSize: cfg.BatchSize,
		batchDur:  cfg.BatchDur,
	}

	go ze.worker()

	return ze
}

// Output returns the output destination for the logger.
func (z *zapEcho) Output() io.Writer {
	return z.output
}

// SetOutput sets the output destination for the logger.
func (z *zapEcho) SetOutput(w io.Writer) {
	z.output = w
}

// Prefix returns the logger prefix.
func (z *zapEcho) Prefix() string {
	return z.prefix
}

// SetPrefix sets the logger prefix.
func (z *zapEcho) SetPrefix(p string) {
	z.prefix = p
}

// Level returns the logger level.
func (z *zapEcho) Level() log.Lvl {
	return z.level
}

// SetLevel sets the logger level.
func (z *zapEcho) SetLevel(v log.Lvl) {
	z.level = v
}

// SetHeader sets the logger header (not used in zap implementation).
func (z *zapEcho) SetHeader(h string) {
	// No-op: zap doesn't use headers in the same way
}

// Print outputs message using Print method.
func (z *zapEcho) Print(i ...any) {
	if z.level > log.INFO {
		return
	}
	z.log(func() {
		z.logger.Sugar().Info(i...)
	})
}

// Printf outputs formatted message using Printf method.
func (z *zapEcho) Printf(format string, args ...any) {
	if z.level > log.INFO {
		return
	}
	z.log(func() {
		z.logger.Sugar().Infof(format, args...)
	})
}

// Printj outputs JSON message using Printj method.
func (z *zapEcho) Printj(j log.JSON) {
	if z.level > log.INFO {
		return
	}
	z.log(func() {
		message, keyValues := keyValues(j)
		z.logger.Sugar().Infow(message, keyValues...)
	})
}

// Debug outputs message at Debug level.
func (z *zapEcho) Debug(i ...any) {
	if z.level > log.DEBUG {
		return
	}
	z.log(func() {
		z.logger.Sugar().Debug(i...)
	})
}

// Debugf outputs formatted message at Debug level.
func (z *zapEcho) Debugf(format string, args ...any) {
	if z.level > log.DEBUG {
		return
	}
	z.log(func() {
		z.logger.Sugar().Debugf(format, args...)
	})
}

// Debugj outputs JSON message at Debug level.
func (z *zapEcho) Debugj(j log.JSON) {
	if z.level > log.DEBUG {
		return
	}
	z.log(func() {
		message, keyValues := keyValues(j)
		z.logger.Sugar().Debugw(message, keyValues...)
	})
}

// Info outputs message at Info level.
func (z *zapEcho) Info(i ...any) {
	if z.level > log.INFO {
		return
	}
	z.log(func() {
		z.logger.Sugar().Info(i...)
	})
}

// Infof outputs formatted message at Info level.
func (z *zapEcho) Infof(format string, args ...any) {
	if z.level > log.INFO {
		return
	}
	z.log(func() {
		z.logger.Sugar().Infof(format, args...)
	})
}

// Infoj outputs JSON message at Info level.
func (z *zapEcho) Infoj(j log.JSON) {
	if z.level > log.INFO {
		return
	}
	z.log(func() {
		message, keyValues := keyValues(j)
		z.logger.Sugar().Infow(message, keyValues...)
	})
}

// Warn outputs message at Warn level.
func (z *zapEcho) Warn(i ...any) {
	if z.level > log.WARN {
		return
	}
	z.log(func() {
		z.logger.Sugar().Warn(i...)
	})
}

// Warnf outputs formatted message at Warn level.
func (z *zapEcho) Warnf(format string, args ...any) {
	if z.level > log.WARN {
		return
	}
	z.log(func() {
		z.logger.Sugar().Warnf(format, args...)
	})
}

// Warnj outputs JSON message at Warn level.
func (z *zapEcho) Warnj(j log.JSON) {
	if z.level > log.WARN {
		return
	}
	z.log(func() {
		message, keyValues := keyValues(j)
		z.logger.Sugar().Warnw(message, keyValues...)
	})
}

// Error outputs message at Error level.
func (z *zapEcho) Error(i ...any) {
	if z.level > log.ERROR {
		return
	}
	z.log(func() {
		z.logger.Sugar().Error(i...)
	})
}

// Errorf outputs formatted message at Error level.
func (z *zapEcho) Errorf(format string, args ...any) {
	if z.level > log.ERROR {
		return
	}
	z.log(func() {
		z.logger.Sugar().Errorf(format, args...)
	})
}

// Errorj outputs JSON message at Error level.
func (z *zapEcho) Errorj(j log.JSON) {
	if z.level > log.ERROR {
		return
	}
	z.log(func() {
		message, keyValues := keyValues(j)
		z.logger.Sugar().Errorw(message, keyValues...)
	})
}

// Fatal outputs message at Fatal level and exits.
func (z *zapEcho) Fatal(i ...any) {
	z.log(func() {
		z.logger.Sugar().Fatal(i...)
	})
}

// Fatalf outputs formatted message at Fatal level and exits.
func (z *zapEcho) Fatalf(format string, args ...any) {
	z.log(func() {
		z.logger.Sugar().Fatalf(format, args...)
	})
}

// Fatalj outputs JSON message at Fatal level and exits.
func (z *zapEcho) Fatalj(j log.JSON) {
	z.log(func() {
		z.logger.Sugar().Fatalw("", j)
	})
}

// Panic outputs message at Panic level and panics.
func (z *zapEcho) Panic(i ...any) {
	z.log(func() {
		z.logger.Sugar().Panic(i...)
	})
}

// Panicf outputs formatted message at Panic level and panics.
func (z *zapEcho) Panicf(format string, args ...any) {
	z.log(func() {
		z.logger.Sugar().Panicf(format, args...)
	})
}

// Panicj outputs JSON message at Panic level and panics.
func (z *zapEcho) Panicj(j log.JSON) {
	z.log(func() {
		message, keyValues := keyValues(j)
		z.logger.Sugar().Panicw(message, keyValues...)
	})
}

// Sync flushes any buffered log entries and waits for completion.
func (z *zapEcho) Sync() error {
	close(z.ch)
	<-z.done

	return z.logger.Sync()
}

func (z *zapEcho) log(fn func()) {
	select {
	case z.ch <- fn:
	default:
		// Drop log if full
	}
}

func (z *zapEcho) worker() {
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

func toGommonLevel(l Level) log.Lvl {
	switch l {
	case DebugLevel:
		return log.DEBUG
	case InfoLevel:
		return log.INFO
	case WarnLevel:
		return log.WARN
	case ErrorLevel:
		return log.ERROR
	default:
		return log.ERROR
	}
}

func keyValues(j log.JSON) (string, []any) {
	keyValues := make([]any, 0)
	var message string
	for key, value := range j {
		if key == "message" {
			message = value.(string)

			continue
		}

		keyValues = append(keyValues, key, value)
	}

	return message, keyValues
}
