# OmniTUI — design plan

The public reference for types, functions, styles, and events is in [API.md](API.md).

## 1. Goal

Build a Go TUI framework inspired by the React mental model, in which:

- the interface is a declarative tree of elements;
- components receive props, context, and children;
- every mounted component instance has its own state;
- every component renders exactly one element (a `Fragment` covers multiple children);
- state changes trigger a new render;
- reconciliation preserves or discards state predictably;
- the public `components` package includes `Row`, `Column`, `Text`, `Input`, `Tabs`, and `List` as official builtins;
- the screen is updated by buffer diffing without redrawing the entire terminal.

The first milestone must prove this model with a small API. Animations, concurrent rendering, and a complete imitation of CSS remain outside the MVP; keyed hooks provide lifecycle effects, synchronized refs, viewport reads, and programmatic focus.

## 2. Assumptions and boundaries

1. Supported platforms are Linux, macOS, and Windows terminals with ANSI/VT100 support. Windows uses virtual terminal processing and polls for console resize changes.
2. “From scratch” means not wrapping Bubble Tea, tview, or another TUI framework. Small, focused utilities such as `golang.org/x/term` for raw mode and a Unicode width library are acceptable.
3. All renders and tree mutations will happen on a single runtime-controlled goroutine. Other goroutines may only enqueue messages.
4. Props and elements are treated as immutable values. The runtime owns mounted state.
5. The first implementation will reconcile the entire logical tree after an update. Subtree optimizations will be added only when measurements justify them.
6. Initial layout will be a small flexbox subset: row/column direction, sizing, padding, gap, alignment, and clipping.
7. IME and advanced accessibility will be later extensions; asynchronous work is started by committed effects, and SGR mouse is supported where the terminal provides it.

## 3. Mental model

There are three different representations, and they must not be mixed:

```text
Immutable elements           Mounted instances           Physical output
(what the user wants)    ->  (identity + state)      ->  (cell buffer)

Component / Box / Text       componentInstance            Cell[x,y]
props / key / children       state / hooks / host         rune / width / style
```

- **Element:** cheap, disposable description created during `Render`.
- **Instance:** persistent internal object while type, position, and key remain compatible. This is where state lives.
- **Host node:** result with user components removed, containing only primitives understood by layout.
- **Cell buffer:** final screen representation used to produce minimal ANSI sequences.

This separation avoids storing state inside `Element`, which is recreated on every render, and avoids coupling components directly to the terminal.

## 4. Public API

The canonical reference for the `omnitui` and `omnitui/components` packages is in [API.md](API.md). It contains supported elements, components, state, context, runtime, geometry, styles, builtins, and events.

This document keeps only architectural decisions and examples that help explain the model.

## 5. Intended usage example

The example assumes `omnitui` as the root package name and `components` as the name of the `omnitui/components` package.

```go
type CounterProps struct {
    Label string
}

type CounterState struct {
    Value int
}

type Counter struct{}

func (Counter) InitialState(CounterProps) CounterState {
    return CounterState{}
}

func (Counter) Render(
    ctx omnitui.Context,
    props CounterProps,
    state CounterState,
    children omnitui.Children,
) omnitui.Element {
    return components.Column(
        components.ColumnProps{Gap: 1},
        components.Text(components.TextProps{
            Content: fmt.Sprintf("%s: %d", props.Label, state.Value),
        }),
        components.Button(components.ButtonProps{
            Label: "Increment",
            OnPress: func(event omnitui.PressEvent) omnitui.EventResult {
                omnitui.UpdateState(ctx, func(current CounterState) CounterState {
                    current.Value++
                    return current
                })
                return omnitui.Consume
            },
        }),
        omnitui.Fragment(children...),
    )
}

var CounterType = omnitui.Define[CounterProps, CounterState]("Counter", Counter{})

root := omnitui.Create(
    CounterType,
    CounterProps{Label: "Clicks"},
    components.Text(components.TextProps{
        Content: "Child received by the component",
    }),
)
```

This example defines the minimum contract that the first integration tests must make possible.

## 6. Reconciliation and identity

When it receives a new element tree, the reconciler compares each new element with the previous mounted instance:

1. If type and key are compatible, reuse the instance and its state.
2. If type or key changes, unmount the previous instance and mount a new one with `InitialState`.
3. Without a key, identity is determined by position among siblings.
4. With a key, lookup happens among siblings and allows reordering without losing state.
5. Duplicate keys among siblings produce an error with the tree path.
6. Props and children are always replaced with the values from the most recent render.

