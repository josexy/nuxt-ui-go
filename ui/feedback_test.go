package nui_test

import (
	"bytes"
	"fmt"
	"image/color"
	"math"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func TestAlertChangesPreserveActionEditor(t *testing.T) {
	showIcon, horizontal, showTitle := false, false, true
	value, changes := "Persistent draft", 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		opts := nui.AlertOptions{Description: "Description", Actions: func() {
			nui.Input(c.Key("draft"), th, &value, nui.InputOptions{Label: "Alert draft", OnChange: func(string) { changes++ }}).Width(150)
		}}
		if showIcon {
			opts.Icon = lucide.Info()
		}
		if showTitle {
			opts.Title = "Title"
		}
		if horizontal {
			opts.Orientation = nui.Set(nui.Horizontal)
		}
		nui.Alert(c.Key("alert"), th, opts)
	}, 500, 220)
	click(t, tt, "Alert draft")
	tt.Key(ui.Ctrl, ui.KeyA)
	for _, field := range []*bool{&showIcon, &horizontal, &showTitle, &horizontal, &showIcon} {
		*field = !*field
		tt.Frame()
		if !tt.Focused("Alert draft") {
			t.Fatalf("changing alert layout lost action focus: icon=%v horizontal=%v title=%v", showIcon, horizontal, showTitle)
		}
	}
	tt.Type("Replacement")
	if value != "Replacement" || changes != 1 {
		t.Fatalf("layout lost selection or repeated notification: value=%q changes=%d", value, changes)
	}
}

func TestAlertEmptyTextActionsKeepTopMargin(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Alert(c, th, nui.AlertOptions{Actions: func() { ui.Text(c, "Action content") }, OnClose: func() {}}).Label("Alert")
	}, 320, 150)
	if got := bounds(t, tt, "Action content").Y - bounds(t, tt, "Alert").Y; got != 26 {
		t.Fatalf("empty title/description action top=%g, want padding 16 + margin 10", got)
	}
}

func TestAlertActionsPartStyles(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Alert(c, th, nui.AlertOptions{Title: "Alert", Actions: func() {
			ui.Box(c).Size(10, 10).Label("First")
			ui.Box(c).Size(10, 10).Label("Second")
		}, UI: nui.Styles{nui.Actions: {Gap: nui.Set(float32(17))}}})
	}, 320, 150)
	a, b := bounds(t, tt, "First"), bounds(t, tt, "Second")
	if b.X-a.X-a.W != 17 {
		t.Fatalf("Actions.Gap did not apply: first=%+v second=%+v", a, b)
	}
}

func TestAlertAvatarPriorityAndSize(t *testing.T) {
	avatar := nui.AvatarOptions{Alt: "Ada Lovelace"}
	show := true
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		opts := nui.AlertOptions{Title: "Alert", Icon: lucide.Info(), Style: func(slot nui.Slot, e ui.Element) {
			if slot == nui.Leading {
				e.Label("Leading")
			}
		}}
		if show {
			opts.Avatar = &avatar
		}
		nui.Alert(c, th, opts)
	}, 320, 150)
	if r := bounds(t, tt, "Leading"); r.W != 44 || r.H != 44 || !tt.HasText("AL") {
		t.Fatalf("default avatar did not replace icon at 44 DIP: %+v", r)
	}
	avatar.Size = nui.SM
	tt.Frame()
	if r := bounds(t, tt, "Leading"); r.W != 28 || r.H != 28 {
		t.Fatalf("explicit avatar size ignored: %+v", r)
	}
	show = false
	tt.Frame()
	if r := bounds(t, tt, "Leading"); r.W != 20 || r.H != 20 || tt.HasText("AL") {
		t.Fatalf("cleared avatar did not restore icon: %+v", r)
	}
}

