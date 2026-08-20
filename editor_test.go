package omnitui

import (
	"testing"

	"github.com/omnitui/omnitui/v2/internal/core"
)

func TestEditorInsertsAndRemovesLineBreaks(t *testing.T) {
	var proposed ValueChangeEvent
	editor := editorInstance("one\ntwo", func(event ValueChangeEvent) EventResult {
		proposed = event
		return Consume
	})
	app := &App{}
	editor.app = app
	editor.editorCursor = 3

	if !app.editorKey(editor, KeyEvent{Key: KeyEnter}) {
		t.Fatal("Enter was not handled")
	}
	if proposed.Value != "one\n\ntwo" || proposed.Source != ChangeKeyboard {
		t.Fatalf("inserted value = %q source=%d", proposed.Value, proposed.Source)
	}
	if editor.editorCursor != 4 {
		t.Fatalf("cursor after Enter = %d, want 4", editor.editorCursor)
	}

	editor.host.Data = core.EditorData{
		Value: "one\ntwo",
		Handlers: core.Handlers{"change": EventHandler[ValueChangeEvent](func(event ValueChangeEvent) EventResult {
			proposed = event
			return Consume
		})},
	}
	editor.editorCursor = 4
	app.editorKey(editor, KeyEvent{Key: KeyBackspace})
	if proposed.Value != "onetwo" || editor.editorCursor != 3 {
		t.Fatalf("Backspace proposed %q with cursor %d, want onetwo at 3", proposed.Value, editor.editorCursor)
	}
}

func TestEditorVerticalNavigationPreservesVisualColumn(t *testing.T) {
	editor := editorInstance("abcd\nx\nabcd", nil)
	editor.rect = Rect{Width: 8, Height: 2}
	editor.editorCursor = 4
	app := &App{}
	editor.app = app

	app.editorKey(editor, KeyEvent{Key: KeyDown})
	if editor.editorCursor != 6 {
		t.Fatalf("cursor on short line = %d, want 6", editor.editorCursor)
	}
	app.editorKey(editor, KeyEvent{Key: KeyDown})
	if editor.editorCursor != 11 {
		t.Fatalf("cursor after second Down = %d, want 11", editor.editorCursor)
	}
}

func TestEditorKeepsCursorVisibleAndAllowsManualScroll(t *testing.T) {
	root := core.NewHost(core.HostEditor, core.EditorData{
		Value:  "one\ntwo\nthree\nfour",
		Width:  core.CellsSize(5),
		Height: core.CellsSize(2),
	}, nil)
	app := New(root, Options{})
	app.width, app.height = 5, 2
	if err := app.render(); err != nil {
		t.Fatal(err)
	}
	editor := app.rootInstance
	editor.editorCursor = len(newEditorDocument("one\ntwo\nthree\nfour").graphemes)
	app.invalidated = true
	if err := app.cycle(); err != nil {
		t.Fatal(err)
	}
	if editor.editorRowOffset != 2 {
		t.Fatalf("automatic row offset = %d, want 2", editor.editorRowOffset)
	}

	app.scrollEditor(editor, 0, -1)
	if err := app.cycle(); err != nil {
		t.Fatal(err)
	}
	if editor.editorRowOffset != 1 {
		t.Fatalf("manual row offset = %d, want 1", editor.editorRowOffset)
	}
}

func TestEditorPaintsSyntaxHighlight(t *testing.T) {
	highlight := Style{Foreground: ANSI(Red), Attributes: Bold}
	root := core.NewHost(core.HostEditor, core.EditorData{
		Value:  "func main",
		Width:  core.CellsSize(9),
		Height: core.CellsSize(1),
		Highlights: [][]core.HighlightSpan{{
			{Start: 0, End: 4, Style: highlight},
		}},
	}, nil)
	app := New(root, Options{})
	app.width, app.height = 9, 1
	if err := app.render(); err != nil {
		t.Fatal(err)
	}

	if got := app.front.Cell(0, 0).Style; got != highlight {
		t.Fatalf("highlighted style = %#v, want %#v", got, highlight)
	}
	if got := app.front.Cell(5, 0).Style; got == highlight {
		t.Fatalf("unhighlighted cell unexpectedly has highlight style %#v", got)
	}
}

func TestEditorMouseUsesVisualColumns(t *testing.T) {
	editor := editorInstance("a🙂b", nil)
	editor.editorCursor = 3
	app := &App{}
	editor.app = app

	app.positionEditorCursor(editor, 1, 0)
	if editor.editorCursor != 1 {
		t.Fatalf("cursor on first emoji cell = %d, want 1", editor.editorCursor)
	}
	app.positionEditorCursor(editor, 2, 0)
	if editor.editorCursor != 2 {
		t.Fatalf("cursor on second emoji cell = %d, want 2", editor.editorCursor)
	}
}

