package nui

import (
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/egoist/mygo/ui"
	plot "github.com/egoist/mygo/ui/chart"
)

type ChartKind uint8

const (
	ChartLine ChartKind = iota
	ChartBar
	ChartArea
	ChartScatter
	ChartPie
	ChartDonut
)

type ChartCurve uint8

const (
	ChartCurveLinear ChartCurve = iota
	ChartCurveMonotone
	ChartCurveStepAfter
	ChartCurveStepBefore
	ChartCurveStep
)

type ChartStackMode uint8

const (
	ChartStackNone ChartStackMode = iota
	ChartStackStandard
	ChartStackNormalized
	ChartStackCenter
)

type ChartSeries struct {
	ID, Label string
	Values    []float64
	Color     Color
}

// ChartSlice is one independently named and colored pie/donut segment.
// Values must be finite and nonnegative. Zero values remain in the legend.
type ChartSlice struct {
	ID, Label string
	Value     float64
	Color     Color
}

// ChartSelection identifies the focused X row or radial slice. Index is -1 when
// nothing is focused; SliceID is empty for Cartesian charts.
type ChartSelection struct {
	Index   int
	SliceID string
}

func (s *ChartState) chartSelection(kind ChartKind) ChartSelection {
	if chartRadial(kind) {
		for i, slice := range s.slices {
			if slice.ID == s.SelectedSlice {
				return ChartSelection{Index: i, SliceID: slice.ID}
			}
		}
		return ChartSelection{Index: -1}
	}
	if s.Selected < 0 || s.Selected >= len(s.x) {
		return ChartSelection{Index: -1}
	}
	return ChartSelection{Index: s.Selected}
}

type chartAxis uint8

const (
	chartNumeric chartAxis = iota
	chartCategory
	chartTime
)

// ChartState owns copied data and per-window selection, visibility and animation.
// SetData, SetCategories and SetTimeData replace Cartesian data; SetSlices replaces
// radial data independently. Publish worker-prepared state through Window.Update;
// never mutate a state while a view is using it, or copy an already rendered state.
type ChartState struct {
	Selected      int
	SelectedSlice string
	x             []float64
	series        []ChartSeries
	categories    []string
	times         []time.Time
	slices        []ChartSlice
	axis          chartAxis
	hidden        map[string]bool
	hiddenSlices  map[string]bool
	revision      uint64
	prepared      *chartPrepared
	animation     chartAnimation
	handle        ui.Handle
	frame         ui.Handle
	piePointer    ui.Point
	pieHovered    bool
}

func chartFinite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

func (s *ChartState) setData(xs []float64, series []ChartSeries, axis chartAxis, labels []string, times []time.Time) error {
	for i, x := range xs {
		if !chartFinite(x) || (i > 0 && x <= xs[i-1]) {
			return fmt.Errorf("nui: chart X must be finite and strictly increasing")
		}
	}
	seen := map[string]bool{}
	copySeries := slices.Clone(series)
	for i, v := range series {
		if v.ID == "" || seen[v.ID] {
			return fmt.Errorf("nui: chart requires unique nonempty series IDs")
		}
		seen[v.ID] = true
		if len(v.Values) != len(xs) {
			return fmt.Errorf("nui: chart series %s has %d values for %d X values", v.ID, len(v.Values), len(xs))
		}
		copySeries[i].Values = slices.Clone(v.Values)
		if copySeries[i].Label == "" {
			copySeries[i].Label = v.ID
		}
		for j, y := range copySeries[i].Values {
			if !chartFinite(y) {
				copySeries[i].Values[j] = math.NaN()
			}
		}
	}
	s.x, s.series, s.axis = slices.Clone(xs), copySeries, axis
	s.categories, s.times = slices.Clone(labels), slices.Clone(times)
	for id := range s.hidden {
		if !seen[id] {
			delete(s.hidden, id)
		}
	}
	s.Selected = -1
	s.invalidateChart()
	return nil
}

// SetData sets a shared numeric X axis. NaN/Inf Y values are gaps, not zeros.
// An error leaves the previous data and visibility unchanged.
func (s *ChartState) SetData(xs []float64, series []ChartSeries) error {
	return s.setData(xs, series, chartNumeric, nil, nil)
}

