package omnitui

import (
	"testing"

	"github.com/omnitui/omnitui/v2/internal/core"
)

func TestFocusedInputClickRepaintsCursor(t *testing.T) {
	root := core.NewHost(
		core.HostInput,
		core.InputData{Value: "abcde", Width: core.CellsSize(5)},
		nil,
	)
	app := New(root, Options{})
	app.width, app.height = 5, 1
	if err := app.render(); err != nil {
		t.Fatal(err)
	}

	input := app.rootInstance
	app.setFocus(input, ProgrammaticFocus)
	input.inputCursor = 5
	if err := app.cycle(); err != nil {
		t.Fatal(err)
	}
	if got := paintedCursorX(app); got != 4 {
		t.Fatalf("initial painted cursor X = %d, want 4", got)
	}

	app.dispatchMouse(MouseEvent{Action: MouseDown, Button: MouseButtonLeft, X: 0, Y: 0})
	if err := app.cycle(); err != nil {
		t.Fatal(err)
	}

	if got := input.inputCursor; got != 1 {
		t.Fatalf("input cursor = %d, want 1", got)
	}
	if got := paintedCursorX(app); got != 0 {
		t.Fatalf("painted cursor X = %d, want 0", got)
	}
}

func TestInputClickUsesVisualGraphemeMidpoints(t *testing.T) {
	input := &instance{
		host: core.Host{
			Kind: core.HostInput,
			Data: core.InputData{Value: "a🙂b"},
		},
	}
	tests := []struct {
		x    int
		want int
	}{
		{x: 0, want: 0},
		{x: 1, want: 1},
		{x: 2, want: 2},
		{x: 3, want: 2},
		{x: 4, want: 3},
	}
	for _, test := range tests {
		input.inputCursor = -1
		app := &App{}

		app.positionInputCursor(input, test.x)

		if got := input.inputCursor; got != test.want {
			t.Errorf("cursor at visual X %d = %d, want %d", test.x, got, test.want)
		}
	}
}

func paintedCursorX(app *App) int {
	for x := 0; x < app.front.Width; x++ {
		if app.front.Cell(x, 0).Style.Attributes&core.AttributeReverse != 0 {
			return x
		}
	}
	return -1
}
