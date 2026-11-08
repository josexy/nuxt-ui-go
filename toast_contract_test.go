package nui_test

import (
	"fmt"
	"image/color"
	"math"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

func TestToastPositionsAndNewestOrderInBothDirections(t *testing.T) {
	for _, rtl := range []bool{false, true} {
		for position := nui.ToastBottomRight; position <= nui.ToastTopCenter; position++ {
			t.Run(fmt.Sprintf("%d/rtl=%v", position, rtl), func(t *testing.T) {
				count := 0
				tt := ui.NewTester(func(c *ui.Context) {
					th := nui.App(c, nui.AppConfig{})
					th.Motion.Reduce = true
					nui.Button(c, th, "Notify", nui.ButtonOptions{}).Absolute().Left(250).Top(180).Width(100).OnClick(func() {
						count++
						nui.Toast(c, th, nui.ToastOptions{Title: fmt.Sprintf("Notice %d", count), Timeout: -1})
					})
					nui.ToastViewport(c, th, nui.ToastViewportOptions{Position: position})
				}, 600, 400)
				if rtl {
					tt.SetDirection(ui.RightToLeft)
				}
				click(t, tt, "Notify")
				click(t, tt, "Notify")
				old, newest := bounds(t, tt, "Notice 1"), bounds(t, tt, "Notice 2")
				wantX := float32(200)
				if position == nui.ToastBottomLeft || position == nui.ToastTopLeft {
					wantX = 16
				}
				if position == nui.ToastBottomCenter || position == nui.ToastTopCenter {
					wantX = 108
				}
				if newest.X != wantX {
					t.Fatalf("physical position x=%v, want %v", newest.X, wantX)
				}
				if position >= nui.ToastTopRight {
					if newest.Y != 16 || newest.Y >= old.Y {
						t.Fatalf("newest top toast not at edge: old=%v new=%v", old, newest)
					}
				} else if newest.Y <= old.Y || newest.Y+newest.H != 384 {
					t.Fatalf("newest bottom toast not at edge: old=%v new=%v", old, newest)
				}
			})
		}
	}
}

func TestToastDescriptionOnlyAndActionSpacing(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		th.Motion.Reduce = true
		nui.Button(c, th, "Notify", nui.ButtonOptions{}).OnClick(func() {
			nui.Toast(c, th, nui.ToastOptions{Description: "Description only", Timeout: -1, Actions: []nui.ToastActionOptions{{Label: "Run"}, {Label: "Keep"}}, UI: nui.Styles{nui.Actions: {Gap: nui.Set(float32(20))}}})
		})
		nui.ToastViewport(c, th, nui.ToastViewportOptions{Position: nui.ToastTopLeft})
	}, 600, 400)
	click(t, tt, "Notify")
	desc, run, keep := bounds(t, tt, "Description only"), bounds(t, tt, "Run"), bounds(t, tt, "Keep")
	if desc.Y != 32 {
		t.Errorf("empty title reserves a line: description=%v", desc)
	}
	if gap := run.Y - desc.Y - desc.H; gap != 10 {
		t.Errorf("vertical action gap=%v, want 10", gap)
	}
	if gap := keep.X - run.X - run.W; gap != 20 {
		t.Errorf("Actions.Gap=%v, want 20", gap)
	}
}

func TestToastHorizontalActionsAreCentered(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		th.Motion.Reduce = true
		nui.Button(c, th, "Notify", nui.ButtonOptions{}).OnClick(func() {
			nui.Toast(c, th, nui.ToastOptions{Title: "Centered notice", Description: "First line\nSecond line", Timeout: -1, Orientation: nui.Set(nui.Horizontal), Actions: []nui.ToastActionOptions{{Label: "Run"}}})
		})
		nui.ToastViewport(c, th, nui.ToastViewportOptions{Position: nui.ToastTopLeft})
	}, 600, 400)
	click(t, tt, "Notify")
	panel, action, close := bounds(t, tt, "Centered notice"), bounds(t, tt, "Run"), bounds(t, tt, "Dismiss Centered notice")
	for _, r := range []ui.Rect{action, close} {
		if math.Abs(float64(r.Y+r.H/2-panel.Y-panel.H/2)) > 1 {
			t.Fatalf("horizontal control is not centered: panel=%v control=%v", panel, r)
		}
	}
}

