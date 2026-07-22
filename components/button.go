package components

import (
	"github.com/omnitui/omnitui/v2"
	"github.com/omnitui/omnitui/v2/internal/core"
)

type ButtonProps struct {
	Label         string
	Plain         bool
	Disabled      bool
	Style         omnitui.Style
	FocusStyle    omnitui.Style
	DisabledStyle omnitui.Style
	Focus         omnitui.FocusHandle
	OnKey         omnitui.EventHandler[omnitui.KeyEvent]
	OnFocus       omnitui.EventHandler[omnitui.FocusEvent]
	OnBlur        omnitui.EventHandler[omnitui.BlurEvent]
	OnPress       omnitui.EventHandler[omnitui.PressEvent]
	OnMouse       omnitui.EventHandler[omnitui.MouseEvent]
}

type buttonComponent struct{}

func (buttonComponent) InitialState(ButtonProps) struct{} { return struct{}{} }
func (buttonComponent) Render(_ omnitui.Context, props ButtonProps, _ struct{}, _ omnitui.Children) omnitui.Element {
	return buttonHost(props)
}

var buttonType = omnitui.Define[ButtonProps, struct{}]("Button", buttonComponent{})

func Button(props ButtonProps) omnitui.Element {
	validateStyle(props.Style)
	validateStyle(props.FocusStyle)
	validateStyle(props.DisabledStyle)
	return omnitui.Create(buttonType, props)
}

func buttonHost(props ButtonProps) omnitui.Element {
	return core.NewHost(core.HostButton, core.ButtonData{Label: props.Label, Plain: props.Plain, Disabled: props.Disabled, Style: props.Style, FocusStyle: props.FocusStyle, DisabledStyle: props.DisabledStyle, Focus: props.Focus, Handlers: handlers(map[string]any{"key": props.OnKey, "focus": props.OnFocus, "blur": props.OnBlur, "press": props.OnPress, "mouse": props.OnMouse})}, nil)
}
