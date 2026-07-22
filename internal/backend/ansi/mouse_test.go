package ansi

import (
	"testing"

	"github.com/omnitui/omnitui/v2/internal/backend"
)

func TestParserSGRMouse(t *testing.T) {
	parser := &Parser{}
	events := parser.Feed([]byte("\x1b[<0;4;5M\x1b[<64;4;5M\x1b[<0;4;5m"))
	if len(events) != 3 {
		t.Fatalf("events = %#v", events)
	}
	down := events[0].Value.(backend.MouseInput)
	if down.Action != 1 || down.Button != 1 || down.X != 3 || down.Y != 4 || down.Buttons != 1 {
		t.Fatalf("down = %#v", down)
	}
	wheel := events[1].Value.(backend.WheelInput)
	if wheel.DeltaY != -1 {
		t.Fatalf("wheel = %#v", wheel)
	}
	up := events[2].Value.(backend.MouseInput)
	if up.Action != 2 || up.Buttons != 0 {
		t.Fatalf("up = %#v", up)
	}
}
