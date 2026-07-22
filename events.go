package omnitui

import (
	"fmt"

	"github.com/omnitui/omnitui/v2/internal/backend"
	"github.com/omnitui/omnitui/v2/internal/core"
	uitext "github.com/omnitui/omnitui/v2/internal/text"
)

func (app *App) handleBackendEvent(value any) error {
	switch event := value.(type) {
	case backend.KeyInput:
		keyEvent := KeyEvent{Key: Key(event.Key), Rune: event.Rune, Modifiers: Modifiers(event.Modifiers), Repeat: event.Repeat}
		app.dispatchKey(keyEvent)
	case backend.MouseInput:
		app.dispatchMouse(MouseEvent{Action: MouseAction(event.Action), Button: MouseButton(event.Button), Buttons: MouseButtons(event.Buttons), X: event.X, Y: event.Y, Modifiers: Modifiers(event.Modifiers)})
	case backend.WheelInput:
		app.dispatchWheel(WheelEvent{X: event.X, Y: event.Y, DeltaX: event.DeltaX, DeltaY: event.DeltaY, Modifiers: Modifiers(event.Modifiers)})
	case backend.ResizeInput:
		app.width, app.height = event.Width, event.Height
		app.invalidated = true
		if target := rootHost(app.rootInstance); target != nil {
			dispatchDirect(target, "resize", ResizeEvent{Width: event.Width, Height: event.Height})
		}
	case backend.PasteInput:
		app.dispatchPaste(PasteEvent{Text: event.Text})
	case KeyEvent:
		app.dispatchKey(event)
	case MouseEvent:
		app.dispatchMouse(event)
	case WheelEvent:
		app.dispatchWheel(event)
	case ResizeEvent:
		app.width, app.height = event.Width, event.Height
		app.invalidated = true
	default:
		return fmt.Errorf("omnitui: unsupported backend event %T", value)
	}
	return nil
}

func (app *App) dispatchKey(event KeyEvent) {
	target := app.focused
	if target == nil {
		target = rootHost(app.rootInstance)
	}
	if target == nil {
		if event.Modifiers&ModCtrl != 0 && (event.Rune == 'c' || event.Rune == 'C' || event.Rune == 0) {
			app.interrupted = true
		}
		return
	}
	if dispatchEvent(target, "key", event) {
		return
	}
	if event.Modifiers&ModCtrl != 0 && (event.Rune == 'c' || event.Rune == 'C' || event.Rune == 0) {
		app.interrupted = true
		return
	}
	if event.Key == KeyTab {
		app.moveFocus(false)
		return
	}
	if event.Key == KeyBacktab {
		app.moveFocus(true)
		return
	}
	if event.Key == KeyRune && event.Rune != 0 {
		if dispatchEvent(target, "text", TextInputEvent{Text: string(event.Rune)}) {
			return
		}
		if input := ancestorHost(target, core.HostInput); input != nil {
			app.inputInsert(input, string(event.Rune), ChangeKeyboard)
		}
		return
	}
	if input := ancestorHost(target, core.HostInput); input != nil {
		if event.Key == KeyEnter {
			dispatchDirect(input, "submit", SubmitEvent{Value: inputValue(input)})
			return
		}
		if app.inputKey(input, event) {
			return
		}
	}
	if list := ancestorHost(target, core.HostList); list != nil && app.listKey(list, event) {
		return
	}
	if tabs := ancestorHost(target, core.HostTabs); tabs != nil && app.tabsKey(tabs, event) {
		return
	}
	control := pressTarget(target)
	if (event.Key == KeyEnter || (event.Key == KeyRune && event.Rune == ' ')) && control != nil {
		source := KeyboardEnter
		if event.Key == KeyRune {
			source = KeyboardSpace
		}
		dispatchEvent(control, "press", PressEvent{Source: source})
	}
}

func (app *App) dispatchPaste(event PasteEvent) {
	target := app.focused
	if target == nil {
		return
	}
	if dispatchEvent(target, "paste", event) {
		return
	}
	if input := ancestorHost(target, core.HostInput); input != nil {
		app.inputInsert(input, event.Text, ChangePaste)
	}
}

