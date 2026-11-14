package nui_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

func withinDIP(a, b float32) bool { return math.Abs(float64(a-b)) <= 1 }

func TestTabsArrowKeysRevealTheNextTriggerMinimally(t *testing.T) {
	value, width := "t1", float32(2000)
	var scroll ui.ScrollState
	items := make([]nui.TabItem, 10)
	for i := range items {
		items[i] = nui.TabItem{ID: fmt.Sprint("t", i+1), Label: fmt.Sprint("Tab ", i+1)}
	}
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.ScrollArea(c.Key("strip"), th, nui.ScrollAreaOptions{Direction: nui.ScrollHorizontal, State: &scroll}, func() {
			nui.Tabs(c.Key("tabs"), th, &value, items, nui.TabsOptions{Label: "Sections", Underline: true})
		}).Width(width).Label("Strip")
	}, 2100, 300)
	origin := bounds(t, tt, "Strip").X
	fourth, fifth := bounds(t, tt, "Tab 4"), bounds(t, tt, "Tab 5")

	// Every trigger visible: moving between them never scrolls.
	click(t, tt, "Tab 1")
	for range 3 {
		tt.Key(0, ui.KeyRight)
	}
	if !tt.Focused("Tab 4") || scroll.X != 0 {
		t.Fatalf("focus moved between visible triggers: Tab 4 focused %v, offset %v", tt.Focused("Tab 4"), scroll.X)
	}

	// Only the first four triggers show.
	width = fourth.X + fourth.W - origin + (fifth.X-(fourth.X+fourth.W))/2
	tt.Frame()
	if scroll.X != 0 || !tt.Focused("Tab 4") {
		t.Fatalf("narrowing scrolled to %v", scroll.X)
	}
	tt.Key(0, ui.KeyRight)
	if !tt.Focused("Tab 5") {
		t.Fatal("Right did not move focus to Tab 5")
	}
	want := fifth.X + fifth.W - origin - width
	if !withinDIP(scroll.X, want) {
		t.Fatalf("revealing Tab 5 scrolled to %v, want the minimum %v", scroll.X, want)
	}
	strip, tab := bounds(t, tt, "Strip"), bounds(t, tt, "Tab 5")
	if tab.X < strip.X-1 || tab.X+tab.W > strip.X+strip.W+1 {
		t.Fatalf("Tab 5 %v is not shown in %v", tab, strip)
	}
}

func TestFormFocusesAndRevealsTheFirstInvalidField(t *testing.T) {
	var form nui.FormState
	var scroll ui.ScrollState
	name, email := "Ada", ""
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.ScrollArea(c.Key("page"), th, nui.ScrollAreaOptions{State: &scroll}, func() {
			nui.Form(c.Key("form"), th, &form, nui.FormOptions{LoadingAuto: nui.Set(false), Validate: func(v nui.FormValues) nui.FormErrors {
				if v["email"] == "" {
					return nui.FormErrors{"email": "Required"}
				}
				return nil
			}}, func(ft nui.Theme) {
				nui.Button(c.Key("submit"), ft, "Submit", nui.ButtonOptions{}).OnClick(form.Submit)
				nui.FormField(c.Key("name"), ft, nui.FormFieldOptions{Name: "name", Value: name, Label: "Name"}, func(ft nui.Theme) { nui.Input(c, ft, &name, nui.InputOptions{Label: "Name input"}) })
				ui.Box(c.Key("spacer")).Height(600)
				nui.FormField(c.Key("email"), ft, nui.FormFieldOptions{Name: "email", Value: email, Label: "Email"}, func(ft nui.Theme) { nui.Input(c, ft, &email, nui.InputOptions{Label: "Email input"}) })
			})
		}).Height(240).Width(400).Label("Page")
	}, 500, 300)
	click(t, tt, "Submit")
	tt.Frame()
	if !tt.Focused("Email input") {
		t.Fatal("the invalid field did not take the focus")
	}
	page, field := bounds(t, tt, "Page"), bounds(t, tt, "Email input")
	if scroll.Y <= 0 || field.Y < page.Y || field.Y+field.H > page.Y+page.H+1 {
		t.Fatalf("the invalid field %v is not scrolled into %v (offset %v)", field, page, scroll.Y)
	}
}

