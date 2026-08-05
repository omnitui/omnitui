package components

import (
	"context"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	omnitui "github.com/omnitui/omnitui/v2"
	"github.com/omnitui/omnitui/v2/internal/core"
)

func TestTreeViewValidatesKeysAndSelection(t *testing.T) {
	mustPanic(t, func() {
		TreeView(TreeViewProps{Nodes: []TreeNode{{Label: "Missing key"}}})
	})
	mustPanic(t, func() {
		TreeView(TreeViewProps{Nodes: []TreeNode{
			{Key: "same"},
			{Key: "parent", Children: []TreeNode{{Key: "same"}}},
		}})
	})
	mustPanic(t, func() {
		TreeView(TreeViewProps{Nodes: []TreeNode{{Key: "known"}}, SelectedKey: "unknown"})
	})
	mustPanic(t, func() {
		TreeView(TreeViewProps{Nodes: []TreeNode{{Key: "known"}}, ExpandedKeys: []string{"unknown"}})
	})
	mustPanic(t, func() {
		TreeView(TreeViewProps{Nodes: []TreeNode{{Key: "leaf"}}, ExpandedKeys: []string{"leaf"}})
	})
	mustPanic(t, func() {
		TreeView(TreeViewProps{
			Nodes:        []TreeNode{{Key: "branch", Children: []TreeNode{{Key: "leaf"}}}},
			ExpandedKeys: []string{"branch", "branch"},
		})
	})
}

func TestTreeViewCopiesNodes(t *testing.T) {
	nodes := []TreeNode{{Key: "root", Label: "Root", Children: []TreeNode{{Key: "child", Label: "Child"}}}}
	expanded := []string{"root"}
	element := TreeView(TreeViewProps{Nodes: nodes, ExpandedKeys: expanded})
	nodes[0].Label = "Changed"
	nodes[0].Children[0].Label = "Changed child"
	expanded[0] = "changed"

	props := core.PropsOf(element).(TreeViewProps)
	if props.Nodes[0].Label != "Root" || props.Nodes[0].Children[0].Label != "Child" {
		t.Fatalf("stored nodes changed with caller slice: %+v", props.Nodes)
	}
	if !reflect.DeepEqual(props.ExpandedKeys, []string{"root"}) {
		t.Fatalf("stored expanded keys changed with caller slice: %+v", props.ExpandedKeys)
	}
	emptyProps := core.PropsOf(TreeView(TreeViewProps{Nodes: props.Nodes, ExpandedKeys: []string{}})).(TreeViewProps)
	if emptyProps.ExpandedKeys == nil {
		t.Fatal("non-nil empty ExpandedKeys became nil")
	}
}

func TestTreeViewFlattensNodesWithConnectors(t *testing.T) {
	lines := flattenTreeNodes([]TreeNode{{
		Key: "root", Label: "Root",
		Children: []TreeNode{
			{Key: "first", Label: "First", Children: []TreeNode{{Key: "leaf", Label: "Leaf"}}},
			{Key: "last", Label: "Last"},
		},
	}}, nil, true)
	want := []treeViewLine{
		{Key: "root", Prefix: "▾ ", Label: "Root", HasChildren: true, Expanded: true},
		{Key: "first", Prefix: "├─ ▾ ", Label: "First", HasChildren: true, Expanded: true, ToggleColumn: 3},
		{Key: "leaf", Prefix: "│  └─ ", Label: "Leaf", ToggleColumn: 6},
		{Key: "last", Prefix: "└─ ", Label: "Last", ToggleColumn: 3},
	}
	if len(lines) != len(want) {
		t.Fatalf("line count = %d, want %d", len(lines), len(want))
	}
	for index := range want {
		if lines[index] != want[index] {
			t.Fatalf("line %d = %+v, want %+v", index, lines[index], want[index])
		}
	}
}

