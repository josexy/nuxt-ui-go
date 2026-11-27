package nui_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/egoist/mygo/i18n"
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

// readerTester is a ui.Tester with a screen reader attached from its first
// frame, for tests of what assistive technology reads and sees.
type readerTester struct {
	*ui.Tester
	t *testing.T
}

// newReaderTester starts view with a screen reader attached and forgets
// what the first frame showed, as a real screen reader attaching to an open
// window reads nothing of it.
func newReaderTester(t *testing.T, view func(*ui.Context), width, height int) *readerTester {
	t.Helper()
	tt := ui.NewTester(view, width, height)
	tt.UseScreenReader()
	tt.Frame()
	r := &readerTester{Tester: tt, t: t}
	r.Announcements()
	r.Alerts()
	return r
}

// expectRead checks that the polite announcements since the last check are
// exactly want, and that nothing was read assertively.
func (r *readerTester) expectRead(step string, want ...string) {
	r.t.Helper()
	got, alerts := r.Announcements(), r.Alerts()
	if !slices.Equal(got, want) || len(alerts) != 0 {
		r.t.Fatalf("%s: read %q and alerted %q, want %q", step, got, alerts, want)
	}
}

// expectSilent checks that nothing was read since the last check.
func (r *readerTester) expectSilent(step string) {
	r.t.Helper()
	r.expectRead(step)
}

// role returns the first element with role and accessible name, failing the
// test when there is none.
func (r *readerTester) role(role ui.Role, name string) ui.AccessInfo {
	r.t.Helper()
	info, ok := r.AccessByRole(role, name)
	if !ok {
		r.t.Fatalf("no element with role %d named %q among %+v", role, name, accessSummary(r.AccessAll()))
	}
	return info
}

// withRole returns every element of the last frame with role, in reading
// order.
func (r *readerTester) withRole(role ui.Role) []ui.AccessInfo {
	var out []ui.AccessInfo
	for _, info := range r.AccessAll() {
		if info.Role == role {
			out = append(out, info)
		}
	}
	return out
}

// accessSummary is a short role/name list for failure messages.
func accessSummary(all []ui.AccessInfo) []string {
	out := make([]string, 0, len(all))
	for _, info := range all {
		out = append(out, roleName(info.Role)+":"+info.Name)
	}
	return out
}

func roleName(role ui.Role) string {
	switch role {
	case ui.RoleStatus:
		return "status"
	case ui.RoleAlert:
		return "alert"
	case ui.RoleMain:
		return "main"
	case ui.RoleBanner:
		return "banner"
	case ui.RoleContentInfo:
		return "contentinfo"
	case ui.RoleTabPanel:
		return "tabpanel"
	case ui.RoleTab:
		return "tab"
	case ui.RoleGrid:
		return "grid"
	case ui.RoleTable:
		return "table"
	case ui.RoleCell:
		return "cell"
	case ui.RoleGroup:
		return "group"
	case ui.RoleText:
		return "text"
	case ui.RoleButton:
		return "button"
	}
	return "role"
}

func TestReaderTesterReadsAKeyedStatusOnce(t *testing.T) {
	text, before := "Draft", false
	r := newReaderTester(t, func(c *ui.Context) {
		ui.Column(c).Children(func() {
			if before {
				ui.Text(c, "Sibling")
			}
			ui.Text(c.Key("status"), text).Role(ui.RoleStatus)
		})
	}, 300, 200)
	if info := r.role(ui.RoleStatus, "Draft"); info.Name != "Draft" {
		t.Fatalf("status %+v", info)
	}
	r.Frame()
	r.expectSilent("unchanged rebuild")
	text = "Saved"
	r.Frame()
	r.expectRead("changed status", "Saved")
	r.Frame()
	r.Frame()
	r.expectSilent("rebuilds after the change")
	before = true
	r.Frame()
	r.expectSilent("sibling inserted before the keyed status")
	if got := r.withRole(ui.RoleStatus); len(got) != 1 {
		t.Fatalf("statuses %v", accessSummary(got))
	}
}

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
	r.expectRead("reading starts", "Reading files…")
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
	r.expectRead("reading again clears the error", "Reading files…")
	release <- nil
	awaitAdvancing(t, r.Tester, func() bool { return !state.Pending() })
	r.Frame()
	r.expectSilent("reading finished without files")
}

