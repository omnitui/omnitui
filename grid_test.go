package omnitui

import (
	"testing"

	"github.com/omnitui/omnitui/v2/internal/core"
)

func TestGridArrangesHorizontalPanelsWithSharedBorder(t *testing.T) {
	app := newGridTestApp(t, 0, 11, 5, 3)
	grid := app.rootInstance

	assertRect(t, grid.children[0].rect, Rect{Width: 6, Height: 5})
	assertRect(t, grid.children[1].rect, Rect{X: 5, Width: 6, Height: 5})
	if got := app.front.Cell(5, 0).Grapheme; got != "╥" {
		t.Fatalf("top divider = %q, want ╥", got)
	}
	if got := app.front.Cell(5, 2).Grapheme; got != "║" {
		t.Fatalf("divider = %q, want ║", got)
	}
	if got := app.front.Cell(5, 4).Grapheme; got != "╨" {
		t.Fatalf("bottom divider = %q, want ╨", got)
	}
}

func TestGridArrangesVerticalPanelsWithSharedBorder(t *testing.T) {
	app := newGridTestApp(t, 1, 9, 11, 3)
	grid := app.rootInstance

	assertRect(t, grid.children[0].rect, Rect{Width: 9, Height: 6})
	assertRect(t, grid.children[1].rect, Rect{Y: 5, Width: 9, Height: 6})
	if got := app.front.Cell(0, 5).Grapheme; got != "╞" {
		t.Fatalf("left divider = %q, want ╞", got)
	}
	if got := app.front.Cell(4, 5).Grapheme; got != "═" {
		t.Fatalf("divider = %q, want ═", got)
	}
	if got := app.front.Cell(8, 5).Grapheme; got != "╡" {
		t.Fatalf("right divider = %q, want ╡", got)
	}

	app.dispatchMouse(MouseEvent{Action: MouseDown, Button: MouseButtonLeft, X: 4, Y: 5})
	app.dispatchMouse(MouseEvent{Action: MouseUp, Button: MouseButtonLeft, X: 4, Y: 7})
	if err := app.cycle(); err != nil {
		t.Fatal(err)
	}
	if grid.gridSizes[0] != 8 || grid.gridSizes[1] != 4 {
		t.Fatalf("vertical dragged sizes = %v, want [8 4]", grid.gridSizes)
	}
}

func TestGridDragResizesOnlyInternalBorders(t *testing.T) {
	app := newGridTestApp(t, 0, 11, 5, 3)
	grid := app.rootInstance

	app.dispatchMouse(MouseEvent{Action: MouseDown, Button: MouseButtonLeft, X: 5, Y: 2})
	if app.capture != grid {
		t.Fatal("grid did not capture the mouse from its internal divider")
	}
	app.dispatchMouse(MouseEvent{Action: MouseMove, Buttons: MouseLeftPressed, X: 9, Y: 2})
	if err := app.cycle(); err != nil {
		t.Fatal(err)
	}
	if grid.gridSizes[0] != 9 || grid.gridSizes[1] != 3 {
		t.Fatalf("dragged sizes = %v, want [9 3]", grid.gridSizes)
	}
	assertRect(t, grid.children[0].rect, Rect{Width: 9, Height: 5})
	assertRect(t, grid.children[1].rect, Rect{X: 8, Width: 3, Height: 5})
	app.dispatchMouse(MouseEvent{Action: MouseUp, Button: MouseButtonLeft, X: 9, Y: 2})
	if grid.gridDragging || app.capture != nil {
		t.Fatal("grid kept mouse capture after release")
	}

	for _, x := range []int{0, 2, 10} {
		before := append([]int(nil), grid.gridSizes...)
		app.dispatchMouse(MouseEvent{Action: MouseDown, Button: MouseButtonLeft, X: x, Y: 2})
		app.dispatchMouse(MouseEvent{Action: MouseMove, Buttons: MouseLeftPressed, X: x + 1, Y: 2})
		app.dispatchMouse(MouseEvent{Action: MouseUp, Button: MouseButtonLeft, X: x + 1, Y: 2})
		if grid.gridSizes[0] != before[0] || grid.gridSizes[1] != before[1] {
			t.Fatalf("non-divider click at X=%d changed sizes from %v to %v", x, before, grid.gridSizes)
		}
	}
}

func TestGridDragLeavesNonAdjacentPanelsUnchanged(t *testing.T) {
	app := newGridTestApp(t, 0, 16, 5, 3, 3)
	grid := app.rootInstance
	last := grid.gridSizes[2]
	divider := grid.gridSizes[0] - 1

	app.dispatchMouse(MouseEvent{Action: MouseDown, Button: MouseButtonLeft, X: divider, Y: 2})
	app.dispatchMouse(MouseEvent{Action: MouseUp, Button: MouseButtonLeft, X: divider + 2, Y: 2})
	if err := app.cycle(); err != nil {
		t.Fatal(err)
	}
	if grid.gridSizes[0] != 8 || grid.gridSizes[1] != 4 || grid.gridSizes[2] != last {
		t.Fatalf("three-panel sizes = %v, want [8 4 %d]", grid.gridSizes, last)
	}
}

