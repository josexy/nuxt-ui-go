package nui_test

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"os/exec"
	"slices"
	"testing"
	"testing/fstest"
	"time"

	"github.com/egoist/mygo/i18n"
	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

// Default values aligned with Nuxt UI v4.12 (fix plan section 1, D1–D15).

func TestStepperIsLinearByDefault(t *testing.T) {
	for _, tc := range []struct {
		linear nui.Optional[bool]
		jumps  bool
	}{{nui.Optional[bool]{}, false}, {nui.Set(true), false}, {nui.Set(false), true}} {
		step := "one"
		tt := ui.NewTester(func(c *ui.Context) {
			th := nui.App(c, nui.AppConfig{})
			nui.Stepper(c, th, &step, []nui.StepItem{{ID: "one", Title: "One"}, {ID: "two", Title: "Two"}, {ID: "three", Title: "Three"}}, nui.StepperOptions{Linear: tc.linear})
		}, 600, 200)
		click(t, tt, "Three")
		if (step == "three") != tc.jumps {
			t.Fatalf("Linear=%+v: clicking the third step selected %q", tc.linear, step)
		}
		click(t, tt, "Two")
		if step != "two" {
			t.Fatalf("Linear=%+v: the next step is always reachable, got %q", tc.linear, step)
		}
	}
}

func TestInputTimeFollowsLocaleHourCycle(t *testing.T) {
	clock := nui.Set(nui.ClockTime{Hour: 13, Minute: 5})
	var s nui.TimeInputState
	var twelve nui.Optional[bool]
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.InputTime(c, th, &clock, &s, nui.InputTimeOptions{Label: "Time", TwelveHour: twelve})
	}, 400, 120)
	tt.SetLocale("en")
	if s.Text != "01:05 PM" {
		t.Fatalf("en: %q, want 12-hour", s.Text)
	}
	tt.SetLocale("zh-Hans")
	if s.Text != "13:05" {
		t.Fatalf("zh-Hans: %q, want 24-hour", s.Text)
	}
	// The user's regional formats win over the language's default.
	tt.SetRegion(i18n.Region{Locale: "en-US", HourCycle: 12})
	tt.Frame()
	if s.Text != "01:05 PM" {
		t.Fatalf("zh-Hans with 12-hour regional formats: %q", s.Text)
	}
	twelve = nui.Set(false)
	tt.Frame()
	if s.Text != "13:05" {
		t.Fatalf("explicit 24-hour: %q", s.Text)
	}
}

func TestToastActionsDefaultToSolidXSInToastColor(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		nui.Button(c, th, "Notify", nui.ButtonOptions{}).OnClick(func() {
			nui.Toast(c, th, nui.ToastOptions{Title: "Saved", Timeout: time.Hour, Color: nui.Error, Actions: []nui.ToastActionOptions{{Label: "Undo"}}})
		})
		nui.ToastViewport(c, th, nui.ToastViewportOptions{})
	}, 800, 600)
	click(t, tt, "Notify")
	tt.Advance(time.Second)
	undo := bounds(t, tt, "Undo")
	if undo.H != 24 {
		t.Fatalf("action is %v DIPs tall, want XS 24", undo.H)
	}
	if c := tt.Image().RGBAAt(int(undo.X)+3, int(undo.Y+undo.H/2)); c != rgba(nui.Light().Error) {
		t.Fatalf("action surface %v, want the toast's solid error color", c)
	}
}

func TestUserWithoutAvatarShowsNone(t *testing.T) {
	avatar := nui.AvatarOptions{}
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(10).Children(func() {
			nui.User(c, th, nui.UserOptions{Name: "Lin Chen", Description: "Admin", Avatar: avatar})
		})
	}, 300, 100)
	// The avatar would take the name as its Alt, so measure the description.
	plain := bounds(t, tt, "Admin").X
	avatar = nui.AvatarOptions{Text: "LC"}
	tt.Frame()
	if with := bounds(t, tt, "Admin").X; plain != 10 || with <= plain {
		t.Fatalf("name at %v without an avatar and %v with one", plain, with)
	}
}

func TestFileUploadAreaIsTheButton(t *testing.T) {
	var s nui.FileUploadState
	calls := 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(20).Children(func() {
			nui.Button(c, th, "Before", nui.ButtonOptions{})
			nui.FileUpload(c, th, &s, nui.FileUploadOptions{Label: "Attachments", Choose: func(context.Context) ([]nui.UploadFile, error) { calls++; return nil, nil }}).Width(300)
		})
	}, 400, 400)
	t.Cleanup(s.Dispose)
	if tt.HasText("Choose files") {
		t.Fatal("the interactive area still shows a Choose button")
	}
	click(t, tt, "Before")
	tt.Key(0, ui.KeyTab)
	if !tt.Focused("Attachments") {
		t.Fatal("the area is not focusable")
	}
	tt.Key(0, ui.KeyEnter)
	awaitCommandResults(t, tt, func() bool { return !s.Pending() })
	tt.Key(0, ui.KeySpace)
	awaitCommandResults(t, tt, func() bool { return !s.Pending() })
	if calls != 2 {
		t.Fatalf("Enter and Space opened the picker %d times, want 2", calls)
	}
}

func TestErrorViewDefaults(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		nui.ErrorView(c, th, nui.ErrorOptions{Status: "404", Title: "Not found", OnRetry: func() {}, Style: func(slot nui.Slot, e ui.Element) {
			if slot == nui.Root {
				e.Label("Error region")
			}
		}})
	}, 600, 500)
	// Below the 640 DIP breakpoint Nuxt uses a 36/40 title. Native fonts
	// retain their own ascent and descent when they exceed the CSS line box.
	if title := bounds(t, tt, "Not found"); title.H < 40 || title.H > 50 {
		t.Fatalf("title line %v, want a 36 DIP title at narrow width", title.H)
	}
	retry := bounds(t, tt, "Try again")
	if retry.H != 36 || tt.Image().RGBAAt(int(retry.X)+3, int(retry.Y+retry.H/2)) != rgba(nui.Light().Primary) {
		t.Fatalf("retry button %v is not a LG primary solid button", retry)
	}
	errorColor := func(c color.RGBA) bool { return near(c, rgba(nui.Light().Error), 12) }
	if countWhere(tt.Image(), ui.Rect{W: 600, H: 500}, errorColor) != 0 {
		t.Fatal("an error icon is drawn although none was given")
	}
}

func TestEmptyTextsAndLegacyOverrides(t *testing.T) {
	var none []string
	var s nui.ListboxState
	listSelection := chosenSelection(none)
	s.SetItems(nil)
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Listbox(c, th, listSelection, &s, nui.ListboxOptions{Label: "Empty"})
	}, 300, 300)
	if !tt.HasText("No data") {
		t.Fatalf("empty Listbox: %v", tt.Texts())
	}
	tt.SetLocale("zh-Hans")
	if !tt.HasText("没有数据") {
		t.Fatalf("zh-Hans empty Listbox: %v", tt.Texts())
	}
}

func TestMarqueeDefaults(t *testing.T) {
	var s nui.MarqueeState
	builds := 0
	repeat := nui.Optional[int]{}
	parts := 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		builds, parts = 0, 0
		nui.Marquee(c, th, &s, nui.MarqueeOptions{Label: "News", Repeat: repeat, Style: func(slot nui.Slot, e ui.Element) {
			if slot == nui.Body {
				parts++
				e.Background(ui.RGB(230, 240, 250))
			}
		}}, func() {
			builds++
			ui.Text(c, "Short").TextColor(ui.RGB(0, 0, 0))
		}).Width(400)
	}, 500, 100)
	tt.Frame() // Geometry is committed after the first frame.
	tt.Advance(time.Second)
	tt.Frame() // Offset is the phase as of the last build.
	if s.Offset == 0 {
		t.Fatal("short content does not scroll")
	}
	if builds != 4 {
		t.Fatalf("content built %d times, want four copies", builds)
	}
	// The edges fade into the background: the first pixel is the background.
	if c := tt.Image().RGBAAt(1, 22); c != white {
		t.Fatalf("edge pixel %v, want the faded background", c)
	}
	region, _ := tt.Find("News")
	content, _ := tt.Find("Short")
	if dy := content.Y + content.H/2 - region.Y - region.H/2; dy < -.1 || dy > .1 {
		t.Fatalf("marquee content is not centered across its motion axis: region=%v content=%v", region, content)
	}
	s.Paused = true
	for _, copies := range []int{1, 2, 6, 0, 4} {
		repeat = nui.Set(copies)
		s.Offset = -1000
		tt.Frame()
		if builds != copies || parts != copies {
			t.Fatalf("Repeat=%d: content=%d, styled parts=%d", copies, builds, parts)
		}
		if s.Offset < 0 || s.Offset >= 464 {
			t.Fatalf("phase outside the content period: %g", s.Offset)
		}
		if copies == 0 && (s.Offset != 0 || tt.HasText("Short")) {
			t.Fatal("Repeat=0 retained content or phase")
		}
	}
}

