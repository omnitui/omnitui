package components

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	omnitui "github.com/omnitui/omnitui/v2"
	"github.com/omnitui/omnitui/v2/internal/core"
)

func TestDropdownValidatesOptions(t *testing.T) {
	mustPanic(t, func() {
		Dropdown(DropdownProps{Options: []DropdownOption{{Label: "Missing key"}}})
	})
	mustPanic(t, func() {
		Dropdown(DropdownProps{Options: []DropdownOption{{Key: "same"}, {Key: "same"}}})
	})
	mustPanic(t, func() {
		Dropdown(DropdownProps{
			Options:     []DropdownOption{{Key: "known"}},
			SelectedKey: "unknown",
		})
	})
	mustPanic(t, func() {
		Dropdown(DropdownProps{
			Options:     []DropdownOption{{Key: "disabled", Disabled: true}},
			SelectedKey: "disabled",
		})
	})
}

func TestDropdownCopiesOptions(t *testing.T) {
	options := []DropdownOption{{Key: "stable", Label: "Stable"}}
	element := Dropdown(DropdownProps{Options: options})
	options[0].Label = "Changed"

	props := core.PropsOf(element).(DropdownProps)
	if got := props.Options[0].Label; got != "Stable" {
		t.Fatalf("stored option label = %q, want Stable", got)
	}
}

func TestDropdownNavigationSkipsDisabledOptions(t *testing.T) {
	options := []DropdownOption{
		{Key: "stable"},
		{Key: "preview", Disabled: true},
		{Key: "nightly"},
	}
	if got := nextEnabledDropdownOption(options, "stable", "preview", false); got != "nightly" {
		t.Fatalf("next enabled option = %q, want nightly", got)
	}
	if got := nextEnabledDropdownOption(options, "nightly", "preview", false); got != "stable" {
		t.Fatalf("previous enabled option = %q, want stable", got)
	}
}

type dropdownHarnessState struct {
	Selected string
}

type dropdownHarness struct {
	changed *string
}

func (dropdownHarness) InitialState(struct{}) dropdownHarnessState {
	return dropdownHarnessState{Selected: "stable"}
}

func (h dropdownHarness) Render(ctx omnitui.Context, _ struct{}, state dropdownHarnessState, _ omnitui.Children) omnitui.Element {
	return Dropdown(DropdownProps{
		Options: []DropdownOption{
			{Key: "stable", Label: "Stable"},
			{Key: "preview", Label: "Preview", Disabled: true},
			{Key: "nightly", Label: "Nightly"},
		},
		SelectedKey: state.Selected,
		OnChange: func(event omnitui.ValueChangeEvent) omnitui.EventResult {
			*h.changed = event.Value
			omnitui.UpdateState(ctx, func(current dropdownHarnessState) dropdownHarnessState {
				current.Selected = event.Value
				return current
			})
			return omnitui.Consume
		},
	})
}

func TestDropdownSelectsWithKeyboard(t *testing.T) {
	if changed := runDropdownHarness(t, "\t\r\x1b[B\r\x03"); changed != "nightly" {
		t.Fatalf("selected option = %q, want nightly", changed)
	}
}

func TestDropdownSelectsWithMouse(t *testing.T) {
	if changed := runDropdownHarness(t, "\t\r\x1b[<0;2;4M\x1b[<0;2;4m\x03"); changed != "nightly" {
		t.Fatalf("selected option = %q, want nightly", changed)
	}
}

func TestDropdownIgnoresDisabledOptionClick(t *testing.T) {
	if changed := runDropdownHarness(t, "\t\r\x1b[<0;2;3M\x1b[<0;2;3m\x03"); changed != "" {
		t.Fatalf("selected option = %q after disabled click, want no change", changed)
	}
}

func runDropdownHarness(t *testing.T, inputSequence string) string {
	t.Helper()
	changed := ""
	typeValue := omnitui.Define[struct{}, dropdownHarnessState]("DropdownHarness", dropdownHarness{changed: &changed})
	input := strings.NewReader(inputSequence)
	app := omnitui.New(
		omnitui.Create(typeValue, struct{}{}),
		omnitui.Options{Input: input, Output: io.Discard},
	)

	if err := app.Run(context.Background()); !errors.Is(err, omnitui.ErrInterrupted) {
		t.Fatalf("Run() error = %v, want ErrInterrupted", err)
	}
	return changed
}
