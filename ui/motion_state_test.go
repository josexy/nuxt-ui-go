package nui_test

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

// switchView is an MD Switch whose track and thumb are labelled.
func switchView(on *bool, loading *bool, changes *int) func(c *ui.Context, th nui.Theme) {
	return func(c *ui.Context, th nui.Theme) {
		nui.Switch(c, th, on, nui.SwitchOptions{Label: "Sync", Loading: *loading, OnChange: func(bool) { *changes++ }, Style: func(slot nui.Slot, e ui.Element) {
			switch slot {
			case nui.Indicator:
				e.Label("Track")
			case nui.Thumb:
				e.Label("Thumb")
			}
		}})
	}
}

func TestSwitchThumbAndTrackAtEachLevel(t *testing.T) {
	for _, l := range motionLevels {
		for _, rtl := range []bool{false, true} {
			name := l.name
			if rtl {
				name += "/rtl"
			}
			t.Run(name, func(t *testing.T) {
				on, loading, changes := false, false, 0
				m := newMotionTester(l, 300, 100, switchView(&on, &loading, &changes))
				if rtl {
					m.SetDirection(ui.RightToLeft)
				}
				m.Advance(300 * time.Millisecond)
				track := bounds(t, m.Tester, "Track")
				trackAt := func() color.RGBA { return m.at(track.X+track.W/2, track.Y+2) }
				idleX, idleColor := bounds(t, m.Tester, "Thumb").X, trackAt()
				click(t, m.Tester, "Sync")
				m.Move(295, 95)
				first := bounds(t, m.Tester, "Thumb").X
				m.Advance(100 * time.Millisecond)
				midX, midColor := bounds(t, m.Tester, "Thumb").X, trackAt()
				m.Advance(150 * time.Millisecond)
				endX, endColor := bounds(t, m.Tester, "Thumb").X, trackAt()
				if on != true || changes != 1 {
					t.Fatalf("toggling set on=%v with %d OnChange calls", on, changes)
				}
				if toward := endX < idleX; toward != rtl || endX == idleX {
					t.Fatalf("the thumb went from %v to %v", idleX, endX)
				}
				if endColor == idleColor {
					t.Fatalf("the track kept its colour %v", endColor)
				}
				switch {
				case l == motionFull:
					if !(min(idleX, endX) < midX && midX < max(idleX, endX)) || !between(midColor, idleColor, endColor) {
						t.Fatalf("at 100 ms the thumb was at %v (from %v to %v) and the track %v", midX, idleX, endX, midColor)
					}
				case l.reduced():
					if first != endX || !between(midColor, idleColor, endColor) {
						t.Fatalf("at Reduced the thumb was at %v, not %v, in the next frame, or the track %v did not fade", first, endX, midColor)
					}
				default:
					if first != endX || midColor != endColor {
						t.Fatalf("at Off the thumb (%v) or the track (%v) did not change at once", first, midColor)
					}
				}
			})
		}
	}
}

func TestLoadingSwitchStillBlocksToggling(t *testing.T) {
	on, loading, changes := false, true, 0
	m := newMotionTester(motionFull, 300, 100, switchView(&on, &loading, &changes))
	before := bounds(t, m.Tester, "Thumb").X
	click(t, m.Tester, "Sync")
	m.Key(0, ui.KeySpace)
	m.Advance(300 * time.Millisecond)
	if on || changes != 0 || bounds(t, m.Tester, "Thumb").X != before {
		t.Fatalf("a loading Switch toggled: on=%v changes=%d", on, changes)
	}
}

// A subtree with motion disabled keeps its Switch still while one outside it
// slides in the same frame.
func TestThemedSubtreeSwitchOptsOut(t *testing.T) {
	inside, outside := false, false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		thumb := func(label string) func(nui.Slot, ui.Element) {
			return func(slot nui.Slot, e ui.Element) {
				if slot == nui.Thumb {
					e.Label(label)
				}
			}
		}
		ui.Column(c).Padding(8).Gap(8).Children(func() {
			nui.Switch(c.Key("outside"), th, &outside, nui.SwitchOptions{Label: "Outside", Style: thumb("Outside thumb")})
			still := th
			still.Motion.Disabled = true
			nui.Themed(c.Key("inside"), still, func(th nui.Theme) {
				nui.Switch(c, th, &inside, nui.SwitchOptions{Label: "Inside", Style: thumb("Inside thumb")})
			})
		})
	}, 300, 200)
	beforeIn, beforeOut := bounds(t, tt, "Inside thumb").X, bounds(t, tt, "Outside thumb").X
	inside, outside = true, true
	tt.Frame()
	tt.Advance(100 * time.Millisecond)
	midIn, midOut := bounds(t, tt, "Inside thumb").X, bounds(t, tt, "Outside thumb").X
	tt.Advance(200 * time.Millisecond)
	endIn, endOut := bounds(t, tt, "Inside thumb").X, bounds(t, tt, "Outside thumb").X
	if midIn != endIn || endIn == beforeIn {
		t.Fatalf("the Switch with motion disabled slid: %v, %v at 100 ms, %v", beforeIn, midIn, endIn)
	}
	if !(beforeOut < midOut && midOut < endOut) {
		t.Fatalf("the Switch outside did not slide: %v, %v at 100 ms, %v", beforeOut, midOut, endOut)
	}
}

