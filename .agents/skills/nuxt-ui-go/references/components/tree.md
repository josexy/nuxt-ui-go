<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Tree

Expand hierarchical data and select its nodes.

- Category: Forms and data
- Version: nuxt-ui-go v0.0.2 plus unreleased changes (MyGo v0.9.0)
- Functions: `Tree`
- Types: `TreeOptions`, `TreeNode`, `TreeState`, `TreeEvent`, `TreeItemState`
- Theme key: `nui.ComponentTree` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

SetNodes indexes nodes by stable, unique IDs. Selection stores selection, Outline.Open stores expansion, and Active is the keyboard highlight. Call SetNodes only when data changes. OnSelect/OnToggle can cancel, while OnChange/OnExpandedChange report actual selection and expansion changes.

Indent guides round Indicator.BorderWidth from DIPs to whole device pixels; an explicit zero hides them.

## Things to know

Linkage covers loaded, enabled nodes, stops at disabled nodes, and does not select children loaded later. Mixed state applies only in multiple mode: BubbleSelect checks mixed descendants; PropagateSelect alone also requires a selected parent. Nested changes guides, backgrounds and hit areas; both modes use native virtualization. Leading/trailing builders replace default icons and should contain display content only; rows and the keyboard retain interaction. Empty branches retain expansion state without fallback folder/chevron icons.

## API

```go
func Tree(c *ui.Context, t Theme, s *TreeState, o TreeOptions) ui.Element
```

### TreeOptions

```go
type TreeOptions struct {
	Label, EmptyText                    string
	Height                              float32
	Color                               Color
	Size                                Size
	Disabled, Multiple                  bool
	SelectionBehavior                   SelectionBehavior // Default: SelectionToggle.
	PropagateSelect, BubbleSelect, Loop bool
	// Nested defaults to true: draw ancestor guides and inset child rows.
	// Both modes use the same flat, virtualized native row model, not nested DOM.
	Nested                                    Optional[bool]
	ExpandOnClick                             Optional[bool] // Default true; Enter/Space only select.
	ExpandedIcon, CollapsedIcon, TrailingIcon *ui.SVG
	// Build replaces non-interactive row content, after the leading icon.
	Build                       func(*ui.Context, Theme, TreeItemState)
	BuildLeading, BuildTrailing func(*ui.Context, Theme, TreeItemState)
	OnSelect, OnToggle          func(*TreeEvent)
	OnChange                    func()         // Once when an interaction changes selection.
	OnExpandedChange            func([]string) // User expansion changes, as sorted stable IDs.
	OnLoad                      func(string, TreeTicket)
	Decoration
}
```

### TreeNode

```go
type TreeNode struct {
	ID, Label       string
	Icon            *ui.SVG
	TrailingIcon    *ui.SVG
	Children        []TreeNode
	Lazy            bool
	Disabled        bool
	DefaultExpanded bool
	OnSelect        func(*TreeEvent)
	OnToggle        func(*TreeEvent)
	Decoration
}
```

### TreeState

```go
type TreeState struct {
	Outline   ui.OutlineState[string]
	Selection ui.Selection[string]

	RowBuilds int
	// Active is the keyboard highlight, independent of selection.
	Active string
}

// Close invalidates pending loads when the owning view is disposed.
func (s *TreeState) Close()

// CompleteLoad must be called on the window's UI thread. Collapsed/replaced requests are rejected.
func (s *TreeState) CompleteLoad(ticket TreeTicket, children []TreeNode, err error) bool

func (s *TreeState) Depth(row int) int

func (s *TreeState) Error(id string) string

func (s *TreeState) Item(row int) string

func (s *TreeState) Retry(id string)

func (s *TreeState) Roots() []string

// Rows, Item and Depth describe the last built visible model. Outline.Open and
// Outline.List remain the persistent expansion and scrolling state; the legacy
// Outline.Rows/Item/Depth methods are no longer populated.
func (s *TreeState) Rows() int

// SetExpanded changes expansion without emitting interaction callbacks. Unlike
// direct Open edits, it invalidates pending loads immediately, even if reopened
// before the next frame. All state methods must run on the window's UI thread.
func (s *TreeState) SetExpanded(id string, expanded bool)

func (s *TreeState) SetNodes(nodes []TreeNode)
```

### TreeEvent

```go
// TreeEvent describes the state before an interaction. A node handler overrides
// the corresponding TreeOptions handler. PreventDefault cancels that action only.
type TreeEvent struct {
	ID                 string
	Selected, Expanded bool
	// contains filtered or unexported fields
}

func (e *TreeEvent) PreventDefault()
```

### TreeItemState

