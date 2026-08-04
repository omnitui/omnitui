package components

import (
	"fmt"
	"strings"

	"github.com/omnitui/omnitui/v2"
	"github.com/omnitui/omnitui/v2/internal/core"
	uitext "github.com/omnitui/omnitui/v2/internal/text"
)

type HighlightSpan struct {
	Start int
	End   int
	Style omnitui.Style
}

type SyntaxHighlighter func(line string, lineIndex int) []HighlightSpan

type EditorProps struct {
	Value       string
	Placeholder string
	Width       omnitui.Size
	Height      omnitui.Size
	Disabled    bool
	ReadOnly    bool
	TabWidth    int
	Scrollbar   ScrollbarMode
	Style       omnitui.Style
	FocusStyle  omnitui.Style
	Highlighter SyntaxHighlighter
	Focus       omnitui.FocusHandle
	OnChange    omnitui.EventHandler[omnitui.ValueChangeEvent]
	OnKey       omnitui.EventHandler[omnitui.KeyEvent]
	OnTextInput omnitui.EventHandler[omnitui.TextInputEvent]
	OnPaste     omnitui.EventHandler[omnitui.PasteEvent]
	OnFocus     omnitui.EventHandler[omnitui.FocusEvent]
	OnBlur      omnitui.EventHandler[omnitui.BlurEvent]
	OnMouse     omnitui.EventHandler[omnitui.MouseEvent]
	OnWheel     omnitui.EventHandler[omnitui.WheelEvent]
}

type editorComponent struct{}

func (editorComponent) InitialState(EditorProps) struct{} { return struct{}{} }
func (editorComponent) Render(_ omnitui.Context, props EditorProps, _ struct{}, _ omnitui.Children) omnitui.Element {
	return editorHost(props)
}

var editorType = omnitui.Define[EditorProps, struct{}]("Editor", editorComponent{})

func Editor(props EditorProps) omnitui.Element {
	if props.TabWidth < 0 {
		panic("omnitui/components: editor TabWidth cannot be negative")
	}
	if props.TabWidth == 0 {
		props.TabWidth = 4
	}
	validateStyle(props.Style)
	validateStyle(props.FocusStyle)
	return omnitui.Create(editorType, props)
}

func editorHost(props EditorProps) omnitui.Element {
	return core.NewHost(core.HostEditor, core.EditorData{
		Value: props.Value, Placeholder: props.Placeholder,
		Width: props.Width, Height: props.Height, Disabled: props.Disabled, ReadOnly: props.ReadOnly,
		TabWidth: props.TabWidth, Scrollbar: uint8(props.Scrollbar), Style: props.Style, FocusStyle: props.FocusStyle,
		Highlights: editorHighlights(props), Focus: props.Focus,
		Handlers: handlers(map[string]any{
			"change": props.OnChange, "key": props.OnKey, "text": props.OnTextInput,
			"paste": props.OnPaste, "focus": props.OnFocus, "blur": props.OnBlur,
			"mouse": props.OnMouse, "wheel": props.OnWheel,
		}),
	}, nil)
}

func editorHighlights(props EditorProps) [][]core.HighlightSpan {
	if props.Highlighter == nil || props.Value == "" {
		return nil
	}
	value := strings.ReplaceAll(props.Value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	lines := strings.Split(value, "\n")
	result := make([][]core.HighlightSpan, len(lines))
	for lineIndex, line := range lines {
		length := len(uitext.Graphemes(line))
		spans := props.Highlighter(line, lineIndex)
		result[lineIndex] = make([]core.HighlightSpan, len(spans))
		for spanIndex, span := range spans {
			if span.Start < 0 || span.End < span.Start || span.End > length {
				panic(fmt.Sprintf("omnitui/components: invalid editor highlight span [%d,%d) on line %d with %d graphemes", span.Start, span.End, lineIndex, length))
			}
			validateStyle(span.Style)
			result[lineIndex][spanIndex] = core.HighlightSpan{Start: span.Start, End: span.End, Style: span.Style}
		}
	}
	return result
}
