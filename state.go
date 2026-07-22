package omnitui

import (
	"fmt"
	"reflect"
)

type stateUpdate struct {
	typ   reflect.Type
	apply func(any) any
}

func SetState[S any](ctx Context, next S) {
	typ := reflect.TypeOf((*S)(nil)).Elem()
	enqueueState(ctx, stateUpdate{typ: typ, apply: func(any) any { return next }})
}

func UpdateState[S any](ctx Context, update func(current S) S) {
	if update == nil {
		panic("omnitui: nil state update")
	}
	typ := reflect.TypeOf((*S)(nil)).Elem()
	enqueueState(ctx, stateUpdate{typ: typ, apply: func(current any) any {
		value, ok := current.(S)
		if !ok {
			panic(fmt.Sprintf("omnitui: state type mismatch: expected %v, got %T", typ, current))
		}
		return update(value)
	}})
}

func enqueueState(ctx Context, update stateUpdate) {
	if ctx.instance == nil || ctx.dispatcher == nil {
		panic("omnitui: state update outside component render")
	}
	if ctx.rendering != nil && ctx.rendering.Load() {
		panic(fmt.Sprintf("omnitui: state update during Render at %s", ctx.instance.path))
	}
	ctx.dispatcher.enqueue(func() {
		instance := ctx.instance
		if !instance.mounted {
			return
		}
		if !stateTypeMatches(instance.state, update.typ) {
			panic(fmt.Sprintf("omnitui: state type mismatch at %s: expected %v, got %T", instance.path, update.typ, instance.state))
		}
		instance.pending = append(instance.pending, update)
		instance.app.invalidated = true
	})
}