func TestCheckboxGroupMaxKeepsItemsFocusable(t *testing.T) {
	selected := []string{"a"}
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.CheckboxGroup(c, th, &selected, []nui.Option[string]{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}}, nui.CheckboxGroupOptions{Label: "Letters", Max: nui.Set(1)})
	}, 300, 200)
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyDown)
	if !tt.Focused("Beta") {
		t.Fatal("the item beyond Max is not focusable")
	}
	tt.Key(0, ui.KeySpace)
	click(t, tt, "Beta")
	if !slices.Equal(selected, []string{"a"}) {
		t.Fatalf("the item beyond Max was checked: %v", selected)
	}
	click(t, tt, "Alpha")
	click(t, tt, "Beta")
	if !slices.Equal(selected, []string{"b"}) {
		t.Fatalf("after unchecking, Beta could not be checked: %v", selected)
	}
}

func TestInputTagsSplitOnCommasOnly(t *testing.T) {
	var tags []string
	var s nui.TagsState
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.InputTags(c, th, &tags, &s, nui.InputTagsOptions{Label: "Tags"})
	}, 400, 200)
	click(t, tt, "Tags")
	tt.Type("a;b,")
	if !slices.Equal(tags, []string{"a;b"}) {
		t.Fatalf("tags %q; want a comma to end the tag, keeping the semicolon", tags)
	}
}

func TestCarouselPositionIsAccessibleNotVisible(t *testing.T) {
	var s nui.CarouselState
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Carousel(c, th, &s, []nui.CarouselItem{{ID: "a", Label: "First"}, {ID: "b", Label: "Second"}}, nui.CarouselOptions{Label: "Slides"}).Width(300)
	}, 400, 300)
	if tt.HasText("1 / 2") {
		t.Fatal("the visible counter is still shown")
	}
	if !tt.HasText("Slides") {
		t.Fatal("the carousel lost its name")
	}
}

// An application's override of a former empty-state key keeps working (D10).
// Registration is process-wide, so the check runs in a child process.
func TestLegacyEmptyTextOverrideStillApplies(t *testing.T) {
	if os.Getenv("NUI_LEGACY_OVERRIDE_CHILD") != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestLegacyEmptyTextOverrideStillApplies$", "-test.count=1")
		cmd.Env = append(os.Environ(), "NUI_LEGACY_OVERRIDE_CHILD=1", "CGO_ENABLED=0")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("child: %v\n%s", err, out)
		}
		return
	}
	app, err := i18n.NewCatalog(fstest.MapFS{"en.json": {Data: []byte(`{"@@locale": "en", "nuxtui.noResults": "Nothing matched"}`)}}, "en")
	if err != nil {
		t.Fatal(err)
	}
	i18n.Register(app)
	var none []string
	var s nui.ListboxState
	listSelection := chosenSelection(none)
	s.SetItems(nil)
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Listbox(c, th, listSelection, &s, nui.ListboxOptions{Label: "Empty"})
	}, 300, 300)
	tt.SetLocale("en")
	if !tt.HasText("Nothing matched") || tt.HasText("No data") {
		t.Fatalf("the application's nuxtui.noResults override was ignored: %v", tt.Texts())
	}
}

// Regressions for parameters that were accepted but ignored, broken rendering
// and right-to-left defects found by the Nuxt UI v4.12 parity review.

func rgba(c ui.Color) color.RGBA { return color.RGBA{c.R, c.G, c.B, 255} }

var white = color.RGBA{255, 255, 255, 255}

// countColor counts the pixels of r that match want exactly.
func countColor(im *image.RGBA, r ui.Rect, want color.RGBA) int {
	n := 0
	for y := int(r.Y); y < int(r.Y+r.H); y++ {
		for x := int(r.X); x < int(r.X+r.W); x++ {
			if im.RGBAAt(x, y) == want {
				n++
			}
		}
	}
	return n
}

// topCorners samples the top row one pixel inside each side: a rounded corner
// shows the background there, while FieldGroup's 1 DIP overlap stays outside.
func topCorners(im *image.RGBA, r ui.Rect) (left, right color.RGBA) {
	y := int(r.Y)
	return im.RGBAAt(int(r.X)+1, y), im.RGBAAt(int(r.X+r.W)-2, y)
}

// chevronPointsRight reports whether the chevron drawn on a white background
// in r points right: its open side, at the left, spreads wider than its tip.
func chevronPointsRight(t *testing.T, im *image.RGBA, r ui.Rect) bool {
	t.Helper()
	ink := func(x, y int) bool { c := im.RGBAAt(x, y); return c.R < 235 || c.G < 235 || c.B < 235 }
	x0, x1, y0, y1 := int(r.X), int(r.X+r.W), int(r.Y), int(r.Y+r.H)
	minX, maxX := math.MaxInt, -1
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			if ink(x, y) {
				minX, maxX = min(minX, x), max(maxX, x)
			}
		}
	}
	if maxX < 0 {
		t.Fatal("no chevron drawn in", r)
	}
	spread := func(x int) int {
		lo, hi := math.MaxInt, -1
		for y := y0; y < y1; y++ {
			if ink(x, y) {
				lo, hi = min(lo, y), max(hi, y)
			}
		}
		return hi - lo
	}
	return spread(minX) > spread(maxX)
}

func TestCheckboxGroupForwardsItemPartStyles(t *testing.T) {
	marker := ui.RGB(255, 0, 255)
	var selected []string
	calls := map[nui.Slot]int{}
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		clear(calls)
		ui.Column(c).Padding(16).Children(func() {
			nui.CheckboxGroup(c, th, &selected, []nui.Option[string]{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}}, nui.CheckboxGroupOptions{Label: "Letters",
				UI:    nui.Styles{nui.Label: {Foreground: nui.Set(marker)}},
				Style: func(slot nui.Slot, e ui.Element) { calls[slot]++ }})
		})
	}, 300, 160)
	if calls[nui.Root] != 1 || calls[nui.Item] != 2 || calls[nui.Label] != 2 || calls[nui.Indicator] != 2 {
		t.Fatalf("group Style calls = %v; want one Root and per-item Item, Label and Indicator", calls)
	}
	if countColor(tt.Image(), bounds(t, tt, "Letters"), rgba(marker)) == 0 {
		t.Fatal("group UI[Label] did not color the item labels")
	}
	click(t, tt, "Beta")
	if len(selected) != 1 || selected[0] != "b" {
		t.Fatalf("selection after styling = %v", selected)
	}
}

func TestCheckboxUsesCheckedIcon(t *testing.T) {
	render := func(icon *ui.SVG) *image.RGBA {
		on := true
		tt := ui.NewTester(func(c *ui.Context) {
			th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
			ui.Column(c).Padding(16).Children(func() { nui.Checkbox(c, th, &on, nui.CheckboxOptions{Label: "Agree", CheckedIcon: icon}) })
		}, 160, 60)
		return tt.Image()
	}
	plain, custom := render(nil), render(lucide.X())
	differs := false
	for y := 16; y < 36 && !differs; y++ {
		for x := 16; x < 36; x++ {
			if plain.RGBAAt(x, y) != custom.RGBAAt(x, y) {
				differs = true
				break
			}
		}
	}
	if !differs {
		t.Fatal("CheckedIcon did not replace the default check mark")
	}
}