For a reused component instance:

1. apply pending state updates;
2. build inherited context;
3. call `Render` with current props, state, and children;
4. reconcile the returned element with the previous subtree.

For a reused host node:

1. update props;
2. reconcile children;
3. produce or update the node used by layout.

Lifecycle is exposed declaratively through keyed hooks. `UseEffect` records intent during `Render`; setup runs only after a successful commit. A changed dependency, a missing hook, unmount, or `App.Run` exit cancels the effect context and invokes cleanup. Refs and focus handles are also scoped to the mounted component instance.

### Pseudocode

```text
reconcile(parent, oldInstance, newElement, inheritedContext):
    if newElement is None:
        unmount(oldInstance)
        return nil

    if identity(oldInstance) != identity(newElement):
        unmount(oldInstance)
        oldInstance = mount(newElement)

    update oldInstance props and children

    if oldInstance is Provider:
        nextContext = inheritedContext.with(key, value)
        reconcile its child with nextContext

    if oldInstance is Component:
        apply queued state updates
        begin keyed hook registration
        output = component.Render(contextFor(oldInstance), props, state, children)
        finish keyed hook registration
        reconcile oldInstance.rendered with output

    if oldInstance is Host:
        reconcileChildren(oldInstance, newElement.children)

    return oldInstance
```

## 7. Rendering pipeline

```text
input / resize / SetState
          |
          v
      runtime queue -- batches updates
          |
          v
  render + reconciliation -- preserves identity and state
          |
          v
      host tree -- only Box, Text, and internal interaction hosts
          |
          v
        layout -- positions, sizes, and clipping
          |
          v
         paint -- back buffer of Cells
          |
          v
     buffer diff -- ANSI sequences for changed cells
          |
          v
        terminal
          |
          v
  committed effect cleanup/setup
```

### Screen buffer

```go
type Cell struct {
    Grapheme string
    Width    int
    Style    Style
}
```

The renderer maintains front and back buffers. After painting, it groups changed cells by row, reduces cursor movements, and emits style changes only when necessary. It swaps the buffers at the end.

Unicode requires handling visual width, combining characters, and continuation cells. This must be covered from the first buffer version, even if the input parser initially supports only basic keyboard input.

## 8. Layout, styles, and builtin components

### 8.1 Component catalog

The `omnitui/components` package provides `Row`, `Column`, `Text`, `Input`, `Tabs`, and `List` as official builtins. Their contracts, props, behavior, and usage examples are in [COMPONENTS.md](COMPONENTS.md).

`Box` and `Button` remain in the same package as lower-level visual building blocks; `Fragment` and `None` belong to the `omnitui` core. `Box` and `Text` create hosts through the opaque `internal/core` boundary; the other builtins use the same `Component` API available to users. `Input` uses an unexported internal host for cursor and editing.

### 8.2 Layout engine

The algorithm has two passes:

1. **Measure, bottom-up:** each node computes its desired size within the received constraints.
2. **Layout, top-down:** the parent distributes space and defines the final rectangle for each child.

Core internal types:

```go
type Constraints struct {
    MinWidth, MaxWidth   int
    MinHeight, MaxHeight int
}

type Rect struct {
    X, Y, Width, Height int
}
```

Generic CSS will not be implemented. Every supported prop will have documented semantics and its own test.

## 9. Events, focus, and concurrency

Public types, handlers, keys, supported events, and propagation rules are in [API.md](API.md).

The main loop is approximately:

```text
for app is running:
    wait for input, resize, external message, state update, or cancellation
    drain queued work
    route events
    reconcile if invalidated
    layout and paint if invalidated or resized
    flush screen diff
```

Only the runtime goroutine touches instances, layout, and buffers. Terminal readers and external producers publish values to channels. Handlers run serially; synchronous updates produced by the same event are drained before reconciliation and appear together in the next frame. Effect setup and cleanup run serially after commit; effect goroutines publish state through the existing dispatcher. `Ref` synchronizes its value without scheduling a frame.

The runtime maintains focus order derived from the visible host tree. Unmounting or disabling the focused node chooses the next valid candidate; resize invalidates layout and paint without reinitializing components.

Mouse events use the already-positioned host tree for hit testing. The search walks paint order from top to bottom, respects clipping rectangles, and chooses the deepest visible node. Mouse down establishes temporary capture until mouse up; unmounting the target cancels capture. The previous path under the pointer is compared with the new path to derive enter and leave without storing state in user components.

