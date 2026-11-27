package nui

import (
	"context"
	"fmt"
	"math"
	"slices"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"golang.org/x/text/cases"
)

func TestChoiceSearchPreservesMatchingAndGrouping(t *testing.T) {
	var state ListboxState
	state.SetItems([]ListboxItem{
		{ID: "heading", Label: "Heading", Group: "one", Kind: ListboxLabel},
		{ID: "contains", Label: "A café", Group: "one"},
		{ID: "prefix", Label: "Cafeteria", Group: "one"},
		{ID: "exact", Label: "CAFÉ", Group: "one", Disabled: true},
		{ID: "separator", Kind: ListboxSeparator, Group: "one"},
		{ID: "later", Label: "Cafe", Group: "one"},
		{ID: "other-heading", Label: "Empty group", Group: "two", Kind: ListboxLabel},
		{ID: "other", Label: "Unrelated", Group: "two"},
		{ID: "again", Label: "Cafe", Group: "one"},
	})
	var filter listboxFilter
	filter.apply(&state.source, "cafe", nil)
	ids := func() []string {
		out := []string{}
		for i := range filter.result.count() {
			out = append(out, filter.result.item(i).ID)
		}
		return out
	}
	want := []string{"heading", "exact", "prefix", "contains", "separator", "later", "again"}
	if !slices.Equal(ids(), want) {
		t.Fatalf("group/rank/structure order = %v, want %v", ids(), want)
	}
	if filter.result.enabledIndex("exact") >= 0 || filter.result.enabledIndex("heading") >= 0 || filter.result.enabledIndex("later") < 0 {
		t.Fatal("filtered metadata made disabled or structural rows selectable")
	}
	index, revision, previous := filter.search.index, filter.result.revision, filter.search.indices
	filter.apply(&state.source, "cafe", nil)
	if filter.result.revision != revision || filter.search.index != index {
		t.Fatal("identical query rebuilt its index or result metadata")
	}
	filter.apply(&state.source, "cafeteria", nil)
	if filter.search.index != index || !slices.Equal(previous, []int{0, 3, 2, 1, 4, 5, 8}) {
		t.Fatal("another query renormalized the source or rewrote published indices")
	}
	filter.apply(&state.source, "cafe", []string{})
	if filter.result.count() != 0 {
		t.Fatal("an explicit empty field list used the default label field")
	}
	filter.apply(&state.source, "one", []string{"group"})
	if filter.result.count() != 7 {
		t.Fatal("group field did not include its matching structural runs")
	}
	state.SetItems([]ListboxItem{{ID: "exact", Label: "Renamed", Description: "cafe"}})
	filter.apply(&state.source, "cafe", nil)
	if filter.result.count() != 0 {
		t.Fatal("same-ID label submission retained stale search text")
	}
	filter.apply(&state.source, "cafe", []string{"description"})
	if !slices.Equal(ids(), []string{"exact"}) {
		t.Fatal("field configuration did not invalidate the index")
	}
}

func TestChoiceSearchKeepsUnrankedComponentRules(t *testing.T) {
	for _, kind := range []choiceSearchKind{searchOptions, searchMenu} {
		data := &choiceSearchData{kind: kind, records: []choiceSearchRecord{
			{id: "contains", label: "A cafe"}, {id: "accent", label: "Café"}, {id: "exact", label: "CAFE"},
		}}
		var cache choiceSearchCache
		cache.apply(data, " CAFE ", nil)
		if !slices.Equal(cache.indices, []int{0, 2}) {
			t.Fatalf("kind %d changed accent or source-order semantics: %v", kind, cache.indices)
		}
		cache.apply(data, "cafe", []string{})
		if len(cache.indices) != 0 {
			t.Fatal("an explicit empty field list used the label default")
		}
	}
	data := &choiceSearchData{kind: searchMenu, records: []choiceSearchRecord{
		{id: "heading", structural: true}, {id: "a", label: "Alpha"}, {id: "separator", structural: true}, {id: "b", label: "Beta", shortcut: "Ctrl+B"},
	}}
	var cache choiceSearchCache
	cache.apply(data, "ctrl+b", []string{"shortcut"})
	if !slices.Equal(cache.indices, []int{0, 2, 3}) {
		t.Fatal("menu shortcut search lost structural rows")
	}
	cache.apply(data, "missing", nil)
	if len(cache.indices) != 0 {
		t.Fatal("menu with no matching actions retained structural rows")
	}
}

