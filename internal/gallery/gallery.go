// Package gallery provides component examples for native windows and automated tests.
package gallery

import (
	"cmp"
	"fmt"
	"image"
	"image/color"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
	nui "github.com/josexy/nuxt-ui-go/ui"
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
	toastViewport                         nui.ToastViewportOptions
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
	nui.TableState[Record]
	TableConfig           nui.TableOptions
	CellCalls             int
	MenuResult            string
	RichOpen, RichSubOpen bool
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
	s.TableState.SetRows(s.Rows, func(v Record) string { return v.ID })
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
	columns := []nui.TableColumn[Record]{
		{ID: "name", Title: "Workspace", Width: 300, MinWidth: 160, MaxWidth: 600, Value: func(v Record) string { return v.Name }, Compare: func(a, b Record) int { return strings.Compare(a.Name, b.Name) }},
		{ID: "status", Title: "Status", Width: 150, Value: func(v Record) string { return v.Status }, Cell: func(c *ui.Context, t nui.Theme, record Record, chosen bool) {
			color := t.Primary
			if record.Status == "Draft" {
				color = t.Muted
			}
			if chosen {
				color = t.Inverted
			}
			ui.Text(c, record.Status).FontSize(12).TextColor(color).Background(color.Alpha(0.10)).Padding(2, 6).Radius(4)
		}},
		{ID: "members", Title: "Members", Width: 110, Value: func(v Record) string { return fmt.Sprint(v.Members) }, Compare: func(a, b Record) int { return cmp.Compare(a.Members, b.Members) }, Cell: func(c *ui.Context, _ nui.Theme, record Record, _ bool) {
			ui.Row(c).Gap(6).Children(func() {
				ui.Icon(c, lucide.Check()).Size(14, 14)
				ui.Textf(c, "%d", record.Members).SingleLine()
			})
		}},
	}
	o := a.TableConfig
	o.Label = "Workspace table"
	if o.Height <= 0 {
		o.Height = 370
	}
	if o.EmptyText == "" {
		o.EmptyText = "No workspace records"
	}
	table := nui.Table(c.Key("workspaces-table"), s, &a.TableState, columns, o).FillWidth()
	a.CellCalls = a.CellBuilds
	return table
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
