package components

import (
	"github.com/viniciusfonseca/omnitui"
	"github.com/viniciusfonseca/omnitui/internal/core"
)

type ColumnProps struct {
	Width, Height        omnitui.Size
	MinWidth, MaxWidth   int
	MinHeight, MaxHeight int
	Padding              omnitui.Spacing
	Gap                  int
	Align                Align
	Justify              Justify
	Clip                 bool
	Style                omnitui.Style
}

type columnComponent struct{}

func (columnComponent) InitialState(ColumnProps) struct{} { return struct{}{} }
func (columnComponent) Render(_ omnitui.Context, props ColumnProps, _ struct{}, children omnitui.Children) omnitui.Element {
	return columnHost(props, children...)
}

var columnType = omnitui.Define[ColumnProps, struct{}]("Column", columnComponent{})

func Column(props ColumnProps, children ...omnitui.Element) omnitui.Element {
	validateBox(props.MinWidth, props.MaxWidth, props.MinHeight, props.MaxHeight, props.Gap, props.Padding)
	validateStyle(props.Style)
	return omnitui.Create(columnType, props, children...)
}

func columnHost(props ColumnProps, children ...omnitui.Element) omnitui.Element {
	return core.NewHost(core.HostBox, core.BoxData{Width: props.Width, Height: props.Height, MinWidth: props.MinWidth, MaxWidth: props.MaxWidth, MinHeight: props.MinHeight, MaxHeight: props.MaxHeight, Padding: props.Padding, Gap: props.Gap, Direction: uint8(Vertical), Align: uint8(props.Align), Justify: uint8(props.Justify), Clip: props.Clip, Style: props.Style}, children)
}
