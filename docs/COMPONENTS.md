# OmniTUI — builtin components

This document describes the official `Row`, `Column`, `Text`, `Input`, `Tabs`, and `List` components exported by the public `omnitui/components` package. Signatures and props are documented in [API.md](API.md); the rendering model is described in [DESIGN.md](DESIGN.md).

## 1. Conventions

- All builtins receive props.
- `Row`, `Column`, and `List` receive children.
- `Tabs` receives its panels as `Element` values inside `TabItem`.
- `Text` and `Input` are leaves and do not receive children.
- `Input`, `Tabs`, and `List` are controlled: the public value comes from props, and events propose changes to the parent component.
- Internal state stores only interaction details such as the cursor, local focus, and scroll offset.
- Handlers return `omnitui.Propagate` or `omnitui.Consume` according to the event contract.
- `UseFocus` handles attach through the `Focus` prop of `Box`, `Button`, `Input`, `Tabs`, and `List`; a `Box` must also set `Focusable`, and a `List` must set `Selectable`.
- `omnitui.Cells(n)` creates a cell-sized `Size`; `omnitui.Fill()` occupies the space available from the parent; `omnitui.All(n)` creates equal spacing on all four sides.
- Examples use `omnitui` for the core and `components` for builtins; they represent snippets from a `Render` method and omit error handling.

```go
import (
    omnitui "github.com/omnitui/omnitui/v2"
    components "github.com/omnitui/omnitui/v2/components"
)
```

| Component | Purpose | Receives children | Internal state |
|---|---|---:|---|
| `Row` | Horizontal layout | Yes | No |
| `Column` | Vertical layout | Yes | No |
| `Text` | Static text | No | No |
| `Input` | Single-line text editing | No | Cursor and horizontal scroll |
| `Tabs` | Navigation between panels | Via `TabItem.Content` | Focused header |
| `List` | Selectable, scrollable list | Yes | Focus and vertical offset |

## 2. `Row`

Arranges children horizontally. It is implemented on top of `Box` with horizontal direction.

