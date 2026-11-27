package nui_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func TestButtonLoadingPartsAndSquareOverrides(t *testing.T) {
	for _, trailing := range []bool{false, true} {
		var hasLeading, hasTrailing bool
		clicks := 0
		tt := ui.NewTester(func(c *ui.Context) {
			th := nui.App(c, nui.AppConfig{})
			ui.Column(c).AlignItems(ui.Start).Children(func() {
				nui.Button(c, th, "Saving", nui.ButtonOptions{Loading: true, LoadingTrailing: trailing, LoadingIcon: lucide.RefreshCw(), Animate: nui.Set(false), Style: func(root, text, before, after ui.Element) {
					text.Label("Loading label")
					hasLeading = before.Valid()
					hasTrailing = after.Valid()
				}}).OnClick(func() { clicks++ })
				nui.Button(c, th, "", nui.ButtonOptions{AccessibleLabel: "Icon", Leading: lucide.Check()})
				nui.Button(c, th, "", nui.ButtonOptions{AccessibleLabel: "Wide icon", Leading: lucide.Check(), Square: true, UI: nui.Styles{nui.Root: {PaddingX: nui.Set(float32(20))}}})
			})
		}, 400, 160)
		if r := bounds(t, tt, "Icon"); r.W != r.H {
			t.Fatalf("icon-only button is not square: %v", r)
		}
		if r := bounds(t, tt, "Wide icon"); r.W != 60 {
			t.Fatalf("square overrode instance padding: %v", r)
		}
		click(t, tt, "Saving")
		if clicks != 0 || bounds(t, tt, "Loading label").W == 0 {
			t.Fatal("loading lost label or accepted activation")
		}
		if trailing && (!hasTrailing || hasLeading) || !trailing && !hasLeading {
			t.Fatal("loading icon on wrong side")
		}
	}
}

func TestSeparatorContentAndPosition(t *testing.T) {
	for _, position := range []nui.SeparatorPosition{nui.SeparatorStart, nui.SeparatorCenter, nui.SeparatorEnd} {
		lines := 0
		tt := ui.NewTester(func(c *ui.Context) {
			th := nui.App(c, nui.AppConfig{})
			lines = 0
			nui.Separator(c, th, nui.SeparatorOptions{Label: "Section", Icon: lucide.Check(), Size: nui.LG, Type: nui.SeparatorDashed, Position: position, Style: func(slot nui.Slot, e ui.Element) {
				if slot == nui.Indicator {
					lines++
					e.Label(fmt.Sprintf("Line %d", lines))
				}
			}}).Width(300)
		}, 400, 100)
		label, line := bounds(t, tt, "Section"), bounds(t, tt, "Line 1")
		want := 1
		if position == nui.SeparatorCenter {
			want = 2
		}
		if lines != want || line.H != 4 {
			t.Fatalf("separator position=%v lines=%d height=%g", position, lines, line.H)
		}
		if position == nui.SeparatorStart && label.X >= line.X || position != nui.SeparatorStart && label.X <= line.X {
			t.Fatal("content on wrong side")
		}
		_ = tt.Image()
	}
}

