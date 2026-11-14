package nui

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
	plot "github.com/egoist/mygo/ui/chart"
)

type ChartOptions struct {
	Label, Description, EmptyText string
	XLabel, YLabel                string
	Kind                          ChartKind
	Curve                         ChartCurve
	// Stack applies to bars and areas; other types ignore it.
	Stack                           ChartStackMode
	Height, Width, AspectRatio      float32
	XAxis, YAxis, Tooltip           Optional[bool]
	TooltipOptions                  *TooltipOptions
	Points                          bool
	Colors                          []ui.Color
	Thickness                       Optional[float32]
	Legend, InteractiveLegend, Grid Optional[bool]
	FormatX, FormatY                func(float64) string
	FormatTime                      func(time.Time) string
	// Animate defaults to true and respects theme/system reduced motion.
	Animate                 Optional[bool]
	AnimationDuration       time.Duration     // default 350ms
	BarPadding              Optional[float32] // category band inner padding, default .25
	DonutInnerRadius        Optional[float64] // Explicit fraction overrides Thickness.
	SlicePadding            Optional[float64] // angle in radians, default .02
	LineWidth, PointSize    Optional[float32]
	OnLegendChange          func(id string, visible bool)
	OnFocusChange, OnSelect func(ChartSelection)
	Center                  func(*ui.Context, Theme)
	Decoration
}

