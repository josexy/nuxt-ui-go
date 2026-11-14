package nui_test

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func treeTester(s *nui.TreeState, o *nui.TreeOptions) *ui.Tester {
	return ui.NewTester(func(c *ui.Context) {
		nui.Tree(c, nui.App(c, nui.AppConfig{}), s, *o).Width(360)
	}, 400, 400)
}

func treeChosen(t *testing.T, s *nui.TreeState, want ...string) {
	t.Helper()
	got := slices.Sorted(s.Selection.All())
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Fatalf("selection %v, want %v", got, want)
	}
}

func TestTreeSelectionAndExpansionEvents(t *testing.T) {
	var s nui.TreeState
	s.SetNodes([]nui.TreeNode{{ID: "root", Label: "Root", Children: []nui.TreeNode{{ID: "a", Label: "Alpha"}, {ID: "b", Label: "Beta"}}}})
	selects, toggles, changes, expansions := 0, 0, 0, 0
	var expanded []string
	o := nui.TreeOptions{Label: "Files", OnSelect: func(e *nui.TreeEvent) { selects++ }, OnToggle: func(e *nui.TreeEvent) { toggles++ }, OnChange: func() { changes++ }, OnExpandedChange: func(ids []string) { expansions++; expanded = ids }}
	tt := treeTester(&s, &o)
	click(t, tt, "Root")
	if !tt.HasText("Alpha") || !s.Outline.Open.Has("root") {
		t.Fatal("row click did not expand")
	}
	treeChosen(t, &s, "root")
	tt.Key(0, ui.KeyDown)
	if s.Active != "a" {
		t.Fatalf("highlight %q", s.Active)
	}
	treeChosen(t, &s, "root")
	tt.Key(0, ui.KeySpace)
	treeChosen(t, &s, "a")
	tt.Key(0, ui.KeySpace)
	treeChosen(t, &s)
	if selects != 3 || toggles != 1 || changes != 3 {
		t.Fatalf("callbacks select=%d toggle=%d change=%d", selects, toggles, changes)
	}
	o.SelectionBehavior = nui.SelectionReplace
	tt.Frame()
	tt.Key(0, ui.KeyEnter)
	tt.Key(0, ui.KeyEnter)
	if selects != 5 || changes != 4 {
		t.Fatal("unchanged activation should emit select once, without change")
	}
	tt.Key(0, ui.KeyLeft)
	if s.Active != "root" {
		t.Fatal("left must highlight parent without selecting")
	}
	tt.Key(0, ui.KeyLeft)
	if tt.HasText("Alpha") {
		t.Fatal("left must collapse parent")
	}
	treeChosen(t, &s, "a")
	if expansions != 2 || len(expanded) != 0 {
		t.Fatalf("expansion notification: %d %v", expansions, expanded)
	}
	s.SetExpanded("root", true)
	tt.Frame()
	if expansions != 2 {
		t.Fatal("external expansion emitted a user event")
	}
}

func TestTreePreventDefaultAndNodeHandlers(t *testing.T) {
	var s nui.TreeState
	nodeSelects, globalSelects := 0, 0
	s.SetNodes([]nui.TreeNode{{ID: "p", Label: "Parent", Children: []nui.TreeNode{{ID: "c", Label: "Child"}}, OnSelect: func(e *nui.TreeEvent) { nodeSelects++; e.PreventDefault() }}})
	o := nui.TreeOptions{OnSelect: func(*nui.TreeEvent) { globalSelects++ }, OnToggle: func(e *nui.TreeEvent) { e.PreventDefault() }}
	tt := treeTester(&s, &o)
	click(t, tt, "Parent")
	if nodeSelects != 1 || globalSelects != 0 || s.Outline.Open.Len() != 0 || s.Selection.Len() != 0 {
		t.Fatal("event prevention / node precedence")
	}
	o.OnToggle = nil
	tt.Frame()
	click(t, tt, "Parent")
	if !tt.HasText("Child") || s.Selection.Len() != 0 {
		t.Fatal("preventing selection must not prevent toggle")
	}
	click(t, tt, "Child")
	treeChosen(t, &s, "c")
	// Replacing the indexed model during selection must not let the old click
	// toggle a new node that happens to reuse the same business ID.
	s.SetExpanded("p", false)
	s.SetNodes([]nui.TreeNode{{ID: "p", Label: "Original", Children: []nui.TreeNode{{ID: "old", Label: "Old child"}}}})
	o.OnSelect = func(*nui.TreeEvent) {
		s.SetNodes([]nui.TreeNode{{ID: "p", Label: "Replacement", Children: []nui.TreeNode{{ID: "new", Label: "New child"}}}})
	}
	tt.Frame()
	click(t, tt, "Original")
	if s.Outline.Open.Has("p") {
		t.Fatal("old selection action expanded a replacement node")
	}
}