func TestSwitchTrackRingOnlyForHighlightOrError(t *testing.T) {
	on := false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(20).Gap(20).Children(func() {
			nui.Switch(c, th, &on, nui.SwitchOptions{Label: "Plain"})
			nui.Switch(c, th, &on, nui.SwitchOptions{Label: "Lit", Highlight: true})
			nui.Switch(c, th, &on, nui.SwitchOptions{Label: "Broken", Invalid: true})
		})
	}, 300, 200)
	th := nui.Light()
	for label, want := range map[string]ui.Color{"Plain": th.Accented, "Lit": th.Primary, "Broken": th.Error} {
		// The middle of the track's top edge, away from its rounded ends and the thumb.
		r := bounds(t, tt, label)
		if got := tt.Image().RGBAAt(int(r.X)+18, int(r.Y)); got != rgba(want) {
			t.Errorf("%s switch track edge = %v, want %v", label, got, rgba(want))
		}
	}
}

func TestFieldGroupKeepsFormFieldSize(t *testing.T) {
	var a, b string
	var groupSize nui.Size
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(16).Children(func() {
			nui.FormField(c, th, nui.FormFieldOptions{Label: "Range", Size: nui.LG}, func(th nui.Theme) {
				nui.FieldGroup(c, th, nui.FieldGroupOptions{Size: groupSize}, []nui.FieldGroupItem{
					{ID: "a", Build: func(c *ui.Context, th nui.Theme) ui.Element {
						return nui.Input(c, th, &a, nui.InputOptions{Label: "From"})
					}},
					{ID: "b", Build: func(c *ui.Context, th nui.Theme) ui.Element {
						return nui.Input(c, th, &b, nui.InputOptions{Label: "To"})
					}},
				})
			})
		})
	}, 400, 200)
	// The labeled editor sits below its root's vertical padding: 8 for LG, 6 for SM and MD.
	editorY := func() float32 { return bounds(t, tt, "From").Y }
	lg := editorY()
	groupSize = nui.SM
	tt.Frame()
	if sm := editorY(); sm != lg-2 {
		t.Fatalf("explicit group size: SM editor at %v, LG at %v", sm, lg)
	}
	groupSize = nui.SizeDefault
	tt.Frame()
	if got := editorY(); got != lg {
		t.Fatalf("unsized FieldGroup replaced FormField LG: editor at %v, want %v", got, lg)
	}
}

func TestFieldGroupCornersFollowDirection(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(20).Children(func() {
			nui.FieldGroup(c, th, nui.FieldGroupOptions{}, []nui.FieldGroupItem{
				{ID: "a", Build: func(c *ui.Context, th nui.Theme) ui.Element {
					return nui.Button(c, th, "First", nui.ButtonOptions{}).Width(80)
				}},
				{ID: "b", Build: func(c *ui.Context, th nui.Theme) ui.Element {
					return nui.Button(c, th, "Second", nui.ButtonOptions{}).Width(80)
				}},
			})
		})
	}, 400, 100)
	green := rgba(nui.Light().Primary)
	rounded := func(label string) (left, right bool) {
		l, r := topCorners(tt.Image(), bounds(t, tt, label))
		return l != green, r != green
	}
	for _, dir := range []ui.Direction{ui.LeftToRight, ui.RightToLeft} {
		tt.SetDirection(dir)
		tt.Frame()
		firstLeft, firstRight := rounded("First")
		lastLeft, lastRight := rounded("Second")
		// Left to right, the first item is leftmost; right to left, rightmost.
		ltr := dir == ui.LeftToRight
		if firstLeft != ltr || firstRight == ltr || lastLeft == ltr || lastRight != ltr {
			t.Fatalf("%v: rounded corners first=(%v,%v) last=(%v,%v); only the group's outer corners should be rounded", dir, firstLeft, firstRight, lastLeft, lastRight)
		}
	}
}

func TestBadgeAndSelectMenuJoinFieldGroup(t *testing.T) {
	var value []string
	var s nui.SelectMenuState
	var text string
	listSelection := chosenSelection(value)
	s.SetItems([]nui.ListboxItem{{ID: "x", Label: "Xray"}})
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(20).Gap(20).Children(func() {
			nui.FieldGroup(c, th, nui.FieldGroupOptions{Size: nui.LG}, []nui.FieldGroupItem{
				{ID: "badge", Build: func(c *ui.Context, th nui.Theme) ui.Element { return nui.Badge(c, th, "https://", nui.BadgeOptions{}) }},
				{ID: "text", Build: func(c *ui.Context, th nui.Theme) ui.Element {
					return nui.Input(c, th, &text, nui.InputOptions{Label: "Host"}).Width(120)
				}},
			})
			nui.FieldGroup(c, th, nui.FieldGroupOptions{}, []nui.FieldGroupItem{
				{ID: "text", Build: func(c *ui.Context, th nui.Theme) ui.Element {
					return nui.Input(c, th, &text, nui.InputOptions{Label: "Name"}).Width(120)
				}},
				{ID: "menu", Build: func(c *ui.Context, th nui.Theme) ui.Element {
					return nui.SelectMenu(c, th, listSelection, &s, nui.SelectMenuOptions{Label: "Domain"}).Width(160)
				}},
			})
		})
	}, 400, 200)
	// Badge has no accessible label; Find returns its text, one LG line tall.
	label := bounds(t, tt, "https://")
	if label.H != 20 {
		t.Fatalf("badge in an LG group has a %v DIP label line, want LG's 20", label.H)
	}
	green := rgba(nui.Light().Primary)
	im := tt.Image()
	// Group items stretch: the badge's label line is centered in the 36 DIP input height.
	host := bounds(t, tt, "Host")
	badge := ui.Rect{X: label.X - 8, Y: host.Y - 8, W: label.W + 16, H: 36}
	if label.Y != host.Y {
		t.Fatalf("badge label at %v is not aligned with its neighbor's text at %v", label.Y, host.Y)
	}
	if l, r := topCorners(im, badge); l == green || r != green {
		t.Fatalf("first badge corners: start %v end %v; want only the start rounded", l, r)
	}
	if l, r := topCorners(im, bounds(t, tt, "Domain")); l == white || r != white {
		t.Fatalf("last SelectMenu trigger corners: start %v end %v; want only the end rounded", l, r)
	}
}

func TestPopoverContentLeavesFieldGroup(t *testing.T) {
	open := true
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(20).Children(func() {
			nui.FieldGroup(c, th, nui.FieldGroupOptions{}, []nui.FieldGroupItem{
				{ID: "a", Build: func(c *ui.Context, th nui.Theme) ui.Element {
					anchor := nui.Button(c, th, "Open", nui.ButtonOptions{})
					nui.Popover(c, th, anchor, &open, nui.PopoverOptions{}, func(inner nui.Theme) { nui.Button(c, inner, "Inside", nui.ButtonOptions{}).Width(80) })
					return anchor
				}},
				{ID: "b", Build: func(c *ui.Context, th nui.Theme) ui.Element { return nui.Button(c, th, "Next", nui.ButtonOptions{}) }},
			})
		})
	}, 400, 200)
	green := rgba(nui.Light().Primary)
	if l, r := topCorners(tt.Image(), bounds(t, tt, "Inside")); l == green || r == green {
		t.Fatal("a control in a group item's popover was given group corners")
	}
}

