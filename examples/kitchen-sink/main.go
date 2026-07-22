package main

import (
	"context"
	"fmt"

	omnitui "github.com/omnitui/omnitui/v2"
	"github.com/omnitui/omnitui/v2/components"
)

var (
	appStyle = omnitui.Style{
		Foreground: omnitui.RGB(225, 231, 239),
		Background: omnitui.RGB(15, 20, 29),
	}
	panelStyle = omnitui.Style{
		Foreground: omnitui.RGB(225, 231, 239),
		Background: omnitui.RGB(25, 32, 46),
	}
	mutedStyle = omnitui.Style{
		Foreground: omnitui.ANSI(omnitui.BrightBlack),
		Attributes: omnitui.Dim,
	}
	accentStyle = omnitui.Style{
		Foreground: omnitui.ANSI(omnitui.BrightCyan),
		Attributes: omnitui.Bold,
	}
	activeTabStyle = omnitui.Style{
		Foreground: omnitui.ANSI(omnitui.Black),
		Background: omnitui.ANSI(omnitui.BrightCyan),
		Attributes: omnitui.Bold,
	}
	inputStyle = omnitui.Style{
		Foreground: omnitui.ANSI(omnitui.White),
		Background: omnitui.RGB(38, 49, 70),
	}
	inputFocusStyle = omnitui.Style{
		Foreground: omnitui.ANSI(omnitui.BrightWhite),
		Background: omnitui.RGB(45, 84, 125),
		Attributes: omnitui.Underline,
	}
	buttonStyle = omnitui.Style{
		Foreground: omnitui.ANSI(omnitui.Black),
		Background: omnitui.ANSI(omnitui.BrightGreen),
		Attributes: omnitui.Bold,
	}
	buttonFocusStyle = omnitui.Style{
		Foreground: omnitui.ANSI(omnitui.Black),
		Background: omnitui.ANSI(omnitui.BrightWhite),
		Attributes: omnitui.Bold | omnitui.Underline,
	}
)

type kitchenState struct {
	ActiveTab string
	Selected  string
	Name      string
	Notice    string
	Clicks    int
}

type kitchenSink struct{}

func (kitchenSink) InitialState(string) kitchenState {
	return kitchenState{
		ActiveTab: "overview",
		Selected:  "runtime",
		Notice:    "Interact with the controls to explore the component states.",
	}
}

