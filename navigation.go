package nui

import (
	"fmt"
	"slices"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

type TabItem struct {
	ID, Label string
	Icon      *ui.SVG
	Disabled  bool
	Content   func(*ui.Context, Theme)
}
type TabsOptions struct {
	Label                            string
	Color                            Color
	Size                             Size
	Underline, Vertical, KeepMounted bool
	ActivationMode                   ActivationMode
	OnChange                         func(string)
	Decoration
}

type ActivationMode uint8

const (
	ActivationAutomatic ActivationMode = iota
	ActivationManual
)

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
	if o.Vertical {
		root.Row().AlignItems(ui.Start)
	}
	finish(t, ComponentTabs, Root, root, PartStyle{Gap: Set(t.spacing(8))}, o.Decoration)
	root.Children(func() {
		parts := ui.TabsBase(c, &index, len(enabled))
		list := parts.List.MinWidth(0).Label(o.Label)
		if o.Vertical {
			list.Column().FocusGroup(ui.Vertical)
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
				var trigger ui.Element
				if it.Disabled {
					trigger = ui.Row(c.Key(it.ID)).Role(ui.RoleTab).Disabled(true)
				} else if o.ActivationMode == ActivationManual {
					trigger = ui.ButtonBase(c.Key(it.ID)).Role(ui.RoleTab).Selected(it.ID == *value).OnClick(func() {
						if *value != it.ID {
							*value = it.ID
							root.SetChanged()
							if o.OnChange != nil {
								o.OnChange(it.ID)
							}
						}
					})
				} else {
					c.Key(it.ID)
					trigger = parts.Tab(slices.Index(enabled, it.ID))
				}
				active := index == slices.Index(enabled, it.ID) && !it.Disabled
				trigger.Label(it.Label).MinWidth(0).AlignItems(ui.Center).FocusRing(true)
				background := ui.Transparent
				foreground := t.Muted
				if active {
					foreground = t.color(config.Color)
					if !o.Underline {
						background = t.color(config.Color)
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
				finish(t, ComponentTabs, Trigger, trigger, PartStyle{Background: Set(background), Foreground: Set(foreground), Radius: Set(t.radius(6)), PaddingX: Set(metrics.X), PaddingY: Set(metrics.Y), Gap: Set(gap), FontSize: Set(metrics.Font), Weight: Set(500)}, o.Decoration)
				trigger.Children(func() {
					if it.Icon != nil {
						finish(t, ComponentTabs, Leading, ui.Icon(c, it.Icon).Size(metrics.Icon, metrics.Icon), PartStyle{}, o.Decoration)
					}
					finish(t, ComponentTabs, Label, ui.Text(c, it.Label).SingleLine().MinWidth(0), PartStyle{}, o.Decoration)
					if o.Underline && active {
						// Past the list's padding, on its border line.
						line := ui.Box(c.Key("indicator")).Absolute().IgnoreInput(true).Background(t.color(config.Color))
						if o.Vertical {
							line.Top(0).Bottom(0).Width(1).Start(-listStyle.PaddingX.Value - 1)
						} else {
							line.Left(0).Right(0).Height(1).Bottom(-listStyle.PaddingY.Value - 1)
						}
					}
				})
			}
		})
		if list.Changed() && index >= 0 && index < len(enabled) {
			*value = enabled[index]
			root.SetChanged()
		}
		for _, it := range items {
			active := it.ID == *value && !it.Disabled
			if it.Content == nil || (!active && !o.KeepMounted) {
				continue
			}
			body := ui.Column(c.Key("content:" + it.ID)).MinWidth(0).Role(ui.RoleGroup).Label(namedMessage(c, "contentNamed", it.Label)).Hidden(!active)
			if o.Vertical {
				body.Grow(1)
			}
			finish(t, ComponentTabs, Body, body, PartStyle{Gap: Set(t.spacing(12))}, o.Decoration)
			body.Children(func() { it.Content(c, t) })
		}
	})
	if o.OnChange != nil && o.ActivationMode != ActivationManual {
		root.OnChange(func() { o.OnChange(*value) })
	}
	return root
}

type BreadcrumbItem struct {
	ID, Label string
	Icon      *ui.SVG
	Disabled  bool
	OnSelect  func()
}
type BreadcrumbOptions struct {
	Label    string
	MaxItems int
	Decoration
}

