package gallery

import (
	"fmt"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

type advancedState struct {
	Command                  nui.CommandPaletteState
	Drawer                   nui.DrawerState
	Carousel                 nui.CarouselState
	Marquee                  nui.MarqueeState
	Files                    nui.FileUploadState
	Chart                    nui.ChartState
	Navigation               nui.NavigationState
	Active                   string
	Collapsed                bool
	Split                    float32
	Result                   string
	ChartKind                nui.ChartKind
	ChartCurve               nui.ChartCurve
	ChartStack               nui.ChartStackMode
	ChartAxis                string
	ChartRevision            int
	ChartNegative, ChartGaps bool
}

func newAdvanced() advancedState {
	s := advancedState{Active: "activity", Split: 62, Result: "Ready", ChartAxis: "category", ChartCurve: nui.ChartCurveMonotone}
	s.updateChartData()
	return s
}
func (a *State) initCommands() {
	a.Advanced.Command.SetIndex(nui.NewCommandIndex([]nui.CommandAction{
		{ID: "activity", Label: "Open activity", Group: "Workspace", Icon: lucide.ArrowRight(), Keywords: "chart graph metrics", OnSelect: func() { a.Page = 15; a.Advanced.Result = "Activity opened" }},
		{ID: "attachments", Label: "Review attachments", Group: "Workspace", Icon: lucide.FolderOpen(), OnSelect: func() { a.Page = 15; a.Advanced.Result = "Attachments opened" }},
		{ID: "appearance", Label: "Appearance", Group: "Preferences", Icon: lucide.Settings(), Children: []nui.CommandAction{
			{ID: "light", Label: "Use light theme", OnSelect: func() { a.Dark = false }},
			{ID: "dark", Label: "Use dark theme", OnSelect: func() { a.Dark = true }},
		}},
		{ID: "inspector", Label: "Open inspector", Group: "Workspace", Icon: lucide.Search(), Key: ui.KeyI, Mods: ui.Ctrl | ui.Shift, Shortcut: "Ctrl Shift I", OnSelect: func() { a.Advanced.Drawer.Open = true }},
	}))
}
func (a *State) advancedOverlays(c *ui.Context, t nui.Theme) {
	nui.CommandPalette(c.Key("workspace-commands"), t, &a.Advanced.Command, nui.CommandPaletteOptions{Title: "Workspace commands", OnSelect: func(id string) { a.Advanced.Result = "Selected " + id }})
	nui.Drawer(c.Key("activity-drawer"), t, &a.Advanced.Drawer, nui.DrawerOptions{Title: "Activity inspector", Description: "Drag the handle or use its arrow keys to resize.", Footer: func(c *ui.Context, t nui.Theme) {
		nui.Button(c, t, "Done reviewing", nui.ButtonOptions{}).OnClick(func() { a.Advanced.Drawer.Open = false })
	}, SnapPoints: []float32{.45, .8}}, func(t nui.Theme) {
		ui.Text(c, "Build review").FontWeight(600)
		ui.Text(c, "This local workspace has no upload connection. Review attachments and activity before choosing your next action.")
		for i := range 16 {
			ui.Text(c, fmt.Sprintf("Run %03d — ready for review", i+1)).Padding(8, 0)
		}
	})
}
func (a *State) advancedInteractions(c *ui.Context, t nui.Theme) {
	heading(c, t, "Advanced interactions", "Find an action, review slides, or open the activity inspector.")
	ui.Row(c).Wrap().Gap(12).Children(func() {
		nui.Button(c, t, "Search commands", nui.ButtonOptions{Leading: lucide.Search()}).OnClick(func() { a.Advanced.Command.Open = true })
		nui.Kbd(c, t, []string{"Ctrl", "K"}, nui.KbdOptions{})
		nui.Button(c, t, "Open inspector", nui.ButtonOptions{Variant: nui.Outline, Color: nui.Neutral}).OnClick(func() { a.Advanced.Drawer.Open = true })
	})
	ui.Text(c, a.Advanced.Result).TextColor(t.Muted)
	nui.Carousel(c, t, &a.Advanced.Carousel, []nui.CarouselItem{
		{ID: "collect", Label: "Collect files", Build: func(c *ui.Context, t nui.Theme) {
			ui.Text(c, "Collect your review files").FontSize(24).FontWeight(600)
			ui.Text(c, "Choose local files or drop them into the attachment area.").TextColor(t.Muted)
		}},
		{ID: "inspect", Label: "Inspect activity", Build: func(c *ui.Context, t nui.Theme) {
			ui.Text(c, "Inspect recent activity").FontSize(24).FontWeight(600)
			ui.Text(c, "Compare builds and reviews, including missing samples.").TextColor(t.Muted)
		}},
		{ID: "review", Label: "Review changes", Build: func(c *ui.Context, t nui.Theme) {
			ui.Text(c, "Review before sharing").FontSize(24).FontWeight(600)
			ui.Text(c, "Keep the inspector open while you read the activity log.").TextColor(t.Muted)
		}},
	}, nui.CarouselOptions{Label: "Review guide", Arrows: nui.Set(true), Active: nui.Set(!a.WindowInactive && !a.Advanced.Command.Open && !a.Advanced.Drawer.Open), Height: 175, Autoplay: 4 * time.Second, Wrap: true})
	nui.Marquee(c, t, &a.Advanced.Marquee, nui.MarqueeOptions{Label: "Workspace updates", Active: nui.Set(!a.WindowInactive && !a.Advanced.Command.Open && !a.Advanced.Drawer.Open), ContentKey: "updates-v1", Height: 44, Speed: 28}, func() {
		ui.Text(c, "Build 240 passed    •    Review queue updated    •    Attachments stay local    •    Workspace ready    •    Use Ctrl K to find an action").NoWrap().Padding(10, 0).TextColor(t.Muted)
	})
	nui.Button(c, t, "Pause updates", nui.ButtonOptions{Variant: nui.Ghost, Color: nui.Neutral}).OnClick(func() { a.Advanced.Marquee.Paused = !a.Advanced.Marquee.Paused })
}
func (a *State) chartsAndFiles(c *ui.Context, t nui.Theme) {
	heading(c, t, "Activity & attachments", "Compare recent runs and prepare local files for review.")
	a.chartDemo(c, t)
	nui.FileUpload(c, t, &a.Advanced.Files, nui.FileUploadOptions{Label: "Review attachments", Description: "Drop text, Markdown or PNG files. Up to 5 files, 10 MB each.", Rules: nui.UploadRules{Accept: []string{".txt", ".md", ".png"}, MaxBytes: 10 << 20, MaxFiles: 5}, Dialog: a.FileDialog})
	nui.Button(c, t, "Add sample attachment", nui.ButtonOptions{Variant: nui.Ghost, Color: nui.Neutral}).OnClick(func() {
		a.Advanced.Files.Add([]nui.UploadFile{{Path: "sample/review.md", Name: "review.md", Size: 2048, MIME: "text/markdown"}}, nui.UploadRules{})
	})
}

var shellNavigation = []nui.NavigationItem{
	{ID: "activity", Label: "Activity", Group: "Workspace", Icon: lucide.Search()},
	{ID: "attachments", Label: "Attachments", Icon: lucide.FolderOpen()},
	{ID: "settings", Label: "Workspace settings", Group: "Manage", Icon: lucide.Settings(), Children: []nui.NavigationItem{{ID: "members", Label: "Members"}, {ID: "permissions", Label: "Access permissions and sharing"}}},
}

func (a *State) shellComponents(c *ui.Context, t nui.Theme) {
	heading(c, t, "Application layout", "A compact shell composed from native layout and navigation components.")
	ui.Column(c).Height(520).MinWidth(0).Border(1, t.Border).Radius(t.Radius).Clip().Children(func() {
		nui.HeaderLayout(c, t, nui.HeaderOptions{Title: "Review workspace", Actions: func(c *ui.Context, t nui.Theme) {
			nui.Button(c, t, "Workspace actions", nui.ButtonOptions{Variant: nui.Outline, Color: nui.Neutral}).OnClick(func() { a.Advanced.Command.Open = true })
		}})
		ui.Row(c).Grow(1).MinHeight(0).MinWidth(0).AlignItems(ui.Stretch).Children(func() {
			nui.Sidebar(c, t, &a.Advanced.Collapsed, nui.SidebarOptions{Title: "Workspace", Width: nui.Set(float32(180))}, func(c *ui.Context, t nui.Theme, compact bool) {
				nui.NavigationMenu(c, t, &a.Advanced.Active, &a.Advanced.Navigation, shellNavigation, nui.NavigationOptions{Label: "Workspace navigation", Vertical: true, Compact: compact, OnNavigate: func(id string) { a.Advanced.Result = "Opened " + id }})
			})
			nui.Splitter(c, t, &a.Advanced.Split, nui.SplitterOptions{Percent: true, Min: nui.Set(float32(25)), Max: nui.Set(float32(80))}, func() {
				ui.Scroll(c).Fill().Padding(12).Gap(12).Children(func() {
					ui.Text(c, "Recent activity").FontSize(18).FontWeight(600)
					nui.Chart(c, t, &a.Advanced.Chart, nui.ChartOptions{Label: "Shell activity", Height: 180})
					ui.Text(c, a.Advanced.Result).TextColor(t.Muted)
				})
			}, func() {
				ui.Scroll(c).Fill().Padding(12).Gap(12).Children(func() {
					ui.Text(c, "Review notes").FontWeight(600)
					ui.Text(c, "Resize the divider with the pointer or arrow keys. Navigation remains bound to the selected workspace section.")
				})
			}).Grow(1)
		})
		nui.FooterLayout(c, t, nui.FooterOptions{Text: "Local workspace · Ready", Actions: func(c *ui.Context, t nui.Theme) {
			nui.Button(c, t, "Inspect", nui.ButtonOptions{Variant: nui.LinkVariant}).OnClick(func() { a.Advanced.Drawer.Open = true })
		}})
	})
	nui.FooterColumns(c, t, []nui.FooterColumn{{ID: "workspace", Title: "Workspace", Items: []nui.NavigationItem{{ID: "activity", Label: "Activity overview"}, {ID: "attachments", Label: "Attachment rules"}}}, {ID: "guide", Title: "Component guide", Items: []nui.NavigationItem{{ID: "commands", Label: "Command palette"}, {ID: "layout", Label: "Layout components"}}}}, nui.FooterColumnsOptions{OnNavigate: func(id string) { a.Advanced.Result = "Opened " + id }})
}

// Dispose releases window-owned work when the containing native window closes.
func (a *State) Dispose() {
	a.Advanced.Command.Close()
	a.Advanced.Files.Dispose()
	a.Props.Files.Dispose()
}
