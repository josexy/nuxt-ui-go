package nui

import (
	"fmt"
	"slices"
	"strings"
	"unicode"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

type TagsState struct {
	Query           string
	Handle          ui.Handle
	focused, pasted bool
}
type InputTagsOptions struct {
	InputOptions
	MaxTags                        int
	Separators                     string
	AddOnTab, AddOnBlur, Duplicate bool
	AddOnPaste                     Optional[bool]
	MaxLength                      int // Native editing limit in UTF-16 units; 0 is unlimited.
	OnTagsChange                   func([]string)
}

// AddTags trims and deduplicates without changing the order of existing tags.
func AddTags(tags []string, text, separators string, limit int) []string {
	return addTags(tags, text, separators, limit, false)
}
func addTags(tags []string, text, separators string, limit int, duplicate bool) []string {
	if separators == "" {
		separators = ",;\n"
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
	emit := func(next []string) {
		if !slices.Equal(*value, next) {
			*value = next
			if o.OnTagsChange != nil {
				o.OnTagsChange(slices.Clone(next))
			}
		}
	}
	parse := func(text string) []string {
		return addTags(*value, text, o.Separators, o.MaxTags, o.Duplicate)
	}
	add := func(text string) {
		if o.ReadOnly || o.Disabled || t.field.Disabled {
			return
		}
		next := parse(text)
		if len(next) != len(*value) {
			emit(next)
			s.Query = ""
		}
	}
	root := ui.Column(c).MinWidth(0).Disabled(o.Disabled || t.field.Disabled)
	// Text editing always indicates focus; tag buttons only do so when reached
	// by keyboard. Match Nuxt's has-focus-visible on the outer field.
	focusVisible := false
	size := selected(defaults(t, ComponentInputTags, ComponentTheme{Size: MD}), ComponentTheme{Size: inheritedSize(t, o.Size)}).Size
	root.Children(func() {
		ui.Row(c.Key("tags")).Wrap().Gap(t.spacing(4)).Children(func() {
			occurrences := map[string]int{}
			for i, tag := range *value {
				key := struct {
					Tag        string
					Occurrence int
				}{tag, occurrences[tag]}
				occurrences[tag]++
				c.Key(key)
				_, del := tagChip(c, t, tag, namedMessage(c, "removeNamed", tag), size, o.ReadOnly, func() { emit(slices.Delete(slices.Clone(*value), i, i+1)) })
				focusVisible = focusVisible || del.FocusVisible()
			}
		})
		in := o.InputOptions
		in.UI = cloneStyles(o.UI)
		// The outer field owns decorations, including the error border. Keep
		// error semantics on the editor without drawing a second border inside.
		in.UI[Root] = PartStyle{BorderWidth: Set(float32(0))}
		changed, submitted := in.OnChange, in.OnSubmit
		previousQuery := s.Query
		in.OnChange = func(text string) {
			if o.MaxLength > 0 {
				text = limitTextEdit(previousQuery, text, o.MaxLength)
				s.Query = text
			}
			sep := o.Separators
			if sep == "" {
				sep = ",;\n"
			}
			if !s.pasted && insertedSeparator(previousQuery, text, sep) {
				add(text)
			}
			s.pasted = false
			if changed != nil {
				changed(text)
			}
		}
		in.OnSubmit = func(text string) {
			add(text)
			if submitted != nil {
				submitted(text)
			}
		}
		style := in.Style
		in.Style = func(root, editor, leading, trailing ui.Element) {
			focusVisible = focusVisible || editor.Focused()
			editor.Bind(&s.Handle).OnPaste(func(text string) bool {
				if !valueOr(o.AddOnPaste, false) {
					s.pasted = text != ""
					return false
				}
				if !o.ReadOnly {
					next := parse(text)
					if len(next) != len(*value) {
						emit(next)
					}
					return true
				}
				return false
			}).OnTextKey(func(k ui.TextKey) bool {
				if o.AddOnTab && k.Key == ui.KeyTab && (k.Mods == 0 || k.Mods == ui.Shift) && !o.ReadOnly {
					if k.Text() == "" {
						return false
					}
					add(k.Text())
					k.SetText(s.Query)
					return true
				}
				if !o.ReadOnly && k.Key == ui.KeyBackspace && k.Text() == "" && len(*value) > 0 {
					emit(slices.Clone((*value)[:len(*value)-1]))
					return true
				}
				return false
			})
			if style != nil {
				style(root, editor, leading, trailing)
			}
		}
		in.Variant = None
		Input(c.Key("query"), t, &s.Query, in)
	})
	focused := s.Handle.Focused(c)
	if s.focused && !focused && o.AddOnBlur {
		add(s.Query)
	}
	s.focused = focused
	config, styles := resolve(t, ComponentInputTags, ComponentTheme{Color: Primary, Size: MD, Variant: Outline}, ComponentTheme{Color: o.Color, Size: inheritedSize(t, o.Size), Variant: o.Variant, UI: o.UI}, func(v ComponentTheme) Styles {
		styles := inputRecipe(t, v, focusVisible, (root.Hovered() || root.FocusWithin()) && !root.IsDisabled(), fieldInvalid(t, o.InputOptions))
		style := styles[Root]
		style.Gap = Set(t.spacing(4))
		styles[Root] = style
		return styles
	})
	styles[Root].apply(root)
	if focusVisible && config.Variant != None {
		color := config.Color
		if fieldInvalid(t, o.InputOptions) {
			color = Error
		}
		root.Shadow(0, 0, 0, 3, t.focusColor(color))
	}
	applyGroup(c, root, t, styles[Root].Radius.Value)
	return root
}

type PinState struct {
	text   string
	seen   string
	Handle ui.Handle
}
type PinInputOptions struct {
	Label                string
	Length               int
	Disabled, Mask       bool
	Numeric, ReadOnly    bool // Text is the default; Numeric keeps only ASCII digits.
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
	root := ui.Column(c).Width(float32(n)*slotSize + float32(n-1)*t.spacing(6) + float32(len(separators))*(t.text(12)+t.spacing(6))).Height(slotSize).Disabled(o.Disabled || t.field.Disabled)
	characters := []rune(*value)
	caret := len(characters)
	focused := s.Handle.Focused(c)
	finish(t, ComponentPinInput, Root, root, PartStyle{}, o.Decoration)
	root.Children(func() {
		Input(c.Key("pin"), t, &s.text, InputOptions{Label: label, Placeholder: o.Placeholder, Password: o.Mask, ReadOnly: o.ReadOnly, OnChange: func(v string) {
			v = normalizePIN(v, n, o.Numeric)
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
			editor.Bind(&s.Handle).InputMethod(!o.Numeric)
			caret, _ = editor.TextSelection()
		}}).Absolute().Left(0).Top(0).Fill().Opacity(.001)
		ui.Row(c.Key("slots")).Absolute().Left(0).Top(0).Gap(t.spacing(6)).Cursor(ui.CursorText).OnClick(func() {
			length := len([]rune(*value))
			c.Resolve(s.Handle).Focus().SetTextSelection(length, length)
		}).Children(func() {
			for i := range n {
				if separators[i] {
					ui.Text(c.Key(fmt.Sprintf("separator:%d", i)), "−").Width(t.text(12)).Height(slotSize).IgnoreInput(true).Role(ui.RoleNone)
				}
				v := o.Placeholder
				if i < len(characters) {
					v = string(characters[i])
					if o.Mask {
						v = "•"
					}
				}
				e := ui.Box(c.Key(i)).Size(slotSize, slotSize).FontSize(buttonMetrics(t, config.Size).Font).Center().Cursor(ui.CursorText).Role(ui.RoleNone).OnClick(func() {
					// Slot geometry differs from the hidden editor's text geometry.
					// Resolve its current element instead of retaining a build value.
					length := len([]rune(*value))
					start := min(i, length)
					end := min(i+1, length)
					c.Resolve(s.Handle).Focus().SetTextSelection(start, end)
				}).Children(func() { ui.Text(c, v).Role(ui.RoleNone).IgnoreInput(true) })
				active := focused && i == min(caret, n-1)
				recipe := inputRecipe(t, config, active, e.Hovered(), fieldInvalid(t, InputOptions{}))[Root]
				recipe.PaddingX, recipe.PaddingY = Set(float32(0)), Set(float32(0))
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
	root := ui.Row(c).Focusable().Role(ui.RoleSlider).Label(o.Label).Range(0, float64(n), float64(*value)).Step(1).Value(fmt.Sprint(*value)).Disabled(o.Disabled || t.field.Disabled).ReadOnly(o.ReadOnly)
	if vert {
		root.Column()
	}
	bindField(c, root, t)
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
	increase, decrease := ui.KeyRight, ui.KeyLeft
	if vert {
		increase, decrease = ui.KeyDown, ui.KeyUp
	} else if c.Direction() == ui.RightToLeft {
		increase, decrease = decrease, increase
	}
	root.OnShortcut(0, increase, func() { set(*value%n + 1) }).OnShortcut(0, decrease, func() { set((max(1, *value)+n-2)%n + 1) }).OnShortcut(0, ui.KeyHome, func() { set(1) }).OnShortcut(0, ui.KeyEnd, func() { set(n) }).OnShortcut(0, ui.KeyPageUp, func() { set(1) }).OnShortcut(0, ui.KeyPageDown, func() { set(n) })
	finish(t, ComponentInputRating, Root, root, PartStyle{Gap: Set(t.spacing(2))}, o.Decoration)
	config := selected(defaults(t, ComponentInputRating, ComponentTheme{Color: Primary, Size: MD}), ComponentTheme{Color: o.Color, Size: inheritedSize(t, o.Size)})
	size := t.text([5]float32{12, 16, 20, 24, 28}[sizeIndex(config.Size)])
	preview := *value
	root.Children(func() {
		for i := 1; i <= n; i++ {
			e := ui.Box(c.Key(i)).Size(size, size).Center().OnClick(func() {
				if o.Clearable && *value == i {
					set(0)
				} else {
					set(i)
				}
			})
			if valueOr(o.Hoverable, false) && e.Hovered() && !e.IsDisabled() && !o.ReadOnly {
				preview = i
			}
			e.Draw(func(p *ui.Painter, r ui.Rect) {
				color, icon := t.Muted, o.EmptyIcon
				if icon == nil {
					icon = o.Icon
				}
				if i <= preview {
					color = t.color(config.Color)
					icon = o.Icon
				}
				if icon == nil {
					icon = lucide.Star()
				}
				p.Icon(icon, r, color)
			})
			finish(t, ComponentInputRating, Item, e, PartStyle{}, o.Decoration)
		}
	})
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
	OnStateChange func(CheckState)
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
