package gallery

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestOverlayNavigationVisualArtifacts(t *testing.T) {
	dir := os.Getenv("NUI_ARTIFACTS")
	if dir == "" {
		t.Skip("NUI_ARTIFACTS not set")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, dark := range []bool{false, true} {
		for _, scale := range []float32{1, 1.25, 1.5, 2} {
			for _, scene := range []string{"menu", "modal", "side", "toast", "activity"} {
				a := newStaticGallery()
				a.Dark, a.Page = dark, 9
				tt := ui.NewTester(a.View, 1080, 820)
				tt.SetScale(scale)
				switch scene {
				case "menu":
					press(t, tt, "Project actions")
					tt.Key(0, ui.KeyEnd)
					tt.Key(0, ui.KeyUp)
					tt.Key(0, ui.KeyRight)
				case "modal":
					press(t, tt, "Review changes")
					press(t, tt, "Review plan")
				case "side":
					press(t, tt, "Right panel")
					tt.Advance(250 * time.Millisecond)
				case "toast":
					press(t, tt, "Show notification")
					tt.Advance(250 * time.Millisecond)
				case "activity":
					press(t, tt, "Navigation & disclosure")
					press(t, tt, "Activity")
				}
				tt.Move(0, 0)
				f, err := os.Create(filepath.Join(dir, fmt.Sprintf("overlays-%s-dark-%v-scale-%g.png", scene, dark, scale)))
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
		}
	}
}