func TestTreeMultipleRangeDisabledAndLoop(t *testing.T) {
	var s nui.TreeState
	s.SetNodes([]nui.TreeNode{{ID: "a", Label: "Alpha"}, {ID: "x", Label: "Disabled", Disabled: true}, {ID: "b", Label: "Beta"}, {ID: "c", Label: "Charlie"}})
	o := nui.TreeOptions{Label: "Files", Multiple: true, Loop: true}
	changes := 0
	o.OnChange = func() { changes++ }
	tt := treeTester(&s, &o)
	click(t, tt, "Alpha")
	click(t, tt, "Beta")
	treeChosen(t, &s, "a", "b")
	click(t, tt, "Disabled")
	treeChosen(t, &s, "a", "b")
	o.SelectionBehavior = nui.SelectionReplace
	tt.Frame()
	click(t, tt, "Alpha")
	if err := tt.ClickWith(ui.Shift, "Charlie"); err != nil {
		t.Fatal(err)
	}
	treeChosen(t, &s, "a", "b", "c")
	before := changes
	if err := tt.ClickWith(ui.Shift, "Charlie"); err != nil {
		t.Fatal(err)
	}
	if changes != before {
		t.Fatal("unchanged range emitted change")
	}
	if err := tt.ClickWith(ui.Ctrl, "Beta"); err != nil {
		t.Fatal(err)
	}
	treeChosen(t, &s, "a", "c")
	tt.Key(0, ui.KeyHome)
	tt.Key(0, ui.KeyUp)
	if s.Active != "c" {
		t.Fatalf("loop %q", s.Active)
	}
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyDown)
	if s.Active != "b" {
		t.Fatal("disabled row not skipped")
	}
	tt.Type("char")
	if s.Active != "c" {
		t.Fatalf("typeahead %q", s.Active)
	}
	tt.Key(ui.Ctrl, ui.KeyA)
	treeChosen(t, &s, "a", "b", "c")
	o.Disabled = true
	tt.Frame()
	click(t, tt, "Alpha")
	tt.Key(0, ui.KeySpace)
	treeChosen(t, &s, "a", "b", "c")
	// Repeated characters cycle; extending a prefix keeps the current match.
	o.Disabled, o.Multiple = false, false
	s.SetNodes([]nui.TreeNode{{ID: "a", Label: "Alpha"}, {ID: "b1", Label: "Bran"}, {ID: "b2", Label: "Bravo"}})
	tt.Frame()
	click(t, tt, "Alpha")
	tt.Advance(2 * time.Second)
	tt.Type("b")
	tt.Type("b")
	if s.Active != "b2" {
		t.Fatal("repeated typeahead character did not cycle")
	}
	tt.Advance(2 * time.Second)
	tt.Type("b")
	tt.Type("r")
	if s.Active != "b1" {
		t.Fatal("extended prefix skipped its current match")
	}
}

