package backend

type KeyInput struct {
	Key       uint16
	Rune      rune
	Modifiers uint8
	Repeat    bool
}

type MouseInput struct {
	Action    uint8
	Button    uint8
	Buttons   uint8
	X, Y      int
	Modifiers uint8
}

type WheelInput struct {
	X, Y, DeltaX, DeltaY int
	Modifiers            uint8
}
type ResizeInput struct{ Width, Height int }
type PasteInput struct{ Text string }
