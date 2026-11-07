package nui

import (
	"slices"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

type CollapsibleOptions struct {
	Label                 string
	Icon                  *ui.SVG
	TrailingIcon          *ui.SVG
	Disabled, KeepMounted bool
	OnChange              func(bool)
	Decoration
}

type disclosureMemory struct{ trigger, body ui.Handle }

func disclosureProgress(c *ui.Context, t Theme, p ui.CollapsibleParts, open bool) float32 {
	if !t.motion(c) {
		v := float32(0)
		if open {
			v = 1
		}
		p.Trigger.Jump("open", v)
		return v
	}
	return p.Progress()
}

// Keep the same panel scopes when mounting policy or motion preferences change.
func disclosureBody(c *ui.Context, t Theme, open, keep bool, progress float32, id Component, d Decoration, base PartStyle, bottom Optional[float32], handle *ui.Handle, build func()) ui.Element {
	if !keep && !open && progress == 0 {
		return ui.Element{}
	}
	clip := ui.Column(c.Key("panel")).Shrink(0)
	var body ui.Element
	clip.Children(func() {
		body = ui.Column(c.Key("body")).Shrink(0).Disabled(!open).Bind(handle)
		if bottom.Set {
			body.Padding(0, 0, bottom.Value, 0)
		}
		finish(t, id, Body, body, base, d)
		body.Children(build)
	})
	if keep {
		clip.Hidden(!open)
	} else if progress < 1 {
		clip.ClipY().Height(progress * body.Bounds().H)
	}
	return body
}

func Collapsible(c *ui.Context, t Theme, open *bool, o CollapsibleOptions, content func(Theme)) ui.Element {
	if open == nil {
		panic("nui: Collapsible requires open state")
	}
	root := ui.Column(c).MinWidth(0)
	memory := ui.Local(root, "disclosure", func() disclosureMemory { return disclosureMemory{} })
	finish(t, ComponentCollapsible, Root, root, PartStyle{Gap: Set(t.spacing(8)), Font: Set(t.Font), FontSize: Set(t.text(14)), Foreground: Set(t.Text)}, o.Decoration)
	root.Children(func() {
		p := ui.CollapsibleBase(c.Key("trigger"), open)
		p.Trigger.Bind(&memory.trigger)
		progress := disclosureProgress(c, t, p, *open)
		disclosureTrigger(c, t, p.Trigger, progress, o, ComponentCollapsible)
		if !*open && memory.body.FocusWithin(c) && !p.Trigger.IsDisabled() {
			memory.trigger.Focus()
		}
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
		disclosureBody(c, t, *open, o.KeepMounted, progress, ComponentCollapsible, o.Decoration, PartStyle{Gap: Set(t.spacing(12)), PaddingY: Set(t.spacing(8))}, Optional[float32]{}, &memory.body, build)
	})
	return root
}
func disclosureTrigger(c *ui.Context, t Theme, e ui.Element, progress float32, o CollapsibleOptions, id Component) {
	e.Label(o.Label).Disabled(o.Disabled).MinWidth(0).AlignItems(ui.Center).FocusRing(true)
	// Nuxt's accordion trigger: no surface or side padding, py-3.5, gap-1.5,
	// and 20 DIP icons in the text color.
	finish(t, id, Trigger, e, PartStyle{Background: Set(ui.Transparent), PaddingY: Set(t.spacing(14)), PaddingX: Set(float32(0)), Gap: Set(t.spacing(6)), Radius: Set(t.radius(6)), Foreground: Set(t.Text), Font: Set(t.Font), FontSize: Set(t.text(14)), LineHeight: Set(float32(20.0 / 14))}, o.Decoration)
	e.Children(func() {
		if o.Icon != nil {
			finish(t, id, Leading, ui.Icon(c.Key("leading"), o.Icon).Size(t.text(20), t.text(20)).Shrink(0), PartStyle{}, o.Decoration)
		}
		finish(t, id, Label, ui.Text(c.Key("label"), o.Label).Grow(1).MinWidth(0), PartStyle{Weight: Set(500)}, o.Decoration)
		icon := o.TrailingIcon
		if icon == nil {
			icon = lucide.ChevronDown()
		}
		finish(t, id, Trailing, ui.Icon(c.Key("trailing"), icon).Size(t.text(20), t.text(20)).Rotate(180*progress).Shrink(0), PartStyle{}, o.Decoration)
	})
}

type AccordionItem struct {
	ID, Label    string
	Icon         *ui.SVG
	TrailingIcon *ui.SVG
	Disabled     bool
	Content      func(*ui.Context, Theme)
	UI           Styles
}
type AccordionOptions struct {
	Multiple, Disabled, KeepMounted bool
	TrailingIcon                    *ui.SVG
	Collapsible                     Optional[bool]
	OnChange                        func()
	Decoration
}
type accordionMemory struct{ handles, bodies map[string]*ui.Handle }

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
	root := ui.Column(c).MinWidth(0)
	memory := ui.Local(root, "accordion", func() accordionMemory {
		return accordionMemory{handles: map[string]*ui.Handle{}, bodies: map[string]*ui.Handle{}}
	})
	restore := false
	for id := range memory.handles {
		if !slices.Contains(ids, id) && memory.handles[id].Focused(c) {
			restore = true
		}
		if !seen[id] {
			restore = restore || memory.bodies[id].FocusWithin(c)
			memory.handles[id].CancelFocus(c)
			memory.bodies[id].CancelFocus(c)
			delete(memory.handles, id)
			delete(memory.bodies, id)
		}
	}
	for id := range seen {
		if memory.handles[id] == nil {
			memory.handles[id] = &ui.Handle{}
			memory.bodies[id] = &ui.Handle{}
		}
	}
	if !o.Multiple && len(*open) > 1 {
		*open = append([]string(nil), (*open)[:1]...)
	}
	finish(t, ComponentAccordion, Root, root, PartStyle{Font: Set(t.Font), FontSize: Set(t.text(14)), Foreground: Set(t.Text)}, o.Decoration)
	root.Children(func() {
		for i, it := range items {
			d := o.Decoration
			d.UI = mergeStyles(o.UI, it.UI)
			row := ui.Column(c.Key(it.ID)).MinWidth(0)
			if i < len(items)-1 {
				row.BorderWidth(0, 0, 1, 0).BorderColor(t.Border)
			}
			finish(t, ComponentAccordion, Item, row, PartStyle{}, d)
			row.Children(func() {
				expanded := slices.Contains(*open, it.ID)
				var p ui.CollapsibleParts
				header := ui.Row(c.Key("header")).MinWidth(0)
				finish(t, ComponentAccordion, Header, header, PartStyle{}, d)
				header.Children(func() { p = ui.CollapsibleBase(c.Key("trigger"), &expanded); p.Trigger.Grow(1) })
				icon := it.TrailingIcon
				if icon == nil {
					icon = o.TrailingIcon
				}
				opts := CollapsibleOptions{Label: it.Label, Icon: it.Icon, TrailingIcon: icon, Disabled: o.Disabled || it.Disabled, Decoration: d}
				progress := disclosureProgress(c, t, p, expanded)
				disclosureTrigger(c, t, p.Trigger, progress, opts, ComponentAccordion)
				p.Trigger.Bind(memory.handles[it.ID])
				if !it.Disabled {
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
						p.Trigger.Jump("open", 1)
						progress = 1
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
				p.Trigger.Expanded(expanded)
				if (!expanded || it.Content == nil) && memory.bodies[it.ID].FocusWithin(c) && !p.Trigger.IsDisabled() {
					memory.handles[it.ID].Focus()
				}
				build := func() {
					if it.Content != nil {
						it.Content(c, t)
					}
				}
				if it.Content != nil {
					disclosureBody(c, t, expanded, o.KeepMounted, progress, ComponentAccordion, d, PartStyle{Gap: Set(t.spacing(12)), Foreground: Set(t.Text)}, Set(t.spacing(14)), memory.bodies[it.ID], build)
				}
			})
		}
	})
	if restore && !o.Disabled && len(ids) > 0 {
		memory.handles[ids[0]].Focus()
	}
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
