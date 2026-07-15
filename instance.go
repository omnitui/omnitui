package omnitui

import (
	"fmt"
	"sync/atomic"

	"github.com/viniciusfonseca/omnitui/internal/core"
)

type instance struct {
	element             core.Element
	parent              *instance
	children            []*instance
	app                 *App
	mounted             bool
	path                string
	def                 *core.ComponentDefinition
	props               any
	state               any
	pending             []stateUpdate
	rendering           atomic.Bool
	host                core.Host
	rect                Rect
	clip                Rect
	style               Style
	inputCursor         int
	inputOffset         int
	tabFocus            string
	listOffset          int
	listAnchorKey       string
	listAnchorOffset    int
	listAnchorIndex     int
	listLastSelected    string
	listManual          bool
	listInvalidSelected string
	listInvalidProposal string
}

func (i *instance) kind() core.Kind { return core.KindOf(i.element) }

func (i *instance) hostKind() core.HostKind { return i.host.Kind }

func (i *instance) disabled() bool {
	switch data := i.host.Data.(type) {
	case core.BoxData:
		return data.Disabled
	case core.ButtonData:
		return data.Disabled
	case core.InputData:
		return data.Disabled
	case core.ListData:
		return data.Disabled
	default:
		return false
	}
}

func (i *instance) focusable() bool {
	if i.kind() != core.KindHost || i.disabled() {
		return false
	}
	switch data := i.host.Data.(type) {
	case core.BoxData:
		return data.Focusable
	case core.ButtonData, core.InputData, core.TabsData, core.ListData:
		return true
	default:
		return false
	}
}

func (i *instance) pressable() bool {
	if i.kind() != core.KindHost || i.disabled() {
		return false
	}
	if i.host.Kind == core.HostButton {
		return true
	}
	return i.handlerValue("press") != nil
}

func (i *instance) handlerValue(name string) any {
	if i.kind() != core.KindHost {
		return nil
	}
	switch data := i.host.Data.(type) {
	case core.BoxData:
		return data.Handlers[name]
	case core.ButtonData:
		return data.Handlers[name]
	case core.InputData:
		return data.Handlers[name]
	case core.TabsData:
		return data.Handlers[name]
	case core.ListData:
		return data.Handlers[name]
	default:
		return nil
	}
}

func eventPath(target *instance) []*instance {
	var reversed []*instance
	for current := target; current != nil; current = current.parent {
		reversed = append(reversed, current)
	}
	result := make([]*instance, len(reversed))
	for index := range reversed {
		result[len(reversed)-1-index] = reversed[index]
	}
	return result
}

func instanceIdentity(old *instance, element Element) bool {
	if old == nil || core.KindOf(element) != old.kind() || core.KeyOf(element) != core.KeyOf(old.element) {
		return false
	}
	if core.KindOf(element) == core.KindComponent {
		return old.def == core.ComponentOf(element)
	}
	if core.KindOf(element) == core.KindProvider {
		return old.elementProviderKey() == core.ProviderOf(element).KeyID
	}
	if core.KindOf(element) == core.KindHost {
		host, _ := core.HostOf(element)
		return old.host.Kind == host.Kind
	}
	return true
}

func (i *instance) elementProviderKey() uint64 { return core.ProviderOf(i.element).KeyID }

func (i *instance) describe() string {
	if i == nil {
		return "<nil>"
	}
	if i.def != nil && i.def.Name != "" {
		return i.def.Name
	}
	if i.kind() == core.KindHost {
		return fmt.Sprintf("host(%d)", i.host.Kind)
	}
	return fmt.Sprintf("element(%d)", i.kind())
}