func TestAlertCloseConfigurationAndNotifications(t *testing.T) {
	count := 0
	custom := nui.ButtonOptions{AccessibleLabel: "Custom close", Size: nui.XL, Color: nui.Error, Variant: nui.Outline}
	custom.Style = func(_, _, leading, _ ui.Element) { leading.Label("Close icon") }
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Alert(c, th, nui.AlertOptions{Title: "Alert", Close: &custom, CloseIcon: lucide.ArrowRight(), OnClose: func() { count++ }})
	}, 320, 150)
	iconWidth := bounds(t, tt, "Close icon").W
	if iconWidth != 24 || custom.UI != nil || custom.Leading != nil {
		t.Fatalf("close options were not applied or were mutated: icon=%g options=%+v", iconWidth, custom)
	}
	click(t, tt, "Custom close")
	tt.Key(0, ui.KeyEnter)
	tt.Key(0, ui.KeySpace)
	if count != 3 || !tt.HasText("Alert") {
		t.Fatalf("close must notify once per activation and leave visibility to caller: count=%d", count)
	}
	custom.Disabled = true
	tt.Frame()
	click(t, tt, "Custom close")
	tt.Key(0, ui.KeyEnter)
	if count != 3 {
		t.Fatal("disabled close notified")
	}
	custom.Disabled, custom.Loading = false, true
	tt.Frame()
	click(t, tt, "Custom close")
	if count != 3 {
		t.Fatal("loading close notified")
	}
}

func TestAlertCloseIconChangesRenderedOutput(t *testing.T) {
	icon := lucide.X()
	var leading *ui.SVG
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Alert(c, th, nui.AlertOptions{Title: "Alert", CloseIcon: icon, Close: &nui.ButtonOptions{Leading: leading}})
	}, 320, 100)
	before := bytes.Clone(tt.Image().Pix)
	icon = lucide.ArrowRight()
	tt.Frame()
	if bytes.Equal(before, tt.Image().Pix) {
		t.Fatal("CloseIcon did not change the rendered icon")
	}
	leading = lucide.Info()
	tt.Frame()
	before = bytes.Clone(tt.Image().Pix)
	icon = lucide.X()
	tt.Frame()
	if !bytes.Equal(before, tt.Image().Pix) {
		t.Fatal("CloseIcon replaced explicit Close.Leading")
	}
}

func TestAlertNarrowLayoutAndDirection(t *testing.T) {
	for _, direction := range []ui.Direction{ui.LeftToRight, ui.RightToLeft} {
		for _, orientation := range []nui.Orientation{nui.Vertical, nui.Horizontal} {
			t.Run(fmt.Sprintf("%v/%v", direction, orientation), func(t *testing.T) {
				tt := ui.NewTester(func(c *ui.Context) {
					th := nui.App(c, nui.AppConfig{})
					nui.Alert(c, th, nui.AlertOptions{Title: "需要复查的项目", Description: "这里是一段较长的中文说明，用于验证提示组件在窄窗口中能够正常换行。", Icon: lucide.Info(), Orientation: nui.Set(orientation), OnClose: func() {}, Actions: func() {
						nui.Button(c.Key("review"), th, "Review", nui.ButtonOptions{Size: nui.XS})
						nui.Button(c.Key("later"), th, "Later", nui.ButtonOptions{Size: nui.XS})
					}, Style: func(slot nui.Slot, e ui.Element) {
						if slot == nui.Body || slot == nui.Actions || slot == nui.Leading || slot == nui.Trailing {
							e.Label(string(slot))
						}
					}}).Label("Alert")
				}, 320, 500)
				tt.SetDirection(direction)
				outer := bounds(t, tt, "Alert")
				for _, label := range []string{"body", "actions", "leading", "trailing", "Review", "Later"} {
					r := bounds(t, tt, label)
					if r.X < outer.X+16 || r.X+r.W > outer.X+outer.W-16 || r.Y < outer.Y+16 || r.Y+r.H > outer.Y+outer.H-16 {
						t.Fatalf("%s outside padded alert: inner=%+v outer=%+v", label, r, outer)
					}
				}
				if orientation == nui.Horizontal {
					a, close := bounds(t, tt, "actions"), bounds(t, tt, "trailing")
					gap := close.X - a.X - a.W
					if direction == ui.RightToLeft {
						gap = a.X - close.X - close.W
					}
					if gap != 6 {
						t.Fatalf("actions/close gap=%g, want 6", gap)
					}
				}
			})
		}
	}
}

