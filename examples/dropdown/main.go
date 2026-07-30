package main

import (
	"context"

	omnitui "github.com/omnitui/omnitui/v2"
	"github.com/omnitui/omnitui/v2/components"
)

type dropdownState struct {
	Channel string
}

type dropdownExample struct{}

func (dropdownExample) InitialState(struct{}) dropdownState {
	return dropdownState{Channel: "stable"}
}

func (dropdownExample) Render(ctx omnitui.Context, _ struct{}, state dropdownState, _ omnitui.Children) omnitui.Element {
	return components.Column(
		components.ColumnProps{
			Gap:     1,
			Padding: omnitui.All(1),
			Style: omnitui.Style{
				Foreground: omnitui.RGB(224, 228, 238),
				Background: omnitui.RGB(20, 24, 33),
			},
		},
		components.Text(components.TextProps{
			Content: "Release channel",
			Style: omnitui.Style{
				Foreground: omnitui.ANSI(omnitui.BrightCyan),
				Attributes: omnitui.Bold,
			},
		}),
		components.Dropdown(components.DropdownProps{
			Options: []components.DropdownOption{
				{Key: "stable", Label: "Stable"},
				{Key: "preview", Label: "Preview"},
				{Key: "nightly", Label: "Nightly", Disabled: true},
			},
			SelectedKey: state.Channel,
			Width:       omnitui.Cells(24),
			MenuHeight:  omnitui.Cells(3),
			Wrap:        true,
			Style: omnitui.Style{
				Foreground: omnitui.ANSI(omnitui.White),
				Background: omnitui.RGB(35, 42, 58),
			},
			FocusStyle: omnitui.Style{
				Foreground: omnitui.ANSI(omnitui.BrightWhite),
				Background: omnitui.RGB(45, 75, 110),
				Attributes: omnitui.Underline,
			},
			MenuStyle: omnitui.Style{
				Foreground: omnitui.ANSI(omnitui.White),
				Background: omnitui.RGB(28, 34, 47),
			},
			SelectedStyle: omnitui.Style{
				Foreground: omnitui.ANSI(omnitui.Black),
				Background: omnitui.ANSI(omnitui.BrightCyan),
				Attributes: omnitui.Bold,
			},
			DisabledOptionStyle: omnitui.Style{
				Foreground: omnitui.ANSI(omnitui.BrightBlack),
				Attributes: omnitui.Dim,
			},
			OnChange: func(event omnitui.ValueChangeEvent) omnitui.EventResult {
				omnitui.UpdateState(ctx, func(current dropdownState) dropdownState {
					current.Channel = event.Value
					return current
				})
				return omnitui.Consume
			},
		}),
		components.Text(components.TextProps{Content: "Selected: " + state.Channel}),
		components.Text(components.TextProps{
			Content: "Enter opens • arrows navigate • Enter selects • Esc closes",
			Style:   omnitui.Style{Foreground: omnitui.ANSI(omnitui.BrightBlack)},
		}),
	)
}

var dropdownExampleType = omnitui.Define[struct{}, dropdownState]("DropdownExample", dropdownExample{})

func main() {
	root := omnitui.Create(dropdownExampleType, struct{}{})
	if err := omnitui.New(root, omnitui.Options{}).Run(context.Background()); err != nil {
		panic(err)
	}
}
