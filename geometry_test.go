package omnitui

import "testing"

func TestFillSize(t *testing.T) {
	if SizeIsAuto(Fill()) {
		t.Fatal("Fill() must not be auto-sized")
	}
	if !SizeIsFill(Fill()) {
		t.Fatal("Fill() must be identified as a fill size")
	}
}
