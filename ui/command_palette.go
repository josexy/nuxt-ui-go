package nui

import (
	"container/heap"
	"context"
	"maps"
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/egoist/mygo/ui"
)

type CommandAction struct {
	Key                                               ui.Key
	Mods                                              ui.Modifiers
	ID, Label, Description, Group, Keywords, Shortcut string
	Prefix, Suffix, Placeholder                       string
	Kbds                                              []string
	Icon                                              *ui.SVG
	Avatar                                            *AvatarOptions
	Chip                                              *ChipOptions
	Loading, Active                                   bool
	Disabled                                          bool
	Children                                          []CommandAction
	OnSelect                                          func()
	// OnSelectEvent can cancel a leaf activation before closing or callbacks.
	OnSelectEvent func(*CommandEvent)
	Decoration
}

type CommandEvent struct {
	ID        string
	prevented bool
}

func (e *CommandEvent) PreventDefault() { e.prevented = true }

type indexedCommand struct {
	action                               CommandAction
	label, description, suffix, keywords []rune
	order                                int
}

// CommandIndex is immutable after construction and safe to search on a worker.
// Rebuild it when actions change, outside rendering. Callbacks run only on activation.
type CommandIndex struct {
	levels    map[string][]indexedCommand
	shortcuts []CommandAction
	count     int
}

func NewCommandIndex(actions []CommandAction) *CommandIndex {
	x := &CommandIndex{levels: map[string][]indexedCommand{}}
	seen := map[string]bool{}
	var add func(string, []CommandAction, int, bool) []CommandAction
	add = func(parent string, items []CommandAction, depth int, disabled bool) []CommandAction {
		if depth > 32 {
			panic("nui: CommandPalette nesting exceeds 32")
		}
		if items == nil {
			return nil
		}
		owned := make([]CommandAction, len(items))
		for i, a := range items {
			checkID(seen, a.ID, "CommandPalette")
			order := x.count
			x.count++
			a.Disabled = a.Disabled || disabled
			a.Kbds = slices.Clone(a.Kbds)
			a.UI = maps.Clone(a.UI)
			if a.Avatar != nil {
				v := *a.Avatar
				v.UI = maps.Clone(v.UI)
				a.Avatar = &v
			}
			if a.Chip != nil {
				v := *a.Chip
				v.UI = maps.Clone(v.UI)
				a.Chip = &v
			}
			a.Children = add(a.ID, a.Children, depth+1, a.Disabled)
			owned[i] = a
			x.levels[parent] = append(x.levels[parent], indexedCommand{action: a, label: []rune(strings.ToLower(a.Label)), description: []rune(strings.ToLower(a.Description)), suffix: []rune(strings.ToLower(a.Suffix)), keywords: []rune(strings.ToLower(a.Keywords)), order: order})
			if a.Key != ui.KeyUnknown {
				x.shortcuts = append(x.shortcuts, a)
			}
		}
		return owned
	}
	add("", actions, 0, false)
	return x
}

type CommandMatch struct {
	Action                                  CommandAction
	Highlights                              []int
	DescriptionHighlights, SuffixHighlights []int
	Score                                   int
	order                                   int
}
type commandHeap []CommandMatch

func (h commandHeap) Len() int { return len(h) }
func (h commandHeap) Less(i, j int) bool {
	if h[i].Score == h[j].Score {
		return h[i].order > h[j].order
	}
	return h[i].Score < h[j].Score
}
func (h commandHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *commandHeap) Push(v any)   { *h = append(*h, v.(CommandMatch)) }
func (h *commandHeap) Pop() any     { a := *h; v := a[len(a)-1]; *h = a[:len(a)-1]; return v }
func fuzzy(label, query []rune, scratch []int) ([]int, int, bool) {
	positions := scratch[:0]
	at := 0
	score := 0
	previous := -2
	for _, q := range query {
		found := false
		for at < len(label) {
			i := at
			at++
			if label[i] != q {
				continue
			}
			score += 10
			if i == previous+1 {
				score += 8
			}
			if i == 0 || unicode.IsSpace(label[i-1]) || label[i-1] == '-' {
				score += 12
			}
			score -= i / 4
			positions = append(positions, i)
			previous = i
			found = true
			break
		}
		if !found {
			return nil, 0, false
		}
	}
	return positions, score, true
}