// The default themes animate: a Switch toggled under a zero AppConfig moves
// its thumb over several frames.
func TestDefaultThemesAnimateTheSwitch(t *testing.T) {
	on := false
	tt := ui.NewTester(func(c *ui.Context) {
		nui.Switch(c, nui.App(c, nui.AppConfig{}), &on, nui.SwitchOptions{Label: "Sync", Style: func(slot nui.Slot, e ui.Element) {
			if slot == nui.Thumb {
				e.Label("Thumb")
			}
		}})
	}, 300, 100)
	seen := map[float32]bool{}
	click(t, tt, "Sync")
	for range 12 {
		seen[bounds(t, tt, "Thumb").X] = true
		tt.Advance(20 * time.Millisecond)
	}
	if len(seen) < 4 {
		t.Fatalf("the thumb showed %d positions while it moved", len(seen))
	}
}

// chevronLabel labels a tree node's trailing icon.
func chevronLabel(label string) nui.Decoration {
	return nui.Decoration{Style: func(slot nui.Slot, e ui.Element) {
		if slot == nui.Trailing {
			e.Label(label)
		}
	}}
}

// The marker icon's bar points up (-90 degrees) at rest and down (90) turned
// over; sideways it is near 0 or 180.
func TestTreeChevronTurnsAtEachLevel(t *testing.T) {
	for _, l := range motionLevels {
		t.Run(l.name, func(t *testing.T) {
			var s nui.TreeState
			s.SetNodes([]nui.TreeNode{{ID: "src", Label: "src", Decoration: chevronLabel("Chevron"), Children: []nui.TreeNode{{ID: "main", Label: "main.go"}}}})
			m := newMotionTester(l, 400, 300, func(c *ui.Context, th nui.Theme) {
				nui.Tree(c, th, &s, nui.TreeOptions{Label: "Files", TrailingIcon: spinnerMarker}).Width(300).Height(200)
			})
			r := bounds(t, m.Tester, "Chevron")
			if a := markerAngle(t, m.Image(), r); math.Abs(a+90) > 15 {
				t.Fatalf("a collapsed chevron points at %.0f degrees", a)
			}
			click(t, m.Tester, "src")
			if !s.Outline.Open.Has("src") {
				t.Fatal("the folder did not expand")
			}
			first := markerAngle(t, m.Image(), r)
			m.Advance(50 * time.Millisecond)
			mid := markerAngle(t, m.Image(), r)
			m.Advance(150 * time.Millisecond)
			if end := markerAngle(t, m.Image(), r); math.Abs(end-90) > 15 {
				t.Fatalf("at 200 ms the chevron points at %.0f degrees, want 90", end)
			}
			if l == motionFull {
				if mid < -70 || mid > 70 {
					t.Fatalf("the chevron was not partway at 50 ms: %.0f degrees", mid)
				}
			} else if math.Abs(first-90) > 15 {
				t.Fatalf("the chevron did not turn over at once: %.0f degrees", first)
			}
		})
	}
}

func TestTreeRowScrolledInAlreadyOpenShowsTurned(t *testing.T) {
	nodes := make([]nui.TreeNode, 60)
	for i := range nodes {
		nodes[i] = nui.TreeNode{ID: fmt.Sprint(i), Label: fmt.Sprintf("Node %02d", i)}
	}
	nodes[50] = nui.TreeNode{ID: "deep", Label: "Deep folder", DefaultExpanded: true, Decoration: chevronLabel("Deep chevron"), Children: []nui.TreeNode{{ID: "leaf", Label: "Leaf"}}}
	var s nui.TreeState
	s.SetNodes(nodes)
	scroll := false
	m := newMotionTester(motionFull, 400, 300, func(c *ui.Context, th nui.Theme) {
		if scroll {
			scroll = false
			s.Outline.List.ScrollTo(50, ui.Start)
		}
		nui.Tree(c, th, &s, nui.TreeOptions{Label: "Files", TrailingIcon: spinnerMarker}).Width(300).Height(200)
	})
	if _, ok := m.Find("Deep chevron"); ok {
		t.Fatal("the deep folder was built before it scrolled into view")
	}
	scroll = true
	m.Frame()
	m.Frame()
	r := bounds(t, m.Tester, "Deep chevron")
	if a := markerAngle(t, m.Image(), r); math.Abs(a-90) > 15 {
		t.Fatalf("an open folder scrolled into view points at %.0f degrees", a)
	}
	if m.FramePending() {
		t.Fatal("an open folder scrolled into view animated its chevron")
	}
}

