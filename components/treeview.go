package components

import (
	"fmt"

	"github.com/omnitui/omnitui/v2"
	"github.com/omnitui/omnitui/v2/internal/core"
)

type TreeNode struct {
	Key      string
	Label    string
	Children []TreeNode
}

type TreeToggleEvent struct {
	Key      string
	Expanded bool
}

type TreeViewProps struct {
	Nodes         []TreeNode
	SelectedKey   string
	ExpandedKeys  []string
	Height        omnitui.Size
	Disabled      bool
	Wrap          bool
	ScrollPadding int
	Scrollbar     ScrollbarMode
	Empty         omnitui.Element
	Style         omnitui.Style
	SelectedStyle omnitui.Style
	Focus         omnitui.FocusHandle

	OnChange   omnitui.EventHandler[omnitui.ValueChangeEvent]
	OnToggle   omnitui.EventHandler[TreeToggleEvent]
	OnActivate omnitui.EventHandler[omnitui.ActivateEvent]
	OnMouse    omnitui.EventHandler[omnitui.MouseEvent]
	OnWheel    omnitui.EventHandler[omnitui.WheelEvent]
}

type treeViewComponent struct{}

func (treeViewComponent) InitialState(TreeViewProps) struct{} { return struct{}{} }
func (treeViewComponent) Render(_ omnitui.Context, props TreeViewProps, _ struct{}, _ omnitui.Children) omnitui.Element {
	expandedKeys, expandAll := treeExpandedKeys(props.ExpandedKeys)
	nodeInfo := treeNodeInfoByKey(props.Nodes)
	emitToggle := func(key string, expanded bool) {
		if props.OnToggle != nil {
			props.OnToggle(TreeToggleEvent{Key: key, Expanded: expanded})
		}
	}
	emitChange := func(key string) {
		if props.OnChange != nil && key != "" && key != props.SelectedKey {
			props.OnChange(omnitui.ValueChangeEvent{Previous: props.SelectedKey, Value: key, Source: omnitui.ChangeKeyboard})
		}
	}
	lines := flattenTreeNodes(props.Nodes, expandedKeys, expandAll)
	items := make([]omnitui.Element, len(lines))
	for index, line := range lines {
		style := omnitui.Style{}
		if line.Key == props.SelectedKey {
			style = props.SelectedStyle
		}
		var mouse omnitui.EventHandler[omnitui.MouseEvent]
		if line.HasChildren {
			key, expanded, toggleColumn := line.Key, line.Expanded, line.ToggleColumn
			mouse = func(event omnitui.MouseEvent) omnitui.EventResult {
				if !props.Disabled && event.Action == omnitui.MouseDown && event.Button == omnitui.MouseButtonLeft && event.LocalX >= toggleColumn && event.LocalX < toggleColumn+2 {
					emitToggle(key, !expanded)
				}
				return omnitui.Propagate
			}
		}
		items[index] = core.NewHost(core.HostTreeLine, core.TreeLineData{
			Prefix: line.Prefix, Label: line.Label, LabelStyle: style,
			Handlers: handlers(map[string]any{"mouse": mouse}),
		}, nil).WithKey(line.Key)
	}
	selectedRowStyle := omnitui.Style{Background: props.SelectedStyle.Background}
	list := List(ListProps{
		SelectedKey:   props.SelectedKey,
		Selectable:    true,
		Height:        props.Height,
		Disabled:      props.Disabled,
		Wrap:          props.Wrap,
		ScrollPadding: props.ScrollPadding,
		Scrollbar:     props.Scrollbar,
		Empty:         props.Empty,
		Style:         props.Style,
		SelectedStyle: selectedRowStyle,
		Focus:         props.Focus,
		OnChange:      props.OnChange,
		OnActivate:    props.OnActivate,
		OnMouse:       props.OnMouse,
		OnWheel:       props.OnWheel,
	}, items...)
	return Box(BoxProps{
		Direction: Vertical,
		Width:     omnitui.Fill(),
		Align:     AlignStretch,
		OnKey: func(event omnitui.KeyEvent) omnitui.EventResult {
			if props.Disabled {
				return omnitui.Propagate
			}
			info, ok := nodeInfo[props.SelectedKey]
			if !ok {
				return omnitui.Propagate
			}
			expanded := treeNodeExpanded(props.SelectedKey, expandedKeys, expandAll)
			switch event.Key {
			case omnitui.KeyRight:
				if !info.HasChildren {
					return omnitui.Consume
				}
				if !expanded {
					emitToggle(props.SelectedKey, true)
				} else {
					emitChange(info.FirstChild)
				}
				return omnitui.Consume
			case omnitui.KeyLeft:
				if info.HasChildren && expanded {
					emitToggle(props.SelectedKey, false)
				} else {
					emitChange(info.Parent)
				}
				return omnitui.Consume
			default:
				return omnitui.Propagate
			}
		},
	}, list)
}

