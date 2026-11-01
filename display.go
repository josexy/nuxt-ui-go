package nui

import (
	"fmt"
	"runtime"
	"strings"
	"unicode"

	"github.com/egoist/mygo/ui"
)

type BadgeOptions struct {
	Color             Color
	Variant           Variant
	Size              Size
	Square            bool
	Leading, Trailing *ui.SVG
	Decoration
}

func Badge(c *ui.Context, t Theme, label string, opts BadgeOptions) ui.Element {
	config, styles := resolve(t, ComponentBadge, ComponentTheme{Color: Primary, Variant: Solid, Size: MD}, ComponentTheme{Color: opts.Color, Variant: opts.Variant, Size: opts.Size, UI: opts.UI}, func(v ComponentTheme) Styles {
		if v.Variant > Subtle {
			panic("nui: Badge supports solid, outline, soft and subtle")
		}
		metrics := [5]controlMetrics{{8, 12, 4, 2, 4, 12}, {10, 12, 6, 4, 4, 12}, {12, 16, 8, 4, 4, 16}, {14, 20, 8, 4, 6, 20}, {16, 24, 10, 4, 6, 24}}
		m := metrics[sizeIndex(v.Size)]
		m.Font = t.text(m.Font)
		m.Line = t.text(m.Line)
		m.X = t.spacing(m.X)
		m.Y = t.spacing(m.Y)
		m.Gap = t.spacing(m.Gap)
		style := mergeStyle(controlStyle(t, m), appearance(t, v.Color, v.Variant, false, false))
		if v.Size == XS || v.Size == SM {
			style.Radius = Set(t.radius(2))
		}
		if opts.Square {
			style.PaddingX = style.PaddingY
		}
		return Styles{Root: style}
	})
	e := ui.Row(c).AlignItems(ui.Center).Shrink(0)
	styles[Root].apply(e)
	size := t.text([5]float32{12, 12, 16, 20, 24}[sizeIndex(config.Size)])
	e.Children(func() {
		if opts.Leading != nil {
			decorated(Leading, ui.Icon(c, opts.Leading).Size(size, size), styles, opts.Style)
		}
		decorated(Label, ui.Text(c, label).SingleLine(), styles, opts.Style)
		if opts.Trailing != nil {
			decorated(Trailing, ui.Icon(c, opts.Trailing).Size(size, size), styles, opts.Style)
		}
	})
	if opts.Style != nil {
		opts.Style(Root, e)
	}
	return e
}

type AvatarOptions struct {
	Image     *ui.Bitmap
	Icon      *ui.SVG
	Text, Alt string
	Color     Color
	Size      Size
	Decoration
}

func avatarSize(t Theme, size Size) float32 {
	return t.text([5]float32{24, 28, 32, 36, 40}[sizeIndex(size)])
}

// Initials returns up to two Unicode initials; a one-word CJK name uses its first two runes.
func Initials(name string) string {
	fields := strings.Fields(name)
	if len(fields) == 0 {
		return "?"
	}
	first := []rune(fields[0])
	last := []rune(fields[len(fields)-1])
	r := []rune{first[0]}
	if len(fields) > 1 {
		r = append(r, last[0])
	} else if len(first) > 1 && unicode.Is(unicode.Han, first[0]) {
		r = append(r, first[1])
	}
	return strings.ToUpper(string(r))
}
func Avatar(c *ui.Context, t Theme, opts AvatarOptions) ui.Element {
	config, styles := resolve(t, ComponentAvatar, ComponentTheme{Color: Neutral, Size: MD}, ComponentTheme{Color: opts.Color, Size: opts.Size, UI: opts.UI}, func(v ComponentTheme) Styles {
		color, bg := t.color(v.Color), t.color(v.Color).Alpha(.10)
		if v.Color == Neutral {
			color, bg = t.Muted, t.Elevated
		}
		return Styles{Root: {Background: Set(bg), Foreground: Set(color), Radius: Set(float32(999)), Font: Set(t.Font), FontSize: Set(t.text([5]float32{12, 14, 16, 18, 20}[sizeIndex(v.Size)])), Weight: Set(500)}}
	})
	size := avatarSize(t, config.Size)
	e := ui.Box(c).Size(size, size).Shrink(0).Center().Clip()
	styles[Root].apply(e)
	alt := opts.Alt
	if alt == "" {
		alt = opts.Text
	}
	if alt != "" {
		e.Role(ui.RoleImage).Label(alt)
	}
	e.Children(func() {
		if opts.Image != nil {
			ui.Image(c, opts.Image).Size(size, size).Fit(ui.Cover)
		} else if opts.Icon != nil {
			decorated(Leading, ui.Icon(c, opts.Icon).Size(size*.6, size*.6), styles, opts.Style)
		} else {
			text := opts.Text
			if text == "" {
				// Nuxt derives fallback initials from alt, while text is literal.
				for word := range strings.SplitSeq(opts.Alt, " ") {
					for _, r := range word {
						text += string(r)
						break
					}
				}
				runes := []rune(text)
				text = string(runes[:min(2, len(runes))])
			}
			font := styles[Root].FontSize.Value
			width, _ := c.MeasureText(1000, ui.Span{Text: text, Font: t.Font, Size: font, Weight: 500})
			if width > size*.75 {
				font *= size * .75 / width
			}
			decorated(Label, ui.Text(c, text).FontSize(font).SingleLine(), styles, opts.Style)
		}
	})
	if opts.Style != nil {
		opts.Style(Root, e)
	}
	return e
}

