package nui

import (
	"fmt"
	"math"
	"time"

	"github.com/egoist/mygo/ui"
)

type sliderInputMemory struct {
	handle   ui.Handle
	dragging bool
	changed  bool
	start    float64
}

var sliderKeys = []ui.Key{ui.KeyLeft, ui.KeyRight, ui.KeyUp, ui.KeyDown, ui.KeyPageUp, ui.KeyPageDown, ui.KeyHome, ui.KeyEnd}

func sliderKeyValue(key ui.Key, shift bool, value, lo, hi, step float64, vert, inverted, rtl bool) float64 {
	if key == ui.KeyHome {
		return lo
	}
	if key == ui.KeyEnd {
		return hi
	}
	delta := step
	if shift || key == ui.KeyPageUp || key == ui.KeyPageDown {
		delta *= 10
	}
	back := false
	switch key {
	case ui.KeyUp, ui.KeyPageUp:
		back = vert && inverted
	case ui.KeyDown, ui.KeyPageDown:
		back = !vert || !inverted
	case ui.KeyRight:
		back = !vert && (rtl != inverted)
	case ui.KeyLeft:
		back = vert || (rtl == inverted)
	}
	if back {
		delta = -delta
	}
	return value + delta
}

type rangeSliderMemory struct {
	root     ui.Handle
	thumbs   [2]ui.Handle
	dragging int
	changed  bool
	start    [2]float64
}

// RangeSlider shares a pointer track and keeps two independent keyboard and
// accessibility targets. Public Range/Step semantics expose actual model values.
func RangeSlider(c *ui.Context, t Theme, low, high *float64, o RangeSliderOptions) ui.Element {
	if low == nil || high == nil {
		panic("nui: RangeSlider requires two values")
	}
	vert := vertical(o.Orientation)
	if o.Min == 0 && o.Max == 0 {
		o.Max = 100
	}
	lo, hi, step := numberBounds(InputNumberOptions{Min: Set(o.Min), Max: Set(o.Max), Step: o.Step})
	gap := float64(max(0, o.MinStepsBetweenThumbs)) * step
	if hi <= lo || gap > hi-lo || math.IsInf(gap, 0) {
		panic("nui: invalid range slider bounds")
	}
	*low = max(lo, min(hi-gap, finite(*low)))
	*high = max(*low+gap, min(hi, finite(*high)))
	config := selected(defaults(t, ComponentSlider, ComponentTheme{Color: Primary, Size: MD}), ComponentTheme{Color: o.Color, Size: inheritedSize(t, o.Size)})
	thumb := t.text([5]float32{12, 14, 16, 18, 20}[sizeIndex(config.Size)])
	thickness := t.spacing([5]float32{6, 7, 8, 9, 10}[sizeIndex(config.Size)])
	root := ui.Box(c).MinWidth(100).Height(max(24, thumb)).Role(ui.RoleGroup).Label(o.Label).Disabled(o.Disabled || t.field.Disabled)
	if vert {
		root.MinWidth(0).Width(max(24, thumb)).Height(160)
	}
	s := ui.Local(root, "range", func() rangeSliderMemory { return rangeSliderMemory{dragging: -1} })
	if o.Disabled || t.field.Disabled {
		s.dragging = -1
	}
	root.Bind(&s.root)
	if s.changed {
		root.SetChanged()
		s.changed = false
	}
	reverse := vert || c.Direction() == ui.RightToLeft
	if o.Inverted {
		reverse = !reverse
	}
	notify := func() {
		s.changed = true
		if o.OnChange != nil {
			o.OnChange(*low, *high)
		}
	}
	set := func(k int, v float64) bool {
		v = max(lo, min(hi, lo+math.Round((v-lo)/step)*step))
		a, b := v, *high
		if k == 1 {
			a, b = *low, v
		}
		if a > b {
			a, b = b, a
		}
		if b-a+1e-10 < gap || (a == *low && b == *high) {
			return false
		}
		*low, *high = a, b
		next := k
		if a != b {
			if v == a {
				next = 0
			} else {
				next = 1
			}
		}
		if s.dragging >= 0 {
			s.dragging = next
		}
		s.thumbs[next].Focus(c)
		notify()
		return true
	}
	commit := func() {
		if o.OnCommit != nil {
			o.OnCommit(*low, *high)
		}
	}
	pointerInput := func(ev ui.InputEvent) bool {
		if o.Disabled || t.field.Disabled {
			return false
		}
		if ev.Kind == ui.InputPointerUp {
			was := s.dragging >= 0
			s.dragging = -1
			if was && s.start != [2]float64{*low, *high} {
				commit()
			}
			return was
		}
		if ev.Button != 0 || (ev.Kind != ui.InputPointerDown && ev.Kind != ui.InputPointerMove) {
			return false
		}
		if ev.Kind == ui.InputPointerMove && s.dragging < 0 {
			return false
		}
		r := s.root.Bounds(c)
		axis, length := ev.X, r.W
		if vert {
			axis, length = ev.Y, r.H
		}
		fraction := max(float32(0), min(float32(1), (axis-thumb/2)/max(float32(1), length-thumb)))
		if reverse {
			fraction = 1 - fraction
		}
		v := lo + float64(fraction)*(hi-lo)
		if ev.Kind == ui.InputPointerDown {
			s.start = [2]float64{*low, *high}
			s.dragging = 0
			if math.Abs(v-*high) < math.Abs(v-*low) || v > *high {
				s.dragging = 1
			}
			s.thumbs[s.dragging].Focus(c)
		}
		set(s.dragging, v)
		return true
	}
	root.HandleInput(pointerInput)
	var focused [2]bool
	root.Children(func() {
		for k, v := range []*float64{low, high} {
			from, to := lo, *high-gap
			if k == 1 {
				from, to = *low+gap, hi
			}
			label := namedMessage(c, "rangeMinimum", o.Label)
			if k == 1 {
				label = namedMessage(c, "rangeMaximum", o.Label)
			}
			e := ui.Box(c.Key(k)).Focusable().Role(ui.RoleSlider).Range(from, to, *v)
			e.Absolute().Left(0).Right(0).Top(0).Bottom(0).Step(step).Label(label).FocusRing(false).Bind(&s.thumbs[k])
			if k == 0 {
				// A FormField label focuses the first thumb, which also carries the
				// field's description and error; each thumb keeps its own name.
				bindField(c, e, t)
				e.Label(label)
			}
			for _, key := range sliderKeys {
				for _, mods := range []ui.Modifiers{0, ui.Shift} {
					e.OnShortcut(mods, key, func() {
						target := k
						if key == ui.KeyHome {
							target = 0
						}
						if key == ui.KeyEnd {
							target = 1
						}
						if set(target, sliderKeyValue(key, mods == ui.Shift, *v, lo, hi, step, vert, o.Inverted, c.Direction() == ui.RightToLeft)) {
							commit()
						}
					})
				}
			}
			focused[k] = e.FocusVisible()
		}
		// A distinct pointer surface prevents both native thumb bindings from
		// interpreting the same press with their individual value ranges.
		ui.Box(c.Key("pointer")).Absolute().Left(0).Right(0).Top(0).Bottom(0).HandleInput(pointerInput)
	})
	f0, f1 := float32((*low-lo)/(hi-lo)), float32((*high-lo)/(hi-lo))
	if reverse {
		f0, f1 = 1-f0, 1-f1
	}
	root.Draw(func(p *ui.Painter, r ui.Rect) {
		track := ui.Rect{X: r.X + thumb/2, Y: r.Y + (r.H-thickness)/2, W: max(0, r.W-thumb), H: thickness}
		if vert {
			track = ui.Rect{X: r.X + (r.W-thickness)/2, Y: r.Y + thumb/2, W: thickness, H: max(0, r.H-thumb)}
		}
		p.Fill(track, t.Accented, 999)
		fill := track
		if vert {
			fill.Y += track.H * min(f0, f1)
			fill.H *= float32(math.Abs(float64(f1 - f0)))
		} else {
			fill.X += track.W * min(f0, f1)
			fill.W *= float32(math.Abs(float64(f1 - f0)))
		}
		p.Fill(fill, t.color(config.Color), 999)
		for k, f := range []float32{f0, f1} {
			dot := ui.Rect{X: track.X + track.W*f - thumb/2, Y: r.Y + (r.H-thumb)/2, W: thumb, H: thumb}
			if vert {
				dot.X = r.X + (r.W-thumb)/2
				dot.Y = track.Y + track.H*f - thumb/2
			}
			sliderThumb(p, dot, t.color(config.Color), t.Background, t.focusColor(config.Color), focused[k])
		}
	})
	if o.Tooltip {
		root.Children(func() {
			for k, v := range []float64{*low, *high} {
				f := []float32{f0, f1}[k]
				c.Key(struct{ Tip int }{k})
				sliderValueTooltip(c, t, s.thumbs[k].Focused(c), fmt.Sprint(v), f, vert, thumb)
			}
		})
	}
	finish(t, ComponentSlider, Root, root, PartStyle{}, o.Decoration)
	return root
}

