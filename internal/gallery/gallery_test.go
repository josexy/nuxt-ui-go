package gallery

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

func newStaticGallery() *State { a := New(); a.Animations = false; return a }
func press(t *testing.T, tt *ui.Tester, label string) {
	t.Helper()
	for range 12 {
		bounds, ok := tt.Find(label)
		if !ok {
			t.Fatalf("missing %q", label)
		}
		height := float32(tt.Image().Bounds().Dy())
		if bounds.Y >= 0 && bounds.Y+bounds.H <= height {
			if err := tt.Click(label); err != nil {
				t.Fatal(err)
			}
			return
		}
		delta := bounds.Y + bounds.H - height + 80
		if bounds.Y < 0 {
			delta = bounds.Y - 80
		}
		tt.Scroll(bounds.X+bounds.W/2, height/2, 0, delta)
		tt.Frame()
	}
	t.Fatalf("could not scroll %q into view", label)
}
func requireInput(t *testing.T, tt *ui.Tester, label, value string) {
	t.Helper()
	press(t, tt, label)
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Key(ui.Ctrl, ui.KeyC)
	if got := tt.Clipboard(); got != value {
		t.Fatalf("%s value %q, want %q", label, got, value)
	}
}
func focus(t *testing.T, tt *ui.Tester, label string) {
	t.Helper()
	for range 120 {
		if tt.Focused(label) {
			return
		}
		tt.Key(0, ui.KeyTab)
	}
	t.Fatalf("could not focus %q", label)
}
func requireText(t *testing.T, tt *ui.Tester, text string) {
	t.Helper()
	if !tt.HasText(text) {
		t.Fatalf("missing %q in %v", text, tt.Texts())
	}
}
func requireTextContains(t *testing.T, tt *ui.Tester, text string) {
	t.Helper()
	for _, v := range tt.Texts() {
		if strings.Contains(v, text) {
			return
		}
	}
	t.Fatalf("missing text containing %q in %v", text, tt.Texts())
}
func sceneTester(t *testing.T, a *State, id, sceneID string) *ui.Tester {
	t.Helper()
	a.Navigate(id)
	p := a.page(id)
	for _, s := range p.Scenes {
		if s.ID == sceneID {
			return ui.NewTester(func(c *ui.Context) {
				theme := nui.App(c, a.themeConfig())
				if a.closeToasts {
					nui.CloseToast(c, "")
					a.closeToasts = false
				}
				for _, notice := range a.pendingToasts {
					nui.Toast(c, theme, notice)
				}
				a.pendingToasts = nil
				nui.ToastViewport(c, theme, a.toastViewport)
				ui.Scroll(c.Key(id + "/" + s.ID)).Fill().Padding(20).Gap(12).Children(func() { s.Build(c, theme) })
			}, 1080, 820)
		}
	}
	t.Fatalf("missing scene %s/%s", id, sceneID)
	return nil
}

func TestGalleryNavigatesAndSettles(t *testing.T) {
	a := newStaticGallery()
	t.Cleanup(a.Dispose)
	tt := ui.NewTester(a.View, 1080, 820)
	for _, p := range componentDirectory {
		a.Navigate(p.ID)
		tt.Frame()
		tt.Advance(time.Second)
		if tt.FramePending() {
			t.Errorf("%s did not settle", p.ID)
		}
	}
	press(t, tt, "Theme")
	press(t, tt, "Dark")
	if !a.Dark {
		t.Fatal("theme did not change")
	}
}

