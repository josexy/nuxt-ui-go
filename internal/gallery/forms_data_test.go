package gallery

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestGalleryTreeCancellationAndStableSelection(t *testing.T) {
	a := newStaticGallery()
	t.Cleanup(a.Dispose)
	tt := sceneTester(t, a, "tree", "lazy")
	press(t, tt, "Load examples")
	tt.Key(0, ui.KeyRight)
	requireTextContains(t, tt, "Simulated load pending")
	a.Navigate("button")
	tt.Frame()
	press(t, tt, "Complete tree load")
	requireTextContains(t, tt, "Stale tree result rejected")
	a.Navigate("tree")
	tt.Frame()
	press(t, tt, "Load examples")
	tt.Key(0, ui.KeyRight)
	press(t, tt, "Complete tree load")
	requireTextContains(t, tt, "Examples loaded")
	requireText(t, tt, "Button example")
	press(t, tt, "Button example")
	requireTextContains(t, tt, "1 selected")
	press(t, tt, "lazy flat rows")
	requireTextContains(t, tt, "1 selected")
	requireText(t, tt, "Button example")
	a.Deactivate()
	tt.Frame()
	requireTextContains(t, tt, "1 selected")
}
func TestGalleryNumberOrientation(t *testing.T) {
	a := newStaticGallery()
	t.Cleanup(a.Dispose)
	tt := sceneTester(t, a, "input-number", "bounded")
	press(t, tt, "bounded seats")
	tt.Type("3")
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("-")
	press(t, tt, "Number orientation")
	press(t, tt, "vertical")
	requireText(t, tt, "Value: {3 true}")
	press(t, tt, "Increase bounded seats")
	requireText(t, tt, "Value: {0 true}")
	press(t, tt, "Increase bounded seats")
	requireText(t, tt, "Value: {0.5 true}")
	press(t, tt, "Number orientation")
	press(t, tt, "horizontal")
	press(t, tt, "Decrease bounded seats")
	requireText(t, tt, "Value: {0 true}")
}
func TestGallerySelectionAndColor(t *testing.T) {
	a := newStaticGallery()
	t.Cleanup(a.Dispose)
	tt := sceneTester(t, a, "select-menu", "single")
	press(t, tt, "single members menu")
	press(t, tt, "Search single members menu")
	tt.Type("Ada")
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	requireText(t, tt, "Selected IDs: [ada]")
	tt = sceneTester(t, a, "color-picker", "palette")
	press(t, tt, "palette accent A")
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("128")
	tt.Frame()
	requireText(t, tt, "Color: {34 197 94 128")
	tt = sceneTester(t, a, "input-date", "calendar")
	press(t, tt, "Choose calendar release date")
	press(t, tt, "calendar release date calendar 2026-11-15")
	requireText(t, tt, "Date: {2026-11-15 true}")
	tt.SetLocale("zh-Hans")
	requireText(t, tt, "Date: {2026-11-15 true}")
	tt.SetLocale("en")
	requireText(t, tt, "Date: {2026-11-15 true}")
}

func TestGalleryListboxEmptyAndDisabled(t *testing.T) {
	a := newStaticGallery()
	t.Cleanup(a.Dispose)
	tt := sceneTester(t, a, "listbox", "states")
	requireText(t, tt, "Empty list")
	requireText(t, tt, "No members")
	requireText(t, tt, "Disabled list")
	requireText(t, tt, "Disabled selected IDs: [lin] · 0 changes")
	press(t, tt, "Ada Lovelace")
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	tt.Key(0, ui.KeySpace)
	requireText(t, tt, "Disabled selected IDs: [lin] · 0 changes")
	if tt.Focused("Disabled members list") {
		t.Fatal("disabled list accepted focus")
	}
	tt.SetSize(720, 820)
	tt.SetPreferences(ui.Preferences{TextScale: 1.5, ReduceMotion: true})
	tt.Frame()
	requireText(t, tt, "No members")
	requireText(t, tt, "Disabled selected IDs: [lin] · 0 changes")
	captureGallery(t, tt, "listbox-empty-and-disabled")
}

