package nui

import (
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

// motionVisible uses committed window geometry. Applications must also set Active=false
// for panels kept mounted but hidden by their own clipping/visibility mechanism.
func motionVisible(c *ui.Context, e ui.Element, active Optional[bool]) bool {
	if active.Set && !active.Value {
		return false
	}
	r := e.Bounds()
	w, h := c.Size()
	return r.W > 0 && r.H > 0 && r.X < w && r.Y < h && r.X+r.W > 0 && r.Y+r.H > 0
}

type DrawerState struct {
	Open              bool
	Snap              int
	height, start     float32
	dragging, wasOpen bool
}
type DrawerOptions struct {
	ModalOptions
	Handle Optional[bool]
	Inset  bool
	// SnapPoints are ascending fractions of available window height, in (0,1].
	SnapPoints                       []float32
	DismissDistance, DismissVelocity float32
	OnSnap                           func(int)
	pinnedRegions                    bool // Sidebar keeps its header/footer outside the body scroller.
}

func drawerPoints(points []float32) []float32 {
	if len(points) == 0 {
		return []float32{.5, .85}
	}
	points = slices.Clone(points)
	for i, p := range points {
		if math.IsNaN(float64(p)) || p <= 0 || p > 1 || (i > 0 && p <= points[i-1]) {
			panic("nui: Drawer snap points must increase within (0,1]")
		}
	}
	return points
}
func drawerSnap(height, velocity, displacement, window float32, points []float32, distance, speed float32, dismiss bool) (int, bool) {
	if distance <= 0 {
		distance = 80
	}
	if speed <= 0 {
		speed = 900
	}
	if dismiss && (height < window*points[0]*.5 || (displacement >= distance && velocity >= speed)) {
		return 0, true
	}
	projected := height - velocity*.15
	best := 0
	delta := float32(math.Inf(1))
	for i, p := range points {
		d := float32(math.Abs(float64(projected - window*p)))
		if d < delta {
			best = i
			delta = d
		}
	}
	return best, false
}

// Drawer is a bottom sheet. Only its handle drags; scrolling body content cannot dismiss it.
// Call every frame, including while closed, as for Slideover.
func Drawer(c *ui.Context, t Theme, s *DrawerState, o DrawerOptions, content func(Theme)) ui.Element {
	if s == nil {
		panic("nui: Drawer requires state")
	}
	for _, value := range []float32{o.DismissDistance, o.DismissVelocity} {
		if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) || value < 0 {
			panic("nui: Drawer dismissal thresholds must be finite and non-negative")
		}
	}
	points := drawerPoints(o.SnapPoints)
	s.Snap = max(0, min(len(points)-1, s.Snap))
	_, h := c.Size()
	if o.Inset {
		h = max(0, h-2*t.spacing(16))
	}
	if !s.Open || !valueOr(o.Handle, true) {
		s.dragging = false
	}
	if s.Open && (!s.wasOpen || !s.dragging) {
		s.height = h * points[s.Snap]
	}
	s.wasOpen = s.Open
	mo := o.ModalOptions
	if !mo.CloseButton.Set {
		mo.CloseButton = Set(mo.Close != nil)
	}
	name := o.Title
	if name == "" {
		name = message(c, "drawer")
	}
	if mo.CloseLabel == "" && o.Title == "" && (mo.Close == nil || mo.Close.AccessibleLabel == "") {
		mo.CloseLabel = namedMessage(c, "closeNamed", name)
	}
	mo.beforeHeader = func(c *ui.Context, t Theme, panel ui.Element) {
		if o.Title == "" {
			panel.Label(name)
		}
		if !valueOr(o.Handle, true) {
			return
		}
		handle := ui.Box(c.Key("drawer-handle")).Size(t.spacing(64), t.spacing(16)).Margin(t.spacing(11), 0, 0, 0).AlignSelf(ui.Center).Focusable().FocusRing(true).Role(ui.RoleSlider).Label(namedMessage(c, "resizeNamed", name)).Range(0, float64(h*points[len(points)-1]), float64(s.height)).Cursor(ui.CursorResizeNS)
		handle.Children(func() {
			finish(t, ComponentDrawer, Indicator, ui.Box(c).Width(t.spacing(48)).Height(t.spacing(6)).Margin(t.spacing(5), t.spacing(8)), PartStyle{Radius: Set(t.radius(999)), Background: Set(t.Accented)}, o.Decoration)
		})
		dx, dy, held := handle.Dragged()
		_ = dx
		if held {
			if !s.dragging {
				s.start = s.height
				s.dragging = true
			}
			s.height = max(0, min(h*points[len(points)-1], s.height-dy))
		}
		_, vy, ended := handle.DragEnded()
		if ended && s.dragging {
			at, close := drawerSnap(s.height, vy, s.start-s.height, h, points, o.DismissDistance, o.DismissVelocity, true)
			prevented := close && !valueOr(o.Dismissible, true)
			if prevented {
				at, _ = drawerSnap(s.height, vy, s.start-s.height, h, points, o.DismissDistance, o.DismissVelocity, false)
				close = false
			}
			changed := at != s.Snap
			s.Snap, s.Open = at, !close
			s.dragging = false
			if prevented && o.OnClosePrevent != nil || close && o.OnClose != nil || !close && changed && o.OnSnap != nil {
				handle.SetChanged().OnChange(func() {
					if prevented && o.OnClosePrevent != nil {
						o.OnClosePrevent()
					}
					if close {
						if o.OnClose != nil {
							o.OnClose()
						}
					} else if changed && o.OnSnap != nil {
						o.OnSnap(at)
					}
				})
			}
			c.After(0)
		} else if s.dragging && !held {
			s.dragging = false
			s.height = h * points[s.Snap]
			c.After(0)
		}
		set := func(i int) {
			next := max(0, min(len(points)-1, i))
			changed := next != s.Snap
			s.Snap = next
			s.dragging = false
			if changed && o.OnSnap != nil {
				o.OnSnap(next)
			}
		}
		handle.OnShortcut(0, ui.KeyUp, func() { set(s.Snap + 1) }).OnShortcut(0, ui.KeyDown, func() { set(s.Snap - 1) }).OnShortcut(0, ui.KeyHome, func() { set(0) }).OnShortcut(0, ui.KeyEnd, func() { set(len(points) - 1) })
		finish(t, ComponentDrawer, Thumb, handle, PartStyle{}, o.Decoration)
	}
	mo.plain = true
	mo.drawerLayout = !o.pinnedRegions
	style := mo.Style
	mo.Style = func(slot Slot, e ui.Element) {
		finish(t, ComponentDrawer, slot, e, PartStyle{}, Decoration{UI: o.UI})
		if style != nil {
			style(slot, e)
		}
	}
	return Slideover(c, t, &s.Open, SlideoverOptions{ModalOptions: mo, Side: SideBottom, Extent: Set(s.height), Inset: o.Inset}, content)
}

