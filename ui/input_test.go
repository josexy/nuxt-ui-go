package nui_test

import (
	"bytes"
	"fmt"
	"image/color"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
	nui "github.com/josexy/nuxt-ui-go/ui"
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
				tt.Advance(200 * time.Millisecond) // the colour transition
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

func TestTextareaRowsAndControlledGrowth(t *testing.T) {
	for _, tc := range []struct {
		rows, maxRows int
		auto          bool
		want          float32
	}{{0, 0, false, 72}, {2, 8, false, 52}, {3, 1, true, 72}, {2, 4, true, 92}, {2, 0, true, 212}} {
		t.Run(fmt.Sprint(tc), func(t *testing.T) {
			value := strings.Repeat("Line\n", 9) + "Last"
			tt := ui.NewTester(func(c *ui.Context) {
				nui.Textarea(c, nui.App(c, nui.AppConfig{}), &value, nui.TextareaOptions{Label: "Notes", Rows: tc.rows, MaxRows: tc.maxRows, AutoResize: tc.auto}).Width(240).Label("Textarea root")
			}, 300, 320)
			if r := bounds(t, tt, "Textarea root"); r.H != tc.want {
				t.Fatalf("rows=%d max=%d auto=%v height=%g, want %g", tc.rows, tc.maxRows, tc.auto, r.H, tc.want)
			}
			value = "Short"
			tt.Frame()
			want := float32(max(tc.rows, 1)*20 + 12)
			if tc.rows <= 0 {
				want = 72
			}
			if r := bounds(t, tt, "Textarea root"); r.H != want {
				t.Fatalf("controlled shrinking height=%g, want %g", r.H, want)
			}
		})
	}
}

func TestTextareaSharedAccessoriesAndEditing(t *testing.T) {
	value, changes, submits, clears := "Original\nNotes", 0, 0, 0
	avatar, auto, password := false, false, false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		th.Motion.Reduce = true
		var av *nui.AvatarOptions
		if avatar {
			av = &nui.AvatarOptions{Alt: "Notes avatar", Text: "AL"}
		}
		nui.Textarea(c, th, &value, nui.TextareaOptions{Label: "Notes", Password: password, Avatar: av, Loading: true, LoadingTrailing: true, Clearable: true, ClearLabel: "Clear notes", MaxLength: 20, UI: nui.Styles{nui.Trailing: {Foreground: nui.Set(ui.RGB(255, 0, 255))}}, OnChange: func(string) { changes++ }, OnSubmit: func(string) { submits++ }, OnClear: func() { clears++ }, Style: func(root, editor, leading, trailing ui.Element) { trailing.Label("Notes loader") }, Rows: 2, AutoResize: auto, MaxRows: 4}).Width(260)
	}, 320, 180)
	click(t, tt, "Notes")
	tt.Key(ui.Ctrl, ui.KeyA)
	avatar, auto = true, true
	tt.Frame()
	if !tt.Focused("Notes") || !tt.HasText("Notes avatar") {
		t.Fatal("textarea accessory or row changes lost focus")
	}
	av, edit, load := bounds(t, tt, "Notes avatar"), bounds(t, tt, "Notes"), bounds(t, tt, "Notes loader")
	if av.Y != edit.Y || load.Y != edit.Y || av.W != 20 {
		t.Fatalf("textarea accessories not aligned at the first row: %+v %+v %+v", av, edit, load)
	}
	if countWhere(tt.Image(), load, func(c color.RGBA) bool { return int(c.R) > int(c.G)+20 && int(c.B) > int(c.G)+20 }) == 0 {
		t.Fatal("textarea loader ignored its part color")
	}
	tt.Type("Replacement")
	tt.Key(0, ui.KeyEnter)
	tt.Key(ui.Shift, ui.KeyEnter)
	if value != "Replacement\n\n" || changes != 3 || submits != 0 {
		t.Fatalf("textarea editing value=%q changes=%d submits=%d", value, changes, submits)
	}
	before := slices.Clone(tt.Image().Pix)
	password = true
	tt.Frame()
	if !bytes.Equal(before, tt.Image().Pix) {
		t.Fatal("shared Password setting unexpectedly changed the multiline display")
	}
	click(t, tt, "Clear notes")
	if value != "" || changes != 4 || clears != 1 || !tt.Focused("Notes") {
		t.Fatal("textarea clear did not notify once and restore focus")
	}
	tt.Type(strings.Repeat("界", 30))
	tt.Type("拒绝")
	if value != strings.Repeat("界", 20) || changes != 5 {
		t.Fatal("textarea length limit notified a fully rejected edit")
	}
}