func Chart(c *ui.Context, t Theme, s *ChartState, o ChartOptions) ui.Element {
	if s == nil {
		panic("nui: Chart requires state")
	}
	if o.Kind > ChartDonut || o.Curve > ChartCurveStep || o.Stack > ChartStackCenter {
		panic("nui: invalid chart kind, curve or stack")
	}
	for _, dimension := range []float32{o.Height, o.Width, o.AspectRatio, valueOr(o.Thickness, float32(24)), valueOr(o.LineWidth, float32(2)), valueOr(o.PointSize, float32(7))} {
		if !chartFinite(float64(dimension)) {
			panic("nui: chart dimensions must be finite")
		}
	}
	if o.Label == "" {
		labels := []string{}
		if chartRadial(o.Kind) {
			for _, slice := range s.slices {
				labels = append(labels, slice.Label)
			}
		} else {
			for _, series := range s.series {
				labels = append(labels, series.Label)
			}
		}
		o.Label = strings.Join(labels, ", ")
		if o.Label == "" {
			o.Label = message(c, "chart")
		}
	}
	// Committed bounds are unknown before the first layout, and stay empty while
	// an ancestor is Hidden (a KeepMounted tab or collapsible). Build with a
	// guess, and let the first paint with a real box ask for one more build:
	// a hidden chart paints nothing, so it never asks.
	frameWidth := s.frame.Bounds(c).W
	guessed := frameWidth <= 0
	if o.Width > 0 {
		frameWidth = o.Width
	}
	if frameWidth <= 0 {
		frameWidth = 640
	}
	h := o.Height
	if h <= 0 {
		h = 300
		if o.AspectRatio > 0 {
			h = frameWidth / o.AspectRatio
		}
	}
	width := int(s.handle.Bounds(c).W)
	if width <= 0 {
		width = int(frameWidth)
		guessed = true
	}
	services := c.Services()
	config := chartConfig{width: width, kind: o.Kind, curve: o.Curve, stack: o.Stack, padding: valueOr(o.BarPadding, .25), pad: valueOr(o.SlicePadding, .02)}
	config.rtl = c.Direction() == ui.RightToLeft
	if o.Kind == ChartDonut {
		outer := chartRadius(ui.Rect{W: frameWidth, H: h})
		if outer > 0 {
			config.inner = max(0, 1-float64(max(0, valueOr(o.Thickness, float32(24))))/float64(outer))
		}
		if o.DonutInnerRadius.Set {
			config.inner = max(0, min(.95, o.DonutInnerRadius.Value))
		}
	}
	if o.Kind != ChartBar && o.Kind != ChartArea {
		config.stack = ChartStackNone
	}
	if !chartFinite(float64(config.padding)) || !chartFinite(config.pad) || !chartFinite(config.inner) {
		panic("nui: chart dimensions must be finite")
	}
	config.padding, config.inner, config.pad = max(0, min(1, config.padding)), max(0, min(1, config.inner)), max(0, config.pad)
	p := s.prepareChart(config)
	palette := slices.Clone(o.Colors)
	duration := o.AnimationDuration
	if duration <= 0 {
		duration = 350 * time.Millisecond
	}
	s.animation.set(p.scene, c.Now(), duration, t.motion(c) && valueOr(o.Animate, true))
	root := ui.Column(c).MinWidth(0).Gap(t.spacing(8)).Role(ui.RoleGroup).Bind(&s.frame)
	if o.Width > 0 {
		root.Width(o.Width)
	}
	finish(t, ComponentChart, Root, root, PartStyle{}, o.Decoration)
	root.Children(func() {
		var description string
		if o.YLabel != "" && !chartRadial(o.Kind) && valueOr(o.YAxis, true) {
			finish(t, ComponentChart, YAxisSlot, ui.Text(c.Key("y-title"), o.YLabel), PartStyle{Foreground: Set(t.Muted), FontSize: Set(t.text(12))}, o.Decoration)
		}
		buildArea := func(area ui.Element) {
			area.Bind(&s.handle).Label(o.Label).Description(o.Description).Role(ui.RoleGroup).Focusable().Clip().FocusRing(true)
			finish(t, ComponentChart, Body, area, PartStyle{}, o.Decoration)
			before := s.chartSelection(o.Kind)
			if chartRadial(o.Kind) && !area.IsDisabled() {
				description = s.selectSlice(c, area, p, o)
			} else if !chartRadial(o.Kind) && !p.scene.empty && !area.IsDisabled() {
				describe := func(i int) string { return s.describePoint(c, p, o, i) }
				s.selectChartPoint(c, area, p.positions, config.rtl, describe)
				description = describe(s.Selected)
			} else if p.scene.empty {
				s.Selected = -1
				s.SelectedSlice = ""
			}
			selection := s.chartSelection(o.Kind)
			if before != selection && o.OnFocusChange != nil {
				area.SetChanged().OnChange(func() { o.OnFocusChange(selection) })
			}
			if o.OnSelect != nil {
				activate := func() { o.OnSelect(selection) }
				area.OnClick(activate).OnShortcut(0, ui.KeyEnter, activate).OnShortcut(0, ui.KeySpace, activate)
			}
			animation := s.animation
			selected := s.SelectedSlice
			ticks := p.y.Ticks(4)
			lineWidth, pointSize := valueOr(o.LineWidth, float32(2)), valueOr(o.PointSize, float32(7))
			measure := guessed
			area.Draw(func(painter *ui.Painter, r ui.Rect) {
				if measure && r.W > 0 && r.H > 0 {
					// Once per build: Draw may run again in frames that only paint.
					measure = false
					services.Invalidate()
				}
				if !chartRadial(o.Kind) && !p.scene.empty && valueOr(o.Grid, true) {
					for _, tick := range ticks {
						painter.Fill(lineRect(ui.Rect{X: r.X, Y: plot.Y(r, tick.At) - .5, W: r.W, H: 1}, painter.Scale(), false, true, 0), t.BorderMuted, 0)
					}
				}
				layers, active := animation.layers(painter.Now())
				for _, layer := range layers {
					drawChartScene(painter, r, t, layer.scene, layer.alpha, selected, max(0, lineWidth), max(0, pointSize), o.Points, palette)
				}
				if active {
					painter.AnimationFrame()
				}
			})
			if p.scene.empty {
				text := o.EmptyText
				if text == "" {
					text = message(c, "noChartData")
				}
				finish(t, ComponentChart, EmptySlot, ui.Text(c, text).Absolute().Left(0).Top(8).Role(ui.RoleStatus), PartStyle{Foreground: Set(t.Muted)}, o.Decoration)
			}
			if description != "" && valueOr(o.Tooltip, true) {
				tooltip := TooltipOptions{}
				if o.TooltipOptions != nil {
					tooltip = *o.TooltipOptions
				}
				if !tooltip.Placement.Set {
					tooltip.Placement = Set(TopCenter)
				}
				if !tooltip.Offset.Set {
					tooltip.Offset = Set(float32(12))
				}
				if !tooltip.DelayDuration.Set {
					tooltip.DelayDuration = Set(time.Duration(0))
				}
				Tooltip(c, t, area, description, tooltip)
			}
			if o.Center != nil {
				center := ui.Column(c.Key("center")).Absolute().Left(0).Right(0).Top(0).Bottom(0).Center().IgnoreInput(true)
				finish(t, ComponentChart, CenterSlot, center, PartStyle{}, o.Decoration)
				center.Children(func() { o.Center(c, t) })
			}
			if !chartRadial(o.Kind) && s.Selected >= 0 && s.Selected < len(p.positions) && !p.scene.empty {
				x := p.positions[s.Selected]
				if config.rtl {
					x = 1 - x
				}
				plot.RuleX(c, x, 1).Background(t.Muted).IgnoreInput(true)
				for i, ys := range p.highs {
					if !math.IsNaN(float64(ys[s.Selected])) {
						index := p.indices[i]
						plot.Point(c, x, ys[s.Selected], 7, 7).Background(chartColorWithPalette(t, s.series[index].Color, index, palette)).Radius(4).IgnoreInput(true)
					}
				}
			}
		}
		if chartRadial(o.Kind) {
			area := ui.Box(c.Key("plot")).Height(h).MinWidth(0)
			area.Children(func() { buildArea(area) })
		} else {
			y := p.y
			y.Format = o.FormatY
			if y.Format == nil && config.stack == ChartStackNormalized {
				y.Format = func(v float64) string { return fmt.Sprintf("%.0f%%", v*100) }
			}
			axes := plot.Axes{}
			if !p.scene.empty {
				count := max(2, min(8, int(float32(width)/t.text(90))))
				if valueOr(o.YAxis, true) {
					if config.rtl {
						axes.Right = y.Ticks(4)
					} else {
						axes.Left = y.Ticks(4)
					}
				}
				if valueOr(o.XAxis, true) {
					axes.Bottom = s.chartTicks(p, o, count)
					if config.rtl {
						for i := range axes.Bottom {
							axes.Bottom[i].At = 1 - axes.Bottom[i].At
						}
					}
				}
				axes.Label = func(side plot.Side, tick plot.Tick) {
					label := ui.Text(c, tick.Label).NoWrap().FontSize(t.text(12)).TextColor(t.Muted)
					if config.rtl {
						label.Direction(ui.RightToLeft)
					}
					slot := YAxisSlot
					if side == plot.Bottom {
						slot = XAxisSlot
					}
					finish(t, ComponentChart, slot, label, PartStyle{}, o.Decoration)
					if side == plot.Bottom {
						label.SingleLine().MaxWidth(max(t.text(20), float32(width)/float32(count)-t.spacing(8))).Shrink(0)
						if tick.At < .05 {
							label.AlignSelf(ui.Start)
						}
						if tick.At > .95 {
							label.AlignSelf(ui.End)
						}
					}
				}
			}
			// Plot axes are physical. Prevent inherited RTL layout from reversing
			// the already mirrored tick positions and Y-axis side a second time.
			ui.Column(c.Key("plot-host")).MinWidth(0).Direction(ui.LeftToRight).Children(func() {
				plot.Plot(c.Key("plot"), axes, buildArea).Height(h).Gap(t.spacing(8)).MinWidth(0)
			})
		}
		if o.XLabel != "" && !chartRadial(o.Kind) && valueOr(o.XAxis, true) {
			finish(t, ComponentChart, XAxisSlot, ui.Text(c.Key("x-title"), o.XLabel).AlignSelf(ui.Center), PartStyle{Foreground: Set(t.Muted), FontSize: Set(t.text(12))}, o.Decoration)
		}
		// Reserve a stable line so pointer selection does not resize the chart or legend.
		finish(t, ComponentChart, Description, ui.Text(c, description).MinHeight(t.text(16)).Role(ui.RoleStatus), PartStyle{Foreground: Set(t.Text), FontSize: Set(t.text(12))}, o.Decoration)
		legendCount := len(s.series)
		if chartRadial(o.Kind) {
			legendCount = len(s.slices)
		}
		if valueOr(o.Legend, legendCount > 1) {
			s.chartLegend(c, t, o)
		}
	})
	return root
}

