package nui

import (
	"maps"
	"slices"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
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
	root := ui.Themed(c, t.Native()).Background(ui.Transparent).MinWidth(0).MinHeight(0).Height(height).Disabled(o.Disabled)
	finish(t, ComponentTable, Root, root, PartStyle{Radius: Set(t.radius(8)), BorderWidth: Set(float32(1)), BorderColor: Set(t.Border)}, o.Decoration)
	root.Children(func() {
		if len(s.rows) == 0 {
			text := o.EmptyText
			if text == "" {
				text = message(c, "noRecords")
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

type TreeNode struct {
	ID, Label string
	Icon      *ui.SVG
	Children  []TreeNode
	Lazy      bool
}
type TreeTicket struct {
	owner   *byte
	id      string
	version uint64
}
type treeRecord struct {
	node     TreeNode
	children []string
	loaded   bool
	loading  bool
	err      string
	version  uint64
}
type TreeState struct {
	identity  *byte
	Outline   ui.OutlineState[string]
	Selection ui.Selection[string]
	roots     []string
	nodes     map[string]*treeRecord
	version   uint64
	RowBuilds int
}

func (s *TreeState) SetNodes(nodes []TreeNode) {
	if s.identity == nil {
		s.identity = new(byte)
	}
	records := map[string]*treeRecord{}
	var add func([]TreeNode) []string
	add = func(items []TreeNode) []string {
		if items == nil {
			return nil
		}
		ids := make([]string, 0, len(items))
		for _, n := range items {
			if n.ID == "" || records[n.ID] != nil {
				panic("nui: Tree needs unique non-empty IDs")
			}
			r := &treeRecord{node: n, loaded: !n.Lazy || n.Children != nil}
			records[n.ID] = r
			r.children = add(n.Children)
			if n.Lazy && r.children == nil {
				r.children = []string{}
			}
			ids = append(ids, n.ID)
		}
		return ids
	}
	roots := add(nodes)
	s.version++
	s.nodes = records
	s.roots = roots
	for id := range s.Selection.All() {
		if records[id] == nil {
			s.Selection.Remove(id)
		}
	}
	for id := range s.Outline.Open.All() {
		if records[id] == nil {
			s.Outline.Open.Remove(id)
		}
	}
}
func (s *TreeState) Roots() []string { return slices.Clone(s.roots) }
func (s *TreeState) Error(id string) string {
	if n := s.nodes[id]; n != nil {
		return n.err
	}
	return ""
}
func (s *TreeState) Retry(id string) {
	if n := s.nodes[id]; n != nil {
		n.err = ""
		n.loaded = false
		n.loading = false
		n.version++
	}
}

// CompleteLoad must be called on the window's UI thread. Collapsed/replaced requests are rejected.
func (s *TreeState) CompleteLoad(ticket TreeTicket, children []TreeNode, err error) bool {
	r := s.nodes[ticket.id]
	if ticket.owner != s.identity || r == nil || !r.loading || r.version != ticket.version {
		return false
	}
	r.loading = false
	if err != nil {
		r.err = err.Error()
		return true
	}
	var staged TreeState
	staged.SetNodes(children)
	for id := range staged.nodes {
		if s.nodes[id] != nil {
			panic("nui: duplicate lazy tree ID")
		}
	}
	maps.Copy(s.nodes, staged.nodes)
	r.children = staged.roots
	if r.children == nil {
		r.children = []string{}
	}
	r.loaded = true
	r.err = ""
	return true
}

type TreeOptions struct {
	Label, EmptyText string
	Height           float32
	Disabled         bool
	OnChange         func()
	OnLoad           func(string, TreeTicket)
	Decoration
}

func Tree(c *ui.Context, t Theme, s *TreeState, o TreeOptions) ui.Element {
	if s == nil {
		panic("nui: Tree requires state")
	}
	s.Outline.List.Selection = &s.Selection
	s.Outline.Indent = t.spacing(16)
	s.Outline.List.Label = func(i int) string { return s.nodes[s.Outline.Item(i)].node.Label }
	// A request is issued once per expansion, and invalidated when its branch closes.
	for id, n := range s.nodes {
		open := s.Outline.Open.Has(id)
		if n.loading && !open {
			n.loading = false
			n.version++
			s.version = max(s.version, n.version)
		}
		if open && !n.loaded && !n.loading && n.err == "" && o.OnLoad != nil {
			s.version++
			n.version = s.version
			n.loading = true
			o.OnLoad(id, TreeTicket{s.identity, id, n.version})
		}
	}
	height := o.Height
	if height <= 0 {
		height = 320
	}
	s.RowBuilds = 0
	root := ui.Themed(c, t.Native()).Background(ui.Transparent).MinWidth(0).MinHeight(0).Height(height).Disabled(o.Disabled)
	finish(t, ComponentTree, Root, root, PartStyle{Radius: Set(t.radius(8)), BorderWidth: Set(float32(1)), BorderColor: Set(t.Border)}, o.Decoration)
	root.Children(func() {
		if len(s.roots) == 0 {
			text := o.EmptyText
			if text == "" {
				text = message(c, "noItems")
			}
			finish(t, ComponentTree, EmptySlot, ui.Text(c, text).Padding(12).Role(ui.RoleStatus), PartStyle{Foreground: Set(t.Muted)}, o.Decoration)
		}
		ui.Outline(c.Key("tree"), &s.Outline, s.roots, func(id string) []string { return s.nodes[id].children }, func(id string) {
			s.RowBuilds++
			n := s.nodes[id]
			row := ui.Row(c).MinWidth(0).AlignItems(ui.Center)
			finish(t, ComponentTree, Item, row, PartStyle{Gap: Set(t.spacing(8))}, o.Decoration)
			row.Children(func() {
				icon := n.node.Icon
				if icon == nil && n.children != nil {
					icon = lucide.FolderOpen()
				}
				if icon != nil {
					ui.Icon(c, icon).Size(16, 16)
				}
				ui.Text(c, n.node.Label).SingleLine()
				if n.loading {
					ui.Text(c, message(c, "loading")).FontSize(12).Role(ui.RoleStatus)
				}
				if n.err != "" {
					Button(c.Key("retry"), t, message(c, "retry"), ButtonOptions{AccessibleLabel: namedMessage(c, "retryNamed", n.node.Label), Variant: Ghost, Size: XS}).Description(n.err).OnClick(func() { s.Retry(id) })
				}
			})
		}).Fill().Label(o.Label).OnChange(o.OnChange)
	})
	return root
}
