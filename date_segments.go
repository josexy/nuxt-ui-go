package nui

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/egoist/mygo/i18n"
	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

type datePart uint8

const (
	partLiteral datePart = iota
	partYear
	partMonth
	partDay
	partHour
	partMinute
	partSecond
	partPeriod
	partCount
)

// Only values and persistent handles survive builds. Parts are keyed by their
// meaning, so a locale change can reorder them without losing focus or a draft.
type segmentState struct {
	values    [partCount]Optional[int]
	handles   [partCount]ui.Handle
	typed     [partCount]string
	selectAll bool
}

type dateSegment struct {
	part    datePart
	literal string
}

// segmentField describes this build's editor. It is never stored in widget state.
type segmentField struct {
	state  *segmentState
	parts  []dateSegment
	label  string
	text   *string
	err    *string
	twelve bool
	clock  bool
	change func(datePart)
	paste  func(string)
}

func segmentPattern(locale string, clock, twelve, seconds bool) []dateSegment {
	skeleton := "yMd"
	if clock {
		skeleton = "Hm"
		if twelve {
			skeleton = "hm"
		}
		if seconds {
			skeleton += "s"
		}
	}
	pattern := i18n.NewDateFormatter(locale, i18n.DateOptions{Skeleton: skeleton}).Pattern()
	var parts []dateSegment
	quoted := false
	runes := []rune(pattern)
	for i := 0; i < len(runes); {
		r := runes[i]
		if r == '\'' {
			if i+1 < len(runes) && runes[i+1] == '\'' {
				r = '\''
				i++
			} else {
				quoted = !quoted
				i++
				continue
			}
		}
		part := partLiteral
		if !quoted {
			switch r {
			case 'y':
				part = partYear
			case 'M', 'L':
				part = partMonth
			case 'd':
				part = partDay
			case 'h', 'H', 'K', 'k':
				part = partHour
			case 'm':
				part = partMinute
			case 's':
				part = partSecond
			case 'a', 'b', 'B':
				part = partPeriod
			}
		}
		i++
		if part != partLiteral {
			for i < len(runes) && runes[i] == r {
				i++
			}
			parts = append(parts, dateSegment{part: part})
		} else if len(parts) > 0 && parts[len(parts)-1].part == partLiteral {
			parts[len(parts)-1].literal += string(r)
		} else {
			parts = append(parts, dateSegment{literal: string(r)})
		}
	}
	return parts
}

func (f segmentField) limits(p datePart) (int, int) {
	switch p {
	case partYear:
		return 1, 9999
	case partMonth:
		return 1, 12
	case partDay:
		y, m := f.state.values[partYear], f.state.values[partMonth]
		if y.Set && m.Set {
			return 1, time.Date(y.Value, time.Month(m.Value)+1, 0, 12, 0, 0, 0, time.UTC).Day()
		}
		return 1, 31
	case partHour:
		if f.twelve {
			return 1, 12
		}
		return 0, 23
	case partPeriod:
		return 0, 1
	default:
		return 0, 59
	}
}

func segmentName(p datePart) string {
	return [...]string{"", "dateYear", "dateMonth", "dateDay", "timeHour", "timeMinute", "timeSecond", "timePeriod"}[p]
}

func (f segmentField) display(c *ui.Context, p datePart) string {
	v := f.state.values[p]
	if !v.Set {
		return message(c, segmentName(p)+"Placeholder")
	}
	if p == partPeriod {
		am, pm := i18n.DayPeriods(c.Region().FormatLocale(c.Locale()), i18n.Abbreviated)
		if v.Value == 1 {
			return pm
		}
		return am
	}
	if p == partYear {
		return fmt.Sprintf("%04d", v.Value)
	}
	return fmt.Sprintf("%02d", v.Value)
}

func (f segmentField) numeric(p datePart, digit byte) bool {
	s := f.state
	lo, hi := f.limits(p)
	text := s.typed[p] + string(digit)
	v, _ := strconv.Atoi(text)
	width := 2
	if p == partYear {
		width = 4
	}
	if len(text) > width || v > hi {
		text = string(digit)
		v = int(digit - '0')
	}
	s.typed[p] = text
	if v >= lo {
		s.values[p] = Set(v)
	} else {
		s.values[p] = Optional[int]{}
	}
	f.change(p)
	return len(text) == width || v*10 > hi
}

