package nui

import (
	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

func listboxItemDecoration(base Decoration, options *ListboxItemOptions) Decoration {
	if options == nil {
		return base
	}
	base.UI = mergeStyles(base.UI, options.UI)
	if options.Style != nil {
		parent := base.Style
		base.Style = func(slot Slot, e ui.Element) {
			if parent != nil {
				parent(slot, e)
			}
			options.Style(slot, e)
		}
	}
	return base
}

func listboxStructure(c *ui.Context, t Theme, size Size, item ListboxItem, width float32, horizontal bool, base Decoration) {
	d := listboxItemDecoration(base, item.Options)
	var part ui.Element
	if item.Kind == ListboxLabel {
		part = ui.Text(c, item.Label).SingleLine()
		finish(t, ComponentListbox, Label, part, optionLabelStyle(t, size), d)
	} else {
		part = ui.Column(c).Padding(t.spacing(4), 0)
		part.Children(func() {
			finishLine(t, ComponentListbox, SeparatorSlot, ui.Box(c.Key("line")).FillWidth().Height(1), PartStyle{Background: Set(t.Border)}, d, false, true, 0)
		})
	}
	if horizontal {
		part.Width(width).MinWidth(width).MaxWidth(width).Shrink(0)
	}
}

func listboxLoading(c *ui.Context, t Theme, size Size, o ListboxOptions) {
	m := optionRecipe(t, size)
	style := PartStyle{Foreground: Set(t.Muted), PaddingY: Set(t.spacing([5]float32{12, 16, 24, 28, 32}[sizeIndex(size)]))}
	style = mergeStyle(mergeStyle(style, t.components[ComponentListbox].UI[LoadingSlot]), o.UI[LoadingSlot])
	// A fixed key keeps the status's identity while rows or an empty message
	// come and go beside it.
	root := ui.Column(c.Key("loading")).FillWidth().Center().Shrink(0).Role(ui.RoleStatus).Label(message(c, "loading"))
	decorated(LoadingSlot, root, Styles{LoadingSlot: style}, o.Style)
	root.Children(func() {
		icon := o.LoadingIcon
		if icon == nil {
			icon = lucide.LoaderCircle()
		}
		part := ui.Box(c).Size(m.Icon, m.Icon).Shrink(0)
		color := mergeValue(style.Foreground, t.components[ComponentListbox].UI[LoadingIconSlot].Foreground)
		color = mergeValue(color, o.UI[LoadingIconSlot].Foreground)
		finish(t, ComponentListbox, LoadingIconSlot, part, PartStyle{}, o.Decoration)
		drawLoadingSVG(part, icon, color.Value, t.motion(c), t.Motion.Spin)
	})
}

func listboxItemContent(c *ui.Context, t Theme, item ListboxItem, o ListboxOptions, d Decoration, size Size, chosen, highlighted bool) {
	m := optionRecipe(t, size)
	switch {
	case item.Icon != nil:
		finish(t, ComponentListbox, Leading, ui.Icon(c.Key("leading"), item.Icon).Size(m.Icon, m.Icon).Shrink(0), PartStyle{Foreground: Set(optionLeadingColor(t, highlighted))}, d)
	case item.Avatar != nil:
		options := *item.Avatar
		diameter := Optional[float32]{}
		if options.Size == SizeDefault {
			diameter = Set(m.Icon)
		}
		finish(t, ComponentListbox, Leading, avatar(c.Key("leading"), t, options, diameter).Shrink(0), PartStyle{}, d)
	case item.Chip != nil:
		options := *item.Chip
		options.Inset = true
		if options.Size == SizeDefault {
			options.Size = [5]Size{SM, SM, MD, MD, LG}[sizeIndex(size)]
		}
		leading := ui.Box(c.Key("leading")).Size(m.Icon, m.Icon).Center().Shrink(0)
		finish(t, ComponentListbox, Leading, leading, PartStyle{}, d)
		leading.Children(func() { Chip(c, t, options, nil) })
	}
	build := o.Build != nil && item.ID != o.actionID
	if build || item.Label != "" || item.Description != "" {
		body := ui.Column(c.Key("body")).Grow(1).MinWidth(0)
		finish(t, ComponentListbox, Body, body, PartStyle{Gap: Set(float32(0))}, d)
		body.Children(func() {
			if build {
				o.Build(c, t, item)
				return
			}
			if item.Label != "" {
				finish(t, ComponentListbox, ItemLabelSlot, ui.Text(c.Key("label"), item.Label).SingleLine().MinWidth(0), PartStyle{}, d)
			}
			if item.Description != "" {
				style := PartStyle{Foreground: Set(t.Muted)}
				if size == XL {
					style.FontSize, style.LineHeight = Set(t.text(14)), Set(float32(20.0/14))
				}
				finish(t, ComponentListbox, Description, ui.Text(c.Key("description"), item.Description).SingleLine().MinWidth(0), style, d)
			}
		})
	} else {
		ui.Box(c.Key("space")).Grow(1)
	}
	trailing := ui.Row(c.Key("trailing")).AlignItems(ui.Center).Shrink(0)
	finish(t, ComponentListbox, Trailing, trailing, PartStyle{Gap: Set(t.spacing(6))}, d)
	trailing.Children(func() {
		if chosen {
			icon := o.SelectedIcon
			if icon == nil {
				icon = lucide.Check()
			}
			finish(t, ComponentListbox, Indicator, ui.Icon(c.Key("selected"), icon).Size(m.Icon, m.Icon).Shrink(0), PartStyle{}, d)
		}
	})
}

func selectMenuClear(c *ui.Context, t Theme, size Size, o SelectMenuOptions, handle *ui.Handle, clear func()) ui.Element {
	root := ui.Row(c).AlignItems(ui.Center).Shrink(0)
	root.Children(func() {
		options := ButtonOptions{}
		if o.ClearOptions != nil {
			options = *o.ClearOptions
		}
		if options.Size == SizeDefault {
			options.Size = size
		}
		if options.Color == ColorDefault {
			options.Color = Neutral
		}
		if options.Variant == VariantDefault {
			options.Variant = LinkVariant
		}
		if options.AccessibleLabel == "" {
			options.AccessibleLabel = message(c, "clearSelection")
		}
		if options.Leading == nil {
			options.Leading = o.ClearIcon
			if options.Leading == nil {
				options.Leading = lucide.X()
			}
		}
		style := options.Style
		options.Style = func(button, label, leading, trailing ui.Element) {
			button.Bind(handle).Padding(0).Shrink(0)
			finish(t, ComponentSelectMenu, Close, button, PartStyle{}, o.Decoration)
			if style != nil {
				style(button, label, leading, trailing)
			}
		}
		Button(c.Key("clear"), t, "", options).KeepFocus().OnClick(clear)
	})
	return root
}

// Trigger parts and row parts have separate names, so styling the trigger's
// label, avatar, or container does not also resize or recolor popup rows.
func selectMenuListDecoration(t Theme, o SelectMenuOptions) Decoration {
	parts := [...]struct{ list, menu Slot }{
		{List, List}, {Item, Item}, {Leading, ItemLeadingSlot}, {Trailing, ItemTrailingSlot},
		{Label, Header}, {ItemLabelSlot, ItemLabelSlot}, {Body, Body}, {Description, Description},
		{Indicator, Indicator}, {SeparatorSlot, SeparatorSlot}, {EmptySlot, EmptySlot},
	}
	d := Decoration{UI: Styles{Root: {BorderWidth: Set(float32(0)), Radius: Set(float32(0)), Background: Set(ui.Transparent)}}}
	for _, part := range parts {
		d.UI[part.list] = mergeStyle(t.components[ComponentSelectMenu].UI[part.menu], o.UI[part.menu])
	}
	if o.Style != nil {
		d.Style = func(slot Slot, e ui.Element) {
			for _, part := range parts {
				if part.list == slot {
					o.Style(part.menu, e)
					return
				}
			}
		}
	}
	return d
}
