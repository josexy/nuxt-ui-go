package gallery

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

func TestPropsGallery(t *testing.T) {
	a := newStaticGallery()
	a.Page = 17
	t.Cleanup(a.Dispose)
	tt := ui.NewTester(a.View, 1080, 1000)
	press(t, tt, "Loading switch")
	press(t, tt, "Sync workspace")
	if a.Props.Enabled {
		t.Fatal("loading preview accepted input")
	}
	press(t, tt, "Selection policies")
	press(t, tt, "Team menu")
	tt.Type("Alpha")
	press(t, tt, "Alpha")
	press(t, tt, "Team menu")
	if !tt.HasText("Beta") {
		t.Fatal("Gallery search reset")
	}
	tt.Key(0, ui.KeyEscape)
	a.Page = 18
	tt.Frame()
	press(t, tt, "Shared orientation")
	press(t, tt, "vertical")
	if a.Props.Direction != nui.Vertical {
		t.Fatal("orientation toggle")
	}
	press(t, tt, "Orientation name")
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("Edited")
	press(t, tt, "Shared orientation")
	press(t, tt, "horizontal")
	if a.Props.Name != "Edited" {
		t.Fatal("orientation value reset")
	}
	for _, width := range []int{720, 1080} {
		tt.SetSize(width, 1000)
		tt.SetPreferences(ui.Preferences{TextScale: 1.5, ReduceMotion: true})
		tt.Move(0, 0)
		tt.Advance(time.Second)
		_ = tt.Image()
	}
}

func TestPropsGalleryArtifacts(t *testing.T) {
	dir := os.Getenv("NUI_PROP_ARTIFACTS")
	if dir == "" {
		t.Skip("NUI_PROP_ARTIFACTS not set")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, dark := range []bool{false, true} {
		for _, scene := range []string{"forms", "selection", "overlays", "horizontal", "vertical", "narrow"} {
			a := newStaticGallery()
			a.Page = 17
			a.Dark = dark
			a.Props.Demo = scene
			width, height := 1080, 1000
			if scene == "horizontal" || scene == "vertical" || scene == "narrow" {
				a.Page = 18
			}
			if scene == "vertical" || scene == "narrow" {
				a.Props.Direction = nui.Vertical
			}
			if scene == "narrow" {
				width, height = 720, 1000
			}
			tt := ui.NewTester(a.View, width, height)
			if scene == "narrow" {
				tt.SetPreferences(ui.Preferences{TextScale: 1.5, ReduceMotion: true})
			}
			tt.Move(0, 0)
			capture := func(suffix string) {
				f, err := os.Create(filepath.Join(dir, fmt.Sprintf("props-%s%s-dark-%v.png", scene, suffix, dark)))
				if err != nil {
					t.Fatal(err)
				}
				err = png.Encode(f, tt.Image())
				closeErr := f.Close()
				if err != nil {
					t.Fatal(err)
				}
				if closeErr != nil {
					t.Fatal(closeErr)
				}
			}
			capture("")
			if a.Page == 18 {
				tt.Scroll(float32(width-30), float32(height-50), 0, 650)
				capture("-bottom")
			}
			a.Dispose()
		}
	}
}