func TestEmptyResponsivePadding(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		nui.Empty(c, nui.App(c, nui.AppConfig{}), nui.EmptyOptions{Title: "Empty title"}).Label("Empty root")
	}, 500, 200)
	for _, check := range []struct {
		width   int
		padding float32
	}{{500, 16}, {800, 24}, {1100, 32}} {
		tt.SetSize(check.width, 200)
		root, title := bounds(t, tt, "Empty root"), bounds(t, tt, "Empty title")
		if got := title.Y - root.Y; got != check.padding {
			t.Fatalf("width %d: top padding %g, want %g", check.width, got, check.padding)
		}
	}
}

func TestEmptyAvatarSizeAndHeaderOverrides(t *testing.T) {
	avatar := nui.AvatarOptions{Text: "AU", Alt: "Empty avatar"}
	header, leading, body := false, false, false
	clicks := 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		opts := nui.EmptyOptions{Title: "Default title", Avatar: &avatar, Size: nui.XL,
			Actions: func() { nui.Button(c, th, "Default action", nui.ButtonOptions{}).OnClick(func() { clicks++ }) },
			Footer:  func() { ui.Text(c, "Custom footer") },
		}
		if header {
			opts.Header = func() { ui.Text(c, "Custom header") }
		}
		if leading {
			opts.Leading = func() { ui.Text(c, "Custom leading") }
		}
		if body {
			opts.Body = func() { nui.Button(c, th, "Custom action", nui.ButtonOptions{}).OnClick(func() { clicks++ }) }
		}
		nui.Empty(c, th, opts)
	}, 500, 360)
	if r := bounds(t, tt, "Empty avatar"); r.W != 48 || r.H != 48 {
		t.Fatalf("XL default avatar: %+v", r)
	}
	avatar.Size = nui.XS
	tt.Frame()
	if r := bounds(t, tt, "Empty avatar"); r.W != 24 || r.H != 24 {
		t.Fatalf("explicit avatar size ignored: %+v", r)
	}
	leading = true
	tt.Frame()
	if !tt.HasText("Custom leading") || !tt.HasText("Default title") || tt.HasText("AU") {
		t.Fatal("leading builder did not replace only the avatar")
	}
	header, body = true, true
	tt.Frame()
	if !tt.HasText("Custom header") || !tt.HasText("Custom footer") || tt.HasText("Custom leading") || tt.HasText("Default title") || tt.HasText("Default action") {
		t.Fatal("header or body builder priority")
	}
	click(t, tt, "Custom action")
	tt.Frame()
	if clicks != 1 {
		t.Fatalf("custom action called %d times", clicks)
	}
}

func TestEmptyHeaderChangeKeepsActionFocus(t *testing.T) {
	showTitle := false
	draft := "Original"
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		opts := nui.EmptyOptions{Actions: func() {
			nui.Input(c.Key("draft"), th, &draft, nui.InputOptions{Label: "Empty action draft"}).Width(180)
		}}
		if showTitle {
			opts.Title = "Empty title"
		}
		nui.Empty(c, th, opts)
	}, 400, 260)
	click(t, tt, "Empty action draft")
	tt.Key(ui.Ctrl, ui.KeyA)
	showTitle = true
	tt.Frame()
	if !tt.Focused("Empty action draft") {
		t.Fatal("header insertion discarded action focus")
	}
	tt.Type("Replacement")
	if draft != "Replacement" {
		t.Fatal("header insertion discarded the action selection")
	}
}

