package components

import (
	"github.com/viniciusfonseca/omnitui"
	"github.com/viniciusfonseca/omnitui/internal/core"
)

type InputProps struct {
	Value       string
	Placeholder string
	Width       omnitui.Size
	Disabled    bool
	ReadOnly    bool
	Mask        rune
	MaxLength   int
	Style       omnitui.Style
	FocusStyle  omnitui.Style
	OnChange    omnitui.EventHandler[omnitui.ValueChangeEvent]
	OnSubmit    omnitui.EventHandler[omnitui.SubmitEvent]
	OnKey       omnitui.EventHandler[omnitui.KeyEvent]
	OnTextInput omnitui.EventHandler[omnitui.TextInputEvent]
	OnPaste     omnitui.EventHandler[omnitui.PasteEvent]
	OnFocus     omnitui.EventHandler[omnitui.FocusEvent]
	OnBlur      omnitui.EventHandler[omnitui.BlurEvent]
	OnMouse     omnitui.EventHandler[omnitui.MouseEvent]
}

type inputComponent struct{}

func (inputComponent) InitialState(InputProps) struct{} { return struct{}{} }
func (inputComponent) Render(_ omnitui.Context, props InputProps, _ struct{}, _ omnitui.Children) omnitui.Element {
	return inputHost(props)
}

var inputType = omnitui.Define[InputProps, struct{}]("Input", inputComponent{})

func Input(props InputProps) omnitui.Element {
	if props.MaxLength < 0 {
		panic("omnitui/components: MaxLength cannot be negative")
	}
	validateStyle(props.Style)
	validateStyle(props.FocusStyle)
	return omnitui.Create(inputType, props)
}

func inputHost(props InputProps) omnitui.Element {
	return core.NewHost(core.HostInput, core.InputData{Value: props.Value, Placeholder: props.Placeholder, Width: props.Width, Disabled: props.Disabled, ReadOnly: props.ReadOnly, Mask: props.Mask, MaxLength: props.MaxLength, Style: props.Style, FocusStyle: props.FocusStyle, Handlers: handlers(map[string]any{"change": props.OnChange, "submit": props.OnSubmit, "key": props.OnKey, "text": props.OnTextInput, "paste": props.OnPaste, "focus": props.OnFocus, "blur": props.OnBlur, "mouse": props.OnMouse})}, nil)
}
