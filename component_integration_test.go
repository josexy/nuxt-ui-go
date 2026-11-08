package nui_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

func click(t *testing.T, tt *ui.Tester, label string) {
	t.Helper()
	if err := tt.Click(label); err != nil {
		t.Fatal(err)
	}
}

// clickOver clicks the center of label's box for whatever covers it there.
// Click requires label itself to take the press; use this only where another
// element deliberately receives it: a modal backdrop, a splitter's grip, the
// PIN slots drawn over their hidden editor, or decoration passing the press
// through to its control.
func clickOver(t *testing.T, tt *ui.Tester, label string) {
	t.Helper()
	r := bounds(t, tt, label)
	tt.ClickAt(r.X+r.W/2, r.Y+r.H/2)
}

func TestButtonActionsStatesAndParts(t *testing.T) {
	var clicks, blocked int
	tt := ui.NewTester(func(c *ui.Context) {
		s := nui.Window(c, nui.Light())
		ui.Column(c).Padding(12).Gap(12).Children(func() {
			nui.Button(c.Key("save"), s, "Save", nui.ButtonOptions{Leading: lucide.Check(), Style: func(root, label, leading, trailing ui.Element) { label.FontWeight(700) }}).Width(170).OnClick(func() { clicks++ })
			nui.Button(c.Key("disabled"), s, "Disabled", nui.ButtonOptions{Disabled: true}).OnClick(func() { blocked++ })
			nui.Button(c.Key("loading"), s, "Loading", nui.ButtonOptions{Loading: true, Animate: nui.Set(false)}).OnClick(func() { blocked++ })
		})
	}, 400, 240)
	click(t, tt, "Save")
	tt.Frame()
	tt.Frame()
	if clicks != 1 {
		t.Fatalf("one click delivered %d actions", clicks)
	}
	click(t, tt, "Disabled")
	click(t, tt, "Loading")
	if blocked != 0 {
		t.Fatalf("blocked buttons activated %d times", blocked)
	}
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyEnter)
	if clicks != 2 {
		t.Fatalf("keyboard activation: clicks=%d", clicks)
	}
	tt.Advance(300 * time.Millisecond)
	if tt.FramePending() {
		t.Fatal("static loading indicator should settle")
	}
}

func TestInputCompositionClipboardReadOnlyAndNotices(t *testing.T) {
	var value string
	var changes int
	var readOnly bool
	tt := ui.NewTester(func(c *ui.Context) {
		s := nui.Window(c, nui.Light())
		ui.Column(c).Padding(16).Children(func() {
			p := nui.Input(c.Key("name"), s, &value, nui.InputOptions{Label: "Name", ReadOnly: readOnly, Leading: lucide.Search(), Trailing: lucide.Check(), Style: func(root, editor, leading, trailing ui.Element) { editor.OnChange(func() { changes++ }) }})
			p.Width(300)
		})
	}, 400, 150)
	click(t, tt, "Name")
	tt.Compose("ni", 2)
	if value != "" {
		t.Fatalf("composition committed prematurely: %q", value)
	}
	tt.Type("你好")
	if value != "你好" {
		t.Fatalf("composition commit: %q", value)
	}
	before := changes
	tt.Frame()
	tt.Frame()
	if changes != before {
		t.Fatal("OnChange repeated on a rebuild")
	}
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.SetClipboard("复制内容")
	tt.Command("paste")
	if value != "复制内容" {
		t.Fatalf("paste: %q", value)
	}
	tt.Command("undo")
	if value != "你好" {
		t.Fatalf("undo: %q", value)
	}
	readOnly = true
	tt.Frame()
	tt.Type("not allowed")
	if value != "你好" {
		t.Fatalf("readonly changed: %q", value)
	}
}

func TestSelectLongListAndDuplicateLabels(t *testing.T) {
	options := make([]nui.Option[int], 120)
	for i := range options {
		options[i] = nui.Option[int]{Value: i, Label: fmt.Sprintf("Plan %03d", i)}
	}
	selected := 0
	tt := ui.NewTester(func(c *ui.Context) {
		s := nui.Window(c, nui.Light())
		ui.Column(c).Padding(20).Children(func() {
			nui.Select(c.Key("select"), s, &selected, options, nui.SelectOptions{Label: "Plan"}).Width(220)
		})
	}, 480, 420)
	click(t, tt, "Plan")
	tt.Key(0, ui.KeyEnd)
	tt.Key(0, ui.KeyEnter)
	if selected != 119 {
		t.Fatalf("End/Enter selected %d", selected)
	}
	if !tt.Focused("Plan") {
		t.Fatal("select did not retain trigger focus")
	}
	options = []nui.Option[int]{{Value: 1, Label: "Same", Description: "One"}, {Value: 2, Label: "Same", Description: "Two"}}
	selected = 1
	tt.Frame()
	click(t, tt, "Plan")
	tt.Key(0, ui.KeyEnd)
	tt.Key(0, ui.KeyEnter)
	if selected != 2 {
		t.Fatalf("duplicate labels lost identity: %d", selected)
	}
}

