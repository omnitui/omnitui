# OmniTUI — planned source structure

This document proposes the folder and file structure for implementing the API defined in [API.md](API.md), the architecture in [DESIGN.md](DESIGN.md), and the examples in [COMPONENTS.md](COMPONENTS.md).

## 1. Package decision

The framework has two public packages:

```go
import (
    omnitui "github.com/viniciusfonseca/omnitui"
    components "github.com/viniciusfonseca/omnitui/components"
)
```

- `omnitui` contains the runtime, elements, components, state, context, and events.
- `omnitui/components` exports `Box`, `Row`, `Column`, `Text`, `Button`, `Input`, `Tabs`, and `List`.

An application may use a short alias without changing the package’s actual name:

```go
import ui "github.com/viniciusfonseca/omnitui/components"

view := ui.Row(ui.RowProps{Gap: 1}, ...)
```

The root package does not import `components`. This direction is required to avoid a cycle: builtin components depend on the runtime, but the runtime does not depend on the component catalog.

## 2. Principles

1. `omnitui` and `components` are the only public APIs in the MVP.
2. `components` imports `omnitui`; the reverse is forbidden.
3. An opaque representation in `internal/core` lets both packages build and consume `Element` without exposing internal hosts.
4. Reconciliation and the runtime remain in the root package while this is the simplest boundary.
5. Independent algorithms and platform details live under `internal/`.
6. Internal packages never import `omnitui` or `components`.
7. Tests stay close to the code; complete scenarios use the headless backend.
8. Folders for deferred features are not created preemptively.

## 3. Planned tree

```text
omnitui/
├── go.mod
├── go.sum
├── README.md
├── LICENSE
│
├── docs/
│   ├── DESIGN.md
│   ├── API.md
│   ├── COMPONENTS.md
│   └── STRUCTURE.md
│
├── app.go
├── options.go
├── element.go
├── component.go
├── context.go
├── state.go
├── event.go
├── event_key.go
├── event_mouse.go
├── style.go
├── size.go
├── geometry.go
│
├── instance.go
├── reconcile.go
├── reconcile_children.go
├── dispatch.go
├── focus.go
├── mouse.go
├── hit_testing.go
├── runtime.go
├── paint.go
│
├── components/
│   ├── doc.go
│   ├── box.go
│   ├── row.go
│   ├── column.go
│   ├── text.go
│   ├── button.go
│   ├── input.go
│   ├── tabs.go
│   ├── list.go
│   ├── row_test.go
│   ├── column_test.go
│   ├── text_test.go
│   ├── input_test.go
│   ├── tabs_test.go
│   └── list_test.go
│
├── internal/
│   ├── core/
│   │   ├── element.go
│   │   ├── component.go
│   │   ├── host.go
│   │   ├── host_box.go
│   │   ├── host_text.go
│   │   ├── host_editable.go
│   │   ├── handler.go
│   │   ├── style.go
│   │   └── geometry.go
│   │
│   ├── backend/
│   │   ├── backend.go
│   │   ├── event.go
│   │   ├── ansi/
│   │   │   ├── backend.go
│   │   │   ├── input.go
│   │   │   ├── parser.go
│   │   │   ├── mouse.go
│   │   │   ├── parser_test.go
│   │   │   ├── mouse_test.go
│   │   │   ├── raw_unix.go
│   │   │   ├── resize_unix.go
│   │   │   └── restore_test.go
│   │   └── headless/
│   │       ├── backend.go
│   │       ├── recorder.go
│   │       └── backend_test.go
│   │
│   ├── layout/
│   │   ├── constraints.go
│   │   ├── node.go
│   │   ├── measure.go
│   │   ├── arrange.go
│   │   ├── clip.go
│   │   └── layout_test.go
│   │
│   ├── screen/
│   │   ├── cell.go
│   │   ├── buffer.go
│   │   ├── diff.go
│   │   ├── ansi.go
│   │   └── diff_test.go
│   │
│   └── text/
│       ├── grapheme.go
│       ├── width.go
│       ├── wrap.go
│       ├── truncate.go
│       └── text_test.go
│
├── examples/
│   ├── counter/
│   │   └── main.go
│   ├── form/
│   │   └── main.go
│   └── catalog/
│       └── main.go
│
├── integration/
│   ├── app_test.go
│   ├── state_test.go
│   ├── events_test.go
│   ├── mouse_test.go
│   ├── components_test.go
│   └── terminal_test.go
│
└── testdata/
    ├── screens/
    │   ├── counter_initial.txt
    │   ├── counter_incremented.txt
    │   ├── form_focused.txt
    │   └── catalog_list_scrolled.txt
    └── input/
        ├── ansi_sequences.json
        └── mouse_sequences.json
```

