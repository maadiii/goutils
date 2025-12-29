package log

import (
	"bufio"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/labstack/gommon/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

// helper to build config writing to a temp file
func tempCfg(t *testing.T, level Level, queue, batch int, batchDur time.Duration) (Config, string) {
	t.Helper()
	file := filepath.Join(t.TempDir(), "zap-echo.log")
	return Config{
		ServiceName: "svc-echo",
		Env:         "test",
		Level:       level,
		Writer: WriterConfig{
			Stdout: false,
			FileConfig: &lumberjack.Logger{
				Filename:   file,
				MaxSize:    10,
				MaxBackups: 1,
				MaxAge:     1,
				Compress:   false,
			},
		},
		QueueSize: queue,
		BatchSize: batch,
		BatchDur:  batchDur,
	}, file
}

func countLines(t *testing.T, path string) int {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		return 0
	}

	s := bufio.NewScanner(f)
	cnt := 0
	for s.Scan() {
		cnt++
	}

	if cerr := f.Close(); cerr != nil {
		t.Fatalf("close: %v", cerr)
	}
	return cnt
}

func TestZapEcho_LevelFiltering(t *testing.T) {
	cfg, file := tempCfg(t, InfoLevel, 10, 2, 20*time.Millisecond)
	l := ZapEcho(cfg)

	// debug should be dropped at info level
	l.Debug("debug-msg")
	// info should pass
	l.Info("info-msg")
	time.Sleep(50 * time.Millisecond)

	if err := l.Sync(); err != nil {
		t.Fatalf("sync error: %v", err)
	}

	lines := countLines(t, file)
	if lines != 1 {
		t.Fatalf("expected 1 line (info only), got %d", lines)
	}
}

func TestZapEcho_SyncFlushesPending(t *testing.T) {
	cfg, file := tempCfg(t, DebugLevel, 10, 5, 500*time.Millisecond)
	l := ZapEcho(cfg)

	// enqueue fewer than batch size, then sync immediately
	l.Info("a")
	l.Info("b")
	if err := l.Sync(); err != nil {
		t.Fatalf("sync error: %v", err)
	}

	lines := countLines(t, file)
	if lines != 2 {
		t.Fatalf("expected 2 flushed lines, got %d", lines)
	}
}

func TestZapEcho_TimerFlushesBatch(t *testing.T) {
	cfg, file := tempCfg(t, DebugLevel, 10, 10, 30*time.Millisecond)
	l := ZapEcho(cfg)

	l.Info("one")
	l.Info("two")
	// wait past batchDur to trigger timer flush (batchSize not reached)
	time.Sleep(80 * time.Millisecond)
	if err := l.Sync(); err != nil {
		t.Fatalf("sync error: %v", err)
	}

	lines := countLines(t, file)
	if lines != 2 {
		t.Fatalf("expected 2 lines after timer flush, got %d", lines)
	}
}

func TestZapEcho_DropsWhenQueueFull(t *testing.T) {
	cfg, file := tempCfg(t, DebugLevel, 1, 100, 200*time.Millisecond)
	l := ZapEcho(cfg)

	for i := 0; i < 10; i++ {
		l.Info("msg", "i", i)
	}

	time.Sleep(50 * time.Millisecond)
	if err := l.Sync(); err != nil {
		t.Fatalf("sync error: %v", err)
	}

	lines := countLines(t, file)
	if lines == 0 {
		t.Fatalf("expected at least one log despite drops, got %d", lines)
	}
	if lines > 2 {
		t.Fatalf("expected drops with tiny queue, got %d lines", lines)
	}
}

func TestZapEcho_BatchFlushOnSize(t *testing.T) {
	cfg, file := tempCfg(t, DebugLevel, 10, 2, 500*time.Millisecond)
	l := ZapEcho(cfg)

	// reach batch size to trigger immediate flush
	l.Info("one")
	l.Info("two")
	time.Sleep(50 * time.Millisecond)

	if err := l.Sync(); err != nil {
		t.Fatalf("sync error: %v", err)
	}

	if got := countLines(t, file); got != 2 {
		t.Fatalf("expected 2 lines after batch flush, got %d", got)
	}
}

func TestZapEcho_SetLevelRaisesThreshold(t *testing.T) {
	cfg, file := tempCfg(t, DebugLevel, 10, 4, 100*time.Millisecond)
	l := ZapEcho(cfg)

	// at debug level, this should log
	l.Debug("d1")

	// raise threshold to warn; subsequent debug should be skipped by pre-filter
	l.SetLevel(log.WARN)
	l.Debug("d2")

	if err := l.Sync(); err != nil {
		t.Fatalf("sync error: %v", err)
	}

	if got := countLines(t, file); got != 1 {
		t.Fatalf("expected only first debug to log after raising level, got %d", got)
	}
}

func TestZapEcho_PrintVariants(t *testing.T) {
	cfg, file := tempCfg(t, InfoLevel, 10, 5, 200*time.Millisecond)
	l := ZapEcho(cfg)

	l.Print("plain")
	l.Printf("formatted %s", "msg")
	l.Printj(log.JSON{"k": "v"})

	if err := l.Sync(); err != nil {
		t.Fatalf("sync error: %v", err)
	}

	if got := countLines(t, file); got < 3 || got > 4 {
		t.Fatalf("expected 3-4 lines for print variants, got %d", got)
	}
}