func TestTreePropagationAndIndeterminate(t *testing.T) {
	var s nui.TreeState
	s.SetNodes([]nui.TreeNode{{ID: "p", Label: "Parent", DefaultExpanded: true, Children: []nui.TreeNode{{ID: "a", Label: "Alpha"}, {ID: "b", Label: "Beta"}, {ID: "x", Label: "Disabled", Disabled: true}}}})
	partial := false
	changes := 0
	o := nui.TreeOptions{Multiple: true, PropagateSelect: true, BubbleSelect: true, ExpandOnClick: nui.Set(false), OnChange: func() { changes++ }, Build: func(c *ui.Context, t nui.Theme, item nui.TreeItemState) {
		ui.Text(c, item.Label)
		if item.ID == "p" {
			partial = item.Indeterminate
		}
	}}
	tt := treeTester(&s, &o)
	click(t, tt, "Alpha")
	treeChosen(t, &s, "a")
	if !partial {
		t.Fatal("partial parent missing")
	}
	click(t, tt, "Beta")
	treeChosen(t, &s, "p", "a", "b")
	if partial {
		t.Fatal("fully selected parent is partial")
	}
	click(t, tt, "Parent")
	treeChosen(t, &s)
	click(t, tt, "Parent")
	treeChosen(t, &s, "p", "a", "b")
	if changes != 4 {
		t.Fatalf("cascade changes %d", changes)
	}
	// Application-controlled edits also refresh mixed state.
	s.Selection.Remove("a")
	tt.Frame()
	if !partial {
		t.Fatal("external selection did not update mixed state")
	}
	o.PropagateSelect, o.BubbleSelect = false, false
	tt.Frame()
	if partial {
		t.Fatal("partial state exposed when both linkage modes are disabled")
	}
	o.BubbleSelect = true
	s.Selection.Clear()
	s.Selection.Add("p")
	tt.Frame()
	if partial {
		t.Fatal("bubble-only parent with no selected descendants is partial")
	}
	o.BubbleSelect, o.PropagateSelect = false, true
	tt.Frame()
	if !partial {
		t.Fatal("propagate-only selected parent lost partial state")
	}
	o.Multiple = false
	tt.Frame()
	if partial {
		t.Fatal("single selection exposed a linked partial state")
	}
}

func TestTreeStableIdentityAndDefaults(t *testing.T) {
	var s nui.TreeState
	nodes := []nui.TreeNode{{ID: "p", Label: "Parent", DefaultExpanded: true, Children: []nui.TreeNode{{ID: "a", Label: "Same"}, {ID: "b", Label: "Same"}}}}
	s.SetNodes(nodes)
	o := nui.TreeOptions{Label: "Files"}
	tt := treeTester(&s, &o)
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyEnd)
	tt.Key(0, ui.KeySpace)
	treeChosen(t, &s, "b")
	nodes[0].Children[1].Label = "Renamed"
	nodes[0].Children = slices.Insert(nodes[0].Children, 0, nui.TreeNode{ID: "z", Label: "Inserted"})
	s.SetNodes(nodes)
	tt.Frame()
	if s.Active != "b" || !tt.Focused("Files") {
		t.Fatal("reordering lost focus identity")
	}
	treeChosen(t, &s, "b")
	s.SetExpanded("p", false)
	tt.Frame()
	if s.Active != "p" || tt.HasText("Renamed") {
		t.Fatal("collapse did not restore parent highlight")
	}
	s.SetNodes(nodes)
	tt.Frame()
	if s.Outline.Open.Has("p") {
		t.Fatal("defaultExpanded reapplied to an existing ID")
	}
	s.SetNodes([]nui.TreeNode{{ID: "new", Label: "New"}})
	tt.Frame()
	treeChosen(t, &s)
	if s.Active != "new" {
		t.Fatal("removed active row not reconciled")
	}
}

