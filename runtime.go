package omnitui

import (
	"strings"

	"github.com/omnitui/omnitui/internal/core"
	"github.com/omnitui/omnitui/internal/screen"
	uitext "github.com/omnitui/omnitui/internal/text"
)

const unbounded = 1 << 29

const tabHorizontalPadding = 1

func (app *App) render() error {
	app.rootInstance = reconcile(nil, app.rootInstance, app.root, nil, app, "root")
	validateListSelections(app.rootInstance)
	rootRect := Rect{Width: app.width, Height: app.height}
	if app.rootInstance != nil {
		if width, height, ok := explicitRootSize(app.rootInstance); ok {
			if width > 0 {
				rootRect.Width = minInt(width, app.width)
			}
			if height > 0 {
				rootRect.Height = minInt(height, app.height)
			}
		}
	}
	if app.rootInstance != nil {
		arrangeNode(app.rootInstance, rootRect, rootRect, Style{}, nil)
	}
	app.recoverFocus()
	app.back = screen.NewBuffer(app.width, app.height, Style{})
	if app.rootInstance != nil {
		paintNode(app.back, app.rootInstance)
	}
	profile := resolveColorProfile(app.options.ColorProfile)
	output := screen.Diff(app.front, app.back, profile)
	if app.backend != nil && len(output) > 0 {
		if err := app.backend.Write(output); err != nil {
			return err
		}
	}
	app.front = app.back
	app.back = nil
	app.invalidated = false
	return nil
}

func validateListSelections(root *instance) {
	if root == nil {
		return
	}
	if root.kind() == core.KindHost && root.host.Kind == core.HostList {
		data, _ := root.host.Data.(core.ListData)
		if !data.Selectable {
			for _, child := range root.children {
				validateListSelections(child)
			}
			return
		}
		candidate := ""
		for _, child := range root.children {
			if key := core.KeyOf(child.element); key != "" {
				candidate = key
				break
			}
		}
		if data.SelectedKey != "" && candidate != "" {
			found := false
			for _, child := range root.children {
				if core.KeyOf(child.element) == data.SelectedKey {
					found = true
					break
				}
			}
			if !found && (root.listInvalidSelected != data.SelectedKey || root.listInvalidProposal != candidate) {
				root.listInvalidSelected, root.listInvalidProposal = data.SelectedKey, candidate
				if handler, ok := root.handlerValue("change").(EventHandler[ValueChangeEvent]); ok && handler != nil {
					handler(ValueChangeEvent{Previous: data.SelectedKey, Value: candidate, Source: ChangeProgrammatic})
				}
			}
		}
		if foundKey(root, data.SelectedKey) {
			root.listInvalidSelected, root.listInvalidProposal = "", ""
		}
	}
	for _, child := range root.children {
		validateListSelections(child)
	}
}

func foundKey(i *instance, key string) bool {
	if key == "" {
		return false
	}
	for _, child := range i.children {
		if core.KeyOf(child.element) == key {
			return true
		}
	}
	return false
}

func explicitRootSize(i *instance) (int, int, bool) {
	if i == nil {
		return 0, 0, false
	}
	if i.kind() != core.KindHost && len(i.children) == 1 {
		return explicitRootSize(i.children[0])
	}
	if i.kind() != core.KindHost {
		return 0, 0, false
	}
	switch data := i.host.Data.(type) {
	case core.BoxData:
		width, height := 0, 0
		if core.SizeModeOf(data.Width) == core.SizeCells {
			width = core.SizeValueOf(data.Width)
		}
		if core.SizeModeOf(data.Height) == core.SizeCells {
			height = core.SizeValueOf(data.Height)
		}
		return width, height, width > 0 || height > 0
	case core.InputData:
		if core.SizeModeOf(data.Width) == core.SizeCells {
			return core.SizeValueOf(data.Width), 1, true
		}
	case core.ListData:
		if core.SizeModeOf(data.Height) == core.SizeCells {
			return 0, core.SizeValueOf(data.Height), true
		}
	}
	return 0, 0, false
}

