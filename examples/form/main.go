package main

import (
	"context"

	omnitui "github.com/omnitui/omnitui/v2"
	"github.com/omnitui/omnitui/v2/components"
)

type formState struct{ Name, Submitted string }
type form struct{}

var formPanelStyle = omnitui.Style{
	Foreground: omnitui.RGB(224, 228, 238),
	Background: omnitui.RGB(20, 24, 33),
}

func (form) InitialState(string) formState { return formState{} }
func (form) Render(ctx omnitui.Context, _ string, state formState, _ omnitui.Children) omnitui.Element {
	return components.Column(components.ColumnProps{Padding: omnitui.All(1), Gap: 1, Style: formPanelStyle},
		components.Text(components.TextProps{Content: "Name", Style: omnitui.Style{Foreground: omnitui.ANSI(omnitui.BrightCyan), Attributes: omnitui.Bold}}),
		components.Input(components.InputProps{
			Value:       state.Name,
			Placeholder: "Enter your name",
			Width:       omnitui.Cells(24),
			Style:       omnitui.Style{Foreground: omnitui.ANSI(omnitui.White), Background: omnitui.RGB(35, 42, 58)},
			FocusStyle:  omnitui.Style{Foreground: omnitui.ANSI(omnitui.BrightWhite), Background: omnitui.RGB(45, 75, 110), Attributes: omnitui.Underline},
			OnChange: func(event omnitui.ValueChangeEvent) omnitui.EventResult {
				omnitui.UpdateState(ctx, func(current formState) formState { current.Name = event.Value; return current })
				return omnitui.Consume
			},
			OnSubmit: func(event omnitui.SubmitEvent) omnitui.EventResult {
				omnitui.UpdateState(ctx, func(current formState) formState { current.Submitted = event.Value; return current })
				return omnitui.Consume
			},
		}),
		components.Text(components.TextProps{Content: state.Submitted, Style: omnitui.Style{Foreground: omnitui.ANSI(omnitui.BrightGreen), Attributes: omnitui.Bold}}),
	)
}

func main() {
	typeValue := omnitui.Define("Form", form{})
	app := omnitui.New(omnitui.Create(typeValue, "form"), omnitui.Options{})
	if err := app.Run(context.Background()); err != nil {
		panic(err)
	}
}