func TestModalTooltipAndToastDefaults(t *testing.T) {
	open, full := true, true
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.ModalWithOptions(c, th, &open, nui.ModalOptions{Title: "Full", Fullscreen: full, Overlay: nui.Set(false)}, func(nui.Theme) { ui.Text(c, "Body") })
	}, 600, 400)
	tt.Advance(time.Second) // Measure after the 200 ms scale-in settles.
	if r := bounds(t, tt, "Full"); r.X != 0 || r.Y != 0 || r.W != 600 || r.H != 400 {
		t.Fatalf("fullscreen panel = %v", r)
	}
	full = false
	tt.Frame()
	if r := bounds(t, tt, "Full"); r.W != 512 || r.H >= 400 {
		t.Fatalf("normal panel = %v", r)
	}
	tt.Key(0, ui.KeyEscape)
	if open {
		t.Fatal("transparent overlay broke dismissal")
	}
	placements := []nui.Placement{nui.BottomStart, nui.BottomCenter, nui.BottomEnd, nui.TopStart, nui.TopCenter, nui.TopEnd, nui.LeftStart, nui.LeftCenter, nui.LeftEnd, nui.RightStart, nui.RightCenter, nui.RightEnd}
	for _, placement := range placements {
		tipOpen := true
		tt := ui.NewTester(func(c *ui.Context) {
			th := nui.App(c, nui.AppConfig{})
			a := nui.Button(c, th, "Anchor", nui.ButtonOptions{}).Absolute().Left(250).Top(200).Size(100, 40)
			nui.Tooltip(c, th, a, "Help", nui.TooltipOptions{Open: &tipOpen, Placement: nui.Set(placement), Style: func(slot nui.Slot, e ui.Element) {
				if slot == nui.Panel {
					e.Label("Tip panel")
				}
			}})
		}, 600, 500)
		a, p := bounds(t, tt, "Anchor"), bounds(t, tt, "Tip panel")
		switch placement {
		case nui.BottomStart, nui.BottomCenter, nui.BottomEnd:
			if p.Y != a.Y+a.H+8 {
				t.Fatalf("bottom placement %v: %v", placement, p)
			}
		case nui.TopStart, nui.TopCenter, nui.TopEnd:
			if p.Y+p.H != a.Y-8 {
				t.Fatalf("top placement %v: %v", placement, p)
			}
		case nui.LeftStart, nui.LeftCenter, nui.LeftEnd:
			if p.X+p.W != a.X-8 {
				t.Fatalf("left placement %v: %v", placement, p)
			}
		case nui.RightStart, nui.RightCenter, nui.RightEnd:
			if p.X != a.X+a.W+8 {
				t.Fatalf("right placement %v: %v", placement, p)
			}
		}
	}
	tt = ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Button(c, th, "Notify", nui.ButtonOptions{}).OnClick(func() { nui.Toast(c, th, nui.ToastOptions{Title: "Saved"}) })
		nui.ToastViewport(c, th, nui.ToastViewportOptions{})
	}, 600, 400)
	click(t, tt, "Notify")
	tt.Move(0, 0)
	tt.Advance(4500 * time.Millisecond)
	if !tt.HasText("Saved") {
		t.Fatal("toast expired before five seconds")
	}
	tt.Advance(time.Second)
	if tt.HasText("Saved") {
		t.Fatal("default toast did not expire")
	}
}

func TestCalendarDirectionAndHiddenRange(t *testing.T) {
	selected := nui.DateSelection{Start: nui.Set(nui.Date{Year: 2026, Month: 11, Day: 15})}
	var state nui.CalendarState
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Calendar(c, th, &selected, &state, nui.CalendarOptions{Label: "Dates"})
	}, 400, 400)
	tt.SetDirection(ui.RightToLeft)
	click(t, tt, "Dates 2026-11-15")
	tt.Key(0, ui.KeyRight)
	if state.Active.Day != 14 {
		t.Fatalf("RTL right = %v", state.Active)
	}
	tt.Key(0, ui.KeyLeft)
	if state.Active.Day != 15 {
		t.Fatalf("RTL left = %v", state.Active)
	}
	var rangeState nui.DateRangeInputState
	hidden := false
	tt = ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.InputDateRange(c, th, &selected, &rangeState, nui.InputDateOptions{Label: "Travel", ShowCalendar: true, HideCalendar: hidden})
	}, 600, 500)
	click(t, tt, "Choose Travel")
	if !rangeState.Open {
		t.Fatal("range calendar did not open")
	}
	hidden = true
	tt.Frame()
	if _, found := tt.Find("Choose Travel"); found || rangeState.Open {
		t.Fatal("HideCalendar retained range trigger/panel")
	}
}

func TestFileUploadFieldDisablingCancelsPending(t *testing.T) {
	var state nui.FileUploadState
	defer state.Dispose()
	disabled := true
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.FormField(c, th, nui.FormFieldOptions{Label: "Files", Disabled: disabled}, func(th nui.Theme) {
			nui.FileUpload(c, th, &state, nui.FileUploadOptions{ChooseLabel: "Choose", Choose: func(ctx context.Context) ([]nui.UploadFile, error) { <-ctx.Done(); return nil, ctx.Err() }})
		})
	}, 500, 350)
	click(t, tt, "Choose")
	if state.Pending() {
		t.Fatal("inherited disabled allowed file picker")
	}
	disabled = false
	tt.Frame()
	click(t, tt, "Choose")
	if !state.Pending() {
		t.Fatal("enabled picker did not start")
	}
	disabled = true
	tt.Frame()
	if state.Pending() {
		t.Fatal("disabled form left file picker task active")
	}
}

