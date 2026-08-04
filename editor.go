package omnitui

import (
	"strings"

	"github.com/omnitui/omnitui/v2/internal/core"
	"github.com/omnitui/omnitui/v2/internal/screen"
	uitext "github.com/omnitui/omnitui/v2/internal/text"
)

type editorLine struct {
	start     int
	graphemes []string
}

type editorDocument struct {
	graphemes []string
	lines     []editorLine
}

func newEditorDocument(value string) editorDocument {
	document := editorDocument{graphemes: uitext.Graphemes(value), lines: []editorLine{{}}}
	for index, grapheme := range document.graphemes {
		if grapheme == "\n" || grapheme == "\r" || grapheme == "\r\n" {
			document.lines = append(document.lines, editorLine{start: index + 1})
			continue
		}
		last := len(document.lines) - 1
		document.lines[last].graphemes = append(document.lines[last].graphemes, grapheme)
	}
	return document
}

func (document editorDocument) position(cursor int) (int, int) {
	cursor = editorClamp(cursor, 0, len(document.graphemes))
	row := 0
	for index := 1; index < len(document.lines); index++ {
		if document.lines[index].start > cursor {
			break
		}
		row = index
	}
	column := cursor - document.lines[row].start
	column = editorClamp(column, 0, len(document.lines[row].graphemes))
	return row, column
}

func (document editorDocument) cursor(row, column int) int {
	row = editorClamp(row, 0, len(document.lines)-1)
	column = editorClamp(column, 0, len(document.lines[row].graphemes))
	return document.lines[row].start + column
}

func editorTabWidth(value int) int {
	if value <= 0 {
		return 4
	}
	return value
}

func editorGraphemeWidth(grapheme string, visualColumn, tabWidth int) int {
	if grapheme == "\t" {
		return tabWidth - visualColumn%tabWidth
	}
	return maxInt(uitext.Width(grapheme), 1)
}

func editorVisualColumn(graphemes []string, column, tabWidth int) int {
	column = editorClamp(column, 0, len(graphemes))
	visual := 0
	for _, grapheme := range graphemes[:column] {
		visual += editorGraphemeWidth(grapheme, visual, tabWidth)
	}
	return visual
}

func editorColumnAtVisual(graphemes []string, target, tabWidth int) int {
	if target <= 0 {
		return 0
	}
	visual := 0
	for index, grapheme := range graphemes {
		width := editorGraphemeWidth(grapheme, visual, tabWidth)
		if target < visual+(width+1)/2 {
			return index
		}
		visual += width
	}
	return len(graphemes)
}

func editorLineWidth(line editorLine, tabWidth int) int {
	return editorVisualColumn(line.graphemes, len(line.graphemes), tabWidth)
}

func measureEditor(data core.EditorData, maxWidth, maxHeight int) (int, int) {
	document := newEditorDocument(data.Value)
	tabWidth := editorTabWidth(data.TabWidth)
	width := 1
	for _, line := range document.lines {
		width = maxInt(width, editorLineWidth(line, tabWidth))
	}
	if data.Value == "" {
		width = maxInt(width, uitext.Width(data.Placeholder))
	}
	height := maxInt(len(document.lines), 1)
	if core.SizeModeOf(data.Height) == core.SizeCells {
		height = core.SizeValueOf(data.Height)
	} else if core.SizeModeOf(data.Height) == core.SizeFill {
		height = maxHeight
	}
	height = minInt(height, maxHeight)
	if core.SizeModeOf(data.Width) == core.SizeCells {
		width = core.SizeValueOf(data.Width)
	} else if core.SizeModeOf(data.Width) == core.SizeFill {
		width = maxWidth
	} else if editorScrollbarVisible(data, len(document.lines), height) {
		width++
	}
	return minInt(width, maxWidth), height
}

