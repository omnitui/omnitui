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
	firstSize := components.GridItemProps{InitialSize: 18, MinSize: 8, MaxSize: 28}
	secondSize := components.GridItemProps{MinSize: 10}
	thirdSize := components.GridItemProps{InitialSize: 18, MinSize: 8, MaxSize: 24}
	if state.Vertical {
		orientation = components.OrientationVertical
		buttonLabel = "Use horizontal orientation"
		firstSize = components.GridItemProps{InitialSize: 4, MinSize: 3, MaxSize: 6}
		secondSize = components.GridItemProps{MinSize: 3}
		thirdSize = components.GridItemProps{InitialSize: 4, MinSize: 3, MaxSize: 6}
	}
	return components.Box(
		components.BoxProps{
			Direction: components.Vertical,
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
			components.GridItem(firstSize, components.Column(
				components.ColumnProps{},
				components.Text(components.TextProps{Content: "Explorer", Style: accentStyle}),
				components.Text(components.TextProps{Content: "Project files and folders", Wrap: components.WrapWord}),
			)),
			components.GridItem(secondSize, components.Column(
				components.ColumnProps{},
				components.Text(components.TextProps{Content: "Editor", Style: accentStyle}),
				components.Text(components.TextProps{Content: "The active document", Wrap: components.WrapWord}),
			)),
			components.GridItem(thirdSize, components.Column(
				components.ColumnProps{},
				components.Text(components.TextProps{Content: "Preview", Style: accentStyle}),
				components.Text(components.TextProps{Content: "Rendered output", Wrap: components.WrapWord}),
			)),
		),
	)
}

func main() {
	typeValue := omnitui.Define("GridExample", gridExample{})
	app := omnitui.New(omnitui.Create(typeValue, struct{}{}), omnitui.Options{})
	if err := app.Run(context.Background()); err != nil {
		panic(err)
	}
}
