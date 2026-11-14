package gallery

import (
	"fmt"
	"slices"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func addTreeContract(p *componentPage) {
	var state nui.TreeState
	var ticket nui.TreeTicket
	multiple, propagate, bubble := true, true, true
	custom, rtl, disabled, blockToggle := false, false, false, false
	changes, expansions, requests := 0, 0, 0
	result := "Ready"
	nodes := []nui.TreeNode{
		{ID: "project", Label: "Contract project", DefaultExpanded: true, Children: []nui.TreeNode{
			{ID: "alpha", Label: "Contract Alpha", Icon: lucide.FileText(), UI: nui.Styles{nui.Label: {Weight: nui.Set(600)}}},
			{ID: "beta", Label: "Contract Beta", TrailingIcon: lucide.Star()},
			{ID: "locked", Label: "Contract locked", Disabled: true},
		}},
		{ID: "remote", Label: "Remote records", Lazy: true},
		{ID: "protected", Label: "Protected tree item", OnSelect: func(e *nui.TreeEvent) { result = "Selection prevented"; e.PreventDefault() }},
	}
	state.SetNodes(nodes)
	p.Deactivate = append(p.Deactivate, state.Close)
	p.Scenes = append(p.Scenes, scene("contract", "Selection, content and lazy loading", "Partial selection follows the linkage options. Node styles override tree-wide styles; content builders replace non-interactive icons. Lazy completion uses tickets, and collapsed requests are rejected.", func(c *ui.Context, t nui.Theme) {
		ui.Row(c).Wrap().Gap(10).Children(func() {
			nui.Checkbox(c, t, &multiple, nui.CheckboxOptions{Label: "Multiple contract tree"})
			nui.Checkbox(c, t, &propagate, nui.CheckboxOptions{Label: "Select tree descendants"})
			nui.Checkbox(c, t, &bubble, nui.CheckboxOptions{Label: "Bubble tree selection"})
			nui.Checkbox(c, t, &blockToggle, nui.CheckboxOptions{Label: "Prevent tree toggle"})
			nui.Checkbox(c, t, &custom, nui.CheckboxOptions{Label: "Custom tree parts"})
			nui.Checkbox(c, t, &rtl, nui.CheckboxOptions{Label: "RTL contract tree"})
			nui.Checkbox(c, t, &disabled, nui.CheckboxOptions{Label: "Disable contract tree"})
		})
		ui.Row(c).Wrap().Gap(8).Children(func() {
			nui.Button(c, t, "Move Alpha last", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() {
				children := nodes[0].Children
				at := slices.IndexFunc(children, func(n nui.TreeNode) bool { return n.ID == "alpha" })
				alpha := children[at]
				nodes[0].Children = append(slices.Delete(slices.Clone(children), at, at+1), alpha)
				state.SetNodes(nodes)
				result = "Nodes reordered"
			})
			nui.Button(c, t, "Collapse remote branch", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { state.SetExpanded("remote", false); result = "Remote branch collapsed" })
			nui.Button(c, t, "Complete empty tree load", nui.ButtonOptions{Variant: nui.Outline, Disabled: requests == 0}).OnClick(func() {
				if state.CompleteLoad(ticket, nil, nil) {
					result = "Remote directory is empty"
				} else {
					result = "Stale tree completion rejected"
				}
			})
		})
		var leading, trailing func(*ui.Context, nui.Theme, nui.TreeItemState)
		if custom {
			leading = func(c *ui.Context, _ nui.Theme, _ nui.TreeItemState) { ui.Icon(c, lucide.Circle()).Size(16, 16) }
			trailing = func(c *ui.Context, _ nui.Theme, item nui.TreeItemState) {
				if item.Indeterminate {
					ui.Text(c, "Mixed")
				} else if item.Expanded {
					ui.Text(c, "Open")
				}
			}
		}
		direction := ui.LeftToRight
		if rtl {
			direction = ui.RightToLeft
		}
		ui.Column(c.Key("direction")).Direction(direction).Children(func() {
			nui.Tree(c.Key("tree"), t, &state, nui.TreeOptions{Label: "Contract tree", Height: 250, Color: nui.Neutral, Multiple: multiple, PropagateSelect: propagate, BubbleSelect: bubble, Disabled: disabled, Loop: true, ExpandOnClick: nui.Set(false), BuildLeading: leading, BuildTrailing: trailing,
				OnChange: func() { changes++ }, OnExpandedChange: func([]string) { expansions++ }, OnToggle: func(e *nui.TreeEvent) {
					if blockToggle {
						e.PreventDefault()
						result = "Toggle prevented"
					}
				}, OnLoad: func(_ string, next nui.TreeTicket) { ticket = next; requests++; result = "Loading remote records" }})
		})
		ui.Text(c, fmt.Sprintf("Tree selection: %v · Changes: %d · Expansions: %d", slices.Sorted(state.Selection.All()), changes, expansions))
		ui.Textf(c, "%s · Loads: %d", result, requests)
	}))
}