func TestRangeSliderBindsFormField(t *testing.T) {
	low, high := 20.0, 80.0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(16).Gap(12).Children(func() {
			nui.Button(c, th, "Before", nui.ButtonOptions{})
			nui.FormField(c, th, nui.FormFieldOptions{Label: "Price", Error: "Too wide"}, func(th nui.Theme) {
				nui.RangeSlider(c, th, &low, &high, nui.RangeSliderOptions{}).Width(240)
			})
		})
	}, 400, 200)
	click(t, tt, "Before")
	label := bounds(t, tt, "Price")
	tt.ClickAt(label.X+8, label.Y+6)
	if !tt.Focused("minimum") {
		t.Fatalf("FormField label did not focus the first thumb; texts %v", tt.Texts())
	}
	tt.Key(0, ui.KeyRight)
	if low != 21 || high != 80 {
		t.Fatalf("keyboard after label focus moved the range to %v-%v", low, high)
	}
}

func TestPinInputShowsFormFieldError(t *testing.T) {
	pin := ""
	var s nui.PinState
	blue := ui.RGB(0, 0, 255)
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(16).Children(func() {
			// Recolor the error text so that only error-colored slots can match.
			nui.FormField(c, th, nui.FormFieldOptions{Label: "Code", Error: "Wrong code", UI: nui.Styles{nui.ErrorSlot: {Foreground: nui.Set(blue)}}}, func(th nui.Theme) {
				nui.PinInput(c, th, &pin, &s, nui.PinInputOptions{Length: 4})
			})
		})
	}, 300, 160)
	if countColor(tt.Image(), ui.Rect{W: 300, H: 160}, rgba(nui.Light().Error)) == 0 {
		t.Fatal("PIN slots in an invalid FormField kept their normal border")
	}
}

func TestBreadcrumbAndFooterColumnsShowItemIcons(t *testing.T) {
	icons := false
	cancel := true
	selected, navigated := 0, 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		var icon *ui.SVG
		if icons {
			icon = lucide.House()
		}
		ui.Column(c).Padding(16).Gap(16).AlignItems(ui.Start).Children(func() {
			nui.Breadcrumb(c, th, []nui.BreadcrumbItem{{ID: "home", Label: "Home", OnSelect: func() {}}, {ID: "here", Label: "Here", Icon: icon}}, nui.BreadcrumbOptions{})
			nui.FooterColumns(c, th, []nui.FooterColumn{{ID: "docs", Title: "Docs", Items: []nui.NavigationItem{{ID: "start", Label: "Start", Icon: icon, OnSelect: func(e *nui.NavigationEvent) {
				selected++
				if cancel {
					e.PreventDefault()
				}
			}}}}}, nui.FooterColumnsOptions{OnNavigate: func(string) { navigated++ }})
		})
	}, 500, 200)
	// Footer items fill their column, so compare the item's pixels instead of its width.
	region := func(r ui.Rect) []color.RGBA {
		var out []color.RGBA
		for y := int(r.Y); y < int(r.Y+r.H); y++ {
			for x := int(r.X); x < int(r.X+r.W); x++ {
				out = append(out, tt.Image().RGBAAt(x, y))
			}
		}
		return out
	}
	here, start := bounds(t, tt, "Here").X, region(bounds(t, tt, "Start"))
	icons = true
	tt.Frame()
	if got := bounds(t, tt, "Here").X; got <= here {
		t.Fatalf("current breadcrumb icon missing: label at %v, was %v", got, here)
	}
	if slices.Equal(region(bounds(t, tt, "Start")), start) {
		t.Fatal("footer item icon missing")
	}
	click(t, tt, "Start")
	if selected != 1 || navigated != 0 {
		t.Fatalf("footer link cancellation: selects=%d navigations=%d", selected, navigated)
	}
	cancel = false
	tt.Key(0, ui.KeyEnter)
	if selected != 2 || navigated != 1 {
		t.Fatalf("footer link activation: selects=%d navigations=%d", selected, navigated)
	}
}

func TestEmptyCardAndBadgeParts(t *testing.T) {
	footer := false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(20).AlignItems(ui.Start).Children(func() {
			o := nui.CardOptions{Title: "Only"}
			if footer {
				o.Footer = func() { ui.Text(c, "Foot") }
			}
			nui.Card(c, th, o, nil).Width(240)
			ui.Text(c, "After card")
			ui.Row(c).Children(func() {
				nui.Badge(c, th, "", nui.BadgeOptions{Leading: lucide.Check()})
				ui.Text(c, "After badge")
			})
		})
	}, 400, 400)
	// Header padding 16 + the title line + 16; no body region or trailing divider.
	header := bounds(t, tt, "Only").H + 32
	if got := bounds(t, tt, "After card").Y - 20; got != header {
		t.Fatalf("card without body is %v DIP tall, want its %v DIP header", got, header)
	}
	// Automatic square padding: 4 + 16 DIP icon + 4, with no label gap.
	if got := bounds(t, tt, "After badge").X - 20; got != 24 {
		t.Fatalf("icon-only badge is %v DIP wide, want 24", got)
	}
	footer = true
	tt.Frame()
	// The header, a single divider, and the footer's padded line.
	want := header + 1 + bounds(t, tt, "Foot").H + 32
	if got := bounds(t, tt, "After card").Y - 20; got != want {
		t.Fatalf("card with header and footer but no body is %v DIP tall, want %v", got, want)
	}
}

func TestCommandPaletteShowsDescriptions(t *testing.T) {
	var s nui.CommandPaletteState
	s.SetIndex(nui.NewCommandIndex([]nui.CommandAction{{ID: "new", Label: "New file", Description: "Create an empty document"}}))
	s.Open = true
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.CommandPalette(c, th, &s, nui.CommandPaletteOptions{})
	}, 700, 500)
	awaitCommandResults(t, tt, func() bool { return tt.HasText("New file") })
	if !tt.HasText("Create an empty document") {
		t.Fatalf("command description missing: %v", tt.Texts())
	}
	// Find returns the labeled row, which must grow to hold both lines.
	row, description := bounds(t, tt, "New file"), bounds(t, tt, "Create an empty document")
	if description.Y < row.Y+row.H/2 || description.Y+description.H > row.Y+row.H {
		t.Fatalf("description %v is not the lower line of row %v", description, row)
	}
}

func TestDirectionalIconsMirrorRightToLeft(t *testing.T) {
	collapsed := false
	var value nui.DateSelection
	var cal nui.CalendarState
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Row(c).Children(func() {
			nui.Sidebar(c, th, &collapsed, nui.SidebarOptions{Title: "Files"}, nil)
			ui.Column(c).Padding(12).Gap(12).Children(func() {
				nui.Breadcrumb(c, th, []nui.BreadcrumbItem{{ID: "a", Label: "Root", OnSelect: func() {}}, {ID: "b", Label: "Leaf"}}, nui.BreadcrumbOptions{})
				nui.Calendar(c, th, &value, &cal, nui.CalendarOptions{Label: "Due"})
			})
		})
	}, 700, 500)
	separator := func() ui.Rect {
		root, leaf := bounds(t, tt, "Root"), bounds(t, tt, "Leaf")
		start, end := root.X+root.W, leaf.X
		if leaf.X < root.X {
			start, end = leaf.X+leaf.W, root.X
		}
		return ui.Rect{X: start, Y: leaf.Y - 2, W: end - start, H: leaf.H + 4}
	}
	pixels := func(r ui.Rect) []color.RGBA {
		var out []color.RGBA
		for y := int(r.Y); y < int(r.Y+r.H); y++ {
			for x := int(r.X); x < int(r.X+r.W); x++ {
				out = append(out, tt.Image().RGBAAt(x, y))
			}
		}
		return out
	}
	sidebar := map[ui.Direction][]color.RGBA{}
	for _, dir := range []ui.Direction{ui.LeftToRight, ui.RightToLeft} {
		tt.SetDirection(dir)
		tt.Frame()
		im := tt.Image()
		forward := dir == ui.LeftToRight
		if chevronPointsRight(t, im, separator()) != forward {
			t.Errorf("%v: breadcrumb separator should point toward the line's end", dir)
		}
		if chevronPointsRight(t, im, bounds(t, tt, "Previous month")) == forward || chevronPointsRight(t, im, bounds(t, tt, "Next month")) != forward {
			t.Errorf("%v: calendar paging chevrons should point toward the earlier and later months", dir)
		}
		sidebar[dir] = pixels(bounds(t, tt, "Toggle sidebar"))
	}
	if slices.Equal(sidebar[ui.LeftToRight], sidebar[ui.RightToLeft]) {
		t.Error("sidebar toggle icon does not mirror right to left")
	}
}

