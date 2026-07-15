package omnitui

import (
	"fmt"

	"github.com/viniciusfonseca/omnitui/internal/core"
)

type Color = core.Color
type ANSIColor = core.ANSIColor

const (
	Black ANSIColor = iota
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
	BrightBlack
	BrightRed
	BrightGreen
	BrightYellow
	BrightBlue
	BrightMagenta
	BrightCyan
	BrightWhite
)

func DefaultColor() Color              { return core.DefaultColorValue() }
func ANSI(color ANSIColor) Color       { return core.ANSIColorValue(color) }
func Indexed(index uint8) Color        { return core.IndexedColor(index) }
func RGB(red, green, blue uint8) Color { return core.RGBColor(red, green, blue) }

type AttributeMask = core.AttributeMask

const (
	Bold          AttributeMask = core.AttributeBold
	Dim                         = core.AttributeDim
	Italic                      = core.AttributeItalic
	Underline                   = core.AttributeUnderline
	Blink                       = core.AttributeBlink
	Reverse                     = core.AttributeReverse
	Hidden                      = core.AttributeHidden
	Strikethrough               = core.AttributeStrikethrough
)

type Style = core.Style

func ResolveStyle(parent, own Style) (Style, error) {
	result, ok := core.ResolveStyle(parent, own)
	if !ok {
		return Style{}, fmt.Errorf("omnitui: style attribute is both set and cleared")
	}
	return result, nil
}

func ValidateStyle(style Style) error {
	if style.Attributes&style.ClearAttributes != 0 {
		return fmt.Errorf("omnitui: style attribute is both set and cleared")
	}
	return nil
}