func TestOverlaysReturnFocusWithoutScrolling(t *testing.T) {
	for _, kind := range []string{"Modal", "DropdownMenu"} {
		t.Run(kind, func(t *testing.T) {
			var scroll ui.ScrollState
			scroll.Y = 200
			open := false
			var menu nui.MenuState
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				nui.ScrollArea(c.Key("page"), th, nui.ScrollAreaOptions{State: &scroll}, func() {
					ui.Box(c.Key("above")).Height(300)
					trigger := nui.Button(c.Key("trigger"), th, "Open", nui.ButtonOptions{})
					if kind == "Modal" {
						trigger.OnClick(func() { open = true })
					} else {
						nui.DropdownMenu(c.Key("menu"), th, trigger, &menu, []nui.MenuItem{{ID: "copy", Label: "Copy"}}, nui.MenuOptions{})
					}
					ui.Box(c.Key("below")).Height(900)
				}).Height(300).Width(400).Label("Page")
				if kind == "Modal" {
					nui.ModalWithOptions(c.Key("dialog"), th, &open, nui.ModalOptions{Title: "Dialog", Transition: nui.Set(false)}, func(nui.Theme) { ui.Text(c, "Dialog body") })
				}
			}, 500, 400)
			if scroll.Y != 200 {
				t.Fatalf("initial offset %v", scroll.Y)
			}
			page, button := bounds(t, tt, "Page"), bounds(t, tt, "Open")
			if button.Y < page.Y || button.Y+button.H > page.Y+page.H {
				t.Fatalf("the trigger %v is not fully visible in %v", button, page)
			}
			click(t, tt, "Open")
			if kind == "Modal" && !tt.HasText("Dialog body") || kind == "DropdownMenu" && !tt.HasText("Copy") {
				t.Fatalf("%s did not open: %v", kind, tt.Texts())
			}
			tt.Key(0, ui.KeyEscape)
			tt.Frame()
			if !tt.Focused("Open") {
				t.Fatalf("focus did not return to the trigger after Escape")
			}
			if scroll.Y != 200 {
				t.Fatalf("returning focus scrolled to %v", scroll.Y)
			}
		})
	}
}

func TestFocusKeptByComponentsLeavesUserScrolling(t *testing.T) {
	var scroll ui.ScrollState
	draft, partial := "", ""
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.ScrollArea(c.Key("page"), th, nui.ScrollAreaOptions{State: &scroll}, func() {
			nui.Input(c.Key("draft"), th, &draft, nui.InputOptions{Label: "Draft"})
			ui.Box(c.Key("gap")).Height(170)
			nui.Input(c.Key("partial"), th, &partial, nui.InputOptions{Label: "Partial"})
			ui.Box(c.Key("below")).Height(900)
		}).Height(220).Width(400).Label("Page")
	}, 500, 300)

	click(t, tt, "Draft")
	page := bounds(t, tt, "Page")
	tt.Scroll(page.X+page.W/2, page.Y+page.H/2, 0, 400)
	tt.Frame()
	scrolled := scroll.Y
	if scrolled <= 0 {
		t.Fatal("the user could not scroll")
	}
	for range 3 {
		tt.Frame()
	}
	if scroll.Y != scrolled || !tt.Focused("Draft") {
		t.Fatalf("rebuilds moved the user's scroll from %v to %v (focused %v)", scrolled, scroll.Y, tt.Focused("Draft"))
	}

	scroll.Y = 0
	tt.Frame()
	// Find gives the visible part of a box: shorter than the same input whole.
	field, whole := bounds(t, tt, "Partial"), bounds(t, tt, "Draft")
	if field.H <= 0 || field.H >= whole.H {
		t.Fatalf("Partial %v is not partially visible in %v (whole input %v)", field, page, whole)
	}
	x, y := field.X+field.W/2, field.Y+field.H/2
	tt.Press(x, y)
	if scroll.Y != 0 {
		t.Fatalf("a pointer press scrolled to %v", scroll.Y)
	}
	tt.Release(x, y)
	tt.Frame()
	if !tt.Focused("Partial") || scroll.Y != 0 {
		t.Fatalf("clicking the visible part: focused %v, offset %v", tt.Focused("Partial"), scroll.Y)
	}
}

// The documented app-side escape hatch: giving the focus back without
// scrolling a container the user moved.
func TestFocusWithPreventScrollKeepsTheOffset(t *testing.T) {
	var scroll ui.ScrollState
	restore := false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.ScrollArea(c.Key("page"), th, nui.ScrollAreaOptions{State: &scroll}, func() {
			cell := ui.Box(c.Key("cell")).Height(40).Focusable().Label("Cell")
			if restore {
				cell.Focus(ui.FocusOptions{PreventScroll: true})
			}
			ui.Box(c.Key("below")).Height(900)
		}).Height(200).Width(300)
	}, 400, 300)
	scroll.Y = 300
	tt.Frame()
	restore = true
	tt.Frame()
	if !tt.Focused("Cell") || scroll.Y != 300 {
		t.Fatalf("PreventScroll focus: focused %v, offset %v", tt.Focused("Cell"), scroll.Y)
	}
}
