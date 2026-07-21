# OmniTUI — public API reference

This document is the canonical source for the framework’s public API. Internal architecture is described in [DESIGN.md](DESIGN.md), complete builtin examples are in [COMPONENTS.md](COMPONENTS.md), and code organization is in [STRUCTURE.md](STRUCTURE.md).

## 1. Public packages

```go
import (
    omnitui "github.com/omnitui/omnitui"
    components "github.com/omnitui/omnitui/components"
)
```

- `omnitui`: elements, components, state, context, runtime, events, geometry, and styles.
- `omnitui/components`: `Box`, `Row`, `Column`, `Text`, `Button`, `Input`, `Tabs`, and `List`.

The module path is `github.com/omnitui/omnitui`, as defined in `go.mod`.

## 2. Elements and components — `omnitui`

### `Element` and children

```go
type Element struct {
    // opaque representation
}

type Children []Element

func (e Element) WithKey(key string) Element
func None() Element
func Fragment(children ...Element) Element
```

- `Element` is a cheap, immutable description of the interface.
- `WithKey` returns a copy; a key must be unique only among siblings.
- `None` represents the absence of content.
- `Fragment` groups multiple children without creating a layout box.
- The zero value of `Element` is equivalent to `None`.

### Component definition

```go
type Component[P, S any] interface {
    InitialState(props P) S
    Render(ctx Context, props P, state S, children Children) Element
}

type ComponentType[P any] struct {
    // opaque, stable identity
}

func Define[P, S any](name string, component Component[P, S]) ComponentType[P]

func Create[P any](
    component ComponentType[P],
    props P,
    children ...Element,
) Element
```

Rules:

- the value passed to `Define` must not hold mutable state;
- each mounted occurrence of a `ComponentType` has independent state;
- props and children are replaced by the most recent render;
- `Render` must be deterministic for the same inputs and may not update state;
- multiple output children must be wrapped in `Fragment` or a container;
- type, position, and key determine whether an instance is preserved.

## 3. State — `omnitui`

```go
func SetState[S any](ctx Context, next S)

func UpdateState[S any](
    ctx Context,
    update func(current S) S,
)
```

- `SetState` replaces the current state.
- `UpdateState` is preferred when the next value depends on the previous one.
- Updates are queued, applied in order, and grouped into a single frame when possible.
- It is safe to call these functions from another goroutine; only the runtime applies the mutation.
- Updates for an unmounted instance are ignored.
- An incorrect type or an update during `Render` is a programming error.

## 4. Context — `omnitui`

```go
type Context struct {
    // current instance, dispatcher, and inherited values
}

type ContextKey[T any] struct {
    // identity and default value
}

func NewContext[T any](defaultValue T) ContextKey[T]
func UseContext[T any](ctx Context, key ContextKey[T]) T
func Provide[T any](key ContextKey[T], value T, child Element) Element
```

`Context` is the framework’s render context and does not replace `context.Context`. Providers use tree scope: the nearest value wins and does not leak to siblings.

## 5. Runtime — `omnitui`

```go
type ColorProfile uint8

const (
    ColorProfileAuto ColorProfile = iota
    ColorProfileANSI16
    ColorProfileANSI256
    ColorProfileTrueColor
)

type Options struct {
    Input        io.Reader
    Output       io.Writer
    ColorProfile ColorProfile
}

type App struct {
    // runtime opaco
}

func New(root Element, options Options) *App
func (app *App) Run(ctx context.Context) error
func (app *App) UpdateRoot(root Element)
func (app *App) Dispatch(message any)
```

- `Input` and `Output` use the current terminal when omitted.
- `ColorProfileAuto` detects the best available capability; an explicit option makes tests and remote environments predictable.
- `Run` owns the terminal until it returns and restores its state on success, cancellation, error, or panic.
- `UpdateRoot` and `Dispatch` only publish work to the queue and may be called from other goroutines.
- `Dispatch(value)` produces a `MessageEvent` at the root host node.

Public MVP error:

```go
var ErrInterrupted = errors.New("omnitui: interrupted")
```