// SetCategories sets a band axis with unique, nonempty category labels.
func (s *ChartState) SetCategories(labels []string, series []ChartSeries) error {
	xs := make([]float64, len(labels))
	seen := map[string]bool{}
	for i, label := range labels {
		if label == "" || seen[label] {
			return fmt.Errorf("nui: chart categories must be unique and nonempty")
		}
		seen[label], xs[i] = true, float64(i)
	}
	return s.setData(xs, series, chartCategory, labels, nil)
}

// SetTimeData sets a strictly increasing time axis; ticks use the first time's location.
func (s *ChartState) SetTimeData(times []time.Time, series []ChartSeries) error {
	xs := make([]float64, len(times))
	for i, tm := range times {
		if i > 0 && !tm.After(times[i-1]) {
			return fmt.Errorf("nui: chart times must be strictly increasing")
		}
		xs[i] = tm.Sub(times[0]).Seconds()
	}
	return s.setData(xs, series, chartTime, nil, times)
}

func (s *ChartState) SetSlices(values []ChartSlice) error {
	seen := map[string]bool{}
	total := 0.0
	for _, v := range values {
		if v.ID == "" || seen[v.ID] {
			return fmt.Errorf("nui: chart requires unique nonempty slice IDs")
		}
		seen[v.ID] = true
		if !chartFinite(v.Value) || v.Value < 0 {
			return fmt.Errorf("nui: chart slice %s must be finite and nonnegative", v.ID)
		}
		total += v.Value
	}
	if !chartFinite(total) {
		return fmt.Errorf("nui: chart slice total is too large")
	}
	s.slices = slices.Clone(values)
	for i := range s.slices {
		if s.slices[i].Label == "" {
			s.slices[i].Label = s.slices[i].ID
		}
	}
	for id := range s.hiddenSlices {
		if !seen[id] {
			delete(s.hiddenSlices, id)
		}
	}
	if !seen[s.SelectedSlice] {
		s.SelectedSlice = ""
	}
	s.invalidateChart()
	return nil
}

func (s *ChartState) invalidateChart() { s.revision++; s.prepared = nil }

func (s *ChartState) SeriesVisible(id string) bool {
	return slices.ContainsFunc(s.series, func(v ChartSeries) bool { return v.ID == id }) && !s.hidden[id]
}

// SetSeriesVisible returns whether a known series changed visibility.
// Visibility follows the ID across data updates and reordering.
func (s *ChartState) SetSeriesVisible(id string, visible bool) bool {
	if !slices.ContainsFunc(s.series, func(v ChartSeries) bool { return v.ID == id }) || s.SeriesVisible(id) == visible {
		return false
	}
	if s.hidden == nil {
		s.hidden = map[string]bool{}
	}
	s.hidden[id] = !visible
	s.invalidateChart()
	return true
}

func (s *ChartState) SliceVisible(id string) bool {
	return slices.ContainsFunc(s.slices, func(v ChartSlice) bool { return v.ID == id }) && !s.hiddenSlices[id]
}

func (s *ChartState) SetSliceVisible(id string, visible bool) bool {
	if !slices.ContainsFunc(s.slices, func(v ChartSlice) bool { return v.ID == id }) || s.SliceVisible(id) == visible {
		return false
	}
	if s.hiddenSlices == nil {
		s.hiddenSlices = map[string]bool{}
	}
	s.hiddenSlices[id] = !visible
	if !visible && s.SelectedSlice == id {
		s.SelectedSlice = ""
	}
	s.invalidateChart()
	return true
}

type chartConfig struct {
	rtl        bool
	width      int
	kind       ChartKind
	curve      ChartCurve
	stack      ChartStackMode
	padding    float32
	inner, pad float64
}

type chartMark struct {
	id          string
	color       Color
	index       int
	x, y, low   []float32
	left, right []float32
}

type chartWedge struct {
	id    string
	color Color
	index int
	arc   plot.Slice
}

