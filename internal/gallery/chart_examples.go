package gallery

import (
	"fmt"
	"slices"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

func addChartContract(p *componentPage) {
	var state nui.ChartState
	series := []nui.ChartSeries{
		{ID: "one", Label: "Contract series one", Values: []float64{3, 7, 4}},
		{ID: "two", Label: "Contract series two", Values: []float64{6, 2, 8}},
	}
	single, reverse, empty := false, false, false
	update := func() {
		current := slices.Clone(series)
		if single {
			current = current[:1]
		}
		if reverse {
			slices.Reverse(current)
		}
		if empty {
			current = nil
		}
		if err := state.SetCategories([]string{"January", "February", "March"}, current); err != nil {
			panic(err)
		}
		wedges := []nui.ChartSlice{{ID: "ready", Label: "Contract ready", Value: 6}, {ID: "pending", Label: "Contract pending", Value: 3}}
		if single {
			wedges = wedges[:1]
		}
		if reverse {
			slices.Reverse(wedges)
		}
		if empty {
			wedges = nil
		}
		if err := state.SetSlices(wedges); err != nil {
			panic(err)
		}
	}
	update()
	kind := nui.ChartLine
	xaxis, yaxis, points, tooltip, center := true, true, false, true, false
	rtl, palette, ratio, forceLegend := false, false, false, false
	focused, selected, legends := 0, 0, 0
	last := "No activation"
	p.Scenes = append(p.Scenes, scene("contract", "Chart axes and selection", "Automatic legends appear only for multiple series or slices. Inspect focused values with arrows, then activate with Enter or Space. Reordering retains visibility by ID; RTL mirrors X positions and physical arrow navigation.", func(c *ui.Context, t nui.Theme) {
		nui.Select(c.Key("kind"), t, &kind, []nui.Option[nui.ChartKind]{{Value: nui.ChartLine, Label: "Contract Line"}, {Value: nui.ChartBar, Label: "Contract Bar"}, {Value: nui.ChartArea, Label: "Contract Area"}, {Value: nui.ChartScatter, Label: "Contract Scatter"}, {Value: nui.ChartPie, Label: "Contract Pie"}, {Value: nui.ChartDonut, Label: "Contract Donut"}}, nui.SelectOptions{Label: "Contract chart kind"}).Width(160)
		ui.Row(c).Wrap().Gap(8).Children(func() {
			nui.Checkbox(c, t, &single, nui.CheckboxOptions{Label: "Single contract series", OnChange: func(bool) { update() }})
			nui.Checkbox(c, t, &reverse, nui.CheckboxOptions{Label: "Reverse contract series", OnChange: func(bool) { update() }})
			nui.Checkbox(c, t, &empty, nui.CheckboxOptions{Label: "Empty contract chart", OnChange: func(bool) { update() }})
			nui.Checkbox(c, t, &forceLegend, nui.CheckboxOptions{Label: "Always show chart legend"})
			nui.Checkbox(c, t, &xaxis, nui.CheckboxOptions{Label: "Show chart X axis"})
			nui.Checkbox(c, t, &yaxis, nui.CheckboxOptions{Label: "Show chart Y axis"})
			nui.Checkbox(c, t, &points, nui.CheckboxOptions{Label: "Show all line points"})
			nui.Checkbox(c, t, &tooltip, nui.CheckboxOptions{Label: "Show chart tooltip"})
			nui.Checkbox(c, t, &center, nui.CheckboxOptions{Label: "Show chart center"})
			nui.Checkbox(c, t, &rtl, nui.CheckboxOptions{Label: "RTL contract chart"})
			nui.Checkbox(c, t, &palette, nui.CheckboxOptions{Label: "Custom chart palette"})
			nui.Checkbox(c, t, &ratio, nui.CheckboxOptions{Label: "Chart aspect ratio"})
		})
		o := nui.ChartOptions{Label: "Contract chart", Description: "Monthly review totals", Kind: kind, Height: 240, XAxis: nui.Set(xaxis), YAxis: nui.Set(yaxis), XLabel: "Review month", YLabel: "Reviewed files", Points: points, Tooltip: nui.Set(tooltip), Animate: nui.Set(false), OnFocusChange: func(nui.ChartSelection) { focused++ }, OnSelect: func(point nui.ChartSelection) {
			selected++
			last = fmt.Sprintf("Activated index %d / slice %s", point.Index, point.SliceID)
		}, OnLegendChange: func(string, bool) { legends++ }}
		if forceLegend {
			o.Legend = nui.Set(true)
		}
		if ratio {
			o.Height = 0
			o.AspectRatio = 2
		}
		if palette {
			o.Colors = []ui.Color{t.Warning, t.Info}
			o.UI = nui.Styles{nui.XAxisSlot: {Foreground: nui.Set(t.Info)}, nui.Indicator: {Radius: nui.Set(float32(1))}}
		}
		if center {
			o.Center = func(c *ui.Context, t nui.Theme) { ui.Text(c, "Review total").TextColor(t.Text) }
		}
		direction := ui.LeftToRight
		if rtl {
			direction = ui.RightToLeft
		}
		ui.Column(c.Key("plot-direction")).Direction(direction).Children(func() { nui.Chart(c.Key("contract-chart"), t, &state, o) })
		ui.Textf(c, "Chart focus: %d · Selections: %d · Legend changes: %d", focused, selected, legends)
		ui.Text(c, last)
	}))
}
