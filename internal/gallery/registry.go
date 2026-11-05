package gallery

import (
	"strings"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/internal/showcase"
)

type pageInfo struct {
	ID, Name, Group   string
	Description, Note showcase.Text
}

// Read the embedded source once. Gallery owns its native scenarios, while the
// website remains the source of component identity and introductory copy.
var componentDirectory = func() []pageInfo {
	groups := map[string]string{"foundations": "Foundations", "inputs": "Inputs", "overlays": "Overlays & navigation", "data": "Forms & data", "advanced": "Advanced & layout"}
	var pages []pageInfo
	for _, d := range showcase.Catalog() {
		pages = append(pages, pageInfo{d.ID, d.Name, groups[d.Category], d.Description, d.Note})
	}
	return append(pages,
		pageInfo{ID: "example/workspace", Name: "Workspace", Group: "Composition examples", Description: showcase.Text{EN: "Save a workspace and manage its documents.", ZH: "保存工作区并管理文档。"}},
		pageInfo{ID: "example/settings", Name: "Settings", Group: "Composition examples", Description: showcase.Text{EN: "Edit application settings and validate a project.", ZH: "编辑应用设置并验证项目。"}},
		pageInfo{ID: "example/task-browser", Name: "Task browser", Group: "Composition examples", Description: showcase.Text{EN: "Filter, sort and edit tasks by stable identity.", ZH: "按稳定标识筛选、排序和编辑任务。"}},
	)
}()

func findPage(id string) (pageInfo, bool) {
	for _, p := range componentDirectory {
		if p.ID == id {
			return p, true
		}
	}
	return pageInfo{}, false
}

func directoryMatches(p pageInfo, query string) bool {
	query = strings.ToLower(strings.TrimSpace(query))
	return strings.Contains(strings.ToLower(p.ID+" "+p.Name+" "+p.Description.EN+" "+p.Description.ZH), query)
}

type exampleScene struct {
	ID, Title, Description string
	Build                  func(*ui.Context, nui.Theme)
}

type componentPage struct {
	Scenes     []exampleScene
	Deactivate []func()
}

func (p *componentPage) stop() {
	for _, stop := range p.Deactivate {
		stop()
	}
}

func (a *State) page(id string) *componentPage {
	if p := a.pages[id]; p != nil {
		return p
	}
	p := a.createPage(id)
	a.pages[id] = p
	return p
}

// Navigate is the only route mutation used by the shell, commands and examples.
// Unknown IDs leave both the current page and its draft untouched.
func (a *State) Navigate(id string) bool {
	if _, ok := findPage(id); !ok || a.disposed {
		return false
	}
	if id != a.ActiveID {
		if p := a.pages[a.ActiveID]; p != nil {
			p.stop()
		}
		a.closeToasts = true
		a.pendingToasts = nil
		a.ActiveID = id
	}
	return true
}

// Deactivate preserves drafts while ending work owned by the visible page.
func (a *State) Deactivate() {
	a.WindowInactive = true
	if p := a.pages[a.ActiveID]; p != nil {
		p.stop()
	}
	a.Sidebar.Open = false
	a.ThemeMenu = nui.MenuState{}
	a.PrimaryMenu = nui.MenuState{}
	a.NeutralMenu = nui.MenuState{}
	a.closeToasts = true
	a.pendingToasts = nil
}

func scene(id, title, description string, build func(*ui.Context, nui.Theme)) exampleScene {
	return exampleScene{ID: id, Title: title, Description: description, Build: build}
}

func (a *State) createPage(id string) *componentPage {
	switch id {
	case "button":
		return buttonPage()
	case "input":
		return inputPage()
	case "example/workspace":
		return &componentPage{Scenes: []exampleScene{scene("workspace", "Workspace overview", "Save changes, open preferences and add or remove a document.", a.workspace)}}
	case "example/settings":
		model := settingsPage()
		return model
	case "example/task-browser":
		return taskBrowserPage()
	}
	info, _ := findPage(id)
	if info.Group == "Foundations" {
		return a.foundationPage(id)
	}
	if info.Group == "Inputs" {
		return inputComponentsPage(id)
	}
	if info.Group == "Overlays & navigation" {
		return a.overlayPage(id)
	}
	if info.Group == "Forms & data" {
		return dataPage(id)
	}
	if info.Group == "Advanced & layout" {
		return a.advancedPage(id)
	}
	return &componentPage{}
}
