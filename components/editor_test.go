package components

import (
	"testing"

	omnitui "github.com/omnitui/omnitui/v2"
	"github.com/omnitui/omnitui/v2/internal/core"
)

func TestEditorBuildsValidatedHighlights(t *testing.T) {
	element := editorHost(EditorProps{
		Value:    "func main",
		TabWidth: 4,
		Highlighter: func(line string, lineIndex int) []HighlightSpan {
			return []HighlightSpan{{Start: 0, End: 4, Style: omnitui.Style{Attributes: omnitui.Bold}}}
		},
	})
	host, ok := core.HostOf(element)
	if !ok {
		t.Fatal("Editor did not create a host")
	}
	data := host.Data.(core.EditorData)
	if len(data.Highlights) != 1 || len(data.Highlights[0]) != 1 {
		t.Fatalf("highlights = %#v", data.Highlights)
	}
	if data.Highlights[0][0].End != 4 {
		t.Fatalf("highlight end = %d, want 4", data.Highlights[0][0].End)
	}
}

func TestEditorRejectsInvalidConfiguration(t *testing.T) {
	mustPanic(t, func() { Editor(EditorProps{TabWidth: -1}) })
	mustPanic(t, func() {
		editorHost(EditorProps{
			Value:    "Go",
			TabWidth: 4,
			Highlighter: func(string, int) []HighlightSpan {
				return []HighlightSpan{{Start: 0, End: 3}}
			},
		})
	})
}

func TestEditorUsesDefaultTabWidth(t *testing.T) {
	element := Editor(EditorProps{Scrollbar: ScrollbarAlways})
	props := core.PropsOf(element).(EditorProps)
	if props.TabWidth != 4 {
		t.Fatalf("TabWidth = %d, want 4", props.TabWidth)
	}
	host, _ := core.HostOf(editorHost(props))
	if scrollbar := host.Data.(core.EditorData).Scrollbar; scrollbar != uint8(ScrollbarAlways) {
		t.Fatalf("Scrollbar = %d, want ScrollbarAlways", scrollbar)
	}
}