func (app *App) inputKey(input *instance, event KeyEvent) bool {
	data, ok := input.host.Data.(core.InputData)
	if !ok || data.Disabled {
		return true
	}
	graphemes := uitext.Graphemes(data.Value)
	input.inputCursor = clamp(input.inputCursor, 0, len(graphemes))
	switch event.Key {
	case KeyLeft:
		if input.inputCursor > 0 {
			input.inputCursor--
		}
		app.invalidated = true
		return true
	case KeyRight:
		if input.inputCursor < len(graphemes) {
			input.inputCursor++
		}
		app.invalidated = true
		return true
	case KeyHome:
		input.inputCursor = 0
		app.invalidated = true
		return true
	case KeyEnd:
		input.inputCursor = len(graphemes)
		app.invalidated = true
		return true
	case KeyBackspace:
		if input.inputCursor > 0 && !data.ReadOnly {
			next := append([]string(nil), graphemes[:input.inputCursor-1]...)
			next = append(next, graphemes[input.inputCursor:]...)
			input.inputCursor--
			app.emitInputChange(input, joinGraphemes(next), ChangeKeyboard)
		}
		return true
	case KeyDelete:
		if input.inputCursor < len(graphemes) && !data.ReadOnly {
			next := append([]string(nil), graphemes[:input.inputCursor]...)
			next = append(next, graphemes[input.inputCursor+1:]...)
			app.emitInputChange(input, joinGraphemes(next), ChangeKeyboard)
		}
		return true
	default:
		return false
	}
}

func (app *App) inputInsert(input *instance, value string, source ChangeSource) {
	data, ok := input.host.Data.(core.InputData)
	if !ok || data.Disabled || data.ReadOnly {
		return
	}
	current := uitext.Graphemes(data.Value)
	insertion := uitext.Graphemes(value)
	cursor := clamp(input.inputCursor, 0, len(current))
	next := append([]string(nil), current[:cursor]...)
	next = append(next, insertion...)
	next = append(next, current[cursor:]...)
	if data.MaxLength > 0 && len(next) > data.MaxLength {
		next = next[:data.MaxLength]
	}
	input.inputCursor = clamp(cursor+len(insertion), 0, len(next))
	app.emitInputChange(input, joinGraphemes(next), source)
}

func (app *App) emitInputChange(input *instance, value string, source ChangeSource) {
	previous := inputValue(input)
	if handler, ok := input.handlerValue("change").(EventHandler[ValueChangeEvent]); ok && handler != nil {
		handler(ValueChangeEvent{Previous: previous, Value: value, Source: source})
	}
	app.invalidated = true
}
func inputValue(input *instance) string {
	if data, ok := input.host.Data.(core.InputData); ok {
		return data.Value
	}
	return ""
}
func joinGraphemes(values []string) string {
	result := ""
	for _, value := range values {
		result += value
	}
	return result
}

func (app *App) listKey(list *instance, event KeyEvent) bool {
	data, ok := list.host.Data.(core.ListData)
	if !ok || data.Disabled || !data.Selectable {
		return true
	}
	if len(list.children) == 0 {
		return true
	}
	index := -1
	for i, child := range list.children {
		if core.KeyOf(child.element) == data.SelectedKey {
			index = i
			break
		}
	}
	next := index
	switch event.Key {
	case KeyUp:
		next--
	case KeyDown:
		next++
	case KeyHome:
		next = 0
	case KeyEnd:
		next = len(list.children) - 1
	case KeyPageUp:
		next = app.pageIndex(list, index, -1)
	case KeyPageDown:
		next = app.pageIndex(list, index, 1)
	case KeyEnter:
		if index >= 0 {
			dispatchDirect(list, "activate", ActivateEvent{Key: core.KeyOf(list.children[index].element), Source: KeyboardEnter})
		}
		return true
	default:
		return false
	}
	if data.Wrap {
		if next < 0 {
			next = len(list.children) - 1
		}
		if next >= len(list.children) {
			next = 0
		}
	} else {
		next = clamp(next, 0, len(list.children)-1)
	}
	if next < 0 || next >= len(list.children) {
		return true
	}
	key := core.KeyOf(list.children[next].element)
	if key != data.SelectedKey {
		app.emitListChange(list, key)
	}
	return true
}

