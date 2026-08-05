package main

import (
	"context"
	"errors"

	omnitui "github.com/omnitui/omnitui/v2"
	"github.com/omnitui/omnitui/v2/components"
)

var projectNodes = []components.TreeNode{
	{
		Key: "omnitui", Label: "omnitui",
		Children: []components.TreeNode{
			{
				Key: "components", Label: "components",
				Children: []components.TreeNode{
					{Key: "treeview.go", Label: "treeview.go"},
					{Key: "list.go", Label: "list.go"},
				},
			},
			{
				Key: "examples", Label: "examples",
				Children: []components.TreeNode{
					{Key: "treeview-example", Label: "treeview/main.go"},
				},
			},
			{Key: "readme", Label: "README.md"},
		},
	},
}

type treeViewState struct {
	Selected string
	Expanded []string
}
type treeViewExample struct{}

func (treeViewExample) InitialState(struct{}) treeViewState {
	return treeViewState{
		Selected: "omnitui",
		Expanded: []string{"omnitui", "components", "examples"},
	}
}

func (treeViewExample) Render(ctx omnitui.Context, _ struct{}, state treeViewState, _ omnitui.Children) omnitui.Element {
	treeFocus := omnitui.UseFocus(ctx, "tree")
	omnitui.UseEffect(ctx, "focus-tree", struct{}{}, func(context.Context) omnitui.Cleanup {
		treeFocus.Request()
		return nil
	})

	return components.Box(
		components.BoxProps{
			Direction: components.Vertical,
			Border:    components.BorderRounded,
			Label:     "Project tree",
			Style: omnitui.Style{
				Foreground: omnitui.RGB(220, 226, 235),
				Background: omnitui.RGB(18, 23, 32),
			},
		},
		components.TreeView(components.TreeViewProps{
			Nodes:         projectNodes,
			SelectedKey:   state.Selected,
			ExpandedKeys:  state.Expanded,
			Height:        omnitui.Cells(8),
			ScrollPadding: 1,
			Scrollbar:     components.ScrollbarAuto,
			Focus:         treeFocus,
			SelectedStyle: omnitui.Style{
				Foreground: omnitui.ANSI(omnitui.BrightWhite),
				Background: omnitui.ANSI(omnitui.BrightBlack),
				Attributes: omnitui.Bold | omnitui.Underline,
			},
			OnChange: func(event omnitui.ValueChangeEvent) omnitui.EventResult {
				omnitui.UpdateState(ctx, func(current treeViewState) treeViewState {
					current.Selected = event.Value
					return current
				})
				return omnitui.Consume
			},
			OnToggle: func(event components.TreeToggleEvent) omnitui.EventResult {
				omnitui.UpdateState(ctx, func(current treeViewState) treeViewState {
					next := make([]string, 0, len(current.Expanded)+1)
					found := false
					for _, key := range current.Expanded {
						if key == event.Key {
							found = true
							if !event.Expanded {
								continue
							}
						}
						next = append(next, key)
					}
					if event.Expanded && !found {
						next = append(next, event.Key)
					}
					current.Expanded = next
					return current
				})
				return omnitui.Consume
			},
		}),
		components.Text(components.TextProps{Content: "Selected: " + state.Selected}),
		components.Text(components.TextProps{Content: "↑/↓ select • ←/→ or triangles collapse/expand • Ctrl+C exits"}),
	)
}

func main() {
	typeValue := omnitui.Define("TreeViewExample", treeViewExample{})
	app := omnitui.New(omnitui.Create(typeValue, struct{}{}), omnitui.Options{})
	if err := app.Run(context.Background()); err != nil && !errors.Is(err, omnitui.ErrInterrupted) {
		panic(err)
	}
}
