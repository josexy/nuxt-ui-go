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
	Avatar             *AvatarOptions // Takes precedence over Icon; defaults to 44 DIP.
	Actions            func()
	Orientation        Optional[Orientation] // Defaults to Vertical.
	OnClose            func()
	Close              *ButtonOptions // A non-nil value also displays the close button without OnClose.
	CloseIcon          *ui.SVG
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
		return Styles{Root: root, Label: {Weight: Set(500)}, Description: {Opacity: Set(float32(.9))}, Actions: {Gap: Set(t.spacing(6))}}
	})
	e := ui.Row(c).FillWidth().MinWidth(0).AlignItems(ui.Start).Clip().Role(ui.RoleStatus).Label(opts.Title)
	if !stacked {
		e.AlignItems(ui.Center)
	}
	styles[Root].apply(e)
	e.Children(func() {
		if opts.Avatar != nil {
			diameter := Optional[float32]{}
			if opts.Avatar.Size == SizeDefault {
				diameter = Set(t.text(44))
			}
			decorated(Leading, avatar(c.Key("leading"), t, *opts.Avatar, diameter), styles, opts.Style)
		} else if opts.Icon != nil {
			decorated(Leading, ui.Icon(c.Key("leading"), opts.Icon).Size(t.text(20), t.text(20)).Shrink(0), styles, opts.Style)
		}
		// Text, actions and close retain the same parents in both orientations.
		// Moving cells instead of reparenting preserves action focus and drafts.
		closable := opts.OnClose != nil || opts.Close != nil
		columns := []ui.Track{ui.Fr(1)}
		if !stacked && opts.Actions != nil {
			columns = append(columns, ui.FitContent())
		}
		if closable {
			columns = append(columns, ui.FitContent())
		}
		content := ui.Grid(c.Key("content")).Grow(1).MinWidth(0).ColumnTracks(columns...).AlignItems(ui.Start)
		if !stacked {
			content.AlignItems(ui.Center)
		}
		startMargin := func(part ui.Element, gap float32) {
			if c.Direction() == ui.RightToLeft {
				part.Margin(0, gap, 0, 0)
			} else {
				part.Margin(0, 0, 0, gap)
			}
		}
		content.Children(func() {
			body := ui.Column(c.Key("body")).MinWidth(0).Gap(t.spacing(4)).ColumnStart(1).RowStart(1)
			decorated(Body, body, styles, opts.Style)
			body.Children(func() {
				if opts.Title != "" {
					decorated(Label, ui.Text(c.Key("title"), opts.Title), styles, opts.Style)
				}
				if opts.Description != "" {
					decorated(Description, ui.Text(c.Key("description"), opts.Description), styles, opts.Style)
				}
			})
			if opts.Actions != nil {
				actions := ui.Row(c.Key("actions")).MinWidth(0).Wrap()
				if stacked {
					actions.ColumnStart(1).RowStart(2).AlignItems(ui.Start).Margin(t.spacing(10), 0, 0, 0)
				} else {
					actions.ColumnStart(2).RowStart(1).AlignItems(ui.Center)
					startMargin(actions, styles[Root].Gap.Value)
				}
				decorated(Actions, actions, styles, opts.Style)
				actions.Children(opts.Actions)
			}
			if closable {
				close := Button(c.Key("close"), t, "", alertCloseOptions(c, opts, styles[Root].Foreground.Value)).OnClick(opts.OnClose)
				close.ColumnStart(len(columns)).RowStart(1).JustifySelf(ui.Start)
				gap := styles[Root].Gap.Value
				if !stacked && opts.Actions != nil {
					gap = styles[Actions].Gap.Value
				} else if stacked && opts.Actions != nil {
					close.RowSpan(2)
				}
				startMargin(close, gap)
				decorated(Trailing, close, styles, opts.Style)
			}
		})
	})
	if opts.Style != nil {
		opts.Style(Root, e)
	}
	return e
}