func TestEmptyLoadingColorUsesLeadingUI(t *testing.T) {
	color := ui.Hex("#dc2626")
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		th.Motion.Disabled = true
		nui.Empty(c, th, nui.EmptyOptions{Loading: true, LoadingIcon: lucide.Check(),
			UI: nui.Styles{nui.Leading: {Foreground: nui.Set(color)}},
			Style: func(slot nui.Slot, e ui.Element) {
				if slot == nui.Leading {
					e.Label("Loading glyph")
				}
			}})
	}, 240, 180)
	if countColor(tt.Image(), bounds(t, tt, "Loading glyph"), rgba(color)) == 0 {
		t.Fatal("loading glyph ignored Leading.Foreground")
	}
}

func TestEmptyFooterKeepsFocusWhenActionsAreAdded(t *testing.T) {
	showActions := false
	draft := "Original"
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		opts := nui.EmptyOptions{Footer: func() { nui.Input(c.Key("draft"), th, &draft, nui.InputOptions{Label: "Footer draft"}).Width(180) }}
		if showActions {
			opts.Actions = func() { nui.Button(c, th, "Action", nui.ButtonOptions{}) }
		}
		nui.Empty(c, th, opts)
	}, 400, 260)
	click(t, tt, "Footer draft")
	tt.Key(ui.Ctrl, ui.KeyA)
	showActions = true
	tt.Frame()
	if !tt.Focused("Footer draft") {
		t.Fatal("adding actions discarded footer focus")
	}
	tt.Type("Replacement")
	if draft != "Replacement" {
		t.Fatal("adding actions discarded footer selection")
	}
}

func TestSkeletonStyleRadiusOverridesUI(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(16).AlignItems(ui.Start).Children(func() {
			nui.Skeleton(c, th, nui.SkeletonOptions{Width: nui.Set(float32(40)), Height: nui.Set(float32(40)), Animate: nui.Set(false), UI: nui.Styles{nui.Root: {Radius: nui.Set(float32(20))}}, Style: func(_ nui.Slot, e ui.Element) { e.Radius(0) }})
		})
	}, 100, 100)
	r := bounds(t, tt, "Loading…")
	if got := tt.Image().RGBAAt(int(r.X)+2, int(r.Y)+2); got != rgba(nui.Light().Elevated) {
		t.Fatalf("Style radius did not override UI radius: corner=%v", got)
	}
}

func TestSkeletonExplicitZeroAndTransparentBackground(t *testing.T) {
	w, h := float32(0), float32(16)
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).AlignSelf(ui.Start).AlignItems(ui.Start).Label("Skeleton layout").Children(func() {
			nui.Skeleton(c, th, nui.SkeletonOptions{Width: nui.Set(w), Height: nui.Set(h), Animate: nui.Set(false), UI: nui.Styles{nui.Root: {Background: nui.Set(ui.Transparent)}}})
			ui.Box(c).Size(4, 4).Label("Following marker")
		})
	}, 200, 100)
	// Empty painted rectangles have no hit bounds; inspect their layout contribution.
	if r := bounds(t, tt, "Skeleton layout"); r.W != 4 || r.H != 20 {
		t.Fatalf("explicit zero width was replaced: %+v", r)
	}
	w, h = 80, 0
	tt.Frame()
	if r := bounds(t, tt, "Skeleton layout"); r.W != 80 || r.H != 4 {
		t.Fatalf("explicit zero height was replaced: %+v", r)
	}
	h = 20
	tt.Frame()
	r := bounds(t, tt, "Loading…")
	if got := tt.Image().RGBAAt(int(r.X+r.W/2), int(r.Y+r.H/2)); got != rgba(nui.Light().Background) {
		t.Fatalf("transparent background was not preserved: %v", got)
	}
}

