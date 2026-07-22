package screen

import (
	"github.com/omnitui/omnitui/v2/internal/core"
	"github.com/omnitui/omnitui/v2/internal/text"
)

type Buffer struct {
	Width, Height int
	cells         []Cell
}

func NewBuffer(width, height int, style core.Style) *Buffer {
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}
	b := &Buffer{Width: width, Height: height, cells: make([]Cell, width*height)}
	for index := range b.cells {
		b.cells[index] = blank(style)
	}
	return b
}

func (b *Buffer) Cell(x, y int) Cell {
	if x < 0 || y < 0 || x >= b.Width || y >= b.Height {
		return Cell{}
	}
	return b.cells[y*b.Width+x]
}

func (b *Buffer) Set(x, y int, grapheme string, style core.Style) {
	if x < 0 || y < 0 || x >= b.Width || y >= b.Height {
		return
	}
	width := text.Width(grapheme)
	if width <= 0 {
		return
	}
	if x+width > b.Width {
		return
	}
	b.cells[y*b.Width+x] = Cell{Grapheme: grapheme, Width: width, Style: style}
	for offset := 1; offset < width; offset++ {
		b.cells[y*b.Width+x+offset] = Cell{Width: 0, Style: style}
	}
}

func (b *Buffer) Fill(rect core.Rect, grapheme string, style core.Style) {
	for y := rect.Y; y < rect.Y+rect.Height; y++ {
		for x := rect.X; x < rect.X+rect.Width; x++ {
			b.Set(x, y, grapheme, style)
		}
	}
}

func (b *Buffer) Clone() *Buffer {
	clone := &Buffer{Width: b.Width, Height: b.Height, cells: make([]Cell, len(b.cells))}
	copy(clone.cells, b.cells)
	return clone
}