func TestNoticeCloseButtonsKeepContrastOnHover(t *testing.T) {
	open := true
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(16).Gap(16).Children(func() {
			nui.Banner(c, th, &open, nui.BannerOptions{Title: "Release notes", CloseLabel: "Close banner"})
			nui.Alert(c, th, nui.AlertOptions{Title: "Saved", OnClose: func() {}})
		})
	}, 500, 260)
	th := nui.Light()
	elevated := rgba(th.Elevated)
	for _, label := range []string{"Close banner", "Dismiss Saved"} {
		r := bounds(t, tt, label)
		tt.Move(r.X+r.W/2, r.Y+r.H/2)
		tt.Advance(time.Second) // Let the hover color transition finish.
		// The button's corner holds no icon stroke; it shows only the hover surface.
		got := tt.Image().RGBAAt(int(r.X)+2, int(r.Y)+2)
		if got == elevated || got.G < got.R+100 {
			t.Errorf("%s hover surface %v hides the inverted icon on the primary color", label, got)
		}
	}
}

func TestTextInputPlaceholdersUseDimmed(t *testing.T) {
	for _, dark := range []bool{false, true} {
		text := ""
		var number nui.Optional[float64]
		var numberState nui.NumberState
		tt := ui.NewTester(func(c *ui.Context) {
			th := nui.App(c, nui.AppConfig{Mode: map[bool]nui.Mode{false: nui.LightMode, true: nui.DarkMode}[dark]})
			ui.Column(c).Padding(16).Gap(16).Children(func() {
				nui.Input(c, th, &text, nui.InputOptions{Label: "Plain", Placeholder: "Search everything"}).Width(260)
				nui.InputNumber(c, th, &number, &numberState, nui.InputNumberOptions{Label: "Count", Placeholder: "Enter a number"}).Width(260)
			})
		}, 320, 160)
		th := nui.Light()
		if dark {
			th = nui.Dark()
		}
		for _, label := range []string{"Plain", "Count"} {
			r := bounds(t, tt, label)
			if countColor(tt.Image(), r, rgba(th.Dimmed)) == 0 || countColor(tt.Image(), r, rgba(th.Muted)) != 0 {
				t.Errorf("dark=%v %s placeholder is not drawn in Dimmed", dark, label)
			}
		}
	}
}

// Shared recipes aligned with Nuxt UI v4.12: overlay surfaces and motion.

func TestPopoverPaddingAndScaleIn(t *testing.T) {
	open := false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(20).Children(func() {
			anchor := nui.Button(c, th, "Details", nui.ButtonOptions{}).OnClick(func() { open = !open })
			nui.Popover(c, th, anchor, &open, nui.PopoverOptions{Style: func(slot nui.Slot, e ui.Element) {
				if slot == nui.Panel {
					e.Label("Popover panel")
				}
			}}, func(nui.Theme) { ui.Text(c, "Inside") })
		})
	}, 400, 300)
	click(t, tt, "Details")
	opening := bounds(t, tt, "Popover panel")
	tt.Advance(50 * time.Millisecond)
	middle := bounds(t, tt, "Popover panel")
	tt.Advance(time.Second)
	settled := bounds(t, tt, "Popover panel")
	if !(opening.W < middle.W && middle.W < settled.W) {
		t.Fatalf("popover did not scale in: widths %v, %v, %v", opening.W, middle.W, settled.W)
	}
	// Scaling from the anchor side keeps the edge facing the anchor in place.
	if opening.Y != settled.Y || opening.X != settled.X {
		t.Fatalf("scale-in moved the anchored corner: %v -> %v", opening, settled)
	}
	if tt.FramePending() {
		t.Fatal("settled popover keeps requesting frames")
	}
	inside := bounds(t, tt, "Inside")
	if inside.X-settled.X != 16 || inside.Y-settled.Y != 16 {
		t.Fatalf("popover content offset (%v, %v), want 16 DIP padding", inside.X-settled.X, inside.Y-settled.Y)
	}

	click(t, tt, "Details")
	tt.SetPreferences(ui.Preferences{ReduceMotion: true})
	click(t, tt, "Details")
	if got := bounds(t, tt, "Popover panel"); got != settled {
		t.Fatalf("reduced motion still animates: %v, want %v", got, settled)
	}
}

func TestTooltipIsOneTruncatedLine(t *testing.T) {
	open := true
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		a := nui.Button(c, th, "Anchor", nui.ButtonOptions{}).Absolute().Left(150).Top(100)
		nui.Tooltip(c, th, a, "A rather long explanation that cannot fit on one tooltip line", nui.TooltipOptions{Open: &open, MaxWidth: nui.Set(float32(160)), Style: func(slot nui.Slot, e ui.Element) {
			if slot == nui.Panel {
				e.Label("Tip")
			}
		}})
	}, 500, 300)
	tt.Advance(time.Second)
	tip := bounds(t, tt, "Tip")
	if tip.H != 24 || tip.W > 160 {
		t.Fatalf("tooltip box %v, want one 24 DIP line within its 160 DIP maximum", tip)
	}
	if text := bounds(t, tt, "A rather long explanation that cannot fit on one tooltip line"); text.H != 16 || text.X+text.W > tip.X+tip.W-10 {
		t.Fatalf("tooltip text %v in %v is not one truncated line inside the 10 DIP padding", text, tip)
	}
}

func TestModalScalesInOverFadingBackdrop(t *testing.T) {
	open := true
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		nui.ModalWithOptions(c, th, &open, nui.ModalOptions{Title: "Confirm"}, func(nui.Theme) { ui.Text(c, "Body") })
	}, 600, 400)
	// The backdrop is Elevated at 75% over white; it starts transparent.
	corner := func() uint8 { return tt.Image().RGBAAt(2, 2).G }
	first, panel := corner(), bounds(t, tt, "Confirm")
	tt.Advance(time.Second)
	settled := bounds(t, tt, "Confirm")
	if first <= corner() {
		t.Fatalf("backdrop did not fade in: corner green %d then %d", first, corner())
	}
	if panel.W >= settled.W || settled.W != 512 {
		t.Fatalf("modal did not scale in to 512 DIPs: %v -> %v", panel, settled)
	}
}

func TestOptionRowsFollowSizeAndHighlight(t *testing.T) {
	for _, tc := range []struct {
		size      nui.Size
		row, line float32
	}{{nui.XS, 24, 16}, {nui.MD, 32, 20}, {nui.XL, 40, 24}} {
		value := "c"
		selectState := submittedSelect([]nui.Option[string]{{Value: "a", Label: "Alpha", Description: "First letter"}, {Value: "b", Label: "Beta"}, {Value: "c", Label: "Gamma"}})
		tt := ui.NewTester(func(c *ui.Context) {
			th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
			ui.Column(c).Padding(20).Children(func() {

				nui.Select(c, th, &value, selectState, nui.SelectOptions{Label: "Letter", Size: tc.size, Popup: nui.PopoverOptions{Label: "Letter options"}}).Width(240)
			})
		}, 400, 400)
		click(t, tt, "Letter")
		tt.Advance(time.Second)
		// Beta and the description appear only in the popup; Gamma is also the trigger's value.
		beta, description := bounds(t, tt, "Beta"), bounds(t, tt, "First letter")
		if beta.H != tc.row || description.H != tc.line {
			t.Fatalf("size %v: row %v, description line %v; want %v and the row's own %v", tc.size, beta.H, description.H, tc.row, tc.line)
		}
		th := nui.Light()
		im := tt.Image()
		// The selected row, after Beta, is highlighted with elevated/50 and checked in the text color.
		gamma := ui.Rect{X: beta.X, Y: beta.Y + beta.H, W: beta.W, H: beta.H}
		got, want := im.RGBAAt(int(gamma.X)+2, int(gamma.Y+gamma.H/2)), rgba(th.Elevated.Mix(th.Background, .5))
		if d := int(got.B) - int(want.B); d < -1 || d > 1 || got == rgba(th.Elevated) {
			t.Fatalf("size %v: highlighted row surface %v, want elevated/50 %v", tc.size, got, want)
		}
		if n := countColor(im, bounds(t, tt, "Letter options"), rgba(th.Primary)); n != 0 {
			t.Fatalf("size %v: the selected check uses the primary color (%d pixels)", tc.size, n)
		}
	}
}

