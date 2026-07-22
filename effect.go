package omnitui

import "context"

// Cleanup releases resources owned by an effect.
type Cleanup func()

type pendingEffect struct {
	dependencies any
	equal        func(any) bool
	setup        func(context.Context) Cleanup
}

type effectSlot struct {
	dependencies any
	cancel       context.CancelFunc
	cleanup      Cleanup
}

// UseEffect registers post-commit work for a keyed component instance. A
// changed dependency cancels and cleans up the previous effect before setup.
func UseEffect[D comparable](ctx Context, key string, dependencies D, setup func(context.Context) Cleanup) {
	instance := hookInstance(ctx, "UseEffect")
	validateHookKey(instance, "UseEffect", key)
	if setup == nil {
		panic("omnitui: UseEffect setup cannot be nil at " + instance.path)
	}
	if instance.pendingEffects == nil {
		instance.pendingEffects = make(map[string]pendingEffect)
	}
	if _, exists := instance.pendingEffects[key]; exists {
		panic("omnitui: duplicate UseEffect key " + key + " at " + instance.path)
	}
	instance.pendingEffects[key] = pendingEffect{
		dependencies: dependencies,
		equal: func(previous any) bool {
			value, ok := previous.(D)
			return ok && value == dependencies
		},
		setup: setup,
	}
	instance.pendingEffectOrder = append(instance.pendingEffectOrder, key)
}

func (app *App) commitEffects() {
	for _, slot := range app.pendingEffectCleanups {
		runEffectCleanup(slot)
	}
	app.pendingEffectCleanups = nil
	for _, instance := range app.effectInstances {
		if !instance.mounted {
			continue
		}
		instance.commitEffects(app.effectContext)
	}
	app.effectInstances = nil
}

func (i *instance) commitEffects(parent context.Context) {
	if parent == nil {
		parent = context.Background()
	}
	for _, key := range i.effectOrder {
		slot := i.effects[key]
		pending, present := i.pendingEffects[key]
		if !present || !pending.equal(slot.dependencies) {
			runEffectCleanup(slot)
			delete(i.effects, key)
		}
	}
	if i.effects == nil && len(i.pendingEffects) > 0 {
		i.effects = make(map[string]*effectSlot)
	}
	i.effectOrder = append(i.effectOrder[:0], i.pendingEffectOrder...)
	for _, key := range i.pendingEffectOrder {
		pending := i.pendingEffects[key]
		if slot, exists := i.effects[key]; exists && pending.equal(slot.dependencies) {
			continue
		}
		effectContext, cancel := context.WithCancel(parent)
		slot := &effectSlot{dependencies: pending.dependencies, cancel: cancel}
		i.effects[key] = slot
		slot.cleanup = pending.setup(effectContext)
	}
	i.pendingEffects = nil
	i.pendingEffectOrder = nil
}

func runEffectCleanup(slot *effectSlot) {
	if slot == nil {
		return
	}
	if slot.cancel != nil {
		slot.cancel()
		slot.cancel = nil
	}
	if slot.cleanup != nil {
		cleanup := slot.cleanup
		slot.cleanup = nil
		cleanup()
	}
}

func (app *App) queueEffectCleanup(i *instance) {
	for _, key := range i.effectOrder {
		if slot := i.effects[key]; slot != nil {
			app.pendingEffectCleanups = append(app.pendingEffectCleanups, slot)
		}
	}
	i.effects = nil
	i.effectOrder = nil
	i.pendingEffects = nil
	i.pendingEffectOrder = nil
}

func (app *App) shutdownEffects() any {
	var firstPanic any
	cleanup := func(slot *effectSlot) {
		if recovered := cleanupEffectSafely(slot); recovered != nil && firstPanic == nil {
			firstPanic = recovered
		}
	}
	for _, slot := range app.pendingEffectCleanups {
		cleanup(slot)
	}
	app.pendingEffectCleanups = nil
	var visit func(*instance)
	visit = func(i *instance) {
		if i == nil {
			return
		}
		for _, child := range i.children {
			visit(child)
		}
		for _, key := range i.effectOrder {
			cleanup(i.effects[key])
		}
		i.effects = nil
		i.effectOrder = nil
	}
	visit(app.rootInstance)
	return firstPanic
}

func cleanupEffectSafely(slot *effectSlot) (recovered any) {
	defer func() { recovered = recover() }()
	runEffectCleanup(slot)
	return nil
}