## 10. Initial code organization

A complete overview of the directory tree, file responsibilities, dependency direction, and phased creation is in [STRUCTURE.md](STRUCTURE.md).

The code exposes two packages: `omnitui` for the runtime and `omnitui/components` for builtins. `components` depends on the core; the core never imports the catalog. Both share only the opaque element representation in `internal/core`.

Extractions into `internal/` should happen only when a boundary is stable and there is a concrete isolation benefit. The terminal backend is the first likely boundary because it enables a headless test backend and, eventually, Windows support.

Internal interfaces worth defining early:

```go
type Backend interface {
    Size() (width, height int, err error)
    Events() <-chan Event
    Write([]byte) error
    Close() error
}

type Clock interface {
    Now() time.Time
}
```

`Backend` enables deterministic tests. `Clock` should be introduced only when animations or timers actually enter the product.

## 11. Dependencies

Initial dependency budget:

- `golang.org/x/term`: raw mode and terminal size;
- a small grapheme/Unicode-width library, selected after a comparative spike;
- no TUI framework, layout, or state-management dependency.

If “from scratch” must literally mean the standard library only, platform-specific code and custom Unicode tables will be required. This significantly increases cost without improving the component model, so it is not the initial recommendation.

## 12. Error handling and diagnostics

- I/O and cancellation errors are returned by `Run`.
- Usage errors — duplicate keys, an incorrect state type, or an update during render — include the component path.
- A component panic must first restore the terminal; it may then be propagated or wrapped according to the policy chosen for `Run`.
- A future diagnostics option may record the cause of each render, phase durations, and changed screen regions.
- There will be no error boundary in the MVP; a clear terminal-recovery contract must come first.

## 13. Testing strategy

The core must be testable without a real terminal.

### Unit tests

- `Element` preserves props, keys, and children without mutation.
- `InitialState` is called once per mount.
- state survives a render with the same type and key.
- state is reinitialized when type or key changes.
- keyed reordering moves instances and preserves their state.
- new props and children reach the reused render.
- the nearest provider wins and its value does not leak to siblings.
- functional updates are applied in order and grouped by frame.
- the parser recognizes every key declared in the MVP and preserves available modifiers.
- the parser recognizes SGR mouse, buttons, movement, release, wheel, coordinates, and modifiers.
- `Consume` prevents bubbling and default behavior; `Propagate` preserves both.
- focus and blur are emitted once, in the correct order, without propagation.
- `Enter` and `Space` generate `PressEvent` only when the `KeyEvent` is not consumed.
- hit testing respects clipping and paint order; capture delivers move/up to the original target.
- enter/leave are derived once per transition, and a left click generates `PressEvent` only for compatible down/up pairs.
- text, paste, change, submit, and activation follow the order declared in [API.md](API.md).
- resizes are coalesced and external messages preserve order.
- resize preserves state and recalculates layout.
- effects run after commit, retain equal dependencies, cancel before cleanup, and clean up on omission, unmount, and runtime exit;
- refs preserve identity per key without invalidating, viewport reads follow resize, and focus handles request and release their bound host;
- measure/layout respects constraints, clipping, and Unicode.
- diff generates ANSI only for changed cells.
- `Row` and `Column` convert their props to `Box` without changing child identity.
- `Text` measures, wraps, and truncates by visual width.
- `Input` preserves its local cursor, keeps a controlled value, and handles editing, paste, and submit.
- `Tabs` validates keys, ignores disabled tabs, and accepts selection by keyboard or click.
- `List` requires keys, preserves keyed selection, and keeps the selected item visible.
- `List` scrolling clamps offsets, preserves its anchor during reordering, and works with items of different heights.
- wheel input moves the `List` without changing `SelectedKey` and propagates when no more space is available.

### Headless integration

A `TestBackend` receives synthetic events and captures frames:

1. mount the example counter;
2. inspect the first frame;
3. send `Tab` and `Enter`;
4. inspect incremented text and focus;
5. reorder keyed components;
6. confirm that their state followed the keys;
7. edit and submit a controlled `Input`;
8. navigate `Tabs` and `List` items by keyboard;
9. click `Button`, `Input`, `Tabs`, and `List` with synthetic coordinates;
10. scroll a `List` by wheel and inspect offset, clipping, and selection;
11. cancel and verify backend shutdown.

### Real terminal

Pseudo-terminal tests validate raw mode, SGR mouse protocol enable/disable, resize, restoration after errors, and ANSI sequences. Snapshots are useful for host trees and buffers; reconciler behavior should prefer less fragile semantic assertions.