func TestListboxRowsRadiusAndEmptyState(t *testing.T) {
	var selected []string
	var s, empty nui.ListboxState
	var menu nui.SelectMenuState
	listSelection := chosenSelection(selected)
	s.SetItems([]nui.ListboxItem{{ID: "a", Label: "Alpha"}})
	empty.SetItems(nil)
	menu.SetItems([]nui.ListboxItem{{ID: "x", Label: "Xray"}})
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(20).Gap(12).Children(func() {
			nui.Listbox(c, th, listSelection, &s, nui.ListboxOptions{Label: "Letters", Height: 120}).Width(240)
			nui.Listbox(c, th, listSelection, &empty, nui.ListboxOptions{Label: "Nothing", EmptyText: "Nothing here", Height: 120}).Width(240)
			nui.SelectMenu(c, th, listSelection, &menu, nui.SelectMenuOptions{Label: "Small menu", Size: nui.XS}).Width(240)
		})
	}, 400, 600)
	if h := bounds(t, tt, "Alpha").H; h != 32 {
		t.Fatalf("listbox row is %v DIPs tall, want Nuxt's 32", h)
	}
	text, root := bounds(t, tt, "Nothing here"), bounds(t, tt, "Nothing")
	if d := (text.X + text.W/2) - (root.X + root.W/2); d < -1 || d > 1 {
		t.Fatalf("empty state %v is not centered in %v", text, root)
	}
	click(t, tt, "Small menu")
	tt.Advance(time.Second)
	if h := bounds(t, tt, "Xray").H; h != 24 {
		t.Fatalf("an XS SelectMenu lists %v DIP rows, want 24", h)
	}
}

func TestMenuRowsUseOptionRecipe(t *testing.T) {
	var state nui.MenuState
	state.SetItems([]nui.MenuItem{{ID: "edit", Label: "Edit", Description: "Change the title"}, {ID: "delete", Label: "Delete", Color: nui.Error}})
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(20).Children(func() {
			anchor := nui.Button(c, th, "Actions", nui.ButtonOptions{})
			nui.DropdownMenu(c, th, anchor, &state, nui.MenuOptions{})
		})
	}, 400, 400)
	click(t, tt, "Actions")
	tt.Advance(time.Second)
	th := nui.Light()
	edit, del := bounds(t, tt, "Edit"), bounds(t, tt, "Delete")
	if del.H != 32 || edit.H != 52 {
		t.Fatalf("menu rows %v and %v; want 32, and 52 with a same-size description", del, edit)
	}
	// The first row has focus: elevated/50, not the opaque elevated surface.
	if got := tt.Image().RGBAAt(int(edit.X)+2, int(edit.Y+edit.H/2)); got == rgba(th.Elevated) || got == white {
		t.Fatalf("highlighted menu row surface %v", got)
	}
	tt.Key(0, ui.KeyDown)
	tt.Frame()
	tt.Advance(200 * time.Millisecond) // the colour transition
	// Error at 10% over white: a light red, not elevated/50.
	if got := tt.Image().RGBAAt(int(del.X)+2, int(del.Y+del.H/2)); got == white || got.R <= got.G || got.R <= got.B {
		t.Fatalf("highlighted destructive row %v is not tinted with its color", got)
	}
}

func TestSmallRadiusTokens(t *testing.T) {
	same := func(t *testing.T, name string, build func(c *ui.Context, th nui.Theme, explicit bool)) {
		t.Helper()
		render := func(explicit bool) []byte {
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
				ui.Column(c).Padding(10).Children(func() { build(c, th, explicit) })
			}, 200, 80)
			return tt.Image().Pix
		}
		if !slices.Equal(render(false), render(true)) {
			t.Errorf("%s default radius is not Nuxt's rounded-sm (4)", name)
		}
	}
	four := nui.Set(float32(4))
	same(t, "Checkbox", func(c *ui.Context, th nui.Theme, explicit bool) {
		on := true
		o := nui.CheckboxOptions{Label: "Box"}
		if explicit {
			o.UI = nui.Styles{nui.Indicator: {Radius: four}}
		}
		nui.Checkbox(c, th, &on, o)
	})
	for _, size := range []nui.Size{nui.XS, nui.SM} {
		same(t, "Badge", func(c *ui.Context, th nui.Theme, explicit bool) {
			o := nui.BadgeOptions{Size: size}
			if explicit {
				o.UI = nui.Styles{nui.Root: {Radius: four}}
			}
			nui.Badge(c, th, "New", o)
		})
	}
	same(t, "Kbd", func(c *ui.Context, th nui.Theme, explicit bool) {
		o := nui.KbdOptions{}
		if explicit {
			o.UI = nui.Styles{nui.Label: {Radius: four}}
		}
		nui.Kbd(c, th, []string{"K"}, o)
	})
}

func TestTitlesUseNuxtsPageSize(t *testing.T) {
	open := true
	collapsed := false
	drawer := nui.DrawerState{Open: true}
	// Containers named after their title come first; rename them or the heading itself.
	rename := func(slot nui.Slot, name string) func(nui.Slot, ui.Element) {
		return func(s nui.Slot, e ui.Element) {
			if s == slot {
				e.Label(name)
			}
		}
	}
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Row(c).Children(func() {
			nui.Sidebar(c, th, &collapsed, nui.SidebarOptions{Title: "Side title", Style: rename(nui.Root, "Sidebar region")}, nil)
			ui.Column(c).Padding(12).Children(func() {
				nui.Card(c, th, nui.CardOptions{Title: "Card title"}, nil).Width(240)
			})
		})
		nui.ModalWithOptions(c, th, &open, nui.ModalOptions{Title: "Modal title", Style: rename(nui.Label, "Modal heading")}, nil)
	}, 900, 600)
	tt.Advance(time.Second)
	for _, title := range []string{"Side title", "Card title", "Modal heading"} {
		if h := bounds(t, tt, title).H; h != 24 {
			t.Errorf("%s line is %v DIPs, want the 16/24 page size", title, h)
		}
	}
	open = false
	tt2 := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Drawer(c, th, &drawer, nui.DrawerOptions{Title: "Drawer title", Style: rename(nui.Label, "Drawer heading")}, nil)
	}, 600, 600)
	tt2.Advance(time.Second)
	if h := bounds(t, tt2, "Drawer heading").H; h != 24 {
		t.Errorf("Drawer title line is %v DIPs, want 24", h)
	}
}

func TestSelectMultipleDrawsOneIcon(t *testing.T) {
	var value []string
	selectState := submittedSelect([]nui.Option[string]{{Value: "a", Label: "Alpha", Description: "First", Icon: lucide.Star()}})
	selectedState := chosenSelection(value)
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(20).Children(func() {
			nui.SelectMultiple(c, th, selectedState, selectState, nui.SelectOptions{Label: "Letters"}).Width(240)
		})
	}, 400, 300)
	click(t, tt, "Letters")
	tt.Advance(time.Second)
	// Padding 6, one 20 DIP icon and a 6 DIP gap precede the texts.
	if offset := bounds(t, tt, "First").X - bounds(t, tt, "Alpha").X; offset != 32 {
		t.Fatalf("option text starts %v DIPs into its row, want 32 after a single icon", offset)
	}
}

