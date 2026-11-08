package gallery

import (
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
	"github.com/josexy/nuxt-ui-go/internal/databrowser"
	"github.com/josexy/nuxt-ui-go/internal/settings"
)

func (a *State) View(c *ui.Context) {
	t := nui.App(c, a.themeConfig())
	if a.closeToasts {
		nui.CloseToast(c, "")
		a.closeToasts = false
	}
	for _, notice := range a.pendingToasts {
		nui.Toast(c, t, notice)
	}
	a.pendingToasts = nil
	nui.ToastViewport(c, t, a.toastViewport)
	ui.Column(c).Fill().Children(func() {
		ui.Row(c.Key("gallery-toolbar")).FillWidth().Padding(12, 20).Gap(10).Wrap().AlignItems(ui.Center).BorderWidth(0, 0, 1, 0).BorderColor(t.Border).Children(func() {
			width, _ := c.Size()
			if width < 880 {
				nui.Button(c.Key("directory-toggle"), t, "", nui.ButtonOptions{AccessibleLabel: "Open component directory", Leading: lucide.Menu(), Square: true, Variant: nui.Ghost, Color: nui.Neutral}).Bind(&a.directoryToggle).OnClick(func() { a.Sidebar.Open = !a.Sidebar.Open })
			}
			ui.Text(c, "NuxtUI Go").FontSize(18).FontWeight(600).TextColor(t.Highlighted)
			ui.Text(c, "Gallery").TextColor(t.Muted).Grow(1)
			a.themeMenu(c, t)
			if a.NewWindow != nil {
				nui.Button(c.Key("new-window"), t, "New window", nui.ButtonOptions{Variant: nui.Outline, Color: nui.Neutral}).OnClick(a.NewWindow)
			}
		})
		ui.Row(c).Grow(1).MinWidth(0).MinHeight(0).AlignItems(ui.Stretch).Children(func() {
			a.componentSidebar(c, t)
			ui.Scroll(c.Key("component/" + a.ActiveID)).Grow(1).MinWidth(0).MinHeight(0).Children(func() {
				nui.Container(c, t, nui.ContainerOptions{MaxWidth: nui.Set(float32(1120)), Padding: nui.Set(float32(28))}, func() {
					info, _ := findPage(a.ActiveID)
					heading(c, t, info.Name, info.Description.EN)
					for _, sample := range a.page(a.ActiveID).Scenes {
						ui.Column(c.Key(sample.ID)).FillWidth().MinWidth(0).Gap(10).Padding(24, 0, 8, 0).Children(func() {
							ui.Text(c, sample.Title).FontSize(18).FontWeight(600).TextColor(t.Highlighted)
							ui.Text(c, sample.Description).TextColor(t.Muted)
							ui.Column(c.Key("preview")).FillWidth().MinWidth(0).Gap(16).Padding(20).Border(1, t.Border).Radius(t.Radius).Children(func() { sample.Build(c, t) })
						})
					}
				})
			})
		})
	})
}

func (a *State) componentSidebar(c *ui.Context, t nui.Theme) {
	width, _ := c.Size()
	items := []nui.NavigationItem{}
	lastGroup := ""
	for _, p := range componentDirectory {
		if !directoryMatches(p, a.DirectorySearch) {
			continue
		}
		group := ""
		if lastGroup != p.Group {
			group, lastGroup = p.Group, p.Group
		}
		items = append(items, nui.NavigationItem{ID: p.ID, Label: p.Name, Group: group, Icon: lucide.Component()})
	}
	a.navigationActive = a.ActiveID
	nui.Sidebar(c.Key("gallery-index"), t, &a.SidebarCollapsed, nui.SidebarOptions{
		Title: "Components", State: &a.Sidebar, Resizable: true,
		MinWidth: nui.Set(float32(200)), MaxWidth: nui.Set(float32(360)), Breakpoint: nui.Set(float32(880)),
		Toggle: nui.Set(width >= 880), NavigationKey: a.ActiveID, ReturnFocus: &a.directoryToggle,
		ToggleLabel: "Toggle gallery sidebar", ResizeLabel: "Resize gallery sidebar",
		Header: func(c *ui.Context, t nui.Theme) {
			if !a.SidebarCollapsed || width < 880 {
				nui.Input(c.Key("directory-search"), t, &a.DirectorySearch, nui.InputOptions{Label: "Search components", Placeholder: "Search components…", Leading: lucide.Search(), Clearable: true}).FillWidth().Bind(&a.directorySearch)
			}
		},
		FooterContent: func(c *ui.Context, t nui.Theme, scope nui.SidebarScope) {
			if !scope.Collapsed {
				ui.Text(c, "MyGo native components").FontSize(12).TextColor(t.Muted)
			}
		},
	}, func(c *ui.Context, t nui.Theme, compact bool) {
		if len(items) == 0 {
			ui.Text(c, "No matching components. Clear the search to browse all.").TextColor(t.Muted)
			return
		}
		nui.NavigationMenu(c.Key("component-directory"), t, &a.navigationActive, &a.Navigation, items, nui.NavigationOptions{Label: "Component directory", Vertical: true, Compact: compact, OnNavigate: func(id string) { a.Navigate(id) }})
	}).Label("Gallery sidebar")
}

func settingsPage() *componentPage {
	s := settings.New()
	return &componentPage{Deactivate: []func(){func() {
		s.EditOpen = false
		s.ConfirmReset = false
		s.FiltersOpen = false
		s.Advanced = false
		s.ActionsMenu = nui.MenuState{}
		s.Wizard.Form.Cancel()
		s.Wizard.Pending = false
		s.Wizard.StartInput.Open = false
		s.Wizard.EndInput.Open = false
	}}, Scenes: []exampleScene{scene("settings", "Application settings", "Edit preferences or open project setup to validate a form.", func(c *ui.Context, t nui.Theme) { s.Panel(c, t).Height(700) })}}
}

func taskBrowserPage() *componentPage {
	s := databrowser.New(10000)
	return &componentPage{Deactivate: []func(){func() { s.Filters = false; s.Editing = false; s.Tree.Close() }}, Scenes: []exampleScene{scene("tasks", "Task browser", "Filter, sort and edit tasks. Selection follows the task ID.", func(c *ui.Context, t nui.Theme) { s.Panel(c, t).Height(550) })}}
}
