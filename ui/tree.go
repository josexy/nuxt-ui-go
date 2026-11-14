package nui

import (
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
)

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

// TreeEvent describes the state before an interaction. A node handler overrides
// the corresponding TreeOptions handler. PreventDefault cancels that action only.
type TreeEvent struct {
	ID                 string
	Selected, Expanded bool
	prevented          bool
}

func (e *TreeEvent) PreventDefault() { e.prevented = true }

// TreeItemState is a snapshot for synchronous, non-interactive row content.
type TreeItemState struct {
	ID, Label                                  string
	Level                                      int
	Selected, Expanded, Indeterminate, Loading bool
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
	parent   string
	search   string
}
type treeRow struct {
	id    string
	depth int
}
type TreeState struct {
	identity  *byte
	Outline   ui.OutlineState[string]
	Selection ui.Selection[string]
	roots     []string
	nodes     map[string]*treeRecord
	loading   map[string]*treeRecord
	version   uint64
	RowBuilds int
	// Active is the keyboard highlight, independent of selection.
	Active         string
	rows           []treeRow
	positions      map[string]int
	open           map[string]bool
	dirty          bool
	anchor, prefix string
	typed          time.Time
	selected       map[string]bool
	partial        map[string]bool
	summaryDirty   bool
	partialMode    uint8
	keyboard       bool
}

