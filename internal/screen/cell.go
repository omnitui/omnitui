package screen

import "github.com/viniciusfonseca/omnitui/internal/core"

type Cell struct {
	Grapheme string
	Width    int
	Style    core.Style
}

func blank(style core.Style) Cell { return Cell{Grapheme: " ", Width: 1, Style: style} }
