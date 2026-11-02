package nui

import (
	"math"
	"slices"
	"time"

	"github.com/egoist/mygo/ui"
	plot "github.com/egoist/mygo/ui/chart"
)

type chartLayer struct {
	scene *chartScene
	alpha float32
}

type chartAnimation struct {
	target   *chartScene
	from     []chartLayer
	start    time.Time
	duration time.Duration
	tween    bool
}

func (a *chartAnimation) set(target *chartScene, now time.Time, duration time.Duration, enabled bool) {
	if !enabled || a.target == nil {
		a.target, a.from = target, nil
		return
	}
	if a.target == target {
		if now.Sub(a.start) >= a.duration {
			a.from = nil
		}
		return
	}
	// Resizing changes sampling resolution, not the data. Layout handles it directly.
	oldConfig, newConfig := a.target.config, target.config
	oldConfig.width, newConfig.width = 0, 0
	if a.target.revision == target.revision && oldConfig == newConfig {
		if _, active := a.layers(now); !active {
			a.target, a.from = target, nil
			return
		}
	}
	current, _ := a.layers(now)
	// Retarget from the displayed frame. Keep at most three outgoing layers even
	// when the user switches types repeatedly before a transition finishes.
	for len(current) > 3 {
		weakest := 0
		for i := range current {
			if current[i].alpha < current[weakest].alpha {
				weakest = i
			}
		}
		current = slices.Delete(current, weakest, weakest+1)
	}
	a.tween = len(current) == 1 && current[0].alpha > .999 && chartCompatible(current[0].scene, target)
	a.target, a.from, a.start, a.duration = target, current, now, duration
}

func (a chartAnimation) layers(now time.Time) ([]chartLayer, bool) {
	if a.target == nil {
		return nil, false
	}
	if len(a.from) == 0 || a.duration <= 0 || now.Sub(a.start) >= a.duration {
		return []chartLayer{{a.target, 1}}, false
	}
	u := float32(max(0, float64(now.Sub(a.start))/float64(a.duration)))
	u = ui.EaseOut(u)
	if a.tween {
		return []chartLayer{{chartInterpolate(a.from[0].scene, a.target, u), 1}}, true
	}
	layers := make([]chartLayer, 0, len(a.from)+1)
	for _, layer := range a.from {
		layer.alpha *= 1 - u
		if layer.alpha > .001 {
			layers = append(layers, layer)
		}
	}
	if u > .001 {
		layers = append(layers, chartLayer{a.target, u})
	}
	return layers, true
}

func chartRadial(kind ChartKind) bool { return kind == ChartPie || kind == ChartDonut }

func chartCompatible(a, b *chartScene) bool {
	if a.empty || b.empty {
		return false
	}
	if chartRadial(a.config.kind) && chartRadial(b.config.kind) {
		if len(a.wedges) != len(b.wedges) {
			return false
		}
		for i := range a.wedges {
			if a.wedges[i].id != b.wedges[i].id {
				return false
			}
		}
		return true
	}
	if a.config.kind != b.config.kind || a.config.curve != b.config.curve || a.config.stack != b.config.stack || len(a.marks) != len(b.marks) {
		return false
	}
	for i, m := range a.marks {
		n := b.marks[i]
		if m.id != n.id || len(m.x) != len(n.x) {
			return false
		}
		for j := range m.x {
			if math.IsNaN(float64(m.x[j])) != math.IsNaN(float64(n.x[j])) || math.IsNaN(float64(m.y[j])) != math.IsNaN(float64(n.y[j])) || math.IsNaN(float64(m.low[j])) != math.IsNaN(float64(n.low[j])) {
				return false
			}
		}
	}
	return true
}

