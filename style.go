package nui

import (
	"maps"

	"github.com/egoist/mygo/ui"
)

// Optional distinguishes inheritance from an explicit zero, false or transparent value.
type Optional[T any] struct {
	Value T
	Set   bool
}

func Set[T any](v T) Optional[T] { return Optional[T]{Value: v, Set: true} }

func valueOr[T any](v Optional[T], fallback T) T {
	if v.Set {
		return v.Value
	}
	return fallback
}

func vertical(o Orientation) bool {
	if o > Vertical {
		panic("nui: invalid orientation")
	}
	return o == Vertical
}

// Orientation describes the arrangement of a component's controls.
type Orientation uint8

const (
	Horizontal Orientation = iota
	Vertical
)

type Size uint8

const (
	SizeDefault Size = iota
	XS
	SM
	MD
	LG
	XL
)

type Color uint8

const (
	ColorDefault Color = iota
	Primary
	Secondary
	Success
	Info
	Warning
	Error
	Neutral
)

type Variant uint8

const (
	VariantDefault Variant = iota
	Solid
	Outline
	Soft
	Subtle
	Ghost
	LinkVariant
	None
)

type ButtonVariant = Variant

type Component string

const (
	ComponentButton         Component = "button"
	ComponentLink           Component = "link"
	ComponentBadge          Component = "badge"
	ComponentCard           Component = "card"
	ComponentAlert          Component = "alert"
	ComponentAvatar         Component = "avatar"
	ComponentAvatarGroup    Component = "avatar-group"
	ComponentChip           Component = "chip"
	ComponentSeparator      Component = "separator"
	ComponentKbd            Component = "kbd"
	ComponentSkeleton       Component = "skeleton"
	ComponentProgress       Component = "progress"
	ComponentProgressGroup  Component = "progress-group"
	ComponentEmpty          Component = "empty"
	ComponentUser           Component = "user"
	ComponentContainer      Component = "container"
	ComponentMain           Component = "main"
	ComponentIcon           Component = "icon"
	ComponentInput          Component = "input"
	ComponentTextarea       Component = "textarea"
	ComponentSelect         Component = "select"
	ComponentInputMenu      Component = "input-menu"
	ComponentPopover        Component = "popover"
	ComponentFormField      Component = "form-field"
	ComponentFieldGroup     Component = "field-group"
	ComponentCheckbox       Component = "checkbox"
	ComponentCheckboxGroup  Component = "checkbox-group"
	ComponentRadioGroup     Component = "radio-group"
	ComponentSwitch         Component = "switch"
	ComponentSlider         Component = "slider"
	ComponentScrollArea     Component = "scroll-area"
	ComponentTooltip        Component = "tooltip"
	ComponentModal          Component = "modal"
	ComponentToast          Component = "toast"
	ComponentDropdownMenu   Component = "dropdown-menu"
	ComponentContextMenu    Component = "context-menu"
	ComponentSlideover      Component = "slideover"
	ComponentTabs           Component = "tabs"
	ComponentCollapsible    Component = "collapsible"
	ComponentAccordion      Component = "accordion"
	ComponentBreadcrumb     Component = "breadcrumb"
	ComponentPagination     Component = "pagination"
	ComponentStepper        Component = "stepper"
	ComponentTimeline       Component = "timeline"
	ComponentBanner         Component = "banner"
	ComponentError          Component = "error"
	ComponentForm           Component = "form"
	ComponentInputNumber    Component = "input-number"
	ComponentInputTags      Component = "input-tags"
	ComponentPinInput       Component = "pin-input"
	ComponentInputRating    Component = "input-rating"
	ComponentListbox        Component = "listbox"
	ComponentSelectMenu     Component = "select-menu"
	ComponentCalendar       Component = "calendar"
	ComponentInputDate      Component = "input-date"
	ComponentInputTime      Component = "input-time"
	ComponentColorPicker    Component = "color-picker"
	ComponentTable          Component = "table"
	ComponentTree           Component = "tree"
	ComponentCommandPalette Component = "command-palette"
	ComponentNavigationMenu Component = "navigation-menu"
	ComponentDrawer         Component = "drawer"
	ComponentFileUpload     Component = "file-upload"
	ComponentChart          Component = "chart"
	ComponentCarousel       Component = "carousel"
	ComponentMarquee        Component = "marquee"
	ComponentSplitter       Component = "splitter"
	ComponentSidebar        Component = "sidebar"
	ComponentHeader         Component = "header"
	ComponentFooter         Component = "footer"
	ComponentFooterColumns  Component = "footer-columns"
)

