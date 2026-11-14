package gallery

import (
	"fmt"
	"slices"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func (a *State) overlayPage(id string) *componentPage {
	p := &componentPage{}
	add := func(id, title, description string, build func(*ui.Context, nui.Theme)) {
		p.Scenes = append(p.Scenes, scene(id, title, description, build))
	}
	if id == "collapsible" || id == "accordion" {
		open, keep, disabled, reduce := true, true, false, false
		ids := []string{"draft"}
		draft := "Retained disclosure draft"
		actions, changes := 0, 0
		add("contract", "Heading and content state", "Disable the heading while using its body, change mounting policy, or close it externally. Custom icons and item styles remain independent of the stored draft.", func(c *ui.Context, t nui.Theme) {
			nui.Switch(c.Key("disable"), t, &disabled, nui.SwitchOptions{Label: "Disable contract heading"})
			nui.Switch(c.Key("keep"), t, &keep, nui.SwitchOptions{Label: "Keep contract disclosure"})
			nui.Switch(c.Key("reduce"), t, &reduce, nui.SwitchOptions{Label: "Reduce disclosure motion"})
			local := t
			local.Motion.Reduce = local.Motion.Reduce || reduce
			body := func(c *ui.Context, inner nui.Theme) {
				nui.Input(c, inner, &draft, nui.InputOptions{Label: "Contract disclosure draft"}).FillWidth()
				nui.Button(c, inner, "Contract body action", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { actions++ })
			}
			if id == "collapsible" {
				nui.Collapsible(c.Key("contract"), local, &open, nui.CollapsibleOptions{Label: "Contract disclosure heading", Icon: lucide.Settings(), TrailingIcon: lucide.Plus(), Disabled: disabled, KeepMounted: keep, OnChange: func(bool) { changes++ }}, func(inner nui.Theme) { body(c, inner) })
			} else {
				nui.Accordion(c.Key("contract"), local, &ids, []nui.AccordionItem{{ID: "draft", Label: "Contract disclosure heading", Icon: lucide.Settings(), TrailingIcon: lucide.ChevronRight(), UI: nui.Styles{nui.Trigger: {Foreground: nui.Set(t.Info)}}, Content: body}, {ID: "empty", Label: "Contract empty heading"}, {ID: "disabled", Label: "Contract disabled heading", Disabled: true}}, nui.AccordionOptions{TrailingIcon: lucide.Plus(), Disabled: disabled, KeepMounted: keep, OnChange: func() { changes++ }})
				open = slices.Contains(ids, "draft")
			}
			nui.Button(c, t, "Toggle contract disclosure externally", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() {
				open = !open
				if open {
					ids = []string{"draft"}
				} else {
					ids = nil
				}
			})
			ui.Textf(c, "Disclosure open: %t · Changes: %d · Actions: %d", open, changes, actions)
		})
	}
	switch id {
	case "tooltip":
		controlled := false
		instant, narrow, keysOnly, policyDisabled, policyOpen := false, false, false, false, false
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
		add("timing", "Delay, width and key hints", "Hover for 200 ms, or use Tab for immediate help. Escape and outside presses close the tooltip.", func(c *ui.Context, t nui.Theme) {
			nui.Switch(c.Key("instant"), t, &instant, nui.SwitchOptions{Label: "Instant tooltip"})
			nui.Switch(c.Key("narrow"), t, &narrow, nui.SwitchOptions{Label: "Narrow tooltip"})
			nui.Switch(c.Key("keys-only"), t, &keysOnly, nui.SwitchOptions{Label: "Shortcut keys only"})
			nui.Switch(c.Key("disable"), t, &policyDisabled, nui.SwitchOptions{Label: "Disable timing tooltip"})
			delay, width := 200*time.Millisecond, float32(260)
			if instant {
				delay = 0
			}
			if narrow {
				width = 160
			}
			nui.Button(c.Key("toggle"), t, "Toggle timing tooltip", nui.ButtonOptions{Disabled: policyDisabled, Variant: nui.Outline}).OnClick(func() { policyOpen = !policyOpen })
			anchor := nui.Button(c.Key("timing-anchor"), t, "Timing tooltip anchor", nui.ButtonOptions{Variant: nui.Subtle}).Description("Workspace action")
			text := "Search workspace files and settings"
			if keysOnly {
				text = ""
			}
			nui.Tooltip(c.Key("timing-tip"), t, anchor, text, nui.TooltipOptions{Open: &policyOpen, DelayDuration: nui.Set(delay), MaxWidth: nui.Set(width), Placement: nui.Set(nui.TopStart), Kbds: []string{"Ctrl", "K"}, Disabled: policyDisabled})
			ui.Text(c, fmt.Sprintf("Timing tooltip open: %t", policyOpen))
		})
		p.Deactivate = append(p.Deactivate, func() { controlled, policyOpen = false, false })
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
			description := map[string]string{
				"basic":  "Open the dialog and edit its name. Escape, the close button or a footer action closes it and returns focus to its trigger.",
				"nested": "Open the dialog. An open Select or InputMenu consumes the first Escape; the next Escape closes the dialog. Focus returns to its trigger.",
				"policy": "This fullscreen dialog has no close button and ignores Escape and outside clicks. Use Cancel or Save in the footer to close it and restore focus.",
			}[sample.id]
			add(sample.id, sample.title, description, func(c *ui.Context, t nui.Theme) {
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
						nui.InputMenu(c, inner, &reviewer, propsOptions, nui.InputMenuOptions{Label: "Reviewer"}).FillWidth()
					} else {
						ui.Text(c, "A native modal focus scope.")
					}
				})
				ui.Textf(c, "Saved %d times", saved)
			})
			p.Deactivate = append(p.Deactivate, func() { open = false })
		}
		open, noHeader, fullscreen, blockDismiss, transparent, noMotion := false, false, false, false, false, false
		draft := "Audit draft"
		closed, prevented := 0, 0
		var trigger ui.Handle
		add("contract", "Header, scrolling and dismissal", "The body scrolls independently. Header and fullscreen changes retain the editor selection; dismissal counters show each action.", func(c *ui.Context, t nui.Theme) {
			nui.Switch(c.Key("blocked"), t, &blockDismiss, nui.SwitchOptions{Label: "Block dialog dismissal"})
			nui.Switch(c.Key("transparent"), t, &transparent, nui.SwitchOptions{Label: "Transparent dialog backdrop"})
			nui.Switch(c.Key("motion"), t, &noMotion, nui.SwitchOptions{Label: "Disable dialog motion"})
			nui.Button(c.Key("open-audit"), t, "Open audited dialog", nui.ButtonOptions{Variant: nui.Outline}).Bind(&trigger).OnClick(func() { open = true })
			title, description, closeButton := "Audited dialog", "Edit the draft and compare header layouts.", true
			if noHeader {
				title, description, closeButton = "", "", false
			}
			nui.ModalWithOptions(c.Key("audit-dialog"), t, &open, nui.ModalOptions{Title: title, Description: description, CloseButton: nui.Set(closeButton), Fullscreen: fullscreen, Overlay: nui.Set(!transparent), Transition: nui.Set(!noMotion), Dismissible: nui.Set(!blockDismiss), Close: &nui.ButtonOptions{AccessibleLabel: "Close audited dialog", Size: nui.XL, Color: nui.Info, Variant: nui.Soft}, CloseIcon: lucide.Check(), ReturnFocus: &trigger, OnClose: func() { closed++ }, OnClosePrevent: func() { prevented++ }, Footer: func(c *ui.Context, inner nui.Theme) {
				nui.Button(c.Key("header"), inner, "Toggle dialog header", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() { noHeader = !noHeader })
				nui.Button(c.Key("fullscreen"), inner, "Toggle dialog fullscreen", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() { fullscreen = !fullscreen })
				nui.Button(c.Key("done"), inner, "Finish dialog review", nui.ButtonOptions{}).OnClick(func() { open = false })
			}}, func(inner nui.Theme) {
				nui.Input(c.Key("draft"), inner, &draft, nui.InputOptions{Label: "Audited dialog draft"})
				for i := range 18 {
					ui.Text(c.Key(i), fmt.Sprintf("Review detail %02d", i+1)).Height(32)
				}
			})
			ui.Text(c, fmt.Sprintf("Dialog notifications: closed=%d prevented=%d", closed, prevented))
			ui.Text(c, "Dialog draft: "+draft)
		})
		p.Deactivate = append(p.Deactivate, func() { open = false })
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
				nui.Slideover(c, t, &open, nui.SlideoverOptions{Side: sample.side, Extent: nui.Set(float32(380)), Title: "Workspace preferences", Description: sample.id + " side", Dismissible: nui.Set(sample.dismissible), ReturnFocus: &trigger, Footer: func(c *ui.Context, inner nui.Theme) {
					nui.Button(c, inner, "Done with preferences", nui.ButtonOptions{}).OnClick(func() { open = false })
				}}, func(inner nui.Theme) {
					nui.Input(c, inner, &name, nui.InputOptions{Label: sample.id + " panel name"}).FillWidth()
				})
			})
			p.Deactivate = append(p.Deactivate, func() { open = false })
		}
		open, inset, noMotion, blocked, customBackdrop := false, false, false, false, false
		side := nui.SideRight
		draft := "Panel draft"
		closed, prevented := 0, 0
		var trigger ui.Handle
		add("contract", "Inset and live placement", "Change the side and inset without losing the editor selection. Escape respects dismissal policy; the custom backdrop follows the transition.", func(c *ui.Context, t nui.Theme) {
			nui.Switch(c.Key("inset"), t, &inset, nui.SwitchOptions{Label: "Inset review panel"})
			nui.Switch(c.Key("motion"), t, &noMotion, nui.SwitchOptions{Label: "Disable panel motion"})
			nui.Switch(c.Key("blocked"), t, &blocked, nui.SwitchOptions{Label: "Block panel dismissal"})
			nui.Switch(c.Key("backdrop"), t, &customBackdrop, nui.SwitchOptions{Label: "Custom panel backdrop"})
			nui.Button(c.Key("open"), t, "Open reviewed panel", nui.ButtonOptions{Variant: nui.Outline}).Bind(&trigger).OnClick(func() { open = true })
			styles := nui.Styles{}
			if customBackdrop {
				styles[nui.Backdrop] = nui.PartStyle{Background: nui.Set(t.Info.Alpha(.35))}
			}
			nui.Slideover(c.Key("review"), t, &open, nui.SlideoverOptions{Side: side, Inset: inset, Extent: nui.Set(float32(340)), Title: "Reviewed panel", Description: "Edit the draft, then change the placement.", Transition: nui.Set(!noMotion), Dismissible: nui.Set(!blocked), ReturnFocus: &trigger, UI: styles, OnClose: func() { closed++ }, OnClosePrevent: func() { prevented++ }, Footer: func(c *ui.Context, inner nui.Theme) {
				nui.Button(c.Key("side"), inner, "Move panel side", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() { side = nui.Side((int(side) + 1) % 4) })
				nui.Button(c.Key("inset"), inner, "Toggle panel inset", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() { inset = !inset })
				nui.Button(c.Key("done"), inner, "Finish panel review", nui.ButtonOptions{}).OnClick(func() { open = false })
			}}, func(inner nui.Theme) {
				nui.Input(c.Key("draft"), inner, &draft, nui.InputOptions{Label: "Reviewed panel draft"})
				for i := range 12 {
					ui.Text(c.Key(i), fmt.Sprintf("Panel detail %02d", i+1)).Height(30)
				}
			})
			ui.Text(c, fmt.Sprintf("Panel notifications: closed=%d prevented=%d", closed, prevented))
			ui.Text(c, "Panel draft: "+draft)
		})
		p.Deactivate = append(p.Deactivate, func() { open = false })
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
		avatarOn, customClose, horizontal, reverse, persistent := false, false, false, false, false
		policyActions, policyClosed := 0, 0
		limit, width := 3, float64(384)
		publish := func() {
			var avatar *nui.AvatarOptions
			if avatarOn {
				avatar = &nui.AvatarOptions{Alt: "Ada Lovelace"}
			}
			var closeOptions *nui.ButtonOptions
			if customClose {
				closeOptions = &nui.ButtonOptions{AccessibleLabel: "Close policy toast", Size: nui.XL, Color: nui.Info, Variant: nui.Soft}
			}
			actions := []nui.ToastActionOptions{{ID: "keep", Label: "Keep", AccessibleLabel: "Keep policy toast", Close: nui.Set(false), OnAction: func() { policyActions++ }}, {ID: "done", Label: "Done", AccessibleLabel: "Done policy toast", OnAction: func() { policyActions++ }}}
			if reverse {
				actions[0], actions[1] = actions[1], actions[0]
			}
			orientation := nui.Vertical
			if horizontal {
				orientation = nui.Horizontal
			}
			timeout := 10 * time.Second
			if persistent {
				timeout = -1
			}
			a.pendingToasts = append(a.pendingToasts, nui.ToastOptions{ID: "gallery-policy", Title: "Policy notification", Description: "A notification with stable actions.", Icon: lucide.Info(), Avatar: avatar, CloseOptions: closeOptions, CloseIcon: lucide.Check(), Color: nui.Info, ProgressOptions: &nui.ToastProgressOptions{UI: nui.Styles{nui.Indicator: {Background: nui.Set(ui.RGB(255, 0, 255))}}}, Actions: actions, Orientation: nui.Set(orientation), Timeout: timeout, OnClose: func() { policyClosed++ }})
		}
		add("contract", "Appearance, viewport and stable actions", "Publish a notice, then update its action order and orientation while keeping focus. Viewport controls apply to existing notices.", func(c *ui.Context, t nui.Theme) {
			nui.Select(c.Key("position"), t, &a.toastViewport.Position, []nui.Option[nui.ToastPosition]{{Value: nui.ToastBottomRight, Label: "Bottom right"}, {Value: nui.ToastBottomLeft, Label: "Bottom left"}, {Value: nui.ToastBottomCenter, Label: "Bottom center"}, {Value: nui.ToastTopRight, Label: "Top right"}, {Value: nui.ToastTopLeft, Label: "Top left"}, {Value: nui.ToastTopCenter, Label: "Top center"}}, nui.SelectOptions{Label: "Toast position"})
			nui.Select(c.Key("limit"), t, &limit, []nui.Option[int]{{Value: 1, Label: "One toast"}, {Value: 2, Label: "Two toasts"}, {Value: 3, Label: "Three toasts"}}, nui.SelectOptions{Label: "Visible toast limit", OnChange: func() { a.toastViewport.Limit = limit }})
			nui.Slider(c.Key("width"), t, &width, nui.SliderOptions{Label: "Toast viewport width", Min: nui.Set(float64(160)), Max: nui.Set(float64(600)), Step: 8, OnChange: func(v float64) { a.toastViewport.Width = nui.Set(float32(v)) }})
			nui.Switch(c.Key("avatar"), t, &avatarOn, nui.SwitchOptions{Label: "Toast avatar"})
			nui.Switch(c.Key("close"), t, &customClose, nui.SwitchOptions{Label: "Custom toast close"})
			nui.Switch(c.Key("orientation"), t, &horizontal, nui.SwitchOptions{Label: "Horizontal policy actions"})
			nui.Switch(c.Key("persistent"), t, &persistent, nui.SwitchOptions{Label: "Persistent policy toast"})
			nui.Button(c.Key("publish"), t, "Show policy toast", nui.ButtonOptions{}).Width(220).AlignSelf(ui.Start).OnClick(publish)
			nui.Button(c.Key("update"), t, "Update focused toast", nui.ButtonOptions{Variant: nui.Outline}).Width(220).AlignSelf(ui.Start).KeepFocus().OnClick(func() { horizontal, reverse = !horizontal, !reverse; publish() })
			ui.Text(c, fmt.Sprintf("Policy toast actions: %d · Closed: %d", policyActions, policyClosed))
		})
		p.Deactivate = append(p.Deactivate, func() { a.toastViewport = nui.ToastViewportOptions{}; limit, width = 3, 384 })
	case "dropdown-menu", "context-menu":
		contract := &nui.MenuState{}
		size, filter, loop, loading := nui.MD, false, false, false
		pinned, reverse := false, false
		checkedChanges, selected, opened, reordered := 0, 0, 0, 0
		result := "No contract action"
		add("contract", "Size, filtering and item identity", "Compare sizes and descriptions, search root and child menus, and keep checked actions open. Stable IDs preserve focus when actions move.", func(c *ui.Context, t nui.Theme) {
			nui.Select(c.Key("size"), t, &size, []nui.Option[nui.Size]{{Value: nui.XS, Label: "XS"}, {Value: nui.SM, Label: "SM"}, {Value: nui.MD, Label: "MD"}, {Value: nui.LG, Label: "LG"}, {Value: nui.XL, Label: "XL"}}, nui.SelectOptions{Label: "Menu size"})
			nui.Switch(c.Key("filter"), t, &filter, nui.SwitchOptions{Label: "Filter menu actions"})
			nui.Switch(c.Key("loop"), t, &loop, nui.SwitchOptions{Label: "Loop menu navigation"})
			nui.Switch(c.Key("loading"), t, &loading, nui.SwitchOptions{Label: "Loading menu item"})
			items := []nui.MenuItem{
				{ID: "team", Kind: nui.MenuLabel, Label: "Project team", Description: "Current workspace", Avatar: &nui.AvatarOptions{Alt: "Team avatar", Text: "AL"}},
				{ID: "edit", Label: "Edit contract project", Description: "Change the name", Avatar: &nui.AvatarOptions{Alt: "Editor avatar", Text: "ED"}, Loading: loading, Shortcut: "Ctrl+E", OnSelect: func() { result = "Editing project"; selected++ }},
				{ID: "copy", Label: "Copy contract project", Icon: lucide.Copy(), OnSelect: func() { result = "Copied project"; selected++ }},
				{ID: "disabled", Label: "Unavailable contract action", Disabled: true},
				{ID: "pin", Label: "Pin contract project", Checked: nui.Set(pinned), Close: nui.Set(false), OnCheckedChange: func(v bool) { pinned = v; checkedChanges++ }, OnSelect: func() { selected++ }},
				{ID: "export", Label: "Contract export", Children: []nui.MenuItem{{ID: "csv", Label: "Contract CSV", OnSelect: func() { result = "Contract CSV selected"; selected++ }}, {ID: "json", Label: "Contract JSON", OnSelect: func() { result = "Contract JSON selected"; selected++ }}}, Content: &nui.MenuContentOptions{Filter: true, Width: nui.Set(float32(240)), FilterInput: &nui.InputOptions{Label: "Format search"}}},
				{ID: "reorder", Label: "Reverse menu items", Close: nui.Set(false), OnSelect: func() { reverse = !reverse; reordered++ }},
			}
			if reverse {
				slices.Reverse(items)
			}
			anchor := nui.Button(c.Key("contract-trigger"), t, "Open contract menu", nui.ButtonOptions{Variant: nui.Outline, Size: size}).AlignSelf(ui.Start)
			opts := nui.MenuOptions{Label: "Contract menu", Size: size, Filter: filter, FilterInput: &nui.InputOptions{Label: "Menu action search", Clearable: true}, FilterFields: []string{"label", "description"}, Loop: loop, CheckedIcon: lucide.Heart(), LoadingIcon: lucide.RefreshCw(), Width: nui.Set(float32(320)), OnOpenChange: func(bool) { opened++ }}
			if id == "dropdown-menu" {
				nui.DropdownMenu(c.Key("contract"), t, anchor, contract, items, opts)
			} else {
				nui.ContextMenu(c.Key("contract"), t, anchor, contract, items, opts)
			}
			ui.Text(c, result)
			ui.Textf(c, "Menu actions: %d · Checked: %d · Open changes: %d · Reorders: %d", selected, checkedChanges, opened, reordered)
		})
		p.Deactivate = append(p.Deactivate, func() { *contract = nui.MenuState{} })
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
		} else {
			pointMenu := &nui.MenuState{}
			blocked, moved := false, false
			clicks, changes := 0, 0
			add("position", "Pointer position and target actions", "Right-click twice at different points. Moving the target keeps the open menu at its original window position; disabling the menu leaves the target's primary action available.", func(c *ui.Context, t nui.Theme) {
				nui.Switch(c.Key("disabled"), t, &blocked, nui.SwitchOptions{Label: "Disable point menu"})
				left := float32(0)
				if moved {
					left = 40
				}
				target := nui.Button(c.Key("target"), t, "Point menu target", nui.ButtonOptions{Variant: nui.Outline}).Width(300).Height(80).AlignSelf(ui.Start).Margin(0, 0, 0, left).OnClick(func() { clicks++ })
				nui.ContextMenu(c.Key("point"), t, target, pointMenu, []nui.MenuItem{{ID: "move", Label: "Move context target", Close: nui.Set(false), OnSelect: func() { moved = !moved }}, {ID: "done", Label: "Finish context action"}}, nui.MenuOptions{Label: "Point context panel", Disabled: blocked, Width: nui.Set(float32(200)), OnOpenChange: func(bool) { changes++ }})
				nui.Button(c.Key("open"), t, "Open context at target", nui.ButtonOptions{Variant: nui.Soft}).AlignSelf(ui.Start).OnClick(func() { pointMenu.Open = true })
				ui.Textf(c, "Target clicks: %d · Menu changes: %d · Moved: %v", clicks, changes, moved)
			})
			p.Deactivate = append(p.Deactivate, func() { *pointMenu = nui.MenuState{} })
		}
	case "tabs":
		value, draft := "account", "Persistent tab draft"
		size, manual, vertical, underline := nui.MD, false, false, false
		content, keep, reverse, removed, customIndicator := true, true, false, false, false
		changes := 0
		add("contract", "Accessories, content and stable state", "Use manual activation, edit the draft, then reorder or change direction without losing the selection. Removing the active tab repairs focus without reporting a user selection.", func(c *ui.Context, t nui.Theme) {
			nui.Select(c.Key("size"), t, &size, []nui.Option[nui.Size]{{Value: nui.XS, Label: "XS"}, {Value: nui.SM, Label: "SM"}, {Value: nui.MD, Label: "MD"}, {Value: nui.LG, Label: "LG"}, {Value: nui.XL, Label: "XL"}}, nui.SelectOptions{Label: "Contract tab size"})
			nui.Switch(c.Key("manual"), t, &manual, nui.SwitchOptions{Label: "Manual tab activation"})
			nui.Switch(c.Key("content"), t, &content, nui.SwitchOptions{Label: "Show tab content"})
			nui.Switch(c.Key("keep"), t, &keep, nui.SwitchOptions{Label: "Keep tab content"})
			nui.Switch(c.Key("underline"), t, &underline, nui.SwitchOptions{Label: "Contract tab underline"})
			nui.Switch(c.Key("indicator"), t, &customIndicator, nui.SwitchOptions{Label: "Custom tab indicator"})
			items := []nui.TabItem{{ID: "account", Label: "Account tab", Avatar: &nui.AvatarOptions{Alt: "Tabs Ada", Text: "AL"}, Badge: "0", Content: func(c *ui.Context, th nui.Theme) {
				nui.Input(c.Key("draft"), th, &draft, nui.InputOptions{Label: "Contract tab draft"})
			}}, {ID: "activity", Label: "Activity tab", Icon: lucide.Clock(), Badge: "3", BadgeOptions: &nui.BadgeOptions{Color: nui.Info, Variant: nui.Soft}, Content: func(c *ui.Context, th nui.Theme) { ui.Text(c, "Activity details") }}, {ID: "blocked", Label: "Blocked tab", Disabled: true}}
			if removed {
				items = items[1:]
			}
			if reverse {
				slices.Reverse(items)
			}
			mode := nui.ActivationAutomatic
			if manual {
				mode = nui.ActivationManual
			}
			var parts nui.Styles
			if customIndicator {
				parts = nui.Styles{nui.Indicator: {Background: nui.Set(ui.RGB(255, 0, 255))}}
			}
			nui.Tabs(c.Key("tabs"), t, &value, items, nui.TabsOptions{Label: "Contract tabs", Size: size, Vertical: vertical, Underline: underline, Content: nui.Set(content), KeepMounted: keep, ActivationMode: mode, OnChange: func(string) { changes++ }, UI: parts})
			nui.Button(c.Key("reverse"), t, "Reverse contract tabs", nui.ButtonOptions{Variant: nui.Outline}).AlignSelf(ui.Start).KeepFocus().OnClick(func() { reverse = !reverse })
			nui.Button(c.Key("direction"), t, "Toggle contract direction", nui.ButtonOptions{Variant: nui.Outline}).AlignSelf(ui.Start).KeepFocus().OnClick(func() { vertical = !vertical })
			label := "Remove account tab"
			if removed {
				label = "Restore account tab"
			}
			nui.Button(c.Key("remove"), t, label, nui.ButtonOptions{Variant: nui.Soft}).AlignSelf(ui.Start).KeepFocus().OnClick(func() { removed = !removed })
			ui.Textf(c, "Selected tab: %s · Changes: %d", value, changes)
		})
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
			description := "Toggle with Enter or Space. KeepMounted retains widget content while closed; the application draft survives either setting."
			if sample.disabled {
				description = "This disclosure is intentionally closed and disabled. Pointer and keyboard input leave it closed and do not invoke change callbacks."
			}
			add(sample.id, sample.title, description, func(c *ui.Context, t nui.Theme) {
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
		truncated, customSeparator, disabled, inactive, icon := true, false, false, false, false
		clicks := 0
		selected := "none"
		add("contract", "Path parameters and actions", "The final item can navigate too. Truncation keeps separators around the ellipsis; an icon takes priority over the avatar.", func(c *ui.Context, t nui.Theme) {
			nui.Switch(c.Key("truncate"), t, &truncated, nui.SwitchOptions{Label: "Truncate contract path"})
			nui.Switch(c.Key("separator"), t, &customSeparator, nui.SwitchOptions{Label: "Custom path separator"})
			nui.Switch(c.Key("disabled"), t, &disabled, nui.SwitchOptions{Label: "Disable final path item"})
			nui.Switch(c.Key("active"), t, &inactive, nui.SwitchOptions{Label: "Remove final path highlight"})
			nui.Switch(c.Key("icon"), t, &icon, nui.SwitchOptions{Label: "Path icon priority"})
			var leading, separator *ui.SVG
			if icon {
				leading = lucide.House()
			}
			if customSeparator {
				separator = lucide.Slash()
			}
			maximum := 0
			if truncated {
				maximum = 2
			}
			nui.Breadcrumb(c, t, []nui.BreadcrumbItem{{ID: "home", Label: "Contract home", Avatar: &nui.AvatarOptions{Alt: "Ada Lovelace"}, Icon: leading, OnSelect: func() { selected = "home"; clicks++ }}, {ID: "project", Label: "Contract project", OnSelect: func() { selected = "project"; clicks++ }}, {ID: "draft", Label: "Contract draft"}, {ID: "current", Label: "Contract current", Active: nui.Set(!inactive), Disabled: disabled, UI: nui.Styles{nui.Label: {Weight: nui.Set(600)}}, OnSelect: func() { selected = "current"; clicks++ }}}, nui.BreadcrumbOptions{Label: "Contract location", MaxItems: maximum, Color: nui.Info, SeparatorIcon: separator})
			ui.Textf(c, "Path: %s · Clicks: %d", selected, clicks)
		})
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
		contractPage, total, changes := 5, 230, 0
		customIcons, edges := true, true
		add("contract", "Control icons and shrinking totals", "Try custom controls, then reduce the total. Page clamping does not emit a user change.", func(c *ui.Context, t nui.Theme) {
			nui.Switch(c.Key("icons"), t, &customIcons, nui.SwitchOptions{Label: "Custom pagination icons"})
			nui.Switch(c.Key("edges"), t, &edges, nui.SwitchOptions{Label: "Contract pagination edges"})
			var first, previous, next, last, ellipsis *ui.SVG
			if customIcons {
				first = lucide.ChevronsUp()
				previous = lucide.ArrowLeft()
				next = lucide.ArrowRight()
				last = lucide.ChevronsDown()
				ellipsis = lucide.EllipsisVertical()
			}
			nui.Pagination(c, t, &contractPage, nui.PaginationOptions{Label: "Contract pages", Total: total, PageSize: 10, SiblingCount: nui.Set(0), ShowEdges: nui.Set(edges), ActiveColor: nui.Info, ActiveVariant: nui.Outline, FirstIcon: first, PrevIcon: previous, NextIcon: next, LastIcon: last, EllipsisIcon: ellipsis, OnChange: func(int) { changes++ }, UI: nui.Styles{nui.Ellipsis: {Foreground: nui.Set(t.Info)}}, Style: func(slot nui.Slot, e ui.Element) {
				switch slot {
				case nui.First:
					e.Label("Contract first page")
				case nui.Previous:
					e.Label("Contract previous page")
				case nui.Next:
					e.Label("Contract next page")
				case nui.Last:
					e.Label("Contract last page")
				}
			}})
			nui.Button(c, t, "Shrink contract total", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() { total = 10 })
			ui.Textf(c, "Contract page: %d · Changes: %d", contractPage, changes)
		})
		for _, sample := range []struct {
			id, title                 string
			siblings                  int
			controls, edges, disabled bool
			size                      nui.Size
		}{{"basic", "Page navigation", 1, true, true, false, nui.MD}, {"zero", "Explicit zero neighbors", 0, true, false, false, nui.SM}, {"states", "Disabled soft pagination", 2, false, true, true, nui.LG}} {
			page := 5
			changes := 0
			description := map[string]string{
				"basic":  "Use the previous/next controls or numbered buttons. One neighboring page is shown on either side of the current page.",
				"zero":   "No neighboring page numbers are shown. Use the first, previous, next and last controls to move between pages.",
				"states": "All page buttons are disabled. Page 5 stays selected, and pointer or keyboard input does not invoke change callbacks.",
			}[sample.id]
			add(sample.id, sample.title, description, func(c *ui.Context, t nui.Theme) {
				nui.Pagination(c, t, &page, nui.PaginationOptions{Label: sample.title, Total: 100, PageSize: 10, SiblingCount: nui.Set(sample.siblings), ShowControls: nui.Set(sample.controls), ShowEdges: nui.Set(sample.edges), Disabled: sample.disabled, Size: sample.size, Variant: nui.Soft, OnChange: func(int) { changes++ }})
				ui.Textf(c, "Page %d · Changes: %d", page, changes)
			})
		}
	case "stepper":
		step, draft := "details", "Retained step draft"
		vertical, reverse, removed, disabled := false, false, false, false
		changes := 0
		add("contract", "Step content and stable identity", "Edit the active body, reorder or change orientation, then remove the current step. Disabled applies to step headings, while the body keeps its own state.", func(c *ui.Context, t nui.Theme) {
			nui.Switch(c.Key("disable"), t, &disabled, nui.SwitchOptions{Label: "Disable contract steps"})
			items := []nui.StepItem{{ID: "details", Title: "Contract details", Description: "Edit project", Icon: lucide.Folder(), UI: nui.Styles{nui.Label: {Foreground: nui.Set(t.Info)}}, Content: func(c *ui.Context, inner nui.Theme) {
				nui.Input(c, inner, &draft, nui.InputOptions{Label: "Contract step draft"}).FillWidth()
			}}, {ID: "review", Title: "Contract review", Description: "Confirm changes", Content: func(c *ui.Context, inner nui.Theme) { ui.Text(c, "Contract review body") }}, {ID: "blocked", Title: "Contract blocked", Disabled: true}}
			if removed {
				items = items[1:]
			}
			if reverse {
				slices.Reverse(items)
			}
			nui.Stepper(c, t, &step, items, nui.StepperOptions{Label: "Contract setup", Vertical: vertical, Linear: nui.Set(false), Disabled: disabled, Color: nui.Neutral, OnChange: func(string) { changes++ }})
			nui.Button(c, t, "Reverse contract steps", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() { reverse = !reverse })
			nui.Button(c, t, "Toggle contract step direction", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() { vertical = !vertical })
			label := "Remove contract details"
			if removed {
				label = "Restore contract details"
			}
			nui.Button(c, t, label, nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() { removed = !removed })
			ui.Textf(c, "Contract step: %s · Changes: %d", step, changes)
		})
		for _, sample := range []struct {
			id, title        string
			linear, vertical bool
		}{{"linear", "Linear workflow", true, false}, {"nonlinear", "Nonlinear workflow", false, false}, {"vertical", "Vertical workflow", false, true}} {
			step := "details"
			changes := 0
			description := "Choose any available step directly. The unavailable Publish step cannot activate."
			if sample.linear {
				description = "Steps unlock in order: choose Review before Approve. The unavailable Publish step cannot activate."
			}
			add(sample.id, sample.title, description, func(c *ui.Context, t nui.Theme) {
				nui.Stepper(c, t, &step, []nui.StepItem{{ID: "details", Title: "Details", Description: "Describe the workspace"}, {ID: "review", Title: "Review", Description: "Check access"}, {ID: "approve", Title: "Approve", Description: "Confirm changes"}, {ID: "locked", Title: "Publish", Disabled: true}}, nui.StepperOptions{Label: sample.title, Linear: nui.Set(sample.linear), Vertical: sample.vertical, OnChange: func(string) { changes++ }})
				ui.Textf(c, "Current step: %s · Changes: %d", step, changes)
			})
		}
	case "timeline":
		currentID := "build"
		reverse, horizontal := false, false
		selections := 0
		add("contract", "Dates, progress and keyboard selection", "CurrentID takes priority over the numeric index. Reverse changes progress without changing the item order. Tab and Enter or Space can select an item.", func(c *ui.Context, t nui.Theme) {
			nui.Switch(c.Key("reverse"), t, &reverse, nui.SwitchOptions{Label: "Reverse contract timeline"})
			nui.Switch(c.Key("direction"), t, &horizontal, nui.SwitchOptions{Label: "Horizontal contract timeline"})
			nui.Timeline(c, t, []nui.TimelineItem{{ID: "create", Title: "Contract created", Time: "09:00", Description: "Project initialized", Avatar: &nui.AvatarOptions{Alt: "Ada Lovelace"}, UI: nui.Styles{nui.DateSlot: {Foreground: nui.Set(t.Info)}}}, {ID: "build", Title: "Contract built", Time: "09:10", Description: "Checks passed", Icon: lucide.Check()}, {ID: "ship", Title: "Contract published", Time: "09:20", Description: "Ready to share", Icon: lucide.Send()}}, nui.TimelineOptions{Label: "Contract activity", Current: nui.Set(0), CurrentID: currentID, Color: nui.Neutral, Horizontal: horizontal, Reverse: reverse, OnSelect: func(id string) { currentID = id; selections++ }})
			ui.Textf(c, "Timeline: %s · Selections: %d", currentID, selections)
		})
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
		localClosed := 0
		add("local", "Local dismissal state", "An explicit close button also works without an application open binding. Unmounting the component resets its local visibility.", func(c *ui.Context, t nui.Theme) {
			nui.Banner(c, t, nil, nui.BannerOptions{Title: "Local announcement", CloseButton: nui.Set(true), CloseIcon: lucide.Check(), CloseLabel: "Close local banner", OnClose: func() { localClosed++ }})
			ui.Textf(c, "Local banner closes: %d", localClosed)
		})
		contractOpen, showIcon, customClose := true, true, false
		contractActions, selected, contractClosed := 0, 0, 0
		add("contract", "Banner activation and close configuration", "Activate the banner, its action and its close button independently. Changing the icon keeps the action identity stable; alignment follows the window breakpoint.", func(c *ui.Context, t nui.Theme) {
			nui.Switch(c.Key("close"), t, &customClose, nui.SwitchOptions{Label: "Custom contract banner close"})
			var icon *ui.SVG
			if showIcon {
				icon = lucide.Info()
			}
			var close *nui.ButtonOptions
			if customClose {
				close = &nui.ButtonOptions{Variant: nui.Outline, Size: nui.SM, Leading: lucide.Check()}
			}
			nui.Banner(c.Key("contract-banner"), t, &contractOpen, nui.BannerOptions{Title: "Contract announcement", Icon: icon, Color: nui.Neutral, CloseLabel: "Close contract banner", Close: close, OnSelect: func() { selected++ }, OnClose: func() { contractClosed++ }, Actions: func(c *ui.Context, inner nui.Theme) {
				nui.Button(c, inner, "Contract banner action", nui.ButtonOptions{}).OnClick(func() { contractActions++ })
			}, Style: func(slot nui.Slot, e ui.Element) {
				if slot == nui.Root {
					e.Label("Contract banner")
				}
			}})
			nui.Button(c, t, "Toggle contract banner icon", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() { showIcon = !showIcon })
			nui.Button(c, t, "Show contract banner", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { contractOpen = true })
			ui.Textf(c, "Banner selections: %d · Actions: %d · Closed: %d", selected, contractActions, contractClosed)
		})
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
		clear, customRetry, customActions := true, false, false
		contractRetries, recoveries := 0, 0
		add("contract", "Error fields and replacement actions", "The default retry can be hidden or customized. A supplied action area takes priority and remains available when the default action is hidden.", func(c *ui.Context, t nui.Theme) {
			nui.Switch(c.Key("clear"), t, &clear, nui.SwitchOptions{Label: "Show contract retry"})
			nui.Switch(c.Key("retry"), t, &customRetry, nui.SwitchOptions{Label: "Custom contract retry"})
			nui.Switch(c.Key("actions"), t, &customActions, nui.SwitchOptions{Label: "Replace contract error actions"})
			var retry *nui.ButtonOptions
			if customRetry {
				retry = &nui.ButtonOptions{Color: nui.Info, Variant: nui.Outline, Size: nui.MD, Leading: lucide.RefreshCw()}
			}
			var actions func(*ui.Context, nui.Theme)
			if customActions {
				actions = func(c *ui.Context, inner nui.Theme) {
					nui.Button(c, inner, "Contract recover", nui.ButtonOptions{}).OnClick(func() { recoveries++ })
				}
			}
			nui.ErrorView(c, t, nui.ErrorOptions{Status: "503", Title: "Missing project", Description: "Retry uses local state.", Icon: lucide.CircleAlert(), RetryLabel: "Contract retry", Clear: nui.Set(clear), Retry: retry, Actions: actions, OnRetry: func() { contractRetries++ }, UI: nui.Styles{nui.StatusSlot: {Foreground: nui.Set(t.Info)}}})
			ui.Textf(c, "Error retries: %d · Recoveries: %d", contractRetries, recoveries)
		})
		retries := 0
		recovered := false
		add("status", "Status code", "An embedded error keeps its message within the preview.", func(c *ui.Context, t nui.Theme) {
			nui.ErrorView(c, t, nui.ErrorOptions{Status: "404", Title: "Project not found", Description: "The local fixture has no record with this ID.", UI: nui.Styles{nui.Label: {FontSize: nui.Set(float32(24))}}})
		})
		add("retry", "Retry counter", "Retry is simulated and performs no network request.", func(c *ui.Context, t nui.Theme) {
			nui.ErrorView(c, t, nui.ErrorOptions{Status: "503", Title: "Activity unavailable", OnRetry: func() { retries++ }, UI: nui.Styles{nui.Label: {FontSize: nui.Set(float32(24))}}})
			ui.Textf(c, "Retried %d times", retries)
		})
		add("recovery", "Recover to content", "The retry action replaces the error with a success alert.", func(c *ui.Context, t nui.Theme) {
			if !recovered {
				nui.ErrorView(c, t, nui.ErrorOptions{Title: "Activity details could not load", Description: "Try loading the local preview again.", RetryLabel: "Restore activity", OnRetry: func() { recovered = true }, UI: nui.Styles{nui.Label: {FontSize: nui.Set(float32(24))}}})
			} else {
				nui.Alert(c, t, nui.AlertOptions{Title: "Activity restored", Color: nui.Success})
				nui.Button(c, t, "Reset recovery", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { recovered = false })
			}
		})
	}
	return p
}