func TestReadOnlyEditorKeepsFocusAndCursorWithoutChangingValue(t *testing.T) {
	changes := 0
	root := core.NewHost(core.HostEditor, core.EditorData{
		Value: "one\ntwo", Width: core.CellsSize(5), Height: core.CellsSize(2), ReadOnly: true,
		Handlers: core.Handlers{"change": EventHandler[ValueChangeEvent](func(ValueChangeEvent) EventResult {
			changes++
			return Consume
		})},
	}, nil)
	app := New(root, Options{})
	app.width, app.height = 5, 2
	if err := app.render(); err != nil {
		t.Fatal(err)
	}
	editor := app.rootInstance

	app.dispatchMouse(MouseEvent{Action: MouseDown, Button: MouseButtonLeft, X: 1, Y: 1})
	app.dispatchMouse(MouseEvent{Action: MouseUp, Button: MouseButtonLeft, X: 1, Y: 1})
	if err := app.cycle(); err != nil {
		t.Fatal(err)
	}
	if app.focused != editor || editor.editorCursor != 5 {
		t.Fatalf("readonly click focused=%v cursor=%d, want focused cursor 5", app.focused == editor, editor.editorCursor)
	}
	if app.front.Cell(1, 1).Style.Attributes&core.AttributeReverse == 0 {
		t.Fatal("readonly cursor was not painted at the clicked position")
	}

	app.dispatchKey(KeyEvent{Key: KeyRune, Rune: 'x'})
	app.dispatchKey(KeyEvent{Key: KeyEnter})
	app.dispatchKey(KeyEvent{Key: KeyBackspace})
	app.dispatchKey(KeyEvent{Key: KeyDelete})
	app.dispatchPaste(PasteEvent{Text: "pasted"})
	if changes != 0 || editor.editorCursor != 5 {
		t.Fatalf("readonly mutation emitted %d changes and moved cursor to %d", changes, editor.editorCursor)
	}

	app.dispatchKey(KeyEvent{Key: KeyLeft})
	app.dispatchKey(KeyEvent{Key: KeyEnd})
	if editor.editorCursor != 7 {
		t.Fatalf("readonly keyboard navigation cursor = %d, want 7", editor.editorCursor)
	}
}

func TestEditorPaintsScrollbarModesWithoutCoveringContent(t *testing.T) {
	tests := []struct {
		name       string
		mode       uint8
		value      string
		height     int
		wantLast   string
		wantBefore string
	}{
		{name: "auto overflow", mode: 0, value: "abcde\ntwo\nthree\nfour", height: 2, wantLast: "█", wantBefore: "d"},
		{name: "always", mode: 1, value: "abcde", height: 2, wantLast: "│", wantBefore: "d"},
		{name: "hidden", mode: 2, value: "abcde\ntwo\nthree", height: 2, wantLast: "e", wantBefore: "d"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := core.NewHost(core.HostEditor, core.EditorData{
				Value: test.value, Width: core.CellsSize(5), Height: core.CellsSize(test.height), Scrollbar: test.mode,
			}, nil)
			app := New(root, Options{})
			app.width, app.height = 5, test.height
			if err := app.render(); err != nil {
				t.Fatal(err)
			}
			if got := app.front.Cell(4, 0).Grapheme; got != test.wantLast {
				t.Fatalf("last column = %q, want %q", got, test.wantLast)
			}
			if got := app.front.Cell(3, 0).Grapheme; got != test.wantBefore {
				t.Fatalf("content before scrollbar = %q, want %q", got, test.wantBefore)
			}
		})
	}
}

func TestEditorKeepsEndCursorVisibleBesideScrollbar(t *testing.T) {
	root := core.NewHost(core.HostEditor, core.EditorData{
		Value: "abcd\nx", Width: core.CellsSize(4), Height: core.CellsSize(1), Scrollbar: 0,
	}, nil)
	app := New(root, Options{})
	app.width, app.height = 4, 1
	if err := app.render(); err != nil {
		t.Fatal(err)
	}
	editor := app.rootInstance
	app.setFocus(editor, ProgrammaticFocus)
	editor.editorCursor = 4
	app.invalidated = true
	if err := app.cycle(); err != nil {
		t.Fatal(err)
	}
	if editor.editorColumnOffset != 2 {
		t.Fatalf("column offset = %d, want 2", editor.editorColumnOffset)
	}
	if app.front.Cell(2, 0).Style.Attributes&core.AttributeReverse == 0 {
		t.Fatal("end cursor is not visible in the content viewport")
	}
	if got := app.front.Cell(3, 0).Grapheme; got != "█" {
		t.Fatalf("scrollbar thumb = %q, want █", got)
	}
}

