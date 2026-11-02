package nui_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

func TestInputClearFieldAndEditing(t *testing.T) {
	value := "Draft"
	changes, clears := 0, 0
	disabled := false
	tt := ui.NewTester(func(c *ui.Context) {
		theme := nui.App(c, nui.AppConfig{})
		ui.Column(c).Fill().Padding(16).Children(func() {
			nui.FormField(c, theme, nui.FormFieldOptions{Label: "Display name", Required: true, Disabled: disabled, Error: "A name is required."}, func(inner nui.Theme) {
				nui.Input(c, inner, &value, nui.InputOptions{Clearable: true, ClearLabel: "Clear name", OnChange: func(string) { changes++ }, OnClear: func() { clears++ }}).Width(300)
			})
		})
	}, 400, 240)
	click(t, tt, "Display name *")
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("")
	tt.Compose("lin", 3)
	if value != "" {
		t.Fatal("composition committed early")
	}
	tt.Type("林晓")
	if value != "林晓" {
		t.Fatalf("label focus or composition: %q", value)
	}
	before := changes
	click(t, tt, "Clear name")
	tt.Frame()
	tt.Frame()
	if value != "" || clears != 1 || changes != before+1 {
		t.Fatalf("clear state=%q callbacks=%d/%d", value, clears, changes-before)
	}
	tt.Type("new")
	if value != "new" {
		t.Fatal("clear lost editor focus")
	}
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.SetClipboard("粘贴文字")
	tt.Command("paste")
	tt.Command("undo")
	if value != "new" {
		t.Fatalf("undo: %q", value)
	}
	tt.Command("redo")
	if value != "粘贴文字" {
		t.Fatalf("redo: %q", value)
	}
	disabled = true
	tt.Frame()
	click(t, tt, "Display name *")
	tt.Type("blocked")
	if value != "粘贴文字" {
		t.Fatal("field disabled ignored")
	}
}

func TestTextareaWrapGrowthAndBounds(t *testing.T) {
	value := "Short"
	readonly := false
	tt := ui.NewTester(func(c *ui.Context) {
		theme := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(12).Children(func() {
			nui.Textarea(c, theme, &value, nui.TextareaOptions{Label: "Notes", ReadOnly: readonly, Rows: 2, MaxRows: 4, AutoResize: true}).Width(240).Label("Notes root")
		})
	}, 320, 220)
	short := bounds(t, tt, "Notes root")
	click(t, tt, "Notes")
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type(strings.Repeat("很长的中文说明与 English content。", 25))
	tall := bounds(t, tt, "Notes root")
	if tall.H <= short.H || tall.H > 94 {
		t.Fatalf("autoresize bounds short=%v tall=%v", short, tall)
	}
	tt.Key(ui.Ctrl, ui.KeyEnd)
	tt.Type("末尾")
	if !strings.HasSuffix(value, "末尾") {
		t.Fatal("capped editor lost input")
	}
	readonly = true
	tt.Frame()
	tt.Type("blocked")
	if strings.HasSuffix(value, "blocked") {
		t.Fatal("readonly textarea mutated")
	}
}

func TestFieldSizesAndGroup(t *testing.T) {
	for i, height := range []float32{24, 28, 32, 36, 40} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			value := "A"
			tt := ui.NewTester(func(c *ui.Context) {
				theme := nui.App(c, nui.AppConfig{})
				nui.FormField(c, theme, nui.FormFieldOptions{Label: "Field", Size: nui.Size(i + 1)}, func(inner nui.Theme) { nui.Input(c, inner, &value, nui.InputOptions{}).Label("Root") })
			}, 240, 180)
			if got := bounds(t, tt, "Root").H; got != height {
				t.Fatalf("height %g want %g", got, height)
			}
		})
	}
	value := "team"
	calls := 0
	disabled, vertical := false, false
	tt := ui.NewTester(func(c *ui.Context) {
		theme := nui.App(c, nui.AppConfig{})
		nui.FieldGroup(c, theme, nui.FieldGroupOptions{Size: nui.LG, Disabled: disabled, Vertical: vertical}, []nui.FieldGroupItem{
			{ID: "name", Build: func(c *ui.Context, t nui.Theme) ui.Element {
				return nui.Input(c, t, &value, nui.InputOptions{Label: "Name"}).Width(200).Label("Input root")
			}},
			{ID: "check", Build: func(c *ui.Context, t nui.Theme) ui.Element {
				return nui.Button(c, t, "Check", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { calls++ })
			}},
		})
	}, 400, 200)
	a, b := bounds(t, tt, "Input root"), bounds(t, tt, "Check")
	if a.H != 36 || b.H != 36 || b.X != a.X+a.W-1 {
		t.Fatalf("joined geometry: %v %v", a, b)
	}
	tt.Key(0, ui.KeyTab)
	tt.Type("!")
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyEnter)
	if value != "!" {
		t.Fatalf("input tab focus %q", value)
	}
	if calls != 1 {
		t.Fatal("group keyboard action")
	}
	disabled = true
	vertical = true
	tt.Frame()
	click(t, tt, "Check")
	if calls != 1 {
		t.Fatal("disabled group activated")
	}
	a, b = bounds(t, tt, "Input root"), bounds(t, tt, "Check")
	if b.Y != a.Y+a.H-1 {
		t.Fatal("vertical joining")
	}
}

