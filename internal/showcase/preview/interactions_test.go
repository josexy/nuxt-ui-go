package preview

import (
	"testing"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/internal/showcase"
)

func exampleTester(t *testing.T, id string) (*ui.Tester, *showcase.Config) {
	t.Helper()
	demo, _ := showcase.Find(id)
	cfg, err := demo.Normalize(showcase.Config{Locale: "en"})
	if err != nil {
		t.Fatal(err)
	}
	example, err := New(id)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(example.Dispose)
	tester := ui.NewTester(func(c *ui.Context) {
		app := cfg.Theme()
		app.Light.Motion.Reduce, app.Dark.Motion.Reduce = true, true
		theme := nui.App(c, app)
		ui.Column(c.Key("preview")).Fill().Padding(24).Gap(16).Children(func() { example.View(c, theme, cfg) })
	}, 760, demo.Height)
	return tester, &cfg
}

func clickExample(t *testing.T, tester *ui.Tester, label string) {
	t.Helper()
	if err := tester.Click(label); err != nil {
		t.Fatal(err)
	}
}

func TestExampleMenusReportEveryAction(t *testing.T) {
	for _, id := range []string{"dropdown-menu", "context-menu"} {
		t.Run(id, func(t *testing.T) {
			tester, cfg := exampleTester(t, id)
			trigger := cfg.String("label")
			for _, action := range []string{"Edit", "Duplicate", "Archive"} {
				if id == "context-menu" {
					if err := tester.RightClick(trigger); err != nil {
						t.Fatal(err)
					}
				} else {
					clickExample(t, tester, trigger)
				}
				clickExample(t, tester, action)
				if !tester.HasText(action) {
					t.Fatalf("%s did not report its action", action)
				}
				if tester.HasText("Project actions") {
					t.Fatal("menu stayed open after selection")
				}
			}
		})
	}
}

func TestExampleFormValidationAndSubmit(t *testing.T) {
	tester, _ := exampleTester(t, "form")
	clickExample(t, tester, "Save project")
	if !tester.HasText("Enter a project name") {
		t.Fatal("empty submission did not validate")
	}
	tester.Type("Website draft")
	clickExample(t, tester, "Save project")
	if !tester.HasText("Saved: Website draft") {
		t.Fatal("form did not submit the edited value")
	}
	tester.Frame()
	if tester.HasText("Enter a project name") {
		t.Fatal("validation error survived successful submit")
	}
}

func TestExampleDisplayGeometry(t *testing.T) {
	t.Run("progress", func(t *testing.T) {
		tester, cfg := exampleTester(t, "progress")
		bounds, ok := tester.Find("Progress")
		if !ok || bounds.H > 16 {
			t.Fatalf("horizontal progress ignores size: %+v", bounds)
		}
		cfg.Values["orientation"] = "vertical"
		tester.Frame()
		bounds, ok = tester.Find("Progress")
		if !ok || bounds.W > 16 || bounds.H < 100 {
			t.Fatalf("vertical progress has incorrect geometry: %+v", bounds)
		}
	})
	for _, id := range []string{"avatar", "avatar-group", "chip", "user"} {
		t.Run(id, func(t *testing.T) {
			tester, _ := exampleTester(t, id)
			if !tester.HasText("AL") {
				t.Fatal("avatar does not show readable initials")
			}
		})
	}
}

func TestExampleSelectionAndDisclosure(t *testing.T) {
	for _, tc := range []struct{ id, click, visible string }{
		{"checkbox", "Enable notifications", "Enabled: true"},
		{"switch", "Enable notifications", "Enabled: true"},
		{"checkbox-group", "Go", "Selected: [go]"},
		{"radio-group", "Rust", "Selected: rust"},
		{"tabs", "Settings", "Project settings"},
		{"accordion", "Usage", "Create a window and pass a theme to components."},
		{"collapsible", "Advanced settings", "The hidden section is now expanded."},
		{"navigation-menu", "Projects", "Active: projects"},
		{"tree", "src", "main.go"},
		{"pagination", "6", "Page 6"},
	} {
		t.Run(tc.id, func(t *testing.T) {
			tester, cfg := exampleTester(t, tc.id)
			clickExample(t, tester, tc.click)
			if tc.id == "tree" {
				tester.Key(0, ui.KeyRight)
			}
			if !tester.HasText(tc.visible) {
				t.Fatalf("interaction did not show %q; texts: %v", tc.visible, tester.Texts())
			}
			cfg.Mode = "dark"
			tester.Frame()
			if !tester.HasText(tc.visible) {
				t.Fatal("theme update lost interactive state")
			}
		})
	}
}

