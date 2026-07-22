package screen

import (
	"fmt"
	"math"

	"github.com/omnitui/omnitui/v2/internal/core"
)

func styleSequence(style core.Style, profile uint8) string {
	parts := []int{0}
	for bit, code := range []int{1, 2, 3, 4, 5, 7, 8, 9} {
		if style.Attributes&(core.AttributeMask(1)<<bit) != 0 {
			parts = append(parts, code)
		}
	}
	if core.ColorKindOf(style.Foreground) != core.ColorUnspecified {
		parts = append(parts, colorCode(style.Foreground, false, profile)...)
	}
	if core.ColorKindOf(style.Background) != core.ColorUnspecified {
		parts = append(parts, colorCode(style.Background, true, profile)...)
	}
	if len(parts) == 1 {
		return ""
	}
	return fmt.Sprintf("\x1b[%sm", joinInts(parts))
}

func colorCode(color core.Color, background bool, profile uint8) []int {
	kind := core.ColorKindOf(color)
	r, g, b, index := core.ColorValueOf(color)
	base := 30
	if background {
		base = 40
	}
	switch kind {
	case core.ColorDefault:
		if background {
			return []int{49}
		}
		return []int{39}
	case core.ColorANSI:
		value := int(index)
		if value >= 8 {
			base += 60
			value -= 8
		}
		return []int{base + value}
	case core.ColorIndexed:
		if profile == 1 {
			return ansiSGR(nearest16(int(index)), background)
		}
		return []int{base + 8, 5, int(index)}
	case core.ColorRGB:
		if profile == 3 {
			if background {
				return []int{48, 2, int(r), int(g), int(b)}
			}
			return []int{38, 2, int(r), int(g), int(b)}
		}
		if profile == 2 {
			if background {
				return []int{48, 5, nearest256(r, g, b)}
			}
			return []int{38, 5, nearest256(r, g, b)}
		}
		return ansiSGR(nearest16(nearest256(r, g, b)), background)
	default:
		return nil
	}
}

func joinInts(values []int) string {
	out := ""
	for index, value := range values {
		if index > 0 {
			out += ";"
		}
		out += fmt.Sprint(value)
	}
	return out
}

func ansiSGR(index int, background bool) []int {
	base := 30
	if background {
		base = 40
	}
	if index >= 8 {
		base += 60
		index -= 8
	}
	return []int{base + index}
}

func nearest16(index int) int {
	if index < 16 {
		return index
	}
	r, g, b := palette256(index)
	best, distance := 0, math.MaxFloat64
	for candidate := 0; candidate < 16; candidate++ {
		cr, cg, cb := ansiRGB(candidate)
		value := colorDistance(r, g, b, cr, cg, cb)
		if value < distance {
			best, distance = candidate, value
		}
	}
	return best
}

func nearest256(r, g, b uint8) int {
	best, distance := 0, math.MaxFloat64
	for candidate := 0; candidate < 256; candidate++ {
		cr, cg, cb := palette256(candidate)
		value := colorDistance(int(r), int(g), int(b), cr, cg, cb)
		if value < distance {
			best, distance = candidate, value
		}
	}
	return best
}

func colorDistance(r, g, b, cr, cg, cb int) float64 {
	dr, dg, db := r-cr, g-cg, b-cb
	return float64(dr*dr + dg*dg + db*db)
}

func ansiRGB(index int) (int, int, int) {
	values := [16][3]int{{0, 0, 0}, {205, 0, 0}, {0, 205, 0}, {205, 205, 0}, {0, 0, 238}, {205, 0, 205}, {0, 205, 205}, {229, 229, 229}, {127, 127, 127}, {255, 0, 0}, {0, 255, 0}, {255, 255, 0}, {92, 92, 255}, {255, 0, 255}, {0, 255, 255}, {255, 255, 255}}
	if index < 0 || index >= len(values) {
		return 0, 0, 0
	}
	return values[index][0], values[index][1], values[index][2]
}

func palette256(index int) (int, int, int) {
	if index < 16 {
		return ansiRGB(index)
	}
	if index < 232 {
		value := index - 16
		return 55 + (value/36)*40, 55 + ((value/6)%6)*40, 55 + (value%6)*40
	}
	gray := 8 + (index-232)*10
	return gray, gray, gray
}
