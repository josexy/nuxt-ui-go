package gallery

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

type overlayNavigationState struct {
	Menu, Context                           nui.MenuState
	Dialog, Side, Preview, Pinned, Advanced bool
	Edge                                    nui.Side
	Tab, Step, Message                      string
	OpenSections                            []string
	Page                                    int
	Banner, Recovered                       bool
}

func (a *State) overlayComponents(c *ui.Context, t nui.Theme) {
	p := &a.Overlays
	heading(c, t, "Workspace actions", "Review a change, choose an export format, or adjust workspace preferences.")
	section(c, t, "Dialogs and contextual help", func() {
		ui.Row(c).Wrap().Gap(12).Children(func() {
			nui.Button(c, t, "Review changes", nui.ButtonOptions{}).OnClick(func() { p.Dialog = true })
			anchor := nui.Button(c, t, "Workspace help", nui.ButtonOptions{Variant: nui.Outline, Color: nui.Neutral}).OnClick(func() { p.Preview = !p.Preview })
			nui.Tooltip(c, t, anchor, "Read about workspace access", nui.TooltipOptions{})
			nui.Popover(c, t, anchor, &p.Preview, nui.PopoverOptions{Width: nui.Set(float32(300)), Label: "Access details"}, func(t nui.Theme) {
				ui.Text(c, "Workspace owners can invite members and change sharing preferences.").TextColor(t.Muted)
			})
			nui.Button(c, t, "Show notification", nui.ButtonOptions{Variant: nui.Soft, Leading: lucide.CircleCheck()}).OnClick(func() {
				nui.Toast(c, t, nui.ToastOptions{Title: "Draft saved", Description: "Your workspace changes are ready to review.", Color: nui.Success, Action: "View draft", OnAction: func() { p.Message = "Draft opened" }})
			})
		})
	})
	section(c, t, "Menus", func() {
		items := []nui.MenuItem{
			{ID: "workspace", Label: "Workspace", Kind: nui.MenuLabel},
			{ID: "open", Label: "Open workspace", Description: "Continue editing your draft", Icon: lucide.FolderOpen(), Shortcut: "Ctrl+O", OnSelect: func() { p.Message = "Workspace opened" }},
			{ID: "pin", Label: "Pin workspace", Checked: nui.Set(p.Pinned), OnSelect: func() { p.Pinned = !p.Pinned }},
			{ID: "export", Label: "Export", Icon: lucide.ArrowRight(), Children: []nui.MenuItem{
				{ID: "csv", Label: "CSV file", Description: "For spreadsheets", OnSelect: func() { p.Message = "CSV export prepared" }},
				{ID: "json", Label: "JSON file", Description: "For other applications", OnSelect: func() { p.Message = "JSON export prepared" }},
				{ID: "pdf", Label: "PDF report", Disabled: true},
			}},
			{ID: "divider", Kind: nui.MenuSeparator},
			{ID: "archive", Label: "Archive workspace", Color: nui.Error, OnSelect: func() { p.Message = "Workspace archived in this example" }},
		}
		anchor := nui.Button(c, t, "Project actions", nui.ButtonOptions{Variant: nui.Outline, Color: nui.Neutral, Trailing: lucide.ChevronDown()})
		nui.DropdownMenu(c, t, anchor, &p.Menu, items, nui.MenuOptions{Label: "Project menu", Width: nui.Set(float32(280))})
		target := ui.Column(c).MinWidth(0).Padding(20).Border(1, t.Border).Radius(8).Label("Project context area").Children(func() {
			ui.Text(c, "Desktop component library").FontWeight(500)
			ui.Text(c, "Right-click here, or focus this area and press Shift+F10.").TextColor(t.Muted)
		})
		nui.ContextMenu(c, t, target, &p.Context, items, nui.MenuOptions{Label: "Project context menu", Width: nui.Set(float32(280))})
		ui.Text(c, p.Message).Role(ui.RoleStatus).TextColor(t.Muted)
	})
	section(c, t, "Side panels", func() {
		ui.Row(c).Wrap().Gap(8).Children(func() {
			for _, entry := range []struct {
				label string
				side  nui.Side
			}{{"Right panel", nui.SideRight}, {"Left panel", nui.SideLeft}, {"Top panel", nui.SideTop}, {"Bottom panel", nui.SideBottom}} {
				nui.Button(c.Key(entry.label), t, entry.label, nui.ButtonOptions{Variant: nui.Outline, Color: nui.Neutral}).OnClick(func() { p.Edge = entry.side; p.Side = true })
			}
		})
	})
	nui.ModalWithOptions(c.Key("review-modal"), t, &p.Dialog, nui.ModalOptions{Title: "Review workspace changes", Description: "Choose a plan before confirming this draft.", Footer: func(c *ui.Context, t nui.Theme) {
		nui.Button(c, t, "Cancel review", nui.ButtonOptions{Variant: nui.Outline, Color: nui.Neutral}).OnClick(func() { p.Dialog = false })
		nui.Button(c, t, "Confirm changes", nui.ButtonOptions{}).OnClick(func() {
			p.Dialog = false
			nui.Toast(c, t, nui.ToastOptions{Title: "Changes confirmed", Color: nui.Success})
		})
	}}, func(t nui.Theme) {
		nui.Select(c, t, &a.Plan, plans, nui.SelectOptions{Label: "Review plan"}).FillWidth()
		nui.InputMenu(c, t, &a.Search, []nui.Option[string]{{Value: "lin", Label: "林晓"}, {Value: "alex", Label: "Alex Chen"}}, nui.InputMenuOptions{Label: "Reviewer", Placeholder: "Choose a reviewer"}).FillWidth()
	})
	nui.Slideover(c.Key("sample-side"), t, &p.Side, nui.SlideoverOptions{Title: "Workspace preferences", Description: "Changes apply to this example only.", Footer: func(c *ui.Context, t nui.Theme) {
		nui.Button(c, t, "Done with preferences", nui.ButtonOptions{}).OnClick(func() { p.Side = false })
	}, Side: p.Edge, Extent: nui.Set(float32(380))}, func(t nui.Theme) {
		nui.Checkbox(c, t, &p.Pinned, nui.CheckboxOptions{Label: "Pin this workspace"})
		nui.Collapsible(c, t, &p.Advanced, nui.CollapsibleOptions{Label: "Advanced options"}, func(t nui.Theme) {
			ui.Text(c, "Pinned workspaces stay near the top of your project list.").TextColor(t.Muted)
		})
	})
}
func (a *State) navigationComponents(c *ui.Context, t nui.Theme) {
	p := &a.Overlays
	nui.Breadcrumb(c, t, []nui.BreadcrumbItem{{ID: "home", Label: "Home", OnSelect: func() { p.Message = "Home selected" }}, {ID: "projects", Label: "Projects", OnSelect: func() { p.Message = "Projects selected" }}, {ID: "review", Label: "Workspace review"}}, nui.BreadcrumbOptions{Label: "Project location"})
	heading(c, t, "Workspace review", "Follow the review, inspect the activity, and decide when to continue.")
	nui.Banner(c, t, &p.Banner, nui.BannerOptions{Title: "Review requested", Description: "Your team can review the draft before it is published.", Icon: lucide.Info(), Color: nui.Info})
	nui.Stepper(c, t, &p.Step, []nui.StepItem{{ID: "details", Title: "Details", Description: "Describe the workspace"}, {ID: "review", Title: "Review", Description: "Check access and activity"}, {ID: "approve", Title: "Approve", Description: "Confirm the changes"}}, nui.StepperOptions{Label: "Review progress", OnChange: func(id string) { p.Message = "Current step: " + id }})
	nui.Tabs(c, t, &p.Tab, []nui.TabItem{
		{ID: "overview", Label: "Overview", Icon: lucide.FolderOpen(), Content: func(c *ui.Context, t nui.Theme) {
			nui.Accordion(c, t, &p.OpenSections, []nui.AccordionItem{
				{ID: "access", Label: "Sharing permissions", Icon: lucide.UserRound(), Content: func(c *ui.Context, t nui.Theme) {
					ui.Text(c, "Members can comment on drafts. Workspace owners manage access.").TextColor(t.Muted)
				}},
				{ID: "history", Label: "Version history", Icon: lucide.Save(), Content: func(c *ui.Context, t nui.Theme) {
					ui.Text(c, "This draft includes the latest saved preferences.").TextColor(t.Muted)
				}},
				{ID: "billing", Label: "Billing settings", Disabled: true},
			}, nui.AccordionOptions{Multiple: true})
		}},
		{ID: "activity", Label: "Activity", Icon: lucide.Check(), Content: func(c *ui.Context, t nui.Theme) {
			nui.Timeline(c, t, []nui.TimelineItem{{ID: "created", Time: "09:10", Title: "Workspace created", Description: "Lin Chen created the desktop workspace.", Icon: lucide.Check()}, {ID: "saved", Time: "10:25", Title: "Preferences saved", Description: "Notification delivery and region were updated.", Icon: lucide.Check()}, {ID: "reviewed", Time: "11:40", Title: "Review started", Description: "Alex Chen is reviewing access.", Icon: lucide.User(), Color: nui.Info}}, nui.TimelineOptions{Label: "Review activity", CurrentID: "reviewed"})
			nui.Pagination(c, t, &p.Page, nui.PaginationOptions{Label: "Activity pages", Total: 120, PageSize: 10, OnChange: func(page int) { p.Message = fmt.Sprintf("Activity page %d", page) }})
		}},
		{ID: "locked", Label: "Audit export", Disabled: true},
	}, nui.TabsOptions{Label: "Review sections"})
	section(c, t, "Loading recovery", func() {
		if !p.Recovered {
			nui.ErrorView(c, t, nui.ErrorOptions{Status: "Unavailable", Title: "Activity details could not load", Description: "The local preview is unavailable. Try loading it again.", OnRetry: func() { p.Recovered = true },
				// Embedded in a page section, the full-page title and copy are scaled down.
				Decoration: nui.Decoration{UI: nui.Styles{nui.Label: {FontSize: nui.Set(float32(24))}, nui.Description: {FontSize: nui.Set(float32(14))}}}})
		} else {
			nui.Alert(c, t, nui.AlertOptions{Title: "Activity restored", Description: "The local preview is ready.", Color: nui.Success, Icon: lucide.CircleCheck()})
		}
	})
	ui.Text(c, p.Message).Role(ui.RoleStatus).TextColor(t.Muted)
}