func TestExampleOverlaysCloseAndReopen(t *testing.T) {
	for _, tc := range []struct{ id, open, close, content string }{
		{"modal", "Open dialog", "Save", "Project name"},
		{"slideover", "Open panel", "Done", "Keep the main view in context."},
		{"drawer", "Open drawer", "Done", "Drag the bottom drawer handle."},
	} {
		t.Run(tc.id, func(t *testing.T) {
			tester, cfg := exampleTester(t, tc.id)
			clickExample(t, tester, tc.open)
			if !tester.HasText(tc.content) {
				t.Fatal("overlay did not open")
			}
			tester.Key(0, ui.KeyEscape)
			if tester.HasText(tc.content) {
				t.Fatal("Escape did not close overlay")
			}
			cfg.Values["dismissible"] = false
			tester.Frame()
			clickExample(t, tester, tc.open)
			tester.Key(0, ui.KeyEscape)
			if !tester.HasText(tc.content) {
				t.Fatal("non-dismissible overlay closed")
			}
			clickExample(t, tester, tc.close)
			if tester.HasText(tc.content) {
				t.Fatal("explicit close did not work")
			}
			clickExample(t, tester, tc.open)
			if !tester.HasText(tc.content) {
				t.Fatal("overlay failed to reopen")
			}
		})
	}
}

func TestExampleMobileContentIsReachable(t *testing.T) {
	for _, tc := range []struct{ id, content string }{
		{"error", "Retry attempts: 0"},
		{"color-picker", "Accent color A"},
		{"file-upload", "design.png"},
		{"footer-columns", "Source"},
		{"checkbox", "Updates are saved locally in this example."},
		{"form-field", "Use a name your team will recognize."},
	} {
		t.Run(tc.id, func(t *testing.T) {
			tester, _ := exampleTester(t, tc.id)
			demo, _ := showcase.Find(tc.id)
			tester.SetSize(352, demo.Height)
			bounds, ok := tester.Find(tc.content)
			if !ok || bounds.H <= 0 || bounds.X+bounds.W > 340 || bounds.Y+bounds.H > float32(demo.Height)-12 {
				t.Fatalf("mobile content clipped: %s %+v", tc.content, bounds)
			}
		})
	}
}

func TestExampleDateRangeRespectsQuickControls(t *testing.T) {
	for _, control := range []string{"disabled", "readOnly"} {
		t.Run(control, func(t *testing.T) {
			tester, cfg := exampleTester(t, "input-date")
			cfg.Values[control] = true
			tester.Frame()
			clickExample(t, tester, "Date range start")
			tester.Type("2030-11-06")
			tester.Key(ui.Ctrl, ui.KeyA)
			tester.Key(ui.Ctrl, ui.KeyC)
			if tester.Clipboard() != "" {
				t.Fatalf("%s range accepted text", control)
			}
			cfg.Values[control] = false
			tester.Frame()
			clickExample(t, tester, "Date range start")
			tester.Type("2030-11-06")
			tester.Key(ui.Ctrl, ui.KeyA)
			tester.Key(ui.Ctrl, ui.KeyC)
			if tester.Clipboard() != "2030-11-06" {
				t.Fatal("enabled range did not accept text")
			}
			cfg.Values["hideCalendar"] = true
			tester.Frame()
			if tester.HasText("Choose Date range") {
				t.Fatal("range calendar trigger remained visible")
			}
		})
	}
}
