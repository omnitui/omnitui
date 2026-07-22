package integration

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	omnitui "github.com/omnitui/omnitui/v2"
	"github.com/omnitui/omnitui/v2/components"
)

type counterProps struct{ Label string }
type counterState struct{ Value int }
type counter struct{}

func (counter) InitialState(counterProps) counterState { return counterState{} }
func (counter) Render(ctx omnitui.Context, props counterProps, state counterState, _ omnitui.Children) omnitui.Element {
	return components.Column(components.ColumnProps{Gap: 1},
		components.Text(components.TextProps{Content: props.Label + ": " + itoa(state.Value)}),
		components.Button(components.ButtonProps{Label: "Incrementar", OnPress: func(omnitui.PressEvent) omnitui.EventResult {
			omnitui.UpdateState(ctx, func(current counterState) counterState { current.Value++; return current })
			return omnitui.Consume
		}}),
	)
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	return "1"
}

func TestCounterHeadlessInputAndState(t *testing.T) {
	typeValue := omnitui.Define[counterProps, counterState]("Counter", counter{})
	var output bytes.Buffer
	app := omnitui.New(omnitui.Create(typeValue, counterProps{Label: "Cliques"}), omnitui.Options{Input: strings.NewReader("\t\r\x03"), Output: &output, ColorProfile: omnitui.ColorProfileANSI16})
	err := app.Run(context.Background())
	if !errors.Is(err, omnitui.ErrInterrupted) {
		t.Fatalf("Run() error = %v, want ErrInterrupted", err)
	}
	if !strings.Contains(output.String(), "\x1b[1;10H1") {
		t.Fatalf("incremented frame not found in output: %q", output.String())
	}
}

type formState struct{ Value string }
type formComponent struct{}

func (formComponent) InitialState(string) formState { return formState{} }
func (formComponent) Render(ctx omnitui.Context, _ string, state formState, _ omnitui.Children) omnitui.Element {
	return components.Input(components.InputProps{Value: state.Value, Width: omnitui.Cells(8), OnChange: func(event omnitui.ValueChangeEvent) omnitui.EventResult {
		omnitui.UpdateState(ctx, func(current formState) formState { current.Value = event.Value; return current })
		return omnitui.Consume
	}})
}

func TestMouseFocusAndControlledInput(t *testing.T) {
	typeValue := omnitui.Define[string, formState]("Form", formComponent{})
	var output bytes.Buffer
	input := "\x1b[<0;1;1M\x1b[<0;1;1ma\x03"
	app := omnitui.New(omnitui.Create(typeValue, "form"), omnitui.Options{Input: strings.NewReader(input), Output: &output})
	if err := app.Run(context.Background()); !errors.Is(err, omnitui.ErrInterrupted) {
		t.Fatalf("Run() error = %v, want ErrInterrupted", err)
	}
	if !strings.Contains(output.String(), "a") {
		t.Fatal("controlled input never painted inserted text")
	}
}

func TestButtonMousePress(t *testing.T) {
	typeValue := omnitui.Define[counterProps, counterState]("Counter", counter{})
	var output bytes.Buffer
	input := "\x1b[<0;1;3M\x1b[<0;1;3m\x03"
	app := omnitui.New(omnitui.Create(typeValue, counterProps{Label: "Cliques"}), omnitui.Options{Input: strings.NewReader(input), Output: &output})
	if err := app.Run(context.Background()); !errors.Is(err, omnitui.ErrInterrupted) {
		t.Fatalf("Run() error = %v, want ErrInterrupted", err)
	}
	if !strings.Contains(output.String(), "\x1b[1;10H1") {
		t.Fatal("mouse press did not update counter")
	}
}

type tabsState struct{ Active string }
type tabsComponent struct{}

func (tabsComponent) InitialState(string) tabsState { return tabsState{Active: "overview"} }
func (tabsComponent) Render(ctx omnitui.Context, _ string, state tabsState, _ omnitui.Children) omnitui.Element {
	return components.Tabs(components.TabsProps{ActiveKey: state.Active, Items: []components.TabItem{{Key: "overview", Label: "Overview", Content: components.Text(components.TextProps{Content: "overview"})}, {Key: "logs", Label: "Logs", Content: components.Text(components.TextProps{Content: "logs"})}}, OnChange: func(event omnitui.ValueChangeEvent) omnitui.EventResult {
		omnitui.UpdateState(ctx, func(current tabsState) tabsState { current.Active = event.Value; return current })
		return omnitui.Consume
	}})
}

