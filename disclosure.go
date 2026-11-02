package nui

import (
	"slices"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

type CollapsibleOptions struct {
	Label                 string
	Icon                  *ui.SVG
	Disabled, KeepMounted bool
	OnChange              func(bool)
	Decoration
}

func Collapsible(c *ui.Context, t Theme, open *bool, o CollapsibleOptions, content func(Theme)) ui.Element {
	if open == nil {
		panic("nui: Collapsible requires open state")
	}
	root := ui.Column(c).MinWidth(0)
	finish(t, ComponentCollapsible, Root, root, PartStyle{Gap: Set(t.spacing(8))}, o.Decoration)
	root.Children(func() {
		p := ui.CollapsibleBase(c, open)
		disclosureTrigger(c, t, p.Trigger, p.Progress(), o, ComponentCollapsible)
		if p.Trigger.Changed() {
			root.SetChanged()
		}
		if o.OnChange != nil {
			root.OnChange(func() { o.OnChange(*open) })
		}
		build := func() {
			if content != nil {
				content(t)
			}
		}
		var body ui.Element
		if o.KeepMounted {
			body = ui.Column(c).Hidden(!*open)
			body.Children(build)
		} else {
			body = p.Panel(build)
		}
		finish(t, ComponentCollapsible, Body, body, PartStyle{Gap: Set(t.spacing(12)), PaddingY: Set(t.spacing(8))}, o.Decoration)
	})
	return root
}
func disclosureTrigger(c *ui.Context, t Theme, e ui.Element, progress float32, o CollapsibleOptions, id Component) {
	e.Label(o.Label).Disabled(o.Disabled).MinWidth(0).AlignItems(ui.Center).FocusRing(true)
	// Nuxt's accordion trigger: no surface or side padding, py-3.5, gap-1.5,
	// and 20 DIP icons in the text color.
	finish(t, id, Trigger, e, PartStyle{Background: Set(ui.Transparent), PaddingY: Set(t.spacing(14)), PaddingX: Set(float32(0)), Gap: Set(t.spacing(6)), Radius: Set(t.radius(6)), Foreground: Set(t.Text)}, o.Decoration)
	e.Children(func() {
		if o.Icon != nil {
			finish(t, id, Leading, ui.Icon(c, o.Icon).Size(20, 20).Shrink(0), PartStyle{}, o.Decoration)
		}
		finish(t, id, Label, ui.Text(c, o.Label).Grow(1).MinWidth(0), PartStyle{Weight: Set(500)}, o.Decoration)
		finish(t, id, Trailing, ui.Icon(c, lucide.ChevronDown()).Size(20, 20).Rotate(180*progress).Shrink(0), PartStyle{}, o.Decoration)
	})
}

type AccordionItem struct {
	ID, Label string
	Icon      *ui.SVG
	Disabled  bool
	Content   func(*ui.Context, Theme)
}
type AccordionOptions struct {
	Multiple, Disabled, KeepMounted bool
	Collapsible                     Optional[bool]
	OnChange                        func()
	Decoration
}
type accordionMemory struct{ handles map[string]*ui.Handle }

func Accordion(c *ui.Context, t Theme, open *[]string, items []AccordionItem, o AccordionOptions) ui.Element {
	if open == nil {
		panic("nui: Accordion requires open IDs")
	}
	ids := make([]string, 0, len(items))
	seen := map[string]bool{}
	for _, it := range items {
		checkID(seen, it.ID, "Accordion")
		if !it.Disabled {
			ids = append(ids, it.ID)
		}
	}
	root := ui.Column(c).MinWidth(0).Disabled(o.Disabled)
	memory := ui.Local(root, "accordion", func() accordionMemory { return accordionMemory{handles: map[string]*ui.Handle{}} })
	for id := range memory.handles {
		if !seen[id] {
			delete(memory.handles, id)
		}
	}
	for _, id := range ids {
		if memory.handles[id] == nil {
			memory.handles[id] = &ui.Handle{}
		}
	}
	if !o.Multiple && len(*open) > 1 {
		*open = append([]string(nil), (*open)[:1]...)
	}
	finish(t, ComponentAccordion, Root, root, PartStyle{}, o.Decoration)
	root.Children(func() {
		for i, it := range items {
			row := ui.Column(c.Key(it.ID)).MinWidth(0)
			if i < len(items)-1 {
				row.BorderWidth(0, 0, 1, 0).BorderColor(t.Border)
			}
			finish(t, ComponentAccordion, Item, row, PartStyle{}, o.Decoration)
			row.Children(func() {
				expanded := slices.Contains(*open, it.ID)
				p := ui.CollapsibleBase(c, &expanded)
				opts := CollapsibleOptions{Label: it.Label, Icon: it.Icon, Disabled: o.Disabled || it.Disabled, Decoration: o.Decoration}
				disclosureTrigger(c, t, p.Trigger, p.Progress(), opts, ComponentAccordion)
				if !it.Disabled {
					p.Trigger.Bind(memory.handles[it.ID])
					index := slices.Index(ids, it.ID)
					for _, key := range []ui.Key{ui.KeyUp, ui.KeyDown, ui.KeyHome, ui.KeyEnd} {
						target := index
						switch key {
						case ui.KeyUp:
							target = (index - 1 + len(ids)) % len(ids)
						case ui.KeyDown:
							target = (index + 1) % len(ids)
						case ui.KeyHome:
							target = 0
						case ui.KeyEnd:
							target = len(ids) - 1
						}
						handle := memory.handles[ids[target]]
						p.Trigger.OnShortcut(0, key, func() { handle.Focus() })
					}
				}
				if p.Trigger.Changed() {
					if !o.Multiple && !valueOr(o.Collapsible, true) && !expanded && slices.Contains(*open, it.ID) {
						expanded = true
					} else {
						next := slices.Clone(*open)
						next = slices.DeleteFunc(next, func(id string) bool { return id == it.ID })
						if expanded {
							if !o.Multiple {
								next = nil
							}
							next = append(next, it.ID)
						}
						*open = next
						root.SetChanged()
					}
				}
				build := func() {
					if it.Content != nil {
						it.Content(c, t)
					}
				}
				var body ui.Element
				if o.KeepMounted {
					body = ui.Column(c).Hidden(!expanded)
					body.Children(build)
				} else {
					body = p.Panel(build)
				}
				// Nuxt's body: the default text color and only pb-3.5.
				body.Padding(0, 0, t.spacing(14), 0)
				finish(t, ComponentAccordion, Body, body, PartStyle{Gap: Set(t.spacing(12)), Foreground: Set(t.Text)}, o.Decoration)
			})
		}
	})
	if o.OnChange != nil {
		root.OnChange(o.OnChange)
	}
	return root
}
func checkID(seen map[string]bool, id, component string) {
	if id == "" || seen[id] {
		panic("nui: " + component + " requires unique non-empty IDs")
	}
	seen[id] = true
}
