package main

import (
	"context"

	"github.com/omnitui/omnitui/v2"
	"github.com/omnitui/omnitui/v2/components"
)

var (
	surfaceStyle = omnitui.Style{
		Foreground: omnitui.RGB(225, 231, 239),
		Background: omnitui.RGB(15, 20, 29),
	}
	panelStyle = omnitui.Style{
		Foreground: omnitui.RGB(205, 214, 226),
		Background: omnitui.RGB(24, 31, 43),
	}
	accentStyle = omnitui.Style{
		Foreground: omnitui.ANSI(omnitui.BrightCyan),
		Attributes: omnitui.Bold,
	}
)

type gridState struct{ Vertical bool }
type gridExample struct{}

func (gridExample) InitialState(struct{}) gridState { return gridState{} }

func (gridExample) Render(ctx omnitui.Context, _ struct{}, state gridState, _ omnitui.Children) omnitui.Element {
	orientation := components.OrientationHorizontal
	buttonLabel := "Use vertical orientation"
	if state.Vertical {
		orientation = components.OrientationVertical
		buttonLabel = "Use horizontal orientation"
	}
	return components.Box(
		components.BoxProps{
			Direction: components.Vertical,
			Padding:   omnitui.All(1),
			Gap:       1,
			Border:    components.BorderRounded,
			Label:     "Resizable grid",
			Style:     surfaceStyle,
		},
		components.Text(components.TextProps{
			Content: "Drag an internal border with the left mouse button.",
			Style:   accentStyle,
		}),
		components.Button(components.ButtonProps{
			Label: buttonLabel,
			OnPress: func(omnitui.PressEvent) omnitui.EventResult {
				omnitui.UpdateState(ctx, func(current gridState) gridState {
					current.Vertical = !current.Vertical
					return current
				})
				return omnitui.Consume
			},
		}),
		components.Grid(
			components.GridProps{
				Width:        omnitui.Fill(),
				Height:       omnitui.Cells(12),
				Orientation:  orientation,
				MinPanelSize: 5,
				Border:       components.BorderSingle,
				Style:        panelStyle,
			},
			panel("Explorer", "Project files and folders"),
			panel("Editor", "The active document"),
			panel("Preview", "Rendered output"),
		),
	)
}

func panel(title, description string) omnitui.Element {
	return components.Column(
		components.ColumnProps{Padding: omnitui.All(1), Gap: 1},
		components.Text(components.TextProps{Content: title, Style: accentStyle}),
		components.Text(components.TextProps{Content: description, Wrap: components.WrapWord}),
	)
}

func main() {
	typeValue := omnitui.Define("GridExample", gridExample{})
	app := omnitui.New(omnitui.Create(typeValue, struct{}{}), omnitui.Options{})
	if err := app.Run(context.Background()); err != nil {
		panic(err)
	}
}