func (f segmentField) step(p datePart, delta int) {
	lo, hi := f.limits(p)
	v := f.state.values[p]
	if !v.Set {
		v = Set(lo)
		if p == partYear {
			v.Value = time.Now().Year()
		}
	} else {
		v.Value += delta
		if v.Value > hi {
			v.Value = lo
		}
		if v.Value < lo {
			v.Value = hi
		}
	}
	f.state.values[p], f.state.typed[p] = v, ""
	f.change(p)
}

type segmentTarget struct {
	field segmentField
	part  datePart
}

func segmentTargets(fields []segmentField) []segmentTarget {
	var targets []segmentTarget
	for _, f := range fields {
		for _, p := range f.parts {
			if p.part != partLiteral {
				targets = append(targets, segmentTarget{f, p.part})
			}
		}
	}
	return targets
}

func segmentInput(c *ui.Context, targets []segmentTarget, index int, o InputOptions) func(ui.InputEvent) bool {
	target := targets[index]
	f, p := target.field, target.part
	s := f.state
	services := c.Services()
	direction := 1
	if c.Direction() == ui.RightToLeft {
		direction = -1
	}
	move := func(next int) {
		if next < 0 || next >= len(targets) {
			return
		}
		s.typed[p], s.selectAll = "", false
		dest := targets[next]
		dest.field.state.typed[dest.part] = ""
		dest.field.state.handles[dest.part].Focus(c)
	}
	command := func(name string) bool {
		switch name {
		case "selectAll":
			s.selectAll = true
			s.typed = [partCount]string{}
			for _, target := range targets {
				if target.field.state == s {
					s.handles[target.part].Focus(c)
					break
				}
			}
			return true
		case "copy":
			services.WriteClipboard(*f.text)
			return true
		case "cut":
			services.WriteClipboard(*f.text)
			if !o.ReadOnly {
				f.paste("")
			}
			return true
		case "paste", "pasteAndMatchStyle":
			if !o.ReadOnly {
				f.paste(services.ReadClipboard())
				s.typed[p] = ""
				s.selectAll = false
			}
			return true
		}
		return false
	}
	return func(ev ui.InputEvent) bool {
		if ev.Kind == ui.InputPointerDown {
			s.typed[p] = ""
			s.selectAll = false
			return false
		}
		if ev.Kind == ui.InputCommand {
			return command(ev.Text)
		}
		if ev.Kind == ui.InputText {
			if o.ReadOnly {
				return true
			}
			text := ev.Text
			digits := text != "" && strings.IndexFunc(text, func(r rune) bool { return r < '0' || r > '9' }) < 0
			if s.selectAll && !digits || len(text) > 1 && strings.ContainsAny(text, "/-:. ") {
				s.selectAll = false
				s.typed[p] = ""
				f.paste(text)
				return true
			}
			if s.selectAll && digits {
				s.selectAll = false
				s.values = [partCount]Optional[int]{}
				s.typed = [partCount]string{}
			}
			current := index
			for _, r := range text {
				dest := targets[current]
				field, part := dest.field, dest.part
				if part == partPeriod {
					if r == 'a' || r == 'A' || r == 'p' || r == 'P' {
						v := 0
						if r == 'p' || r == 'P' {
							v = 1
						}
						field.state.values[part] = Set(v)
						field.change(part)
					}
				} else if r >= '0' && r <= '9' {
					if field.numeric(part, byte(r)) && current+1 < len(targets) {
						current++
						move(current)
					}
				} else if r == '/' || r == '-' || r == ':' || unicode.IsSpace(r) {
					if current+1 < len(targets) {
						current++
						move(current)
					}
				}
			}
			return true
		}
		if ev.Kind != ui.InputKeyDown {
			return false
		}
		if ev.Mods == ui.Ctrl || ev.Mods == ui.Super {
			switch ev.Key {
			case ui.KeyA:
				return command("selectAll")
			case ui.KeyC:
				return command("copy")
			case ui.KeyX:
				return command("cut")
			case ui.KeyV:
				return command("paste")
			}
		}
		if ev.Mods != 0 {
			s.typed[p] = ""
			return false
		}
		switch ev.Key {
		case ui.KeyLeft:
			move(index - direction)
			return true
		case ui.KeyRight:
			move(index + direction)
			return true
		case ui.KeyTab:
			s.typed[p] = ""
			s.selectAll = false
			return false
		case ui.KeyEnter:
			valid := true
			for _, target := range targets {
				valid = valid && *target.field.err == ""
			}
			if valid && o.OnSubmit != nil {
				o.OnSubmit(*f.text)
			}
			return true
		case ui.KeyUp, ui.KeyDown:
			if !o.ReadOnly {
				delta := 1
				if ev.Key == ui.KeyDown {
					delta = -1
				}
				f.step(p, delta)
			}
			return true
		case ui.KeyHome, ui.KeyEnd:
			if !o.ReadOnly {
				lo, hi := f.limits(p)
				v := lo
				if ev.Key == ui.KeyEnd {
					v = hi
				}
				s.values[p], s.typed[p] = Set(v), ""
				f.change(p)
			}
			return true
		case ui.KeyBackspace, ui.KeyDelete:
			if !o.ReadOnly {
				if s.selectAll {
					f.paste("")
					s.selectAll = false
				} else {
					s.values[p], s.typed[p] = Optional[int]{}, ""
					f.change(p)
				}
			}
			return true
		}
		return false
	}
}

