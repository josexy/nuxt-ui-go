package nui_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

func TestSidesResizeReduceMotionAndFocusTrap(t *testing.T) {
	for _, edge := range []nui.Side{nui.SideRight, nui.SideLeft, nui.SideTop, nui.SideBottom} {
		t.Run(fmt.Sprint(edge), func(t *testing.T) {
			open := false
			tt := ui.NewTester(func(c *ui.Context) {
				theme := nui.App(c, nui.AppConfig{})
				nui.Button(c, theme, "Open", nui.ButtonOptions{}).OnClick(func() { open = true })
				nui.Slideover(c, theme, &open, nui.SlideoverOptions{Side: edge, Extent: nui.Set(float32(300)), Title: "Side", Footer: func(c *ui.Context, t nui.Theme) {
					nui.Button(c, t, "Finish", nui.ButtonOptions{}).OnClick(func() { open = false })
				}}, func(t nui.Theme) { nui.Button(c, t, "Inside", nui.ButtonOptions{}) })
			}, 700, 500)
			tt.SetPreferences(ui.Preferences{ReduceMotion: true})
			click(t, tt, "Open")
			for _, size := range [][2]int{{700, 500}, {360, 320}} {
				tt.SetSize(size[0], size[1])
				r := bounds(t, tt, "Side")
				if r.X < 0 || r.Y < 0 || r.X+r.W > float32(size[0]) || r.Y+r.H > float32(size[1]) {
					t.Fatalf("outside window: %v", r)
				}
				if (edge == nui.SideLeft || edge == nui.SideRight) && r.W != 300 {
					t.Fatalf("side width %v", r)
				}
				if (edge == nui.SideTop || edge == nui.SideBottom) && r.H != 300 {
					t.Fatalf("side height %v", r)
				}
			}
			click(t, tt, "Inside")
			for range 8 {
				tt.Key(0, ui.KeyTab)
				if tt.Focused("Open") {
					t.Fatal("focus escaped modal scope")
				}
			}
			tt.Key(0, ui.KeyEscape)
			if open || tt.HasText("Inside") || !tt.Focused("Open") {
				t.Fatal("reduced motion close or focus")
			}
		})
	}
}

func TestMenuHoverKeyboardAndRemoval(t *testing.T) {
	var state nui.MenuState
	remove := false
	tt := ui.NewTester(func(c *ui.Context) {
		theme := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(20).Gap(12).Children(func() {
			anchor := nui.Button(c, theme, "Actions", nui.ButtonOptions{})
			items := []nui.MenuItem{{ID: "export", Label: "Export", Children: []nui.MenuItem{{ID: "csv", Label: "CSV"}}}, {ID: "rename", Label: "Rename"}}
			if remove {
				items = items[1:]
			}
			nui.DropdownMenu(c, theme, anchor, &state, items, nui.MenuOptions{})
			nui.Button(c, theme, "After menu", nui.ButtonOptions{})
		})
	}, 600, 400)
	click(t, tt, "Actions")
	r := bounds(t, tt, "Export")
	tt.Move(r.X+5, r.Y+5)
	tt.Advance(200 * time.Millisecond)
	tt.Key(0, ui.KeyRight)
	tt.Key(0, ui.KeyEscape)
	tt.Advance(400 * time.Millisecond)
	if tt.HasText("CSV") || !tt.Focused("Export") {
		t.Fatal("hover reopened dismissed submenu")
	}
	tt.Key(0, ui.KeyRight)
	remove = true
	tt.Frame()
	if tt.HasText("CSV") {
		t.Fatal("removed submenu survived")
	}
	tt.Key(0, ui.KeyTab)
	if state.Open || !tt.Focused("After menu") {
		t.Fatal("Tab did not dismiss menu and continue after trigger")
	}
	click(t, tt, "Actions")
	tt.Key(ui.Shift, ui.KeyTab)
	if state.Open || !tt.Focused("Actions") {
		t.Fatal("Shift+Tab did not dismiss menu and return to trigger")
	}
}

