package log

import (
	"context"
	"testing"
	"time"
)

type spyLogger struct {
	calls []string
}

func (s *spyLogger) Debug(msg string, fields ...any) {
	s.calls = append(s.calls, "debug:"+msg)
}

func (s *spyLogger) Info(msg string, fields ...any) {
	s.calls = append(s.calls, "info:"+msg)
}

func (s *spyLogger) Warn(msg string, fields ...any) {
	s.calls = append(s.calls, "warn:"+msg)
}

func (s *spyLogger) Error(msg string, fields ...any) {
	s.calls = append(s.calls, "error:"+msg)
}
func (s *spyLogger) Sync() error { return nil }

func TestLevel_String(t *testing.T) {
	if DebugLevel.String() != "debug" {
		t.Fatalf("expected debug level string")
	}
	if InfoLevel.String() != "info" {
		t.Fatalf("expected info level string")
	}
}

func TestWithContext_FromContext(t *testing.T) {
	s := &spyLogger{}
	ctx := WithContext(context.Background(), s)

	got := FromContext(ctx)
	if got == nil {
		t.Fatalf("FromContext returned nil")
	}

	// verify it implements goutils.Logger by calling a method
	got.Info("test-msg", "k", "v")

	// type assert to our spy to inspect calls
	if sp, ok := got.(*spyLogger); ok {
		if len(sp.calls) != 1 || sp.calls[0] != "info:test-msg" {
			t.Fatalf("unexpected spy calls: %#v", sp.calls)
		}
	}
}

func TestZapLogger_BasicLoggingAndSync(t *testing.T) {
	cfg := Config{
		ServiceName: "svc",
		Env:         "test",
		Level:       DebugLevel,
		Writer: WriterConfig{
			Stdout: false,
		},
		QueueSize: 10,
		BatchSize: 2,
		BatchDur:  50 * time.Millisecond,
	}

	l := Zap(cfg)

	// basic calls shouldn't panic
	l.Debug("dmsg", "k", "v")
	l.Info("imsg", "k", "v")
	l.Warn("wmsg", "k", "v")
	l.Error("emsg", "k", "v")

	// allow worker to process
	time.Sleep(100 * time.Millisecond)

	if err := l.Sync(); err != nil {
		t.Fatalf("Sync returned error: %v", err)
	}
}

func TestZapLogger_DropWhenQueueFull(t *testing.T) {
	cfg := Config{
		ServiceName: "svc",
		Env:         "test",
		Level:       DebugLevel,
		Writer: WriterConfig{
			Stdout: false,
		},
		QueueSize: 1,
		BatchSize: 1000,
		BatchDur:  200 * time.Millisecond,
	}

	l := Zap(cfg)

	// fill queue quickly with more messages than capacity
	for i := range 10 {
		l.Info("msg", "i", i)
	}

	// give some time and then sync
	time.Sleep(50 * time.Millisecond)
	if err := l.Sync(); err != nil {
		t.Fatalf("Sync returned error after overflow: %v", err)
	}
}

func TestZapLogger_TimerFlushBatch(t *testing.T) {
	// Test that timer flushes batches even when batch size not reached
	cfg := Config{
		ServiceName: "timer-test",
		Env:         "test",
		Level:       DebugLevel,
		Writer: WriterConfig{
			Stdout: false,
		},
		QueueSize: 10,
		BatchSize: 100, // large batch size
		BatchDur:  50 * time.Millisecond,
	}

	l := Zap(cfg)

	// send fewer messages than batch size
	l.Info("msg1")
	l.Info("msg2")

	// wait for timer to trigger flush
	time.Sleep(100 * time.Millisecond)

	if err := l.Sync(); err != nil {
		t.Fatalf("Sync returned error: %v", err)
	}
}

func TestZap_InvalidLevel(t *testing.T) {
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

	_ = Zap(cfg)
}