type AvatarItem struct {
	ID     string
	Avatar AvatarOptions
}
type avatarItemKey string
type AvatarGroupOptions struct {
	Max     Optional[int]
	Size    Size
	Color   Color
	Overlap Optional[float32]
	Decoration
}

func AvatarGroup(c *ui.Context, t Theme, items []AvatarItem, opts AvatarGroupOptions) ui.Element {
	config := selected(defaults(t, ComponentAvatarGroup, ComponentTheme{Size: MD, Color: Neutral}), ComponentTheme{Size: opts.Size, Color: opts.Color})
	limit := len(items)
	if opts.Max.Set && opts.Max.Value > 0 {
		limit = opts.Max.Value
	}
	overlap := t.spacing([5]float32{2, 6, 6, 6, 8}[sizeIndex(config.Size)])
	ring := [5]float32{1, 2, 2, 2, 3}[sizeIndex(config.Size)]
	if opts.Overlap.Set {
		overlap = opts.Overlap.Value
	}
	seen := map[string]bool{}
	for _, item := range items {
		if item.ID == "" || seen[item.ID] {
			panic("nui: AvatarGroup requires unique non-empty IDs")
		}
		seen[item.ID] = true
	}
	// Reverse both iteration and layout: keep caller order while painting the
	// first avatar over its following neighbor, as Nuxt's row-reverse group does.
	e := ui.Row(c).Reverse().Justify(ui.End).AlignItems(ui.Center)
	finish(t, ComponentAvatarGroup, Root, e, PartStyle{Gap: Set(-overlap)}, opts.Decoration)
	e.Children(func() {
		if len(items) > limit {
			Avatar(c.Key("overflow"), t, AvatarOptions{Text: fmt.Sprintf("+%d", len(items)-limit), Size: config.Size, Color: config.Color,
				Alt: formatMessage(c, "morePeople", map[string]any{"count": len(items) - limit})}).Shadow(0, 0, 0, ring, t.Background)
		}
		for i := min(limit, len(items)) - 1; i >= 0; i-- {
			item := items[i]
			o := item.Avatar
			if o.Size == SizeDefault {
				o.Size = config.Size
			}
			if o.Color == ColorDefault {
				o.Color = config.Color
			}
			Avatar(c.Key(avatarItemKey(item.ID)), t, o).Shadow(0, 0, 0, ring, t.Background)
		}
	})
	return e
}

type ChipPosition uint8

const (
	TopRight ChipPosition = iota
	TopLeft
	BottomRight
	BottomLeft
)

type ChipOptions struct {
	Color         Color
	Size          Size
	Text          string
	Position      ChipPosition
	Inset, Hidden bool
	Decoration
}

func Chip(c *ui.Context, t Theme, opts ChipOptions, content func()) ui.Element {
	e := ui.Box(c).Shrink(0)
	e.Children(func() {
		if content != nil {
			content()
		}
		if opts.Hidden {
			return
		}
		config := selected(defaults(t, ComponentChip, ComponentTheme{Color: Primary, Size: MD}), ComponentTheme{Color: opts.Color, Size: opts.Size})
		h := t.text([5]float32{6, 7, 8, 9, 10}[sizeIndex(config.Size)])
		w := h
		font := h
		if opts.Text != "" {
			tw, _ := c.MeasureText(1000, ui.Span{Text: opts.Text, Font: t.Font, Size: font, Weight: 500})
			w = max(h, tw)
		}
		chip := ui.Box(c).Size(w, h).Radius(999).Center().IgnoreInput(true)
		if content != nil {
			chip.Absolute()
			left := opts.Position == TopLeft || opts.Position == BottomLeft
			bottom := opts.Position == BottomLeft || opts.Position == BottomRight
			x, y := w/2, -h/2
			if left {
				chip.Left(0)
				x = -x
			} else {
				chip.Right(0)
			}
			if bottom {
				chip.Bottom(0)
				y = -y
			} else {
				chip.Top(0)
			}
			if !opts.Inset {
				chip.Translate(x, y)
			}
		}
		_, styles := resolve(t, ComponentChip, ComponentTheme{Color: Primary, Size: MD}, ComponentTheme{Color: opts.Color, Size: opts.Size, UI: opts.UI}, func(v ComponentTheme) Styles {
			return Styles{Indicator: {Background: Set(t.color(v.Color)), Foreground: Set(t.Inverted), FontSize: Set(font), Weight: Set(500)}}
		})
		decorated(Indicator, chip, styles, opts.Style)
		chip.Shadow(0, 0, 0, 1, t.Background)
		if opts.Text != "" {
			chip.Children(func() { ui.Text(c, opts.Text).SingleLine() })
		}
	})
	finish(t, ComponentChip, Root, e, PartStyle{}, opts.Decoration)
	return e
}