func (s *ChartState) chartTicks(p *chartPrepared, o ChartOptions, count int) []plot.Tick {
	var ticks []plot.Tick
	switch s.axis {
	case chartCategory:
		step := max(1, (len(s.categories)+max(1, count)-1)/max(1, count))
		for i := 0; i < len(s.categories); i += step {
			ticks = append(ticks, plot.Tick{At: p.band.At(i), Value: float64(i), Label: s.categories[i]})
		}
	case chartTime:
		x := p.time
		x.Format = o.FormatTime
		ticks = x.Ticks(count)
	default:
		x := p.x
		x.Format = o.FormatX
		ticks = x.Ticks(count)
	}
	// Scale tick counts are approximate; enforce a density limit for native labels.
	step := max(1, (len(ticks)+max(1, count)-1)/max(1, count))
	if step == 1 {
		return ticks
	}
	sparse := make([]plot.Tick, 0, count)
	for i := 0; i < len(ticks); i += step {
		sparse = append(sparse, ticks[i])
	}
	return sparse
}

func (s *ChartState) describePoint(c *ui.Context, p *chartPrepared, o ChartOptions, i int) string {
	if i < 0 || i >= len(s.x) {
		return ""
	}
	x := fmt.Sprintf("%g", s.x[i])
	if o.FormatX != nil {
		x = o.FormatX(s.x[i])
	}
	if s.axis == chartCategory {
		x = s.categories[i]
	}
	if s.axis == chartTime {
		x = s.times[i].Format("2006-01-02 15:04:05")
		if o.FormatTime != nil {
			x = o.FormatTime(s.times[i])
		}
	}
	parts := []string{x}
	for _, index := range p.indices {
		series := s.series[index]
		v := series.Values[i]
		value := fmt.Sprintf("%.3g", v)
		if o.FormatY != nil {
			value = o.FormatY(v)
		}
		if !chartFinite(v) {
			value = message(c, "missingValue")
		}
		parts = append(parts, series.Label+": "+value)
	}
	return strings.Join(parts, " · ")
}