func measureNode(i *instance, maxWidth, maxHeight int) (int, int) {
	if i == nil {
		return 0, 0
	}
	switch i.kind() {
	case core.KindComponent, core.KindProvider:
		if len(i.children) == 0 {
			return 0, 0
		}
		return measureNode(i.children[0], maxWidth, maxHeight)
	case core.KindFragment:
		width, height := 0, 0
		for _, child := range i.children {
			childWidth, childHeight := measureNode(child, maxWidth, maxHeight)
			if childWidth > width {
				width = childWidth
			}
			height += childHeight
		}
		return width, height
	case core.KindHost:
		return measureHost(i, maxWidth, maxHeight)
	default:
		return 0, 0
	}
}

func measureHost(i *instance, maxWidth, maxHeight int) (int, int) {
	switch data := i.host.Data.(type) {
	case core.TextData:
		lines := uitext.Wrap(data.Content, maxInt(maxWidth, 1), data.Wrap == 1)
		if data.Wrap == 2 {
			lines = uitext.Wrap(data.Content, maxInt(maxWidth, 1), false)
		}
		if data.Wrap == 0 {
			lines = strings.Split(data.Content, "\n")
		}
		if data.MaxLines > 0 && len(lines) > data.MaxLines {
			lines = lines[:data.MaxLines]
		}
		width := 0
		for _, line := range lines {
			if value := uitext.Width(line); value > width {
				width = value
			}
		}
		return minInt(width, maxWidth), len(lines)
	case core.ButtonData:
		padding := 0
		if !data.Plain {
			padding = 4
		}
		return minInt(uitext.Width(data.Label)+padding, maxWidth), 1
	case core.InputData:
		width := uitext.Width(data.Value)
		if width == 0 {
			width = uitext.Width(data.Placeholder)
		}
		if width < 1 {
			width = 1
		}
		if core.SizeModeOf(data.Width) == core.SizeCells {
			width = core.SizeValueOf(data.Width)
		}
		return minInt(width, maxWidth), 1
	case core.TabsData:
		return measureTabs(i, data, maxWidth, maxHeight)
	case core.ListData:
		return measureList(i, data, maxWidth, maxHeight)
	case core.BoxData:
		return measureBox(i, data, maxWidth, maxHeight)
	default:
		return 0, 0
	}
}

func measureBox(i *instance, data core.BoxData, maxWidth, maxHeight int) (int, int) {
	children := flattenLayoutChildren(i.children)
	border := 0
	if data.Border != 0 {
		border = 2
	}
	innerWidth := maxInt(maxWidth-border-data.Padding.Left-data.Padding.Right, 1)
	innerHeight := maxInt(maxHeight-border-data.Padding.Top-data.Padding.Bottom, 1)
	width, height := 0, 0
	if data.Direction == 0 {
		if data.Wrap {
			width, height = measureWrapped(children, data, innerWidth, innerHeight)
		} else {
			for index, child := range children {
				childWidth, childHeight := measureNode(child, innerWidth, innerHeight)
				width += childWidth
				if index > 0 {
					width += data.Gap
				}
				if childHeight > height {
					height = childHeight
				}
			}
		}
	} else {
		for index, child := range children {
			childWidth, childHeight := measureNode(child, innerWidth, innerHeight)
			if childWidth > width {
				width = childWidth
			}
			height += childHeight
			if index > 0 {
				height += data.Gap
			}
		}
	}
	width += border + data.Padding.Left + data.Padding.Right
	height += border + data.Padding.Top + data.Padding.Bottom
	if core.SizeModeOf(data.Width) == core.SizeCells {
		width = core.SizeValueOf(data.Width)
	}
	if core.SizeModeOf(data.Height) == core.SizeCells {
		height = core.SizeValueOf(data.Height)
	}
	width = clamp(width, data.MinWidth, data.MaxWidth)
	height = clamp(height, data.MinHeight, data.MaxHeight)
	return minInt(width, maxWidth), minInt(height, maxHeight)
}