func TestTreeLazyAncestorCollapseRetryAndDispose(t *testing.T) {
	var s nui.TreeState
	s.SetNodes([]nui.TreeNode{{ID: "p", Label: "Parent", DefaultExpanded: true, Children: []nui.TreeNode{{ID: "lazy", Label: "Lazy", Lazy: true, DefaultExpanded: true}}}})
	var ticket nui.TreeTicket
	loads := 0
	o := nui.TreeOptions{OnLoad: func(_ string, v nui.TreeTicket) { ticket = v; loads++ }}
	tt := treeTester(&s, &o)
	old := ticket
	s.SetExpanded("p", false)
	s.SetExpanded("p", true)
	tt.Frame()
	if loads != 2 || s.CompleteLoad(old, nil, nil) {
		t.Fatal("ancestor close/reopen accepted stale ticket")
	}
	if !s.CompleteLoad(ticket, nil, errors.New("offline")) {
		t.Fatal("load error rejected")
	}
	tt.Frame()
	click(t, tt, "Retry Lazy")
	if loads != 3 || s.Error("lazy") != "" {
		t.Fatalf("retry loads=%d error=%s", loads, s.Error("lazy"))
	}
	s.Close()
	if s.CompleteLoad(ticket, nil, nil) {
		t.Fatal("disposed ticket accepted")
	}
	tt.Frame()
	if !s.CompleteLoad(ticket, []nui.TreeNode{{ID: "leaf", Label: "Loaded"}}, nil) {
		t.Fatal("fresh request rejected")
	}
	tt.Frame()
	if !tt.HasText("Loaded") {
		t.Fatal("loaded rows not indexed")
	}
	leading := 0
	o.Style = func(slot nui.Slot, _ ui.Element) {
		if slot == nui.Leading {
			leading++
		}
	}
	s.SetNodes([]nui.TreeNode{{ID: "empty", Label: "Empty branch", Lazy: true, DefaultExpanded: true}})
	tt.Frame()
	if leading == 0 {
		t.Fatal("unresolved branch lost its folder icon")
	}
	if !s.CompleteLoad(ticket, nil, nil) {
		t.Fatal("empty result rejected")
	}
	leading = 0
	tt.Frame()
	if leading != 0 || !s.Outline.Open.Has("empty") {
		t.Fatal("empty branch icons or expansion state")
	}
}

func TestTreeRTLHiddenRowsAndInheritedDisabled(t *testing.T) {
	var s nui.TreeState
	s.SetNodes([]nui.TreeNode{
		{ID: "p", Label: "Parent", Children: []nui.TreeNode{{ID: "d", Label: "Disabled child", Disabled: true}}},
		{ID: "q", Label: "Other", Children: []nui.TreeNode{{ID: "c", Label: "Child"}}},
	})
	disabled := false
	accent := ui.RGB(126, 34, 206)
	tt := ui.NewTester(func(c *ui.Context) {
		theme := nui.App(c, nui.AppConfig{})
		theme.InvertedBackground, theme.HighContrast = accent, true
		ui.Column(c).Direction(ui.RightToLeft).Disabled(disabled).Children(func() {
			nui.FormField(c, theme, nui.FormFieldOptions{Label: "Files"}, func(ft nui.Theme) {
				nui.Tree(c, ft, &s, nui.TreeOptions{Color: nui.Neutral, ExpandOnClick: nui.Set(false)}).Width(360)
			})
		})
	}, 400, 400)
	click(t, tt, "Parent")
	if s.Outline.Open.Has("p") {
		t.Fatal("ExpandOnClick false ignored")
	}
	tt.Key(0, ui.KeyLeft)
	if !s.Outline.Open.Has("p") {
		t.Fatal("RTL expand")
	}
	tt.Key(0, ui.KeyLeft)
	if s.Active != "p" {
		t.Fatal("entering disabled children escaped to a different branch")
	}
	if !tt.Focused("Files") {
		t.Fatal("tree did not inherit its field name")
	}
	r := bounds(t, tt, "Parent")
	if countWhere(tt.Image(), r, func(v color.RGBA) bool { return v == rgba(accent) }) == 0 {
		t.Fatalf("neutral tree focus missing in %v", r)
	}
	// Same-size external changes must invalidate the visible index too.
	s.Outline.Open.Remove("p")
	s.Outline.Open.Add("q")
	tt.Frame()
	if tt.HasText("Disabled child") || !tt.HasText("Child") {
		t.Fatal("external expansion not reflected")
	}
	disabled = true
	tt.Frame()
	click(t, tt, "Other")
	tt.Key(0, ui.KeySpace)
	treeChosen(t, &s, "p")
}