func TestNavigationMenuTurnsOneChevron(t *testing.T) {
	for _, custom := range []bool{false, true} {
		t.Run(fmt.Sprint("custom=", custom), func(t *testing.T) {
			active := "home"
			var nav nui.NavigationState
			o := nui.NavigationOptions{}
			if custom {
				o.TrailingIcon = spinnerMarker
			}
			m := newMotionTester(motionFull, 600, 300, func(c *ui.Context, th nui.Theme) {
				nui.NavigationMenu(c, th, &active, &nav, []nui.NavigationItem{{ID: "home", Label: "Home"}, {ID: "docs", Label: "Docs", Children: []nui.NavigationItem{{ID: "guide", Label: "Guide"}}}}, o)
			})
			r := bounds(t, m.Tester, "Docs")
			size := float32(20)
			icon := ui.Rect{X: r.X + r.W - 10 - size, Y: r.Y + (r.H-size)/2, W: size, H: size}
			closed := clone(m)
			click(t, m.Tester, "Docs")
			m.Move(595, 295)
			m.Advance(50 * time.Millisecond)
			mid := clone(m)
			m.Advance(300 * time.Millisecond)
			if custom {
				// The marker's bar turns from up (-90 degrees) to down (90).
				if a := markerAngle(t, closed, icon); math.Abs(a+90) > 15 {
					t.Fatalf("closed, the icon points at %.0f degrees", a)
				}
				if a := markerAngle(t, mid, icon); a < -70 || a > 70 {
					t.Fatalf("at 50 ms the icon was not partway: %.0f degrees", a)
				}
				if a := markerAngle(t, m.Image(), icon); math.Abs(a-90) > 15 {
					t.Fatalf("open, the icon points at %.0f degrees", a)
				}
				return
			}
			// Turned partway, the chevron is neither the closed nor the open
			// one: it turns rather than being swapped for an up chevron.
			open := clone(m)
			differs := func(a, b *image.RGBA) bool {
				for y := int(icon.Y); y < int(icon.Y+icon.H); y++ {
					for x := int(icon.X); x < int(icon.X+icon.W); x++ {
						if d := int(a.RGBAAt(x, y).R) - int(b.RGBAAt(x, y).R); d > 80 || d < -80 {
							return true
						}
					}
				}
				return false
			}
			if !differs(mid, closed) || !differs(mid, open) || !differs(open, closed) {
				t.Fatal("the chevron did not turn from closed to open")
			}
		})
	}
}

func verticalNavigation(active *string, nav *nui.NavigationState, navigated *[]string) func(c *ui.Context, th nui.Theme) {
	items := []nui.NavigationItem{
		{ID: "docs", Label: "Docs", Children: []nui.NavigationItem{{ID: "guide", Label: "Guide"}, {ID: "api", Label: "API"}}},
		{ID: "blog", Label: "Blog"},
	}
	return func(c *ui.Context, th nui.Theme) {
		nui.NavigationMenu(c, th, active, nav, items, nui.NavigationOptions{Vertical: true, OnNavigate: func(id string) { *navigated = append(*navigated, id) }}).Width(240)
	}
}

func TestVerticalNavigationChildrenAtEachLevel(t *testing.T) {
	for _, l := range motionLevels {
		t.Run(l.name, func(t *testing.T) {
			active, navigated := "blog", []string{}
			var nav nui.NavigationState
			m := newMotionTester(l, 400, 400, verticalNavigation(&active, &nav, &navigated))
			closed := bounds(t, m.Tester, "Blog").Y
			click(t, m.Tester, "Docs")
			first := bounds(t, m.Tester, "Blog").Y
			m.Advance(100 * time.Millisecond)
			mid := bounds(t, m.Tester, "Blog").Y
			m.Advance(200 * time.Millisecond)
			open := bounds(t, m.Tester, "Blog").Y
			if open <= closed+40 {
				t.Fatalf("the children did not open: Blog moved from %v to %v", closed, open)
			}
			if l == motionFull {
				if !(closed < mid && mid < open) {
					t.Fatalf("at 100 ms Blog was at %v, not between %v and %v", mid, closed, open)
				}
				return
			}
			if first != open {
				t.Fatalf("the children did not open in one frame: Blog at %v, then %v", first, open)
			}
			if l.off {
				m.idle(t, time.Second, "open children at Off")
			}
		})
	}
}

