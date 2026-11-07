package nui_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
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