func TestFileUploadHiddenMessagesStaySilent(t *testing.T) {
	for _, kind := range []string{"kept-mounted tab", "access-hidden parent"} {
		t.Run(kind, func(t *testing.T) {
			visible := false
			var state nui.FileUploadState
			t.Cleanup(state.Dispose)
			r := newReaderTester(t, func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				build := func(c *ui.Context, th nui.Theme) {
					nui.FileUpload(c.Key("files"), th, &state, nui.FileUploadOptions{Label: "Files"})
				}
				if kind == "kept-mounted tab" {
					value := "home"
					if visible {
						value = "upload"
					}
					nui.Tabs(c.Key("tabs"), th, &value, []nui.TabItem{
						{ID: "home", Label: "Home", Content: func(c *ui.Context, _ nui.Theme) { ui.Text(c, "Home content") }},
						{ID: "upload", Label: "Upload", Content: build},
					}, nui.TabsOptions{KeepMounted: true})
				} else {
					ui.Column(c.Key("parent")).AccessHidden(!visible).Children(func() { build(c, th) })
				}
			}, 700, 500)

			state.Error = "Upload failed"
			for range 3 {
				r.Frame()
			}
			if _, ok := r.AccessByRole(ui.RoleStatus, state.Error); ok {
				t.Fatal("the hidden error is accessible")
			}
			r.expectSilent("an error appearing in a hidden parent")
			visible = true
			r.Frame()
			r.role(ui.RoleStatus, state.Error)
			r.expectSilent("exposing an existing status")
			state.Error = "Upload quota reached"
			r.Frame()
			r.expectRead("a visible error changes", state.Error)
			state.Error = ""
			r.Frame()
			r.expectSilent("clearing the error")
			// The parent hides in the same frame as the error appears;
			// last-frame geometric visibility cannot guard this case.
			visible, state.Error = false, "Retry failed"
			r.Frame()
			r.expectSilent("an error appearing as its parent hides")
			state.Error = ""
			r.Frame()
			visible = true
			r.Frame()
			state.Error = "Network unavailable"
			r.Frame()
			r.expectRead("a new error after showing again", state.Error)
			r.Frame()
			r.expectSilent("unchanged visible error")
		})
	}
}

