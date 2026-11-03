package preview

import (
	"testing"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/internal/showcase"
)

func TestEveryExampleBuildsAndRebuilds(t *testing.T) {
	for _, demo := range showcase.Catalog() {
		t.Run(demo.ID, func(t *testing.T) {
			example, err := New(demo.ID)
			if err != nil {
				t.Fatal(err)
			}
			defer example.Dispose()
			cfg, err := demo.Normalize(showcase.Config{})
			if err != nil {
				t.Fatal(err)
			}
			tt := ui.NewTester(func(c *ui.Context) {
				theme := cfg.Theme()
				theme.Light.Motion.Reduce = true
				theme.Dark.Motion.Reduce = true
				t := nui.App(c, theme)
				ui.Column(c.Key("preview")).Fill().Padding(24).Gap(16).Children(func() { example.View(c, t, cfg) })
			}, 760, demo.Height)
			tt.Frame()
			cfg.Mode = "dark"
			tt.Frame()
			tt.Frame()
			// The panel must never offer a value that the real component rejects.
			// Exercise live updates on the same instance, including explicit zero
			// and optional false values, rather than only parsing generated code.
			for _, control := range demo.Controls {
				values := []any{control.Default}
				switch control.Kind {
				case "select":
					for _, choice := range control.Choices {
						values = append(values, choice.Value)
					}
				case "boolean":
					values = append(values, true, false)
				case "optionalBoolean":
					values = append(values, nil, true, false)
				case "number", "integer":
					values = append(values, control.Min, control.Max)
				case "text":
					values = append(values, "", "中文 \"quoted\"")
				}
				for _, value := range values {
					t.Logf("%s=%v", control.Key, value)
					cfg.Values[control.Key] = value
					tt.Frame()
					_ = tt.Image()
				}
				cfg.Values[control.Key] = control.Default
			}
			if tt.Image().Bounds().Empty() {
				t.Fatal("empty render")
			}
		})
	}
}

func TestParameterUpdatesPreserveInputAndActions(t *testing.T) {
	demo, _ := showcase.Find("input")
	cfg, _ := demo.Normalize(showcase.Config{})
	example, _ := New("input")
	defer example.Dispose()
	tt := ui.NewTester(func(c *ui.Context) {
		theme := nui.App(c, cfg.Theme())
		ui.Column(c).Fill().Padding(24).Children(func() { example.View(c, theme, cfg) })
	}, 760, 360)
	if err := tt.Click("Project name"); err != nil {
		t.Fatal(err)
	}
	tt.Type("Draft stays")
	cfg.Values["size"] = "lg"
	cfg.Values["variant"] = "soft"
	cfg.Mode = "dark"
	tt.Frame()
	if _, ok := tt.Find("Value: Draft stays"); !ok {
		t.Fatal("parameter update discarded input")
	}
	tt.Type("!")
	if _, ok := tt.Find("Value: Draft stays!"); !ok {
		t.Fatal("input lost focus after configuration")
	}

	button, _ := showcase.Find("button")
	cfg, _ = button.Normalize(showcase.Config{})
	buttons, _ := New("button")
	defer buttons.Dispose()
	bt := ui.NewTester(func(c *ui.Context) {
		theme := nui.App(c, cfg.Theme())
		ui.Column(c).Fill().Padding(24).Children(func() { buttons.View(c, theme, cfg) })
	}, 760, 360)
	if err := bt.Click("Save changes"); err != nil {
		t.Fatal(err)
	}
	cfg.Values["variant"] = "outline"
	bt.Frame()
	bt.Frame()
	if _, ok := bt.Find("Clicked 1 times"); !ok {
		t.Fatal("callback count changed during rebuild")
	}
	cfg.Values["disabled"] = true
	bt.Frame()
	if err := bt.Click("Save changes"); err != nil {
		t.Fatal(err)
	}
	if _, ok := bt.Find("Clicked 1 times"); !ok {
		t.Fatal("disabled button activated")
	}
}

func TestDialogKeyboardAndRemount(t *testing.T) {
	demo, _ := showcase.Find("modal")
	cfg, _ := demo.Normalize(showcase.Config{})
	example, _ := New("modal")
	defer func() { example.Dispose() }()
	tt := ui.NewTester(func(c *ui.Context) {
		theme := nui.App(c, cfg.Theme())
		ui.Column(c).Fill().Padding(24).Children(func() { example.View(c, theme, cfg) })
	}, 760, 480)
	if err := tt.Click("Open dialog"); err != nil {
		t.Fatal(err)
	}
	if _, ok := tt.Find("Project name"); !ok {
		t.Fatal("dialog did not open")
	}
	tt.Key(0, ui.KeyEscape)
	if _, ok := tt.Find("Project name"); ok {
		t.Fatal("Escape did not dismiss dialog")
	}
	example.Dispose()
	example, _ = New("modal")
	tt.Frame()
	if err := tt.Click("Open dialog"); err != nil {
		t.Fatal(err)
	}
	if _, ok := tt.Find("Project name"); !ok {
		t.Fatal("remounted dialog failed")
	}
}

func TestPopoverClickDismissAndRemount(t *testing.T) {
	demo, _ := showcase.Find("popover")
	cfg, _ := demo.Normalize(showcase.Config{})
	example, _ := New("popover")
	defer func() { example.Dispose() }()
	tt := ui.NewTester(func(c *ui.Context) {
		app := cfg.Theme()
		app.Light.Motion.Reduce = true
		app.Dark.Motion.Reduce = true
		theme := nui.App(c, app)
		ui.Column(c.Key("preview")).Fill().Padding(24).Gap(16).Children(func() { example.View(c, theme, cfg) })
	}, 760, demo.Height)
	visible := func(want bool) {
		t.Helper()
		tt.Frame()
		if _, found := tt.Find("Additional information"); found != want {
			t.Fatalf("popover visible = %v, want %v", found, want)
		}
	}
	click := func(label string) {
		t.Helper()
		if err := tt.Click(label); err != nil {
			t.Fatal(err)
		}
	}
	visible(false)
	click("Open details")
	visible(true)
	cfg.Values["dismissible"] = false
	tt.Frame()
	tt.Key(0, ui.KeyEscape)
	visible(true)
	click("Close")
	visible(false)
	cfg.Values["dismissible"] = nil
	tt.Frame()
	tt.Key(0, ui.KeyEnter)
	visible(true)
	tt.Key(0, ui.KeyEscape)
	visible(false)
	click("Open details")
	visible(true)
	tt.ClickAt(720, 380)
	visible(false)
	example.Dispose()
	example, _ = New("popover")
	tt.Frame()
	click("Open details")
	visible(true)
}
