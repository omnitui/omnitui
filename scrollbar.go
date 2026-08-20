package omnitui

import "github.com/omnitui/omnitui/v2/internal/core"

type verticalScrollbarMetrics struct {
	thumbStart    int
	thumbSize     int
	maxThumbStart int
	maxOffset     int
}

func verticalScrollbar(viewport, total, offset int) (verticalScrollbarMetrics, bool) {
	if viewport <= 0 || total <= viewport {
		return verticalScrollbarMetrics{}, false
	}
	thumbSize := maxInt(viewport*viewport/maxInt(total, 1), 1)
	maxThumbStart := maxInt(viewport-thumbSize, 0)
	maxOffset := total - viewport
	offset = maxInt(0, minInt(offset, maxOffset))
	return verticalScrollbarMetrics{
		thumbStart:    offset * maxThumbStart / maxOffset,
		thumbSize:     thumbSize,
		maxThumbStart: maxThumbStart,
		maxOffset:     maxOffset,
	}, true
}

func (metrics verticalScrollbarMetrics) offsetAt(thumbStart int) int {
	if metrics.maxThumbStart <= 0 {
		return 0
	}
	thumbStart = maxInt(0, minInt(thumbStart, metrics.maxThumbStart))
	return (thumbStart*metrics.maxOffset + metrics.maxThumbStart/2) / metrics.maxThumbStart
}

func (app *App) scrollbarMouse(scrollable *instance, event MouseEvent) bool {
	metrics, ok := scrollbarMetrics(scrollable)
	if !ok {
		scrollable.scrollbarDragging = false
		return false
	}
	localY := event.Y - scrollable.rect.Y
	switch event.Action {
	case MouseDown:
		if event.Button != MouseButtonLeft || event.X != scrollable.rect.X+scrollable.rect.Width-1 || !scrollable.clip.Contains(event.X, event.Y) {
			return false
		}
		if localY < metrics.thumbStart || localY >= metrics.thumbStart+metrics.thumbSize {
			return false
		}
		scrollable.scrollbarDragging = true
		scrollable.scrollbarDragGrab = localY - metrics.thumbStart
		app.capture = scrollable
		app.pressTarget = nil
		return true
	case MouseMove:
		if !scrollable.scrollbarDragging {
			return false
		}
		app.setScrollbarOffset(scrollable, metrics.offsetAt(localY-scrollable.scrollbarDragGrab))
		return true
	case MouseUp:
		if event.Button != MouseButtonLeft || !scrollable.scrollbarDragging {
			return false
		}
		app.setScrollbarOffset(scrollable, metrics.offsetAt(localY-scrollable.scrollbarDragGrab))
		scrollable.scrollbarDragging = false
		return true
	default:
		return false
	}
}

func scrollbarMetrics(scrollable *instance) (verticalScrollbarMetrics, bool) {
	switch data := scrollable.host.Data.(type) {
	case core.EditorData:
		if data.Disabled {
			return verticalScrollbarMetrics{}, false
		}
		total := len(newEditorDocument(data.Value).lines)
		if !editorScrollbarVisible(data, total, scrollable.rect.Height) {
			return verticalScrollbarMetrics{}, false
		}
		return verticalScrollbar(scrollable.rect.Height, total, scrollable.editorRowOffset)
	case core.ListData:
		if data.Disabled || data.Scrollbar == 2 {
			return verticalScrollbarMetrics{}, false
		}
		return verticalScrollbar(scrollable.rect.Height, listContentHeight(scrollable, data), scrollable.listOffset)
	default:
		return verticalScrollbarMetrics{}, false
	}
}

func (app *App) setScrollbarOffset(scrollable *instance, offset int) {
	switch scrollable.host.Data.(type) {
	case core.EditorData:
		if offset == scrollable.editorRowOffset {
			return
		}
		scrollable.editorRowOffset = offset
		scrollable.editorManualScroll = true
	case core.ListData:
		if offset == scrollable.listOffset {
			return
		}
		scrollable.listOffset = offset
		scrollable.listManual = true
		scrollable.listAnchorKey = ""
	default:
		return
	}
	app.invalidated = true
}

func listContentHeight(list *instance, data core.ListData) int {
	total := 0
	for index, child := range list.children {
		_, height := measureNode(child, list.rect.Width, list.rect.Height)
		total += height
		if index > 0 {
			total += data.Gap
		}
	}
	return total
}
