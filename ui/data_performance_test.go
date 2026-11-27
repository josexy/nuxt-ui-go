package nui_test

import (
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func TestListboxSubmittedEdits(t *testing.T) {
	items := []nui.ListboxItem{{ID: "a", Label: "Alpha", Group: "First"}, {ID: "b", Label: "Beta", Group: "Second"}}
	selected := []string{"b"}
	var state nui.ListboxState
	shown := true
	listSelection := chosenSelection(selected)
	state.SetItems(items)
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		if shown {
			nui.Listbox(c.Key("choices"), th, listSelection, &state, nui.ListboxOptions{Label: "Choices"}).Width(320)
		}
	}, 400, 300)
	if state.Active != "b" {
		t.Fatal("initial highlight did not follow the selected item")
	}
	click(t, tt, "Alpha")
	items[0].Label, items[0].Group, items[0].Disabled = "Renamed", "New group", true
	state.SetItems(items)
	listSelection.Set([]string{"b"})

	tt.Frame()
	if tt.HasText("Alpha") || tt.HasText("First") || !tt.HasText("Renamed") || !tt.HasText("New group") || state.Active != "b" {
		t.Fatalf("stale item cache: active=%q texts=%v", state.Active, tt.Texts())
	}
	click(t, tt, "Beta")
	if listSelection.Len() != 0 {
		t.Fatal("external selection was not reflected in the next activation")
	}
	shown = false
	tt.Frame()
	tt.Frame()
	items[0], items[1] = items[1], items[0]
	state.SetItems(items)
	shown = true
	tt.Frame()
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyHome)
	tt.Key(0, ui.KeyEnter)
	if !slices.Equal(listSelection.Values(), []string{"b"}) {
		t.Fatalf("reordered/remounted IDs: %v", listSelection.Values())
	}
	items = items[:1]
	state.SetItems(items)
	tt.Frame()
	if tt.HasText("Renamed") {
		t.Fatal("removed item remained cached")
	}
}

func TestListboxLargeHorizontalViewport(t *testing.T) {
	for _, rtl := range []bool{false, true} {
		t.Run(fmt.Sprint(rtl), func(t *testing.T) {
			items := make([]nui.ListboxItem, 10000)
			for i := range items {
				items[i] = nui.ListboxItem{ID: fmt.Sprint(i), Label: fmt.Sprintf("Choice %05d", i), Group: fmt.Sprintf("Group %d", i/100)}
			}
			// Business IDs must not collide with virtualization spacers.
			items[0].ID, items[1].ID = "before", "after"
			var state nui.ListboxState
			var selected []string
			builds := 0
			listSelection := chosenSelection(selected)
			state.SetItems(items)
			tt := ui.NewTester(func(c *ui.Context) {
				builds = 0
				nui.Listbox(c, nui.App(c, nui.AppConfig{}), listSelection, &state, nui.ListboxOptions{
					Label: "Choices", Orientation: nui.Set(nui.Horizontal), ItemWidth: 120,
					Style: func(slot nui.Slot, _ ui.Element) {
						if slot == nui.Item {
							builds++
						}
					},
				}).Width(400)
			}, 500, 200)
			if rtl {
				tt.SetDirection(ui.RightToLeft)
			}
			tt.Key(0, ui.KeyTab)
			tt.Key(0, ui.KeyEnd)
			tt.Key(0, ui.KeyEnter)
			tt.Frame()
			if builds > 10 || !slices.Equal(listSelection.Values(), []string{"9999"}) {
				t.Fatalf("horizontal virtualization: builds=%d selected=%v", builds, listSelection.Values())
			}
			r := bounds(t, tt, "Choice 09999")
			if r.X < 0 || r.X+r.W > 501 {
				t.Fatalf("keyboard target outside viewport: %+v", r)
			}
			parityScreenshot(t, tt, fmt.Sprintf("virtual-horizontal-rtl-%v", rtl))
			tt.Key(0, ui.KeyHome)
			tt.Frame()
			list := bounds(t, tt, "Choices")
			delta := float32(2400)
			if rtl {
				delta = -delta
			}
			tt.Scroll(list.X+list.W/2, list.Y+list.H/2, delta, 0)
			tt.Frame()
			if builds > 10 || tt.HasText("Choice 00000") {
				t.Fatalf("wheel did not replace the visible rows: builds=%d texts=%v", builds, tt.Texts())
			}
			items = items[:3]
			state.SetItems(items)
			tt.Frame()
			if !tt.HasText("Choice 00000") {
				t.Fatal("shrinking the data left the viewport empty")
			}
		})
	}
}