func alertCloseOptions(c *ui.Context, opts AlertOptions, foreground ui.Color) ButtonOptions {
	close := ButtonOptions{}
	if opts.Close != nil {
		close = *opts.Close
	}
	if close.Size == SizeDefault {
		close.Size = MD
	}
	if close.Color == ColorDefault {
		close.Color = Neutral
	}
	if close.Variant == VariantDefault {
		close.Variant = LinkVariant
	}
	if close.AccessibleLabel == "" {
		close.AccessibleLabel = namedMessage(c, "dismissNamed", opts.Title)
	}
	if close.Leading == nil {
		close.Leading = opts.CloseIcon
		if close.Leading == nil {
			close.Leading = lucide.X()
		}
	}
	base := PartStyle{PaddingX: Set(float32(0)), PaddingY: Set(float32(0))}
	if close.Color == Neutral && close.Variant == LinkVariant && !close.Active && !close.UI[Root].Foreground.Set {
		base.Foreground = Set(foreground)
		custom := close.Style
		close.Style = func(root, label, leading, trailing ui.Element) {
			if !root.IsDisabled() && (root.Hovered() || root.Pressed()) {
				root.TextColor(foreground.Alpha(.75))
			}
			if custom != nil {
				custom(root, label, leading, trailing)
			}
		}
	}
	close.UI = cloneStyles(close.UI)
	close.UI[Root] = mergeStyle(base, close.UI[Root])
	return close
}

type EmptyOptions struct {
	Title, Description string
	Icon               *ui.SVG
	Avatar             *AvatarOptions
	Size               Size
	Variant            Variant
	Loading            bool
	LoadingIcon        *ui.SVG
	Header, Leading    func()
	Body               func()
	Actions, Footer    func()
	Decoration
}

func Empty(c *ui.Context, t Theme, opts EmptyOptions) ui.Element {
	windowWidth, _ := c.Size()
	padding := float32(16)
	if windowWidth >= 1024 {
		padding = 32
	} else if windowWidth >= 640 {
		padding = 24
	}
	config, styles := resolve(t, ComponentEmpty, ComponentTheme{Variant: Outline, Size: MD}, ComponentTheme{Variant: opts.Variant, Size: opts.Size, UI: opts.UI}, func(v ComponentTheme) Styles {
		s := surface(t, v.Variant)
		s.PaddingX = Set(t.spacing(padding))
		s.PaddingY = Set(t.spacing(padding))
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
		return Styles{Root: s, Header: {Gap: Set(t.spacing(8))}, Body: {Gap: Set(t.spacing(16))}, Actions: {Gap: Set(t.spacing(8))}, Footer: {Gap: Set(t.spacing(8))}, Label: {Foreground: Set(fg), FontSize: Set(t.text([5]float32{14, 14, 16, 16, 18}[sizeIndex(v.Size)])), Weight: Set(500)}, Description: {Foreground: Set(muted), FontSize: Set(t.text([5]float32{12, 12, 14, 14, 16}[sizeIndex(v.Size)]))}}
	})
	e := ui.Column(c).FillWidth().MinWidth(0).AlignItems(ui.Center).Justify(ui.Center)
	styles[Root].apply(e)
	e.Children(func() {
		if opts.Header != nil || opts.Leading != nil || opts.Avatar != nil || opts.Loading || opts.Icon != nil || opts.Title != "" || opts.Description != "" {
			header := ui.Column(c.Key("header")).MinWidth(0).MaxWidth(384).AlignItems(ui.Center)
			decorated(Header, header, styles, opts.Style)
			header.Children(func() {
				if opts.Header != nil {
					// A custom header owns its leading content, title and description.
					opts.Header()
					return
				}
				if opts.Leading != nil {
					leading := ui.Column(c.Key("leading")).AlignItems(ui.Center)
					decorated(Leading, leading, styles, opts.Style)
					leading.Children(opts.Leading)
				} else if opts.Avatar != nil || opts.Loading || opts.Icon != nil {
					ui.Box(c.Key("avatar")).Margin(0, 0, t.spacing(8), 0).Children(func() {
						emptyAvatar(c, t, opts, config.Size, styles[Leading])
					})
				}
				if opts.Title != "" {
					decorated(Label, ui.Text(c.Key("title"), opts.Title).TextAlign(ui.Center).Role(ui.RoleHeading).Level(2), styles, opts.Style)
				}
				if opts.Description != "" {
					decorated(Description, ui.Text(c.Key("description"), opts.Description).TextAlign(ui.Center), styles, opts.Style)
				}
			})
		}
		if opts.Body != nil || opts.Actions != nil {
			body := ui.Column(c.Key("body")).MaxWidth(384).MinWidth(0).AlignItems(ui.Center)
			decorated(Body, body, styles, opts.Style)
			body.Children(func() {
				if opts.Body != nil {
					opts.Body()
					return
				}
				actions := ui.Row(c.Key("actions")).Wrap().Justify(ui.Center)
				decorated(Actions, actions, styles, opts.Style)
				actions.Children(opts.Actions)
			})
		}
		if opts.Footer != nil {
			footer := ui.Column(c.Key("footer")).MaxWidth(384).MinWidth(0).AlignItems(ui.Center)
			decorated(Footer, footer, styles, opts.Style)
			footer.Children(opts.Footer)
		}
	})
	if opts.Style != nil {
		opts.Style(Root, e)
	}
	return e
}

