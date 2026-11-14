package preview

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/internal/showcase"
	nui "github.com/josexy/nuxt-ui-go/ui"
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

// clickExampleOver clicks the center of label's box for whatever covers it
// there, as a splitter's grip over its one-DIP focusable divider. Use
// clickExample wherever label itself takes the click.
func clickExampleOver(t *testing.T, tester *ui.Tester, label string) {
	t.Helper()
	r, ok := tester.Find(label)
	if !ok {
		t.Fatalf("missing %q", label)
	}
	tester.ClickAt(r.X+r.W/2, r.Y+r.H/2)
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
	tester, cfg := exampleTester(t, "form")
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
	cfg.Values["async"] = true
	tester.Frame()
	clickExample(t, tester, "Save project")
	cfg.Values["async"] = false
	tester.Frame()
	clickExample(t, tester, "Repeat submit")
	if !tester.HasText("Requests: 2 · Errors: 1 · Dirty: false") {
		t.Fatal("pending example submitted twice")
	}
	clickExample(t, tester, "Complete request")
	if !tester.HasText("Saved: Website draft") {
		t.Fatal("loading invalidated the example ticket")
	}
	cfg.Values["loadingAuto"] = false
	cfg.Values["async"] = true
	tester.Frame()
	clickExample(t, tester, "Save project")
	clickExample(t, tester, "Project name")
	tester.Type(" changed")
	clickExample(t, tester, "Complete request")
	if !tester.HasText("Stale result rejected") {
		t.Fatal("edited example accepted stale results")
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
	for _, id := range []string{"avatar", "avatar-group", "user"} {
		t.Run(id, func(t *testing.T) {
			tester, _ := exampleTester(t, id)
			if !tester.HasText("AL") {
				t.Fatal("avatar does not show readable initials")
			}
		})
	}
	t.Run("chip", func(t *testing.T) {
		tester, _ := exampleTester(t, "chip")
		host, hostOK := tester.Find("Notifications 3")
		marker, markerOK := tester.Find("3")
		if !hostOK || !markerOK || marker.X < host.X+host.W-8 {
			t.Fatalf("chip is missing its button host or trailing marker: host=%+v marker=%+v", host, marker)
		}
	})
}

func TestExampleSelectionAndDisclosure(t *testing.T) {
	t.Run("sidebar", func(t *testing.T) {
		tester, cfg := exampleTester(t, "sidebar")
		clickExample(t, tester, "Sidebar header action")
		if !tester.HasText("Header actions: 1") || !tester.HasText("Project navigation") {
			t.Fatal("sidebar header regions missing")
		}
		clickExample(t, tester, "Sidebar draft")
		tester.Key(ui.Ctrl, ui.KeyA)
		tester.Type("Preserved sidebar draft")
		cfg.Values["breakpoint"] = float64(1000)
		tester.Frame()
		if !tester.Focused("Toggle sidebar") {
			t.Fatal("entering the closed mobile layout lost editor focus")
		}
		clickExample(t, tester, "Toggle sidebar")
		clickExample(t, tester, "Projects")
		if !tester.Focused("Toggle sidebar") || !tester.HasText("Current page: projects") {
			t.Fatal("sidebar navigation did not close and return focus")
		}
		cfg.Values["breakpoint"] = float64(0)
		tester.Frame()
		clickExample(t, tester, "Sidebar draft")
		tester.Key(ui.Ctrl, ui.KeyA)
		tester.Key(ui.Ctrl, ui.KeyC)
		if tester.Clipboard() != "Preserved sidebar draft" {
			t.Fatal("sidebar breakpoint changed the draft")
		}
	})
	t.Run("header", func(t *testing.T) {
		tester, cfg := exampleTester(t, "header")
		clickExample(t, tester, "Header draft")
		tester.Type("Saved header draft")
		cfg.Values["leading"], cfg.Values["styled"] = true, true
		tester.Frame()
		if !tester.Focused("Header draft") {
			t.Fatal("header leading toggle lost action focus")
		}
		clickExample(t, tester, "Open composed header menu")
		clickExample(t, tester, "Header menu settings")
		if !tester.HasText("Header page: settings") || !tester.Focused("Open composed header menu") {
			t.Fatal("composed header menu navigation/focus")
		}
		if !tester.HasText("Header draft: Saved header draft") {
			t.Fatal("header menu discarded draft")
		}
	})
	t.Run("footer", func(t *testing.T) {
		tester, cfg := exampleTester(t, "footer")
		clickExample(t, tester, "Review")
		cfg.Values["styled"], cfg.Values["rtl"] = true, true
		tester.Frame()
		clickExample(t, tester, "Help")
		if !tester.HasText("Actions: 2 · Help selected") {
			t.Fatal("footer actions did not run")
		}
	})
	t.Run("footer-columns", func(t *testing.T) {
		tester, cfg := exampleTester(t, "footer-columns")
		clickExample(t, tester, "Examples")
		cfg.Values["cancel"], cfg.Values["styled"], cfg.Values["rtl"] = true, true, true
		tester.Frame()
		clickExample(t, tester, "Components")
		clickExample(t, tester, "Unavailable")
		if !tester.HasText("Active: examples · Selects: 2 · Navigations: 1") {
			t.Fatal("footer selection cancellation or disabled state")
		}
	})
	t.Run("marquee", func(t *testing.T) {
		tester, cfg := exampleTester(t, "marquee")
		cfg.Values["repeat"] = float64(0)
		tester.Frame()
		if tester.HasText("Release 0") {
			t.Fatal("zero copies retained content")
		}
		cfg.Values["repeat"], cfg.Values["styled"], cfg.Values["orientation"] = float64(2), true, "vertical"
		tester.Frame()
		clickExample(t, tester, "Replace content")
		clickExample(t, tester, "Pause / resume")
		if !tester.HasText("Release 1") || !tester.HasText("Paused: true · Content: 1") {
			t.Fatalf("marquee example controls did not update content/state: %v", tester.Texts())
		}
	})
	t.Run("splitter", func(t *testing.T) {
		tester, cfg := exampleTester(t, "splitter")
		cfg.Values["disabled"] = true
		tester.Frame()
		clickExample(t, tester, "Pane draft")
		tester.Key(ui.Ctrl, ui.KeyA)
		tester.Type("Editable while locked")
		if !tester.HasText("Editable while locked") {
			t.Fatal("locked divider blocked the panel editor")
		}
		cfg.Values["disabled"] = false
		tester.Frame()
		clickExampleOver(t, tester, "Panel divider")
		tester.Key(0, ui.KeyEnd)
		tester.Key(0, ui.KeyRight)
		if !tester.HasText("Value: 70 · Changes: 1") {
			t.Fatal("splitter boundary notification count")
		}
		clickExampleOver(t, tester, "Nested panel divider")
		tester.Key(0, ui.KeyDown)
		if !tester.HasText("Nested value: 60 · Changes: 1") {
			t.Fatal("nested splitter did not resize independently")
		}
		cfg.Values["orientation"], cfg.Values["rtl"], cfg.Values["styled"] = "vertical", true, true
		tester.Frame()
		clickExample(t, tester, "External maximum")
		if !tester.HasText("Editable while locked") || !tester.HasText("Value: 70 · Changes: 1") {
			t.Fatal("splitter layout/external model update")
		}
	})
	t.Run("carousel", func(t *testing.T) {
		tester, cfg := exampleTester(t, "carousel")
		cfg.Values["arrows"], cfg.Values["custom"] = true, true
		tester.Frame()
		clickExample(t, tester, "Next Slides")
		if !tester.HasText("Index: 1 · Changes: 1 · Paused: false") {
			t.Fatal("carousel navigation result missing")
		}
		clickExample(t, tester, "Remove Build")
		if !tester.Focused("Slides slides") {
			t.Fatal("carousel removal lost viewport focus")
		}
		if !tester.HasText("Index: 1 · Changes: 1 · Paused: false") {
			t.Fatal("external removal emitted navigation")
		}
		cfg.Values["orientation"], cfg.Values["rtl"], cfg.Values["avatar"] = "vertical", true, true
		tester.Frame()
		clickExample(t, tester, "Restore slides")
		clickExample(t, tester, "Previous Slides")
		if !tester.HasText("Index: 0 · Changes: 1 · Paused: false") {
			t.Fatal("bounded carousel changed index")
		}
	})
	t.Run("chart", func(t *testing.T) {
		tester, cfg := exampleTester(t, "chart")
		clickExample(t, tester, "Weekly activity")
		tester.Key(0, ui.KeyHome)
		tester.Key(0, ui.KeyEnter)
		if !tester.HasText("Activated index 0 / slice ") {
			t.Fatal("chart activation did not report the original X row")
		}
		cfg.Values["xaxis"], cfg.Values["tooltip"], cfg.Values["rtl"], cfg.Values["points"] = false, false, true, true
		tester.Frame()
		if tester.HasText("Weekday") {
			t.Fatal("hidden chart axis retained title")
		}
		cfg.Values["kind"], cfg.Values["center"] = "donut", true
		tester.Frame()
		if !tester.HasText("Total") {
			t.Fatal("chart center content missing")
		}
	})
	t.Run("file-upload", func(t *testing.T) {
		tester, cfg := exampleTester(t, "file-upload")
		if !tester.HasText("MD") {
			t.Fatal("file avatar text was replaced by the fallback icon")
		}
		await := func(label string) {
			t.Helper()
			deadline := time.Now().Add(2 * time.Second)
			for !tester.HasText(label) && time.Now().Before(deadline) {
				time.Sleep(time.Millisecond)
				tester.Frame()
			}
			if !tester.HasText(label) {
				t.Fatalf("missing selection result %s", label)
			}
		}
		clickExample(t, tester, "Project files")
		await("Changes: 1 · Rejected: 1")
		cfg.Values["position"] = "inside"
		tester.Frame()
		clickExample(t, tester, "Remove notes.txt")
		await("Changes: 2 · Rejected: 1")
		if !tester.Focused("Project files") {
			t.Fatal("file removal lost picker focus")
		}
		cfg.Values["outcome"] = "error"
		tester.Frame()
		clickExample(t, tester, "Project files")
		await("sample selection failed")
		if !tester.HasText("Changes: 2 · Rejected: 1") {
			t.Fatal("failed chooser changed queue")
		}
	})
	t.Run("table", func(t *testing.T) {
		tester, cfg := exampleTester(t, "table")
		clickExample(t, tester, "Project")
		clickExample(t, tester, "Documentation")
		tester.Key(0, ui.KeyEnter)
		if !tester.HasText("Sort changes: 1 · Selection changes: 1 · Activated: Documentation") {
			t.Fatal("table model notifications missing")
		}
		clickExample(t, tester, "Restore input order")
		if !tester.HasText("First: NuxtUI Go · Selected: 1") {
			t.Fatal("table restore lost input order or selection")
		}
		cfg.Values["loading"], cfg.Values["empty"] = true, true
		tester.Frame()
		if !tester.HasText("Loading projects") {
			t.Fatal("table loading content missing")
		}
		cfg.Values["loading"] = false
		cfg.Values["rtl"] = true
		cfg.Mode = "dark"
		tester.Frame()
		if !tester.HasText("No data") {
			t.Fatal("table empty content missing")
		}
	})
	for _, tc := range []struct{ id, click, visible string }{
		{"checkbox", "Enable notifications", "Enabled: true"},
		{"switch", "Enable notifications", "Enabled: true"},
		{"checkbox-group", "Go", "Selected: [go]"},
		{"radio-group", "Rust", "Selected: rust"},
		{"tabs", "Settings", "Project settings"},
		{"accordion", "Usage", "Create a window and pass a theme to components."},
		{"collapsible", "Advanced settings", "Open: false · Changes: 1 · Actions: 0"},
		{"navigation-menu", "Projects", "Active: projects"},
		{"tree", "src", "main.go"},
		{"pagination", "6", "Page 6"},
	} {
		t.Run(tc.id, func(t *testing.T) {
			tester, cfg := exampleTester(t, tc.id)
			clickExample(t, tester, tc.click)
			if tc.id == "tree" {
				tester.Key(0, ui.KeyRight)
				cfg.Values["protect"] = true
				tester.Frame()
				tester.Key(0, ui.KeyLeft)
				if !tester.HasText("src toggle prevented · Loads: 0") {
					t.Fatal("tree toggle prevention missing")
				}
				cfg.Values["protect"] = false
				cfg.Values["multiple"], cfg.Values["bubble"], cfg.Values["custom"] = true, true, true
				tester.Frame()
				clickExample(t, tester, "main.go")
				if !tester.HasText("Mixed") {
					t.Fatal("tree custom trailing state missing")
				}
				clickExample(t, tester, "Remote files")
				clickExample(t, tester, "Collapse remote")
				clickExample(t, tester, "Complete empty load")
				if !tester.HasText("Stale completion rejected · Loads: 1") {
					t.Fatal("collapsed tree result accepted")
				}
				clickExample(t, tester, "Remote files")
				clickExample(t, tester, "Complete empty load")
				if !tester.HasText("Remote directory is empty · Loads: 2") {
					t.Fatal("empty tree completion missing")
				}
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
		{"slideover", "Open panel", "Done", "Panel draft"},
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
			if tc.id == "drawer" {
				draft, ok := tester.Find("Drawer draft")
				if !ok {
					t.Fatal("drawer draft missing")
				}
				tester.Scroll(draft.X+draft.W/2, draft.Y+draft.H/2, 0, 1000)
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
		{"error", "Retry attempts: 0 · Recoveries: 0"},
		{"color-picker", "Accent color A"},
		{"file-upload", "design.png"},
		{"footer-columns", "Unavailable"},
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
			// "Date range start" names the group of segments; a press reaches one of them.
			clickExample(t, tester, "Date range start day")
			tester.Type("2030-11-06")
			tester.Key(ui.Ctrl, ui.KeyA)
			tester.Key(ui.Ctrl, ui.KeyC)
			if tester.Clipboard() != "" {
				t.Fatalf("%s range accepted text", control)
			}
			cfg.Values[control] = false
			tester.Frame()
			clickExample(t, tester, "Date range start day")
			tester.Type("2030-11-06")
			tester.Key(ui.Ctrl, ui.KeyA)
			tester.Key(ui.Ctrl, ui.KeyC)
			if tester.Clipboard() != "2030-11-06" {
				t.Fatal("enabled range did not accept text")
			}
			cfg.Values["showCalendar"] = true
			tester.Frame()
			if _, found := tester.Find("Choose Date range"); !found {
				t.Fatal("ShowCalendar did not enable the range trigger")
			}
			clickExample(t, tester, "Choose Date range")
			if _, found := tester.Find("Date range calendar"); !found {
				t.Fatal("range calendar did not open")
			}
			cfg.Values["showCalendar"] = false
			tester.Frame()
			if _, found := tester.Find("Choose Date range"); found {
				t.Fatal("range calendar trigger remained visible")
			}
			if _, found := tester.Find("Date range calendar"); found {
				t.Fatal("range calendar remained open after ShowCalendar was disabled")
			}
		})
	}
}