func TestTimelineAndMarquee(t *testing.T) {
	current := nui.Set(1)
	reverse, horizontal := false, false
	chosen := ""
	markers, connectors := 0, 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		markers, connectors = 0, 0
		nui.Timeline(c, th, []nui.TimelineItem{{ID: "a", Title: "Start"}, {ID: "b", Title: "Build"}, {ID: "c", Title: "Finish"}}, nui.TimelineOptions{Current: current, Reverse: reverse, Horizontal: horizontal, Size: nui.LG, OnSelect: func(id string) { chosen = id }, Style: func(slot nui.Slot, e ui.Element) {
			if slot == nui.Indicator {
				markers++
				e.Label(fmt.Sprintf("Marker %d", markers))
			}
			if slot == nui.SeparatorSlot {
				connectors++
				e.Label(fmt.Sprintf("Connector %d", connectors))
			}
		}}).Width(480)
	}, 600, 400)
	pixel := func(name string) [4]uint32 {
		r := bounds(t, tt, name)
		a, b, c, d := tt.Image().At(int(r.X+r.W/2), int(r.Y+r.H/2)).RGBA()
		return [4]uint32{a, b, c, d}
	}
	if markers != 3 || connectors != 2 || bounds(t, tt, "Marker 1").W != 36 {
		t.Fatal("timeline size or connectors")
	}
	if pixel("Marker 1") != pixel("Marker 2") || pixel("Marker 1") == pixel("Marker 3") {
		t.Fatal("timeline progress colors")
	}
	reverse = true
	tt.Frame()
	if pixel("Marker 2") != pixel("Marker 3") || pixel("Marker 1") == pixel("Marker 3") {
		t.Fatal("reverse timeline progress")
	}
	click(t, tt, "Build")
	if chosen != "b" || current.Value != 1 {
		t.Fatal("timeline selection callback changed controlled progress")
	}
	horizontal = true
	tt.Frame()
	if bounds(t, tt, "Marker 1").Y != bounds(t, tt, "Marker 2").Y || bounds(t, tt, "Connector 1").H != 2 {
		t.Fatal("horizontal timeline geometry")
	}
	var marquee nui.MarqueeState
	pause := false
	tt = ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Marquee(c, th, &marquee, nui.MarqueeOptions{Label: "News", PauseOnHover: pause}, func() { ui.Text(c, strings.Repeat("Update ", 50)).NoWrap() }).Width(300)
	}, 400, 100)
	tt.Move(20, 20)
	tt.Advance(time.Second)
	tt.Frame() // Offset is the phase as of the last build.
	if marquee.Offset == 0 {
		t.Fatal("marquee default unexpectedly pauses on hover")
	}
	pause = true
	tt.Frame()
	before := marquee.Offset
	tt.Advance(time.Second)
	if marquee.Offset != before {
		t.Fatal("pauseOnHover ignored")
	}
}

func TestProgressDirectionAndSegmentClamping(t *testing.T) {
	vertical, inverted := false, false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		bar := nui.Progress(c, th, 25, nui.ProgressOptions{Label: "Progress", Vertical: vertical, Inverted: inverted})
		if !vertical {
			bar.Width(200)
		}
	}, 300, 200)
	barLabel := "Progress"
	sample := func(x, y int) [4]uint32 {
		r := bounds(t, tt, barLabel)
		a, b, c, d := tt.Image().At(int(r.X)+x, int(r.Y)+y).RGBA()
		return [4]uint32{a, b, c, d}
	}
	fill, empty := sample(20, 4), sample(180, 4)
	if fill == empty {
		t.Fatal("progress has no distinct fill")
	}
	inverted = true
	tt.Frame()
	if sample(20, 4) != empty || sample(180, 4) != fill {
		t.Fatal("horizontal inverted fill")
	}
	tt.SetDirection(ui.RightToLeft)
	tt.Frame()
	if sample(20, 4) != fill || sample(180, 4) != empty {
		t.Fatal("RTL inverted fill")
	}
	vertical, inverted = true, false
	tt.SetDirection(ui.LeftToRight)
	tt.Frame()
	if sample(4, 20) != fill || sample(4, 140) != empty {
		t.Fatal("vertical progress must fill from top")
	}
	inverted = true
	tt.Frame()
	if sample(4, 20) != empty || sample(4, 140) != fill {
		t.Fatal("vertical inverted progress must fill from bottom")
	}
	barLabel = "Segments"
	tt = ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.ProgressGroup(c, th, []nui.ProgressSegment{{ID: "a", Label: "First", Value: 300, Color: nui.Primary}, {ID: "b", Label: "Second", Value: 100, Color: nui.Secondary}}, nui.ProgressGroupOptions{Label: "Segments", Max: nui.Set(200.0)}).Width(200)
	}, 300, 200)
	if !tt.HasText("100%") || !tt.HasText("50%") {
		t.Fatal("legend is not clamped percentages")
	}
	first, second := sample(20, 4), sample(180, 4)
	if sample(0, 0) == first {
		t.Fatal("segment paint escaped the rounded track clip")
	}
	if first == second || sample(150, 4) != second {
		t.Fatal("segments were not clamped individually before shrinking")
	}
	tt.SetDirection(ui.RightToLeft)
	tt.Frame()
	if sample(20, 4) != second || sample(180, 4) != first {
		t.Fatal("RTL segments not mirrored")
	}
}

