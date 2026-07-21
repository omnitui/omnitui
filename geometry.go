package omnitui

import (
	"fmt"

	"github.com/omnitui/omnitui/internal/core"
)

type Size = core.Size

func Auto() Size { return core.AutoSize() }

func Cells(value int) Size {
	if value < 0 {
		panic("omnitui: size cannot be negative")
	}
	return core.CellsSize(value)
}

func Fill() Size { return core.FillSize() }

func SizeIsAuto(value Size) bool    { return core.SizeModeOf(value) == core.SizeAuto }
func SizeIsFill(value Size) bool    { return core.SizeModeOf(value) == core.SizeFill }
func SizeCellsValue(value Size) int { return core.SizeValueOf(value) }

type Spacing = core.Spacing

func All(value int) Spacing {
	if value < 0 {
		panic("omnitui: spacing cannot be negative")
	}
	return Spacing{Top: value, Right: value, Bottom: value, Left: value}
}

func XY(horizontal, vertical int) Spacing {
	if horizontal < 0 || vertical < 0 {
		panic("omnitui: spacing cannot be negative")
	}
	return Spacing{Top: vertical, Right: horizontal, Bottom: vertical, Left: horizontal}
}

type Rect = core.Rect

func ValidateSpacing(value Spacing) error {
	if value.Top < 0 || value.Right < 0 || value.Bottom < 0 || value.Left < 0 {
		return fmt.Errorf("omnitui: spacing cannot be negative")
	}
	return nil
}
