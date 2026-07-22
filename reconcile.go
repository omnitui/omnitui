package omnitui

import (
	"fmt"
	"reflect"
	"sync/atomic"

	"github.com/omnitui/omnitui/internal/core"
)

func reconcile(parent, old *instance, element Element, values map[uint64]any, app *App, path string) *instance {
	if core.KindOf(element) == core.KindNone {
		unmount(old)
		return nil
	}
	if !instanceIdentity(old, element) {
		unmount(old)
		old = nil
	}
	current := old
	if current == nil {
		current = &instance{app: app, mounted: true}
	}
	current.parent = parent
	current.element = element
	current.path = path
	current.mounted = true
	switch core.KindOf(element) {
	case core.KindComponent:
		current.def = core.ComponentOf(element)
		current.props = core.PropsOf(element)
		if old == nil {
			current.state = current.def.InitialState(current.props)
		}
		applyPending(current)
		children := core.ChildrenOf(element)
		rendering := &atomic.Bool{}
		context := Context{instance: current, dispatcher: app.dispatcher, values: cloneContextValues(values), rendering: rendering}
		rendering.Store(true)
		var output Element
		func() {
			defer rendering.Store(false)
			current.beginHooks()
			output = current.def.Render(context, current.props, current.state, Children(children))
			current.finishHooks()
		}()
		current.children = reconcileChildren(current, current.children, []Element{output}, values, app, path+"/render")
	case core.KindProvider:
		provider := core.ProviderOf(element)
		next := cloneContextValues(values)
		if next == nil {
			next = map[uint64]any{}
		}
		next[provider.KeyID] = provider.Value
		current.children = reconcileChildren(current, current.children, core.ChildrenOf(element), next, app, path+"/provider")
	case core.KindFragment:
		current.children = reconcileChildren(current, current.children, core.ChildrenOf(element), values, app, path)
	case core.KindHost:
		previousFocus := focusBindingOf(current)
		if old != nil && old.host.Kind == core.HostList {
			captureListAnchor(old)
		}
		current.host, _ = core.HostOf(element)
		syncFocusedBinding(current, previousFocus)
		current.children = reconcileChildren(current, current.children, logicalHostChildren(current.host), values, app, path)
	}
	return current
}

func captureListAnchor(i *instance) {
	if i == nil || i.host.Kind != core.HostList {
		return
	}
	for index, child := range i.children {
		if core.KeyOf(child.element) == "" {
			continue
		}
		if child.rect.Y+child.rect.Height > i.rect.Y {
			i.listAnchorKey = core.KeyOf(child.element)
			i.listAnchorOffset = child.rect.Y - i.rect.Y
			i.listAnchorIndex = index
			return
		}
	}
	i.listAnchorKey, i.listAnchorOffset, i.listAnchorIndex = "", 0, 0
}

func logicalHostChildren(host core.Host) []Element {
	if host.Kind == core.HostTabs {
		data, _ := host.Data.(core.TabsData)
		key := data.ActiveKey
		if key == "" {
			for _, item := range data.Items {
				if !item.Disabled {
					key = item.Key
					break
				}
			}
		}
		for _, item := range data.Items {
			if item.Key == key {
				return []Element{item.Content.WithKey(key)}
			}
		}
		return nil
	}
	if host.Kind == core.HostList && len(host.Children) == 0 {
		if data, ok := host.Data.(core.ListData); ok && core.KindOf(data.Empty) != core.KindNone {
			return []Element{data.Empty}
		}
		return nil
	}
	return core.ChildrenOf(core.NewHost(host.Kind, host.Data, host.Children))
}

func reconcileChildren(parent *instance, old []*instance, elements []Element, values map[uint64]any, app *App, path string) []*instance {
	seen := map[string]bool{}
	for index, element := range elements {
		key := core.KeyOf(element)
		if key != "" {
			if seen[key] {
				panic(fmt.Sprintf("omnitui: duplicate key %q at %s/%d", key, path, index))
			}
			seen[key] = true
		}
	}
	used := make([]bool, len(old))
	result := make([]*instance, len(elements))
	for index, element := range elements {
		var previous *instance
		key := core.KeyOf(element)
		if key != "" {
			for oldIndex, candidate := range old {
				if !used[oldIndex] && core.KeyOf(candidate.element) == key {
					previous = candidate
					used[oldIndex] = true
					break
				}
			}
		} else if index < len(old) && !used[index] && core.KeyOf(old[index].element) == "" {
			previous = old[index]
			used[index] = true
		}
		result[index] = reconcile(parent, previous, element, values, app, fmt.Sprintf("%s/%d", path, index))
	}
	for index, candidate := range old {
		if !used[index] {
			unmount(candidate)
		}
	}
	return compactInstances(result)
}

func compactInstances(values []*instance) []*instance {
	result := values[:0]
	for _, value := range values {
		if value != nil {
			result = append(result, value)
		}
	}
	return result
}

func applyPending(i *instance) {
	for _, update := range i.pending {
		if !stateTypeMatches(i.state, update.typ) {
			panic(fmt.Sprintf("omnitui: state type mismatch at %s", i.path))
		}
		i.state = update.apply(i.state)
	}
	i.pending = nil
}

func stateTypeMatches(value any, expected reflect.Type) bool {
	if value == nil {
		return expected == nil || expected.Kind() == reflect.Interface
	}
	return reflect.TypeOf(value) == expected
}

func unmount(i *instance) {
	if i == nil || !i.mounted {
		return
	}
	for _, child := range i.children {
		unmount(child)
	}
	i.mounted = false
	if i.app != nil {
		i.app.queueEffectCleanup(i)
		for _, binding := range i.focusHandles {
			binding.active.Store(false)
			binding.focused.Store(false)
		}
		i.focusHandles = nil
		i.refs = nil
		if i.app.focused == i {
			i.app.setFocus(nil, ElementRemoved)
			i.app.focusLost = true
		}
		if i.app.capture == i {
			i.app.capture = nil
		}
		if i.app.pressTarget == i {
			i.app.pressTarget = nil
		}
		filtered := i.app.hoverPath[:0]
		for _, hovered := range i.app.hoverPath {
			if !isDescendant(hovered, i) {
				filtered = append(filtered, hovered)
			}
		}
		i.app.hoverPath = filtered
	}
}

func isDescendant(value, ancestor *instance) bool {
	for current := value; current != nil; current = current.parent {
		if current == ancestor {
			return true
		}
	}
	return false
}
