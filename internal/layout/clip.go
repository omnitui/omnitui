package layout

import "github.com/omnitui/omnitui/internal/core"

func Clip(rect, clip core.Rect) core.Rect { return core.IntersectRect(rect, clip) }