`Run` returns `ErrInterrupted` when `Ctrl+C` is not consumed.

## 6. Geometry — `omnitui`

```go
type Size struct {
    // opaque mode and value
}

func Auto() Size
func Cells(value int) Size

type Spacing struct {
    Top, Right, Bottom, Left int
}

func All(value int) Spacing
func XY(horizontal, vertical int) Spacing

type Rect struct {
    X, Y, Width, Height int
}
```

- The zero value of `Size` is equivalent to `Auto()`.
- Negative sizes and spacing are invalid.
- Percentages and flexible units are not part of the MVP.

## 7. Styles — `omnitui`

### Colors

```go
type Color struct {
    // opaque kind and channels
}

type ANSIColor uint8

const (
    Black ANSIColor = iota
    Red
    Green
    Yellow
    Blue
    Magenta
    Cyan
    White
    BrightBlack
    BrightRed
    BrightGreen
    BrightYellow
    BrightBlue
    BrightMagenta
    BrightCyan
    BrightWhite
)

func DefaultColor() Color
func ANSI(color ANSIColor) Color
func Indexed(index uint8) Color
func RGB(red, green, blue uint8) Color
```

The zero value of `Color` means **unspecified** and inherits the parent’s resolved color. `DefaultColor()` is different: it explicitly requests the terminal’s default color and stops inheritance.

Supported profiles:

| Profile | Colors |
|---|---:|
| `ColorProfileANSI16` | 16 cores ANSI |
| `ColorProfileANSI256` | paleta indexada de 256 cores |
| `ColorProfileTrueColor` | RGB de 24 bits |

When a color exceeds the active profile, the renderer chooses the visually closest entry in the available palette. Alpha, gradients, and color blending are not supported.

### Text attributes

```go
type AttributeMask uint16

const (
    Bold AttributeMask = 1 << iota
    Dim
    Italic
    Underline
    Blink
    Reverse
    Hidden
    Strikethrough
)

type Style struct {
    Foreground      Color
    Background      Color
    Attributes      AttributeMask
    ClearAttributes AttributeMask
}
```

| Attribute | Expected effect | Notes |
|---|---|---|
| `Bold` | Strong intensity | May select a bright color on older terminals |
| `Dim` | Reduced intensity | Not every terminal distinguishes it from normal color |
| `Italic` | Slanted text | May be ignored by the terminal |
| `Underline` | Simple underline | Widely supported |
| `Blink` | Blinking text | Frequently disabled by terminals |
| `Reverse` | Swaps foreground and background | Widely supported |
| `Hidden` | Hides visual content | Cells still occupy layout space |
| `Strikethrough` | Struck-through text | May be ignored by the terminal |

The framework emits the corresponding SGR codes but does not visually simulate attributes ignored by the terminal. `DoubleUnderline`, `Overline`, hyperlinks, fonts, and blink speed are outside the MVP.

### Inheritance and composition

To resolve a node’s style:

1. start with the parent’s resolved style;
2. replace foreground/background when the color is non-zero;
3. remove the `ClearAttributes` bits;
4. add the `Attributes` bits.

The same bit cannot appear in `Attributes` and `ClearAttributes`; this produces a props error. The zero value of `Style` fully inherits the parent’s style.

```go
titleStyle := omnitui.Style{
    Foreground: omnitui.ANSI(omnitui.Cyan),
    Attributes: omnitui.Bold | omnitui.Underline,
}

normalChild := omnitui.Style{
    ClearAttributes: omnitui.Bold,
}
```

Props such as `FocusStyle`, `ActiveStyle`, and `SelectedStyle` are applied after `Style`, using the same rules. Padding, gap, border, alignment, and clipping are component properties, not `Style` attributes.

## 8. Builtin components — `components`

### Shared enums

