package nui_test

import (
	"fmt"
	"image"
	"image/color"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

// arrivedShare returns how many pixels differ between before and after, and
// which share of them already show their final colour in got: a colour
// transition partway shows few of them there, and a change at once all of
// them.
func arrivedShare(before, got, after *image.RGBA) (changed int, share float64) {
	arrived := 0
	for i := 0; i+3 < len(before.Pix); i += 4 {
		a, b, g := before.Pix[i:i+3], after.Pix[i:i+3], got.Pix[i:i+3]
		if a[0] == b[0] && a[1] == b[1] && a[2] == b[2] {
			continue
		}
		changed++
		if g[0] == b[0] && g[1] == b[1] && g[2] == b[2] {
			arrived++
		}
	}
	if changed == 0 {
		return 0, 0
	}
	return changed, float64(arrived) / float64(changed)
}

// clone copies the Tester's last frame.
func clone(m *motionTester) *image.RGBA {
	im := m.Image()
	return &image.RGBA{Pix: append([]uint8(nil), im.Pix...), Stride: im.Stride, Rect: im.Rect}
}

// checkFade runs act on a settled window and checks that what it changed
// fades over the 150 ms colour transition when fade is set, and shows in the
// next frame otherwise.
func checkFade(t *testing.T, m *motionTester, fade bool, act func()) {
	t.Helper()
	m.Advance(500 * time.Millisecond)
	before := clone(m)
	act()
	first := clone(m)
	m.Advance(75 * time.Millisecond)
	mid := clone(m)
	m.Advance(600 * time.Millisecond)
	after := clone(m)
	changed, now := arrivedShare(before, first, after)
	_, halfway := arrivedShare(before, mid, after)
	switch {
	case changed == 0:
		t.Fatal("the change painted nothing")
	case fade && (now > .05 || halfway > .25):
		t.Fatalf("%d changed pixels did not fade: %.0f%% final at once, %.0f%% at 75 ms", changed, now*100, halfway*100)
	case !fade && now < .95:
		t.Fatalf("only %.0f%% of %d changed pixels showed in the next frame", now*100, changed)
	}
}

func labelRoot(name string) func(root, _, _, _ ui.Element) {
	return func(root, _, _, _ ui.Element) { root.Label(name) }
}

func TestInputBorderFadesOnFocus(t *testing.T) {
	for _, l := range motionLevels {
		t.Run(l.name, func(t *testing.T) {
			text := ""
			m := newMotionTester(l, 400, 120, func(c *ui.Context, th nui.Theme) {
				nui.Input(c, th, &text, nui.InputOptions{Label: "Name", Style: labelRoot("Field")}).Width(240)
			})
			r := bounds(t, m.Tester, "Field")
			x, y := r.X, r.Y+r.H/2
			idle, focused := rgba(nui.Light().BorderAccented), rgba(nui.Light().Primary)
			if got := m.at(x, y); got != idle {
				t.Fatalf("idle border %v, want %v", got, idle)
			}
			m.Key(0, ui.KeyTab)
			if l.off {
				// MyGo's own caret still blinks in a focused editor.
				if got := m.at(x, y); got != focused {
					t.Fatalf("at Off the focused border did not show in the next frame: %v", got)
				}
				return
			}
			m.Advance(75 * time.Millisecond)
			if mid := m.at(x, y); !between(mid, idle, focused) {
				t.Fatalf("at 75 ms the border was %v, not between %v and %v", mid, idle, focused)
			}
			m.Advance(75 * time.Millisecond)
			if got := m.at(x, y); got != focused {
				t.Fatalf("at 150 ms the border was %v, want %v", got, focused)
			}
		})
	}
}

func TestZeroTransitionLeavesOtherMotion(t *testing.T) {
	text, open := "", false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		th.Motion.Transition = 0
		ui.Column(c).Padding(8).Gap(8).Children(func() {
			nui.Input(c, th, &text, nui.InputOptions{Label: "Name", Style: labelRoot("Field")}).Width(240)
			a := nui.Button(c, th, "Details", nui.ButtonOptions{}).OnClick(func() { open = true })
			nui.Popover(c, th, a, &open, nui.PopoverOptions{Width: nui.Set(float32(200)), Style: func(slot nui.Slot, e ui.Element) {
				if slot == nui.Panel {
					e.Label("Panel")
				}
			}}, func(nui.Theme) { ui.Box(c).Size(10, 40) })
		})
	}, 400, 300)
	r := bounds(t, tt, "Field")
	tt.Key(0, ui.KeyTab)
	if got := tt.Image().RGBAAt(int(r.X), int(r.Y+r.H/2)); got != rgba(nui.Light().Primary) {
		t.Fatalf("with Motion.Transition 0 the focused border did not show at once: %v", got)
	}
	click(t, tt, "Details")
	tt.Advance(50 * time.Millisecond)
	if w := bounds(t, tt, "Panel").W; w >= 200 {
		t.Fatalf("with Motion.Transition 0 the Popover stopped scaling in: width %v", w)
	}
}

