package nui_test

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func TestSelectMenuResetAndOpeningKeys(t *testing.T) {
	for _, keep := range []bool{false, true} {
		var state nui.SelectMenuState
		var selected []string
		var queries []string
		var openings []bool
		disabled, search := false, true
		tt := ui.NewTester(func(c *ui.Context) {
			theme := nui.App(c, nui.AppConfig{})
			ui.Column(c).Padding(20).Children(func() {
				nui.SelectMenu(c, theme, &selected, &state, []nui.ListboxItem{{ID: "a", Label: "Alpha"}, {ID: "b", Label: "Beta"}}, nui.SelectMenuOptions{
					Label: "Choose", Disabled: disabled, SearchInput: nui.Set(search),
					ResetSearchTermOnBlur: nui.Set(!keep), ResetSearchTermOnSelect: nui.Set(!keep), OnQuery: func(q string) { queries = append(queries, q) },
					OnOpenChange: func(open bool) { openings = append(openings, open) },
				}).Width(300)
				nui.Button(c, theme, "Outside", nui.ButtonOptions{}).Absolute().Top(440).Left(350)
			})
		}, 500, 500)
		click(t, tt, "Choose")
		if len(openings) != 1 || !openings[0] {
			t.Fatalf("opening notifications: %v", openings)
		}
		tt.Type("Alpha")
		click(t, tt, "Alpha")
		tt.Advance(500 * time.Millisecond)
		if state.Open || len(selected) != 1 || selected[0] != "a" {
			t.Fatal("selection did not close")
		}
		if len(openings) != 2 || openings[1] {
			t.Fatalf("selection close notifications: %v", openings)
		}
		if (state.Query != "") != keep {
			t.Fatalf("keep=%v query=%q", keep, state.Query)
		}
		click(t, tt, "Choose")
		if tt.HasText("Beta") == keep {
			t.Fatal("reopen filtering does not reflect reset option")
		}
		tt.Key(ui.Ctrl, ui.KeyA)
		tt.Type("Beta")
		beforeClose := len(openings)
		tt.Key(0, ui.KeyEscape)
		tt.Advance(500 * time.Millisecond)
		if len(openings) != beforeClose+1 || openings[len(openings)-1] {
			t.Fatalf("Escape notifications: %v", openings)
		}
		if !tt.Focused("Choose") {
			t.Fatal("focus did not return")
		}
		if (state.Query != "") != keep {
			t.Fatal("Escape did not apply reset option")
		}
		for _, key := range []ui.Key{ui.KeyUp, ui.KeyDown} {
			tt.Key(0, key)
			if !state.Open {
				t.Fatalf("%v did not open", key)
			}
			tt.Key(0, ui.KeyEscape)
			tt.Advance(500 * time.Millisecond)
		}
		search = false
		tt.Frame()
		tt.Key(0, ui.KeyUp)
		if tt.HasText("Search Choose") || !tt.HasText("Beta") {
			t.Fatal("hidden search still filters/appears")
		}
		tt.Key(0, ui.KeyEnd)
		tt.Key(0, ui.KeyEnter)
		if len(selected) != 1 || selected[0] != "b" {
			t.Fatalf("searchless keyboard selection: %v", selected)
		}
		tt.Advance(500 * time.Millisecond)
		click(t, tt, "Choose")
		search = true
		tt.Frame()
		click(t, tt, "Search Choose")
		tt.Type("Alpha")
		click(t, tt, "Outside")
		tt.Advance(500 * time.Millisecond)
		if state.Open || ((state.Query != "") != keep) || !tt.Focused("Outside") {
			t.Fatalf("outside dismissal: open=%v query=%q keep=%v focused=%v", state.Open, state.Query, keep, tt.Focused("Outside"))
		}
		disabled = true
		tt.Frame()
		tt.Key(0, ui.KeyUp)
		if state.Open {
			t.Fatal("disabled trigger opened")
		}
		before := len(queries)
		tt.Frame()
		tt.Frame()
		if len(queries) != before {
			t.Fatal("reset callback repeated on rebuild")
		}
	}
}

func TestPaginationExplicitZeroAndControls(t *testing.T) {
	page := 5
	edges, controls := true, true
	tt := ui.NewTester(func(c *ui.Context) {
		theme := nui.App(c, nui.AppConfig{})
		nui.Pagination(c, theme, &page, nui.PaginationOptions{Total: 100, PageSize: 10, SiblingCount: nui.Set(0), ShowEdges: nui.Set(edges), ShowControls: nui.Set(controls)})
	}, 1000, 150)
	if tt.HasText("4") || tt.HasText("6") || !tt.HasText("5") || !tt.HasText("10") {
		t.Fatal("zero neighbors / edges")
	}
	edges, controls = false, false
	tt.Frame()
	if tt.HasText("1") || tt.HasText("10") || tt.HasText("Next") || !tt.HasText("5") {
		t.Fatal("hidden controls/edges")
	}
	page = 99
	tt.Frame()
	if page != 10 || !tt.HasText("10") {
		t.Fatal("page clamping")
	}
}

func TestPaginationEdgeWindowsAndActiveVariant(t *testing.T) {
	page := 1
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		nui.Pagination(c, th, &page, nui.PaginationOptions{Total: 100, PageSize: 10, SiblingCount: nui.Set(0), ShowEdges: nui.Set(true), ActiveColor: nui.Error, ActiveVariant: nui.Outline})
	}, 1000, 150)
	if !tt.HasText("3") || tt.HasText("4") {
		t.Fatal("first edge window")
	}
	page = 10
	tt.Frame()
	if !tt.HasText("8") || tt.HasText("7") {
		t.Fatal("last edge window")
	}
	r := bounds(t, tt, "Page 10")
	red, green, blue, _ := tt.Image().At(int(r.X+6), int(r.Y+6)).RGBA()
	if red < 60000 || green < 60000 || blue < 60000 {
		t.Fatalf("active outline overridden by Selected fill: %d,%d,%d", red, green, blue)
	}
}
