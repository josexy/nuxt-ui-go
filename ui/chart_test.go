package nui

import (
	"crypto/sha256"
	"math"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	plot "github.com/egoist/mygo/ui/chart"
)

func TestChartAxesAndAtomicUpdates(t *testing.T) {
	var s ChartState
	labels := []string{"First", "Second", "Third"}
	values := []float64{3, 7, math.Inf(1)}
	series := []ChartSeries{{ID: "a", Label: "Alpha", Values: values}}
	if err := s.SetCategories(labels, series); err != nil {
		t.Fatal(err)
	}
	labels[0], values[0] = "mutated", 90
	p := s.prepareChart(chartConfig{width: 600, kind: ChartBar, padding: .25})
	if s.categories[0] != "First" || s.series[0].Values[0] != 3 || !math.IsNaN(s.series[0].Values[2]) {
		t.Fatal("input ownership or gap normalization")
	}
	if !(p.positions[0] > 0 && p.positions[2] < 1 && p.positions[0] < p.positions[1]) {
		t.Fatal("band positions", p.positions)
	}
	if ticks := s.chartTicks(p, ChartOptions{}, 4); len(ticks) != 3 || ticks[1].Label != "Second" {
		t.Fatal(ticks)
	}
	s.SetSeriesVisible("a", false)
	revision := s.revision
	if err := s.SetCategories([]string{"duplicate", "duplicate"}, nil); err == nil {
		t.Fatal("duplicate labels accepted")
	}
	if err := s.SetData([]float64{1, 2}, series); err == nil {
		t.Fatal("unequal series accepted")
	}
	if s.revision != revision || s.SeriesVisible("a") || s.categories[0] != "First" {
		t.Fatal("failed update mutated state")
	}
	zone := time.FixedZone("test +08", 8*60*60)
	times := []time.Time{time.Date(2026, 10, 31, 0, 0, 0, 0, zone), time.Date(2026, 11, 1, 0, 0, 0, 0, zone), time.Date(2026, 11, 3, 0, 0, 0, 0, zone)}
	if err := s.SetTimeData(times, series); err != nil {
		t.Fatal(err)
	}
	s.SetSeriesVisible("a", true)
	p = s.prepareChart(chartConfig{width: 600})
	if math.Abs(float64(p.positions[1])-1.0/3) > 1e-6 {
		t.Fatal("time spacing", p.positions)
	}
	for _, tick := range s.chartTicks(p, ChartOptions{FormatTime: func(tm time.Time) string { return tm.Format("01-02 MST") }}, 4) {
		if !strings.Contains(tick.Label, "test +08") {
			t.Fatal("time zone lost", tick.Label)
		}
	}
	if err := s.SetTimeData([]time.Time{times[1], times[0]}, nil); err == nil {
		t.Fatal("decreasing times accepted")
	}
}

