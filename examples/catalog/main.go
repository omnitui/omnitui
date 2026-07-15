package main

import (
	"context"

	omnitui "github.com/viniciusfonseca/omnitui"
	"github.com/viniciusfonseca/omnitui/components"
)

type catalogState struct{ Selected string }
type catalog struct{}

func (catalog) InitialState(string) catalogState { return catalogState{Selected: "omnitui"} }
func (catalog) Render(ctx omnitui.Context, _ string, state catalogState, _ omnitui.Children) omnitui.Element {
	return components.Column(components.ColumnProps{Padding: omnitui.All(1), Gap: 1, Style: omnitui.Style{Foreground: omnitui.RGB(224, 228, 238), Background: omnitui.RGB(20, 24, 33)}},
		components.Text(components.TextProps{Content: "Catálogo de exemplos", Style: omnitui.Style{Foreground: omnitui.ANSI(omnitui.BrightCyan), Attributes: omnitui.Bold}}),
		components.List(components.ListProps{
			SelectedKey:   state.Selected,
			Height:        omnitui.Cells(6),
			ScrollPadding: 1,
			Style:         omnitui.Style{Foreground: omnitui.ANSI(omnitui.White)},
			SelectedStyle: omnitui.Style{Foreground: omnitui.ANSI(omnitui.Black), Background: omnitui.ANSI(omnitui.BrightCyan), Attributes: omnitui.Bold},
			OnChange: func(event omnitui.ValueChangeEvent) omnitui.EventResult {
				omnitui.UpdateState(ctx, func(current catalogState) catalogState { current.Selected = event.Value; return current })
				return omnitui.Consume
			},
		},
			components.Text(components.TextProps{Content: "OmniTUI"}).WithKey("omnitui"),
			components.Text(components.TextProps{Content: "CLI Tools"}).WithKey("cli-tools"),
			components.Text(components.TextProps{Content: "Experimentos"}).WithKey("labs"),
			components.Text(components.TextProps{Content: "Documentação"}).WithKey("docs"),
			components.Text(components.TextProps{Content: "Benchmarks"}).WithKey("benchmarks"),
		),
	)
}

func main() {
	typeValue := omnitui.Define[string, catalogState]("Catalog", catalog{})
	app := omnitui.New(omnitui.Create(typeValue, "catalog"), omnitui.Options{})
	if err := app.Run(context.Background()); err != nil {
		panic(err)
	}
}
