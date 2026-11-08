package settings

import (
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestSettingsEditConfirmationUndoAndFilters(t *testing.T) {
	a := New()
	tt := ui.NewTester(a.View, 960, 900)
	press(t, tt, "Actions")
	tt.Key(0, ui.KeyEnter)
	if !a.EditOpen {
		t.Fatal("menu did not open edit dialog")
	}
	// "Profile name *" names the whole field; its visible label focuses the input.
	press(t, tt, "Profile name")
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Key(0, ui.KeyBackspace)
	press(t, tt, "Apply profile")
	if !a.EditOpen || !tt.HasText("Enter your display name.") {
		t.Fatal("invalid edit accepted")
	}
	press(t, tt, "Profile name")
	tt.Type("New profile")
	press(t, tt, "Apply profile")
	if a.Name != "New profile" || a.EditOpen || !tt.Focused("Actions") || !tt.HasText("Profile updated") {
		t.Fatal("apply, toast or menu opener focus")
	}
	press(t, tt, "Edit profile")
	press(t, tt, "Profile name")
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("Unsaved draft")
	tt.Key(0, ui.KeyEscape)
	if a.EditOpen || a.Name != "New profile" || !tt.Focused("Edit profile") {
		t.Fatal("cancel edit/focus")
	}
	a.Notify, a.Volume, a.Channels = false, 85, []string{"desktop"}
	tt.Frame()
	press(t, tt, "Actions")
	tt.Key(0, ui.KeyEnd)
	tt.Key(0, ui.KeyEnter)
	if !a.ConfirmReset {
		t.Fatal("reset confirmation missing")
	}
	press(t, tt, "Reset notifications")
	if !a.Notify || a.Volume != 40 || !tt.Focused("Actions") {
		t.Fatal("reset or focus restore")
	}
	tt.Advance(200 * time.Millisecond)
	press(t, tt, "Undo reset")
	if a.Notify || a.Volume != 85 || !slices.Equal(a.Channels, []string{"desktop"}) {
		t.Fatal("undo lost old preferences")
	}
	press(t, tt, "Notification filters")
	tt.Advance(250 * time.Millisecond)
	press(t, tt, "Deliver notifications")
	press(t, tt, "Sound")
	press(t, tt, "Filter volume")
	tt.Key(0, ui.KeyHome)
	press(t, tt, "Advanced delivery")
	tt.Advance(250 * time.Millisecond)
	press(t, tt, "Done filtering")
	tt.Advance(250 * time.Millisecond)
	if a.FiltersOpen || !a.Notify || a.Volume != 0 || !tt.Focused("Notification filters") {
		t.Fatalf("filters open=%v notify=%v volume=%v focus=%v", a.FiltersOpen, a.Notify, a.Volume, tt.Focused("Notification filters"))
	}
}

func TestSettingsNarrowOverlays(t *testing.T) {
	a := New()
	a.Name = "很长的显示名称 / A long profile name for a narrow window"
	tt := ui.NewTester(a.View, 720, 900)
	tt.SetPreferences(ui.Preferences{TextScale: 1.5, ReduceMotion: true})
	press(t, tt, "Edit profile")
	for _, label := range []string{"Close Edit profile details", "Profile name *", "Apply profile"} {
		r, ok := tt.Find(label)
		if !ok || r.X < 0 || r.Y < 0 || r.X+r.W > 720 || r.Y+r.H > 900 {
			t.Fatalf("%s outside narrow modal: %v", label, r)
		}
	}
	press(t, tt, "Apply profile")
	press(t, tt, "Notification filters")
	for _, label := range []string{"Deliver notifications", "Done filtering"} {
		r, ok := tt.Find(label)
		if !ok || r.X < 0 || r.Y < 0 || r.X+r.W > 720 || r.Y+r.H > 900 {
			t.Fatalf("%s outside narrow side panel: %v", label, r)
		}
	}
	// The preceding save toast stays below the modal layer and cannot swallow this click.
	press(t, tt, "Done filtering")
	if a.FiltersOpen {
		t.Fatal("toast or layout blocked the side-panel footer")
	}
}
