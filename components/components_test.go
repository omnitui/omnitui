package components

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	omnitui "github.com/omnitui/omnitui"
	"github.com/omnitui/omnitui/internal/core"
)

type focusForwardProbe struct{ seen *bool }

func (focusForwardProbe) InitialState(struct{}) struct{} { return struct{}{} }
func (probe focusForwardProbe) Render(ctx omnitui.Context, _ struct{}, _ struct{}, _ omnitui.Children) omnitui.Element {
	handle := omnitui.UseFocus(ctx, "control")
	hosts := []omnitui.Element{
		Box(BoxProps{Focusable: true, Focus: handle}),
		buttonHost(ButtonProps{Focus: handle}),
		inputHost(InputProps{Focus: handle}),
		tabsHost(TabsProps{Focus: handle}),
		listHost(ListProps{Focus: handle}),
	}
	for _, element := range hosts {
		host, ok := core.HostOf(element)
		if !ok || hostFocus(host.Data) != handle {
			return Text(TextProps{Content: "focus forwarding failed"})
		}
	}
	*probe.seen = true
	return hosts[0]
}

func hostFocus(data any) omnitui.FocusHandle {
	switch value := data.(type) {
	case core.BoxData:
		return value.Focus.(omnitui.FocusHandle)
	case core.ButtonData:
		return value.Focus.(omnitui.FocusHandle)
	case core.InputData:
		return value.Focus.(omnitui.FocusHandle)
	case core.TabsData:
		return value.Focus.(omnitui.FocusHandle)
	case core.ListData:
		return value.Focus.(omnitui.FocusHandle)
	default:
		return omnitui.FocusHandle{}
	}
}

func TestRowAndColumnAreBoxHosts(t *testing.T) {
	child := Text(TextProps{Content: "child"}).WithKey("stable")
	for name, element := range map[string]omnitui.Element{"row": Row(RowProps{Gap: 1}, child), "column": Column(ColumnProps{Gap: 1}, child)} {
		if core.KindOf(element) != core.KindComponent || core.ComponentOf(element).Name != nameTitle(name) {
			t.Fatalf("%s is not a builtin component", name)
		}
	}
}

func nameTitle(value string) string {
	if value == "row" {
		return "Row"
	}
	return "Column"
}

func TestTabsAndListValidateKeys(t *testing.T) {
	mustPanic(t, func() { Tabs(TabsProps{Items: []TabItem{{Key: "same"}, {Key: "same"}}}) })
	mustPanic(t, func() { List(ListProps{Selectable: true}, Text(TextProps{Content: "missing"})) })
	Tabs(TabsProps{ActiveKey: "ok", Items: []TabItem{{Key: "ok"}, {Key: "disabled", Disabled: true}}})
}

func TestStyleConflictIsRejected(t *testing.T) {
	mustPanic(t, func() { Text(TextProps{Style: omnitui.Style{Attributes: omnitui.Bold, ClearAttributes: omnitui.Bold}}) })
}

func TestFocusableComponentsForwardFocusHandle(t *testing.T) {
	seen := false
	typeValue := omnitui.Define[struct{}, struct{}]("FocusForwardProbe", focusForwardProbe{seen: &seen})
	app := omnitui.New(omnitui.Create(typeValue, struct{}{}), omnitui.Options{Input: strings.NewReader("\x03"), Output: io.Discard})
	if err := app.Run(context.Background()); !errors.Is(err, omnitui.ErrInterrupted) {
		t.Fatalf("Run() error = %v, want ErrInterrupted", err)
	}
	if !seen {
		t.Fatal("focus handle was not forwarded by every focusable component")
	}
}

func mustPanic(t *testing.T, action func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	action()
}
