package omnitui

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/omnitui/omnitui/v2/internal/backend"
	"github.com/omnitui/omnitui/v2/internal/backend/headless"
	"github.com/omnitui/omnitui/v2/internal/core"
	"github.com/omnitui/omnitui/v2/internal/screen"
)

type probeProps struct{ Value string }
type probeState struct{ Count int }
type probe struct {
	initial *int
	seen    *[]string
}

func (p probe) InitialState(probeProps) probeState { *p.initial++; return probeState{} }
func (p probe) Render(_ Context, props probeProps, state probeState, children Children) Element {
	*p.seen = append(*p.seen, fmt.Sprintf("%s:%d:%d", props.Value, state.Count, len(children)))
	return None()
}

func TestReconcilePreservesStateAndUpdatesProps(t *testing.T) {
	initial, seen := 0, []string{}
	typeValue := Define[probeProps, probeState]("Probe", probe{initial: &initial, seen: &seen})
	app := New(Create(typeValue, probeProps{Value: "first"}), Options{Input: nil, Output: nil})
	app.width, app.height = 20, 2
	if err := app.render(); err != nil {
		t.Fatal(err)
	}
	app.root = Create(typeValue, probeProps{Value: "second"}, None())
	app.invalidated = true
	if err := app.render(); err != nil {
		t.Fatal(err)
	}
	if initial != 1 {
		t.Fatalf("InitialState calls = %d, want 1", initial)
	}
	if got := seen[len(seen)-1]; got != "second:0:1" {
		t.Fatalf("last render = %q", got)
	}
}

type contextProbe struct{ seen *[]int }

func (p contextProbe) InitialState(ContextKey[int]) struct{} { return struct{}{} }
func (p contextProbe) Render(ctx Context, key ContextKey[int], _ struct{}, _ Children) Element {
	*p.seen = append(*p.seen, UseContext(ctx, key))
	return None()
}

func TestProviderScopeDoesNotLeakToSibling(t *testing.T) {
	key := NewContext(7)
	seen := []int{}
	typeValue := Define[ContextKey[int], struct{}]("ContextProbe", contextProbe{seen: &seen})
	root := Fragment(Provide(key, 11, Create(typeValue, key)), Create(typeValue, key))
	app := New(root, Options{})
	app.width, app.height = 10, 2
	if err := app.render(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(seen, []int{11, 7}) {
		t.Fatalf("context values = %v", seen)
	}
}

type renderUpdate struct{}

func (renderUpdate) InitialState(int) int { return 0 }
func (renderUpdate) Render(ctx Context, _ int, _ int, _ Children) Element {
	SetState(ctx, 1)
	return None()
}

func TestStateUpdateDuringRenderPanics(t *testing.T) {
	typeValue := Define[int, int]("RenderUpdate", renderUpdate{})
	app := New(Create(typeValue, 0), Options{})
	app.width, app.height = 1, 1
	defer func() {
		if recover() == nil {
			t.Fatal("expected render state update panic")
		}
	}()
	_ = app.render()
}

func TestHeadlessBackendFeedsRuntime(t *testing.T) {
	headlessBackend := headless.New(8, 2)
	app := newWithBackend(core.NewHost(core.HostBox, core.BoxData{}, nil), Options{}, func() (backend.Backend, error) { return headlessBackend, nil })
	headlessBackend.Resize(4, 1)
	headlessBackend.Send(backend.KeyInput{Modifiers: 1})
	if err := app.Run(context.Background()); !errors.Is(err, ErrInterrupted) {
		t.Fatalf("Run() error = %v, want ErrInterrupted", err)
	}
	if len(headlessBackend.Frames()) == 0 {
		t.Fatal("headless backend did not record a frame")
	}
}

func TestResizeEventsCoalesce(t *testing.T) {
	widths := []int{}
	root := core.NewHost(core.HostBox, core.BoxData{Handlers: core.Handlers{"resize": EventHandler[ResizeEvent](func(event ResizeEvent) EventResult { widths = append(widths, event.Width); return Consume })}}, nil)
	headlessBackend := headless.New(8, 2)
	headlessBackend.Resize(3, 2)
	headlessBackend.Resize(4, 2)
	headlessBackend.Send(backend.KeyInput{Modifiers: 1})
	app := newWithBackend(root, Options{}, func() (backend.Backend, error) { return headlessBackend, nil })
	if err := app.Run(context.Background()); !errors.Is(err, ErrInterrupted) {
		t.Fatalf("Run() error = %v", err)
	}
	if !reflect.DeepEqual(widths, []int{4}) {
		t.Fatalf("resize callbacks = %v", widths)
	}
}

func TestFillUsesAvailableSpace(t *testing.T) {
	tests := []struct {
		name          string
		parent        core.BoxData
		child         core.Element
		width, height int
		wantWidth     int
		wantHeight    int
	}{
		{
			name:   "box",
			parent: core.BoxData{Padding: core.Spacing{Top: 1, Right: 1, Bottom: 1, Left: 1}},
			child:  core.NewHost(core.HostBox, core.BoxData{Width: core.FillSize(), Height: core.FillSize()}, nil),
			width:  10, height: 4,
			wantWidth: 8, wantHeight: 2,
		},
		{
			name:  "input",
			child: core.NewHost(core.HostInput, core.InputData{Width: core.FillSize()}, nil),
			width: 10, height: 4,
			wantWidth: 10, wantHeight: 1,
		},
		{
			name:  "empty list",
			child: core.NewHost(core.HostList, core.ListData{Height: core.FillSize()}, nil),
			width: 10, height: 4,
			wantWidth: 0, wantHeight: 4,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := New(core.NewHost(core.HostBox, test.parent, []core.Element{test.child}), Options{})
			app.width, app.height = test.width, test.height
			if err := app.render(); err != nil {
				t.Fatal(err)
			}
			got := app.rootInstance.children[0].rect
			if got.Width != test.wantWidth || got.Height != test.wantHeight {
				t.Fatalf("child rect = %#v, want width=%d height=%d", got, test.wantWidth, test.wantHeight)
			}
		})
	}
}

