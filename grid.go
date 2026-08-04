package omnitui

import (
	"github.com/omnitui/omnitui/v2/internal/core"
	"github.com/omnitui/omnitui/v2/internal/screen"
)

func measureGrid(grid *instance, data core.GridData, maxWidth, maxHeight int) (int, int) {
	width, height := 0, 0
	for index, child := range grid.children {
		childWidth, childHeight := measureNode(child, maxWidth, maxHeight)
		if data.Orientation == 0 {
			width += childWidth
			if index > 0 {
				width--
			}
			height = maxInt(height, childHeight)
		} else {
			width = maxInt(width, childWidth)
			height += childHeight
			if index > 0 {
				height--
			}
		}
	}
	width = maxInt(width, 0)
	height = maxInt(height, 0)
	if core.SizeModeOf(data.Width) == core.SizeCells {
		width = core.SizeValueOf(data.Width)
	} else if core.SizeModeOf(data.Width) == core.SizeFill {
		width = maxWidth
	}
	if core.SizeModeOf(data.Height) == core.SizeCells {
		height = core.SizeValueOf(data.Height)
	} else if core.SizeModeOf(data.Height) == core.SizeFill {
		height = maxHeight
	}
	return minInt(width, maxWidth), minInt(height, maxHeight)
}

func arrangeGrid(grid *instance, data core.GridData) {
	count := len(grid.children)
	if count == 0 {
		grid.gridSizes = nil
		grid.gridExtent = 0
		grid.gridDragging = false
		return
	}
	extent := grid.rect.Width
	if data.Orientation == 1 {
		extent = grid.rect.Height
	}
	if len(grid.gridSizes) != count || grid.gridOrientation != data.Orientation {
		grid.gridSizes = make([]int, count)
		grid.gridDragging = false
	}
	if grid.gridExtent != extent {
		grid.gridDragging = false
	}
	grid.gridSizes = fitGridSizes(grid.gridSizes, extent+count-1, gridMinimum(data))
	grid.gridExtent = extent
	grid.gridOrientation = data.Orientation

	cursor := 0
	clip := core.IntersectRect(grid.clip, grid.rect)
	for index, child := range grid.children {
		size := grid.gridSizes[index]
		panel := Rect{X: grid.rect.X, Y: grid.rect.Y, Width: grid.rect.Width, Height: grid.rect.Height}
		if data.Orientation == 0 {
			panel.X += cursor
			panel.Width = size
		} else {
			panel.Y += cursor
			panel.Height = size
		}
		arrangeNode(child, panel, clip, grid.style, nil)
		cursor += maxInt(size-1, 0)
	}
}

func fitGridSizes(current []int, total, minimum int) []int {
	count := len(current)
	if count == 0 {
		return nil
	}
	total = maxInt(total, 0)
	if total < count*minimum {
		minimum = total / count
	}
	result := append([]int(nil), current...)
	if gridSizeSum(result) == 0 {
		for index := range result {
			result[index] = total / count
			if index < total%count {
				result[index]++
			}
		}
		return result
	}
	for index := range result {
		result[index] = maxInt(result[index], minimum)
	}
	sum := gridSizeSum(result)
	for sum > total {
		changed := false
		for index := len(result) - 1; index >= 0 && sum > total; index-- {
			if result[index] <= minimum {
				continue
			}
			result[index]--
			sum--
			changed = true
		}
		if !changed {
			break
		}
	}
	if sum < total {
		result[len(result)-1] += total - sum
	}
	return result
}

func gridSizeSum(values []int) int {
	total := 0
	for _, value := range values {
		total += value
	}
	return total
}

func gridMinimum(data core.GridData) int {
	if data.MinPanelSize < 3 {
		return 3
	}
	return data.MinPanelSize
}

