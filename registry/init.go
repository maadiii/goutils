package registry

import (
	"context"
	"reflect"
)

type Handler[E any] interface {
	Invoke(ctx context.Context)
}

func Register[E any](domainModel E, f registryFn) {
	typ := reflect.TypeOf(domainModel)

	registry[typ] = f
}

func Service[E any](domainModel E) Handler[E] {
	typ := reflect.TypeOf(domainModel)

	return registry[typ](domainModel)
}

type registryFn func(domainModel any) Handler[any]

var registry = make(map[reflect.Type]registryFn, 0)