func TestListScrollPaddingAtStartDoesNotShiftItems(t *testing.T) {
	items := []Element{
		core.NewHost(core.HostText, core.TextData{Content: "first"}, nil).WithKey("first"),
		core.NewHost(core.HostText, core.TextData{Content: "second"}, nil).WithKey("second"),
		core.NewHost(core.HostText, core.TextData{Content: "third"}, nil).WithKey("third"),
	}
	root := core.NewHost(
		core.HostList,
		core.ListData{
			Height:        core.CellsSize(3),
			Selectable:    true,
			SelectedKey:   "first",
			ScrollPadding: 1,
		},
		items,
	)
	app := New(root, Options{})
	app.width, app.height = 10, 3

	if err := app.render(); err != nil {
		t.Fatal(err)
	}
	if got := app.rootInstance.listOffset; got != 0 {
		t.Fatalf("list offset = %d, want 0", got)
	}
	if got := app.rootInstance.children[0].rect.Y; got != 0 {
		t.Fatalf("first item Y = %d, want 0", got)
	}
}

func TestOverlayDoesNotConsumeLayoutAndPaintsAboveSiblings(t *testing.T) {
	anchor := core.NewHost(core.HostText, core.TextData{Content: "anchor"}, nil)
	menu := core.NewHost(core.HostText, core.TextData{Content: "menu"}, nil)
	overlay := core.NewHost(core.HostOverlay, core.OverlayData{}, []Element{menu})
	below := core.NewHost(core.HostText, core.TextData{Content: "below"}, nil)
	root := core.NewHost(
		core.HostBox,
		core.BoxData{Direction: 1},
		[]Element{anchor, overlay, below},
	)
	app := New(root, Options{})
	app.width, app.height = 20, 4

	if err := app.render(); err != nil {
		t.Fatal(err)
	}
	overlayInstance := app.rootInstance.children[1]
	belowInstance := app.rootInstance.children[2]
	if got, want := belowInstance.rect.Y, 1; got != want {
		t.Fatalf("sibling Y = %d, want %d", got, want)
	}
	if got, want := overlayInstance.children[0].rect.Y, belowInstance.rect.Y; got != want {
		t.Fatalf("overlay Y = %d, want sibling Y %d", got, want)
	}
	if got := app.front.Cell(0, 1).Grapheme; got != "m" {
		t.Fatalf("painted grapheme = %q, want overlay grapheme m", got)
	}
	if got := app.targetAt(0, 1); got != overlayInstance.children[0] {
		t.Fatalf("hit target = %v, want overlay child", got)
	}
}