func TestVerticalNavigationCollapseWithFocusInside(t *testing.T) {
	active, navigated := "blog", []string{}
	var nav nui.NavigationState
	m := newMotionTester(motionFull, 400, 400, verticalNavigation(&active, &nav, &navigated))
	click(t, m.Tester, "Docs")
	m.Advance(300 * time.Millisecond)
	m.Key(0, ui.KeyDown) // the menu is a vertical focus group
	if !m.Focused("Guide") {
		t.Fatal("Down did not focus the first child")
	}
	nav.Expanded["docs"] = false
	m.Frame()
	if !m.Focused("Docs") {
		t.Fatal("collapsing did not move focus to the parent at once")
	}
	m.Advance(17 * time.Millisecond)
	guide, ok := m.Find("Guide")
	if !ok {
		t.Fatal("the children vanished instead of closing")
	}
	// The first child still shows at the top of the closing clip, above the
	// item after it.
	if blog := bounds(t, m.Tester, "Blog"); blog.Y <= guide.Y+3 {
		t.Fatalf("Guide is no longer visible while closing: Guide %v, Blog %v", guide, blog)
	}
	m.ClickAt(guide.X+guide.W/2, guide.Y+3)
	if len(navigated) != 0 || active != "blog" {
		t.Fatalf("closing children took a click: navigated %v, active %q", navigated, active)
	}
	m.Advance(300 * time.Millisecond)
	if _, ok := m.Find("Guide"); ok {
		t.Fatal("the children stayed after closing")
	}
}

// sidebarView builds a Sidebar beside a Main area whose content is labelled.
func sidebarView(collapsed *bool, state *nui.SidebarState, transition nui.Optional[bool], draft *string) func(c *ui.Context, th nui.Theme) {
	return func(c *ui.Context, th nui.Theme) {
		ui.Row(c).Fill().Children(func() {
			nui.Sidebar(c.Key("sidebar"), th, collapsed, nui.SidebarOptions{Title: "Files", State: state, Resizable: true, Breakpoint: nui.Set(float32(500)), Transition: transition}, func(c *ui.Context, th nui.Theme, compact bool) {
				ui.Text(c, "Entry")
			})
			nui.Main(c.Key("main"), th, func() {
				ui.Box(c).Height(10).FillWidth().Label("Main start")
				nui.Input(c.Key("draft"), th, draft, nui.InputOptions{Label: "Draft"})
			}).Grow(1)
		})
	}
}

func TestSidebarCollapseAtEachLevel(t *testing.T) {
	type run struct {
		level      motionLevel
		transition nui.Optional[bool]
	}
	runs := []run{{motionFull, nui.Optional[bool]{}}, {motionReducedDesktop, nui.Optional[bool]{}}, {motionReducedTheme, nui.Optional[bool]{}}, {motionOff, nui.Optional[bool]{}}, {motionFull, nui.Set(false)}}
	for _, r := range runs {
		t.Run(fmt.Sprintf("%s/transition=%v", r.level.name, r.transition), func(t *testing.T) {
			collapsed, draft := false, ""
			var state nui.SidebarState
			m := newMotionTester(r.level, 900, 400, sidebarView(&collapsed, &state, r.transition, &draft))
			if w := bounds(t, m.Tester, "Files").W; w != 256 {
				t.Fatalf("expanded width %v", w)
			}
			click(t, m.Tester, "Toggle sidebar")
			first := bounds(t, m.Tester, "Files")
			m.Advance(100 * time.Millisecond)
			mid, main := bounds(t, m.Tester, "Files"), bounds(t, m.Tester, "Main start")
			if mid.X+mid.W > main.X+.5 || mid.X+mid.W < main.X-.5 {
				t.Fatalf("Main's start %v does not follow the Sidebar's end %v", main.X, mid.X+mid.W)
			}
			m.Advance(200 * time.Millisecond)
			if w := bounds(t, m.Tester, "Files").W; w != 64 || !collapsed {
				t.Fatalf("collapsed width %v (collapsed=%v)", w, collapsed)
			}
			if r.level == motionFull && !r.transition.Set {
				if !(64 < mid.W && mid.W < 256) {
					t.Fatalf("at 100 ms the width was %v, not between 64 and 256", mid.W)
				}
			} else if first.W != 64 {
				t.Fatalf("the width did not change at once: %v", first.W)
			}
		})
	}
}

func TestSidebarDragFollowsThePointer(t *testing.T) {
	collapsed, draft := false, ""
	var state nui.SidebarState
	m := newMotionTester(motionFull, 900, 400, sidebarView(&collapsed, &state, nui.Optional[bool]{}, &draft))
	r := bounds(t, m.Tester, "Resize sidebar")
	x, y := r.X+r.W/2, r.Y+r.H/2
	m.Press(x, y)
	m.Move(x+20, y)
	m.Move(x+40, y)
	if w := bounds(t, m.Tester, "Files").W; w != 296 {
		t.Fatalf("a drag of 40 DIPs showed a width of %v in its frame, want 296", w)
	}
	m.Release(x+40, y)
	if b, _ := m.span(300 * time.Millisecond); b > 12 {
		t.Fatalf("the released drag kept building: %d", b)
	}
}

