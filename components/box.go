package components

import (
	"fmt"

	"github.com/viniciusfonseca/omnitui"
	"github.com/viniciusfonseca/omnitui/internal/core"
)

type Direction uint8

const (
	Horizontal Direction = iota
	Vertical
)

type Align uint8

const (
	AlignStart Align = iota
	AlignCenter
	AlignEnd
	AlignStretch
)

type Justify uint8

const (
	JustifyStart Justify = iota
	JustifyCenter
	JustifyEnd
	JustifySpaceBetween
	JustifySpaceAround
)

type BorderStyle uint8

const (
	BorderNone BorderStyle = iota
	BorderSingle
	BorderRounded
	BorderDouble
	BorderHeavy
)

type BoxProps struct {
	Width, Height        omnitui.Size
	MinWidth, MaxWidth   int
	MinHeight, MaxHeight int
	Padding              omnitui.Spacing
	Gap                  int
	Direction            Direction
	Align                Align
	Justify              Justify
	Wrap                 bool
	Clip                 bool
	Border               BorderStyle
	Style                omnitui.Style
	Focusable            bool
	Disabled             bool
	OnKey                omnitui.EventHandler[omnitui.KeyEvent]
	OnTextInput          omnitui.EventHandler[omnitui.TextInputEvent]
	OnPaste              omnitui.EventHandler[omnitui.PasteEvent]
	OnFocus              omnitui.EventHandler[omnitui.FocusEvent]
	OnBlur               omnitui.EventHandler[omnitui.BlurEvent]
	OnPress              omnitui.EventHandler[omnitui.PressEvent]
	OnMouse              omnitui.EventHandler[omnitui.MouseEvent]
	OnWheel              omnitui.EventHandler[omnitui.WheelEvent]
	OnResize             omnitui.EventHandler[omnitui.ResizeEvent]
	OnMessage            omnitui.EventHandler[omnitui.MessageEvent]
}

func Box(props BoxProps, children ...omnitui.Element) omnitui.Element {
	validateBox(props.MinWidth, props.MaxWidth, props.MinHeight, props.MaxHeight, props.Gap, props.Padding)
	validateStyle(props.Style)
	return core.NewHost(core.HostBox, core.BoxData{
		Width: props.Width, Height: props.Height, MinWidth: props.MinWidth, MaxWidth: props.MaxWidth,
		MinHeight: props.MinHeight, MaxHeight: props.MaxHeight, Padding: props.Padding, Gap: props.Gap,
		Direction: uint8(props.Direction), Align: uint8(props.Align), Justify: uint8(props.Justify),
		Wrap: props.Wrap, Clip: props.Clip, Border: uint8(props.Border), Style: props.Style,
		Focusable: props.Focusable, Disabled: props.Disabled,
		Handlers: handlers(map[string]any{
			"key": props.OnKey, "text": props.OnTextInput, "paste": props.OnPaste, "focus": props.OnFocus,
			"blur": props.OnBlur, "press": props.OnPress, "mouse": props.OnMouse, "wheel": props.OnWheel,
			"resize": props.OnResize, "message": props.OnMessage,
		}),
	}, children)
}

func validateStyle(style omnitui.Style) {
	if err := omnitui.ValidateStyle(style); err != nil {
		panic(err)
	}
}

func validateBox(minWidth, maxWidth, minHeight, maxHeight, gap int, padding omnitui.Spacing) {
	if minWidth < 0 || maxWidth < 0 || minHeight < 0 || maxHeight < 0 || gap < 0 {
		panic("omnitui/components: dimensions and gap cannot be negative")
	}
	if minWidth > 0 && maxWidth > 0 && minWidth > maxWidth {
		panic(fmt.Sprintf("omnitui/components: min width %d exceeds max width %d", minWidth, maxWidth))
	}
	if minHeight > 0 && maxHeight > 0 && minHeight > maxHeight {
		panic(fmt.Sprintf("omnitui/components: min height %d exceeds max height %d", minHeight, maxHeight))
	}
	if err := omnitui.ValidateSpacing(padding); err != nil {
		panic(err)
	}
}

func handlers(values map[string]any) core.Handlers {
	result := core.Handlers{}
	for name, value := range values {
		if value != nil {
			result[name] = value
		}
	}
	return result
}
