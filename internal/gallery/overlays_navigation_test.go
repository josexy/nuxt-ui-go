package gallery

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestGalleryNestedModalEscapeAndDraft(t *testing.T) {
	a := newStaticGallery()
	t.Cleanup(a.Dispose)
	tt := sceneTester(t, a, "modal", "nested")
	press(t, tt, "Edit workspace")
	press(t, tt, "Dialog workspace name")
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("Modal 中文 draft")
	press(t, tt, "Dialog plan")
	tt.Key(0, ui.KeyEscape)
	requireText(t, tt, "Dialog workspace name")
	tt.Key(0, ui.KeyEscape)
	if !tt.Focused("Edit workspace") {
		t.Fatal("modal did not restore trigger focus")
	}
	press(t, tt, "Edit workspace")
	requireInput(t, tt, "Dialog workspace name", "Modal 中文 draft")
	a.Deactivate()
	tt.Frame()
	if tt.HasText("Dialog workspace name") {
		t.Fatal("hidden page retained modal")
	}
	a.WindowInactive = false
	press(t, tt, "Edit workspace")
	requireInput(t, tt, "Dialog workspace name", "Modal 中文 draft")
}
func TestGalleryToastActionsAndCleanup(t *testing.T) {
	a := newStaticGallery()
	t.Cleanup(a.Dispose)
	tt := sceneTester(t, a, "toast", "actions")
	press(t, tt, "Show action toast")
	press(t, tt, "Review")
	requireText(t, tt, "Reviews: 1 · Closed: 0")
	press(t, tt, "Done")
	requireText(t, tt, "Reviews: 1 · Closed: 1")
	press(t, tt, "Show action toast")
	a.Deactivate()
	tt.Frame()
	if tt.HasText("Workspace ready") {
		t.Fatal("toast survived deactivation")
	}
}
func TestGalleryManualTabsAndAccordion(t *testing.T) {
	a := newStaticGallery()
	t.Cleanup(a.Dispose)
	tt := sceneTester(t, a, "tabs", "manual")
	press(t, tt, "Overview")
	tt.Key(0, ui.KeyRight)
	requireText(t, tt, "Selected: overview · Changes: 0")
	tt.Key(0, ui.KeyEnter)
	requireText(t, tt, "Selected: settings · Changes: 1")
	press(t, tt, "manual tab draft")
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("Persistent tab edit")
	press(t, tt, "Overview")
	press(t, tt, "Settings")
	requireInput(t, tt, "manual tab draft", "Persistent tab edit")
	tt = sceneTester(t, a, "accordion", "multiple")
	press(t, tt, "Version history")
	requireText(t, tt, "Open IDs: [access history]")
	press(t, tt, "Billing settings")
	requireText(t, tt, "Open IDs: [access history]")
}
