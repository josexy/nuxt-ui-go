package nui_test

import (
	"testing"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

func TestEmptyResponsivePadding(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		nui.Empty(c, nui.App(c, nui.AppConfig{}), nui.EmptyOptions{Title: "Empty title"}).Label("Empty root")
	}, 500, 200)
	for _, check := range []struct {
		width   int
		padding float32
	}{{500, 16}, {800, 24}, {1100, 32}} {
		tt.SetSize(check.width, 200)
		root, title := bounds(t, tt, "Empty root"), bounds(t, tt, "Empty title")
		if got := title.Y - root.Y; got != check.padding {
			t.Fatalf("width %d: top padding %g, want %g", check.width, got, check.padding)
		}
	}
}

func TestEmptyAvatarSizeAndHeaderOverrides(t *testing.T) {
	avatar := nui.AvatarOptions{Text: "AU", Alt: "Empty avatar"}
	header, leading, body := false, false, false
	clicks := 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		opts := nui.EmptyOptions{Title: "Default title", Avatar: &avatar, Size: nui.XL,
			Actions: func() { nui.Button(c, th, "Default action", nui.ButtonOptions{}).OnClick(func() { clicks++ }) },
			Footer:  func() { ui.Text(c, "Custom footer") },
		}
		if header {
			opts.Header = func() { ui.Text(c, "Custom header") }
		}
		if leading {
			opts.Leading = func() { ui.Text(c, "Custom leading") }
		}
		if body {
			opts.Body = func() { nui.Button(c, th, "Custom action", nui.ButtonOptions{}).OnClick(func() { clicks++ }) }
		}
		nui.Empty(c, th, opts)
	}, 500, 360)
	if r := bounds(t, tt, "Empty avatar"); r.W != 48 || r.H != 48 {
		t.Fatalf("XL default avatar: %+v", r)
	}
	avatar.Size = nui.XS
	tt.Frame()
	if r := bounds(t, tt, "Empty avatar"); r.W != 24 || r.H != 24 {
		t.Fatalf("explicit avatar size ignored: %+v", r)
	}
	leading = true
	tt.Frame()
	if !tt.HasText("Custom leading") || !tt.HasText("Default title") || tt.HasText("AU") {
		t.Fatal("leading builder did not replace only the avatar")
	}
	header, body = true, true
	tt.Frame()
	if !tt.HasText("Custom header") || !tt.HasText("Custom footer") || tt.HasText("Custom leading") || tt.HasText("Default title") || tt.HasText("Default action") {
		t.Fatal("header or body builder priority")
	}
	click(t, tt, "Custom action")
	tt.Frame()
	if clicks != 1 {
		t.Fatalf("custom action called %d times", clicks)
	}
}

func TestEmptyHeaderChangeKeepsActionFocus(t *testing.T) {
	showTitle := false
	draft := "Original"
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		opts := nui.EmptyOptions{Actions: func() {
			nui.Input(c.Key("draft"), th, &draft, nui.InputOptions{Label: "Empty action draft"}).Width(180)
		}}
		if showTitle {
			opts.Title = "Empty title"
		}
		nui.Empty(c, th, opts)
	}, 400, 260)
	click(t, tt, "Empty action draft")
	tt.Key(ui.Ctrl, ui.KeyA)
	showTitle = true
	tt.Frame()
	if !tt.Focused("Empty action draft") {
		t.Fatal("header insertion discarded action focus")
	}
	tt.Type("Replacement")
	if draft != "Replacement" {
		t.Fatal("header insertion discarded the action selection")
	}
}

func TestEmptyLoadingColorUsesLeadingUI(t *testing.T) {
	color := ui.Hex("#dc2626")
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		th.Motion.Reduce = true
		nui.Empty(c, th, nui.EmptyOptions{Loading: true, LoadingIcon: lucide.Check(),
			UI: nui.Styles{nui.Leading: {Foreground: nui.Set(color)}},
			Style: func(slot nui.Slot, e ui.Element) {
				if slot == nui.Leading {
					e.Label("Loading glyph")
				}
			}})
	}, 240, 180)
	if countColor(tt.Image(), bounds(t, tt, "Loading glyph"), rgba(color)) == 0 {
		t.Fatal("loading glyph ignored Leading.Foreground")
	}
}

func TestEmptyFooterKeepsFocusWhenActionsAreAdded(t *testing.T) {
	showActions := false
	draft := "Original"
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		opts := nui.EmptyOptions{Footer: func() { nui.Input(c.Key("draft"), th, &draft, nui.InputOptions{Label: "Footer draft"}).Width(180) }}
		if showActions {
			opts.Actions = func() { nui.Button(c, th, "Action", nui.ButtonOptions{}) }
		}
		nui.Empty(c, th, opts)
	}, 400, 260)
	click(t, tt, "Footer draft")
	tt.Key(ui.Ctrl, ui.KeyA)
	showActions = true
	tt.Frame()
	if !tt.Focused("Footer draft") {
		t.Fatal("adding actions discarded footer focus")
	}
	tt.Type("Replacement")
	if draft != "Replacement" {
		t.Fatal("adding actions discarded footer selection")
	}
}