func measureWrapped(children []*instance, data core.BoxData, innerWidth, innerHeight int) (int, int) {
	lineWidth, lineHeight, width, height := 0, 0, 0, 0
	for index, child := range children {
		childWidth, childHeight := measureNode(child, innerWidth, innerHeight)
		extra := childWidth
		if lineWidth > 0 {
			extra += data.Gap
		}
		if lineWidth > 0 && lineWidth+extra > innerWidth {
			if lineWidth > width {
				width = lineWidth
			}
			height += lineHeight + data.Gap
			lineWidth, lineHeight = childWidth, childHeight
		} else {
			lineWidth += extra
			if childHeight > lineHeight {
				lineHeight = childHeight
			}
		}
		if index == len(children)-1 {
			if lineWidth > width {
				width = lineWidth
			}
			height += lineHeight
		}
	}
	return width, height
}

func flattenLayoutChildren(values []*instance) []*instance {
	var result []*instance
	for _, child := range values {
		if child.kind() == core.KindFragment {
			result = append(result, flattenLayoutChildren(child.children)...)
			continue
		}
		if (child.kind() == core.KindComponent || child.kind() == core.KindProvider) && len(child.children) == 1 && child.children[0].kind() == core.KindFragment {
			result = append(result, flattenLayoutChildren(child.children[0].children)...)
			continue
		}
		result = append(result, child)
	}
	return result
}

func arrangeFragmentShells(values []*instance, rect, clip Rect, style Style) {
	for _, child := range values {
		if child.kind() == core.KindFragment || child.kind() == core.KindComponent || child.kind() == core.KindProvider {
			child.rect, child.clip, child.style = rect, clip, style
			arrangeFragmentShells(child.children, rect, clip, style)
		}
	}
}

func measureTabs(i *instance, data core.TabsData, maxWidth, maxHeight int) (int, int) {
	headerWidth, headerHeight := 0, 1
	for _, item := range data.Items {
		size := uitext.Width(item.Label)
		if data.Orientation == 0 {
			headerWidth += size + tabHorizontalPadding*2
		} else if size+tabHorizontalPadding*2 > headerWidth {
			headerWidth = size + tabHorizontalPadding*2
		}
		if data.Orientation == 1 {
			headerHeight++
		}
	}
	contentWidth, contentHeight := 0, 0
	if len(i.children) > 0 {
		contentWidth, contentHeight = measureNode(i.children[0], maxWidth, maxHeight)
	}
	if data.Orientation == 0 {
		return minInt(maxInt(headerWidth, contentWidth), maxWidth), minInt(headerHeight+contentHeight, maxHeight)
	}
	return minInt(headerWidth+contentWidth, maxWidth), minInt(maxInt(headerHeight, contentHeight), maxHeight)
}

func measureList(i *instance, data core.ListData, maxWidth, maxHeight int) (int, int) {
	width, height := 0, 0
	if len(i.children) == 0 {
		return 0, 0
	}
	for index, child := range i.children {
		childWidth, childHeight := measureNode(child, maxWidth, maxHeight)
		if childWidth > width {
			width = childWidth
		}
		height += childHeight
		if index > 0 {
			height += data.Gap
		}
	}
	if core.SizeModeOf(data.Height) == core.SizeCells {
		height = core.SizeValueOf(data.Height)
	}
	return minInt(width, maxWidth), minInt(height, maxHeight)
}

func arrangeNode(i *instance, rect, clip Rect, inherited Style, override *Style) {
	if i == nil {
		return
	}
	i.rect, i.clip = rect, core.IntersectRect(clip, rect)
	own := hostStyle(i)
	if override != nil {
		own, _ = core.ResolveStyle(own, *override)
	}
	i.style, _ = core.ResolveStyle(inherited, own)
	switch i.kind() {
	case core.KindComponent, core.KindProvider:
		if len(i.children) > 0 {
			arrangeNode(i.children[0], rect, clip, i.style, nil)
		}
	case core.KindFragment:
		y := rect.Y
		for _, child := range i.children {
			_, height := measureNode(child, rect.Width, rect.Height)
			arrangeNode(child, Rect{X: rect.X, Y: y, Width: rect.Width, Height: height}, clip, i.style, nil)
			y += height
		}
	case core.KindHost:
		switch data := i.host.Data.(type) {
		case core.BoxData:
			arrangeBox(i, data)
		case core.TabsData:
			arrangeTabs(i, data)
		case core.ListData:
			arrangeList(i, data)
		default:
			for _, child := range i.children {
				arrangeNode(child, rect, clip, i.style, nil)
			}
		}
	}
}