func TestChartStackVisibilityAndDenseGaps(t *testing.T) {
	var s ChartState
	if err := s.SetData([]float64{0, 1, 2}, []ChartSeries{
		{ID: "a", Values: []float64{2, 4, 3}}, {ID: "b", Values: []float64{3, -2, math.NaN()}}, {ID: "c", Values: []float64{1, -1, 2}},
	}); err != nil {
		t.Fatal(err)
	}
	check := func(p *chartPrepared, series, point int, lo, hi float64) {
		t.Helper()
		m := p.scene.marks[series]
		if math.Abs(p.y.Invert(m.low[point])-lo) > 1e-5 || math.Abs(p.y.Invert(m.y[point])-hi) > 1e-5 {
			t.Fatalf("series %d point %d: %g..%g, want %g..%g", series, point, p.y.Invert(m.low[point]), p.y.Invert(m.y[point]), lo, hi)
		}
	}
	config := chartConfig{width: 600, kind: ChartBar, stack: ChartStackStandard}
	p := s.prepareChart(config)
	check(p, 1, 0, 2, 5)
	check(p, 2, 1, -3, -2)
	if !math.IsNaN(float64(p.scene.marks[1].y[2])) {
		t.Fatal("stack bridged gap")
	}
	config.stack = ChartStackNormalized
	p = s.prepareChart(config)
	check(p, 2, 0, 5.0/6, 1)
	check(p, 1, 1, -2.0/7, 0)
	config.stack = ChartStackCenter
	check(s.prepareChart(config), 0, 0, -3, -1)
	config.stack = ChartStackStandard
	s.SetSeriesVisible("a", false)
	p = s.prepareChart(config)
	check(p, 0, 0, 0, 3)
	check(p, 1, 0, 3, 4)
	if p.y.Hi >= 6 {
		t.Fatal("hidden series remained in domain")
	}
	s.SetSeriesVisible("b", false)
	s.SetSeriesVisible("c", false)
	if !s.prepareChart(config).scene.empty {
		t.Fatal("all hidden not empty")
	}
	if !s.SetSeriesVisible("c", true) || s.prepareChart(config).scene.empty {
		t.Fatal("cannot restore hidden series")
	}

	const n, width = 100001, 1000
	xs, a, b := make([]float64, n), make([]float64, n), make([]float64, n)
	for i := range xs {
		xs[i], a[i], b[i] = float64(i), 2, 3
	}
	a[54], b[51] = 100, math.NaN()
	if err := s.SetData(xs, []ChartSeries{{ID: "a", Values: a}, {ID: "b", Values: b}}); err != nil {
		t.Fatal(err)
	}
	s.SetSeriesVisible("a", true)
	s.SetSeriesVisible("b", true)
	config.width = width
	p = s.prepareChart(config)
	if s.SampleCount() > 8*width {
		t.Fatal("unbounded stacked samples", s.SampleCount())
	}
	if !slices.Equal(p.scene.marks[0].x, p.scene.marks[1].x) {
		t.Fatal("stack boundaries sampled at different X")
	}
	if !slices.Contains(p.scene.marks[0].x, p.positions[54]) || p.scene.marks[0].x[len(p.scene.marks[0].x)-1] != p.positions[n-1] {
		t.Fatal("lost spike or final point")
	}
	if !math.IsNaN(float64(p.scene.marks[1].y[0])) {
		t.Fatal("dense bucket connected a gap")
	}
}

func TestChartRadialDataAndVisibility(t *testing.T) {
	var s ChartState
	data := []ChartSlice{{ID: "a", Value: 3}, {ID: "b", Value: 1}, {ID: "zero"}}
	if err := s.SetSlices(data); err != nil {
		t.Fatal(err)
	}
	data[0].Value = 99
	config := chartConfig{width: 300, kind: ChartDonut, inner: .6}
	p := s.prepareChart(config)
	if len(p.scene.wedges) != 2 || math.Abs(p.scene.wedges[0].arc.End-1.5*math.Pi) > 1e-6 {
		t.Fatal("radial proportions or copied data")
	}
	s.SetSliceVisible("a", false)
	if err := s.SetSlices([]ChartSlice{{ID: "b", Value: 2}, {ID: "a", Value: 4}}); err != nil {
		t.Fatal(err)
	}
	if s.SliceVisible("a") || len(s.prepareChart(config).scene.wedges) != 1 {
		t.Fatal("visibility did not follow ID")
	}
	revision := s.revision
	for _, value := range []float64{-1, math.NaN(), math.Inf(1)} {
		if err := s.SetSlices([]ChartSlice{{ID: "bad", Value: value}}); err == nil {
			t.Fatal("invalid slice", value)
		}
	}
	if err := s.SetSlices([]ChartSlice{{ID: "duplicate"}, {ID: "duplicate"}}); err == nil {
		t.Fatal("duplicate ID accepted")
	}
	if s.revision != revision {
		t.Fatal("invalid slices mutated state")
	}
}

