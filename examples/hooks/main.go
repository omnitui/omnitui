package main

import (
	"context"
	"fmt"
	"time"

	omnitui "github.com/omnitui/omnitui"
	"github.com/omnitui/omnitui/components"
)

type theme struct {
	Title  omnitui.Style
	Muted  omnitui.Style
	Accent omnitui.Style
}

var themeContext = omnitui.NewContext(theme{})

type hooksProps struct {
	TickInterval time.Duration
}

type hooksState struct {
	Query         string
	Ticks         int
	FocusRequests int
}

type hooksDemo struct{}

func (hooksDemo) InitialState(hooksProps) hooksState { return hooksState{} }

func (hooksDemo) Render(ctx omnitui.Context, props hooksProps, state hooksState, _ omnitui.Children) omnitui.Element {
	colors := omnitui.UseContext(ctx, themeContext)
	viewport := omnitui.UseViewport(ctx)
	searchFocus := omnitui.UseFocus(ctx, "search")
	focusRequests := omnitui.UseRef(ctx, "focus-requests", 0)

	omnitui.UseEffect(ctx, "ticker", props.TickInterval, func(effectContext context.Context) omnitui.Cleanup {
		searchFocus.Request()
		ticker := time.NewTicker(props.TickInterval)
		go func() {
			for {
				select {
				case <-effectContext.Done():
					return
				case <-ticker.C:
					omnitui.UpdateState(ctx, func(current hooksState) hooksState {
						current.Ticks++
						return current
					})
				}
			}
		}()
		return ticker.Stop
	})

	focusStatus := "not focused"
	if searchFocus.Focused() {
		focusStatus = "focused"
	}
	help := "Tab moves focus • Enter activates • Ctrl+C exits"
	if viewport.Width < 55 {
		help = "Tab • Enter • Ctrl+C"
	}

	return components.Column(
		components.ColumnProps{Padding: omnitui.All(1), Gap: 1},
		components.Text(components.TextProps{Content: "OmniTUI hooks", Style: colors.Title}),
		components.Text(components.TextProps{
			Content: fmt.Sprintf("viewport %dx%d • efeito: %d ticks", viewport.Width, viewport.Height, state.Ticks),
			Style:   colors.Muted,
		}),
		components.Input(components.InputProps{
			Value:       state.Query,
			Placeholder: "Type to update the state...",
			Width:       omnitui.Fill(),
			Focus:       searchFocus,
			FocusStyle:  colors.Accent,
			OnChange: func(event omnitui.ValueChangeEvent) omnitui.EventResult {
				omnitui.UpdateState(ctx, func(current hooksState) hooksState {
					current.Query = event.Value
					return current
				})
				return omnitui.Consume
			},
		}),
		components.Row(
			components.RowProps{Gap: 1, Wrap: true},
			components.Button(components.ButtonProps{
				Label:      "Focus input",
				FocusStyle: colors.Accent,
				OnPress: func(omnitui.PressEvent) omnitui.EventResult {
					requests := focusRequests.Update(func(current int) int { return current + 1 })
					searchFocus.Request()
					omnitui.UpdateState(ctx, func(current hooksState) hooksState {
						current.FocusRequests = requests
						return current
					})
					return omnitui.Consume
				},
			}),
			components.Button(components.ButtonProps{
				Label:      "Clear focus",
				FocusStyle: colors.Accent,
				OnPress: func(omnitui.PressEvent) omnitui.EventResult {
					searchFocus.Blur()
					return omnitui.Consume
				},
			}),
		),
		components.Text(components.TextProps{
			Content: fmt.Sprintf("input %s • %d focus requests • value: %q", focusStatus, state.FocusRequests, state.Query),
		}),
		components.Text(components.TextProps{Content: help, Style: colors.Muted}),
	)
}

var hooksType = omnitui.Define[hooksProps, hooksState]("HooksDemo", hooksDemo{})

func main() {
	colors := theme{
		Title:  omnitui.Style{Foreground: omnitui.ANSI(omnitui.BrightCyan), Attributes: omnitui.Bold},
		Muted:  omnitui.Style{Foreground: omnitui.ANSI(omnitui.BrightBlack)},
		Accent: omnitui.Style{Foreground: omnitui.ANSI(omnitui.Black), Background: omnitui.ANSI(omnitui.BrightCyan), Attributes: omnitui.Bold},
	}
	root := omnitui.Provide(
		themeContext,
		colors,
		omnitui.Create(hooksType, hooksProps{TickInterval: time.Second}),
	)
	if err := omnitui.New(root, omnitui.Options{}).Run(context.Background()); err != nil {
		panic(err)
	}
}
