package gallery

import (
	"fmt"
	"math"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

func (s *advancedState) updateChartData() {
	labels := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	xs, times := make([]float64, len(labels)), make([]time.Time, len(labels))
	series := []nui.ChartSeries{{ID: "build", Label: "Builds"}, {ID: "review", Label: "Reviews"}, {ID: "deploy", Label: "Deploys"}}
	for k := range series {
		series[k].Values = make([]float64, len(labels))
		for i := range labels {
			xs[i] = float64(i + 1)
			times[i] = time.Date(2026, time.Month(i+1), 1, 0, 0, 0, 0, time.UTC)
			v := 30 - float64(k)*7 + (16-float64(k)*3)*math.Sin(float64(i)*.65+float64(k)+float64(s.ChartRevision)*1.2)
			if s.ChartNegative {
				v -= 22
			}
			series[k].Values[i] = math.Round(v)
		}
	}
	if s.ChartGaps {
		series[0].Values[5] = math.NaN()
	}
	var err error
	switch s.ChartAxis {
	case "category":
		err = s.Chart.SetCategories(labels, series)
	case "time":
		err = s.Chart.SetTimeData(times, series)
	default:
		err = s.Chart.SetData(xs, series)
	}
	if err != nil {
		panic(err)
	}
	phase := float64(s.ChartRevision) * 1.2
	err = s.Chart.SetSlices([]nui.ChartSlice{
		{ID: "passed", Label: "Passed", Value: math.Round(65 + 18*math.Sin(phase)), Color: nui.Primary},
		{ID: "queued", Label: "Queued", Value: math.Round(18 + 8*math.Cos(phase)), Color: nui.Secondary},
		{ID: "failed", Label: "Failed", Value: math.Round(12 + 7*math.Sin(phase+1)), Color: nui.Error},
		{ID: "skipped", Label: "Skipped", Value: math.Round(8 + 5*math.Cos(phase+2)), Color: nui.Warning},
	})
	if err != nil {
		panic(err)
	}
}

func (a *State) chartDemo(c *ui.Context, t nui.Theme) {
	s := &a.Advanced
	radial := s.ChartKind == nui.ChartPie || s.ChartKind == nui.ChartDonut
	ui.Row(c.Key("chart-kinds")).Wrap().Gap(6).Children(func() {
		for _, it := range []struct {
			label string
			kind  nui.ChartKind
		}{
			{"Line", nui.ChartLine}, {"Bars", nui.ChartBar}, {"Area", nui.ChartArea},
			{"Scatter", nui.ChartScatter}, {"Pie", nui.ChartPie}, {"Donut", nui.ChartDonut},
		} {
			variant, color := nui.Outline, nui.Neutral
			if s.ChartKind == it.kind {
				variant, color = nui.Soft, nui.Primary
			}
			nui.Button(c.Key(it.label), t, it.label, nui.ButtonOptions{Variant: variant, Color: color}).OnClick(func() { s.ChartKind = it.kind })
		}
	})
	ui.Row(c.Key("chart-options")).Wrap().Gap(12).Children(func() {
		ui.Column(c.Key("axis")).Gap(4).Children(func() {
			ui.Text(c, "X axis").FontSize(12).TextColor(t.Muted)
			nui.Select(c, t, &s.ChartAxis, []nui.Option[string]{{Value: "numeric", Label: "Numeric"}, {Value: "category", Label: "Categories"}, {Value: "time", Label: "Time"}}, nui.SelectOptions{Label: "Chart X axis", Disabled: radial, OnChange: s.updateChartData}).Width(150)
		})
		ui.Column(c.Key("curve")).Gap(4).Children(func() {
			ui.Text(c, "Curve").FontSize(12).TextColor(t.Muted)
			nui.Select(c, t, &s.ChartCurve, []nui.Option[nui.ChartCurve]{
				{Value: nui.ChartCurveLinear, Label: "Linear"}, {Value: nui.ChartCurveMonotone, Label: "Monotone"},
				{Value: nui.ChartCurveStepAfter, Label: "Step after"}, {Value: nui.ChartCurveStepBefore, Label: "Step before"}, {Value: nui.ChartCurveStep, Label: "Step middle"},
			}, nui.SelectOptions{Label: "Chart curve", Disabled: s.ChartKind != nui.ChartLine && s.ChartKind != nui.ChartArea}).Width(160)
		})
		ui.Column(c.Key("stack")).Gap(4).Children(func() {
			ui.Text(c, "Stack").FontSize(12).TextColor(t.Muted)
			nui.Select(c, t, &s.ChartStack, []nui.Option[nui.ChartStackMode]{
				{Value: nui.ChartStackNone, Label: "None"}, {Value: nui.ChartStackStandard, Label: "Standard"},
				{Value: nui.ChartStackNormalized, Label: "Normalized"}, {Value: nui.ChartStackCenter, Label: "Centered"},
			}, nui.SelectOptions{Label: "Chart stacking", Disabled: s.ChartKind != nui.ChartBar && s.ChartKind != nui.ChartArea}).Width(160)
		})
	})
	ui.Row(c.Key("chart-actions")).Wrap().AlignItems(ui.Center).Gap(12).Children(func() {
		nui.Button(c, t, "Update data", nui.ButtonOptions{}).OnClick(func() { s.ChartRevision++; s.updateChartData() })
		nui.Button(c, t, "Reset data", nui.ButtonOptions{Variant: nui.Ghost, Color: nui.Neutral}).OnClick(func() {
			s.ChartRevision, s.ChartNegative, s.ChartGaps = 0, false, false
			s.updateChartData()
			for _, id := range []string{"build", "review", "deploy"} {
				s.Chart.SetSeriesVisible(id, true)
			}
			for _, id := range []string{"passed", "queued", "failed", "skipped"} {
				s.Chart.SetSliceVisible(id, true)
			}
		})
		nui.Checkbox(c, t, &s.ChartNegative, nui.CheckboxOptions{Label: "Negative values", Disabled: radial, OnChange: func(bool) { s.updateChartData() }})
		nui.Checkbox(c, t, &s.ChartGaps, nui.CheckboxOptions{Label: "Data gaps", Disabled: radial, OnChange: func(bool) { s.updateChartData() }})
		nui.Checkbox(c, t, &a.Animations, nui.CheckboxOptions{Label: "Animate charts"})
	})
	t.Motion.Reduce = t.Motion.Reduce || !a.Animations
	nui.Chart(c.Key("activity-chart"), t, &s.Chart, nui.ChartOptions{
		Label: "Workspace activity", Kind: s.ChartKind, Curve: s.ChartCurve, Stack: s.ChartStack, Height: 220,
		FormatTime: func(v time.Time) string { return v.Format("Jan 2006") },
	})
	ui.Text(c, fmt.Sprintf("Dataset %d · Click a legend to hide or show it. Focus the chart and use arrow keys to inspect values.", s.ChartRevision+1)).FontSize(12).TextColor(t.Muted)
}