func TestTabsKeyboardSelection(t *testing.T) {
	typeValue := omnitui.Define[string, tabsState]("TabsScreen", tabsComponent{})
	var output bytes.Buffer
	app := omnitui.New(omnitui.Create(typeValue, "tabs"), omnitui.Options{Input: strings.NewReader("\t\x1b[C\r\x03"), Output: &output})
	if err := app.Run(context.Background()); !errors.Is(err, omnitui.ErrInterrupted) {
		t.Fatalf("Run() error = %v", err)
	}
	if !strings.Contains(output.String(), "logs") {
		t.Fatal("tabs did not paint the selected panel")
	}
}

type mouseTabsComponent struct{ changes *int }

func (mouseTabsComponent) InitialState(string) tabsState { return tabsState{Active: "overview"} }
func (c mouseTabsComponent) Render(ctx omnitui.Context, _ string, state tabsState, _ omnitui.Children) omnitui.Element {
	return components.Tabs(components.TabsProps{
		ActiveKey: state.Active,
		Items: []components.TabItem{
			{Key: "overview", Label: "Overview", Content: components.Text(components.TextProps{Content: "overview"})},
			{Key: "logs", Label: "Logs", Content: components.Text(components.TextProps{Content: "logs"})},
		},
		OnChange: func(event omnitui.ValueChangeEvent) omnitui.EventResult {
			*c.changes = *c.changes + 1
			omnitui.UpdateState(ctx, func(current tabsState) tabsState { current.Active = event.Value; return current })
			return omnitui.Consume
		},
	})
}

func TestTabsMouseSelectionUsesHeaderRow(t *testing.T) {
	changes := 0
	typeValue := omnitui.Define[string, tabsState]("MouseTabsScreen", mouseTabsComponent{changes: &changes})
	var output bytes.Buffer
	input := "\x1b[<0;14;2M\x1b[<0;14;2m\x1b[<0;14;1M\x1b[<0;14;1m\x03"
	app := omnitui.New(omnitui.Create(typeValue, "tabs"), omnitui.Options{Input: strings.NewReader(input), Output: &output})
	if err := app.Run(context.Background()); !errors.Is(err, omnitui.ErrInterrupted) {
		t.Fatalf("Run() error = %v", err)
	}
	if changes != 1 {
		t.Fatalf("tab changes = %d, want 1 (header click only)", changes)
	}
	if !strings.Contains(output.String(), "logs") {
		t.Fatal("header click did not select the second tab")
	}
}

type listState struct{ Selected string }
type listComponent struct{}

func (listComponent) InitialState(string) listState { return listState{Selected: "one"} }
func (listComponent) Render(ctx omnitui.Context, _ string, state listState, _ omnitui.Children) omnitui.Element {
	return components.List(components.ListProps{SelectedKey: state.Selected, Selectable: true, Height: omnitui.Cells(2), OnChange: func(event omnitui.ValueChangeEvent) omnitui.EventResult {
		omnitui.UpdateState(ctx, func(current listState) listState { current.Selected = event.Value; return current })
		return omnitui.Consume
	}}, components.Text(components.TextProps{Content: "one"}).WithKey("one"), components.Text(components.TextProps{Content: "two"}).WithKey("two"), components.Text(components.TextProps{Content: "three"}).WithKey("three"))
}

func TestListKeyboardSelection(t *testing.T) {
	typeValue := omnitui.Define[string, listState]("ListScreen", listComponent{})
	var output bytes.Buffer
	app := omnitui.New(omnitui.Create(typeValue, "list"), omnitui.Options{Input: strings.NewReader("\t\x1b[B\x03"), Output: &output})
	if err := app.Run(context.Background()); !errors.Is(err, omnitui.ErrInterrupted) {
		t.Fatalf("Run() error = %v", err)
	}
	if !strings.Contains(output.String(), "two") {
		t.Fatal("list did not paint the selected item")
	}
}

func TestListWheelScrollsWithoutChangingSelection(t *testing.T) {
	typeValue := omnitui.Define[string, listState]("ListScreen", listComponent{})
	var output bytes.Buffer
	app := omnitui.New(omnitui.Create(typeValue, "list"), omnitui.Options{Input: strings.NewReader("\x1b[<65;1;2M\x03"), Output: &output})
	if err := app.Run(context.Background()); !errors.Is(err, omnitui.ErrInterrupted) {
		t.Fatalf("Run() error = %v", err)
	}
	if !strings.Contains(output.String(), "hree") {
		t.Fatal("wheel did not reveal the third item")
	}
}
