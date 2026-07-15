package omnitui

type Key uint16

const (
	KeyRune Key = iota
	KeyEnter
	KeyEscape
	KeyTab
	KeyBacktab
	KeyBackspace
	KeyDelete
	KeyInsert
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyHome
	KeyEnd
	KeyPageUp
	KeyPageDown
	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
)

type Modifiers uint8

const (
	ModCtrl Modifiers = 1 << iota
	ModAlt
	ModShift
)

type KeyEvent struct {
	Key       Key
	Rune      rune
	Modifiers Modifiers
	Repeat    bool
}