func TestFieldSurfacesFadeTheirColors(t *testing.T) {
	type field struct {
		name string
		view func(c *ui.Context, th nui.Theme)
		act  func(t *testing.T, m *motionTester)
	}
	text, tags, choice, query := "", []string{}, "a", nui.SearchState{}
	var number nui.Optional[float64]
	var numberState nui.NumberState
	var tagsState nui.TagsState
	var menu nui.SelectMenuState
	var picked []string
	var pin string
	var pinState nui.PinState
	focus := func(t *testing.T, m *motionTester) { m.Key(0, ui.KeyTab) }
	hover := func(label string) func(t *testing.T, m *motionTester) {
		return func(t *testing.T, m *motionTester) {
			r := bounds(t, m.Tester, label)
			m.Move(r.X+r.W/2, r.Y+r.H/2)
		}
	}
	options := []nui.Option[string]{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}}
	fields := []field{
		{"Input", func(c *ui.Context, th nui.Theme) {
			nui.Input(c, th, &text, nui.InputOptions{Label: "Field"}).Width(240)
		}, focus},
		{"Textarea", func(c *ui.Context, th nui.Theme) {
			nui.Textarea(c, th, &text, nui.TextareaOptions{Label: "Field"}).Width(240)
		}, focus},
		{"InputNumber", func(c *ui.Context, th nui.Theme) {
			nui.InputNumber(c, th, &number, &numberState, nui.InputNumberOptions{Label: "Field"}).Width(240)
		}, focus},
		{"InputTags", func(c *ui.Context, th nui.Theme) {
			nui.InputTags(c, th, &tags, &tagsState, nui.InputTagsOptions{Label: "Field"}).Width(240)
		}, focus},
		{"InputMenu", func(c *ui.Context, th nui.Theme) {
			nui.InputMenu(c, th, &query, options, nui.InputMenuOptions{Label: "Field", OpenOnFocus: nui.Set(false)}).Width(240)
		}, focus},
		{"Select", func(c *ui.Context, th nui.Theme) {
			nui.Select(c, th, &choice, options, nui.SelectOptions{Label: "Field"}).Width(240)
		}, hover("Field")},
		{"SelectMenu", func(c *ui.Context, th nui.Theme) {
			nui.SelectMenu(c, th, &picked, &menu, []nui.ListboxItem{{ID: "a", Label: "Alpha"}}, nui.SelectMenuOptions{Label: "Field"}).Width(240)
		}, hover("Field")},
		{"PinInput", func(c *ui.Context, th nui.Theme) {
			nui.PinInput(c, th, &pin, &pinState, nui.PinInputOptions{Label: "Code", Length: 4})
		}, focus},
	}
	for _, f := range fields {
		for _, l := range motionLevels {
			t.Run(f.name+"/"+l.name, func(t *testing.T) {
				m := newMotionTester(l, 400, 160, f.view)
				checkFade(t, m, !l.off, func() { f.act(t, m) })
			})
		}
	}
}