func (app *App) pageIndex(list *instance, index, direction int) int {
	if index < 0 {
		if direction < 0 {
			return 0
		}
		return len(list.children) - 1
	}
	rows := maxInt(list.rect.Height-1, 1)
	target := index
	consumed := 0
	if direction > 0 {
		for target+1 < len(list.children) && consumed < rows {
			target++
			_, h := measureNode(list.children[target], list.rect.Width, list.rect.Height)
			consumed += h + 1
		}
	} else {
		for target-1 >= 0 && consumed < rows {
			target--
			_, h := measureNode(list.children[target], list.rect.Width, list.rect.Height)
			consumed += h + 1
		}
	}
	return target
}
func (app *App) emitListChange(list *instance, key string) {
	data, _ := list.host.Data.(core.ListData)
	if handler, ok := list.handlerValue("change").(EventHandler[ValueChangeEvent]); ok && handler != nil {
		handler(ValueChangeEvent{Previous: data.SelectedKey, Value: key, Source: ChangeKeyboard})
	}
	app.invalidated = true
}

func (app *App) tabsKey(tabs *instance, event KeyEvent) bool {
	data, ok := tabs.host.Data.(core.TabsData)
	if !ok {
		return true
	}
	focus := tabs.tabFocus
	if focus == "" {
		focus = data.ActiveKey
		if focus == "" {
			focus = firstEnabledTab(data.Items)
		}
	}
	index := -1
	for i, item := range data.Items {
		if item.Key == focus {
			index = i
			break
		}
	}
	direction := 0
	if data.Orientation == 0 {
		if event.Key == KeyLeft {
			direction = -1
		}
		if event.Key == KeyRight {
			direction = 1
		}
	} else {
		if event.Key == KeyUp {
			direction = -1
		}
		if event.Key == KeyDown {
			direction = 1
		}
	}
	if direction != 0 {
		for {
			index += direction
			if index < 0 || index >= len(data.Items) {
				break
			}
			if !data.Items[index].Disabled {
				tabs.tabFocus = data.Items[index].Key
				app.invalidated = true
				break
			}
		}
		return true
	}
	if event.Key == KeyEnter || (event.Key == KeyRune && event.Rune == ' ') {
		for _, item := range data.Items {
			if item.Key == focus && !item.Disabled {
				app.emitTabsChange(tabs, focus)
				break
			}
		}
		return true
	}
	return false
}
func (app *App) emitTabsChange(tabs *instance, key string) {
	data, _ := tabs.host.Data.(core.TabsData)
	if handler, ok := tabs.handlerValue("change").(EventHandler[ValueChangeEvent]); ok && handler != nil {
		handler(ValueChangeEvent{Previous: data.ActiveKey, Value: key, Source: ChangeKeyboard})
	}
	app.invalidated = true
}

func (app *App) dispatchMouse(event MouseEvent) {
	actual := app.targetAt(event.X, event.Y)
	path := eventPath(actual)
	app.updateHover(path, event.X, event.Y)
	target := actual
	if app.capture != nil && event.Action != MouseEnter && event.Action != MouseLeave {
		target = app.capture
	}
	if target == nil {
		return
	}
	consumed := dispatchEvent(target, "mouse", event)
	if event.Action == MouseDown && event.Button == MouseButtonLeft {
		if !consumed {
			if focus := focusTarget(target); focus != nil {
				app.setFocus(focus, ProgrammaticFocus)
			}
			if input := ancestorHost(target, core.HostInput); input != nil {
				app.positionInputCursor(input, event.X-input.rect.X)
			}
			app.capture = target
			app.pressTarget = pressTarget(target)
		}
	}
	if event.Action == MouseUp {
		if app.pressTarget != nil && event.Button == MouseButtonLeft && !consumed && samePressControl(app.pressTarget, actual) {
			dispatchEvent(app.pressTarget, "press", PressEvent{Source: MouseLeft})
		}
		if app.capture != nil && event.Button == MouseButtonLeft {
			app.capture = nil
			app.pressTarget = nil
		}
	}
	if !consumed && !app.defaultMouse(target, event) {
		return
	}
}