```go
// TreeItemState is a snapshot for synchronous, non-interactive row content.
type TreeItemState struct {
	ID, Label                                  string
	Level                                      int
	Selected, Expanded, Indeterminate, Loading bool
}
```

### Related types

```go
type SelectionBehavior uint8

const (
	SelectionToggle SelectionBehavior = iota
	SelectionReplace
)
```

```go
type TreeTicket struct {
	// contains filtered or unexported fields
}
```

Types documented in other references: `Color` → [theme.md](theme.md), `Decoration` → [theme.md](theme.md), `Optional` → [theme.md](theme.md), `Size` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Forms, dates, and data components

#### Tree

##### Store nodes, selection, and expansion

TreeState.SetNodes takes a list of TreeNodes and checks that every ID in the tree is non-empty and unique. The state fields are:

| State | Purpose |
| --- | --- |
| `state.Outline.Open` | IDs of expanded nodes |
| `state.Selection` | IDs of selected nodes |
| `state.Active` | ID of the node highlighted by the keyboard |

The default is single selection, and clicking a row both selects it and toggles its expansion. Arrow keys only navigate; Enter and Space select. Multiple enables multiple selection, and SelectionBehavior chooses between toggling and replacing the selection. PropagateSelect and BubbleSelect control parent-child linking, and Loop controls whether navigation wraps around.

Nodes can be Disabled, and OnSelect and OnToggle can cancel the current action. Parent-child linking applies only to loaded nodes, and disabled nodes block it. Children loaded later do not inherit the selection automatically.

##### Indentation and virtualization

Nested defaults to true, which shows ancestor guide lines and indents each child's background and click area by level. With `nui.Set(false)`, the background and click area span the full row while the content stays indented. Switching keeps the selection, expansion, focus, and scroll position.

Both modes use a native virtual list that builds only the rows near the viewport. The index of visible nodes is rebuilt when the data or the expansion state changes. Because apps may edit the public collections directly, every build still checks the expanded IDs, and also the selected IDs when indeterminate states are needed. Customize row content with Build.

##### Lazy loading

A node with Lazy=true and Children=nil calls `OnLoad(id, ticket)` once when it is visible and expanded. Your app starts loading in the background and returns the result in `Window.Update` with `CompleteLoad(ticket, children, err)`.

A failed load shows Retry. Collapsing the node or an ancestor with SetExpanded, replacing the data, retrying, or calling Close all invalidate the old ticket. A true result from CompleteLoad only means the result was accepted; it may still contain an error.

When SetNodes replaces the data, it removes selected and expanded IDs that no longer exist. An empty root list shows EmptyText.

##### Migrating from the old Outline wrapper

| Behavior you need | Current code |
| --- | --- |
| Multiple selection | Set `Multiple: true` explicitly |
| Multiple selection that replaces | `SelectionBehavior: nui.SelectionReplace` |
| Click selects without expanding | `ExpandOnClick: nui.Set(false)` |
| Listen for selection / expansion | Use OnChange / OnToggle |
| Query visible rows | Use `state.Rows()`, `Item()`, and `Depth()` instead of the methods of the same names on `state.Outline` |

Outline.List keeps the scroll state and Handle, while Tree manages Selected, Selection, and Label. Drag-and-drop moving and a full business permission model are not provided.

OnExpandedChange fires after a user action actually changes the expansion state, with the expanded IDs sorted as strings. Cancellation, initialization, and outside SetExpanded calls do not trigger it. OnChange still reports only selection changes. If a selection handler replaces the nodes, the earlier click no longer expands a newly created node with the same ID.

A TreeNode's UI and Style can override Item, Leading, Label, and Trailing; node overrides are applied after the tree instance's overrides. BuildLeading and BuildTrailing replace the leading and trailing icons, and Build replaces the main content; these callbacks should build display content only. TreeItemState.Indeterminate is used only for multiple-selection linking. With BubbleSelect, a node is indeterminate when its available descendants are a mix of selected and unselected; with only PropagateSelect, the parent must also be selected. In single selection, or with both linking options off, it is false.

A loaded empty branch keeps its expansion state but does not get the default folder icon and arrow; set Icon explicitly if you want a folder look. Type-ahead search cycles through matches when the same character is repeated, and keeps the current match while the user continues typing a prefix.

#### Style part index (excerpt)

| Component | Parts |
| --- | --- |
| Tree | Root, List, Item, Leading, Label, Trailing, Indicator, EmptySlot |

### Options and defaults

#### Common options by component (excerpt)

