package gallery

import (
	"context"
	"fmt"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

func addCommandContract(p *componentPage) {
	var state nui.CommandPaletteState
	var opener ui.Handle
	protect, keepOpen, external, rtl := true, false, false, false
	showInput, showIcon := true, true
	preserve := false
	delay := time.Duration(0)
	size := nui.MD
	selected := "None"
	calls, edits, requests := 0, 0, 0
	status := "Ready"
	var current, previous nui.QueryTicket
	var currentCtx context.Context
	query, parent := "", ""
	index := nui.NewCommandIndex([]nui.CommandAction{
		{ID: "create", Label: "Create document", Prefix: "File", Suffix: "Local", Description: "Start a blank invoice", Keywords: "new", Group: "Files", Icon: lucide.FilePlus(), Kbds: []string{"Ctrl", "N"}},
		{ID: "appearance", Label: "Appearance options", Placeholder: "Choose a theme…", Group: "Preferences", Children: []nui.CommandAction{{ID: "light", Label: "Light workspace", Icon: lucide.Sun()}, {ID: "dark", Label: "Dark workspace", Icon: lucide.Moon()}}},
		{ID: "profile", Label: "Personal workspace", Description: "Your account and documents", Group: "Files", Avatar: &nui.AvatarOptions{Text: "LC", Alt: "Lin Chen"}, Decoration: nui.Decoration{UI: nui.Styles{nui.ItemLabelSlot: {Weight: nui.Set(600)}}}},
		{ID: "status", Label: "Service status", Group: "Status", Chip: &nui.ChipOptions{Color: nui.Success}},
		{ID: "sync", Label: "Background sync", Group: "Status", Loading: true, Disabled: true},
		{ID: "protected", Label: "Protected command", Group: "Actions", OnSelectEvent: func(e *nui.CommandEvent) {
			if protect {
				e.PreventDefault()
				status = "Activation prevented"
			}
		}},
	})
	state.SetIndex(index)
	p.Deactivate = append(p.Deactivate, state.Close)
	p.Scenes = append(p.Scenes, scene("contract", "Search, nested actions and cancellation", "Search invoice or Local to match descriptions and suffixes. File actions stay in one group. Empty Backspace returns from Appearance. In manual search mode, edit twice and try the stale ticket before completing the current one.", func(c *ui.Context, t nui.Theme) {
		ui.Row(c).Wrap().Gap(10).Children(func() {
			nui.Checkbox(c, t, &protect, nui.CheckboxOptions{Label: "Protect command activation"})
			nui.Checkbox(c, t, &keepOpen, nui.CheckboxOptions{Label: "Keep commands open"})
			nui.Checkbox(c, t, &external, nui.CheckboxOptions{Label: "Manual command results"})
			nui.Checkbox(c, t, &preserve, nui.CheckboxOptions{Label: "Preserve command groups"})
			nui.Checkbox(c, t, &showInput, nui.CheckboxOptions{Label: "Show command input"})
			nui.Checkbox(c, t, &showIcon, nui.CheckboxOptions{Label: "Show command search icon"})
			nui.Checkbox(c, t, &rtl, nui.CheckboxOptions{Label: "RTL command palette"})
		})
		ui.Row(c).Wrap().Gap(8).Children(func() {
			nui.Button(c, t, "Toggle command delay", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() {
				if delay == 0 {
					delay = 150 * time.Millisecond
				} else {
					delay = 0
				}
			})
			nui.Button(c, t, "Cycle command size", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() {
				size++
				if size > nui.XL {
					size = nui.XS
				}
			})
			nui.Button(c, t, "Open contract commands", nui.ButtonOptions{Leading: lucide.Search()}).Bind(&opener).OnClick(func() { state.Open = true })
		})
		var search func(context.Context, nui.QueryTicket, string, string)
		if external {
			search = func(ctx context.Context, ticket nui.QueryTicket, text, level string) {
				previous, current = current, ticket
				currentCtx = ctx
				query, parent = text, level
				requests++
				status = "Waiting for result"
			}
		}
		direction := ui.LeftToRight
		if rtl {
			direction = ui.RightToLeft
		}
		ui.Column(c.Key("direction")).Direction(direction).Children(func() {
			nui.CommandPalette(c.Key("commands"), t, &state, nui.CommandPaletteOptions{Title: "Contract commands", Size: size, SearchDelay: delay, Search: search, SearchInput: nui.Set(showInput), ShowIcon: nui.Set(showIcon), PreserveGroupOrder: preserve, DisableShortcut: true, CloseOnSelect: nui.Set(!keepOpen), Loop: true,
				OnSelect: func(id string) { selected = id; calls++; status = "Command completed" }, OnSearchTermChange: func(string) { edits++ },
				Modal: nui.ModalOptions{ReturnFocus: &opener, Footer: func(c *ui.Context, t nui.Theme) {
					if external {
						ui.Row(c).Wrap().Gap(8).Children(func() {
							nui.Button(c, t, "Try stale command result", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() {
								if state.Results.Complete(previous, nil, nil) {
									status = "Unexpected stale completion"
								} else {
									status = "Stale command result rejected"
								}
							})
							nui.Button(c, t, "Complete command search", nui.ButtonOptions{Disabled: currentCtx == nil}).OnClick(func() {
								if state.Results.Complete(current, index.Search(currentCtx, query, parent, 200), nil) {
									status = "Result received"
								} else {
									status = "Obsolete command request"
								}
							})
						})
					}
					ui.Textf(c, "Command: %s · Calls: %d · Edits: %d", selected, calls, edits)
					ui.Textf(c, "%s · Requests: %d", status, requests)
				}}})
		})
		ui.Text(c, fmt.Sprintf("Command: %s · Calls: %d · Delay: %s", selected, calls, delay))
	}))
}
