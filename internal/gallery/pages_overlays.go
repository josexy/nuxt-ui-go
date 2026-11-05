package gallery

import (
	"fmt"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

func (a *State) overlayPage(id string) *componentPage {
	p := &componentPage{}
	add := func(id, title, description string, build func(*ui.Context, nui.Theme)) {
		p.Scenes = append(p.Scenes, scene(id, title, description, build))
	}
	switch id {
	case "tooltip":
		controlled := false
		add("hover", "Hover help", "Hover over the button to reveal its tooltip.", func(c *ui.Context, t nui.Theme) {
			anchor := nui.Button(c, t, "Workspace help", nui.ButtonOptions{Variant: nui.Outline})
			nui.Tooltip(c, t, anchor, "Read about workspace access", nui.TooltipOptions{})
		})
		add("keyboard", "Keyboard focus", "Tab to this button. Shortcut hints are formatted by the component.", func(c *ui.Context, t nui.Theme) {
			anchor := nui.Button(c, t, "Search shortcut", nui.ButtonOptions{Leading: lucide.Search()})
			nui.Tooltip(c, t, anchor, "Search", nui.TooltipOptions{Kbds: []string{"Ctrl", "K"}, Placement: nui.Set(nui.BottomStart)})
		})
		add("controlled", "Controlled and disabled", "The switch controls one tooltip. The disabled tooltip remains hidden.", func(c *ui.Context, t nui.Theme) {
			nui.Switch(c, t, &controlled, nui.SwitchOptions{Label: "Show controlled tooltip"})
			anchor := nui.Button(c, t, "Shortcut help", nui.ButtonOptions{Variant: nui.Ghost})
			nui.Tooltip(c, t, anchor, "Controlled search help", nui.TooltipOptions{Open: &controlled, Kbds: []string{"Ctrl", "K"}})
			disabled := nui.Button(c, t, "Disabled tooltip anchor", nui.ButtonOptions{Variant: nui.Outline})
			nui.Tooltip(c, t, disabled, "Hidden help", nui.TooltipOptions{Disabled: true})
		})
		p.Deactivate = append(p.Deactivate, func() { controlled = false })
	case "modal":
		for _, sample := range []struct {
			id, title                      string
			dismissible, close, fullscreen bool
		}{{"basic", "Basic dialog", true, true, false}, {"nested", "Edit workspace", true, true, false}, {"policy", "Explicit close and fullscreen", false, false, true}} {
			open := false
			plan, name := "personal", "Nuxt workspace"
			reviewer := nui.SearchState{}
			saved := 0
			var trigger ui.Handle
			add(sample.id, sample.title, "Open the dialog. Nested Select and InputMenu consume the first Escape; a dismissible dialog consumes the second. Focus returns to its trigger.", func(c *ui.Context, t nui.Theme) {
				label := "Open " + sample.id + " dialog"
				if sample.id == "nested" {
					label = "Edit workspace"
				}
				nui.Button(c, t, label, nui.ButtonOptions{}).Bind(&trigger).OnClick(func() { open = true })
				nui.ModalWithOptions(c, t, &open, nui.ModalOptions{Title: sample.title, Description: "Changes apply to this scenario only.", Dismissible: nui.Set(sample.dismissible), CloseButton: nui.Set(sample.close), Fullscreen: sample.fullscreen, ReturnFocus: &trigger, Footer: func(c *ui.Context, inner nui.Theme) {
					nui.Button(c, inner, "Cancel "+sample.id, nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { open = false })
					nui.Button(c, inner, "Save "+sample.id, nui.ButtonOptions{}).OnClick(func() { saved++; open = false })
				}}, func(inner nui.Theme) {
					nui.Input(c, inner, &name, nui.InputOptions{Label: "Dialog workspace name"}).FillWidth()
					if sample.id == "nested" {
						nui.Select(c, inner, &plan, plans, nui.SelectOptions{Label: "Dialog plan"}).FillWidth()
						nui.InputMenu(c, inner, &reviewer, propsOptions, nui.InputMenuOptions{InputOptions: nui.InputOptions{Label: "Reviewer"}}).FillWidth()
					} else {
						ui.Text(c, "A native modal focus scope.")
					}
				})
				ui.Textf(c, "Saved %d times", saved)
			})
			p.Deactivate = append(p.Deactivate, func() { open = false })
		}
	case "slideover":
		for _, sample := range []struct {
			id          string
			side        nui.Side
			dismissible bool
		}{{"right", nui.SideRight, true}, {"left", nui.SideLeft, true}, {"top", nui.SideTop, true}, {"bottom", nui.SideBottom, false}} {
			open := false
			name := "Workspace"
			var trigger ui.Handle
			add(sample.id, sample.id+" panel", "Open a side panel and edit its draft. The bottom panel requires its Done action; other panels allow Escape.", func(c *ui.Context, t nui.Theme) {
				nui.Button(c, t, "Open "+sample.id+" panel", nui.ButtonOptions{Variant: nui.Outline}).Bind(&trigger).OnClick(func() { open = true })
				nui.Slideover(c, t, &open, nui.SlideoverOptions{Side: sample.side, Extent: nui.Set(float32(380)), ModalOptions: nui.ModalOptions{Title: "Workspace preferences", Description: sample.id + " side", Dismissible: nui.Set(sample.dismissible), ReturnFocus: &trigger, Footer: func(c *ui.Context, inner nui.Theme) {
					nui.Button(c, inner, "Done with preferences", nui.ButtonOptions{}).OnClick(func() { open = false })
				}}}, func(inner nui.Theme) {
					nui.Input(c, inner, &name, nui.InputOptions{Label: sample.id + " panel name"}).FillWidth()
				})
			})
			p.Deactivate = append(p.Deactivate, func() { open = false })
		}
	case "toast":
		actions, closed, queued := 0, 0, 0
		add("colors", "Semantic notifications", "Each button queues a native notification in this window.", func(c *ui.Context, t nui.Theme) {
			comparison(c, func() {
				for _, v := range semanticColors {
					nui.Button(c.Key(v.Name), t, "Notify "+v.Name, nui.ButtonOptions{Color: v.Value, Variant: nui.Soft}).OnClick(func() {
						a.pendingToasts = append(a.pendingToasts, nui.ToastOptions{Title: v.Name + " notification", Color: v.Value})
					})
				}
			})
		})
		add("actions", "Actions and close", "Review keeps the toast open; Done closes it. Close callbacks update the result.", func(c *ui.Context, t nui.Theme) {
			nui.Button(c, t, "Show action toast", nui.ButtonOptions{}).OnClick(func() {
				a.pendingToasts = append(a.pendingToasts, nui.ToastOptions{ID: "gallery-action", Title: "Workspace ready", Description: "Review or dismiss this update.", Timeout: -1, OnClose: func() { closed++ }, Actions: []nui.ToastActionOptions{{Label: "Review", Close: nui.Set(false), OnAction: func() { actions++ }}, {Label: "Done"}}})
			})
			ui.Textf(c, "Reviews: %d · Closed: %d", actions, closed)
		})
		add("queue", "Queue and timeout", "Queue three 3-second notices. Hover or focus pauses a timeout. Leaving the page clears its notices.", func(c *ui.Context, t nui.Theme) {
			nui.Button(c, t, "Queue three notices", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() {
				for range 3 {
					queued++
					a.pendingToasts = append(a.pendingToasts, nui.ToastOptions{Title: fmt.Sprintf("Queued notice %d", queued), Timeout: 3 * time.Second})
				}
			})
			ui.Textf(c, "Queued: %d", queued)
		})
	case "dropdown-menu", "context-menu":
		for _, sample := range []struct{ id, title string }{{"basic", "Actions and disabled items"}, {"nested", "Nested export menu"}, {"checks", "Checked and radio items"}} {
			state := &nui.MenuState{}
			result := "No action"
			pinned := false
			format := "csv"
			add(sample.id, sample.title, "Use arrows, Enter and Escape. Context targets also support Shift+F10. Results use stable item IDs.", func(c *ui.Context, t nui.Theme) {
				items := []nui.MenuItem{{ID: "workspace", Label: "Workspace", Kind: nui.MenuLabel}, {ID: "open", Label: "Open workspace", Description: "Continue editing", Shortcut: "Ctrl+O", OnSelect: func() { result = "Workspace opened" }}, {ID: "unavailable", Label: "Unavailable action", Disabled: true}}
				if sample.id == "nested" {
					items = append(items, nui.MenuItem{ID: "export", Label: "Export", Children: []nui.MenuItem{{ID: "csv", Label: "CSV file", OnSelect: func() { result = "CSV export prepared" }}, {ID: "json", Label: "JSON file", OnSelect: func() { result = "JSON export prepared" }}, {ID: "pdf", Label: "PDF report", Disabled: true}}})
				}
				if sample.id == "checks" {
					items = append(items, nui.MenuItem{ID: "pin", Label: "Pin workspace", Checked: nui.Set(pinned), OnSelect: func() { pinned = !pinned; result = fmt.Sprintf("Pinned: %t", pinned) }})
					for _, v := range []string{"csv", "json"} {
						items = append(items, nui.MenuItem{ID: v, Label: "Format " + v, Radio: true, Checked: nui.Set(format == v), OnSelect: func() { format = v; result = "Format: " + v }})
					}
				}
				if id == "dropdown-menu" {
					anchor := nui.Button(c, t, "Project actions "+sample.id, nui.ButtonOptions{Variant: nui.Outline, Trailing: lucide.ChevronDown()})
					nui.DropdownMenu(c, t, anchor, state, items, nui.MenuOptions{Label: "Project menu " + sample.id, Width: nui.Set(float32(280))})
				} else {
					target := ui.Column(c).FillWidth().Padding(20).Border(1, t.Border).Radius(8).Label("Project context " + sample.id).Children(func() { ui.Text(c, "Right-click or press Shift+F10 here.") })
					nui.ContextMenu(c, t, target, state, items, nui.MenuOptions{Label: "Project context menu " + sample.id})
				}
				ui.Text(c, result).Role(ui.RoleStatus)
			})
			p.Deactivate = append(p.Deactivate, func() { *state = nui.MenuState{} })
		}
		if id == "dropdown-menu" {
			model := &State{}
			add("rich", "Custom rich actions", "A PopoverBase composition with a nested export panel; built-in DropdownMenu scenarios above provide standard keyboard menu behavior.", func(c *ui.Context, t nui.Theme) {
				anchor := nui.Button(c, t, "Rich actions", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { model.RichOpen = true })
				model.richActions(c, t, anchor)
				ui.Text(c, model.MenuResult)
			})
			p.Deactivate = append(p.Deactivate, func() { model.RichOpen = false; model.RichSubOpen = false })
		}
	case "tabs":
		for _, sample := range []struct {
			id, title                         string
			underline, vertical, manual, keep bool
		}{{"pill", "Pill tabs", false, false, false, false}, {"underline", "Vertical underline tabs", true, true, false, false}, {"manual", "Manual activation and mounted content", false, false, true, true}} {
			selected, draft := "overview", "Persistent tab draft"
			changes := 0
			add(sample.id, sample.title, "Arrows move focus. In manual mode press Enter or Space to select. A disabled tab is skipped; the model preserves edited text.", func(c *ui.Context, t nui.Theme) {
				mode := nui.ActivationAutomatic
				if sample.manual {
					mode = nui.ActivationManual
				}
				nui.Tabs(c, t, &selected, []nui.TabItem{{ID: "overview", Label: "Overview", Content: func(c *ui.Context, inner nui.Theme) { ui.Text(c, "Overview selected") }}, {ID: "settings", Label: "Settings", Content: func(c *ui.Context, inner nui.Theme) {
					nui.Input(c, inner, &draft, nui.InputOptions{Label: sample.id + " tab draft"}).FillWidth()
				}}, {ID: "locked", Label: "Audit export", Disabled: true}}, nui.TabsOptions{Label: sample.title, Underline: sample.underline, Vertical: sample.vertical, ActivationMode: mode, KeepMounted: sample.keep, OnChange: func(string) { changes++ }})
				ui.Textf(c, "Selected: %s · Changes: %d", selected, changes)
			})
		}
	case "collapsible":
		for _, sample := range []struct {
			id, title      string
			disabled, keep bool
		}{{"controlled", "Controlled disclosure", false, false}, {"disabled", "Disabled disclosure", true, false}, {"mounted", "Keep mounted content", false, true}} {
			open := false
			draft := "Retained model"
			changes := 0
			add(sample.id, sample.title, "Toggle with Enter or Space. KeepMounted retains widget content while closed; the application draft survives either setting.", func(c *ui.Context, t nui.Theme) {
				nui.Collapsible(c, t, &open, nui.CollapsibleOptions{Label: sample.title, Disabled: sample.disabled, KeepMounted: sample.keep, OnChange: func(bool) { changes++ }}, func(inner nui.Theme) {
					nui.Input(c, inner, &draft, nui.InputOptions{Label: sample.id + " disclosure draft"}).FillWidth()
				})
				ui.Textf(c, "Open: %t · Changes: %d", open, changes)
			})
		}
	case "accordion":
		for _, sample := range []struct {
			id, title                   string
			multiple, collapsible, keep bool
		}{{"single", "Single open section", false, true, false}, {"multiple", "Multiple open sections", true, true, false}, {"required", "Keep one section open", false, false, true}} {
			open := []string{"access"}
			draft := "Shared with desktop team"
			changes := 0
			add(sample.id, sample.title, "Use arrows between headings and Enter to toggle. Billing is disabled.", func(c *ui.Context, t nui.Theme) {
				nui.Accordion(c, t, &open, []nui.AccordionItem{{ID: "access", Label: "Sharing permissions", Content: func(c *ui.Context, inner nui.Theme) {
					nui.Input(c, inner, &draft, nui.InputOptions{Label: sample.id + " sharing draft"}).FillWidth()
				}}, {ID: "history", Label: "Version history", Content: func(c *ui.Context, inner nui.Theme) {
					ui.Text(c, "This longer description wraps with the available width. Draft preferences remain available when the section is rebuilt.")
				}}, {ID: "billing", Label: "Billing settings", Disabled: true}}, nui.AccordionOptions{Multiple: sample.multiple, Collapsible: nui.Set(sample.collapsible), KeepMounted: sample.keep, OnChange: func() { changes++ }})
				ui.Textf(c, "Open IDs: %v · Changes: %d", open, changes)
			})
		}
	case "breadcrumb":
		for _, sample := range []struct {
			id, title string
			max       int
		}{{"basic", "Basic path", 0}, {"truncated", "Truncated path", 3}, {"actions", "Navigation callbacks", 4}} {
			result := "Current: Workspace review"
			add(sample.id, sample.title, "Activate a parent crumb to update the local result. The final crumb represents the current location.", func(c *ui.Context, t nui.Theme) {
				var items []nui.BreadcrumbItem
				for i, label := range []string{"Home", "Projects", "Desktop", "Drafts", "Workspace review"} {
					it := nui.BreadcrumbItem{ID: fmt.Sprintf("crumb-%d", i), Label: label}
					if i < 4 {
						it.OnSelect = func() { result = label + " selected" }
					}
					items = append(items, it)
				}
				nui.Breadcrumb(c, t, items, nui.BreadcrumbOptions{Label: sample.title, MaxItems: sample.max})
				ui.Text(c, result)
			})
		}
	case "pagination":
		for _, sample := range []struct {
			id, title                 string
			siblings                  int
			controls, edges, disabled bool
			size                      nui.Size
		}{{"basic", "Page navigation", 1, true, true, false, nui.MD}, {"zero", "Explicit zero neighbors", 0, true, false, false, nui.SM}, {"states", "Disabled soft pagination", 2, false, true, true, nui.LG}} {
			page := 5
			changes := 0
			add(sample.id, sample.title, "Use the controls or numbered buttons. SiblingCount=Set(0) intentionally shows no neighbors.", func(c *ui.Context, t nui.Theme) {
				nui.Pagination(c, t, &page, nui.PaginationOptions{Label: sample.title, Total: 100, PageSize: 10, SiblingCount: nui.Set(sample.siblings), ShowControls: nui.Set(sample.controls), ShowEdges: nui.Set(sample.edges), Disabled: sample.disabled, Size: sample.size, Variant: nui.Soft, OnChange: func(int) { changes++ }})
				ui.Textf(c, "Page %d · Changes: %d", page, changes)
			})
		}
	case "stepper":
		for _, sample := range []struct {
			id, title        string
			linear, vertical bool
		}{{"linear", "Linear workflow", true, false}, {"nonlinear", "Nonlinear workflow", false, false}, {"vertical", "Vertical workflow", false, true}} {
			step := "details"
			changes := 0
			add(sample.id, sample.title, "Linear steps unlock in order. The unavailable final step cannot activate.", func(c *ui.Context, t nui.Theme) {
				nui.Stepper(c, t, &step, []nui.StepItem{{ID: "details", Title: "Details", Description: "Describe the workspace"}, {ID: "review", Title: "Review", Description: "Check access"}, {ID: "approve", Title: "Approve", Description: "Confirm changes"}, {ID: "locked", Title: "Publish", Disabled: true}}, nui.StepperOptions{Label: sample.title, Linear: nui.Set(sample.linear), Vertical: sample.vertical, OnChange: func(string) { changes++ }})
				ui.Textf(c, "Current step: %s · Changes: %d", step, changes)
			})
		}
	case "timeline":
		items := []nui.TimelineItem{{ID: "created", Time: "09:10", Title: "Workspace created", Description: "Lin created the workspace.", Icon: lucide.Check()}, {ID: "saved", Time: "10:25", Title: "Preferences saved", Description: "Region updated.", Icon: lucide.Save()}, {ID: "reviewed", Time: "11:40", Title: "Review started", Description: "Alex is checking access.", Icon: lucide.UserRound(), Color: nui.Info}}
		add("events", "Default events", "Each event has a stable identity, time and description.", func(c *ui.Context, t nui.Theme) {
			nui.Timeline(c, t, items, nui.TimelineOptions{Label: "Workspace events"})
		})
		selected := "saved"
		add("current", "Current event and selection", "Click an event to update the current ID.", func(c *ui.Context, t nui.Theme) {
			nui.Timeline(c, t, items, nui.TimelineOptions{Label: "Selectable events", CurrentID: selected, OnSelect: func(id string) { selected = id }})
			ui.Text(c, "Current event: "+selected)
		})
		add("horizontal", "Horizontal timeline", "A bounded horizontal scroll keeps all events reachable in a narrow preview.", func(c *ui.Context, t nui.Theme) {
			nui.ScrollArea(c, t, nui.ScrollAreaOptions{Direction: nui.ScrollHorizontal}, func() {
				nui.Timeline(c, t, items, nui.TimelineOptions{Label: "Horizontal events", Horizontal: true, Current: nui.Set(1)}).MinWidth(600)
			}).Height(200).FillWidth()
		})
	case "banner":
		actions := 0
		open := true
		closed := 0
		add("colors", "Semantic banners", "Banners provide a prominent status strip.", func(c *ui.Context, t nui.Theme) {
			for _, v := range semanticColors {
				nui.Banner(c.Key(v.Name), t, nil, nui.BannerOptions{Title: v.Name + " update", Color: v.Value, Icon: lucide.Info()})
			}
		})
		add("actions", "Action result", "The action callback updates only this scene.", func(c *ui.Context, t nui.Theme) {
			nui.Banner(c, t, nil, nui.BannerOptions{Title: "Review requested", Description: "Check the draft before publishing.", Actions: func(c *ui.Context, inner nui.Theme) {
				nui.Button(c, inner, "Review banner", nui.ButtonOptions{Size: nui.SM}).OnClick(func() { actions++ })
			}})
			ui.Textf(c, "Reviews: %d", actions)
		})
		add("dismiss", "Close and reopen", "The close button hides this banner; Reopen restores it.", func(c *ui.Context, t nui.Theme) {
			nui.Button(c, t, "Reopen banner", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { open = true })
			nui.Banner(c, t, &open, nui.BannerOptions{Title: "Dismissible update", CloseLabel: "Close sample banner", OnClose: func() { closed++ }})
			ui.Textf(c, "Open: %t · Closed: %d", open, closed)
		})
	case "error":
		retries := 0
		recovered := false
		add("status", "Status code", "An embedded error keeps its message within the preview.", func(c *ui.Context, t nui.Theme) {
			nui.ErrorView(c, t, nui.ErrorOptions{Status: "404", Title: "Project not found", Description: "The local fixture has no record with this ID.", Decoration: nui.Decoration{UI: nui.Styles{nui.Label: {FontSize: nui.Set(float32(24))}}}})
		})
		add("retry", "Retry counter", "Retry is simulated and performs no network request.", func(c *ui.Context, t nui.Theme) {
			nui.ErrorView(c, t, nui.ErrorOptions{Status: "503", Title: "Activity unavailable", OnRetry: func() { retries++ }, Decoration: nui.Decoration{UI: nui.Styles{nui.Label: {FontSize: nui.Set(float32(24))}}}})
			ui.Textf(c, "Retried %d times", retries)
		})
		add("recovery", "Recover to content", "The retry action replaces the error with a success alert.", func(c *ui.Context, t nui.Theme) {
			if !recovered {
				nui.ErrorView(c, t, nui.ErrorOptions{Title: "Activity details could not load", Description: "Try loading the local preview again.", RetryLabel: "Restore activity", OnRetry: func() { recovered = true }, Decoration: nui.Decoration{UI: nui.Styles{nui.Label: {FontSize: nui.Set(float32(24))}}}})
			} else {
				nui.Alert(c, t, nui.AlertOptions{Title: "Activity restored", Color: nui.Success})
				nui.Button(c, t, "Reset recovery", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { recovered = false })
			}
		})
	}
	return p
}