func TestChoiceSearchASCIIAndCancellation(t *testing.T) {
	// Exhaust the ASCII fast path against Unicode case folding, including
	// punctuation and controls, rather than assuming it only sees letters.
	for ch := byte(0); ch < 128; ch++ {
		text := string([]byte{'A', ch, 'Z'})
		if got, want := foldListboxSearch(text), cases.Fold().String(text); got != want {
			t.Fatalf("fold %q = %q, want %q", text, got, want)
		}
	}
	if foldListboxSearch("Straße CAFÉ Ａ") != "strasse cafe a" {
		t.Fatal("Unicode compatibility/case/accent folding changed")
	}
	data := &choiceSearchData{kind: searchListbox, records: make([]choiceSearchRecord, 10000)}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if normalizeChoices(ctx, data, 1) != nil {
		t.Fatal("normalization ignored cancellation")
	}
	index := normalizeChoices(context.Background(), data, 1)
	if index.indices(ctx, "x") != nil {
		t.Fatal("matching ignored cancellation")
	}
}

type heldChoiceSearch struct {
	worker  *choiceSearchWorker
	started chan choiceSearchJob
	calls   atomic.Int32
	release func()
}

func holdFirstChoiceSearch(t *testing.T) *heldChoiceSearch {
	t.Helper()
	h := &heldChoiceSearch{worker: newChoiceSearchWorker(), started: make(chan choiceSearchJob, 100)}
	released := make(chan struct{})
	var once sync.Once
	h.release = func() { once.Do(func() { close(released) }) }
	h.worker.work = func(job choiceSearchJob, index *choiceSearchIndex) choiceSearchResult {
		first := h.calls.Add(1) == 1
		h.started <- job
		if first {
			<-released
		}
		return runChoiceSearch(job, index)
	}
	t.Cleanup(func() { h.worker.cancel(); h.release(); waitChoiceWorker(t, h.worker) })
	return h
}

func (h *heldChoiceSearch) next(t *testing.T) choiceSearchJob {
	t.Helper()
	select {
	case job := <-h.started:
		return job
	case <-time.After(5 * time.Second):
		t.Fatal("search worker did not start")
		return choiceSearchJob{}
	}
}

func waitChoiceWorker(t testing.TB, w *choiceSearchWorker) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		w.mu.Lock()
		running := w.running
		w.mu.Unlock()
		if !running {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("search worker did not stop")
}

func awaitChoiceView(t *testing.T, tt *ui.Tester, pending func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		tt.Advance(16 * time.Millisecond)
		if !pending() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("search result did not reach the UI")
}

func TestChoiceSearchExecutionPolicy(t *testing.T) {
	for _, test := range []struct {
		n          int
		mode       SearchExecution
		background bool
	}{
		{2048, SearchAuto, false}, {2049, SearchAuto, true}, {10000, SearchImmediate, false}, {1, SearchBackground, true},
	} {
		t.Run(fmt.Sprintf("%d/%d", test.n, test.mode), func(t *testing.T) {
			var state ListboxState
			state.SetItems(performanceItems(test.n))
			held := holdFirstChoiceSearch(t)
			cache := choiceSearchCache{worker: held.worker}
			cache.execute(state.source.search, "Item", nil, test.mode)
			if test.background {
				held.next(t)
				if !cache.pending || cache.index != nil || cache.known {
					t.Fatal("first normalization ran on the caller or reported ready")
				}
				held.release()
				waitChoiceWorker(t, held.worker)
				cache.execute(state.source.search, "Item", nil, test.mode)
			} else if held.calls.Load() != 0 {
				t.Fatal("Immediate/small Auto started a worker")
			}
			if cache.pending || len(cache.indices) != test.n {
				t.Fatal("policy lost matching results")
			}
			index, revision := cache.index, cache.revision
			cache.execute(state.source.search, "Item", nil, test.mode)
			if cache.index != index || cache.revision != revision {
				t.Fatal("unchanged request repeated search")
			}
			cache.reset()
		})
	}
}