func TestTreeVirtualHighlightSurvivesUnmount(t *testing.T) {
	var s nui.TreeState
	s.SetNodes([]nui.TreeNode{{ID: "a", Label: "Alpha"}, {ID: "b", Label: "Beta"}})
	show := true
	tt := ui.NewTester(func(c *ui.Context) {
		theme := nui.App(c, nui.AppConfig{})
		if show {
			nui.Tree(c.Key("files"), theme, &s, nui.TreeOptions{Label: "Files"}).Width(360)
		}
	}, 400, 400)
	click(t, tt, "Beta")
	show = false
	tt.Frame()
	tt.Frame()
	show = true
	tt.Frame()
	tt.Frame()
	treeChosen(t, &s, "b")
	if s.Active != "b" {
		t.Fatal("unmount lost highlight")
	}
}

func TestTreeSelectAllOverlappingBranchesAndDisabledBoundary(t *testing.T) {
	var s nui.TreeState
	s.SetNodes([]nui.TreeNode{{ID: "root", Label: "Root", DefaultExpanded: true, Children: []nui.TreeNode{
		{ID: "branch", Label: "Branch", DefaultExpanded: true, Children: []nui.TreeNode{{ID: "leaf", Label: "Leaf"}}},
		{ID: "disabled", Label: "Disabled", Disabled: true, DefaultExpanded: true, Children: []nui.TreeNode{{ID: "enabled", Label: "Enabled child"}}},
	}}})
	changes := 0
	o := nui.TreeOptions{Multiple: true, PropagateSelect: true, OnChange: func() { changes++ }}
	tt := treeTester(&s, &o)
	tt.Key(0, ui.KeyTab)
	tt.Key(ui.Ctrl, ui.KeyA)
	treeChosen(t, &s, "root", "branch", "leaf", "enabled")
	tt.Key(ui.Ctrl, ui.KeyA)
	if changes != 1 {
		t.Fatalf("select all callbacks %d", changes)
	}
}

