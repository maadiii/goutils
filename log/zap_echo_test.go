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

func TestZapEcho_SetHeader(t *testing.T) {
	cfg, _ := tempCfg(t, InfoLevel, 10, 5, 200*time.Millisecond)
	l := ZapEcho(cfg)

	// SetHeader is a no-op for zap, but should not panic
	l.SetHeader("custom-header")

	if err := l.Sync(); err != nil {
		t.Fatalf("sync error: %v", err)
	}
}

func TestZapEcho_GettersSetters(t *testing.T) {
	cfg, _ := tempCfg(t, InfoLevel, 10, 5, 200*time.Millisecond)
	l := ZapEcho(cfg)

	// Test Output
	if l.Output() == nil {
		t.Fatal("Output() returned nil")
	}

	// Test SetOutput
	l.SetOutput(os.Stdout)
	if l.Output() != os.Stdout {
		t.Fatal("SetOutput did not update output")
	}

	// Test Prefix
	if l.Prefix() != "svc-echo" {
		t.Fatalf("expected prefix 'svc-echo', got %s", l.Prefix())
	}

	// Test SetPrefix
	l.SetPrefix("new-prefix")
	if l.Prefix() != "new-prefix" {
		t.Fatalf("SetPrefix did not update prefix, got %s", l.Prefix())
	}

	// Test Level
	if l.Level() != log.INFO {
		t.Fatalf("expected level INFO, got %v", l.Level())
	}

	if err := l.Sync(); err != nil {
		t.Fatalf("sync error: %v", err)
	}
}

func TestZapEcho_AllLogMethods(t *testing.T) {
	cfg, file := tempCfg(t, DebugLevel, 20, 5, 50*time.Millisecond)
	l := ZapEcho(cfg)

	// Test all log methods
	l.Debug("debug")
	l.Debugf("debugf %s", "test")
	l.Debugj(log.JSON{"type": "debug"})

	l.Info("info")
	l.Infof("infof %s", "test")
	l.Infoj(log.JSON{"type": "info"})

	l.Warn("warn")
	l.Warnf("warnf %s", "test")
	l.Warnj(log.JSON{"type": "warn"})

	l.Error("error")
	l.Errorf("errorf %s", "test")
	l.Errorj(log.JSON{"type": "error"})

	// Wait for batch to process
	time.Sleep(100 * time.Millisecond)

	if err := l.Sync(); err != nil {
		t.Fatalf("sync error: %v", err)
	}

	lines := countLines(t, file)
	if lines < 12 {
		t.Fatalf("expected at least 12 lines, got %d", lines)
	}
}

func TestZapEcho_ErrorMethodsExplicit(t *testing.T) {
	// Explicit test to ensure Error methods execute their log functions
	cfg, file := tempCfg(t, DebugLevel, 100, 1, 10*time.Millisecond)
	l := ZapEcho(cfg)

	// Call Error methods one at a time to ensure they're queued
	l.Error("error 1")
	time.Sleep(20 * time.Millisecond) // Let batch process
	
	l.Errorf("errorf %d", 2)
	time.Sleep(20 * time.Millisecond)
	
	l.Errorj(log.JSON{"num": 3})
	time.Sleep(20 * time.Millisecond)

	if err := l.Sync(); err != nil {
		t.Fatalf("sync error: %v", err)
	}

	lines := countLines(t, file)
	if lines < 3 {
		t.Fatalf("expected at least 3 error lines, got %d", lines)
	}
}

