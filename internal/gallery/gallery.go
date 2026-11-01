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
	"github.com/josexy/nuxt-ui-go/internal/settings"
)

type Record struct {
	ID, Name, Status string
	Members          int
}

type State struct {
	Props                               propsState
	WindowInactive                      bool
	Advanced                            advancedState
	FileDialog                          mygo.OpenDialogOptions
	FormsData                           formsDataState
	Overlays                            overlayNavigationState
	Settings                            *settings.State
	Dark                                bool
	Page                                int
	Name, Invalid, ReadOnly             string
	Saved                               int
	Plan                                string
	Search                              nui.SearchState
	Dialog, LocalDialog                 bool
	LocalPlan                           string
	Rows                                []Record
	List                                ui.ListState
	Selection                           ui.Selection[string]
	Sort, appliedSort                   ui.SortOrder
	CellCalls                           int
	MenuResult                          string
	RichOpen, RichSubOpen               bool
	NewWindow                           func()
	System, Compact, Animations         bool
	Primary, Neutral                    string
	ThemeMenu, PrimaryMenu, NeutralMenu nui.MenuState
	Notice, HasDocument                 bool
	Percent                             float64
	Message                             string
	AvatarImage                         *ui.Bitmap
}

func New() *State {
	s := &State{Name: "Nuxt workspace", Invalid: "team@", ReadOnly: "workspace_01", Plan: "team", LocalPlan: "personal", Primary: "green", Neutral: "slate", Animations: true}
	s.SetRows(10_000)
	s.Settings = settings.New()
	s.FormsData = newFormsData()
	s.Advanced = newAdvanced()
	s.Props = newProps()
	s.initCommands()
	s.FileDialog = mygo.OpenDialogOptions{Title: "Choose review attachments", Multiple: true}
	s.Overlays = overlayNavigationState{Banner: true, Step: "details", Tab: "overview", Page: 1}
	s.Notice, s.Percent = true, 42
	im := image.NewRGBA(image.Rect(0, 0, 64, 96))
	for y := range 96 {
		for x := range 64 {
			im.SetRGBA(x, y, color.RGBA{uint8(50 + x*2), uint8(90 + y), 170, 255})
		}
	}
	s.AvatarImage = ui.NewBitmap(im)
	return s
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

func (a *State) View(c *ui.Context) {
	if (a.Page != 17 || a.Props.Demo != "overlays") && a.Props.Files.Pending() {
		a.Props.Files.Dispose()
	}
	if a.Page != 15 && a.Advanced.Files.Pending() {
		a.Advanced.Files.Dispose()
	}
	s := nui.App(c, a.themeConfig())
	t := s
	nui.ToastViewport(c, t, nui.ToastViewportOptions{})

	ui.Column(c).Fill().Children(func() {
		ui.Row(c).Padding(16, 24).Gap(12).AlignItems(ui.Center).
			BorderWidth(0, 0, 1, 0).BorderColor(t.Border).Children(func() {
			ui.Text(c, "NuxtUI Go").FontSize(18).FontWeight(600).TextColor(t.Highlighted)
			ui.Text(c, "Native components").TextColor(t.Muted).Grow(1)
			language, nextLanguage := "English", "zh-Hans"
			if strings.HasPrefix(c.Locale(), "zh") {
				language, nextLanguage = "简体中文", "en"
			}
			nui.Button(c.Key("language-toggle"), s, language, nui.ButtonOptions{AccessibleLabel: "Language / 语言", Variant: nui.Outline, Neutral: true}).Description(language).OnClick(func() { mygo.App.SetLocale(nextLanguage) })
			a.themeMenu(c, s)
			if a.NewWindow != nil {
				nui.Button(c.Key("new-window"), s, "New window", nui.ButtonOptions{Variant: nui.Outline, Neutral: true}).OnClick(a.NewWindow)
			}
		})
		ui.Row(c).Grow(1).AlignItems(ui.Stretch).Children(func() {
			ui.Scroll(c.Key("gallery-index")).Width(188).MinHeight(0).Padding(20, 12).Gap(6).BorderWidth(0, 1, 0, 0).BorderColor(t.Border).Children(func() {
				for index, name := range []string{"Components", "Inputs & overlays", "Data & menus", "Theme & buttons", "Display", "Feedback", "Workspace", "Forms & settings", "Form states", "Overlays & menus", "Navigation & disclosure", "Advanced forms", "Dates & color", "Task browser", "Advanced interactions", "Activity & attachments", "Application layout", "Props & behavior", "Orientation"} {
					page := index
					opts := nui.ButtonOptions{Variant: nui.Outline, Neutral: true}
					if a.Page == page {
						opts = nui.ButtonOptions{Variant: nui.Soft}
					}
					nui.Button(c.Key(fmt.Sprintf("nav-%d", page)), s, name, opts).FillWidth().Justify(ui.Start).OnClick(func() { a.Page = page })
				}
				ui.Box(c).Grow(1)
				ui.Text(c, "Component gallery").FontSize(12).TextColor(t.Muted)
				ui.Text(c, "MyGo native UI").FontSize(12).TextColor(t.Muted)
			})
			ui.Scroll(c.Key(fmt.Sprintf("page-%d", a.Page))).Grow(1).Padding(28).Gap(24).Children(func() {
				switch a.Page {
				case 0:
					a.components(c, s)
				case 1:
					a.inputs(c, s)
				case 2:
					a.data(c, s)
				case 3:
					a.foundations(c, s)
				case 4:
					a.display(c, s)
				case 5:
					a.feedback(c, s)
				case 6:
					a.workspace(c, s)
				case 7:
					a.Settings.Panel(c, s).Height(700)
				case 8:
					a.formStates(c, s)
				case 9:
					a.overlayComponents(c, s)
				case 10:
					a.navigationComponents(c, s)
				case 11:
					a.advancedForms(c, s)
				case 12:
					a.dateComponents(c, s)
				case 14:
					a.advancedInteractions(c, s)
				case 15:
					a.chartsAndFiles(c, s)
				case 16:
					a.shellComponents(c, s)
				case 17:
					a.commonProps(c, s)
				case 18:
					a.orientationProps(c, s)
				case 13:
					heading(c, s, "Task browser", "Filter, sort and edit tasks by their stable identity.")
					a.FormsData.Browser.Panel(c, s).Height(550)
				}
			})
		})
	})
	a.advancedOverlays(c, t)
}

func heading(c *ui.Context, s nui.Theme, title, description string) {
	t := s
	ui.Column(c).Gap(6).Children(func() {
		ui.Text(c, title).FontSize(24).FontWeight(600).TextColor(t.Highlighted)
		ui.Text(c, description).TextColor(t.Muted)
	})
}

func section(c *ui.Context, s nui.Theme, title string, children func()) {
	t := s
	ui.Column(c).Gap(14).Padding(0, 0, 24, 0).BorderWidth(0, 0, 1, 0).BorderColor(t.Border).Children(func() {
		ui.Text(c, title).FontWeight(600).TextColor(t.Highlighted)
		children()
	})
}

func (a *State) components(c *ui.Context, s nui.Theme) {
	t := s
	heading(c, s, "Native components", "Nuxt styling on MyGo controls. Try the keyboard, then switch themes.")
	section(c, s, "Button variants", func() {
		ui.Row(c).Gap(12).Children(func() {
			nui.Button(c.Key("save"), s, "Save changes", nui.ButtonOptions{Leading: lucide.Check()}).OnClick(func() { a.Saved++ })
			nui.Button(c, s, "Outline", nui.ButtonOptions{Variant: nui.Outline})
			nui.Button(c, s, "Soft", nui.ButtonOptions{Variant: nui.Soft})
			nui.Button(c, s, "Neutral", nui.ButtonOptions{Variant: nui.Outline, Neutral: true})
		})
		ui.Textf(c, "Saved %d times", a.Saved).FontSize(12).TextColor(t.Muted)
	})
	section(c, s, "States and custom parts", func() {
		ui.Row(c).Gap(12).Children(func() {
			nui.Button(c, s, "Disabled", nui.ButtonOptions{Disabled: true})
			nui.Button(c, s, "Loading", nui.ButtonOptions{Loading: true})
			nui.Button(c, s, "Custom label", nui.ButtonOptions{Variant: nui.Soft, Style: func(root, label, leading, trailing ui.Element) { label.FontWeight(700) }}).Width(170)
		})
		ui.Text(c, "Tab highlights keyboard focus. Loading blocks repeated activation.").FontSize(12).TextColor(t.Muted)
	})
	section(c, s, "Lucide icons", func() {
		ui.Row(c).Gap(20).Children(func() {
			for _, icon := range []*ui.SVG{lucide.Check(), lucide.ChevronDown(), lucide.Search(), lucide.X(), lucide.Save(), lucide.Ellipsis(), lucide.Sun(), lucide.Moon()} {
				ui.Icon(c, icon).Size(24, 24).TextColor(t.Muted)
			}
		})
	})
	section(c, s, "Local theme", func() {
		local := nui.Dark()
		if a.Dark {
			local = nui.Light()
		}
		nui.Themed(c.Key("local-theme"), local, func(inner nui.Theme) {
			ui.Text(c, "A separate theme inside this window").FontWeight(500)
			ui.Row(c).Gap(12).Children(func() {
				nui.Select(c.Key("local-plan"), inner, &a.LocalPlan, plans, nui.SelectOptions{Label: "Local plan"}).Width(220)
				nui.Button(c, inner, "Local dialog", nui.ButtonOptions{Variant: nui.Soft}).OnClick(func() { a.LocalDialog = true })
			})
			nui.Modal(c.Key("local-dialog"), inner, &a.LocalDialog, "Local theme dialog", func(ms nui.Theme) {
				ui.Text(c, "This dialog keeps the theme of its origin.")
				nui.Button(c, ms, "Close local dialog", nui.ButtonOptions{}).OnClick(func() { a.LocalDialog = false })
			})
		}).Padding(20).Gap(16).Radius(10)
	})
}

func (a *State) inputs(c *ui.Context, s nui.Theme) {
	t := s
	heading(c, s, "Inputs & overlays", "Editing, duplicate labels and nested focus scopes.")
	section(c, s, "Workspace details", func() {
		ui.Text(c, "Workspace name").FontWeight(500)
		nui.Input(c.Key("workspace-name"), s, &a.Name, nui.InputOptions{Label: "Workspace name", Leading: lucide.Search(), Trailing: lucide.Check()}).Width(440)
		ui.Text(c, "错误状态与中文输入").FontWeight(500)
		nui.Input(c.Key("invalid-email"), s, &a.Invalid, nui.InputOptions{Label: "Email", Invalid: true, ErrorMessage: "Enter a complete email address"}).Width(440)
		ui.Text(c, "Enter a complete email address").FontSize(12).TextColor(t.Error)
		nui.Input(c.Key("readonly"), s, &a.ReadOnly, nui.InputOptions{Label: "Workspace ID", ReadOnly: true}).Width(440)
	})
	section(c, s, "Selection", func() {
		ui.Row(c).Gap(16).Children(func() {
			nui.Select(c.Key("plan"), s, &a.Plan, plans, nui.SelectOptions{Label: "Plan"}).Width(240)
			nui.SearchSelect(c.Key("search"), s, &a.Search, plans, "Search workspaces").Width(300)
		})
		ui.Textf(c, "Selected ID: %s   Search ID: %s", a.Plan, a.Search.Selected).FontSize(12).TextColor(t.Muted)
	})
	nui.Button(c.Key("open-dialog"), s, "Edit workspace", nui.ButtonOptions{}).OnClick(func() { a.Dialog = true })
	nui.Modal(c.Key("edit-dialog"), s, &a.Dialog, "Edit workspace", func(ms nui.Theme) {
		ui.Text(c, "Choose a plan, then use Escape to close each layer.").TextColor(t.Muted)
		nui.Input(c.Key("dialog-name"), ms, &a.Name, nui.InputOptions{Label: "Dialog workspace name"}).FillWidth()
		nui.Select(c.Key("dialog-plan"), ms, &a.Plan, plans, nui.SelectOptions{Label: "Dialog plan"}).FillWidth()
		ui.Row(c).Gap(8).Justify(ui.End).Children(func() {
			nui.Button(c, ms, "Cancel", nui.ButtonOptions{Variant: nui.Outline, Neutral: true}).OnClick(func() { a.Dialog = false })
			nui.Button(c, ms, "Save workspace", nui.ButtonOptions{}).OnClick(func() { a.Saved++; a.Dialog = false })
		})
	})
}

func (a *State) data(c *ui.Context, s nui.Theme) {
	t := s
	heading(c, s, "Data & menus", "10,000 workspaces with sorting, resizing and stable selection.")
	ui.Row(c).Gap(10).Children(func() {
		button := nui.Button(c.Key("native-menu"), s, "Workspace actions", nui.ButtonOptions{Variant: nui.Outline, Neutral: true, Trailing: lucide.ChevronDown()})
		button.PopoverMenuButton(func(m *ui.Menu) {
			if m.Item("Duplicate workspace").Shortcut(ui.Ctrl, ui.KeyD).Chosen() {
				a.MenuResult = "Duplicated"
			}
			m.Item("Unavailable action").Disabled(true)
			m.Submenu("Export", func(sub *ui.Menu) {
				if sub.Item("CSV file").Chosen() {
					a.MenuResult = "Exported CSV"
				}
				if sub.Item("JSON file").Chosen() {
					a.MenuResult = "Exported JSON"
				}
			})
		})
		anchor := nui.Button(c.Key("rich-actions"), s, "Rich actions", nui.ButtonOptions{Variant: nui.Soft, Trailing: lucide.Ellipsis()})
		anchor.OnClick(func() { a.RichOpen = !a.RichOpen })
		a.richActions(c, s, anchor)
		ui.Text(c, a.MenuResult).TextColor(t.Muted)
	})
	a.Table(c, s)
	ui.Textf(c, "%d selected", a.Selection.Len()).FontSize(12).TextColor(t.Muted)
}

// Table intentionally uses MyGo's public Table without copying its internals.
// Header builders and row paint hooks are not exposed by this upstream version.
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
