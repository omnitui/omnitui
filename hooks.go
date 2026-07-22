package omnitui

import "fmt"

func hookInstance(ctx Context, name string) *instance {
	if ctx.instance == nil || ctx.dispatcher == nil {
		panic(fmt.Sprintf("omnitui: %s called outside component render", name))
	}
	if ctx.rendering == nil || !ctx.rendering.Load() {
		panic(fmt.Sprintf("omnitui: %s called outside Render at %s", name, ctx.instance.path))
	}
	return ctx.instance
}

func validateHookKey(instance *instance, name, key string) {
	if key == "" {
		panic(fmt.Sprintf("omnitui: %s key cannot be empty at %s", name, instance.path))
	}
}

func (i *instance) beginHooks() {
	i.pendingEffects = nil
	i.pendingEffectOrder = i.pendingEffectOrder[:0]
	i.seenRefs = nil
	i.seenFocus = nil
}

func (i *instance) finishHooks() {
	for key := range i.refs {
		if _, ok := i.seenRefs[key]; !ok {
			delete(i.refs, key)
		}
	}
	for key, binding := range i.focusHandles {
		if _, ok := i.seenFocus[key]; ok {
			continue
		}
		binding.active.Store(false)
		binding.focused.Store(false)
		delete(i.focusHandles, key)
	}
	i.seenRefs = nil
	i.seenFocus = nil
	if len(i.effects) > 0 || len(i.pendingEffects) > 0 {
		i.app.effectInstances = append(i.app.effectInstances, i)
	}
}
