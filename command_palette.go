package nui

import (
	"container/heap"
	"context"
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

type CommandAction struct {
	Key                                               ui.Key
	Mods                                              ui.Modifiers
	ID, Label, Description, Group, Keywords, Shortcut string
	Icon                                              *ui.SVG
	Disabled                                          bool
	Children                                          []CommandAction
	OnSelect                                          func()
}
type indexedCommand struct {
	action          CommandAction
	label, keywords []rune
	order           int
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
	var add func(string, []CommandAction, int)
	add = func(parent string, items []CommandAction, depth int) {
		if depth > 32 {
			panic("nui: CommandPalette nesting exceeds 32")
		}
		for _, a := range items {
			checkID(seen, a.ID, "CommandPalette")
			children := a.Children
			a.Children = slices.Clone(children)
			x.levels[parent] = append(x.levels[parent], indexedCommand{a, []rune(strings.ToLower(a.Label)), []rune(strings.ToLower(a.Keywords)), x.count})
			x.count++
			if a.Key != ui.KeyUnknown {
				x.shortcuts = append(x.shortcuts, a)
			}
			add(a.ID, children, depth+1)
		}
	}
	add("", actions, 0)
	return x
}

type CommandMatch struct {
	Action     CommandAction
	Highlights []int
	Score      int
	order      int
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
	if x == nil {
		return nil
	}
	limit = commandLimit(limit)
	q := []rune(strings.ToLower(strings.TrimSpace(query)))
	h := commandHeap{}
	scratch := make([]int, 0, len(q))
	for i, it := range x.levels[parent] {
		if i%128 == 0 {
			select {
			case <-ctx.Done():
				return nil
			default:
			}
		}
		marks, score, ok := fuzzy(it.label, q, scratch)
		if !ok {
			_, score, ok = fuzzy(it.keywords, q, scratch)
			marks = nil
			score -= 20
		}
		if !ok {
			continue
		}
		if len(h) >= limit && score <= h[0].Score {
			continue
		}
		m := CommandMatch{it.action, slices.Clone(marks), score, it.order}
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
type CommandPaletteState struct {
	Open                  bool
	Query                 string
	Results               QueryState[CommandMatch]
	index                 *CommandIndex
	level                 []string
	lastQuery, lastParent string
	dirty, wasOpen        bool
	mailbox               chan commandResult
	list                  ListboxState
	editor                ui.Handle
	selected              []string
}

func (s *CommandPaletteState) SetIndex(index *CommandIndex) {
	s.Results.Cancel()
	s.index = index
	s.dirty = true
	s.level = nil
	s.Query = ""
	s.Results.Items = nil
	s.mailbox = nil
}
func (s *CommandPaletteState) Close() { s.Open = false; s.Results.Cancel(); s.mailbox = nil }
func (s *CommandPaletteState) parent() string {
	if len(s.level) > 0 {
		return s.level[len(s.level)-1]
	}
	return ""
}
func (s *CommandPaletteState) Back() {
	if len(s.level) > 0 {
		s.level = s.level[:len(s.level)-1]
		s.Query = ""
		s.dirty = true
		s.editor.Focus()
	}
}

type CommandPaletteOptions struct {
	Title, Placeholder string
	Limit              int
	ShortcutKey        ui.Key
	ShortcutMods       ui.Modifiers
	DisableShortcut    bool
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
	identity := ui.Local(host, "command", func() byte { return 0 })
	if !o.DisableShortcut {
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
	if s.index != nil {
		for _, a := range s.index.shortcuts {
			if a.Disabled || len(a.Children) > 0 {
				continue
			}
			c.OnShortcut(a.Mods, a.Key, func() {
				s.Close()
				if a.OnSelect != nil {
					a.OnSelect()
				}
				if o.OnSelect != nil {
					o.OnSelect(a.ID)
				}
			})
		}
	}
	if !s.Open {
		if s.wasOpen {
			s.Close()
		}
		s.wasOpen = false
		return host
	}
	if !s.wasOpen {
		s.editor.Focus(c)
		s.dirty = true
	}
	s.wasOpen = true
	if s.mailbox != nil {
		select {
		case r := <-s.mailbox:
			s.Results.Complete(r.ticket, r.items, nil)
			s.mailbox = nil
		default:
		}
	}
	if s.dirty || s.Query != s.lastQuery || s.parent() != s.lastParent {
		s.dirty = false
		s.lastQuery, s.lastParent = s.Query, s.parent()
		s.selected = nil
		s.list.Active = ""
		s.list.List.ScrollTo(0, ui.Start)
		ctx, ticket := s.Results.Start(context.Background(), s.Query)
		s.Results.Items = nil
		if o.Search != nil {
			o.Search(ctx, ticket, s.Query, s.parent())
		} else {
			index, query, parent, limit := s.index, s.Query, s.parent(), o.Limit
			mailbox := make(chan commandResult, 1)
			s.mailbox = mailbox
			go func() {
				items := index.Search(ctx, query, parent, limit)
				select {
				case <-ctx.Done():
					return
				case mailbox <- commandResult{ticket, items}:
				}
			}()
		}
	}
	if s.mailbox != nil {
		c.After(16 * time.Millisecond)
	}
	title := o.Title
	if title == "" {
		title = message(c, "commandPalette")
	}
	activate := func(id string) {
		if s.Results.Loading || s.Query != s.lastQuery || s.parent() != s.lastParent {
			return
		}
		for _, m := range s.Results.Items {
			a := m.Action
			if a.ID != id || a.Disabled {
				continue
			}
			if len(a.Children) > 0 {
				s.level = append(s.level, id)
				s.Query = ""
				s.dirty = true
				s.editor.Focus(c)
				return
			}
			s.Close()
			if a.OnSelect != nil {
				a.OnSelect()
			}
			if o.OnSelect != nil {
				o.OnSelect(id)
			}
			return
		}
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
		if len(s.level) > 0 {
			Button(c.Key("back"), t, message(c, "back"), ButtonOptions{Variant: Ghost}).OnClick(s.Back)
		}
		placeholder := o.Placeholder
		if placeholder == "" {
			placeholder = message(c, "searchActions")
		}
		Input(c.Key("search"), t, &s.Query, InputOptions{Label: namedMessage(c, "searchNamed", title), Placeholder: placeholder, Leading: lucide.Search(), OnSubmit: func(string) { activate(s.list.Active) }, Style: func(root, editor, leading, trailing ui.Element) {
			editor.Bind(&s.editor).OnTextKey(func(k ui.TextKey) bool {
				if k.Key == ui.KeyDown && k.Mods == 0 {
					s.list.List.Handle.Focus(c)
					return true
				}
				return false
			})
			finish(t, ComponentCommandPalette, Editor, root, PartStyle{}, o.Decoration)
		}})
		if s.Results.Loading {
			ui.Text(c, message(c, "searching")).Role(ui.RoleStatus).TextColor(t.Muted)
		}
		if s.Results.Error != "" {
			ui.Text(c, s.Results.Error).Role(ui.RoleStatus).TextColor(t.Error)
		}
		items := make([]ListboxItem, 0, len(s.Results.Items))
		matches := map[string]CommandMatch{}
		limit := o.Limit
		limit = commandLimit(limit)
		for _, m := range s.Results.Items[:min(limit, len(s.Results.Items))] {
			a := m.Action
			items = append(items, ListboxItem{ID: a.ID, Label: a.Label, Description: a.Description, Group: a.Group, Icon: a.Icon, Disabled: a.Disabled})
			matches[a.ID] = m
		}
		Listbox(c, t, &s.selected, &s.list, items, ListboxOptions{Style: func(slot Slot, e ui.Element) { finish(t, ComponentCommandPalette, slot, e, PartStyle{}, o.Decoration) }, Label: namedMessage(c, "resultsNamed", title), Height: 280, OnChange: func(ids []string) {
			if len(ids) > 0 {
				activate(ids[0])
			}
		}, Build: func(c *ui.Context, t Theme, it ListboxItem) {
			m := matches[it.ID]
			runes := []rune(it.Label)
			spans := []ui.Span{}
			start := 0
			for start < len(runes) {
				marked := slices.Contains(m.Highlights, start)
				end := start + 1
				for end < len(runes) && slices.Contains(m.Highlights, end) == marked {
					end++
				}
				span := ui.Span{Text: string(runes[start:end])}
				if marked {
					span.Color = t.Primary
					span.Background = t.Primary.Alpha(.12)
					span.Radius = 2
				}
				spans = append(spans, span)
				start = end
			}
			ui.RichText(c, spans...).Grow(1).MinWidth(0).SingleLine()

			if m.Action.Shortcut != "" {
				Kbd(c, t, []string{m.Action.Shortcut}, KbdOptions{})
			}
		}})
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
