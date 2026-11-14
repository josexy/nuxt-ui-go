package gallery

import (
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/internal/showcase"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func TestComponentDirectoryMatchesCatalog(t *testing.T) {
	seen := map[string]bool{}
	for _, p := range componentDirectory {
		if p.ID == "" || seen[p.ID] {
			t.Fatalf("duplicate or empty ID %q", p.ID)
		}
		seen[p.ID] = true
	}
	for _, d := range showcase.Catalog() {
		p, ok := findPage(d.ID)
		if !ok || p.Name != d.Name || p.Description != d.Description || p.Note != d.Note || p.Group == "" {
			t.Fatalf("metadata mismatch: %s", d.ID)
		}
	}
	if len(seen) != 76 {
		t.Fatalf("directory: %d", len(seen))
	}
	a := newStaticGallery()
	t.Cleanup(a.Dispose)
	if a.Navigate("unknown") || a.ActiveID != "button" || len(a.pages) != 0 {
		t.Fatal("unknown navigation mutated the application")
	}
	for _, query := range []string{"BUTTON", " button ", "加载"} {
		p, _ := findPage("button")
		if !directoryMatches(p, query) {
			t.Errorf("no match for %q", query)
		}
	}
}

func TestComponentScenesBuildAndRebuild(t *testing.T) {
	a := newStaticGallery()
	t.Cleanup(a.Dispose)
	for _, info := range componentDirectory {
		if strings.HasPrefix(info.ID, "example/") {
			continue
		}
		t.Run(info.ID, func(t *testing.T) {
			p := a.page(info.ID)
			if len(p.Scenes) < 3 {
				t.Fatalf("%s has %d scenes", info.ID, len(p.Scenes))
			}
			seen := map[string]bool{}
			for _, s := range p.Scenes {
				if s.ID == "" || seen[s.ID] || s.Title == "" || s.Description == "" || s.Build == nil {
					t.Fatalf("invalid scene: %+v", s)
				}
				seen[s.ID] = true
				t.Run(s.ID, func(t *testing.T) {
					tt := ui.NewTester(func(c *ui.Context) {
						theme := nui.App(c, a.themeConfig())
						ui.Scroll(c.Key(info.ID + "/" + s.ID)).Fill().Padding(20).Gap(12).Children(func() { s.Build(c, theme) })
					}, 1080, 820)
					tt.Frame()
					if im := tt.Image(); im.Bounds().Empty() {
						t.Fatal("empty rendering")
					}
				})
			}
			p.stop()
		})
	}
}

func TestPageCacheDraftAndWindowIsolation(t *testing.T) {
	a, b := newStaticGallery(), newStaticGallery()
	t.Cleanup(a.Dispose)
	t.Cleanup(b.Dispose)
	tt := ui.NewTester(a.View, 1080, 820)
	if len(a.pages) != 1 || a.ActiveID != "button" {
		t.Fatalf("eager pages: %d", len(a.pages))
	}
	a.Navigate("input")
	tt.Frame()
	first := a.page("input")
	press(t, tt, "Workspace name")
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("中文 workspace draft")
	a.Navigate("button")
	tt.Frame()
	a.Navigate("input")
	tt.Frame()
	if first != a.page("input") || !tt.HasText("Current value: 中文 workspace draft") {
		t.Fatal("return lost draft or recreated page")
	}
	ui.NewTester(b.View, 1080, 820)
	if b.ActiveID != "button" || a.page("button") == b.page("button") {
		t.Fatal("windows share page state")
	}
	a.Dispose()
	a.Dispose()
	if a.Navigate("card") {
		t.Fatal("disposed navigation accepted")
	}
}

func TestFormExampleRejectsCancelledCompletions(t *testing.T) {
	s := &formExample{}
	s.Form.Register("name", "Native workspace")
	s.Ticket = s.Form.BeginAsync()
	s.Form.Cancel()
	s.finish()
	if s.Success != 0 || s.Result != "Stale request rejected" {
		t.Fatal("cancelled form completion published")
	}
	s.Ticket = s.Form.BeginAsync()
	s.finish()
	if s.Success != 1 {
		t.Fatal("new request could not complete")
	}
}
