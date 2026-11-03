package gallery

import (
	"slices"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

// The order matches State.Page; IDs keep widget identity independent of labels.
var galleryNavigation = []nui.NavigationItem{
	{ID: "components", Label: "Components", Icon: lucide.Boxes()},
	{ID: "inputs-overlays", Label: "Inputs & overlays", Icon: lucide.TextCursorInput()},
	{ID: "data-menus", Label: "Data & menus", Icon: lucide.Table2()},
	{ID: "theme-buttons", Label: "Theme & buttons", Icon: lucide.Palette()},
	{ID: "display", Label: "Display", Icon: lucide.PanelsTopLeft()},
	{ID: "feedback", Label: "Feedback", Icon: lucide.Bell()},
	{ID: "workspace", Label: "Workspace", Icon: lucide.LayoutDashboard()},
	{ID: "forms-settings", Label: "Forms & settings", Icon: lucide.Settings()},
	{ID: "form-states", Label: "Form states", Icon: lucide.ListChecks()},
	{ID: "overlays-menus", Label: "Overlays & menus", Icon: lucide.Layers()},
	{ID: "navigation-disclosure", Label: "Navigation & disclosure", Icon: lucide.Compass()},
	{ID: "advanced-forms", Label: "Advanced forms", Icon: lucide.ListFilter()},
	{ID: "dates-color", Label: "Dates & color", Icon: lucide.CalendarDays()},
	{ID: "task-browser", Label: "Task browser", Icon: lucide.ClipboardList()},
	{ID: "advanced-interactions", Label: "Advanced interactions", Icon: lucide.Workflow()},
	{ID: "activity-attachments", Label: "Activity & attachments", Icon: lucide.ChartNoAxesCombined()},
	{ID: "application-layout", Label: "Application layout", Icon: lucide.AppWindow()},
	{ID: "props-behavior", Label: "Props & behavior", Icon: lucide.SlidersHorizontal()},
	{ID: "orientation", Label: "Orientation", Icon: lucide.Move()},
}

func (a *State) sidebar(c *ui.Context, t nui.Theme) {
	// Other examples and commands can change Page directly. Keep the navigation
	// selection in sync while binding the widget to a persistent model field.
	a.navigationActive = ""
	if a.Page >= 0 && a.Page < len(galleryNavigation) {
		a.navigationActive = galleryNavigation[a.Page].ID
	}
	nui.Sidebar(c.Key("gallery-index"), t, &a.SidebarCollapsed, nui.SidebarOptions{
		Title: "Gallery", State: &a.Sidebar, Resizable: true,
		MinWidth: nui.Set(float32(200)), MaxWidth: nui.Set(float32(360)),
		ToggleLabel: "Toggle gallery sidebar", ResizeLabel: "Resize gallery sidebar",
		FooterContent: func(c *ui.Context, t nui.Theme, scope nui.SidebarScope) {
			if !scope.Collapsed {
				ui.Text(c, "MyGo native UI").FontSize(12).TextColor(t.Muted)
			}
		},
	}, func(c *ui.Context, t nui.Theme, compact bool) {
		t = t.WithComponent(nui.ComponentButton, nui.ComponentTheme{Size: nui.SM})
		nui.NavigationMenu(c.Key("pages"), t, &a.navigationActive, &a.Navigation, galleryNavigation, nui.NavigationOptions{
			Label: "Gallery pages", Vertical: true, Compact: compact,
			OnNavigate: func(id string) {
				if page := slices.IndexFunc(galleryNavigation, func(item nui.NavigationItem) bool { return item.ID == id }); page >= 0 {
					a.Page = page
				}
			},
		})
	}).Label("Gallery sidebar")
}