// sliderThumb paints Nuxt's thumb: the background color within a 2 DIP ring
// of the slider's color drawn outside it, and while focused, a 3 DIP outline
// 2 DIPs out, which the ring fills.
func sliderThumb(p *ui.Painter, dot ui.Rect, color, thumb, focusColor ui.Color, focus bool) {
	grow := func(r ui.Rect, by float32) ui.Rect {
		return ui.Rect{X: r.X - by, Y: r.Y - by, W: r.W + 2*by, H: r.H + 2*by}
	}
	if focus {
		p.Fill(grow(dot, 5), focusColor, 999)
	}
	p.Fill(grow(dot, 2), color, 999)
	p.Fill(dot, thumb, 999)
}

func sliderValueTooltip(c *ui.Context, t Theme, focused bool, text string, f float32, vert bool, thumb float32) {
	anchor := ui.Box(c).Absolute().Size(thumb, thumb).PassThrough()
	if vert {
		anchor.LeftPercent(50).TopPercent(f*100).Translate(-thumb/2, -thumb*f)
	} else {
		anchor.LeftPercent(f*100).TopPercent(50).Translate(-thumb*f, -thumb/2)
	}
	_, _, over := anchor.PointerPosition()
	state := ui.Local(anchor, "tip", func() tooltipMemory { return tooltipMemory{} })
	if !over && !focused {
		state.suppressed = false
		return
	}
	if state.suppressed {
		return
	}
	ui.Overlay(c, func() {
		tip := ui.Text(c, text).Role(ui.RoleTooltip).AttachTo(anchor, ui.AnchorTop, ui.AnchorBottom).Margin(6, 0).Padding(4, 8).Background(t.Background).TextColor(t.Text).Radius(t.radius(4)).IgnoreInput(true)
		if tip.OverlayShortcut(0, ui.KeyEscape) {
			state.suppressed = true
			c.After(time.Nanosecond)
		}
	})
}
