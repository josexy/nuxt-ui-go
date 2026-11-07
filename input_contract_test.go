package nui_test

import (
	"fmt"
	"image/color"
	"testing"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

func TestInputLoadingUsesPartForeground(t *testing.T) {
	value := ""
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		th.Motion.Reduce = true
		nui.Input(c, th, &value, nui.InputOptions{Label: "Loading field", Loading: true, UI: nui.Styles{nui.Leading: {Foreground: nui.Set(ui.RGB(255, 0, 255))}}, Style: func(root, editor, leading, trailing ui.Element) {
			leading.Label("Input loader")
		}}).Width(240)
	}, 300, 80)
	r := bounds(t, tt, "Input loader")
	if countWhere(tt.Image(), r, func(c color.RGBA) bool { return int(c.R) > int(c.G)+20 && int(c.B) > int(c.G)+20 }) == 0 {
		t.Fatal("loading input ignored Leading.Foreground")
	}
}

func TestInputNeutralRingAndNoneHighlight(t *testing.T) {
	for _, variant := range []nui.Variant{nui.Outline, nui.None} {
		t.Run(fmt.Sprint(variant), func(t *testing.T) {
			value := "Text"
			accent := ui.RGB(126, 34, 206)
			highlight := variant == nui.None
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
				th.InvertedBackground = accent
				ui.Column(c).Padding(16).Children(func() {
					nui.Input(c, th, &value, nui.InputOptions{Label: "Field", Color: nui.Neutral, Variant: variant, Highlight: highlight}).Width(240).Label("Input root")
				})
			}, 300, 90)
			if !highlight {
				click(t, tt, "Field")
			}
			r := bounds(t, tt, "Input root")
			if got := tt.Image().RGBAAt(int(r.X+r.W/2), int(r.Y)); got != rgba(accent) {
				t.Fatalf("variant=%v highlight=%v border=%v, want %v", variant, highlight, got, rgba(accent))
			}
		})
	}
}

func TestInputAvatarAndLoadingPriority(t *testing.T) {
	for i, size := range []nui.Size{nui.XS, nui.SM, nui.MD, nui.LG, nui.XL} {
		value := "Draft"
		av := nui.AvatarOptions{Alt: "Input avatar", Text: "AL"}
		opts := nui.InputOptions{Label: "Account", Size: size, Avatar: &av, Clearable: true, Trailing: lucide.Check(), Style: func(root, editor, leading, trailing ui.Element) { trailing.Label("Trailing part") }}
		tt := ui.NewTester(func(c *ui.Context) {
			th := nui.App(c, nui.AppConfig{})
			th.Motion.Reduce = true
			nui.Input(c.Key("account"), th, &value, opts).Width(260)
		}, 320, 90)
		want := []float32{16, 16, 20, 20, 24}[i]
		if r := bounds(t, tt, "Input avatar"); r.W != want || r.H != want {
			t.Fatalf("size %d: avatar=%+v, want %g", size, r, want)
		}
		click(t, tt, "Account")
		tt.Key(ui.Ctrl, ui.KeyA)
		opts.Loading, opts.LoadingTrailing = true, true
		tt.Frame()
		if !tt.HasText("Input avatar") || !tt.Focused("Account") {
			t.Fatal("trailing loading removed the avatar or editor focus")
		}
		opts.Leading = lucide.Search()
		tt.Frame()
		if tt.HasText("Input avatar") {
			t.Fatal("explicit leading icon did not move the loader before the input")
		}
		tt.Type("Edited while loading")
		if value != "Edited while loading" {
			t.Fatal("loading changed editing or selection")
		}
		opts.Loading, opts.Leading = false, nil
		av.Size = nui.XS
		tt.Frame()
		if r := bounds(t, tt, "Input avatar"); r.W != 24 || r.H != 24 {
			t.Fatalf("explicit avatar size was overridden: %+v", r)
		}
	}
}