func TestBoxLabelOverlaysAndTruncatesTopBorder(t *testing.T) {
	tests := []struct {
		name  string
		width int
		label string
		want  string
	}{
		{name: "label", width: 12, label: "Panel", want: "┌ Panel ───┐"},
		{name: "truncated", width: 7, label: "Settings", want: "┌ Se… ┐"},
		{name: "first line", width: 10, label: "Title\nignored", want: "┌ Title ─┐"},
		{name: "wide grapheme", width: 9, label: "🙂 UI", want: "┌ 🙂 UI ┐"},
		{name: "too narrow", width: 4, label: "Panel", want: "┌──┐"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			box := &instance{
				host:  core.Host{Kind: core.HostBox, Data: core.BoxData{Border: 1, Label: test.label}},
				rect:  Rect{Width: test.width, Height: 3},
				clip:  Rect{Width: test.width, Height: 3},
				style: Style{},
			}
			buffer := screen.NewBuffer(test.width, 3, Style{})

			paintBox(buffer, box, box.host.Data.(core.BoxData))

			got := ""
			for x := 0; x < test.width; x++ {
				cell := buffer.Cell(x, 0)
				if cell.Width > 0 {
					got += cell.Grapheme
				}
			}
			if got != test.want {
				t.Fatalf("top border = %q, want %q", got, test.want)
			}
		})
	}
}

func TestBoxLabelContributesToAutoWidth(t *testing.T) {
	box := &instance{
		host: core.Host{
			Kind: core.HostBox,
			Data: core.BoxData{Border: 1, Label: "Settings"},
		},
	}

	width, height := measureHost(box, 20, 5)

	if width != 12 || height != 2 {
		t.Fatalf("size = %dx%d, want 12x2", width, height)
	}
}

func TestTabsHorizontalPaddingIsPaintedAndClickable(t *testing.T) {
	activeStyle := core.Style{
		Foreground: core.ANSIColorValue(1),
		Background: core.ANSIColorValue(14),
	}
	tabs := &instance{
		host: core.Host{
			Kind: core.HostTabs,
			Data: core.TabsData{
				Items: []core.TabData{
					{Key: "one", Label: "One"},
					{Key: "two", Label: "Two"},
				},
				ActiveKey:   "one",
				ActiveStyle: activeStyle,
			},
		},
		rect: Rect{Width: 10, Height: 1},
	}
	data := tabs.host.Data.(core.TabsData)
	buffer := screen.NewBuffer(10, 1, core.Style{})
	paintTabs(buffer, tabs, data)

	for _, x := range []int{0, 4} {
		if got := buffer.Cell(x, 0).Style; got != activeStyle {
			t.Fatalf("active tab padding at x=%d has style %#v, want %#v", x, got, activeStyle)
		}
	}
	if got := tabsAtForTest(tabs, 0, 0); got != "one" {
		t.Fatalf("left active padding hit tab %q, want one", got)
	}
	if got := tabsAtForTest(tabs, 4, 0); got != "one" {
		t.Fatalf("right active padding hit tab %q, want one", got)
	}
	if got := tabsAtForTest(tabs, 5, 0); got != "two" {
		t.Fatalf("left inactive padding hit tab %q, want two", got)
	}
}

func tabsAtForTest(tabs *instance, x, y int) string {
	return (&App{}).tabAt(tabs, x, y)
}