func TestSkeletonPulseChangesPixelsAndStops(t *testing.T) {
	off := false
	tt := ui.NewTester(func(c *ui.Context) {
		nui.Skeleton(c, nui.App(c, nui.AppConfig{Mode: nui.LightMode, DisableAnimations: off}), nui.SkeletonOptions{})
	}, 200, 60)
	first := bytes.Clone(tt.Image().Pix)
	changed := false
	for range 4 {
		tt.Advance(250 * time.Millisecond)
		changed = changed || !bytes.Equal(first, tt.Image().Pix)
	}
	if !changed {
		t.Fatal("pulse requested frames without changing pixels")
	}
	// The pulse moves nothing: it goes on under reduced motion, as Nuxt's
	// animate-pulse does, and stops with animations off.
	tt.SetPreferences(ui.Preferences{ReduceMotion: true})
	before := bytes.Clone(tt.Image().Pix)
	tt.Advance(time.Second)
	if bytes.Equal(before, tt.Image().Pix) || !tt.FramePending() {
		t.Fatal("reduced motion stopped the pulse")
	}
	off = true
	tt.Frame()
	still := bytes.Clone(tt.Image().Pix)
	tt.Advance(time.Second)
	if !bytes.Equal(still, tt.Image().Pix) || tt.FramePending() {
		t.Fatal("disabled animations did not stop the pulse")
	}
}

func TestSkeletonThemesAndParentWidth(t *testing.T) {
	for name, mode := range map[string]nui.Mode{"light": nui.LightMode, "dark": nui.DarkMode} {
		t.Run(name, func(t *testing.T) {
			var th nui.Theme
			tt := ui.NewTester(func(c *ui.Context) {
				th = nui.App(c, nui.AppConfig{Mode: mode})
				ui.Column(c).Width(100).Padding(8).Children(func() {
					nui.Skeleton(c, th, nui.SkeletonOptions{Animate: nui.Set(false)}).FillWidth()
				})
			}, 180, 80)
			tt.SetPreferences(ui.Preferences{TextScale: 1.5})
			r := bounds(t, tt, "Loading…")
			if r.W != 84 || r.H != 24 {
				t.Fatalf("parent width or text scaling: %+v", r)
			}
			if got := tt.Image().RGBAAt(int(r.X+r.W/2), int(r.Y+r.H/2)); got != rgba(th.Elevated) {
				t.Fatalf("theme background=%v want=%v", got, th.Elevated)
			}
		})
	}
}

func TestProgressNeutralUsesInvertedBackground(t *testing.T) {
	color := ui.Hex("#7e22ce")
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		th.InvertedBackground = color
		nui.Progress(c, th, 50, nui.ProgressOptions{Label: "Progress", Color: nui.Neutral, Animate: nui.Set(false)}).Width(200)
	}, 240, 60)
	r := bounds(t, tt, "Progress")
	if got := tt.Image().RGBAAt(int(r.X)+30, int(r.Y)+4); got != rgba(color) {
		t.Fatalf("Neutral progress ignored InvertedBackground: %v", got)
	}
}

func TestProgressRejectsInvalidMaximum(t *testing.T) {
	for _, maximum := range []float64{0, -1, math.NaN(), math.Inf(1), math.Inf(-1)} {
		t.Run(fmt.Sprint(maximum), func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid maximum was silently accepted")
				}
			}()
			ui.NewTester(func(c *ui.Context) {
				nui.Progress(c, nui.App(c, nui.AppConfig{}), 20, nui.ProgressOptions{Max: nui.Set(maximum)})
			}, 240, 60)
		})
	}
}

func TestProgressValueTransitionSettles(t *testing.T) {
	value := float64(25)
	tt := ui.NewTester(func(c *ui.Context) {
		nui.Progress(c, nui.App(c, nui.AppConfig{}), value, nui.ProgressOptions{Label: "Progress"}).Width(200)
	}, 240, 60)
	value = 75
	tt.Frame()
	if !tt.FramePending() {
		t.Fatal("value change did not start a transition")
	}
	tt.Advance(250 * time.Millisecond)
	if tt.FramePending() {
		t.Fatal("completed value transition kept requesting frames")
	}
	r := bounds(t, tt, "Progress")
	if got := tt.Image().RGBAAt(int(r.X)+140, int(r.Y)+4); got != rgba(nui.Light().Primary) {
		t.Fatal("value transition did not reach its target")
	}
}