// Search checks cancellation while scanning and retains at most limit matches.
// Highlight offsets are Unicode rune indexes, not UTF-8 byte offsets.
func (x *CommandIndex) Search(ctx context.Context, query, parent string, limit int) []CommandMatch {
	return x.search(ctx, query, parent, limit, commandDefaultFields, false)
}

func (x *CommandIndex) search(ctx context.Context, query, parent string, limit int, fields commandFields, ignore bool) []CommandMatch {
	if x == nil {
		return nil
	}
	limit = commandLimit(limit)
	q := []rune(strings.ToLower(strings.TrimSpace(query)))
	h := commandHeap{}
	var scratch [4][]int
	for i := range scratch {
		scratch[i] = make([]int, 0, len(q))
	}
	for i, it := range x.levels[parent] {
		if i%128 == 0 {
			select {
			case <-ctx.Done():
				return nil
			default:
			}
		}
		marks, score, ok := matchCommand(it, q, scratch, fields, ignore)
		if !ok {
			continue
		}
		if len(h) >= limit && score <= h[0].Score {
			continue
		}
		m := CommandMatch{Action: it.action, Highlights: slices.Clone(marks[0]), DescriptionHighlights: slices.Clone(marks[1]), SuffixHighlights: slices.Clone(marks[2]), Score: score, order: it.order}
		if len(h) < limit {
			heap.Push(&h, m)
		} else if score > h[0].Score {
			h[0] = m
			heap.Fix(&h, 0)
		}
		if len(q) == 0 && len(h) == limit {
			break
		}
	}
	result := []CommandMatch(h)
	slices.SortStableFunc(result, func(a, b CommandMatch) int {
		if a.Score != b.Score {
			return b.Score - a.Score
		}
		return a.order - b.order
	})
	return result
}

type commandResult struct {
	ticket QueryTicket
	items  []CommandMatch
}
type commandLevel struct{ id, label, placeholder string }
type CommandPaletteState struct {
	Open                                   bool
	Query                                  string
	Results                                QueryState[CommandMatch]
	index                                  *CommandIndex
	level                                  []commandLevel
	lastQuery, lastParent                  string
	lastLimit                              int
	lastFields                             commandFields
	lastIgnore, lastExternal, lastDisabled bool
	lastDelay                              time.Duration
	waiting                                bool
	readyAt                                time.Time
	dirty, wasOpen                         bool
	mailbox                                chan commandResult
	list                                   ListboxState
	editor                                 ui.Handle
	selected, previous                     []string
	rows                                   commandRowsCache
}

func (s *CommandPaletteState) SetIndex(index *CommandIndex) {
	s.Refresh()
	s.index = index
	s.dirty = true
	s.level = nil
	s.Query = ""
	s.Results.Items = nil
	s.mailbox = nil
}

// Refresh cancels pending work and searches again on the next build. Use it
// after replacing a custom Search closure without changing Query or SetIndex.
func (s *CommandPaletteState) Refresh() {
	s.Results.Cancel()
	s.Results.Items = nil
	s.mailbox = nil
	s.waiting = false
	s.dirty = true
}

// Searching includes the debounce interval before a worker request starts.
func (s *CommandPaletteState) Searching() bool { return s.waiting || s.Results.Loading }
func (s *CommandPaletteState) Close() {
	s.Open, s.wasOpen = false, false
	s.Results.Cancel()
	s.mailbox = nil
	s.dirty = true
	s.waiting = false
	s.editor.CancelFocus()
	s.list.List.CancelFocus()
}
func (s *CommandPaletteState) parent() string {
	if len(s.level) > 0 {
		return s.level[len(s.level)-1].id
	}
	return ""
}
func (s *CommandPaletteState) Back() {
	if len(s.level) > 0 {
		s.level = s.level[:len(s.level)-1]
		s.Query = ""
		s.Refresh()
		s.editor.Focus()
	}
}

