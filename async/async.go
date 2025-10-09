package async

import (
	"context"

	"utils/errors"
)

type Future[E any] interface {
	Await() (E, error)
}

type FutureResult[E any] struct {
	Value E
	Err   error
}

func Spawn[E any](ctx context.Context, f func(context.Context) (E, error)) Future[E] {
	a := &async[E]{
		result: make(chan FutureResult[E]),
	}

	go func() {
		defer handlePanic(ctx, a)

		value, err := f(ctx)
		result := FutureResult[E]{
			Value: value,
			Err:   err,
		}

		select {
		case a.result <- result:
		case <-ctx.Done():
		}
	}()

	return a
}

type async[E any] struct {
	result chan FutureResult[E]
}

func (a *async[E]) Await() (E, error) {
	result := <-a.result

	return result.Value, result.Err
}

func handlePanic[E any](ctx context.Context, a *async[E]) {
	if r := recover(); r != nil {
		result := FutureResult[E]{
			Err: errors.New("%v", r),
		}

		select {
		case a.result <- result:
		case <-ctx.Done():
		}
	}
}