func (kitchenSink) Render(ctx omnitui.Context, _ string, state kitchenState, _ omnitui.Children) omnitui.Element {
	return components.Box(
		components.BoxProps{
			Direction: components.Vertical,
			Padding:   omnitui.All(1),
			Gap:       1,
			Border:    components.BorderDouble,
			Clip:      true,
			Style:     appStyle,
		},
		components.Text(components.TextProps{
			Content: "OmniTUI • Kitchen sink",
			Style: omnitui.Style{
				Foreground: omnitui.ANSI(omnitui.BrightCyan),
				Attributes: omnitui.Bold | omnitui.Underline,
			},
		}),
		components.Text(components.TextProps{
			Content:  "Every builtin component in one interactive screen: layout, state, focus, input, selection, borders, colors, and text rendering.",
			Style:    mutedStyle,
			Wrap:     components.WrapWord,
			MaxLines: 2,
			Truncate: components.TruncateEllipsis,
		}),
		components.Tabs(components.TabsProps{
			ActiveKey:   state.ActiveTab,
			Style:       omnitui.Style{Foreground: omnitui.ANSI(omnitui.BrightBlack)},
			ActiveStyle: activeTabStyle,
			Items: []components.TabItem{
				{
					Key:   "overview",
					Label: "Overview",
					Content: components.Column(
						components.ColumnProps{Gap: 1, Padding: omnitui.All(1), Style: panelStyle},
						components.Text(components.TextProps{Content: "Declarative UI, deterministic reconciliation, and incremental terminal rendering.", Style: accentStyle, Wrap: components.WrapWord}),
						components.Text(components.TextProps{Content: "This panel is a controlled tab: its active key lives in the parent state.", Style: mutedStyle, Wrap: components.WrapWord}),
					),
				},
				{
					Key:   "events",
					Label: "Events",
					Content: components.Column(
						components.ColumnProps{Gap: 1, Padding: omnitui.All(1), Style: panelStyle},
						components.Text(components.TextProps{Content: "Keyboard, mouse, wheel, text input, paste, focus, and submit events are supported.", Style: omnitui.Style{Foreground: omnitui.ANSI(omnitui.BrightYellow), Attributes: omnitui.Bold}, Wrap: components.WrapWord}),
						components.Text(components.TextProps{Content: "Try Tab, arrow keys, Enter, Space, and the mouse.", Style: mutedStyle}),
					),
				},
				{
					Key:   "rendering",
					Label: "Rendering",
					Content: components.Column(
						components.ColumnProps{Gap: 1, Padding: omnitui.All(1), Style: panelStyle},
						components.Text(components.TextProps{Content: "Cells carry grapheme width and inherited styles; only changed cells are flushed.", Style: omnitui.Style{Foreground: omnitui.ANSI(omnitui.BrightGreen), Attributes: omnitui.Bold}, Wrap: components.WrapWord}),
						components.Text(components.TextProps{Content: "The examples below exercise borders, clipping, alignment, wrapping, and truncation.", Style: mutedStyle, Wrap: components.WrapWord}),
					),
				},
			},
			OnChange: func(event omnitui.ValueChangeEvent) omnitui.EventResult {
				omnitui.UpdateState(ctx, func(current kitchenState) kitchenState {
					current.ActiveTab = event.Value
					return current
				})
				return omnitui.Consume
			},
		}),
		components.Row(
			components.RowProps{Gap: 1, Align: components.AlignStart},
			listPanel(ctx, state),
			stylePanel(),
		),
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
					omnitui.UpdateState(ctx, func(current kitchenState) kitchenState {
						current.Name = event.Value
						current.Notice = "Unsaved input: " + event.Value
						return current
					})
					return omnitui.Consume
				},
				OnSubmit: func(event omnitui.SubmitEvent) omnitui.EventResult {
					omnitui.UpdateState(ctx, func(current kitchenState) kitchenState {
						current.Notice = "Submitted: " + event.Value
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
					omnitui.UpdateState(ctx, func(current kitchenState) kitchenState {
						current.Notice = "Saved: " + current.Name
						return current
					})
					return omnitui.Consume
				},
			}),
			components.Button(components.ButtonProps{
				Label:      fmt.Sprintf("Clicks: %d", state.Clicks),
				Style:      omnitui.Style{Foreground: omnitui.ANSI(omnitui.Black), Background: omnitui.ANSI(omnitui.BrightMagenta), Attributes: omnitui.Bold},
				FocusStyle: omnitui.Style{Foreground: omnitui.ANSI(omnitui.Black), Background: omnitui.ANSI(omnitui.BrightWhite), Attributes: omnitui.Bold | omnitui.Reverse},
				OnPress: func(omnitui.PressEvent) omnitui.EventResult {
					omnitui.UpdateState(ctx, func(current kitchenState) kitchenState {
						current.Clicks++
						current.Notice = "Button pressed; local state updated."
						return current
					})
					return omnitui.Consume
				},
			}),
		),
		components.Text(components.TextProps{Content: state.Notice, Style: mutedStyle, Wrap: components.WrapWord}),
	)
}