func fillHost(buffer *screen.Buffer, i *instance) {
	if core.ColorKindOf(i.style.Background) == core.ColorUnspecified && i.style.Attributes == 0 {
		return
	}
	for y := i.clip.Y; y < i.clip.Y+i.clip.Height; y++ {
		for x := i.clip.X; x < i.clip.X+i.clip.Width; x++ {
			buffer.Set(x, y, " ", i.style)
		}
	}
}

func arrangeBox(i *instance, data core.BoxData) {
	border := 0
	if data.Border != 0 {
		border = 1
	}
	inner := Rect{X: i.rect.X + border + data.Padding.Left, Y: i.rect.Y + border + data.Padding.Top, Width: maxInt(i.rect.Width-2*border-data.Padding.Left-data.Padding.Right, 0), Height: maxInt(i.rect.Height-2*border-data.Padding.Top-data.Padding.Bottom, 0)}
	children := flattenLayoutChildren(i.children)
	if len(children) == 0 {
		return
	}
	arrangeFragmentShells(i.children, inner, childClip(i, inner), i.style)
	if data.Direction == 0 && data.Wrap {
		arrangeWrappedBox(children, data, inner, i)
		return
	}
	widths := make([]int, len(children))
	heights := make([]int, len(children))
	total := 0
	cross := 0
	for index, child := range children {
		widths[index], heights[index] = measureNode(child, inner.Width, inner.Height)
		if data.Direction == 0 {
			total += widths[index]
			if heights[index] > cross {
				cross = heights[index]
			}
		} else {
			total += heights[index]
			if widths[index] > cross {
				cross = widths[index]
			}
		}
	}
	if len(children) > 1 {
		total += data.Gap * (len(children) - 1)
	}
	free := inner.Width - total
	if data.Direction == 1 {
		free = inner.Height - total
	}
	if free < 0 {
		free = 0
	}
	if free > 0 {
		flexGrow := make([]int, len(children))
		totalGrow := 0
		for index, child := range children {
			flexGrow[index] = flexGrowOf(child)
			totalGrow += flexGrow[index]
		}
		if totalGrow > 0 {
			distributed := 0
			for index, grow := range flexGrow {
				if grow == 0 {
					continue
				}
				amount := free * grow / totalGrow
				distributed += amount
				if data.Direction == 0 {
					widths[index] += amount
				} else {
					heights[index] += amount
				}
			}
			for index := 0; distributed < free; index = (index + 1) % len(flexGrow) {
				if flexGrow[index] == 0 {
					continue
				}
				if data.Direction == 0 {
					widths[index]++
				} else {
					heights[index]++
				}
				distributed++
			}
			free = 0
		}
	}
	lead, between := justify(data.Justify, free, len(children))
	main := lead
	for index, child := range children {
		childWidth, childHeight := widths[index], heights[index]
		if data.Direction == 0 {
			y := aligned(inner.Y, inner.Height, childHeight, data.Align)
			if data.Align == 3 {
				childHeight = inner.Height
				y = inner.Y
			}
			arrangeNode(child, Rect{X: inner.X + main, Y: y, Width: childWidth, Height: childHeight}, childClip(i, inner), i.style, nil)
			main += childWidth + data.Gap + between
		} else {
			x := aligned(inner.X, inner.Width, childWidth, data.Align)
			if data.Align == 3 {
				childWidth = inner.Width
				x = inner.X
			}
			arrangeNode(child, Rect{X: x, Y: inner.Y + main, Width: childWidth, Height: childHeight}, childClip(i, inner), i.style, nil)
			main += childHeight + data.Gap + between
		}
	}
}

func flexGrowOf(i *instance) int {
	if i == nil {
		return 0
	}
	switch i.kind() {
	case core.KindHost:
		if data, ok := i.host.Data.(core.BoxData); ok {
			return data.FlexGrow
		}
	case core.KindComponent, core.KindProvider, core.KindFragment:
		if len(i.children) == 1 {
			return flexGrowOf(i.children[0])
		}
	}
	return 0
}