func TestToastFocusBackgroundPauseAndLimit(t *testing.T) {
	count := 0
	tt := ui.NewTester(func(c *ui.Context) {
		theme := nui.App(c, nui.AppConfig{})
		nui.Button(c, theme, "Queue", nui.ButtonOptions{}).OnClick(func() {
			count++
			nui.Toast(c, theme, nui.ToastOptions{ID: fmt.Sprint(count), Title: fmt.Sprintf("Notice %d", count), Timeout: time.Second, Action: "Undo"})
		})
		nui.ToastViewport(c, theme, nui.ToastViewportOptions{})
	}, 600, 600)
	// Toast actions use their own per-toast styles; keyboard navigation enters the viewport.
	for range 4 {
		click(t, tt, "Queue")
	}
	if tt.HasText("Notice 1") || !tt.HasText("Notice 4") {
		t.Fatal("native newest-three queue")
	}
	tt.Move(0, 0)
	tt.SetFocused(false)
	tt.Advance(2 * time.Second)
	if !tt.HasText("Notice 4") {
		t.Fatal("background window timeout was not paused")
	}
	tt.SetFocused(true)
	tt.Key(0, ui.KeyTab)
	if !tt.Focused("Undo") {
		t.Fatalf("toast action not keyboard reachable: %v", tt.Texts())
	}
	tt.Advance(2 * time.Second)
	if !tt.HasText("Notice 4") {
		t.Fatal("focused toast timeout was not paused")
	}
	click(t, tt, "Queue")
	tt.Move(0, 0)
	tt.Advance(1100 * time.Millisecond)
	if tt.HasText("Notice 4") || tt.HasText("Notice 5") {
		t.Fatal("queue did not resume expiry")
	}
}

func TestMenusKeyboardIdentityAndSubmenus(t *testing.T) {
	var menu nui.MenuState
	checked := false
	chosen := ""
	calls := 0
	swap := false
	tt := ui.NewTester(func(c *ui.Context) {
		theme := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(12).Gap(12).Children(func() {
			anchor := nui.Button(c, theme, "Actions", nui.ButtonOptions{})
			items := []nui.MenuItem{
				{ID: "rename", Label: "Rename", OnSelect: func() { chosen = "rename"; calls++ }},
				{ID: "disabled", Label: "Unavailable", Disabled: true, OnSelect: func() { t.Error("disabled action") }},
				{ID: "pin", Label: "Pinned", Checked: nui.Set(checked), OnSelect: func() { checked = !checked; calls++ }},
				{ID: "export", Label: "Export", Description: "Choose a format", Children: []nui.MenuItem{{ID: "csv", Label: "CSV", OnSelect: func() { chosen = "csv"; calls++ }}, {ID: "advanced", Label: "Advanced formats", Children: []nui.MenuItem{{ID: "json", Label: "JSON", OnSelect: func() { chosen = "json"; calls++ }}}}}},
				{ID: "delete", Label: "Delete", Color: nui.Error, OnSelect: func() { chosen = "delete"; calls++ }},
			}
			if swap {
				items[0], items[4] = items[4], items[0]
			}
			nui.DropdownMenu(c, theme, anchor, &menu, items, nui.MenuOptions{Label: "Actions menu"})
			nui.Button(c, theme, "Outside", nui.ButtonOptions{})
		})
	}, 600, 480)
	click(t, tt, "Actions")
	if !menu.Open || !tt.Focused("Rename") {
		t.Fatalf("menu initial focus: open=%v texts=%v", menu.Open, tt.Texts())
	}
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeySpace)
	if !checked || menu.Open || calls != 1 {
		t.Fatalf("checkbox menu: checked=%v open=%v calls=%d", checked, menu.Open, calls)
	}
	if !tt.Focused("Actions") {
		t.Fatal("menu did not restore trigger focus")
	}
	click(t, tt, "Actions")
	tt.Type("exp")
	tt.Key(0, ui.KeyRight)
	if !tt.Focused("CSV") {
		t.Fatalf("typeahead/submenu focus: %v", tt.Texts())
	}
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyRight)
	if !tt.Focused("JSON") {
		t.Fatal("nested submenu focus")
	}
	tt.Key(0, ui.KeyEscape)
	if !menu.Open || !tt.Focused("Advanced formats") {
		t.Fatal("Escape did not close only grandchild")
	}
	tt.Key(0, ui.KeyLeft)
	if !tt.Focused("Export") {
		t.Fatal("Left did not restore submenu trigger")
	}
	tt.Key(0, ui.KeyHome)
	swap = true
	tt.Frame()
	tt.Key(0, ui.KeyEnter)
	if chosen != "rename" || calls != 2 {
		t.Fatalf("reorder retargeted action %q/%d", chosen, calls)
	}
	click(t, tt, "Actions")
	click(t, tt, "Outside")
	if menu.Open {
		t.Fatal("outside dismissal")
	}
	tt.Advance(time.Second)
	if tt.FramePending() {
		t.Fatal("menu did not settle")
	}
}

