package components

import (
	"fmt"

	"github.com/omnitui/omnitui/v2"
	"github.com/omnitui/omnitui/v2/internal/core"
)

type GridItemProps struct {
	InitialSize int
	MinSize     int
	MaxSize     int
}

type gridItemData struct{ Props GridItemProps }
type gridItemComponent struct{}

func (gridItemComponent) InitialState(gridItemData) struct{} { return struct{}{} }
func (gridItemComponent) Render(_ omnitui.Context, _ gridItemData, _ struct{}, children omnitui.Children) omnitui.Element {
	return children[0]
}

var gridItemType = omnitui.Define[gridItemData, struct{}]("GridItem", gridItemComponent{})

func GridItem(props GridItemProps, child omnitui.Element) omnitui.Element {
	validateGridItemSize("InitialSize", props.InitialSize)
	validateGridItemSize("MinSize", props.MinSize)
	validateGridItemSize("MaxSize", props.MaxSize)
	if props.MinSize > 0 && props.MaxSize > 0 && props.MaxSize < props.MinSize {
		panic("omnitui/components: grid item MaxSize cannot be smaller than MinSize")
	}
	if props.InitialSize > 0 && props.MinSize > 0 && props.InitialSize < props.MinSize {
		panic("omnitui/components: grid item InitialSize cannot be smaller than MinSize")
	}
	if props.InitialSize > 0 && props.MaxSize > 0 && props.InitialSize > props.MaxSize {
		panic("omnitui/components: grid item InitialSize cannot be greater than MaxSize")
	}
	element := omnitui.Create(gridItemType, gridItemData{Props: props}, child)
	if key := core.KeyOf(child); key != "" {
		element = element.WithKey(key)
	}
	return element
}

func validateGridItemSize(name string, value int) {
	if value < 0 {
		panic(fmt.Sprintf("omnitui/components: grid item %s cannot be negative", name))
	}
	if value > 0 && value < 3 {
		panic(fmt.Sprintf("omnitui/components: grid item %s must be at least 3", name))
	}
}

func gridItemProps(element omnitui.Element) (GridItemProps, bool) {
	if core.KindOf(element) != core.KindComponent {
		return GridItemProps{}, false
	}
	data, ok := core.PropsOf(element).(gridItemData)
	return data.Props, ok
}
