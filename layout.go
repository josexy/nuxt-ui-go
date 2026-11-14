package nui

import "github.com/egoist/mygo/ui"

type ContainerOptions struct {
	MaxWidth, Padding Optional[float32]
	Decoration
}

func Container(c *ui.Context, t Theme, opts ContainerOptions, children func()) ui.Element {
	width := t.ContainerWidth
	if opts.MaxWidth.Set {
		width = opts.MaxWidth.Value
	}
	windowWidth, _ := c.Size()
	padding := float32(16)
	if windowWidth >= 1024 {
		padding = 32
	} else if windowWidth >= 640 {
		padding = 24
	}
	padding = t.spacing(padding)
	if opts.Padding.Set {
		padding = opts.Padding.Value
	}
	e := ui.Column(c).FillWidth().MaxWidth(width).AlignSelf(ui.Center).MinWidth(0)
	style := mergeStyle(PartStyle{PaddingX: Set(padding), Gap: Set(t.spacing(16))}, t.components[ComponentContainer].UI[Root])
	style.PaddingX = mergeValue(style.PaddingX, opts.Padding)
	decorated(Root, e, Styles{Root: mergeStyle(style, opts.UI[Root])}, opts.Style)
	if children != nil {
		e.Children(children)
	}
	return e
}
func Main(c *ui.Context, t Theme, children func()) ui.Element {
	// The window's main landmark, as Nuxt UI's <main>. A nested instance opts
	// out by replacing the role on the returned element.
	e := ui.Column(c).Grow(1).MinWidth(0).MinHeight(0).Role(ui.RoleMain)
	finish(t, ComponentMain, Root, e, PartStyle{Foreground: Set(t.Text), Font: Set(t.Font), FontSize: Set(t.text(14)), LineHeight: Set(float32(20.0 / 14))}, Decoration{})
	if children != nil {
		e.Children(children)
	}
	return e
}

type CardOptions struct {
	Variant            Variant
	Title, Description string
	Header, Footer     func()
	Decoration
}

func surface(t Theme, variant Variant) PartStyle {
	s := PartStyle{Radius: Set(t.radius(8)), Background: Set(t.Background), Foreground: Set(t.Text), BorderWidth: Set(float32(0)), BorderColor: Set(t.Border)}
	switch variant {
	case Solid:
		s.Background = Set(t.InvertedBackground)
		s.Foreground = Set(t.Inverted)
	case Soft:
		s.Background = Set(t.Elevated.Alpha(.5))
	case Subtle:
		s.Background = Set(t.Elevated.Alpha(.5))
		s.BorderWidth = Set(float32(1))
	default:
		s.BorderWidth = Set(float32(1))
	}
	return s
}
func Card(c *ui.Context, t Theme, opts CardOptions, body func()) ui.Element {
	windowWidth, _ := c.Size()
	contentPadding := float32(16)
	if windowWidth >= 640 {
		contentPadding = 24
	}
	config, styles := resolve(t, ComponentCard, ComponentTheme{Variant: Outline}, ComponentTheme{Variant: opts.Variant, UI: opts.UI}, func(v ComponentTheme) Styles {
		fg, muted := t.Highlighted, t.Muted
		if v.Variant == Solid {
			fg = t.Inverted
			muted = t.Dimmed
		}
		pad := PartStyle{PaddingX: Set(t.spacing(contentPadding)), PaddingY: Set(t.spacing(16)), Gap: Set(t.spacing(8))}
		header := mergeStyle(pad, PartStyle{Gap: Set(t.spacing(4))})
		return Styles{Root: surface(t, v.Variant), Header: header, Body: mergeStyle(pad, PartStyle{PaddingY: Set(t.spacing(contentPadding))}), Footer: pad, Label: {Foreground: Set(fg), Weight: Set(600), FontSize: Set(t.text(16)), LineHeight: Set(float32(24.0 / 16))}, Description: {Foreground: Set(muted), FontSize: Set(t.text(14))}}
	})
	e := ui.Column(c).MinWidth(0).Clip()
	styles[Root].apply(e)
	e.Children(func() {
		// Like Nuxt's divide-y, lines separate only the sections that render.
		sections := 0
		divide := func() {
			if sections > 0 && config.Variant != Solid {
				ui.Box(c.Key(sections)).Height(1).Material(lineMaterial{color: t.Border, hairline: true})
			}
			sections++
		}
		if opts.Header != nil || opts.Title != "" || opts.Description != "" {
			divide()
			h := ui.Column(c.Key("header"))
			styles[Header].apply(h)
			h.Children(func() {
				if opts.Header != nil {
					opts.Header()
				} else {
					if opts.Title != "" {
						decorated(Label, ui.Text(c, opts.Title), styles, opts.Style)
					}
					if opts.Description != "" {
						description := ui.Text(c, opts.Description)
						if opts.Title == "" {
							description.Margin(t.spacing(4), 0, 0, 0)
						}
						decorated(Description, description, styles, opts.Style)
					}
				}
			})
			if opts.Style != nil {
				opts.Style(Header, h)
			}
		}
		if body != nil {
			divide()
			b := ui.Column(c.Key("body")).MinWidth(0)
			styles[Body].apply(b)
			b.Children(body)
			if opts.Style != nil {
				opts.Style(Body, b)
			}
		}
		if opts.Footer != nil {
			divide()
			f := ui.Row(c.Key("footer")).AlignItems(ui.Center).Wrap()
			styles[Footer].apply(f)
			f.Children(opts.Footer)
			if opts.Style != nil {
				opts.Style(Footer, f)
			}
		}
	})
	if opts.Style != nil {
		opts.Style(Root, e)
	}
	return e
}