func paintGrid(buffer *screen.Buffer, grid *instance, data core.GridData) {
	for _, child := range grid.children {
		paintNode(buffer, child)
	}
	if len(grid.gridSizes) < 2 || grid.rect.Empty() {
		return
	}
	glyphs := gridDividerGlyphs(data.Border, data.Orientation)
	cursor := 0
	for index := 0; index < len(grid.gridSizes)-1; index++ {
		cursor += maxInt(grid.gridSizes[index]-1, 0)
		if data.Orientation == 0 {
			for row := 0; row < grid.rect.Height; row++ {
				glyph := glyphs[1]
				if row == 0 {
					glyph = glyphs[0]
				} else if row == grid.rect.Height-1 {
					glyph = glyphs[2]
				}
				setGridCell(buffer, grid, grid.rect.X+cursor, grid.rect.Y+row, glyph)
			}
		} else {
			for column := 0; column < grid.rect.Width; column++ {
				glyph := glyphs[1]
				if column == 0 {
					glyph = glyphs[0]
				} else if column == grid.rect.Width-1 {
					glyph = glyphs[2]
				}
				setGridCell(buffer, grid, grid.rect.X+column, grid.rect.Y+cursor, glyph)
			}
		}
	}
}

func gridDividerGlyphs(border, orientation uint8) [3]string {
	if orientation == 0 {
		if border == 3 {
			return [3]string{"╦", "║", "╩"}
		}
		return [3]string{"╥", "║", "╨"}
	}
	if border == 3 {
		return [3]string{"╠", "═", "╣"}
	}
	return [3]string{"╞", "═", "╡"}
}

func setGridCell(buffer *screen.Buffer, grid *instance, x, y int, grapheme string) {
	if grid.rect.Contains(x, y) && grid.clip.Contains(x, y) {
		buffer.Set(x, y, grapheme, grid.style)
	}
}

func (app *App) gridMouse(grid *instance, event MouseEvent) bool {
	data, ok := grid.host.Data.(core.GridData)
	if !ok || len(grid.gridSizes) < 2 {
		return false
	}
	coordinate := event.X
	if data.Orientation == 1 {
		coordinate = event.Y
	}
	if event.Action == MouseDown && event.Button == MouseButtonLeft {
		divider := gridDividerAt(grid, data, coordinate)
		if divider < 0 {
			return false
		}
		grid.gridDragging = true
		grid.gridDragIndex = divider
		grid.gridDragOrigin = coordinate
		grid.gridDragFirst = grid.gridSizes[divider]
		grid.gridDragSecond = grid.gridSizes[divider+1]
		app.capture = grid
		app.pressTarget = nil
		return true
	}
	if event.Action == MouseUp && event.Button == MouseButtonLeft {
		if !grid.gridDragging {
			return false
		}
		app.resizeGrid(grid, data, coordinate)
		grid.gridDragging = false
		return true
	}
	if event.Action != MouseMove || !grid.gridDragging {
		return false
	}
	app.resizeGrid(grid, data, coordinate)
	return true
}

func (app *App) resizeGrid(grid *instance, data core.GridData, coordinate int) {
	pair := grid.gridDragFirst + grid.gridDragSecond
	minimum := gridMinimum(data)
	if pair < minimum*2 {
		minimum = pair / 2
	}
	delta := coordinate - grid.gridDragOrigin
	delta = gridLimit(delta, minimum-grid.gridDragFirst, grid.gridDragSecond-minimum)
	first := grid.gridDragFirst + delta
	second := grid.gridDragSecond - delta
	index := grid.gridDragIndex
	if first == grid.gridSizes[index] && second == grid.gridSizes[index+1] {
		return
	}
	grid.gridSizes[index] = first
	grid.gridSizes[index+1] = second
	app.invalidated = true
}

func gridDividerAt(grid *instance, data core.GridData, coordinate int) int {
	start := grid.rect.X
	if data.Orientation == 1 {
		start = grid.rect.Y
	}
	cursor := start
	for index := 0; index < len(grid.gridSizes)-1; index++ {
		cursor += maxInt(grid.gridSizes[index]-1, 0)
		if coordinate == cursor {
			return index
		}
	}
	return -1
}

func gridLimit(value, minimum, maximum int) int {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}
