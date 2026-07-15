package main

import (
	"context"

	omnitui "github.com/viniciusfonseca/omnitui"
	"github.com/viniciusfonseca/omnitui/components"
)

var (
	surfaceStyle = omnitui.Style{
		Foreground: omnitui.RGB(224, 228, 238),
		Background: omnitui.RGB(20, 24, 33),
	}
	titleStyle = omnitui.Style{
		Foreground: omnitui.ANSI(omnitui.BrightCyan),
		Attributes: omnitui.Bold | omnitui.Underline,
	}
	mutedStyle = omnitui.Style{
		Foreground: omnitui.ANSI(omnitui.BrightBlack),
		Attributes: omnitui.Dim,
	}
	tabStyle = omnitui.Style{
		Foreground: omnitui.ANSI(omnitui.BrightBlack),
	}
	activeTabStyle = omnitui.Style{
		Foreground: omnitui.ANSI(omnitui.BrightCyan),
		Attributes: omnitui.Bold | omnitui.Underline,
	}
	inputStyle = omnitui.Style{
		Foreground: omnitui.ANSI(omnitui.White),
		Background: omnitui.RGB(35, 42, 58),
	}
	inputFocusStyle = omnitui.Style{
		Foreground: omnitui.ANSI(omnitui.BrightWhite),
		Background: omnitui.RGB(45, 75, 110),
		Attributes: omnitui.Underline,
	}
	buttonStyle = omnitui.Style{
		Foreground: omnitui.ANSI(omnitui.Black),
		Background: omnitui.ANSI(omnitui.BrightCyan),
		Attributes: omnitui.Bold,
	}
	buttonFocusStyle = omnitui.Style{
		Foreground: omnitui.ANSI(omnitui.Black),
		Background: omnitui.ANSI(omnitui.BrightWhite),
		Attributes: omnitui.Bold | omnitui.Underline,
	}
)

type styledState struct {
	ActiveKey string
	Name      string
	Saved     bool
}

type styledExample struct{}

func (styledExample) InitialState(string) styledState {
	return styledState{ActiveKey: "dashboard"}
}

func (styledExample) Render(ctx omnitui.Context, _ string, state styledState, _ omnitui.Children) omnitui.Element {
	status := "Not saved yet"
	if state.Saved {
		status = "Saved for " + state.Name
	}

	return components.Box(
		components.BoxProps{
			Direction: components.Vertical,
			Padding:   omnitui.All(1),
			Gap:       1,
			Border:    components.BorderRounded,
			Style:     surfaceStyle,
		},
		components.Text(components.TextProps{
			Content: "OmniTUI • Styled components",
			Style:   titleStyle,
		}),
		components.Text(components.TextProps{
			Content: "Colors, attributes, borders, and focus styles on one screen.",
			Style:   mutedStyle,
		}),
		components.Tabs(components.TabsProps{
			ActiveKey:   state.ActiveKey,
			Style:       tabStyle,
			ActiveStyle: activeTabStyle,
			Items: []components.TabItem{
				{Key: "dashboard", Label: "Dashboard", Content: components.Text(components.TextProps{Content: "● All services are operational", Style: omnitui.Style{Foreground: omnitui.ANSI(omnitui.BrightGreen), Attributes: omnitui.Bold}})},
				{Key: "metrics", Label: "Metrics", Content: components.Text(components.TextProps{Content: "CPU  24%   Memory  61%   Latency  18ms", Style: omnitui.Style{Foreground: omnitui.ANSI(omnitui.BrightYellow)}})},
				{Key: "help", Label: "Help", Content: components.Text(components.TextProps{Content: "Use Tab to move focus and Enter to activate controls."})},
			},
			OnChange: func(event omnitui.ValueChangeEvent) omnitui.EventResult {
				omnitui.UpdateState(ctx, func(current styledState) styledState {
					current.ActiveKey = event.Value
					return current
				})
				return omnitui.Consume
			},
		}),
		components.Row(
			components.RowProps{Gap: 1, Align: components.AlignCenter},
			components.Text(components.TextProps{Content: "Name:", Style: omnitui.Style{Attributes: omnitui.Bold}}),
			components.Input(components.InputProps{
				Value:       state.Name,
				Placeholder: "Enter your name",
				Width:       omnitui.Cells(20),
				Style:       inputStyle,
				FocusStyle:  inputFocusStyle,
				OnChange: func(event omnitui.ValueChangeEvent) omnitui.EventResult {
					omnitui.UpdateState(ctx, func(current styledState) styledState {
						current.Name = event.Value
						current.Saved = false
						return current
					})
					return omnitui.Consume
				},
			}),
			components.Button(components.ButtonProps{
				Label:      "Save",
				Style:      buttonStyle,
				FocusStyle: buttonFocusStyle,
				OnPress: func(omnitui.PressEvent) omnitui.EventResult {
					omnitui.UpdateState(ctx, func(current styledState) styledState {
						current.Saved = true
						return current
					})
					return omnitui.Consume
				},
			}),
		),
		components.Text(components.TextProps{Content: status, Style: mutedStyle}),
	)
}

func main() {
	styledType := omnitui.Define("StyledExample", styledExample{})
	app := omnitui.New(omnitui.Create(styledType, "styled"), omnitui.Options{})
	if err := app.Run(context.Background()); err != nil {
		panic(err)
	}
}
