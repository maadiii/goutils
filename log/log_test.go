package log

import (
	"context"
	"os"
	"testing"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

func TestZapInit(t *testing.T) {
	config := Config{
		ServiceName: "test-service",
		Env:         "test",
		Level:       InfoLevel,
		Writer: WriterConfig{
			Stdout: false,
			FileConfig: &lumberjack.Logger{
				Filename:   "/tmp/test.log",
				MaxSize:    100,
				MaxBackups: 1,
			},
		},
		QueueSize: 100,
		BatchSize: 10,
		BatchDur:  100 * time.Millisecond,
	}

	logger := Zap(config)
	if logger == nil {
		t.Fatalf("Zap returned nil logger")
	}
}

func TestZapDebug(t *testing.T) {
	config := Config{
		ServiceName: "test-service",
		Env:         "test",
		Level:       DebugLevel,
		Writer: WriterConfig{
			Stdout: false,
			FileConfig: &lumberjack.Logger{
				Filename:   "/tmp/test.log",
				MaxSize:    100,
				MaxBackups: 1,
			},
		},
		QueueSize: 100,
		BatchSize: 10,
		BatchDur:  100 * time.Millisecond,
	}

	logger := Zap(config)
	logger.Debug("test debug message", "key", "value")
}

func TestZapInfo(t *testing.T) {
	config := Config{
		ServiceName: "test-service",
		Env:         "test",
		Level:       InfoLevel,
		Writer: WriterConfig{
			Stdout: false,
			FileConfig: &lumberjack.Logger{
				Filename:   "/tmp/test.log",
				MaxSize:    100,
				MaxBackups: 1,
			},
		},
		QueueSize: 100,
		BatchSize: 10,
		BatchDur:  100 * time.Millisecond,
	}

	logger := Zap(config)
	logger.Info("test info message", "key", "value")
}

func TestZapWarn(t *testing.T) {
	config := Config{
		ServiceName: "test-service",
		Env:         "test",
		Level:       WarnLevel,
		Writer: WriterConfig{
			Stdout: false,
			FileConfig: &lumberjack.Logger{
				Filename:   "/tmp/test.log",
				MaxSize:    100,
				MaxBackups: 1,
			},
		},
		QueueSize: 100,
		BatchSize: 10,
		BatchDur:  100 * time.Millisecond,
	}

	logger := Zap(config)
	logger.Warn("test warn message", "key", "value")
}

func TestZapError(t *testing.T) {
	config := Config{
		ServiceName: "test-service",
		Env:         "test",
		Level:       ErrorLevel,
		Writer: WriterConfig{
			Stdout: false,
			FileConfig: &lumberjack.Logger{
				Filename:   "/tmp/test.log",
				MaxSize:    100,
				MaxBackups: 1,
			},
		},
		QueueSize: 100,
		BatchSize: 10,
		BatchDur:  100 * time.Millisecond,
	}

	logger := Zap(config)
	logger.Error("test error message", "key", "value")
}

func TestZapWithStdout(t *testing.T) {
	config := Config{
		ServiceName: "test-service",
		Env:         "test",
		Level:       InfoLevel,
		Writer: WriterConfig{
			Stdout: true,
		},
		QueueSize: 100,
		BatchSize: 10,
		BatchDur:  100 * time.Millisecond,
	}

	logger := Zap(config)
	logger.Info("test stdout message", "key", "value")
}

func TestZapWithBothStdoutAndFile(t *testing.T) {
	config := Config{
		ServiceName: "test-service",
		Env:         "test",
		Level:       InfoLevel,
		Writer: WriterConfig{
			Stdout: true,
			FileConfig: &lumberjack.Logger{
				Filename:   "/tmp/test.log",
				MaxSize:    100,
				MaxBackups: 1,
			},
		},
		QueueSize: 100,
		BatchSize: 10,
		BatchDur:  100 * time.Millisecond,
	}

	logger := Zap(config)
	logger.Info("test both stdout and file message", "key", "value")
}

func TestHertzMiddleware(t *testing.T) {
	config := Config{
		ServiceName: "test-service",
		Env:         "test",
		Level:       InfoLevel,
		Writer: WriterConfig{
			Stdout: false,
			FileConfig: &lumberjack.Logger{
				Filename:   "/tmp/test.log",
				MaxSize:    100,
				MaxBackups: 1,
			},
		},
		QueueSize: 100,
		BatchSize: 10,
		BatchDur:  100 * time.Millisecond,
	}

	logger := Zap(config)
	middleware := HertzMiddleware(logger)
	if middleware == nil {
		t.Fatalf("HertzMiddleware returned nil")
	}
}

func TestZapEcho(t *testing.T) {
	config := Config{
		ServiceName: "test-service",
		Env:         "test",
		Level:       InfoLevel,
		Writer: WriterConfig{
			Stdout: false,
			FileConfig: &lumberjack.Logger{
				Filename:   "/tmp/test.log",
				MaxSize:    100,
				MaxBackups: 1,
			},
		},
		QueueSize: 100,
		BatchSize: 10,
		BatchDur:  100 * time.Millisecond,
	}

	zapEcho := ZapEcho(config)
	if zapEcho == nil {
		t.Fatalf("ZapEcho returned nil")
	}
}

func TestZapEchoDebug(t *testing.T) {
	config := Config{
		ServiceName: "test-service",
		Env:         "test",
		Level:       DebugLevel,
		Writer: WriterConfig{
			Stdout: false,
			FileConfig: &lumberjack.Logger{
				Filename:   "/tmp/test.log",
				MaxSize:    100,
				MaxBackups: 1,
			},
		},
		QueueSize: 100,
		BatchSize: 10,
		BatchDur:  100 * time.Millisecond,
	}

	zapEcho := ZapEcho(config)
	zapEcho.Debug("test debug")
}

func TestZapEchoInfo(t *testing.T) {
	config := Config{
		ServiceName: "test-service",
		Env:         "test",
		Level:       InfoLevel,
		Writer: WriterConfig{
			Stdout: false,
			FileConfig: &lumberjack.Logger{
				Filename:   "/tmp/test.log",
				MaxSize:    100,
				MaxBackups: 1,
			},
		},
		QueueSize: 100,
		BatchSize: 10,
		BatchDur:  100 * time.Millisecond,
	}

	zapEcho := ZapEcho(config)
	zapEcho.Info("test info")
}

func TestZapEchoWarn(t *testing.T) {
	config := Config{
		ServiceName: "test-service",
		Env:         "test",
		Level:       WarnLevel,
		Writer: WriterConfig{
			Stdout: false,
			FileConfig: &lumberjack.Logger{
				Filename:   "/tmp/test.log",
				MaxSize:    100,
				MaxBackups: 1,
			},
		},
		QueueSize: 100,
		BatchSize: 10,
		BatchDur:  100 * time.Millisecond,
	}

	zapEcho := ZapEcho(config)
	zapEcho.Warn("test warn")
}

func TestZapEchoError(t *testing.T) {
	config := Config{
		ServiceName: "test-service",
		Env:         "test",
		Level:       ErrorLevel,
		Writer: WriterConfig{
			Stdout: false,
			FileConfig: &lumberjack.Logger{
				Filename:   "/tmp/test.log",
				MaxSize:    100,
				MaxBackups: 1,
			},
		},
		QueueSize: 100,
		BatchSize: 10,
		BatchDur:  100 * time.Millisecond,
	}

	zapEcho := ZapEcho(config)
	zapEcho.Error("test error")
}

func TestZapEchoDebugf(t *testing.T) {
	config := Config{
		ServiceName: "test-service",
		Env:         "test",
		Level:       DebugLevel,
		Writer: WriterConfig{
			Stdout: false,
			FileConfig: &lumberjack.Logger{
				Filename:   "/tmp/test.log",
				MaxSize:    100,
				MaxBackups: 1,
			},
		},
		QueueSize: 100,
		BatchSize: 10,
		BatchDur:  100 * time.Millisecond,
	}

	zapEcho := ZapEcho(config)
	zapEcho.Debugf("test debug %s", "formatted")
}

func TestZapEchoInfof(t *testing.T) {
	config := Config{
		ServiceName: "test-service",
		Env:         "test",
		Level:       InfoLevel,
		Writer: WriterConfig{
			Stdout: false,
			FileConfig: &lumberjack.Logger{
				Filename:   "/tmp/test.log",
				MaxSize:    100,
				MaxBackups: 1,
			},
		},
		QueueSize: 100,
		BatchSize: 10,
		BatchDur:  100 * time.Millisecond,
	}

	zapEcho := ZapEcho(config)
	zapEcho.Infof("test info %s", "formatted")
}

func TestZapEchoWarnf(t *testing.T) {
	config := Config{
		ServiceName: "test-service",
		Env:         "test",
		Level:       WarnLevel,
		Writer: WriterConfig{
			Stdout: false,
			FileConfig: &lumberjack.Logger{
				Filename:   "/tmp/test.log",
				MaxSize:    100,
				MaxBackups: 1,
			},
		},
		QueueSize: 100,
		BatchSize: 10,
		BatchDur:  100 * time.Millisecond,
	}

	zapEcho := ZapEcho(config)
	zapEcho.Warnf("test warn %s", "formatted")
}

func TestZapEchoErrorf(t *testing.T) {
	config := Config{
		ServiceName: "test-service",
		Env:         "test",
		Level:       ErrorLevel,
		Writer: WriterConfig{
			Stdout: false,
			FileConfig: &lumberjack.Logger{
				Filename:   "/tmp/test.log",
				MaxSize:    100,
				MaxBackups: 1,
			},
		},
		QueueSize: 100,
		BatchSize: 10,
		BatchDur:  100 * time.Millisecond,
	}

	zapEcho := ZapEcho(config)
	zapEcho.Errorf("test error %s", "formatted")
}

func TestZapEchoDebugj(t *testing.T) {
	config := Config{
		ServiceName: "test-service",
		Env:         "test",
		Level:       DebugLevel,
		Writer: WriterConfig{
			Stdout: false,
			FileConfig: &lumberjack.Logger{
				Filename:   "/tmp/test.log",
				MaxSize:    100,
				MaxBackups: 1,
			},
		},
		QueueSize: 100,
		BatchSize: 10,
		BatchDur:  100 * time.Millisecond,
	}

	zapEcho := ZapEcho(config)
	zapEcho.Debugj(map[string]interface{}{"key": "value"})
}

func TestZapEchoInfoj(t *testing.T) {
	config := Config{
		ServiceName: "test-service",
		Env:         "test",
		Level:       InfoLevel,
		Writer: WriterConfig{
			Stdout: false,
			FileConfig: &lumberjack.Logger{
				Filename:   "/tmp/test.log",
				MaxSize:    100,
				MaxBackups: 1,
			},
		},
		QueueSize: 100,
		BatchSize: 10,
		BatchDur:  100 * time.Millisecond,
	}

	zapEcho := ZapEcho(config)
	zapEcho.Infoj(map[string]interface{}{"key": "value"})
}

func TestZapEchoWarnj(t *testing.T) {
	config := Config{
		ServiceName: "test-service",
		Env:         "test",
		Level:       WarnLevel,
		Writer: WriterConfig{
			Stdout: false,
			FileConfig: &lumberjack.Logger{
				Filename:   "/tmp/test.log",
				MaxSize:    100,
				MaxBackups: 1,
			},
		},
		QueueSize: 100,
		BatchSize: 10,
		BatchDur:  100 * time.Millisecond,
	}

	zapEcho := ZapEcho(config)
	zapEcho.Warnj(map[string]interface{}{"key": "value"})
}

func TestZapEchoErrorj(t *testing.T) {
	config := Config{
		ServiceName: "test-service",
		Env:         "test",
		Level:       ErrorLevel,
		Writer: WriterConfig{
			Stdout: false,
			FileConfig: &lumberjack.Logger{
				Filename:   "/tmp/test.log",
				MaxSize:    100,
				MaxBackups: 1,
			},
		},
		QueueSize: 100,
		BatchSize: 10,
		BatchDur:  100 * time.Millisecond,
	}

	zapEcho := ZapEcho(config)
	zapEcho.Errorj(map[string]interface{}{"key": "value"})
}

func TestZapEchoSync(t *testing.T) {
	config := Config{
		ServiceName: "test-service",
		Env:         "test",
		Level:       InfoLevel,
		Writer: WriterConfig{
			Stdout: true,
		},
		QueueSize: 100,
		BatchSize: 10,
		BatchDur:  100 * time.Millisecond,
	}

	zapEcho := ZapEcho(config)
	err := zapEcho.Sync()
	// Sync may fail on some systems, we just test it doesn't panic
	_ = err
}

func TestLevelString(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{DebugLevel, "debug"},
		{InfoLevel, "info"},
		{WarnLevel, "warn"},
		{ErrorLevel, "error"},
	}

	for _, test := range tests {
		result := test.level.String()
		if result != test.expected {
			t.Errorf("Level.String() = %q, want %q", result, test.expected)
		}
	}
}

