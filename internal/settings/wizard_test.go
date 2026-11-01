package settings

import (
	"testing"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

func TestWizardValidationCancelAndHiddenFields(t *testing.T) {
	a := New()
	a.Page = 3
	tt := ui.NewTester(a.View, 960, 900)
	press(t, tt, "Continue")
	if a.Wizard.Step != 0 || !tt.Focused("Project name input") {
		t.Fatal("empty project advanced")
	}
	tt.Type("Release")
	press(t, tt, "Custom project key")
	press(t, tt, "Continue")
	if a.Wizard.Step != 0 {
		t.Fatal("invalid visible key")
	}
	press(t, tt, "Custom project key")
	press(t, tt, "Continue")
	if a.Wizard.Step != 1 {
		t.Fatal("hidden field blocked step")
	}
	press(t, tt, "Create project")
	if a.Wizard.Pending {
		t.Fatal("empty schedule accepted")
	}
	a.Wizard.Start = nui.Set(nui.Date{Year: 2026, Month: 11, Day: 10})
	a.Wizard.End = nui.Set(nui.Date{Year: 2026, Month: 11, Day: 1})
	tt.Frame()
	press(t, tt, "Create project")
	if a.Wizard.Pending {
		t.Fatal("reverse schedule accepted")
	}
	a.Wizard.End = nui.Set(nui.Date{Year: 2026, Month: 11, Day: 20})
	tt.Frame()
	press(t, tt, "Create project")
	if !a.Wizard.Pending {
		t.Fatal("valid schedule rejected")
	}
	press(t, tt, "Cancel creation")
	if a.Wizard.Form.Submitting {
		t.Fatal("cancel pending")
	}
	press(t, tt, "Create project")
	press(t, tt, "Confirm project")
	if a.Wizard.Status != "Project created" {
		t.Fatal("confirm")
	}
	press(t, tt, "Back")
	if a.Wizard.Name != "Release" {
		t.Fatal("hidden model lost")
	}
}