func TestChoiceSearchWorkerKeepsOnlyLatestRequest(t *testing.T) {
	var state ListboxState
	state.SetItems(performanceItems(10000))
	held := holdFirstChoiceSearch(t)
	cache := choiceSearchCache{worker: held.worker}
	cache.execute(state.source.search, "Item 000001", nil, SearchBackground)
	first := held.next(t)
	for i := range 100 {
		cache.execute(state.source.search, fmt.Sprintf("Item %06d", i), nil, SearchBackground)
	}
	if first.ctx.Err() == nil || held.calls.Load() != 1 {
		t.Fatal("superseded work remained live or requests launched parallel workers")
	}
	held.worker.mu.Lock()
	pending := held.worker.pending
	held.worker.mu.Unlock()
	if pending == nil || pending.key.query != "Item 000099" {
		t.Fatal("latest pending request was not retained")
	}
	held.release()
	waitChoiceWorker(t, held.worker)
	cache.execute(state.source.search, "Item 000099", nil, SearchBackground)
	if held.calls.Load() != 2 || cache.pending || len(cache.indices) != 1 || cache.indices[0] != 99 {
		t.Fatal("intermediate or stale request was published")
	}
	previous := cache.indices
	index := cache.index
	cache.execute(state.source.search, "Item 000100", nil, SearchBackground)
	waitChoiceWorker(t, held.worker)
	cache.execute(state.source.search, "Item 000100", nil, SearchBackground)
	if cache.index != index || previous[0] != 99 || cache.indices[0] != 100 {
		t.Fatal("query changed normalized data or rewrote a published result")
	}
	cache.reset()
}

func TestChoiceSearchTypingStaysAvailableWhileWorkerHeld(t *testing.T) {
	var state ListboxState
	var selected SelectionState[string]
	state.SetItems(performanceItems(10000))
	held := holdFirstChoiceSearch(t)
	state.filter.search.worker = held.worker
	changes, activations := 0, 0
	tt := ui.NewTester(func(c *ui.Context) {
		Listbox(c.Key("choices"), App(c, AppConfig{DisableAnimations: true}), &selected, &state, ListboxOptions{Filter: true, FilterInput: &InputOptions{Label: "Search"}, Height: 300, OnChange: func() { changes++ }, OnSelect: func(string) { activations++ }}).Width(400)
	}, 440, 360)
	defer state.Close()
	if held.calls.Load() != 0 {
		t.Fatal("empty query started local search")
	}
	tt.Key(0, ui.KeyTab)
	tt.Type("Item")
	held.next(t)
	if !state.Searching() || state.filter.search.index != nil || !tt.Focused("Search") || tt.HasText("Item 000000") {
		t.Fatal("first large search blocked editing or exposed stale results")
	}

	tt.Type(" 009999")
	tt.Key(0, ui.KeyEnter)
	if state.Query != "Item 009999" || selected.Len() != 0 || changes != 0 || activations != 0 {
		t.Fatal("pending Enter activated an old item or lost typing")
	}
	held.release()
	awaitChoiceView(t, tt, state.Searching)
	if !tt.HasText("Item 009999") || tt.HasText("Item 000000") {
		t.Fatal("latest query did not replace results")
	}
	tt.Key(0, ui.KeyEnter)
	if !selected.Has("9999") || changes != 1 || activations != 1 {
		t.Fatal("completed result did not activate once")
	}
}

func TestChoiceSearchCancellationRejectsQueuedPublication(t *testing.T) {
	var state ListboxState
	state.SetItems(performanceItems(100))
	held := holdFirstChoiceSearch(t)
	cache := choiceSearchCache{worker: held.worker}
	cache.execute(state.source.search, "Item", nil, SearchBackground)
	old := held.next(t)
	cache.reset()
	if cache.pending || old.ctx.Err() == nil {
		t.Fatal("reset did not cancel the active request")
	}
	state.SetItems([]ListboxItem{{ID: "new", Label: "New"}})
	cache.execute(state.source.search, "New", nil, SearchBackground)
	// Even a completion delivered after cancellation must fail the UI ticket.
	held.worker.mailbox <- choiceSearchResult{ticket: old.ticket, index: &choiceSearchIndex{data: old.key.data}, indices: []int{0}}
	cache.execute(state.source.search, "New", nil, SearchBackground)
	if !cache.pending || cache.known {
		t.Fatal("old result crossed a reset/data generation")
	}
	held.release()
	waitChoiceWorker(t, held.worker)
	cache.execute(state.source.search, "New", nil, SearchBackground)
	if cache.pending || cache.index.data != state.source.search || len(cache.indices) != 1 {
		t.Fatal("replacement data did not publish")
	}
	cache.reset()
}