func TestToastWidthFitsCustomViewportPadding(t *testing.T) {
	padding := float32(48)
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		th.Motion.Reduce = true
		nui.Button(c, th, "Notify", nui.ButtonOptions{}).OnClick(func() { nui.Toast(c, th, nui.ToastOptions{Title: "Padded notice", Timeout: -1}) })
		nui.ToastViewport(c, th, nui.ToastViewportOptions{UI: nui.Styles{nui.Root: {PaddingX: nui.Set(padding)}}})
	}, 400, 300)
	click(t, tt, "Notify")
	if r := bounds(t, tt, "Padded notice"); r.X < 48 || r.X+r.W > 352 || r.W > 304 {
		t.Fatalf("toast escapes the configured viewport padding: %v", r)
	}
	padding = 0
	tt.Frame()
	if r := bounds(t, tt, "Padded notice"); r.W != 384 || r.X+r.W != 400 {
		t.Fatalf("zero padding retained a hidden width limit: %v", r)
	}
}

func TestToastAvatarCloseAndProgressSnapshots(t *testing.T) {
	closed := 0
	avatar := nui.AvatarOptions{Text: "AB", Alt: "Toast avatar"}
	close := nui.ButtonOptions{Size: nui.XL, AccessibleLabel: "Configured close", Loading: true, Style: func(_, _, leading, _ ui.Element) { leading.Label("Configured close icon") }}
	progress := nui.ToastProgressOptions{Color: nui.Error, UI: nui.Styles{nui.Indicator: {Background: nui.Set(ui.RGB(255, 0, 255))}}}
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		th.Motion.Reduce = true
		nui.Button(c, th, "Notify", nui.ButtonOptions{}).OnClick(func() {
			nui.Toast(c, th, nui.ToastOptions{ID: "stable", Title: "Snapshot notice", Description: "Description", Icon: lucide.Info(), Avatar: &avatar, CloseOptions: &close, CloseIcon: lucide.Check(), ProgressOptions: &progress, Timeout: time.Hour, OnClose: func() { closed++ }})
		})
		nui.ToastViewport(c, th, nui.ToastViewportOptions{})
	}, 600, 400)
	click(t, tt, "Notify")
	if r := bounds(t, tt, "Toast avatar"); r.W != 44 || r.H != 44 {
		t.Fatalf("toast avatar default=%v", r)
	}
	if !tt.HasText("AB") || bounds(t, tt, "Configured close icon").W != 24 {
		t.Fatal("avatar or configured close size missing")
	}
	panel := bounds(t, tt, "Snapshot notice")
	bar := func() color.RGBA { return tt.Image().RGBAAt(int(panel.X+panel.W/2), int(panel.Y+panel.H-2)) }
	if bar() != rgba(ui.RGB(255, 0, 255)) {
		t.Fatalf("progress UI override not drawn: %v", bar())
	}
	avatar.Text = "ZZ"
	close.Loading = false
	progress.UI[nui.Indicator] = nui.PartStyle{Background: nui.Set(ui.RGB(0, 255, 0))}
	tt.Frame()
	click(t, tt, "Configured close")
	if closed != 0 || !tt.HasText("AB") || tt.HasText("ZZ") || bar() != rgba(ui.RGB(255, 0, 255)) {
		t.Fatal("queued appearance changed with caller-owned configuration")
	}
	click(t, tt, "Notify")
	if !tt.HasText("ZZ") {
		t.Fatal("same-ID replacement did not update avatar")
	}
	click(t, tt, "Configured close")
	if closed != 1 {
		t.Fatalf("custom close notifications=%d", closed)
	}
}