func TestChoiceGroupsAndSwitch(t *testing.T) {
	selected := []string{"a"}
	radio := "a"
	on := false
	changes := 0
	groupDisabled := false
	options := []nui.Option[string]{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Blocked", Disabled: true}, {Value: "c", Label: "Charlie"}}
	tt := ui.NewTester(func(c *ui.Context) {
		theme := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(12).Gap(16).Children(func() {
			nui.CheckboxGroup(c, theme, &selected, options, nui.CheckboxGroupOptions{Label: "Channels", Disabled: groupDisabled, OnChange: func() { changes++ }})
			nui.RadioGroup(c, theme, &radio, []nui.Option[string]{{Value: "a", Label: "Radio A"}, {Value: "b", Label: "Radio blocked", Disabled: true}, {Value: "c", Label: "Radio C"}}, nui.RadioGroupOptions{Label: "Plan", Disabled: groupDisabled})
			nui.FormField(c, theme, nui.FormFieldOptions{Label: "Sync policy"}, func(inner nui.Theme) { nui.Switch(c, inner, &on, nui.SwitchOptions{Label: "Sync enabled"}) })
		})
	}, 420, 400)
	click(t, tt, "Charlie")
	tt.Frame()
	tt.Frame()
	if len(selected) != 2 || changes != 1 {
		t.Fatalf("checkbox group %v changes=%d", selected, changes)
	}
	click(t, tt, "Blocked")
	if len(selected) != 2 {
		t.Fatal("disabled choice selected")
	}
	options[0], options[2] = options[2], options[0]
	tt.Frame()
	click(t, tt, "Alpha")
	if len(selected) != 1 || selected[0] != "c" {
		t.Fatalf("reordered selection %v", selected)
	}
	click(t, tt, "Radio A")
	tt.Key(0, ui.KeyRight)
	if radio != "c" || !tt.Focused("Radio C") {
		t.Fatalf("radio skip/focus %s", radio)
	}
	tt.Key(0, ui.KeyHome)
	if radio != "a" {
		t.Fatal("radio Home")
	}
	label := bounds(t, tt, "Sync policy")
	tt.ClickAt(label.X+8, label.Y+6)
	tt.Frame()
	if !on {
		t.Fatal("field label did not toggle switch")
	}
	groupDisabled = true
	tt.Frame()
	click(t, tt, "Charlie")
	click(t, tt, "Radio C")
	if len(selected) != 1 || radio != "a" {
		t.Fatal("group disabled ignored")
	}
}

func TestSliderStepOrientationAndDisabled(t *testing.T) {
	value := float64(10)
	disabled, vertical := false, false
	changes := 0
	tt := ui.NewTester(func(c *ui.Context) {
		theme := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(20).Children(func() {
			nui.Slider(c, theme, &value, nui.SliderOptions{Label: "Volume", Min: nui.Set(float64(10)), Max: nui.Set(float64(90)), Step: 5, Disabled: disabled, Vertical: vertical, OnChange: func(float64) { changes++ }}).Width(240)
		})
	}, 300, 240)
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyRight)
	if value != 15 || changes != 1 {
		t.Fatalf("step %g changes=%d", value, changes)
	}
	tt.Key(0, ui.KeyEnd)
	if value != 90 {
		t.Fatal("End")
	}
	tt.Key(0, ui.KeyHome)
	if value != 10 {
		t.Fatal("Home")
	}
	vertical = true
	tt.Frame()
	tt.Key(0, ui.KeyUp)
	if value != 15 {
		t.Fatal("vertical arrow")
	}
	disabled = true
	tt.Frame()
	tt.Key(0, ui.KeyUp)
	if value != 15 {
		t.Fatal("disabled slider")
	}
}