type SeparatorPosition uint8

const (
	SeparatorCenter SeparatorPosition = iota
	SeparatorStart
	SeparatorEnd
)

type SeparatorType uint8

const (
	SeparatorSolid SeparatorType = iota
	SeparatorDashed
	SeparatorDotted
)

type SeparatorOptions struct {
	Label       string
	Icon        *ui.SVG
	Avatar      *AvatarOptions
	Content     func()
	Color       Color
	Size        Size // XS..XL correspond to 1..5 DIP.
	Type        SeparatorType
	Position    SeparatorPosition
	Vertical    bool // Kept for existing callers; either this or Orientation can select vertical.
	Orientation Orientation
	Decoration
}

func Separator(c *ui.Context, t Theme, opts SeparatorOptions) ui.Element {
	return separator(c, t, opts, false)
}

func separator(c *ui.Context, t Theme, opts SeparatorOptions, hairline bool) ui.Element {
	if opts.Position > SeparatorEnd || opts.Type > SeparatorDotted {
		panic("nui: invalid separator position or type")
	}
	vert := vertical(opts.Orientation) || opts.Vertical
	config := selected(defaults(t, ComponentSeparator, ComponentTheme{Color: Neutral, Size: XS}), ComponentTheme{Color: opts.Color, Size: opts.Size})
	color := t.color(config.Color)
	if config.Color == Neutral {
		color = t.Border
	}
	thickness := float32(sizeIndex(config.Size) + 1)
	e := ui.Row(c).AlignItems(ui.Center).Gap(t.spacing(12))
	if vert {
		e.Column().Height(48).Gap(t.spacing(8))
	} else {
		e.FillWidth()
	}
	e.Children(func() {
		lineIndex := 0
		line := func() {
			l := ui.Box(c.Key(lineIndex)).Grow(1)
			lineIndex++
			if vert {
				l.Width(thickness)
			} else {
				l.Height(thickness)
			}
			style := mergeStyle(PartStyle{Background: Set(color)}, t.components[ComponentSeparator].UI[Indicator])
			style = mergeStyle(style, opts.UI[Indicator])
			style.apply(l)
			l.Material(lineMaterial{color: style.Background.Value, vertical: vert, hairline: hairline && config.Size == XS, pattern: opts.Type})
			if opts.Style != nil {
				opts.Style(Indicator, l)
			}
		}
		hasContent := opts.Content != nil || opts.Label != "" || opts.Icon != nil || opts.Avatar != nil
		content := func() {
			container := ui.Row(c.Key("content")).Shrink(0)
			finish(t, ComponentSeparator, Body, container, PartStyle{Foreground: Set(t.Text), Weight: Set(500)}, opts.Decoration)
			container.Children(func() {
				switch {
				case opts.Content != nil:
					opts.Content()
				case opts.Label != "":
					label := ui.Text(c, opts.Label).TextAlign(ui.Center)
					if !vert {
						label.SingleLine()
					}
					finish(t, ComponentSeparator, Label, label, PartStyle{FontSize: Set(t.text(14))}, opts.Decoration)
				case opts.Icon != nil:
					icon := Icon(c, t, opts.Icon, IconOptions{Size: Set(t.text(20)), Color: Set(t.Text)})
					finish(t, ComponentSeparator, Leading, icon, PartStyle{}, opts.Decoration)
				case opts.Avatar != nil:
					diameter := Optional[float32]{}
					if opts.Avatar.Size == SizeDefault {
						diameter = Set(t.text(20))
					}
					finish(t, ComponentSeparator, Leading, avatar(c, t, *opts.Avatar, diameter), PartStyle{}, opts.Decoration)
				}
			})
		}
		if hasContent && opts.Position == SeparatorStart {
			content()
		}
		line()
		if hasContent && opts.Position != SeparatorStart {
			content()
			if opts.Position == SeparatorCenter {
				line()
			}
		}
	})
	finish(t, ComponentSeparator, Root, e, PartStyle{}, opts.Decoration)
	return e
}
