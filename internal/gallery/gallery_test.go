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

func newStaticGallery() *State {
	a := New()
	a.Animations = false
	return a
}

func press(t *testing.T, tt *ui.Tester, label string) {
	t.Helper()
	if err := tt.Click(label); err != nil {
		t.Fatal(err)
	}
}

func TestGalleryNavigatesAndSettles(t *testing.T) {
	a := newStaticGallery()
	tt := ui.NewTester(a.View, 1080, 820)
	for _, label := range []string{"Inputs & overlays", "Data & menus", "Theme & buttons", "Display", "Feedback", "Workspace", "Forms & settings", "Form states", "Overlays & menus", "Navigation & disclosure", "Advanced forms", "Dates & color", "Task browser", "Advanced interactions", "Activity & attachments", "Application layout", "Components", "Theme"} {
		press(t, tt, label)
		tt.Advance(300 * time.Millisecond)
		if tt.FramePending() {
			t.Fatalf("%s did not settle", label)
		}
	}
	press(t, tt, "Dark")
	if !a.Dark {
		t.Fatal("theme did not change")
	}
}

func TestWorkspaceCompositionAndThemeControls(t *testing.T) {
	a := newStaticGallery()
	a.Page = 6
	tt := ui.NewTester(a.View, 1080, 820)
	press(t, tt, "Save workspace")
	if a.Saved != 1 {
		t.Fatal("save action")
	}
	press(t, tt, "Add example document")
	if !a.HasDocument {
		t.Fatal("empty state did not change")
	}
	press(t, tt, "Foundation component notes")
	if a.Message != "Document opened" {
		t.Fatal("document link")
	}
	press(t, tt, "Remove document")
	if a.HasDocument {
		t.Fatal("remove action")
	}
	press(t, tt, "Preferences")
	if a.Page != 3 {
		t.Fatal("preferences navigation")
	}
	press(t, tt, "Choose primary color")
	tt.Type("blue")
	tt.Key(0, ui.KeyEnter)
	press(t, tt, "Toggle density")
	press(t, tt, "Mode: light")
	press(t, tt, "Mode: dark")
	if a.Primary != "blue" || !a.Compact || !a.System {
		t.Fatal("theme configuration")
	}
	tt.SetDark(true)
	tt.SetPreferences(ui.Preferences{TextScale: 1.5, ReduceMotion: true})
	// Repeated responsive layout and dynamic removal must not retain build values.
	for _, width := range []int{720, 1440, 1080} {
		tt.SetSize(width, 820)
		press(t, tt, "Workspace")
		tt.Scroll(float32(width-100), 400, 0, 600)
		press(t, tt, "Add example document")
		tt.Scroll(float32(width-100), 400, 0, 600)
		press(t, tt, "Remove document")
	}
}

func TestTableVirtualizesSortsAndPreservesSelection(t *testing.T) {
	for _, count := range []int{100, 10_000, 100_000} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			a := newStaticGallery()
			a.SetRows(count)
			tt := ui.NewTester(func(c *ui.Context) {
				s := nui.Window(c, nui.Light())
				ui.Column(c).Fill().Padding(16).Children(func() { a.Table(c, s) })
			}, 700, 450)
			if a.CellCalls > 150 {
				t.Fatalf("built %d cells for %d rows", a.CellCalls, count)
			}
			t.Logf("rows=%d last-build-cells=%d", count, a.CellCalls)
			press(t, tt, "Workspace 00001")
			if !a.Selection.Has("workspace-0") {
				t.Fatal("row not selected by stable ID")
			}
			press(t, tt, "Workspace")
			press(t, tt, "Workspace")
			if !a.Sort.Descending || a.Rows[0].ID == "workspace-0" {
				t.Fatalf("sorting failed: %+v", a.Sort)
			}
			if !a.Selection.Has("workspace-0") {
				t.Fatal("sorting lost selection")
			}
			// Move the first column edge in the table's committed coordinates.
			header, ok := tt.Find("Workspace")
			if !ok {
				t.Fatal("header missing")
			}
			box, ok := tt.Find("Workspace table")
			if !ok {
				t.Fatal("table missing")
			}
			x, y := box.X+300, header.Y+header.H/2
			tt.Press(x, y)
			tt.Move(x+40, y)
			tt.Release(x+40, y)
			if a.List.Columns.Widths["name"] <= 300 {
				t.Fatalf("resize missing: %+v", a.List.Columns)
			}
		})
	}
}

func TestDrawnMenuSubmenuAndRichComposition(t *testing.T) {
	a := newStaticGallery()
	a.Page = 2
	tt := ui.NewTester(a.View, 1080, 820)
	press(t, tt, "Workspace actions")
	press(t, tt, "Export")
	press(t, tt, "CSV file")
	if a.MenuResult != "Exported CSV" {
		t.Fatalf("menu result %q", a.MenuResult)
	}
	press(t, tt, "Rich actions")
	press(t, tt, "Export options")
	press(t, tt, "Export rich CSV")
	if a.MenuResult != "Rich CSV exported" || a.RichOpen || a.RichSubOpen {
		t.Fatalf("rich composition state %+v", a.MenuResult)
	}
}

// Set NUI_ARTIFACTS to capture repeatable software-rendered views for review.
func TestGalleryVisualArtifacts(t *testing.T) {
	directory := os.Getenv("NUI_ARTIFACTS")
	if directory == "" {
		t.Skip("NUI_ARTIFACTS not set")
	}
	if err := os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	for _, dark := range []bool{false, true} {
		for _, scale := range []float32{1, 1.25, 1.5, 2} {
			for page := range 17 {
				a := newStaticGallery()
				a.Dark = dark
				a.Page = page
				tt := ui.NewTester(a.View, 1080, 820)
				tt.SetScale(scale)
				name := fmt.Sprintf("gallery-dark-%v-page-%d-scale-%g.png", dark, page, scale)
				f, err := os.Create(filepath.Join(directory, name))
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