func TestTreeViewHidesChildrenOfCollapsedNodes(t *testing.T) {
	nodes := []TreeNode{{
		Key: "root", Label: "Root",
		Children: []TreeNode{{Key: "branch", Label: "Branch", Children: []TreeNode{{Key: "leaf", Label: "Leaf"}}}},
	}}
	lines := flattenTreeNodes(nodes, map[string]struct{}{"root": {}}, false)
	want := []treeViewLine{
		{Key: "root", Prefix: "▾ ", Label: "Root", HasChildren: true, Expanded: true},
		{Key: "branch", Prefix: "└─ ▸ ", Label: "Branch", HasChildren: true, ToggleColumn: 3},
	}
	if !reflect.DeepEqual(lines, want) {
		t.Fatalf("collapsed lines = %+v, want %+v", lines, want)
	}
}

func TestTreeViewBuildsSelectableListWithoutGaps(t *testing.T) {
	selectedStyle := omnitui.Style{
		Background: omnitui.ANSI(omnitui.BrightCyan),
		Attributes: omnitui.Underline,
	}
	props := TreeViewProps{
		Nodes:         []TreeNode{{Key: "root", Label: "Root", Children: []TreeNode{{Key: "child", Label: "Child"}}}},
		SelectedKey:   "child",
		ExpandedKeys:  []string{"root"},
		SelectedStyle: selectedStyle,
	}
	element := (treeViewComponent{}).Render(omnitui.Context{}, props, struct{}{}, nil)
	wrapper, ok := core.HostOf(element)
	if !ok || wrapper.Kind != core.HostBox || len(wrapper.Children) != 1 {
		t.Fatalf("tree wrapper = %+v", wrapper)
	}
	wrapperData := wrapper.Data.(core.BoxData)
	if core.SizeModeOf(wrapperData.Width) != core.SizeFill || wrapperData.Align != uint8(AlignStretch) {
		t.Fatalf("tree wrapper layout = %+v", wrapperData)
	}
	list := wrapper.Children[0]
	listProps := core.PropsOf(list).(ListProps)
	wantRowStyle := omnitui.Style{Background: selectedStyle.Background}
	if !listProps.Selectable || listProps.Gap != 0 || listProps.SelectedKey != "child" || listProps.SelectedStyle != wantRowStyle {
		t.Fatalf("list props = %+v", listProps)
	}
	items := core.ChildrenOf(list)
	if len(items) != 2 || core.KeyOf(items[1]) != "child" {
		t.Fatalf("tree items = %+v", items)
	}
	selectedItem, ok := core.HostOf(items[1])
	if !ok || selectedItem.Kind != core.HostTreeLine {
		t.Fatalf("selected tree item = %+v", selectedItem)
	}
	selectedLine := selectedItem.Data.(core.TreeLineData)
	if selectedLine.Prefix != "└─ " || selectedLine.Label != "Child" {
		t.Fatalf("child item did not contain the rendered tree line")
	}
	if selectedLine.LabelStyle != selectedStyle {
		t.Fatalf("selected label style = %+v, want %+v", selectedLine.LabelStyle, selectedStyle)
	}
	unselectedItem, _ := core.HostOf(items[0])
	if unselectedItem.Data.(core.TreeLineData).LabelStyle != (omnitui.Style{}) {
		t.Fatalf("unselected label style = %+v, want zero", unselectedItem.Data.(core.TreeLineData).LabelStyle)
	}
}

type treeViewHarnessState struct {
	Selected string
	Expanded []string
}
type treeViewHarness struct {
	changed *string
	toggled *TreeToggleEvent
}

func (treeViewHarness) InitialState(struct{}) treeViewHarnessState {
	return treeViewHarnessState{Selected: "root", Expanded: []string{"root"}}
}

