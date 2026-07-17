---
name: omnitui-interface-builder
description: Build, extend, style, and validate terminal user interfaces with the OmniTUI Go framework. Use when creating or modifying OmniTUI components, examples, dashboards, forms, menus, tabs, lists, interactive terminal screens, layout composition, event handling, or terminal styling in a Go repository that uses github.com/viniciusfonseca/omnitui.
---

# OmniTUI Interface Builder

Build terminal interfaces as small, declarative component trees. Favor explicit state, controlled inputs, predictable layout, accessible keyboard focus, and styles that remain legible in ANSI terminals.

## Workflow

### 1. Inspect the repository

- Read `go.mod` to confirm the module path and Go version.
- Read `README.md` and the relevant files in `docs/`, especially `docs/API.md` and `docs/COMPONENTS.md`.
- Inspect existing examples before adding a new pattern. Reuse their package aliases and `main`/`Run` structure.
- Use `rg` to find the actual exported API instead of guessing names or signatures.

### 2. Define the screen state

Keep application state in the user component, not inside props or package-level mutable variables.

```go
type screenState struct {
	ActiveTab string
	Selected  string
	Query     string
	Notice    string
}

type screen struct{}

func (screen) InitialState(string) screenState {
	return screenState{ActiveTab: "overview", Selected: "first"}
}
```

Use `omnitui.UpdateState(ctx, func(current S) S { ... })` in event handlers. Return `omnitui.Consume` after handling an event; return `omnitui.Propagate` only when an ancestor or default behavior should also receive it.

### 3. Compose the tree

Start with one structural container and divide the screen into named regions:

- Use `Box` when direction, borders, clipping, focus, or lower-level layout control is needed.
- Use `Column` for vertical sections such as header, content, and footer.
- Use `Row` for toolbars, form fields, metric cards, and side-by-side panels.
- Use `Text` for labels, descriptions, status lines, and empty states.
- Use `Input` for controlled single-line editing.
- Use `Button` for explicit actions.
- Use `Tabs` for mutually exclusive panels.
- Use `List` for keyed selection and scrolling.

Keep a predictable hierarchy: outer surface → header → navigation/content → controls → status. Extract repeated regions into helper functions that accept state and context.

### 4. Make interactive components controlled

`Input`, `Tabs`, and `List` receive their public value from props. Their events propose a new value; accept it by updating the parent state.

```go
components.Tabs(components.TabsProps{
	ActiveKey: state.ActiveTab,
	Items: []components.TabItem{
		{Key: "overview", Label: "Overview", Content: overviewPanel()},
		{Key: "logs", Label: "Logs", Content: logsPanel()},
	},
	OnChange: func(event omnitui.ValueChangeEvent) omnitui.EventResult {
		omnitui.UpdateState(ctx, func(current screenState) screenState {
			current.ActiveTab = event.Value
			return current
		})
		return omnitui.Consume
	},
})
```

Apply the same pattern to `Input.OnChange`, `Input.OnSubmit`, `List.OnChange`, and `List.OnActivate`. Give every direct `List` child a unique, stable `.WithKey(...)`; use stable tab keys as well.

### 5. Add visual hierarchy

Define a small style vocabulary instead of styling every element independently:

```go
var (
	surfaceStyle = omnitui.Style{
		Foreground: omnitui.RGB(225, 231, 239),
		Background: omnitui.RGB(15, 20, 29),
	}
	mutedStyle = omnitui.Style{
		Foreground: omnitui.ANSI(omnitui.BrightBlack),
		Attributes: omnitui.Dim,
	}
	accentStyle = omnitui.Style{
		Foreground: omnitui.ANSI(omnitui.BrightCyan),
		Attributes: omnitui.Bold,
	}
)
```

Use style roles such as surface, panel, accent, muted, selected, focus, success, warning, and error. Apply `FocusStyle` to every interactive field or button. Ensure focused text remains readable against its focused background.

### 6. Validate and polish

Run the narrowest relevant checks while iterating, then run the full suite:

```bash
gofmt -w path/to/changed.go
go test ./...
go test -race ./...
go vet ./...
go build ./examples/...
```

If the interface is interactive, exercise keyboard traversal, Enter/Space activation, mouse clicks, wheel scrolling, input submission, resize, and Ctrl+C. Check that changing state does not reset focus or selection unexpectedly.

## Layout rules

- Set `Gap` and `Padding` deliberately; use `omnitui.All(n)` for equal inset and `omnitui.XY(horizontal, vertical)` for asymmetric spacing.
- Use `omnitui.Cells(n)` for stable terminal dimensions and `omnitui.Auto()` when content should determine size.
- Keep fixed-width sibling panels narrow enough for common terminal widths. Avoid adding several unconstrained fixed-width columns in one `Row`.
- Use `Align` for the cross-axis and `Justify` for the main axis. Prefer `AlignCenter` for compact controls and `JustifySpaceBetween` for toolbars.
- Use `WrapWord` for prose and `TextAlignCenter` only for short labels or intentionally centered copy.
- Use `MaxLines` and `TruncateEllipsis` for status text that must not grow a panel unexpectedly.
- Set `Clip: true` on bounded panels when overflowing children must not paint outside their rectangle.
- Prefer one outer `Box` with a border and background over many nested borders. Use nested borders only when they clarify distinct regions.
- Design for the smallest reasonable terminal first; a clipped but coherent layout is better than a screen that requires a wide terminal to be usable.

