package nui

import (
	"context"
	"runtime"
	"sync"

	"github.com/egoist/mygo/ui"
)

// SearchExecution controls where local choice filtering runs. Empty queries
// and application-filtered data bypass local search under every policy.
type SearchExecution uint8

const (
	// SearchAuto runs normalization and matching in the background above
	// 2,048 submitted candidates, and immediately for smaller datasets.
	SearchAuto SearchExecution = iota
	SearchImmediate
	SearchBackground
)

func (execution SearchExecution) background(candidates int) bool {
	if execution > SearchBackground {
		panic("nui: invalid search execution policy")
	}
	return execution == SearchBackground || execution == SearchAuto && candidates > 2048
}

type choiceSearchKey struct {
	data      *choiceSearchData
	fields    choiceSearchFields
	query     string
	execution SearchExecution
}

type choiceSearchJob struct {
	key      choiceSearchKey
	ticket   uint64
	ctx      context.Context
	cancel   context.CancelFunc
	services ui.Services
}

type choiceSearchResult struct {
	ticket     uint64
	index      *choiceSearchIndex
	indices    []int
	listbox    listboxItems
	options    optionMetadata[string]
	membership map[int]bool
	menu       menuMetadata
}

// Metadata contains only logical IDs/indices. Preparing it alongside matching
// avoids a second full candidate/result walk when the UI receives a result.
func runChoiceSearch(job choiceSearchJob, index *choiceSearchIndex) choiceSearchResult {
	key := job.key
	if index == nil || index.data != key.data || index.fields != key.fields {
		index = normalizeChoices(job.ctx, key.data, key.fields)
	}
	out := choiceSearchResult{ticket: job.ticket, index: index}
	if index == nil || job.ctx.Err() != nil {
		return out
	}
	out.indices = index.indices(job.ctx, choiceQuery(key.data.kind, key.query))
	if job.ctx.Err() != nil {
		return out
	}
	count := len(out.indices)
	switch key.data.kind {
	case searchOptions:
		out.options = optionMetadata[string]{identities: make([]optionIdentity[string], count), index: make(map[string]int, count), keys: make([]any, count)}
		out.membership = make(map[int]bool, count)
	case searchListbox:
		out.listbox = listboxItems{order: out.indices, index: make(map[string]int, count), positions: make([]int, count), rows: make([]listboxRow, 0, count)}
	case searchMenu:
		out.menu.index = make(map[string]int, count)
	}
	previousGroup, groupNumber := "", 0
	for i, source := range out.indices {
		if choiceSearchCancelled(job.ctx, i) {
			return choiceSearchResult{index: index}
		}
		record := key.data.records[source]
		switch key.data.kind {
		case searchOptions:
			out.options.identities[i] = optionIdentity[string]{record.id, record.label, record.description, record.disabled}
			out.options.index[record.id], out.options.keys[i] = i, optionKey[string]{record.id}
			out.membership[source] = true
			if record.description != "" {
				out.options.descriptions++
			}
			if !record.disabled {
				out.options.enabledCount++
			}
		case searchListbox:
			if record.group != "" && record.group != previousGroup {
				groupNumber++
				out.listbox.rows = append(out.listbox.rows, listboxRow{item: -1, group: record.group, key: struct {
					Group  string
					Number int
				}{record.group, groupNumber}})
			}
			out.listbox.index[record.id], out.listbox.positions[i] = i, len(out.listbox.rows)
			out.listbox.rows = append(out.listbox.rows, listboxRow{item: i, key: record.id})
			if !record.structural && !record.disabled {
				out.listbox.enabled = append(out.listbox.enabled, i)
			}
			previousGroup = record.group
		case searchMenu:
			out.menu.index[record.id] = i
			if !record.structural && !record.disabled {
				out.menu.enabled = append(out.menu.enabled, i)
			}
		}
	}
	return out
}

// The mutex protects scheduling only. There is at most one running worker and
// one replacement request, including across close/remount transitions. Workers
// publish into a bounded mailbox; only the owning UI build updates the model.
type choiceSearchWorker struct {
	mu      sync.Mutex
	running bool
	active  context.CancelFunc
	pending *choiceSearchJob
	index   *choiceSearchIndex
	mailbox chan choiceSearchResult
	work    func(choiceSearchJob, *choiceSearchIndex) choiceSearchResult
}

