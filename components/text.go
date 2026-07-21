package components

import (
	"github.com/omnitui/omnitui"
	"github.com/omnitui/omnitui/internal/core"
)

type TextWrap uint8

const (
	WrapNone TextWrap = iota
	WrapWord
	WrapGrapheme
)

type TextAlign uint8

const (
	TextAlignStart TextAlign = iota
	TextAlignCenter
	TextAlignEnd
)

type TruncateMode uint8

const (
	TruncateClip TruncateMode = iota
	TruncateEllipsis
)

type TextProps struct {
	Content  string
	Style    omnitui.Style
	Wrap     TextWrap
	Align    TextAlign
	MaxLines int
	Truncate TruncateMode
}

func Text(props TextProps) omnitui.Element {
	if props.MaxLines < 0 {
		panic("omnitui/components: MaxLines cannot be negative")
	}
	validateStyle(props.Style)
	return core.NewHost(core.HostText, core.TextData{Content: props.Content, Style: props.Style, Wrap: uint8(props.Wrap), Align: uint8(props.Align), MaxLines: props.MaxLines, Truncate: uint8(props.Truncate)}, nil)
}
