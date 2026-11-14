package nui

import (
	"math"

	"github.com/egoist/mygo/ui"
)

type SplitterOptions struct {
	Label                       string
	Orientation                 Orientation
	Vertical, Percent, Disabled bool
	Min, Max                    Optional[float32]
	// KeyboardResizeBy uses the bound value's unit. Unset moves ten DIPs; zero locks arrow keys.
	KeyboardResizeBy Optional[float32]
	OnChange         func(float32)
	Decoration
}

type splitterLocal struct {
	vertical, rtl, initialized, ignoreDrag bool
}

// Splitter binds the first pane's size in DIPs, or a percentage when Percent is true.
// Disabled locks only the divider; disabling an ancestor also disables pane contents.
func Splitter(c *ui.Context, t Theme, value *float32, o SplitterOptions, first, second func()) ui.Element {
	if value == nil {
		panic("nui: Splitter requires value")
	}
	vert := vertical(o.Orientation) || o.Vertical
	rtl := !vert && c.Direction() == ui.RightToLeft
	lo, hi := float32(0), float32(1e7)
	if o.Percent {
		hi = 100
	}
	for _, limit := range []Optional[float32]{o.Min, o.Max, o.KeyboardResizeBy} {
		if limit.Set && (math.IsNaN(float64(limit.Value)) || math.IsInf(float64(limit.Value), 0)) {
			panic("nui: Splitter constraints and step must be finite")
		}
	}
	if o.Min.Set {
		lo = max(lo, o.Min.Value)
	}
	if o.Max.Set {
		hi = min(hi, o.Max.Value)
	}
	if hi < lo || (o.KeyboardResizeBy.Set && o.KeyboardResizeBy.Value < 0) {
		panic("nui: invalid Splitter constraints or step")
	}
	clamp := func(v float32) float32 { return max(lo, min(hi, float32(finite(float64(v))))) }
	*value = clamp(*value)
	root := ui.Themed(c, t.Native()).Background(ui.Transparent).AlignSelf(ui.Stretch).MinWidth(0).MinHeight(0)
	finish(t, ComponentSplitter, Root, root, PartStyle{}, o.Decoration)
	state := ui.Local(root, "splitter", func() splitterLocal { return splitterLocal{} })
	label := o.Label
	if label == "" {
		label = message(c, "divider")
	}
	root.Children(func() {
		host := ui.Row(c.Key("panes")).Fill().MinWidth(0).MinHeight(0).AlignItems(ui.Stretch)
		if vert {
			host.Column()
		}
		total := host.Bounds().W
		if vert {
			total = host.Bounds().H
		}
		gripSize, minPane := t.spacing(8), t.spacing(40)
		room := max(float32(0), total-1)
		toDIP := func(v float32) float32 {
			if o.Percent {
				return v * room / 100
			}
			return v
		}
		toValue := func(v float32) float32 {
			if o.Percent {
				if room <= 0 {
					return *value
				}
				return v * 100 / room
			}
			return v
		}
		fit := func(v float32) float32 {
			if total > 0 {
				v = min(v, room-minPane)
			}
			return max(v, minPane)
		}
		shown := toDIP(*value)
		if o.Percent || *value > minPane {
			shown = fit(shown)
		}
		step := toValue(t.spacing(10))
		if o.KeyboardResizeBy.Set {
			step = o.KeyboardResizeBy.Value
		}
		set := func(v float32) bool {
			v = clamp(toValue(fit(toDIP(v))))
			if v == *value {
				return false
			}
			*value = v
			return true
		}
		change := func(v float32) {
			if set(v) && o.OnChange != nil {
				o.OnChange(*value)
			}
		}
		length := func(e ui.Element, v float32) {
			if vert {
				e.Height(v)
			} else {
				e.Width(v)
			}
		}
		minLength := func(e ui.Element, v float32) {
			if vert {
				e.MinHeight(v)
			} else {
				e.MinWidth(v)
			}
		}
		host.Children(func() {
			firstPane := ui.Column(c.Key("first")).Shrink(1).Clip().MinWidth(0).MinHeight(0)
			if o.Percent {
				firstPane.Grow(*value / 100)
				minLength(firstPane, minPane)
			} else {
				length(firstPane, *value)
				minLength(firstPane, min(*value, minPane))
			}
			divider := ui.Box(c.Key("divider")).Shrink(0).Focusable().FocusRing(false).Role(ui.RoleSplitter).Label(label).Disabled(o.Disabled)
			length(divider, 1)
			divider.Range(float64(lo), float64(hi), float64(*value)).Step(float64(step))
			back, forth := ui.KeyLeft, ui.KeyRight
			if vert {
				back, forth = ui.KeyUp, ui.KeyDown
			} else if rtl {
				back, forth = forth, back
			}
			current := toValue(shown)
			divider.OnShortcut(0, back, func() {
				if step > 0 {
					change(current - step)
				}
			}).OnShortcut(0, forth, func() {
				if step > 0 {
					change(current + step)
				}
			})
			divider.OnShortcut(0, ui.KeyHome, func() { change(lo) }).OnShortcut(0, ui.KeyEnd, func() { change(hi) })
			secondPane := ui.Column(c.Key("second")).Grow(1).Shrink(1).MinWidth(0).MinHeight(0)
			minLength(secondPane, minPane)
			if o.Percent {
				firstPane.Grow(*value / 100)
				secondPane.Grow(1 - *value/100)
			} else {
				length(firstPane, *value)
			}
			finish(t, ComponentSplitter, Leading, firstPane, PartStyle{}, o.Decoration)
			if first != nil {
				firstPane.Children(first)
			}
			finish(t, ComponentSplitter, Trailing, secondPane, PartStyle{}, o.Decoration)
			secondPane.Children(func() {
				// Only the content clips. The hit area overlaps both panes while
				// the focusable divider reserves exactly one DIP in the layout.
				body := ui.Column(c.Key("content")).Fill().Clip().MinWidth(0).MinHeight(0)
				if second != nil {
					body.Children(second)
				}
				hit := ui.Box(c.Key("hit-area")).Role(ui.RoleNone).Disabled(o.Disabled)
				if vert {
					hit.Attach(ui.AnchorTop, ui.AnchorCenter).Top(-.5).WidthPercent(100).Height(gripSize).Cursor(ui.CursorResizeNS)
				} else {
					hit.Attach(ui.AnchorStart, ui.AnchorCenter).Start(-.5).HeightPercent(100).Width(gripSize).Cursor(ui.CursorResizeEW)
				}
				if hit.IsDisabled() {
					hit.Cursor(ui.CursorDefault)
				}
				finish(t, ComponentSplitter, SeparatorSlot, hit, PartStyle{}, o.Decoration)
				dx, dy, held := hit.Dragged()
				if state.initialized && (state.vertical != vert || state.rtl != rtl) {
					state.ignoreDrag = held
				}
				state.initialized, state.vertical, state.rtl = true, vert, rtl
				if state.ignoreDrag {
					state.ignoreDrag = held
					held = false
				}
				if held && !divider.IsDisabled() {
					divider.Focus()
					delta := dx
					if vert {
						delta = dy
					} else if rtl {
						delta = -dx
					}
					if delta != 0 && set(toValue(shown+delta)) {
						next := *value
						divider.SetChanged().OnChange(func() {
							if o.OnChange != nil {
								o.OnChange(next)
							}
						})
					}
				}
				lineColor := t.Border
				if !divider.IsDisabled() && (held || hit.Hovered() || divider.Focused()) {
					lineColor = t.Primary
				}
				hit.Children(func() {
					line := ui.Box(c.Key("line")).Attach(ui.AnchorCenter, ui.AnchorCenter).IgnoreInput(true).Role(ui.RoleNone)
					if vert {
						line.WidthPercent(100).Height(1)
					} else {
						line.HeightPercent(100).Width(1)
					}
					finishLine(t, ComponentSplitter, Indicator, line, PartStyle{Background: Set(lineColor)}, o.Decoration, !vert, false, 0)
				})
				focused := divider.FocusVisible()
				hit.DrawOver(func(p *ui.Painter, r ui.Rect) {
					if focused {
						p.FocusRing(r, [4]float32{})
					}
				})
			})
		})
	})
	return root
}
