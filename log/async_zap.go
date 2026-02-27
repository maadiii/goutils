package log

import (
	"io"
	"sync"
	"sync/atomic"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LoggerConfig struct {
	Writer       io.Writer
	Level        string
	QueueSize    int
	BatchSize    int
	WorkerNum    int
	IncludeStack bool
}

var fieldPool = sync.Pool{
	New: func() any {
		s := make([]zap.Field, 0, 32)
		return &s
	},
}

type logItem struct {
	entry  zapcore.Entry
	fields []zapcore.Field
}

type AsyncCore struct {
	core      zapcore.Core
	queue     chan logItem
	wg        sync.WaitGroup
	batchSize int
	dropped   atomic.Uint64
}

func NewAsyncCore(core zapcore.Core, workerNum, buffer, batchSize int) *AsyncCore {
	ac := &AsyncCore{
		core:      core,
		queue:     make(chan logItem, buffer),
		batchSize: batchSize,
	}
	ac.wg.Add(workerNum)
	for range workerNum {
		go ac.worker()
	}
	return ac
}

func (a *AsyncCore) worker() {
	defer a.wg.Done()
	batch := make([]logItem, 0, a.batchSize)

	for {
		item, ok := <-a.queue
		if !ok {
			a.flushBatch(batch)
			return
		}
		batch = append(batch, item)
		if len(batch) >= a.batchSize {
			a.flushBatch(batch)
			batch = batch[:0]
		}
	}
}

func (a *AsyncCore) flushBatch(batch []logItem) {
	for _, item := range batch {
		_ = a.core.Write(item.entry, item.fields)
		for i := range item.fields {
			item.fields[i] = zap.Field{}
		}
		fieldPool.Put(&item.fields)
	}
}

func (a *AsyncCore) Enabled(l zapcore.Level) bool {
	return a.core.Enabled(l)
}

func (a *AsyncCore) With(fields []zapcore.Field) zapcore.Core {
	return &AsyncCore{
		core:      a.core.With(fields),
		queue:     a.queue,
		batchSize: a.batchSize,
	}
}

func (a *AsyncCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if a.Enabled(ent.Level) {
		return ce.AddCore(ent, a)
	}
	return ce
}

func (a *AsyncCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	item := logItem{entry: ent, fields: fields}

	if ent.Level >= zapcore.ErrorLevel {
		a.queue <- item // blocking
		return nil
	}

	select {
	case a.queue <- item:
		return nil
	default:
		a.dropped.Add(1)
		return nil
	}
}

func (a *AsyncCore) Sync() error {
	close(a.queue)
	a.wg.Wait()
	return a.core.Sync()
}

func (a *AsyncCore) Dropped() uint64 {
	return a.dropped.Load()
}

func NewLoggerWithConfig(cfg LoggerConfig) (*zap.Logger, *AsyncCore, error) {
	encoderCfg := zapcore.EncoderConfig{
		TimeKey:       "ts",
		LevelKey:      "level",
		CallerKey:     "caller",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		EncodeLevel:   zapcore.LowercaseLevelEncoder,
		EncodeTime:    zapcore.ISO8601TimeEncoder,
		EncodeCaller:  zapcore.ShortCallerEncoder,
	}

	encoder := zapcore.NewJSONEncoder(encoderCfg)
	ws := zapcore.AddSync(cfg.Writer)

	level := zapcore.InfoLevel
	switch cfg.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	}

	core := zapcore.NewCore(encoder, ws, level)

	queueSize := cfg.QueueSize
	if queueSize <= 0 {
		queueSize = 100_000
	}

	batchSize := cfg.BatchSize
	if batchSize <= 0 {
		batchSize = 32
	}

	asyncCore := NewAsyncCore(core, cfg.WorkerNum, queueSize, batchSize)

	opts := []zap.Option{zap.AddCaller()}
	if cfg.IncludeStack {
		opts = append(opts, zap.AddStacktrace(zap.ErrorLevel))
	}

	logger := zap.New(asyncCore, opts...)

	return logger, asyncCore, nil
}