func arrangeWrappedBox(children []*instance, data core.BoxData, inner Rect, parent *instance) {
	type line struct {
		indexes       []int
		width, height int
	}
	var lines []line
	for index, child := range children {
		width, height := measureNode(child, inner.Width, inner.Height)
		last := len(lines) - 1
		if last < 0 || (lines[last].width > 0 && lines[last].width+data.Gap+width > inner.Width) {
			lines = append(lines, line{})
			last++
		}
		if lines[last].width > 0 {
			lines[last].width += data.Gap
		}
		lines[last].width += width
		if height > lines[last].height {
			lines[last].height = height
		}
		lines[last].indexes = append(lines[last].indexes, index)
	}
	y := inner.Y
	for _, current := range lines {
		widths := make([]int, len(current.indexes))
		totalGrow := 0
		for position, index := range current.indexes {
			widths[position], _ = measureNode(children[index], inner.Width, inner.Height)
			totalGrow += flexGrowOf(children[index])
		}
		free := maxInt(inner.Width-current.width, 0)
		if totalGrow > 0 {
			distributed := 0
			for position, index := range current.indexes {
				grow := flexGrowOf(children[index])
				if grow == 0 {
					continue
				}
				amount := free * grow / totalGrow
				widths[position] += amount
				distributed += amount
			}
			for position := 0; distributed < free; position = (position + 1) % len(widths) {
				if flexGrowOf(children[current.indexes[position]]) == 0 {
					continue
				}
				widths[position]++
				distributed++
			}
			free = 0
		}
		lead, between := justify(data.Justify, free, len(current.indexes))
		x := inner.X + lead
		for position, index := range current.indexes {
			child := children[index]
			width := widths[position]
			_, height := measureNode(child, inner.Width, inner.Height)
			childY := aligned(y, current.height, height, data.Align)
			if data.Align == 3 {
				height = current.height
				childY = y
			}
			arrangeNode(child, Rect{X: x, Y: childY, Width: width, Height: height}, childClip(parent, inner), parent.style, nil)
			x += width + data.Gap + between
		}
		y += current.height + data.Gap
	}
}

func childClip(i *instance, inner Rect) Rect {
	if data, ok := i.host.Data.(core.BoxData); ok && data.Clip {
		return core.IntersectRect(i.clip, inner)
	}
	return i.clip
}

func arrangeTabs(i *instance, data core.TabsData) {
	header := Rect{X: i.rect.X, Y: i.rect.Y, Width: i.rect.Width, Height: 1}
	panel := Rect{X: i.rect.X, Y: i.rect.Y + 1, Width: i.rect.Width, Height: maxInt(i.rect.Height-1, 0)}
	if data.Orientation == 1 {
		width := 0
		for _, item := range data.Items {
			if value := uitext.Width(item.Label) + tabHorizontalPadding*2; value > width {
				width = value
			}
		}
		header = Rect{X: i.rect.X, Y: i.rect.Y, Width: width, Height: i.rect.Height}
		panel = Rect{X: i.rect.X + width, Y: i.rect.Y, Width: maxInt(i.rect.Width-width, 0), Height: i.rect.Height}
	}
	if len(i.children) > 0 {
		arrangeNode(i.children[0], panel, core.IntersectRect(i.clip, panel), i.style, nil)
	}
	_ = header
}

