package main

import (
	"context"
	"fmt"

	omnitui "github.com/viniciusfonseca/omnitui"
	"github.com/viniciusfonseca/omnitui/components"
)

type counterProps struct{ Label string }
type counterState struct{ Value int }
type counter struct{}

var counterPanelStyle = omnitui.Style{
	Foreground: omnitui.RGB(224, 228, 238),
	Background: omnitui.RGB(20, 24, 33),
}

func (counter) InitialState(counterProps) counterState { return counterState{} }
func (counter) Render(ctx omnitui.Context, props counterProps, state counterState, children omnitui.Children) omnitui.Element {
	return components.Column(
		components.ColumnProps{Gap: 1, Padding: omnitui.All(1), Style: counterPanelStyle},
		components.Text(components.TextProps{
			Content: fmt.Sprintf("%s: %d", props.Label, state.Value),
			Style:   omnitui.Style{Foreground: omnitui.ANSI(omnitui.BrightCyan), Attributes: omnitui.Bold},
		}),
		components.Button(components.ButtonProps{
			Label:      "Incrementar",
			Style:      omnitui.Style{Foreground: omnitui.ANSI(omnitui.Black), Background: omnitui.ANSI(omnitui.BrightGreen), Attributes: omnitui.Bold},
			FocusStyle: omnitui.Style{Foreground: omnitui.ANSI(omnitui.Black), Background: omnitui.ANSI(omnitui.BrightWhite), Attributes: omnitui.Bold | omnitui.Underline},
			OnPress: func(omnitui.PressEvent) omnitui.EventResult {
				omnitui.UpdateState(ctx, func(current counterState) counterState { current.Value++; return current })
				return omnitui.Consume
			},
		}),
		omnitui.Fragment(children...),
	)
}

var counterType = omnitui.Define[counterProps, counterState]("Counter", counter{})

func main() {
	app := omnitui.New(omnitui.Create(counterType, counterProps{Label: "Cliques"}), omnitui.Options{})
	if err := app.Run(context.Background()); err != nil {
		panic(err)
	}
}
