package components

import (
	"testing"

	omnitui "github.com/viniciusfonseca/omnitui"
	"github.com/viniciusfonseca/omnitui/internal/core"
)

func TestRowAndColumnAreBoxHosts(t *testing.T) {
	child := Text(TextProps{Content: "child"}).WithKey("stable")
	for name, element := range map[string]omnitui.Element{"row": Row(RowProps{Gap: 1}, child), "column": Column(ColumnProps{Gap: 1}, child)} {
		if core.KindOf(element) != core.KindComponent || core.ComponentOf(element).Name != nameTitle(name) {
			t.Fatalf("%s is not a builtin component", name)
		}
	}
}

func nameTitle(value string) string {
	if value == "row" {
		return "Row"
	}
	return "Column"
}

func TestTabsAndListValidateKeys(t *testing.T) {
	mustPanic(t, func() { Tabs(TabsProps{Items: []TabItem{{Key: "same"}, {Key: "same"}}}) })
	mustPanic(t, func() { List(ListProps{Selectable: true}, Text(TextProps{Content: "missing"})) })
	Tabs(TabsProps{ActiveKey: "ok", Items: []TabItem{{Key: "ok"}, {Key: "disabled", Disabled: true}}})
}

func TestStyleConflictIsRejected(t *testing.T) {
	mustPanic(t, func() { Text(TextProps{Style: omnitui.Style{Attributes: omnitui.Bold, ClearAttributes: omnitui.Bold}}) })
}

func mustPanic(t *testing.T, action func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	action()
}
