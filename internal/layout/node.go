package layout

import "github.com/omnitui/omnitui/internal/core"

type Node struct {
	Rect     core.Rect
	Clip     core.Rect
	Children []*Node
}