func TestProgressRestoringAnimationKeepsCurrentValue(t *testing.T) {
	value, animate := float64(25), true
	tt := ui.NewTester(func(c *ui.Context) {
		nui.Progress(c, nui.App(c, nui.AppConfig{Mode: nui.LightMode}), value, nui.ProgressOptions{Label: "Progress", Animate: nui.Set(animate)}).Width(200)
	}, 240, 60)
	value, animate = 75, false
	tt.Frame()
	r := bounds(t, tt, "Progress")
	before := tt.Image().RGBAAt(int(r.X)+140, int(r.Y)+4)
	if before != rgba(nui.Light().Primary) {
		t.Fatal("disabled animation did not show the current value immediately")
	}
	animate = true
	tt.Frame()
	if after := tt.Image().RGBAAt(int(r.X)+140, int(r.Y)+4); after != before {
		t.Fatalf("restoring animation reverted the current progress: %v -> %v", before, after)
	}
}

func TestProgressIndeterminateHalfTrack(t *testing.T) {
	animated, vertical := false, false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		e := nui.Progress(c, th, 0, nui.ProgressOptions{Label: "Progress", Indeterminate: true, Animate: nui.Set(animated), Vertical: vertical})
		if vertical {
			e.Height(200)
		} else {
			e.Width(200)
		}
	}, 240, 240)
	assertHalf := func() {
		t.Helper()
		r := bounds(t, tt, "Progress")
		for _, offset := range []int{60, 140} {
			x, y := int(r.X)+offset, int(r.Y)+4
			if vertical {
				x, y = int(r.X)+4, int(r.Y)+offset
			}
			if got := tt.Image().RGBAAt(x, y); got != rgba(nui.Light().Primary) {
				t.Fatalf("half-track indicator missing at %d: %v", offset, got)
			}
		}
	}
	assertHalf()
	animated = true
	tt.Frame()
	phase := time.Duration(tt.Now().UnixNano() % int64(2*time.Second))
	tt.Advance(3*time.Second - phase)
	tt.Frame()
	assertHalf()
	vertical = true
	tt.Frame()
	assertHalf()
}

func TestProgressGroupNeutralUsesInvertedBackground(t *testing.T) {
	color := ui.Hex("#7e22ce")
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		th.InvertedBackground = color
		nui.ProgressGroup(c, th, []nui.ProgressSegment{{ID: "first", Label: "First", Value: 50, Color: nui.Neutral}}, nui.ProgressGroupOptions{Label: "Group"}).Width(200)
	}, 240, 100)
	r := bounds(t, tt, "Group")
	if got := tt.Image().RGBAAt(int(r.X)+30, int(r.Y)+4); got != rgba(color) {
		t.Fatalf("Neutral segment ignored InvertedBackground: %v", got)
	}
}

func TestProgressGroupIconAndItemStyles(t *testing.T) {
	showIcon := true
	size := nui.XS
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		segment := nui.ProgressSegment{ID: "first", Value: 50, UI: nui.Styles{nui.Segment: {Background: nui.Set(ui.Hex("#dc2626"))}}}
		if showIcon {
			segment.Icon = lucide.Check()
		}
		nui.ProgressGroup(c, th, []nui.ProgressSegment{segment}, nui.ProgressGroupOptions{Label: "Group", Size: size, Style: func(slot nui.Slot, e ui.Element) {
			if slot == nui.Leading {
				e.Label("Legend marker")
			}
		}}).Width(200)
	}, 240, 140)
	if r := bounds(t, tt, "Legend marker"); r.W != 12 || r.H != 12 {
		t.Fatalf("icon-only item did not create an XS legend: %+v", r)
	}
	r := bounds(t, tt, "Group")
	if got := tt.Image().RGBAAt(int(r.X)+30, int(r.Y)+1); got != rgba(ui.Hex("#dc2626")) {
		t.Fatalf("per-segment fill style ignored: %v", got)
	}
	size = nui.XL
	tt.Frame()
	if r := bounds(t, tt, "Legend marker"); r.W != 20 || r.H != 20 {
		t.Fatalf("XL icon size: %+v", r)
	}
	showIcon = false
	tt.Frame()
	if tt.HasText("Legend marker") {
		t.Fatal("empty item left an unnecessary legend")
	}
}