func TestEmptyStatusesKeepTheirIdentity(t *testing.T) {
	t.Run("SelectMenu message changes while shown", func(t *testing.T) {
		var values []string
		s := nui.SelectMenuState{Open: true}
		listSelection1 := chosenSelection(values)
		s.SetItems(nil)
		r := newReaderTester(t, func(c *ui.Context) {
			th := nui.App(c, nui.AppConfig{})
			nui.SelectMenu(c.Key("menu"), th, listSelection1, &s, nui.SelectMenuOptions{Label: "Pick"})
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
		listSelection2 := chosenSelection(selected)
		s.SetItems(items)
		r := newReaderTester(t, func(c *ui.Context) {
			th := nui.App(c, nui.AppConfig{})
			nui.Listbox(c.Key("list"), th, listSelection2, &s, nui.ListboxOptions{Label: "Options", Filter: true})
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
	s.SetItems([]nui.MenuItem{{ID: "copy", Label: "Copy"}, {ID: "paste", Label: "Paste"}})
	r := newReaderTester(t, func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		a := nui.Button(c.Key("actions"), th, "Actions", nui.ButtonOptions{})
		nui.DropdownMenu(c.Key("menu"), th, a, &s, nui.MenuOptions{Filter: true})
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

func inside(inner, outer ui.Rect) bool {
	return inner.W > 0 && inner.H > 0 && inner.X >= outer.X && inner.Y >= outer.Y && inner.X+inner.W <= outer.X+outer.W && inner.Y+inner.H <= outer.Y+outer.H
}

func TestShellComponentsAreLandmarks(t *testing.T) {
	optOut := false
	var header, main, footer ui.Handle
	var bounds [3]ui.Rect
	r := newReaderTester(t, func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Fill().Children(func() {
			h := nui.HeaderLayout(c.Key("header"), th, nui.HeaderOptions{Title: "Project", Description: "Overview"}).Bind(&header)
			m := nui.Main(c.Key("main"), th, func() { ui.Text(c, "Quarterly report") }).Bind(&main)
			f := nui.FooterLayout(c.Key("footer"), th, nui.FooterOptions{Text: "All changes saved"}).Bind(&footer)
			if optOut {
				h.Role(ui.RoleGroup)
				m.Role(ui.RoleGroup)
				f.Role(ui.RoleGroup)
			}
			bounds = [3]ui.Rect{header.Bounds(c), main.Bounds(c), footer.Bounds(c)}
		})
	}, 800, 600)
	r.Frame()
	before := bounds

	for _, role := range []ui.Role{ui.RoleBanner, ui.RoleMain, ui.RoleContentInfo} {
		if got := r.withRole(role); len(got) != 1 {
			t.Fatalf("%s landmarks: %v", roleName(role), accessSummary(r.AccessAll()))
		}
	}
	banner := r.role(ui.RoleBanner, "")
	heading := r.role(ui.RoleHeading, "Project")
	if heading.Level != 1 || !inside(heading.Bounds, banner.Bounds) {
		t.Fatalf("title %+v is not a level-1 heading inside the banner %+v", heading, banner.Bounds)
	}
	mainInfo := r.role(ui.RoleMain, "")
	body := r.role(ui.RoleText, "Quarterly report")
	if !inside(body.Bounds, mainInfo.Bounds) {
		t.Fatalf("page content %v is outside the main landmark %v", body.Bounds, mainInfo.Bounds)
	}
	info := r.role(ui.RoleContentInfo, "")
	saved := r.role(ui.RoleText, "All changes saved")
	if !inside(saved.Bounds, info.Bounds) {
		t.Fatalf("footer text %v is outside the content-info landmark %v", saved.Bounds, info.Bounds)
	}

	optOut = true
	r.Frame()
	r.Frame()
	for _, role := range []ui.Role{ui.RoleBanner, ui.RoleMain, ui.RoleContentInfo} {
		if got := r.withRole(role); len(got) != 0 {
			t.Fatalf("opted-out %s landmark is still reported: %v", roleName(role), accessSummary(got))
		}
	}
	if bounds != before {
		t.Fatalf("opting out moved the shell: before %v after %v", before, bounds)
	}
	if heading := r.role(ui.RoleHeading, "Project"); heading.Level != 1 {
		t.Fatalf("opted-out header lost its heading: %+v", heading)
	}
	r.expectSilent("landmark changes")
}

func TestTabPanelsAreNamedAndControlledByTheirTabs(t *testing.T) {
	value, keep, content := "account", false, true
	withContent := true
	r := newReaderTester(t, func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		items := []nui.TabItem{{ID: "account", Label: "Account"}, {ID: "password", Label: "Password"}}
		if withContent {
			items[0].Content = func(c *ui.Context, t nui.Theme) { ui.Text(c, "Account settings") }
			items[1].Content = func(c *ui.Context, t nui.Theme) { ui.Text(c, "Password settings") }
		}
		nui.Tabs(c.Key("tabs"), th, &value, items, nui.TabsOptions{Label: "Settings", KeepMounted: keep, Content: nui.Set(content)})
	}, 600, 400)

	panel := r.role(ui.RoleTabPanel, "Account")
	if !inside(r.role(ui.RoleText, "Account settings").Bounds, panel.Bounds) {
		t.Fatalf("panel %+v does not hold its content", panel)
	}
	if tab := r.role(ui.RoleTab, "Account"); !slices.Equal(tab.Controls, []string{"Account"}) || !tab.Selected {
		t.Fatalf("Account tab %+v does not control its panel", tab)
	}
	if tab := r.role(ui.RoleTab, "Password"); len(tab.Controls) != 0 {
		t.Fatalf("Password tab controls %q while its panel is not built", tab.Controls)
	}
	if got := r.withRole(ui.RoleTabPanel); len(got) != 1 {
		t.Fatalf("tab panels %v", accessSummary(got))
	}

	keep = true
	r.Frame()
	if got := r.withRole(ui.RoleTabPanel); len(got) != 1 || got[0].Name != "Account" {
		t.Fatalf("a hidden kept-mounted panel is reported: %v", accessSummary(got))
	}
	if _, ok := r.AccessByRole(ui.RoleText, "Password settings"); ok {
		t.Fatal("the hidden panel's content is reported")
	}
	r.Key(0, ui.KeyTab)
	r.Key(0, ui.KeyRight)
	if value != "password" {
		t.Fatalf("arrow key selected %q", value)
	}
	if got := r.withRole(ui.RoleTabPanel); len(got) != 1 || got[0].Name != "Password" {
		t.Fatalf("panels after switching: %v", accessSummary(got))
	}
	if tab := r.role(ui.RoleTab, "Password"); !slices.Equal(tab.Controls, []string{"Password"}) {
		t.Fatalf("Password tab %+v does not control its panel", tab)
	}

	for _, step := range []struct {
		name               string
		withContent, shown bool
	}{{"Content=false", true, false}, {"items without content", false, true}} {
		withContent, content = step.withContent, step.shown
		r.Frame()
		if got := r.withRole(ui.RoleTabPanel); len(got) != 0 {
			t.Fatalf("%s: panels %v", step.name, accessSummary(got))
		}
		for _, name := range []string{"Account", "Password"} {
			if tab := r.role(ui.RoleTab, name); len(tab.Controls) != 0 {
				t.Fatalf("%s: %s tab controls %q", step.name, name, tab.Controls)
			}
		}
	}
	r.expectSilent("tab panel changes")
}

func TestTabPanelKeepsAnOverriddenLegacyName(t *testing.T) {
	if os.Getenv("NUI_TAB_PANEL_OVERRIDE_CHILD") != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestTabPanelKeepsAnOverriddenLegacyName$", "-test.count=1")
		cmd.Env = append(os.Environ(), "NUI_TAB_PANEL_OVERRIDE_CHILD=1", "CGO_ENABLED=0")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("child: %v\n%s", err, out)
		}
		return
	}
	app, err := i18n.NewCatalog(fstest.MapFS{"en.json": {Data: []byte(`{"@@locale": "en", "nuxtui.contentNamed": "{name} pane"}`)}}, "en")
	if err != nil {
		t.Fatal(err)
	}
	i18n.Register(app)
	value := "account"
	r := newReaderTester(t, func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Tabs(c, th, &value, []nui.TabItem{{ID: "account", Label: "Account", Content: func(c *ui.Context, t nui.Theme) { ui.Text(c, "Account settings") }}}, nui.TabsOptions{})
	}, 600, 400)
	r.SetLocale("en")
	r.role(ui.RoleTabPanel, "Account pane")
	if tab := r.role(ui.RoleTab, "Account"); !slices.Equal(tab.Controls, []string{"Account pane"}) {
		t.Fatalf("tab %+v does not control the renamed panel", tab)
	}
}

// focusLabel tabs through the view until the element labeled name has the
// keyboard focus.
func focusLabel(t *testing.T, tt *ui.Tester, name string) {
	t.Helper()
	for range 20 {
		if tt.Focused(name) {
			return
		}
		tt.Key(0, ui.KeyTab)
	}
	t.Fatalf("could not focus %q", name)
}

func TestCalendarGridIsAnInteractiveGrid(t *testing.T) {
	day := nui.Date{Year: 2026, Month: time.November, Day: 13}
	for _, readOnly := range []bool{false, true} {
		name := "editable"
		if readOnly {
			name = "read-only"
		}
		t.Run(name, func(t *testing.T) {
			selection := nui.DateSelection{Start: nui.Set(day)}
			var state nui.CalendarState
			changes := 0
			r := newReaderTester(t, func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				nui.Calendar(c.Key("calendar"), th, &selection, &state, nui.CalendarOptions{Label: "Due", ReadOnly: readOnly, OnChange: func(nui.DateSelection) { changes++ }})
			}, 640, 560)

			grid := r.role(ui.RoleGrid, "Due")
			if grid.ReadOnly != readOnly {
				t.Fatalf("grid read-only %v, want %v", grid.ReadOnly, readOnly)
			}
			if _, ok := r.AccessByRole(ui.RoleTable, ""); ok {
				t.Fatal("the day grid is still reported as a table")
			}
			for _, weekday := range []string{"Sunday", "Monday", "Saturday"} {
				r.role(ui.RoleColumnHeader, weekday)
			}
			cell := r.role(ui.RoleCell, "Due 2026-11-13")
			if !cell.Selected || cell.ReadOnly != readOnly || !inside(cell.Bounds, grid.Bounds) {
				t.Fatalf("13 November cell %+v in grid %v", cell, grid.Bounds)
			}
			if other := r.role(ui.RoleCell, "Due 2026-11-14"); other.Selected || other.ReadOnly != readOnly {
				t.Fatalf("14 November cell %+v", other)
			}

			focusLabel(t, r.Tester, "Due")
			if active := r.role(ui.RoleCell, "Due 2026-11-13"); !active.Focused {
				t.Fatalf("the grid does not point assistive technology at the active day: %+v", active)
			}
			r.Key(0, ui.KeyRight)
			if want := (nui.Date{Year: 2026, Month: time.November, Day: 14}); state.Active != want || !r.role(ui.RoleCell, "Due 2026-11-14").Focused {
				t.Fatalf("Right moved the focused date to %v, want %v", state.Active, want)
			}
			r.Key(0, ui.KeyEnter)
			if readOnly {
				if selection.Start.Value != day || changes != 0 {
					t.Fatalf("a read-only calendar changed its value to %+v (%d changes)", selection, changes)
				}
			} else if selection.Start.Value != state.Active || changes != 1 {
				t.Fatalf("Enter chose %+v (%d changes)", selection, changes)
			}
		})
	}
}

func TestCalendarMonthAndYearGridsAreGrids(t *testing.T) {
	selection := nui.DateSelection{Start: nui.Set(nui.Date{Year: 2026, Month: time.November, Day: 13})}
	state := nui.CalendarState{View: nui.CalendarMonthView}
	r := newReaderTester(t, func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Calendar(c.Key("calendar"), th, &selection, &state, nui.CalendarOptions{Label: "Due", ReadOnly: true})
	}, 640, 560)
	if grid := r.role(ui.RoleGrid, "Due"); !grid.ReadOnly {
		t.Fatalf("month grid %+v", grid)
	}
	if cell := r.role(ui.RoleCell, "Due Nov 2026"); !cell.ReadOnly {
		t.Fatalf("month cell %+v", cell)
	}
	state.View = nui.CalendarYearView
	r.Frame()
	if grid := r.role(ui.RoleGrid, "Due"); !grid.ReadOnly {
		t.Fatalf("year grid %+v", grid)
	}
	r.role(ui.RoleCell, "Due 2026")
}

func TestReadOnlyControlsReportReadOnly(t *testing.T) {
	rating := 3
	date := nui.Set(nui.Date{Year: 2026, Month: time.November, Day: 13})
	clock := nui.Set(nui.ClockTime{Hour: 9, Minute: 30})
	var dateState nui.DateInputState
	var timeState nui.TimeInputState
	r := newReaderTester(t, func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Gap(12).Children(func() {
			nui.InputRating(c.Key("rating"), th, &rating, nui.InputRatingOptions{Label: "Score", ReadOnly: true})
			nui.InputDate(c.Key("date"), th, &date, &dateState, nui.InputDateOptions{Label: "Due", ReadOnly: true})
			nui.InputTime(c.Key("time"), th, &clock, &timeState, nui.InputTimeOptions{Label: "Start"})
		})
	}, 640, 400)

	score := r.role(ui.RoleSlider, "Score")
	if !score.ReadOnly || score.Now != 3 || score.Value != "3" {
		t.Fatalf("read-only rating %+v", score)
	}
	if month := r.role(ui.RoleStepper, "Due month"); !month.ReadOnly {
		t.Fatalf("read-only InputDate month segment %+v", month)
	}
	if hour := r.role(ui.RoleStepper, "Start hour"); hour.ReadOnly {
		t.Fatalf("editable InputTime hour segment reports read-only: %+v", hour)
	}
}