func TestMenuPointerContextAndEmpty(t *testing.T) {
	var menu, context nui.MenuState
	result := ""
	empty := false
	tt := ui.NewTester(func(c *ui.Context) {
		theme := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(16).Gap(24).Children(func() {
			anchor := nui.Button(c, theme, "More", nui.ButtonOptions{})
			items := []nui.MenuItem{{ID: "export", Label: "Export", Children: []nui.MenuItem{{ID: "csv", Label: "Export CSV", OnSelect: func() { result = "csv" }}}}}
			if empty {
				items = nil
			}
			nui.DropdownMenu(c, theme, anchor, &menu, items, nui.MenuOptions{})
			target := nui.Button(c, theme, "Context target", nui.ButtonOptions{}).Width(240).Height(80)
			nui.ContextMenu(c, theme, target, &context, []nui.MenuItem{{ID: "copy", Label: "Copy path", OnSelect: func() { result = "copy" }}}, nui.MenuOptions{Label: "Context actions"})
		})
	}, 600, 400)
	click(t, tt, "More")
	r := bounds(t, tt, "Export")
	tt.Move(r.X+5, r.Y+5)
	tt.Advance(50 * time.Millisecond)
	tt.Move(580, 380)
	tt.Advance(200 * time.Millisecond)
	if tt.HasText("Export CSV") {
		t.Fatal("submenu opened after pointer left")
	}
	tt.Move(r.X+5, r.Y+5)
	tt.Advance(200 * time.Millisecond)
	if !tt.HasText("Export CSV") {
		t.Fatal("hover submenu absent")
	}
	click(t, tt, "Export CSV")
	if result != "csv" || menu.Open {
		t.Fatal("submenu pointer choice")
	}
	if err := tt.RightClick("Context target"); err != nil {
		t.Fatal(err)
	}
	if !context.Open || !tt.Focused("Copy path") {
		t.Fatal("right click context menu")
	}
	tt.Key(0, ui.KeyEnter)
	if result != "copy" || context.Open {
		t.Fatal("context action")
	}
	click(t, tt, "Context target")
	tt.Key(ui.Shift, ui.KeyF10)
	if !context.Open {
		t.Fatal("Shift+F10")
	}
	tt.Key(0, ui.KeyEscape)
	tt.Key(0, ui.KeyContextMenu)
	if !context.Open {
		t.Fatal("Menu key")
	}
	tt.Key(0, ui.KeyEscape)
	empty = true
	click(t, tt, "More")
	if !tt.HasText("No actions") {
		t.Fatal("empty menu")
	}
	tt.Key(0, ui.KeyEscape)
	if menu.Open {
		t.Fatal("empty menu Escape")
	}
}