func emptyAvatar(c *ui.Context, t Theme, opts EmptyOptions, size Size, leading PartStyle) {
	av := AvatarOptions{}
	if opts.Avatar != nil {
		av = *opts.Avatar
	}
	if av.Icon == nil {
		av.Icon = opts.Icon
		if opts.Loading {
			av.Icon = opts.LoadingIcon
			if av.Icon == nil {
				av.Icon = lucide.LoaderCircle()
			}
		}
	}
	av.UI = cloneStyles(av.UI)
	av.UI[Leading] = mergeStyle(av.UI[Leading], leading)
	custom := av.Style
	av.Style = func(slot Slot, e ui.Element) {
		if custom != nil {
			custom(slot, e)
		}
		if slot == Leading && opts.Style != nil {
			opts.Style(Leading, e)
		}
	}
	diameter := Optional[float32]{}
	if av.Size == SizeDefault {
		diameter = Set(t.text([5]float32{32, 36, 40, 44, 48}[sizeIndex(size)]))
	}
	var buildIcon func(*ui.Context, *ui.SVG, float32, ui.Color) ui.Element
	if opts.Loading {
		buildIcon = func(c *ui.Context, icon *ui.SVG, size float32, color ui.Color) ui.Element {
			return loadingSVG(c, icon, size, color, t.motion(c), t.Motion.Spin)
		}
	}
	avatarWithIcon(c, t, av, diameter, buildIcon)
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
	e := ui.Box(c).Size(w, h).Clip().Role(ui.RoleStatus).Label(message(c, "loading"))
	animated := t.motion(c)
	if opts.Animate.Set {
		animated = animated && opts.Animate.Value
	}
	style := mergeStyle(PartStyle{Radius: Set(t.Radius), Background: Set(t.Elevated)}, t.components[ComponentSkeleton].UI[Root])
	style = mergeStyle(style, opts.UI[Root])
	color := style.Background.Value
	style.Background = Set(ui.Transparent)
	style.apply(e)
	e.Children(func() {
		// Child drawing uses the root's final clip; a root Draw runs before it.
		ui.Box(c.Key(skeletonPaintKey{})).Absolute().Left(0).Top(0).Right(0).Bottom(0).IgnoreInput(true).Role(ui.RoleNone).Draw(func(p *ui.Painter, r ui.Rect) {
			alpha := float32(1)
			if animated {
				p.AnimationFrame()
				alpha = .5 + .5*float32((1+math.Cos(2*math.Pi*phase(p.Now(), t.Motion.Pulse)))/2)
			}
			p.Fill(r, color.Alpha(alpha), 0)
		})
	})
	if opts.Style != nil {
		opts.Style(Root, e)
	}
	return e
}

