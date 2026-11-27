package nui

import (
	"math"

	"github.com/egoist/mygo/ui"
)

type inputMenuTags struct {
	list                            ui.ListState
	labels                          map[string]string
	dataRevision, selectionRevision uint64
	pending                         string
	measuredWidth                   float32
}

func (s *SearchState) updateTagLabels() {
	tags := &s.memory.tags
	if tags.dataRevision == s.source.revision && tags.selectionRevision == s.SelectedValues.Revision() {
		return
	}
	labels := make(map[string]string, s.SelectedValues.Len())
	for _, id := range s.SelectedValues.values {
		label, known := tags.labels[id]
		if at, exists := s.source.metadata.index[id]; exists {
			label = s.source.items[at].Label
		} else if !known {
			label = id
		}
		labels[id] = label
	}
	tags.labels = labels
	tags.dataRevision, tags.selectionRevision = s.source.revision, s.SelectedValues.Revision()
	if s.SelectedValues.Len() == 0 {
		tags.labels, tags.pending = nil, ""
	}
}

// Only a bounded number of short-selection labels need measurement. Longer
// selections always fill the viewport; the native List measures visible rows.
func (s *SearchState) tagViewportHeight(c *ui.Context, t Theme, size Size) float32 {
	maximum := t.text(96)
	i := sizeIndex(size)
	font := t.text([5]float32{10, 10, 12, 12, 14}[i])
	line := t.text([5]float32{12, 12, 16, 16, 20}[i])
	minimum := line + t.spacing(6)
	if float32(s.SelectedValues.Len())*minimum >= maximum {
		return maximum
	}
	width := s.memory.tags.list.Bounds(c).W
	if width <= 0 {
		width, _ = c.Size()
	}
	width = max(t.text(12), width-t.spacing(14)-t.text(16))
	span := ui.Span{Text: "M", Font: t.Font, Size: font, Weight: 500}
	_, baseline := c.MeasureText(0, span)
	height := float32(0)
	for _, id := range s.SelectedValues.values {
		span.Text = s.memory.tags.labels[id]
		_, measured := c.MeasureText(width, span)
		lines := max(1, float32(math.Round(float64(measured/max(1, baseline)))))
		height += lines*line + t.spacing(6)
		if height >= maximum {
			return maximum
		}
	}
	return height
}

func inputMenuSelectedTags(c *ui.Context, t Theme, state *SearchState, o InputMenuOptions, size Size) {
	tags, selected := &state.memory.tags, &state.SelectedValues
	if selected.Len() == 0 {
		return
	}
	if tags.pending != "" && !selected.Has(tags.pending) {
		tags.pending = ""
		state.memory.focusRequested = true
	}
	request := func(index int) {
		if index < 0 || index >= selected.Len() {
			return
		}
		tags.pending = selected.values[index]
		tags.list.ScrollIntoView(index)
	}
	back, forth := ui.KeyLeft, ui.KeyRight
	if c.Direction() == ui.RightToLeft {
		back, forth = forth, back
	}
	list := ui.List(c.Key("selected-tags"), &tags.list, selected.Len()).ItemKey(func(i int) any { return selected.values[i] })
	height := state.tagViewportHeight(c, t, size)
	list.FillWidth().Grow(0).Basis(height).Shrink(0).MinHeight(0).Height(height).KeepFocus().Role(ui.RoleList).Label(namedMessage(c, "selectedOptionsNamed", o.Label))
	list.OnUnmount(func() { tags.pending = "" })
	if o.ReadOnly || list.IsDisabled() {
		tags.pending = ""
	}
	services := c.Services()
	list.Draw(func(_ *ui.Painter, r ui.Rect) {
		if tags.measuredWidth != r.W {
			tags.measuredWidth = r.W
			services.Invalidate()
		}
	})
	list.Rows(func(row ui.ListRow) {
		id := selected.values[row.Index]
		label := tags.labels[id]
		remove := func() {
			if selected.Remove(id) {
				tags.pending = ""
				state.memory.focusRequested = true
				if o.OnValuesChange != nil {
					o.OnValuesChange()
				}
			}
		}
		wrapper := ui.Column(row.Context).FillWidth().MinWidth(0).PaddingY(t.spacing(1))
		wrapper.Children(func() {
			chip, del := tagChipStyled(row.Context.Key("tag"), t, label, namedMessage(c, "removeNamed", label), size, o.ReadOnly, o.DeleteIcon, remove, func(slot Slot, part ui.Element) {
				if slot == Label {
					part.Grow(1)
				}
			})
			chip.FillWidth()
			for _, key := range []ui.Key{ui.KeyDelete, ui.KeyBackspace} {
				del.OnShortcut(0, key, remove)
			}
			for _, key := range []ui.Key{ui.KeyUp, back} {
				del.OnShortcut(0, key, func() { request(selected.positions[id] - 1) })
			}
			for _, key := range []ui.Key{ui.KeyDown, forth} {
				del.OnShortcut(0, key, func() { request(selected.positions[id] + 1) })
			}
			del.OnShortcut(0, ui.KeyHome, func() { request(0) }).OnShortcut(0, ui.KeyEnd, func() { request(selected.Len() - 1) })
			if row.Index+1 < selected.Len() {
				del.OnShortcut(0, ui.KeyTab, func() { request(selected.positions[id] + 1) })
			}
			if row.Index > 0 {
				del.OnShortcut(ui.Shift, ui.KeyTab, func() { request(selected.positions[id] - 1) })
			}
			if tags.pending == id && !del.IsDisabled() {
				del.Focus()
				tags.pending = ""
			}
		})
	})
}
