package screen

import (
	"bytes"
	"testing"

	"github.com/omnitui/omnitui/v2/internal/core"
)

func TestDiffSkipsUnchangedCells(t *testing.T) {
	old := NewBuffer(4, 1, core.Style{})
	next := old.Clone()
	next.Set(2, 0, "x", core.Style{})
	output := Diff(old, next, 1)
	if !bytes.Contains(output, []byte("\x1b[1;3Hx")) {
		t.Fatalf("diff = %q", output)
	}
	if bytes.Contains(output, []byte("\x1b[1;1H")) || bytes.Contains(output, []byte("\x1b[1;2H")) {
		t.Fatalf("diff rewrote unchanged cells: %q", output)
	}
}