func TestChoiceSearchWaitingDoesNotPollViews(t *testing.T) {
	var state ListboxState
	var selection SelectionState[string]
	state.SetItems(performanceItems(10000))
	state.Query = "Item"
	held := holdFirstChoiceSearch(t)
	state.filter.search.worker = held.worker
	builds := 0
	tt := ui.NewTester(func(c *ui.Context) {
		builds++
		// No focused editor: its native caret blink legitimately schedules builds.
		Listbox(c.Key("choices"), App(c, AppConfig{DisableAnimations: true}), &selection, &state, ListboxOptions{Height: 300}).Width(400)
	}, 440, 360)
	defer state.Close()
	held.next(t)
	tt.Frame()
	before := builds
	tt.Advance(time.Second)
	if builds != before {
		t.Fatal("pending search introduced a polling build loop")
	}
	held.release()
	awaitChoiceView(t, tt, state.Searching)
	if builds <= before || !tt.HasText("Item 000000") {
		t.Fatal("worker did not invalidate for publication")
	}
}

type searchControlFixture struct {
	build                            func(*ui.Context)
	cache                            func() *choiceSearchCache
	close, open, clearDraft, replace func()
	bypass                           bool
	activations                      int
	chosen                           string
}

func newSearchControlFixture(kind string) *searchControlFixture {
	f := &searchControlFixture{}
	var list ListboxState
	var menu SelectMenuState
	var input SearchState
	var actions MenuState
	var selected SelectionState[string]
	input.Open, menu.Open, actions.Open = true, true, true
	setData := func(replacement bool) {
		items := performanceItems(100)
		if replacement {
			items[99].Label = "Current target"
		}
		switch kind {
		case "Listbox":
			list.SetItems(items)
		case "SelectMenu":
			menu.SetItems(items)
		case "InputMenu":
			options := make([]Option[string], len(items))
			for i, item := range items {
				options[i] = Option[string]{Value: item.ID, Label: item.Label}
			}
			input.SetOptions(options)
		case "Menu":
			commands := make([]MenuItem, len(items))
			for i, item := range items {
				commands[i] = MenuItem{ID: item.ID, Label: item.Label, OnSelect: func() { f.activations++; f.chosen = item.ID }}
			}
			actions.SetItems(commands)
		}
	}
	setData(false)
	f.replace = func() { setData(true) }
	f.clearDraft = func() { list.Query, menu.Query, input.Query, actions.Query = "", "", "", "" }
	f.open = func() { input.Open, menu.Open, actions.Open = true, true, true }
	f.cache = func() *choiceSearchCache {
		switch kind {
		case "Listbox":
			return &list.filter.search
		case "SelectMenu":
			return &menu.filterCache.search
		case "InputMenu":
			return &input.memory.search
		default:
			return &actions.levels[""].search.search
		}
	}
	f.close = func() {
		switch kind {
		case "Listbox":
			list.Close()
		case "SelectMenu":
			menu.Close()
		case "InputMenu":
			input.Close()
		case "Menu":
			actions.Close()
		}
	}
	f.build = func(c *ui.Context) {
		t := App(c, AppConfig{DisableAnimations: true})
		onSelect := func(id string) { f.activations++; f.chosen = id }
		switch kind {
		case "Listbox":
			Listbox(c.Key("control"), t, &selected, &list, ListboxOptions{Label: "Choices", Filter: true, FilterInput: &InputOptions{Label: "Search"}, SearchExecution: SearchBackground, IgnoreFilter: f.bypass, OnSelect: onSelect}).Width(400)
		case "SelectMenu":
			SelectMenu(c.Key("control"), t, &selected, &menu, SelectMenuOptions{Label: "Choices", FilterInput: &InputOptions{Label: "Search"}, SearchExecution: SearchBackground, ExternalFilter: f.bypass, OnSelect: onSelect}).Width(400)
		case "InputMenu":
			InputMenu(c.Key("control"), t, &input, InputMenuOptions{Label: "Search", SearchExecution: SearchBackground, ExternalFilter: f.bypass, OnSelect: onSelect}).Width(400)
		case "Menu":
			anchor := ui.Button(c.Key("trigger"), "Actions")
			DropdownMenu(c.Key("control"), t, anchor, &actions, MenuOptions{Filter: true, FilterInput: &InputOptions{Label: "Search"}, SearchExecution: SearchBackground, IgnoreFilter: f.bypass})
		}
	}
	return f
}