This is the target structure for the MVP. Section 9 indicates when each group should be created.

## 4. Public `omnitui` package

### 4.1 Core API

The canonical list of public types, functions, and behavior is in [API.md](API.md). This section records only the physical responsibility of each file.

| File | Responsibility |
|---|---|
| `app.go` | `App`, `New`, `Run`, `UpdateRoot`, and `Dispatch` |
| `options.go` | `Options`, defaults, and validation |
| `element.go` | `Element`, `Children`, `WithKey`, `None`, and `Fragment` |
| `component.go` | `Component`, `ComponentType`, `Define`, and `Create` |
| `context.go` | Render context and typed providers |
| `state.go` | `SetState`, `UpdateState`, and pending updates |
| `event.go` | Events, handlers, `Propagate`, and `Consume` |
| `event_key.go` | Public keys, runes, and modifiers |
| `event_mouse.go` | Actions, buttons, coordinates, `MouseEvent`, and `WheelEvent` |
| `style.go` | Colors, attributes, and `Style` |
| `size.go` | `Size`, `Cells`, and automatic dimensions |
| `geometry.go` | `Spacing`, `Rect`, and helpers such as `All` |

The root package does not export visual components. In particular, there are no `omnitui.Text`, `omnitui.Row`, or `omnitui.List`; those symbols belong to `components`.

### 4.2 Runtime and reconciliation

| File | Responsibility |
|---|---|
| `instance.go` | Mounted instances, identity, state, props, and diagnostic path |
| `reconcile.go` | Mount, update, replace, and unmount |
| `reconcile_children.go` | Positional and keyed reconciliation |
| `dispatch.go` | Serialized queue for state, messages, resize, and input |
| `focus.go` | Focus order, targets, and recovery after unmount |
| `mouse.go` | Hover path, capture, button state, and default behavior |
| `hit_testing.go` | Target lookup by position, clipping, and paint order |
| `runtime.go` | Main loop and phase coordination |
| `paint.go` | Conversion of the positioned host tree to the back buffer |

These files remain in the root package and keep their internal symbols unexported.

## 5. Public `components` package

### 5.1 Exported API

| File | Main symbols |
|---|---|
| `doc.go` | Overview and package import example |
| `box.go` | `Box`, `BoxProps` |
| `row.go` | `Row`, `RowProps` |
| `column.go` | `Column`, `ColumnProps` |
| `text.go` | `Text`, `TextProps`, wrapping, and truncation |
| `button.go` | `Button`, `ButtonProps` |
| `input.go` | `Input`, `InputProps` |
| `tabs.go` | `Tabs`, `TabsProps`, `TabItem` |
| `list.go` | `List`, `ListProps`, `ScrollbarMode` |

Signatures, props, and enums are centralized in [API.md](API.md). This section records only the location of each implementation.

### 5.2 Implementation

- `Row`, `Column`, `Button`, `Input`, `Tabs`, and `List` use the `omnitui.Component` contract.
- `Box` and `Text` create hosts through `internal/core`.
- `Input` uses the private editable host in `internal/core`.
- Public props use fundamental `omnitui` types such as `Style`, `Size`, `Spacing`, and events.
- No `components` API is required to start or run an application.

The package does not access instances, the reconciler, queues, or the backend. It describes elements and reacts to events like any user-component library.

## 6. Shared `internal/core` boundary

Separating builtins creates a Go-specific problem: unexported fields in `omnitui.Element` would not be accessible to `components`, while exporting host constructors would pollute the runtime API.

`internal/core` solves this by:

- defining the concrete opaque representation of `Element`;
- defining kinds and payloads for `Box`, `Text`, and editable hosts;
- storing type-erased handlers;
- containing shared neutral style and geometry values;
- knowing nothing about the runtime, state, context, or builtin components.

The root package re-exports non-generic public aliases for the opaque element representation and shared style and geometry values. The names and contracts of these aliases are defined exclusively in [API.md](API.md).

Because `internal/core` is under the module tree, `components` can import it but external consumers cannot. The catalog can therefore create hosts without making that construction part of the public API.

`ComponentType[P]`, `Context`, and generic helpers remain defined in `omnitui`; they do not need to be aliases for internal types.

## 7. Other internal packages

### 7.1 `internal/backend`

Defines the neutral contract between the runtime and platforms:

```go
type Backend interface {
    Size() (width, height int, err error)
    Events() <-chan Event
    Write([]byte) error
    Close() error
}
```

