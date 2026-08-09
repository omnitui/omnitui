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
			childWidth = gridMeasuredTrackSize(data, index, childWidth)
			width += childWidth
			if index > 0 {
				width -= gridBorderOverlap(data)
			}
			height = maxInt(height, childHeight)
		} else {
			childHeight = gridMeasuredTrackSize(data, index, childHeight)
			width = maxInt(width, childWidth)
			height += childHeight
			if index > 0 {
				height -= gridBorderOverlap(data)
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
	total := extent + gridBorderOverlap(data)*(count-1)
	reset := len(grid.gridSizes) != count || grid.gridOrientation != data.Orientation
	if reset {
		grid.gridSizes = initialGridSizes(data, count, total)
		grid.gridDragging = false
	} else {
		grid.gridSizes = fitGridSizes(grid.gridSizes, data, total)
	}
	if grid.gridExtent != extent {
		grid.gridDragging = false
	}
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
		cursor += maxInt(size-gridBorderOverlap(data), 0)
	}
}

func gridBorderOverlap(_ core.GridData) int {
	return 1
}

func gridMeasuredTrackSize(data core.GridData, index, content int) int {
	minimum, maximum := gridTrackConstraint(data, index)
	size := content
	if initial := gridTrack(data, index).InitialSize; initial > 0 {
		size = initial
	}
	size = maxInt(size, minimum)
	if maximum > 0 {
		size = minInt(size, maximum)
	}
	return size
}

func initialGridSizes(data core.GridData, count, total int) []int {
	if count == 0 {
		return nil
	}
	total = maxInt(total, 0)
	minimums, maximums := gridTrackBounds(data, count, total)
	result := append([]int(nil), minimums...)
	automatic := make([]int, 0, count)
	for index := range result {
		track := gridTrack(data, index)
		if track.InitialSize == 0 {
			automatic = append(automatic, index)
			continue
		}
		result[index] = gridLimit(track.InitialSize, minimums[index], maximums[index])
	}
	shrinkGridSizes(result, minimums, total)
	growGridSizesEvenly(result, maximums, automatic, total)
	growGridSizesFromEnd(result, maximums, total)
	return result
}

func fitGridSizes(current []int, data core.GridData, total int) []int {
	count := len(current)
	if count == 0 {
		return nil
	}
	total = maxInt(total, 0)
	minimums, maximums := gridTrackBounds(data, count, total)
	result := append([]int(nil), current...)
	for index := range result {
		result[index] = gridLimit(result[index], minimums[index], maximums[index])
	}
	shrinkGridSizes(result, minimums, total)
	growGridSizesFromEnd(result, maximums, total)
	return result
}

func gridTrackBounds(data core.GridData, count, total int) ([]int, []int) {
	minimums := make([]int, count)
	for index := range minimums {
		minimums[index], _ = gridTrackConstraint(data, index)
	}
	for gridSizeSum(minimums) > total {
		largest := -1
		for index, minimum := range minimums {
			if minimum > 0 && (largest < 0 || minimum > minimums[largest]) {
				largest = index
			}
		}
		if largest < 0 {
			break
		}
		minimums[largest]--
	}
	maximums := make([]int, count)
	for index := range maximums {
		_, maximum := gridTrackConstraint(data, index)
		if maximum == 0 {
			maximum = total
		}
		maximums[index] = maxInt(maximum, minimums[index])
	}
	return minimums, maximums
}

func shrinkGridSizes(values, minimums []int, total int) {
	sum := gridSizeSum(values)
	for sum > total {
		changed := false
		for index := len(values) - 1; index >= 0 && sum > total; index-- {
			if values[index] <= minimums[index] {
				continue
			}
			values[index]--
			sum--
			changed = true
		}
		if !changed {
			break
		}
	}
}

func growGridSizesEvenly(values, maximums, indexes []int, total int) {
	sum := gridSizeSum(values)
	for sum < total {
		changed := false
		for _, index := range indexes {
			if sum >= total {
				break
			}
			if values[index] >= maximums[index] {
				continue
			}
			values[index]++
			sum++
			changed = true
		}
		if !changed {
			break
		}
	}
}

func growGridSizesFromEnd(values, maximums []int, total int) {
	remaining := total - gridSizeSum(values)
	for index := len(values) - 1; index >= 0 && remaining > 0; index-- {
		growth := minInt(maximums[index]-values[index], remaining)
		if growth <= 0 {
			continue
		}
		values[index] += growth
		remaining -= growth
	}
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

func gridTrack(data core.GridData, index int) core.GridTrackData {
	if index < 0 || index >= len(data.Tracks) {
		return core.GridTrackData{}
	}
	return data.Tracks[index]
}

func gridTrackConstraint(data core.GridData, index int) (int, int) {
	track := gridTrack(data, index)
	minimum := track.MinSize
	if minimum == 0 {
		minimum = gridMinimum(data)
	}
	maximum := track.MaxSize
	if maximum > 0 && maximum < minimum {
		maximum = minimum
	}
	return minimum, maximum
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
	minimums, maximums := gridTrackBounds(data, len(grid.gridSizes), gridSizeSum(grid.gridSizes))
	delta := coordinate - grid.gridDragOrigin
	index := grid.gridDragIndex
	lower := minimums[index] - grid.gridDragFirst
	lower = maxInt(lower, grid.gridDragSecond-maximums[index+1])
	upper := grid.gridDragSecond - minimums[index+1]
	upper = minInt(upper, maximums[index]-grid.gridDragFirst)
	if lower > upper {
		return
	}
	delta = gridLimit(delta, lower, upper)
	first := grid.gridDragFirst + delta
	second := grid.gridDragSecond - delta
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
