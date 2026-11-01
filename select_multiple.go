package nui

import (
	"fmt"
	"maps"

	"github.com/egoist/mygo/ui"
)

type multipleSelectMemory[T comparable] struct {
	menu SelectMenuState
	ids  map[T]string
	next uint64
}

// SelectMultiple is Select's slice-valued counterpart. It reuses the searchable
// list implementation with search disabled, preserving typed option identities.
func SelectMultiple[T comparable](c *ui.Context, t Theme, value *[]T, options []Option[T], o SelectOptions) ui.Element {
	if value == nil {
		panic("nui: SelectMultiple requires a slice pointer")
	}
	validateOptions(options)
	config := inputDefaults(t, ComponentSelect, InputOptions{Color: o.Color, Variant: o.Variant, Size: o.Size})
	// The adapter remains a Select for theme resolution, including slot overrides.
	menuTheme := t
	menuTheme.components = make(map[Component]ComponentTheme, len(t.components)+1)
	maps.Copy(menuTheme.components, t.components)
	styles := cloneStyles(t.Component(ComponentSelect).UI)
	for slot, style := range o.UI {
		styles[slot] = mergeStyle(styles[slot], style)
	}
	// Select.Root is its trigger; SelectMenu.Root is an outer container.
	// Map slots explicitly so trigger dimensions never constrain the popup list.
	menuTheme.components[ComponentSelectMenu] = ComponentTheme{Color: config.Color, Variant: config.Variant, Size: config.Size}
	decoration := Decoration{UI: cloneStyles(styles)}
	delete(decoration.UI, Root)
	delete(decoration.UI, Panel)
	decoration.UI[Trigger] = mergeStyle(styles[Root], styles[Trigger])
	if o.Style != nil {
		decoration.Style = func(slot Slot, e ui.Element) {
			if slot == Root {
				return
			}
			if slot == Trigger {
				slot = Root
			}
			o.Style(slot, e)
		}
	}
	popup := o.Popup
	popup.UI = cloneStyles(popup.UI)
	popup.UI[Panel] = mergeStyle(popup.UI[Panel], styles[Panel])
	popup.Style = func(slot Slot, e ui.Element) {
		if o.Popup.Style != nil {
			o.Popup.Style(slot, e)
		}
		if o.Style != nil {
			o.Style(slot, e)
		}
	}
	root := ui.Column(c).MinWidth(0)
	s := ui.Local(root, "multiple", func() multipleSelectMemory[T] { return multipleSelectMemory[T]{ids: map[T]string{}} })
	valid := map[T]bool{}
	for _, it := range options {
		valid[it.Value] = true
	}
	for _, v := range *value {
		valid[v] = true
	}
	for v := range s.ids {
		if !valid[v] {
			delete(s.ids, v)
		}
	}
	id := func(v T) string {
		if s.ids[v] == "" {
			s.next++
			s.ids[v] = fmt.Sprint(s.next)
		}
		return s.ids[v]
	}
	items := make([]ListboxItem, 0, len(options))
	byID := map[string]T{}
	byOption := map[string]Option[T]{}
	for _, it := range options {
		key := id(it.Value)
		items = append(items, ListboxItem{ID: key, Label: it.Label, Description: it.Description, Disabled: it.Disabled, Icon: it.Icon})
		byID[key] = it.Value
		byOption[key] = it
	}
	selected := make([]string, 0, len(*value))
	for _, v := range *value {
		key := id(v)
		selected = append(selected, key)
		byID[key] = v
		if s.menu.labels == nil {
			s.menu.labels = map[string]string{}
		}
		if s.menu.labels[key] == "" {
			s.menu.labels[key] = fmt.Sprint(v)
		}
	}
	root.Children(func() {
		SelectMenu(c, menuTheme, &selected, &s.menu, items, SelectMenuOptions{
			Label: o.Label, Multiple: true, Disabled: o.Disabled, Decoration: decoration, Build: func(c *ui.Context, t Theme, item ListboxItem) {
				optionContent(c, t, byOption[item.ID], false, inputDefaults(t, ComponentSelect, InputOptions{Size: o.Size}).Size)
			}, OnChange: func(ids []string) {
				next := make([]T, 0, len(ids))
				for _, key := range ids {
					if v, ok := byID[key]; ok {
						next = append(next, v)
					}
				}
				*value = next
				if o.OnChange != nil {
					o.OnChange()
				}
			},
			Placeholder: o.Placeholder, SearchInput: Set(false), Color: config.Color, Variant: config.Variant, Size: config.Size, Popup: popup, Invalid: o.Invalid, ErrorMessage: o.ErrorMessage,
		})
	})
	if o.ErrorMessage != "" {
		root.Error(o.ErrorMessage)
	} else if o.Invalid {
		root.Error(message(c, "checkValue"))
	}
	return root
}
