package components

import (
	"testing"

	"github.com/omnitui/omnitui/v2"
	"github.com/omnitui/omnitui/v2/internal/core"
)

func TestGridWrapsChildrenInBorderedPanels(t *testing.T) {
	child := Text(TextProps{Content: "first"}).WithKey("first")
	element := gridHost(GridProps{
		Orientation:  OrientationHorizontal,
		MinPanelSize: 3,
		Border:       BorderHeavy,
	}, child)
	host, ok := core.HostOf(element)
	if !ok || host.Kind != core.HostGrid {
		t.Fatal("Grid did not create a grid host")
	}
	if len(host.Children) != 1 {
		t.Fatalf("grid children = %d, want 1", len(host.Children))
	}
	panel, ok := core.HostOf(host.Children[0])
	if !ok {
		t.Fatal("grid child is not a host panel")
	}
	data := panel.Data.(core.BoxData)
	if data.Border != uint8(BorderHeavy) || !data.Clip {
		t.Fatalf("panel border=%d clip=%v", data.Border, data.Clip)
	}
	if core.KeyOf(host.Children[0]) != "first" {
		t.Fatalf("panel key = %q, want first", core.KeyOf(host.Children[0]))
	}
}

func TestGridDefaultsAndValidation(t *testing.T) {
	element := Grid(GridProps{}, omnitui.None())
	props := core.PropsOf(element).(GridProps)
	if props.MinPanelSize != 3 || props.Border != BorderSingle {
		t.Fatalf("defaults = MinPanelSize %d, Border %d", props.MinPanelSize, props.Border)
	}
	mustPanic(t, func() { Grid(GridProps{Orientation: Orientation(2)}) })
	mustPanic(t, func() { Grid(GridProps{MinPanelSize: -1}) })
	mustPanic(t, func() { Grid(GridProps{MinPanelSize: 2}) })
	mustPanic(t, func() { Grid(GridProps{Border: BorderStyle(99)}) })
}
