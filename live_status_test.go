package nui_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

func TestChartSelectionIsAnnouncedOnceByKeyboardOnly(t *testing.T) {
	var s nui.ChartState
	if err := s.SetCategories([]string{"Jan", "Feb", "Mar"}, []nui.ChartSeries{{ID: "a", Label: "Alpha", Values: []float64{2, 8, 4}}}); err != nil {
		t.Fatal(err)
	}
	if err := s.SetSlices([]nui.ChartSlice{{ID: "one", Label: "One", Value: 3}, {ID: "two", Label: "Two", Value: 1}}); err != nil {
		t.Fatal(err)
	}
	kind := nui.ChartLine
	r := newReaderTester(t, func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Chart(c, th, &s, nui.ChartOptions{Label: "Metrics", Kind: kind, Height: 220, Animate: nui.Set(false)})
	}, 640, 400)

	r.Key(0, ui.KeyTab)
	if !r.Focused("Metrics") {
		t.Fatal("Tab did not focus the chart")
	}
	r.Key(0, ui.KeyHome)
	r.expectRead("Home", "Jan · Alpha: 2")
	r.Key(0, ui.KeyRight)
	r.expectRead("Right", "Feb · Alpha: 8")
	r.Frame()
	r.Frame()
	r.expectSilent("rebuilds after a keyboard move")
	if !r.HasText("Feb · Alpha: 8") {
		t.Fatalf("description line lost: %q", r.Texts())
	}

	plot, ok := r.Find("Metrics")
	if !ok {
		t.Fatal("no plot area")
	}
	seen := map[string]bool{}
	for _, at := range []float32{.05, .5, .95, .5} {
		r.Move(plot.X+plot.W*at, plot.Y+plot.H/2)
		for _, want := range []string{"Jan · Alpha: 2", "Feb · Alpha: 8", "Mar · Alpha: 4"} {
			if r.HasText(want) {
				seen[want] = true
			}
		}
	}
	r.expectSilent("pointer moves")
	if len(seen) != 3 {
		t.Fatalf("the description did not follow the pointer: %v", seen)
	}
	r.Move(0, 0)
	r.expectSilent("pointer leaving")

	kind = nui.ChartPie
	r.Frame()
	r.expectSilent("switching to a pie chart")
	if !r.Focused("Metrics") {
		r.Key(0, ui.KeyTab)
	}
	r.Key(0, ui.KeyHome)
	r.expectRead("pie Home", "One: 3 (75.0%)")
	r.Key(0, ui.KeyRight)
	r.expectRead("pie Right", "Two: 1 (25.0%)")
	r.Frame()
	r.expectSilent("pie rebuild")
}

func TestCommandPaletteMessagesAreReadOnceWhenTheyAppear(t *testing.T) {
	state := nui.CommandPaletteState{}
	var ticket nui.QueryTicket
	r := newReaderTester(t, func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.CommandPalette(c.Key("palette"), th, &state, nui.CommandPaletteOptions{Search: func(_ context.Context, tk nui.QueryTicket, _, _ string) { ticket = tk }})
	}, 700, 620)
	r.expectSilent("closed palette")

	state.Open = true
	r.Frame()
	r.expectRead("opening starts a search", "Searching…")
	r.Frame()
	r.Frame()
	r.expectSilent("rebuilds while searching")
	if info := r.role(ui.RoleStatus, "Searching…"); info.Name != "Searching…" {
		t.Fatalf("searching status %+v", info)
	}
	// The results list shows its own loading status meanwhile; a language
	// change rewrites both texts, so each is read once.
	r.SetLocale("zh-Hans")
	r.expectRead("the searching text changes while shown", "搜索中…", "加载中…")
	r.SetLocale("en")
	r.expectRead("the searching text changes back", "Searching…", "Loading…")

	if !state.Results.Complete(ticket, nil, errors.New("Network unavailable")) {
		t.Fatal("the search ticket was rejected")
	}
	r.Frame()
	r.expectRead("an error replaces the searching message", "Network unavailable")
	for range 3 {
		r.Frame()
	}
	r.expectSilent("rebuilds with the same error")

	r.Type("x")
	r.Advance(time.Second)
	r.expectRead("a new query clears the error and searches", "Searching…")
	if !state.Results.Complete(ticket, nil, nil) {
		t.Fatal("the second search ticket was rejected")
	}
	r.Frame()
	r.expectSilent("results clearing the searching message")
	if _, ok := r.AccessByRole(ui.RoleStatus, "Searching…"); ok {
		t.Fatalf("the searching status is left after the search: %v", accessSummary(r.AccessAll()))
	}

	state.Close()
	r.Frame()
	r.Advance(time.Second)
	r.Announcements()
	state.Open = true
	r.Frame()
	r.expectRead("reopening searches again", "Searching…")
}

