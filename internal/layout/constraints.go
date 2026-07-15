package layout

import "github.com/viniciusfonseca/omnitui/internal/core"

type Constraints struct {
	MinWidth, MaxWidth   int
	MinHeight, MaxHeight int
}

func Unbounded() Constraints { return Constraints{MaxWidth: 1 << 30, MaxHeight: 1 << 30} }

func (c Constraints) Clamp(width, height int) core.Rect {
	if width < c.MinWidth {
		width = c.MinWidth
	}
	if width > c.MaxWidth {
		width = c.MaxWidth
	}
	if height < c.MinHeight {
		height = c.MinHeight
	}
	if height > c.MaxHeight {
		height = c.MaxHeight
	}
	return core.Rect{Width: width, Height: height}
}
