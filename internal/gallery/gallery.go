// Package gallery provides component examples for native windows and automated tests.
package gallery

import (
	"fmt"
	"image"
	"image/color"
	"slices"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

type Record struct {
	ID, Name, Status string
	Members          int
}

type State struct {
	ActiveID, DirectorySearch             string
	pages                                 map[string]*componentPage
	directoryToggle, directorySearch      ui.Handle
	closeToasts, disposed, WindowInactive bool
	pendingToasts                         []nui.ToastOptions
	FileDialog                            mygo.OpenDialogOptions
	Sidebar                               nui.SidebarState
	SidebarCollapsed                      bool
	Navigation                            nui.NavigationState
	navigationActive                      string
	NewWindow                             func()
	Dark, System, Compact, Animations     bool
	Primary, Neutral                      string
	ThemeMenu, PrimaryMenu, NeutralMenu   nui.MenuState
	Saved                                 int
	HasDocument                           bool
	Message                               string
	AvatarImage                           *ui.Bitmap
	Advanced                              advancedState
	Rows                                  []Record
	List                                  ui.ListState
	Selection                             ui.Selection[string]
	Sort, appliedSort                     ui.SortOrder
	CellCalls                             int
	MenuResult                            string
	RichOpen, RichSubOpen                 bool
}

func New() *State {
	return &State{ActiveID: "button", pages: make(map[string]*componentPage), Primary: "green", Neutral: "slate", Animations: true, FileDialog: mygo.OpenDialogOptions{Title: "Choose review attachments", Multiple: true}}
}

func (a *State) avatarImage() *ui.Bitmap {
	if a.AvatarImage == nil {
		im := image.NewRGBA(image.Rect(0, 0, 64, 96))
		for y := range 96 {
			for x := range 64 {
				im.SetRGBA(x, y, color.RGBA{uint8(50 + x*2), uint8(90 + y), 170, 255})
			}
		}
		a.AvatarImage = ui.NewBitmap(im)
	}
	return a.AvatarImage
}

func (s *State) SetRows(n int) {
	s.Rows = make([]Record, n)
	for i := range s.Rows {
		status := "Active"
		if i%3 == 0 {
			status = "Draft"
		}
		s.Rows[i] = Record{ID: fmt.Sprintf("workspace-%d", i), Name: fmt.Sprintf("Workspace %05d", i+1), Status: status, Members: i%87 + 1}
	}
	s.List.Key = func(i int) any { return s.Rows[i].ID }
	s.List.Label = func(i int) string { return s.Rows[i].Name }
	s.List.Selection, s.List.Sort = &s.Selection, &s.Sort
}

var plans = []nui.Option[string]{
	{Value: "personal", Label: "Personal", Description: "For your own projects"},
	{Value: "team", Label: "Team", Description: "Shared workspace"},
	{Value: "team-eu", Label: "Team", Description: "European workspace"},
}

func heading(c *ui.Context, s nui.Theme, title, description string) {
	t := s
	ui.Column(c).Gap(6).Children(func() {
		ui.Text(c, title).FontSize(24).FontWeight(600).TextColor(t.Highlighted)
		ui.Text(c, description).TextColor(t.Muted)
	})
}

func (a *State) Table(c *ui.Context, s nui.Theme) ui.Element {
	t := s
	a.CellCalls = 0
	columns := []ui.TableColumn{
		{ID: "name", Title: "Workspace", Width: 300, MinWidth: 160, MaxWidth: 600, Sortable: true},
		{ID: "status", Title: "Status", Width: 150},
		{ID: "members", Title: "Members", Width: 110, Sortable: true},
	}
	return ui.Table(c.Key("workspaces-table"), &a.List, columns, len(a.Rows), func(row, col int) {
		a.CellCalls++
		record := a.Rows[row]
		chosen := a.Selection.Has(record.ID)
		switch col {
		case 0:
			ui.Text(c, record.Name).SingleLine()
		case 1:
			color := t.Primary
			if record.Status == "Draft" {
				color = t.Muted
			}
			if chosen {
				color = t.Inverted
			}
			ui.Text(c, record.Status).FontSize(12).TextColor(color).Background(color.Alpha(0.10)).Padding(2, 6).Radius(4)
		case 2:
			ui.Row(c).Gap(6).Children(func() {
				ui.Icon(c, lucide.Check()).Size(14, 14)
				ui.Textf(c, "%d", record.Members).SingleLine()
			})
		}
	}).Label("Workspace table").Height(370).FillWidth().Border(1, t.Border).Radius(t.Radius).OnChange(func() {
		if a.Sort == a.appliedSort {
			return
		}
		a.appliedSort = a.Sort
		slices.SortStableFunc(a.Rows, func(left, right Record) int {
			order := strings.Compare(left.Name, right.Name)
			if a.Sort.Column == "members" {
				order = left.Members - right.Members
			}
			if a.Sort.Descending {
				order = -order
			}
			return order
		})
	})
}

// richActions is an experiment in composition, not a released DropdownMenu.
// Native PopoverMenu above supplies the full built-in submenu baseline.
func (a *State) richActions(c *ui.Context, s nui.Theme, anchor ui.Element) {
	t := s
	ui.PopoverBase(c, anchor, &a.RichOpen, func(panel ui.Element) {
		panel.Column().Padding(6).Gap(4).Width(270).Radius(8).Background(t.Background).Border(1, t.Border).Shadow(0, 4, 12, 0, ui.RGBA(0, 0, 0, 0.12))
		action := nui.Button(c, s, "", nui.ButtonOptions{Variant: nui.Soft, AccessibleLabel: "Rich save"}).FillWidth().Justify(ui.Start).Label("Rich save")
		action.Children(func() {
			ui.Icon(c, lucide.Save()).Size(20, 20)
			ui.Column(c).Gap(2).Children(func() {
				ui.Text(c, "Save a copy")
				ui.Text(c, "Keep the current workspace").FontSize(12).TextColor(t.Muted)
			})
		})
		action.OnClick(func() { a.MenuResult = "Copy saved"; a.RichOpen = false })
		sub := nui.Button(c, s, "Export options", nui.ButtonOptions{Variant: nui.Outline, Neutral: true, Trailing: lucide.ChevronDown()}).FillWidth()
		sub.OnClick(func() { a.RichSubOpen = !a.RichSubOpen })
		ui.PopoverBase(c, sub, &a.RichSubOpen, func(child ui.Element) {
			child.Column().AttachTo(sub, ui.AnchorTopEnd, ui.AnchorTopStart).Padding(6).Gap(4).Radius(8).Background(t.Background).Border(1, t.Border)
			nui.Button(c, s, "Export rich CSV", nui.ButtonOptions{Variant: nui.Soft}).OnClick(func() { a.MenuResult = "Rich CSV exported"; a.RichSubOpen = false; a.RichOpen = false })
		})
	})
	if !a.RichOpen {
		a.RichSubOpen = false
	}
}