func Breadcrumb(c *ui.Context, t Theme, items []BreadcrumbItem, o BreadcrumbOptions) ui.Element {
	seen := map[string]bool{}
	for _, it := range items {
		checkID(seen, it.ID, "Breadcrumb")
	}
	root := ui.Row(c).Wrap().MinWidth(0).AlignItems(ui.Center).Role(ui.RoleGroup).Label(o.Label)
	finish(t, ComponentBreadcrumb, Root, root, PartStyle{Gap: Set(t.spacing(6))}, o.Decoration)
	m := buttonMetrics(t, MD)
	current := t.color(defaults(t, ComponentBreadcrumb, ComponentTheme{Color: Primary}).Color)
	root.Children(func() {
		for i, it := range items {
			if o.MaxItems >= 2 && len(items) > o.MaxItems && i > 0 && i < len(items)-o.MaxItems+1 {
				if i == 1 {
					ui.Icon(c, lucide.Ellipsis()).Size(m.Icon, m.Icon).TextColor(t.Muted)
				}
				continue
			}
			if i > 0 {
				separator := lucide.ChevronRight()
				if c.Direction() == ui.RightToLeft {
					separator = lucide.ChevronLeft()
				}
				ui.Icon(c, separator).Size(m.Icon, m.Icon).Shrink(0).TextColor(t.Muted)
			}
			if i == len(items)-1 {
				item := ui.Row(c.Key(it.ID)).MinWidth(0).AlignItems(ui.Center).Gap(m.Gap)
				item.Children(func() {
					if it.Icon != nil {
						ui.Icon(c, it.Icon).Size(m.Icon, m.Icon).Shrink(0)
					}
					ui.Text(c, it.Label).SingleLine().MinWidth(0).Description(message(c, "currentPage"))
				})
				finish(t, ComponentBreadcrumb, Item, item, PartStyle{Foreground: Set(current), Weight: Set(600), FontSize: Set(m.Font), LineHeight: Set(m.Line / m.Font)}, o.Decoration)
			} else {
				e := Button(c.Key(it.ID), t, it.Label, ButtonOptions{Variant: LinkVariant, Color: Neutral, Leading: it.Icon, Disabled: it.Disabled, Size: MD, UI: Styles{Root: {PaddingX: Set(float32(0)), PaddingY: Set(float32(0))}}}).Role(ui.RoleLink).OnClick(it.OnSelect)
				finish(t, ComponentBreadcrumb, Item, e, PartStyle{}, o.Decoration)
			}
		}
	})
	return root
}

