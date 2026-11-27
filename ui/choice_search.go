package nui

import (
	"context"
	"runtime"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/egoist/mygo/ui"
	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

type choiceSearchKind uint8

const (
	searchOptions choiceSearchKind = iota
	searchListbox
	searchMenu
)

// Search workers receive only immutable scalar records, never UI builders,
// application callbacks, style maps or a build Context.
type choiceSearchRecord struct {
	id, label, description, group, shortcut string
	structural                              bool
	disabled                                bool
}

type choiceSearchData struct {
	candidates int
	kind       choiceSearchKind
	records    []choiceSearchRecord
}

var emptyOptionsSearch = &choiceSearchData{kind: searchOptions}
var emptyMenuSearch = &choiceSearchData{kind: searchMenu}

type menuView struct {
	items []MenuItem
	order []int
	menuMetadata
}

// The UI and search worker share immutable logical navigation metadata.
type menuMetadata struct {
	index   map[string]int
	enabled []int
}

func (v menuView) enabledAt(id string) int {
	at, ok := v.index[id]
	if !ok || !menuEnabled(v.item(at)) {
		return -1
	}
	return at
}

func (v menuView) count() int {
	if v.order != nil {
		return len(v.order)
	}
	return len(v.items)
}

func (v menuView) item(at int) MenuItem {
	if v.order != nil {
		at = v.order[at]
	}
	return v.items[at]
}

type menuSearchCache struct {
	search choiceSearchCache
}

func (s *menuSearchCache) apply(source submittedMenuLevel, query string, o MenuOptions) menuView {
	if !o.Filter || o.IgnoreFilter || strings.TrimSpace(query) == "" {
		s.search.pause()
		return source.view()
	}
	data := source.search
	if data == nil {
		data = emptyMenuSearch
	}
	s.search.execute(data, query, o.FilterFields, o.SearchExecution)
	if s.search.pending {
		return menuView{}
	}
	return menuView{items: source.items, order: s.search.indices, menuMetadata: s.search.result.menu}
}

type choiceSearchFields uint8

func choiceFields(kind choiceSearchKind, fields []string) choiceSearchFields {
	if fields == nil {
		return 1
	}
	var mask choiceSearchFields
	for _, field := range fields {
		switch {
		case field == "label":
			mask |= 1
		case field == "description":
			mask |= 2
		case field == "value" && kind != searchMenu, field == "id" && kind != searchOptions:
			mask |= 4
		case field == "group" && kind == searchListbox:
			mask |= 8
		case field == "shortcut" && kind == searchMenu:
			mask |= 16
		default:
			// Rich menus historically ignore unknown fields. Other choice
			// controls treat an unknown field as a programming error.
			if kind != searchMenu {
				panic("nui: unknown option filter field")
			}
		}
	}
	return mask
}

func foldListboxSearch(text string) string {
	ascii := true
	for i := range len(text) {
		if text[i] >= utf8.RuneSelf {
			ascii = false
			break
		}
	}
	if ascii {
		return strings.ToLower(text)
	}
	text = norm.NFKD.String(cases.Fold().String(text))
	return strings.Map(func(r rune) rune {
		if unicode.Is(unicode.Mn, r) {
			return -1
		}
		return r
	}, text)
}

func choiceQuery(kind choiceSearchKind, query string) string {
	if kind == searchListbox {
		return foldListboxSearch(query)
	}
	return strings.ToLower(strings.TrimSpace(query))
}

type choiceSearchIndex struct {
	data   *choiceSearchData
	fields choiceSearchFields
	text   [][5]string
}

// Go's js/wasm goroutines share the browser event loop. Only asynchronous
// jobs carry this context: Immediate mode must never yield inside a UI build.
// Time-bounded batches let new input/cancellation reach a background job.
type choiceWorkerContext struct {
	context.Context
	yielded time.Time
}

func choiceSearchCancelled(ctx context.Context, at int) bool {
	if at%128 != 0 {
		return false
	}
	if ctx.Err() != nil {
		return true
	}
	if runtime.GOOS == "js" {
		if job, ok := ctx.(*choiceWorkerContext); ok && time.Since(job.yielded) >= 2*time.Millisecond {
			time.Sleep(time.Millisecond)
			job.yielded = time.Now()
		}
	}
	return ctx.Err() != nil
}

func normalizeChoices(ctx context.Context, data *choiceSearchData, fields choiceSearchFields) *choiceSearchIndex {
	index := &choiceSearchIndex{data: data, fields: fields, text: make([][5]string, len(data.records))}
	for i, record := range data.records {
		if choiceSearchCancelled(ctx, i) {
			return nil
		}
		for field, value := range [5]string{record.label, record.description, record.id, record.group, record.shortcut} {
			if fields&(1<<field) == 0 {
				continue
			}
			if data.kind == searchListbox {
				value = foldListboxSearch(value)
			} else {
				value = strings.ToLower(value)
			}
			index.text[i][field] = value
		}
	}
	return index
}

func (x *choiceSearchIndex) score(at int, query string) int {
	best := 3
	for field, value := range x.text[at] {
		if x.fields&(1<<field) == 0 {
			continue
		}
		switch {
		case value == query:
			best = 0
		case strings.HasPrefix(value, query):
			best = min(best, 1)
		case strings.Contains(value, query):
			best = min(best, 2)
		}
	}
	return best
}

// indices preserves each component's matching contract. Every returned slice
// has its own storage; later requests cannot rewrite a published result.
func (x *choiceSearchIndex) indices(ctx context.Context, query string) []int {
	out := make([]int, 0)
	records := x.data.records
	if x.data.kind != searchListbox {
		matches := 0
		for i, record := range records {
			if choiceSearchCancelled(ctx, i) {
				return nil
			}
			if record.structural {
				out = append(out, i)
			} else if x.score(i, query) < 3 {
				out = append(out, i)
				matches++
			}
		}
		if matches == 0 {
			out = out[:0]
		}
		return out
	}
	// Rank each option run independently. Structural rows stay between runs;
	// an entire consecutive group disappears when no option in it matches.
	for start := 0; start < len(records); {
		end := start + 1
		for end < len(records) && records[end].group == records[start].group {
			if choiceSearchCancelled(ctx, end) {
				return nil
			}
			end++
		}
		groupStart, matches := len(out), 0
		for run := start; run < end; {
			if choiceSearchCancelled(ctx, run) {
				return nil
			}
			if records[run].structural {
				out = append(out, run)
				run++
				continue
			}
			var buckets [3][]int
			for run < end && !records[run].structural {
				if choiceSearchCancelled(ctx, run) {
					return nil
				}
				if score := x.score(run, query); score < 3 {
					buckets[score] = append(buckets[score], run)
					matches++
				}
				run++
			}
			for _, bucket := range buckets {
				out = append(out, bucket...)
			}
		}
		if matches == 0 {
			out = out[:groupStart]
		}
		start = end
	}
	return out
}

type choiceSearchCache struct {
	worker             *choiceSearchWorker
	mount              *byte
	services           ui.Services
	request            choiceSearchKey
	ticket             uint64
	requested, pending bool
	result             choiceSearchResult
	index              *choiceSearchIndex
	indices            []int
	query              string
	known              bool
	revision           uint64
}

func (s *choiceSearchCache) apply(data *choiceSearchData, query string, fields []string) bool {
	return s.execute(data, query, fields, SearchImmediate)
}