func TestWorkspaceCompositionAndThemeControls(t *testing.T) {
	a := newStaticGallery()
	t.Cleanup(a.Dispose)
	a.Navigate("example/workspace")
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
	if a.ActiveID != "theme" {
		t.Fatal("preferences navigation")
	}
	press(t, tt, "Choose primary color")
	tt.Type("blue")
	tt.Key(0, ui.KeyEnter)
	press(t, tt, "Toggle density")
	press(t, tt, "Cycle theme mode")
	press(t, tt, "Cycle theme mode")
	if a.Primary != "blue" || !a.Compact || !a.System {
		t.Fatal("theme configuration")
	}
	tt.SetDark(true)
	tt.SetPreferences(ui.Preferences{TextScale: 1.5, ReduceMotion: true})
	for _, width := range []int{720, 1440, 1080} {
		tt.SetSize(width, 820)
		a.Navigate("example/workspace")
		tt.Frame()
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
			if !a.Sort.Descending || a.RowID(0) == "workspace-0" {
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
			x, y := box.X+297, header.Y+header.H/2
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
	t.Cleanup(a.Dispose)
	tt := sceneTester(t, a, "dropdown-menu", "nested")
	press(t, tt, "Project actions nested")
	press(t, tt, "Export")
	press(t, tt, "CSV file")
	requireText(t, tt, "CSV export prepared")
	tt = sceneTester(t, a, "dropdown-menu", "rich")
	press(t, tt, "Rich actions")
	press(t, tt, "Export options")
	press(t, tt, "Export rich CSV")
	requireText(t, tt, "Rich CSV exported")
}

func captureGallery(t *testing.T, tt *ui.Tester, name string) {
	t.Helper()
	directory := os.Getenv("NUI_ARTIFACTS")
	if directory == "" {
		return
	}
	if err := os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(directory, name+".png"))
	if err != nil {
		t.Fatal(err)
	}
	err = png.Encode(f, tt.Image())
	closeErr := f.Close()
	if err != nil || closeErr != nil {
		t.Fatalf("capture: %v %v", err, closeErr)
	}
}

// Screenshot names follow the same stable component and scene registry as the UI.
func TestGalleryVisualArtifacts(t *testing.T) {
	if os.Getenv("NUI_ARTIFACTS") == "" {
		t.Skip("NUI_ARTIFACTS not set")
	}
	for _, dark := range []bool{false, true} {
		a := newStaticGallery()
		a.Dark = dark
		t.Cleanup(a.Dispose)
		for _, p := range componentDirectory {
			for _, s := range a.page(p.ID).Scenes {
				tt := sceneTester(t, a, p.ID, s.ID)
				captureGallery(t, tt, fmt.Sprintf("scene-%s-%s-dark-%t", strings.ReplaceAll(p.ID, "/", "-"), s.ID, dark))
			}
		}
		for _, width := range []int{720, 1080, 1440} {
			for _, scale := range []float32{1, 1.5} {
				for _, id := range []string{"button", "input", "modal", "table", "chart", "sidebar"} {
					cell := newStaticGallery()
					cell.Dark = dark
					cell.Navigate(id)
					tt := ui.NewTester(cell.View, width, 900)
					tt.SetPreferences(ui.Preferences{TextScale: scale, ReduceMotion: true})
					switch id {
					case "button":
						press(t, tt, "Save changes")
						requireText(t, tt, "Saved 1 times")
					case "input":
						press(t, tt, "Workspace name")
						tt.Key(ui.Ctrl, ui.KeyA)
						tt.Type("Review / 中文 draft")
					case "modal":
						press(t, tt, "Open basic dialog")
					case "table":
						press(t, tt, "Select first record")
						requireTextContains(t, tt, "1 selected")
					case "chart":
						press(t, tt, "Update data")
					case "sidebar":
						if width < 880 {
							press(t, tt, "Open inner slideover sidebar")
						}
					}
					if cell.ActiveID != id {
						t.Fatalf("visual action navigated away from %s", id)
					}
					captureGallery(t, tt, fmt.Sprintf("shell-%s-%d-text-%g-dark-%t", id, width, scale, dark))
					cell.Dispose()
				}
			}
		}
		for _, id := range []string{"button", "modal", "sidebar"} {
			cell := newStaticGallery()
			cell.Dark = dark
			cell.Navigate(id)
			tt := ui.NewTester(cell.View, 720, 900)
			tt.SetDirection(ui.RightToLeft)
			tt.SetPreferences(ui.Preferences{TextScale: 1.5, ReduceMotion: true})
			if id == "modal" {
				press(t, tt, "Open basic dialog")
			}
			if id == "sidebar" {
				press(t, tt, "Open inner drawer sidebar")
			}
			captureGallery(t, tt, fmt.Sprintf("rtl-%s-dark-%t", id, dark))
			cell.Dispose()
		}

	}
}