func TestProgressGroupBuilderKeepsFocusAcrossReorderAndOrientation(t *testing.T) {
	draft := "Original"
	items := []nui.ProgressSegment{
		{ID: "draft", Value: 50, Build: func(c *ui.Context, th nui.Theme) {
			nui.Input(c.Key("draft"), th, &draft, nui.InputOptions{Label: "Segment draft"}).Width(140)
		}},
		{ID: "other", Label: "Other", Value: 20},
	}
	orientation := nui.Horizontal
	tt := ui.NewTester(func(c *ui.Context) {
		nui.ProgressGroup(c, nui.App(c, nui.AppConfig{}), items, nui.ProgressGroupOptions{Orientation: orientation}).Width(240)
	}, 300, 240)
	click(t, tt, "Segment draft")
	tt.Key(ui.Ctrl, ui.KeyA)
	slices.Reverse(items)
	orientation = nui.Vertical
	tt.Frame()
	if !tt.Focused("Segment draft") {
		t.Fatal("segment reorder or orientation discarded focus")
	}
	tt.Type("Replacement")
	if draft != "Replacement" {
		t.Fatalf("segment reorder discarded selection: %q", draft)
	}
}

func TestProgressGroupLegendFollowsSize(t *testing.T) {
	size := nui.XS
	tt := ui.NewTester(func(c *ui.Context) {
		nui.ProgressGroup(c, nui.App(c, nui.AppConfig{}), []nui.ProgressSegment{{ID: "first", Label: "First", Value: 50}}, nui.ProgressGroupOptions{Size: size}).Width(200)
	}, 240, 100)
	small := bounds(t, tt, "First")
	size = nui.XL
	tt.Frame()
	large := bounds(t, tt, "First")
	if large.H <= small.H {
		t.Fatalf("legend size unchanged: XS=%+v XL=%+v", small, large)
	}
}

func TestProgressGroupVerticalAndUnusedSpace(t *testing.T) {
	for _, orientation := range []nui.Orientation{nui.Horizontal, nui.Vertical} {
		items := []nui.ProgressSegment{{ID: "used", Value: 25}, {ID: "zero", Value: 0, Color: nui.Error}}
		tt := ui.NewTester(func(c *ui.Context) {
			nui.ProgressGroup(c, nui.App(c, nui.AppConfig{Mode: nui.LightMode}), items, nui.ProgressGroupOptions{Label: "Group", Orientation: orientation, Height: 200}).Width(200)
		}, 240, 240)
		sample := func(at float32) color.RGBA {
			r := bounds(t, tt, "Group")
			x, y := r.X+at, r.Y+4
			if orientation == nui.Vertical {
				x, y = r.X+4, r.Y+at
			}
			p := tt.Image().RGBAAt(int(x), int(y))
			return p
		}
		if sample(20) != rgba(nui.Light().Primary) || sample(150) != rgba(nui.Light().Accented) {
			t.Fatalf("orientation %d: wrong fill or unused space", orientation)
		}
		items = nil
		tt.Frame()
		if sample(20) != rgba(nui.Light().Accented) || sample(150) != rgba(nui.Light().Accented) {
			t.Fatalf("orientation %d: empty segments left a visible fill", orientation)
		}
	}
}
