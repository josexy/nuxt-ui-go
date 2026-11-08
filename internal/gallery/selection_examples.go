package gallery

import (
	"context"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

func addSelectionContract(p *componentPage, menu bool) {
	values := []string{"tea"}
	var list nui.ListboxState
	var popup nui.SelectMenuState
	filter, loading, protect, reversed, styled := true, false, true, false, false
	focusSearch, resetModel, trailing, focusAfter := true, true, true, false
	changes, queries, selected, attempts, highlights, clears, openings := 0, 0, 0, 0, 0, 0, 0
	var leave ui.Handle
	protected := &nui.ListboxItemOptions{OnSelect: func(e *nui.ListboxEvent) {
		attempts++
		if protect {
			e.PreventDefault()
		}
	}, UI: nui.Styles{nui.ItemLabelSlot: {Weight: nui.Set(600)}}}
	p.Deactivate = append(p.Deactivate, func() { popup.Open = false })
	p.Scenes = append(p.Scenes, scene("contract", "Selection, filtering and item parts", "Search for cafe: the exact accented match comes first. The protected option can cancel activation. Reordering keeps IDs, while labels, avatars, chips and separators have their own parts.", func(c *ui.Context, t nui.Theme) {
		ui.Row(c).Wrap().Gap(12).Children(func() {
			nui.Switch(c.Key("protect"), t, &protect, nui.SwitchOptions{Label: "Protect cafe selection"})
			nui.Switch(c.Key("loading"), t, &loading, nui.SwitchOptions{Label: "Contract choices loading"})
			nui.Switch(c.Key("styles"), t, &styled, nui.SwitchOptions{Label: "Custom choice parts"})
		})
		if menu {
			ui.Row(c).Wrap().Gap(12).Children(func() {
				nui.Switch(c.Key("focus"), t, &focusSearch, nui.SwitchOptions{Label: "Focus menu search"})
				nui.Switch(c.Key("reset"), t, &resetModel, nui.SwitchOptions{Label: "Clear resets selection"})
				nui.Switch(c.Key("trailing"), t, &trailing, nui.SwitchOptions{Label: "Show menu trailing control"})
				nui.Switch(c.Key("after"), t, &focusAfter, nui.SwitchOptions{Label: "Focus outside after selection"})
			})
		}
		items := []nui.ListboxItem{
			{ID: "heading", Label: "Drinks", Kind: nui.ListboxLabel},
			{ID: "prefix", Label: "Cafeteria", Description: "Prefix match", Icon: lucide.Heart()},
			{ID: "contains", Label: "A café", Description: "Contains the search term", Chip: &nui.ChipOptions{Color: nui.Success}},
			{ID: "exact", Label: "Café", Description: "Exact match; activation can be cancelled", Avatar: &nui.AvatarOptions{Alt: "Cafe avatar", Text: "CF"}, Options: protected},
			{ID: "separator", Kind: nui.ListboxSeparator},
			{ID: "tea", Label: "Tea", Description: "Initially selected", Chip: &nui.ChipOptions{Color: nui.Warning}},
		}
		if reversed {
			items[1], items[3] = items[3], items[1]
		}
		o := nui.ListboxOptions{Label: "Contract choices", Color: nui.Neutral, Size: nui.MD, Highlight: true, SelectedIcon: lucide.Star(),
			OnChange: func([]string) { changes++ }, OnSelect: func(string) {
				selected++
				if focusAfter {
					leave.Focus(c)
				}
			}, OnHighlight: func(string) { highlights++ }, OnSearchTermChange: func(string) { queries++ }}
		if styled {
			o.UI = nui.Styles{nui.ItemLabelSlot: {Foreground: nui.Set(t.Info)}, nui.Description: {Foreground: nui.Set(t.Warning)}}
		}
		o.FilterInput = &nui.InputOptions{Label: "Contract filter", Placeholder: "Search drinks"}
		if menu {
			nui.SelectMenu(c.Key("choices"), t, &values, &popup, items, nui.SelectMenuOptions{ListboxOptions: o, Color: nui.Neutral,
				Avatar: &nui.AvatarOptions{Alt: "Menu avatar", Text: "ME"}, Trailing: lucide.ChevronUp(), Loading: loading, LoadingIcon: lucide.Clock(),
				SearchInput: nui.Set(filter), FocusSearch: nui.Set(focusSearch), ResetSearchTermOnBlur: nui.Set(false),
				Clearable: true, ClearIcon: lucide.Minus(), ResetModelValueOnClear: nui.Set(resetModel), ShowTrailing: nui.Set(trailing),
				OnClear: func() { clears++ }, OnOpenChange: func(bool) { openings++ }})
		} else {
			o.Filter, o.Loading, o.Height = filter, loading, 280
			nui.Listbox(c.Key("choices"), t, &values, &list, items, o)
		}
		ui.Row(c).Wrap().Gap(8).Children(func() {
			nui.Button(c.Key("filter"), t, "Toggle contract filter", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() { filter = !filter })
			nui.Button(c.Key("reorder"), t, "Reorder contract choices", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() { reversed = !reversed })
			nui.Button(c.Key("leave"), t, "Leave contract selection", nui.ButtonOptions{Variant: nui.Outline}).Bind(&leave)
		})
		ui.Textf(c, "Selected: %v · Changes: %d · Activations: %d", values, changes, selected)
		ui.Textf(c, "Queries: %d · Highlights: %d · Attempts: %d", queries, highlights, attempts)
		if menu {
			ui.Textf(c, "Menu opens/closes: %d · Clears: %d", openings, clears)
		}
	}))
	if menu {
		addRemoteSelection(p)
	}
}

func addRemoteSelection(p *componentPage) {
	var menu nui.SelectMenuState
	var values []string
	var results nui.QueryState[nui.ListboxItem]
	var current, previous nui.QueryTicket
	status := "No request"
	p.Deactivate = append(p.Deactivate, func() { menu.Open = false; results.Cancel() })
	p.Scenes = append(p.Scenes, scene("remote-contract", "Simulated remote search", "Type two different queries, then complete the previous request. Only the latest ticket can replace the options. Completion is an explicit action; this example makes no network request.", func(c *ui.Context, t nui.Theme) {
		nui.SelectMenu(c.Key("remote"), t, &values, &menu, results.Items, nui.SelectMenuOptions{Label: "Remote choices", ExternalFilter: true, Loading: results.Loading,
			ResetSearchTermOnBlur: nui.Set(false), ResetSearchTermOnSelect: nui.Set(false), OnQuery: func(query string) {
				previous = current
				_, current = results.Start(context.Background(), query)
				status = "Request pending: " + query
			}})
		ui.Row(c).Wrap().Gap(8).Children(func() {
			nui.Button(c.Key("previous"), t, "Complete stale options", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() {
				if results.Complete(previous, []nui.ListboxItem{{ID: "stale", Label: "Stale result"}}, nil) {
					status = "Previous result accepted"
				} else {
					status = "Stale options rejected"
				}
			})
			nui.Button(c.Key("current"), t, "Complete current options", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() {
				if results.Complete(current, []nui.ListboxItem{{ID: "current", Label: "Result: " + current.Query}}, nil) {
					status = "Current options accepted"
				} else {
					status = "Current options rejected"
				}
			})
			nui.Button(c.Key("cancel"), t, "Cancel option request", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { results.Cancel(); status = "Option request cancelled" })
		})
		ui.Text(c, status)
		ui.Textf(c, "Remote selected: %v", values)
	}))
}
