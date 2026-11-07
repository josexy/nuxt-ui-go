package nui_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
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

// This records the pinned dependency's selected-text IME cancellation limitation, not acceptance
// of data loss. Numeric PINs disable composition; the new text mode exposes it.
func TestUpstreamPinSelectedIMECancellationParity(t *testing.T) {
	value := "Draft"
	var state nui.PinState
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.PinInput(c, th, &value, &state, nui.PinInputOptions{Label: "PIN", Style: func(slot nui.Slot, e ui.Element) {
			if slot == nui.Item {
				e.Label("Slot")
			}
		}})
	}, 400, 100)
	click(t, tt, "Slot")
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Compose("lin", 3)
	tt.Compose("", 0)
	if value != "" {
		t.Fatalf("upstream cancellation changed to %q; reassess selected-text IME cancellation", value)
	}
	tt.Command("undo")
	if value != "Draft" {
		t.Fatal("undo did not restore PIN text")
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

func TestTextareaUnboundedAndInputLength(t *testing.T) {
	text := strings.Repeat("line\n", 11) + "last"
	value, observed := "", ""
	maxRows := 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Width(320).Children(func() {
			nui.Textarea(c, th, &text, nui.TextareaOptions{Label: "Text", AutoResize: true, MaxRows: maxRows})
			nui.Input(c, th, &value, nui.InputOptions{Label: "Limited", MaxLength: 4, Loading: true, OnChange: func(v string) { observed = v }})
		})
	}, 400, 600)
	full := bounds(t, tt, "Text").H
	maxRows = 8
	tt.Frame()
	if full <= bounds(t, tt, "Text").H {
		t.Fatal("maxrows=0 still capped at eight lines")
	}
	click(t, tt, "Limited")
	tt.Type("a😀bc")
	if value != "a😀b" || observed != value {
		t.Fatalf("loading must remain editable and maxlength must precede notification: %q / %q", value, observed)
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
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.SelectMenu(c, th, &values, &s, items, nui.SelectMenuOptions{Label: "Pick", FilterFields: fields, CreateWhenAlways: always, CreateAtTop: true, OnCreate: func(v string) { created = v }})
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
