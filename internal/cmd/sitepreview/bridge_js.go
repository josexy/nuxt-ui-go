//go:build js && wasm

package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/egoist/mygo"
	"github.com/josexy/nuxt-ui-go/internal/showcase"
	"github.com/josexy/nuxt-ui-go/internal/showcase/preview"
)

type request struct {
	Type      string          `json:"type"`
	ID        int             `json:"id"`
	Component string          `json:"component"`
	Reset     bool            `json:"reset"`
	Config    showcase.Config `json:"config"`
}

func send(value map[string]any) {
	value["channel"] = "nuxt-ui-go"
	encoded, err := json.Marshal(value)
	if err != nil {
		return
	}
	parent := js.Global().Get("parent")
	parent.Call("postMessage", js.Global().Get("JSON").Call("parse", string(encoded)), js.Global().Get("location").Get("origin"))
}

func connect(a *app) {
	report := func(name string, args []string) { send(map[string]any{"type": "event", "name": name, "args": args}) }
	a.example.SetReport(report)
	listener := js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) == 0 {
			return nil
		}
		event := args[0]
		if event.Get("origin").String() != js.Global().Get("location").Get("origin").String() || !event.Get("source").Equal(js.Global().Get("parent")) {
			return nil
		}
		data := event.Get("data")
		if data.Type() != js.TypeObject || data.IsNull() || data.Get("channel").String() != "nuxt-ui-go" {
			return nil
		}
		var req request
		if err := json.Unmarshal([]byte(js.Global().Get("JSON").Call("stringify", data).String()), &req); err != nil {
			return nil
		}
		if req.Type != "configure" || req.Component != a.demo.ID {
			return nil
		}
		cfg, err := a.demo.Normalize(req.Config)
		if err != nil {
			send(map[string]any{"type": "error", "id": req.ID, "message": err.Error()})
			return nil
		}
		snippet, source, err := a.demo.Source(cfg)
		if err != nil {
			send(map[string]any{"type": "error", "id": req.ID, "message": err.Error()})
			return nil
		}
		a.window.Update(func() {
			if a.closed {
				return
			}
			if req.Reset {
				next, err := preview.New(a.demo.ID)
				if err != nil {
					send(map[string]any{"type": "error", "id": req.ID, "message": err.Error()})
					return
				}
				a.example.Dispose()
				a.example = next
				a.example.SetWindow(a.window)
				a.example.SetReport(report)
			}
			if err := a.example.Configure(cfg); err != nil {
				send(map[string]any{"type": "error", "id": req.ID, "message": err.Error()})
				return
			}
			a.config = cfg
			mygo.App.SetLocale(cfg.Locale)
			send(map[string]any{"type": "applied", "id": req.ID, "snippet": snippet, "source": source})
		})
		return nil
	})
	js.Global().Call("addEventListener", "message", listener)
	a.window.OnClosed(func() { js.Global().Call("removeEventListener", "message", listener); listener.Release() })
	send(map[string]any{"type": "ready", "component": a.demo.ID})
}