```go
type Direction uint8
const (
    Horizontal Direction = iota
    Vertical
)

type Align uint8
const (
    AlignStart Align = iota
    AlignCenter
    AlignEnd
    AlignStretch
)

type Justify uint8
const (
    JustifyStart Justify = iota
    JustifyCenter
    JustifyEnd
    JustifySpaceBetween
    JustifySpaceAround
)

type TextWrap uint8
const (
    WrapNone TextWrap = iota
    WrapWord
    WrapGrapheme
)

type TextAlign uint8
const (
    TextAlignStart TextAlign = iota
    TextAlignCenter
    TextAlignEnd
)

type TruncateMode uint8
const (
    TruncateClip TruncateMode = iota
    TruncateEllipsis
)

type Orientation uint8
const (
    OrientationHorizontal Orientation = iota
    OrientationVertical
)
```

### `Box`

```go
type BorderStyle uint8

const (
    BorderNone BorderStyle = iota
    BorderSingle
    BorderRounded
    BorderDouble
    BorderHeavy
)

type BoxProps struct {
    Width, Height        omnitui.Size
    MinWidth, MaxWidth   int
    MinHeight, MaxHeight int
    Padding              omnitui.Spacing
    Gap                  int
    Direction            Direction
    Align                Align
    Justify              Justify
    Wrap                 bool
    Clip                 bool
    Border               BorderStyle
    Style                omnitui.Style

    Focusable bool
    Disabled  bool

    OnKey       omnitui.EventHandler[omnitui.KeyEvent]
    OnTextInput omnitui.EventHandler[omnitui.TextInputEvent]
    OnPaste     omnitui.EventHandler[omnitui.PasteEvent]
    OnFocus     omnitui.EventHandler[omnitui.FocusEvent]
    OnBlur      omnitui.EventHandler[omnitui.BlurEvent]
    OnPress     omnitui.EventHandler[omnitui.PressEvent]
    OnMouse     omnitui.EventHandler[omnitui.MouseEvent]
    OnWheel     omnitui.EventHandler[omnitui.WheelEvent]
    OnResize    omnitui.EventHandler[omnitui.ResizeEvent]
    OnMessage   omnitui.EventHandler[omnitui.MessageEvent]
}

func Box(props BoxProps, children ...omnitui.Element) omnitui.Element
```

`OnResize` and `OnMessage` are called only when the `Box` is the root host node. When different from `BorderNone`, `Border` occupies one cell on each of the four sides.

### `Row`

```go
type RowProps struct {
    Width, Height        omnitui.Size
    MinWidth, MaxWidth   int
    MinHeight, MaxHeight int
    Padding              omnitui.Spacing
    Gap                  int
    Align                Align
    Justify              Justify
    Wrap                 bool
    Clip                 bool
    Style                omnitui.Style
}

func Row(props RowProps, children ...omnitui.Element) omnitui.Element
```

### `Column`

```go
type ColumnProps struct {
    Width, Height        omnitui.Size
    MinWidth, MaxWidth   int
    MinHeight, MaxHeight int
    Padding              omnitui.Spacing
    Gap                  int
    Align                Align
    Justify              Justify
    Clip                 bool
    Style                omnitui.Style
}

func Column(props ColumnProps, children ...omnitui.Element) omnitui.Element
```

### `Text`

```go
type TextProps struct {
    Content  string
    Style    omnitui.Style
    Wrap     TextWrap
    Align    TextAlign
    MaxLines int
    Truncate TruncateMode
}

func Text(props TextProps) omnitui.Element
```

`MaxLines == 0` means no limit. Wrapping and truncation operate on graphemes and visual width.

### `Button`

```go
type ButtonProps struct {
    Label         string
    Disabled      bool
    Style         omnitui.Style
    FocusStyle    omnitui.Style
    DisabledStyle omnitui.Style

    OnKey   omnitui.EventHandler[omnitui.KeyEvent]
    OnFocus omnitui.EventHandler[omnitui.FocusEvent]
    OnBlur  omnitui.EventHandler[omnitui.BlurEvent]
    OnPress omnitui.EventHandler[omnitui.PressEvent]
    OnMouse omnitui.EventHandler[omnitui.MouseEvent]
}

func Button(props ButtonProps) omnitui.Element
```

`Button` is focusable by default. `Enter`, `Space`, and a complete left click produce `PressEvent` when enabled.

### `Input`