Signature and props: [API.md — `Row`](API.md#row).

By default, it measures the sum of the children’s widths, uses the greatest height, and does not wrap. `Justify` distributes space along the horizontal axis; `Align` positions children along the vertical axis.

### Example

```go
return components.Row(
    components.RowProps{
        Gap:     1,
        Align:   components.AlignCenter,
        Justify: components.JustifyEnd,
    },
    components.Text(components.TextProps{Content: "Save changes?"}),
    components.Button(components.ButtonProps{
        Label: "Cancel",
        OnPress: func(event omnitui.PressEvent) omnitui.EventResult {
            cancel()
            return omnitui.Consume
        },
    }),
    components.Button(components.ButtonProps{
        Label: "Save",
        OnPress: func(event omnitui.PressEvent) omnitui.EventResult {
            save()
            return omnitui.Consume
        },
    }),
)
```

## 3. `Column`

Arranges children vertically. It shares `Row`’s semantics with the axes reversed.

Signature and props: [API.md — `Column`](API.md#column).

`Row` and `Column` have separate props to keep the API readable, even though they convert internally to the same layout structure.

### Example

```go
return components.Column(
    components.ColumnProps{
        Gap:     1,
        Padding: omnitui.All(1),
    },
    components.Text(components.TextProps{Content: "Profile"}),
    components.Text(components.TextProps{Content: "Name: Ada Lovelace"}),
    components.Text(components.TextProps{Content: "Role: Engineer"}),
)
```

## 4. `Text`

Renders static text, does not receive focus, and does not accept children.

Signature and props: [API.md — `Text`](API.md#text).

Content is segmented into graphemes before measurement. `MaxLines == 0` means no limit. Wrapping and truncation respect visual width rather than byte or rune count.

### Example

```go
return components.Text(components.TextProps{
    Content:  "A long description that may occupy more than one line.",
    Wrap:     components.WrapWord,
    MaxLines: 2,
    Truncate: components.TruncateEllipsis,
    Style: omnitui.Style{
        Foreground: omnitui.ANSI(omnitui.Cyan),
        Attributes: omnitui.Bold,
    },
})
```

## 5. `Input`

Controlled, focusable, single-line text field. The displayed value always comes from `Value`; `OnChange` proposes a new value, and the parent must update its state to accept it.

Signature and props: [API.md — `Input`](API.md#input).

Internal state stores the cursor and horizontal offset, never a second copy of `Value`. Insertion, Backspace, Delete, arrow keys, Home, End, paste, and Enter submission are part of the first version. A left click positions the cursor at the nearest visual grapheme; dragging to select text is outside the MVP. `MaxLength` counts graphemes. `Mask` changes painting only.

### Example

```go
type FormState struct {
    Name      string
    Submitted string
}

func renderNameInput(ctx omnitui.Context, state FormState) omnitui.Element {
    return components.Input(components.InputProps{
        Value:       state.Name,
        Placeholder: "Enter your name",
        MaxLength:   80,
        OnChange: func(event omnitui.ValueChangeEvent) omnitui.EventResult {
            omnitui.UpdateState(ctx, func(current FormState) FormState {
                current.Name = event.Value
                return current
            })
            return omnitui.Consume
        },
        OnSubmit: func(event omnitui.SubmitEvent) omnitui.EventResult {
            omnitui.UpdateState(ctx, func(current FormState) FormState {
                current.Submitted = event.Value
                return current
            })
            return omnitui.Consume
        },
    })
}
```

## 6. `Tabs`

Displays a tab bar and the active panel. Selection is controlled by `ActiveKey`.

Signature and props: [API.md — `Tabs`](API.md#tabs).

Keys must be unique and stable. `ActiveKey == ""` displays the first enabled tab. A missing or disabled key is a props error. Arrow keys move focus; `Enter`, `Space`, or a left click on a header proposes a new key. Every header includes one terminal column of horizontal padding on both sides, and that padded area is part of the header hit target.

### Example

```go
type ScreenState struct {
    ActiveTab string
}

func renderTabs(ctx omnitui.Context, state ScreenState) omnitui.Element {
    return components.Tabs(components.TabsProps{
        ActiveKey: state.ActiveTab,
        Items: []components.TabItem{
            {
                Key:   "overview",
                Label: "Overview",
                Content: components.Text(components.TextProps{
                    Content: "Project summary",
                }),
            },
            {
                Key:   "logs",
                Label: "Logs",
                Content: components.Text(components.TextProps{
                    Content: "No errors found",
                }),
            },
        },
        OnChange: func(event omnitui.ValueChangeEvent) omnitui.EventResult {
            omnitui.UpdateState(ctx, func(current ScreenState) ScreenState {
                current.ActiveTab = event.Value
                return current
            })
            return omnitui.Consume
        },
    })
}
```

## 7. `List`

Displays children as selectable items in a vertical viewport. Every direct child must have `WithKey`; the key identifies selection and preserves identity during reordering.

Signature and props: [API.md — `List`](API.md#list).

Arrow keys, Home, End, PageUp, and PageDown propose a new selection. `Enter` emits `ActivateEvent`. Internal state maintains focus and scroll offset; `SelectedKey` remains controlled by the parent. `Empty` is rendered when there are no items.

### Scrolling

`List` creates a scrollable viewport when `Height` resolves to a finite height and the combined visual height of the items and gaps exceeds that space. With automatic height, the list grows to fit its items and does not scroll.

The offset is measured in **terminal rows**, not indexes. This supports multi-line items with different heights. Internally, state uses the key of the first visible item and its intra-item offset as an anchor, so inserting or reordering items before the viewport does not cause an unnecessary jump.

`ScrollbarMode` values: [API.md — `List`](API.md#list).

- `ScrollbarAuto` occupies a column only when overflow exists.
- `ScrollbarAlways` reserves a column even when all content fits.
- `ScrollbarHidden` keeps scrolling enabled but does not draw the indicator.
- `ScrollPadding` attempts to keep this number of free rows above and below the selected item. The constraint is relaxed at the edges or when the item is taller than the viewport.

#### Navigation and automatic scrolling

- `Up` and `Down` propose the previous or next item. When the parent accepts the proposal in `SelectedKey`, the list applies the smallest offset needed to make it visible.
- `PageUp` and `PageDown` propose the first eligible item approximately one viewport height above or below, subtracting one overlap row.
- `Home` proposes the first item; when accepted, it moves the viewport to the top.
- `End` proposes the last item; when accepted, it moves the viewport to the bottom.
- With `Wrap`, crossing an edge proposes the opposite edge and, if accepted, moves the viewport there.
- Changing `SelectedKey` externally reveals the corresponding item during the next layout. This is the declarative way to perform programmatic scrolling in the MVP.
- A left click on an item proposes its key as the selection and moves focus to the `List`.
- Wheel input changes the scroll anchor directly without changing `SelectedKey`; selection can therefore be temporarily outside the viewport.
- `OnWheel` runs before the default behavior. `Consume` prevents scrolling; at the upper or lower limit, an unconsumed wheel event propagates to allow nested scrollable containers in the future.

When selection changes, an item is considered visible when its entire rectangle fits in the viewport. If it is taller than the viewport, its first row is aligned to the top and the remainder is clipped. Manual wheel scrolling does not force this invariant until the next `SelectedKey` change.

#### Tree changes and resize

- If items are inserted or reordered, the key anchor preserves the visible region whenever possible.
- If the anchor item disappears, the runtime uses the next item, then the previous item, and finally the top.
- If the selected item disappears, the component proposes the nearest eligible key; it does not silently change `SelectedKey`.
- Resize recalculates the viewport and clamps the offset to the new range. If the selection was visible before resize, it remains visible; a selection already moved away by wheel input does not snap automatically.
- An empty list resets the offset and renders `Empty` without a scrollbar.

All children are reconciled and measured in the MVP, including those outside the viewport; only painting is clipped. Virtualization is deferred until benchmarks and a specific API for on-demand items exist. Wheel input is part of the MVP; scrollbar dragging and inertia remain deferred.

### Example

```go
type ProjectState struct {
    SelectedProject string
    OpenProject     string
}

func renderProjects(ctx omnitui.Context, state ProjectState) omnitui.Element {
    return components.List(
        components.ListProps{
            SelectedKey:   state.SelectedProject,
            Height:        omnitui.Cells(4),
            Wrap:          true,
            ScrollPadding: 1,
            Scrollbar:     components.ScrollbarAuto,
            Empty: components.Text(components.TextProps{
                Content: "No projects",
            }),
            OnChange: func(event omnitui.ValueChangeEvent) omnitui.EventResult {
                omnitui.UpdateState(ctx, func(current ProjectState) ProjectState {
                    current.SelectedProject = event.Value
                    return current
                })
                return omnitui.Consume
            },
            OnActivate: func(event omnitui.ActivateEvent) omnitui.EventResult {
                omnitui.UpdateState(ctx, func(current ProjectState) ProjectState {
                    current.OpenProject = event.Key
                    return current
                })
                return omnitui.Consume
            },
        },
        components.Text(components.TextProps{Content: "OmniTUI"}).WithKey("omnitui"),
        components.Text(components.TextProps{Content: "CLI Tools"}).WithKey("cli-tools"),
        components.Text(components.TextProps{Content: "Experiments"}).WithKey("labs"),
        components.Text(components.TextProps{Content: "Documentation"}).WithKey("docs"),
        components.Text(components.TextProps{Content: "Benchmarks"}).WithKey("benchmarks"),
        components.Text(components.TextProps{Content: "Archive"}).WithKey("archive"),
    )
}
```

## 8. Lower-level building blocks

- `Box`: exported by `components`; a configurable container with direction, size, padding, gap, alignment, border, style, and clipping.
- `Button`: exported by `components`; a focusable control with a label and `OnPress`.
- `Fragment`: belongs to `omnitui` and groups elements without creating layout.
- `None`: belongs to `omnitui` and represents the absence of an element.

`Row` and `Column` should be the usual layout choices. `Box` is available when direction must be dynamic or lower-level capabilities are needed.

### Programmatic focus

Create a focus handle during `Render`, attach it to one focusable control, and request focus from an event handler or effect:

```go
searchFocus := omnitui.UseFocus(ctx, "search")

return components.Column(
    components.ColumnProps{Gap: 1},
    components.Input(components.InputProps{
        Value: state.Query,
        Focus: searchFocus,
    }),
    components.Button(components.ButtonProps{
        Label: "Focus search",
        OnPress: func(omnitui.PressEvent) omnitui.EventResult {
            searchFocus.Request()
            return omnitui.Consume
        },
    }),
)
```

The handle keeps the same binding while its component instance and hook key are preserved. Calling `Blur` releases focus only when its bound host is currently focused.

## 9. Builtin acceptance criteria

1. All are exported by `omnitui/components` and use the core reconciler.
2. Props and children are never mutated internally.
3. `Row` and `Column` preserve the identity and keys of their children.
4. `Text` correctly measures, wraps, and truncates graphemes with variable width.
5. `Input` keeps a valid cursor when `Value` changes externally.
6. `Tabs` validates keys and never activates a disabled tab.
7. `List` preserves its key anchor and selection during insertion and reordering.
8. `List` reveals an item after selection changes and preserves its visibility during resize when it was already visible.
9. `List` correctly clamps its offset with variable-height items, a small viewport, and an empty list.
10. `Input`, `Tabs`, and `List` respond to clicks; `List` responds to wheel input without changing selection.
11. All events follow the ordering and propagation defined in [API.md](API.md).
12. All work with the headless backend and have examples compiled as tests.
