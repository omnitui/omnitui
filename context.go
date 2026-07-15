package omnitui

import (
	"sync/atomic"

	"github.com/viniciusfonseca/omnitui/internal/core"
)

type Context struct {
	instance   *instance
	dispatcher *dispatcher
	values     map[uint64]any
}

type ContextKey[T any] struct {
	id           uint64
	defaultValue T
}

var nextContextID atomic.Uint64

func NewContext[T any](defaultValue T) ContextKey[T] {
	return ContextKey[T]{id: nextContextID.Add(1), defaultValue: defaultValue}
}

func UseContext[T any](ctx Context, key ContextKey[T]) T {
	if ctx.values != nil {
		if value, ok := ctx.values[key.id]; ok {
			return value.(T)
		}
	}
	return key.defaultValue
}

func Provide[T any](key ContextKey[T], value T, child Element) Element {
	return core.ProviderElement(key.id, value, child)
}

func cloneContextValues(values map[uint64]any) map[uint64]any {
	if len(values) == 0 {
		return nil
	}
	clone := make(map[uint64]any, len(values))
	for key, value := range values {
		clone[key] = value
	}
	return clone
}