func TestAvatarLiteralAltAndGroupInheritance(t *testing.T) {
	limit := nui.Optional[int]{}
	items := make([]nui.AvatarItem, 7)
	for i := range items {
		items[i] = nui.AvatarItem{ID: fmt.Sprint(i), Avatar: nui.AvatarOptions{Alt: fmt.Sprintf("Person %d", i)}}
	}
	items[0].Avatar.Size = nui.XL
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).AlignItems(ui.Start).Children(func() {
			nui.Avatar(c, th, nui.AvatarOptions{Text: "+12", Alt: "Overflow"})
			nui.Avatar(c, th, nui.AvatarOptions{Alt: "Ada Byron Lovelace"})
			nui.AvatarGroup(c, th, items, nui.AvatarGroupOptions{Max: limit, Size: nui.SM, Color: nui.Secondary})
		})
	}, 500, 200)
	bounds(t, tt, "+12")
	bounds(t, tt, "AB")
	bounds(t, tt, "Person 6")
	if bounds(t, tt, "Person 0").W != 40 || bounds(t, tt, "Person 1").W != 28 {
		t.Fatal("group size must be a fallback, not override explicit child size")
	}
	if a, b := bounds(t, tt, "Person 0"), bounds(t, tt, "Person 1"); b.X-a.X != a.W-6 {
		t.Fatalf("group order or SM overlap: %v / %v", a, b)
	}
	for _, n := range []int{0, -1, 3} {
		limit = nui.Set(n)
		tt.Frame()
		_, visible := tt.Find("Person 6")
		if visible != (n <= 0) {
			t.Fatalf("max=%d last avatar visible=%v", n, visible)
		}
	}
	bounds(t, tt, "+4")
}

// MaxLength itself is covered by the TestInputLimit* tests.
func TestTextareaUnboundedAndLoadingInputEditable(t *testing.T) {
	text := strings.Repeat("line\n", 11) + "last"
	value, observed := "", ""
	maxRows := 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Width(320).Children(func() {
			nui.Textarea(c, th, &text, nui.TextareaOptions{Label: "Text", AutoResize: true, MaxRows: maxRows})
			nui.Input(c, th, &value, nui.InputOptions{Label: "Loading", Loading: true, OnChange: func(v string) { observed = v }})
		})
	}, 400, 600)
	full := bounds(t, tt, "Text").H
	maxRows = 8
	tt.Frame()
	if full <= bounds(t, tt, "Text").H {
		t.Fatal("maxrows=0 still capped at eight lines")
	}
	click(t, tt, "Loading")
	tt.Type("a😀b")
	if value != "a😀b" || observed != value {
		t.Fatalf("a loading input must remain editable: %q / %q", value, observed)
	}
}

func TestTriCheckboxFormFieldActivation(t *testing.T) {
	value := nui.Indeterminate
	changes := 0
	disabled := false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.FormField(c, th, nui.FormFieldOptions{Label: "Permissions", Disabled: disabled}, func(inner nui.Theme) {
			nui.CheckboxTri(c, inner, &value, nui.CheckboxTriOptions{OnStateChange: func(nui.CheckState) { changes++ }})
		})
	}, 300, 180)
	label := bounds(t, tt, "Permissions")
	activate := func() {
		tt.Press(label.X+5, label.Y+8)
		tt.Release(label.X+5, label.Y+8)
		tt.Frame()
	}
	activate()
	if value != nui.Checked || changes != 1 {
		t.Fatalf("mixed label click: value=%v changes=%d", value, changes)
	}
	activate()
	if value != nui.Unchecked || changes != 2 {
		t.Fatal("label cannot toggle checked state")
	}
	disabled = true
	tt.Frame()
	activate()
	if value != nui.Unchecked || changes != 2 {
		t.Fatal("disabled label changed state")
	}
}