func arrangeList(i *instance, data core.ListData) {
	viewport := i.rect
	total := 0
	heights := make([]int, len(i.children))
	if len(i.children) == 1 && core.KeyOf(i.children[0].element) == "" {
		arrangeNode(i.children[0], viewport, i.clip, i.style, nil)
		i.listOffset = 0
		return
	}
	for index, child := range i.children {
		_, heights[index] = measureNode(child, viewport.Width, viewport.Height)
		total += heights[index]
		if index > 0 {
			total += data.Gap
		}
	}
	reserve := data.Scrollbar == 1
	overflow := total > viewport.Height
	if data.Scrollbar == 0 && overflow {
		reserve = true
	}
	content := viewport
	if reserve {
		content.Width = maxInt(content.Width-1, 0)
	}
	maxOffset := maxInt(total-content.Height, 0)
	if i.listOffset > maxOffset {
		i.listOffset = maxOffset
	}
	if i.listOffset < 0 {
		i.listOffset = 0
	}
	selected := ""
	if data.Selectable {
		selected = data.SelectedKey
	}
	selectionChanged := selected != i.listLastSelected
	if selectionChanged {
		i.listOffset = revealListSelection(i, data, heights, content, i.listOffset)
		i.listLastSelected = selected
		i.listManual = false
	}
	if i.listAnchorKey != "" && !i.listManual && !selectionChanged {
		offset, ok := listItemOffset(i, i.listAnchorKey, heights, data.Gap)
		if !ok && len(i.children) > 0 {
			index := clamp(i.listAnchorIndex, 0, len(i.children)-1)
			offset, _ = listItemOffset(i, core.KeyOf(i.children[index].element), heights, data.Gap)
			ok = true
		}
		if ok {
			i.listOffset = clamp(offset-i.listAnchorOffset, 0, maxOffset)
		}
	}
	if i.listOffset > maxOffset {
		i.listOffset = maxOffset
	}
	y := content.Y - i.listOffset
	for index, child := range i.children {
		var override *Style
		if core.KeyOf(child.element) == selected {
			override = &data.SelectedStyle
		}
		arrangeNode(child, Rect{X: content.X, Y: y, Width: content.Width, Height: heights[index]}, core.IntersectRect(i.clip, content), i.style, override)
		y += heights[index] + data.Gap
	}
	rememberListAnchor(i, data, heights)
	i.listManual = false
}

func rememberListAnchor(i *instance, data core.ListData, heights []int) {
	if len(i.children) == 0 || (len(i.children) == 1 && core.KeyOf(i.children[0].element) == "") {
		i.listAnchorKey, i.listAnchorIndex = "", 0
		return
	}
	offset := 0
	for index, child := range i.children {
		if core.KeyOf(child.element) != "" && offset+heights[index] > i.listOffset {
			i.listAnchorKey = core.KeyOf(child.element)
			i.listAnchorOffset = offset - i.listOffset
			i.listAnchorIndex = index
			return
		}
		offset += heights[index] + data.Gap
	}
	i.listAnchorKey, i.listAnchorIndex = "", 0
}

func revealListSelection(i *instance, data core.ListData, heights []int, content Rect, current int) int {
	if !data.Selectable || data.SelectedKey == "" {
		return current
	}
	offset, ok := listItemOffset(i, data.SelectedKey, heights, data.Gap)
	if !ok {
		return current
	}
	height := 0
	for index, child := range i.children {
		if core.KeyOf(child.element) == data.SelectedKey {
			height = heights[index]
			break
		}
	}
	pad := data.ScrollPadding
	if height >= content.Height {
		return offset
	}
	if offset < current+pad {
		current = offset - pad
	}
	if offset+height > current+content.Height-pad {
		current = offset + height - content.Height + pad
	}
	return current
}

func listItemOffset(i *instance, key string, heights []int, gap int) (int, bool) {
	offset := 0
	for index, child := range i.children {
		if core.KeyOf(child.element) == key {
			return offset, true
		}
		offset += heights[index] + gap
	}
	return 0, false
}

func justify(value uint8, free, count int) (int, int) {
	if count <= 1 {
		switch value {
		case 1:
			return free / 2, 0
		case 2:
			return free, 0
		case 4:
			return free / 2, 0
		default:
			return 0, 0
		}
	}
	switch value {
	case 1:
		return free / 2, 0
	case 2:
		return free, 0
	case 3:
		return 0, free / (count - 1)
	case 4:
		return free / (count * 2), free / count
	default:
		return 0, 0
	}
}
func aligned(start, available, size int, value uint8) int {
	switch value {
	case 1:
		return start + maxInt(available-size, 0)/2
	case 2:
		return start + maxInt(available-size, 0)
	default:
		return start
	}
}

