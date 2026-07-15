package components

import (
	"fmt"

	"github.com/viniciusfonseca/omnitui"
	"github.com/viniciusfonseca/omnitui/internal/core"
)

type Orientation uint8

const (
	OrientationHorizontal Orientation = iota
	OrientationVertical
)

type TabItem struct {
	Key, Label string
	Content    omnitui.Element
	Disabled   bool
}

type TabsProps struct {
	Items       []TabItem
	ActiveKey   string
	Orientation Orientation
	Style       omnitui.Style
	ActiveStyle omnitui.Style
	OnChange    omnitui.EventHandler[omnitui.ValueChangeEvent]
}

type tabsComponent struct{}

func (tabsComponent) InitialState(TabsProps) struct{} { return struct{}{} }
func (tabsComponent) Render(_ omnitui.Context, props TabsProps, _ struct{}, _ omnitui.Children) omnitui.Element {
	return tabsHost(props)
}

var tabsType = omnitui.Define[TabsProps, struct{}]("Tabs", tabsComponent{})

func Tabs(props TabsProps) omnitui.Element {
	props.Items = append([]TabItem(nil), props.Items...)
	validateStyle(props.Style)
	validateStyle(props.ActiveStyle)
	items := make([]core.TabData, len(props.Items))
	seen := map[string]struct{}{}
	firstEnabled := ""
	for index, item := range props.Items {
		if item.Key == "" {
			panic(fmt.Sprintf("omnitui/components: empty tab key at index %d", index))
		}
		if _, ok := seen[item.Key]; ok {
			panic(fmt.Sprintf("omnitui/components: duplicate tab key %q", item.Key))
		}
		seen[item.Key] = struct{}{}
		if !item.Disabled && firstEnabled == "" {
			firstEnabled = item.Key
		}
		items[index] = core.TabData{Key: item.Key, Label: item.Label, Content: item.Content, Disabled: item.Disabled}
	}
	if props.ActiveKey != "" {
		item, ok := seen[props.ActiveKey]
		if !ok {
			panic(fmt.Sprintf("omnitui/components: unknown active tab %q", props.ActiveKey))
		}
		for _, candidate := range props.Items {
			if candidate.Key == props.ActiveKey && candidate.Disabled {
				panic(fmt.Sprintf("omnitui/components: active tab %q is disabled", props.ActiveKey))
			}
		}
		_ = item
	} else {
		_ = firstEnabled
	}
	return omnitui.Create(tabsType, props)
}

func tabsHost(props TabsProps) omnitui.Element {
	items := make([]core.TabData, len(props.Items))
	for index, item := range props.Items {
		items[index] = core.TabData{Key: item.Key, Label: item.Label, Content: item.Content, Disabled: item.Disabled}
	}
	return core.NewHost(core.HostTabs, core.TabsData{Items: items, ActiveKey: props.ActiveKey, Orientation: uint8(props.Orientation), Style: props.Style, ActiveStyle: props.ActiveStyle, Handlers: handlers(map[string]any{"change": props.OnChange})}, nil)
}