func chartInterpolate(a, b *chartScene, u float32) *chartScene {
	if u <= 0 {
		return a
	}
	if u >= 1 {
		return b
	}
	out := *b
	lerp := func(x, y float32) float32 { return x + (y-x)*u }
	blend := func(x, y []float32) []float32 {
		values := make([]float32, len(y))
		for i := range values {
			values[i] = lerp(x[i], y[i])
		}
		return values
	}
	if chartRadial(b.config.kind) {
		out.config.inner = a.config.inner + (b.config.inner-a.config.inner)*float64(u)
		out.wedges = slices.Clone(b.wedges)
		for i := range out.wedges {
			x, y := a.wedges[i].arc, b.wedges[i].arc
			out.wedges[i].arc = plot.Slice{Start: x.Start + (y.Start-x.Start)*float64(u), End: x.End + (y.End-x.End)*float64(u), Pad: x.Pad + (y.Pad-x.Pad)*float64(u)}
		}
	} else {
		out.marks = slices.Clone(b.marks)
		for i := range out.marks {
			x, y := a.marks[i], b.marks[i]
			out.marks[i].x, out.marks[i].y = blend(x.x, y.x), blend(x.y, y.y)
			out.marks[i].low = blend(x.low, y.low)
			out.marks[i].left, out.marks[i].right = blend(x.left, y.left), blend(x.right, y.right)
		}
	}
	return &out
}

func chartRadius(r ui.Rect) float32 { return max(0, min(r.W, r.H)/2-10) }

func drawChartScene(p *ui.Painter, r ui.Rect, t Theme, scene *chartScene, alpha float32, selectedSlice string, lineWidth, pointSize float32) {
	if scene.empty {
		return
	}
	if chartRadial(scene.config.kind) {
		outer := chartRadius(r)
		inner := outer * float32(scene.config.inner)
		for _, wedge := range scene.wedges {
			radius := outer
			if wedge.id == selectedSlice {
				radius += 4
			}
			p.FillPath(plot.Sector(r.X+r.W/2, r.Y+r.H/2, inner, radius, wedge.arc), chartColor(t, wedge.color, wedge.index).Alpha(alpha))
		}
		return
	}
	for _, mark := range scene.marks {
		color := chartColor(t, mark.color, mark.index).Alpha(alpha)
		switch scene.config.kind {
		case ChartBar:
			for i, x := range mark.x {
				if !chartFinite(float64(x)) || !chartFinite(float64(mark.y[i])) || !chartFinite(float64(mark.low[i])) {
					continue
				}
				top, bottom := plot.Y(r, max(mark.y[i], mark.low[i])), plot.Y(r, min(mark.y[i], mark.low[i]))
				p.Fill(ui.Rect{X: plot.X(r, mark.left[i]), Y: top, W: (mark.right[i] - mark.left[i]) * r.W, H: bottom - top}, color, 2)
			}
		case ChartScatter:
			for i, x := range mark.x {
				if !chartFinite(float64(x)) || !chartFinite(float64(mark.y[i])) {
					continue
				}
				p.Fill(ui.Rect{X: plot.X(r, x) - pointSize/2, Y: plot.Y(r, mark.y[i]) - pointSize/2, W: pointSize, H: pointSize}, color, pointSize/2)
			}
		default:
			if scene.config.kind == ChartArea {
				opacity := float32(.16)
				if scene.config.stack != ChartStackNone {
					opacity = .55
				}
				p.FillPath(plot.Area(r, mark.x, mark.low, mark.y, plot.Curve(scene.config.curve)), color.Alpha(opacity))
			}
			if lineWidth > 0 {
				p.StrokePath(plot.Line(r, mark.x, mark.y, plot.Curve(scene.config.curve)), lineWidth, color)
			}
		}
	}
}

// chartPalettes are Nuxt's --ui-chart-2..6 defaults: blue, violet, amber, rose and cyan.
var chartPalettes = [5]Palette{BluePalette(), VioletPalette(), AmberPalette(), RosePalette(), CyanPalette()}

// chartColor follows Nuxt's six distinct series colors, starting with primary;
// the fixed hues use shade 500, or 400 in dark themes.
func chartColor(t Theme, color Color, index int) ui.Color {
	if color != ColorDefault {
		return t.color(color)
	}
	i := index % 6
	if i == 0 {
		return t.Primary
	}
	shade := 500
	if t.Dark {
		shade = 400
	}
	return chartPalettes[i-1].Shade(shade)
}