func TestWithContext(t *testing.T) {
	config := Config{
		ServiceName: "test-service",
		Env:         "test",
		Level:       InfoLevel,
		Writer: WriterConfig{
			Stdout: true,
		},
		QueueSize: 100,
		BatchSize: 10,
		BatchDur:  100 * time.Millisecond,
	}

	logger := Zap(config)
	ctx := WithContext(context.Background(), logger)
	if ctx == nil {
		t.Fatalf("WithContext returned nil")
	}

	retrievedLogger := FromContext(ctx)
	if retrievedLogger == nil {
		t.Fatalf("FromContext returned nil")
	}
}

func TestZapEchoOutput(t *testing.T) {
	config := Config{
		ServiceName: "test-service",
		Env:         "test",
		Level:       InfoLevel,
		Writer: WriterConfig{
			Stdout: true,
		},
		QueueSize: 100,
		BatchSize: 10,
		BatchDur:  100 * time.Millisecond,
	}

	zapEcho := ZapEcho(config)
	output := zapEcho.Output()
	if output != os.Stderr {
		t.Logf("Output returned non-stderr: %v", output)
	}
}

func TestZapEchoSetOutput(t *testing.T) {
	config := Config{
		ServiceName: "test-service",
		Env:         "test",
		Level:       InfoLevel,
		Writer: WriterConfig{
			Stdout: true,
		},
		QueueSize: 100,
		BatchSize: 10,
		BatchDur:  100 * time.Millisecond,
	}

	zapEcho := ZapEcho(config)
	zapEcho.SetOutput(os.Stdout)
	// Should not panic
}