func arrangeEditor(editor *instance, data core.EditorData) {
	document := newEditorDocument(data.Value)
	if editor.editorLastValue != data.Value {
		editor.editorCursor = editorClamp(editor.editorCursor, 0, len(document.graphemes))
		editor.editorManualScroll = false
		editor.editorLastValue = data.Value
	}
	row, column := document.position(editor.editorCursor)
	tabWidth := editorTabWidth(data.TabWidth)
	visualColumn := editorVisualColumn(document.lines[row].graphemes, column, tabWidth)
	contentWidth := editorContentWidth(editor, data, document)
	maxRowOffset := maxInt(len(document.lines)-editor.rect.Height, 0)
	maxColumnOffset := editorMaxColumnOffset(document, tabWidth, contentWidth)
	editor.editorRowOffset = editorClamp(editor.editorRowOffset, 0, maxRowOffset)
	editor.editorColumnOffset = editorClamp(editor.editorColumnOffset, 0, maxColumnOffset)
	if editor.editorManualScroll {
		return
	}
	if row < editor.editorRowOffset {
		editor.editorRowOffset = row
	} else if row >= editor.editorRowOffset+editor.rect.Height {
		editor.editorRowOffset = row - editor.rect.Height + 1
	}
	if contentWidth > 0 {
		if visualColumn < editor.editorColumnOffset {
			editor.editorColumnOffset = visualColumn
		} else if visualColumn >= editor.editorColumnOffset+contentWidth {
			editor.editorColumnOffset = visualColumn - contentWidth + 1
		}
	}
	editor.editorRowOffset = editorClamp(editor.editorRowOffset, 0, maxRowOffset)
	editor.editorColumnOffset = editorClamp(editor.editorColumnOffset, 0, maxColumnOffset)
}

func paintEditor(buffer *screen.Buffer, editor *instance, data core.EditorData) {
	document := newEditorDocument(data.Value)
	tabWidth := editorTabWidth(data.TabWidth)
	contentWidth := editorContentWidth(editor, data, document)
	for screenRow := 0; screenRow < editor.rect.Height; screenRow++ {
		lineIndex := editor.editorRowOffset + screenRow
		if lineIndex >= len(document.lines) {
			break
		}
		line := document.lines[lineIndex]
		styles := editorLineStyles(editor.style, line.graphemes, data.Highlights, lineIndex)
		paintEditorLine(buffer, editor, line.graphemes, styles, screenRow, tabWidth, contentWidth)
	}
	if data.Value == "" && data.Placeholder != "" && editor.editorRowOffset == 0 {
		graphemes := uitext.Graphemes(data.Placeholder)
		styles := make([]Style, len(graphemes))
		for index := range styles {
			styles[index] = editor.style
		}
		paintEditorLine(buffer, editor, graphemes, styles, 0, tabWidth, contentWidth)
	}
	paintEditorCursor(buffer, editor, document, tabWidth, contentWidth)
	paintEditorScrollbar(buffer, editor, data, len(document.lines))
}

func editorLineStyles(base Style, graphemes []string, highlights [][]core.HighlightSpan, lineIndex int) []Style {
	styles := make([]Style, len(graphemes))
	for index := range styles {
		styles[index] = base
	}
	if lineIndex >= len(highlights) {
		return styles
	}
	for _, span := range highlights[lineIndex] {
		start := editorClamp(span.Start, 0, len(styles))
		end := editorClamp(span.End, start, len(styles))
		for index := start; index < end; index++ {
			styles[index], _ = core.ResolveStyle(styles[index], span.Style)
		}
	}
	return styles
}

func paintEditorLine(buffer *screen.Buffer, editor *instance, graphemes []string, styles []Style, screenRow, tabWidth, contentWidth int) {
	visual := 0
	left := editor.editorColumnOffset
	right := left + contentWidth
	y := editor.rect.Y + screenRow
	for index, grapheme := range graphemes {
		width := editorGraphemeWidth(grapheme, visual, tabWidth)
		start, end := visual, visual+width
		visual = end
		if end <= left {
			continue
		}
		if start >= right {
			break
		}
		if grapheme == "\t" {
			for column := maxInt(start, left); column < minInt(end, right); column++ {
				editorSetCell(buffer, editor, editor.rect.X+column-left, y, " ", styles[index])
			}
			continue
		}
		if start < left || end > right {
			continue
		}
		editorSetCell(buffer, editor, editor.rect.X+start-left, y, grapheme, styles[index])
	}
}

func paintEditorCursor(buffer *screen.Buffer, editor *instance, document editorDocument, tabWidth, contentWidth int) {
	if !appFocused(editor) || editor.rect.Empty() {
		return
	}
	row, column := document.position(editor.editorCursor)
	if row < editor.editorRowOffset || row >= editor.editorRowOffset+editor.rect.Height {
		return
	}
	visual := editorVisualColumn(document.lines[row].graphemes, column, tabWidth)
	x := editor.rect.X + visual - editor.editorColumnOffset
	y := editor.rect.Y + row - editor.editorRowOffset
	if x < editor.rect.X || x >= editor.rect.X+contentWidth || !editor.rect.Contains(x, y) || !editor.clip.Contains(x, y) {
		return
	}
	cell := buffer.Cell(x, y)
	cell.Style.Attributes |= core.AttributeReverse
	if cell.Grapheme == "" {
		cell.Grapheme = " "
	}
	buffer.Set(x, y, cell.Grapheme, cell.Style)
}