type CarouselItem struct {
	ID, Label string
	Build     func(*ui.Context, Theme)
	Decoration
}
type CarouselState struct {
	Index                int
	Paused               bool
	handle               ui.Handle
	root                 ui.Handle
	offset               float32
	held                 bool
	vertical, ignoreDrag bool
	rtl                  bool
	last                 time.Time
	identity             string
}
type CarouselOptions struct {
	Label    string
	Height   float32
	Active   Optional[bool]
	Autoplay time.Duration
	// Wrap wraps explicit/automatic next and previous actions; drag stays bounded.
	Wrap                         bool
	Orientation                  Orientation
	Arrows                       Optional[bool]
	Dots                         bool
	Prev, Next                   *ButtonOptions
	PrevIcon, NextIcon           *ui.SVG
	WatchDrag, StopOnInteraction Optional[bool]
	OnChange                     func(int)
	Decoration
}

func Carousel(c *ui.Context, t Theme, s *CarouselState, items []CarouselItem, o CarouselOptions) ui.Element {
	vert := vertical(o.Orientation)
	if s == nil {
		panic("nui: Carousel requires state")
	}
	seen := map[string]bool{}
	for _, it := range items {
		checkID(seen, it.ID, "Carousel")
	}
	s.Index = max(0, min(max(0, len(items)-1), s.Index))
	height := o.Height
	if math.IsNaN(float64(height)) || math.IsInf(float64(height), 0) {
		panic("nui: carousel height must be finite")
	}
	if height <= 0 {
		height = 180
	}
	root := ui.Column(c).MinWidth(0).Gap(t.spacing(8)).Label(o.Label).Role(ui.RoleGroup).Bind(&s.root)
	finish(t, ComponentCarousel, Root, root, PartStyle{}, o.Decoration)
	now := c.Now()
	setIndex := func(index int) bool {
		if len(items) == 0 {
			return false
		}
		if o.Wrap {
			index = (index%len(items) + len(items)) % len(items)
		}
		index = max(0, min(len(items)-1, index))
		changed := index != s.Index
		if changed {
			s.Index = index
		}
		s.offset = 0
		s.last = now
		return changed
	}
	change := func(index int) {
		if setIndex(index) && o.OnChange != nil {
			o.OnChange(s.Index)
		}
	}
	stop := func() {
		if o.Autoplay > 0 && valueOr(o.StopOnInteraction, true) {
			s.Paused = true
		}
	}
	navigate := func(index int) { stop(); change(index) }
	root.Children(func() {
		var pending Optional[int]
		// The arrows are siblings over the viewport, so pressing them never drags it.
		stage := ui.Box(c.Key("stage")).MinWidth(0)
		var view ui.Element
		stage.Children(func() {
			view = ui.Box(c).Height(height).MinWidth(0).FillWidth().Clip().Focusable().FocusRing(true).Bind(&s.handle).Label(namedMessage(c, "slidesNamed", o.Label))
		})
		finish(t, ComponentCarousel, Body, view, PartStyle{Radius: Set(t.radius(8)), Background: Set(t.Elevated)}, o.Decoration)
		back, next := ui.KeyLeft, ui.KeyRight
		if vert {
			back, next = ui.KeyUp, ui.KeyDown
		} else if c.Direction() == ui.RightToLeft {
			back, next = next, back
		}
		view.OnShortcut(0, back, func() { navigate(s.Index - 1) }).OnShortcut(0, next, func() { navigate(s.Index + 1) }).OnShortcut(0, ui.KeyHome, func() { navigate(0) }).OnShortcut(0, ui.KeyEnd, func() { navigate(len(items) - 1) })
		if len(items) == 0 {
			view.Children(func() {
				finish(t, ComponentCarousel, EmptySlot, ui.Text(c, message(c, "noSlides")), PartStyle{PaddingX: Set(t.spacing(16)), PaddingY: Set(t.spacing(16))}, o.Decoration)
			})
			s.last = time.Time{}
			s.identity, s.offset, s.held, s.ignoreDrag = "", 0, false, false
			return
		}
		running := o.Autoplay > 0 && len(items) > 1 && !root.IsDisabled() && !s.Paused && !s.held && !t.Motion.Reduce && !c.Preferences().ReduceMotion && motionVisible(c, view, o.Active) && !root.Hovered() && !s.root.FocusWithin(c)
		if running {
			if s.last.IsZero() {
				s.last = c.Now()
			}
			if c.Now().Sub(s.last) >= o.Autoplay {
				if o.Wrap || s.Index < len(items)-1 {
					if setIndex(s.Index + 1) {
						pending = Set(s.Index)
					}
				} else {
					s.last = time.Time{}
					running = false
				}
			}
			if running {
				c.After(max(time.Millisecond, o.Autoplay-c.Now().Sub(s.last)))
			}
		} else {
			s.last = time.Time{}
		}
		if s.identity != items[s.Index].ID {
			if s.handle.FocusWithin(c) && !s.handle.Focused(c) {
				s.handle.Focus(c)
			}
			s.identity = items[s.Index].ID
			s.offset = 0
			s.held = false
		}
		w := view.Bounds().W
		if vert {
			w = view.Bounds().H
		}
		if w <= 0 {
			w = 320
		}
		var dx, dy, vx, vy float32
		var held, ended bool
		if valueOr(o.WatchDrag, true) && !view.IsDisabled() {
			dx, dy, held = view.Dragged()
			vx, vy, ended = view.DragEnded()
		} else {
			s.offset, s.held = 0, false
		}
		rtl := c.Direction() == ui.RightToLeft
		if s.vertical != vert || s.rtl != rtl {
			s.vertical, s.rtl, s.offset, s.held, s.ignoreDrag = vert, rtl, 0, false, held
		}
		if s.ignoreDrag {
			s.ignoreDrag = held
			held = false
		}
		if vert {
			dx = dy
		}
		if held {
			s.held = true
			s.offset = max(-w, min(w, s.offset+dx))
			s.last = time.Time{}
		}
		if vert {
			vx = vy
		}
		if ended && s.held {
			delta := 0
			if s.offset < -w*.2 || vx < -600 {
				delta = 1
			} else if s.offset > w*.2 || vx > 600 {
				delta = -1
			}
			if !vert && c.Direction() == ui.RightToLeft {
				delta = -delta
			}
			stop()
			if setIndex(max(0, min(len(items)-1, s.Index+delta))) {
				pending = Set(s.Index)
			}
			s.held = false
		} else if !held && s.held {
			s.held = false
			s.offset = 0
		}
		view.Children(func() {
			for i := max(0, s.Index-1); i <= min(len(items)-1, s.Index+1); i++ {
				it := items[i]
				pane := ui.Column(c.Key(it.ID)).Absolute().WidthPercent(100).MinWidth(0).Label(it.Label)
				if vert {
					pane.Left(0).TopPercent(float32((i-s.Index)*100)).HeightPercent(100).Translate(0, s.offset)
				} else {
					offset := (i - s.Index) * 100
					if c.Direction() == ui.RightToLeft {
						offset = -offset
					}
					pane.LeftPercent(float32(offset)).Top(0).Bottom(0).Translate(s.offset, 0)
				}
				if i != s.Index {
					pane.Disabled(true).Role(ui.RoleNone)
				}
				style := mergeStyle(mergeStyle(mergeStyle(PartStyle{PaddingX: Set(t.spacing(20)), PaddingY: Set(t.spacing(20))}, t.components[ComponentCarousel].UI[Item]), o.UI[Item]), it.UI[Item])
				style.apply(pane)
				if o.Style != nil {
					o.Style(Item, pane)
				}
				if it.Style != nil {
					it.Style(Item, pane)
				}
				pane.Children(func() {
					if it.Build != nil {
						it.Build(c, t)
					} else {
						ui.Text(c, it.Label).FontSize(t.text(18))
					}
				})
			}
		})
		if valueOr(o.Arrows, false) {
			stage.Children(func() {
				// Nuxt's round outline arrows, 16 DIPs in from the viewport's
				// sides (or its top and bottom), centered across it.
				previous, following := lucide.ArrowLeft(), lucide.ArrowRight()
				if vert {
					previous, following = lucide.ArrowUp(), lucide.ArrowDown()
				} else if c.Direction() == ui.RightToLeft {
					previous, following = following, previous
				}
				arrow := func(key, label string, icon *ui.SVG, disabled bool, delta int, start bool) {
					custom, override, slot := o.Next, o.NextIcon, Next
					if start {
						custom, override, slot = o.Prev, o.PrevIcon, Previous
					}
					options := carouselButtonOptions(custom, override, icon, label, disabled)
					b := Button(c.Key(key), t, "", options).Absolute().OnClick(func() { navigate(s.Index + delta) })
					switch {
					case vert && start:
						b.Attach(ui.AnchorTop, ui.AnchorTop).Top(t.spacing(16))
					case vert:
						b.Attach(ui.AnchorBottom, ui.AnchorBottom).Bottom(t.spacing(16))
					case start != rtl:
						b.Attach(ui.AnchorLeft, ui.AnchorLeft).Left(t.spacing(16))
					default:
						b.Attach(ui.AnchorRight, ui.AnchorRight).Right(t.spacing(16))
					}
					finish(t, ComponentCarousel, slot, b, PartStyle{}, o.Decoration)
				}
				arrow("previous", namedMessage(c, "previousNamed", o.Label), previous, !o.Wrap && s.Index == 0, -1, true)
				arrow("next", namedMessage(c, "nextNamed", o.Label), following, !o.Wrap && s.Index == len(items)-1, 1, false)
			})
		}
		root.Value(formatMessage(c, "slidePosition", map[string]any{"index": s.Index + 1, "total": len(items)}))
		if o.Dots || o.Autoplay > 0 {
			controls := ui.Row(c.Key("controls")).Wrap().AlignItems(ui.Center).Justify(ui.Center)
			finish(t, ComponentCarousel, Actions, controls, PartStyle{Gap: Set(t.spacing(12))}, o.Decoration)
			controls.Children(func() {
				if o.Dots {
					// Nuxt's 12 DIP dots, accented, and inverted for the current slide.
					dots := ui.Row(c.Key("dots")).Wrap().AlignItems(ui.Center)
					finish(t, ComponentCarousel, List, dots, PartStyle{Gap: Set(t.spacing(12))}, o.Decoration)
					dots.Children(func() {
						for i, it := range items {
							dot := ui.ButtonBase(c.Key("dot:"+it.ID)).Label(formatMessage(c, "goToSlide", map[string]any{"number": i + 1})).Selected(s.Index == i).FocusRing(false).Size(12, 12).Radius(999).OnClick(func() { change(i) })
							color := t.Accented
							if s.Index == i {
								color = t.InvertedBackground
							}
							if dot.FocusVisible() {
								dot.Shadow(0, 0, 0, 3, t.InvertedBackground.Alpha(.25))
							}
							finish(t, ComponentCarousel, Indicator, dot, PartStyle{Background: Set(color)}, o.Decoration)
						}
					})
				}
				if o.Autoplay > 0 {
					label := namedMessage(c, "pauseNamed", o.Label)
					if s.Paused {
						label = namedMessage(c, "playNamed", o.Label)
					}
					button := Button(c.Key("play-pause"), t, label, ButtonOptions{Variant: Ghost, Color: Neutral}).OnClick(func() { s.Paused = !s.Paused; s.last = time.Time{} })
					finish(t, ComponentCarousel, Trigger, button, PartStyle{}, o.Decoration)
				}
			})
		}
		if pending.Set && o.OnChange != nil {
			index := pending.Value
			view.SetChanged().OnChange(func() { o.OnChange(index) })
		}
	})
	return root
}

