package preview

import (
	"context"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/internal/showcase"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func submitPreviewSentinel(example Example) {
	options := []nui.Option[string]{{Value: "sentinel", Label: "Sentinel", Description: "Submitted dataset survives"}}
	items := []nui.ListboxItem{{ID: "sentinel", Label: "Sentinel", Description: "Submitted dataset survives"}}
	menus := []nui.MenuItem{{ID: "sentinel", Label: "Sentinel", Description: "Submitted dataset survives"}}
	switch s := example.(type) {
	case *selectExample:
		s.SingleOptions.SetOptions(options)
		s.Open = true
	case *listboxExample:
		s.List.SetItems(items)
	case *select_menuExample:
		s.Menu.SetItems(items)
		s.Menu.Open = true
	case *input_menuExample:
		s.Search.SetOptions(options)
		s.Open = true
	case *dropdown_menuExample:
		s.Menu.SetItems(menus)
		s.Menu.Open = true
	case *context_menuExample:
		s.Menu.SetItems(menus)
		s.Menu.Open = true
	}
}

func openPreviewChoices(example Example) {
	switch s := example.(type) {
	case *selectExample:
		s.Open, s.ManyOpen = true, false
	case *select_menuExample:
		s.Menu.Open = true
	case *input_menuExample:
		s.Open = true
	case *dropdown_menuExample:
		s.Menu.Open = true
	case *context_menuExample:
		s.Menu.Open = true
	}
}

func TestChoicePreviewsKeepSubmittedDataAcrossBuilds(t *testing.T) {
	for _, id := range []string{"select", "listbox", "select-menu", "input-menu", "dropdown-menu", "context-menu"} {
		t.Run(id, func(t *testing.T) {
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
			if err := example.Configure(cfg); err != nil {
				t.Fatal(err)
			}
			tt := ui.NewTester(func(c *ui.Context) {
				app := cfg.Theme()
				app.DisableAnimations = true
				th := nui.App(c, app)
				ui.Column(c.Key("preview")).Fill().Padding(24).Gap(16).Children(func() { example.View(c, th, cfg) })
			}, 760, demo.Height)
			submitPreviewSentinel(example)
			tt.Frame()
			if input, ok := example.(*input_menuExample); ok {
				// Replacing InputMenu data dismisses the previous popup. Reopen
				// after that update to inspect the newly submitted rows.
				input.Open = true
				tt.Frame()
			}
			if !tt.HasText("Submitted dataset survives") {
				t.Fatal("view replaced the explicitly submitted dataset")
			}
			for range 4 {
				tt.Frame()
			}
			cfg.Mode = "dark"
			cfg.Values["size"] = "lg"
			tt.Frame()
			if !tt.HasText("Submitted dataset survives") {
				t.Fatal("appearance or settled builds resubmitted example defaults")
			}
			if selectExample, ok := example.(*selectExample); ok {
				selectExample.Open, selectExample.ManyOpen = false, true
				tt.Frame()
				if tt.HasText("Submitted dataset survives") {
					t.Fatal("single and multiple preview controls shared their dataset")
				}
			}
			// A fresh instance starts from its own submitted defaults, even while
			// another window has custom state at the same component keys.
			fresh, err := New(id)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(fresh.Dispose)
			if err := fresh.Configure(cfg); err != nil {
				t.Fatal(err)
			}
			openPreviewChoices(fresh)
			other := ui.NewTester(func(c *ui.Context) {
				fresh.View(c, nui.App(c, nui.AppConfig{DisableAnimations: true}), cfg)
			}, 760, demo.Height)
			openPreviewChoices(example)
			tt.Frame()
			if !tt.HasText("Submitted dataset survives") {
				t.Fatal("creating another window replaced the first instance's dataset")
			}
			example.Dispose()
			other.Frame()
			original := map[string]string{"select": "Backend services", "listbox": "Native applications", "select-menu": "Native applications", "input-menu": "Backend services", "dropdown-menu": "Change project details", "context-menu": "Change project details"}[id]
			if other.HasText("Submitted dataset survives") || !other.HasText(original) {
				t.Fatal("preview instances shared submitted data")
			}
		})
	}
}