func (app *App) positionInputCursor(input *instance, x int) {
	data, ok := input.host.Data.(core.InputData)
	if !ok {
		return
	}
	graphemes := uitext.Graphemes(data.Value)
	input.inputCursor = clamp(input.inputOffset, 0, len(graphemes))
	used := 0
	for index := input.inputOffset; index < len(graphemes); index++ {
		width := uitext.Width(graphemes[index])
		if x < used+maxInt(width, 1)/2 {
			input.inputCursor = index
			return
		}
		used += width
		input.inputCursor = index + 1
	}
}

func (app *App) defaultMouse(target *instance, event MouseEvent) bool {
	for _, node := range eventPath(target) {
		if node.hostKind() == core.HostTabs && event.Action == MouseDown && event.Button == MouseButtonLeft {
			if key := app.tabAt(node, event.X, event.Y); key != "" {
				app.emitTabsChange(node, key)
				return true
			}
		}
		if node.hostKind() == core.HostList && event.Action == MouseDown && event.Button == MouseButtonLeft {
			data, ok := node.host.Data.(core.ListData)
			if !ok || !data.Selectable {
				continue
			}
			if item := app.listItemAt(node, event.X, event.Y); item != nil {
				app.emitListChange(node, core.KeyOf(item.element))
				app.setFocus(node, ProgrammaticFocus)
				return true
			}
		}
	}
	return true
}

func samePressControl(a, b *instance) bool {
	if a == nil || b == nil {
		return false
	}
	if a == b {
		return true
	}
	for current := b; current != nil; current = current.parent {
		if current == a {
			return true
		}
	}
	return false
}
func pressTarget(target *instance) *instance {
	for current := target; current != nil; current = current.parent {
		if current.pressable() {
			return current
		}
	}
	return nil
}
func focusTarget(target *instance) *instance {
	for current := target; current != nil; current = current.parent {
		if current.focusable() {
			return current
		}
	}
	return nil
}

func (app *App) dispatchWheel(event WheelEvent) {
	target := app.targetAt(event.X, event.Y)
	if target == nil {
		return
	}
	consumed := dispatchEvent(target, "wheel", event)
	if consumed {
		return
	}
	for current := target; current != nil; current = current.parent {
		if current.hostKind() == core.HostList {
			app.scrollList(current, event.DeltaY)
			return
		}
	}
}
func (app *App) scrollList(list *instance, delta int) {
	data, ok := list.host.Data.(core.ListData)
	if !ok {
		return
	}
	total := 0
	for index, child := range list.children {
		_, height := measureNode(child, list.rect.Width, list.rect.Height)
		total += height
		if index > 0 {
			total += data.Gap
		}
	}
	maxOffset := maxInt(total-list.rect.Height, 0)
	next := clamp(list.listOffset+delta, maxOffset*-1, maxOffset)
	if next == list.listOffset {
		return
	}
	if next < 0 {
		next = 0
	}
	list.listOffset = next
	list.listManual = true
	list.listAnchorKey = ""
	app.invalidated = true
}

func dispatchEvent[E any](target *instance, name string, event E) bool {
	for current := target; current != nil; current = current.parent {
		if handler, ok := current.handlerValue(name).(EventHandler[E]); ok && handler != nil {
			if withLocal, ok := any(event).(MouseEvent); ok {
				withLocal.LocalX = withLocal.X - current.rect.X
				withLocal.LocalY = withLocal.Y - current.rect.Y
				event = any(withLocal).(E)
			}
			if withWheel, ok := any(event).(WheelEvent); ok {
				withWheel.LocalX = withWheel.X - current.rect.X
				withWheel.LocalY = withWheel.Y - current.rect.Y
				event = any(withWheel).(E)
			}
			if handler(event) == Consume {
				return true
			}
		}
	}
	return false
}
func dispatchDirect[E any](target *instance, name string, event E) {
	if handler, ok := target.handlerValue(name).(EventHandler[E]); ok && handler != nil {
		handler(event)
	}
}

