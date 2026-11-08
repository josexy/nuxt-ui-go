package nui_test

import (
	"context"
	"image/color"
	"os"
	"os/exec"
	"slices"
	"testing"
	"testing/fstest"
	"time"

	"github.com/egoist/mygo/i18n"
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
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
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Listbox(c, th, &none, &s, nil, nui.ListboxOptions{Label: "Empty"})
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
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Listbox(c, th, &none, &s, nil, nui.ListboxOptions{Label: "Empty"})
	}, 300, 300)
	tt.SetLocale("en")
	if !tt.HasText("Nothing matched") || tt.HasText("No data") {
		t.Fatalf("the application's nuxtui.noResults override was ignored: %v", tt.Texts())
	}
}