func (s *ChartState) selectSlice(c *ui.Context, area ui.Element, p *chartPrepared, o ChartOptions) string {
	area.Focusable().FocusRing(true)
	wedges := p.scene.wedges
	index := -1
	for i, wedge := range wedges {
		if wedge.id == s.SelectedSlice {
			index = i
		}
	}
	before := index
	switch {
	case len(wedges) == 0:
		index = -1
	case area.Shortcut(0, ui.KeyRight), area.Shortcut(0, ui.KeyDown):
		index = min(len(wedges)-1, index+1)
	case area.Shortcut(0, ui.KeyLeft), area.Shortcut(0, ui.KeyUp):
		if index < 0 {
			index = len(wedges)
		}
		index = max(0, index-1)
	case area.Shortcut(0, ui.KeyHome):
		index = 0
	case area.Shortcut(0, ui.KeyEnd):
		index = len(wedges) - 1
	case area.Shortcut(0, ui.KeyEscape):
		index = -1
	}
	keyboard := before != index
	x, y, over := area.PointerPosition()
	if !keyboard && over && (!s.pieHovered || s.piePointer.X != x || s.piePointer.Y != y) {
		bounds := area.Bounds()
		arcs := make([]plot.Slice, len(wedges))
		for i, wedge := range wedges {
			arcs[i] = wedge.arc
		}
		outer := chartRadius(bounds)
		if at, ok := plot.SliceAt(arcs, x-bounds.W/2, y-bounds.H/2, outer*float32(p.scene.config.inner), outer); ok {
			index = at
		} else {
			index = -1
		}
	} else if !over && s.pieHovered && !keyboard {
		index = -1
	}
	s.pieHovered, s.piePointer = over, ui.Point{X: x, Y: y}
	if index < 0 || index >= len(wedges) {
		s.SelectedSlice = ""
		return ""
	}
	s.SelectedSlice = wedges[index].id
	v := s.slices[wedges[index].index]
	total := 0.0
	for _, wedge := range wedges {
		total += s.slices[wedge.index].Value
	}
	value := fmt.Sprintf("%.3g", v.Value)
	if o.FormatY != nil {
		value = o.FormatY(v.Value)
	}
	description := fmt.Sprintf("%s: %s (%.1f%%)", v.Label, value, v.Value/total*100)
	area.Value(description)
	if keyboard {
		c.Announce(description)
	}
	return description
}