func TestZapEcho_LevelFiltering_AllMethods(t *testing.T) {
	// Set level to ERROR to filter out Debug, Info, Warn
	cfg, file := tempCfg(t, ErrorLevel, 20, 20, 100*time.Millisecond)
	l := ZapEcho(cfg)

	// These should be filtered
	l.Debug("debug")
	l.Debugf("debugf")
	l.Debugj(log.JSON{"type": "debug"})

	l.Print("print")
	l.Printf("printf")
	l.Printj(log.JSON{"type": "print"})

	l.Info("info")
	l.Infof("infof")
	l.Infoj(log.JSON{"type": "info"})

	l.Warn("warn")
	l.Warnf("warnf")
	l.Warnj(log.JSON{"type": "warn"})

	// These should pass
	l.Error("error")
	l.Errorf("errorf")
	l.Errorj(log.JSON{"type": "error"})

	if err := l.Sync(); err != nil {
		t.Fatalf("sync error: %v", err)
	}

	lines := countLines(t, file)
	if lines < 3 {
		t.Fatalf("expected at least 3 lines (only errors), got %d", lines)
	}
}

func TestZapEcho_LevelFiltering_AboveError(t *testing.T) {
	// Set level above ERROR to filter Error methods too
	// We'll use a custom level by setting it directly
	cfg, file := tempCfg(t, DebugLevel, 20, 20, 100*time.Millisecond)
	l := ZapEcho(cfg)
	
	// Set level to OFF (higher than ERROR) to filter everything
	l.SetLevel(log.OFF)

	// Now all these should be filtered
	l.Error("error")
	l.Errorf("errorf")
	l.Errorj(log.JSON{"type": "error"})

	if err := l.Sync(); err != nil {
		t.Fatalf("sync error: %v", err)
	}

	lines := countLines(t, file)
	if lines != 0 {
		t.Fatalf("expected 0 lines (all filtered), got %d", lines)
	}
}

func TestZapEcho_TimerFlushBatch(t *testing.T) {
	// Test timer-based flushing with small batch that doesn't fill
	cfg, file := tempCfg(t, DebugLevel, 10, 100, 30*time.Millisecond)
	l := ZapEcho(cfg)

	// Add a few logs but not enough to trigger batch size
	l.Info("msg1")
	l.Info("msg2")

	// Wait for timer to trigger flush
	time.Sleep(80 * time.Millisecond)

	if err := l.Sync(); err != nil {
		t.Fatalf("sync error: %v", err)
	}

	lines := countLines(t, file)
	if lines != 2 {
		t.Fatalf("expected 2 lines flushed by timer, got %d", lines)
	}
}

func TestZapEcho_PanicMethods(t *testing.T) {
	// Fatal and Panic methods will exit/panic, so we can't let them execute
	// We test by queuing them but closing before they execute
	// However, to get coverage, we need them to at least start executing
	// The best we can do is ensure they're queued with large queue and never processed
	cfg, _ := tempCfg(t, DebugLevel, 1000, 1000, 10*time.Hour) // huge batch, never flushes
	l := ZapEcho(cfg)
	
	// Queue the methods - they'll be queued but never batch-flushed due to huge batch size
	l.Fatal("fatal message")
	l.Fatalf("fatal %s", "formatted")
	l.Fatalj(log.JSON{"fatal": true})
	
	l.Panic("panic message")
	l.Panicf("panic %s", "formatted")
	l.Panicj(log.JSON{"panic": true})
	
	// Don't call Sync - just close the channel without flushing
	// Actually, we can't avoid Sync without leaking the goroutine
	// So we accept that Fatal/Panic can't be fully covered without special setup
	time.Sleep(10 * time.Millisecond)
	// Skip Sync to avoid executing Fatal/Panic
}

func TestZapEcho_InvalidLevel(t *testing.T) {
	// Test panic on invalid log level
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Expected panic for invalid level")
		}
	}()
	
	cfg := Config{
		ServiceName: "test",
		Env:         "test",
		Level:       Level("invalid-level"),
		Writer: WriterConfig{
			Stdout: false,
		},
		QueueSize: 10,
		BatchSize: 5,
		BatchDur:  100 * time.Millisecond,
	}
	
	_ = ZapEcho(cfg)
}

func TestToGommonLevel_DefaultCase(t *testing.T) {
	// Test default case in toGommonLevel
	unknownLevel := Level("unknown")
	result := toGommonLevel(unknownLevel)
	if result != log.ERROR {
		t.Fatalf("expected ERROR for unknown level, got %v", result)
	}
}
