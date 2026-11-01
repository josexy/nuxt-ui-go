package nui

import (
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/egoist/mygo/ui"
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
	// SnapPoints are ascending fractions of window height, in (0,1].
	SnapPoints                       []float32
	DismissDistance, DismissVelocity float32
	OnSnap                           func(int)
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
	points := drawerPoints(o.SnapPoints)
	s.Snap = max(0, min(len(points)-1, s.Snap))
	_, h := c.Size()
	if !s.Open {
		s.dragging = false
	}
	if s.Open && (!s.wasOpen || !s.dragging) {
		s.height = h * points[s.Snap]
	}
	s.wasOpen = s.Open
	mo := o.ModalOptions
	originalHeader := mo.Header
	mo.Header = func(c *ui.Context, t Theme) {
		handle := ui.Box(c.Key("drawer-handle")).Size(64, 16).AlignSelf(ui.Center).Focusable().FocusRing(true).Role(ui.RoleSlider).Label(namedMessage(c, "resizeNamed", o.Title)).Range(0, float64(h*points[len(points)-1]), float64(s.height)).Cursor(ui.CursorResizeNS)
		handle.Children(func() { ui.Box(c).Width(40).Height(4).Margin(6, 12).Radius(2).Background(t.BorderAccented) })
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
			s.Snap, s.Open = func() (int, bool) {
				at, close := drawerSnap(s.height, vy, s.start-s.height, h, points, o.DismissDistance, o.DismissVelocity, !o.Dismissible.Set || o.Dismissible.Value)
				return at, !close
			}()
			s.dragging = false
			if !s.Open && o.OnClose != nil {
				handle.SetChanged().OnChange(o.OnClose)
			} else if o.OnSnap != nil {
				handle.SetChanged().OnChange(func() { o.OnSnap(s.Snap) })
			}
		} else if s.dragging && !held {
			s.dragging = false
			s.height = h * points[s.Snap]
		}
		set := func(i int) {
			s.Snap = max(0, min(len(points)-1, i))
			s.dragging = false
			if o.OnSnap != nil {
				o.OnSnap(s.Snap)
			}
		}
		handle.OnShortcut(0, ui.KeyUp, func() { set(s.Snap + 1) }).OnShortcut(0, ui.KeyDown, func() { set(s.Snap - 1) }).OnShortcut(0, ui.KeyHome, func() { set(0) }).OnShortcut(0, ui.KeyEnd, func() { set(len(points) - 1) })
		finish(t, ComponentDrawer, Thumb, handle, PartStyle{}, o.Decoration)
		if originalHeader != nil {
			originalHeader(c, t)
		} else {
			ui.Text(c, o.Title).Role(ui.RoleHeading).Level(2).FontSize(t.text(18)).FontWeight(600)
			if o.Description != "" {
				ui.Text(c, o.Description).TextColor(t.Muted)
			}
		}
	}
	style := mo.Style
	mo.Style = func(slot Slot, e ui.Element) {
		finish(t, ComponentDrawer, slot, e, PartStyle{}, Decoration{UI: o.UI})
		if style != nil {
			style(slot, e)
		}
	}
	return Slideover(c, t, &s.Open, SlideoverOptions{ModalOptions: mo, Side: SideBottom, Extent: Set(s.height)}, content)
}