type PaginationOptions struct {
	Label                     string
	Total, PageSize, Siblings int
	// SiblingCount overrides legacy Siblings and can explicitly request zero.
	SiblingCount            Optional[int]
	ShowControls, ShowEdges Optional[bool]
	Color, ActiveColor      Color
	Variant, ActiveVariant  Variant
	Size                    Size
	Disabled                bool
	OnChange                func(int)
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
	finish(t, ComponentPagination, Root, root, PartStyle{Gap: Set(t.spacing(4))}, o.Decoration)
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
	icons := map[string]*ui.SVG{"first": first, "previous": previous, "next": next, "last": last}
	root.Children(func() {
		pageButton := func(key, label string, next int, disabled, active bool) {
			variant := config.Variant
			color := config.Color
			if active {
				variant = activeConfig.Variant
				color = activeConfig.Color
			}
			opts := ButtonOptions{Variant: variant, Color: color, Size: config.Size, Disabled: disabled || o.Disabled, Square: true, Leading: icons[key]}
			if opts.Leading == nil {
				// Page numbers are square buttons with a centered label at least 20 wide.
				opts.Style = func(root, label, leading, trailing ui.Element) { label.MinWidth(t.text(20)).TextAlign(ui.Center) }
			} else {
				label = ""
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
				style.apply(gap)
				gap.Children(func() { ui.Icon(c, lucide.Ellipsis()).Size(m.Icon, m.Icon) })
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
	root := ui.Row(c).MinWidth(0).AlignItems(ui.Start).Role(ui.RoleGroup).Label(o.Label).FocusGroup(ui.Horizontal).Disabled(o.Disabled)
	if o.Vertical {
		root.Column().AlignItems(ui.Stretch).FocusGroup(ui.Vertical)
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
	gap := float32(0)
	if o.Vertical {
		gap = t.spacing(16)
	}
	finish(t, ComponentStepper, Root, root, PartStyle{Gap: Set(gap)}, o.Decoration)
	root.Children(func() {
		for i, it := range items {
			blocked := it.Disabled || o.Disabled || (valueOr(o.Linear, true) && i > current+1)
			item := ui.Column(c.Key(it.ID)).MinWidth(0)
			itemGap := float32(0)
			if o.Vertical {
				item.Row().AlignItems(ui.Stretch)
				itemGap = wrapperGap
			} else {
				item.Grow(1).Basis(0).TextAlign(ui.Center)
			}
			finish(t, ComponentStepper, Item, item, PartStyle{Gap: Set(itemGap)}, o.Decoration)
			item.Children(func() {
				bg, fg := t.Elevated, t.Muted
				if i <= current {
					bg = t.color(config.Color)
					fg = t.Inverted
				}
				header := ui.Column(c.Key("header")).MinWidth(0).AlignItems(ui.Center)
				if o.Vertical {
					header.Width(diameter).Shrink(0)
				}
				finish(t, ComponentStepper, Header, header, PartStyle{}, o.Decoration)
				header.Children(func() {
					trigger := ui.ButtonBase(c.Key("trigger")).Size(diameter, diameter).Disabled(blocked).Role(ui.RoleButton).Label(it.Title).Description(it.Description).Selected(i == current).FocusRing(true)
					finish(t, ComponentStepper, Trigger, trigger, PartStyle{Background: Set(bg), Foreground: Set(fg), Radius: Set(float32(999)), FontSize: Set(fontSize), Weight: Set(600)}, o.Decoration)
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
						finish(t, ComponentStepper, Indicator, marker, PartStyle{}, o.Decoration)
						marker.Children(func() {
							if it.Icon != nil {
								finish(t, ComponentStepper, Leading, ui.Icon(c, it.Icon).Size(iconSize, iconSize), PartStyle{}, o.Decoration)
							} else {
								ui.Textf(c, "%d", i+1)
							}
						})
					})
					if i < len(items)-1 {
						color := t.Accented
						if i < current {
							color = t.color(config.Color)
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
							finish(t, ComponentStepper, SeparatorSlot, separator, PartStyle{Background: Set(color), Radius: Set(float32(999)), Opacity: Set(opacity)}, o.Decoration)
						})
					}
				})
				wrapper := ui.Column(c.Key("wrapper")).MinWidth(0).Margin(wrapperGap, 0, 0, 0)
				if o.Vertical {
					wrapper.Grow(1)
				}
				finish(t, ComponentStepper, Body, wrapper, PartStyle{FontSize: Set(textSize), LineHeight: Set(lineHeight)}, o.Decoration)
				wrapper.Children(func() {
					if it.Title != "" {
						finish(t, ComponentStepper, Label, ui.Text(c, it.Title), PartStyle{Foreground: Set(t.Text), Weight: Set(500)}, o.Decoration)
					}
					if it.Description != "" {
						finish(t, ComponentStepper, Description, ui.Text(c, it.Description), PartStyle{Foreground: Set(t.Muted)}, o.Decoration)
					}
				})
			})
		}
	})
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
	finish(t, ComponentTimeline, Root, root, PartStyle{Gap: Set(t.spacing(6))}, o.Decoration)
	root.Children(func() {
		for i, it := range items {
			d := o.Decoration
			d.UI = cloneStyles(o.UI)
			if d.UI == nil {
				d.UI = Styles{}
			}
			for slot, style := range it.UI {
				d.UI[slot] = mergeStyle(d.UI[slot], style)
			}
			color := it.Color
			if color == ColorDefault {
				color = config.Color
			}
			active := current == i
			completed := current >= 0 && ((!o.Reverse && i < current) || (o.Reverse && i > current))
			row := ui.Row(c.Key(it.ID)).MinWidth(0).AlignItems(ui.Stretch).Role(ui.RoleListItem).Label(it.Title)
			if o.OnSelect != nil {
				row.Focusable().OnClick(func() { o.OnSelect(it.ID) })
			}
			if o.Horizontal {
				row.Column().Grow(1)
			}
			finish(t, ComponentTimeline, Item, row, PartStyle{Gap: Set(t.spacing(12))}, d)
			row.Children(func() {
				container := ui.Column(c).AlignItems(ui.Center).Gap(t.spacing(6)).Shrink(0)
				if o.Horizontal {
					container.Row().FillWidth()
				}
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
					marker := Avatar(c, t, avatar)
					bg, fg := t.Elevated, t.Muted
					if active || completed {
						bg, fg = t.color(color), t.Inverted
					}
					finish(t, ComponentTimeline, Indicator, marker, PartStyle{Background: Set(bg), Foreground: Set(fg)}, d)
					if i < len(items)-1 {
						separator := ui.Box(c).Grow(1).Radius(999)
						if o.Horizontal {
							separator.Height(2)
						} else {
							separator.Width(2).MinHeight(t.spacing(12))
						}
						bg := t.Elevated
						if completed || (o.Reverse && active) {
							bg = t.color(color)
						}
						finish(t, ComponentTimeline, SeparatorSlot, separator, PartStyle{Background: Set(bg)}, d)
					}
				})
				body := ui.Column(c).Grow(1).MinWidth(0)
				pad := t.spacing([5]float32{22, 24, 26, 28, 30}[sizeIndex(config.Size)])
				if o.Horizontal {
					body.PaddingEnd(pad)
				} else {
					body.Padding(t.spacing([5]float32{2, 4, 6, 8, 10}[sizeIndex(config.Size)]), 0, pad, 0)
				}
				finish(t, ComponentTimeline, Body, body, PartStyle{Gap: Set(t.spacing(4))}, d)
				body.Children(func() {
					if it.Time != "" {
						ui.Text(c, it.Time).TextColor(t.Dimmed).FontSize(t.text(12))
					}
					finish(t, ComponentTimeline, Label, ui.Text(c, it.Title), PartStyle{Foreground: Set(t.Highlighted), Weight: Set(500), FontSize: Set(t.text(14))}, d)
					if it.Description != "" {
						finish(t, ComponentTimeline, Description, ui.Text(c, it.Description), PartStyle{Foreground: Set(t.Muted), FontSize: Set(t.text(14))}, d)
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