func TestInputTrailingLoaderColorAndRTL(t *testing.T) {
	for _, rtl := range []bool{false, true} {
		value := "Draft"
		tt := ui.NewTester(func(c *ui.Context) {
			th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
			th.Motion.Reduce = true
			nui.Input(c, th, &value, nui.InputOptions{Label: "Account", Avatar: &nui.AvatarOptions{Text: "AL", Alt: "Input avatar"}, Loading: true, LoadingTrailing: true, UI: nui.Styles{nui.Trailing: {Foreground: nui.Set(ui.RGB(255, 0, 255))}}, Style: func(root, editor, leading, trailing ui.Element) { trailing.Label("Trailing loader") }}).Width(260)
		}, 320, 90)
		if rtl {
			tt.SetDirection(ui.RightToLeft)
		}
		av, edit, load := bounds(t, tt, "Input avatar"), bounds(t, tt, "Account"), bounds(t, tt, "Trailing loader")
		if !rtl && !(av.X < edit.X && edit.X < load.X) || rtl && !(av.X > edit.X && edit.X > load.X) {
			t.Fatalf("RTL=%v layout: avatar=%+v editor=%+v loader=%+v", rtl, av, edit, load)
		}
		if countWhere(tt.Image(), load, func(c color.RGBA) bool { return int(c.R) > int(c.G)+20 && int(c.B) > int(c.G)+20 }) == 0 {
			t.Fatal("trailing loader ignored part color")
		}
	}
}

func TestInputControlledUpdatesAndDisabledClear(t *testing.T) {
	value, changes, submits, clears := "Original", 0, 0, 0
	disabled, readonly := false, false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Box(c).Disabled(disabled).Children(func() {
			nui.Input(c, th, &value, nui.InputOptions{Label: "Account", ReadOnly: readonly, Clearable: true, ClearLabel: "Clear account", OnChange: func(string) { changes++ }, OnSubmit: func(string) { submits++ }, OnClear: func() { clears++ }}).Width(260)
		})
	}, 320, 90)
	click(t, tt, "Account")
	value = "External"
	tt.Frame()
	tt.Frame()
	if changes != 0 || submits != 0 || clears != 0 || !tt.Focused("Account") {
		t.Fatal("controlled update notified or lost focus")
	}
	disabled = true
	tt.Frame()
	click(t, tt, "Clear account")
	if value != "External" || changes != 0 || clears != 0 {
		t.Fatal("disabled clear changed the value")
	}
	disabled, readonly = false, true
	tt.Frame()
	if _, ok := tt.Find("Clear account"); ok {
		t.Fatal("read-only input displayed a clear action")
	}
	readonly = false
	tt.Frame()
	click(t, tt, "Clear account")
	tt.Frame()
	if value != "" || changes != 1 || clears != 1 || !tt.Focused("Account") {
		t.Fatalf("clear value=%q callbacks=%d/%d", value, changes, clears)
	}
	tt.Type("Updated")
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if value != "Updated" || changes != 2 || submits != 1 {
		t.Fatalf("input or submit value=%q callbacks=%d/%d", value, changes, submits)
	}
}

func TestInputRejectedLengthEditDoesNotNotify(t *testing.T) {
	value, changes := "A", 0
	tt := ui.NewTester(func(c *ui.Context) {
		nui.Input(c, nui.App(c, nui.AppConfig{}), &value, nui.InputOptions{Label: "Limited", MaxLength: 1, Clearable: true, ClearLabel: "Clear limited", OnChange: func(string) { changes++ }}).Width(200)
	}, 260, 90)
	click(t, tt, "Limited")
	tt.Key(0, ui.KeyEnd)
	tt.Type("B")
	if value != "A" || changes != 0 {
		t.Fatalf("rejected length edit value=%q changes=%d", value, changes)
	}
	click(t, tt, "Clear limited")
	if value != "" || changes != 1 {
		t.Fatal("length limit suppressed the clear notification")
	}
	tt.Type("中文")
	if value != "中" || changes != 2 {
		t.Fatal("partial accepted edit did not notify once")
	}
}
