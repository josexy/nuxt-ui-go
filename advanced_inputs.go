package nui

import (
	"fmt"
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

type TagsState struct {
	Query           string
	Handle          ui.Handle
	focused, pasted bool
	active          Optional[tagKey]
	removers        map[tagKey]*ui.Handle
}

type tagKey struct {
	value      string
	occurrence int
}

func tagKeys(values []string) []tagKey {
	keys := make([]tagKey, len(values))
	counts := map[string]int{}
	for i, value := range values {
		keys[i] = tagKey{value, counts[value]}
		counts[value]++
	}
	return keys
}

type InputTagsOptions struct {
	InputOptions
	MaxTags                          int
	Separators                       string
	AddOnTab, AddOnBlur, Duplicate   bool
	AddOnPaste                       Optional[bool]
	MaxLength                        int // Native editing limit in UTF-16 units; 0 is unlimited.
	OnTagsChange                     func([]string)
	DeleteIcon                       *ui.SVG
	ConvertValue, DisplayValue       func(string) string
	OnAddTag, OnRemoveTag, OnInvalid func(string)
}

// AddTags trims and deduplicates without changing the order of existing tags.
func AddTags(tags []string, text, separators string, limit int) []string {
	return addTags(tags, text, separators, limit, false)
}
func addTags(tags []string, text, separators string, limit int, duplicate bool) []string {
	if separators == "" {
		separators = ","
	}
	out := slices.Clone(tags)
	for _, v := range strings.FieldsFunc(text, func(r rune) bool { return strings.ContainsRune(separators, r) }) {
		v = strings.TrimSpace(v)
		if v != "" && (duplicate || !slices.Contains(out, v)) && (limit <= 0 || len(out) < limit) {
			out = append(out, v)
		}
	}
	return out
}

func insertedSeparator(before, after, separators string) bool {
	a, b := []rune(before), []rune(after)
	start := 0
	for start < len(a) && start < len(b) && a[start] == b[start] {
		start++
	}
	endA, endB := len(a), len(b)
	for endA > start && endB > start && a[endA-1] == b[endB-1] {
		endA--
		endB--
	}
	return strings.ContainsAny(string(b[start:endB]), separators)
}

func InputTags(c *ui.Context, t Theme, value *[]string, s *TagsState, o InputTagsOptions) ui.Element {
	if value == nil || s == nil {
		panic("nui: InputTags requires value and state")
	}
	keys := tagKeys(*value)
	if s.active.Set && !slices.Contains(keys, s.active.Value) {
		s.active = Optional[tagKey]{}
	}
	if s.removers == nil {
		s.removers = map[tagKey]*ui.Handle{}
	}
	mutable := !o.ReadOnly && !o.Disabled && !t.field.Disabled
	emit := func(next []string) {
		if !slices.Equal(*value, next) {
			*value = next
			if o.OnTagsChange != nil {
				o.OnTagsChange(slices.Clone(next))
			}
		}
	}
	separators := o.Separators
	if separators == "" {
		separators = ","
	}
	add := func(text string, clearDraft bool) bool {
		if !mutable {
			return false
		}
		next := slices.Clone(*value)
		var added []string
		for _, candidate := range strings.FieldsFunc(text, func(r rune) bool { return strings.ContainsRune(separators, r) }) {
			candidate = strings.TrimSpace(candidate)
			if candidate == "" {
				continue
			}
			if o.ConvertValue != nil {
				candidate = o.ConvertValue(candidate)
			}
			if candidate == "" || o.MaxTags > 0 && len(next) >= o.MaxTags || !o.Duplicate && slices.Contains(next, candidate) {
				if o.OnInvalid != nil {
					o.OnInvalid(candidate)
				}
				continue
			}
			next = append(next, candidate)
			added = append(added, candidate)
		}
		if len(added) == 0 {
			return false
		}
		emit(next)
		if clearDraft {
			s.Query = ""
		}
		s.active = Optional[tagKey]{}
		if o.OnAddTag != nil {
			for _, tag := range added {
				o.OnAddTag(tag)
			}
		}
		return true
	}
	remove := func(key tagKey) {
		if !mutable {
			return
		}
		i := slices.Index(tagKeys(*value), key)
		if i < 0 {
			return
		}
		removed := (*value)[i]
		emit(slices.Delete(slices.Clone(*value), i, i+1))
		s.active = Optional[tagKey]{}
		if o.OnRemoveTag != nil {
			o.OnRemoveTag(removed)
		}
		s.Handle.Focus(c)
	}
	in := o.InputOptions
	in.MaxLength = o.MaxLength
	in.focusWithin = true
	changed, submitted, style := in.OnChange, in.OnSubmit, in.Style
	previousQuery := s.Query
	in.OnChange = func(text string) {
		s.active = Optional[tagKey]{}
		if !s.pasted && insertedSeparator(previousQuery, text, separators) {
			if !add(text, true) {
				s.Query = strings.TrimFunc(text, func(r rune) bool { return strings.ContainsRune(separators, r) })
			}
		}
		s.pasted = false
		if changed != nil {
			changed(text)
		}
	}
	in.OnSubmit = func(text string) {
		add(text, true)
		if submitted != nil {
			submitted(text)
		}
	}
	config := inputDefaults(t, ComponentInputTags, in)
	m := buttonMetrics(t, config.Size)
	in.Style = func(root, editor, leading, trailing ui.Element) {
		editor.Bind(&s.Handle).Basis(t.spacing(80))
		editor.OnPaste(func(text string) bool {
			if !valueOr(o.AddOnPaste, false) {
				s.pasted = text != ""
				return false
			}
			if !mutable {
				return false
			}
			add(text, false)
			return true
		}).OnTextKey(func(k ui.TextKey) bool {
			if !mutable || c.Resolve(s.Handle).Composing() {
				return false
			}
			if o.AddOnTab && k.Key == ui.KeyTab && (k.Mods == 0 || k.Mods == ui.Shift) {
				if k.Text() == "" {
					return false
				}
				s.Query = k.Text()
				add(k.Text(), true)
				k.SetText(s.Query)
				return true
			}
			if k.Mods != 0 {
				s.active = Optional[tagKey]{}
				return false
			}
			start, end := c.Resolve(s.Handle).TextSelection()
			if start != 0 || end != 0 {
				s.active = Optional[tagKey]{}
				return false
			}
			current := tagKeys(*value)
			index := -1
			if s.active.Set {
				index = slices.Index(current, s.active.Value)
			}
			if len(current) == 0 {
				return false
			}
			switch k.Key {
			case ui.KeyBackspace, ui.KeyDelete:
				if index >= 0 {
					remove(current[index])
					remaining := tagKeys(*value)
					if len(remaining) > 0 {
						s.active = Set(remaining[min(index, len(remaining)-1)])
					}
					return true
				}
				if k.Key == ui.KeyBackspace {
					s.active = Set(current[len(current)-1])
					return true
				}
			case ui.KeyLeft, ui.KeyRight:
				back := (k.Key == ui.KeyLeft) != (c.Direction() == ui.RightToLeft)
				if index < 0 {
					if back {
						s.active = Set(current[len(current)-1])
						return true
					}
					return false
				}
				if back {
					index = max(0, index-1)
				} else {
					index++
				}
				if index >= len(current) {
					s.active = Optional[tagKey]{}
				} else {
					s.active = Set(current[index])
				}
				return true
			case ui.KeyHome, ui.KeyEnd:
				if index >= 0 {
					index = 0
					if k.Key == ui.KeyEnd {
						index = len(current) - 1
					}
					s.active = Set(current[index])
					return true
				}
			case ui.KeyUp, ui.KeyDown:
				return index >= 0
			default:
				s.active = Optional[tagKey]{}
			}
			return false
		})
		if style != nil {
			style(root, editor, leading, trailing)
		}
	}
	root := inputShell(c, t, &s.Query, in, ComponentInputTags, 0, 0, func() ui.Element {
		var editor ui.Element
		body := ui.Row(c.Key("body")).Grow(1).MinWidth(0).Wrap().AlignItems(ui.Center).Gap(m.Gap)
		finish(t, ComponentInputTags, Body, body, PartStyle{}, Decoration{UI: o.UI})
		body.Children(func() {
			for i, tag := range *value {
				key := keys[i]
				if s.removers[key] == nil {
					s.removers[key] = &ui.Handle{}
				}
				label := tag
				if o.DisplayValue != nil {
					label = o.DisplayValue(tag)
				}
				_, del := tagChipStyled(c.Key(key), t, label, namedMessage(c, "removeNamed", label), config.Size, o.ReadOnly, o.DeleteIcon, func() { remove(key) }, func(slot Slot, e ui.Element) {
					if slot == Item {
						active := s.active.Set && s.active.Value == key
						e.Font(t.Font).Selected(active).Background(t.Elevated).TextColor(t.Text)
						if active {
							e.Background(t.Accented)
						}
					}
					finish(t, ComponentInputTags, slot, e, PartStyle{}, Decoration{UI: o.UI})
				})
				del.Bind(s.removers[key])
			}
			editor = ui.TextInputBase(c.Key("query"), &s.Query)
		})
		return editor
	}, nil)
	for key, handle := range s.removers {
		if !slices.Contains(keys, key) {
			if !root.IsDisabled() && handle.Focused(c) {
				s.Handle.Focus(c)
			}
			handle.CancelFocus(c)
			delete(s.removers, key)
		}
	}
	focused := s.Handle.Focused(c)
	if s.focused && !focused {
		s.active = Optional[tagKey]{}
		if o.AddOnBlur && !root.IsDisabled() && add(s.Query, true) {
			c.Services().Invalidate()
		}
	}
	s.focused = focused
	return root
}

type PinState struct {
	text       string
	seen       string
	start, end int
	backward   bool
	Handle     ui.Handle
}

type pinCaretState struct {
	focused    bool
	start, end int
	text       string
	since      time.Time
}
type PinInputOptions struct {
	Label                string
	Length               int
	Disabled, Mask       bool
	Numeric, ReadOnly    bool // Text is the default; Numeric keeps only ASCII digits.
	Highlight            bool
	Placeholder          string
	Color                Color
	Variant              Variant
	Size                 Size
	Separator            []int // Insert a separator after these 1-based slots.
	SeparatorEvery       int   // Insert after every N slots; Separator also applies.
	OnChange, OnComplete func(string)
	Decoration
}

func normalizePIN(v string, n int, numeric bool) string {
	var b strings.Builder
	count := 0
	for _, r := range v {
		if count < n && !unicode.IsControl(r) && (!numeric || r >= '0' && r <= '9') {
			b.WriteRune(r)
			count++
		}
	}
	return b.String()
}

// Native text insertion replaces the selection and shifts its suffix. A PIN
// instead overwrites as many slots as the accepted input occupies. Recognize
// that insertion from its old selection and new caret; leave deletion and
// restored undo/redo selections to the native editor.
func replacePINSlots(previous, next string, start, end, caret, n int, numeric bool) (string, int) {
	old, raw := []rune(previous), []rune(next)
	start, end = min(start, len(old)), min(end, len(old))
	added := len(raw) - len(old) + end - start
	if added > 0 && caret == start+added && caret <= len(raw) &&
		string(raw[:start]) == string(old[:start]) && string(raw[caret:]) == string(old[end:]) {
		insert := []rune(normalizePIN(string(raw[start:caret]), n-start, numeric))
		if len(insert) == 0 {
			return previous, start
		}
		tail := min(max(end, start+len(insert)), len(old))
		return normalizePIN(string(old[:start])+string(insert)+string(old[tail:]), n, numeric), start + len(insert)
	}
	return normalizePIN(next, n, numeric), caret
}

// PinInput uses one real editor, preserving paste, selection and Backspace across all slots.
func PinInput(c *ui.Context, t Theme, value *string, s *PinState, o PinInputOptions) ui.Element {
	if value == nil || s == nil {
		panic("nui: PinInput requires value and state")
	}
	n := o.Length
	if n == 0 {
		n = 5
	}
	if n < 1 || n > 32 {
		panic("nui: PIN length must be 1..32")
	}
	// Keep a localized name for standalone PINs without overriding application
	// labels or FormField binding. The visible placeholder remains caller-owned.
	label := o.Label
	if label == "" && t.field.Label == "" {
		key := "pinCharacters"
		if o.Numeric {
			key = "pinDigits"
		}
		label = formatMessage(c, key, map[string]any{"count": n})
	}
	*value = normalizePIN(*value, n, o.Numeric)
	if s.seen != *value {
		s.text = *value
		s.seen = *value
	}
	config := selected(defaults(t, ComponentPinInput, ComponentTheme{Color: Primary, Variant: Outline, Size: MD}), ComponentTheme{Color: o.Color, Variant: o.Variant, Size: inheritedSize(t, o.Size)})
	slotSize := t.text([5]float32{24, 28, 32, 36, 40}[sizeIndex(config.Size)])
	rootStyle := mergeStyle(t.components[ComponentPinInput].UI[Root], o.UI[Root])
	gap := valueOr(rootStyle.Gap, t.spacing(6))
	paddingX, paddingY := valueOr(rootStyle.PaddingX, 0), valueOr(rootStyle.PaddingY, 0)
	separators := map[int]bool{}
	for _, at := range o.Separator {
		if at > 0 && at < n {
			separators[at] = true
		}
	}
	if o.SeparatorEvery > 0 {
		for at := o.SeparatorEvery; at < n; at += o.SeparatorEvery {
			separators[at] = true
		}
	}
	// Match the separator's text advance instead of reserving a fixed-width
	// dash box. Keep both layers absolute so slots receive pointer input.
	separatorStyle := mergeStyle(PartStyle{Font: Set(t.Font), FontSize: Set(t.text(16)), Foreground: Set(t.Dimmed)}, t.components[ComponentPinInput].UI[SeparatorSlot])
	separatorStyle = mergeStyle(separatorStyle, o.UI[SeparatorSlot])
	var separatorWidth float32
	contentHeight := slotSize
	if len(separators) > 0 {
		var separatorHeight float32
		separatorWidth, separatorHeight = c.MeasureText(0, ui.Span{Text: "•", Font: separatorStyle.Font.Value, Size: separatorStyle.FontSize.Value, Weight: valueOr(separatorStyle.Weight, 400)})
		separatorWidth += 2 * valueOr(separatorStyle.PaddingX, 0)
		if separatorStyle.LineHeight.Set {
			separatorHeight = max(separatorHeight, separatorStyle.FontSize.Value*separatorStyle.LineHeight.Value)
		}
		contentHeight = max(contentHeight, separatorHeight+2*valueOr(separatorStyle.PaddingY, 0))
	}
	root := ui.Column(c).Width(float32(n)*slotSize + float32(n-1)*gap + float32(len(separators))*(separatorWidth+gap) + 2*paddingX).Height(contentHeight + 2*paddingY).Disabled(o.Disabled || t.field.Disabled)
	characters := []rune(*value)
	caret := len(characters)
	focused := s.Handle.Focused(c)
	previousStart, previousEnd := s.start, s.end
	resetCaret := false
	caretState := ui.Local(root, "caret", func() pinCaretState { return pinCaretState{} })
	finish(t, ComponentPinInput, Root, root, PartStyle{}, o.Decoration)
	root.Children(func() {
		Input(c.Key("pin"), t, &s.text, InputOptions{Label: label, Placeholder: o.Placeholder, Password: o.Mask, ReadOnly: o.ReadOnly, OnChange: func(v string) {
			editor := c.Resolve(s.Handle)
			start, end := editor.TextSelection()
			if start == end {
				v, start = replacePINSlots(*value, v, previousStart, previousEnd, start, n, o.Numeric)
				length := len([]rune(v))
				start = min(start, length, n-1)
				end = min(start+1, length)
				editor.SetTextSelection(start, end)
			} else {
				v = normalizePIN(v, n, o.Numeric)
			}
			s.text = v
			if v != *value {
				*value = v
				s.seen = v
				if o.OnChange != nil {
					o.OnChange(v)
				}
				if len([]rune(v)) == n && o.OnComplete != nil {
					o.OnComplete(v)
				}
			}
		}, Style: func(root, editor, leading, trailing ui.Element) {
			editor.Bind(&s.Handle).InputMethod(!o.Numeric).OnTextKey(func(k ui.TextKey) bool {
				s.backward = k.Mods == 0 && k.Key == ui.KeyLeft
				return false
			})
			start, end := editor.TextSelection()
			// Selecting the current occupied slot makes the next keystroke a
			// replacement, including after typing, arrows and keyboard focus.
			if focused && !o.ReadOnly && !editor.IsDisabled() && !editor.Composing() && s.text == *value && start == end {
				// Left first collapses the native one-character selection. Move
				// one additional position to reach the preceding PIN slot.
				if s.backward && previousEnd == previousStart+1 {
					start = max(start-1, 0)
				}
				start = min(start, len(characters), n-1)
				end = min(start+1, len(characters))
				if start != end {
					editor.SetTextSelection(start, end)
					resetCaret = true
				}
			}
			s.backward = false
			s.start, s.end = start, end
			caret = start
		}}).Absolute().Left(0).Top(0).Fill().Opacity(.001)
		editing := focused && !o.ReadOnly && !root.IsDisabled()
		if resetCaret || editing != caretState.focused || s.start != caretState.start || s.end != caretState.end || *value != caretState.text {
			*caretState = pinCaretState{focused: editing, start: s.start, end: s.end, text: *value, since: c.Now()}
		}
		ui.Row(c.Key("slots")).Absolute().Left(paddingX).Top(paddingY).Height(contentHeight).AlignItems(ui.Center).Gap(gap).Cursor(ui.CursorText).OnClick(func() {
			length := len([]rune(*value))
			c.Resolve(s.Handle).Focus().SetTextSelection(length, length)
		}).Children(func() {
			for i := range n {
				if separators[i] {
					separator := ui.Text(c.Key(fmt.Sprintf("separator:%d", i)), "•").Shrink(0).IgnoreInput(true).Role(ui.RoleNone)
					separatorStyle.apply(separator)
					if o.Style != nil {
						o.Style(SeparatorSlot, separator)
					}
				}
				active := focused && !root.IsDisabled() && i == min(caret, n-1)
				v := o.Placeholder
				if active {
					v = ""
				}
				if i < len(characters) {
					v = string(characters[i])
					if o.Mask {
						v = "•"
					}
				}
				e := ui.Box(c.Key(i)).Size(slotSize, slotSize).Shrink(0).FontSize(buttonMetrics(t, config.Size).Font).Center().Cursor(ui.CursorText).Role(ui.RoleNone).OnClick(func() {
					// Slot geometry differs from the hidden editor's text geometry.
					// Resolve its current element instead of retaining a build value.
					length := len([]rune(*value))
					start := min(i, length)
					end := min(i+1, length)
					c.Resolve(s.Handle).Focus().SetTextSelection(start, end)
				}).Children(func() {
					finish(t, ComponentPinInput, Label, ui.Text(c, v).SingleLine().MinWidth(0).Role(ui.RoleNone).IgnoreInput(true), PartStyle{}, o.Decoration)
				})
				invalid := fieldInvalid(t, InputOptions{})
				recipe := inputRecipe(t, config, active, e.Hovered() && !e.IsDisabled(), invalid)[Root]
				recipe.PaddingX, recipe.PaddingY = Set(float32(0)), Set(float32(0))
				if o.Highlight && !invalid {
					recipe.BorderWidth, recipe.BorderColor = Set(float32(1)), Set(inputColor(t, config.Color))
				}
				if active && config.Variant != None {
					color := config.Color
					if invalid {
						color = Error
					}
					e.Shadow(0, 0, 0, 3, inputFocusColor(t, color))
				}
				if active && editing {
					var textWidth float32
					if i < len(characters) {
						textWidth, _ = c.MeasureText(0, ui.Span{Text: v, Font: t.Font, Size: buttonMetrics(t, config.Size).Font})
					}
					since, motion := caretState.since, t.motion(c)
					height := buttonMetrics(t, config.Size).Line
					e.DrawOver(func(p *ui.Painter, r ui.Rect) {
						const blink = 530 * time.Millisecond
						phase := p.Now().Sub(since)
						if !motion || (phase/blink)%2 == 0 {
							p.Fill(lineRect(ui.Rect{X: r.X + (r.W+textWidth)/2, Y: r.Y + (r.H-height)/2, W: 1, H: height}, p.Scale(), true, false, 0), t.Highlighted, 0)
						}
						if motion && phase < 30*time.Second {
							p.After(blink - phase%blink)
						}
					})
				}
				finish(t, ComponentPinInput, Item, e, recipe, o.Decoration)
			}
		})
	})
	return root
}

type InputRatingOptions struct {
	Label              string
	Max                int
	Disabled, ReadOnly bool
	Color              Color
	Size               Size
	Orientation        Orientation
	Clearable          bool
	Hoverable          Optional[bool]
	Icon, EmptyIcon    *ui.SVG
	OnChange           func(int)
	Decoration
}

// ratingMemory is the star under the pointer and the hover preview, as Nuxt's
// hoveredRating: set when the pointer enters a star, kept over the gaps
// between stars, and reset when it leaves the rating or a click clears it.
type ratingMemory struct{ over, hover int }

func InputRating(c *ui.Context, t Theme, value *int, o InputRatingOptions) ui.Element {
	vert := vertical(o.Orientation)
	if value == nil {
		panic("nui: InputRating requires value")
	}
	n := o.Max
	if n == 0 {
		n = 5
	}
	if n < 1 || n > 20 {
		panic("nui: rating maximum must be 1..20")
	}
	*value = max(0, min(n, *value))
	// Nuxt rings the current star, not the whole control.
	root := ui.Row(c).Focusable().FocusRing(false).Role(ui.RoleSlider).Label(o.Label).Range(0, float64(n), float64(*value)).Step(1).Value(fmt.Sprint(*value)).Disabled(o.Disabled || t.field.Disabled).ReadOnly(o.ReadOnly)
	if vert {
		root.Column()
	}
	bindField(c, root, t)
	memory := ui.Local(root, "rating", func() ratingMemory { return ratingMemory{} })
	interactive := !root.IsDisabled() && !o.ReadOnly
	set := func(v int) {
		if o.ReadOnly {
			return
		}
		v = max(0, min(n, v))
		if v != *value {
			*value = v
			if o.OnChange != nil {
				o.OnChange(v)
			}
		}
	}
	clear := func() {
		// A cleared rating previews nothing until the pointer enters a star again.
		memory.hover = 0
		set(0)
	}
	increase, decrease := ui.KeyRight, ui.KeyLeft
	if vert {
		increase, decrease = ui.KeyDown, ui.KeyUp
	} else if c.Direction() == ui.RightToLeft {
		increase, decrease = decrease, increase
	}
	// As Nuxt's, the arrow keys stop at the first and the last star: only
	// clearing returns to 0. Home and End set the value, as a slider's do.
	root.OnShortcut(0, increase, func() {
		if *value < n {
			set(*value + 1)
		}
	}).OnShortcut(0, decrease, func() {
		if *value > 1 {
			set(*value - 1)
		}
	}).OnShortcut(0, ui.KeyHome, func() { set(1) }).OnShortcut(0, ui.KeyEnd, func() { set(n) }).OnShortcut(0, ui.KeyPageUp, func() { set(1) }).OnShortcut(0, ui.KeyPageDown, func() { set(n) })
	// Space acts on the current star, as on Nuxt's focused one: it rates an
	// empty rating 1, and clears the current star when Clearable.
	root.OnShortcut(0, ui.KeySpace, func() {
		switch {
		case *value == 0:
			set(1)
		case o.Clearable:
			clear()
		}
	})
	tone := o.Color
	if t.field.Error != "" {
		tone = Error // An invalid field colors the stars and the ring, as Nuxt's.
	}
	config, styles := resolve(t, ComponentInputRating, ComponentTheme{Color: Primary, Size: MD}, ComponentTheme{Color: tone, Size: inheritedSize(t, o.Size), UI: o.UI}, func(v ComponentTheme) Styles {
		return Styles{Root: {Gap: Set(t.spacing(2))}, Indicator: {Foreground: Set(t.color(v.Color))}}
	})
	tone = config.Color
	size := t.text([5]float32{12, 16, 20, 24, 28}[sizeIndex(config.Size)])
	// Nuxt's inline-flex rating takes its stars' size instead of stretching
	// across its container, so hover and focus end where the stars do.
	style := styles[Root]
	length := float32(n)*size + float32(n-1)*style.Gap.Value
	w, h := length, size
	if vert {
		w, h = size, length
	}
	root.Size(w+2*style.PaddingX.Value, h+2*style.PaddingY.Value)
	decorated(Root, root, styles, o.Style)
	focus := root.FocusVisible()
	preview := *value
	now := 0 // The star under the pointer in this build.
	icon, emptyIcon := o.Icon, o.EmptyIcon
	if icon == nil {
		icon = lucide.Star()
	}
	if emptyIcon == nil {
		emptyIcon = icon
	}
	// These elements are used only while configuring this build. Keeping both
	// layers lets the filled preview change without replacing item identities.
	indicators := make([]ui.Element, n)
	root.Children(func() {
		for i := 1; i <= n; i++ {
			e := ui.Box(c.Key(i)).Size(size, size).Center().Radius(t.radius(4)).OnClick(func() {
				memory.over = i // The pointer is on this star, which must not count as entering it.
				if o.Clearable && *value == i {
					clear()
				} else {
					set(i)
				}
			})
			if o.ReadOnly && !root.IsDisabled() {
				e.Cursor(ui.CursorDefault)
			}
			if e.Hovered() {
				now = i
			}
			// Interactive stars grow on hover, as Nuxt's hover:scale-110.
			scale := float32(1)
			if interactive && e.Hovered() {
				scale = 1.1
			}
			if t.motion(c) && t.Motion.Transition > 0 {
				scale = e.AnimateWith("nui-rating-scale", scale, t.Motion.Transition, ui.EaseOut)
			} else {
				e.Jump("nui-rating-scale", scale)
			}
			e.TransformOrigin(.5, .5).Scale(scale, scale)
			// The ring marks the current star: the rated one, or the first of an empty rating.
			if focus && i == max(1, *value) {
				e.Shadow(0, 0, 0, 3, inputFocusColor(t, tone))
			}
			decorated(Item, e, styles, o.Style)
			e.Children(func() {
				decorated(EmptyIconSlot, ui.Icon(c.Key("empty"), emptyIcon).Fill().TextColor(t.Muted), styles, o.Style)
				indicator := ui.Box(c.Key("indicator")).Absolute().Left(0).Top(0).Fill().Clip()
				decorated(Indicator, indicator, styles, o.Style)
				indicator.Children(func() {
					decorated(IconSlot, ui.Icon(c.Key("icon"), icon).Fill(), styles, o.Style)
				})
				indicators[i-1] = indicator
			})
		}
	})
	if interactive && valueOr(o.Hoverable, false) {
		switch {
		case now > 0 && now != memory.over:
			memory.hover = now // The pointer entered a star.
		case !root.Hovered():
			memory.hover = 0 // It left the rating; between stars, the preview stays.
		}
		memory.over = now
		if memory.hover > 0 {
			preview = memory.hover
		}
	} else {
		memory.over, memory.hover = now, 0
	}
	for i, indicator := range indicators {
		indicator.Hidden(i+1 > preview)
	}
	return root
}

type CheckState uint8

const (
	Unchecked CheckState = iota
	Checked
	Indeterminate
)

type CheckboxTriOptions struct {
	CheckboxOptions
	// IndeterminateIcon replaces the minus in a visible mixed-state indicator.
	IndeterminateIcon *ui.SVG
	OnStateChange     func(CheckState)
}

func CheckboxTri(c *ui.Context, t Theme, value *CheckState, o CheckboxTriOptions) ui.Element {
	if value == nil || *value > Indeterminate {
		panic("nui: invalid checkbox state")
	}
	mixed := *value == Indeterminate
	on := *value == Checked
	e := ui.CheckboxBase(c, &on).Disabled(o.Disabled || t.field.Disabled).FocusRing(false)
	if mixed {
		e.Mixed()
	}
	p := o.CheckboxOptions
	p.mixed = mixed
	p.indeterminateIcon = o.IndeterminateIcon
	togglePresentation(c, t, e, on || mixed, ComponentCheckbox, false, false, p)
	fieldActivation(c, e, t, func() bool {
		if o.Disabled || t.field.Disabled {
			return false
		}
		on = *value != Checked
		return true
	})
	if e.Changed() {
		if on {
			*value = Checked
		} else {
			*value = Unchecked
		}
	}
	e.OnChange(func() {
		if o.OnChange != nil {
			o.OnChange(*value == Checked)
		}
		if o.OnStateChange != nil {
			o.OnStateChange(*value)
		}
	})
	return e
}

type RangeSliderOptions struct {
	Label                 string
	Min, Max, Step        float64
	Disabled              bool
	Orientation           Orientation
	Inverted, Tooltip     bool
	MinStepsBetweenThumbs int
	Color                 Color
	Size                  Size
	OnChange              func(float64, float64)
	OnCommit              func(float64, float64)
	Decoration
}
