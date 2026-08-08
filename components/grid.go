package components

import (
	"fmt"

	"github.com/omnitui/omnitui/v2"
	"github.com/omnitui/omnitui/v2/internal/core"
)

type GridProps struct {
	Width, Height omnitui.Size
	FlexGrow      int
	Orientation   Orientation
	MinPanelSize  int
	Border        BorderStyle
	Style         omnitui.Style
}

type gridComponent struct{}

func (gridComponent) InitialState(GridProps) struct{} { return struct{}{} }
func (gridComponent) Render(_ omnitui.Context, props GridProps, _ struct{}, children omnitui.Children) omnitui.Element {
	return gridHost(props, children...)
}

var gridType = omnitui.Define[GridProps, struct{}]("Grid", gridComponent{})

func Grid(props GridProps, children ...omnitui.Element) omnitui.Element {
	if props.Orientation > OrientationVertical {
		panic("omnitui/components: invalid grid orientation")
	}
	if props.MinPanelSize < 0 {
		panic("omnitui/components: grid MinPanelSize cannot be negative")
	}
	if props.FlexGrow < 0 {
		panic("omnitui/components: grid FlexGrow cannot be negative")
	}
	if props.MinPanelSize == 0 {
		props.MinPanelSize = 3
	}
	if props.MinPanelSize < 3 {
		panic("omnitui/components: grid MinPanelSize must be at least 3")
	}
	if props.Border == BorderNone {
		props.Border = BorderSingle
	}
	if props.Border > BorderHeavy {
		panic("omnitui/components: invalid grid border")
	}
	validateStyle(props.Style)
	for index, child := range children {
		item, configured := gridItemProps(child)
		if !configured {
			continue
		}
		minimum := item.MinSize
		if minimum == 0 {
			minimum = props.MinPanelSize
		}
		if item.InitialSize > 0 && item.InitialSize < minimum {
			panic(fmt.Sprintf("omnitui/components: grid item %d InitialSize cannot be smaller than its effective minimum", index))
		}
		if item.MaxSize > 0 && item.MaxSize < minimum {
			panic(fmt.Sprintf("omnitui/components: grid item %d MaxSize cannot be smaller than its effective minimum", index))
		}
	}
	return omnitui.Create(gridType, props, children...)
}

func gridHost(props GridProps, children ...omnitui.Element) omnitui.Element {
	panels := make([]omnitui.Element, len(children))
	tracks := make([]core.GridTrackData, len(children))
	for index, child := range children {
		if item, ok := gridItemProps(child); ok {
			tracks[index] = core.GridTrackData{
				InitialSize: item.InitialSize,
				MinSize:     item.MinSize,
				MaxSize:     item.MaxSize,
			}
		}
		panel := Box(BoxProps{
			Align:  AlignStretch,
			Border: props.Border,
			Clip:   true,
			Style:  props.Style,
		}, child)
		if key := core.KeyOf(child); key != "" {
			panel = panel.WithKey(key)
		}
		panels[index] = panel
	}
	return core.NewHost(core.HostGrid, core.GridData{
		Width: props.Width, Height: props.Height, FlexGrow: props.FlexGrow,
		Orientation: uint8(props.Orientation), MinPanelSize: props.MinPanelSize,
		Border: uint8(props.Border), Style: props.Style, Tracks: tracks,
	}, panels)
}