func TestModalSlideoverFocusAndDismiss(t *testing.T) {
	modal, side := false, false
	dismissible := false
	outside, closed := 0, 0
	value := "a"
	tt := ui.NewTester(func(c *ui.Context) {
		theme := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(16).Gap(12).Children(func() {
			nui.Button(c, theme, "Open modal", nui.ButtonOptions{}).OnClick(func() { modal = true })
			nui.Button(c, theme, "Open side", nui.ButtonOptions{}).OnClick(func() { side = true })
			nui.Button(c, theme, "Outside", nui.ButtonOptions{}).OnClick(func() { outside++ })
			nui.ModalWithOptions(c, theme, &modal, nui.ModalOptions{Title: "Edit workspace", Description: "Changes stay local", Dismissible: nui.Set(dismissible), OnClose: func() { closed++ }, Footer: func(c *ui.Context, t nui.Theme) {
				nui.Button(c, t, "Done", nui.ButtonOptions{}).OnClick(func() { modal = false })
			}}, func(t nui.Theme) {
				nui.Select(c, t, &value, []nui.Option[string]{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}}, nui.SelectOptions{Label: "Plan"})
			})
			nui.Slideover(c, theme, &side, nui.SlideoverOptions{Title: "Filters", Extent: nui.Set(float32(300))}, func(t nui.Theme) { nui.Button(c, t, "Apply", nui.ButtonOptions{}).OnClick(func() { side = false }) })
		})
	}, 700, 500)
	click(t, tt, "Open modal")
	tt.Key(0, ui.KeyEscape)
	if !modal || closed != 0 {
		t.Fatal("non dismissible modal escaped")
	}
	click(t, tt, "Outside")
	if outside != 0 || !modal {
		t.Fatal("modal background interactive")
	}
	click(t, tt, "Plan")
	tt.Key(0, ui.KeyEscape)
	if !modal {
		t.Fatal("child Escape closed modal")
	}
	dismissible = true
	tt.Frame()
	tt.Key(0, ui.KeyEscape)
	if modal || closed != 1 || !tt.Focused("Open modal") {
		t.Fatalf("modal close state %v/%d", modal, closed)
	}
	click(t, tt, "Open side")
	tt.Advance(300 * time.Millisecond)
	r := bounds(t, tt, "Filters")
	if r.X < 390 || r.W != 300 {
		t.Fatalf("side geometry %v", r)
	}
	click(t, tt, "Outside")
	if outside != 0 || side {
		t.Fatal("slideover backdrop")
	}
	tt.Advance(300 * time.Millisecond)
	// Removing the modal can start the underlying button's hover transition.
	tt.Advance(250 * time.Millisecond)
	if tt.FramePending() {
		t.Fatal("slideover exit kept requesting frames")
	}
	click(t, tt, "Open side")
	tt.Advance(300 * time.Millisecond)
	click(t, tt, "Apply")
	tt.Advance(300 * time.Millisecond)
	if side || !tt.Focused("Open side") {
		t.Fatal("slideover close focus")
	}
}

func TestTooltipTimingAndFocus(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		theme := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(30).Children(func() {
			anchor := nui.Button(c, theme, "Help", nui.ButtonOptions{})
			nui.Tooltip(c, theme, anchor, "Workspace help", nui.TooltipOptions{})
		})
	}, 400, 240)
	r := bounds(t, tt, "Help")
	tt.Move(r.X+4, r.Y+4)
	tt.Advance(500 * time.Millisecond)
	if tt.HasText("Workspace help") {
		t.Fatal("tooltip too early")
	}
	tt.Advance(150 * time.Millisecond)
	if !tt.HasText("Workspace help") {
		t.Fatal("tooltip absent")
	}
	tt.Key(0, ui.KeyEscape)
	if tt.HasText("Workspace help") {
		t.Fatal("tooltip Escape")
	}
	tt.Move(390, 220)
	tt.Key(0, ui.KeyTab)
	if !tt.HasText("Workspace help") {
		t.Fatal("keyboard tooltip")
	}
	click(t, tt, "Help")
	if tt.HasText("Workspace help") {
		t.Fatal("tooltip press dismissal")
	}
}

