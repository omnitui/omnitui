package omnitui

import (
	"github.com/omnitui/omnitui/v2/internal/core"
)

type Element = core.Element
type Children []Element

func None() Element                        { return core.None() }
func Fragment(children ...Element) Element { return core.Fragment(children...) }