func TestSidebarBreakpointKeepsFocusAndModel(t *testing.T) {
	collapsed, draft := false, ""
	var state nui.SidebarState
	m := newMotionTester(motionFull, 900, 400, sidebarView(&collapsed, &state, nui.Optional[bool]{}, &draft))
	click(t, m.Tester, "Draft")
	m.Type("kept")
	m.SetSize(400, 400)
	m.Advance(300 * time.Millisecond)
	if !m.Focused("Draft") || draft != "kept" || collapsed || state.Open {
		t.Fatalf("crossing the breakpoint lost focus (%v) or the model: draft %q collapsed %v open %v", m.Focused("Draft"), draft, collapsed, state.Open)
	}
	m.SetSize(900, 400)
	m.Advance(300 * time.Millisecond)
	if !m.Focused("Draft") || draft != "kept" || bounds(t, m.Tester, "Files").W != 256 {
		t.Fatalf("crossing back lost focus (%v), the draft %q or the width %v", m.Focused("Draft"), draft, bounds(t, m.Tester, "Files").W)
	}
}

// drawerTop returns where the Drawer's handle, at the top of its panel, is.
func drawerTop(t *testing.T, m *motionTester) float32 {
	t.Helper()
	return bounds(t, m.Tester, "Resize Sheet").Y
}

func TestDrawerSnapsAtEachLevel(t *testing.T) {
	for _, l := range motionLevels {
		t.Run(l.name, func(t *testing.T) {
			s := nui.DrawerState{Open: true}
			snaps := []int{}
			m := newMotionTester(l, 500, 600, func(c *ui.Context, th nui.Theme) {
				nui.Drawer(c, th, &s, nui.DrawerOptions{Title: "Sheet", OnSnap: func(i int) { snaps = append(snaps, i) }}, func(nui.Theme) {
					ui.Text(c, "Content")
				})
			})
			m.Advance(400 * time.Millisecond) // the entrance
			low := drawerTop(t, m)
			click(t, m.Tester, "Resize Sheet")
			m.Key(0, ui.KeyUp)
			if s.Snap != 1 || len(snaps) != 1 || snaps[0] != 1 {
				t.Fatalf("Up snapped to %d with OnSnap %v", s.Snap, snaps)
			}
			first := drawerTop(t, m)
			m.Advance(100 * time.Millisecond)
			mid := drawerTop(t, m)
			m.Advance(100 * time.Millisecond)
			high := drawerTop(t, m)
			if d := low - high; d < 600*(.85-.5)-2 || d > 600*(.85-.5)+2 {
				t.Fatalf("the panel grew %v DIPs, want 210", d)
			}
			if l == motionFull {
				if !(high < mid && mid < low) {
					t.Fatalf("at 100 ms the panel's top was %v, not between %v and %v", mid, low, high)
				}
			} else if first != high {
				t.Fatalf("the snap did not show at once: %v, then %v", first, high)
			}
			if len(snaps) != 1 {
				t.Fatalf("OnSnap ran %d times", len(snaps))
			}
		})
	}
}

func TestDrawerDragFollowsThePointer(t *testing.T) {
	s := nui.DrawerState{Open: true}
	m := newMotionTester(motionFull, 500, 600, func(c *ui.Context, th nui.Theme) {
		nui.Drawer(c, th, &s, nui.DrawerOptions{Title: "Sheet"}, func(nui.Theme) { ui.Text(c, "Content") })
	})
	m.Advance(400 * time.Millisecond)
	r := bounds(t, m.Tester, "Resize Sheet")
	x, y := r.X+r.W/2, r.Y+r.H/2
	m.Press(x, y)
	m.Move(x, y-20)
	m.Move(x, y-40)
	m.Move(x, y-60)
	// A move shows in the frame after the one that reads it.
	if top := drawerTop(t, m); top > r.Y-35 || top < r.Y-61 {
		t.Fatalf("dragging 60 DIPs up moved the panel's top from %v to %v", r.Y, top)
	}
	if m.FramePending() {
		t.Fatal("a held drag animated the height")
	}
	dragged := drawerTop(t, m)
	m.Release(x, y-60)
	m.Advance(17 * time.Millisecond)
	// Released, it eases back to its snap from where it was let go.
	if top := drawerTop(t, m); top <= dragged-1 || top >= r.Y {
		t.Fatalf("the release jumped: dragged to %v, then %v, snap at %v", dragged, top, r.Y)
	}
	m.Advance(time.Second)
	if top := drawerTop(t, m); top != r.Y {
		t.Fatalf("a short drag did not settle back at its snap: %v, want %v", top, r.Y)
	}
}

