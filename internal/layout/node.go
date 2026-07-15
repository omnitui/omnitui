package layout

import "github.com/viniciusfonseca/omnitui/internal/core"

type Node struct {
	Rect     core.Rect
	Clip     core.Rect
	Children []*Node
}