func TestFileUploadMessagesAreReadOnceWhenTheyAppear(t *testing.T) {
	release := make(chan error)
	var state nui.FileUploadState
	defer state.Dispose()
	sibling := false
	r := newReaderTester(t, func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Children(func() {
			if sibling {
				ui.Text(c, "Attach the signed contract")
			}
			nui.FileUpload(c.Key("upload"), th, &state, nui.FileUploadOptions{Label: "Files", Choose: func(ctx context.Context) ([]nui.UploadFile, error) {
				select {
				case err := <-release:
					return nil, err
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			}})
		})
	}, 500, 400)

	click(t, r.Tester, "Files")
	if !state.Pending() {
		t.Fatal("the chooser did not start")
	}
	r.expectRead("reading starts", "Reading file information…")
	r.Frame()
	r.Frame()
	r.expectSilent("rebuilds while reading")
	sibling = true
	r.Frame()
	r.expectSilent("a sibling inserted before the upload while reading")

	release <- errors.New("File too large")
	awaitAdvancing(t, r.Tester, func() bool { return state.Error == "File too large" })
	r.Frame()
	r.expectRead("an error replaces the reading message", "File too large")
	for range 3 {
		r.Frame()
	}
	r.expectSilent("rebuilds with the same error")
	sibling = false
	r.Frame()
	r.expectSilent("a sibling removed before the upload")
	state.Error = "Upload quota reached"
	r.Frame()
	r.expectRead("the error text changes while shown", "Upload quota reached")

	click(t, r.Tester, "Files")
	r.expectRead("reading again clears the error", "Reading file information…")
	release <- nil
	awaitAdvancing(t, r.Tester, func() bool { return !state.Pending() })
	r.Frame()
	r.expectSilent("reading finished without files")
}

func TestEmptyStatusesKeepTheirIdentity(t *testing.T) {
	t.Run("SelectMenu message changes while shown", func(t *testing.T) {
		var values []string
		s := nui.SelectMenuState{Open: true}
		r := newReaderTester(t, func(c *ui.Context) {
			th := nui.App(c, nui.AppConfig{})
			nui.SelectMenu(c.Key("menu"), th, &values, &s, nil, nui.SelectMenuOptions{Label: "Pick"})
		}, 500, 400)
		r.role(ui.RoleStatus, "No data")
		r.Frame()
		r.expectSilent("unchanged empty menu")
		r.Type("z")
		r.Frame()
		if s.Query != "z" {
			t.Fatalf("typing did not reach the menu's search: %q", s.Query)
		}
		r.expectRead("query turns the empty message into no match", "No matching data")
		r.Frame()
		r.Frame()
		r.expectSilent("rebuilds with the same query")
	})

	t.Run("Listbox rebuild is silent", func(t *testing.T) {
		var selected []string
		s := nui.ListboxState{Query: "zzz"}
		items := []nui.ListboxItem{{ID: "a", Label: "Alpha"}, {ID: "b", Label: "Beta"}}
		r := newReaderTester(t, func(c *ui.Context) {
			th := nui.App(c, nui.AppConfig{})
			nui.Listbox(c.Key("list"), th, &selected, &s, items, nui.ListboxOptions{Label: "Options", Filter: true})
		}, 500, 400)
		r.role(ui.RoleStatus, "No matching data")
		for range 4 {
			r.Frame()
		}
		r.expectSilent("unchanged no-match rebuilds")
	})

	t.Run("Tree empty on first render", func(t *testing.T) {
		var s nui.TreeState
		r := newReaderTester(t, func(c *ui.Context) {
			th := nui.App(c, nui.AppConfig{})
			nui.Tree(c.Key("tree"), th, &s, nui.TreeOptions{Label: "Files"})
		}, 500, 400)
		r.expectSilent("first render")
		if got := r.withRole(ui.RoleStatus); len(got) != 1 || got[0].Name != "No data" {
			t.Fatalf("empty tree statuses %v", accessSummary(got))
		}
		r.Frame()
		r.expectSilent("rebuild")
	})
}

func TestFeedbackStatusesReadChangesOnce(t *testing.T) {
	for _, kind := range []string{"Alert", "Banner"} {
		t.Run(kind, func(t *testing.T) {
			shown, title, description := false, "Sync", "Saving drafts"
			r := newReaderTester(t, func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				ui.Column(c).Fill().Children(func() {
					ui.Text(c, "Page")
					if !shown {
						return
					}
					if kind == "Alert" {
						nui.Alert(c.Key("notice"), th, nui.AlertOptions{Title: title, Description: description})
					} else {
						nui.Banner(c.Key("notice"), th, nil, nui.BannerOptions{Title: title, Description: description})
					}
				})
			}, 900, 300)
			shown = true
			r.Frame()
			r.expectSilent("first show")
			if info := r.role(ui.RoleStatus, "Sync"); info.Name != "Sync" {
				t.Fatalf("status %+v", info)
			}
			r.Frame()
			r.expectSilent("unchanged rebuild")
			description = "All drafts saved"
			r.Frame()
			got := r.Announcements()
			if len(got) != 1 || !strings.Contains(got[0], "All drafts saved") || !strings.Contains(got[0], "Sync") {
				t.Fatalf("description change read %q", got)
			}
			r.Frame()
			r.expectSilent("rebuild after the change")
			title = "Sync paused"
			r.Frame()
			got = r.Announcements()
			if len(got) != 1 || !strings.Contains(got[0], "Sync paused") {
				t.Fatalf("title change read %q", got)
			}
			r.role(ui.RoleStatus, "Sync paused")
			r.Frame()
			r.expectSilent("rebuild after the title change")
		})
	}
}

func TestDropdownMenuFilterEmptyStatusIsQuiet(t *testing.T) {
	var s nui.MenuState
	r := newReaderTester(t, func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		a := nui.Button(c.Key("actions"), th, "Actions", nui.ButtonOptions{})
		nui.DropdownMenu(c.Key("menu"), th, a, &s, []nui.MenuItem{{ID: "copy", Label: "Copy"}, {ID: "paste", Label: "Paste"}}, nui.MenuOptions{Filter: true})
	}, 500, 400)
	click(t, r.Tester, "Actions")
	r.Announcements()
	r.Type("zz")
	r.Frame()
	r.role(ui.RoleStatus, "No matching data")
	r.expectSilent("the empty message appearing")
	r.Type("z")
	r.Frame()
	r.Frame()
	r.expectSilent("typing that keeps the empty message")
}