func (s *ChartState) chartLegend(c *ui.Context, t Theme, o ChartOptions) {
	legend := ui.Row(c.Key("legend")).Wrap().Justify(ui.Center)
	finish(t, ComponentChart, List, legend, PartStyle{Gap: Set(t.spacing(16))}, o.Decoration)
	legend.Children(func() {
		item := func(id, label string, color Color, index int, visible bool, toggle func()) {
			var row ui.Element
			if valueOr(o.InteractiveLegend, true) {
				// The button owns its identity; no build value escapes to an action.
				row = ui.ButtonBase(c.Key("toggle-" + id)).Role(ui.RoleCheckBox).Checked(visible).Label(label).Gap(6).AlignItems(ui.Center).FocusRing(true).OnClick(toggle)
			} else {
				row = ui.Row(c.Key(id)).Gap(6).AlignItems(ui.Center)
			}
			foreground := t.Muted
			if !visible {
				foreground = t.Dimmed
			}
			finish(t, ComponentChart, Item, row, PartStyle{PaddingX: Set(t.spacing(6)), PaddingY: Set(t.spacing(4)), Radius: Set(t.radius(4)), Foreground: Set(foreground)}, o.Decoration)
			row.Children(func() {
				dot := ui.Box(c).Size(t.spacing(8), t.spacing(8)).Radius(t.radius(4)).Background(chartColorWithPalette(t, color, index, o.Colors))
				if !visible {
					dot.Opacity(.25)
				}
				finish(t, ComponentChart, Indicator, dot, PartStyle{}, o.Decoration)
				finish(t, ComponentChart, Label, ui.Text(c, label), PartStyle{Foreground: Set(foreground), FontSize: Set(t.text(12))}, o.Decoration)
			})
		}
		if chartRadial(o.Kind) {
			for i, v := range s.slices {
				visible := !s.hiddenSlices[v.ID]
				item(v.ID, v.Label, v.Color, i, visible, func() {
					s.SetSliceVisible(v.ID, !visible)
					if o.OnLegendChange != nil {
						o.OnLegendChange(v.ID, !visible)
					}
				})
			}
		} else {
			for i, v := range s.series {
				visible := !s.hidden[v.ID]
				item(v.ID, v.Label, v.Color, i, visible, func() {
					s.SetSeriesVisible(v.ID, !visible)
					if o.OnLegendChange != nil {
						o.OnLegendChange(v.ID, !visible)
					}
				})
			}
		}
	})
}