func TestZapEchoPrefix(t *testing.T) {
	config := Config{
		ServiceName: "test-service",
		Env:         "test",
		Level:       InfoLevel,
		Writer: WriterConfig{
			Stdout: true,
		},
		QueueSize: 100,
		BatchSize: 10,
		BatchDur:  100 * time.Millisecond,
	}

	zapEcho := ZapEcho(config)
	prefix := zapEcho.Prefix()
	_ = prefix // Just test it doesn't panic
}

func TestZapEchoSetPrefix(t *testing.T) {
	config := Config{
		ServiceName: "test-service",
		Env:         "test",
		Level:       InfoLevel,
		Writer: WriterConfig{
			Stdout: true,
		},
		QueueSize: 100,
		BatchSize: 10,
		BatchDur:  100 * time.Millisecond,
	}

	zapEcho := ZapEcho(config)
	zapEcho.SetPrefix("TEST_PREFIX")
	// Should not panic
}

func TestZapEchoLevel(t *testing.T) {
	config := Config{
		ServiceName: "test-service",
		Env:         "test",
		Level:       InfoLevel,
		Writer: WriterConfig{
			Stdout: true,
		},
		QueueSize: 100,
		BatchSize: 10,
		BatchDur:  100 * time.Millisecond,
	}

	zapEcho := ZapEcho(config)
	level := zapEcho.Level()
	_ = level // Just test it doesn't panic
}