- `backend/ansi` implements raw mode, alternate screen, ANSI input, SGR mouse, wheel, resize, and Unix restoration.
- `backend/headless` receives synthetic events and captures frames for tests.
- backends know nothing about `omnitui`, `components`, focus, or state.

### 7.2 `internal/layout`

Runs `measure` and `arrange` over a normalized tree. It knows nothing about components, events, or ANSI.

### 7.3 `internal/screen`

Owns cells, buffers, clipping, diffing, and deterministic ANSI encoding. `Style` is compiled into internal attributes before painting.

### 7.4 `internal/text`

Centralizes graphemes, visual width, wrapping, truncation, and cursor movement. Only this package depends directly on the selected Unicode library.

## 8. Dependency direction

```text
examples/ -------------------------------> omnitui
examples/ -------------------------------> components
integration/ ----------------------------> omnitui
integration/ ----------------------------> components
integration/ -----> internal/backend/headless

components -----> omnitui
components -----> internal/core

omnitui -----> internal/core
omnitui -----> internal/backend
omnitui -----> internal/backend/ansi
omnitui -----> internal/layout
omnitui -----> internal/screen
omnitui -----> internal/text

internal/backend/ansi -----> internal/backend
internal/backend/headless --> internal/backend

omnitui -X-> components
internal/* -X-> omnitui
internal/* -X-> components
```

Required rules:

- `components` may import `omnitui`; `omnitui` never imports `components`;
- internal packages do not import any public module package;
- `internal/core` is shared data structure, not a second runtime;
- `ansi` and `headless` depend only on the `backend` contract;
- examples import only public packages;
- integration tests may import the headless backend;
- do not create generic `util`, `common`, `shared`, or `helpers` packages.

## 9. Incremental creation

### Phases 0 and 1 — contract and reconciliation

```text
go.mod
element.go
component.go
context.go
state.go
event.go
app.go
instance.go
reconcile.go
reconcile_children.go
runtime.go
internal/core/element.go
internal/core/component.go
internal/backend/backend.go
internal/backend/headless/backend.go
examples/counter/main.go
```

### Phase 2 — hosts, layout, and basic components

```text
style.go
size.go
geometry.go
internal/core/host*.go
internal/core/style.go
internal/core/geometry.go
internal/layout/
internal/screen/
internal/text/
components/box.go
components/row.go
components/column.go
components/text.go
testdata/screens/
```

### Phases 3 and 4 — terminal and interaction

```text
event_key.go
event_mouse.go
dispatch.go
focus.go
mouse.go
hit_testing.go
paint.go
components/button.go
internal/backend/ansi/
testdata/input/mouse_sequences.json
integration/events_test.go
integration/mouse_test.go
integration/terminal_test.go
```

### Phase 5 — stateful components

```text
internal/core/host_editable.go
components/input.go
components/tabs.go
components/list.go
examples/form/
examples/catalog/
integration/components_test.go
testdata/screens/catalog_list_scrolled.txt
```

### Phase 6 — hardening

Add benchmarks, fuzz tests, and new files only where measurements indicate a need. Memoization, virtualization, and a scheduler do not get preventive folders.

## 10. Tests

- runtime tests stay with the root package;
- each builtin has tests in `components/*_test.go`;
- algorithm tests live under the corresponding internal package;
- `integration/` uses `omnitui` and `components` as an external consumer;
- `integration/mouse_test.go` covers hit testing, clipping, bubbling, capture, hover, press, and wheel;
- snapshots live in `testdata/screens`;
- parser fuzzing, including SGR mouse sequences, lives in `internal/backend/ansi/parser_fuzz_test.go`.

`components` tests should prefer the public contract and the headless backend. Direct access to `internal/core` is acceptable only for testing prop translation into hosts.

## 11. Deliberately deferred structures

```text
internal/backend/windows/  # native Windows support
internal/virtual/          # virtualized lists
internal/animation/        # clock, scheduler, and transitions
internal/effects/          # lifecycle and cleanup
```

These names are conceptual markers, not reserved directories.

## 12. Acceptance criteria

1. Applications import `omnitui` for the runtime and `components` for builtin UI.
2. No visual component is exported by the root package.
3. `components` depends on `omnitui`, never the reverse.
4. No import cycle is required.
5. Internal hosts are not exposed to external consumers.
6. ANSI and headless backends implement the same contract.
7. The reconciler is testable without builtin components or a terminal.
8. Builtins use the same instance and state model as user components.
9. SGR mouse is isolated in the backend; hit testing and capture remain in the runtime.
10. `go test ./...` covers both public packages and integration tests.
11. `go test -race ./...` finds no mutation outside the runtime goroutine.
