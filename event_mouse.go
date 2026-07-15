package omnitui

type MouseAction uint8

const (
	MouseMove MouseAction = iota
	MouseDown
	MouseUp
	MouseEnter
	MouseLeave
)

type MouseButton uint8

const (
	MouseButtonNone MouseButton = iota
	MouseButtonLeft
	MouseButtonMiddle
	MouseButtonRight
)

type MouseButtons uint8

const (
	MouseLeftPressed MouseButtons = 1 << iota
	MouseMiddlePressed
	MouseRightPressed
)

type MouseEvent struct {
	Action         MouseAction
	Button         MouseButton
	Buttons        MouseButtons
	X, Y           int
	LocalX, LocalY int
	Modifiers      Modifiers
}

type WheelEvent struct {
	X, Y           int
	LocalX, LocalY int
	DeltaX, DeltaY int
	Modifiers      Modifiers
}