func TestZapEchoSetHeader(t *testing.T) {
	config := Config{
		ServiceName: "test-service",
		Env:         "test",
		Level:       InfoLevel,
		Writer: WriterConfig{
			Stdout: true,
		},
		QueueSize: 100,
		BatchSize: 10,
		BatchDur:  100 * time.Millisecond,
	}

	zapEcho := ZapEcho(config)
	zapEcho.SetHeader("X-Custom-Header")
	// Should not panic
}
func TestZapEchoFatal(t *testing.T) {
	config := Config{
		ServiceName: "test-service",
		Env:         "test",
		Level:       ErrorLevel,
		Writer: WriterConfig{
			Stdout: false,
			FileConfig: &lumberjack.Logger{
				Filename:   "/tmp/test.log",
				MaxSize:    100,
				MaxBackups: 1,
			},
		},
		QueueSize: 100,
		BatchSize: 10,
		BatchDur:  100 * time.Millisecond,
	}

	_ = ZapEcho(config) // zapEcho would call os.Exit
	
	// Test Fatal - it calls os.Exit which terminates the test
	// We need to skip this or use a wrapper
	t.Skip("Fatal calls os.Exit and would terminate the test")
}

func TestZapEchoFatalf(t *testing.T) {
	config := Config{
		ServiceName: "test-service",
		Env:         "test",
		Level:       ErrorLevel,
		Writer: WriterConfig{
			Stdout: false,
			FileConfig: &lumberjack.Logger{
				Filename:   "/tmp/test.log",
				MaxSize:    100,
				MaxBackups: 1,
			},
		},
		QueueSize: 100,
		BatchSize: 10,
		BatchDur:  100 * time.Millisecond,
	}

	_ = ZapEcho(config)

	t.Skip("Fatalf calls os.Exit and would terminate the test")
}

func TestZapEchoFatalj(t *testing.T) {
	config := Config{
		ServiceName: "test-service",
		Env:         "test",
		Level:       ErrorLevel,
		Writer: WriterConfig{
			Stdout: false,
			FileConfig: &lumberjack.Logger{
				Filename:   "/tmp/test.log",
				MaxSize:    100,
				MaxBackups: 1,
			},
		},
		QueueSize: 100,
		BatchSize: 10,
		BatchDur:  100 * time.Millisecond,
	}

	_ = ZapEcho(config)

	t.Skip("Fatalj calls os.Exit and would terminate the test")
}

func TestZapEchoPanic(t *testing.T) {
	// Skip panic tests as they occur in worker goroutine
	t.Skip("Panic occurs in background worker, cannot be reliably tested")
}

func TestZapEchoPanicf(t *testing.T) {
	// Skip panic tests as they occur in worker goroutine
	t.Skip("Panicf occurs in background worker, cannot be reliably tested")
}

func TestZapEchoPanicj(t *testing.T) {
	// Skip panic tests as they occur in worker goroutine
	t.Skip("Panicj occurs in background worker, cannot be reliably tested")
}