Tree's OnSelect and OnToggle can cancel an operation; OnChange and OnExpandedChange report actual changes to selection and expansion. TreeNode.UI and Style style an individual node's parts, and BuildLeading and BuildTrailing replace the non-interactive leading and trailing content. For indeterminate states with multiple selection, lazy loading, and empty branches, see Tree.

### Compatibility and known limitations

#### Feature scope › Data and files (excerpt)

| Area | Supported | Not provided, or left to the app |
| --- | --- | --- |
| Table / Tree | Native virtual lists and stable IDs. Table supports multi-column and manual sorting, separate model notifications, and loading content. Tree links parent and child selection for loaded nodes | Table headers have no keyboard focus of their own. No full TanStack configuration, frozen columns, multi-level headers, row grouping and aggregation, or tree drag and drop. See [Table](table.md) and Tree |

#### Accessibility and motion (excerpt)

| Area | Current limitation |
| --- | --- |
| Tree | Provides Tree/TreeItem roles, levels, selection, and active item semantics. The current virtual list lacks the assistive technology expand/collapse actions built into `ui.Outline` |

## Complete example

A runnable program. App state lives in `example`, outside the view, and the view only builds UI.

```go
package main

import (
	"log"
	"slices"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

type example struct {
	Tree                       nui.TreeState
	Ticket                     nui.TreeTicket
	Changes, Expansions, Loads int
	Result                     string
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	direction := ui.LeftToRight

	ui.Column(c.Key("direction")).Direction(direction).Children(func() {
		nui.Tree(c.Key("tree"), t, &s.Tree, nui.TreeOptions{Label: "Project files", Height: 280, Nested: nui.Set(true), SelectionBehavior: nui.SelectionToggle, ExpandOnClick: nui.Set(true),
			OnSelect: func(e *nui.TreeEvent) {
			},
			OnToggle: func(e *nui.TreeEvent) {
			},
			OnChange: func() { s.Changes++ }, OnExpandedChange: func([]string) { s.Expansions++ },
			OnLoad: func(_ string, ticket nui.TreeTicket) { s.Ticket = ticket; s.Loads++; s.Result = "Loading remote files" }})
	})
	ui.Row(c).Wrap().Gap(8).Children(func() {
		nui.Button(c.Key("collapse"), t, "Collapse remote", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { s.Tree.SetExpanded("remote", false); s.Result = "Remote branch collapsed" })
		nui.Button(c.Key("complete"), t, "Complete empty load", nui.ButtonOptions{Variant: nui.Outline, Disabled: s.Loads == 0}).OnClick(func() {
			if s.Tree.CompleteLoad(s.Ticket, nil, nil) {
				s.Result = "Remote directory is empty"
			} else {
				s.Result = "Stale completion rejected"
			}
		})
	})
	ui.Textf(c, "Selected IDs: %v", slices.Sorted(s.Tree.Selection.All()))
	ui.Textf(c, "Selection changes: %d · Expansion changes: %d", s.Changes, s.Expansions)
	ui.Textf(c, "%s · Loads: %d", s.Result, s.Loads)
}

func main() {
	s := &example{}

	s.Tree.SetNodes([]nui.TreeNode{
		{ID: "src", Label: "src", DefaultExpanded: true, Children: []nui.TreeNode{
			{ID: "main", Label: "main.go", Icon: lucide.FileCode(), Decoration: nui.Decoration{UI: nui.Styles{nui.Label: {Weight: nui.Set(600)}}}},
			{ID: "theme", Label: "theme.go", Icon: lucide.FileCode()},
			{ID: "components", Label: "components", DefaultExpanded: true, Children: []nui.TreeNode{{ID: "button-file", Label: "button.go", Icon: lucide.FileCode()}, {ID: "input-file", Label: "input.go", Icon: lucide.FileCode()}}},
			{ID: "locked", Label: "generated.go (disabled)", Disabled: true},
		}},
		{ID: "docs", Label: "docs", Children: []nui.TreeNode{{ID: "readme", Label: "README.md", Icon: lucide.FileText()}}},
		{ID: "remote", Label: "Remote files", Lazy: true},
	})
	s.Result = "Ready"
	defer s.Tree.Close()
	palettes := nui.DefaultPalettes()
	palettes.Primary, _ = nui.PaletteByName("green")
	palettes.Neutral, _ = nui.PaletteByName("slate")
	config := nui.AppConfig{Mode: nui.LightMode, Light: nui.NewTheme(false, palettes), Dark: nui.NewTheme(true, palettes)}
	mygo.App.SetLanguages(nui.SupportedLocales()...)
	mygo.App.WhenReady(func() {
		mygo.App.SetLocale("en")
		mygo.NewWindow(mygo.WindowOptions{Title: "Tree · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