func TestChoiceSearchUnmountCloseAndRemount(t *testing.T) {
	for _, kind := range []string{"Listbox", "SelectMenu", "InputMenu", "Menu"} {
		for _, explicitClose := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/Close=%v", kind, explicitClose), func(t *testing.T) {
				f := newSearchControlFixture(kind)
				shown := true
				tt := ui.NewTester(func(c *ui.Context) {
					if shown {
						f.build(c)
					}
				}, 600, 500)
				defer f.close()
				held := holdFirstChoiceSearch(t)
				cache := f.cache()
				cache.worker = held.worker
				if err := tt.Click("Search"); err != nil {
					t.Fatal(err)
				}
				tt.Type("Item")
				old := held.next(t)
				tt.Key(0, ui.KeyEnter)
				if !cache.pending || f.activations != 0 {
					t.Fatal("pending search activated an old choice")
				}
				if explicitClose {
					f.close()
				}
				shown = false
				tt.Frame()
				if old.ctx.Err() == nil || cache.pending || cache.mount != nil {
					t.Fatal("unmount/Close left work or the old mount live")
				}
				f.clearDraft()
				f.replace()
				shown = true
				tt.Frame()
				f.open()
				tt.Frame()
				if err := tt.Click("Search"); err != nil {
					t.Fatal(err)
				}
				tt.Type("Current target")
				if f.cache().worker != held.worker || held.calls.Load() != 1 {
					t.Fatal("remount started a parallel worker while cancellation was still finishing")
				}
				held.release()
				awaitChoiceView(t, tt, func() bool { return f.cache().pending })
				if !tt.HasText("Current target") || tt.HasText("Item 000000") || f.activations != 0 {
					t.Fatal("old data/results crossed the mount generation")
				}
				tt.Key(0, ui.KeyEnter)
				if f.activations != 1 || f.chosen != "99" {
					t.Fatalf("current result activation = %q/%d", f.chosen, f.activations)
				}
			})
		}
	}
}

func TestChoiceSearchExternalBypassCancelsLocalWork(t *testing.T) {
	for _, kind := range []string{"Listbox", "SelectMenu", "InputMenu", "Menu"} {
		t.Run(kind, func(t *testing.T) {
			f := newSearchControlFixture(kind)
			tt := ui.NewTester(f.build, 600, 500)
			defer f.close()
			held := holdFirstChoiceSearch(t)
			cache := f.cache()
			cache.worker = held.worker
			if err := tt.Click("Search"); err != nil {
				t.Fatal(err)
			}
			tt.Type("Item")
			old := held.next(t)
			f.bypass = true
			tt.Frame()
			if cache.pending || old.ctx.Err() == nil || !tt.HasText("Item 000000") {
				t.Fatal("application filtering did not bypass/cancel local search")
			}
			tt.Type("never matches")
			if held.calls.Load() != 1 || cache.pending || !tt.HasText("Item 000000") {
				t.Fatal("external query duplicated local matching")
			}
			held.release()
			waitChoiceWorker(t, held.worker)
			tt.Frame()
			if !tt.HasText("Item 000000") || f.activations != 0 {
				t.Fatal("cancelled local result replaced application-owned choices")
			}
		})
	}
}

