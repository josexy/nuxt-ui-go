package gallery

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestGalleryPopoverFocusAndIndependentScroll(t *testing.T) {
	a := newStaticGallery()
	t.Cleanup(a.Dispose)
	tt := sceneTester(t, a, "popover", "click")
	press(t, tt, "Open click popover")
	press(t, tt, "click popover plan")
	tt.Key(0, ui.KeyEscape)
	requireText(t, tt, "Close click popover")
	tt.Key(0, ui.KeyEscape)
	if !tt.Focused("Open click popover") {
		t.Fatal("popover failed to restore focus")
	}
	a.Navigate("scroll-area")
	tt = ui.NewTester(a.View, 1080, 900)
	firstLabel := "Long vertical list · Row 03 / 独立滚动区域"
	otherLabel := "Wide scrollbar · Row 01 / 独立滚动区域"
	first, other := galleryBounds(t, tt, firstLabel), galleryBounds(t, tt, otherLabel)
	tt.Scroll(first.X+30, first.Y+10, 0, 20)
	tt.Frame()
	moved, unchanged := galleryBounds(t, tt, firstLabel), galleryBounds(t, tt, otherLabel)
	if moved.Y >= first.Y || unchanged.Y != other.Y {
		t.Fatalf("scroll leaked between models: %v -> %v; %v -> %v", first, moved, other, unchanged)
	}
}

func TestGalleryButtonActivationAndBlockedCallbacks(t *testing.T) {
	a := newStaticGallery()
	t.Cleanup(a.Dispose)
	tt := sceneTester(t, a, "button", "usage")
	press(t, tt, "Save changes")
	tt.Key(0, ui.KeyEnter)
	tt.Key(0, ui.KeySpace)
	requireText(t, tt, "Saved 3 times")
	a.Dark = true
	tt.Frame()
	requireText(t, tt, "Saved 3 times")
	tt = sceneTester(t, a, "button", "states")
	press(t, tt, "Disabled")
	tt.Key(0, ui.KeyEnter)
	tt.Key(0, ui.KeySpace)
	press(t, tt, "Loading")
	tt.Key(0, ui.KeyEnter)
	tt.Key(0, ui.KeySpace)
	requireText(t, tt, "Blocked callbacks: 0")
}
func TestGalleryInheritedFieldsAndDisabledChoices(t *testing.T) {
	a := newStaticGallery()
	t.Cleanup(a.Dispose)
	tt := sceneTester(t, a, "form-field", "inheritance")
	press(t, tt, "Locked team")
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("Blocked edit")
	requireText(t, tt, "Value: Locked workspace")
	tt = sceneTester(t, a, "field-group", "inheritance")
	press(t, tt, "Validate inheritance")
	requireText(t, tt, "Not submitted")
	press(t, tt, "Large disabled group name")
	tt.Type("Blocked")
	requireText(t, tt, "Value: Workspace")
	tt = sceneTester(t, a, "checkbox-group", "selection")
	press(t, tt, "Test")
	press(t, tt, "Deploy")
	press(t, tt, "Preview")
	requireText(t, tt, "Selected IDs: [build test] · Changes: 1")
	press(t, tt, "Build")
	press(t, tt, "Preview")
	requireText(t, tt, "Selected IDs: [test preview] · Changes: 3")
	tt = sceneTester(t, a, "radio-group", "selection")
	press(t, tt, "Test")
	requireText(t, tt, "Selected ID: test · Changes: 1")
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeySpace)
	requireText(t, tt, "Selected ID: build · Changes: 2")
}
func TestGalleryControlledInputsAndIsolation(t *testing.T) {
	a := newStaticGallery()
	t.Cleanup(a.Dispose)
	tt := sceneTester(t, a, "checkbox", "boolean")
	press(t, tt, "Accept terms")
	requireText(t, tt, "Checked: true · Changes: 1")
	tt.Key(0, ui.KeySpace)
	requireText(t, tt, "Checked: false · Changes: 2")
	tt = sceneTester(t, a, "switch", "controlled")
	press(t, tt, "Email notifications")
	requireText(t, tt, "Notifications: false · Changes: 1")
	tt = sceneTester(t, a, "switch", "states")
	press(t, tt, "Loading switch")
	tt.Key(0, ui.KeySpace)
	press(t, tt, "Disabled switch")
	requireText(t, tt, "Disabled: true · Loading: false")
	tt = sceneTester(t, a, "select-menu", "single")
	press(t, tt, "single members menu")
	press(t, tt, "Search single members menu")
	tt.Type("Ada")
	press(t, tt, "Ada Lovelace")
	requireText(t, tt, "Selected IDs: [ada]")
	press(t, tt, "single members menu")
	requireText(t, tt, "Grace Hopper")
	tt.Key(0, ui.KeyEscape)
	tt = sceneTester(t, a, "input", "usage")
	press(t, tt, "Workspace name")
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("Independent 中文 draft")
	tt = sceneTester(t, a, "input", "variants")
	if tt.HasText("Independent 中文 draft") {
		t.Fatal("scenes share draft")
	}
	tt = sceneTester(t, a, "input", "usage")
	requireText(t, tt, "Current value: Independent 中文 draft")
}