type CommandPaletteOptions struct {
	Title, Placeholder                                      string
	Limit                                                   int
	ShortcutKey                                             ui.Key
	ShortcutMods                                            ui.Modifiers
	DisableShortcut                                         bool
	Disabled                                                bool
	Size                                                    Size
	SearchDelay                                             time.Duration
	FilterFields                                            []string
	IgnoreFilter, PreserveGroupOrder                        bool
	Autofocus, SearchInput, Back                            Optional[bool]
	SearchInputOptions                                      *InputOptions
	Icon, ChildrenIcon, BackIcon, LoadingIcon, SelectedIcon *ui.SVG
	ShowIcon                                                Optional[bool]
	Loading, Loop                                           bool
	HighlightOnHover                                        Optional[bool]
	BackOptions                                             *ButtonOptions
	EmptyText                                               string
	BuildEmpty                                              func(*ui.Context, Theme, string)
	OnSearchTermChange, OnHighlight                         func(string)
	OnSelectEvent                                           func(*CommandEvent)
	CloseOnSelect                                           Optional[bool]
	// Search starts application work; publish Results.Complete on the UI thread.
	// Nil uses the immutable index on a worker and a bounded mailbox.
	Search   func(context.Context, QueryTicket, string, string)
	OnSelect func(string)
	Modal    ModalOptions
	Decoration
}