func TestGridUsesPerPanelInitialSizesOnBothAxes(t *testing.T) {
	tracks := []core.GridTrackData{
		{InitialSize: 10, MinSize: 8, MaxSize: 12},
		{MinSize: 5, MaxSize: 13},
	}
	horizontal := newGridTestAppWithTracks(t, 0, 20, 5, 3, tracks)
	if got := horizontal.rootInstance.gridSizes; got[0] != 10 || got[1] != 11 {
		t.Fatalf("horizontal initial sizes = %v, want [10 11]", got)
	}
	assertRect(t, horizontal.rootInstance.children[0].rect, Rect{Width: 10, Height: 5})
	assertRect(t, horizontal.rootInstance.children[1].rect, Rect{X: 9, Width: 11, Height: 5})

	vertical := newGridTestAppWithTracks(t, 1, 5, 20, 3, tracks)
	if got := vertical.rootInstance.gridSizes; got[0] != 10 || got[1] != 11 {
		t.Fatalf("vertical initial sizes = %v, want [10 11]", got)
	}
	assertRect(t, vertical.rootInstance.children[0].rect, Rect{Width: 5, Height: 10})
	assertRect(t, vertical.rootInstance.children[1].rect, Rect{Y: 9, Width: 5, Height: 11})
}

func TestGridInitialSizeControlsAutomaticMeasurement(t *testing.T) {
	data := core.GridData{
		MinPanelSize: 3,
		Tracks: []core.GridTrackData{
			{InitialSize: 8, MinSize: 5, MaxSize: 10},
		},
	}
	if got := gridMeasuredTrackSize(data, 0, 20); got != 8 {
		t.Fatalf("measured explicit track = %d, want 8", got)
	}
	if got := gridMeasuredTrackSize(data, 1, 20); got != 20 {
		t.Fatalf("measured automatic track = %d, want 20", got)
	}
}

func TestGridDragHonorsPerPanelMinimumsAndMaximums(t *testing.T) {
	tracks := []core.GridTrackData{
		{InitialSize: 10, MinSize: 8, MaxSize: 12},
		{MinSize: 5, MaxSize: 13},
	}
	app := newGridTestAppWithTracks(t, 0, 20, 5, 3, tracks)
	grid := app.rootInstance

	app.dispatchMouse(MouseEvent{Action: MouseDown, Button: MouseButtonLeft, X: 9, Y: 2})
	app.dispatchMouse(MouseEvent{Action: MouseUp, Button: MouseButtonLeft, X: 17, Y: 2})
	if err := app.cycle(); err != nil {
		t.Fatal(err)
	}
	if got := grid.gridSizes; got[0] != 12 || got[1] != 9 {
		t.Fatalf("sizes after maximum drag = %v, want [12 9]", got)
	}

	app.dispatchMouse(MouseEvent{Action: MouseDown, Button: MouseButtonLeft, X: 11, Y: 2})
	app.dispatchMouse(MouseEvent{Action: MouseUp, Button: MouseButtonLeft, X: 0, Y: 2})
	if err := app.cycle(); err != nil {
		t.Fatal(err)
	}
	if got := grid.gridSizes; got[0] != 8 || got[1] != 13 {
		t.Fatalf("sizes after minimum drag = %v, want [8 13]", got)
	}
}

func TestGridFitsExistingSizesWithinPerPanelBounds(t *testing.T) {
	data := core.GridData{
		MinPanelSize: 3,
		Tracks: []core.GridTrackData{
			{InitialSize: 10, MinSize: 8, MaxSize: 12},
			{MinSize: 5, MaxSize: 13},
		},
	}
	if got := fitGridSizes([]int{10, 11}, data, 16); got[0] != 8 || got[1] != 8 {
		t.Fatalf("shrunk sizes = %v, want [8 8]", got)
	}
	if got := fitGridSizes([]int{8, 8}, data, 31); got[0] != 12 || got[1] != 13 {
		t.Fatalf("grown capped sizes = %v, want [12 13]", got)
	}
}

func newGridTestApp(t *testing.T, orientation uint8, width, height, minimum int, panelCount ...int) *App {
	t.Helper()
	panel := func(label string) Element {
		return core.NewHost(core.HostBox, core.BoxData{Border: 1, Clip: true}, []Element{
			core.NewHost(core.HostText, core.TextData{Content: label}, nil),
		})
	}
	count := 2
	if len(panelCount) > 0 {
		count = panelCount[0]
	}
	panels := make([]Element, count)
	for index := range panels {
		panels[index] = panel("panel")
	}
	root := core.NewHost(core.HostGrid, core.GridData{
		Width: core.CellsSize(width), Height: core.CellsSize(height),
		Orientation: orientation, MinPanelSize: minimum, Border: 1,
	}, panels)
	app := New(root, Options{})
	app.width, app.height = width, height
	if err := app.render(); err != nil {
		t.Fatal(err)
	}
	return app
}

func newGridTestAppWithTracks(t *testing.T, orientation uint8, width, height, minimum int, tracks []core.GridTrackData) *App {
	t.Helper()
	panel := func() Element {
		return core.NewHost(core.HostBox, core.BoxData{Border: 1, Clip: true}, []Element{
			core.NewHost(core.HostText, core.TextData{Content: "panel"}, nil),
		})
	}
	panels := make([]Element, len(tracks))
	for index := range panels {
		panels[index] = panel()
	}
	root := core.NewHost(core.HostGrid, core.GridData{
		Width: core.CellsSize(width), Height: core.CellsSize(height),
		Orientation: orientation, MinPanelSize: minimum, Border: 1,
		Tracks: tracks,
	}, panels)
	app := New(root, Options{})
	app.width, app.height = width, height
	if err := app.render(); err != nil {
		t.Fatal(err)
	}
	return app
}

func assertRect(t *testing.T, got, want Rect) {
	t.Helper()
	if got != want {
		t.Fatalf("rect = %#v, want %#v", got, want)
	}
}