func TestChoiceSearchPendingDoesNotCreateValues(t *testing.T) {
	for _, kind := range []string{"InputMenu", "SelectMenu"} {
		t.Run(kind, func(t *testing.T) {
			var input SearchState
			var menu SelectMenuState
			var selected SelectionState[string]
			input.SetOptions([]Option[string]{{Value: "alpha", Label: "Alpha"}})
			menu.SetItems([]ListboxItem{{ID: "alpha", Label: "Alpha"}})
			input.Open, menu.Open = true, true
			defer input.Close()
			defer menu.Close()
			created, changes := "", 0
			tt := ui.NewTester(func(c *ui.Context) {
				th := App(c, AppConfig{DisableAnimations: true})
				onCreate := func(query string) { created = query }
				if kind == "InputMenu" {
					InputMenu(c.Key("choices"), th, &input, InputMenuOptions{Label: "Search", Multiple: true, SearchExecution: SearchBackground, OnCreate: onCreate, OnValuesChange: func() { changes++ }}).Width(400)
				} else {
					SelectMenu(c.Key("choices"), th, &selected, &menu, SelectMenuOptions{Label: "Choices", Multiple: true, FilterInput: &InputOptions{Label: "Search"}, SearchExecution: SearchBackground, OnCreate: onCreate, OnChange: func() { changes++ }}).Width(400)
				}
			}, 480, 400)
			cache := &input.memory.search
			if kind == "SelectMenu" {
				cache = &menu.filterCache.search
			}
			held := holdFirstChoiceSearch(t)
			cache.worker = held.worker
			if err := tt.Click("Search"); err != nil {
				t.Fatal(err)
			}
			tt.Type("missing")
			held.next(t)
			tt.Key(0, ui.KeyEnter)
			if !tt.HasText("Loading") || created != "" || changes != 0 || selected.Len() != 0 || input.SelectedValues.Len() != 0 {
				t.Fatal("pending search exposed an empty/create action or changed selection")
			}
			held.release()
			awaitChoiceView(t, tt, func() bool { return cache.pending })
			if !tt.HasText("Create") {
				t.Fatal("completed empty result did not expose creation")
			}
			tt.Key(0, ui.KeyEnter)
			if created != "missing" || changes != 0 || selected.Len() != 0 || input.SelectedValues.Len() != 0 {
				t.Fatal("creation inserted a business value or emitted a selection change")
			}
		})
	}
}

func TestChoiceSearchPublishesIntoFocusedList(t *testing.T) {
	var state SelectMenuState
	var selected SelectionState[string]
	state.SetItems([]ListboxItem{{ID: "a", Label: "Alpha"}, {ID: "b", Label: "Beta"}})
	state.Open, state.Query = true, "Beta"
	held := holdFirstChoiceSearch(t)
	state.filterCache.search.worker = held.worker
	defer state.Close()
	changes := 0
	tt := ui.NewTester(func(c *ui.Context) {
		SelectMenu(c.Key("choices"), App(c, AppConfig{DisableAnimations: true}), &selected, &state, SelectMenuOptions{
			Label: "Choices", FocusSearch: Set(false), SearchExecution: SearchBackground, OnChange: func() { changes++ },
		}).Width(400)
	}, 480, 400)
	held.next(t)
	tt.Key(0, ui.KeyEnter)
	if selected.Len() != 0 || changes != 0 {
		t.Fatal("focused pending list activated an old row")
	}
	held.release()
	awaitChoiceView(t, tt, state.Searching)
	tt.Key(0, ui.KeyEnter)
	if !selected.Has("b") || changes != 1 || state.Open || !tt.Focused("Choices") {
		t.Fatal("result publication lost list focus, activation or return focus")
	}
}

func expectSubmissionPanic(t *testing.T, submit func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Error("invalid submission was accepted")
		}
	}()
	submit()
}