func CommandPalette(c *ui.Context, t Theme, s *CommandPaletteState, o CommandPaletteOptions) ui.Element {
	if s == nil {
		panic("nui: CommandPalette requires state")
	}
	host := ui.Box(c).Absolute().Size(0, 0).IgnoreInput(true)
	if o.SearchDelay < 0 {
		panic("nui: CommandPalette SearchDelay must be non-negative")
	}
	host.Disabled(o.Disabled || t.field.Disabled)
	disabled := host.IsDisabled()
	fields := commandFieldMask(o.FilterFields)
	config := selected(defaults(t, ComponentCommandPalette, ComponentTheme{Size: MD}), ComponentTheme{Size: o.Size})
	identity := ui.Local(host, "command", func() byte { return 0 })
	shown := ui.Local(host, "nui-command-status", func() shownMessages { return shownMessages{} })
	if !disabled && !o.DisableShortcut {
		key, mods := o.ShortcutKey, o.ShortcutMods
		if key == ui.KeyUnknown {
			key = ui.KeyK
			mods = ui.Ctrl
		}
		c.OnShortcut(mods, key, func() {
			if s.Open {
				s.Close()
			} else {
				s.Open = true
			}
		})
	}
	if s.index != nil && !disabled {
		for _, a := range s.index.shortcuts {
			if a.Disabled || len(a.Children) > 0 {
				continue
			}
			c.OnShortcut(a.Mods, a.Key, func() { activateCommand(s, a, o) })
		}
	}
	if !s.Open {
		if s.wasOpen {
			s.Close()
		}
		s.wasOpen = false
		// Messages showing again in a reopened palette are news again.
		*shown = shownMessages{}
		// Keep the modal host mounted for its explicit ReturnFocus transition.
		ModalWithOptions(c.Key(identity), t, &s.Open, o.Modal, nil)
		return host
	}
	if !s.wasOpen {
		if !disabled && valueOr(o.Autofocus, true) {
			if valueOr(o.SearchInput, true) {
				s.editor.Focus(c)
			} else {
				s.list.List.Handle.Focus(c)
			}
		}
		s.dirty = true
	}
	s.wasOpen = true
	if disabled != s.lastDisabled {
		if disabled {
			s.Results.Cancel()
			s.mailbox = nil
			s.waiting = false
		} else {
			s.dirty = true
		}
		s.lastDisabled = disabled
	}
	limit := commandLimit(o.Limit)
	wasWaiting := s.waiting
	if !disabled && (s.dirty || s.Query != s.lastQuery || s.parent() != s.lastParent || limit != s.lastLimit || fields != s.lastFields || o.IgnoreFilter != s.lastIgnore || (o.Search != nil) != s.lastExternal) {
		debounce := !s.dirty && s.parent() == s.lastParent && s.Query != s.lastQuery
		s.Results.Cancel()
		s.Results.Items = nil
		s.Results.Error = ""
		s.mailbox = nil
		s.waiting = true
		s.readyAt = c.Now()
		if debounce {
			s.readyAt = s.readyAt.Add(o.SearchDelay)
		}
		s.dirty = false
		s.lastQuery, s.lastParent = s.Query, s.parent()
		s.lastLimit = limit
		s.lastFields, s.lastIgnore, s.lastExternal = fields, o.IgnoreFilter, o.Search != nil
		s.selected = nil
		s.list.Active = ""
		s.list.List.ScrollTo(0, ui.Start)
	}
	if s.waiting && wasWaiting && s.lastDelay != o.SearchDelay {
		s.readyAt = c.Now().Add(o.SearchDelay)
	}
	s.lastDelay = o.SearchDelay
	if s.waiting && !disabled && !c.Now().Before(s.readyAt) {
		s.waiting = false
		ctx, ticket := s.Results.Start(context.Background(), s.Query)
		if o.Search != nil {
			o.Search(ctx, ticket, s.Query, s.parent())
		} else {
			index, query, parent, limit, ignore := s.index, s.Query, s.parent(), limit, o.IgnoreFilter
			mailbox := make(chan commandResult, 1)
			s.mailbox = mailbox
			services := c.Services()
			go func() {
				items := index.search(ctx, query, parent, limit, fields, ignore)
				select {
				case <-ctx.Done():
					return
				case mailbox <- commandResult{ticket, items}:
					// The next build takes the result; nothing polls for it.
					services.Invalidate()
				}
			}()
		}
	}
	if s.waiting {
		c.After(s.readyAt.Sub(c.Now()))
	}
	if s.mailbox != nil {
		select {
		case r := <-s.mailbox:
			s.Results.Complete(r.ticket, r.items, nil)
			s.mailbox = nil
		default:
		}
	}
	title := o.Title
	if title == "" {
		title = message(c, "commandPalette")
	}
	activate := func(id string) bool {
		if disabled || s.waiting || s.Results.Loading || s.Query != s.lastQuery || s.parent() != s.lastParent {
			return false
		}
		for _, m := range s.Results.Items {
			a := m.Action
			if a.ID != id || a.Disabled {
				continue
			}
			if len(a.Children) > 0 {
				s.level = append(s.level, commandLevel{id, a.Label, a.Placeholder})
				before := s.Query
				s.Query = ""
				s.Refresh()
				if before != "" && o.OnSearchTermChange != nil {
					o.OnSearchTermChange("")
				}
				s.editor.Focus(c)
				return true
			}
			return activateCommand(s, a, o)
		}
		return false
	}
	mo := o.Modal
	mo.Title = title
	onClose := mo.OnClose
	mo.OnClose = func() {
		s.Close()
		if onClose != nil {
			onClose()
		}
	}
	panel := ModalWithOptions(c.Key(identity), t, &s.Open, mo, func(t Theme) {
		commandInput(c, t, s, o, title, disabled, config.Size, activate)
		searching := message(c, "searching")
		if announceShown(c, &shown.progress, s.Results.Loading || s.waiting, searching) {
			statusText(c, "status-searching", searching).TextColor(t.Muted)
		}
		if announceShown(c, &shown.err, s.Results.Error != "", s.Results.Error) {
			statusText(c, "status-error", s.Results.Error).TextColor(t.Error)
		}
		commandRows(c, t, s, o, title, disabled, config.Size, activate)
	})
	if panel.Valid() {
		finish(t, ComponentCommandPalette, Root, panel, PartStyle{}, o.Decoration)
	}
	return panel
}

func commandLimit(n int) int {
	if n <= 0 {
		return 200
	}
	return min(n, 1000)
}