// FormatKey formats familiar key tokens for a target GOOS without registering a shortcut.
func FormatKey(key, goos string) string {
	mac := goos == "darwin"
	switch strings.ToLower(key) {
	case "cmdorctrl", "primary":
		if mac {
			return "⌘"
		}
		return "Ctrl"
	case "cmd", "meta", "super":
		if mac {
			return "⌘"
		}
		return "Win"
	case "ctrl", "control":
		if mac {
			return "⌃"
		}
		return "Ctrl"
	case "alt", "option":
		if mac {
			return "⌥"
		}
		return "Alt"
	case "shift":
		if mac {
			return "⇧"
		}
		return "Shift"
	case "enter", "return":
		return "↵"
	case "escape", "esc":
		return "Esc"
	}
	return key
}

type KbdOptions struct {
	Size     Size
	Color    Color
	Variant  Variant
	Platform string
	Decoration
}

func Kbd(c *ui.Context, t Theme, keys []string, opts KbdOptions) ui.Element {
	platform := opts.Platform
	if platform == "" {
		platform = runtime.GOOS
	}
	config, styles := resolve(t, ComponentKbd, ComponentTheme{Size: MD, Color: Neutral, Variant: Outline}, ComponentTheme{Size: opts.Size, Color: opts.Color, Variant: opts.Variant, UI: opts.UI}, func(v ComponentTheme) Styles {
		return Styles{Label: mergeStyle(appearance(t, v.Color, v.Variant, false, false), PartStyle{Font: Set(t.Font), FontSize: Set(t.text([5]float32{10, 10, 11, 12, 12}[sizeIndex(v.Size)])), Weight: Set(500)})}
	})
	idx := sizeIndex(config.Size)
	height := t.text([5]float32{16, 16, 20, 24, 24}[idx])
	e := ui.Row(c).Gap(t.spacing(4))
	e.Children(func() {
		for _, key := range keys {
			chip := ui.Box(c).Height(height).MinWidth(height).PaddingX(t.spacing(4)).Radius(t.radius(2)).Center()
			decorated(Label, chip, styles, opts.Style)
			chip.Children(func() { ui.Text(c, FormatKey(key, platform)).SingleLine() })
		}
	})
	finish(t, ComponentKbd, Root, e, PartStyle{}, opts.Decoration)
	return e
}

type UserOptions struct {
	Name, Description  string
	Avatar             AvatarOptions
	Size               Size
	Vertical, Disabled bool
	OnActivate         func()
	Decoration
}

func User(c *ui.Context, t Theme, opts UserOptions) ui.Element {
	var e ui.Element
	if opts.OnActivate != nil {
		e = ui.ButtonBase(c).Disabled(opts.Disabled).FocusRing(false).OnClick(opts.OnActivate).Label(opts.Name)
	} else {
		e = ui.Row(c).Disabled(opts.Disabled)
	}
	e.AlignItems(ui.Center).MinWidth(0)
	if opts.Vertical {
		e.Column().AlignItems(ui.Start)
	}
	config, styles := resolve(t, ComponentUser, ComponentTheme{Size: MD}, ComponentTheme{Size: opts.Size, UI: opts.UI}, func(v ComponentTheme) Styles {
		return Styles{Root: {Gap: Set(t.spacing([5]float32{6, 8, 8, 10, 10}[sizeIndex(v.Size)])), Radius: Set(t.Radius)}, Label: {Foreground: Set(t.Highlighted), FontSize: Set(t.text([5]float32{12, 12, 14, 14, 16}[sizeIndex(v.Size)])), Weight: Set(500)}, Description: {Foreground: Set(t.Muted), FontSize: Set(t.text([5]float32{12, 12, 12, 14, 14}[sizeIndex(v.Size)]))}}
	})
	decorated(Root, e, styles, opts.Style)
	if e.FocusVisible() {
		e.Shadow(0, 0, 0, 3, t.focusColor(Primary))
	}
	e.Children(func() {
		avatar := opts.Avatar
		avatar.Size = config.Size
		if avatar.Alt == "" {
			avatar.Alt = opts.Name
		}
		Avatar(c, t, avatar)
		wrapper := ui.Column(c).MinWidth(0).Grow(1)
		if config.Size == XS {
			wrapper.Row().AlignItems(ui.Center).Gap(t.spacing(6))
		}
		wrapper.Children(func() {
			decorated(Label, ui.Text(c, opts.Name).SingleLine(), styles, opts.Style)
			if opts.Description != "" {
				decorated(Description, ui.Text(c, opts.Description).SingleLine(), styles, opts.Style)
			}
		})
	})
	return e
}