func TestSelectStableHighlightDisabledAndEmptyValue(t *testing.T) {
	value := nui.Optional[string]{}
	options := []nui.Option[nui.Optional[string]]{{Value: nui.Optional[string]{}, Label: "None"}, {Value: nui.Set("a"), Label: "Same", Description: "Asia"}, {Value: nui.Set("b"), Label: "Blocked", Disabled: true}, {Value: nui.Set("c"), Label: "Same", Description: "Europe"}}
	tt := ui.NewTester(func(c *ui.Context) {
		theme := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(12).Children(func() { nui.Select(c, theme, &value, options, nui.SelectOptions{Label: "Region"}).Width(250) })
	}, 400, 340)
	click(t, tt, "Region")
	tt.Key(0, ui.KeyDown)
	options[1], options[3] = options[3], options[1]
	tt.Frame()
	tt.Key(0, ui.KeyEnter)
	if value != nui.Set("a") {
		t.Fatalf("reorder retargeted highlight %v", value)
	}
	click(t, tt, "Region")
	tt.Key(0, ui.KeyHome)
	tt.Key(0, ui.KeyEnter)
	if value.Set {
		t.Fatal("empty optional value")
	}
	click(t, tt, "Region")
	tt.Key(0, ui.KeyEnd)
	tt.Key(0, ui.KeyEnter)
	if value != nui.Set("a") {
		t.Fatal("disabled keyboard filtering")
	}
	options = nil
	tt.Frame()
	click(t, tt, "Region")
	if !tt.HasText("No data") {
		t.Fatal("empty select hidden")
	}
	tt.Key(0, ui.KeyEscape)
}

func TestInputMenuEmptyDisabledCustomAndReorder(t *testing.T) {
	state := nui.SearchState{}
	selected, submitted, cleared := 0, 0, 0
	options := []nui.Option[string]{{Value: "a", Label: "Same", Description: "Asia"}, {Value: "b", Label: "Same", Description: "Europe"}, {Value: "x", Label: "Blocked", Disabled: true}}
	tt := ui.NewTester(func(c *ui.Context) {
		theme := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(12).Children(func() {
			nui.InputMenu(c, theme, &state, options, nui.InputMenuOptions{Label: "Owner", Clearable: true, OnClear: func() { cleared++ }, OnSubmit: func(string) { submitted++ }, AllowCustom: true, FilterFields: []string{"label", "description"}, OnSelect: func(string) { selected++ }}).Width(260)
		})
	}, 400, 350)
	click(t, tt, "Owner")
	tt.Type("Europe")
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	tt.Frame()
	if state.Selected != "b" || state.Query != "Same" || selected != 1 || submitted != 0 {
		t.Fatalf("selection %+v callbacks %d/%d", state, selected, submitted)
	}
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("missing")
	if !tt.HasText("No matching data") {
		t.Fatal("empty popup missing")
	}
	tt.Key(0, ui.KeyEnter)
	if state.Selected != "b" || selected != 1 || submitted != 1 {
		t.Fatal("custom query changed chosen ID or repeated callbacks")
	}
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("Blocked")
	if !tt.HasText("No matching data") {
		t.Fatal("disabled-only popup missing")
	}
	tt.Key(0, ui.KeyEnter)
	if state.Selected != "b" {
		t.Fatal("disabled option selected")
	}
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("Same")
	options[0], options[1] = options[1], options[0]
	tt.Frame()
	if tt.HasText("Europe") {
		t.Fatal("changed data should dismiss index-based popup")
	}
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	if state.Selected != "b" {
		t.Fatalf("reopen identity %v", state)
	}
	click(t, tt, "Clear Owner")
	tt.Frame()
	if state.Query != "" || state.Selected != "" || cleared != 1 {
		t.Fatalf("clear query and identity: %+v, callbacks=%d", state, cleared)
	}
	tt.Key(0, ui.KeyEscape)
	tt.Advance(time.Second)
	if tt.FramePending() {
		t.Fatal("menu did not settle")
	}
}

