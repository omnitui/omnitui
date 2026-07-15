package text

import "testing"

func TestGraphemeWidthAndWrapping(t *testing.T) {
	graphemes := Graphemes("e\u0301🙂")
	if len(graphemes) != 2 {
		t.Fatalf("graphemes = %#v", graphemes)
	}
	if Width(graphemes[0]) != 1 || Width(graphemes[1]) != 2 {
		t.Fatalf("widths = %d,%d", Width(graphemes[0]), Width(graphemes[1]))
	}
	lines := Wrap("um dois três", 6, true)
	if len(lines) != 3 {
		t.Fatalf("wrapped lines = %#v", lines)
	}
	if got := Truncate("abcdef", 4, true); got != "abc…" {
		t.Fatalf("truncate = %q", got)
	}
}
