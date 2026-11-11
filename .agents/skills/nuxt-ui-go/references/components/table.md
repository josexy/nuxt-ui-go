<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Table

Display rows and columns with sorting, selection and scrolling.

- Category: Forms and data
- Version: nuxt-ui-go v0.0.2 (MyGo v0.8.3)
- Functions: `Table`
- Types: `TableOptions`, `TableColumn`, `TableState`
- Theme key: `nui.ComponentTable` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Set data with TableState.SetRows and stable, unique row IDs. Value supplies text, Compare defines sorting, and Cell customizes cells; Sortable can explicitly disable sorting. Clearing sorting restores the latest SetRows input order. Point List.SortBy at a persistent slice for multi-column sorting; Sort mirrors its first entry. ManualSorting preserves input order so the application can fetch data for the requested sort.

## Things to know

The header stays above the internal scrolling body and native virtualization is always enabled. Fixed only prevents column movement and resizing. Headers are not separate keyboard stops, so this example provides sorting buttons. OnSelectionChange reports selected IDs; OnActivate runs on Enter or double-click. These differ from Nuxt’s row-click select event. Filtering, pagination and requests belong to the application; pinned columns, grouped headers, expanded rows and aggregation are not provided.

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
type TableLoadingAnimation uint8

const (
	TableLoadingCarousel TableLoadingAnimation = iota
	TableLoadingCarouselInverse
	TableLoadingSwing
	TableLoadingElastic
)
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
| `Row` / `RowID` / `Len` | Read a row, a row ID, and the row count in the current view |

Columns can also set Width, MinWidth, MaxWidth, and Align. Sorting changes row positions, so when editing data, look up the business model by RowID or by the IDs in Selection. CellBuilds counts the cell builder calls in the current build.

Clearing the sort restores the input order from the latest SetRows, which is also kept for rows that compare as equal. With multi-column sorting, Shift-click or Ctrl-click (Cmd-click on macOS) on a header adds a sort column; further clicks switch it to descending and then remove it. An ordinary click toggles that column between ascending and descending. Removing a column or disabling its sorting removes the invalid sort entries.

OnSortChange and OnSelectionChange report sort and selection changes caused by the user; outside model assignments and cleanup by SetRows do not trigger them. Selection notifications return stable IDs sorted as strings. OnActivate returns a row ID on Enter or double-click, and OnChange keeps the native combined notification. These semantics differ from Nuxt UI's select event on row click. List.Selected is an optional active-row index; when it is not bound, OnActivate uses an internal index. Still use IDs when acting on business data.

Table sets List.Key, Selection, Sort, and Label; do not override these bindings in your app. List.SortBy, Columns, Selected, KeepOffset, FollowEnd, HeaderMenu, CellMenu, and similar fields follow MyGo's rules. List.CellSelect is a native extension, and this wrapper has no RowHeader column. Full cell selection semantics differ from row selection notifications.

##### Filtering and scope

For search, filtering, and pagination, your app computes the current rows first and then calls SetRows.

EmptyText and BuildEmpty set the empty content below the header. Loading shows a loading bar and keeps the existing data. With no data, BuildLoading's content is shown if provided; otherwise the empty content is still shown. LoadingColor defaults to Primary. LoadingAnimation can be Carousel, CarouselInverse, Swing, or Elastic, and a static bar is used with reduced motion. Caption is the table's description for assistive technology, and BuildCaption can show a visible title. Height defaults to 320 DIP, including the title; the header stays fixed and the body scrolls inside.

Virtualization is always on. It reduces only the number of cells built; generating data, validating IDs, sorting, and filtering still scale with the data size. MyGo handles row layout. Cells can override Item and Label through a column's UI and Style, and column overrides are applied after table-level overrides.

`Fixed` **only prevents moving and resizing the column**; it does not freeze the column during scrolling. Frozen columns, multi-level headers, arbitrary row or header builders, expandable rows, grouping and aggregation, and full TanStack semantics are not yet supported. Native headers have no keyboard focus of their own; to support keyboard sorting, provide buttons that change Sort or SortBy. The underlying List measures row heights, so custom cells can be taller; your app should still verify complex dynamic content.

#### Style part index (excerpt)

| Component | Parts |
| --- | --- |
| Table | Root, CaptionSlot, Body, Panel, Item, Label, EmptySlot, LoadingSlot, Indicator |

InputNumber's and InputTags' `UI` support every part in the table, while the embedded `InputOptions.Style` still receives only the Root, Editor, Leading, and Trailing nodes. Configure InputNumber's button callbacks through IncrementOptions and DecrementOptions. Table's internal native regions and ColorPicker's unexposed internal parts use the Native theme. RangeSlider draws its track and thumbs from the Theme and Color; apart from Indicator.Background and Thumb.Background, other Indicator and Thumb fields have no effect. Parts that are not exposed cannot be changed by writing their names yourself. For the precedence of global WithComponent settings and instance UI/Style, see [Style overrides](theme.md).

### Options and defaults

#### Common options by component (excerpt)

Table's Sortable, ManualSorting, and List.SortBy control, respectively, whether a column can be sorted, whether the app sorts the data, and multi-column sorting. Loading keeps the existing rows and shows a loading bar; BuildLoading replaces the empty content only while an empty table is loading. Caption provides a description and BuildCaption a visible title. For sort, selection, and activation notifications and native limits, see Table.

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
	nui "github.com/josexy/nuxt-ui-go"
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