```go
type InputProps struct {
    Value       string
    Placeholder string
    Width       omnitui.Size
    Disabled    bool
    ReadOnly    bool
    Mask        rune
    MaxLength   int
    Style       omnitui.Style
    FocusStyle  omnitui.Style

    OnChange    omnitui.EventHandler[omnitui.ValueChangeEvent]
    OnSubmit    omnitui.EventHandler[omnitui.SubmitEvent]
    OnKey       omnitui.EventHandler[omnitui.KeyEvent]
    OnTextInput omnitui.EventHandler[omnitui.TextInputEvent]
    OnPaste     omnitui.EventHandler[omnitui.PasteEvent]
    OnFocus     omnitui.EventHandler[omnitui.FocusEvent]
    OnBlur      omnitui.EventHandler[omnitui.BlurEvent]
    OnMouse     omnitui.EventHandler[omnitui.MouseEvent]
}

func Input(props InputProps) omnitui.Element
```

`Input` is controlled by `Value`; `OnChange` only proposes a new value. `MaxLength` counts graphemes. `Mask` changes painting only. A left click positions the cursor at the nearest visual grapheme.

### `Tabs`

```go
type TabItem struct {
    Key      string
    Label    string
    Content  omnitui.Element
    Disabled bool
}

type TabsProps struct {
    Items       []TabItem
    ActiveKey   string
    Orientation Orientation
    Style       omnitui.Style
    ActiveStyle omnitui.Style
    OnChange    omnitui.EventHandler[omnitui.ValueChangeEvent]
}

func Tabs(props TabsProps) omnitui.Element
```

Keys must be unique. `ActiveKey == ""` uses the first enabled tab; a missing or disabled key is a props error. A left click on a header proposes its key through `OnChange`.

### `List`

```go
type ScrollbarMode uint8

const (
    ScrollbarAuto ScrollbarMode = iota
    ScrollbarAlways
    ScrollbarHidden
)

type ListProps struct {
    SelectedKey   string
    Height        omnitui.Size
    Gap           int
    Disabled      bool
    Wrap          bool
    ScrollPadding int
    Scrollbar     ScrollbarMode
    Empty         omnitui.Element
    Style         omnitui.Style
    SelectedStyle omnitui.Style

    OnChange   omnitui.EventHandler[omnitui.ValueChangeEvent]
    OnActivate omnitui.EventHandler[omnitui.ActivateEvent]
    OnMouse    omnitui.EventHandler[omnitui.MouseEvent]
    OnWheel    omnitui.EventHandler[omnitui.WheelEvent]
}

func List(props ListProps, items ...omnitui.Element) omnitui.Element
```