func TestListboxSteadyFrameAllocations(t *testing.T) {
	items := make([]nui.ListboxItem, 100000)
	selected := make([]string, len(items))
	for i := range items {
		selected[i] = fmt.Sprint(i)
		items[i] = nui.ListboxItem{ID: selected[i], Label: "Choice"}
	}
	var state nui.ListboxState
	listSelection := chosenSelection(selected)
	state.SetItems(items)
	tt := ui.NewTester(func(c *ui.Context) {
		nui.Listbox(c, nui.App(c, nui.AppConfig{}), listSelection, &state, nui.ListboxOptions{Height: 320}).Width(640)
	}, 700, 400)
	tt.Move(690, 390)
	tt.Frame()
	if allocations := testing.AllocsPerRun(5, tt.Frame); allocations > 256 {
		t.Fatalf("unchanged data allocated per item: %.0f allocations/frame", allocations)
	}
}

func TestHorizontalListboxResizeAtEnd(t *testing.T) {
	items := make([]nui.ListboxItem, 100)
	for i := range items {
		items[i] = nui.ListboxItem{ID: fmt.Sprint(i), Label: fmt.Sprintf("Item %02d", i)}
	}
	var state nui.ListboxState
	var selected []string
	width := float32(240)
	listSelection := chosenSelection(selected)
	state.SetItems(items)
	tt := ui.NewTester(func(c *ui.Context) {
		nui.Listbox(c, nui.App(c, nui.AppConfig{}), listSelection, &state, nui.ListboxOptions{Label: "Choices", Orientation: nui.Set(nui.Horizontal), ItemWidth: 120}).Width(width)
	}, 1200, 200)
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyEnd)
	width = 1080
	tt.Frame()
	if !tt.HasText("Item 91") || !tt.HasText("Item 99") {
		t.Fatalf("resized end viewport has gaps: %v", tt.Texts())
	}
}

func TestSelectMenuCachesTrackLabelsAndFilterFields(t *testing.T) {
	items := []nui.ListboxItem{{ID: "a", Label: "Alpha", Description: "One"}, {ID: "b", Label: "Beta", Description: "Two"}}
	values := []string{"a"}
	state := nui.SelectMenuState{}
	fields := []string{"label"}
	listSelection := chosenSelection(values)
	state.SetItems(items)
	tt := ui.NewTester(func(c *ui.Context) {
		nui.SelectMenu(c, nui.App(c, nui.AppConfig{}), listSelection, &state, nui.SelectMenuOptions{Label: "Menu", FilterFields: fields}).Width(300)
	}, 400, 350)
	items[0].Label = "Renamed"
	state.SetItems(items)
	tt.Frame()
	if !tt.HasText("Renamed") {
		t.Fatal("cached label did not update")
	}
	items = items[1:]
	state.SetItems(items)
	tt.Frame()
	if !tt.HasText("Renamed") {
		t.Fatal("missing async option lost its last known label")
	}
	listSelection.Set([]string{"b"})
	tt.Frame()
	if !tt.HasText("Beta") || tt.HasText("Renamed") {
		t.Fatal("submitted selection change was ignored")
	}
	click(t, tt, "Menu")
	tt.Type("Two")
	if !tt.HasText("No matching data") {
		t.Fatalf("label-only filter: %v", tt.Texts())
	}
	fields[0] = "description"
	tt.Frame()
	if tt.HasText("No matching data") {
		t.Fatal("filter field change did not invalidate the cached results")
	}
	items[0].Description = "Changed"
	state.SetItems(items)
	tt.Frame()
	if !tt.HasText("No matching data") {
		t.Fatal("in-place item edit did not invalidate filtering")
	}
}