func TestMenuHighlightCrossFades(t *testing.T) {
	var state nui.MenuState
	m := newMotionTester(motionFull, 400, 300, func(c *ui.Context, th nui.Theme) {
		a := nui.Button(c, th, "Actions", nui.ButtonOptions{})
		nui.DropdownMenu(c, th, a, &state, []nui.MenuItem{{ID: "edit", Label: "Edit"}, {ID: "copy", Label: "Copy"}, {ID: "share", Label: "Share"}}, nui.MenuOptions{})
	})
	click(t, m.Tester, "Actions")
	m.Advance(300 * time.Millisecond) // the entrance and the first highlight
	first, second := bounds(t, m.Tester, "Edit"), bounds(t, m.Tester, "Copy")
	at := func(r ui.Rect) color.RGBA { return m.at(r.X+3, r.Y+r.H/2) }
	lit, idle := at(first), at(second)
	if lit == idle {
		t.Fatalf("the first row is not highlighted: %v", lit)
	}
	m.Key(0, ui.KeyDown)
	m.Advance(75 * time.Millisecond)
	in, out := at(second), at(first)
	if !between(in, idle, lit) || !between(out, lit, idle) {
		t.Fatalf("at 75 ms the rows did not cross-fade: second %v, first %v (idle %v, lit %v)", in, out, idle, lit)
	}
	m.Advance(200 * time.Millisecond)
	if at(second) != lit || at(first) != idle {
		t.Fatalf("the highlight did not arrive: second %v, first %v", at(second), at(first))
	}
}

func TestCardCheckboxFadesToChecked(t *testing.T) {
	for _, l := range motionLevels {
		t.Run(l.name, func(t *testing.T) {
			checked := false
			m := newMotionTester(l, 400, 140, func(c *ui.Context, th nui.Theme) {
				nui.Checkbox(c, th, &checked, nui.CheckboxOptions{Label: "Card", Description: "A card choice", Variant: nui.ChoiceCard}).Width(260)
			})
			checkFade(t, m, !l.off, func() {
				checked = true
				m.Frame()
			})
		})
	}
}

// Holding Down in a long list keeps colour fades on the rows that change
// and does the same as with animations off.
func TestHeldKeyInLongSelectMenuMatchesOff(t *testing.T) {
	items := make([]nui.ListboxItem, 500)
	for i := range items {
		items[i] = nui.ListboxItem{ID: fmt.Sprint(i), Label: fmt.Sprintf("Item %03d", i)}
	}
	type outcome struct {
		active          string
		first, last     int
		highlights      int
		builds, settled int
	}
	run := func(l motionLevel) outcome {
		var value []string
		var state nui.SelectMenuState
		highlights := 0
		m := newMotionTester(l, 500, 500, func(c *ui.Context, th nui.Theme) {
			nui.SelectMenu(c, th, &value, &state, items, nui.SelectMenuOptions{Label: "Items", OnHighlight: func(string) { highlights++ }}).Width(300)
		})
		click(t, m.Tester, "Items")
		m.Advance(300 * time.Millisecond)
		// The Tester settles each key with up to 20 builds at a standing
		// clock while anything animates, which a window paces at the
		// display's rate: count the builds while the clock runs.
		held := 0
		for range 30 {
			m.KeyRepeat(0, ui.KeyDown)
			b, _ := m.span(33 * time.Millisecond)
			held += b
		}
		settled, _ := m.span(500 * time.Millisecond)
		first, last := state.Listbox.List.Visible()
		return outcome{state.Listbox.Active, first, last, highlights, held, settled}
	}
	full, off := run(motionFull), run(motionOff)
	if full.active != off.active || full.first != off.first || full.last != off.last || full.highlights != off.highlights {
		t.Fatalf("holding Down differs from Off: full %+v, off %+v", full, off)
	}
	if full.active != "29" && full.active != "30" {
		t.Fatalf("held Down reached item %q", full.active)
	}
	// About a second of keys at 60 Hz: at most a build a frame while rows
	// fade, and none once they arrive.
	if full.builds > 66 || full.settled > 12 {
		t.Fatalf("held Down built %d times, and %d more after the keys (Off: %d)", full.builds, full.settled, off.builds)
	}
}