func TestTreeNestedGuidesAndModeSwitch(t *testing.T) {
	for _, rtl := range []bool{false, true} {
		for _, dark := range []bool{false, true} {
			t.Run(fmt.Sprintf("rtl=%v/dark=%v", rtl, dark), func(t *testing.T) {
				var s nui.TreeState
				s.SetNodes([]nui.TreeNode{
					{ID: "root", Label: "Root", DefaultExpanded: true, Children: []nui.TreeNode{
						{ID: "branch", Label: "Branch", DefaultExpanded: true, Children: []nui.TreeNode{{ID: "deep", Label: "Deep file"}}},
						{ID: "sibling", Label: "Sibling"},
					}},
					{ID: "tail", Label: "Tail"},
				})
				marker := ui.RGB(255, 0, 255)
				o := nui.TreeOptions{Label: "Files", UI: nui.Styles{nui.Indicator: {BorderColor: nui.Set(marker)}}, Build: func(c *ui.Context, _ nui.Theme, item nui.TreeItemState) {
					ui.Text(c, "Text: "+item.Label).SingleLine().Grow(1).MinWidth(0)
				}}
				theme := nui.Light()
				if dark {
					theme = nui.Dark()
				}
				theme.TextScale = 1.5
				tt := ui.NewTester(func(c *ui.Context) {
					th := nui.App(c, nui.AppConfig{Mode: nui.LightMode, Light: theme})
					direction := ui.LeftToRight
					if rtl {
						direction = ui.RightToLeft
					}
					ui.Column(c).Direction(direction).Children(func() { nui.Tree(c.Key("files"), th, &s, o).Width(360) })
				}, 400, 400)
				click(t, tt, "Deep file")
				tt.Move(390, 390)
				deep := bounds(t, tt, "Text: Deep file")
				columns := treeGuideColumns(tt.Image(), int(deep.Y+deep.H/2), marker)
				if len(columns) != 2 {
					t.Fatalf("nested leaf guide columns %v", columns)
				}
				sibling := bounds(t, tt, "Text: Sibling")
				if got := treeGuideColumns(tt.Image(), int(sibling.Y+sibling.H/2), marker); len(got) != 1 {
					t.Fatalf("sibling guides %v", got)
				}
				tail := bounds(t, tt, "Text: Tail")
				if got := treeGuideColumns(tt.Image(), int(tail.Y+tail.H/2), marker); len(got) != 0 {
					t.Fatalf("guide leaked into next root: %v", got)
				}
				if !rtl && float32(columns[1]) >= deep.X || rtl && float32(columns[0]) <= deep.X+deep.W {
					t.Fatal("guides not on logical start side")
				}
				tt.ClickAt(float32(columns[0]), deep.Y+deep.H/2)
				treeChosen(t, &s, "deep") // Nested ancestor gutters are outside the row's hit area.
				o.Nested = nui.Set(false)
				tt.Frame()
				flat := bounds(t, tt, "Text: Deep file")
				if deep != flat {
					t.Fatalf("content moved on mode switch: nested=%v flat=%v", deep, flat)
				}
				if got := treeGuideColumns(tt.Image(), int(flat.Y+flat.H/2), marker); len(got) != 0 {
					t.Fatal("flat mode retained guides")
				}
				if got := tt.Image().RGBAAt(columns[0], int(flat.Y+flat.H/2)); got != rgba(theme.Elevated) {
					t.Fatalf("flat selection does not cover gutter: %v", got)
				}
				if !tt.Focused("Files") || s.Active != "deep" || s.Outline.Open.Len() != 2 {
					t.Fatal("mode switch lost tree state or focus")
				}
				treeChosen(t, &s, "deep")
				tt.ClickAt(float32(columns[0]), flat.Y+flat.H/2)
				treeChosen(t, &s) // Flat rows include the same gutter in their hit area.
				o.Nested = nui.Set(true)
				o.UI[nui.Indicator] = nui.PartStyle{BorderColor: nui.Set(marker), BorderWidth: nui.Set(float32(0))}
				tt.Frame()
				if got := treeGuideColumns(tt.Image(), int(deep.Y+deep.H/2), marker); len(got) != 0 {
					t.Fatal("explicit zero guide width ignored")
				}
			})
		}
	}
}

func treeGuideColumns(im *image.RGBA, y int, marker ui.Color) []int {
	var columns []int
	for x := im.Bounds().Min.X; x < im.Bounds().Max.X; x++ {
		if im.RGBAAt(x, y) == rgba(marker) {
			columns = append(columns, x)
		}
	}
	return columns
}

func TestTreeNestedGuidesWithOffscreenAncestors(t *testing.T) {
	children := make([]nui.TreeNode, 10000)
	for i := range children {
		children[i] = nui.TreeNode{ID: fmt.Sprint(i), Label: fmt.Sprintf("File %d", i)}
	}
	var s nui.TreeState
	s.SetNodes([]nui.TreeNode{{ID: "root", Label: "Root", DefaultExpanded: true, Children: children}})
	marker := ui.RGB(255, 0, 255)
	o := nui.TreeOptions{Label: "Files", Height: 200, UI: nui.Styles{nui.Indicator: {BorderColor: nui.Set(marker)}}}
	tt := treeTester(&s, &o)
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyEnd)
	tt.Key(0, ui.KeySpace)
	tt.Move(390, 390)
	r := bounds(t, tt, "File 9999")
	if got := treeGuideColumns(tt.Image(), int(r.Y+r.H/2), marker); len(got) != 1 {
		t.Fatalf("offscreen ancestor guide: %v", got)
	}
	if s.RowBuilds > 24 {
		t.Fatalf("nested built %d rows", s.RowBuilds)
	}
	first, last := s.Outline.List.Visible()
	o.Nested = nui.Set(false)
	tt.Frame()
	nextFirst, nextLast := s.Outline.List.Visible()
	if first != nextFirst || last != nextLast || bounds(t, tt, "File 9999").Y != r.Y {
		t.Fatal("mode switch lost scroll position")
	}
	treeChosen(t, &s, "9999")
	if !tt.Focused("Files") {
		t.Fatal("mode switch lost focus")
	}
}