func editorScrollbarVisible(data core.EditorData, totalLines, viewportHeight int) bool {
	if data.Scrollbar == 2 || viewportHeight <= 0 {
		return false
	}
	return data.Scrollbar == 1 || totalLines > viewportHeight
}

func editorContentWidth(editor *instance, data core.EditorData, document editorDocument) int {
	width := editor.rect.Width
	if editorScrollbarVisible(data, len(document.lines), editor.rect.Height) && width > 0 {
		width--
	}
	return maxInt(width, 0)
}

func editorMaxColumnOffset(document editorDocument, tabWidth, contentWidth int) int {
	maximum := 0
	for _, line := range document.lines {
		maximum = maxInt(maximum, editorLineWidth(line, tabWidth)-contentWidth+1)
	}
	return maxInt(maximum, 0)
}

func paintEditorScrollbar(buffer *screen.Buffer, editor *instance, data core.EditorData, totalLines int) {
	height := editor.rect.Height
	if !editorScrollbarVisible(data, totalLines, height) || editor.rect.Width <= 0 {
		return
	}
	x := editor.rect.X + editor.rect.Width - 1
	for row := editor.rect.Y; row < editor.rect.Y+height; row++ {
		if editor.clip.Contains(x, row) {
			buffer.Set(x, row, "│", editor.style)
		}
	}
	if totalLines <= height {
		return
	}
	thumb := maxInt(height*height/maxInt(totalLines, 1), 1)
	start := editor.rect.Y + editor.editorRowOffset*maxInt(height-thumb, 0)/maxInt(totalLines-height, 1)
	for row := start; row < start+thumb; row++ {
		if editor.clip.Contains(x, row) {
			buffer.Set(x, row, "█", editor.style)
		}
	}
}

func editorSetCell(buffer *screen.Buffer, editor *instance, x, y int, grapheme string, style Style) {
	width := maxInt(uitext.Width(grapheme), 1)
	end := x + width - 1
	if !editor.rect.Contains(x, y) || !editor.rect.Contains(end, y) || !editor.clip.Contains(x, y) || !editor.clip.Contains(end, y) {
		return
	}
	buffer.Set(x, y, grapheme, style)
}

func (app *App) editorInsert(editor *instance, value string, source ChangeSource) {
	data, ok := editor.host.Data.(core.EditorData)
	if !ok || data.Disabled || data.ReadOnly {
		return
	}
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	document := newEditorDocument(data.Value)
	cursor := editorClamp(editor.editorCursor, 0, len(document.graphemes))
	insertion := uitext.Graphemes(value)
	next := append([]string(nil), document.graphemes[:cursor]...)
	next = append(next, insertion...)
	next = append(next, document.graphemes[cursor:]...)
	editor.editorCursor = cursor + len(insertion)
	editor.editorGoalSet = false
	editor.editorManualScroll = false
	app.emitEditorChange(editor, strings.Join(next, ""), source)
}

func (app *App) editorKey(editor *instance, event KeyEvent) bool {
	data, ok := editor.host.Data.(core.EditorData)
	if !ok || data.Disabled {
		return true
	}
	document := newEditorDocument(data.Value)
	editor.editorCursor = editorClamp(editor.editorCursor, 0, len(document.graphemes))
	row, column := document.position(editor.editorCursor)
	resetGoal := true
	switch event.Key {
	case KeyEnter:
		app.editorInsert(editor, "\n", ChangeKeyboard)
		return true
	case KeyLeft:
		editor.editorCursor = maxInt(editor.editorCursor-1, 0)
	case KeyRight:
		editor.editorCursor = minInt(editor.editorCursor+1, len(document.graphemes))
	case KeyHome:
		editor.editorCursor = document.cursor(row, 0)
	case KeyEnd:
		editor.editorCursor = document.cursor(row, len(document.lines[row].graphemes))
	case KeyUp:
		app.moveEditorVertically(editor, document, row, column, -1)
		resetGoal = false
	case KeyDown:
		app.moveEditorVertically(editor, document, row, column, 1)
		resetGoal = false
	case KeyPageUp:
		app.moveEditorVertically(editor, document, row, column, -maxInt(editor.rect.Height-1, 1))
		resetGoal = false
	case KeyPageDown:
		app.moveEditorVertically(editor, document, row, column, maxInt(editor.rect.Height-1, 1))
		resetGoal = false
	case KeyBackspace:
		if !data.ReadOnly && editor.editorCursor > 0 {
			app.deleteEditorRange(editor, document, editor.editorCursor-1, editor.editorCursor)
		}
		return true
	case KeyDelete:
		if !data.ReadOnly && editor.editorCursor < len(document.graphemes) {
			app.deleteEditorRange(editor, document, editor.editorCursor, editor.editorCursor+1)
		}
		return true
	default:
		return false
	}
	if resetGoal {
		editor.editorGoalSet = false
	}
	editor.editorManualScroll = false
	app.invalidated = true
	return true
}

