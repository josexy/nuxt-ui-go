package gallery

import (
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

func tableParityExample() func(*ui.Context, nui.Theme) {
	var state nui.TableState[string]
	rows := []string{"B2", "A2", "B1", "A1"}
	state.SetRows(rows, func(v string) string { return v })
	orders := []ui.SortOrder{}
	state.List.SortBy = &orders
	manual, loading, empty, rtl, disabled := false, false, false, false, false
	animation := nui.TableLoadingCarousel
	sorts, selections := 0, 0
	activated := "None"
	return func(c *ui.Context, t nui.Theme) {
		ui.Row(c).Wrap().Gap(8).Children(func() {
			nui.Checkbox(c, t, &manual, nui.CheckboxOptions{Label: "Manual table sorting"})
			nui.Checkbox(c, t, &loading, nui.CheckboxOptions{Label: "Loading contract table"})
			nui.Checkbox(c, t, &disabled, nui.CheckboxOptions{Label: "Disable contract table"})
			nui.Checkbox(c, t, &rtl, nui.CheckboxOptions{Label: "RTL contract table"})
			nui.Checkbox(c, t, &empty, nui.CheckboxOptions{Label: "Empty contract table", OnChange: func(v bool) {
				if v {
					state.SetRows(nil, func(v string) string { return v })
				} else {
					state.SetRows(rows, func(v string) string { return v })
				}
			}})
		})
		ui.Row(c).Wrap().Gap(8).Children(func() {
			nui.Button(c, t, "Sort group then ID", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { orders = []ui.SortOrder{{Column: "group"}, {Column: "id"}} })
			nui.Button(c, t, "Restore input order", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { orders = nil })
			nui.Button(c, t, "Cycle loading animation", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { animation = (animation + 1) % 4 })
		})
		direction := ui.LeftToRight
		if rtl {
			direction = ui.RightToLeft
		}
		ui.Column(c.Key("direction")).Direction(direction).Children(func() {
			nui.Table(c.Key("table"), t, &state, []nui.TableColumn[string]{
				{ID: "group", Title: "Group", Width: 160, MinWidth: 100, Value: func(v string) string { return v[:1] }, Compare: func(a, b string) int { return strings.Compare(a[:1], b[:1]) }},
				{ID: "id", Title: "Record ID", Width: 180, MinWidth: 100, Value: func(v string) string { return v }, Compare: strings.Compare, Decoration: nui.Decoration{UI: nui.Styles{nui.Label: {Weight: nui.Set(600)}}}},
			}, nui.TableOptions{Label: "Contract table", Caption: "Four records in input order B2, A2, B1, A1", Height: 220, ManualSorting: manual, Loading: loading, LoadingColor: nui.Info, LoadingAnimation: animation, Disabled: disabled,
				BuildCaption: func(c *ui.Context, t nui.Theme) { ui.Text(c, "Project records").TextColor(t.Highlighted) },
				BuildEmpty:   func(c *ui.Context, _ nui.Theme) { ui.Text(c, "No contract records") },
				BuildLoading: func(c *ui.Context, _ nui.Theme) { ui.Text(c, "Fetching contract records") },
				OnSortChange: func([]ui.SortOrder) { sorts++ }, OnSelectionChange: func([]string) { selections++ }, OnActivate: func(id string) { activated = id }})
		})
		first := "None"
		if state.Len() > 0 {
			first = state.RowID(0)
		}
		ui.Text(c, fmt.Sprintf("First: %s · Sort events: %d · Selection events: %d · Activated: %s", first, sorts, selections, activated))
	}
}