func TestToastCountdownUpdatesWithReducedMotion(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		th.Motion.Reduce = true
		nui.Button(c, th, "Notify", nui.ButtonOptions{}).OnClick(func() { nui.Toast(c, th, nui.ToastOptions{Title: "Timer", Timeout: time.Second, Color: nui.Error}) })
		nui.ToastViewport(c, th, nui.ToastViewportOptions{})
	}, 600, 400)
	click(t, tt, "Notify")
	r := bounds(t, tt, "Timer")
	pixel := func() color.RGBA { return tt.Image().RGBAAt(int(r.X+r.W*.75), int(r.Y+r.H-2)) }
	before := pixel()
	tt.Advance(550 * time.Millisecond)
	if pixel() == before {
		t.Fatal("reduced-motion countdown did not update before expiry")
	}
}

func TestToastActionIdentityAcrossReplacementAndOrientation(t *testing.T) {
	swapped, horizontal := false, false
	chosen := ""
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		th.Motion.Reduce = true
		publish := func() {
			actions := []nui.ToastActionOptions{{ID: "keep", Label: "Keep", Close: nui.Set(false), OnAction: func() { chosen = "keep" }}, {ID: "review", Label: "Review", Close: nui.Set(false), OnAction: func() { chosen = "review" }}}
			if swapped {
				actions[0], actions[1] = actions[1], actions[0]
			}
			orientation := nui.Vertical
			if horizontal {
				orientation = nui.Horizontal
			}
			nui.Toast(c, th, nui.ToastOptions{ID: "actions", Title: "Actions notice", Actions: actions, Orientation: nui.Set(orientation), Timeout: -1})
		}
		nui.Button(c.Key("notify"), th, "Notify", nui.ButtonOptions{}).OnClick(publish)
		nui.Button(c.Key("update"), th, "Update notice", nui.ButtonOptions{}).KeepFocus().OnClick(publish)
		nui.ToastViewport(c, th, nui.ToastViewportOptions{})
	}, 800, 500)
	click(t, tt, "Notify")
	click(t, tt, "Keep")
	swapped = true
	click(t, tt, "Update notice")
	if !tt.Focused("Keep") {
		t.Fatal("action reordering moved focus to another action")
	}
	horizontal = true
	click(t, tt, "Update notice")
	if !tt.Focused("Keep") {
		t.Fatal("orientation replacement discarded focused action")
	}
	chosen = ""
	tt.Key(0, ui.KeyEnter)
	if chosen != "keep" {
		t.Fatalf("focused action activated %q", chosen)
	}
}

func TestToastDoesNotRetainSourceFieldDisabledState(t *testing.T) {
	pending := false
	actions := 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		th.Motion.Reduce = true
		nui.Button(c, th, "Notify", nui.ButtonOptions{}).OnClick(func() { pending = true })
		nui.FieldGroup(c, th, nui.FieldGroupOptions{Disabled: true}, []nui.FieldGroupItem{{ID: "source", Build: func(c *ui.Context, inner nui.Theme) ui.Element {
			if pending {
				pending = false
				nui.Toast(c, inner, nui.ToastOptions{Title: "Field result", Timeout: -1, Actions: []nui.ToastActionOptions{{ID: "run", Label: "Run action", Close: nui.Set(false), OnAction: func() { actions++ }}}})
			}
			return nui.Button(c, inner, "Disabled field", nui.ButtonOptions{})
		}}})
		nui.ToastViewport(c, th, nui.ToastViewportOptions{})
	}, 600, 400)
	click(t, tt, "Notify")
	click(t, tt, "Run action")
	if actions != 1 {
		t.Fatal("source field disabled the independent toast action")
	}
	click(t, tt, "Dismiss Field result")
	if tt.HasText("Field result") {
		t.Fatal("source field disabled the toast close action")
	}
}
