package gallery

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

func addNavigationContract(p *componentPage) {
	var state nui.NavigationState
	active := "overview"
	vertical, compact, disabled, highlight, single, protect, rtl := true, false, false, true, false, true, false
	tooltip, popover, collapsible := true, true, true
	color, variant := nui.Primary, nui.Soft
	changes, expansions := 0, 0
	result := "Ready"
	items := []nui.NavigationItem{
		{ID: "overview", Label: "Overview page", Icon: lucide.House(), Badge: "4", Group: "Workspace"},
		{ID: "projects", Label: "Project library", Icon: lucide.Folder(), Chip: &nui.ChipOptions{Color: nui.Success, Inset: true}},
		{ID: "manage", Label: "Manage workspace", Icon: lucide.Settings(), DefaultOpen: true, Children: []nui.NavigationItem{
			{ID: "members", Label: "Workspace members", Description: "Invite people and assign permissions", Avatar: &nui.AvatarOptions{Text: "LC", Alt: "Lin Chen"}},
			{ID: "protected", Label: "Protected destination", Icon: lucide.Lock(), OnSelect: func(e *nui.NavigationEvent) {
				if protect {
					e.PreventDefault()
					result = "Navigation prevented"
				}
			}},
			{ID: "unavailable", Label: "Unavailable destination", Disabled: true},
		}},
		{ID: "help-label", Label: "Help section", LabelOnly: true},
		{ID: "help", Label: "Help center", Icon: lucide.Info(), Tooltip: nui.Set(true), Children: []nui.NavigationItem{{ID: "quickstart", Label: "Quickstart guide"}, {ID: "releases", Label: "Release notes"}}, UI: nui.Styles{nui.ItemLabelSlot: {Weight: nui.Set(600)}}},
	}
	p.Scenes = append(p.Scenes, scene("contract", "Navigation state and item content", "Compare horizontal, expanded and compact navigation. Protected destination cancels navigation. Single mode limits expanded siblings; collapsible controls whether the last open branch can close. Compact switching preserves expansion and returns focus from hidden children.", func(c *ui.Context, t nui.Theme) {
		ui.Row(c).Wrap().Gap(10).Children(func() {
			nui.Checkbox(c, t, &vertical, nui.CheckboxOptions{Label: "Vertical contract navigation"})
			nui.Checkbox(c, t, &compact, nui.CheckboxOptions{Label: "Compact contract navigation"})
			nui.Checkbox(c, t, &disabled, nui.CheckboxOptions{Label: "Disable contract navigation"})
			nui.Checkbox(c, t, &highlight, nui.CheckboxOptions{Label: "Highlight contract navigation"})
			nui.Checkbox(c, t, &single, nui.CheckboxOptions{Label: "Single navigation branch"})
			nui.Checkbox(c, t, &collapsible, nui.CheckboxOptions{Label: "Allow navigation collapse"})
			nui.Checkbox(c, t, &protect, nui.CheckboxOptions{Label: "Protect navigation action"})
			nui.Checkbox(c, t, &tooltip, nui.CheckboxOptions{Label: "Navigation tooltips"})
			nui.Checkbox(c, t, &popover, nui.CheckboxOptions{Label: "Navigation popovers"})
			nui.Checkbox(c, t, &rtl, nui.CheckboxOptions{Label: "RTL contract navigation"})
		})
		ui.Row(c).Wrap().Gap(8).Children(func() {
			nui.Button(c, t, "Toggle navigation color", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() {
				if color == nui.Primary {
					color = nui.Neutral
				} else {
					color = nui.Primary
				}
			})
			nui.Button(c, t, "Toggle navigation variant", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() {
				if variant == nui.Soft {
					variant = nui.LinkVariant
				} else {
					variant = nui.Soft
				}
			})
			nui.Button(c, t, "Collapse navigation externally", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() {
				if state.Expanded != nil {
					state.Expanded["manage"] = false
				}
			})
		})
		direction := ui.LeftToRight
		if rtl {
			direction = ui.RightToLeft
		}
		frame := ui.Column(c.Key("direction")).Direction(direction)
		if vertical && compact {
			frame.Width(64)
		}
		frame.Children(func() {
			nui.NavigationMenu(c.Key("navigation"), t, &active, &state, items, nui.NavigationOptions{Label: "Contract navigation", Vertical: vertical, Compact: compact, Disabled: disabled, Color: color, Variant: variant, Highlight: highlight, Single: single, Collapsible: nui.Set(collapsible), Tooltip: nui.Set(tooltip), Popover: nui.Set(popover), OnNavigate: func(id string) { changes++; result = "Opened " + id }, OnExpandedChange: func([]string) { expansions++ }})
		})
		ui.Text(c, fmt.Sprintf("Destination: %s · Actions: %d · Expansion changes: %d", active, changes, expansions))
		ui.Text(c, result)
	}))
}