func hostStyle(i *instance) Style {
	if i.kind() != core.KindHost {
		return Style{}
	}
	var style Style
	switch data := i.host.Data.(type) {
	case core.BoxData:
		style = data.Style
	case core.TextData:
		style = data.Style
	case core.ButtonData:
		style = data.Style
		if appFocused(i) {
			style, _ = core.ResolveStyle(style, data.FocusStyle)
		}
		if data.Disabled {
			style, _ = core.ResolveStyle(style, data.DisabledStyle)
		}
	case core.InputData:
		style = data.Style
		if appFocused(i) {
			style, _ = core.ResolveStyle(style, data.FocusStyle)
		}
	case core.TabsData:
		style = data.Style
	case core.ListData:
		style = data.Style
	}
	return style
}

func appFocused(i *instance) bool { return i.app != nil && i.app.focused == i }

func paintNode(buffer *screen.Buffer, i *instance) {
	if i == nil || i.clip.Empty() {
		return
	}
	switch i.kind() {
	case core.KindComponent, core.KindProvider, core.KindFragment:
		for _, child := range i.children {
			paintNode(buffer, child)
		}
	case core.KindHost:
		fillHost(buffer, i)
		switch data := i.host.Data.(type) {
		case core.BoxData:
			paintBox(buffer, i, data)
		case core.TextData:
			paintText(buffer, i, data.Content, data.Wrap, data.Align, data.MaxLines, data.Truncate, i.style)
		case core.ButtonData:
			label := data.Label
			if !data.Plain {
				label = "[ " + label + " ]"
			}
			paintText(buffer, i, label, 0, 0, 1, 0, i.style)
		case core.InputData:
			paintInput(buffer, i, data)
		case core.TabsData:
			paintTabs(buffer, i, data)
		case core.ListData:
			paintList(buffer, i, data)
		}
	}
}

func paintBox(buffer *screen.Buffer, i *instance, data core.BoxData) {
	if data.Border != 0 {
		glyphs := []string{"┌", "┐", "└", "┘", "─", "│"}
		if data.Border == 2 {
			glyphs = []string{"╭", "╮", "╰", "╯", "─", "│"}
		}
		if data.Border == 3 {
			glyphs = []string{"╔", "╗", "╚", "╝", "═", "║"}
		}
		if data.Border == 4 {
			glyphs = []string{"┏", "┓", "┗", "┛", "━", "┃"}
		}
		x, y, w, h := i.rect.X, i.rect.Y, i.rect.Width, i.rect.Height
		buffer.Set(x, y, glyphs[0], i.style)
		buffer.Set(x+w-1, y, glyphs[1], i.style)
		buffer.Set(x, y+h-1, glyphs[2], i.style)
		buffer.Set(x+w-1, y+h-1, glyphs[3], i.style)
		for col := x + 1; col < x+w-1; col++ {
			buffer.Set(col, y, glyphs[4], i.style)
			buffer.Set(col, y+h-1, glyphs[4], i.style)
		}
		for row := y + 1; row < y+h-1; row++ {
			buffer.Set(x, row, glyphs[5], i.style)
			buffer.Set(x+w-1, row, glyphs[5], i.style)
		}
	}
	for _, child := range i.children {
		paintNode(buffer, child)
	}
}

func paintText(buffer *screen.Buffer, i *instance, value string, wrap, align uint8, maxLines int, truncate uint8, style Style) {
	width := maxInt(i.rect.Width, 1)
	lines := uitext.Wrap(value, width, wrap == 1)
	if wrap == 2 {
		lines = uitext.Wrap(value, width, false)
	}
	if wrap == 0 {
		lines = strings.Split(value, "\n")
	}
	limited := false
	if maxLines > 0 && len(lines) > maxLines {
		lines = lines[:maxLines]
		limited = true
	}
	for row, line := range lines {
		if row >= i.rect.Height {
			break
		}
		if truncate != 0 {
			line = uitext.Truncate(line, width, true)
			if limited && row == len(lines)-1 {
				line = uitext.Truncate(line, width, true)
			}
		}
		lineWidth := uitext.Width(line)
		x := i.rect.X
		if align == 1 {
			x += maxInt(width-lineWidth, 0) / 2
		} else if align == 2 {
			x += maxInt(width-lineWidth, 0)
		}
		paintGraphemes(buffer, x, i.rect.Y+row, line, style)
	}
}