func TestSelectAndInputMenuVirtualizedNavigation(t *testing.T) {
	for _, kind := range []string{"select", "input"} {
		t.Run(kind, func(t *testing.T) {
			options := make([]nui.Option[string], 10000)
			builds, changes := 0, 0
			for i := range options {
				label := fmt.Sprintf("Option %05d", i)
				options[i] = nui.Option[string]{Value: fmt.Sprint(i), Label: label, Build: func(c *ui.Context, _ nui.Theme) {
					builds++
					ui.Text(c, label)
				}}
			}
			options[9998].Disabled = true
			selected := "0"
			var search nui.SearchState
			selectState := submittedSelect(options)
			search.SetOptions(options)
			tt := ui.NewTester(func(c *ui.Context) {
				builds = 0
				th := nui.App(c, nui.AppConfig{})
				if kind == "select" {
					nui.Select(c, th, &selected, selectState, nui.SelectOptions{Label: "Choices", OnChange: func() { changes++ }}).Width(320)
				} else {
					nui.InputMenu(c, th, &search, nui.InputMenuOptions{Label: "Choices", OnSelect: func(string) { changes++ }}).Width(320)
				}
			}, 450, 400)
			tt.Key(0, ui.KeyTab)
			if kind == "select" {
				tt.Key(0, ui.KeyDown)
				tt.Key(0, ui.KeyEnd)
			} else {
				tt.Key(0, ui.KeyUp)
			}
			tt.Frame()
			if builds > 24 || !tt.HasText("Option 09999") {
				t.Fatalf("end navigation: builds=%d texts=%v", builds, tt.Texts())
			}
			r := bounds(t, tt, "Option 09999")
			if r.W < 100 || r.H < 10 || r.Y < 0 || r.Y+r.H > 400 {
				t.Fatalf("virtualized option is not visibly laid out: %+v", r)
			}
			parityScreenshot(t, tt, "virtual-"+kind+"-end")
			tt.Key(0, ui.KeyUp)
			tt.Key(0, ui.KeyEnter)
			got := selected
			if kind == "input" {
				got = search.Selected
			}
			if got != "9997" || changes != 1 || !tt.Focused("Choices") {
				t.Fatalf("virtualized choice: got=%q changes=%d focus=%v", got, changes, tt.Focused("Choices"))
			}
		})
	}
}

func TestTreeLazyCompletionAfterCollapseBeforeFrame(t *testing.T) {
	var state nui.TreeState
	state.SetNodes([]nui.TreeNode{{ID: "root", Label: "Root", Lazy: true}})
	state.Outline.Open.Add("root")
	var ticket nui.TreeTicket
	tt := ui.NewTester(func(c *ui.Context) {
		nui.Tree(c, nui.App(c, nui.AppConfig{}), &state, nui.TreeOptions{OnLoad: func(_ string, next nui.TreeTicket) { ticket = next }})
	}, 400, 350)
	state.Outline.Open.Remove("root")
	if state.CompleteLoad(ticket, []nui.TreeNode{{ID: "stale", Label: "Stale"}}, nil) {
		t.Fatal("collapsed request accepted before the next build")
	}
	tt.Frame()
	state.Outline.Open.Add("root")
	tt.Frame()
	if !state.CompleteLoad(ticket, []nui.TreeNode{{ID: "fresh", Label: "Fresh"}}, nil) {
		t.Fatal("new expansion request was not tracked")
	}
	tt.Frame()
	if !tt.HasText("Fresh") || tt.HasText("Stale") {
		t.Fatal("lazy load visibility")
	}
}