func TestTreeAdjacentSelectionSurfaceGap(t *testing.T) {
	for _, nested := range []bool{false, true} {
		for _, dark := range []bool{false, true} {
			for _, size := range []nui.Size{nui.XS, nui.SM, nui.MD, nui.LG, nui.XL} {
				t.Run(fmt.Sprintf("nested=%v/dark=%v/size=%d", nested, dark, size), func(t *testing.T) {
					var s nui.TreeState
					s.SetNodes([]nui.TreeNode{{ID: "root", Label: "Root", DefaultExpanded: true, Children: []nui.TreeNode{{ID: "a", Label: "Alpha"}, {ID: "b", Label: "Beta"}, {ID: "c", Label: "Charlie"}}}})
					s.Selection.Add("a")
					s.Selection.Add("b")
					theme := nui.Light()
					if dark {
						theme = nui.Dark()
					}
					marker := ui.RGB(255, 0, 255)
					changes := 0
					tt := ui.NewTester(func(c *ui.Context) {
						th := nui.App(c, nui.AppConfig{Mode: nui.LightMode, Light: theme})
						nui.Tree(c, th, &s, nui.TreeOptions{Label: "Files", Size: size, Multiple: true, Nested: nui.Set(nested), OnChange: func() { changes++ }, UI: nui.Styles{nui.Indicator: {BorderColor: nui.Set(marker)}}}).Width(360)
					}, 400, 400)
					a, b := bounds(t, tt, "Alpha"), bounds(t, tt, "Beta")
					wantHeight := []float32{24, 28, 32, 36, 40}[int(size-nui.XS)]
					if a.H != wantHeight || b.Y != a.Y+a.H {
						t.Fatalf("gap changed row layout: a=%v b=%v", a, b)
					}
					x, y := int(a.X+a.W-24), int(b.Y)
					for _, offset := range []int{-1, 0} {
						if got := tt.Image().RGBAAt(x, y+offset); got != rgba(theme.Background) {
							t.Fatalf("adjacent surfaces touch at offset %d: %v", offset, got)
						}
					}
					for _, offset := range []int{-2, 1} {
						if got := tt.Image().RGBAAt(x, y+offset); got != rgba(theme.Elevated) {
							t.Fatalf("selection missing at offset %d: %v", offset, got)
						}
					}
					if nested {
						for _, offset := range []int{-1, 0} {
							if got := treeGuideColumns(tt.Image(), y+offset, marker); len(got) != 1 {
								t.Fatalf("surface gap interrupted guide: %v", got)
							}
						}
					}
					// The visually empty bottom pixel still belongs to Alpha.
					tt.ClickAt(float32(x), b.Y-.5)
					treeChosen(t, &s, "b")
					if changes != 1 {
						t.Fatalf("gap click notified %d times", changes)
					}
					charlie := bounds(t, tt, "Charlie")
					tt.Move(float32(x), charlie.Y+charlie.H/2)
					if got := tt.Image().RGBAAt(x, int(charlie.Y)); got != rgba(theme.Background) {
						t.Fatalf("hover filled top inset: %v", got)
					}
					if got := tt.Image().RGBAAt(x, int(charlie.Y)+1); got == rgba(theme.Background) {
						t.Fatal("hover surface missing")
					}
				})
			}
		}
	}
}