func TestPopoverNestedDismissAndEdgePosition(t *testing.T) {
	search := nui.SearchState{}
	open := false
	dialog := false
	chosen := "a"
	outside := 0
	tt := ui.NewTester(func(c *ui.Context) {
		theme := nui.App(c, nui.AppConfig{})
		ui.Column(c).Fill().Padding(12).Children(func() {
			nui.Button(c, theme, "Outside", nui.ButtonOptions{}).OnClick(func() { outside++ })
			nui.Button(c, theme, "Dialog", nui.ButtonOptions{}).OnClick(func() { dialog = true })
			ui.Box(c).Grow(1)
			anchor := nui.Button(c, theme, "Details", nui.ButtonOptions{}).OnClick(func() { open = !open })
			nui.Popover(c, theme, anchor, &open, nui.PopoverOptions{Label: "Details panel", Width: nui.Set(float32(230))}, func(inner nui.Theme) {
				ui.Text(c, "Popover content")
				nui.Button(c, inner, "Dismiss", nui.ButtonOptions{}).OnClick(func() { open = false })
			})
			nui.Modal(c, theme, &dialog, "Nested dialog", func(inner nui.Theme) {
				nui.InputMenu(c, inner, &search, nil, nui.InputMenuOptions{Label: "Empty search"})
				nui.Select(c, inner, &chosen, []nui.Option[string]{{Value: "a", Label: "Option A"}, {Value: "b", Label: "Option B"}}, nui.SelectOptions{Label: "Nested select"})
			})
		})
	}, 500, 380)
	click(t, tt, "Details")
	r := bounds(t, tt, "Details panel")
	if r.Y+r.H > 380 || r.Y < 0 {
		t.Fatalf("edge popup %v", r)
	}
	click(t, tt, "Outside")
	if open || outside != 1 {
		t.Fatal("outside press should dismiss and reach target")
	}
	click(t, tt, "Details")
	click(t, tt, "Dismiss")
	tt.Key(0, ui.KeyEnter)
	if !open {
		t.Fatal("popover focus was not restored to opener")
	}
	tt.Key(0, ui.KeyEscape)
	click(t, tt, "Dialog")
	click(t, tt, "Empty search")
	tt.Type("missing")
	// Without options, the query cannot match anything: "No data" (D10).
	if !tt.HasText("No data") {
		t.Fatal("nested empty search hidden")
	}
	tt.Key(0, ui.KeyEscape)
	if !dialog || tt.HasText("No data") {
		t.Fatal("empty search Escape must close only the child popup")
	}
	click(t, tt, "Nested select")
	tt.Key(0, ui.KeyEscape)
	if !dialog {
		t.Fatal("nested select closed dialog")
	}
	tt.Key(0, ui.KeyEscape)
	if dialog {
		t.Fatal("dialog Escape")
	}
}

func TestInputReadonlyAndCompositionSubmission(t *testing.T) {
	value := "Draft"
	readonly, submits := false, 0
	tt := ui.NewTester(func(c *ui.Context) {
		theme := nui.App(c, nui.AppConfig{})
		nui.Input(c, theme, &value, nui.InputOptions{Label: "Name", ReadOnly: readonly, OnSubmit: func(string) { submits++ }}).Width(260)
	}, 400, 150)
	click(t, tt, "Name")
	tt.Key(0, ui.KeyEnd)
	tt.Compose("lin", 3)
	// Tester.Type is the documented candidate commit. Tester.Key injects an
	// ordinary unconsumed application key, bypassing the operating system IME.
	tt.Type("林")
	if submits != 0 {
		t.Fatal("committing composition submitted the form")
	}
	tt.Key(0, ui.KeyEnter)
	if submits != 1 {
		t.Fatal("ordinary Enter did not submit exactly once")
	}
	readonly = true
	tt.Frame()
	before := value
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("blocked")
	tt.SetClipboard("paste")
	tt.Command("paste")
	if value != before {
		t.Fatal("readonly input accepted editing")
	}
	if active, _ := tt.TextInputActive(); active {
		t.Fatal("readonly input left the input method active")
	}
}

func TestPopoverSidePlacement(t *testing.T) {
	for _, placement := range []nui.Placement{nui.RightStart, nui.LeftStart} {
		open := true
		tt := ui.NewTester(func(c *ui.Context) {
			theme := nui.App(c, nui.AppConfig{})
			ui.Box(c).Fill().Center().Children(func() {
				anchor := nui.Button(c, theme, "Anchor", nui.ButtonOptions{})
				nui.Popover(c, theme, anchor, &open, nui.PopoverOptions{Label: "Panel", Placement: placement, Width: nui.Set(float32(140))}, func(inner nui.Theme) {
					ui.Text(c, "Side popup content").Height(80)
				})
			})
		}, 600, 360)
		a, p := bounds(t, tt, "Anchor"), bounds(t, tt, "Panel")
		if p.Y != a.Y || (placement == nui.RightStart && p.X < a.X+a.W) || (placement == nui.LeftStart && p.X+p.W > a.X) {
			t.Fatalf("placement %v: anchor=%v panel=%v", placement, a, p)
		}
	}
}
