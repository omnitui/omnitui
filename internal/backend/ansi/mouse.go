package ansi

import (
	"strconv"
	"strings"

	"github.com/omnitui/omnitui/v2/internal/backend"
)

func parseMouse(value string) (any, bool) {
	if len(value) == 0 {
		return nil, false
	}
	final := value[len(value)-1]
	fields := strings.Split(value[:len(value)-1], ";")
	if len(fields) != 3 {
		return nil, false
	}
	b, e1 := strconv.Atoi(fields[0])
	x, e2 := strconv.Atoi(fields[1])
	y, e3 := strconv.Atoi(fields[2])
	if e1 != nil || e2 != nil || e3 != nil {
		return nil, false
	}
	modifiers := uint8(0)
	if b&4 != 0 {
		modifiers |= 4
	}
	if b&8 != 0 {
		modifiers |= 2
	}
	if b&16 != 0 {
		modifiers |= 1
	}
	if b&64 != 0 {
		delta := -1
		if b&1 != 0 {
			delta = 1
		}
		return backend.WheelInput{X: x - 1, Y: y - 1, DeltaY: delta, Modifiers: modifiers}, true
	}
	button := uint8((b & 3) + 1)
	action := uint8(1)
	if final == 'm' {
		action = 2
	}
	if b&32 != 0 {
		action = 0
		button = 0
	}
	return backend.MouseInput{Action: action, Button: button, X: x - 1, Y: y - 1, Modifiers: modifiers}, true
}