// Selected rows show only their trailing check: MyGo's Selected paints the
// accent color, which the option recipe must replace when the row is not
// highlighted, as in Nuxt, where only highlighted rows have a surface.
func TestSelectedOptionRowsHaveNoAccentSurface(t *testing.T) {
	members := []nui.ListboxItem{{ID: "lin", Label: "Lin Chen", Group: "Design"}, {ID: "ada", Label: "Ada Lovelace", Group: "Engineering"}, {ID: "grace", Label: "Grace Hopper", Group: "Engineering"}, {ID: "guest", Label: "Guest account", Group: "External", Disabled: true}}
	th := nui.Light()
	// surface samples a row just inside its start edge, away from text and the check.
	surface := func(tt *ui.Tester, label string) color.RGBA {
		r := bounds(t, tt, label)
		return tt.Image().RGBAAt(int(r.X)+2, int(r.Y+r.H/2))
	}
	accent := func(c color.RGBA) bool { return near(c, rgba(th.Primary), 24) }

	t.Run("listbox", func(t *testing.T) {
		var chosen []string
		var s nui.ListboxState
		listSelection := chosenSelection(chosen)
		s.SetItems(members)
		tt := ui.NewTester(func(c *ui.Context) {
			th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
			ui.Column(c).Padding(20).Children(func() {
				nui.Listbox(c, th, listSelection, &s, nui.ListboxOptions{Label: "Members list", Multiple: true, Height: 230}).Width(300)
			})
		}, 400, 320)
		click(t, tt, "Lin Chen")
		click(t, tt, "Grace Hopper")
		ada := bounds(t, tt, "Ada Lovelace")
		tt.Move(ada.X+ada.W/2, ada.Y+ada.H/2)
		tt.Frame()
		tt.Advance(200 * time.Millisecond) // the colour transition
		parityScreenshot(t, tt, "listbox-selected-rows")
		if !slices.Equal(listSelection.Values(), []string{"lin", "grace"}) {
			t.Fatalf("chosen %v", listSelection.Values())
		}
		for _, label := range []string{"Lin Chen", "Grace Hopper"} {
			if c := surface(tt, label); accent(c) {
				t.Errorf("selected, unhighlighted %s row is filled with the accent color: %v", label, c)
			}
		}
		if c := surface(tt, "Ada Lovelace"); c == white || accent(c) {
			t.Errorf("the hovered row lost its elevated/50 highlight: %v", c)
		}
	})

	t.Run("select", func(t *testing.T) {
		value := "a"
		selectState := submittedSelect([]nui.Option[string]{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}})
		tt := ui.NewTester(func(c *ui.Context) {
			th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
			ui.Column(c).Padding(20).Children(func() {
				nui.Select(c, th, &value, selectState, nui.SelectOptions{Label: "Letter"}).Width(240)
			})
		}, 400, 300)
		click(t, tt, "Letter")
		tt.Advance(time.Second)
		beta := bounds(t, tt, "Beta")
		tt.Move(beta.X+beta.W/2, beta.Y+beta.H/2)
		tt.Frame()
		// The trigger also shows "Alpha"; the popup row is the second match.
		alpha := ui.Rect{X: beta.X, Y: beta.Y - beta.H, W: beta.W, H: beta.H}
		if c := tt.Image().RGBAAt(int(alpha.X)+2, int(alpha.Y+alpha.H/2)); accent(c) {
			t.Errorf("the selected Select option is filled with the accent color once unhighlighted: %v", c)
		}
	})

	t.Run("input-menu", func(t *testing.T) {
		state := nui.SearchState{}
		state.SelectedValues.Set([]string{"a", "b"})
		state.SetOptions([]nui.Option[string]{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}, {Value: "c", Label: "Gamma"}})
		tt := ui.NewTester(func(c *ui.Context) {
			th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
			ui.Column(c).Padding(20).Children(func() {
				nui.InputMenu(c, th, &state, nui.InputMenuOptions{Label: "Letters", Multiple: true}).Width(300)
			})
		}, 400, 400)
		click(t, tt, "Show options")
		tt.Advance(time.Second)
		gamma := bounds(t, tt, "Gamma")
		tt.Move(gamma.X+gamma.W/2, gamma.Y+gamma.H/2)
		tt.Frame()
		// Chips repeat the selected labels; the popup rows sit above Gamma.
		for i, label := range []string{"Alpha", "Beta"} {
			y := gamma.Y - float32(2-i)*gamma.H + gamma.H/2
			if c := tt.Image().RGBAAt(int(gamma.X)+2, int(y)); accent(c) {
				t.Errorf("selected %s option is filled with the accent color: %v", label, c)
			}
		}
	})
}

// InputRating behavior aligned with Nuxt UI v4.12 (Reka UI 2.11 Rating).

// labelStars names each star "Star i" for lookups; the counter restarts every build.
func labelStars() func(nui.Slot, ui.Element) {
	star := 0
	return func(slot nui.Slot, e ui.Element) {
		switch slot {
		case nui.Root:
			star = 0
		case nui.Item:
			star++
			e.Label(fmt.Sprintf("Star %d", star))
		}
	}
}

func TestInputRatingRestoringMotionKeepsCurrentScale(t *testing.T) {
	value, reduce := 0, false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		th.Motion.Reduce = reduce
		ui.Column(c).Padding(20).Children(func() {
			nui.InputRating(c, th, &value, nui.InputRatingOptions{Label: "Stars", Style: labelStars()})
		})
	}, 300, 100)
	star := bounds(t, tt, "Star 1")
	reduce = true
	tt.Frame()
	tt.Move(star.X+star.W/2, star.Y+star.H/2)
	before := bounds(t, tt, "Star 1")
	if before.W <= star.W {
		t.Fatal("reduced motion lost the hover scale")
	}
	reduce = false
	tt.Frame()
	if after := bounds(t, tt, "Star 1"); after.W != before.W {
		t.Fatalf("restoring motion reverted the hovered star: %+v -> %+v", before, after)
	}
}

func TestInputRatingKeyboardMatchesNuxt(t *testing.T) {
	for _, tc := range []struct {
		name                string
		start               int
		clearable, readOnly bool
		keys                []ui.Key
		want, changes       int
	}{
		{"right from empty rates one", 0, false, false, []ui.Key{ui.KeyRight}, 1, 1},
		{"left from empty stays empty", 0, false, false, []ui.Key{ui.KeyLeft}, 0, 0},
		{"left at one stops", 1, false, false, []ui.Key{ui.KeyLeft}, 1, 0},
		{"right at max stops", 5, false, false, []ui.Key{ui.KeyRight}, 5, 0},
		{"space rates an empty rating one", 0, false, false, []ui.Key{ui.KeySpace}, 1, 1},
		{"space keeps the rating without clearable", 3, false, false, []ui.Key{ui.KeySpace}, 3, 0},
		{"space clears the current star", 3, true, false, []ui.Key{ui.KeySpace}, 0, 1},
		// Nuxt's focus stays on the cleared star; Go's current star of an
		// empty rating is the first, so Space then rates it 1.
		{"space after clearing rates one", 3, true, false, []ui.Key{ui.KeySpace, ui.KeySpace}, 1, 2},
		{"enter does nothing", 3, true, false, []ui.Key{ui.KeyEnter}, 3, 0},
		{"read-only ignores space", 3, true, true, []ui.Key{ui.KeySpace}, 3, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			value, changes := tc.start, 0
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				ui.Column(c).Padding(10).Children(func() {
					nui.Button(c, th, "Before", nui.ButtonOptions{})
					nui.InputRating(c, th, &value, nui.InputRatingOptions{Label: "Stars", Clearable: tc.clearable, ReadOnly: tc.readOnly, OnChange: func(int) { changes++ }})
				})
			}, 300, 120)
			click(t, tt, "Before")
			tt.Key(0, ui.KeyTab)
			if !tt.Focused("Stars") {
				t.Fatal("the rating did not take focus")
			}
			for _, k := range tc.keys {
				tt.Key(0, k)
			}
			if value != tc.want || changes != tc.changes {
				t.Fatalf("value=%d changes=%d, want %d and %d", value, changes, tc.want, tc.changes)
			}
		})
	}
}

