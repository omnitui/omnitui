package layout

import "github.com/omnitui/omnitui/v2/internal/core"

func Clip(rect, clip core.Rect) core.Rect { return core.IntersectRect(rect, clip) }