## Component contracts

| Need | Component | Important rules |
|---|---|---|
| Structural container | `Box` | Configure `Direction`, `Padding`, `Gap`, `Border`, `Clip`, and optional handlers. |
| Horizontal group | `Row` | Children share a horizontal axis; use `Align` and `Justify`. |
| Vertical group | `Column` | Children share a vertical axis; use `Gap` and `Padding`. |
| Static content | `Text` | Leaf component; use wrapping and truncation for prose. |
| Text editing | `Input` | Controlled by `Value`; accept changes through `OnChange`. |
| Action | `Button` | Handle `OnPress`; define both `Style` and `FocusStyle`. |
| Panel navigation | `Tabs` | Use unique keys, controlled `ActiveKey`, and an `OnChange` handler. |
| Selectable viewport | `List` | Key every item, control `SelectedKey`, and use `OnActivate` for Enter. |

`Box` and `Button` are lower-level building blocks; `Row`, `Column`, `Text`, `Input`, `Tabs`, and `List` are the usual application-level components.

## Interaction patterns

### Forms

- Keep one state field per controlled input.
- Update state in `OnChange`; use `OnSubmit` for commit actions.
- Put the label, input, and action in a `Row` or a compact `Column`.
- Render validation or submission feedback as a separate `Text` with a semantic style.
- Avoid mutating state in `Render`.

### Tabs

- Keep `ActiveKey` in parent state.
- Use short labels and stable keys.
- Keep each panel’s content inside a `Column` or `Box` so padding and background are explicit.
- Style inactive and active headers separately.
- Verify that clicking a tab header selects the tab without treating the content row as part of the header hitbox.

### Lists

- Use stable keys that identify the underlying item, not its current index.
- Set a finite `Height` when demonstrating scrolling.
- Use `SelectedStyle` with a strong foreground/background contrast.
- Keep `OnChange` responsible for selection and `OnActivate` responsible for opening or confirming an item.
- Use `ScrollPadding` to keep keyboard selection visible and `ScrollbarAuto` or `ScrollbarAlways` when the viewport benefits from an indicator.

### Buttons and focus

- Make the label describe the result: `Save`, `Open`, `Retry`, or `Delete`.
- Update local state in `OnPress` and consume the event.
- Provide a visible `FocusStyle`; do not rely on color alone, especially with ANSI16 output.
- Place focusable controls in a logical order that matches the visual order.

## Styling reference

Use the public constructors for terminal colors:

- `omnitui.ANSI(omnitui.BrightCyan)` for portable ANSI colors.
- `omnitui.Indexed(45)` for a 256-color palette entry when the color profile supports it.
- `omnitui.RGB(45, 84, 125)` for true-color accents.
- `omnitui.DefaultColor()` when a foreground or background should inherit the terminal default.

Use attributes intentionally: `Bold`, `Dim`, `Italic`, `Underline`, `Blink`, `Reverse`, `Hidden`, and `Strikethrough`. Avoid combining `Attributes` and `ClearAttributes` for the same bit. Validate that foreground/background choices remain legible in ANSI16, ANSI256, and true-color profiles.

## Canonical screen skeleton

Use this shape as a starting point and replace the regions with product-specific content:

```go
func (screen) Render(ctx omnitui.Context, _ string, state screenState, _ omnitui.Children) omnitui.Element {
	return components.Box(
		components.BoxProps{
			Direction: components.Vertical,
			Padding:   omnitui.All(1),
			Gap:       1,
			Border:    components.BorderRounded,
			Clip:      true,
			Style:     surfaceStyle,
		},
		components.Text(components.TextProps{Content: "Application title", Style: accentStyle}),
		navigation(ctx, state),
		content(ctx, state),
		footer(state),
	)
}
```

Keep helper functions pure with respect to rendering: accept `ctx` and state, return elements, and perform mutations only inside event handlers.

## Common failure modes

- Do not store mutable state in a component value passed to `omnitui.Define`; mounted occurrences must own independent state.
- Do not pass children to `Text` or `Input`; they are leaves.
- Do not omit keys from `List` items or reuse keys for unrelated items.
- Do not make `Tabs` or `List` appear interactive without wiring the corresponding controlled state handler.
- Do not update state during `Render`.
- Do not use negative dimensions, gaps, padding, or `MaxLines`.
- Do not create a giant `Row` of unconstrained content that cannot fit a normal terminal.
- Do not use a border, color, or attribute as the only way to communicate focus or selection.
- Do not assume mouse support is available; every important action must remain keyboard accessible.

## Completion checklist

- [ ] Read the repository API and component documentation.
- [ ] Define explicit component state and stable keys.
- [ ] Build a clear `Box`/`Row`/`Column` hierarchy.
- [ ] Keep `Input`, `Tabs`, and `List` controlled by parent state.
- [ ] Add visible focus and selection styles.
- [ ] Handle keyboard traversal and activation.
- [ ] Exercise mouse and wheel behavior where supported.
- [ ] Verify wrapping, truncation, clipping, and resize behavior.
- [ ] Run `gofmt`, `go test ./...`, `go test -race ./...`, `go vet ./...`, and `go build ./examples/...`.
