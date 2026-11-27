package nui

import (
	"github.com/egoist/mygo/ui"
)

// TableLoadingAnimation is the LoadingAnimation of Table's loading bar.
type TableLoadingAnimation = LoadingAnimation

// The Table names of the LoadingAnimation presets.
const (
	TableLoadingCarousel        = LoadingCarousel
	TableLoadingCarouselInverse = LoadingCarouselInverse
	TableLoadingSwing           = LoadingSwing
	TableLoadingElastic         = LoadingElastic
)

type tableEvents struct {
	sort        []ui.SortOrder
	initialized bool
}

// MyGo does not expose its header divider as a part. Cover its full
// one-DIP rectangle before painting the replacement; simply drawing a hairline
// on top leaves a second native-colored pixel at fractional scales.
func tableHeaderDivider(c *ui.Context, t Theme, table ui.Element, headerHeight float32, d Decoration) {
	background := t.Background
	for _, slot := range []Slot{Root, Body, Panel} {
		style := mergeStyle(t.components[ComponentTable].UI[slot], d.UI[slot])
		if style.Background.Set && style.Background.Value.A == 255 {
			background = style.Background.Value
		}
	}
	table.Children(func() {
		line := ui.Box(c.Key("header-divider")).Absolute().Top(headerHeight).Left(0).Right(0).Height(1).PassThrough()
		style := mergeStyle(PartStyle{Background: Set(t.BorderAccented)}, t.components[ComponentTable].UI[SeparatorSlot])
		style = mergeStyle(style, d.UI[SeparatorSlot])
		style.apply(line)
		line.Material(tableDividerMaterial{background: background, color: style.Background.Value})
		if d.Style != nil {
			d.Style(SeparatorSlot, line)
		}
	})
}

type tableDividerMaterial struct{ background, color ui.Color }

func (m tableDividerMaterial) PaintMaterial(p *ui.Painter, r ui.Rect, _ [4]float32) {
	p.Fill(r, m.background, 0)
	p.Fill(lineRect(r, p.Scale(), false, true, -1), m.color, 0)
}

// tableCellStyle is a column's cell styling for one build, merged once rather
// than for every visible cell: global, then table, then column overrides.
type tableCellStyle struct {
	item, label PartStyle
	// direct puts a Value column's text straight into the native cell, which
	// aligns it as the decorated Item row would. Label styling still applies
	// to that text; Item styling, Style callbacks and custom cells need the
	// decorated container.
	direct bool
}

func newTableCellStyle[T any](t Theme, table Decoration, column TableColumn[T]) tableCellStyle {
	global := t.components[ComponentTable].UI
	s := tableCellStyle{
		item:  mergeStyle(mergeStyle(mergeStyle(PartStyle{Gap: Set(t.spacing(6))}, global[Item]), table.UI[Item]), column.UI[Item]),
		label: mergeStyle(mergeStyle(mergeStyle(PartStyle{}, global[Label]), table.UI[Label]), column.UI[Label]),
	}
	s.direct = column.Value != nil && column.Cell == nil && column.BuildCell == nil &&
		table.Style == nil && column.Style == nil &&
		global[Item] == (PartStyle{}) && table.UI[Item] == (PartStyle{}) && column.UI[Item] == (PartStyle{})
	return s
}

func finishTableCell(slot Slot, e ui.Element, style PartStyle, table, column Decoration) {
	style.apply(e)
	if table.Style != nil {
		table.Style(slot, e)
	}
	if column.Style != nil {
		column.Style(slot, e)
	}
}

func tableLoading[K TableKey](c *ui.Context, t Theme, y float32, o KeyedTableOptions[K]) {
	config := selected(defaults(t, ComponentTable, ComponentTheme{Color: Primary}), ComponentTheme{Color: o.LoadingColor})
	style := PartStyle{Foreground: Set(inputColor(t, config.Color))}
	style = mergeStyle(style, t.components[ComponentTable].UI[Indicator])
	style = mergeStyle(style, o.UI[Indicator])
	line := ui.Box(c).Absolute().Left(0).Right(0).Top(y).Height(t.spacing(1)).Clip().PassThrough().Role(ui.RoleProgress).Label(message(c, "loading")).Range(0, 100, -1).Value(message(c, "inProgress"))
	style.apply(line)
	// The bar travels at Full, pulses across the header under reduced
	// motion, as Nuxt's motion-reduce:animate-pulse, and stands still at Off.
	moving, pulsing, rtl := t.moves(c), t.motionLevel(c) == motionReduced, c.Direction() == ui.RightToLeft
	pulse := t.Motion.Pulse
	line.Draw(func(p *ui.Painter, r ui.Rect) {
		start, width, color := float32(0), float32(1), style.Foreground.Value
		switch {
		case moving:
			p.AnimationFrame()
			start, width = loadingSpan(o.LoadingAnimation, float32(phase(p.Now(), loadingCycle)), rtl)
		case pulsing:
			color = color.Alpha(pulseAlpha(p, pulse))
		}
		fill := r
		fill.X += start * r.W
		fill.W *= width
		p.Clip(r, 0, func() { p.Fill(fill, color, 0) })
	})
	if o.Style != nil {
		o.Style(Indicator, line)
	}
}