func TestCarouselDotFadesToCurrent(t *testing.T) {
	items := []nui.CarouselItem{
		{ID: "a", Label: "Slide A", Build: func(c *ui.Context, _ nui.Theme) { ui.Text(c, "A") }},
		{ID: "b", Label: "Slide B", Build: func(c *ui.Context, _ nui.Theme) { ui.Text(c, "B") }},
		{ID: "c", Label: "Slide C", Build: func(c *ui.Context, _ nui.Theme) { ui.Text(c, "C") }},
	}
	for _, l := range motionLevels {
		t.Run(l.name, func(t *testing.T) {
			var s nui.CarouselState
			m := newMotionTester(l, 500, 300, func(c *ui.Context, th nui.Theme) {
				nui.Carousel(c, th, &s, items, nui.CarouselOptions{Label: "Slides", Dots: true, Height: 120}).Width(400)
			})
			one, two := bounds(t, m.Tester, "Go to slide 1"), bounds(t, m.Tester, "Go to slide 2")
			at := func(r ui.Rect) color.RGBA { return m.at(r.X+r.W/2, r.Y+r.H/2) }
			current, idle := at(one), at(two)
			if current == idle {
				t.Fatalf("the current dot looks idle: %v", current)
			}
			click(t, m.Tester, "Go to slide 2")
			m.Move(490, 290)
			if !l.off {
				m.Advance(75 * time.Millisecond)
				if in, out := at(two), at(one); !between(in, idle, current) || !between(out, current, idle) {
					t.Fatalf("at 75 ms the dots did not fade: second %v, first %v (idle %v, current %v)", in, out, idle, current)
				}
				m.Advance(300 * time.Millisecond)
			}
			if at(two) != current || at(one) != idle || s.Index != 1 {
				t.Fatalf("the dots did not arrive: second %v, first %v, index %d", at(two), at(one), s.Index)
			}
		})
	}
}

// A file dragged over the drop area takes the hover background on the same
// element and transition as the pointer does. The Tester cannot hold a file
// drag over an element (DropFiles drags and drops at once), so the pointer
// drives the zone here.
func TestFileUploadZoneFadesItsBackground(t *testing.T) {
	for _, l := range motionLevels {
		t.Run(l.name, func(t *testing.T) {
			var s nui.FileUploadState
			m := newMotionTester(l, 500, 300, func(c *ui.Context, th nui.Theme) {
				nui.FileUpload(c, th, &s, nui.FileUploadOptions{Label: "Attachments", Description: "Drop files here"}).Width(400)
			})
			r := bounds(t, m.Tester, "Drop files here")
			checkFade(t, m, !l.off, func() { m.Move(r.X+r.W/2, r.Y+r.H/2) })
		})
	}
}

func TestClickableBannerFadesOnHover(t *testing.T) {
	for _, l := range motionLevels {
		t.Run(l.name, func(t *testing.T) {
			m := newMotionTester(l, 500, 200, func(c *ui.Context, th nui.Theme) {
				nui.Banner(c, th, nil, nui.BannerOptions{Title: "Update ready", OnSelect: func() {}})
			})
			r := bounds(t, m.Tester, "Update ready")
			checkFade(t, m, !l.off, func() { m.Move(r.X+r.W/2, r.Y+r.H/2) })
		})
	}
}

func TestSidebarRailFadesOnHover(t *testing.T) {
	for _, l := range motionLevels {
		t.Run(l.name, func(t *testing.T) {
			collapsed := false
			var state nui.SidebarState
			m := newMotionTester(l, 900, 400, func(c *ui.Context, th nui.Theme) {
				ui.Row(c).Fill().Children(func() {
					nui.Sidebar(c.Key("sidebar"), th, &collapsed, nui.SidebarOptions{Title: "Files", State: &state, Resizable: true, ResizeLabel: "Rail"}, func(c *ui.Context, th nui.Theme, compact bool) {
						ui.Text(c, "Entry")
					})
				})
			})
			r := bounds(t, m.Tester, "Rail")
			checkFade(t, m, !l.off, func() { m.Move(r.X+r.W/2, r.Y+r.H/2) })
		})
	}
}

func TestLinkHoverShowsAtOnce(t *testing.T) {
	m := newMotionTester(motionFull, 300, 100, func(c *ui.Context, th nui.Theme) {
		nui.Link(c, th, "Documentation", nui.LinkOptions{})
	})
	r := bounds(t, m.Tester, "Documentation")
	m.Advance(300 * time.Millisecond)
	before := clone(m)
	m.Move(r.X+r.W/2, r.Y+r.H/2)
	first := clone(m)
	m.idle(t, time.Second, "a hovered Link")
	if changed, share := arrivedShare(before, first, clone(m)); changed == 0 || share != 1 {
		t.Fatalf("the Link's text colour did not change in the next frame: %d changed, %.0f%% there", changed, share*100)
	}
}