func newChoiceSearchWorker() *choiceSearchWorker {
	return &choiceSearchWorker{mailbox: make(chan choiceSearchResult, 1), work: runChoiceSearch}
}

func (w *choiceSearchWorker) submit(job choiceSearchJob) {
	w.mu.Lock()
	if w.active != nil {
		w.active()
	}
	if w.pending != nil {
		w.pending.cancel()
	}
	w.pending = &job
	start := !w.running
	w.running = true
	w.mu.Unlock()
	if start {
		go w.run()
	}
}

func (w *choiceSearchWorker) cancel() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.active != nil {
		w.active()
	}
	if w.pending != nil {
		w.pending.cancel()
	}
	w.pending, w.index = nil, nil
	select {
	case <-w.mailbox:
	default:
	}
}

func (w *choiceSearchWorker) run() {
	for {
		w.mu.Lock()
		if w.pending == nil {
			w.running, w.active = false, nil
			w.mu.Unlock()
			return
		}
		job, index := *w.pending, w.index
		w.pending, w.active = nil, job.cancel
		w.mu.Unlock()

		if runtime.GOOS == "js" {
			job.ctx = &choiceWorkerContext{Context: job.ctx}
		}
		result := w.work(job, index)

		w.mu.Lock()
		valid := job.ctx.Err() == nil
		// Reuse completed normalization when only the query was superseded.
		if result.index != nil && (valid || w.pending != nil && w.pending.key.data == job.key.data && w.pending.key.fields == job.key.fields) {
			w.index = result.index
		}
		if valid {
			select {
			case <-w.mailbox:
			default:
			}
			w.mailbox <- result
		}
		job.cancel()
		w.active = nil
		w.mu.Unlock()
		if valid {
			job.services.Invalidate()
		}
	}
}

func (s *choiceSearchCache) bind(c *ui.Context, host ui.Element, unmount func()) {
	mount := ui.Local(host, "nui-choice-search-mount", func() byte { return 0 })
	if s.mount != mount {
		s.reset()
		s.mount = mount
	}
	s.services = c.Services()
	host.OnUnmount(func() {
		if s.mount == mount {
			s.reset()
			s.mount = nil
			if unmount != nil {
				unmount()
			}
		}
	})
}

func (s *choiceSearchCache) pause() {
	if s.requested {
		s.ticket++
		if s.worker != nil {
			s.worker.cancel()
		}
	}
	s.requested, s.pending, s.known = false, false, false
}

func (s *choiceSearchCache) reset() {
	s.pause()
	s.request = choiceSearchKey{}
	s.index, s.indices, s.result = nil, nil, choiceSearchResult{}
}

func (s *choiceSearchCache) ready(query string) bool {
	return !s.pending && (!s.requested || choiceQuery(s.request.data.kind, query) == choiceQuery(s.request.data.kind, s.request.query))
}

func (s *choiceSearchCache) publish(result choiceSearchResult) {
	s.result, s.index, s.indices = result, result.index, result.indices
	s.query, s.known, s.pending = choiceQuery(s.request.data.kind, s.request.query), true, false
	s.revision++
}

func (s *choiceSearchCache) execute(data *choiceSearchData, query string, fields []string, execution SearchExecution) bool {
	background := execution.background(data.candidates)
	key := choiceSearchKey{data: data, fields: choiceFields(data.kind, fields), query: query, execution: execution}
	before := s.revision
	if !s.requested || s.request != key {
		s.ticket++
		s.request, s.requested, s.pending, s.known = key, true, background, false
		ctx, cancel := context.WithCancel(context.Background())
		job := choiceSearchJob{key: key, ticket: s.ticket, ctx: ctx, cancel: cancel, services: s.services}
		if background {
			if s.worker == nil {
				s.worker = newChoiceSearchWorker()
			}
			s.worker.submit(job)
		} else {
			if s.worker != nil {
				s.worker.cancel()
			}
			s.publish(runChoiceSearch(job, s.index))
			cancel()
		}
	}
	if s.worker != nil {
		select {
		case result := <-s.worker.mailbox:
			if s.pending && result.ticket == s.ticket {
				s.publish(result)
			}
		default:
		}
	}
	return before != s.revision
}