Every direct item must have `WithKey`. `List` is controlled by `SelectedKey`; a left click proposes the item and wheel input moves only the viewport. Detailed scrolling and navigation are in [COMPONENTS.md](COMPONENTS.md#scrolling).

## 9. Events — `omnitui`

### Handler contract

```go
type EventResult uint8

const (
    Propagate EventResult = iota
    Consume
)

type EventHandler[E any] func(event E) EventResult
```

- `Propagate` sends the event to the next eligible ancestor.
- `Consume` stops bubbling and prevents the associated default behavior.
- A missing handler is equivalent to `Propagate`.
- Handlers run on the runtime goroutine and should return quickly.
- Events are immutable values for consumers.

For events without propagation, the return value is ignored, but the signature remains uniform.

### Events supported in the MVP

| Event | Source | Initial target | Propagation |
|---|---|---|---|
| `KeyEvent` | Key or ANSI sequence | Focused element | Target through ancestors |
| `TextInputEvent` | Normalized printable input | Focused `Input` | Target through ancestors |
| `PasteEvent` | Bracketed paste | Focused `Input` | Target through ancestors |
| `MouseEvent` | Pointer movement, button, enter, or leave | Host under pointer or capture target | Target through ancestors, except enter/leave |
| `WheelEvent` | Terminal wheel or scroll gesture | Host under pointer | Target through ancestors |
| `FocusEvent` | Element receives focus | New focus | Does not propagate |
| `BlurEvent` | Element loses focus | Previous focus | Does not propagate |
| `PressEvent` | Control activation | Pressable `Button` or `Box` | Target through ancestors |
| `ValueChangeEvent` | `Input`, `Tabs`, or `List` proposes a value | Emitting builtin | Does not propagate |
| `SubmitEvent` | `Enter` in an `Input` | Focused `Input` | Does not propagate |
| `ActivateEvent` | `Enter` on a `List` item | Focused `List` | Does not propagate |
| `ResizeEvent` | Terminal dimensions change | Root host | Does not propagate |
| `MessageEvent` | `App.Dispatch` | Root host | Does not propagate |

### Keyboard

```go
type Key uint16

const (
    KeyRune Key = iota
    KeyEnter
    KeyEscape
    KeyTab
    KeyBacktab
    KeyBackspace
    KeyDelete
    KeyInsert
    KeyUp
    KeyDown
    KeyLeft
    KeyRight
    KeyHome
    KeyEnd
    KeyPageUp
    KeyPageDown
    KeyF1
    KeyF2
    KeyF3
    KeyF4
    KeyF5
    KeyF6
    KeyF7
    KeyF8
    KeyF9
    KeyF10
    KeyF11
    KeyF12
)

type Modifiers uint8

const (
    ModCtrl Modifiers = 1 << iota
    ModAlt
    ModShift
)

type KeyEvent struct {
    Key       Key
    Rune      rune
    Modifiers Modifiers
    Repeat    bool
}
```

Terminals do not provide portable `KeyUp` events. Modifiers and repetition are reported only when distinguishable in the received input.

Default behavior after an unconsumed `KeyEvent`:

- `Tab` and `Backtab` move focus;
- `Enter` and `Space` generate `PressEvent` on pressable controls;
- `Ctrl+C` exits `Run` with `ErrInterrupted`.

### Focus

```go
type FocusCause uint8

const (
    ProgrammaticFocus FocusCause = iota
    ForwardTraversal
    BackwardTraversal
    ElementRemoved
)

type FocusEvent struct {
    Cause FocusCause
}

type BlurEvent struct {
    Cause FocusCause
}
```

`FocusEvent` and `BlurEvent` do not propagate.

### Press

```go
type PressSource uint8

const (
    KeyboardEnter PressSource = iota
    KeyboardSpace
    MouseLeft
    ProgrammaticPress
)

type PressEvent struct {
    Source PressSource
}
```

Disabled controls neither receive nor propagate `PressEvent`.

### Mouse and wheel

```go
type MouseAction uint8

const (
    MouseMove MouseAction = iota
    MouseDown
    MouseUp
    MouseEnter
    MouseLeave
)

type MouseButton uint8

const (
    MouseButtonNone MouseButton = iota
    MouseButtonLeft
    MouseButtonMiddle
    MouseButtonRight
)

type MouseButtons uint8

const (
    MouseLeftPressed MouseButtons = 1 << iota
    MouseMiddlePressed
    MouseRightPressed
)

type MouseEvent struct {
    Action    MouseAction
    Button    MouseButton
    Buttons   MouseButtons
    X, Y      int
    LocalX    int
    LocalY    int
    Modifiers Modifiers
}

type WheelEvent struct {
    X, Y      int
    LocalX    int
    LocalY    int
    DeltaX    int
    DeltaY    int
    Modifiers Modifiers
}
```

- Coordinates are zero-based; `X`/`Y` are screen-relative. The runtime gives each handler a copy with `LocalX`/`LocalY` relative to the node whose handler is running, including during bubbling.
- `MouseMove`, `MouseDown`, `MouseUp`, and `WheelEvent` propagate from the target to its ancestors.
- `MouseEnter` and `MouseLeave` are derived by comparing ancestor paths. Each node that was entered or left receives its own event without bubbling; a `Box` can therefore detect hover even when a child is the deepest target.
- `Buttons` describes buttons held during movement; `Button` identifies the button that changed on down/up.
- `DeltaY < 0` scrolls up and `DeltaY > 0` scrolls down; `DeltaX < 0` scrolls left and `DeltaX > 0` scrolls right.
- The backend normalizes wheel input into logical lines; the magnitude may be greater than 1 when the terminal reports multiple steps.

Without capture, hit testing walks hosts from the last painted to the first, respects accumulated ancestor clipping, and chooses the deepest visible node. The target does not depend on a cell containing a grapheme: the entire layout rectangle participates. During capture, move/up are delivered to the capture target; enter/leave continue to be calculated from the host actually under the pointer.

On `MouseDown`, the target receives automatic capture until the corresponding `MouseUp`. Events continue reaching it even outside its rectangle. Capture is canceled if the target is unmounted or disabled.

Unconsumed default behavior:

- a left click focuses a focusable target;
- a left down followed by an up still inside the same pressable control generates `PressEvent{Source: MouseLeft}`;
- a click on a `Tabs` header proposes its key through `ValueChangeEvent`;
- a click on a `List` item proposes its selection;
- wheel input over a `List` moves the viewport without changing `SelectedKey`.

The MVP enables the SGR extended mouse protocol and motion tracking. Double-click, triple-click, semantic drag, drag selection, and scrollbar manipulation are outside the MVP; applications may still interpret the raw down/move/up sequence.

### Text and paste

```go
type TextInputEvent struct {
    Text string
}

type PasteEvent struct {
    Text string
}
```

For printable input, the runtime delivers `KeyEvent`, then `TextInputEvent`, and, if both allow default behavior, the `Input` proposes `ValueChangeEvent`. Paste remains a single event and respects the backend’s input limit.

### Change, submit, and activation

```go
type ChangeSource uint8

const (
    ChangeKeyboard ChangeSource = iota
    ChangePaste
    ChangeProgrammatic
)

type ValueChangeEvent struct {
    Previous string
    Value    string
    Source   ChangeSource
}

type SubmitEvent struct {
    Value string
}

type ActivateEvent struct {
    Key    string
    Source PressSource
}
```

`ValueChangeEvent` is a proposal: it does not change props automatically. `SubmitEvent` does not clear the input, and `ActivateEvent` does not change selection.

### Resize and messages

```go
type ResizeEvent struct {
    Width  int
    Height int
}

type MessageEvent struct {
    Value any
}
```

Pending resizes may be coalesced to the most recent size. Messages preserve order and are not coalesced.

### Processing order

1. Normalize backend input.
2. Resolve the target by focus, root, mouse capture, or hit testing.
3. Run the target handler.
4. Propagate while the result is `Propagate`.
5. Apply default behavior if the event was not consumed.
6. Drain state updates produced by handlers.
7. Reconcile and produce at most one frame.

### Events outside the MVP

- `TickEvent`;
- `KeyUp`;
- semantic `DoubleClickEvent` and `DragEvent`;
- `Mount`, `Update`, and `Unmount` lifecycle.

Lifecycle will belong to a future effects and cleanup API, not the input event system.

## 10. Handler matrix by component

| Component | Public handlers |
|---|---|
| `Box` | `OnKey`, `OnTextInput`, `OnPaste`, `OnFocus`, `OnBlur`, `OnPress`, `OnMouse`, `OnWheel`, `OnResize`, `OnMessage` |
| `Button` | `OnKey`, `OnFocus`, `OnBlur`, `OnPress`, `OnMouse` |
| `Input` | `OnKey`, `OnTextInput`, `OnPaste`, `OnFocus`, `OnBlur`, `OnMouse`, `OnChange`, `OnSubmit` |
| `Tabs` | `OnChange` |
| `List` | `OnMouse`, `OnWheel`, `OnChange`, `OnActivate` |
| `Row`, `Column`, `Text` | No direct handlers |

To make `Row` or `Column` interactive, use a `Box` configured as focusable or create a composite component that renders an interactive surface.

## 11. Usage errors

The following situations are programming errors and must include the component path:

- duplicate key among siblings;
- incompatible state type;
- state update during `Render`;
- children passed to a leaf component;
- props with negative sizes;
- missing or disabled active tab;
- `List` item without a key;
- attribute present simultaneously in `Attributes` and `ClearAttributes`.

I/O, cancellation, and interruption errors are returned by `App.Run`.
