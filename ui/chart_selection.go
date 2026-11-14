package nui

import (
	"github.com/egoist/mygo/ui"
	plot "github.com/egoist/mygo/ui/chart"
)

type chartPointerMemory struct {
	over bool
	x    float32
}

// Keep X data in logical order while mapping physical arrows and pointer positions.
func (s *ChartState) selectChartPoint(c *ui.Context, area ui.Element, xs []float32, rtl bool, describe func(int) string) {
	area.Focusable()
	memory := ui.Local(area, "nui-chart-pointer", func() chartPointerMemory { return chartPointerMemory{} })
	before, n := s.Selected, len(xs)
	if s.Selected < -1 || s.Selected >= n {
		s.Selected = -1
	}
	keyboard := false
	move := func(next int) { s.Selected = max(0, min(n-1, next)); keyboard = true }
	if n > 0 {
		switch {
		case area.Shortcut(0, ui.KeyHome):
			move(0)
		case area.Shortcut(0, ui.KeyEnd):
			move(n - 1)
		case area.Shortcut(0, ui.KeyRight):
			if s.Selected < 0 {
				move(0)
			} else if rtl {
				move(s.Selected - 1)
			} else {
				move(s.Selected + 1)
			}
		case area.Shortcut(0, ui.KeyLeft):
			if s.Selected < 0 {
				move(n - 1)
			} else if rtl {
				move(s.Selected + 1)
			} else {
				move(s.Selected - 1)
			}
		case area.Shortcut(0, ui.KeyEscape):
			s.Selected = -1
			keyboard = true
		}
	}
	x, _, over := area.PointerPosition()
	if over {
		if width := area.Bounds().W; width > 0 && n > 0 && !keyboard && (!memory.over || x != memory.x) {
			position := x / width
			if rtl {
				position = 1 - position
			}
			s.Selected = plot.Nearest(xs, position)
		}
	} else if memory.over && !keyboard {
		s.Selected = -1
	}
	memory.over, memory.x = over, x
	if s.Selected >= 0 {
		value := describe(s.Selected)
		area.Value(value)
		if keyboard && before != s.Selected {
			c.Announce(value)
		}
	}
}