func TestSelectSubmissionOwnsDataAndValidatesBeforeCommit(t *testing.T) {
	var s SelectState[float64]
	called := ""
	options := []Option[float64]{{Value: 0, Label: "Zero", Build: func(*ui.Context, Theme) { called = "original" }}}
	s.SetOptions(options)
	options[0].Label = "Changed"
	options[0].Build = func(*ui.Context, Theme) { called = "replacement" }
	s.source.items[0].Build(nil, Theme{})
	if s.source.items[0].Label != "Zero" || called != "original" || s.source.metadata.index[0] != 0 {
		t.Fatal("submission aliased descriptors or lost a typed zero")
	}
	for _, invalid := range [][]Option[float64]{
		{{Value: 1}, {Value: 1}}, {{Value: 1}, {Value: math.NaN()}},
	} {
		expectSubmissionPanic(t, func() { s.SetOptions(invalid) })
		if s.source.revision != 1 || len(s.source.items) != 1 || s.source.items[0].Label != "Zero" {
			t.Fatal("invalid update replaced the usable option model")
		}
	}
	s.SetOptions(options)
	s.source.items[0].Build(nil, Theme{})
	if s.source.items[0].Label != "Changed" || called != "replacement" || s.source.revision != 2 {
		t.Fatal("same-ID descriptor replacement was ignored")
	}
	s.SetOptions(options[:0])
	if s.source.items != nil || s.source.metadata.index != nil || s.source.metadata.keys != nil {
		t.Fatal("empty submission retained option storage")
	}
}

func TestInputMenuSubmissionKeepsDraftAndRejectsEmptyIDs(t *testing.T) {
	s := SearchState{Query: "draft", Selected: "a", Open: true}
	s.SetOptions([]Option[string]{{Value: "a", Label: "Alpha"}})
	expectSubmissionPanic(t, func() { s.SetOptions([]Option[string]{{Value: "", Label: "Invalid"}}) })
	if s.source.items[0].Label != "Alpha" || s.source.revision != 1 {
		t.Fatal("invalid input-menu data was committed")
	}
	s.SetOptions(nil)
	if s.Query != "draft" || s.Selected != "a" || !s.Open || s.source.items != nil || s.source.metadata.index != nil {
		t.Fatal("empty options reset independent input state or retained data")
	}
}

func TestListboxSubmissionOwnsNestedDescriptors(t *testing.T) {
	color := ui.RGB(10, 20, 30)
	styles := Styles{Item: {Background: Set(color)}}
	called := ""
	items := []ListboxItem{{ID: "a", Label: "Alpha",
		Avatar:  &AvatarOptions{Text: "A", Decoration: Decoration{UI: styles}},
		Chip:    &ChipOptions{Text: "Chip", Decoration: Decoration{UI: styles}},
		Options: &ListboxItemOptions{Decoration: Decoration{UI: styles}, OnSelect: func(*ListboxEvent) { called = "original" }},
	}}
	var s ListboxState
	s.SetItems(items)
	styles[Item] = PartStyle{Background: Set(ui.Transparent)}
	items[0].Label, items[0].Avatar.Text, items[0].Chip.Text = "Changed", "B", "Other"
	items[0].Options.OnSelect = func(*ListboxEvent) { called = "replacement" }
	owned := s.source.items[0]
	owned.Options.OnSelect(&ListboxEvent{})
	if owned.Label != "Alpha" || owned.Avatar.Text != "A" || owned.Chip.Text != "Chip" || called != "original" {
		t.Fatal("listbox submission aliased a nested descriptor")
	}
	for _, style := range []Styles{owned.Options.UI, owned.Avatar.UI, owned.Chip.UI} {
		if style[Item].Background.Value != color {
			t.Fatal("listbox submission aliased a style map")
		}
	}
	s.SetItems(items)
	s.source.items[0].Options.OnSelect(&ListboxEvent{})
	if called != "replacement" || s.source.items[0].Label != "Changed" {
		t.Fatal("listbox same-ID replacement was ignored")
	}
	for _, invalid := range [][]ListboxItem{{{ID: ""}}, {{ID: "a"}, {ID: "a"}}, {{ID: "a", Kind: 255}}} {
		expectSubmissionPanic(t, func() { s.SetItems(invalid) })
		if s.source.revision != 2 || s.source.items[0].Label != "Changed" {
			t.Fatal("invalid listbox update damaged the previous data")
		}
	}
	s.Query = "retained query"
	s.SetItems(items[:0])
	if s.source.items != nil || s.source.index != nil || s.Query != "retained query" {
		t.Fatal("empty listbox submission retained data or reset the query")
	}
}

