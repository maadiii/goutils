package util

type Pointer[T any] interface {
	*T
}

func GetPtrValue[T any, P Pointer[T]](p P) T {
	if p == nil {
		var zero T

		return zero
	}

	return *p
}