// Hover changes that change only text or icon colours show at once at every
// level, as MyGo's transitions do not reach text colour.
func TestTextOnlyHoverShowsAtOnce(t *testing.T) {
	tab := "one"
	nav := "home"
	var navState nui.NavigationState
	cases := []struct {
		name, target string
		view         func(c *ui.Context, th nui.Theme)
	}{
		{"Breadcrumb", "Projects", func(c *ui.Context, th nui.Theme) {
			nui.Breadcrumb(c, th, []nui.BreadcrumbItem{{ID: "home", Label: "Home", OnSelect: func() {}}, {ID: "projects", Label: "Projects", OnSelect: func() {}}, {ID: "here", Label: "Here"}}, nui.BreadcrumbOptions{})
		}},
		{"Tabs", "Two", func(c *ui.Context, th nui.Theme) {
			nui.Tabs(c, th, &tab, []nui.TabItem{{ID: "one", Label: "One"}, {ID: "two", Label: "Two"}}, nui.TabsOptions{Label: "Sections", Content: nui.Set(false)}).Width(300)
		}},
		{"NavigationMenu link", "Settings", func(c *ui.Context, th nui.Theme) {
			nui.NavigationMenu(c, th, &nav, &navState, []nui.NavigationItem{{ID: "home", Label: "Home"}, {ID: "settings", Label: "Settings"}}, nui.NavigationOptions{Variant: nui.LinkVariant})
		}},
		{"User", "Ada Lovelace", func(c *ui.Context, th nui.Theme) {
			nui.User(c, th, nui.UserOptions{Name: "Ada Lovelace", Description: "Analyst", OnActivate: func() {}})
		}},
	}
	for _, tc := range cases {
		for _, l := range motionLevels {
			t.Run(tc.name+"/"+l.name, func(t *testing.T) {
				m := newMotionTester(l, 500, 160, tc.view)
				r := bounds(t, m.Tester, tc.target)
				checkFade(t, m, false, func() { m.Move(r.X+r.W/2, r.Y+r.H/2) })
			})
		}
	}
}