// Scenes contain only bounded drawing geometry, never the full source arrays.
type chartScene struct {
	revision uint64
	config   chartConfig
	marks    []chartMark
	wedges   []chartWedge
	empty    bool
}

type chartPrepared struct {
	scene     *chartScene
	x, y      plot.Linear
	time      plot.Time
	band      plot.Band
	positions []float32
	highs     [][]float32
	indices   []int
}

// Prepare prewarms the default line chart cache. Other options are prepared by Chart.
func (s *ChartState) Prepare(width int) { s.prepareChart(chartConfig{width: width, padding: .25}) }

func (s *ChartState) SampleCount() int {
	if s.prepared == nil {
		return 0
	}
	n := len(s.prepared.scene.wedges)
	for _, mark := range s.prepared.scene.marks {
		n += len(mark.x)
	}
	return n
}

func (s *ChartState) prepareChart(config chartConfig) *chartPrepared {
	config.width = max(1, min(2048, config.width))
	if s.prepared != nil && s.prepared.scene.config == config && s.prepared.scene.revision == s.revision {
		return s.prepared
	}
	p := &chartPrepared{scene: &chartScene{revision: s.revision, config: config, empty: true}}
	s.prepared = p
	if config.kind == ChartPie || config.kind == ChartDonut {
		values := []float64{}
		for i, v := range s.slices {
			if !s.hiddenSlices[v.ID] && v.Value > 0 {
				values = append(values, v.Value)
				p.scene.wedges = append(p.scene.wedges, chartWedge{id: v.ID, color: v.Color, index: i})
			}
		}
		for i, arc := range plot.Pie(values, config.pad) {
			p.scene.wedges[i].arc = arc
		}
		p.scene.empty = len(values) == 0
		return p
	}
	n := len(s.x)
	p.x = plot.NewLinear(0, 1)
	if n > 0 {
		lo, hi := s.x[0], s.x[n-1]
		if n == 1 {
			lo, hi = lo-.5, hi+.5
		} else if config.kind == ChartBar && s.axis != chartCategory {
			lo -= (s.x[1] - s.x[0]) / 2
			hi += (s.x[n-1] - s.x[n-2]) / 2
		}
		p.x = plot.NewLinear(lo, hi)
		if s.axis == chartTime {
			p.time = plot.NewTime(s.times[0].Add(time.Duration(lo*float64(time.Second))), s.times[0].Add(time.Duration(hi*float64(time.Second))))
		}
	}
	p.band = plot.NewBand(s.categories...).Padding(config.padding, .15)
	p.positions = make([]float32, n)
	for i, x := range s.x {
		p.positions[i] = p.x.At(x)
		switch s.axis {
		case chartCategory:
			p.positions[i] = p.band.At(i)
		case chartTime:
			p.positions[i] = p.time.At(s.times[i])
		}
	}
	values := [][]float64{}
	for i, v := range s.series {
		if !s.hidden[v.ID] {
			values = append(values, v.Values)
			p.indices = append(p.indices, i)
		}
	}
	stacked := [][]plot.Segment(nil)
	if config.stack != ChartStackNone {
		stacked = plot.Stack(values, plot.Stacking(config.stack-1))
	}
	lo, hi := 0.0, 0.0
	for i, ys := range values {
		for j, y := range ys {
			if !chartFinite(y) {
				continue
			}
			a, b := 0.0, y
			if stacked != nil {
				a, b = stacked[i][j].Lo, stacked[i][j].Hi
			}
			if !chartFinite(a) || !chartFinite(b) {
				continue
			}
			p.scene.empty = false
			lo, hi = min(lo, a, b), max(hi, a, b)
		}
	}
	if lo == hi {
		hi = lo + 1
	}
	p.y = plot.NewLinear(lo, hi).Nice(4)
	lows := make([][]float32, len(values))
	p.highs = make([][]float32, len(values))
	for i, ys := range values {
		lows[i], p.highs[i] = make([]float32, n), make([]float32, n)
		for j, y := range ys {
			a, b := 0.0, y
			if stacked != nil {
				a, b = stacked[i][j].Lo, stacked[i][j].Hi
			}
			lows[i][j], p.highs[i][j] = p.y.At(a), p.y.At(b)
			if !chartFinite(y) || !chartFinite(a) || !chartFinite(b) {
				lows[i][j], p.highs[i][j] = float32(math.NaN()), float32(math.NaN())
			}
		}
	}
	// Stack boundaries share sample positions. Preserve each bucket's endpoints
	// and aggregate envelope extrema; a series' gaps remain gaps for that bucket.
	type bucket struct {
		from, to int
		samples  []int
	}
	var buckets []bucket
	if stacked != nil {
		step := max(1, (n+config.width-1)/config.width)
		envelope := make([]float64, n)
		for i := range values {
			for j, segment := range stacked[i] {
				if chartFinite(values[i][j]) {
					envelope[j] += math.Abs(segment.Hi - segment.Lo)
				}
			}
		}
		for from := 0; from < n; from += step {
			to, low, high := min(n, from+step), from, from
			for j := from + 1; j < to; j++ {
				if envelope[j] < envelope[low] {
					low = j
				}
				if envelope[j] > envelope[high] {
					high = j
				}
			}
			samples := []int{from, low, high, to - 1}
			slices.Sort(samples)
			buckets = append(buckets, bucket{from, to, slices.Compact(samples)})
		}
	}
	for i, ys := range p.highs {
		index := p.indices[i]
		series := s.series[index]
		mark := chartMark{id: series.ID, color: series.Color, index: index}
		if stacked == nil {
			mark.x, mark.y = chartDownsample(p.positions, ys, config.width)
		} else {
			for _, bucket := range buckets {
				gap := false
				for j := bucket.from; j < bucket.to; j++ {
					gap = gap || math.IsNaN(float64(ys[j]))
				}
				for _, j := range bucket.samples {
					mark.x = append(mark.x, p.positions[j])
					value := ys[j]
					if gap {
						value = float32(math.NaN())
					}
					mark.y = append(mark.y, value)
				}
			}
		}
		for _, x := range mark.x {
			j := plot.Nearest(p.positions, x)
			if j < 0 || math.IsNaN(float64(x)) {
				mark.low = append(mark.low, float32(math.NaN()))
				mark.left, mark.right = append(mark.left, x), append(mark.right, x)
				continue
			}
			mark.low = append(mark.low, lows[i][j])
			bw := float32(.7)
			if n > 1 {
				bw *= p.positions[min(n-1, j+1)] - p.positions[max(0, j-1)]
				if j > 0 && j < n-1 {
					bw /= 2
				}
			}
			if s.axis == chartCategory {
				bw = p.band.Width()
			}
			left := x - bw/2
			if stacked == nil {
				bw /= float32(max(1, len(values)))
				left += float32(i) * bw
			}
			mark.left, mark.right = append(mark.left, left), append(mark.right, left+bw)
		}
		p.scene.marks = append(p.scene.marks, mark)
	}
	if config.rtl {
		for i := range p.scene.marks {
			mark := &p.scene.marks[i]
			for j := range mark.x {
				mark.x[j] = 1 - mark.x[j]
				mark.left[j], mark.right[j] = 1-mark.right[j], 1-mark.left[j]
			}
		}
	}
	return p
}

func chartDownsample(xs, ys []float32, width int) ([]float32, []float32) {
	x, y := plot.MinMax(nil, nil, xs, ys, width)
	if len(x) <= 4*width+4 {
		return x, y
	}
	x, y = nil, nil
	step := max(1, (len(ys)+width-1)/width)
	for from := 0; from < len(ys); from += step {
		to, low, high, gap := min(len(ys), from+step), from, from, false
		for j := from; j < to; j++ {
			gap = gap || math.IsNaN(float64(ys[j]))
			if ys[j] < ys[low] {
				low = j
			}
			if ys[j] > ys[high] {
				high = j
			}
		}
		if gap {
			x, y = append(x, xs[from]), append(y, float32(math.NaN()))
			continue
		}
		if low > high {
			low, high = high, low
		}
		x, y = append(x, xs[low]), append(y, ys[low])
		if low != high {
			x, y = append(x, xs[high]), append(y, ys[high])
		}
	}
	return x, y
}
