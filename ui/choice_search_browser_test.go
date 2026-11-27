//go:build js && wasm

package nui

import (
	"context"
	"strings"
	"syscall/js"
	"testing"
)

func TestChoiceSearchWorkerYieldsToJavaScript(t *testing.T) {
	data := &choiceSearchData{kind: searchListbox, candidates: 30000, records: make([]choiceSearchRecord, 30000)}
	label := strings.Repeat("Straße café Ａ ", 100)
	for i := range data.records {
		data.records[i].label = label
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	worker := newChoiceSearchWorker()
	defer worker.cancel()
	tick := make(chan bool, 1)
	timer := js.FuncOf(func(js.Value, []js.Value) any {
		worker.mu.Lock()
		running := worker.running
		worker.mu.Unlock()
		cancel()
		tick <- running
		return nil
	})
	defer timer.Release()
	worker.submit(choiceSearchJob{key: choiceSearchKey{data: data, fields: 1, query: "cafe", execution: SearchBackground}, ctx: ctx, cancel: cancel, ticket: 1})
	js.Global().Call("setTimeout", timer, 0)
	if !<-tick {
		t.Fatal("background normalization monopolized the JavaScript event loop until completion")
	}
	waitChoiceWorker(t, worker)
	select {
	case <-worker.mailbox:
		t.Fatal("JavaScript cancellation arrived too late to reject the result")
	default:
	}
}