type Slot string

const (
	Root          Slot = "root"
	Label         Slot = "label"
	Leading       Slot = "leading"
	Trailing      Slot = "trailing"
	Header        Slot = "header"
	Body          Slot = "body"
	Footer        Slot = "footer"
	Description   Slot = "description"
	Hint          Slot = "hint"
	Help          Slot = "help"
	ErrorSlot     Slot = "error"
	Indicator     Slot = "indicator"
	Editor        Slot = "editor"
	Panel         Slot = "panel"
	Item          Slot = "item"
	EmptySlot     Slot = "empty"
	Thumb         Slot = "thumb"
	Backdrop      Slot = "backdrop"
	Trigger       Slot = "trigger"
	List          Slot = "list"
	Close         Slot = "close"
	Actions       Slot = "actions"
	SeparatorSlot Slot = "separator"
)

// PartStyle contains appearance and spacing only. It cannot remove roles or input behavior.
type PartStyle struct {
	Background, Foreground, BorderColor           Optional[ui.Color]
	Radius, BorderWidth, Opacity                  Optional[float32]
	FontSize, LineHeight, PaddingX, PaddingY, Gap Optional[float32]
	Font                                          Optional[string]
	Weight                                        Optional[int]
	Shadow                                        Optional[ShadowTokens]
}
type Styles map[Slot]PartStyle
type ComponentTheme struct {
	Color   Color
	Size    Size
	Variant Variant
	UI      Styles
}

// Decoration is shared by components with several synchronous styling slots.
type Decoration struct {
	UI    Styles
	Style func(Slot, ui.Element)
}

func mergeValue[T any](a, b Optional[T]) Optional[T] {
	if b.Set {
		return b
	}
	return a
}
func mergeStyle(a, b PartStyle) PartStyle {
	a.Background = mergeValue(a.Background, b.Background)
	a.Foreground = mergeValue(a.Foreground, b.Foreground)
	a.BorderColor = mergeValue(a.BorderColor, b.BorderColor)
	a.Radius = mergeValue(a.Radius, b.Radius)
	a.BorderWidth = mergeValue(a.BorderWidth, b.BorderWidth)
	a.Opacity = mergeValue(a.Opacity, b.Opacity)
	a.FontSize = mergeValue(a.FontSize, b.FontSize)
	a.LineHeight = mergeValue(a.LineHeight, b.LineHeight)
	a.PaddingX = mergeValue(a.PaddingX, b.PaddingX)
	a.PaddingY = mergeValue(a.PaddingY, b.PaddingY)
	a.Gap = mergeValue(a.Gap, b.Gap)
	a.Font = mergeValue(a.Font, b.Font)
	a.Weight = mergeValue(a.Weight, b.Weight)
	a.Shadow = mergeValue(a.Shadow, b.Shadow)
	return a
}
func (s PartStyle) apply(e ui.Element) {
	if !e.Valid() {
		return
	}
	if s.Background.Set {
		e.Background(s.Background.Value)
	}
	if s.Foreground.Set {
		e.TextColor(s.Foreground.Value)
	}
	if s.Radius.Set {
		e.Radius(s.Radius.Value)
	}
	if s.Opacity.Set {
		e.Opacity(s.Opacity.Value)
	}
	if s.Font.Set {
		e.Font(s.Font.Value)
	}
	if s.FontSize.Set {
		e.FontSize(s.FontSize.Value)
	}
	if s.LineHeight.Set {
		e.LineHeight(s.LineHeight.Value)
	}
	if s.Weight.Set {
		e.FontWeight(s.Weight.Value)
	}
	if s.PaddingX.Set {
		e.PaddingX(s.PaddingX.Value)
	}
	if s.PaddingY.Set {
		e.PaddingY(s.PaddingY.Value)
	}
	if s.Gap.Set {
		e.Gap(s.Gap.Value)
	}
	if s.Shadow.Set {
		v := s.Shadow.Value
		e.Shadow(0, v.OffsetY, v.Blur, 0, v.Color)
	}
	if s.BorderWidth.Set && s.BorderWidth.Value > 0 {
		color := ui.Transparent
		if s.BorderColor.Set {
			color = s.BorderColor.Value
		}
		e.InsetShadow(0, 0, 0, s.BorderWidth.Value, color)
	}
}
func cloneStyles(src Styles) Styles {
	out := make(Styles, len(src))
	maps.Copy(out, src)
	return out
}

