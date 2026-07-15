package layout

import "github.com/viniciusfonseca/omnitui/internal/core"

func Clip(rect, clip core.Rect) core.Rect { return core.IntersectRect(rect, clip) }
