package omnitui

import "sync"

// Ref stores a synchronized value without invalidating the interface.
type Ref[T any] struct {
	mu    sync.RWMutex
	value T
}

// UseRef returns the ref associated with key for the current component instance.
func UseRef[T any](ctx Context, key string, initial T) *Ref[T] {
	instance := hookInstance(ctx, "UseRef")
	validateHookKey(instance, "UseRef", key)
	if instance.seenRefs == nil {
		instance.seenRefs = make(map[string]struct{})
	}
	if _, exists := instance.seenRefs[key]; exists {
		panic("omnitui: duplicate UseRef key " + key + " at " + instance.path)
	}
	instance.seenRefs[key] = struct{}{}
	if existing, ok := instance.refs[key]; ok {
		ref, ok := existing.(*Ref[T])
		if !ok {
			panic("omnitui: UseRef type changed for key " + key + " at " + instance.path)
		}
		return ref
	}
	if instance.refs == nil {
		instance.refs = make(map[string]any)
	}
	ref := &Ref[T]{value: initial}
	instance.refs[key] = ref
	return ref
}

// Get returns the current value.
func (r *Ref[T]) Get() T {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.value
}

// Set replaces the current value.
func (r *Ref[T]) Set(value T) {
	r.mu.Lock()
	r.value = value
	r.mu.Unlock()
}

// Swap replaces the current value and returns the previous value.
func (r *Ref[T]) Swap(value T) T {
	r.mu.Lock()
	previous := r.value
	r.value = value
	r.mu.Unlock()
	return previous
}

// Update replaces the value atomically and returns the result. The callback
// must not access the same ref because it runs while the ref is locked.
func (r *Ref[T]) Update(update func(current T) T) T {
	if update == nil {
		panic("omnitui: nil Ref update")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.value = update(r.value)
	return r.value
}
