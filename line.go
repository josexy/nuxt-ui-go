package nui

import (
	"math"

	"github.com/egoist/mygo/ui"
)

// lineMaterial keeps layout in DIPs while quantizing stroke width separately
// from position. A material also lets a final Style Background/Material replace
// the entire stroke, instead of painting a second background underneath it.
type lineMaterial struct {
	color              ui.Color
	vertical, hairline bool
	pattern            SeparatorType
	edge               int // -1: start edge, 0: center, 1: end edge (physical coordinates).
}

func pixelRound(v float32) float32 { return float32(math.Round(float64(v))) }

func lineRect(r ui.Rect, scale float32, vertical, hairline bool, edge int) ui.Rect {
	position, width := r.Y, r.H
	if vertical {
		position, width = r.X, r.W
	}
	if width <= 0 {
		return ui.Rect{}
	}
	pixels := max(float32(1), pixelRound(width*scale))
	if hairline {
		pixels = 1
	}
	start := pixelRound((position+width/2)*scale - pixels/2)
	if edge < 0 {
		start = pixelRound(position * scale)
	}
	if edge > 0 {
		start = pixelRound((position+width)*scale) - pixels
	}
	if vertical {
		r.X, r.W = start/scale, pixels/scale
	} else {
		r.Y, r.H = start/scale, pixels/scale
	}
	return r
}

func (m lineMaterial) PaintMaterial(p *ui.Painter, r ui.Rect, radii [4]float32) {
	r = lineRect(r, p.Scale(), m.vertical, m.hairline, m.edge)
	if r.W <= 0 || r.H <= 0 {
		return
	}
	if m.pattern == SeparatorSolid {
		p.FillRadii(r, radii, m.color)
		return
	}
	// Integer-pixel dashes and gaps never alternate length as the line moves.
	scale := p.Scale()
	start, length, width := r.X, r.W, r.H
	if m.vertical {
		start, length, width = r.Y, r.H, r.W
	}
	end := pixelRound((start + length) * scale)
	start = pixelRound(start * scale)
	width = pixelRound(width * scale)
	dash, radius := 3*width, float32(0)
	if m.pattern == SeparatorDotted {
		dash, radius = width, width/(2*scale)
	}
	for at := start; at < end; at += dash + 2*width {
		segment := ui.Rect{X: at / scale, Y: r.Y, W: min(dash, end-at) / scale, H: r.H}
		if m.vertical {
			segment = ui.Rect{X: r.X, Y: at / scale, W: r.W, H: min(dash, end-at) / scale}
		}
		p.Fill(segment, m.color, radius)
	}
}

func styleLine(e ui.Element, style PartStyle, vertical, hairline bool, edge int) {
	style.apply(e)
	e.Material(lineMaterial{color: style.Background.Value, vertical: vertical, hairline: hairline, edge: edge})
}

func finishLine(t Theme, id Component, slot Slot, e ui.Element, base PartStyle, d Decoration, vertical, hairline bool, edge int) {
	base = mergeStyle(mergeStyle(base, t.components[id].UI[slot]), d.UI[slot])
	styleLine(e, base, vertical, hairline, edge)
	if d.Style != nil {
		d.Style(slot, e)
	}
}