func TestSelectMenuFilterAndCreatePolicy(t *testing.T) {
	var values []string
	s := nui.SelectMenuState{Open: true, Query: "team"}
	fields := []string(nil)
	always := false
	created := ""
	items := []nui.ListboxItem{{ID: "alpha", Label: "Alpha", Description: "team", Group: "team"}}
	listSelection := chosenSelection(values)
	s.SetItems(items)
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.SelectMenu(c, th, listSelection, &s, nui.SelectMenuOptions{Label: "Pick", FilterFields: fields, CreateWhenAlways: always, CreateAtTop: true, OnCreate: func(v string) { created = v }})
	}, 500, 400)
	if _, found := tt.Find("Alpha"); found {
		t.Fatal("default filter unexpectedly searched group or description")
	}
	fields = []string{"description"}
	tt.Frame()
	bounds(t, tt, "Alpha")
	if _, found := tt.Find("Create “team”"); found {
		t.Fatal("default creation must be hidden when results exist")
	}
	always = true
	tt.Frame()
	create := bounds(t, tt, "Create “team”")
	if create.W == 0 || create.Y >= bounds(t, tt, "Alpha").Y {
		t.Fatal("create-at-top missing or below results")
	}
	click(t, tt, "Create “team”")
	if created != "team" || !s.Open || s.Query != "team" {
		t.Fatalf("create changed search/selection: %q %+v", created, s)
	}
	s.Query = "alpha"
	fields = []string{"value"}
	tt.Frame()
	if tt.HasText("Create “alpha”") {
		t.Fatal("always creation duplicated an existing value")
	}
}

func TestPinInputTextNumericAndReadOnly(t *testing.T) {
	value, completed := "", ""
	var state nui.PinState
	numeric, readonly := false, false
	count := 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		count = 0
		nui.PinInput(c, th, &value, &state, nui.PinInputOptions{Label: "Code", Numeric: numeric, ReadOnly: readonly, SeparatorEvery: 2, OnComplete: func(v string) { completed = v }, Style: func(part nui.Slot, e ui.Element) {
			if part == nui.Item {
				count++
				e.Label(fmt.Sprintf("Slot %d", count))
			}
		}})
	}, 400, 100)
	if count != 5 {
		t.Fatalf("default length = %d", count)
	}
	click(t, tt, "Slot 1")
	tt.Type("A中😀bcZ")
	if value != "A中😀bc" || completed != value {
		t.Fatalf("text paste and completion = %q / %q", value, completed)
	}
	click(t, tt, "Slot 3")
	tt.Type("X")
	if value != "A中Xbc" {
		t.Fatalf("Unicode slot replacement = %q", value)
	}
	readonly = true
	tt.Frame()
	click(t, tt, "Slot 2")
	tt.Key(0, ui.KeyBackspace)
	if value != "A中Xbc" {
		t.Fatal("readonly PIN changed")
	}
	readonly, numeric, value, completed = false, true, "", ""
	tt.Frame()
	click(t, tt, "Slot 1")
	tt.Type("A12中34😀56")
	if value != "12345" || completed != value {
		t.Fatalf("numeric filtering = %q / %q", value, completed)
	}
}

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
	selectState := submittedSelect(options)
	tt := ui.NewTester(func(c *ui.Context) {
		s := nui.Window(c, nui.Light())
		ui.Column(c).Padding(20).Children(func() {
			nui.Select(c.Key("select"), s, &selected, selectState, nui.SelectOptions{Label: "Plan"}).Width(220)
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
	selectState.SetOptions(options)
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
	state.SetOptions(options)
	tt := ui.NewTester(func(c *ui.Context) {
		s := nui.Window(c, nui.Light())
		ui.Column(c).Padding(16).Children(func() { nui.SearchSelect(c.Key("search"), s, &state, "Search").Width(280) })
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
	selectState := submittedSelect(options)
	tt := ui.NewTester(func(c *ui.Context) {
		s := nui.Window(c, nui.Light())
		ui.Column(c).Padding(16).Children(func() {
			nui.Button(c.Key("open"), s, "Open dialog", nui.ButtonOptions{}).OnClick(func() { open = true })
			nui.Modal(c.Key("dialog"), s, &open, "Dialog", func(inner nui.Theme) {
				nui.Select(c.Key("choice"), inner, &selected, selectState, nui.SelectOptions{Label: "Dialog select"}).FillWidth()
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

// These helpers run in fixture setup, before the view is mounted.
func submittedSelect[T comparable](options []nui.Option[T]) *nui.SelectState[T] {
	s := &nui.SelectState[T]{}
	s.SetOptions(options)
	return s
}

func chosenSelection[T comparable](values []T) *nui.SelectionState[T] {
	s := &nui.SelectionState[T]{}
	s.Set(values)
	return s
}