func (app *App) moveEditorVertically(editor *instance, document editorDocument, row, column, delta int) {
	data, _ := editor.host.Data.(core.EditorData)
	if !editor.editorGoalSet {
		editor.editorGoalColumn = editorVisualColumn(document.lines[row].graphemes, column, editorTabWidth(data.TabWidth))
		editor.editorGoalSet = true
	}
	nextRow := editorClamp(row+delta, 0, len(document.lines)-1)
	nextColumn := editorColumnAtVisual(document.lines[nextRow].graphemes, editor.editorGoalColumn, editorTabWidth(data.TabWidth))
	editor.editorCursor = document.cursor(nextRow, nextColumn)
}

func (app *App) deleteEditorRange(editor *instance, document editorDocument, start, end int) {
	start = editorClamp(start, 0, len(document.graphemes))
	end = editorClamp(end, start, len(document.graphemes))
	next := append([]string(nil), document.graphemes[:start]...)
	next = append(next, document.graphemes[end:]...)
	editor.editorCursor = start
	editor.editorGoalSet = false
	editor.editorManualScroll = false
	app.emitEditorChange(editor, strings.Join(next, ""), ChangeKeyboard)
}

func (app *App) emitEditorChange(editor *instance, value string, source ChangeSource) {
	data, _ := editor.host.Data.(core.EditorData)
	if value == data.Value {
		return
	}
	if handler, ok := editor.handlerValue("change").(EventHandler[ValueChangeEvent]); ok && handler != nil {
		handler(ValueChangeEvent{Previous: data.Value, Value: value, Source: source})
	}
	app.invalidated = true
}

func (app *App) positionEditorCursor(editor *instance, x, y int) {
	data, ok := editor.host.Data.(core.EditorData)
	if !ok {
		return
	}
	document := newEditorDocument(data.Value)
	row := editorClamp(editor.editorRowOffset+y, 0, len(document.lines)-1)
	localX := maxInt(x, 0)
	if editorScrollbarVisible(data, len(document.lines), editor.rect.Height) {
		contentWidth := editorContentWidth(editor, data, document)
		if x >= contentWidth {
			return
		}
		localX = editorClamp(localX, 0, maxInt(contentWidth-1, 0))
	}
	target := editor.editorColumnOffset + localX
	column := editorColumnAtVisual(document.lines[row].graphemes, target, editorTabWidth(data.TabWidth))
	cursor := document.cursor(row, column)
	if cursor != editor.editorCursor {
		editor.editorCursor = cursor
		app.invalidated = true
	}
	editor.editorGoalSet = false
	editor.editorManualScroll = false
}

func (app *App) scrollEditor(editor *instance, deltaX, deltaY int) {
	data, ok := editor.host.Data.(core.EditorData)
	if !ok {
		return
	}
	document := newEditorDocument(data.Value)
	maxRow := maxInt(len(document.lines)-editor.rect.Height, 0)
	contentWidth := editorContentWidth(editor, data, document)
	maxColumn := editorMaxColumnOffset(document, editorTabWidth(data.TabWidth), contentWidth)
	nextRow := editorClamp(editor.editorRowOffset+deltaY, 0, maxRow)
	nextColumn := editorClamp(editor.editorColumnOffset+deltaX, 0, maxColumn)
	if nextRow == editor.editorRowOffset && nextColumn == editor.editorColumnOffset {
		return
	}
	editor.editorRowOffset = nextRow
	editor.editorColumnOffset = nextColumn
	editor.editorManualScroll = true
	app.invalidated = true
}

func editorClamp(value, minimum, maximum int) int {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}