func TestToastQueueClockPauseAndIsolation(t *testing.T) {
	closed, actions := 0, 0
	seq := 0
	view := func(c *ui.Context) {
		theme := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(12).Children(func() {
			nui.Button(c, theme, "Notify", nui.ButtonOptions{}).OnClick(func() {
				seq++
				nui.Toast(c, theme, nui.ToastOptions{ID: "saved", Title: "Saved", Description: "Local changes", Timeout: time.Second, Action: "Undo", OnAction: func() { actions++ }, OnClose: func() { closed++ }})
			})
			nui.Button(c, theme, "Clear toasts", nui.ButtonOptions{}).OnClick(func() { nui.CloseToast(c, "") })
		})
		nui.ToastViewport(c, theme, nui.ToastViewportOptions{})
	}
	tt := ui.NewTester(view, 500, 400)
	other := ui.NewTester(func(c *ui.Context) {
		theme := nui.App(c, nui.AppConfig{})
		nui.ToastViewport(c, theme, nui.ToastViewportOptions{})
	}, 500, 400)
	click(t, tt, "Notify")
	tt.Advance(200 * time.Millisecond)
	if !tt.HasText("Saved") || other.HasText("Saved") {
		t.Fatal("window toast isolation")
	}
	r := bounds(t, tt, "Saved")
	tt.Move(r.X+10, r.Y+10)
	tt.Advance(2 * time.Second)
	if !tt.HasText("Saved") || closed != 0 {
		t.Fatal("hover did not pause timeout")
	}
	tt.Move(0, 0)
	tt.Advance(1100 * time.Millisecond)
	if tt.HasText("Saved") || closed != 1 {
		t.Fatalf("timeout %d", closed)
	}
	click(t, tt, "Notify")
	tt.Advance(200 * time.Millisecond)
	click(t, tt, "Undo")
	tt.Frame()
	if actions != 1 || closed != 2 {
		t.Fatal("toast action/close repeated")
	}
	click(t, tt, "Notify")
	click(t, tt, "Notify")
	click(t, tt, "Clear toasts")
	if closed != 3 {
		t.Fatal("same-ID replacement or clear callbacks")
	}
	tt.Advance(2 * time.Second)
	if tt.FramePending() {
		t.Fatal("cleared toast left animation")
	}
}

