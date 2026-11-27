package nui

import (
	"cmp"
	"fmt"
	"reflect"
	"slices"
	"strconv"

	"github.com/egoist/mygo/ui"
)

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
	// BuildCell takes precedence over Cell and Value. The Element parameter is
	// the existing styled content container, valid only during this callback.
	BuildCell func(*ui.Context, Theme, ui.Element, TableCell[T])
	Decoration
}

// TableCell describes the displayed row supplied to BuildCell.
type TableCell[T any] struct {
	Value    T
	Index    int
	Selected bool
}
type tableRecord[T any, K TableKey] struct {
	id      K
	value   T
	ordinal int
}

// TableKey is a stable string or integer row identity. Numeric zero is valid.
type TableKey interface {
	~string | ~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// TableState preserves the string-keyed table API.
type TableState[T any] = KeyedTableState[T, string]

// KeyedTableState keeps typed row identities, selection, sorting and layout.
// SetRows owns a copy; SetBorrowedRows reads caller-owned data without copying it.
type KeyedTableState[T any, K TableKey] struct {
	List      ui.ListState
	Selection ui.Selection[K]
	Sort      ui.SortOrder
	rows      []tableRecord[T, K]
	// pendingIDs validates a snapshot before changing rows. ids is the last
	// validated membership set; unchanged IDs can reuse it without rebuilding.
	pendingIDs                []K
	ids                       map[K]bool
	idsCapacity               int
	idsValid                  bool
	applied                   []ui.SortOrder
	revision, sortedRevision  uint64
	inputOrder, manualApplied bool
	active                    int
	activeInitialized         bool
	CellBuilds                int

	borrowed      []T
	borrowedIDs   []K
	borrowedOrder []int
	borrowing     bool
	selectionPeak int
}

// SetRows copies row values and validates identities once, outside rendering.
// Removed rows are removed from selection; retain the full data separately for filtering.
func (s *KeyedTableState[T, K]) SetRows(rows []T, key func(T) K) {
	if key == nil {
		panic("nui: Table requires row IDs")
	}
	n := len(rows)
	// Retain working storage across updates, but release large snapshots after
	// a substantial shrink. Empty snapshots release all row storage.
	limit := max(64, 4*n)
	keys := s.pendingIDs
	if n == 0 {
		keys = nil
	} else if cap(keys) > limit {
		keys = make([]K, n)
	} else {
		keys = slices.Grow(keys[:0], n)[:n]
	}
	committed := false
	defer func() {
		clear(keys)
		if committed {
			s.pendingIDs = keys
		} else {
			// A bad ID or a panicking key function must leave the visible
			// snapshot intact and must not retain values from the failed one.
			s.pendingIDs = nil
			if !s.idsValid {
				s.ids, s.idsCapacity = nil, 0
			}
		}
	}()
	unchanged := n == len(s.rows)
	for i, v := range rows {
		id := key(v)
		if i < len(s.rows) && id == s.rows[i].id {
			// Keep the existing string when a key function formats the same
			// ID again. The caller's temporary can then be collected.
			id = s.rows[i].id
		} else {
			unchanged = false
		}
		keys[i] = id
	}
	if !unchanged || !s.idsValid {
		s.idsValid = false
		if n == 0 {
			s.ids, s.idsCapacity = nil, 0
		} else if s.ids == nil || s.idsCapacity < n || s.idsCapacity > limit {
			capacity := n
			if s.idsCapacity < n {
				capacity = max(n, 2*s.idsCapacity)
			}
			s.ids, s.idsCapacity = make(map[K]bool, capacity), capacity
		} else {
			clear(s.ids)
		}
		for _, id := range keys {
			checkTableID(s.ids, id)
		}
	}
	activeID, hasActive := s.activeRowID()
	// No callbacks or validation remain: it is now safe to overwrite records.
	s.borrowed, s.borrowedIDs, s.borrowedOrder, s.borrowing = nil, nil, nil, false
	s.selectionPeak = max(s.selectionPeak, s.Selection.Len())
	if n == 0 {
		s.rows = nil
	} else if cap(s.rows) > limit {
		s.rows = make([]tableRecord[T, K], n)
	} else {
		clear(s.rows[min(n, len(s.rows)):])
		s.rows = slices.Grow(s.rows[:0], n)[:n]
	}
	for i, v := range rows {
		s.rows[i] = tableRecord[T, K]{keys[i], v, i}
	}
	s.idsValid = true
	s.revision++
	s.inputOrder = true
	for id := range s.Selection.All() {
		if !s.ids[id] {
			s.Selection.Remove(id)
		}
	}
	s.restoreActiveRow(activeID, hasActive)
	s.compactSelection()
	committed = true
}
func (s *KeyedTableState[T, K]) Len() int {
	if s.borrowing {
		return len(s.borrowed)
	}
	return len(s.rows)
}

func (s *KeyedTableState[T, K]) Row(i int) T {
	if s.borrowing {
		if !s.inputOrder && s.borrowedOrder != nil {
			i = s.borrowedOrder[i]
		}
		return s.borrowed[i]
	}
	return s.rows[i].value
}

func (s *KeyedTableState[T, K]) RowID(i int) K {
	if s.borrowing {
		if !s.inputOrder && s.borrowedOrder != nil {
			i = s.borrowedOrder[i]
		}
		return s.borrowedIDs[i]
	}
	return s.rows[i].id
}

// Keep the internally owned cursor aligned with every snapshot and sort,
// including several updates before MyGo builds the table again. App-owned
// Selected bindings retain MyGo's own cursor semantics.
func (s *KeyedTableState[T, K]) activeRowID() (K, bool) {
	if s.List.Selected == &s.active && s.active >= 0 && s.active < s.Len() {
		return s.RowID(s.active), true
	}
	var zero K
	return zero, false
}

func (s *KeyedTableState[T, K]) restoreActiveRow(id K, present bool) {
	if present {
		index := -1
		for i := range s.Len() {
			if s.RowID(i) == id {
				index = i
				break
			}
		}
		// The wrapper has reconciled the cursor, including numeric identity zero.
		s.List.Choose(index)
	}
}

func tableIDLabel[K TableKey](id K) string {
	switch value := any(id).(type) {
	case string:
		return value
	case uint64:
		return strconv.FormatUint(value, 10)
	case int:
		return strconv.Itoa(value)
	default:
		return fmt.Sprint(id)
	}
}

func checkTableID[K TableKey](seen map[K]bool, id K) {
	var zero K
	if id == zero && reflect.TypeFor[K]().Kind() == reflect.String {
		panic("nui: Table rows require non-empty IDs")
	}
	if seen[id] {
		panic(fmt.Sprintf("nui: duplicate Table row ID %v", id))
	}
	seen[id] = true
}

// InvalidateSort requests one new sort after a comparator's captured inputs
// change. SetRows already invalidates sorting when the row data changes.
func (s *KeyedTableState[T, K]) InvalidateSort() { s.revision++ }

func (s *KeyedTableState[T, K]) sort(columns []TableColumn[T], manual bool) {
	by := map[string]TableColumn[T]{}
	for _, col := range columns {
		if valueOr(col.Sortable, col.Compare != nil) {
			by[col.ID] = col
		}
	}
	orders := []ui.SortOrder{s.Sort}
	if s.List.SortBy != nil {
		orders = *s.List.SortBy
	}
	valid := make([]ui.SortOrder, 0, len(orders))
	seen := map[string]bool{}
	for _, order := range orders {
		if _, ok := by[order.Column]; ok && !seen[order.Column] {
			valid = append(valid, order)
			seen[order.Column] = true
		}
	}
	if s.List.SortBy != nil && !slices.Equal(*s.List.SortBy, valid) {
		*s.List.SortBy = slices.Clone(valid)
	}
	s.Sort = ui.SortOrder{}
	if len(valid) > 0 {
		s.Sort = valid[0]
	}
	if slices.Equal(s.applied, valid) && s.sortedRevision == s.revision && s.manualApplied == manual {
		return
	}
	defer s.restoreActiveRow(s.activeRowID())
	s.applied = slices.Clone(valid)
	s.sortedRevision = s.revision
	s.manualApplied = manual
	if s.borrowing {
		s.sortBorrowed(valid, by, manual)
		return
	}
	if manual || len(valid) == 0 {
		if !s.inputOrder {
			slices.SortFunc(s.rows, func(a, b tableRecord[T, K]) int { return cmp.Compare(a.ordinal, b.ordinal) })
		}
		s.inputOrder = true
		return
	}
	slices.SortFunc(s.rows, func(a, b tableRecord[T, K]) int {
		for _, order := range valid {
			compare := by[order.Column].Compare
			left, right := a.value, b.value
			if order.Descending {
				left, right = right, left
			}
			if v := compare(left, right); v != 0 {
				return v
			}
		}
		// Equal values keep SetRows order, independently of previous sorts.
		return cmp.Compare(a.ordinal, b.ordinal)
	})
	s.inputOrder = false
}

// TableOptions preserves the string-keyed table options and callbacks.
type TableOptions = KeyedTableOptions[string]

// KeyedTableOptions configures Table with callbacks using its row identity type.
type KeyedTableOptions[K TableKey] struct {
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
	OnSelectionChange                      func([]K)
	// OnActivate reports the row ID on Enter or double-click.
	OnActivate func(K)
	OnChange   func()
	Decoration
}

func Table[T any, K TableKey](c *ui.Context, t Theme, s *KeyedTableState[T, K], columns []TableColumn[T], o KeyedTableOptions[K]) ui.Element {
	if s == nil {
		panic("nui: Table requires state")
	}
	defer s.compactSelection()
	checkLoadingAnimation(o.LoadingAnimation, "table")
	if o.Label == "" {
		o.Label = t.field.Label
	}
	if o.Label == "" {
		o.Label = o.Caption
	}
	seen := map[string]bool{}
	native := make([]ui.TableColumn, len(columns))
	for i, col := range columns {
		checkID(seen, col.ID, "Table columns")
		sortable := valueOr(col.Sortable, col.Compare != nil)
		if sortable && col.Compare == nil && !o.ManualSorting {
			panic("nui: sortable Table columns require Compare or ManualSorting")
		}
		native[i] = ui.TableColumn{ID: col.ID, Title: col.Title, Width: col.Width, MinWidth: col.MinWidth, MaxWidth: col.MaxWidth, Align: col.Align, Sortable: sortable, Fixed: col.Fixed}
	}
	s.sort(columns, o.ManualSorting)
	s.List.Key = func(i int) any { return s.RowID(i) }
	s.List.Selection = &s.Selection
	s.List.Sort = &s.Sort
	if o.OnActivate != nil && s.List.Selected == nil {
		if !s.activeInitialized {
			s.active, s.activeInitialized = -1, true
		}
		s.List.Selected = &s.active
	}
	s.List.Label = func(i int) string {
		if len(columns) > 0 && columns[0].Value != nil {
			return columns[0].Value(s.Row(i))
		}
		return tableIDLabel(s.RowID(i))
	}
	height := o.Height
	if height <= 0 {
		height = 320
	}
	s.CellBuilds = 0
	// Nuxt's th is highlighted and its td muted. The native table draws its
	// header in the theme's TextMuted and its cells in Text, so swap them here.
	theme := t.Native()
	theme.TextMuted, theme.Text = t.Highlighted, t.Muted
	// Native header and row padding share Spacing. Keep their minimum large
	// enough for scaled text, even when a compact density was requested.
	theme.Spacing = max(theme.Spacing, t.text(4))
	root := ui.Themed(c, theme).Background(ui.Transparent).MinWidth(0).MinHeight(0).Height(height).Disabled(o.Disabled || t.field.Disabled)
	events := ui.Local(root, "events", func() tableEvents[K] { return tableEvents[K]{} })
	finish(t, ComponentTable, Root, root, PartStyle{Radius: Set(t.radius(8)), BorderWidth: Set(float32(1)), BorderColor: Set(t.Border)}, o.Decoration)
	root.Children(func() {
		if o.BuildCaption != nil {
			caption := ui.Column(c.Key("caption")).MinWidth(0)
			finish(t, ComponentTable, CaptionSlot, caption, PartStyle{}, o.Decoration)
			caption.Children(func() { o.BuildCaption(c, t) })
		}
		body := ui.Box(c.Key("body")).Grow(1).MinWidth(0).MinHeight(0)
		finish(t, ComponentTable, Body, body, PartStyle{}, o.Decoration)
		body.Children(func() {
			table := ui.Table(c.Key("table"), &s.List, native, s.Len(), func(row, col int) {
				// Native section headers call column zero, even with no columns.
				if len(columns) == 0 {
					return
				}
				s.CellBuilds++
				value, id := s.Row(row), s.RowID(row)
				column := columns[col]
				chosen := s.Selection.Has(id)
				cell := ui.Row(c).MinWidth(0).AlignItems(ui.Center)
				finishTableCell(t, Item, cell, PartStyle{Gap: Set(t.spacing(6))}, o.Decoration, column.Decoration)
				cell.Children(func() {
					if column.BuildCell != nil {
						column.BuildCell(c, t, cell, TableCell[T]{Value: value, Index: row, Selected: chosen})
					} else if column.Cell != nil {
						column.Cell(c, t, value, chosen)
					} else if column.Value != nil {
						label := ui.Text(c, column.Value(value)).SingleLine()
						finishTableCell(t, Label, label, PartStyle{}, o.Decoration, column.Decoration)
					}
				})
			}).Fill().Label(o.Label).Description(o.Caption)
			tableHeaderDivider(c, t, table, theme.Space(8), o.Decoration)
			finish(t, ComponentTable, Panel, table, PartStyle{}, o.Decoration)
			// Changed distinguishes a native input notice from external model edits.
			// The mounted snapshot survives notice/rebuild passes without holding Parts.
			if !table.Changed() || !events.initialized {
				if !events.initialized || !slices.Equal(events.sort, s.applied) {
					events.sort = slices.Clone(s.applied)
				}
				// Every build compares, so that outside edits never count as the
				// user's; an unchanged choice reads it without copying it.
				if o.OnSelectionChange != nil && (!events.initialized || tableSelectionDiffers(events.selection, &s.Selection)) {
					events.selection = snapshotTableSelection(events.selection, &s.Selection)
				}
				events.initialized = true
			}
			table.OnChange(func() {
				previousSort := events.sort
				s.sort(columns, o.ManualSorting)
				events.sort = slices.Clone(s.applied)
				selectionChanged := false
				if o.OnSelectionChange != nil {
					selectionChanged = tableSelectionDiffers(events.selection, &s.Selection)
					if selectionChanged {
						events.selection = snapshotTableSelection(events.selection, &s.Selection)
					}
				}
				if o.OnSortChange != nil && !slices.Equal(previousSort, events.sort) {
					o.OnSortChange(slices.Clone(events.sort))
				}
				if selectionChanged {
					ids := slices.Clone(events.selection)
					slices.Sort(ids)
					o.OnSelectionChange(ids)
				}
				if o.OnChange != nil {
					o.OnChange()
				}
			})
			if o.OnActivate != nil && table.Submitted() && s.List.Selected != nil {
				index := *s.List.Selected
				if index >= 0 && index < s.Len() {
					id := s.RowID(index)
					table.OnSubmit(func() { o.OnActivate(id) })
				}
			}
			headerHeight := theme.Space(8) + 1
			if s.Len() == 0 {
				slot, build := EmptySlot, o.BuildEmpty
				if o.Loading && o.BuildLoading != nil {
					slot, build = LoadingSlot, o.BuildLoading
				}
				empty := ui.Column(c.Key("empty")).Absolute().Left(0).Right(0).Top(headerHeight).Bottom(0).MinWidth(0).Center().PassThrough().Role(ui.RoleStatus)
				finish(t, ComponentTable, slot, empty, PartStyle{Foreground: Set(t.Muted), PaddingX: Set(t.spacing(16)), PaddingY: Set(t.spacing(24))}, o.Decoration)
				empty.Children(func() {
					if build != nil {
						build(c, t)
					} else {
						text := o.EmptyText
						if text == "" {
							text = emptyMessage(c, "noData", "noRecords")
						}
						ui.Text(c, text).MinWidth(0)
					}
				})
			}
			if o.Loading {
				tableLoading(c.Key("loading"), t, headerHeight-1, o)
			}
		})
	})
	return root
}
