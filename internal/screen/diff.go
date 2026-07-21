package screen

import (
	"bytes"
	"fmt"

	"github.com/omnitui/omnitui/internal/core"
)

// Diff emits only changed cells and groups adjacent changes on each row.
func Diff(previous, next *Buffer, profile uint8) []byte {
	if next == nil {
		return nil
	}
	var output bytes.Buffer
	current := core.Style{}
	hasStyle := false
	for y := 0; y < next.Height; y++ {
		for x := 0; x < next.Width; {
			if !changed(previous, next, x, y) {
				x++
				continue
			}
			start := x
			for x < next.Width && changed(previous, next, x, y) {
				x++
			}
			output.WriteString(fmt.Sprintf("\x1b[%d;%dH", y+1, start+1))
			for column := start; column < x; column++ {
				cell := next.Cell(column, y)
				if !hasStyle || cell.Style != current {
					if hasStyle {
						output.WriteString("\x1b[0m")
					}
					if sequence := styleSequence(cell.Style, profile); sequence != "" {
						output.WriteString(sequence)
						hasStyle = true
					} else {
						hasStyle = false
					}
					current = cell.Style
				}
				if cell.Width != 0 {
					output.WriteString(cell.Grapheme)
				}
			}
		}
	}
	if hasStyle {
		output.WriteString("\x1b[0m")
	}
	return output.Bytes()
}

func changed(previous, next *Buffer, x, y int) bool {
	if previous == nil || previous.Width != next.Width || previous.Height != next.Height {
		return true
	}
	return previous.Cell(x, y) != next.Cell(x, y)
}