func TestTabsAccordionAndRemount(t *testing.T) {
	tab := "profile"
	value := "Draft"
	var expanded []string
	collapse := false
	changes := 0
	tt := ui.NewTester(func(c *ui.Context) {
		theme := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(12).Gap(12).Children(func() {
			nui.Tabs(c, theme, &tab, []nui.TabItem{{ID: "profile", Label: "Profile", Content: func(c *ui.Context, t nui.Theme) { nui.Input(c, t, &value, nui.InputOptions{Label: "Draft"}) }}, {ID: "blocked", Label: "Blocked", Disabled: true}, {ID: "activity", Label: "Activity", Content: func(c *ui.Context, t nui.Theme) { ui.Text(c, "Recent activity") }}}, nui.TabsOptions{Label: "Sections", KeepMounted: true, OnChange: func(string) { changes++ }})
			nui.Accordion(c, theme, &expanded, []nui.AccordionItem{{ID: "general", Label: "General", Content: func(c *ui.Context, t nui.Theme) { ui.Text(c, "General details") }}, {ID: "blocked", Label: "Unavailable", Disabled: true}, {ID: "privacy", Label: "Privacy", Content: func(c *ui.Context, t nui.Theme) { ui.Text(c, "Privacy details") }}}, nui.AccordionOptions{})
			nui.Collapsible(c, theme, &collapse, nui.CollapsibleOptions{Label: "Advanced"}, func(t nui.Theme) { nui.Input(c, t, &value, nui.InputOptions{Label: "Advanced draft"}) })
		})
	}, 600, 500)
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyRight)
	if tab != "activity" || changes != 1 {
		t.Fatalf("tabs skipped disabled: %s/%d", tab, changes)
	}
	click(t, tt, "Profile")
	click(t, tt, "Draft")
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("Edited")
	click(t, tt, "Activity")
	click(t, tt, "Profile")
	if value != "Edited" {
		t.Fatal("hidden content value")
	}
	click(t, tt, "Draft")
	tt.Key(ui.Ctrl, ui.KeyZ)
	if value != "Draft" {
		t.Fatal("KeepMounted lost editor undo history")
	}
	click(t, tt, "General")
	tt.Advance(250 * time.Millisecond)
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	tt.Advance(250 * time.Millisecond)
	if len(expanded) != 1 || expanded[0] != "privacy" || tt.HasText("General details") {
		t.Fatalf("accordion single open %v", expanded)
	}
	click(t, tt, "Advanced")
	tt.Advance(250 * time.Millisecond)
	click(t, tt, "Advanced")
	tt.Advance(250 * time.Millisecond)
	if tt.HasText("Advanced draft") || tt.FramePending() {
		t.Fatal("collapsible did not unmount/settle")
	}
}

func TestTabsReorderRemoveAndDisabled(t *testing.T) {
	value := "b"
	items := []nui.TabItem{{ID: "a", Label: "First"}, {ID: "b", Label: "Second"}, {ID: "c", Label: "Third"}}
	tt := ui.NewTester(func(c *ui.Context) {
		theme := nui.App(c, nui.AppConfig{})
		nui.Tabs(c, theme, &value, items, nui.TabsOptions{})
	}, 500, 300)
	click(t, tt, "Second")
	items[0], items[1] = items[1], items[0]
	tt.Frame()
	tt.Key(0, ui.KeyRight)
	if value != "a" {
		t.Fatalf("reordered focused identity: %s", value)
	}
	items = items[:1]
	tt.Frame()
	if value != "b" {
		t.Fatal("removed selected ID did not normalize")
	}
	items[0].Disabled = true
	tt.Frame()
	tt.Key(0, ui.KeyRight)
	if value != "b" {
		t.Fatal("all-disabled list changed value")
	}
}

func TestModalIdentitySurvivesOtherOverlayRemoval(t *testing.T) {
	open := false
	draft := "Draft"
	tt := ui.NewTester(func(c *ui.Context) {
		theme := nui.App(c, nui.AppConfig{})
		nui.ToastViewport(c, theme, nui.ToastViewportOptions{})
		nui.Button(c, theme, "Open editor", nui.ButtonOptions{}).OnClick(func() {
			nui.Toast(c, theme, nui.ToastOptions{Title: "Transient", Timeout: 300 * time.Millisecond})
			open = true
		})
		nui.ModalWithOptions(c.Key("editor"), theme, &open, nui.ModalOptions{Title: "Editor"}, func(t nui.Theme) {
			nui.Input(c, t, &draft, nui.InputOptions{Label: "Draft text"})
		})
	}, 600, 500)
	click(t, tt, "Open editor")
	click(t, tt, "Draft text")
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("Edited")
	tt.Move(0, 0)
	tt.Advance(time.Second)
	tt.Advance(250 * time.Millisecond)
	if tt.HasText("Transient") || !tt.Focused("Draft text") {
		t.Fatal("unrelated overlay removal moved modal focus")
	}
	tt.Key(ui.Ctrl, ui.KeyZ)
	if draft != "Draft" {
		t.Fatal("unrelated overlay removal remounted modal editor")
	}
	tt.Key(0, ui.KeyEscape)
	if !tt.Focused("Open editor") {
		t.Fatal("modal opener identity")
	}
}

