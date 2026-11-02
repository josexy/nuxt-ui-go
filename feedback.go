package nui

import (
	"fmt"
	"math"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

type AlertOptions struct {
	Title, Description string
	Color              Color
	Variant            Variant
	Icon               *ui.SVG
	Actions            func()
	Orientation        Optional[Orientation] // Defaults to Vertical.
	OnClose            func()
	Decoration
}

func Alert(c *ui.Context, t Theme, opts AlertOptions) ui.Element {
	stacked := vertical(valueOr(opts.Orientation, Vertical))
	_, styles := resolve(t, ComponentAlert, ComponentTheme{Color: Primary, Variant: Solid}, ComponentTheme{Color: opts.Color, Variant: opts.Variant, UI: opts.UI}, func(v ComponentTheme) Styles {
		if v.Variant > Subtle {
			panic("nui: Alert supports solid, outline, soft and subtle")
		}
		root := appearance(t, v.Color, v.Variant, false, false)
		// Alert's recipe differs from Button for outline rings and neutral surfaces.
		if v.Variant == Outline && v.Color != Neutral {
			root.BorderColor = Set(t.color(v.Color).Alpha(.25))
		}
		if v.Color == Neutral && v.Variant != Solid {
			root.Foreground = Set(t.Highlighted)
			if v.Variant == Soft || v.Variant == Subtle {
				root.Background = Set(t.Elevated.Alpha(.5))
			}
			if v.Variant == Outline {
				root.BorderColor = Set(t.Border)
			}
		}
		root.Radius = Set(t.radius(8))
		root.PaddingX = Set(t.spacing(16))
		root.PaddingY = Set(t.spacing(16))
		root.Gap = Set(t.spacing(10))
		root.FontSize = Set(t.text(14))
		root.LineHeight = Set(float32(20.0 / 14))
		return Styles{Root: root, Label: {Weight: Set(500)}, Description: {Opacity: Set(float32(.9))}}
	})
	e := ui.Row(c).FillWidth().MinWidth(0).AlignItems(ui.Start).Role(ui.RoleStatus).Label(opts.Title)
	if !stacked {
		e.AlignItems(ui.Center)
	}
	styles[Root].apply(e)
	e.Children(func() {
		if opts.Icon != nil {
			decorated(Leading, ui.Icon(c, opts.Icon).Size(t.text(20), t.text(20)).Shrink(0), styles, opts.Style)
		}
		ui.Column(c).Grow(1).MinWidth(0).Gap(t.spacing(4)).Children(func() {
			if opts.Title != "" {
				decorated(Label, ui.Text(c, opts.Title), styles, opts.Style)
			}
			if opts.Description != "" {
				decorated(Description, ui.Text(c, opts.Description), styles, opts.Style)
			}
			if opts.Actions != nil && stacked {
				ui.Row(c).Wrap().Gap(t.spacing(6)).Margin(t.spacing(6), 0, 0, 0).Children(opts.Actions)
			}
		})
		if opts.Actions != nil && !stacked {
			ui.Row(c.Key("actions")).Wrap().Gap(t.spacing(6)).Children(opts.Actions)
		}
		if opts.OnClose != nil {
			// Nuxt uses an unpadded md link button: no hover surface that could
			// hide the icon. It keeps the alert's text color and dims on hover.
			fg := styles[Root].Foreground.Value
			close := Button(c, t, "", ButtonOptions{AccessibleLabel: namedMessage(c, "dismissNamed", opts.Title), Color: Neutral, Variant: LinkVariant, Size: MD, Square: true, Leading: lucide.X(), UI: Styles{Root: {Foreground: Set(fg), PaddingX: Set(float32(0)), PaddingY: Set(float32(0))}}, Style: func(root, label, leading, trailing ui.Element) {
				if !root.IsDisabled() && (root.Hovered() || root.Pressed()) {
					root.TextColor(fg.Alpha(.75))
				}
			}}).OnClick(opts.OnClose)
			decorated(Trailing, close, styles, opts.Style)
		}
	})
	if opts.Style != nil {
		opts.Style(Root, e)
	}
	return e
}

type EmptyOptions struct {
	Title, Description string
	Icon               *ui.SVG
	Size               Size
	Variant            Variant
	Loading            bool
	LoadingIcon        *ui.SVG
	Actions, Footer    func()
	Decoration
}

func Empty(c *ui.Context, t Theme, opts EmptyOptions) ui.Element {
	config, styles := resolve(t, ComponentEmpty, ComponentTheme{Variant: Outline, Size: MD}, ComponentTheme{Variant: opts.Variant, Size: opts.Size, UI: opts.UI}, func(v ComponentTheme) Styles {
		s := surface(t, v.Variant)
		s.PaddingX = Set(t.spacing(32))
		s.PaddingY = Set(t.spacing(32))
		s.Gap = Set(t.spacing(16))
		fg, muted := t.Highlighted, t.Muted
		switch v.Variant {
		case Solid:
			fg, muted = t.Inverted, t.Dimmed
		case Soft, Subtle:
			muted = t.Toned
		case None:
			s.Background, s.BorderWidth = Set(ui.Transparent), Set(float32(0))
		}
		return Styles{Root: s, Label: {Foreground: Set(fg), FontSize: Set(t.text([5]float32{14, 14, 16, 16, 18}[sizeIndex(v.Size)])), Weight: Set(500)}, Description: {Foreground: Set(muted), FontSize: Set(t.text([5]float32{12, 12, 14, 14, 16}[sizeIndex(v.Size)]))}}
	})
	e := ui.Column(c).FillWidth().MinWidth(0).AlignItems(ui.Center)
	styles[Root].apply(e)
	e.Children(func() {
		// Nuxt's empty avatar: the icon at half the size of an elevated circle.
		size := t.text([5]float32{32, 36, 40, 44, 48}[sizeIndex(config.Size)])
		if opts.Loading || opts.Icon != nil {
			circle := ui.Box(c.Key("avatar")).Size(size, size).Radius(999).Background(t.Elevated).Center().Shrink(0)
			circle.Children(func() {
				if opts.Loading {
					icon := opts.LoadingIcon
					if icon == nil {
						icon = lucide.LoaderCircle()
					}
					decorated(Leading, loadingSVG(c, icon, size/2, t.Muted, t.motion(c), t.Motion.Spin), styles, opts.Style)
				} else {
					decorated(Leading, ui.Icon(c, opts.Icon).Size(size/2, size/2).TextColor(t.Muted), styles, opts.Style)
				}
			})
		}
		if opts.Title != "" || opts.Description != "" {
			ui.Column(c.Key("header")).MinWidth(0).AlignItems(ui.Center).Gap(t.spacing(8)).Children(func() {
				if opts.Title != "" {
					decorated(Label, ui.Text(c, opts.Title).TextAlign(ui.Center), styles, opts.Style)
				}
				if opts.Description != "" {
					decorated(Description, ui.Text(c, opts.Description).MaxWidth(384).TextAlign(ui.Center), styles, opts.Style)
				}
			})
		}
		if opts.Actions != nil {
			ui.Row(c).Wrap().Justify(ui.Center).Gap(t.spacing(8)).Children(opts.Actions)
		}
		if opts.Footer != nil {
			ui.Column(c).AlignItems(ui.Center).Children(opts.Footer)
		}
	})
	if opts.Style != nil {
		opts.Style(Root, e)
	}
	return e
}

type SkeletonOptions struct {
	Width, Height Optional[float32]
	Animate       Optional[bool]
	Decoration
}

func Skeleton(c *ui.Context, t Theme, opts SkeletonOptions) ui.Element {
	w, h := float32(160), t.text(16)
	if opts.Width.Set {
		w = opts.Width.Value
	}
	if opts.Height.Set {
		h = opts.Height.Value
	}
	e := ui.Box(c).Size(w, h).Role(ui.RoleStatus).Label(message(c, "loading"))
	animated := t.motion(c)
	if opts.Animate.Set {
		animated = animated && opts.Animate.Value
	}
	style := mergeStyle(PartStyle{Radius: Set(t.Radius), Background: Set(t.Elevated)}, t.components[ComponentSkeleton].UI[Root])
	style = mergeStyle(style, opts.UI[Root])
	color := style.Background.Value
	radius := style.Radius.Value
	style.Background = Set(ui.Transparent)
	style.apply(e)
	e.Draw(func(p *ui.Painter, r ui.Rect) {
		alpha := float32(1)
		if animated {
			p.AnimationFrame()
			alpha = .5 + .5*float32((1+math.Cos(2*math.Pi*phase(p.Now(), t.Motion.Pulse)))/2)
		}
		p.Fill(r, color.Alpha(alpha), radius)
	})
	if opts.Style != nil {
		opts.Style(Root, e)
	}
	return e
}
func phase(now time.Time, period time.Duration) float64 {
	if period <= 0 {
		period = time.Second
	}
	return float64(now.UnixNano()%int64(period)) / float64(period)
}
func finite(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	return v
}
func progressMax(v Optional[float64]) float64 {
	max := float64(100)
	if v.Set {
		max = finite(v.Value)
	}
	if max <= 0 {
		panic("nui: progress Max must be finite and positive")
	}
	return max
}

type ProgressOptions struct {
	Label                   string
	Color                   Color
	Size                    Size
	Max                     Optional[float64]
	Indeterminate, Vertical bool
	Orientation             Orientation
	Inverted                bool
	Animate                 Optional[bool]
	Decoration
}

func Progress(c *ui.Context, t Theme, value float64, opts ProgressOptions) ui.Element {
	opts.Vertical = vertical(opts.Orientation) || opts.Vertical
	config := selected(defaults(t, ComponentProgress, ComponentTheme{Color: Primary, Size: MD}), ComponentTheme{Color: opts.Color, Size: opts.Size})
	max := progressMax(opts.Max)
	value = math.Max(0, math.Min(max, finite(value)))
	ratio := float32(math.Round(value/max*100) / 100)
	thickness := t.spacing([5]float32{2, 4, 8, 12, 16}[sizeIndex(config.Size)])
	e := ui.Box(c).Radius(999).Clip().Role(ui.RoleProgress).Label(opts.Label)
	if opts.Vertical {
		e.Width(thickness).Height(160)
	} else {
		e.FillWidth().Height(thickness)
	}
	if !opts.Indeterminate {
		e.Range(0, max, value)
	} else {
		// MyGo represents unknown progress with Now < Min, matching ui.Progress.
		e.Range(0, max, -1).Value(message(c, "inProgress")).Description(message(c, "inProgress"))
	}
	finish(t, ComponentProgress, Root, e, PartStyle{Background: Set(t.Accented)}, opts.Decoration)
	color := t.color(config.Color)
	if s := t.components[ComponentProgress].UI[Indicator]; s.Background.Set && opts.Color == ColorDefault {
		color = s.Background.Value
	}
	if s := opts.UI[Indicator]; s.Background.Set {
		color = s.Background.Value
	}
	animated := t.motion(c)
	if opts.Animate.Set {
		animated = animated && opts.Animate.Value
	}
	reverse := opts.Inverted
	if !opts.Vertical && c.Direction() == ui.RightToLeft {
		reverse = !reverse
	}
	if animated && !opts.Indeterminate {
		ratio = e.AnimateWith("nui-progress", ratio, 200*time.Millisecond, ui.EaseOut)
	}
	e.Draw(func(p *ui.Painter, r ui.Rect) {
		fraction, start := ratio, float32(0)
		if opts.Indeterminate {
			fraction = .3
			start = .35
			if animated {
				p.AnimationFrame()
				start = float32(phase(p.Now(), 2*time.Second))*1.3 - .3
			}
		}
		fill := r
		if opts.Vertical {
			fill.H = r.H * fraction
			fill.Y = r.Y + r.H*start
			if reverse {
				fill.Y = r.Y + r.H - fill.H - r.H*start
			}
		} else {
			fill.W = r.W * fraction
			fill.X = r.X + r.W*start
			if reverse {
				fill.X = r.X + r.W - fill.W - r.W*start
			}
		}
		p.Clip(r, 999, func() { p.Fill(fill, color, 999) })
	})
	return e
}

type ProgressSegment struct {
	ID, Label string
	Value     float64
	Color     Color
}
type ProgressGroupOptions struct {
	Label       string
	Max         Optional[float64]
	Color       Color
	Size        Size
	HideLegend  bool
	Orientation Orientation
	Height      float32
	Decoration
}

func ProgressGroup(c *ui.Context, t Theme, segments []ProgressSegment, opts ProgressGroupOptions) ui.Element {
	vert := vertical(opts.Orientation)
	config := selected(defaults(t, ComponentProgressGroup, ComponentTheme{Color: Primary, Size: MD}), ComponentTheme{Color: opts.Color, Size: opts.Size})
	max := progressMax(opts.Max)
	items := append([]ProgressSegment(nil), segments...)
	// Normalize before summing so even multiple MaxFloat64 values stay finite.
	scale := max
	seen := map[string]bool{}
	for i := range items {
		if items[i].ID == "" || seen[items[i].ID] {
			panic("nui: ProgressGroup requires unique non-empty IDs")
		}
		seen[items[i].ID] = true
		items[i].Value = math.Min(max, math.Max(0, finite(items[i].Value)))
		scale = math.Max(scale, items[i].Value)
		if items[i].Color == ColorDefault {
			items[i].Color = config.Color
		}
	}
	sum := float64(0)
	hasLegend := false
	for _, item := range items {
		sum += item.Value / scale
		hasLegend = hasLegend || item.Label != ""
	}
	denom := math.Max(max/scale, sum)
	value := max
	if sum < max/scale {
		value = sum * scale
	}
	thickness := t.spacing([5]float32{2, 4, 8, 12, 16}[sizeIndex(config.Size)])
	e := ui.Column(c).Gap(t.spacing(8)).MinWidth(0)
	rtl := !vert && c.Direction() == ui.RightToLeft
	if vert {
		e.Row().AlignItems(ui.Start)
	}
	e.Children(func() {
		bar := ui.Box(c).FillWidth().Height(thickness).Radius(999).Clip().Role(ui.RoleProgress).Label(opts.Label).Range(0, max, value).Background(t.Accented)
		if vert {
			height := opts.Height
			if height <= 0 {
				height = 160
			}
			bar.Width(thickness).Height(height)
		}
		bar.Draw(func(p *ui.Painter, r ui.Rect) {
			p.Clip(r, 999, func() {
				at := r.X
				if vert {
					at = r.Y
				} else if rtl {
					at += r.W
				}
				for _, item := range items {
					if vert {
						h := r.H * float32((item.Value/scale)/denom)
						p.Fill(ui.Rect{X: r.X, Y: at, W: r.W, H: h}, t.color(item.Color), 0)
						at += h
						continue
					}
					w := r.W * float32((item.Value/scale)/denom)
					if rtl {
						at -= w
					}
					p.Fill(ui.Rect{X: at, Y: r.Y, W: w, H: r.H}, t.color(item.Color), 0)
					if !rtl {
						at += w
					}
				}
			})
		})
		finish(t, ComponentProgressGroup, Indicator, bar, PartStyle{}, opts.Decoration)
		if !opts.HideLegend && hasLegend {
			ui.Column(c).Gap(t.spacing(4)).Grow(1).MinWidth(0).Children(func() {
				for _, item := range items {
					ui.Row(c.Key(item.ID)).Gap(t.spacing(6)).AlignItems(ui.Center).Children(func() {
						ui.Box(c).Size(8, 8).Radius(999).Background(t.color(item.Color))
						ui.Text(c, item.Label).SingleLine().Grow(1).MinWidth(0)
						ui.Text(c, fmt.Sprintf("%.0f%%", math.Round(item.Value/max*100))).TextColor(t.Dimmed)
					})
				}
			})
		}
	})
	finish(t, ComponentProgressGroup, Root, e, PartStyle{}, opts.Decoration)
	return e
}
