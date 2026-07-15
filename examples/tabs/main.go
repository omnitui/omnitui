package main

import (
	"context"

	omnitui "github.com/viniciusfonseca/omnitui"
	"github.com/viniciusfonseca/omnitui/components"
)

type tabsState struct {
	ActiveKey string
}

type tabsExample struct{}

var tabsPanelStyle = omnitui.Style{
	Foreground: omnitui.RGB(224, 228, 238),
	Background: omnitui.RGB(20, 24, 33),
}

func (tabsExample) InitialState(string) tabsState {
	return tabsState{ActiveKey: "overview"}
}

func (tabsExample) Render(ctx omnitui.Context, _ string, state tabsState, _ omnitui.Children) omnitui.Element {
	return components.Column(
		components.ColumnProps{Gap: 1, Padding: omnitui.All(1), Style: tabsPanelStyle},
		components.Text(components.TextProps{Content: "Use ←/→ para trocar de aba e Enter para selecionar", Style: omnitui.Style{Foreground: omnitui.ANSI(omnitui.BrightBlack), Attributes: omnitui.Dim}}),
		components.Tabs(components.TabsProps{
			ActiveKey:   state.ActiveKey,
			Style:       omnitui.Style{Foreground: omnitui.ANSI(omnitui.BrightBlack)},
			ActiveStyle: omnitui.Style{Foreground: omnitui.ANSI(omnitui.BrightCyan), Attributes: omnitui.Bold | omnitui.Underline},
			Items: []components.TabItem{
				{
					Key:   "overview",
					Label: "Visão geral",
					Content: components.Column(
						components.ColumnProps{Gap: 1},
						components.Text(components.TextProps{Content: "Projeto: OmniTUI", Style: omnitui.Style{Foreground: omnitui.ANSI(omnitui.BrightCyan), Attributes: omnitui.Bold}}),
						components.Text(components.TextProps{Content: "Status: em desenvolvimento", Style: omnitui.Style{Foreground: omnitui.ANSI(omnitui.BrightGreen)}}),
					),
				},
				{
					Key:   "logs",
					Label: "Logs",
					Content: components.Column(
						components.ColumnProps{Gap: 1},
						components.Text(components.TextProps{Content: "[info] aplicação iniciada", Style: omnitui.Style{Foreground: omnitui.ANSI(omnitui.BrightYellow)}}),
						components.Text(components.TextProps{Content: "[info] nenhuma falha encontrada", Style: omnitui.Style{Foreground: omnitui.ANSI(omnitui.BrightGreen)}}),
					),
				},
				{
					Key:     "settings",
					Label:   "Configurações",
					Content: components.Text(components.TextProps{Content: "Tema: padrão | Mouse: habilitado", Style: omnitui.Style{Foreground: omnitui.ANSI(omnitui.BrightYellow)}}),
				},
			},
			OnChange: func(event omnitui.ValueChangeEvent) omnitui.EventResult {
				omnitui.UpdateState(ctx, func(current tabsState) tabsState {
					current.ActiveKey = event.Value
					return current
				})
				return omnitui.Consume
			},
		}),
	)
}

func main() {
	tabsType := omnitui.Define("TabsExample", tabsExample{})
	app := omnitui.New(omnitui.Create(tabsType, "tabs"), omnitui.Options{})
	if err := app.Run(context.Background()); err != nil {
		panic(err)
	}
}