// tabRect returns the bounds of the tab trigger named label.
func tabRect(t *testing.T, m *motionTester, label string) ui.Rect {
	t.Helper()
	info, ok := m.AccessByRole(ui.RoleTab, label)
	if !ok {
		t.Fatalf("missing tab %q", label)
	}
	return info.Bounds
}

func tabsView(value *string, o nui.TabsOptions, labelled string) func(c *ui.Context, th nui.Theme) {
	items := []nui.TabItem{{ID: "a", Label: "First"}, {ID: "b", Label: "Second"}, {ID: "c", Label: "Third"}}
	o.Label, o.Content = "Sections", nui.Set(false)
	style := o.Style
	o.Style = func(slot nui.Slot, e ui.Element) {
		if slot == nui.Indicator {
			e.Label(labelled)
		}
		if style != nil {
			style(slot, e)
		}
	}
	return func(c *ui.Context, th nui.Theme) {
		nui.Tabs(c.Key(labelled), th, value, items, o).Width(420)
	}
}

func TestTabsPillSlidesUnderTheLabels(t *testing.T) {
	value := "a"
	m := newMotionTester(motionFull, 500, 200, tabsView(&value, nui.TabsOptions{}, "Pill"))
	first, third := tabRect(t, m, "First"), tabRect(t, m, "Third")
	label := bounds(t, m.Tester, "Third")
	before := clone(m)
	click(t, m.Tester, "Third")
	m.Move(495, 195)
	m.Advance(100 * time.Millisecond)
	pill, mid := bounds(t, m.Tester, "Pill"), clone(m)
	m.Advance(200 * time.Millisecond)
	end, after := bounds(t, m.Tester, "Pill"), clone(m)
	if !(first.X < pill.X && pill.X < third.X) || end.X != third.X || end.W != third.W {
		t.Fatalf("the pill went to %v at 100 ms and %v at rest (first %v, third %v)", pill, end, first, third)
	}
	// The whitest pixel of the third label, a glyph's core, shows the text's
	// colour, on its way from idle to active at 100 ms.
	var at [2]int
	best := -1
	for y := int(label.Y); y < int(label.Y+label.H); y++ {
		for x := int(label.X); x < int(label.X+label.W); x++ {
			p := after.RGBAAt(x, y)
			if v := int(min(p.R, p.G, p.B)); v > best {
				best, at = v, [2]int{x, y}
			}
		}
	}
	idle, mids, active := before.RGBAAt(at[0], at[1]), mid.RGBAAt(at[0], at[1]), after.RGBAAt(at[0], at[1])
	if active != rgba(nui.Light().Inverted) || !between(mids, idle, active) {
		t.Fatalf("the third label's colour went %v -> %v -> %v", idle, mids, active)
	}
	// Drawn under the labels: the active label's text shows over it.
	if countColor(after, label, rgba(nui.Light().Inverted)) == 0 {
		t.Fatal("the pill covers the label it slid under")
	}
}

func TestTabsVerticalUnderlineTravels(t *testing.T) {
	value := "a"
	m := newMotionTester(motionFull, 500, 300, tabsView(&value, nui.TabsOptions{Underline: true, Vertical: true}, "Line"))
	first, third := tabRect(t, m, "First"), tabRect(t, m, "Third")
	start := bounds(t, m.Tester, "Line")
	value = "c"
	m.Frame()
	m.Advance(100 * time.Millisecond)
	mid := bounds(t, m.Tester, "Line")
	m.Advance(100 * time.Millisecond)
	end := bounds(t, m.Tester, "Line")
	if start.Y != first.Y || end.Y != third.Y || !(start.Y < mid.Y && mid.Y < end.Y) || mid.X != start.X {
		t.Fatalf("the line went %v -> %v -> %v (first %v, third %v)", start, mid, end, first, third)
	}
}

func TestTwoTabsKeepTheirIndicators(t *testing.T) {
	one, two := "a", "c"
	m := newMotionTester(motionFull, 500, 300, func(c *ui.Context, th nui.Theme) {
		tabsView(&one, nui.TabsOptions{}, "One")(c, th)
		tabsView(&two, nui.TabsOptions{}, "Two")(c, th)
	})
	stay := bounds(t, m.Tester, "Two")
	start := bounds(t, m.Tester, "One")
	one = "b"
	m.Frame()
	for range 12 {
		m.Advance(20 * time.Millisecond)
		if got := bounds(t, m.Tester, "Two"); got != stay {
			t.Fatalf("the second Tabs' indicator moved: %v -> %v", stay, got)
		}
		if got := bounds(t, m.Tester, "One"); got.Y != start.Y {
			t.Fatalf("the first Tabs' indicator left its list: %v -> %v", start, got)
		}
	}
}

