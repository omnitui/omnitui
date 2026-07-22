package omnitui

import (
	"fmt"
	"sync/atomic"

	"github.com/omnitui/omnitui/internal/core"
)

type focusBinding struct {
	app     *App
	owner   *instance
	key     string
	active  atomic.Bool
	focused atomic.Bool
}

// FocusHandle controls one focusable host bound by its Focus prop. Its zero
// value is inert.
type FocusHandle struct {
	binding   *focusBinding
	rendering *atomic.Bool
}

// UseFocus returns a keyed focus binding for the current component instance.
func UseFocus(ctx Context, key string) FocusHandle {
	instance := hookInstance(ctx, "UseFocus")
	validateHookKey(instance, "UseFocus", key)
	if instance.seenFocus == nil {
		instance.seenFocus = make(map[string]struct{})
	}
	if _, exists := instance.seenFocus[key]; exists {
		panic("omnitui: duplicate UseFocus key " + key + " at " + instance.path)
	}
	instance.seenFocus[key] = struct{}{}
	if binding := instance.focusHandles[key]; binding != nil {
		binding.active.Store(true)
		return FocusHandle{binding: binding, rendering: ctx.rendering}
	}
	if instance.focusHandles == nil {
		instance.focusHandles = make(map[string]*focusBinding)
	}
	binding := &focusBinding{app: instance.app, owner: instance, key: key}
	binding.active.Store(true)
	instance.focusHandles[key] = binding
	return FocusHandle{binding: binding, rendering: ctx.rendering}
}

// Request queues focus for the bound host when it is visible and enabled.
func (handle FocusHandle) Request() {
	binding := handle.binding
	if binding == nil || !binding.active.Load() {
		return
	}
	if handle.rendering != nil && handle.rendering.Load() {
		panic("omnitui: FocusHandle.Request called during Render at " + binding.owner.path)
	}
	binding.app.dispatcher.enqueue(func() {
		if !binding.active.Load() || !binding.owner.mounted {
			return
		}
		target := findFocusTarget(binding.app.rootInstance, binding)
		if target == nil || !target.focusable() || target.clip.Empty() {
			return
		}
		binding.app.setFocus(target, ProgrammaticFocus)
	})
}

// Blur queues focus removal when the bound host is currently focused.
func (handle FocusHandle) Blur() {
	binding := handle.binding
	if binding == nil || !binding.active.Load() {
		return
	}
	if handle.rendering != nil && handle.rendering.Load() {
		panic("omnitui: FocusHandle.Blur called during Render at " + binding.owner.path)
	}
	binding.app.dispatcher.enqueue(func() {
		if binding.active.Load() && focusBindingOf(binding.app.focused) == binding {
			binding.app.setFocus(nil, ProgrammaticFocus)
		}
	})
}

// Focused reports whether the bound host currently owns focus.
func (handle FocusHandle) Focused() bool {
	return handle.binding != nil && handle.binding.active.Load() && handle.binding.focused.Load()
}

func findFocusTarget(root *instance, binding *focusBinding) *instance {
	if root == nil {
		return nil
	}
	if focusBindingOf(root) == binding {
		return root
	}
	for _, child := range root.children {
		if target := findFocusTarget(child, binding); target != nil {
			return target
		}
	}
	return nil
}

func focusBindingOf(i *instance) *focusBinding {
	if i == nil || i.kind() != core.KindHost {
		return nil
	}
	var value any
	switch data := i.host.Data.(type) {
	case core.BoxData:
		value = data.Focus
	case core.ButtonData:
		value = data.Focus
	case core.InputData:
		value = data.Focus
	case core.TabsData:
		value = data.Focus
	case core.ListData:
		value = data.Focus
	}
	handle, _ := value.(FocusHandle)
	return handle.binding
}

func syncFocusedBinding(i *instance, previous *focusBinding) {
	if i == nil || i.app == nil || i.app.focused != i {
		return
	}
	next := focusBindingOf(i)
	if previous == next {
		return
	}
	if previous != nil {
		previous.focused.Store(false)
	}
	if next != nil {
		next.focused.Store(true)
	}
}

func validateFocusBindings(root *instance) {
	expected := make(map[*focusBinding]struct{})
	attached := make(map[*focusBinding]int)
	var visit func(*instance)
	visit = func(i *instance) {
		if i == nil {
			return
		}
		for _, binding := range i.focusHandles {
			if binding.active.Load() {
				expected[binding] = struct{}{}
			}
		}
		if binding := focusBindingOf(i); binding != nil && binding.active.Load() {
			attached[binding]++
		}
		for _, child := range i.children {
			visit(child)
		}
	}
	visit(root)
	for binding := range expected {
		if attached[binding] != 1 {
			panic(fmt.Sprintf("omnitui: UseFocus key %q at %s is attached to %d hosts, want 1", binding.key, binding.owner.path, attached[binding]))
		}
	}
	for binding := range attached {
		if _, ok := expected[binding]; !ok {
			panic("omnitui: FocusHandle belongs to a different component tree")
		}
	}
}