func listPanel(ctx omnitui.Context, state kitchenState) omnitui.Element {
	return components.Box(
		components.BoxProps{
			Width:     omnitui.Cells(36),
			Direction: components.Vertical,
			Padding:   omnitui.All(1),
			Gap:       1,
			Border:    components.BorderRounded,
			Style:     panelStyle,
		},
		components.Text(components.TextProps{Content: "List • selectable + scrollable", Style: accentStyle}),
		components.List(
			components.ListProps{
				SelectedKey:   state.Selected,
				Selectable:    true,
				Height:        omnitui.Cells(5),
				ScrollPadding: 1,
				Scrollbar:     components.ScrollbarAlways,
				Style:         omnitui.Style{Foreground: omnitui.ANSI(omnitui.White)},
				SelectedStyle: omnitui.Style{Foreground: omnitui.ANSI(omnitui.Black), Background: omnitui.ANSI(omnitui.BrightCyan), Attributes: omnitui.Bold},
				Empty:         components.Text(components.TextProps{Content: "No items"}),
				OnChange: func(event omnitui.ValueChangeEvent) omnitui.EventResult {
					omnitui.UpdateState(ctx, func(current kitchenState) kitchenState {
						current.Selected = event.Value
						current.Notice = "Selected item: " + event.Value
						return current
					})
					return omnitui.Consume
				},
				OnActivate: func(event omnitui.ActivateEvent) omnitui.EventResult {
					omnitui.UpdateState(ctx, func(current kitchenState) kitchenState {
						current.Notice = "Activated item: " + event.Key
						return current
					})
					return omnitui.Consume
				},
			},
			components.Text(components.TextProps{Content: "Runtime"}).WithKey("runtime"),
			components.Text(components.TextProps{Content: "Components"}).WithKey("components"),
			components.Text(components.TextProps{Content: "Terminal"}).WithKey("terminal"),
			components.Text(components.TextProps{Content: "Headless tests"}).WithKey("tests"),
			components.Text(components.TextProps{Content: "Documentation"}).WithKey("docs"),
		),
	)
}

func stylePanel() omnitui.Element {
	return components.Box(
		components.BoxProps{
			Width:     omnitui.Cells(42),
			Direction: components.Vertical,
			Padding:   omnitui.All(1),
			Gap:       1,
			Border:    components.BorderHeavy,
			Clip:      true,
			Style:     panelStyle,
		},
		components.Text(components.TextProps{Content: "Text • colors + attributes", Style: accentStyle}),
		components.Text(components.TextProps{Content: "ANSI bright colors", Style: omnitui.Style{Foreground: omnitui.ANSI(omnitui.BrightYellow)}}),
		components.Text(components.TextProps{Content: "RGB foreground and background", Style: omnitui.Style{Foreground: omnitui.RGB(255, 160, 90), Background: omnitui.RGB(65, 32, 45)}}),
		components.Text(components.TextProps{Content: "Bold • italic • underline", Style: omnitui.Style{Attributes: omnitui.Bold | omnitui.Italic | omnitui.Underline}}),
		components.Text(components.TextProps{Content: "Reverse and strikethrough", Style: omnitui.Style{Foreground: omnitui.ANSI(omnitui.BrightMagenta), Attributes: omnitui.Reverse | omnitui.Strikethrough}}),
		components.Text(components.TextProps{
			Content:  "Centered text with word wrapping demonstrates clipping and truncation in a constrained box.",
			Align:    components.TextAlignCenter,
			Wrap:     components.WrapWord,
			MaxLines: 2,
			Truncate: components.TruncateEllipsis,
			Style:    omnitui.Style{Foreground: omnitui.ANSI(omnitui.BrightGreen)},
		}),
		components.Row(
			components.RowProps{Gap: 1, Align: components.AlignCenter, Justify: components.JustifySpaceAround},
			borderSample(components.BorderSingle, "Single", omnitui.ANSI(omnitui.BrightCyan)),
			borderSample(components.BorderRounded, "Rounded", omnitui.ANSI(omnitui.BrightGreen)),
			borderSample(components.BorderHeavy, "Heavy", omnitui.ANSI(omnitui.BrightYellow)),
		),
	)
}

func borderSample(border components.BorderStyle, label string, color omnitui.Color) omnitui.Element {
	return components.Box(
		components.BoxProps{
			Width:   omnitui.Cells(12),
			Padding: omnitui.All(1),
			Border:  border,
			Style:   omnitui.Style{Foreground: color},
		},
		components.Text(components.TextProps{Content: label, Align: components.TextAlignCenter}),
	)
}

func main() {
	kitchenType := omnitui.Define("KitchenSink", kitchenSink{})
	app := omnitui.New(omnitui.Create(kitchenType, "kitchen-sink"), omnitui.Options{})
	if err := app.Run(context.Background()); err != nil {
		panic(err)
	}
}