func TestSearchUsesStableIDsAndFilters(t *testing.T) {
	state := nui.SearchState{}
	options := []nui.Option[string]{{Value: "a", Label: "Team", Description: "Asia"}, {Value: "b", Label: "Team", Description: "Europe"}, {Value: "c", Label: "Personal"}}
	tt := ui.NewTester(func(c *ui.Context) {
		s := nui.Window(c, nui.Light())
		ui.Column(c).Padding(16).Children(func() { nui.SearchSelect(c.Key("search"), s, &state, options, "Search").Width(280) })
	}, 440, 350)
	click(t, tt, "Search")
	tt.Type("Europe")
	tt.Key(0, ui.KeyEnter)
	if state.Selected != "b" || state.Query != "Team" {
		t.Fatalf("chosen %#v", state)
	}
	// Clear then pick the other duplicate. Identity must not be the label.
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("Asia")
	tt.Key(0, ui.KeyEnter)
	if state.Selected != "a" {
		t.Fatalf("chosen duplicate %q", state.Selected)
	}
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("no results")
	if tt.HasText("Europe") || tt.HasText("Asia") {
		t.Fatal("old options remain after filtering")
	}
	tt.Key(0, ui.KeyEscape)
	if state.Selected != "a" {
		t.Fatal("uncommitted query changed selection")
	}
}

func TestModalNestedSelectAndFocusRestoration(t *testing.T) {
	var open bool
	selected := "one"
	options := []nui.Option[string]{{Value: "one", Label: "First"}, {Value: "two", Label: "Second"}}
	tt := ui.NewTester(func(c *ui.Context) {
		s := nui.Window(c, nui.Light())
		ui.Column(c).Padding(16).Children(func() {
			nui.Button(c.Key("open"), s, "Open dialog", nui.ButtonOptions{}).OnClick(func() { open = true })
			nui.Modal(c.Key("dialog"), s, &open, "Dialog", func(inner nui.Theme) {
				nui.Select(c.Key("choice"), inner, &selected, options, nui.SelectOptions{Label: "Dialog select"}).FillWidth()
				nui.Button(c, inner, "Close", nui.ButtonOptions{}).OnClick(func() { open = false })
			})
		})
	}, 600, 420)
	click(t, tt, "Open dialog")
	click(t, tt, "Dialog select")
	if !tt.HasText("Second") {
		t.Fatal("nested popup not open")
	}
	tt.Key(0, ui.KeyEscape)
	if !open || tt.HasText("Second") {
		t.Fatal("Escape must close only the select first")
	}
	tt.Key(0, ui.KeyEscape)
	if open || !tt.Focused("Open dialog") {
		t.Fatalf("dialog open=%v, trigger focus=%v", open, tt.Focused("Open dialog"))
	}
	click(t, tt, "Open dialog")
	click(t, tt, "Close")
	tt.Frame()
	if open {
		t.Fatal("dialog did not close after recreation")
	}
}

func TestThemeScopesRestoreAndWindowsStayIndependent(t *testing.T) {
	light, dark := nui.Light(), nui.Dark()
	var inside, after ui.Color
	newView := func(theme nui.Theme) *ui.Tester {
		return ui.NewTester(func(c *ui.Context) {
			nui.Window(c, theme)
			ui.Column(c).Padding(10).Gap(12).Children(func() {
				nui.Themed(c, dark, func(inner nui.Theme) {
					inside = c.Theme().Background
					nui.Button(c, inner, "Local", nui.ButtonOptions{})
				}).Padding(12)
				after = c.Theme().Background
				ui.Text(c, "Window")
			})
		}, 340, 200)
	}
	first := newView(light)
	if inside != dark.Background || after != light.Background {
		t.Fatal("subtree theme leaked")
	}
	second := newView(dark)
	first.Frame()
	if after != light.Background {
		t.Fatal("another window replaced the first theme")
	}
	if first.Image().RGBAAt(0, 0) == second.Image().RGBAAt(0, 0) {
		t.Fatal("independent windows have same background")
	}
}

func TestKeyedEditorsSurviveReorderAndHide(t *testing.T) {
	type row struct{ id, value string }
	rows := []*row{{id: "a", value: "A"}, {id: "b", value: "B"}}
	shown := true
	tt := ui.NewTester(func(c *ui.Context) {
		s := nui.Window(c, nui.Light())
		ui.Column(c).Padding(12).Gap(12).Children(func() {
			if shown {
				for _, r := range rows {
					nui.Input(c.Key(r.id), s, &r.value, nui.InputOptions{Label: r.id}).Width(250)
				}
			}
		})
	}, 400, 260)
	click(t, tt, "b")
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("Edited")
	rows[0], rows[1] = rows[1], rows[0]
	tt.Frame()
	tt.Type(" again")
	if rows[0].value != "Edited again" || rows[1].value != "A" {
		t.Fatalf("reordered values: %+v %+v", rows[0], rows[1])
	}
	shown = false
	tt.Frame()
	tt.Frame()
	shown = true
	tt.Frame()
	click(t, tt, "a")
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("Restored")
	if rows[1].value != "Restored" {
		t.Fatal("reappeared field lost its model")
	}
}
