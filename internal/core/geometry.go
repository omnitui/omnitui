package core

type Size struct {
	mode  uint8
	value int
}

const (
	SizeAuto uint8 = iota
	SizeCells
	SizeFill
)

func AutoSize() Size { return Size{} }

func CellsSize(value int) Size { return Size{mode: SizeCells, value: value} }

func FillSize() Size { return Size{mode: SizeFill} }

func SizeModeOf(value Size) uint8 { return value.mode }
func SizeValueOf(value Size) int  { return value.value }

type Spacing struct {
	Top, Right, Bottom, Left int
}

type Rect struct {
	X, Y, Width, Height int
}

func (r Rect) Contains(x, y int) bool {
	return x >= r.X && y >= r.Y && x < r.X+r.Width && y < r.Y+r.Height
}

func (r Rect) Empty() bool { return r.Width <= 0 || r.Height <= 0 }

func IntersectRect(a, b Rect) Rect {
	x := max(a.X, b.X)
	y := max(a.Y, b.Y)
	right := min(a.X+a.Width, b.X+b.Width)
	bottom := min(a.Y+a.Height, b.Y+b.Height)
	if right <= x || bottom <= y {
		return Rect{X: x, Y: y}
	}
	return Rect{X: x, Y: y, Width: right - x, Height: bottom - y}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
