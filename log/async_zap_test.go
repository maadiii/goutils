package log

import (
	"io"
	"sync"
	"testing"

	"go.uber.org/zap"
)

func BenchmarkAsyncLoggerHighRPS(b *testing.B) {
	logger, asyncCore, err := NewLoggerWithConfig(LoggerConfig{
		Writer:    io.Discard,
		Level:     "info",
		QueueSize: 500_000,
		BatchSize: 256,
		WorkerNum: 4,
	})
	if err != nil {
		b.Fatal(err)
	}
	defer asyncCore.Sync()

	b.ResetTimer()

	numGoroutines := 20
	logsPerGoroutine := b.N / numGoroutines
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for g := 0; g < numGoroutines; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < logsPerGoroutine; i++ {
				fieldsPtr := fieldPool.Get().(*[]zap.Field)
				fields := (*fieldsPtr)[:0]

				fields = append(fields,
					zap.String("name", "http"),
					zap.Int("status", 200),
					zap.String("method", "GET"),
					zap.String("route", "/api/test"),
					zap.String("path", "/api/test"),
					zap.String("query", "foo=bar"),
					zap.String("agent", "bench-agent"),
					zap.String("ip", "127.0.0.1"),
				)

				logger.Info("benchmark log", fields...)
			}
		}()
	}

	wg.Wait()

	b.Logf("Dropped logs: %d", asyncCore.Dropped())
}
