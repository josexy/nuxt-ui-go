package nui

import (
	"math"
	"time"

	"github.com/egoist/mygo/ui"
)

type TableLoadingAnimation uint8

const (
	TableLoadingCarousel TableLoadingAnimation = iota
	TableLoadingCarouselInverse
	TableLoadingSwing
	TableLoadingElastic
)

type tableEvents struct {
	sort        []ui.SortOrder
	selection   map[string]struct{}
	initialized bool
}

func tableSelection(selection *ui.Selection[string]) map[string]struct{} {
	result := make(map[string]struct{}, selection.Len())
	for id := range selection.All() {
		result[id] = struct{}{}
	}
	return result
}

func finishTableCell(t Theme, slot Slot, e ui.Element, base PartStyle, table, column Decoration) {
	base = mergeStyle(base, t.components[ComponentTable].UI[slot])
	base = mergeStyle(base, table.UI[slot])
	base = mergeStyle(base, column.UI[slot])
	base.apply(e)
	if table.Style != nil {
		table.Style(slot, e)
	}
	if column.Style != nil {
		column.Style(slot, e)
	}
}

func tableLoading(c *ui.Context, t Theme, y float32, o TableOptions) {
	config := selected(defaults(t, ComponentTable, ComponentTheme{Color: Primary}), ComponentTheme{Color: o.LoadingColor})
	style := PartStyle{Foreground: Set(inputColor(t, config.Color))}
	style = mergeStyle(style, t.components[ComponentTable].UI[Indicator])
	style = mergeStyle(style, o.UI[Indicator])
	line := ui.Box(c).Absolute().Left(0).Right(0).Top(y).Height(t.spacing(1)).Clip().PassThrough().Role(ui.RoleProgress).Label(message(c, "loading")).Range(0, 100, -1).Value(message(c, "inProgress"))
	style.apply(line)
	animated, rtl := t.motion(c), c.Direction() == ui.RightToLeft
	line.Draw(func(p *ui.Painter, r ui.Rect) {
		start, width := float32(0), float32(1)
		if animated {
			p.AnimationFrame()
			phase := float32(phase(p.Now(), 2*time.Second))
			width = .5
			switch o.LoadingAnimation {
			case TableLoadingCarousel, TableLoadingCarouselInverse:
				start = 1.5*phase - .5
				if o.LoadingAnimation == TableLoadingCarouselInverse {
					start = .5 - start
				}
				if rtl {
					start = 1 - start - width
				}
			case TableLoadingSwing, TableLoadingElastic:
				x := float32(1 - math.Abs(float64(2*phase-1)))
				x = x * x * (3 - 2*x)
				if o.LoadingAnimation == TableLoadingSwing {
					start = -.125 + .75*x
				} else {
					width = .5 + .5*x
					start = (1 - width) / 2
				}
			}
		}
		fill := r
		fill.X += start * r.W
		fill.W *= width
		p.Clip(r, 0, func() { p.Fill(fill, style.Foreground.Value, 0) })
	})
	if o.Style != nil {
		o.Style(Indicator, line)
	}
}
