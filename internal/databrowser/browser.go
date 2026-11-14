// Package databrowser demonstrates caller-owned filtering, paging and editing by stable ID.
package databrowser

import (
	"fmt"
	"slices"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

type Task struct{ ID, Name, Status, Team string }
type State struct {
	Config                       nui.AppConfig
	Table                        nui.TableState[Task]
	Tree                         nui.TreeState
	All                          []Task
	Query                        string
	Page, Total                  int
	ActiveOnly, Filters, Editing bool
	EditID, EditName, Message    string
}

func New(count int) *State {
	s := &State{Config: nui.DefaultApp(), Page: 1}
	s.All = make([]Task, count)
	for i := range s.All {
		status := "Active"
		if i%3 == 0 {
			status = "Draft"
		}
		s.All[i] = Task{fmt.Sprintf("task-%d", i), fmt.Sprintf("Task %05d", i+1), status, []string{"Design", "Engineering", "Operations"}[i%3]}
	}
	s.Tree.SetNodes([]nui.TreeNode{{ID: "teams", Label: "Teams", Children: []nui.TreeNode{{ID: "Design", Label: "Design"}, {ID: "Engineering", Label: "Engineering"}, {ID: "Operations", Label: "Operations"}}}})
	s.Tree.Outline.Open.Add("teams")
	s.Apply()
	return s
}
func (s *State) Apply() {
	rows := []Task{}
	for _, task := range s.All {
		if s.ActiveOnly && task.Status != "Active" {
			continue
		}
		if s.Query != "" && !strings.Contains(strings.ToLower(task.Name), strings.ToLower(s.Query)) {
			continue
		}
		if s.Tree.Selection.Len() > 0 && !s.Tree.Selection.Has("teams") && !s.Tree.Selection.Has(task.Team) {
			continue
		}
		rows = append(rows, task)
	}
	s.Total = len(rows)
	s.Page = max(1, min(max(1, (len(rows)+99)/100), s.Page))
	start := (s.Page - 1) * 100
	end := min(len(rows), start+100)
	s.Table.SetRows(rows[start:end], func(v Task) string { return v.ID })
}
func (s *State) selected() string {
	if s.Table.Selection.Len() != 1 {
		return ""
	}
	for id := range s.Table.Selection.All() {
		return id
	}
	return ""
}
func (s *State) beginEdit() {
	s.EditID = s.selected()
	for _, task := range s.All {
		if task.ID == s.EditID {
			s.EditName = task.Name
			s.Editing = true
			break
		}
	}
}
func (s *State) save() {
	for i := range s.All {
		if s.All[i].ID == s.EditID {
			s.All[i].Name = strings.TrimSpace(s.EditName)
			break
		}
	}
	s.Editing = false
	s.Message = "Task updated"
	s.Apply()
}
func (s *State) View(c *ui.Context) {
	t := nui.App(c, s.Config)
	nui.ToastViewport(c, t, nui.ToastViewportOptions{Position: nui.ToastTopRight})
	ui.Column(c).Fill().Padding(24).Gap(20).Children(func() {
		ui.Row(c).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "Task browser").FontSize(24).FontWeight(600).Grow(1)
			nui.Button(c, t, "Theme", nui.ButtonOptions{Variant: nui.Outline, Leading: lucide.Moon()}).OnClick(func() {
				if t.Dark {
					s.Config.Mode = nui.LightMode
				} else {
					s.Config.Mode = nui.DarkMode
				}
			})
		})
		s.Panel(c, t).Grow(1)
	})
}
func (s *State) Panel(c *ui.Context, t nui.Theme) ui.Element {
	return ui.Column(c).MinWidth(0).MinHeight(0).Gap(16).Children(func() {
		ui.Row(c).Gap(8).Children(func() {
			nui.Input(c, t, &s.Query, nui.InputOptions{Label: "Search tasks", Placeholder: "Search tasks", Leading: lucide.Search(), Clearable: true, OnChange: func(string) { s.Page = 1; s.Apply() }}).Grow(1)
			nui.Button(c, t, "Filters", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { s.Filters = true })
			nui.Button(c, t, "Edit task", nui.ButtonOptions{Disabled: s.selected() == ""}).OnClick(s.beginEdit)
		})
		ui.Row(c).Gap(16).AlignItems(ui.Stretch).Grow(1).MinHeight(240).Children(func() {
			nui.Tree(c, t, &s.Tree, nui.TreeOptions{Label: "Task teams", Height: 360, OnChange: func() { s.Page = 1; s.Apply() }}).Width(190)
			columns := []nui.TableColumn[Task]{{ID: "name", Title: "Task", Width: 300, MinWidth: 160, Value: func(v Task) string { return v.Name }, Compare: func(a, b Task) int { return strings.Compare(a.Name, b.Name) }}, {ID: "team", Title: "Team", Width: 150, Value: func(v Task) string { return v.Team }, Compare: func(a, b Task) int { return strings.Compare(a.Team, b.Team) }}, {ID: "status", Title: "Status", Width: 120, Value: func(v Task) string { return v.Status }, Cell: func(c *ui.Context, t nui.Theme, v Task, selected bool) {
				color := nui.Success
				if v.Status == "Draft" {
					color = nui.Neutral
				}
				nui.Badge(c, t, v.Status, nui.BadgeOptions{Color: color, Variant: nui.Soft})
			}}}
			nui.Table(c, t, &s.Table, columns, nui.TableOptions{Label: "Tasks", Height: 360}).Grow(1)
		})
		ui.Row(c).AlignItems(ui.Center).Children(func() {
			ui.Textf(c, "%d tasks · %d selected", s.Total, s.Table.Selection.Len()).TextColor(t.Muted).Grow(1)
			nui.Pagination(c, t, &s.Page, nui.PaginationOptions{Total: s.Total, PageSize: 100, OnChange: func(int) { s.Apply() }})
		})
		if s.Message != "" {
			ui.Text(c, s.Message).Role(ui.RoleStatus).TextColor(t.Muted)
		}
		nui.Slideover(c, t, &s.Filters, nui.SlideoverOptions{Title: "Task filters", Footer: func(c *ui.Context, t nui.Theme) {
			nui.Button(c, t, "Apply filters", nui.ButtonOptions{}).OnClick(func() { s.Filters = false; s.Page = 1; s.Apply() })
		}}, func(inner nui.Theme) {
			nui.Checkbox(c, inner, &s.ActiveOnly, nui.CheckboxOptions{Label: "Active tasks only"})
			nui.Button(c, inner, "Clear filters", nui.ButtonOptions{Variant: nui.Ghost}).OnClick(func() { s.ActiveOnly = false; s.Tree.Selection.Clear(); s.Query = ""; s.Page = 1; s.Apply() })
		})
		nui.ModalWithOptions(c, t, &s.Editing, nui.ModalOptions{Title: "Edit task details", Footer: func(c *ui.Context, t nui.Theme) {
			nui.Button(c, t, "Cancel task edit", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { s.Editing = false })
			nui.Button(c, t, "Save task", nui.ButtonOptions{Disabled: strings.TrimSpace(s.EditName) == ""}).OnClick(func() {
				s.save()
				nui.Toast(c, t, nui.ToastOptions{Title: "Task updated", Description: s.EditName, Color: nui.Success})
			})
		}}, func(inner nui.Theme) { nui.Input(c, inner, &s.EditName, nui.InputOptions{Label: "Task name"}) })
	})
}

// SelectedIDs is an owned snapshot useful to caller commands.
func (s *State) SelectedIDs() []string {
	ids := []string{}
	for id := range s.Table.Selection.All() {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	return ids
}