func TestTabsIndicatorAtEachLevel(t *testing.T) {
	for _, l := range motionLevels {
		for _, underline := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/underline=%v", l.name, underline), func(t *testing.T) {
				value := "a"
				m := newMotionTester(l, 500, 200, tabsView(&value, nui.TabsOptions{Underline: underline}, "Indicator"))
				third := tabRect(t, m, "Third")
				click(t, m.Tester, "Third")
				got := bounds(t, m.Tester, "Indicator")
				if arrived := got.X == third.X; arrived == (l == motionFull) {
					t.Fatalf("in the next frame the indicator was at %v (third at %v)", got, third)
				}
			})
		}
	}
}

func TestTabsIndicatorStylesFollowTheSlide(t *testing.T) {
	info := ui.RGB(0, 120, 255)
	for _, source := range []string{"theme", "instance"} {
		t.Run(source, func(t *testing.T) {
			value := "a"
			o := nui.TabsOptions{}
			if source == "instance" {
				o.UI = nui.Styles{nui.Indicator: {Background: nui.Set(info)}}
			}
			view := tabsView(&value, o, "Pill")
			m := newMotionTester(motionFull, 500, 200, func(c *ui.Context, th nui.Theme) {
				if source == "theme" {
					th = th.WithComponent(nui.ComponentTabs, nui.ComponentTheme{UI: nui.Styles{nui.Indicator: {Background: nui.Set(info)}}})
				}
				view(c, th)
			})
			click(t, m.Tester, "Third")
			m.Move(495, 195)
			for _, step := range []time.Duration{40, 40, 300} {
				m.Advance(step * time.Millisecond)
				r := bounds(t, m.Tester, "Pill")
				if got := m.at(r.X+r.W/2, r.Y+r.H-3); got != rgba(info) {
					t.Fatalf("the pill at %v was %v, want %v", r, got, rgba(info))
				}
			}
		})
	}
}

func TestTabsTriggerBackgroundStillWins(t *testing.T) {
	magenta := ui.RGB(255, 0, 255)
	value := "a"
	m := newMotionTester(motionFull, 500, 200, tabsView(&value, nui.TabsOptions{UI: nui.Styles{nui.Trigger: {Background: nui.Set(magenta)}}}, "Pill"))
	first := tabRect(t, m, "First")
	if got := m.at(first.X+first.W/2, first.Y+first.H-3); got != rgba(magenta) {
		t.Fatalf("the active trigger's surface was %v, want its Trigger background %v", got, rgba(magenta))
	}
}

func TestTabsStyleIndicatorReceivesThePill(t *testing.T) {
	value := "a"
	m := newMotionTester(motionFull, 500, 200, tabsView(&value, nui.TabsOptions{}, "Pill"))
	// The trigger keeps its own name, so the labelled box is another element,
	// filling it.
	first := tabRect(t, m, "First")
	if pill := bounds(t, m.Tester, "Pill"); pill != first {
		t.Fatalf("Style(Indicator) did not receive a box filling the active trigger: %v (trigger %v)", pill, first)
	}
}

// carouselView is a five-page Carousel whose pages count their builds.
func carouselView(s *nui.CarouselState, builds map[string]int, changes *[]int, o nui.CarouselOptions) func(c *ui.Context, th nui.Theme) {
	items := make([]nui.CarouselItem, 5)
	for i := range items {
		id := fmt.Sprintf("Page %d", i+1)
		items[i] = nui.CarouselItem{ID: id, Label: id, Build: func(c *ui.Context, _ nui.Theme) {
			builds[id]++
			ui.Text(c, id)
		}}
	}
	o.Label, o.Height, o.Arrows, o.Dots = "Slides", 120, nui.Set(true), true
	o.OnChange = func(i int) { *changes = append(*changes, i) }
	return func(c *ui.Context, th nui.Theme) {
		nui.Carousel(c, th, s, items, o).Width(400)
	}
}

func TestCarouselNextSlidesAtEachDirection(t *testing.T) {
	for _, dir := range []string{"ltr", "rtl", "vertical"} {
		t.Run(dir, func(t *testing.T) {
			var s nui.CarouselState
			builds, changes := map[string]int{}, []int{}
			o := nui.CarouselOptions{}
			if dir == "vertical" {
				o.Orientation = nui.Vertical
			}
			m := newMotionTester(motionFull, 600, 400, carouselView(&s, builds, &changes, o))
			if dir == "rtl" {
				m.SetDirection(ui.RightToLeft)
			}
			view := bounds(t, m.Tester, "Page 1")
			click(t, m.Tester, "Next Slides")
			m.Move(595, 395)
			m.Advance(100 * time.Millisecond)
			m.Frame() // bounds as the frames that only paint show them
			// Bounds are what shows through the viewport: midway both pages
			// show in part, page 1 on the start side and page 2 on the end.
			one, two := bounds(t, m.Tester, "Page 1"), bounds(t, m.Tester, "Page 2")
			along := func(r ui.Rect) (at, length float32) {
				if dir == "vertical" {
					return r.Y, r.H
				}
				return r.X, r.W
			}
			a, al := along(one)
			b, bl := along(two)
			_, size := along(view)
			ordered := a < b
			if dir == "rtl" {
				ordered = a > b
			}
			if !(al > 0 && al < size && bl > 0 && bl < size && ordered) {
				t.Fatalf("at 100 ms page 1 showed %v and page 2 %v of %v", one, two, view)
			}
			m.Advance(200 * time.Millisecond)
			m.Frame()
			if r := bounds(t, m.Tester, "Page 2"); r != view || len(changes) != 1 || changes[0] != 1 {
				t.Fatalf("page 2 arrived at %v (want %v), OnChange %v", r, view, changes)
			}
		})
	}
}