func TestMenuPreviewSubmitsOnlyChangedConfigurationOrActions(t *testing.T) {
	demo, _ := showcase.Find("dropdown-menu")
	cfg, _ := demo.Normalize(showcase.Config{Locale: "en"})
	example, _ := New("dropdown-menu")
	s := example.(*dropdown_menuExample)
	defer s.Dispose()
	tt := ui.NewTester(func(c *ui.Context) { s.View(c, nui.App(c, nui.AppConfig{DisableAnimations: true}), cfg) }, 760, demo.Height)
	submissions, submit := 0, s.SubmitItems
	s.SubmitItems = func() { submissions++; submit() }
	for range 4 {
		tt.Frame()
	}
	cfg.Values["size"] = "lg"
	tt.Frame()
	if submissions != 0 {
		t.Fatal("settled or size-only changes resubmitted data")
	}
	cfg.Values["avatar"] = !cfg.Bool("avatar")
	tt.Frame()
	if submissions != 1 {
		t.Fatal("descriptor configuration did not submit once")
	}
	s.Menu.Open = true
	tt.Frame()
	clickExample(t, tt, "Pinned")
	if submissions != 2 || !s.Pinned || s.CheckedChanges != 1 {
		t.Fatal("checked action did not submit its new controlled value once")
	}
	clickExample(t, tt, "Reverse actions")
	if submissions != 3 || !s.Reverse {
		t.Fatal("reorder did not submit once")
	}
	tt.Frame()
	if submissions != 3 {
		t.Fatal("reordered rows were resubmitted by the next build")
	}
}

func TestDisposedRemotePreviewCannotPublishIntoReset(t *testing.T) {
	first, _ := New("select-menu")
	old := first.(*select_menuExample)
	_, ticket := old.Results.Start(context.Background(), "old")
	first.Dispose()
	second, _ := New("select-menu")
	fresh := second.(*select_menuExample)
	defer second.Dispose()
	if old.Results.Complete(ticket, []nui.ListboxItem{{ID: "old", Label: "Stale"}}, nil) || fresh.Results.Complete(ticket, nil, nil) {
		t.Fatal("disposed or foreign request published into preview state")
	}
}

func TestChoicePreviewDisposalRetainsDataAndDraft(t *testing.T) {
	for _, id := range []string{"input-menu", "listbox", "select-menu", "dropdown-menu", "context-menu"} {
		t.Run(id, func(t *testing.T) {
			example, err := New(id)
			if err != nil {
				t.Fatal(err)
			}
			switch s := example.(type) {
			case *input_menuExample:
				s.Open, s.Search.Open, s.Search.Query = true, true, "draft"
				s.Search.SelectedValues.Add("go")
				example.Dispose()
				if s.Open || s.Search.Open || s.Search.Searching() || s.Search.Query != "draft" || !s.Search.SelectedValues.Has("go") {
					t.Fatal("InputMenu disposal left its external binding open or cleared the app model")
				}
			case *listboxExample:
				s.List.Query = "draft"
				s.Values.Add("go")
				example.Dispose()
				if s.List.Searching() || s.List.Query != "draft" || !s.Values.Has("go") {
					t.Fatal("Listbox disposal changed the app model")
				}
			case *select_menuExample:
				s.Menu.Open, s.Remote.Open, s.Menu.Query = true, true, "draft"
				s.Values.Add("go")
				example.Dispose()
				if s.Menu.Open || s.Remote.Open || s.Menu.Searching() || s.Remote.Searching() || s.Menu.Query != "draft" || !s.Values.Has("go") {
					t.Fatal("SelectMenu disposal did not close both local and remote controls")
				}
			case *dropdown_menuExample:
				s.Menu.Open, s.Menu.Query = true, "draft"
				example.Dispose()
				if s.Menu.Open || s.Menu.Searching() || s.Menu.Query != "draft" {
					t.Fatal("DropdownMenu disposal did not preserve the draft while closing")
				}
			case *context_menuExample:
				s.Menu.Open, s.Menu.Query = true, "draft"
				example.Dispose()
				if s.Menu.Open || s.Menu.Searching() || s.Menu.Query != "draft" {
					t.Fatal("ContextMenu disposal did not preserve the draft while closing")
				}
			}
		})
	}
}
