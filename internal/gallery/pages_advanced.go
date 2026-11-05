package gallery

import (
	"errors"
	"fmt"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

func (a *State) advancedPage(id string) *componentPage {
	p := &componentPage{}
	add := func(id, title, description string, build func(*ui.Context, nui.Theme)) {
		p.Scenes = append(p.Scenes, scene(id, title, description, build))
	}
	switch id {
	case "command-palette":
		for _, kind := range []string{"search", "groups", "empty"} {
			state := &nui.CommandPaletteState{}
			result := "No command selected"
			commands := []nui.CommandAction{{ID: "activity", Label: "Open activity", Group: "Workspace", Keywords: "chart graph metrics", OnSelect: func() { a.Navigate("chart") }}, {ID: "attachments", Label: "Review attachments", Group: "Workspace", OnSelect: func() { a.Navigate("file-upload") }}, {ID: "appearance", Label: "Appearance", Group: "Preferences", Children: []nui.CommandAction{{ID: "light", Label: "Use light theme", OnSelect: func() { a.Dark = false; a.System = false }}, {ID: "dark", Label: "Use dark theme", OnSelect: func() { a.Dark = true; a.System = false }}}}, {ID: "inspect", Label: "Inspect workspace", Group: "Workspace", Shortcut: "Ctrl Shift I", Key: ui.KeyI, Mods: ui.Ctrl | ui.Shift, OnSelect: func() { result = "Inspector command selected" }}, {ID: "publish", Label: "Publish", Disabled: true}}
			if kind == "empty" {
				commands = nil
			}
			state.SetIndex(nui.NewCommandIndex(commands))
			p.Deactivate = append(p.Deactivate, state.Close)
			add(kind, "Commands · "+kind, "Type to search labels and keywords, Enter activates, Escape cancels; grouped Appearance opens a child level. Ctrl K opens the first palette; Ctrl Shift I runs its inspector action. Navigation uses component IDs even when the directory is filtered.", func(c *ui.Context, t nui.Theme) {
				nui.Button(c, t, "Open "+kind+" commands", nui.ButtonOptions{Leading: lucide.Search()}).OnClick(func() { state.Open = true })
				nui.CommandPalette(c, t, state, nui.CommandPaletteOptions{Title: kind + " workspace commands", DisableShortcut: kind != "search", OnSelect: func(id string) { result = "Selected " + id }})
				ui.Text(c, result)
			})
		}
	case "navigation-menu":
		for _, kind := range []string{"horizontal", "vertical", "compact"} {
			active := "activity"
			state := &nui.NavigationState{}
			result := "Ready"
			items := append([]nui.NavigationItem{}, shellNavigation...)
			items = append(items, nui.NavigationItem{ID: "locked", Label: "Publish locked", Disabled: true, Icon: lucide.Lock()})
			add(kind, "Navigation · "+kind, "Arrows/Home/End move among items, Enter navigates. Compact icons expose names through tooltips and settings through a submenu. Selection is independent of the Gallery directory.", func(c *ui.Context, t nui.Theme) {
				nui.NavigationMenu(c, t, &active, state, items, nui.NavigationOptions{Label: kind + " workspace navigation", Vertical: kind != "horizontal", Compact: kind == "compact", OnNavigate: func(id string) { result = "Opened " + id }})
				ui.Text(c, result)
			})
		}
	case "drawer":
		for _, kind := range []string{"basic", "snap", "policy"} {
			state := &nui.DrawerState{}
			handle := &ui.Handle{}
			draft := "Review build"
			snaps := 0
			p.Deactivate = append(p.Deactivate, func() { state.Open = false })
			add(kind, "Drawer · "+kind, "Drag the handle or focus it and use Up/Down to move between snap points. Escape/outside click dismiss unless policy disables dismissal; Done always closes and restores the opener.", func(c *ui.Context, t nui.Theme) {
				nui.Button(c, t, "Open "+kind+" drawer", nui.ButtonOptions{}).Bind(handle).OnClick(func() { state.Open = true })
				points := []float32{.45, .8}
				if kind == "snap" {
					points = []float32{.3, .55, .9}
				}
				nui.Drawer(c, t, state, nui.DrawerOptions{ModalOptions: nui.ModalOptions{Title: kind + " inspector", ReturnFocus: handle, Dismissible: nui.Set(kind != "policy"), Footer: func(c *ui.Context, t nui.Theme) {
					nui.Button(c, t, "Done "+kind+" reviewing", nui.ButtonOptions{}).OnClick(func() { state.Open = false })
				}}, SnapPoints: points, OnSnap: func(int) { snaps++ }}, func(t nui.Theme) {
					nui.Input(c, t, &draft, nui.InputOptions{Label: kind + " review notes"})
					for i := range 12 {
						ui.Textf(c, "Run %03d · ready", i+1)
					}
				})
				ui.Textf(c, "Snap %d · %d changes", state.Snap, snaps)
			})
		}
	case "file-upload":
		for _, kind := range []string{"native", "rules", "simulation"} {
			state := &nui.FileUploadState{}
			ticket := nui.UploadTicket{}
			result := "Files remain local"
			progress := 0.0
			preview, remove, grid := true, true, false
			serial := 0
			p.Deactivate = append(p.Deactivate, state.Dispose)
			add(kind, "Attachments · "+kind, "Choose local files using this window's native dialog. Accept .txt/.md/.png, at most 5 files and 10 MB each. Simulation uses explicit progress/finish/fail actions; leaving or hiding cancels jobs and rejects late reports. No network upload is performed.", func(c *ui.Context, t nui.Theme) {
				rules := nui.UploadRules{Accept: []string{".txt", ".md", ".png"}, MaxFiles: 5, MaxBytes: 10 << 20}
				layout := nui.UploadList
				if grid {
					layout = nui.UploadGrid
				}
				ui.Row(c).Wrap().Gap(8).Children(func() {
					nui.Switch(c, t, &preview, nui.SwitchOptions{Label: kind + " file previews"})
					nui.Switch(c, t, &remove, nui.SwitchOptions{Label: kind + " allow removal"})
					nui.Switch(c, t, &grid, nui.SwitchOptions{Label: kind + " grid layout"})
				})
				nui.FileUpload(c, t, state, nui.FileUploadOptions{Label: kind + " review attachments", ChooseLabel: "Choose " + kind + " files", Dialog: a.FileDialog, Rules: rules, Layout: layout, Preview: nui.Set(preview), FileDelete: nui.Set(remove), OnRemove: func(string) { result = "File removed" }, OnCancel: func(string) { result = "Upload cancelled" }})
				ui.Row(c).Wrap().Gap(8).Children(func() {
					nui.Button(c, t, "Add "+kind+" sample", nui.ButtonOptions{}).OnClick(func() {
						serial++
						state.Add([]nui.UploadFile{{Path: fmt.Sprintf("sample/%s-%d.md", kind, serial), Name: fmt.Sprintf("review-%d.md", serial), Size: 2048, MIME: "text/markdown"}}, rules)
					})
					if kind == "rules" {
						nui.Button(c, t, "Add rejected executable", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() {
							state.Add([]nui.UploadFile{{Path: "sample/application.exe", Name: "application.exe", Size: 20 << 20}}, rules)
						})
					}
					if kind == "simulation" {
						nui.Button(c, t, "Start simulated upload", nui.ButtonOptions{Disabled: len(state.Files) == 0}).OnClick(func() { _, ticket = state.Begin(state.Files[0].ID); progress = 0; result = "Simulation started" })
						nui.Button(c, t, "Advance progress", nui.ButtonOptions{}).OnClick(func() {
							progress = min(1, progress+.25)
							if state.Report(ticket, progress, progress == 1, nil) {
								result = fmt.Sprintf("Simulated progress %.0f%%", progress*100)
							} else {
								result = "Stale upload result rejected"
							}
						})
						nui.Button(c, t, "Fail simulated upload", nui.ButtonOptions{Color: nui.Error}).OnClick(func() {
							if state.Report(ticket, progress, false, errors.New("simulated transport failure")) {
								result = "Simulation failed"
							} else {
								result = "Stale upload result rejected"
							}
						})
						nui.Button(c, t, "Cancel simulated upload", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { state.Dispose(); result = "Simulation cancelled" })
					}
				})
				ui.Text(c, result)
			})
		}
	case "chart":
		model := &State{Advanced: newAdvanced()}
		add("explorer", "Interactive chart explorer", "Compare all six chart kinds, numeric/category/time axes, five curves and four stacking modes. Click legends or focus the chart and use arrows. Update changes data without recreating the model; negative values and missing samples are cartesian only.", model.chartDemo)
		for _, kind := range []string{"time", "donut"} {
			state := newAdvanced()
			state.ChartAxis = kind
			if kind == "time" {
				state.ChartKind = nui.ChartBar
				state.ChartNegative = true
				state.updateChartData()
			} else {
				state.ChartKind = nui.ChartDonut
			}
			revision := 0
			add(kind, "Chart · "+kind, "This independent dataset retains legend visibility across builds. Time values use months in 2026; radial slices use stable IDs. Update changes only this scene.", func(c *ui.Context, t nui.Theme) {
				nui.Chart(c, t, &state.Chart, nui.ChartOptions{Label: kind + " activity chart", Kind: state.ChartKind, Height: 220, FormatTime: func(v time.Time) string { return v.Format("Jan 2006") }})
				nui.Button(c, t, "Update "+kind+" dataset", nui.ButtonOptions{}).OnClick(func() { revision++; state.ChartRevision = revision; state.updateChartData() })
				ui.Textf(c, "Dataset %d", revision+1)
			})
		}
	case "carousel":
		for _, kind := range []string{"horizontal", "vertical", "autoplay"} {
			state := &nui.CarouselState{}
			arrows, dots, wrap := true, true, kind != "vertical"
			changes := 0
			items := []nui.CarouselItem{}
			for i, label := range []string{"Collect files", "Inspect activity", "Review changes"} {
				items = append(items, nui.CarouselItem{ID: fmt.Sprint(i), Label: label, Build: func(c *ui.Context, t nui.Theme) { ui.Text(c, label).FontSize(22); ui.Textf(c, "Step %d of 3", i+1) }})
			}
			add(kind, "Carousel · "+kind, "Use arrow keys, drag or navigation buttons/dots. Wrap controls explicit and automatic movement; drag remains bounded. Autoplay respects visibility, pause, window inactivity and reduced motion.", func(c *ui.Context, t nui.Theme) {
				ui.Row(c).Wrap().Gap(8).Children(func() {
					nui.Switch(c, t, &arrows, nui.SwitchOptions{Label: kind + " arrows"})
					nui.Switch(c, t, &dots, nui.SwitchOptions{Label: kind + " dots"})
					nui.Switch(c, t, &wrap, nui.SwitchOptions{Label: kind + " wrap"})
					nui.Button(c, t, "Pause / resume "+kind, nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { state.Paused = !state.Paused })
				})
				o := nui.CarouselOptions{Label: kind + " review guide", Height: 180, Arrows: nui.Set(arrows), Dots: dots, Wrap: wrap, Active: nui.Set(!a.WindowInactive && a.Animations), OnChange: func(int) { changes++ }}
				if kind == "vertical" {
					o.Orientation = nui.Vertical
				}
				if kind == "autoplay" {
					o.Autoplay = 4 * time.Second
				}
				nui.Carousel(c, t, state, items, o)
				ui.Textf(c, "Slide %d · %d changes · paused %t", state.Index+1, changes, state.Paused)
			})
		}
	case "marquee":
		for _, kind := range []string{"horizontal", "vertical", "static"} {
			state := &nui.MarqueeState{}
			reverse := false
			speed := float64(28)
			add(kind, "Marquee · "+kind, "Toggle direction and pause/resume. Moving content is presentation-only, pauses on hover, and respects reduced motion. Static stays still regardless of the global animation preference.", func(c *ui.Context, t nui.Theme) {
				nui.Switch(c, t, &reverse, nui.SwitchOptions{Label: kind + " reverse"})
				nui.Slider(c, t, &speed, nui.SliderOptions{Label: kind + " speed", Min: nui.Set(10.0), Max: nui.Set(60.0), Step: 5})
				nui.Button(c, t, "Pause / resume "+kind, nui.ButtonOptions{}).OnClick(func() { state.Paused = !state.Paused })
				nui.Marquee(c, t, state, nui.MarqueeOptions{Label: kind + " updates", ContentKey: "gallery-updates-v1", Vertical: kind == "vertical", Reverse: reverse, Speed: float32(speed), Height: 80, PauseOnHover: true, Active: nui.Set(kind != "static" && !a.WindowInactive && a.Animations)}, func() {
					ui.Text(c, "Build passed    •    Queue updated    •    Files stay local    •    Ready for review").NoWrap()
				})
				ui.Textf(c, "Paused: %t", state.Paused)
			})
		}
	case "splitter":
		for _, kind := range []string{"horizontal", "vertical", "disabled"} {
			width, height := float32(40), float32(55)
			minimum, maximum := float32(25), float32(80)
			title := "Left pane + stacked right panes"
			if kind == "vertical" {
				title = "Constrained nested panes"
				minimum, maximum = 35, 65
			}
			if kind == "disabled" {
				title = "Disabled nested panes"
			}
			changes := 0
			draft := "Persistent pane draft"
			notes := "Independent notes"
			add(kind, title, fmt.Sprintf("The left pane sits beside a right pane split into top and bottom. Drag the vertical divider or use Left/Right for width; drag the horizontal divider or use Up/Down for height. Both positions are constrained to %.0f–%.0f%%; native panes retain a 40 DIP minimum.", minimum, maximum), func(c *ui.Context, t nui.Theme) {
				o := nui.SplitterOptions{Percent: true, Disabled: kind == "disabled", Min: nui.Set(minimum), Max: nui.Set(maximum), OnChange: func(float32) { changes++ }}
				nui.Splitter(c.Key("pane-widths"), t, &width, o, func() {
					ui.Scroll(c.Key("editor")).Fill().Padding(12).Gap(8).Children(func() {
						ui.Text(c, "Left pane").Bold()
						nui.Input(c.Key("draft"), t, &draft, nui.InputOptions{Label: kind + " pane draft"})
						ui.Text(c, "Edit here to update the preview on the right.")
					})
				}, func() {
					vertical := o
					vertical.Vertical = true
					nui.Splitter(c.Key("pane-heights"), t, &height, vertical, func() {
						ui.Scroll(c.Key("preview")).Fill().Padding(12).Gap(8).Children(func() {
							ui.Text(c, "Right · top").Bold()
							ui.Text(c, "Live preview")
							ui.Text(c, draft)
						})
					}, func() {
						ui.Scroll(c.Key("notes")).Fill().Padding(12).Gap(8).Children(func() {
							ui.Text(c, "Right · bottom").Bold()
							ui.Text(c, "Notes")
							nui.Input(c.Key("notes"), t, &notes, nui.InputOptions{Label: kind + " pane notes"})
						})
					}).Fill().Label(kind + " pane heights")
				}).Height(360).Border(1, t.Border).Label(kind + " pane widths")
				ui.Textf(c, "Left width %.0f%% · Right top height %.0f%% · %d changes", width, height, changes)
			})
		}
	case "sidebar":
		for _, kind := range []string{"slideover", "modal", "drawer"} {
			state := &nui.SidebarState{}
			collapsed := false
			active := "activity"
			navigation := &nui.NavigationState{}
			right := false
			changes := 0
			draft := "Inner workspace"
			opener := &ui.Handle{}
			p.Deactivate = append(p.Deactivate, func() { state.Open = false })
			add(kind, "Sidebar · "+kind, "Each inner sidebar owns collapse, width and overlay state. At window widths below 880 DIP use its opener; selecting a route closes the overlay and returns focus. Resize on desktop; toggle side to compare RTL mirroring.", func(c *ui.Context, t nui.Theme) {
				nui.Switch(c, t, &right, nui.SwitchOptions{Label: kind + " sidebar on right"})
				nui.Button(c, t, "Open inner "+kind+" sidebar", nui.ButtonOptions{}).Bind(opener).OnClick(func() { state.Open = true })
				ui.Row(c).Height(290).Border(1, t.Border).MinWidth(0).Children(func() {
					side := nui.SideLeft
					if right {
						side = nui.SideRight
					}
					mode := nui.SidebarSlideover
					if kind == "modal" {
						mode = nui.SidebarModal
					}
					if kind == "drawer" {
						mode = nui.SidebarDrawer
					}
					buildSidebar := func() {
						nui.Sidebar(c, t, &collapsed, nui.SidebarOptions{Title: kind + " workspace sidebar", ToggleLabel: "Toggle inner " + kind + " sidebar", ResizeLabel: "Resize inner " + kind + " sidebar", Width: nui.Set(float32(220)), MinWidth: nui.Set(float32(160)), MaxWidth: nui.Set(float32(320)), State: state, Resizable: true, Side: nui.Set(side), Breakpoint: nui.Set(float32(880)), Mode: mode, NavigationKey: active, ReturnFocus: opener, OnResize: func(float32) { changes++ }, OnCollapsedChange: func(bool) { changes++ }}, func(c *ui.Context, t nui.Theme, compact bool) {
							nui.NavigationMenu(c, t, &active, navigation, shellNavigation, nui.NavigationOptions{Label: kind + " inner navigation", Vertical: true, Compact: compact})
						})
					}
					body := func() {
						ui.Column(c).Grow(1).MinWidth(0).Padding(12).Gap(12).Children(func() {
							ui.Text(c, "Selected: "+active)
							nui.Input(c, t, &draft, nui.InputOptions{Label: kind + " inner workspace name"})
						})
					}
					if right {
						body()
						buildSidebar()
					} else {
						buildSidebar()
						body()
					}
				})
				ui.Textf(c, "Collapsed %t · %d changes", collapsed, changes)
			})
		}
	case "header", "footer", "footer-columns":
		for _, kind := range []string{"simple", "actions", "constrained"} {
			result := "Ready"
			clicks := 0
			add(kind, "Layout · "+kind, "Compare basic content, local actions and a constrained 340 DIP composition. Actions update only this section; wrapped content remains reachable with larger text.", func(c *ui.Context, t nui.Theme) {
				root := ui.Column(c).MinWidth(0).Border(1, t.Border).Gap(8)
				if kind == "constrained" {
					root.MaxWidth(340)
				}
				root.Children(func() {
					action := func(c *ui.Context, t nui.Theme) {
						nui.Button(c, t, kind+" layout action", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { clicks++; result = "Action completed" })
					}
					switch id {
					case "header":
						o := nui.HeaderOptions{Title: "Review workspace", Description: "Native header with responsive actions"}
						if kind != "simple" {
							o.Actions = action
							o.Leading = func(c *ui.Context, t nui.Theme) { ui.Icon(c, lucide.FolderOpen()).Size(24, 24) }
						}
						nui.HeaderLayout(c, t, o)
						nui.Main(c, t, func() { ui.Text(c, "Workspace body").Padding(20) }).MinHeight(100)
					case "footer":
						nui.Main(c, t, func() {
							ui.Text(c, "Review local changes before sharing. Long body content remains above the status and action area.").Padding(20)
						}).MinHeight(100)
						o := nui.FooterOptions{Text: "Local workspace · Ready"}
						if kind != "simple" {
							o.Actions = action
						}
						nui.FooterLayout(c, t, o)
					case "footer-columns":
						nui.FooterColumns(c, t, []nui.FooterColumn{{ID: "workspace", Title: "Workspace", Items: []nui.NavigationItem{{ID: "activity", Label: "Activity overview"}, {ID: "attachments", Label: "Attachment rules"}}}, {ID: "guide", Title: "Component guide", Items: []nui.NavigationItem{{ID: "commands", Label: "Command palette"}, {ID: "layout", Label: "Layout components"}, {ID: "disabled", Label: "Unavailable", Disabled: true}}}}, nui.FooterColumnsOptions{MinColumnWidth: 140, OnNavigate: func(id string) { clicks++; result = "Opened " + id }})
					}
				})
				ui.Textf(c, "%s · %d actions", result, clicks)
			})
		}
	}
	return p
}