func TestChartAnimationRetargetAndBoundedLayers(t *testing.T) {
	var s ChartState
	config := chartConfig{width: 600, kind: ChartLine}
	data := func(values []float64) *chartScene {
		t.Helper()
		if err := s.SetData([]float64{0, 1, 2}, []ChartSeries{{ID: "a", Values: values}}); err != nil {
			t.Fatal(err)
		}
		return s.prepareChart(config).scene
	}
	now := time.Unix(0, 0)
	var a chartAnimation
	first := data([]float64{2, 8, 3})
	a.set(first, now, time.Second, true)
	target := data([]float64{8, 2, 3})
	a.set(target, now, time.Second, true)
	mid, active := a.layers(now.Add(300 * time.Millisecond))
	if !active || !a.tween || mid[0].scene.marks[0].y[0] <= first.marks[0].y[0] || mid[0].scene.marks[0].y[0] >= target.marks[0].y[0] {
		t.Fatal("data did not interpolate")
	}
	a.set(data([]float64{4, 8, 2}), now.Add(300*time.Millisecond), time.Second, true)
	retarget, _ := a.layers(now.Add(300 * time.Millisecond))
	if !slices.Equal(retarget[0].scene.marks[0].y, mid[0].scene.marks[0].y) {
		t.Fatal("interrupted data animation jumped")
	}
	for i := range 20 {
		config.kind = []ChartKind{ChartBar, ChartArea, ChartScatter, ChartLine}[i%4]
		at := now.Add(time.Duration(400+i*10) * time.Millisecond)
		a.set(s.prepareChart(config).scene, at, time.Second, true)
		layers, _ := a.layers(at.Add(5 * time.Millisecond))
		if len(layers) > 4 {
			t.Fatal("unbounded outgoing layers", len(layers))
		}
	}
	if layers, active := a.layers(now.Add(3 * time.Second)); active || len(layers) != 1 {
		t.Fatal("transition did not settle")
	}
	a.set(first, now, time.Second, false)
	if layers, active := a.layers(now); active || layers[0].scene != first {
		t.Fatal("disabled animation did not snap")
	}
	if err := s.SetSlices([]ChartSlice{{ID: "a", Value: 3}, {ID: "b", Value: 1}}); err != nil {
		t.Fatal(err)
	}
	pie := s.prepareChart(chartConfig{kind: ChartPie, width: 600}).scene
	a.set(pie, now, time.Second, false)
	if err := s.SetSlices([]ChartSlice{{ID: "a", Value: 1}, {ID: "b", Value: 3}}); err != nil {
		t.Fatal(err)
	}
	donut := s.prepareChart(chartConfig{kind: ChartDonut, width: 600, inner: .6}).scene
	a.set(donut, now, time.Second, true)
	mid, active = a.layers(now.Add(300 * time.Millisecond))
	if !active || !a.tween || mid[0].scene.config.inner <= 0 || mid[0].scene.config.inner >= .6 || mid[0].scene.wedges[0].arc.End >= pie.wedges[0].arc.End || mid[0].scene.wedges[0].arc.End <= donut.wedges[0].arc.End {
		t.Fatal("pie/donut data and radius did not interpolate")
	}
}

