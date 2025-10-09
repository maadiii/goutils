package service

import (
	"context"
	"reflect"
)

type Service[E any] interface {
	Invoke(ctx context.Context) error
}

func Register[E any](domainModel E, f factory) {
	typ := reflect.TypeOf(domainModel)

	registry[typ] = f
}

func Factory[E any](domainModel E) Service[E] {
	typ := reflect.TypeOf(domainModel)

	return registry[typ](domainModel)
}

type factory func(domainModel any) Service[any]

var registry = make(map[reflect.Type]factory, 0)
