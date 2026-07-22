package ansi

import (
	"testing"

	"github.com/omnitui/omnitui/v2/internal/backend"
)

func TestParserRecognizesKeyboardSequences(t *testing.T) {
	parser := &Parser{}
	events := parser.Feed([]byte("a\r\t\x1b[A\x1b[Z\x1b[15~"))
	keys := make([]backend.KeyInput, 0, len(events))
	for _, event := range events {
		if key, ok := event.Value.(backend.KeyInput); ok {
			keys = append(keys, key)
		}
	}
	if len(keys) != 6 {
		t.Fatalf("parsed %d keys, want 6", len(keys))
	}
	if keys[0].Rune != 'a' || keys[1].Key != 1 || keys[2].Key != 3 || keys[3].Key != 8 || keys[4].Key != 4 || keys[5].Key != 16 {
		t.Fatalf("unexpected keys: %#v", keys)
	}
}

func TestParserBracketedPaste(t *testing.T) {
	parser := &Parser{}
	events := parser.Feed([]byte("\x1b[200~a\n\x1b[201~"))
	if len(events) != 1 {
		t.Fatalf("events = %#v", events)
	}
	paste, ok := events[0].Value.(backend.PasteInput)
	if !ok || paste.Text != "a\n" {
		t.Fatalf("paste = %#v", events[0].Value)
	}
}
