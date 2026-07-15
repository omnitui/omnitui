package core

type ColorKind uint8

const (
	ColorUnspecified ColorKind = iota
	ColorDefault
	ColorANSI
	ColorIndexed
	ColorRGB
)

type ANSIColor uint8

type Color struct {
	kind    ColorKind
	value   uint8
	r, g, b uint8
}

func DefaultColorValue() Color                        { return Color{kind: ColorDefault} }
func ANSIColorValue(value ANSIColor) Color            { return Color{kind: ColorANSI, value: uint8(value)} }
func IndexedColor(value uint8) Color                  { return Color{kind: ColorIndexed, value: value} }
func RGBColor(r, g, b uint8) Color                    { return Color{kind: ColorRGB, r: r, g: g, b: b} }
func ColorKindOf(value Color) ColorKind               { return value.kind }
func ColorValueOf(value Color) (r, g, b, index uint8) { return value.r, value.g, value.b, value.value }

type AttributeMask uint16

const (
	AttributeBold AttributeMask = 1 << iota
	AttributeDim
	AttributeItalic
	AttributeUnderline
	AttributeBlink
	AttributeReverse
	AttributeHidden
	AttributeStrikethrough
)

type Style struct {
	Foreground      Color
	Background      Color
	Attributes      AttributeMask
	ClearAttributes AttributeMask
}

func ResolveStyle(parent, own Style) (Style, bool) {
	if own.Attributes&own.ClearAttributes != 0 {
		return Style{}, false
	}
	result := parent
	if own.Foreground.kind != ColorUnspecified {
		result.Foreground = own.Foreground
	}
	if own.Background.kind != ColorUnspecified {
		result.Background = own.Background
	}
	result.Attributes &^= own.ClearAttributes
	result.Attributes |= own.Attributes
	return result, true
}
