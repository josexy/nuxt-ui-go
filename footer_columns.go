package nui

import (
	"math"

	"github.com/egoist/mygo/ui"
)

type FooterColumn struct {
	ID, Title string
	Items     []NavigationItem
}

type FooterColumnsOptions struct {
	MinColumnWidth float32
	OnNavigate     func(string)
	Decoration
}

func FooterColumns(c *ui.Context, t Theme, columns []FooterColumn, o FooterColumnsOptions) ui.Element {
	w := o.MinColumnWidth
	if math.IsNaN(float64(w)) || math.IsInf(float64(w), 0) {
		panic("nui: FooterColumns width must be finite")
	}
	if w <= 0 {
		w = 160
	}
	seen := map[string]bool{}
	root := ui.Row(c).Wrap().AlignItems(ui.Start).MinWidth(0).MaxWidthPercent(100)
	finish(t, ComponentFooterColumns, Root, root, PartStyle{Gap: Set(t.spacing(32))}, o.Decoration)
	root.Children(func() {
		for _, col := range columns {
			checkID(seen, col.ID, "FooterColumns")
			column := ui.Column(c.Key(col.ID)).Grow(1).Basis(w).MinWidth(0).MaxWidthPercent(100)
			finish(t, ComponentFooterColumns, Body, column, PartStyle{Gap: Set(t.spacing(24))}, o.Decoration)
			column.Children(func() {
				if col.Title != "" {
					finish(t, ComponentFooterColumns, Label, ui.Text(c.Key("title"), col.Title).MinWidth(0).Role(ui.RoleHeading).Level(3), PartStyle{Foreground: Set(t.Highlighted), FontSize: Set(t.text(14)), Weight: Set(600)}, o.Decoration)
				}
				ids := map[string]bool{}
				list := ui.Column(c.Key("links")).MinWidth(0).AlignItems(ui.Start)
				finish(t, ComponentFooterColumns, List, list, PartStyle{Gap: Set(t.spacing(16))}, o.Decoration)
				list.Children(func() {
					for _, it := range col.Items {
						checkID(ids, it.ID, "FooterColumns item")
						link := ui.ButtonBase(c.Key(it.ID)).Role(ui.RoleLink).Label(it.Label).Disabled(it.Disabled).FocusRing(false).MinWidth(0).MaxWidthPercent(100).AlignItems(ui.Center)
						foreground, weight := t.Muted, 400
						if it.Active {
							foreground, weight = t.Primary, 500
						} else if !link.IsDisabled() && link.Hovered() {
							foreground = t.Text
						}
						if link.FocusVisible() {
							link.Shadow(0, 0, 0, 3, t.focusColor(Primary))
						}
						finishFooterLinkPart(t, Item, link, PartStyle{Foreground: Set(foreground), Weight: Set(weight), Font: Set(t.Font), FontSize: Set(t.text(14)), LineHeight: Set(float32(20.0 / 14)), Radius: Set(t.radius(2)), Gap: Set(t.spacing(6))}, o.Decoration, it.Decoration)
						link.Children(func() {
							if it.Icon != nil {
								icon := ui.Icon(c.Key("icon"), it.Icon).Size(t.text(20), t.text(20)).Shrink(0)
								finishFooterLinkPart(t, ItemLeadingSlot, icon, PartStyle{}, o.Decoration, it.Decoration)
							}
							if it.Label != "" {
								label := ui.Text(c.Key("label"), it.Label).SingleLine().MinWidth(0)
								finishFooterLinkPart(t, ItemLabelSlot, label, PartStyle{}, o.Decoration, it.Decoration)
							}
						})
						link.OnClick(func() {
							event := &NavigationEvent{ID: it.ID}
							if it.OnSelect != nil {
								it.OnSelect(event)
							}
							if !event.prevented && o.OnNavigate != nil {
								o.OnNavigate(it.ID)
							}
						})
					}
				})
			})
		}
	})
	return root
}

func finishFooterLinkPart(t Theme, slot Slot, e ui.Element, base PartStyle, component, item Decoration) {
	base = mergeStyle(base, t.components[ComponentFooterColumns].UI[slot])
	base = mergeStyle(base, component.UI[slot])
	base = mergeStyle(base, item.UI[slot])
	base.apply(e)
	if component.Style != nil {
		component.Style(slot, e)
	}
	if item.Style != nil {
		item.Style(slot, e)
	}
}
