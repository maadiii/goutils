package async

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestSpawnAwait_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	f := Spawn(ctx, func(ctx context.Context) (int, error) {
		return 42, nil
	})

	v, err := f.Await()
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if v != 42 {
		t.Fatalf("expected value 42, got %v", v)
	}
}

func TestSpawn_PanicHandled(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	f := Spawn(ctx, func(ctx context.Context) (int, error) {
		panic("boom!")
	})

	_, err := f.Await()
	if err == nil {
		t.Fatalf("expected error from panic, got nil")
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected panic message to be propagated, got %v", err)
	}
}

func TestSpawn_ContextCanceled_PreventsSend(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Use a synchronization channel so we control when the spawned function returns.
	startReturn := make(chan struct{})

	f := Spawn(ctx, func(c context.Context) (int, error) {
		<-startReturn
		return 1, nil
	})

	// Cancel before letting the function return so the goroutine's select should choose
	// the ctx.Done() case and not send the result.
	cancel()
	close(startReturn)

	// give the spawned goroutine time to execute the select after the function returns
	time.Sleep(20 * time.Millisecond)

	done := make(chan struct{})
	go func() {
		f.Await()
		close(done)
	}()

	select {
	case <-done:
		t.Fatalf("expected Await to block when context cancelled before send")
	case <-time.After(100 * time.Millisecond):
		// expected path: still blocked
	}
}

func TestSpawn_PanicWithContextCanceled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	startPanic := make(chan struct{})

	f := Spawn(ctx, func(c context.Context) (int, error) {
		<-startPanic
		panic("crash!")
	})

	// Cancel context before allowing the panic
	cancel()
	close(startPanic)

	// Give time for panic to be handled with cancelled context
	time.Sleep(20 * time.Millisecond)

	// Await should still block because the panic's select chose ctx.Done() path
	done := make(chan struct{})
	go func() {
		f.Await()
		close(done)
	}()

	select {
	case <-done:
		t.Fatalf("expected Await to block when panic recovery skips send due to cancelled context")
	case <-time.After(100 * time.Millisecond):
		// expected: still blocked
	}
}

func TestSpawn_ErrorReturned(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	expectedErr := context.Canceled
	f := Spawn(ctx, func(ctx context.Context) (string, error) {
		return "", expectedErr
	})

	v, err := f.Await()
	if err != expectedErr {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
	if v != "" {
		t.Fatalf("expected empty string, got %v", v)
	}
}