func TestInputRatingRingsCurrentStarAndHugsStars(t *testing.T) {
	value := 3
	var orientation nui.Orientation
	var tone nui.Color
	var parts nui.Styles
	highContrast := false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		th.HighContrast = highContrast
		ui.Column(c).Padding(20).Gap(10).Children(func() {
			nui.Button(c, th, "Before", nui.ButtonOptions{})
			nui.InputRating(c, th, &value, nui.InputRatingOptions{Label: "Stars", Orientation: orientation, Color: tone, UI: parts, Style: labelStars()})
		})
	}, 300, 260)
	// Five 20 DIP stars 2 apart, instead of the column's 260 DIP width.
	if r := bounds(t, tt, "Stars"); r.W != 108 || r.H != 20 {
		t.Fatalf("horizontal rating %v, want 108x20", r)
	}
	click(t, tt, "Before")
	tt.Key(0, ui.KeyTab)
	tt.Frame()
	ring := func(label string) bool {
		r := bounds(t, tt, label)
		c := tt.Image().RGBAAt(int(r.X+r.W/2), int(r.Y)-2)
		return c != white
	}
	if !ring("Star 3") || ring("Star 1") || ring("Star 5") {
		t.Fatal("the focus ring is not on the current star alone")
	}
	root := bounds(t, tt, "Stars")
	if c := tt.Image().RGBAAt(int(root.X+root.W)+3, int(root.Y+root.H/2)); c != white {
		t.Fatalf("a ring is still drawn around the whole rating: %v", c)
	}
	value = 0
	tt.Frame()
	if !ring("Star 1") || ring("Star 3") {
		t.Fatal("an empty rating should ring its first star")
	}
	orientation = nui.Vertical
	tt.Frame()
	if r := bounds(t, tt, "Stars"); r.W != 20 || r.H != 108 {
		t.Fatalf("vertical rating %v, want 20x108", r)
	}
	orientation, tone, highContrast, value = nui.Horizontal, nui.Neutral, true, 3
	tt.Frame()
	star := bounds(t, tt, "Star 3")
	if got := tt.Image().RGBAAt(int(star.X+star.W/2), int(star.Y)-2); got != rgba(nui.Light().InvertedBackground) {
		t.Fatalf("neutral focus ring = %v, want inverted background", got)
	}
	red, blue := ui.RGB(255, 0, 0), ui.RGB(0, 0, 255)
	parts = nui.Styles{nui.Indicator: {Foreground: nui.Set(red)}, nui.EmptyIconSlot: {Foreground: nui.Set(blue)}}
	tt.Frame()
	if countColor(tt.Image(), bounds(t, tt, "Star 3"), rgba(red)) == 0 || countColor(tt.Image(), bounds(t, tt, "Star 5"), rgba(blue)) == 0 {
		t.Fatal("rating layers ignored their foreground styles")
	}
	parts[nui.IconSlot] = nui.PartStyle{Foreground: nui.Set(blue)}
	tt.Frame()
	if countColor(tt.Image(), bounds(t, tt, "Star 3"), rgba(red)) != 0 {
		t.Fatal("icon foreground did not override the filled layer")
	}
	tt.SetDirection(ui.RightToLeft)
	tt.Key(0, ui.KeyLeft)
	if value != 4 {
		t.Fatal("RTL left arrow did not increase the rating")
	}
	tt.Key(0, ui.KeyRight)
	if value != 3 {
		t.Fatal("RTL right arrow did not decrease the rating")
	}
}

func TestInputRatingHoverPreview(t *testing.T) {
	value := 3
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(20).Children(func() {
			// A 10 DIP gap leaves room between the stars for the pointer.
			nui.InputRating(c, th, &value, nui.InputRatingOptions{Label: "Stars", Clearable: true, Hoverable: nui.Set(true), Style: labelStars(), UI: nui.Styles{nui.Root: {Gap: nui.Set(float32(10))}}})
		})
	}, 300, 120)
	// A filled star is green (primary), an empty one slate (muted). Judge by
	// hue: the hovered star is scaled, so its edges blend with the background.
	green := func(c color.RGBA) bool { return int(c.G) > int(c.R)+40 && int(c.G) > int(c.B)+20 }
	filled := func(i int) bool {
		return countWhere(tt.Image(), bounds(t, tt, fmt.Sprintf("Star %d", i)), green) > 10
	}
	moveTo := func(x, y float32) {
		tt.Move(x, y)
		tt.Advance(time.Second)
	}
	center := func(label string) (float32, float32) {
		r := bounds(t, tt, label)
		return r.X + r.W/2, r.Y + r.H/2
	}
	moveTo(center("Star 4"))
	if !filled(4) || filled(5) {
		t.Fatal("hovering star 4 does not preview four stars")
	}
	// Between stars 4 and 5 the preview stays, without falling back to the value.
	four := bounds(t, tt, "Star 4")
	moveTo(four.X+four.W+5, four.Y+four.H/2)
	if !filled(4) {
		t.Fatal("the preview flickered back while crossing the gap")
	}
	moveTo(150, 100)
	if filled(4) || !filled(3) {
		t.Fatal("leaving the rating did not restore the value")
	}
	// Clearing the current star shows an empty rating while the pointer stays.
	x, y := center("Star 3")
	moveTo(x, y)
	tt.ClickAt(x, y)
	tt.Advance(time.Second)
	if value != 0 || filled(1) || filled(3) {
		t.Fatalf("after clearing: value %d, star 1 filled %v, star 3 filled %v", value, filled(1), filled(3))
	}
	moveTo(center("Star 2"))
	if !filled(2) || filled(3) {
		t.Fatal("entering another star after clearing does not preview it")
	}
}

func TestInputRatingStarsGrowOnHover(t *testing.T) {
	value := 0
	readOnly := false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(20).Children(func() {
			nui.InputRating(c, th, &value, nui.InputRatingOptions{Label: "Stars", ReadOnly: readOnly, Style: labelStars(), UI: nui.Styles{nui.Root: {Gap: nui.Set(float32(10))}}})
		})
	}, 300, 120)
	// extent is the width of the star's drawing, inked pixels included.
	extent := func() int {
		r := bounds(t, tt, "Star 2")
		lo, hi := 1<<30, -1
		for y := int(r.Y) - 3; y < int(r.Y+r.H)+3; y++ {
			for x := int(r.X) - 3; x < int(r.X+r.W)+3; x++ {
				if tt.Image().RGBAAt(x, y) != white {
					lo, hi = min(lo, x), max(hi, x)
				}
			}
		}
		return hi - lo
	}
	rest := extent()
	r := bounds(t, tt, "Star 2")
	tt.Move(r.X+r.W/2, r.Y+r.H/2)
	tt.Advance(time.Second)
	if grown := extent(); grown <= rest {
		t.Fatalf("hovered star spans %d DIPs, unhovered %d; want it scaled up", grown, rest)
	}
	readOnly = true
	tt.Frame()
	tt.Advance(time.Second)
	if got := extent(); got != rest {
		t.Fatalf("a read-only star grows on hover: %d, want %d", got, rest)
	}
	if got := tt.Cursor(); got != ui.CursorDefault {
		t.Fatalf("read-only rating cursor = %v, want default", got)
	}
}

func TestInputRatingFieldErrorColor(t *testing.T) {
	value := 3
	blue := ui.RGB(0, 0, 255)
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(20).Children(func() {
			// Recolor the error text so that only the stars can use the error color.
			nui.FormField(c, th, nui.FormFieldOptions{Label: "Score", Error: "Too low", UI: nui.Styles{nui.ErrorSlot: {Foreground: nui.Set(blue)}}}, func(th nui.Theme) {
				nui.InputRating(c, th, &value, nui.InputRatingOptions{})
			})
		})
	}, 300, 200)
	th := nui.Light()
	all := ui.Rect{W: 300, H: 200}
	if countColor(tt.Image(), all, rgba(th.Error)) == 0 || countColor(tt.Image(), all, rgba(th.Primary)) != 0 {
		t.Fatal("the rated stars of an invalid field are not drawn in the error color")
	}
}