## 14. Incremental plan

### Phase 0 — executable contract

Deliverables:

- initialize the Go module and basic CI;
- write the `Counter` example as a compilation test;
- define `Element`, `Component`, props, state, context, and children;
- create a minimal headless backend.

Completion criterion: the example compiles and mounts an inspectable tree, still without a terminal.

### Phase 1 — reconciliation and state

Deliverables:

- mount, update, and unmount instances;
- implement positional and keyed identity;
- implement the `SetState`/`UpdateState` queue;
- implement context providers/consumers;
- add diagnostics for paths and duplicate keys.

Completion criterion: all identity, state, props, children, and context tests pass in memory.

### Phase 2 — host tree and layout

Deliverables:

- `Text`, `Box`, `Row`, `Column`, `Fragment`, and `None`;
- measure/layout with row, column, constraints, padding, gap, and clipping;
- cell buffer with Unicode and styles;
- deterministic screen snapshots.

Completion criterion: headless trees produce correct buffers at different sizes.

### Phase 3 — interactive terminal

Deliverables:

- Unix backend, raw mode, and alternate screen;
- key, SGR mouse, wheel, and resize parser;
- mouse tracking enablement and terminal mode restoration;
- front/back ANSI diff;
- robust terminal restoration.

Completion criterion: the counter runs in a real terminal, resizes without losing state, and always restores the terminal on exit.

### Phase 4 — interaction

Deliverables:

- focus, bubbling, and event consumption;
- `KeyEvent`, `MouseEvent`, `WheelEvent`, `FocusEvent`, `BlurEvent`, `PressEvent`, `TextInputEvent`, `PasteEvent`, `ResizeEvent`, and `MessageEvent`;
- hit testing, hover path, and mouse capture;
- cancelable default behavior and ordered dispatch;
- keyboard- and mouse-accessible `Button`;
- end-to-end pseudo-terminal tests;
- render/layout/paint timing.

Completion criterion: keyboard and mouse interaction is deterministic, events reach the correct target under clipping, and a normal frame does not rewrite unchanged cells.

### Phase 5 — stateful and selectable builtins

Deliverables:

- controlled `Input`, cursor, editing, paste, submit, and click positioning;
- controlled `Tabs`, header navigation, clicks, and active panel;
- controlled `List`, viewport, navigation, clicks, activation, scrollbar, wheel, and automatic scrolling;
- `ValueChangeEvent`, `SubmitEvent`, and `ActivateEvent`;
- documentation and composition examples for all builtins.

Completion criterion: `Row`, `Column`, `Text`, `Input`, `Tabs`, and `List` are exported by `omnitui/components`, do not create a reverse runtime dependency, and pass the same reconciliation tests as user components.

### Phase 6 — hardening before expansion

Deliverables:

- race detector, ANSI parser fuzzing, and panic tests;
- public documentation and two larger examples;
- benchmarks with deep trees, keyed lists, and full screens;
- measurement-based decision about memoization or subtree rendering.

Completion criterion: the minimal API is stable and bottlenecks are known from benchmarks rather than assumptions.

## 15. Deliberately deferred decisions

- asynchronous components or suspense;
- render concorrente;
- public memoization;
- portals and overlays outside the normal tree;
- double-click, semantic drag, scrollbar dragging, direct clipboard access, and IME;
- list virtualization;
- animations and timers;
- custom markup or DSL.

Each item should be added only with a use case, defined semantics, and a test. The core should not anticipate every React capability: the primary inspiration is declarative trees, data flow, identity, and predictable reconciliation.

## 16. MVP success criteria

The MVP is ready when tests and an executable example can demonstrate that:

1. a component receives typed props, context, and children;
2. two instances of the same component maintain independent state;
3. a state update causes a new render without blocking or corrupting the loop;
4. type, position, and key correctly determine whether state is preserved;
5. components recursively compose primitives and other components;
6. resize and events do not reinitialize the tree;
7. only buffer differences are written to the terminal;
8. the terminal is restored on every exit path;
9. `go test -race ./...` passes;
10. `Row`, `Column`, `Text`, `Input`, `Tabs`, and `List` are exported exclusively by `omnitui/components`;
11. SGR mouse, hit testing, capture, click press, and wheel work in the real and headless backends;
12. the `Counter` example API remains small and understandable.

These criteria define the release boundary. Features that do not directly help meet them should wait until the first end-to-end version works.