func TestEditorScrollbarThumbFollowsVerticalOffset(t *testing.T) {
	root := core.NewHost(core.HostEditor, core.EditorData{
		Value: "one\ntwo\nthree\nfour", Width: core.CellsSize(5), Height: core.CellsSize(2), Scrollbar: 0,
	}, nil)
	app := New(root, Options{})
	app.width, app.height = 5, 2
	if err := app.render(); err != nil {
		t.Fatal(err)
	}
	app.scrollEditor(app.rootInstance, 0, 2)
	if err := app.cycle(); err != nil {
		t.Fatal(err)
	}
	if got := app.front.Cell(4, 0).Grapheme; got != "│" {
		t.Fatalf("top track = %q, want │", got)
	}
	if got := app.front.Cell(4, 1).Grapheme; got != "█" {
		t.Fatalf("bottom thumb = %q, want █", got)
	}
}

func TestEditorScrollbarClickDoesNotMoveCursor(t *testing.T) {
	editor := editorInstance("one\ntwo\nthree", nil)
	editor.rect = Rect{Width: 5, Height: 2}
	editor.editorCursor = 1
	app := &App{}
	editor.app = app

	app.positionEditorCursor(editor, 4, 0)
	if editor.editorCursor != 1 {
		t.Fatalf("cursor after scrollbar click = %d, want 1", editor.editorCursor)
	}
}

func TestEditorScrollbarDragScrollsWithoutMovingCursor(t *testing.T) {
	root := core.NewHost(core.HostEditor, core.EditorData{
		Value: "0\n1\n2\n3\n4\n5\n6\n7\n8\n9", Width: core.CellsSize(5), Height: core.CellsSize(4), Scrollbar: 0,
	}, nil)
	app := New(root, Options{})
	app.width, app.height = 5, 4
	if err := app.render(); err != nil {
		t.Fatal(err)
	}
	editor := app.rootInstance
	editor.editorCursor = 1

	app.dispatchMouse(MouseEvent{Action: MouseDown, Button: MouseButtonLeft, X: 4, Y: 0})
	app.dispatchMouse(MouseEvent{Action: MouseMove, Buttons: MouseLeftPressed, X: 8, Y: 20})
	app.dispatchMouse(MouseEvent{Action: MouseUp, Button: MouseButtonLeft, X: 8, Y: 20})

	if editor.editorRowOffset != 6 {
		t.Fatalf("row offset after scrollbar drag = %d, want 6", editor.editorRowOffset)
	}
	if editor.editorCursor != 1 {
		t.Fatalf("cursor after scrollbar drag = %d, want 1", editor.editorCursor)
	}
	if app.capture != nil {
		t.Fatal("editor kept mouse capture after scrollbar release")
	}
}

func TestEditorScrollbarDragKeepsTheThumbGrabPoint(t *testing.T) {
	root := core.NewHost(core.HostEditor, core.EditorData{
		Value: "0\n1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11", Width: core.CellsSize(5), Height: core.CellsSize(6), Scrollbar: 0,
	}, nil)
	app := New(root, Options{})
	app.width, app.height = 5, 6
	if err := app.render(); err != nil {
		t.Fatal(err)
	}

	app.dispatchMouse(MouseEvent{Action: MouseDown, Button: MouseButtonLeft, X: 4, Y: 2})
	app.dispatchMouse(MouseEvent{Action: MouseMove, Buttons: MouseLeftPressed, X: 4, Y: 3})
	app.dispatchMouse(MouseEvent{Action: MouseUp, Button: MouseButtonLeft, X: 4, Y: 3})

	if got := app.rootInstance.editorRowOffset; got != 2 {
		t.Fatalf("row offset after dragging from the thumb bottom = %d, want 2", got)
	}
}

func editorInstance(value string, onChange EventHandler[ValueChangeEvent]) *instance {
	handlers := core.Handlers{}
	if onChange != nil {
		handlers["change"] = onChange
	}
	element := core.NewHost(core.HostEditor, core.EditorData{Value: value, TabWidth: 4, Handlers: handlers}, nil)
	host, _ := core.HostOf(element)
	return &instance{
		element: element,
		host:    host,
	}
}