func TestChartInteractionAndAnimationLifecycle(t *testing.T) {
	var s ChartState
	setData := func(a, b float64) {
		t.Helper()
		if err := s.SetCategories([]string{"Jan", "Feb", "Mar"}, []ChartSeries{{ID: "a", Label: "Alpha", Values: []float64{a, b, 4}}, {ID: "b", Label: "Beta", Values: []float64{3, 6, 2}}}); err != nil {
			t.Fatal(err)
		}
	}
	setData(2, 8)
	if err := s.SetSlices([]ChartSlice{{ID: "one", Label: "One", Value: 3}, {ID: "two", Label: "Two", Value: 1}}); err != nil {
		t.Fatal(err)
	}
	shown, offscreen, enabled, interactive := true, false, true, true
	rtl := false
	kind := ChartLine
	legendCalls := 0
	focusCalls, selectCalls := 0, 0
	var activated ChartSelection
	var plotBounds ui.Rect
	tt := ui.NewTester(func(c *ui.Context) {
		theme := App(c, AppConfig{})
		direction := ui.LeftToRight
		if rtl {
			direction = ui.RightToLeft
		}
		ui.Scroll(c).Fill().Direction(direction).Children(func() {
			if offscreen {
				ui.Box(c).Height(700)
			}
			if !shown {
				return
			}
			Chart(c, theme, &s, ChartOptions{Label: "Metrics", Kind: kind, Height: 220, Animate: Set(enabled), InteractiveLegend: Set(interactive), OnLegendChange: func(string, bool) { legendCalls++ }, OnFocusChange: func(ChartSelection) { focusCalls++ }, OnSelect: func(point ChartSelection) { selectCalls++; activated = point }})
			plotBounds = s.handle.Bounds(c)
		})
	}, 640, 400)
	before := sha256.Sum256(tt.Image().Pix)
	setData(8, 2)
	tt.Frame()
	tt.Advance(100 * time.Millisecond)
	mid := sha256.Sum256(tt.Image().Pix)
	if !tt.FramePending() || mid == before {
		t.Fatal("no visible data animation")
	}
	tt.Advance(400 * time.Millisecond)
	if after := sha256.Sum256(tt.Image().Pix); mid == after || tt.FramePending() {
		t.Fatal("data animation not settled or unchanged")
	}
	kind = ChartBar
	tt.Frame()
	tt.Advance(100 * time.Millisecond)
	mid = sha256.Sum256(tt.Image().Pix)
	if !tt.FramePending() {
		t.Fatal("no type transition")
	}
	tt.Advance(400 * time.Millisecond)
	if after := sha256.Sum256(tt.Image().Pix); mid == after || tt.FramePending() {
		t.Fatal("type transition did not finish")
	}
	click := func(label string) {
		t.Helper()
		if err := tt.Click(label); err != nil {
			t.Fatal(err)
		}
	}
	click("Alpha")
	if s.SeriesVisible("a") || legendCalls != 1 {
		t.Fatal("legend click")
	}
	tt.Key(0, ui.KeySpace)
	if !s.SeriesVisible("a") || legendCalls != 2 {
		t.Fatal("keyboard legend")
	}
	click("Alpha")
	click("Beta")
	if _, ok := tt.Find("No chart data"); !ok {
		t.Fatal("all hidden empty state")
	}
	click("Alpha")
	if !s.SeriesVisible("a") {
		t.Fatal("all hidden legend could not restore")
	}
	interactive = false
	tt.Frame()
	click("Alpha")
	if !s.SeriesVisible("a") {
		t.Fatal("noninteractive legend toggled")
	}
	interactive = true
	kind = ChartDonut
	tt.Frame()
	tt.Advance(time.Second)
	click("Metrics")
	focusCalls, selectCalls = 0, 0
	tt.Key(0, ui.KeyHome)
	if s.SelectedSlice != "one" {
		t.Fatal("radial Home", s.SelectedSlice)
	}
	tt.Key(0, ui.KeyHome)
	tt.Key(0, ui.KeyEnter)
	tt.Key(0, ui.KeySpace)
	if focusCalls != 1 || selectCalls != 2 || activated.SliceID != "one" {
		t.Fatalf("chart notifications: focus=%d select=%d point=%+v", focusCalls, selectCalls, activated)
	}
	tt.Key(0, ui.KeyRight)
	if s.SelectedSlice != "two" {
		t.Fatal("radial Right", s.SelectedSlice)
	}
	// Pointer hit testing uses the ring, excluding the hole.
	arc := s.prepared.scene.wedges[0].arc
	x, y := plot.Polar(plotBounds.X+plotBounds.W/2, plotBounds.Y+plotBounds.H/2, chartRadius(plotBounds)*.8, arc.Mid())
	tt.Move(x, y)
	if s.SelectedSlice != "one" {
		t.Fatal("radial pointer", s.SelectedSlice)
	}
	tt.Move(plotBounds.X+plotBounds.W/2, plotBounds.Y+plotBounds.H/2)
	if s.SelectedSlice != "" {
		t.Fatal("donut hole selected a slice")
	}
	click("One")
	if s.SliceVisible("one") {
		t.Fatal("radial legend")
	}
	tt.Move(0, 0)

	kind = ChartLine
	tt.Frame()
	tt.Advance(time.Second)
	enabled = false
	setData(2, 8)
	tt.Frame()
	tt.Advance(50 * time.Millisecond)
	if tt.FramePending() {
		t.Fatal("Animate=false ignored")
	}
	enabled = true
	tt.SetPreferences(ui.Preferences{ReduceMotion: true})
	setData(8, 2)
	tt.Frame()
	tt.Advance(50 * time.Millisecond)
	if tt.FramePending() {
		t.Fatal("reduced motion ignored")
	}
	tt.SetPreferences(ui.Preferences{})
	setData(2, 8)
	tt.Frame()
	offscreen = true
	tt.Frame()
	tt.Advance(50 * time.Millisecond)
	if tt.FramePending() {
		t.Fatal("offscreen chart keeps scheduling")
	}
	offscreen = false
	tt.Frame()
	shown = false
	tt.Frame()
	tt.Advance(50 * time.Millisecond)
	if tt.FramePending() {
		t.Fatal("unmounted chart keeps scheduling")
	}
	shown = true
	tt.Frame()
	tt.Advance(time.Second)
	if tt.FramePending() {
		t.Fatal("returning chart did not settle")
	}
	rtl, enabled = true, false
	tt.Frame()
	click("Metrics")
	tt.Key(0, ui.KeyHome)
	tt.Key(0, ui.KeyLeft)
	if s.Selected != 1 {
		t.Fatalf("RTL left selected %d", s.Selected)
	}
	tt.Key(0, ui.KeyRight)
	if s.Selected != 0 {
		t.Fatal("RTL right did not return to first X value")
	}
	tt.Move(plotBounds.X+plotBounds.W*(1-s.prepared.positions[2]), plotBounds.Y+plotBounds.H/2)
	if s.Selected != 2 {
		t.Fatalf("RTL pointer selected %d", s.Selected)
	}
	jan, _ := tt.Find("Jan")
	mar, _ := tt.Find("Mar")
	if jan.X <= mar.X {
		t.Fatalf("RTL axis labels do not follow reversed data: Jan=%+v Mar=%+v", jan, mar)
	}
	ytick, ok := tt.Find("8")
	if !ok || ytick.X < plotBounds.X+plotBounds.W-1 {
		t.Fatalf("RTL Y labels are not on the right: tick=%+v plot=%+v", ytick, plotBounds)
	}
}

func TestChartDefaultSeriesColorsAreDistinct(t *testing.T) {
	for _, th := range []Theme{Light(), Dark()} {
		shade := 500
		if th.Dark {
			shade = 400
		}
		want := []ui.Color{th.Primary, BluePalette().Shade(shade), VioletPalette().Shade(shade), AmberPalette().Shade(shade), RosePalette().Shade(shade), CyanPalette().Shade(shade)}
		seen := map[ui.Color]bool{}
		for i, color := range want {
			got := chartColor(th, ColorDefault, i)
			if got != color || seen[got] {
				t.Fatalf("dark=%v series %d color %v, want distinct %v", th.Dark, i, got, color)
			}
			seen[got] = true
		}
		if chartColor(th, ColorDefault, 6) == th.Primary || chartColor(th, Error, 3) != th.Error || chartColor(th, Neutral, 0) != th.InvertedBackground {
			t.Fatal("later series must use lighter colors while explicit colors and neutral remain distinct")
		}
	}
}
