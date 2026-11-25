<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Table

Display tabular data with virtual scrolling, single/multi-column sorting, and row selection.

- Category: Forms and data
- Version: nuxt-ui-go v0.1.1 plus unreleased changes (MyGo v0.10.0)
- Functions: `Table`
- Types: `TableOptions`, `TableColumn`, `TableState`
- Theme key: `nui.ComponentTable` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Load data with `TableState.SetRows`, giving each row a stable, unique `ID`. Columns define text via `Value`, local sorting via `Compare`, and custom rendering via `Cell`. Point `List.SortBy` at a persistent slice for multi-column sorting, or enable `ManualSorting` for server-side sorting.

## Things to know

- Headers stay pinned above the virtualized body. `Fixed` locks column reordering and resizing—it does not freeze columns while horizontally scrolling.
- `OnSelectionChange` reports selected IDs; `OnActivate` fires on `Enter` or double-click (unlike Nuxt UI's single-click `select`).
- Filtering, pagination, and fetching belong to your app (use `List.EndReached` + `NearEnd` for load-more). Frozen columns, grouped headers, row expansion, and aggregation are not supported.

## API

```go
func Table[T any](c *ui.Context, t Theme, s *TableState[T], columns []TableColumn[T], o TableOptions) ui.Element
```

### TableOptions

```go
type TableOptions struct {
	Label, EmptyText string
	Caption          string
	Height           float32
	Disabled         bool
	// ManualSorting keeps the SetRows order while headers update sort state.
	ManualSorting                          bool
	Loading                                bool
	LoadingColor                           Color
	LoadingAnimation                       TableLoadingAnimation
	BuildCaption, BuildEmpty, BuildLoading func(*ui.Context, Theme)
	OnSortChange                           func([]ui.SortOrder)
	OnSelectionChange                      func([]string)
	// OnActivate reports the row ID on Enter or double-click.
	OnActivate func(string)
	OnChange   func()
	Decoration
}
```

### TableColumn

```go
type TableColumn[T any] struct {
	ID, Title                 string
	Width, MinWidth, MaxWidth float32
	Align                     ui.Align
	Fixed                     bool // Prevents resizing/reordering; it does not freeze a scrolling column.
	// Sortable defaults to whether Compare is provided. Explicit true without
	// Compare requires ManualSorting; false disables this column's sort UI.
	Sortable Optional[bool]
	Value    func(T) string
	Compare  func(T, T) int
	Cell     func(*ui.Context, Theme, T, bool)
	Decoration
}
```

### TableState

```go
type TableState[T any] struct {
	List      ui.ListState
	Selection ui.Selection[string]
	Sort      ui.SortOrder

	CellBuilds int
	// contains filtered or unexported fields
}

// InvalidateSort requests one new sort after a comparator's captured inputs
// change. SetRows already invalidates sorting when the row data changes.
func (s *TableState[T]) InvalidateSort()

func (s *TableState[T]) Len() int

func (s *TableState[T]) Row(i int) T

func (s *TableState[T]) RowID(i int) string

// SetRows copies row values and validates identities once, outside rendering.
// Removed rows are removed from selection; retain the full data separately for filtering.
func (s *TableState[T]) SetRows(rows []T, key func(T) string)
```

### Related types

```go
// LoadingAnimation is how an indeterminate bar moves, as Nuxt UI's Progress
// animation presets: Progress's indicator and Table's loading bar.
type LoadingAnimation uint8

const (
	// LoadingCarousel slides a half-length indicator from the start to the end.
	LoadingCarousel LoadingAnimation = iota
	// LoadingCarouselInverse slides it from the end to the start.
	LoadingCarouselInverse
	// LoadingSwing swings it to the end and back.
	LoadingSwing
	// LoadingElastic grows it from the middle to the whole track and back.
	LoadingElastic
)
```

```go
// TableLoadingAnimation is the LoadingAnimation of Table's loading bar.
type TableLoadingAnimation = LoadingAnimation
```

Types documented in other references: `Color` → [theme.md](theme.md), `Decoration` → [theme.md](theme.md), `Optional` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Forms, dates, and data components

#### Table

##### Prepare rows and columns

TableState[T] stores the current row data, sorting, selection, and column layout. Call SetRows when the data changes; do not set the rows again on every view build.

```go
// Declare the type and state in your app model.
type TaskRow struct {
    ID string
    Title string
}

var tasks nui.TableState[TaskRow]

// Run on initialization or when the data changes.
tasks.SetRows([]TaskRow{
    {ID: "task-1", Title: "Collect requirements"},
    {ID: "task-2", Title: "Review the docs"},
}, func(row TaskRow) string { return row.ID })
```

Define the displayed columns and create the table in the view:

```go
nui.Table(c, theme, &tasks, []nui.TableColumn[TaskRow]{
    {
        ID: "title",
        Title: "Task",
        Width: 240,
        Value: func(row TaskRow) string { return row.Title },
    },
}, nui.TableOptions{Label: "Task list", Height: 320})
```

SetRows copies the row values and checks that each ID is non-empty and unique. If T contains pointers or slices, your app still owns the data they point to. Replacing the data removes selected IDs that no longer exist.

##### Sorting, selection, and custom cells

| Field or method | Purpose |
| --- | --- |
| `TableColumn.Compare` | Stable local sorting; returns the comparison of two rows |
| `TableColumn.Sortable` | When unset, determined by Compare. An explicit false prevents sorting the column; true without Compare requires ManualSorting |
| `TableColumn.Cell` | A custom cell that receives the Context, the Theme, the row value, and whether it is selected |
| `Selection` / `Sort` | A `ui.Selection[string]` and a `ui.SortOrder`, respectively |
| `state.List.SortBy` | Point it at a persistent `[]ui.SortOrder` to enable multi-column sorting; Sort mirrors the first entry |
| `ManualSorting` | Keeps the input order from SetRows; your app loads or computes the data based on the sort state |
| `InvalidateSort` | Requests a re-sort when data captured by a comparator changes. SetRows invalidates the sort automatically |
| `state.List.Columns` | Stores column widths and order; your app can persist it |
| `state.List.RowHeight` | Above 0, every body row is that many DIPs high, with its cells in the middle; the header keeps its height. 0 keeps the default row height |
| `state.List.EndReached` / `NearEnd` | EndReached is called once, with the index of the last row in view, when the user's scrolling brings the rows in view within NearEnd rows of the end; see Filtering and scope |
| `Row` / `RowID` / `Len` | Read a row, a row ID, and the row count in the current view |

Columns can also set Width, MinWidth, MaxWidth, and Align. Sorting changes row positions, so when editing data, look up the business model by RowID or by the IDs in Selection. CellBuilds counts the cell builder calls in the current build.

Clearing the sort restores the input order from the latest SetRows, which is also kept for rows that compare as equal. With multi-column sorting, Shift-click or Ctrl-click (Cmd-click on macOS) on a header adds a sort column; further clicks switch it to descending and then remove it. An ordinary click toggles that column between ascending and descending. Removing a column or disabling its sorting removes the invalid sort entries.

OnSortChange and OnSelectionChange report sort and selection changes caused by the user; outside model assignments and cleanup by SetRows do not trigger them. Selection notifications return stable IDs sorted as strings. OnActivate returns a row ID on Enter or double-click, and OnChange keeps the native combined notification. These semantics differ from Nuxt UI's select event on row click. List.Selected is an optional active-row index; when it is not bound, OnActivate uses an internal index. Still use IDs when acting on business data.

OnChange fires once for each user change: a header click that sorts (with OnSortChange), a selection change (with OnSelectionChange), and, since MyGo v0.9.0, a finished column move, column resize, or double-click on a column edge that fits the column to its cells. A drag reports nothing until the pointer is released, and a drag that ends with the same order and widths, or a gesture on a Fixed column, reports nothing. Column gestures call neither OnSortChange nor OnSelectionChange, and rebuilding without user input calls no callback. Save the column layout from OnChange:

```go
nui.Table(c.Key("files"), t, &app.files, columns, nui.TableOptions{
	Label: "Files",
	OnChange: func() {
		app.settings.Columns = app.files.List.Columns
		app.saveSettings()
	},
})
```

Restore it by assigning the saved `ui.TableLayout` to `List.Columns` before the first build. The internal active row that OnActivate reads follows its record: after sorting or several SetRows calls before the next build, Enter activates the same record even though its row moved. When SetRows removes that record, no row is active and Enter does nothing until the user chooses another row.

Table sets List.Key, Selection, Sort, and Label; do not override these bindings in your app. List.SortBy, Columns, Selected, KeepOffset, FollowEnd, RowHeight, EndReached, NearEnd, HeaderMenu, CellMenu, and similar fields follow MyGo's rules. List.MoveAnimation is not supported through Table yet: nui's motion levels do not govern it, and MyGo turns it off only for the desktop's reduced-motion setting, so rows would still glide when `AppConfig.DisableAnimations` turns nui's motion off. List.CellSelect is a native extension, and this wrapper has no RowHeader column. Full cell selection semantics differ from row selection notifications.

##### Filtering and scope

For search, filtering, and pagination, your app computes the current rows first and then calls SetRows.

To load more rows as the user scrolls, set `state.List.EndReached`, and `NearEnd` for how many rows before the last count as near the end. Start the request in the background, then append the page with SetRows in `Window.Update`. EndReached is called once per approach to the end: not again while the rows in view stay there, and not when the table first shows or rebuilds, when SetRows appends rows, when the app scrolls with `List.ScrollToEnd`, or when every row already fits. After a page is appended, the user's next scroll near the new end calls it once more. It calls none of OnChange, OnSortChange, OnSelectionChange, or OnActivate.

EmptyText and BuildEmpty set the empty content below the header. Loading shows a loading bar and keeps the existing data. With no data, BuildLoading's content is shown if provided; otherwise the empty content is still shown. LoadingColor defaults to Primary. LoadingAnimation can be Carousel, CarouselInverse, Swing, or Elastic, the same presets as Progress's Animation (`TableLoadingAnimation` is an alias of `LoadingAnimation`). Under reduced motion the bar spans the header and pulses its opacity, as Nuxt's `motion-reduce:animate-pulse`; at the Off motion level it is a static full-width bar. Caption is the table's description for assistive technology, and BuildCaption can show a visible title. Height defaults to 320 DIP, including the title; the header stays fixed and the body scrolls inside.

Virtualization is always on. It reduces only the number of cells built; generating data, validating IDs, sorting, and filtering still scale with the data size. MyGo handles row layout. Cells can override Item and Label through a column's UI and Style, and column overrides are applied after table-level overrides.

`Fixed` **only prevents moving and resizing the column**; it does not freeze the column during scrolling. Frozen columns, multi-level headers, arbitrary row or header builders, expandable rows, grouping and aggregation, and full TanStack semantics are not yet supported. Native headers have no keyboard focus of their own; to support keyboard sorting, provide buttons that change Sort or SortBy. The underlying List measures row heights, so custom cells can be taller; your app should still verify complex dynamic content.

#### Style part index (excerpt)

| Component | Parts |
| --- | --- |
| Table | Root, CaptionSlot, Body, Panel, Item, Label, EmptySlot, LoadingSlot, Indicator, SeparatorSlot |

InputNumber's and InputTags' `UI` support every part in the table, while the embedded `InputOptions.Style` still receives only the Root, Editor, Leading, and Trailing nodes. Configure InputNumber's button callbacks through IncrementOptions and DecrementOptions. Table's header divider paints one device pixel and takes its stroke color from `UI[SeparatorSlot].Background` (default BorderAccented). Its replacement covers MyGo's one-DIP divider using the theme background or an opaque Root/Body/Panel UI background; custom translucent surfaces or raw Style background changes should provide their own SeparatorSlot material. Table's other internal native regions and ColorPicker's unexposed internal parts use the Native theme. RangeSlider draws its track and thumbs from the Theme and Color; apart from Indicator.Background and Thumb.Background, other Indicator and Thumb fields have no effect. Parts that are not exposed cannot be changed by writing their names yourself. For the precedence of global WithComponent settings and instance UI/Style, see [Style overrides](theme.md).

### Options and defaults

#### Common options by component (excerpt)

Table's Sortable, ManualSorting, and List.SortBy control, respectively, whether a column can be sorted, whether the app sorts the data, and multi-column sorting. Loading keeps the existing rows and shows a loading bar; BuildLoading replaces the empty content only while an empty table is loading. Caption provides a description and BuildCaption a visible title. For sort, selection, and activation notifications and native limits, see Table.

### Themes and basic components

#### Components and parts › Progress and animation (excerpt)

Determinate progress fills to the rounded integer percentage, while assistive technology receives the exact amount completed. The indeterminate indicator moves in a two-second cycle along one of Nuxt UI's presets, set with `Animation`: `LoadingCarousel` (the default, a half-length indicator sliding from the start to the end), `LoadingCarouselInverse`, `LoadingSwing` or `LoadingElastic`. The carousels run from the other end right to left and when Inverted, and all presets run along a vertical track; Table's loading bar uses the same presets. Under reduced motion the indicator spans the whole track and pulses its opacity instead, as Nuxt's `motion-reduce:animate-pulse`; with Animate=false or at the Off motion level it stays still, half the track's length, in the center. A determinate value moves to its new amount over 200 ms only at the Full motion level. An out-of-range Animation panics, as Table's LoadingAnimation does. Internally, `Now=-1 < Min=0` represents unknown progress, with an "In progress" value or description. Windows UIA clamps the unknown value to the minimum; how screen readers announce it has not been verified.

**When animations stop:** Continuous animations request the next frame only while a visible component actually draws. They stop when the component scrolls out of view, unmounts, has Animate turned off, or at the Off motion level; spinners and the Skeleton pulse go on under reduced motion. Button's colour transition follows `Motion.Transition` at Full and Reduced and is immediate at Off, and `Animate: nui.Set(false)` controls the loading spinner. Once a transition ends, no more frames are requested. Indeterminate Progress and Table's loading bar pulse under reduced motion. In MyGo a colour transition builds the view in each frame while the colour moves (about ten builds for 150 ms), so an app with an expensive view can set `Motion.Transition` to 0 or turn animation off. Components do not use timer goroutines for animation.

### Inputs, selection, and fields

#### Popover and ScrollArea › ScrollArea: give content a scrolling region (excerpt)

ScrollArea supports nesting and two-way scrolling. It does not provide upstream's items/virtualize, edge fade shadows, infinite loading, or custom scrollbar interaction. All content is built every time, so use the virtual lists in Listbox or Table for long lists. Put long content in a Popover inside a ScrollArea yourself.

### Compatibility and known limitations

#### Feature scope › Data and files (excerpt)

| Area | Supported | Not provided, or left to the app |
| --- | --- | --- |
| Table / Tree | Native virtual lists and stable IDs. Table supports multi-column and manual sorting, separate model notifications, and loading content. Tree links parent and child selection for loaded nodes | Table headers have no keyboard focus of their own. No full TanStack configuration, frozen columns, multi-level headers, row grouping and aggregation, or tree drag and drop. See Table and [Tree](tree.md) |

## Complete example

A runnable program. App state lives in `example`, outside the view, and the view only builds UI.

```go
package main

import (
	"cmp"
	"fmt"
	"log"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

type example struct {
	Table                         nui.TableState[string]
	Orders                        []ui.SortOrder
	Multiple                      bool
	SortChanges, SelectionChanges int
	Activated                     string
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	direction := ui.LeftToRight

	ui.Row(c).Wrap().Gap(8).Children(func() {
		nui.Button(c.Key("sort"), t, "Sort Project descending", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() {
			s.Table.Sort = ui.SortOrder{Column: "name", Descending: true}
			if s.Multiple {
				s.Orders = []ui.SortOrder{s.Table.Sort}
			}
		})
		nui.Button(c.Key("restore"), t, "Restore input order", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { s.Table.Sort = ui.SortOrder{}; s.Orders = nil })
	})
	ui.Column(c.Key("direction")).Direction(direction).Children(func() {
		nui.Table(c.Key("table"), t, &s.Table, []nui.TableColumn[string]{
			{ID: "name", Title: "Project", Width: 240, MinWidth: 100, Value: func(v string) string { return v }, Compare: func(a, b string) int { return strings.Compare(a, b) }},
			{ID: "length", Title: "Length", Width: 100, MinWidth: 80, Value: func(v string) string { return fmt.Sprint(len(v)) }, Compare: func(a, b string) int { return cmp.Compare(len(a), len(b)) }},
		}, nui.TableOptions{Label: "Projects", Caption: "Project records", Height: float32(240), LoadingAnimation: nui.TableLoadingCarousel, BuildLoading: func(c *ui.Context, _ nui.Theme) { ui.Text(c, "Loading projects") }, OnSortChange: func([]ui.SortOrder) { s.SortChanges++ }, OnSelectionChange: func([]string) { s.SelectionChanges++ }, OnActivate: func(id string) { s.Activated = id }})
	})
	first := "None"
	if s.Table.Len() > 0 {
		first = s.Table.RowID(0)
	}
	ui.Textf(c, "First: %s · Selected: %d", first, s.Table.Selection.Len())
	ui.Textf(c, "Sort changes: %d · Selection changes: %d · Activated: %s", s.SortChanges, s.SelectionChanges, s.Activated)
}

func main() {
	s := &example{}

	s.Table.SetRows([]string{"NuxtUI Go", "MyGo", "Documentation"}, func(v string) string { return v })
	s.Activated = "None"
	palettes := nui.DefaultPalettes()
	palettes.Primary, _ = nui.PaletteByName("green")
	palettes.Neutral, _ = nui.PaletteByName("slate")
	config := nui.AppConfig{Mode: nui.LightMode, Light: nui.NewTheme(false, palettes), Dark: nui.NewTheme(true, palettes)}
	mygo.App.SetLanguages(nui.SupportedLocales()...)
	mygo.App.WhenReady(func() {
		mygo.App.SetLocale("en")
		mygo.NewWindow(mygo.WindowOptions{Title: "Table · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