// Items highlight, hover or select with a colour fade at Full and Reduced,
// and at once at Off.
func TestItemSurfacesFadeTheirColors(t *testing.T) {
	type item struct {
		name    string
		view    func(c *ui.Context, th nui.Theme)
		prepare func(t *testing.T, m *motionTester)
		act     func(t *testing.T, m *motionTester)
	}
	hover := func(label string) func(t *testing.T, m *motionTester) {
		return func(t *testing.T, m *motionTester) {
			r := bounds(t, m.Tester, label)
			m.Move(r.X+r.W/2, r.Y+r.H/2)
		}
	}
	press := func(label string) func(t *testing.T, m *motionTester) {
		return func(t *testing.T, m *motionTester) { click(t, m.Tester, label) }
	}
	none := func(*testing.T, *motionTester) {}
	options := []nui.Option[string]{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}, {Value: "c", Label: "Gamma"}}
	listItems := []nui.ListboxItem{{ID: "a", Label: "Alpha"}, {ID: "b", Label: "Beta"}, {ID: "c", Label: "Gamma"}}
	actions := []nui.CommandAction{{ID: "a", Label: "Alpha"}, {ID: "b", Label: "Beta"}, {ID: "c", Label: "Gamma"}}
	month := nui.Date{Year: 2024, Month: 2, Day: 1}
	items := func() []item {
		var listSel, menuSel, groupSel []string
		var listState nui.ListboxState
		var menuState nui.SelectMenuState
		var search nui.SearchState
		var palette nui.CommandPaletteState
		var tree nui.TreeState
		var nav nui.NavigationState
		var context nui.MenuState
		var calendar nui.CalendarState
		var date nui.DateSelection
		choice, radio, current := "a", "a", "home"
		selectOpen, inputOpen := true, true
		palette.SetIndex(nui.NewCommandIndex(actions))
		palette.Open = true
		tree.SetNodes([]nui.TreeNode{{ID: "a", Label: "Alpha"}, {ID: "b", Label: "Beta"}})
		calendar.Month, calendar.Active = month, month
		return []item{
			{"Button", func(c *ui.Context, th nui.Theme) { nui.Button(c, th, "Save", nui.ButtonOptions{}) }, none, hover("Save")},
			{"Listbox", func(c *ui.Context, th nui.Theme) {
				nui.Listbox(c, th, &listSel, &listState, listItems, nui.ListboxOptions{Label: "List"}).Width(240)
			}, none, hover("Beta")},
			{"Select rows", func(c *ui.Context, th nui.Theme) {
				nui.Select(c, th, &choice, options, nui.SelectOptions{Label: "Choice", Open: &selectOpen}).Width(240)
			}, none, hover("Gamma")},
			{"InputMenu rows", func(c *ui.Context, th nui.Theme) {
				nui.InputMenu(c, th, &search, options, nui.InputMenuOptions{Label: "Search", Open: &inputOpen}).Width(240)
			}, none, hover("Gamma")},
			{"SelectMenu rows", func(c *ui.Context, th nui.Theme) {
				nui.SelectMenu(c, th, &menuSel, &menuState, listItems, nui.SelectMenuOptions{Label: "Menu"}).Width(240)
			}, press("Menu"), hover("Gamma")},
			{"CommandPalette rows", func(c *ui.Context, th nui.Theme) {
				nui.CommandPalette(c, th, &palette, nui.CommandPaletteOptions{SearchInput: nui.Set(false)})
			}, func(t *testing.T, m *motionTester) {
				awaitAdvancing(t, m.Tester, func() bool { return !palette.Searching() && len(palette.Results.Items) > 0 })
			}, hover("Gamma")},
			{"ContextMenu rows", func(c *ui.Context, th nui.Theme) {
				a := ui.Box(c).Size(200, 60).Label("Area")
				nui.ContextMenu(c, th, a, &context, []nui.MenuItem{{ID: "a", Label: "Alpha"}, {ID: "b", Label: "Beta"}}, nui.MenuOptions{})
			}, func(t *testing.T, m *motionTester) {
				if err := m.RightClick("Area"); err != nil {
					t.Fatal(err)
				}
			}, func(t *testing.T, m *motionTester) { m.Key(0, ui.KeyDown) }},
			{"NavigationMenu pill", func(c *ui.Context, th nui.Theme) {
				nui.NavigationMenu(c, th, &current, &nav, []nui.NavigationItem{{ID: "home", Label: "Home"}, {ID: "settings", Label: "Settings"}}, nui.NavigationOptions{})
			}, none, hover("Settings")},
			{"Tree rows", func(c *ui.Context, th nui.Theme) {
				nui.Tree(c, th, &tree, nui.TreeOptions{Label: "Files"}).Width(240).Height(120)
			}, none, hover("Beta")},
			{"Calendar cells", func(c *ui.Context, th nui.Theme) {
				nui.Calendar(c, th, &date, &calendar, nui.CalendarOptions{Label: "Calendar", Today: nui.Set(month)})
			}, none, hover("Calendar 2024-02-14")},
			{"RadioGroup card", func(c *ui.Context, th nui.Theme) {
				nui.RadioGroup(c, th, &radio, options, nui.RadioGroupOptions{Label: "Plan", Variant: nui.ChoiceCard}).Width(240)
			}, none, press("Beta")},
			{"CheckboxGroup card", func(c *ui.Context, th nui.Theme) {
				nui.CheckboxGroup(c, th, &groupSel, options, nui.CheckboxGroupOptions{Label: "Channels", Variant: nui.ChoiceCard}).Width(240)
			}, none, press("Beta")},
		}
	}
	for i := range len(items()) {
		for _, l := range motionLevels {
			it := items()[i]
			t.Run(it.name+"/"+l.name, func(t *testing.T) {
				m := newMotionTester(l, 700, 560, it.view)
				it.prepare(t, m)
				m.Move(695, 555)
				checkFade(t, m, !l.off, func() { it.act(t, m) })
				if l.off && it.name == "Calendar cells" {
					m.idle(t, time.Second, "a hovered Calendar cell at Off")
				}
			})
		}
	}
}
