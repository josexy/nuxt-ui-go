package nui_test

import (
	"bytes"
	"fmt"
	"image/color"
	"slices"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

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