func TestKeyedOptionalOverlaysPreserveFollowingEditor(t *testing.T) {
	var state nui.MenuState
	popover := false
	draft := "Draft"
	tt := ui.NewTester(func(c *ui.Context) {
		theme := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(20).Gap(12).Children(func() {
			anchor := nui.Button(c, theme, "Actions", nui.ButtonOptions{})
			nui.Tooltip(c.Key("help"), theme, anchor, "Action help", nui.TooltipOptions{})
			nui.DropdownMenu(c.Key("menu"), theme, anchor, &state, []nui.MenuItem{{ID: "one", Label: "One"}}, nui.MenuOptions{})
			help := nui.Button(c, theme, "Help", nui.ButtonOptions{}).OnClick(func() { popover = !popover })
			nui.Popover(c.Key("help-panel"), theme, help, &popover, nui.PopoverOptions{}, func(t nui.Theme) { ui.Text(c, "Details") })
			nui.Input(c, theme, &draft, nui.InputOptions{Label: "Following editor"})
		})
	}, 500, 400)
	click(t, tt, "Following editor")
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("Edited")
	click(t, tt, "Actions")
	tt.Key(0, ui.KeyEscape)
	click(t, tt, "Following editor")
	tt.Key(ui.Ctrl, ui.KeyZ)
	if draft != "Draft" {
		t.Fatal("optional overlay leaked caller key into following editor")
	}
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("Again")
	click(t, tt, "Help")
	tt.Key(0, ui.KeyEscape)
	click(t, tt, "Following editor")
	tt.Key(ui.Ctrl, ui.KeyZ)
	if draft != "Draft" {
		t.Fatal("keyed Popover lost following editor history")
	}
}

func TestPaginationStepperAndRecovery(t *testing.T) {
	page := 1
	step := "one"
	banner := true
	retried, navigated := 0, 0
	allow := false
	tt := ui.NewTester(func(c *ui.Context) {
		theme := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(12).Gap(12).Children(func() {
			nui.Pagination(c, theme, &page, nui.PaginationOptions{Total: 1000, PageSize: 10})
			nui.Stepper(c, theme, &step, []nui.StepItem{{ID: "one", Title: "Details"}, {ID: "two", Title: "Review"}, {ID: "three", Title: "Finish"}}, nui.StepperOptions{CanSelect: func(string) bool { return allow }})
			nui.Breadcrumb(c, theme, []nui.BreadcrumbItem{{ID: "home", Label: "Home", OnSelect: func() { navigated++ }}, {ID: "settings", Label: "Settings"}}, nui.BreadcrumbOptions{})
			nui.Banner(c, theme, &banner, nui.BannerOptions{Title: "Pending changes"})
			nui.ErrorView(c, theme, nui.ErrorOptions{Title: "Could not load", Description: "Try loading the local draft again.", OnRetry: func() { retried++ }})
		})
	}, 800, 620)
	click(t, tt, "Last page")
	if page != 100 {
		t.Fatal("last page")
	}
	click(t, tt, "Next page")
	if page != 100 {
		t.Fatal("disabled next")
	}
	click(t, tt, "First page")
	if page != 1 {
		t.Fatal("first page")
	}
	click(t, tt, "Review")
	if step != "one" {
		t.Fatal("step policy")
	}
	allow = true
	tt.Frame()
	click(t, tt, "Finish")
	if step != "one" {
		t.Fatal("linear future step")
	}
	click(t, tt, "Review")
	if step != "two" {
		t.Fatal("step advance")
	}
	click(t, tt, "Home")
	click(t, tt, "Dismiss banner")
	click(t, tt, "Try again")
	if banner || retried != 1 || navigated != 1 {
		t.Fatal("navigation/recovery")
	}
}
