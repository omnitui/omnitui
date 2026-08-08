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
	if data.Border != uint8(BorderHeavy) || !data.Clip || data.Align != uint8(AlignStretch) {
		t.Fatalf("panel border=%d clip=%v align=%d", data.Border, data.Clip, data.Align)
	}
	if core.KeyOf(host.Children[0]) != "first" {
		t.Fatalf("panel key = %q, want first", core.KeyOf(host.Children[0]))
	}
}

func TestGridForwardsPerItemSizes(t *testing.T) {
	child := GridItem(GridItemProps{InitialSize: 12, MinSize: 6, MaxSize: 20},
		Text(TextProps{Content: "first"}).WithKey("first"),
	)
	element := gridHost(GridProps{MinPanelSize: 3, Border: BorderSingle}, child)
	host, ok := core.HostOf(element)
	if !ok {
		t.Fatal("Grid did not create a host")
	}
	data := host.Data.(core.GridData)
	if len(data.Tracks) != 1 || data.Tracks[0] != (core.GridTrackData{InitialSize: 12, MinSize: 6, MaxSize: 20}) {
		t.Fatalf("grid tracks = %+v", data.Tracks)
	}
	if core.KeyOf(host.Children[0]) != "first" {
		t.Fatalf("panel key = %q, want first", core.KeyOf(host.Children[0]))
	}
}

func TestGridForwardsFlexGrow(t *testing.T) {
	element := gridHost(GridProps{FlexGrow: 1, MinPanelSize: 3, Border: BorderSingle}, omnitui.None())
	host, ok := core.HostOf(element)
	if !ok {
		t.Fatal("Grid did not create a host")
	}
	data := host.Data.(core.GridData)
	if data.FlexGrow != 1 {
		t.Fatalf("grid flex grow = %d, want 1", data.FlexGrow)
	}
	mustPanic(t, func() { Grid(GridProps{FlexGrow: -1}) })
}

func TestGridItemValidation(t *testing.T) {
	invalid := []GridItemProps{
		{InitialSize: -1},
		{InitialSize: 2},
		{MinSize: 2},
		{MaxSize: 2},
		{MinSize: 8, MaxSize: 7},
		{InitialSize: 7, MinSize: 8},
		{InitialSize: 9, MaxSize: 8},
	}
	for _, props := range invalid {
		props := props
		mustPanic(t, func() { GridItem(props, omnitui.None()) })
	}
	mustPanic(t, func() {
		Grid(GridProps{MinPanelSize: 8}, GridItem(GridItemProps{MaxSize: 7}, omnitui.None()))
	})
	mustPanic(t, func() {
		Grid(GridProps{MinPanelSize: 8}, GridItem(GridItemProps{InitialSize: 7}, omnitui.None()))
	})
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