func paintInput(buffer *screen.Buffer, i *instance, data core.InputData) {
	graphemes := uitext.Graphemes(data.Value)
	i.inputCursor = clamp(i.inputCursor, 0, len(graphemes))
	if i.rect.Width > 0 {
		if i.inputCursor < i.inputOffset {
			i.inputOffset = i.inputCursor
		}
		if i.inputCursor-i.inputOffset >= i.rect.Width {
			i.inputOffset = i.inputCursor - i.rect.Width + 1
		}
	}
	value := data.Value
	if value == "" {
		value = data.Placeholder
	}
	if data.Mask != 0 && data.Value != "" {
		value = strings.Repeat(string(data.Mask), len(graphemes))
	}
	if i.inputOffset > 0 {
		visible := uitext.Graphemes(value)
		if i.inputOffset < len(visible) {
			value = joinGraphemes(visible[i.inputOffset:])
		} else {
			value = ""
		}
	}
	paintText(buffer, i, value, 0, 0, 1, 0, i.style)
	if appFocused(i) && i.rect.Width > 0 {
		cursor := i.rect.X + i.inputCursor - i.inputOffset
		if cursor < i.rect.X {
			cursor = i.rect.X
		}
		if cursor >= i.rect.X+i.rect.Width {
			cursor = i.rect.X + i.rect.Width - 1
		}
		cell := buffer.Cell(cursor, i.rect.Y)
		cell.Style.Attributes |= core.AttributeReverse
		buffer.Set(cursor, i.rect.Y, cell.Grapheme, cell.Style)
	}
}

func paintTabs(buffer *screen.Buffer, i *instance, data core.TabsData) {
	x, y := i.rect.X, i.rect.Y
	for _, item := range data.Items {
		label := item.Label
		style := i.style
		key := item.Key
		if key == data.ActiveKey || (data.ActiveKey == "" && firstEnabledTab(data.Items) == key) {
			style, _ = core.ResolveStyle(style, data.ActiveStyle)
		}
		paintGraphemes(buffer, x, y, " ", style)
		paintGraphemes(buffer, x+tabHorizontalPadding, y, label, style)
		paintGraphemes(buffer, x+tabHorizontalPadding+uitext.Width(label), y, " ", style)
		if data.Orientation == 0 {
			x += uitext.Width(label) + tabHorizontalPadding*2
		} else {
			y++
		}
	}
	for _, child := range i.children {
		paintNode(buffer, child)
	}
}
func firstEnabledTab(items []core.TabData) string {
	for _, item := range items {
		if !item.Disabled {
			return item.Key
		}
	}
	return ""
}

func paintList(buffer *screen.Buffer, i *instance, data core.ListData) {
	for _, child := range i.children {
		paintNode(buffer, child)
	}
	total := 0
	for index, child := range i.children {
		_, height := measureNode(child, i.rect.Width, i.rect.Height)
		total += height
		if index > 0 {
			total += data.Gap
		}
	}
	if data.Scrollbar != 2 && (data.Scrollbar == 1 || total > i.rect.Height) && i.rect.Width > 0 {
		x := i.rect.X + i.rect.Width - 1
		if data.Scrollbar == 1 || total > i.rect.Height {
			for row := i.rect.Y; row < i.rect.Y+i.rect.Height; row++ {
				buffer.Set(x, row, "│", i.style)
			}
			if total > i.rect.Height {
				thumb := maxInt(i.rect.Height*i.rect.Height/maxInt(total, 1), 1)
				start := i.rect.Y + i.listOffset*maxInt(i.rect.Height-thumb, 0)/maxInt(total-i.rect.Height, 1)
				for row := start; row < start+thumb; row++ {
					buffer.Set(x, row, "█", i.style)
				}
			}
		}
	}
}

func paintGraphemes(buffer *screen.Buffer, x, y int, value string, style Style) {
	for _, grapheme := range uitext.Graphemes(value) {
		if x >= buffer.Width {
			return
		}
		buffer.Set(x, y, grapheme, style)
		x += uitext.Width(grapheme)
	}
}

func clamp(value, minValue, maxValue int) int {
	if minValue > 0 && value < minValue {
		value = minValue
	}
	if maxValue > 0 && value > maxValue {
		value = maxValue
	}
	return value
}
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