func TestCarouselDragReleaseContinues(t *testing.T) {
	var s nui.CarouselState
	builds, changes := map[string]int{}, []int{}
	m := newMotionTester(motionFull, 600, 400, carouselView(&s, builds, &changes, nui.CarouselOptions{}))
	view := bounds(t, m.Tester, "Page 1")
	x, y := view.X+view.W/2, view.Y+view.H/2
	m.Press(x, y)
	for step := float32(1); step <= 6; step++ {
		m.Move(x-step*view.W*.05, y)
	}
	dragged := bounds(t, m.Tester, "Page 2").X
	if want := view.X + view.W*.7; dragged < want-2 || dragged > want+2 {
		t.Fatalf("a 30%% drag showed page 2 at %v, want %v", dragged, want)
	}
	m.Release(x-view.W*.3, y)
	prev := bounds(t, m.Tester, "Page 2").X
	if prev > dragged+1 {
		t.Fatalf("the release jumped back: %v -> %v", dragged, prev)
	}
	for range 15 {
		m.Advance(16 * time.Millisecond)
		got := bounds(t, m.Tester, "Page 2").X
		if got > prev+.5 {
			t.Fatalf("the slide went back: %v -> %v", prev, got)
		}
		prev = got
	}
	m.Advance(100 * time.Millisecond)
	m.Frame()
	if got := bounds(t, m.Tester, "Page 2").X; got != view.X || len(changes) != 1 || changes[0] != 1 {
		t.Fatalf("the release did not land on page 2: %v, OnChange %v", got, changes)
	}
}

func TestCarouselDotJumpSlidesInOnePage(t *testing.T) {
	var s nui.CarouselState
	builds, changes := map[string]int{}, []int{}
	m := newMotionTester(motionFull, 600, 400, carouselView(&s, builds, &changes, nui.CarouselOptions{}))
	view := bounds(t, m.Tester, "Page 1")
	click(t, m.Tester, "Go to slide 5")
	if s.Index != 4 {
		t.Fatalf("the dot went to %d", s.Index)
	}
	clear(builds) // the press built the first pages before the jump
	m.Move(595, 395)
	var xs []float32
	for range 8 {
		xs = append(xs, bounds(t, m.Tester, "Page 5").X-view.X)
		m.Advance(16 * time.Millisecond)
	}
	m.Advance(200 * time.Millisecond)
	m.Frame() // bounds as the frames that only paint show them
	if xs[0] > view.W+1 || xs[2] <= 0 || xs[2] >= view.W {
		t.Fatalf("page 5 did not slide in across one page: %v", xs)
	}
	if bounds(t, m.Tester, "Page 5") != view || len(changes) != 1 || changes[0] != 4 {
		t.Fatalf("the jump did not land on page 5: OnChange %v", changes)
	}
	for id, n := range builds {
		if id != "Page 1" && id != "Page 4" && id != "Page 5" && n > 0 {
			t.Fatalf("after the jump %s was built %d times: %v", id, n, builds)
		}
	}
}

func TestCarouselPageChangeAtEachLevel(t *testing.T) {
	for _, l := range motionLevels {
		t.Run(l.name, func(t *testing.T) {
			var s nui.CarouselState
			builds, changes := map[string]int{}, []int{}
			m := newMotionTester(l, 600, 400, carouselView(&s, builds, &changes, nui.CarouselOptions{}))
			view := bounds(t, m.Tester, "Page 1")
			click(t, m.Tester, "Next Slides")
			m.Move(595, 395)
			arrived := bounds(t, m.Tester, "Page 2") == view
			if arrived == (l == motionFull) {
				t.Fatalf("page 2 arrived=%v in the next frame", arrived)
			}
			if l.off {
				if b, _ := m.span(2 * time.Second); b > 2 || m.FramePending() {
					t.Fatalf("a page change at Off built %d times in 2 s (pending=%v)", b, m.FramePending())
				}
			}
		})
	}
}
