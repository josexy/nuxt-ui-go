package gallery

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

func TestGalleryNumberOrientation(t *testing.T) {
	a := newStaticGallery()
	a.Page = 11
	tt := ui.NewTester(a.View, 1080, 820)
	press(t, tt, "Seats")
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("-")
	press(t, tt, "Number orientation")
	press(t, tt, "vertical")
	if a.FormsData.NumberOrientation != nui.Vertical || a.FormsData.NumberState.Text != "-" || a.FormsData.Number.Value != 3 {
		t.Fatal("Gallery orientation selector changed the number draft or value")
	}
	press(t, tt, "Increase Seats")
	if a.FormsData.Number.Value != 4 {
		t.Fatal("Gallery vertical increase did not update the value")
	}
	press(t, tt, "Number orientation")
	press(t, tt, "horizontal")
	press(t, tt, "Decrease Seats")
	if a.FormsData.NumberOrientation != nui.Horizontal || a.FormsData.Number.Value != 3 {
		t.Fatal("Gallery horizontal decrease did not update the value")
	}
}

func TestGallerySelectionAndColor(t *testing.T) {
	a := newStaticGallery()
	a.Page = 11
	tt := ui.NewTester(a.View, 1080, 820)
	press(t, tt, "Project members")
	press(t, tt, "Search Project members")
	tt.Type("Ada")
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	if len(a.FormsData.Chosen) != 1 || a.FormsData.Chosen[0] != "ada" {
		t.Fatalf("select menu %v open=%v active=%s searchFocus=%v", a.FormsData.Chosen, a.FormsData.Select.Open, a.FormsData.Select.Listbox.Active, tt.Focused("Search Project members"))
	}
	tt.Key(0, ui.KeyEscape)
	press(t, tt, "Dates & color")
	before := a.FormsData.Color
	press(t, tt, "Accent color A")
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("128")
	if a.FormsData.Color.A != 128 || a.FormsData.Color.R != before.R {
		t.Fatal("alpha overwrote RGB")
	}
	tt.Frame()
	if a.FormsData.Color.A != 128 {
		t.Fatal("native picker overwrote alpha")
	}
	press(t, tt, "Choose Release date")
	press(t, tt, "Release date calendar 2026-11-15")
	if a.FormsData.Date.Value.Day != 15 {
		t.Fatal("popup date sync")
	}
}
func TestFormsDataVisualArtifacts(t *testing.T) {
	dir := os.Getenv("NUI_ARTIFACTS")
	if dir == "" {
		t.Skip("NUI_ARTIFACTS not set")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, dark := range []bool{false, true} {
		for _, scale := range []float32{1, 1.25, 1.5, 2} {
			for _, scene := range []string{"select-menu", "date-popup", "wizard", "data-edit", "narrow"} {
				a := newStaticGallery()
				a.Dark = dark
				a.Page = 11
				tt := ui.NewTester(a.View, 1080, 820)
				tt.SetScale(scale)
				switch scene {
				case "select-menu":
					press(t, tt, "Project members")
				case "date-popup":
					a.Page = 12
					tt.Frame()
					press(t, tt, "Choose Release date")
				case "wizard":
					a.Page = 7
					a.Settings.Page = 3
					tt.Frame()
					press(t, tt, "Continue")
				case "data-edit":
					a.Page = 13
					tt.Frame()
					press(t, tt, "Task 00001")
					press(t, tt, "Edit task")
				case "narrow":
					tt.SetSize(720, 900)
					tt.SetPreferences(ui.Preferences{TextScale: 1.5, ReduceMotion: true})
				}
				tt.Move(0, 0)
				f, err := os.Create(filepath.Join(dir, fmt.Sprintf("forms-data-%s-dark-%v-scale-%g.png", scene, dark, scale)))
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

func TestNarrowSectionsDoNotOverlap(t *testing.T) {
	a := newStaticGallery()
	a.Page = 11
	tt := ui.NewTester(a.View, 720, 900)
	tt.SetPreferences(ui.Preferences{TextScale: 1.5, ReduceMotion: true})
	slider, ok := tt.Find("Budget range")
	if !ok {
		t.Fatal("range missing")
	}
	trigger, ok := tt.Find("Project members")
	if !ok {
		t.Fatal("trigger missing")
	}
	if trigger.Y < slider.Y+slider.H {
		t.Fatalf("overlap: range=%+v trigger=%+v", slider, trigger)
	}
}