// WithComponent returns an isolated theme snapshot. Caller maps are copied.
func (t Theme) WithComponent(component Component, patch ComponentTheme) Theme {
	out := make(map[Component]ComponentTheme, len(t.components)+1)
	maps.Copy(out, t.components)
	base := out[component]
	if patch.Color != ColorDefault {
		base.Color = patch.Color
	}
	if patch.Size != SizeDefault {
		base.Size = patch.Size
	}
	if patch.Variant != VariantDefault {
		base.Variant = patch.Variant
	}
	base.UI = cloneStyles(base.UI)
	for slot, style := range patch.UI {
		base.UI[slot] = mergeStyle(base.UI[slot], style)
	}
	out[component] = base
	t.components = out
	return t
}
func (t Theme) Component(component Component) ComponentTheme {
	v := t.components[component]
	v.UI = cloneStyles(v.UI)
	return v
}
func sizeIndex(size Size) int {
	if size < XS || size > XL {
		return 2
	}
	return int(size - XS)
}
func defaults(t Theme, id Component, fallback ComponentTheme) ComponentTheme {
	global := t.components[id]
	if global.Color != ColorDefault {
		fallback.Color = global.Color
	}
	if global.Size != SizeDefault {
		fallback.Size = global.Size
	}
	if global.Variant != VariantDefault {
		fallback.Variant = global.Variant
	}
	return fallback
}
func selected(base, instance ComponentTheme) ComponentTheme {
	if instance.Color != ColorDefault {
		base.Color = instance.Color
	}
	if instance.Size != SizeDefault {
		base.Size = instance.Size
	}
	if instance.Variant != VariantDefault {
		base.Variant = instance.Variant
	}
	return base
}