// segmentedInput shares the ordinary field recipe, while each segment owns its
// focus and stepper semantics. Style's editor argument is the first segment.
func segmentedInput(c *ui.Context, t Theme, fields []segmentField, o InputOptions, id Component, after func()) ui.Element {
	config := inputDefaults(t, id, o)
	m := buttonMetrics(t, config.Size)
	root := ui.Row(c).MinWidth(0).AlignItems(ui.Center).Disabled(o.Disabled || t.field.Disabled).Role(ui.RoleGroup).Label(o.Label).Font(t.Font).FontSize(m.Font).LineHeight(m.Line / m.Font)
	var first, leading, trailing ui.Element
	var segments []ui.Element
	var segmentStyles []PartStyle
	targets := segmentTargets(fields)
	gap, width := float32(2), float32(36)
	switch config.Size {
	case XS:
		gap, width = 1, 32
	case SM:
		width = 32
	case LG:
		gap = 3
	case XL:
		gap, width = 3, 40
	}
	root.Children(func() {
		if o.Loading {
			icon := o.LoadingIcon
			if icon == nil {
				icon = lucide.LoaderCircle()
			}
			leading = loadingSVG(c.Key("loading"), icon, m.Icon, t.Dimmed, t.motion(c), t.Motion.Spin)
		} else if o.Leading != nil {
			leading = ui.Icon(c.Key("leading"), o.Leading).Size(m.Icon, m.Icon).Shrink(0)
		}
		index := 0
		for n, f := range fields {
			fieldFocused := false
			for _, part := range f.parts {
				fieldFocused = fieldFocused || f.state.handles[part.part].Focused(c)
			}
			if !fieldFocused {
				f.state.selectAll = false
			}
			if n > 0 {
				ui.Icon(c.Key("separator"), lucide.Minus()).Size(t.text(16), t.text(16)).TextColor(t.Muted).Shrink(0)
			}
			group := ui.Row(c.Key(n)).Gap(t.spacing(gap)).AlignItems(ui.Center).Shrink(0)
			if len(fields) > 1 {
				group.Role(ui.RoleGroup).Label(f.label).Value(*f.text)
			}
			group.Children(func() {
				for i, part := range f.parts {
					p := part.part
					if p == partLiteral {
						ui.Text(c.Key(fmt.Sprintf("literal-%d", i)), part.literal).TextColor(t.Muted).Role(ui.RoleNone)
						continue
					}
					name := strings.TrimSpace(f.label + " " + message(c, segmentName(p)))
					w := width
					if p == partYear {
						w += 8
					}
					text := f.display(c, p)
					if p == partPeriod {
						w = max(w, float32(len([]rune(text)))*14)
					}
					e := ui.Box(c.Key(p)).Width(t.text(w)).MinHeight(m.Line).Center().Focusable().FocusRing(false).Cursor(ui.CursorText).ReadOnly(o.ReadOnly).Role(ui.RoleStepper).Label(name).Value(text).Bind(&f.state.handles[p]).TextCaret(ui.Rect{}).InputMethod(false)
					if !f.state.handles[p].Focused(c) {
						f.state.typed[p] = ""
					}
					lo, hi := f.limits(p)
					if v := f.state.values[p]; v.Set {
						e.Range(float64(lo), float64(hi), float64(v.Value)).Step(1)
					}
					semantics := o
					semantics.Label = name
					fieldSemantics(c, e, t, semantics)
					e.HandleInput(segmentInput(c, targets, index, o))
					index++
					if !first.Valid() {
						first = e
					}
					base := PartStyle{Radius: Set(t.radius(2))}
					if !f.state.values[p].Set {
						base.Foreground = Set(t.Dimmed)
					}
					if fieldInvalid(t, o) {
						base.Foreground = Set(t.Error)
					}
					if e.Focused() || f.state.selectAll {
						background := t.Elevated
						if config.Variant == Subtle || config.Variant == Soft {
							background = t.Accented
						}
						base.Background = Set(background)
					}
					e.Children(func() { ui.Text(c, text).Role(ui.RoleNone) })
					segments = append(segments, e)
					segmentStyles = append(segmentStyles, base)
				}
			})
		}
		if o.Clearable && !o.ReadOnly {
			hasText := false
			for _, f := range fields {
				hasText = hasText || *f.text != ""
			}
			if hasText {
				label := o.ClearLabel
				if label == "" {
					label = namedMessage(c, "clearNamed", o.Label)
				}
				ui.ButtonBase(c.Key("clear")).Label(label).Size(m.Icon, m.Icon).FocusRing(true).OnClick(func() {
					for _, f := range fields {
						f.paste("")
					}
					targets[0].field.state.handles[targets[0].part].Focus(c)
					if o.OnClear != nil {
						o.OnClear()
					}
				}).Children(func() { ui.Icon(c, lucide.X()).Size(m.Icon, m.Icon).TextColor(t.Muted) })
			}
		}
		if o.Trailing != nil {
			trailing = ui.Icon(c.Key("trailing"), o.Trailing).Size(m.Icon, m.Icon).Shrink(0)
		}
		if after != nil {
			after()
		}
	})
	focused := root.FocusWithin()
	_, styles := resolve(t, id, ComponentTheme{Color: Primary, Size: MD, Variant: Outline}, ComponentTheme{Color: o.Color, Size: inheritedSize(t, o.Size), Variant: o.Variant, UI: o.UI}, func(v ComponentTheme) Styles {
		s := inputRecipe(t, v, focused, root.Hovered() && !root.IsDisabled(), fieldInvalid(t, o))
		if o.Highlight && !fieldInvalid(t, o) && v.Variant != None {
			r := s[Root]
			r.BorderWidth, r.BorderColor = Set(float32(1)), Set(t.color(v.Color))
			s[Root] = r
		}
		return s
	})
	styles[Root].apply(root)
	styles[Leading].apply(leading)
	styles[Trailing].apply(trailing)
	for i, e := range segments {
		styles[Editor].apply(e)
		finish(t, id, Segment, e, segmentStyles[i], Decoration{UI: o.UI})
	}
	if focused && config.Variant != None {
		color := config.Color
		if fieldInvalid(t, o) {
			color = Error
		}
		root.Shadow(0, 0, 0, 3, t.focusColor(color))
	}
	if root.Clicked() && !root.IsDisabled() && !focused {
		targets[0].field.state.handles[targets[0].part].Focus(c)
	}
	applyGroup(c, root, t, styles[Root].Radius.Value)
	if o.Style != nil {
		o.Style(root, first, leading, trailing)
	}
	texts := make([]string, len(fields))
	for i, f := range fields {
		texts[i] = *f.text
	}
	root.Value(strings.Join(texts, " – "))
	// Display-only hour-cycle/seconds changes must not mark a form dirty or
	// invalidate an asynchronous validation ticket. Invalid drafts keep their text.
	for i, f := range fields {
		if f.clock && *f.err == "" && *f.text != "" {
			h := f.state.values[partHour].Value
			if f.twelve {
				h = h%12 + 12*f.state.values[partPeriod].Value
			}
			texts[i] = (ClockTime{h, f.state.values[partMinute].Value, f.state.values[partSecond].Value}).String()
		}
	}
	err, key := o.ErrorMessage, o.errorKey
	if o.Invalid && err == "" {
		err, key = message(c, "checkValue"), "checkValue"
	}
	trackFieldEdit(t, strings.Join(texts, "\x00"), err, key)
	return root
}