var treeViewType = omnitui.Define[TreeViewProps, struct{}]("TreeView", treeViewComponent{})

func TreeView(props TreeViewProps) omnitui.Element {
	if props.ScrollPadding < 0 {
		panic("omnitui/components: tree view ScrollPadding cannot be negative")
	}
	validateStyle(props.Style)
	validateStyle(props.SelectedStyle)
	props.Nodes = cloneTreeNodes(props.Nodes)
	if props.ExpandedKeys != nil {
		props.ExpandedKeys = append([]string{}, props.ExpandedKeys...)
	}
	seen := make(map[string]struct{})
	validateTreeNodes(props.Nodes, seen)
	if props.SelectedKey != "" {
		if _, ok := seen[props.SelectedKey]; !ok {
			panic(fmt.Sprintf("omnitui/components: unknown tree node %q", props.SelectedKey))
		}
	}
	if props.ExpandedKeys != nil {
		nodeInfo := treeNodeInfoByKey(props.Nodes)
		expandedSeen := make(map[string]struct{}, len(props.ExpandedKeys))
		for _, key := range props.ExpandedKeys {
			info, ok := nodeInfo[key]
			if !ok {
				panic(fmt.Sprintf("omnitui/components: unknown expanded tree node %q", key))
			}
			if !info.HasChildren {
				panic(fmt.Sprintf("omnitui/components: expanded tree node %q has no children", key))
			}
			if _, exists := expandedSeen[key]; exists {
				panic(fmt.Sprintf("omnitui/components: duplicate expanded tree node %q", key))
			}
			expandedSeen[key] = struct{}{}
		}
	}
	return omnitui.Create(treeViewType, props)
}

type treeViewLine struct {
	Key          string
	Prefix       string
	Label        string
	HasChildren  bool
	Expanded     bool
	ToggleColumn int
}

func flattenTreeNodes(nodes []TreeNode, expandedKeys map[string]struct{}, expandAll bool) []treeViewLine {
	lines := make([]treeViewLine, 0)
	var appendNodes func([]TreeNode, string, bool)
	appendNodes = func(current []TreeNode, prefix string, roots bool) {
		for index, node := range current {
			last := index == len(current)-1
			linePrefix := prefix
			childPrefix := prefix
			if !roots {
				if last {
					linePrefix += "└─ "
					childPrefix += "   "
				} else {
					linePrefix += "├─ "
					childPrefix += "│  "
				}
			}
			hasChildren := len(node.Children) > 0
			expanded := hasChildren && treeNodeExpanded(node.Key, expandedKeys, expandAll)
			indicator := ""
			if expanded {
				indicator = "▾ "
			} else if hasChildren {
				indicator = "▸ "
			}
			lines = append(lines, treeViewLine{
				Key: node.Key, Prefix: linePrefix + indicator, Label: node.Label,
				HasChildren: hasChildren, Expanded: expanded, ToggleColumn: len([]rune(linePrefix)),
			})
			if expanded {
				appendNodes(node.Children, childPrefix, false)
			}
		}
	}
	appendNodes(nodes, "", true)
	return lines
}

type treeNodeInfo struct {
	Parent      string
	FirstChild  string
	HasChildren bool
}

func treeNodeInfoByKey(nodes []TreeNode) map[string]treeNodeInfo {
	result := make(map[string]treeNodeInfo)
	var collect func([]TreeNode, string)
	collect = func(current []TreeNode, parent string) {
		for _, node := range current {
			info := treeNodeInfo{Parent: parent, HasChildren: len(node.Children) > 0}
			if info.HasChildren {
				info.FirstChild = node.Children[0].Key
			}
			result[node.Key] = info
			collect(node.Children, node.Key)
		}
	}
	collect(nodes, "")
	return result
}

func treeExpandedKeys(keys []string) (map[string]struct{}, bool) {
	if keys == nil {
		return nil, true
	}
	result := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		result[key] = struct{}{}
	}
	return result, false
}

func treeNodeExpanded(key string, expandedKeys map[string]struct{}, expandAll bool) bool {
	if expandAll {
		return true
	}
	_, ok := expandedKeys[key]
	return ok
}

func validateTreeNodes(nodes []TreeNode, seen map[string]struct{}) {
	for _, node := range nodes {
		if node.Key == "" {
			panic("omnitui/components: tree node has an empty key")
		}
		if _, ok := seen[node.Key]; ok {
			panic(fmt.Sprintf("omnitui/components: duplicate tree node key %q", node.Key))
		}
		seen[node.Key] = struct{}{}
		validateTreeNodes(node.Children, seen)
	}
}

func cloneTreeNodes(nodes []TreeNode) []TreeNode {
	if len(nodes) == 0 {
		return nil
	}
	cloned := make([]TreeNode, len(nodes))
	for index, node := range nodes {
		cloned[index] = node
		cloned[index].Children = cloneTreeNodes(node.Children)
	}
	return cloned
}
