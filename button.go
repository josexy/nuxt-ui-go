package nui

import (
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

type ButtonOptions struct {
	Variant Variant
	Color   Color
	Size    Size
	// Neutral selects the neutral palette; Color: Neutral is preferred.
	Neutral, Disabled, Loading, Block, Square bool
	AccessibleLabel                           string
	Leading, Trailing                         *ui.SVG
	LoadingIcon                               *ui.SVG
	LoadingTrailing                           bool // Place loading at the trailing edge unless Leading is set.
	Active                                    bool
	ActiveColor                               Color
	ActiveVariant                             Variant
	Animate                                   Optional[bool]
	UI                                        Styles
	Style                                     func(root, label, leading, trailing ui.Element)
}

type controlMetrics struct{ Font, Line, X, Y, Gap, Icon float32 }

func buttonMetrics(t Theme, size Size) controlMetrics {
	metrics := [5]controlMetrics{{12, 16, 8, 4, 4, 16}, {12, 16, 10, 6, 6, 16}, {14, 20, 10, 6, 6, 20}, {14, 20, 12, 8, 8, 20}, {16, 24, 12, 8, 8, 24}}
	m := metrics[sizeIndex(size)]
	m.Font = t.text(m.Font)
	m.Line = t.text(m.Line)
	m.X = t.spacing(m.X)
	m.Y = t.spacing(m.Y)
	m.Gap = t.spacing(m.Gap)
	m.Icon = t.text(m.Icon)
	return m
}
func controlStyle(t Theme, m controlMetrics) PartStyle {
	return PartStyle{Font: Set(t.Font), FontSize: Set(m.Font), LineHeight: Set(m.Line / m.Font), Weight: Set(500), Radius: Set(t.Radius), PaddingX: Set(m.X), PaddingY: Set(m.Y), Gap: Set(m.Gap)}
}

func Button(c *ui.Context, t Theme, label string, opts ButtonOptions) ui.Element {
	chosen := Optional[bool]{}
	if opts.Active {
		chosen = Set(true)
	}
	return button(c, t, label, opts, chosen)
}

// Apply selection semantics before the Nuxt recipe: MyGo Selected also changes
// colors, which must not override an explicit active variant or color.
func button(c *ui.Context, t Theme, label string, opts ButtonOptions, chosen Optional[bool]) ui.Element {
	opts.Size = inheritedSize(t, opts.Size)
	opts.Disabled = opts.Disabled || t.field.Disabled
	if label == "" && opts.AccessibleLabel == "" {
		panic("nui: icon-only Button requires AccessibleLabel")
	}
	if opts.Neutral {
		opts.Color = Neutral
	}
	if chosen.Set && chosen.Value {
		if opts.ActiveColor != ColorDefault {
			opts.Color = opts.ActiveColor
		}
		if opts.ActiveVariant != VariantDefault {
			opts.Variant = opts.ActiveVariant
		}
	}
	e := ui.ButtonBase(c).Disabled(opts.Disabled || opts.Loading).FocusRing(false)
	if chosen.Set {
		e.Selected(chosen.Value)
	}
	hover := !e.IsDisabled() && (e.Hovered() || e.Pressed())
	focus := e.FocusVisible()
	config, styles := resolve(t, ComponentButton, ComponentTheme{Color: Primary, Size: MD, Variant: Solid}, ComponentTheme{Color: opts.Color, Size: opts.Size, Variant: opts.Variant, UI: opts.UI}, func(v ComponentTheme) Styles {
		m := buttonMetrics(t, v.Size)
		root := mergeStyle(controlStyle(t, m), appearance(t, v.Color, v.Variant, hover, focus))
		if opts.Square || label == "" {
			root.PaddingX = root.PaddingY
		}
		return Styles{Root: root}
	})
	m := buttonMetrics(t, config.Size)
	if t.motion(c) && t.Motion.Transition > 0 {
		s := styles[Root]
		if s.Background.Set {
			s.Background.Value = e.AnimateColor("background", s.Background.Value, ui.Animation{Duration: t.Motion.Transition})
		}
		styles[Root] = s
	}
	styles[Root].apply(e)
	if opts.Block {
		e.FillWidth().Justify(ui.Center)
	}
	if opts.AccessibleLabel != "" {
		e.Label(opts.AccessibleLabel)
	} else {
		e.Label(label)
	}
	if focus {
		e.Shadow(0, 0, 0, 3, t.focusColor(config.Color))
	}
	var labelPart, leadingPart, trailingPart ui.Element
	animated := t.motion(c)
	if opts.Animate.Set {
		animated = animated && opts.Animate.Value
	}
	e.Children(func() {
		loadingLeading := opts.Loading && (!opts.LoadingTrailing || opts.Leading != nil)
		spinner := opts.LoadingIcon
		if spinner == nil {
			spinner = lucide.LoaderCircle()
		}
		if loadingLeading {
			leadingPart = loadingSVG(c, spinner, m.Icon, mergeValue(styles[Root].Foreground, styles[Leading].Foreground).Value, animated, t.Motion.Spin)
		} else if opts.Leading != nil {
			leadingPart = ui.Icon(c, opts.Leading).Size(m.Icon, m.Icon).Shrink(0)
		}
		if label != "" {
			labelPart = ui.Text(c, label).SingleLine().MinWidth(0)
		}
		if opts.Loading && !loadingLeading {
			trailingPart = loadingSVG(c, spinner, m.Icon, mergeValue(styles[Root].Foreground, styles[Trailing].Foreground).Value, animated, t.Motion.Spin)
		} else if opts.Trailing != nil {
			trailingPart = ui.Icon(c, opts.Trailing).Size(m.Icon, m.Icon).Shrink(0)
		}
		if opts.Block && trailingPart.Valid() {
			// Nuxt's ms-auto: a block's trailing icon ends the row.
			trailingPart.MarginStart(ui.Auto)
		}
	})
	styles[Label].apply(labelPart)
	styles[Leading].apply(leadingPart)
	styles[Trailing].apply(trailingPart)
	applyGroup(c, e, t, styles[Root].Radius.Value)
	if opts.Style != nil {
		opts.Style(e, labelPart, leadingPart, trailingPart)
	}
	return e
}

// Paint-time scheduling stops automatically when clipped or absent from the scene.
func loadingSVG(c *ui.Context, icon *ui.SVG, size float32, color ui.Color, animated bool, period time.Duration) ui.Element {
	return ui.Box(c).Size(size, size).Shrink(0).Draw(func(p *ui.Painter, r ui.Rect) {
		angle := float32(0)
		if animated {
			p.AnimationFrame()
			angle = float32(phase(p.Now(), period)) * 360
		}
		p.Transform(ui.Rotation(angle), r.X+r.W/2, r.Y+r.H/2, func() { p.Icon(icon, r, color) })
	})
}

type LinkOptions struct {
	URL              string
	Color            Color
	Active, Disabled bool
	Decoration
}

// Link opens URL only on activation. With URL empty, attach an application OnClick.
func Link(c *ui.Context, t Theme, label string, opts LinkOptions) ui.Element {
	e := ui.ButtonBase(c).Role(ui.RoleLink).Label(label).Disabled(opts.Disabled).FocusRing(false)
	config, styles := resolve(t, ComponentLink, ComponentTheme{Color: Primary}, ComponentTheme{Color: opts.Color, UI: opts.UI}, func(v ComponentTheme) Styles {
		foreground := t.Muted
		if opts.Active {
			foreground = t.color(v.Color)
		} else if !e.IsDisabled() && e.Hovered() {
			foreground = t.Text
		}
		return Styles{Root: {Foreground: Set(foreground), Font: Set(t.Font), FontSize: Set(t.text(14)), LineHeight: Set(float32(20.0 / 14)), Radius: Set(t.Radius)}}
	})
	decorated(Root, e, styles, opts.Style)
	if e.FocusVisible() {
		e.Shadow(0, 0, 0, 3, t.focusColor(config.Color))
	}
	e.Children(func() {
		text := ui.Text(c, label).SingleLine()
		decorated(Label, text, styles, opts.Style)
	})
	if opts.URL != "" {
		services := c.Services()
		e.OnClick(func() { services.OpenURL(opts.URL) })
	}
	return e
}

type IconOptions struct {
	Size  Optional[float32]
	Color Optional[ui.Color]
	Label string
	Decoration
}

func Icon(c *ui.Context, t Theme, svg *ui.SVG, opts IconOptions) ui.Element {
	size := t.text(20)
	if opts.Size.Set {
		size = opts.Size.Value
	}
	e := ui.Icon(c, svg).Size(size, size).Shrink(0)
	if opts.Label != "" {
		e.Role(ui.RoleImage).Label(opts.Label)
	}
	base := t.components[ComponentIcon].UI[Root]
	base.Foreground = mergeValue(base.Foreground, opts.Color)
	base = mergeStyle(base, opts.UI[Root])
	base.apply(e)
	if opts.Style != nil {
		opts.Style(Root, e)
	}
	return e
}
