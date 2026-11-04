package nui

import (
	"slices"

	"github.com/egoist/mygo/ui"
)

type TableColumn[T any] struct {
	ID, Title                 string
	Width, MinWidth, MaxWidth float32
	Align                     ui.Align
	Fixed                     bool // Prevents resizing/reordering; it does not freeze a scrolling column.
	Value                     func(T) string
	Compare                   func(T, T) int
	Cell                      func(*ui.Context, Theme, T, bool)
}
type tableRecord[T any] struct {
	id    string
	value T
}
type TableState[T any] struct {
	List                     ui.ListState
	Selection                ui.Selection[string]
	Sort                     ui.SortOrder
	rows                     []tableRecord[T]
	applied                  ui.SortOrder
	revision, sortedRevision uint64
	CellBuilds               int
}

// SetRows copies row values and validates identities once, outside rendering.
// Removed rows are removed from selection; retain the full data separately for filtering.
func (s *TableState[T]) SetRows(rows []T, key func(T) string) {
	if key == nil {
		panic("nui: Table requires row IDs")
	}
	seen := map[string]bool{}
	next := make([]tableRecord[T], len(rows))
	for i, v := range rows {
		id := key(v)
		checkID(seen, id, "Table rows")
		next[i] = tableRecord[T]{id, v}
	}
	s.rows = next
	s.revision++
	for id := range s.Selection.All() {
		if !seen[id] {
			s.Selection.Remove(id)
		}
	}
}
func (s *TableState[T]) Len() int           { return len(s.rows) }
func (s *TableState[T]) Row(i int) T        { return s.rows[i].value }
func (s *TableState[T]) RowID(i int) string { return s.rows[i].id }
func (s *TableState[T]) sort(columns []TableColumn[T]) {
	if s.applied == s.Sort && s.sortedRevision == s.revision {
		return
	}
	s.applied = s.Sort
	s.sortedRevision = s.revision
	for _, col := range columns {
		if col.ID == s.Sort.Column && col.Compare != nil {
			slices.SortStableFunc(s.rows, func(a, b tableRecord[T]) int {
				v := col.Compare(a.value, b.value)
				if s.Sort.Descending {
					return -v
				}
				return v
			})
			break
		}
	}
}

type TableOptions struct {
	Label, EmptyText string
	Height           float32
	Disabled         bool
	OnChange         func()
	Decoration
}

func Table[T any](c *ui.Context, t Theme, s *TableState[T], columns []TableColumn[T], o TableOptions) ui.Element {
	if s == nil {
		panic("nui: Table requires state")
	}
	seen := map[string]bool{}
	native := make([]ui.TableColumn, len(columns))
	for i, col := range columns {
		checkID(seen, col.ID, "Table columns")
		native[i] = ui.TableColumn{ID: col.ID, Title: col.Title, Width: col.Width, MinWidth: col.MinWidth, MaxWidth: col.MaxWidth, Align: col.Align, Sortable: col.Compare != nil, Fixed: col.Fixed}
	}
	s.sort(columns)
	s.List.Key = func(i int) any { return s.rows[i].id }
	s.List.Selection = &s.Selection
	s.List.Sort = &s.Sort
	s.List.Label = func(i int) string {
		if len(columns) > 0 && columns[0].Value != nil {
			return columns[0].Value(s.rows[i].value)
		}
		return s.rows[i].id
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
	root := ui.Themed(c, theme).Background(ui.Transparent).MinWidth(0).MinHeight(0).Height(height).Disabled(o.Disabled)
	finish(t, ComponentTable, Root, root, PartStyle{Radius: Set(t.radius(8)), BorderWidth: Set(float32(1)), BorderColor: Set(t.Border)}, o.Decoration)
	root.Children(func() {
		if len(s.rows) == 0 {
			text := o.EmptyText
			if text == "" {
				text = emptyMessage(c, "noData", "noRecords")
			}
			finish(t, ComponentTable, EmptySlot, ui.Text(c, text).Padding(16).Role(ui.RoleStatus), PartStyle{Foreground: Set(t.Muted)}, o.Decoration)
		}
		ui.Table(c.Key("table"), &s.List, native, len(s.rows), func(row, col int) {
			s.CellBuilds++
			record := s.rows[row]
			column := columns[col]
			chosen := s.Selection.Has(record.id)
			cell := ui.Row(c).MinWidth(0).AlignItems(ui.Center)
			finish(t, ComponentTable, Item, cell, PartStyle{Gap: Set(t.spacing(6))}, o.Decoration)
			cell.Children(func() {
				if column.Cell != nil {
					column.Cell(c, t, record.value, chosen)
				} else if column.Value != nil {
					ui.Text(c, column.Value(record.value)).SingleLine()
				}
			})
		}).Fill().Label(o.Label).OnChange(func() {
			s.sort(columns)
			if o.OnChange != nil {
				o.OnChange()
			}
		})
	})
	return root
}