type CarouselItem struct {
	ID, Label string
	Build     func(*ui.Context, Theme)
}
type CarouselState struct {
	Index                int
	Paused               bool
	handle               ui.Handle
	offset               float32
	held                 bool
	vertical, ignoreDrag bool
	last                 time.Time
	identity             string
}
type CarouselOptions struct {
	Label    string
	Height   float32
	Active   Optional[bool]
	Autoplay time.Duration
	// Wrap wraps explicit/automatic next and previous actions; drag stays bounded.
	Wrap        bool
	Orientation Orientation
	Arrows      Optional[bool]
	Dots        bool
	OnChange    func(int)
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
	if height <= 0 {
		height = 180
	}
	root := ui.Column(c).MinWidth(0).Gap(8).Label(o.Label).Role(ui.RoleGroup)
	finish(t, ComponentCarousel, Root, root, PartStyle{}, o.Decoration)
	change := func(index int) {
		if len(items) == 0 {
			return
		}
		if o.Wrap {
			index = (index%len(items) + len(items)) % len(items)
		}
		index = max(0, min(len(items)-1, index))
		if index != s.Index {
			s.Index = index
			if o.OnChange != nil {
				o.OnChange(index)
			}
		}
		s.offset = 0
		s.last = c.Now()
	}
	root.Children(func() {
		view := ui.Box(c).Height(height).MinWidth(0).Clip().Focusable().FocusRing(true).Bind(&s.handle).Label(namedMessage(c, "slidesNamed", o.Label))
		finish(t, ComponentCarousel, Body, view, PartStyle{Radius: Set(t.radius(8)), Background: Set(t.Elevated)}, o.Decoration)
		back, next := ui.KeyLeft, ui.KeyRight
		if vert {
			back, next = ui.KeyUp, ui.KeyDown
		} else if c.Direction() == ui.RightToLeft {
			back, next = next, back
		}
		view.OnShortcut(0, back, func() { change(s.Index - 1) }).OnShortcut(0, next, func() { change(s.Index + 1) }).OnShortcut(0, ui.KeyHome, func() { change(0) }).OnShortcut(0, ui.KeyEnd, func() { change(len(items) - 1) })
		if len(items) == 0 {
			view.Children(func() { ui.Text(c, message(c, "noSlides")).Padding(16) })
			s.last = time.Time{}
			return
		}
		if s.identity != items[s.Index].ID {
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
		dx, dy, held := view.Dragged()
		if s.vertical != vert {
			s.vertical, s.offset, s.held, s.ignoreDrag = vert, 0, false, held
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
		vx, vy, ended := view.DragEnded()
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
			change(s.Index + delta)
			s.held = false
		} else if !held && s.held {
			s.held = false
			s.offset = 0
		}
		view.Children(func() {
			for i := max(0, s.Index-1); i <= min(len(items)-1, s.Index+1); i++ {
				it := items[i]
				pane := ui.Column(c.Key(it.ID)).Absolute().WidthPercent(100).Padding(20).MinWidth(0).Label(it.Label)
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
				finish(t, ComponentCarousel, Item, pane, PartStyle{}, o.Decoration)
				pane.Children(func() {
					if it.Build != nil {
						it.Build(c, t)
					} else {
						ui.Text(c, it.Label).FontSize(t.text(18))
					}
				})
			}
		})
		running := o.Autoplay > 0 && len(items) > 1 && !s.Paused && !s.held && !t.Motion.Reduce && !c.Preferences().ReduceMotion && motionVisible(c, view, o.Active) && !root.Hovered() && !s.handle.FocusWithin(c)
		if running {
			if s.last.IsZero() {
				s.last = c.Now()
			}
			if c.Now().Sub(s.last) >= o.Autoplay {
				if o.Wrap || s.Index < len(items)-1 {
					change(s.Index + 1)
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
		ui.Row(c).Wrap().AlignItems(ui.Center).Gap(8).Children(func() {
			if valueOr(o.Arrows, false) {
				Button(c.Key("previous"), t, message(c, "previous"), ButtonOptions{AccessibleLabel: namedMessage(c, "previousNamed", o.Label), Variant: Outline, Color: Neutral, Disabled: !o.Wrap && s.Index == 0}).OnClick(func() { change(s.Index - 1) })
			}
			ui.Text(c, fmt.Sprintf("%d / %d", s.Index+1, len(items))).Grow(1).Role(ui.RoleStatus).TextColor(t.Muted)
			if o.Autoplay > 0 {
				label := namedMessage(c, "pauseNamed", o.Label)
				if s.Paused {
					label = namedMessage(c, "playNamed", o.Label)
				}
				Button(c.Key("play-pause"), t, label, ButtonOptions{Variant: Ghost, Color: Neutral}).OnClick(func() { s.Paused = !s.Paused; s.last = time.Time{} })
			}
			if valueOr(o.Arrows, false) {
				Button(c.Key("next"), t, message(c, "next"), ButtonOptions{AccessibleLabel: namedMessage(c, "nextNamed", o.Label), Variant: Outline, Color: Neutral, Disabled: !o.Wrap && s.Index == len(items)-1}).OnClick(func() { change(s.Index + 1) })
			}
			if o.Dots {
				for i, it := range items {
					button(c.Key("dot:"+it.ID), t, "", ButtonOptions{AccessibleLabel: formatMessage(c, "goToSlide", map[string]any{"number": i + 1}), Square: true, Variant: Ghost, Color: Neutral}, Set(s.Index == i)).Size(20, 20).OnClick(func() { change(i) }).Children(func() {
						color := t.BorderAccented
						if s.Index == i {
							color = t.Primary
						}
						ui.Box(c).Size(6, 6).Radius(999).Background(color)
					})
				}
			}
		})
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
	// ContentKey resets the phase when content changes. Children must be presentation-only.
	ContentKey string
	Decoration
}

func Marquee(c *ui.Context, t Theme, s *MarqueeState, o MarqueeOptions, content func()) ui.Element {
	if s == nil {
		panic("nui: Marquee requires state")
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
		gap = 24
	}
	root := ui.Box(c).MinWidth(0).Height(h).Clip().Focusable().FocusRing(true).Label(o.Label).Role(ui.RoleGroup)
	finish(t, ComponentMarquee, Root, root, PartStyle{}, o.Decoration)
	root.OnShortcut(0, ui.KeySpace, func() { s.Paused = !s.Paused })
	if s.content != o.ContentKey {
		s.content = o.ContentKey
		s.Offset = 0
		s.last = time.Time{}
	}
	var measure ui.Element
	root.Children(func() {
		measure = ui.Row(c.Key("content")).Absolute().Left(0).Top(0).Gap(gap).IgnoreInput(true)
		if o.Vertical {
			measure.Column()
		}
		measure.Children(func() {
			if content != nil {
				content()
			}
		})
		length := measure.Bounds().W
		if o.Vertical {
			length = measure.Bounds().H
		}
		extent := root.Bounds().W
		if o.Vertical {
			extent = root.Bounds().H
		}
		period := length + gap
		running := period > gap && length > extent && !s.Paused && (!o.PauseOnHover || !root.Hovered()) && !root.FocusWithin() && !t.Motion.Reduce && !c.Preferences().ReduceMotion && motionVisible(c, root, o.Active)
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
		if t.Motion.Reduce || c.Preferences().ReduceMotion || length <= extent {
			offset = 0
			s.Offset = 0
		}
		if o.Vertical {
			measure.Translate(0, -offset)
		} else {
			measure.Translate(-offset, 0)
		}
		if length > extent && period > gap {
			copy := ui.Row(c.Key("repeat")).Absolute().Left(0).Top(0).Gap(gap).Role(ui.RoleNone).IgnoreInput(true)
			if o.Vertical {
				copy.Column().Translate(0, period-offset)
			} else {
				copy.Translate(period-offset, 0)
			}
			copy.Children(func() {
				if content != nil {
					content()
				}
			})
		}
	})
	return root
}
