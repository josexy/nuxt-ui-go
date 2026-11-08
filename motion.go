package nui

import (
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

// motionVisible reports whether some of e showed in the frame before, within
// the window and the scroll containers and clips around it. Applications must
// also set Active=false for panels they hide by other means, such as covering them.
func motionVisible(e ui.Element, active Optional[bool]) bool {
	if active.Set && !active.Value {
		return false
	}
	return e.InView()
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
		running := o.Autoplay > 0 && len(items) > 1 && !root.IsDisabled() && !s.Paused && !s.held && !t.Motion.Reduce && !c.Preferences().ReduceMotion && motionVisible(view, o.Active) && !root.Hovered() && !s.root.FocusWithin(c)
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
	// Offset is the content's phase within one period of it and its gap, as
	// of the last build: between builds the content moves in frames that only
	// paint. Set it to move the content there.
	Offset  float32
	Paused  bool
	content string
	written float32 // Offset as the last build left it; another value is the app's.
}

// marqueeTravel is how far one run of a Marquee strip's pose heads, in DIPs,
// and marqueeRenew how far from 0 a pose goes before a strip starts again from
// 0: float32 keeps poses that far exact to a few thousandths of a DIP. Tests
// shorten them to see runs and strips start again.
var (
	marqueeTravel float32 = 65536
	marqueeRenew  float64 = 1 << 20
)

// marqueeRun is how a Marquee moves its strip, kept with the marquee's root:
// the strip is posed in every build, at rest at p, or moving since t0 from p0
// in direction dir at speed, while the phase went on from phase0. A build
// then translates the strip so that its phase shows exactly, and wraps it
// back within its copies as it goes.
type marqueeRun struct {
	t0                 time.Time
	phase0, p0, p      float32
	dir, speed, length float32
	moving             bool
	// A strip of another generation is a new element: its first pose shows at
	// once, so it moves only from the next frame, freshAt being its first.
	// renew asks the next build for one, at 0.
	gen          int
	built, renew bool
	freshAt      time.Time
	vertical     bool
}

type marqueeStripKey struct{ gen int }

func marqueeRunTime(speed float32) time.Duration {
	return time.Duration(float64(marqueeTravel/speed) * float64(time.Second))
}

// pose returns where the strip's pose is at now.
func (r *marqueeRun) pose(now time.Time) float32 {
	if !r.moving {
		return r.p
	}
	return r.p0 + r.dir*r.speed*float32(min(now.Sub(r.t0), marqueeRunTime(r.speed)).Seconds())
}

// phase returns how far the content went along its period at now.
func (r *marqueeRun) phase(now time.Time, period float32) float32 {
	d := r.phase0 + r.speed*float32(now.Sub(r.t0).Seconds())
	phase := float32(math.Mod(finite(float64(d)), float64(period)))
	if phase >= period {
		phase = 0 // rounded up to float32
	}
	return phase
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
	}
	run := ui.Local(root, "nui-marquee-run", func() marqueeRun { return marqueeRun{} })
	if repeat == 0 {
		s.Offset, s.written = 0, 0
		*run = marqueeRun{gen: run.gen + 1}
		return root
	}
	root.Children(func() {
		now := c.Now()
		extent := root.Bounds().W
		if o.Vertical {
			extent = root.Bounds().H
		}
		if run.renew || run.built && run.vertical != o.Vertical {
			// Start again from 0 with a strip of a new generation.
			*run = marqueeRun{gen: run.gen + 1, length: run.length}
		}
		// The strip moves every copy at once: posed, it moves in frames that
		// only paint, without building the view.
		strip := ui.Box(c.Key(marqueeStripKey{run.gen})).Absolute().Left(0).Top(0).Right(0).Bottom(0).IgnoreInput(true).Role(ui.RoleNone)
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
		// Every copy is built even before the first committed measurement.
		var measure ui.Element
		copies := make([]ui.Element, 0, repeat-1)
		strip.Children(func() {
			measure = copyOf("content")
			for i := 1; i < repeat; i++ {
				copies = append(copies, copyOf(fmt.Sprintf("repeat-%d", i)).Role(ui.RoleNone))
			}
		})
		length := measure.Bounds().W
		if o.Vertical {
			length = measure.Bounds().H
		}
		if length > 0 {
			run.length = length
		} else if !run.built {
			length = run.length // a strip of a new generation is not measured yet
		}
		period := length + gap
		// The strip wraps back within its copies, which repeat every period.
		span := period * float32(max(1, repeat-1))
		if period > 0 {
			s.Offset = float32(math.Mod(finite(float64(s.Offset)), float64(period)))
			if s.Offset < 0 {
				s.Offset += period
			}
			if s.Offset >= period {
				s.Offset = 0 // rounded up to float32
			}
		}
		external := s.Offset != s.written
		if run.moving && !external {
			s.Offset = run.phase(now, period)
		}
		motion := !t.Motion.Reduce && !c.Preferences().ReduceMotion
		running := period > gap && !root.IsDisabled() && !s.Paused && (!o.PauseOnHover || !root.Hovered()) && !root.FocusWithin() && motion && motionVisible(root, o.Active) && !c.Instant()
		dir := float32(-1)
		if o.Reverse {
			dir = 1
		}
		far := math.Abs(float64(run.pose(now))) > marqueeRenew
		if run.moving && (!running || external || far || run.dir != dir || run.speed != speed || now.Sub(run.t0) >= marqueeRunTime(run.speed)-time.Second) {
			// The run stops where it got to; one that goes on starts again
			// from there. A pose so far from 0 that float32 blurs it starts
			// again from 0, in the next build, once this one has shown the
			// phase where the run left it.
			run.p, run.moving, run.renew = run.pose(now), false, far
		}
		if !motion {
			s.Offset = 0
		}
		if running && !run.moving && run.built && !run.renew && !run.freshAt.Equal(now) {
			run.moving, run.t0, run.phase0, run.p0, run.dir, run.speed = true, now, s.Offset, run.p, dir, speed
		}
		if !run.built {
			run.built, run.freshAt, run.vertical = true, now, o.Vertical
		}
		target, a := run.p, ui.Animation{Duration: time.Millisecond, Ease: ui.Linear}
		if run.moving {
			target, a = run.p0+run.dir*marqueeTravel, ui.Animation{Duration: marqueeRunTime(run.speed), Ease: ui.Linear}
		}
		pose := ui.Pose{X: target}
		if o.Vertical {
			pose = ui.Pose{Y: target}
		}
		strip.AnimatePose(pose, a)
		// Where the phase shows: content going backwards comes from the far
		// copy, which repeats the first.
		shown := -s.Offset
		if o.Reverse && motion {
			shown = s.Offset - span
		}
		at := shown - run.pose(now)
		if o.Vertical {
			strip.Translate(0, at)
		} else {
			strip.Translate(at, 0)
		}
		for i, copy := range copies {
			if o.Vertical {
				copy.Translate(0, float32(i+1)*period)
			} else {
				copy.Translate(float32(i+1)*period, 0)
			}
		}
		switch {
		case run.moving:
			// Build again before the strip leaves its copies, and at least
			// once a second to see whether the marquee still shows.
			wrap := time.Duration(float64((span-s.Offset)/speed) * float64(time.Second))
			c.After(max(time.Millisecond, min(time.Second, wrap)))
		case running:
			c.After(time.Millisecond) // a new strip moves from the next frame
		}
		if !root.InView() {
			// Nothing has measured the marquee before its first frame: once
			// it paints, build again with its size. One out of view never
			// paints, so it never asks.
			services, ask := c.Services(), true
			strip.Draw(func(_ *ui.Painter, r ui.Rect) {
				if ask && r.W > 0 && r.H > 0 {
					ask = false
					services.Invalidate()
				}
			})
		}
		s.written = s.Offset
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
