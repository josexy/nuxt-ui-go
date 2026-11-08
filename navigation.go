package nui

import (
	"fmt"
	"slices"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

type TabItem struct {
	ID, Label    string
	Icon         *ui.SVG
	Avatar       *AvatarOptions
	Badge        string
	BadgeOptions *BadgeOptions
	Disabled     bool
	Content      func(*ui.Context, Theme)
	UI           Styles
}
type TabsOptions struct {
	Label                            string
	Color                            Color
	Size                             Size
	Underline, Vertical, KeepMounted bool
	Content                          Optional[bool] // Defaults to true.
	ActivationMode                   ActivationMode
	OnChange                         func(string)
	Decoration
}

type ActivationMode uint8

const (
	ActivationAutomatic ActivationMode = iota
	ActivationManual
)

type tabsMemory struct {
	triggers, bodies map[string]*ui.Handle
	active           string
	changed          bool
}

func Tabs(c *ui.Context, t Theme, value *string, items []TabItem, o TabsOptions) ui.Element {
	if o.ActivationMode > ActivationManual {
		panic("nui: invalid tab activation mode")
	}
	if value == nil {
		panic("nui: Tabs requires a selected ID")
	}
	seen := map[string]bool{}
	enabled := []string{}
	for _, it := range items {
		checkID(seen, it.ID, "Tabs")
		if !it.Disabled {
			enabled = append(enabled, it.ID)
		}
	}
	index := slices.Index(enabled, *value)
	if index < 0 && len(enabled) > 0 {
		index = 0
		*value = enabled[0]
	}
	config := selected(defaults(t, ComponentTabs, ComponentTheme{Color: Primary, Size: MD}), ComponentTheme{Color: o.Color, Size: o.Size})
	root := ui.Column(c).MinWidth(0)
	memory := ui.Local(root, "nui-tabs", func() tabsMemory {
		return tabsMemory{triggers: map[string]*ui.Handle{}, bodies: map[string]*ui.Handle{}}
	})
	if memory.changed {
		root.SetChanged()
		memory.changed = false
	}
	focused := ""
	for id, handle := range memory.triggers {
		if handle.Focused(c) {
			focused = id
		}
	}
	restore := focused != "" && !slices.Contains(enabled, focused)
	if previous := memory.bodies[memory.active]; previous != nil && previous.FocusWithin(c) {
		hasContent := false
		for _, it := range items {
			if it.ID == *value {
				hasContent = it.Content != nil
				break
			}
		}
		if memory.active != *value || !valueOr(o.Content, true) || !hasContent {
			restore = true
		}
	}
	for id := range memory.triggers {
		if !seen[id] {
			memory.triggers[id].CancelFocus(c)
			memory.bodies[id].CancelFocus(c)
			delete(memory.triggers, id)
			delete(memory.bodies, id)
		}
	}
	for _, it := range items {
		if memory.triggers[it.ID] == nil {
			memory.triggers[it.ID] = &ui.Handle{}
			memory.bodies[it.ID] = &ui.Handle{}
		}
	}
	memory.active = *value
	index = slices.Index(enabled, *value)
	if o.Vertical {
		root.Row().AlignItems(ui.Center)
	}
	finish(t, ComponentTabs, Root, root, PartStyle{Gap: Set(t.spacing(8)), Font: Set(t.Font), FontSize: Set(t.text(t.FontSize)), Foreground: Set(t.Text)}, o.Decoration)
	root.Children(func() {
		parts := ui.TabsBase(c, &index, len(enabled))
		list := parts.List.MinWidth(0).Label(o.Label)
		if o.Vertical {
			list.Column().AlignItems(ui.Stretch).FocusGroup(ui.Vertical)
		} else {
			list.FillWidth()
		}
		// Pill lists are an elevated rounded-lg surface; underline lists have
		// Nuxt's border-b (border-s when vertical), which the indicator covers.
		listStyle := PartStyle{Background: Set(t.Elevated), Radius: Set(t.radius(8)), PaddingX: Set(t.spacing(4)), PaddingY: Set(t.spacing(4))}
		if o.Underline {
			listStyle.Background, listStyle.Radius = Set(ui.Transparent), Set(float32(0))
			if o.Vertical {
				list.BorderStartWidth(1)
			} else {
				list.BorderWidth(0, 0, 1, 0)
			}
			list.BorderColor(t.Border)
		}
		finish(t, ComponentTabs, List, list, listStyle, o.Decoration)
		listStyle = mergeStyle(mergeStyle(listStyle, t.components[ComponentTabs].UI[List]), o.UI[List])
		list.Children(func() {
			for _, it := range items {
				decoration := o.Decoration
				decoration.UI = mergeStyles(o.UI, it.UI)
				var trigger ui.Element
				choose := func(id string) {
					if *value != id {
						*value = id
						memory.changed = true
						if o.OnChange != nil {
							o.OnChange(id)
						}
					}
				}
				if it.Disabled {
					trigger = ui.Row(c.Key(it.ID)).Role(ui.RoleTab).Disabled(true)
				} else {
					trigger = ui.ButtonBase(c.Key(it.ID)).Role(ui.RoleTab).Selected(it.ID == *value).OnClick(func() {
						choose(it.ID)
					})
					move := func(delta int, edge bool) {
						next := delta
						if !edge {
							next = (slices.Index(enabled, it.ID) + delta + len(enabled)) % len(enabled)
						}
						id := enabled[next]
						if o.ActivationMode == ActivationAutomatic {
							choose(id)
						}
						memory.triggers[id].Focus()
					}
					back, forward := ui.KeyLeft, ui.KeyRight
					if c.Direction() == ui.RightToLeft {
						back, forward = forward, back
					}
					if o.Vertical {
						back, forward = ui.KeyUp, ui.KeyDown
					}
					trigger.OnShortcut(0, back, func() { move(-1, false) }).OnShortcut(0, forward, func() { move(1, false) }).
						OnShortcut(0, ui.KeyHome, func() { move(0, true) }).OnShortcut(0, ui.KeyEnd, func() { move(len(enabled)-1, true) })
				}
				active := index == slices.Index(enabled, it.ID) && !it.Disabled
				trigger.Label(it.Label).MinWidth(0).AlignItems(ui.Center).FocusRing(true).Bind(memory.triggers[it.ID])
				background := ui.Transparent
				foreground := t.Muted
				indicatorColor := t.color(config.Color)
				if config.Color == Neutral {
					indicatorColor = t.InvertedBackground
				}
				if active {
					foreground = t.color(config.Color)
					if !o.Underline {
						foreground = t.Inverted
					}
				}
				if trigger.Hovered() && !it.Disabled && !active {
					foreground = t.Text
				}
				metrics := buttonMetrics(t, config.Size)
				if config.Size == MD {
					metrics.X = t.spacing(12)
				}
				if !o.Underline {
					trigger.Grow(1).Justify(ui.Center)
				}
				gap := t.spacing([5]float32{4, 6, 6, 8, 8}[sizeIndex(config.Size)])
				finish(t, ComponentTabs, Trigger, trigger, PartStyle{Background: Set(background), Foreground: Set(foreground), Radius: Set(t.radius(6)), PaddingX: Set(metrics.X), PaddingY: Set(metrics.Y), Gap: Set(gap), Font: Set(t.Font), FontSize: Set(metrics.Font), LineHeight: Set(metrics.Line / metrics.Font), Weight: Set(500)}, decoration)
				trigger.Children(func() {
					if active {
						// Absolute children paint after normal children in MyGo. The
						// pill surface must style the trigger itself to stay behind text.
						indicator := trigger
						style := PartStyle{Background: Set(indicatorColor), Radius: Set(t.radius(999))}
						if o.Underline {
							indicator = ui.Box(c.Key("indicator")).Absolute().IgnoreInput(true)
							if o.Vertical {
								indicator.Top(0).Bottom(0).Width(1).Start(-listStyle.PaddingX.Value - 1)
							} else {
								indicator.Left(0).Right(0).Height(1).Bottom(-listStyle.PaddingY.Value - 1)
							}
						} else {
							style.Radius = Set(t.radius(6))
							style.Shadow = Set(ShadowTokens{Color: ui.RGBA(0, 0, 0, .05), Blur: t.spacing(2), OffsetY: t.spacing(1)})
							// Keep existing Trigger background overrides useful; Indicator
							// remains the more specific selection-surface override.
							if t.components[ComponentTabs].UI[Trigger].Background.Set || decoration.UI[Trigger].Background.Set {
								style.Background = Optional[ui.Color]{}
							}
						}
						finish(t, ComponentTabs, Indicator, indicator, style, decoration)
					}
					if it.Icon != nil {
						finish(t, ComponentTabs, Leading, ui.Icon(c.Key("leading"), it.Icon).Size(metrics.Icon, metrics.Icon).Shrink(0), PartStyle{}, decoration)
					} else if it.Avatar != nil {
						diameter := Optional[float32]{}
						if it.Avatar.Size == SizeDefault {
							diameter = Set(t.text([5]float32{16, 16, 20, 20, 24}[sizeIndex(config.Size)]))
						}
						finish(t, ComponentTabs, Leading, avatar(c.Key("leading"), t, *it.Avatar, diameter).Shrink(0), PartStyle{}, decoration)
					}
					if it.Label != "" {
						finish(t, ComponentTabs, Label, ui.Text(c.Key("label"), it.Label).SingleLine().MinWidth(0), PartStyle{}, decoration)
					}
					if it.Badge != "" || it.BadgeOptions != nil {
						badge := BadgeOptions{}
						if it.BadgeOptions != nil {
							badge = *it.BadgeOptions
						}
						if badge.Size == SizeDefault {
							badge.Size = SM
						}
						if badge.Color == ColorDefault {
							badge.Color = Neutral
						}
						if badge.Variant == VariantDefault {
							badge.Variant = Outline
						}
						finish(t, ComponentTabs, Trailing, Badge(c.Key("badge"), t, it.Badge, badge).Shrink(0), PartStyle{}, decoration)
					}
				})
			}
		})
		for _, it := range items {
			active := it.ID == *value && !it.Disabled
			if !valueOr(o.Content, true) || it.Content == nil || (!active && !o.KeepMounted) {
				continue
			}
			body := ui.Column(c.Key("content:" + it.ID)).MinWidth(0).Role(ui.RoleGroup).Label(namedMessage(c, "contentNamed", it.Label)).Focusable().FocusRing(true).Hidden(!active).Bind(memory.bodies[it.ID])
			if o.Vertical {
				body.Grow(1)
			} else {
				body.FillWidth()
			}
			decoration := o.Decoration
			decoration.UI = mergeStyles(o.UI, it.UI)
			finish(t, ComponentTabs, Body, body, PartStyle{Gap: Set(t.spacing(12)), Radius: Set(t.radius(6))}, decoration)
			body.Children(func() { it.Content(c, t) })
		}
	})
	if restore && len(enabled) > 0 {
		memory.triggers[*value].Focus()
	}
	return root
}

type BreadcrumbItem struct {
	ID, Label string
	Icon      *ui.SVG
	Avatar    *AvatarOptions
	Active    Optional[bool]
	Disabled  bool
	OnSelect  func()
	UI        Styles
}
type BreadcrumbOptions struct {
	Label         string
	MaxItems      int
	Color         Color
	SeparatorIcon *ui.SVG
	Decoration
}

func Breadcrumb(c *ui.Context, t Theme, items []BreadcrumbItem, o BreadcrumbOptions) ui.Element {
	seen := map[string]bool{}
	for _, it := range items {
		checkID(seen, it.ID, "Breadcrumb")
	}
	label := o.Label
	if label == "" {
		label = message(c, "breadcrumbLabel")
	}
	root := ui.Row(c).MinWidth(0).AlignItems(ui.Center).Role(ui.RoleGroup).Label(label)
	finish(t, ComponentBreadcrumb, Root, root, PartStyle{Gap: Set(t.spacing(6)), Font: Set(t.Font), FontSize: Set(t.text(14)), LineHeight: Set(float32(20.0 / 14))}, o.Decoration)
	finish(t, ComponentBreadcrumb, List, root, PartStyle{}, o.Decoration)
	color := selected(defaults(t, ComponentBreadcrumb, ComponentTheme{Color: Primary}), ComponentTheme{Color: o.Color}).Color
	current := t.color(color)
	separator := o.SeparatorIcon
	if separator == nil {
		separator = lucide.ChevronRight()
		if c.Direction() == ui.RightToLeft {
			separator = lucide.ChevronLeft()
		}
	}
	visible := make([]int, 0, len(items))
	for i := range items {
		if o.MaxItems >= 2 && len(items) > o.MaxItems && i > 0 && i < len(items)-o.MaxItems+1 {
			if i == 1 {
				visible = append(visible, -1)
			}
			continue
		}
		visible = append(visible, i)
	}
	root.Children(func() {
		for position, index := range visible {
			if position > 0 {
				d := o.Decoration
				if previous := visible[position-1]; previous >= 0 {
					d.UI = mergeStyles(o.UI, items[previous].UI)
				}
				finish(t, ComponentBreadcrumb, SeparatorSlot, ui.Icon(c.Key(fmt.Sprintf("separator:%d", index)), separator).Size(t.text(20), t.text(20)).Shrink(0), PartStyle{Foreground: Set(t.Muted)}, d)
			}
			if index < 0 {
				finish(t, ComponentBreadcrumb, Ellipsis, ui.Icon(c.Key("ellipsis"), lucide.Ellipsis()).Size(t.text(20), t.text(20)).Shrink(0), PartStyle{Foreground: Set(t.Muted)}, o.Decoration)
				continue
			}
			it := items[index]
			d := o.Decoration
			d.UI = mergeStyles(o.UI, it.UI)
			var item ui.Element
			if it.OnSelect != nil {
				item = ui.ButtonBase(c.Key("item:" + it.ID)).Role(ui.RoleLink).OnClick(it.OnSelect).FocusRing(false)
			} else {
				item = ui.Row(c.Key("item:" + it.ID))
			}
			if it.OnSelect != nil {
				item.Label(it.Label)
			}
			item.Disabled(it.Disabled).MinWidth(0).Shrink(1).AlignItems(ui.Center)
			active := valueOr(it.Active, index == len(items)-1)
			fg, weight := t.Muted, 500
			if active {
				fg, weight = current, 600
			} else if it.OnSelect != nil && item.Hovered() && !item.IsDisabled() {
				fg = t.Text
			}
			base := PartStyle{Foreground: Set(fg), Weight: Set(weight), FontSize: Set(t.text(14)), LineHeight: Set(float32(20.0 / 14)), Gap: Set(t.spacing(6)), Radius: Set(t.radius(6))}
			if item.IsDisabled() {
				base.Opacity = Set(float32(.75))
			}
			if item.FocusVisible() {
				item.Shadow(0, 0, 0, 3, t.focusColor(color))
			}
			finish(t, ComponentBreadcrumb, Item, item, base, d)
			finish(t, ComponentBreadcrumb, Trigger, item, PartStyle{}, d)
			if active && index == len(items)-1 && it.OnSelect != nil {
				item.Description(message(c, "currentPage"))
			}
			item.Children(func() {
				if it.Icon != nil {
					finish(t, ComponentBreadcrumb, Leading, ui.Icon(c.Key("leading"), it.Icon).Size(t.text(20), t.text(20)).Shrink(0), PartStyle{}, d)
				} else if it.Avatar != nil {
					diameter := Optional[float32]{}
					if it.Avatar.Size == SizeDefault {
						diameter = Set(t.text(20))
					}
					finish(t, ComponentBreadcrumb, Leading, avatar(c.Key("leading"), t, *it.Avatar, diameter).Shrink(0), PartStyle{}, d)
				}
				if it.Label != "" {
					text := ui.Text(c.Key("label"), it.Label).SingleLine().MinWidth(0)
					if active && index == len(items)-1 && it.OnSelect == nil {
						text.Description(message(c, "currentPage"))
					}
					finish(t, ComponentBreadcrumb, Label, text, PartStyle{}, d)
				}
			})
		}
	})
	return root
}

type PaginationOptions struct {
	Label                     string
	Total, PageSize, Siblings int
	// SiblingCount overrides legacy Siblings and can explicitly request zero.
	SiblingCount                                          Optional[int]
	ShowControls, ShowEdges                               Optional[bool]
	Color, ActiveColor                                    Color
	Variant, ActiveVariant                                Variant
	Size                                                  Size
	Disabled                                              bool
	FirstIcon, PrevIcon, NextIcon, LastIcon, EllipsisIcon *ui.SVG
	OnChange                                              func(int)
	Decoration
}

func Pagination(c *ui.Context, t Theme, page *int, o PaginationOptions) ui.Element {
	if page == nil {
		panic("nui: Pagination requires page state")
	}
	size := o.PageSize
	if size <= 0 {
		size = 10
	}
	pages := 1
	if o.Total > 0 {
		pages = 1 + (o.Total-1)/size
	}
	*page = max(1, min(*page, pages))
	siblings := o.Siblings
	if siblings <= 0 {
		siblings = 2
	}
	siblings = min(siblings, 10)
	if o.SiblingCount.Set {
		siblings = max(0, min(o.SiblingCount.Value, 10))
	}
	config := selected(defaults(t, ComponentPagination, ComponentTheme{Color: Neutral, Variant: Outline, Size: MD}), ComponentTheme{Color: o.Color, Variant: o.Variant, Size: o.Size})
	activeConfig := selected(ComponentTheme{Color: Primary, Variant: Solid}, ComponentTheme{Color: o.ActiveColor, Variant: o.ActiveVariant})
	root := ui.Row(c).Wrap().MinWidth(0).AlignItems(ui.Center).Role(ui.RoleGroup).Label(o.Label).Disabled(o.Disabled)
	finish(t, ComponentPagination, Root, root, PartStyle{Gap: Set(t.spacing(4)), Font: Set(t.Font)}, o.Decoration)
	finish(t, ComponentPagination, List, root, PartStyle{}, o.Decoration)
	choose := func(next int) {
		if next == *page {
			return
		}
		*page = next
		if o.OnChange != nil {
			o.OnChange(next)
		}
	}
	first, previous, next, last := lucide.ChevronsLeft(), lucide.ChevronLeft(), lucide.ChevronRight(), lucide.ChevronsRight()
	if c.Direction() == ui.RightToLeft {
		first, previous, next, last = last, next, previous, first
	}
	if o.FirstIcon != nil {
		first = o.FirstIcon
	}
	if o.PrevIcon != nil {
		previous = o.PrevIcon
	}
	if o.NextIcon != nil {
		next = o.NextIcon
	}
	if o.LastIcon != nil {
		last = o.LastIcon
	}
	icons := map[string]*ui.SVG{"first": first, "previous": previous, "next": next, "last": last}
	controlSlots := map[string]Slot{"first": First, "previous": Previous, "next": Next, "last": Last}
	root.Children(func() {
		pageButton := func(key, label string, next int, disabled, active bool) {
			variant := config.Variant
			color := config.Color
			if active {
				variant = activeConfig.Variant
				color = activeConfig.Color
			}
			opts := ButtonOptions{Variant: variant, Color: color, Size: config.Size, Disabled: disabled || o.Disabled, Square: true, Leading: icons[key]}
			if opts.Leading != nil {
				label = ""
			}
			opts.Style = func(root, label, leading, trailing ui.Element) {
				if icons[key] == nil {
					label.MinWidth(t.text(20)).TextAlign(ui.Center)
					finish(t, ComponentPagination, Label, label, PartStyle{}, o.Decoration)
				} else {
					finish(t, ComponentPagination, Leading, leading, PartStyle{}, o.Decoration)
				}
			}
			accessible := formatMessage(c, "pageNumber", map[string]any{"number": next})
			switch key {
			case "first":
				accessible = message(c, "pageFirst")
			case "previous":
				accessible = message(c, "pagePrevious")
			case "next":
				accessible = message(c, "pageNext")
			case "last":
				accessible = message(c, "pageLast")
			}
			opts.AccessibleLabel = accessible
			e := button(c.Key(key), t, label, opts, Set(active)).OnClick(func() { choose(next) })
			finish(t, ComponentPagination, Item, e, PartStyle{}, o.Decoration)
			if slot, ok := controlSlots[key]; ok {
				finish(t, ComponentPagination, slot, e, PartStyle{}, o.Decoration)
			}
		}
		if valueOr(o.ShowControls, true) {
			pageButton("first", "", 1, *page == 1, false)
			pageButton("previous", "", max(1, *page-1), *page == 1, false)
		}
		candidates := paginationPages(*page, pages, siblings, valueOr(o.ShowEdges, false))
		before := 0
		for _, n := range candidates {
			if before > 0 && n-before > 1 {
				// A non-interactive square with the buttons' look, as Nuxt's.
				m := buttonMetrics(t, config.Size)
				gap := ui.Row(c.Key(fmt.Sprintf("gap-%d", n))).Shrink(0).AlignItems(ui.Center).IgnoreInput(true)
				style := mergeStyle(controlStyle(t, m), appearance(t, config.Color, config.Variant, false, false))
				style.PaddingX = style.PaddingY
				finish(t, ComponentPagination, Ellipsis, gap, style, o.Decoration)
				icon := o.EllipsisIcon
				if icon == nil {
					icon = lucide.Ellipsis()
				}
				gap.Children(func() {
					finish(t, ComponentPagination, Leading, ui.Icon(c, icon).Size(m.Icon, m.Icon), PartStyle{}, o.Decoration)
				})
			}
			pageButton(fmt.Sprint(n), fmt.Sprint(n), n, false, n == *page)
			before = n
		}
		if valueOr(o.ShowControls, true) {
			pageButton("next", "", min(pages, *page+1), *page == pages, false)
			pageButton("last", "", pages, *page == pages, false)
		}
	})
	return root
}

func paginationPages(page, total, siblings int, edges bool) []int {
	start, end := max(1, page-siblings), min(total, page+siblings)
	// Match Reka's bounded window: near an edge use the otherwise empty
	// ellipsis slot for pages; small sets fit without ellipses at all.
	if edges {
		if total <= 2*siblings+5 {
			start, end = 1, total
		} else if start <= 3 {
			start, end = 1, 2*siblings+3
		} else if end >= total-2 {
			start, end = total-2*siblings-2, total
		}
	}
	count := min(total, 2*siblings+1)
	if !edges && start == 1 {
		end = count
	}
	if !edges && end == total {
		start = total - count + 1
	}
	pages := []int{}
	if edges && start > 1 {
		pages = append(pages, 1)
	}
	for n := start; n <= end; n++ {
		pages = append(pages, n)
	}
	if edges && end < total {
		pages = append(pages, total)
	}
	return pages
}

type StepItem struct {
	ID, Title, Description string
	Icon                   *ui.SVG
	Disabled               bool
	UI                     Styles
	Content                func(*ui.Context, Theme)
}
type StepperOptions struct {
	Label              string
	Vertical, Disabled bool
	// Linear defaults to true, as Nuxt's: a step can be reached only after
	// the one before it. Set(false) allows any available step.
	Linear    Optional[bool]
	Color     Color
	Size      Size
	CanSelect func(string) bool
	OnChange  func(string)
	Decoration
}

type stepperMemory struct {
	triggers map[string]*ui.Handle
	body     ui.Handle
	active   string
}

func Stepper(c *ui.Context, t Theme, value *string, items []StepItem, o StepperOptions) ui.Element {
	if value == nil {
		panic("nui: Stepper requires a selected ID")
	}
	seen := map[string]bool{}
	current := -1
	for i, it := range items {
		checkID(seen, it.ID, "Stepper")
		if it.ID == *value {
			current = i
		}
	}
	label := o.Label
	if label == "" {
		label = message(c, "stepperLabel")
	}
	root := ui.Column(c).MinWidth(0).Role(ui.RoleGroup).Label(label)
	if o.Vertical {
		root.Row().AlignItems(ui.Start)
	}
	memory := ui.Local(root, "stepper", func() stepperMemory { return stepperMemory{triggers: map[string]*ui.Handle{}} })
	restore := memory.body.FocusWithin(c) && (memory.active != *value || current < 0 || items[current].Content == nil)
	for id, handle := range memory.triggers {
		if !seen[id] {
			restore = restore || handle.Focused(c)
			handle.CancelFocus(c)
			delete(memory.triggers, id)
		}
	}
	for id := range seen {
		if memory.triggers[id] == nil {
			memory.triggers[id] = &ui.Handle{}
		}
	}
	config := selected(defaults(t, ComponentStepper, ComponentTheme{Color: Primary, Size: MD}), ComponentTheme{Color: o.Color, Size: o.Size})
	size := sizeIndex(config.Size)
	diameter := t.spacing([]float32{24, 32, 40, 48, 56}[size])
	iconSize := t.spacing([]float32{12, 16, 20, 24, 28}[size])
	fontSize := t.text([]float32{12, 14, 16, 18, 20}[size])
	textSize := t.text([]float32{12, 12, 14, 16, 18}[size])
	lineHeight := []float32{16.0 / 12, 16.0 / 12, 20.0 / 14, 24.0 / 16, 28.0 / 18}[size]
	wrapperGap := t.spacing([]float32{6, 8, 10, 12, 14}[size])
	connectorInset := t.spacing([]float32{16, 20, 28, 32, 36}[size])
	finish(t, ComponentStepper, Root, root, PartStyle{Gap: Set(t.spacing(16)), Font: Set(t.Font), Foreground: Set(t.Text)}, o.Decoration)
	accent := t.color(config.Color)
	if config.Color == Neutral {
		accent = t.InvertedBackground
	}
	root.Children(func() {
		list := ui.Row(c.Key("steps")).MinWidth(0).AlignItems(ui.Start).FocusGroup(ui.Horizontal).Disabled(o.Disabled)
		gap := float32(0)
		if o.Vertical {
			list.Column().AlignItems(ui.Stretch).FocusGroup(ui.Vertical).Shrink(0)
			gap = t.spacing(16)
		} else {
			list.FillWidth()
		}
		finish(t, ComponentStepper, List, list, PartStyle{Gap: Set(gap)}, o.Decoration)
		list.Children(func() {
			for i, it := range items {
				d := o.Decoration
				d.UI = mergeStyles(o.UI, it.UI)
				blocked := it.Disabled || o.Disabled || (valueOr(o.Linear, true) && i > current+1)
				item := ui.Column(c.Key(it.ID)).MinWidth(0)
				itemGap := float32(0)
				if o.Vertical {
					item.Row().AlignItems(ui.Stretch)
					itemGap = wrapperGap
				} else {
					item.Grow(1).Basis(0).TextAlign(ui.Center)
				}
				finish(t, ComponentStepper, Item, item, PartStyle{Gap: Set(itemGap)}, d)
				item.Children(func() {
					bg, fg := t.Elevated, t.Muted
					if i <= current {
						bg = accent
						fg = t.Inverted
					}
					header := ui.Column(c.Key("header")).MinWidth(0).AlignItems(ui.Center)
					if o.Vertical {
						header.Width(diameter).Shrink(0)
					}
					finish(t, ComponentStepper, Header, header, PartStyle{}, d)
					header.Children(func() {
						trigger := ui.ButtonBase(c.Key("trigger")).Size(diameter, diameter).Disabled(blocked).Role(ui.RoleButton).Label(it.Title).Description(it.Description).Selected(i == current).FocusRing(false).Bind(memory.triggers[it.ID])
						if blocked && memory.triggers[it.ID].Focused(c) {
							restore = true
						}
						if trigger.FocusVisible() {
							trigger.Shadow(0, 0, 0, 3, t.focusColor(config.Color))
						}
						finish(t, ComponentStepper, Trigger, trigger, PartStyle{Background: Set(bg), Foreground: Set(fg), Radius: Set(float32(999)), FontSize: Set(fontSize), Weight: Set(600)}, d)
						trigger.OnClick(func() {
							if o.CanSelect != nil && !o.CanSelect(it.ID) {
								return
							}
							if *value != it.ID {
								*value = it.ID
								if o.OnChange != nil {
									o.OnChange(it.ID)
								}
							}
						})
						trigger.Children(func() {
							marker := ui.Box(c).Fill().Center().Radius(999)
							finish(t, ComponentStepper, Indicator, marker, PartStyle{}, d)
							marker.Children(func() {
								if it.Icon != nil {
									finish(t, ComponentStepper, Leading, ui.Icon(c, it.Icon).Size(iconSize, iconSize), PartStyle{}, d)
								} else {
									ui.Textf(c, "%d", i+1)
								}
							})
						})
						if i < len(items)-1 {
							color := t.Accented
							if i < current {
								color = accent
							}
							connector := ui.Box(c.Key("connector")).Absolute()
							if o.Vertical {
								connector.Top(diameter + t.spacing(6)).Bottom(-t.spacing(10)).Start(diameter/2 - 1).Width(2)
							} else {
								// The absolute track spans adjacent equal-width items. Padding
								// leaves the same clearance around both circles as Nuxt UI.
								connector.StartPercent(50).EndPercent(-50).Top(diameter/2 - 2).Height(2).PaddingX(connectorInset)
							}
							connector.Children(func() {
								separator := ui.Box(c).Fill()
								opacity := float32(1)
								if it.Disabled || o.Disabled {
									opacity = .75
								}
								finish(t, ComponentStepper, SeparatorSlot, separator, PartStyle{Background: Set(color), Radius: Set(float32(999)), Opacity: Set(opacity)}, d)
							})
						}
					})
					wrapper := ui.Column(c.Key("wrapper")).MinWidth(0).Margin(wrapperGap, 0, 0, 0)
					if o.Vertical {
						wrapper.Grow(1)
					}
					finish(t, ComponentStepper, Body, wrapper, PartStyle{FontSize: Set(textSize), LineHeight: Set(lineHeight)}, d)
					wrapper.Children(func() {
						if it.Title != "" {
							finish(t, ComponentStepper, Label, ui.Text(c.Key("title"), it.Title), PartStyle{Foreground: Set(t.Text), Weight: Set(500)}, d)
						}
						if it.Description != "" {
							finish(t, ComponentStepper, Description, ui.Text(c.Key("description"), it.Description), PartStyle{Foreground: Set(t.Muted)}, d)
						}
					})
				})
			}
		})
		if current >= 0 && items[current].Content != nil {
			it := items[current]
			panel := ui.Column(c.Key("content:" + it.ID)).MinWidth(0).Bind(&memory.body)
			if o.Vertical {
				panel.Grow(1)
			} else {
				panel.FillWidth()
			}
			finish(t, ComponentStepper, Panel, panel, PartStyle{Gap: Set(t.spacing(12))}, o.Decoration)
			panel.Children(func() { it.Content(c, t) })
		}
	})
	if restore && !o.Disabled {
		index := current
		if index < 0 || items[index].Disabled {
			index = -1
			for i, it := range items {
				if !it.Disabled && (!valueOr(o.Linear, true) || i <= current+1) {
					index = i
					break
				}
			}
		}
		if index >= 0 {
			memory.triggers[items[index].ID].Focus()
		}
	}
	memory.active = *value
	return root
}

type TimelineItem struct {
	ID, Title, Description, Time string
	Icon                         *ui.SVG
	Avatar                       *AvatarOptions
	Color                        Color
	UI                           Styles
	Content                      func(*ui.Context, Theme)
}
type TimelineOptions struct {
	Label      string
	Horizontal bool
	Color      Color
	Size       Size
	Current    Optional[int] // Zero-based step; unset means no active/completed steps.
	CurrentID  string        // Takes precedence over Current when non-empty.
	Reverse    bool          // Changes progress interpretation, not item order.
	OnSelect   func(string)
	Decoration
}

func Timeline(c *ui.Context, t Theme, items []TimelineItem, o TimelineOptions) ui.Element {
	seen := map[string]bool{}
	for _, it := range items {
		checkID(seen, it.ID, "Timeline")
	}
	config := selected(defaults(t, ComponentTimeline, ComponentTheme{Color: Primary, Size: MD}), ComponentTheme{Color: o.Color, Size: o.Size})
	current := -1
	if o.Current.Set {
		current = o.Current.Value
		if o.Reverse {
			current = len(items) - 1 - current
		}
	}
	if o.CurrentID != "" {
		current = -1
		for i, it := range items {
			if it.ID == o.CurrentID {
				current = i
				break
			}
		}
	}
	if current < 0 || current >= len(items) {
		current = -1
	}
	root := ui.Column(c).MinWidth(0).Role(ui.RoleList).Label(o.Label)
	if o.Horizontal {
		root.Row().AlignItems(ui.Stretch)
	}
	finish(t, ComponentTimeline, Root, root, PartStyle{Gap: Set(t.spacing(6)), Font: Set(t.Font), FontSize: Set(t.text(14)), LineHeight: Set(float32(20.0 / 14)), Foreground: Set(t.Text)}, o.Decoration)
	root.Children(func() {
		for i, it := range items {
			d := o.Decoration
			d.UI = mergeStyles(o.UI, it.UI)
			color := it.Color
			if color == ColorDefault {
				color = config.Color
			}
			accent := t.color(color)
			if color == Neutral {
				accent = t.InvertedBackground
			}
			active := current == i
			completed := current >= 0 && ((!o.Reverse && i < current) || (o.Reverse && i > current))
			row := ui.Row(c.Key(it.ID)).MinWidth(0).AlignItems(ui.Stretch).Role(ui.RoleListItem).Label(it.Title)
			if o.OnSelect != nil {
				row.Focusable().FocusRing(true).OnClick(func() { o.OnSelect(it.ID) })
			}
			if o.Horizontal {
				row.Column().Grow(1).Basis(0)
			}
			finish(t, ComponentTimeline, Item, row, PartStyle{Gap: Set(t.spacing(12))}, d)
			row.Children(func() {
				container := ui.Column(c.Key("container")).AlignItems(ui.Center).Shrink(0)
				if o.Horizontal {
					container.Row().FillWidth()
				}
				finish(t, ComponentTimeline, Header, container, PartStyle{Gap: Set(t.spacing(6))}, d)
				container.Children(func() {
					avatar := AvatarOptions{Size: config.Size, Icon: it.Icon}
					if it.Avatar != nil {
						avatar = *it.Avatar
						if avatar.Size == SizeDefault {
							avatar.Size = config.Size
						}
						if avatar.Icon == nil {
							avatar.Icon = it.Icon
						}
					}
					marker := Avatar(c.Key("indicator"), t, avatar)
					bg, fg := t.Elevated, t.Muted
					if active || completed {
						bg, fg = accent, t.Inverted
					}
					finish(t, ComponentTimeline, Indicator, marker, PartStyle{Background: Set(bg), Foreground: Set(fg)}, d)
					if i < len(items)-1 {
						separator := ui.Box(c.Key("separator")).Grow(1).Radius(999)
						if o.Horizontal {
							separator.Height(2)
						} else {
							separator.Width(2)
						}
						bg := t.Elevated
						if completed || (o.Reverse && active) {
							bg = accent
						}
						finish(t, ComponentTimeline, SeparatorSlot, separator, PartStyle{Background: Set(bg)}, d)
					}
				})
				body := ui.Column(c.Key("body")).Grow(1).MinWidth(0)
				pad := t.spacing([5]float32{22, 24, 26, 28, 30}[sizeIndex(config.Size)])
				if o.Horizontal {
					body.PaddingEnd(pad)
				} else {
					body.Margin(t.spacing([5]float32{2, 4, 6, 8, 10}[sizeIndex(config.Size)]), 0, 0, 0).Padding(0, 0, pad, 0)
				}
				finish(t, ComponentTimeline, Body, body, PartStyle{}, d)
				body.Children(func() {
					if it.Time != "" {
						finish(t, ComponentTimeline, DateSlot, ui.Text(c.Key("date"), it.Time), PartStyle{Foreground: Set(t.Dimmed), FontSize: Set(t.text(12)), LineHeight: Set(float32(20.0 / 12))}, d)
					}
					if it.Title != "" {
						finish(t, ComponentTimeline, Label, ui.Text(c.Key("title"), it.Title), PartStyle{Foreground: Set(t.Highlighted), Weight: Set(500), FontSize: Set(t.text(14)), LineHeight: Set(float32(20.0 / 14))}, d)
					}
					if it.Description != "" {
						finish(t, ComponentTimeline, Description, ui.Text(c.Key("description"), it.Description), PartStyle{Foreground: Set(t.Muted), FontSize: Set(t.text(14)), LineHeight: Set(float32(20.0 / 14))}, d)
					}
					if it.Content != nil {
						it.Content(c, t)
					}
				})
			})
		}
	})
	return root
}