type skeletonPaintKey struct{}

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
	if config.Color == Neutral {
		color = t.InvertedBackground
	}
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
	if !opts.Indeterminate {
		if animated {
			ratio = e.AnimateWith("nui-progress", ratio, 200*time.Millisecond, ui.EaseOut)
		} else {
			e.Jump("nui-progress", ratio)
		}
	}
	e.Draw(func(p *ui.Painter, r ui.Rect) {
		fraction, start := ratio, float32(0)
		if opts.Indeterminate {
			fraction = .5
			start = .25
			if animated {
				p.AnimationFrame()
				start = float32(phase(p.Now(), 2*time.Second))*1.5 - .5
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
	Icon      *ui.SVG
	UI        Styles
	Build     func(*ui.Context, Theme) // Optional synchronous legend content builder.
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
		hasLegend = hasLegend || item.Label != "" || item.Icon != nil || item.Build != nil
	}
	denom := math.Max(max/scale, sum)
	value := max
	if sum < max/scale {
		value = sum * scale
	}
	thickness := t.spacing([5]float32{2, 4, 8, 12, 16}[sizeIndex(config.Size)])
	fontSize := t.text([5]float32{12, 14, 14, 14, 16}[sizeIndex(config.Size)])
	iconSize := t.text([5]float32{12, 16, 16, 16, 20}[sizeIndex(config.Size)])
	dotSize := t.text([5]float32{6, 8, 8, 8, 10}[sizeIndex(config.Size)])
	fillColor := func(color Color) ui.Color {
		if color == Neutral {
			return t.InvertedBackground
		}
		return t.color(color)
	}
	stylePart := func(slot Slot, part ui.Element, base PartStyle, itemUI Styles) {
		style := mergeStyle(base, t.components[ComponentProgressGroup].UI[slot])
		style = mergeStyle(style, opts.UI[slot])
		style = mergeStyle(style, itemUI[slot])
		style.apply(part)
		if opts.Style != nil {
			opts.Style(slot, part)
		}
	}
	e := ui.Column(c).Gap(t.spacing(8)).MinWidth(0)
	if vert {
		e.Row().AlignItems(ui.Start)
	}
	e.Children(func() {
		bar := ui.Row(c.Key("track")).FillWidth().Height(thickness).Radius(999).Clip().Role(ui.RoleProgress).Label(opts.Label).Range(0, max, value).Background(t.Accented)
		if vert {
			height := opts.Height
			if height <= 0 {
				height = 160
			}
			bar.Column().Width(thickness).Height(height)
		}
		bar.Children(func() {
			for _, item := range items {
				share := float32((item.Value/scale)/denom) * 100
				part := ui.Box(c.Key(item.ID)).Shrink(0)
				if vert {
					part.FillWidth().HeightPercent(share)
				} else {
					part.WidthPercent(share).FillHeight()
				}
				stylePart(Segment, part, PartStyle{Background: Set(fillColor(item.Color))}, item.UI)
			}
		})
		finish(t, ComponentProgressGroup, Indicator, bar, PartStyle{}, opts.Decoration)
		if !opts.HideLegend && hasLegend {
			legend := ui.Column(c.Key("legend")).Gap(t.spacing(4)).Grow(1).MinWidth(0).FontSize(fontSize)
			stylePart(List, legend, PartStyle{}, nil)
			legend.Children(func() {
				for _, item := range items {
					row := ui.Row(c.Key(item.ID)).Gap(t.spacing(6)).AlignItems(ui.Center)
					stylePart(Item, row, PartStyle{}, item.UI)
					row.Children(func() {
						if item.Build != nil {
							item.Build(c, t)
							return
						}
						if item.Icon != nil {
							part := ui.Icon(c.Key("leading"), item.Icon).Size(iconSize, iconSize).Shrink(0).Role(ui.RoleNone)
							stylePart(Leading, part, PartStyle{Foreground: Set(t.color(item.Color))}, item.UI)
						} else {
							part := ui.Box(c.Key("leading")).Size(dotSize, dotSize).Shrink(0).Radius(999)
							stylePart(Leading, part, PartStyle{Background: Set(fillColor(item.Color))}, item.UI)
						}
						stylePart(Label, ui.Text(c.Key("label"), item.Label).SingleLine().Grow(1).MinWidth(0), PartStyle{}, item.UI)
						stylePart(Trailing, ui.Text(c.Key("trailing"), fmt.Sprintf("%.0f%%", math.Round(item.Value/max*100))).Shrink(0), PartStyle{Foreground: Set(t.Dimmed)}, item.UI)
					})
				}
			})
		}
	})
	finish(t, ComponentProgressGroup, Root, e, PartStyle{}, opts.Decoration)
	return e
}