func TestSelectMenuSubmissionUpdatesBothModels(t *testing.T) {
	s := SelectMenuState{Query: "draft", Open: true}
	items := []ListboxItem{{ID: "a", Label: "Alpha"}}
	s.SetItems(items)
	items[0].Label = "Replacement"
	if s.data.items[0].Label != "Alpha" || s.Listbox.source.items[0].Label != "Alpha" {
		t.Fatal("trigger or list aliased the caller's data")
	}
	s.SetItems(items)
	if s.data.items[0].Label != "Replacement" || s.Listbox.source.items[0].Label != "Replacement" {
		t.Fatal("trigger and list did not receive the same revision")
	}
	s.SetItems(nil)
	if s.data.items != nil || s.Listbox.source.items != nil || s.Query != "draft" || !s.Open {
		t.Fatal("empty submission retained data or reset independent popup state")
	}
}

func TestMenuSubmissionOwnsTreeAndBindings(t *testing.T) {
	query, called := "draft", ""
	styles := Styles{Item: {Radius: Set(float32(9))}}
	fields := []string{"label"}
	items := []MenuItem{{ID: "parent", Label: "Parent", UI: styles, Avatar: &AvatarOptions{Text: "P", Decoration: Decoration{UI: styles}},
		Content:  &MenuContentOptions{SearchTerm: &query, FilterFields: fields, Decoration: Decoration{UI: styles}, FilterInput: &InputOptions{Label: "Search", UI: styles, Avatar: &AvatarOptions{Text: "Q", Decoration: Decoration{UI: styles}}}},
		Children: []MenuItem{{ID: "child", Label: "Child", OnSelect: func() { called = "original" }}},
	}}
	var s MenuState
	s.SetItems(items)
	items[0].Children[0].Label = "Changed"
	items[0].Children[0].OnSelect = func() { called = "replacement" }
	items[0].Avatar.Text, items[0].Content.FilterInput.Label = "Other", "Other search"
	fields[0] = "description"
	styles[Item] = PartStyle{Radius: Set(float32(42))}
	owned := s.source.items[0]
	owned.Children[0].OnSelect()
	if owned.Children[0].Label != "Child" || called != "original" || owned.Avatar.Text != "P" || owned.Content.FilterInput.Label != "Search" || owned.Content.FilterFields[0] != "label" {
		t.Fatal("menu submission aliased nested data")
	}
	for _, style := range []Styles{owned.UI, owned.Avatar.UI, owned.Content.UI, owned.Content.FilterInput.UI, owned.Content.FilterInput.Avatar.UI} {
		if style[Item].Radius.Value != 9 {
			t.Fatal("menu submission aliased a style map")
		}
	}
	query = "live binding"
	if *owned.Content.SearchTerm != query {
		t.Fatal("submission detached an intentional app-owned query binding")
	}
	s.SetItems(items)
	s.source.byID["child"].item.OnSelect()
	if called != "replacement" || s.source.byID["child"].item.Label != "Changed" {
		t.Fatal("same-ID menu callback/label replacement was ignored")
	}
	for _, invalid := range [][]MenuItem{
		{{ID: ""}}, {{ID: "x", Children: []MenuItem{{ID: "x"}}}}, {{ID: "bad", Kind: 255}},
	} {
		expectSubmissionPanic(t, func() { s.SetItems(invalid) })
		if s.source.revision != 2 || s.source.byID["child"].item.Label != "Changed" {
			t.Fatal("invalid menu update damaged the previous tree")
		}
	}
	deep := []MenuItem{{ID: "leaf"}}
	for i := range 33 {
		deep = []MenuItem{{ID: strconv.Itoa(i), Children: deep}}
	}
	expectSubmissionPanic(t, func() { s.SetItems(deep) })
	s.SetItems(nil)
	if s.source.items != nil || s.source.byID != nil || s.source.levels != nil {
		t.Fatal("empty menu submission retained data")
	}
}

// Test-only conveniences that run local search synchronously.
func (s *choiceSearchCache) apply(data *choiceSearchData, query string, fields []string) bool {
	return s.execute(data, query, fields, SearchImmediate)
}

func (s *listboxFilter) apply(data *listboxItems, query string, fields []string) {
	s.visible(data, query, fields, SearchImmediate)
}

func (s *listboxItems) update(items []ListboxItem) bool {
	if slices.Equal(s.items, items) {
		return false
	}
	s.items, s.index = slices.Clone(items), indexListboxItems(items)
	s.revision++
	s.prepareRows()
	s.prepareSearch()
	return true
}