// resolve applies global parts before explicit instance semantics and final instance parts.
func resolve(t Theme, id Component, fallback, instance ComponentTheme, recipe func(ComponentTheme) Styles) (ComponentTheme, Styles) {
	base := defaults(t, id, fallback)
	actual := selected(base, instance)
	styles := recipe(base)
	for slot, style := range t.components[id].UI {
		styles[slot] = mergeStyle(styles[slot], style)
	}
	if instance.Color != ColorDefault || instance.Variant != VariantDefault || instance.Size != SizeDefault {
		for slot, style := range recipe(actual) {
			var overlay PartStyle
			if instance.Color != ColorDefault || instance.Variant != VariantDefault {
				overlay.Background = style.Background
				overlay.Foreground = style.Foreground
				overlay.BorderColor = style.BorderColor
				overlay.BorderWidth = style.BorderWidth
			}
			if instance.Size != SizeDefault {
				overlay.FontSize = style.FontSize
				overlay.LineHeight = style.LineHeight
				overlay.Weight = style.Weight
				overlay.Radius = style.Radius
				overlay.PaddingX = style.PaddingX
				overlay.PaddingY = style.PaddingY
				overlay.Gap = style.Gap
			}
			styles[slot] = mergeStyle(styles[slot], overlay)
		}
	}
	for slot, style := range instance.UI {
		styles[slot] = mergeStyle(styles[slot], style)
	}
	return actual, styles
}
func finish(t Theme, id Component, slot Slot, e ui.Element, base PartStyle, d Decoration) {
	base = mergeStyle(base, t.components[id].UI[slot])
	base = mergeStyle(base, d.UI[slot])
	base.apply(e)
	if d.Style != nil {
		d.Style(slot, e)
	}
}
func decorated(slot Slot, e ui.Element, styles Styles, callback func(Slot, ui.Element)) {
	styles[slot].apply(e)
	if callback != nil {
		callback(slot, e)
	}
}
func (t Theme) color(c Color) ui.Color {
	switch c {
	case Secondary:
		return t.Secondary
	case Success:
		return t.Success
	case Info:
		return t.Info
	case Warning:
		return t.Warning
	case Error:
		return t.Error
	case Neutral:
		return t.Highlighted
	default:
		return t.Primary
	}
}

func appearance(t Theme, color Color, variant Variant, hover, focus bool) PartStyle {
	c := t.color(color)
	s := PartStyle{Background: Set(ui.Transparent), Foreground: Set(c), BorderWidth: Set(float32(0)), BorderColor: Set(ui.Transparent)}
	switch variant {
	case Solid:
		s.Background = Set(c)
		s.Foreground = Set(t.Inverted)
		if hover {
			if color == Neutral {
				s.Background = Set(c.Alpha(.90))
			} else {
				s.Background = Set(c.Alpha(.75))
			}
		}
	case Outline:
		s.BorderWidth = Set(float32(1))
		s.BorderColor = Set(c.Alpha(.5))
		if focus {
			s.BorderColor = Set(c)
		}
		if color == Neutral {
			s.Background = Set(t.Background)
			s.Foreground = Set(t.Text)
			s.BorderColor = Set(t.BorderAccented)
			if focus {
				s.BorderColor = Set(t.Highlighted)
			}
		}
		if hover {
			if color == Neutral {
				s.Background = Set(t.Elevated)
			} else {
				s.Background = Set(c.Alpha(.10))
			}
		}
	case Soft, Subtle:
		s.Background = Set(c.Alpha(.10))
		if hover {
			s.Background = Set(c.Alpha(.15))
		}
		if color == Neutral {
			s.Background = Set(t.Elevated)
			s.Foreground = Set(t.Text)
			if hover {
				s.Background = Set(t.Accented.Alpha(.75))
			}
		}
		if variant == Subtle {
			s.BorderWidth = Set(float32(1))
			s.BorderColor = Set(c.Alpha(.25))
			if color == Neutral {
				s.BorderColor = Set(t.BorderAccented)
			}
			if focus {
				s.BorderColor = Set(c)
			}
		}
	case Ghost:
		if color == Neutral {
			s.Foreground = Set(t.Text)
		}
		if hover {
			if color == Neutral {
				s.Background = Set(t.Elevated)
			} else {
				s.Background = Set(c.Alpha(.10))
			}
		}
	case LinkVariant:
		if color == Neutral {
			s.Foreground = Set(t.Muted)
		}
		if hover {
			s.Foreground = Set(c.Alpha(.75))
			if color == Neutral {
				s.Foreground = Set(t.Text)
			}
		}
	}
	if t.HighContrast {
		if variant == Solid {
			s.Background = Set(c)
		}
		if s.BorderWidth.Value > 0 {
			s.BorderColor = Set(c)
		}
	}
	return s
}