type MarqueeState struct {
	Offset  float32
	Paused  bool
	last    time.Time
	content string
}
type MarqueeOptions struct {
	Label              string
	Vertical, Reverse  bool
	Orientation        Orientation
	PauseOnHover       bool
	Speed, Gap, Height float32
	Active             Optional[bool]
	// Repeat is the number of content copies. Unset uses four; explicit zero is empty.
	Repeat Optional[int]
	// Overlay fades the content in and out at both edges; defaults to true.
	Overlay Optional[bool]
	// ContentKey resets the phase when content changes. Children must be presentation-only.
	ContentKey string
	Decoration
}

func Marquee(c *ui.Context, t Theme, s *MarqueeState, o MarqueeOptions, content func()) ui.Element {
	if s == nil {
		panic("nui: Marquee requires state")
	}
	for _, v := range []float32{o.Height, o.Speed, o.Gap} {
		if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
			panic("nui: Marquee dimensions and speed must be finite")
		}
	}
	repeat := valueOr(o.Repeat, 4)
	if repeat < 0 {
		panic("nui: Marquee repeat must be non-negative")
	}
	o.Vertical = vertical(o.Orientation) || o.Vertical
	if !o.Vertical && c.Direction() == ui.RightToLeft {
		o.Reverse = !o.Reverse
	}
	h := o.Height
	if h <= 0 {
		h = 44
	}
	speed := o.Speed
	if speed <= 0 {
		speed = 32
	}
	gap := o.Gap
	if gap <= 0 {
		gap = t.spacing(64)
	}
	root := ui.Box(c).MinWidth(0).Height(h).Clip().Focusable().FocusRing(true).Label(o.Label).Role(ui.RoleGroup)
	finish(t, ComponentMarquee, Root, root, PartStyle{}, o.Decoration)
	root.OnShortcut(0, ui.KeySpace, func() { s.Paused = !s.Paused })
	if s.content != o.ContentKey {
		s.content = o.ContentKey
		s.Offset = 0
		s.last = time.Time{}
	}
	if repeat == 0 {
		s.Offset, s.last = 0, time.Time{}
		return root
	}
	var measure ui.Element
	root.Children(func() {
		extent := root.Bounds().W
		if o.Vertical {
			extent = root.Bounds().H
		}
		// Each copy is at least as long as the marquee, its items spread
		// around, as Nuxt's w-full min-w-max justify-around content.
		copyOf := func(key string) ui.Element {
			e := ui.Row(c.Key(key)).Absolute().IgnoreInput(true).Justify(ui.SpaceAround).AlignItems(ui.Center)
			if o.Vertical {
				e.Column().MinHeight(extent).Attach(ui.AnchorTop, ui.AnchorTop)
			} else {
				e.MinWidth(extent).Attach(ui.AnchorLeft, ui.AnchorLeft)
			}
			finish(t, ComponentMarquee, Body, e, PartStyle{Gap: Set(gap)}, o.Decoration)
			e.Children(func() {
				if content != nil {
					content()
				}
			})
			return e
		}
		measure = copyOf("content")
		length := measure.Bounds().W
		if o.Vertical {
			length = measure.Bounds().H
		}
		period := length + gap
		if period > 0 {
			s.Offset = float32(math.Mod(finite(float64(s.Offset)), float64(period)))
			if s.Offset < 0 {
				s.Offset += period
			}
		}
		running := period > gap && !root.IsDisabled() && !s.Paused && (!o.PauseOnHover || !root.Hovered()) && !root.FocusWithin() && !t.Motion.Reduce && !c.Preferences().ReduceMotion && motionVisible(c, root, o.Active)
		if running {
			if !s.last.IsZero() {
				delta := min(c.Now().Sub(s.last), 100*time.Millisecond)
				s.Offset = float32(math.Mod(float64(s.Offset+speed*float32(delta.Seconds())), float64(period)))
			}
			s.last = c.Now()
			c.AnimationFrame()
		} else {
			s.last = time.Time{}
		}
		offset := s.Offset
		if o.Reverse {
			offset = period - offset
		}
		if t.Motion.Reduce || c.Preferences().ReduceMotion {
			offset = 0
			s.Offset = 0
		}
		if o.Vertical {
			measure.Translate(0, -offset)
		} else {
			measure.Translate(-offset, 0)
		}
		// Every copy is built even before the first committed measurement.
		for i := 1; i < repeat; i++ {
			copy := copyOf(fmt.Sprintf("repeat-%d", i)).Role(ui.RoleNone)
			at := float32(i)*period - offset
			if o.Vertical {
				copy.Translate(0, at)
			} else {
				copy.Translate(at, 0)
			}
		}
		if valueOr(o.Overlay, true) {
			// The edges fade into the background over a third of the marquee.
			fade := func(key string, start bool) {
				e := ui.Box(c.Key(key)).Absolute().IgnoreInput(true).Role(ui.RoleNone)
				angle := float32(90)
				if o.Vertical {
					e.Left(0).Right(0).HeightPercent(33.33)
					angle = 180
					if start {
						e.Top(0)
					} else {
						e.Bottom(0)
						angle = 0
					}
				} else {
					e.Top(0).Bottom(0).WidthPercent(33.33)
					toward := start != (c.Direction() == ui.RightToLeft)
					if start {
						e.Start(0)
					} else {
						e.End(0)
					}
					if !toward {
						angle = 270
					}
				}
				e.LinearGradient(ui.LinearGradient{From: t.Background, To: t.Background.Alpha(0), Angle: angle})
			}
			fade("fade-start", true)
			fade("fade-end", false)
		}
	})
	return root
}