func (s *TreeState) SetNodes(nodes []TreeNode) {
	if s.identity == nil {
		s.identity = new(byte)
	}
	records := map[string]*treeRecord{}
	var add func([]TreeNode, string) []string
	add = func(items []TreeNode, parent string) []string {
		if items == nil {
			return nil
		}
		ids := make([]string, 0, len(items))
		for _, n := range items {
			if n.ID == "" || records[n.ID] != nil {
				panic("nui: Tree needs unique non-empty IDs")
			}
			r := &treeRecord{node: n, parent: parent, search: strings.ToLower(n.Label), loaded: !n.Lazy || n.Children != nil}
			r.node.Children = nil // Own the indexed model, not the caller's slices.
			records[n.ID] = r
			r.children = add(n.Children, n.ID)
			if n.Lazy && r.children == nil {
				r.children = []string{}
			}
			ids = append(ids, n.ID)
		}
		return ids
	}
	roots := add(nodes, "")
	for id, n := range records {
		if s.nodes[id] == nil && n.node.DefaultExpanded && n.children != nil {
			s.Outline.Open.Add(id)
		}
	}
	s.version++
	s.nodes = records
	s.roots = roots
	s.loading = nil
	s.dirty, s.summaryDirty = true, true
	for id := range s.Selection.All() {
		if records[id] == nil {
			s.Selection.Remove(id)
		}
	}
	for id := range s.Outline.Open.All() {
		if records[id] == nil || records[id].children == nil {
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
	if n := s.nodes[id]; n != nil && !n.loaded && n.err != "" {
		n.err = ""
		n.loaded = false
		n.loading = false
		s.version++
		n.version = s.version
		delete(s.loading, id)
	}
}

// CompleteLoad must be called on the window's UI thread. Collapsed/replaced requests are rejected.
func (s *TreeState) CompleteLoad(ticket TreeTicket, children []TreeNode, err error) bool {
	r := s.nodes[ticket.id]
	if ticket.owner != s.identity || r == nil || !r.loading || r.version != ticket.version || !s.expandedPath(ticket.id) {
		return false
	}
	r.loading = false
	delete(s.loading, ticket.id)
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
	for _, id := range staged.roots {
		s.nodes[id].parent = ticket.id
	}
	for id := range staged.Outline.Open.All() {
		s.Outline.Open.Add(id)
	}
	r.children = staged.roots
	if r.children == nil {
		r.children = []string{}
	}
	r.loaded = true
	r.err = ""
	s.dirty, s.summaryDirty = true, true
	return true
}

// SetExpanded changes expansion without emitting interaction callbacks. Unlike
// direct Open edits, it invalidates pending loads immediately, even if reopened
// before the next frame. All state methods must run on the window's UI thread.
func (s *TreeState) SetExpanded(id string, expanded bool) {
	n := s.nodes[id]
	if n == nil || n.children == nil || s.Outline.Open.Has(id) == expanded {
		return
	}
	if expanded {
		s.Outline.Open.Add(id)
	} else {
		s.Outline.Open.Remove(id)
	}
	s.dirty = true
	s.cancelHiddenLoads()
}

// Close invalidates pending loads when the owning view is disposed.
func (s *TreeState) Close() {
	for id, n := range s.loading {
		n.loading = false
		delete(s.loading, id)
	}
}

func (s *TreeState) expandedPath(id string) bool {
	for id != "" {
		n := s.nodes[id]
		if n == nil || !s.Outline.Open.Has(id) {
			return false
		}
		id = n.parent
	}
	return true
}

func (s *TreeState) cancelHiddenLoads() {
	for id, n := range s.loading {
		if !s.expandedPath(id) {
			n.loading = false
			delete(s.loading, id)
		}
	}
}

func (s *TreeState) visibleRows() {
	changed := s.dirty || len(s.open) != s.Outline.Open.Len()
	if !changed {
		for id := range s.Outline.Open.All() {
			if !s.open[id] {
				changed = true
				break
			}
		}
	}
	if !changed {
		return
	}
	s.dirty = false
	if s.open == nil {
		s.open = map[string]bool{}
	}
	clear(s.open)
	for id := range s.Outline.Open.All() {
		s.open[id] = true
	}
	if s.positions == nil {
		s.positions = map[string]int{}
	}
	clear(s.positions)
	s.rows = s.rows[:0]
	var visit func([]string, int)
	visit = func(ids []string, depth int) {
		for _, id := range ids {
			s.positions[id] = len(s.rows)
			s.rows = append(s.rows, treeRow{id, depth})
			if s.open[id] {
				visit(s.nodes[id].children, depth+1)
			}
		}
	}
	visit(s.roots, 0)
	// A collapsed ancestor receives the highlight; selection stays by ID.
	for id := s.Active; id != ""; {
		n := s.nodes[id]
		if n == nil {
			break
		}
		if _, ok := s.positions[id]; ok && !n.node.Disabled {
			s.Active = id
			return
		}
		id = n.parent
	}
	s.Active = ""
}

// Rows, Item and Depth describe the last built visible model. Outline.Open and
// Outline.List remain the persistent expansion and scrolling state; the legacy
// Outline.Rows/Item/Depth methods are no longer populated.
func (s *TreeState) Rows() int           { return len(s.rows) }
func (s *TreeState) Item(row int) string { return s.rows[row].id }
func (s *TreeState) Depth(row int) int   { return s.rows[row].depth }

func (s *TreeState) selectionSummary(o TreeOptions) {
	var mode uint8
	if o.Multiple {
		if o.BubbleSelect {
			mode = 1
		} else if o.PropagateSelect {
			mode = 2
		}
	}
	changed := s.summaryDirty || s.partialMode != mode || len(s.selected) != s.Selection.Len()
	if !changed {
		for id := range s.Selection.All() {
			if !s.selected[id] {
				changed = true
				break
			}
		}
	}
	if !changed {
		return
	}
	s.summaryDirty = false
	s.partialMode = mode
	if s.selected == nil {
		s.selected, s.partial = map[string]bool{}, map[string]bool{}
	}
	clear(s.selected)
	clear(s.partial)
	for id := range s.Selection.All() {
		s.selected[id] = true
	}
	var visit func(string) (int, int)
	visit = func(id string) (int, int) {
		n := s.nodes[id]
		if n.node.Disabled {
			return 0, 0
		}
		total, chosen := 0, 0
		for _, child := range n.children {
			a, b := visit(child)
			total += a
			chosen += b
		}
		// Bubbling describes mixed descendants; propagation describes an
		// explicitly selected parent whose descendants are not all selected.
		s.partial[id] = (mode == 1 && chosen > 0 || mode == 2 && s.Selection.Has(id)) && chosen < total
		if s.Selection.Has(id) {
			chosen++
		}
		return total + 1, chosen
	}
	for _, id := range s.roots {
		visit(id)
	}
}

func (s *TreeState) next(at, delta int, loop bool) int {
	for step := 1; step <= len(s.rows); step++ {
		i := at + step*delta
		if loop {
			i = (i%len(s.rows) + len(s.rows)) % len(s.rows)
		}
		if i < 0 || i >= len(s.rows) {
			break
		}
		if !s.nodes[s.rows[i].id].node.Disabled {
			return i
		}
	}
	return at
}

func (s *TreeState) changeSelection(id string, on bool, propagate bool) bool {
	n := s.nodes[id]
	if n == nil || n.node.Disabled {
		return false
	}
	changed := s.Selection.Has(id) != on
	if on {
		s.Selection.Add(id)
	} else {
		s.Selection.Remove(id)
	}
	if propagate {
		for _, child := range n.children {
			changed = s.changeSelection(child, on, true) || changed
		}
	}
	return changed
}

// selectRows visits overlapping subtrees once. A range or Select All can include
// both a parent and its descendants; recursively selecting every row separately
// would make a deeply nested, expanded tree quadratic.
func (s *TreeState) selectRows(rows []treeRow, propagate bool) bool {
	var visited map[string]bool
	if propagate {
		visited = make(map[string]bool)
	}
	var selectOne func(string) bool
	selectOne = func(id string) bool {
		n := s.nodes[id]
		if n.node.Disabled || visited[id] {
			return false
		}
		changed := s.changeSelection(id, true, false)
		if propagate {
			visited[id] = true
			for _, child := range n.children {
				changed = selectOne(child) || changed
			}
		}
		return changed
	}
	changed := false
	for _, row := range rows {
		changed = selectOne(row.id) || changed
	}
	return changed
}

func (s *TreeState) bubble() bool {
	changed := false
	var visit func(string)
	visit = func(id string) {
		n := s.nodes[id]
		if n.node.Disabled || len(n.children) == 0 {
			return
		}
		all, enabled := true, false
		for _, child := range n.children {
			if s.nodes[child].node.Disabled {
				continue
			}
			visit(child)
			enabled = true
			all = all && s.Selection.Has(child)
		}
		if enabled {
			changed = s.changeSelection(id, all, false) || changed
		}
	}
	for _, id := range s.roots {
		visit(id)
	}
	return changed
}

func (s *TreeState) bubbleAncestors(id string) bool {
	changed := false
	for id = s.nodes[id].parent; id != ""; id = s.nodes[id].parent {
		n := s.nodes[id]
		if n.node.Disabled {
			break
		}
		all, enabled := true, false
		for _, child := range n.children {
			if s.nodes[child].node.Disabled {
				continue
			}
			enabled = true
			all = all && s.Selection.Has(child)
		}
		if enabled {
			changed = s.changeSelection(id, all, false) || changed
		}
	}
	return changed
}
