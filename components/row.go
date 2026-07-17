package components

import (
	"github.com/viniciusfonseca/omnitui"
	"github.com/viniciusfonseca/omnitui/internal/core"
)

type RowProps struct {
	Width, Height        omnitui.Size
	MinWidth, MaxWidth   int
	MinHeight, MaxHeight int
	FlexGrow             int
	Padding              omnitui.Spacing
	Gap                  int
	Align                Align
	Justify              Justify
	Wrap                 bool
	Clip                 bool
	Style                omnitui.Style
}

type rowComponent struct{}

func (rowComponent) InitialState(RowProps) struct{} { return struct{}{} }
func (rowComponent) Render(_ omnitui.Context, props RowProps, _ struct{}, children omnitui.Children) omnitui.Element {
	return rowHost(props, children...)
}

var rowType = omnitui.Define[RowProps, struct{}]("Row", rowComponent{})

func Row(props RowProps, children ...omnitui.Element) omnitui.Element {
	validateBox(props.MinWidth, props.MaxWidth, props.MinHeight, props.MaxHeight, props.FlexGrow, props.Gap, props.Padding)
	validateStyle(props.Style)
	return omnitui.Create(rowType, props, children...)
}

func rowHost(props RowProps, children ...omnitui.Element) omnitui.Element {
	return core.NewHost(core.HostBox, core.BoxData{Width: props.Width, Height: props.Height, MinWidth: props.MinWidth, MaxWidth: props.MaxWidth, MinHeight: props.MinHeight, MaxHeight: props.MaxHeight, FlexGrow: props.FlexGrow, Padding: props.Padding, Gap: props.Gap, Direction: uint8(Horizontal), Align: uint8(props.Align), Justify: uint8(props.Justify), Wrap: props.Wrap, Clip: props.Clip, Style: props.Style}, children)
}
