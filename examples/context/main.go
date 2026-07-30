package main

import (
	"context"
	"fmt"

	omnitui "github.com/omnitui/omnitui/v2"
	"github.com/omnitui/omnitui/v2/components"
)

type counterValue struct {
	Count     int
	Increment func()
}

var counterContext = omnitui.NewContext(counterValue{
	Increment: func() {},
})

type providerState struct {
	Count int
}

type counterProvider struct{}

func (counterProvider) InitialState(struct{}) providerState { return providerState{} }

func (counterProvider) Render(ctx omnitui.Context, _ struct{}, state providerState, _ omnitui.Children) omnitui.Element {
	value := counterValue{
		Count: state.Count,
		Increment: func() {
			omnitui.UpdateState(ctx, func(current providerState) providerState {
				current.Count++
				return current
			})
		},
	}

	return omnitui.Provide(
		counterContext,
		value,
		omnitui.Create(counterType, struct{}{}),
	)
}

var providerType = omnitui.Define("CounterProvider", counterProvider{})

type counter struct{}

func (counter) InitialState(struct{}) struct{} { return struct{}{} }

func (counter) Render(ctx omnitui.Context, _ struct{}, _ struct{}, _ omnitui.Children) omnitui.Element {
	value := omnitui.UseContext(ctx, counterContext)

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
			Content: fmt.Sprintf("Context count: %d", value.Count),
			Style: omnitui.Style{
				Foreground: omnitui.ANSI(omnitui.BrightCyan),
				Attributes: omnitui.Bold,
			},
		}),
		components.Button(components.ButtonProps{
			Label: "Increment through context",
			Style: omnitui.Style{
				Foreground: omnitui.ANSI(omnitui.Black),
				Background: omnitui.ANSI(omnitui.BrightGreen),
				Attributes: omnitui.Bold,
			},
			FocusStyle: omnitui.Style{
				Foreground: omnitui.ANSI(omnitui.Black),
				Background: omnitui.ANSI(omnitui.BrightWhite),
				Attributes: omnitui.Bold | omnitui.Underline,
			},
			OnPress: func(omnitui.PressEvent) omnitui.EventResult {
				value.Increment()
				return omnitui.Consume
			},
		}),
	)
}

var counterType = omnitui.Define("Counter", counter{})

func main() {
	root := omnitui.Create(providerType, struct{}{})
	if err := omnitui.New(root, omnitui.Options{}).Run(context.Background()); err != nil {
		panic(err)
	}
}