func BenchmarkDataComponents(b *testing.B) {
	for _, kind := range []string{"Listbox", "HorizontalListbox", "SelectMenuClosed", "SelectOpen", "InputMenuOpen", "TreeCollapsed", "TreeExpanded", "TreeFlat"} {
		for _, count := range []int{100, 10000, 100000} {
			b.Run(fmt.Sprintf("%s/%d", kind, count), func(b *testing.B) {
				var items []nui.ListboxItem
				var options []nui.Option[string]
				var nodes []nui.TreeNode
				var values []string
				switch kind {
				case "Listbox", "HorizontalListbox", "SelectMenuClosed":
					items = make([]nui.ListboxItem, count)
					if kind == "SelectMenuClosed" {
						values = make([]string, count)

					}
				case "SelectOpen", "InputMenuOpen":
					options = make([]nui.Option[string], count)
				default:
					nodes = make([]nui.TreeNode, count)
				}
				for i := range count {
					id, label := fmt.Sprint(i), fmt.Sprintf("Item %06d", i)
					if items != nil {
						items[i] = nui.ListboxItem{ID: id, Label: label}
					}
					if options != nil {
						options[i] = nui.Option[string]{Value: id, Label: label}
					}
					if nodes != nil {
						nodes[i] = nui.TreeNode{ID: id, Label: label}
					}
					if values != nil {
						values[i] = id
					}
				}
				var list nui.ListboxState
				var menu nui.SelectMenuState
				var search nui.SearchState
				var tree nui.TreeState
				var selected []string
				single := "0"
				builds := 0
				switch kind {
				case "TreeCollapsed", "TreeExpanded":
					tree.SetNodes([]nui.TreeNode{{ID: "root", Label: "Root", Children: nodes}})
					if kind == "TreeExpanded" {
						tree.SetExpanded("root", true)
					}
				case "TreeFlat":
					tree.SetNodes(nodes)
				}
				var selectState nui.SelectState[string]
				if kind == "SelectOpen" {
					selectState.SetOptions(options)
				}
				listSelection1 := chosenSelection(selected)
				if kind == "Listbox" || kind == "HorizontalListbox" {
					list.SetItems(items)
				}
				listSelection2 := chosenSelection(values)
				if kind == "SelectMenuClosed" {
					menu.SetItems(items)
				}
				if kind == "InputMenuOpen" {
					search.SetOptions(options)
				}
				tt := ui.NewTester(func(c *ui.Context) {
					builds++
					th := nui.App(c, nui.AppConfig{DisableAnimations: true})
					switch kind {
					case "Listbox", "HorizontalListbox":
						o := nui.ListboxOptions{Label: "Choices", Height: 320}
						if kind == "HorizontalListbox" {
							o.Orientation = nui.Set(nui.Horizontal)
						}
						nui.Listbox(c, th, listSelection1, &list, o).Width(640)
					case "SelectMenuClosed":
						nui.SelectMenu(c, th, listSelection2, &menu, nui.SelectMenuOptions{Label: "Choices", Multiple: true}).Width(640)
					case "SelectOpen":
						nui.Select(c, th, &single, &selectState, nui.SelectOptions{Label: "Choices"}).Width(640)
					case "InputMenuOpen":
						nui.InputMenu(c, th, &search, nui.InputMenuOptions{Label: "Choices"}).Width(640)
					default:
						nui.Tree(c, th, &tree, nui.TreeOptions{Height: 320}).Width(640)
					}
				}, 700, 400)
				if kind == "SelectOpen" || kind == "InputMenuOpen" {
					tt.Key(0, ui.KeyTab)
					tt.Key(0, ui.KeyDown)
				}
				tt.Move(690, 390)
				tt.Frame()
				before := builds
				tt.Frame()
				if builds-before != 1 {
					b.Fatalf("settled Frame performed %d builds, want 1", builds-before)
				}
				builds = 0
				b.ReportAllocs()
				b.ResetTimer()
				for b.Loop() {
					tt.Frame()
				}
				b.ReportMetric(float64(builds)/float64(b.N), "builds/op")
				if builds != b.N {
					b.Fatalf("%d operations performed %d builds", b.N, builds)
				}
			})
		}
	}
}

// Motion is a separate workload. Frame delivers the controlled-open change at
// a fixed clock; Advance then runs scheduled ticks at their simulated times.
// These are software-rendered Tester operations, not GPU frames.
func BenchmarkDataAnimations(b *testing.B) {
	for _, reduced := range []bool{false, true} {
		b.Run(fmt.Sprintf("SelectPopup/Reduced=%v", reduced), func(b *testing.B) {
			open, builds := false, 0
			value := "a"
			options := []nui.Option[string]{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}}
			selectState := submittedSelect(options)
			tt := ui.NewTesterWith(func(c *ui.Context) {
				builds++
				nui.Select(c.Key("choices"), nui.App(c, nui.AppConfig{}), &value, selectState,
					nui.SelectOptions{Label: "Choices", Open: &open}).Width(320)
			}, 700, 400, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: reduced}})
			tt.Move(690, 390)
			tt.Advance(time.Second)
			inputBuilds, tickBuilds := 0, 0
			b.ReportAllocs()
			for b.Loop() {
				for _, next := range []bool{true, false} {
					open = next
					before := builds
					tt.Frame()
					inputBuilds += builds - before
					before = builds
					tt.Advance(300 * time.Millisecond)
					tickBuilds += builds - before
				}
			}
			b.ReportMetric(float64(inputBuilds)/float64(b.N), "input-builds/op")
			b.ReportMetric(float64(tickBuilds)/float64(b.N), "tick-builds/op")
			b.ReportMetric(float64(600*time.Millisecond), "simulated-ns/op")
		})
	}
}
