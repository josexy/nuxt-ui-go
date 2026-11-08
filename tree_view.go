package nui

import (
	"slices"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

type TreeOptions struct {
	Label, EmptyText                    string
	Height                              float32
	Color                               Color
	Size                                Size
	Disabled, Multiple                  bool
	SelectionBehavior                   SelectionBehavior // Default: SelectionToggle.
	PropagateSelect, BubbleSelect, Loop bool
	// Nested defaults to true: draw ancestor guides and inset child rows.
	// Both modes use the same flat, virtualized native row model, not nested DOM.
	Nested                                    Optional[bool]
	ExpandOnClick                             Optional[bool] // Default true; Enter/Space only select.
	ExpandedIcon, CollapsedIcon, TrailingIcon *ui.SVG
	// Build replaces non-interactive row content, after the leading icon.
	Build                       func(*ui.Context, Theme, TreeItemState)
	BuildLeading, BuildTrailing func(*ui.Context, Theme, TreeItemState)
	OnSelect, OnToggle          func(*TreeEvent)
	OnChange                    func()         // Once when an interaction changes selection.
	OnExpandedChange            func([]string) // User expansion changes, as sorted stable IDs.
	OnLoad                      func(string, TreeTicket)
	Decoration
}

func Tree(c *ui.Context, t Theme, s *TreeState, o TreeOptions) ui.Element {
	if s == nil {
		panic("nui: Tree requires state")
	}
	if o.SelectionBehavior > SelectionReplace {
		panic("nui: invalid selection behavior")
	}
	if o.Label == "" {
		o.Label = t.field.Label
	}
	s.cancelHiddenLoads()
	disabled := o.Disabled || t.field.Disabled
	height := o.Height
	if height <= 0 {
		height = 320
	}
	root := ui.Column(c).MinWidth(0).MinHeight(0).Height(height).Disabled(disabled)
	disabled = root.IsDisabled()
	if !disabled && o.OnLoad != nil {
		for id := range s.Outline.Open.All() {
			n := s.nodes[id]
			if n != nil && !n.node.Disabled && !n.loaded && !n.loading && n.err == "" && s.expandedPath(id) {
				s.version++
				n.version, n.loading = s.version, true
				if s.loading == nil {
					s.loading = map[string]*treeRecord{}
				}
				s.loading[id] = n
				o.OnLoad(id, TreeTicket{s.identity, id, n.version})
			}
		}
	}
	s.visibleRows()
	if o.Build != nil || o.BuildLeading != nil || o.BuildTrailing != nil || (o.Multiple && (o.BubbleSelect || o.PropagateSelect)) {
		s.selectionSummary(o)
	}
	_, ok := s.positions[s.Active]
	if !ok || s.nodes[s.Active].node.Disabled {
		at := s.next(-1, 1, false)
		if at >= 0 {
			s.Active = s.rows[at].id
		}
	}
	notify := func(changed bool) {
		if changed {
			s.summaryDirty = true
			if o.OnChange != nil {
				o.OnChange()
			}
		}
	}
	selectItem := func(id string, mods ui.Modifiers) {
		n := s.nodes[id]
		if disabled || n == nil || n.node.Disabled {
			return
		}
		event := TreeEvent{ID: id, Selected: s.Selection.Has(id), Expanded: s.Outline.Open.Has(id)}
		handler := n.node.OnSelect
		if handler == nil {
			handler = o.OnSelect
		}
		if handler != nil {
			handler(&event)
		}
		if event.prevented || s.nodes[id] != n {
			return
		}
		propagate := o.Multiple && o.PropagateSelect
		toggle := o.SelectionBehavior == SelectionToggle || mods&(ui.Ctrl|ui.Super) != 0
		on := !toggle || !s.Selection.Has(id)
		changed := false
		from, anchorOK := s.positions[s.anchor]
		to, visible := s.positions[id]
		rangeSelection := o.Multiple && mods&ui.Shift != 0 && anchorOK && visible
		var before map[string]bool
		if rangeSelection {
			if o.SelectionBehavior == SelectionReplace {
				before = make(map[string]bool, s.Selection.Len())
				for key := range s.Selection.All() {
					before[key] = true
				}
				changed = s.Selection.Len() > 0
				s.Selection.Clear()
			}
			changed = s.selectRows(s.rows[min(from, to):max(from, to)+1], propagate) || changed
		} else {
			if !o.Multiple || !toggle {
				if s.Selection.Len() != 1 || !s.Selection.Has(id) {
					changed = s.Selection.Len() > 0
					s.Selection.Clear()
				}
			}
			changed = s.changeSelection(id, on, propagate) || changed
			s.anchor = id
		}
		if o.Multiple && o.BubbleSelect {
			if rangeSelection {
				changed = s.bubble() || changed
			} else {
				changed = s.bubbleAncestors(id) || changed
			}
		}
		if before != nil && len(before) == s.Selection.Len() {
			changed = false
			for key := range s.Selection.All() {
				if !before[key] {
					changed = true
					break
				}
			}
		}
		notify(changed)
	}
	toggleItem := func(id string) {
		n := s.nodes[id]
		if disabled || n == nil || n.node.Disabled || n.children == nil {
			return
		}
		event := TreeEvent{ID: id, Selected: s.Selection.Has(id), Expanded: s.Outline.Open.Has(id)}
		handler := n.node.OnToggle
		if handler == nil {
			handler = o.OnToggle
		}
		if handler != nil {
			handler(&event)
		}
		if !event.prevented && s.nodes[id] == n {
			s.SetExpanded(id, !event.Expanded)
			if o.OnExpandedChange != nil && s.Outline.Open.Has(id) != event.Expanded {
				o.OnExpandedChange(slices.Sorted(s.Outline.Open.All()))
			}
		}
	}
	config := selected(defaults(t, ComponentTree, ComponentTheme{Color: Primary, Size: MD}), ComponentTheme{Color: o.Color, Size: o.Size})
	m := optionRecipe(t, config.Size)
	pad := t.spacing([5]float32{8, 10, 10, 12, 12}[sizeIndex(config.Size)])
	indent := t.spacing([5]float32{22, 24, 26, 28, 30}[sizeIndex(config.Size)])
	if s.Outline.Indent > 0 {
		indent = s.Outline.Indent
	}
	nested := valueOr(o.Nested, true)
	s.RowBuilds = 0
	finish(t, ComponentTree, Root, root, PartStyle{}, o.Decoration)
	root.Children(func() {
		if len(s.roots) == 0 {
			text := o.EmptyText
			if text == "" {
				text = emptyMessage(c, "noData", "noItems")
			}
			finish(t, ComponentTree, EmptySlot, ui.Text(c, text).Padding(12).Role(ui.RoleStatus), PartStyle{Foreground: Set(t.Muted)}, o.Decoration)
		}
		// Interaction belongs to Tree. List supplies measured virtualization,
		// stable row identities, scroll anchoring and persistent focus only.
		s.Outline.List.Selection, s.Outline.List.Selected, s.Outline.List.Label = nil, nil, nil
		list := ui.List(c.Key("tree"), &s.Outline.List, len(s.rows)).ItemKey(func(i int) any { return s.rows[i].id }).Fill().Focusable().FocusRing(false).Role(ui.RoleTree).Label(o.Label).InputMethod(false)
		finish(t, ComponentTree, List, list, PartStyle{}, o.Decoration)
		move := func(i int, mods ui.Modifiers) {
			if i < 0 || i >= len(s.rows) || s.nodes[s.rows[i].id].node.Disabled {
				return
			}
			if mods&ui.Shift != 0 && s.anchor == "" {
				s.anchor = s.Active
			}
			s.Active = s.rows[i].id
			s.Outline.List.ScrollIntoView(i)
			if o.Multiple && mods&ui.Shift != 0 {
				selectItem(s.Active, mods)
			}
		}
		back, forth := ui.KeyLeft, ui.KeyRight
		if c.Direction() == ui.RightToLeft {
			back, forth = forth, back
		}
		list.HandleInput(func(ev ui.InputEvent) bool {
			if disabled {
				return false
			}
			at, found := s.positions[s.Active]
			if !found {
				at = -1
			}
			if ev.Kind == ui.InputText && ev.Text != "" && ev.Mods&(ui.Ctrl|ui.Super|ui.Alt) == 0 {
				s.keyboard = true
				if ev.Time.Sub(s.typed) > time.Second {
					s.prefix = ""
				}
				s.typed = ev.Time
				s.prefix += strings.ToLower(ev.Text)
				if i := typeaheadIndex(s.prefix, at, len(s.rows), func(i int) (string, bool) {
					n := s.nodes[s.rows[i].id]
					return n.search, !n.node.Disabled
				}); i >= 0 {
					move(i, 0)
				}
				return true
			}
			if ev.Kind != ui.InputKeyDown {
				return false
			}
			s.keyboard = true
			if o.Multiple && ev.Key == ui.KeyA && (ev.Mods == ui.Ctrl || ev.Mods == ui.Super) {
				changed := s.selectRows(s.rows, o.PropagateSelect)
				if o.BubbleSelect {
					changed = s.bubble() || changed
				}
				notify(changed)
				return true
			}
			if ev.Mods & ^(ui.Shift|ui.Ctrl|ui.Super) != 0 {
				return false
			}
			switch ev.Key {
			case ui.KeyDown:
				move(s.next(at, 1, o.Loop), ev.Mods)
			case ui.KeyUp:
				move(s.next(at, -1, o.Loop), ev.Mods)
			case ui.KeyHome:
				move(s.next(-1, 1, false), ev.Mods)
			case ui.KeyEnd:
				move(s.next(len(s.rows), -1, false), ev.Mods)
			case ui.KeyPageUp, ui.KeyPageDown:
				delta := max(1, int(height/(m.Line+2*m.Pad))-1)
				if ev.Key == ui.KeyPageUp {
					delta = -delta
				}
				target := max(0, min(len(s.rows)-1, at+delta))
				if delta > 0 {
					move(s.next(target-1, 1, false), ev.Mods)
				} else {
					move(s.next(target+1, -1, false), ev.Mods)
				}
			case ui.KeyEnter, ui.KeySpace:
				selectItem(s.Active, ev.Mods)
			case forth:
				if at >= 0 {
					if !s.Outline.Open.Has(s.Active) {
						toggleItem(s.Active)
					} else {
						for i := at + 1; i < len(s.rows) && s.rows[i].depth > s.rows[at].depth; i++ {
							if s.rows[i].depth == s.rows[at].depth+1 && !s.nodes[s.rows[i].id].node.Disabled {
								move(i, 0)
								break
							}
						}
					}
				}
			case back:
				if at >= 0 {
					if s.Outline.Open.Has(s.Active) {
						toggleItem(s.Active)
					} else if p := s.nodes[s.Active].parent; p != "" {
						move(s.positions[p], 0)
					}
				}
			default:
				return false
			}
			return true
		})
		list.Rows(func(r ui.ListRow) {
			s.RowBuilds++
			entry := s.rows[r.Index]
			id, n := entry.id, s.nodes[entry.id]
			chosen, expanded := s.Selection.Has(id), s.Outline.Open.Has(id)
			if nested && entry.depth > 0 {
				treeGuides(r.Context.Key("guides"), t, entry.depth, indent, o.Decoration)
			}
			row := ui.Column(r.Context.Key("item")).MinWidth(0).Role(ui.RoleTreeItem).Level(entry.depth + 1).Label(n.node.Label).Selected(chosen).Background(ui.Transparent).Disabled(n.node.Disabled).KeepFocus()
			if nested {
				row.MarginStart(float32(entry.depth) * indent)
			}
			if n.children != nil {
				row.Expanded(expanded)
			}
			if id == s.Active {
				list.ActiveDescendant(row)
			}
			mods := row.ClickModifiers()
			row.OnClick(func() {
				s.keyboard = false
				s.Active = id
				s.Outline.List.Focus(c)
				selectItem(id, mods)
				if s.nodes[id] == n && valueOr(o.ExpandOnClick, true) {
					toggleItem(id)
				}
			})
			style := optionItemStyle(t, m, row.Hovered())
			style.PaddingX = Set(pad)
			if chosen {
				style.Background, style.Foreground = Set(t.Elevated), Set(t.color(config.Color))
			}
			if id == s.Active && !disabled && (list.FocusVisible() || s.keyboard && list.FocusWithin()) {
				style.BorderWidth, style.BorderColor = Set(float32(2)), Set(inputFocusColor(t, config.Color))
			}
			row.Children(func() {
				// Nuxt's before:inset-y-px leaves a 2 DIP visual gap between
				// adjacent highlights. Keep the outer row's full hit area and
				// height, and inset only its styled surface. Guides stay continuous.
				style.PaddingY = Set(max(0, m.Pad-1))
				surface := ui.Row(r.Context.Key("surface")).MinWidth(0).AlignItems(ui.Center).Margin(1, 0).Role(ui.RoleNone)
				finishTreeItem(t, Item, surface, style, o.Decoration, n.node.Decoration)
				surface.Children(func() {
					if !nested && entry.depth > 0 {
						// The row's gap after this spacer is part of the indentation.
						ui.Box(r.Context.Key("indent")).Width(max(0, float32(entry.depth)*indent-m.Gap)).Shrink(0)
					}
					itemState := TreeItemState{ID: id, Label: n.node.Label, Level: entry.depth + 1, Selected: chosen, Expanded: expanded, Indeterminate: s.partial[id], Loading: n.loading}
					icon := n.node.Icon
					// Empty loaded branches retain expansion state, as Reka does,
					// but only nonempty or unresolved branches get fallback icons.
					hasContent := len(n.children) > 0 || !n.loaded
					if icon == nil && hasContent {
						icon = o.CollapsedIcon
						if expanded {
							icon = o.ExpandedIcon
						}
						if icon == nil {
							if expanded {
								icon = lucide.FolderOpen()
							} else {
								icon = lucide.Folder()
							}
						}
					}
					if o.BuildLeading != nil {
						leading := ui.Row(r.Context.Key("leading")).Shrink(0).AlignItems(ui.Center)
						finishTreeItem(t, Leading, leading, PartStyle{}, o.Decoration, n.node.Decoration)
						leading.Children(func() { o.BuildLeading(r.Context, t, itemState) })
					} else if icon != nil {
						finishTreeItem(t, Leading, ui.Icon(r.Context.Key("leading"), icon).Size(m.Icon, m.Icon).Shrink(0), PartStyle{}, o.Decoration, n.node.Decoration)
					}
					if o.Build != nil {
						o.Build(r.Context, t, itemState)
					} else {
						finishTreeItem(t, Label, ui.Text(r.Context.Key("label"), n.node.Label).SingleLine().Grow(1).MinWidth(0), PartStyle{}, o.Decoration, n.node.Decoration)
					}
					if n.loading {
						ui.Text(r.Context, message(c, "loading")).FontSize(t.text(12)).Role(ui.RoleStatus)
					}
					if n.err != "" {
						Button(r.Context.Key("retry"), t, message(c, "retry"), ButtonOptions{AccessibleLabel: namedMessage(c, "retryNamed", n.node.Label), Variant: Ghost, Size: XS}).Description(n.err).OnClick(func() { s.Retry(id) })
					}
					trailing := n.node.TrailingIcon
					if trailing == nil && hasContent {
						trailing = o.TrailingIcon
						if trailing == nil {
							trailing = lucide.ChevronDown()
						}
					}
					if o.BuildTrailing != nil {
						trailing := ui.Row(r.Context.Key("trailing")).Shrink(0).AlignItems(ui.Center)
						finishTreeItem(t, Trailing, trailing, PartStyle{}, o.Decoration, n.node.Decoration)
						trailing.Children(func() { o.BuildTrailing(r.Context, t, itemState) })
					} else if trailing != nil {
						arrow := ui.Icon(r.Context.Key("trailing"), trailing).Size(m.Icon, m.Icon).Shrink(0)
						if n.children != nil {
							if expanded {
								arrow.Rotate(180)
							}
							arrow.KeepFocus().OnClick(func() { s.keyboard = false; s.Active = id; s.Outline.List.Focus(c); toggleItem(id) })
						}
						finishTreeItem(t, Trailing, arrow, PartStyle{}, o.Decoration, n.node.Decoration)
					}
				})
			})
		})
	})
	return root
}

// Each virtual row paints only the segments of its ancestors' group borders
// that cross that row. This works even when the ancestors are offscreen, without
// constructing offscreen groups or allocating a widget for each ancestor.
func treeGuides(c *ui.Context, t Theme, depth int, indent float32, decoration Decoration) {
	style := mergeStyle(PartStyle{BorderWidth: Set(float32(1)), BorderColor: Set(t.Border)}, t.components[ComponentTree].UI[Indicator])
	style = mergeStyle(style, decoration.UI[Indicator])
	width, color := max(0, style.BorderWidth.Value), style.BorderColor.Value
	style.BorderWidth = Set(float32(0)) // The border is a set of lines, not a box.
	guide := ui.Box(c).Absolute().Left(0).Right(0).Top(0).Bottom(0).Role(ui.RoleNone).PassThrough()
	rtl := c.Direction() == ui.RightToLeft
	inset := min(t.spacing(6), indent/2)
	guide.Draw(func(p *ui.Painter, r ui.Rect) {
		if width <= 0 {
			return
		}
		for level := 0; level < depth; level++ {
			offset := float32(level+1)*indent - inset
			if offset >= r.W {
				break
			} // Deep trees never draw beyond the viewport.
			x := r.X + offset
			if rtl {
				x = r.X + r.W - offset - width
			}
			p.Fill(ui.Rect{X: x, Y: r.Y, W: width, H: r.H}, color, 0)
		}
	})
	style.apply(guide)
	if decoration.Style != nil {
		decoration.Style(Indicator, guide)
	}
}