func ancestorHost(target *instance, kind core.HostKind) *instance {
	for current := target; current != nil; current = current.parent {
		if current.kind() == core.KindHost && current.host.Kind == kind {
			return current
		}
	}
	return nil
}
func (app *App) targetAt(x, y int) *instance {
	var best *instance
	var visit func(*instance)
	visit = func(current *instance) {
		if current == nil || current.clip.Empty() || !current.clip.Contains(x, y) {
			return
		}
		if current.kind() == core.KindHost && current.rect.Contains(x, y) {
			best = current
		}
		for _, child := range current.children {
			visit(child)
		}
	}
	visit(app.rootInstance)
	return best
}
func (app *App) listItemAt(list *instance, x, y int) *instance {
	for _, child := range list.children {
		if child.rect.Contains(x, y) && child.clip.Contains(x, y) {
			return child
		}
	}
	return nil
}
func (app *App) tabAt(tabs *instance, x, y int) string {
	data, ok := tabs.host.Data.(core.TabsData)
	if !ok {
		return ""
	}
	cursor := tabs.rect.X
	if data.Orientation == 1 {
		for index, item := range data.Items {
			width := uitext.Width(item.Label) + tabHorizontalPadding*2
			if y == tabs.rect.Y+index && x >= tabs.rect.X && x < tabs.rect.X+width && !item.Disabled {
				return item.Key
			}
		}
		return ""
	}
	if y != tabs.rect.Y {
		return ""
	}
	for _, item := range data.Items {
		width := uitext.Width(item.Label) + tabHorizontalPadding*2
		if x >= cursor && x < cursor+width && !item.Disabled {
			return item.Key
		}
		cursor += width
	}
	return ""
}

func (app *App) updateHover(path []*instance, x, y int) {
	old := app.hoverPath
	common := 0
	for common < len(old) && common < len(path) && old[common] == path[common] {
		common++
	}
	for index := len(old) - 1; index >= common; index-- {
		node := old[index]
		if handler, ok := node.handlerValue("mouse").(EventHandler[MouseEvent]); ok && handler != nil {
			handler(MouseEvent{Action: MouseLeave, X: x, Y: y, LocalX: x - node.rect.X, LocalY: y - node.rect.Y})
		}
	}
	for index := common; index < len(path); index++ {
		node := path[index]
		if handler, ok := node.handlerValue("mouse").(EventHandler[MouseEvent]); ok && handler != nil {
			handler(MouseEvent{Action: MouseEnter, X: x, Y: y, LocalX: x - node.rect.X, LocalY: y - node.rect.Y})
		}
	}
	app.hoverPath = path
}

func (app *App) recoverFocus() {
	if app.focused == nil && app.focusLost {
		values := app.focusables()
		if len(values) > 0 {
			app.setFocus(values[0], ElementRemoved)
		}
		app.focusLost = false
		return
	}
	if app.focused != nil && app.focused.mounted && app.focused.focusable() && !app.focused.clip.Empty() {
		return
	}
	if app.focused != nil {
		previous := app.focused
		values := app.focusables()
		var next *instance
		for _, candidate := range values {
			if candidate != previous {
				next = candidate
				break
			}
		}
		app.setFocus(next, ElementRemoved)
	}
}
func (app *App) focusables() []*instance {
	var result []*instance
	var visit func(*instance)
	visit = func(i *instance) {
		if i == nil || i.clip.Empty() {
			return
		}
		if i.focusable() {
			result = append(result, i)
		}
		for _, child := range i.children {
			visit(child)
		}
	}
	visit(app.rootInstance)
	return result
}
func (app *App) moveFocus(backward bool) {
	values := app.focusables()
	if len(values) == 0 {
		return
	}
	index := -1
	for i, value := range values {
		if value == app.focused {
			index = i
			break
		}
	}
	if backward {
		if index <= 0 {
			index = len(values) - 1
		} else {
			index--
		}
	} else {
		index = (index + 1) % len(values)
	}
	cause := ForwardTraversal
	if backward {
		cause = BackwardTraversal
	}
	app.setFocus(values[index], cause)
}
func (app *App) setFocus(next *instance, cause FocusCause) {
	if app.focused == next {
		return
	}
	if app.focused != nil {
		if binding := focusBindingOf(app.focused); binding != nil {
			binding.focused.Store(false)
		}
		dispatchDirect(app.focused, "blur", BlurEvent{Cause: cause})
	}
	app.focused = next
	if next != nil {
		if binding := focusBindingOf(next); binding != nil {
			binding.focused.Store(true)
		}
		dispatchDirect(next, "focus", FocusEvent{Cause: cause})
	}
	app.invalidated = true
}