func (h treeViewHarness) Render(ctx omnitui.Context, _ struct{}, state treeViewHarnessState, _ omnitui.Children) omnitui.Element {
	treeFocus := omnitui.UseFocus(ctx, "tree")
	omnitui.UseEffect(ctx, "focus-tree", struct{}{}, func(context.Context) omnitui.Cleanup {
		treeFocus.Request()
		return nil
	})
	return TreeView(TreeViewProps{
		Nodes: []TreeNode{{
			Key: "root", Label: "Root",
			Children: []TreeNode{{Key: "child", Label: "Child"}},
		}},
		SelectedKey:  state.Selected,
		ExpandedKeys: state.Expanded,
		Focus:        treeFocus,
		OnChange: func(event omnitui.ValueChangeEvent) omnitui.EventResult {
			if h.changed != nil {
				*h.changed = event.Value
			}
			omnitui.UpdateState(ctx, func(current treeViewHarnessState) treeViewHarnessState {
				current.Selected = event.Value
				return current
			})
			return omnitui.Consume
		},
		OnToggle: func(event TreeToggleEvent) omnitui.EventResult {
			if h.toggled != nil {
				*h.toggled = event
			}
			omnitui.UpdateState(ctx, func(current treeViewHarnessState) treeViewHarnessState {
				if event.Expanded {
					current.Expanded = append(current.Expanded, event.Key)
				} else {
					current.Expanded = []string{}
				}
				return current
			})
			return omnitui.Consume
		},
	})
}

func TestTreeViewSelectsWithKeyboard(t *testing.T) {
	if changed := runTreeViewHarness(t, "\x1b[B\x03"); changed != "child" {
		t.Fatalf("selected node = %q, want child", changed)
	}
	if changed := runTreeViewHarness(t, "\x1b[C\x03"); changed != "child" {
		t.Fatalf("selected node after right arrow = %q, want child", changed)
	}
}

func TestTreeViewSelectsWithMouse(t *testing.T) {
	if changed := runTreeViewHarness(t, "\x1b[<0;2;2M\x1b[<0;2;2m\x03"); changed != "child" {
		t.Fatalf("selected node = %q, want child", changed)
	}
}

func TestTreeViewTogglesWithKeyboardAndMouse(t *testing.T) {
	keyboard := runTreeViewToggleHarness(t, "\x1b[D\x03")
	if keyboard != (TreeToggleEvent{Key: "root", Expanded: false}) {
		t.Fatalf("keyboard toggle = %+v", keyboard)
	}
	mouse := runTreeViewToggleHarness(t, "\x1b[<0;1;1M\x1b[<0;1;1m\x03")
	if mouse != (TreeToggleEvent{Key: "root", Expanded: false}) {
		t.Fatalf("mouse toggle = %+v", mouse)
	}
	reopened := runTreeViewToggleHarness(t, "\x1b[D\x1b[C\x03")
	if reopened != (TreeToggleEvent{Key: "root", Expanded: true}) {
		t.Fatalf("reopen toggle = %+v", reopened)
	}
}

func runTreeViewHarness(t *testing.T, inputSequence string) string {
	t.Helper()
	changed := ""
	typeValue := omnitui.Define[struct{}, treeViewHarnessState]("TreeViewHarness", treeViewHarness{changed: &changed})
	app := omnitui.New(
		omnitui.Create(typeValue, struct{}{}),
		omnitui.Options{Input: strings.NewReader(inputSequence), Output: io.Discard},
	)
	if err := app.Run(context.Background()); !errors.Is(err, omnitui.ErrInterrupted) {
		t.Fatalf("Run() error = %v, want ErrInterrupted", err)
	}
	return changed
}

func runTreeViewToggleHarness(t *testing.T, inputSequence string) TreeToggleEvent {
	t.Helper()
	toggled := TreeToggleEvent{}
	typeValue := omnitui.Define[struct{}, treeViewHarnessState]("TreeViewToggleHarness", treeViewHarness{toggled: &toggled})
	app := omnitui.New(
		omnitui.Create(typeValue, struct{}{}),
		omnitui.Options{Input: strings.NewReader(inputSequence), Output: io.Discard},
	)
	if err := app.Run(context.Background()); !errors.Is(err, omnitui.ErrInterrupted) {
		t.Fatalf("Run() error = %v, want ErrInterrupted", err)
	}
	return toggled
}
