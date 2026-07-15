package omnitui

type EventResult uint8

const (
	Propagate EventResult = iota
	Consume
)

type EventHandler[E any] func(event E) EventResult

type ChangeSource uint8

const (
	ChangeKeyboard ChangeSource = iota
	ChangePaste
	ChangeProgrammatic
)

type TextInputEvent struct{ Text string }
type PasteEvent struct{ Text string }

type ValueChangeEvent struct {
	Previous string
	Value    string
	Source   ChangeSource
}

type SubmitEvent struct{ Value string }

type PressSource uint8

const (
	KeyboardEnter PressSource = iota
	KeyboardSpace
	MouseLeft
	ProgrammaticPress
)

type PressEvent struct{ Source PressSource }
type ActivateEvent struct {
	Key    string
	Source PressSource
}

type FocusCause uint8

const (
	ProgrammaticFocus FocusCause = iota
	ForwardTraversal
	BackwardTraversal
	ElementRemoved
)

type FocusEvent struct{ Cause FocusCause }
type BlurEvent struct{ Cause FocusCause }

type ResizeEvent struct{ Width, Height int }
type MessageEvent struct{ Value any }
