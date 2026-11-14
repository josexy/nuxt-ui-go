package nui

import (
	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

func finishNavigationPart(t Theme, slot Slot, e ui.Element, base PartStyle, global, node Decoration) {
	base = mergeStyle(base, t.components[ComponentNavigationMenu].UI[slot])
	base = mergeStyle(base, global.UI[slot])
	base = mergeStyle(base, node.UI[slot])
	base.apply(e)
	if global.Style != nil {
		global.Style(slot, e)
	}
	if node.Style != nil {
		node.Style(slot, e)
	}
}

func navigationMenuSlot(slot Slot) Slot {
	switch slot {
	case Header:
		return Label
	case Label:
		return ItemLabelSlot
	case Leading:
		return ItemLeadingSlot
	case Trailing:
		return ItemTrailingSlot
	default:
		return slot
	}
}

func navigationMenuDecoration(t Theme, o NavigationOptions, item NavigationItem, base Decoration) Decoration {
	styles := mergeStyles(mergeStyles(t.components[ComponentNavigationMenu].UI, o.UI), item.UI)
	mapped := Styles{}
	for _, slot := range []Slot{Header, Item, Body, Leading, Label, Description, Trailing, Indicator} {
		if style, ok := styles[navigationMenuSlot(slot)]; ok {
			mapped[slot] = style
		}
	}
	base.UI = mergeStyles(base.UI, mapped)
	previous := base.Style
	base.Style = func(slot Slot, e ui.Element) {
		if previous != nil {
			previous(slot, e)
		}
		slot = navigationMenuSlot(slot)
		if o.Style != nil {
			o.Style(slot, e)
		}
		if item.Style != nil {
			item.Style(slot, e)
		}
	}
	return base
}

func navigationButton(c *ui.Context, t Theme, it NavigationItem, o NavigationOptions, config ComponentTheme, label string, active, open, compact bool, depth int) ui.Element {
	if it.Label == "" {
		panic("nui: NavigationMenu items require labels")
	}
	b := ui.ButtonBase(c).Label(it.Label).Description(it.Description).Disabled(it.Disabled).Selected(active).MinWidth(0).AlignItems(ui.Center).FocusRing(false)
	fg, bg := t.Muted, ui.Transparent
	hover := !b.IsDisabled() && (b.Hovered() || b.Pressed())
	if active {
		fg = t.color(config.Color)
		if config.Color == Neutral {
			fg = t.Highlighted
		}
	} else if hover || open {
		fg = t.Highlighted
	}
	if config.Variant == Soft {
		if active && !o.Highlight {
			bg = t.Elevated
		} else if hover || (open && !active) {
			bg = t.Elevated.Alpha(.5)
		}
	}
	padding := t.spacing(10)
	if compact {
		padding = t.spacing(6)
	}
	if b.FocusVisible() {
		b.Shadow(0, 0, 0, 3, inputFocusColor(t, config.Color))
	}
	finishNavigationPart(t, Item, b, PartStyle{Font: Set(t.Font), FontSize: Set(t.text(14)), LineHeight: Set(float32(20.0 / 14)), Weight: Set(500), PaddingX: Set(padding), PaddingY: Set(t.spacing(6)), Gap: Set(t.spacing(6)), Foreground: Set(fg), Background: Set(bg), Radius: Set(t.radius(6))}, o.Decoration, it.Decoration)
	b.Children(func() {
		var leading ui.Element
		if it.Avatar != nil {
			diameter := Optional[float32]{}
			if it.Avatar.Size == SizeDefault {
				diameter = Set(t.text(20))
			}
			leading = avatar(c.Key("avatar"), t, *it.Avatar, diameter)
		} else if it.Icon != nil {
			icon := func() { ui.Icon(c, it.Icon).Size(t.text(20), t.text(20)).Shrink(0) }
			if it.Chip != nil {
				chip := *it.Chip
				if chip.Size == SizeDefault {
					chip.Size = SM
				}
				leading = Chip(c.Key("chip"), t, chip, icon)
			} else {
				leading = ui.Icon(c.Key("icon"), it.Icon).Size(t.text(20), t.text(20)).Shrink(0)
			}
		}
		if leading.Valid() {
			ink := t.Dimmed
			if active {
				ink = fg
			} else if hover || open {
				ink = t.Text
			}
			finishNavigationPart(t, ItemLeadingSlot, leading, PartStyle{Foreground: Set(ink)}, o.Decoration, it.Decoration)
		}
		if label != "" {
			finishNavigationPart(t, ItemLabelSlot, ui.Text(c.Key("label"), label).SingleLine().MinWidth(0), PartStyle{}, o.Decoration, it.Decoration)
		}
		if !compact && (it.Badge != "" || it.BadgeOptions != nil || len(it.Children) > 0 || it.TrailingIcon != nil) {
			trailing := ui.Row(c.Key("trailing")).MarginStart(ui.Auto).Shrink(0).Gap(t.spacing(6)).AlignItems(ui.Center)
			finishNavigationPart(t, ItemTrailingSlot, trailing, PartStyle{}, o.Decoration, it.Decoration)
			trailing.Children(func() {
				if it.Badge != "" || it.BadgeOptions != nil {
					opts := BadgeOptions{}
					if it.BadgeOptions != nil {
						opts = *it.BadgeOptions
					}
					if opts.Color == ColorDefault {
						opts.Color = Neutral
					}
					if opts.Size == SizeDefault {
						opts.Size = SM
					}
					if opts.Variant == VariantDefault {
						opts.Variant = Outline
					}
					Badge(c.Key("badge"), t, it.Badge, opts)
				}
				icon := it.TrailingIcon
				if icon == nil && len(it.Children) > 0 {
					icon = o.TrailingIcon
					if icon == nil {
						if open {
							icon = lucide.ChevronUp()
						} else {
							icon = lucide.ChevronDown()
						}
					}
				}
				if icon != nil {
					ui.Icon(c.Key("trailing-icon"), icon).Size(t.text(20), t.text(20))
				}
			})
		}
		if o.Highlight && active && (!o.Vertical || depth > 0) {
			color := o.HighlightColor
			if color == ColorDefault {
				color = config.Color
			}
			ink := t.color(color)
			if color == Neutral {
				ink = t.InvertedBackground
			}
			indicator := ui.Box(c.Key("indicator")).Absolute().IgnoreInput(true).Background(ink)
			if o.Vertical {
				indicator.Width(1).Start(-t.spacing(6)).Top(t.spacing(2)).Bottom(t.spacing(2))
			} else {
				indicator.Height(1).Bottom(-t.spacing(8)).Start(t.spacing(10)).End(t.spacing(10))
			}
			style := mergeStyle(PartStyle{Background: Set(ink)}, t.components[ComponentNavigationMenu].UI[Indicator])
			style = mergeStyle(mergeStyle(style, o.UI[Indicator]), it.UI[Indicator])
			styleLine(indicator, style, o.Vertical, false, 0)
			if o.Style != nil {
				o.Style(Indicator, indicator)
			}
			if it.Style != nil {
				it.Style(Indicator, indicator)
			}
		}
	})
	return b
}
