package components

import (
	"fmt"

	"github.com/viniciusfonseca/omnitui"
	"github.com/viniciusfonseca/omnitui/internal/core"
)

type ScrollbarMode uint8

const (
	ScrollbarAuto ScrollbarMode = iota
	ScrollbarAlways
	ScrollbarHidden
)

type ListProps struct {
	SelectedKey   string
	Selectable    bool
	Height        omnitui.Size
	Gap           int
	Disabled      bool
	Wrap          bool
	ScrollPadding int
	Scrollbar     ScrollbarMode
	Empty         omnitui.Element
	Style         omnitui.Style
	SelectedStyle omnitui.Style
	OnChange      omnitui.EventHandler[omnitui.ValueChangeEvent]
	OnActivate    omnitui.EventHandler[omnitui.ActivateEvent]
	OnMouse       omnitui.EventHandler[omnitui.MouseEvent]
	OnWheel       omnitui.EventHandler[omnitui.WheelEvent]
}

type listComponent struct{}

func (listComponent) InitialState(ListProps) struct{} { return struct{}{} }
func (listComponent) Render(_ omnitui.Context, props ListProps, _ struct{}, children omnitui.Children) omnitui.Element {
	return listHost(props, children...)
}

var listType = omnitui.Define[ListProps, struct{}]("List", listComponent{})

func List(props ListProps, items ...omnitui.Element) omnitui.Element {
	if props.Gap < 0 || props.ScrollPadding < 0 {
		panic("omnitui/components: list spacing cannot be negative")
	}
	validateStyle(props.Style)
	validateStyle(props.SelectedStyle)
	seen := map[string]struct{}{}
	for index, item := range items {
		key := core.KeyOf(item)
		if props.Selectable && key == "" {
			panic(fmt.Sprintf("omnitui/components: list item %d has no key", index))
		}
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			panic(fmt.Sprintf("omnitui/components: duplicate list item key %q", key))
		}
		seen[key] = struct{}{}
	}
	return omnitui.Create(listType, props, items...)
}

func listHost(props ListProps, items ...omnitui.Element) omnitui.Element {
	return core.NewHost(core.HostList, core.ListData{SelectedKey: props.SelectedKey, Selectable: props.Selectable, Height: props.Height, Gap: props.Gap, Disabled: props.Disabled, Wrap: props.Wrap, ScrollPadding: props.ScrollPadding, Scrollbar: uint8(props.Scrollbar), Empty: props.Empty, Style: props.Style, SelectedStyle: props.SelectedStyle, Handlers: handlers(map[string]any{"change": props.OnChange, "activate": props.OnActivate, "mouse": props.OnMouse, "wheel": props.OnWheel})}, items)
}