func TestGalleryPopulatedReadOnlyExamples(t *testing.T) {
	a := newStaticGallery()
	t.Cleanup(a.Dispose)
	for _, sample := range []struct {
		id, label, summary string
		covered            bool // the PIN slots take the press for the named editor
	}{
		{"pin-input", "states verification code", "Read-only code: 1234 · Disabled code: 1234", true},
		{"input-date", "states release date month", "Date: {2026-11-05 true}", false},
		{"input-time", "states reminder hour", "Time: {09:30:00 true}", false},
	} {
		t.Run(sample.id, func(t *testing.T) {
			tt := sceneTester(t, a, sample.id, "states")
			requireText(t, tt, sample.summary)
			if sample.covered {
				pressOver(t, tt, sample.label)
			} else {
				press(t, tt, sample.label)
			}
			tt.Key(ui.Ctrl, ui.KeyA)
			tt.Type("99")
			tt.Key(0, ui.KeyUp)
			tt.Key(0, ui.KeyBackspace)
			tt.Frame()
			requireText(t, tt, sample.summary)
		})
	}
	tt := sceneTester(t, a, "input-menu", "states")
	requireInput(t, tt, "Read-only search", "Team")
	tt.Type("Blocked")
	tt.Key(0, ui.KeyBackspace)
	requireInput(t, tt, "Read-only search", "Team")
}

func TestGalleryEmptyExamplesRecover(t *testing.T) {
	a := newStaticGallery()
	t.Cleanup(a.Dispose)
	tt := sceneTester(t, a, "select", "empty")
	requireText(t, tt, "Choose a plan")
	press(t, tt, "Unselected plan")
	press(t, tt, "Personal")
	requireText(t, tt, "Personal")
	requireText(t, tt, "No options available")
	captureGallery(t, tt, "select-empty-and-unselected")

	tt = sceneTester(t, a, "table", "empty")
	requireText(t, tt, "No workspace records")
	press(t, tt, "Select first record")
	requireTextContains(t, tt, "0 rows · 0 selected")
	press(t, tt, "Add sample record")
	if tt.HasText("No workspace records") {
		t.Fatal("empty message survived adding a record")
	}
	requireText(t, tt, "Workspace 00001")
	press(t, tt, "Select first record")
	requireTextContains(t, tt, "1 rows · 1 selected")
	press(t, tt, "Remove sample record")
	requireText(t, tt, "No workspace records")
	requireTextContains(t, tt, "0 rows · 0 selected")
	tt.SetSize(720, 820)
	tt.SetPreferences(ui.Preferences{TextScale: 1.5, ReduceMotion: true})
	captureGallery(t, tt, "table-empty-restored")
}

func TestGalleryFormValidationCancellationAndRecovery(t *testing.T) {
	a := newStaticGallery()
	t.Cleanup(a.Dispose)
	tt := sceneTester(t, a, "form", "validation")
	press(t, tt, "Submit validation")
	requireText(t, tt, "Workspace name is required")
	press(t, tt, "validation workspace name")
	tt.Type("Native workspace")
	press(t, tt, "Submit validation")
	requireText(t, tt, "Workspace submitted · 1 successful submissions")
	tt = sceneTester(t, a, "form", "async")
	press(t, tt, "Submit async")
	requireText(t, tt, "Simulated request pending")
	a.Deactivate()
	tt.Frame()
	press(t, tt, "Complete simulated request")
	requireText(t, tt, "Stale request rejected · 0 successful submissions")
	a.WindowInactive = false
	press(t, tt, "Submit async")
	press(t, tt, "Complete simulated request")
	requireText(t, tt, "Simulated request completed · 1 successful submissions")
}
func TestGalleryRangeOrientationAndCommit(t *testing.T) {
	a := newStaticGallery()
	t.Cleanup(a.Dispose)
	tt := sceneTester(t, a, "slider", "range")
	press(t, tt, "Range orientation")
	press(t, tt, "vertical")
	focus(t, tt, "Workspace range minimum")
	tt.Key(0, ui.KeyUp)
	requireText(t, tt, "Confirmed: 25 – 70")
	press(t, tt, "Reverse range")
	focus(t, tt, "Workspace range minimum")
	tt.Key(0, ui.KeyUp)
	requireText(t, tt, "Confirmed: 20 – 70")
}
func TestNarrowSectionsDoNotOverlap(t *testing.T) {
	a := newStaticGallery()
	t.Cleanup(a.Dispose)
	a.Navigate("input")
	tt := ui.NewTester(a.View, 720, 900)
	tt.SetPreferences(ui.Preferences{TextScale: 1.5, ReduceMotion: true})
	input := galleryBounds(t, tt, "Workspace name")
	next := galleryBounds(t, tt, "Outline input")
	if next.Y < input.Y+input.H {
		t.Fatalf("overlap: %v %v", input, next)
	}
}
