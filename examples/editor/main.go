package main

import (
	"context"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/omnitui/omnitui/v2"
	"github.com/omnitui/omnitui/v2/components"
	"github.com/rivo/uniseg"
)

var (
	surfaceStyle = omnitui.Style{
		Foreground: omnitui.RGB(220, 225, 235),
		Background: omnitui.RGB(18, 22, 30),
	}
	editorStyle = omnitui.Style{
		Foreground: omnitui.RGB(220, 225, 235),
		Background: omnitui.RGB(28, 34, 46),
	}
	editorFocusStyle = omnitui.Style{
		Background: omnitui.RGB(34, 42, 58),
	}
	keywordStyle = omnitui.Style{
		Foreground: omnitui.ANSI(omnitui.BrightMagenta),
		Attributes: omnitui.Bold,
	}
	commentStyle = omnitui.Style{
		Foreground: omnitui.ANSI(omnitui.BrightBlack),
		Attributes: omnitui.Italic,
	}
	stringStyle = omnitui.Style{
		Foreground: omnitui.ANSI(omnitui.BrightGreen),
	}
	numberStyle = omnitui.Style{
		Foreground: omnitui.ANSI(omnitui.BrightYellow),
	}
)

var goKeywords = map[string]struct{}{
	"break": {}, "case": {}, "chan": {}, "const": {}, "continue": {}, "default": {},
	"defer": {}, "else": {}, "fallthrough": {}, "for": {}, "func": {}, "go": {},
	"goto": {}, "if": {}, "import": {}, "interface": {}, "map": {}, "package": {},
	"range": {}, "return": {}, "select": {}, "struct": {}, "switch": {}, "type": {},
	"var": {},
}

const initialSource = `package main

import "fmt"

func main() {
    // Edit this code with the keyboard or mouse.
    answer := 42
    fmt.Println("answer:", answer)
}`

type editorState struct{ Source string }
type editorExample struct{}

func (editorExample) InitialState(string) editorState {
	return editorState{Source: initialSource}
}

func (editorExample) Render(ctx omnitui.Context, _ string, state editorState, _ omnitui.Children) omnitui.Element {
	lineCount := strings.Count(state.Source, "\n") + 1
	return components.Box(
		components.BoxProps{
			Direction: components.Vertical,
			Padding:   omnitui.All(1),
			Gap:       1,
			Border:    components.BorderRounded,
			Label:     "Go editor",
			Style:     surfaceStyle,
		},
		components.Text(components.TextProps{
			Content: "Type to edit • arrows/Home/End navigate • wheel or scrollbar drag scrolls",
			Style:   omnitui.Style{Foreground: omnitui.ANSI(omnitui.BrightCyan)},
		}),
		components.Editor(components.EditorProps{
			Value:       state.Source,
			Width:       omnitui.Fill(),
			Height:      omnitui.Cells(12),
			TabWidth:    4,
			Scrollbar:   components.ScrollbarAuto,
			Style:       editorStyle,
			FocusStyle:  editorFocusStyle,
			Highlighter: highlightGo,
			OnChange: func(event omnitui.ValueChangeEvent) omnitui.EventResult {
				omnitui.UpdateState(ctx, func(current editorState) editorState {
					current.Source = event.Value
					return current
				})
				return omnitui.Consume
			},
		}),
		components.Text(components.TextProps{
			Content: fmt.Sprintf("%d lines", lineCount),
			Style:   omnitui.Style{Foreground: omnitui.ANSI(omnitui.BrightBlack)},
		}),
	)
}

func highlightGo(line string, _ int) []components.HighlightSpan {
	graphemes := splitGraphemes(line)
	var spans []components.HighlightSpan
	for index := 0; index < len(graphemes); {
		if graphemes[index] == "/" && index+1 < len(graphemes) && graphemes[index+1] == "/" {
			spans = append(spans, components.HighlightSpan{Start: index, End: len(graphemes), Style: commentStyle})
			break
		}
		if graphemes[index] == "\"" || graphemes[index] == "`" {
			quote := graphemes[index]
			end := index + 1
			for end < len(graphemes) {
				if graphemes[end] == "\\" && quote == "\"" && end+1 < len(graphemes) {
					end += 2
					continue
				}
				end++
				if graphemes[end-1] == quote {
					break
				}
			}
			spans = append(spans, components.HighlightSpan{Start: index, End: end, Style: stringStyle})
			index = end
			continue
		}
		if isIdentifierStart(graphemes[index]) {
			end := index + 1
			for end < len(graphemes) && isIdentifierPart(graphemes[end]) {
				end++
			}
			if _, keyword := goKeywords[strings.Join(graphemes[index:end], "")]; keyword {
				spans = append(spans, components.HighlightSpan{Start: index, End: end, Style: keywordStyle})
			}
			index = end
			continue
		}
		if isDigit(graphemes[index]) {
			end := index + 1
			for end < len(graphemes) && isDigit(graphemes[end]) {
				end++
			}
			spans = append(spans, components.HighlightSpan{Start: index, End: end, Style: numberStyle})
			index = end
			continue
		}
		index++
	}
	return spans
}

func splitGraphemes(value string) []string {
	iterator := uniseg.NewGraphemes(value)
	var result []string
	for iterator.Next() {
		result = append(result, iterator.Str())
	}
	return result
}

func isIdentifierStart(grapheme string) bool {
	value, _ := utf8.DecodeRuneInString(grapheme)
	return value == '_' || unicode.IsLetter(value)
}

func isIdentifierPart(grapheme string) bool {
	value, _ := utf8.DecodeRuneInString(grapheme)
	return value == '_' || unicode.IsLetter(value) || unicode.IsDigit(value)
}

func isDigit(grapheme string) bool {
	value, _ := utf8.DecodeRuneInString(grapheme)
	return unicode.IsDigit(value)
}

func main() {
	typeValue := omnitui.Define("EditorExample", editorExample{})
	app := omnitui.New(omnitui.Create(typeValue, "editor"), omnitui.Options{})
	if err := app.Run(context.Background()); err != nil {
		panic(err)
	}
}
