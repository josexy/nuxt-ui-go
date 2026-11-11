// Package nativeprobe exercises the built gallery through MyGo's opt-in socket.
package nativeprobe

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/png"
	"math"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

type node struct {
	Description            string
	Min, Max, Now          *float64
	ID, Role, Label, Value string
	Key                    any
	States                 []string
	Actions                []string
	Bounds                 struct{ X, Y, Width, Height float64 }
}
type tree struct {
	Focus string
	Nodes []node
}
type client struct {
	t      *testing.T
	conn   net.Conn
	dec    *json.Decoder
	enc    *json.Encoder
	window int
}

func (c *client) call(method string, params map[string]any, out any) {
	c.t.Helper()
	if params == nil {
		params = map[string]any{}
	}
	if method != "windows" {
		params["window"] = c.window
	}
	if err := c.conn.SetDeadline(time.Now().Add(10 * time.Second)); err != nil {
		c.t.Fatal(err)
	}
	if err := c.enc.Encode(map[string]any{"id": 1, "method": method, "params": params}); err != nil {
		c.t.Fatal(err)
	}
	var reply struct {
		Result json.RawMessage
		Error  *struct{ Code, Message string }
	}
	if err := c.dec.Decode(&reply); err != nil {
		c.t.Fatal(method, err)
	}
	if reply.Error != nil {
		c.t.Fatalf("%s: %+v", method, reply.Error)
	}
	if out != nil {
		if err := json.Unmarshal(reply.Result, out); err != nil {
			c.t.Fatal(err)
		}
	}
}
func (c *client) idle() {
	// Windows only schedules the first frame of a hidden window. Capture uses
	// the public forced-frame path to flush pending changes. It leaves a
	// presentation pending until shown, so idle is only meaningful when visible.
	if os.Getenv("NUI_NATIVE_HIDDEN") == "1" {
		c.call("capture", nil, nil)
		return
	}
	c.call("idle", map[string]any{"timeout": 5000}, nil)
}
func (c *client) tree() tree { c.idle(); var tr tree; c.call("tree", nil, &tr); return tr }
func (c *client) find(label string) node {
	c.t.Helper()
	nodes := c.tree().Nodes
	for _, n := range nodes {
		if n.Label == label && n.Role != "text" && n.Role != "group" {
			return n
		}
	}
	for _, n := range nodes {
		if n.Label == label {
			return n
		}
	}
	labels := []string{}
	for _, n := range nodes {
		labels = append(labels, n.Label)
	}
	c.t.Fatalf("window %d: missing %q in %v", c.window, label, labels)
	return node{}
}
func (c *client) press(label string) {
	n := c.find(label)
	for _, candidate := range c.tree().Nodes {
		if n.Role == "listItem" && !slices.Contains(n.Actions, "press") && candidate.Label == label && slices.Contains(candidate.Actions, "press") {
			n = candidate
			break
		}
	}
	c.call("act", map[string]any{"id": n.ID, "action": "scrollIntoView"}, nil)
	c.idle()
	if slices.Contains(n.States, "focusable") {
		c.call("act", map[string]any{"id": n.ID, "action": "focus"}, nil)
		c.idle()
	}
	c.call("act", map[string]any{"id": n.ID, "action": "press"}, nil)
	c.idle()
}
func (c *client) key(key string) {
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "key", "key": key}}}, nil)
	c.idle()
}

func (c *client) edit(label, value string) {
	n := c.find(label)
	c.call("act", map[string]any{"id": n.ID, "action": "focus"}, nil)
	c.idle()
	c.key("Ctrl+A")
	c.key("Backspace")
	if value != "" {
		c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": value}}}, nil)
		c.idle()
	}
}
func (c *client) navigate(id, label string) {
	if !c.has("Search components") {
		c.press("Open component directory")
	}
	search := c.find("Search components")
	// Directory lookup is test setup. Set its filter directly so an unrelated
	// editor's focus cannot turn navigation into typing in the component under test.
	c.call("act", map[string]any{"id": search.ID, "action": "setValue", "text": id}, nil)
	c.idle()
	search = c.find("Search components")
	for _, n := range c.tree().Nodes {
		if n.Label == label && n.Role == "button" && n.Bounds.Y > search.Bounds.Y && n.Bounds.X < search.Bounds.X+search.Bounds.Width {
			c.call("act", map[string]any{"id": n.ID, "action": "scrollIntoView"}, nil)
			c.idle()
			c.call("act", map[string]any{"id": n.ID, "action": "focus"}, nil)
			c.idle()
			c.call("act", map[string]any{"id": n.ID, "action": "press"}, nil)
			c.idle()
			if c.has("Search components") {
				c.call("act", map[string]any{"id": c.find("Search components").ID, "action": "setValue", "text": ""}, nil)
				c.idle()
			}
			return
		}
	}
	c.t.Fatalf("directory is missing %s (%s): filter=%q focus=%s", label, id, search.Value, c.tree().Focus)
}
func (c *client) has(label string) bool {
	for _, n := range c.tree().Nodes {
		if n.Label == label {
			return true
		}
	}
	return false
}

func (c *client) checkNestedSplitters() {
	c.navigate("splitter", "Splitter")
	left, top, bottom := c.find("Left pane").Bounds, c.find("Right · top").Bounds, c.find("Right · bottom").Bounds
	if top.X <= left.X+left.Width || top.Y != left.Y || bottom.X != top.X || bottom.Y <= top.Y+top.Height {
		c.t.Fatal("native Splitter panes are not left + stacked right")
	}
	outer, inner := c.find("horizontal pane widths").Bounds, c.find("horizontal pane heights").Bounds
	x, y := inner.X-.5, outer.Y+outer.Height/2
	for _, event := range []map[string]any{{"type": "pointerDown", "x": x, "y": y}, {"type": "pointerMove", "x": x + 30, "y": y}, {"type": "pointerUp", "x": x + 30, "y": y}} {
		c.call("input", map[string]any{"events": []any{event}}, nil)
		c.idle()
	}
	wider := c.find("horizontal pane heights").Bounds
	if wider.X <= inner.X || wider.Y != inner.Y || wider.Height != inner.Height {
		c.capture("splitter-drag-failure")
		c.t.Fatalf("native outer Splitter did not resize independently: outer=%+v before=%+v after=%+v drag=(%.2f, %.2f)", outer, inner, wider, x, y)
	}
	c.key("Right")
	inner = c.find("horizontal pane heights").Bounds
	focused := false
	for _, n := range c.tree().Nodes {
		if n.Role == "splitter" && n.Bounds.Width > n.Bounds.Height && n.Bounds.X >= inner.X && n.Bounds.Y >= inner.Y && n.Bounds.Y < inner.Y+inner.Height {
			c.call("act", map[string]any{"id": n.ID, "action": "focus"}, nil)
			focused = true
			break
		}
	}
	if !focused {
		c.t.Fatal("native inner Splitter is missing")
	}
	before := c.find("Right · bottom").Bounds
	c.key("Down")
	if after := c.find("Right · bottom").Bounds; after.Y <= before.Y || c.find("horizontal pane heights").Bounds.X != inner.X {
		c.t.Fatal("native inner Splitter did not resize independently")
	}
	c.edit("horizontal pane draft", "嵌套分栏预览")
	c.find("嵌套分栏预览")
	c.capture("splitter-nested")
	c.navigate("button", "Button")
}

func (c *client) checkNarrowDirectory() {
	c.press("Open component directory")
	c.key("Escape")
	if c.tree().Focus != c.find("Open component directory").ID {
		c.t.Fatal("directory Escape did not return focus")
	}
	c.navigate("input", "Input")
	if c.has("Search components") {
		c.t.Fatal("directory remained open after navigation")
	}
	c.edit("Workspace name", "窄窗口草稿")
	c.navigate("button", "Button")
	c.press("Save changes")
	c.find("Saved 1 times")
	c.navigate("input", "Input")
	if c.find("Workspace name").Value != "窄窗口草稿" {
		c.t.Fatal("narrow navigation lost draft")
	}
	c.capture("input-narrow-draft")
	c.press("Open component directory")
	c.edit("Search components", "")
	first := c.find("App")
	c.call("act", map[string]any{"id": first.ID, "action": "focus"}, nil)
	c.key("End")
	c.key("Enter")
	c.find("Task 00001")
	c.press("Open component directory")
	last := c.find("Task browser")
	c.call("act", map[string]any{"id": last.ID, "action": "focus"}, nil)
	c.key("Home")
	c.key("Enter")
	c.find("Light and dark")
	for _, item := range []struct{ id, label string }{{"modal", "Modal"}, {"table", "Table"}, {"chart", "Chart"}, {"sidebar", "Sidebar"}} {
		c.navigate(item.id, item.label)
		switch item.id {
		case "modal":
			c.press("Edit workspace")
			c.press("Dialog plan")
			c.key("Escape")
			c.key("Escape")
		case "table":
			c.press("Select first record")
		case "chart":
			c.press("Update data")
		case "sidebar":
			for _, mode := range []string{"slideover", "modal", "drawer"} {
				label := "Open inner " + mode + " sidebar"
				c.press(label)
				c.key("Escape")
				if c.tree().Focus != c.find(label).ID {
					c.t.Fatalf("%s sidebar did not restore focus", mode)
				}
			}
		}
		c.capture(item.id + "-narrow")
	}
	c.checkNestedSplitters()
	c.press("Open component directory")
	c.edit("Search components", "no-such-component")
	c.find("No matching components. Clear the search to browse all.")
	c.capture("directory-narrow-empty")
	c.key("Escape")
}
func (c *client) checkSidebar() {
	c.press("Workspace settings")
	c.find("Access permissions and sharing")
	c.capture("shell-compact-submenu")
	c.press("Access permissions and sharing")
	c.find("Selected: permissions")
	c.key("Escape")
	c.press("Toggle inner slideover sidebar")
	handle := c.find("Resize inner slideover sidebar")
	initial := handle.Bounds.X
	c.call("act", map[string]any{"id": handle.ID, "action": "focus"}, nil)
	c.key("Right")
	if n := c.find("Resize inner slideover sidebar"); math.Abs(n.Bounds.X-initial-10) > .1 {
		c.t.Fatalf("sidebar keyboard resize: %+v", n)
	}
	handle = c.find("Resize inner slideover sidebar")
	x, y := handle.Bounds.X+handle.Bounds.Width/2, handle.Bounds.Y+handle.Bounds.Height/2
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "pointerDown", "x": x, "y": y}}}, nil)
	c.idle()
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "pointerMove", "x": x + 40, "y": y}}}, nil)
	c.idle()
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "pointerUp", "x": x + 40, "y": y}}}, nil)
	if n := c.find("Resize inner slideover sidebar"); math.Abs(n.Bounds.X-initial-50) > .1 {
		c.t.Fatalf("sidebar pointer resize: %+v", n)
	}
	c.capture("shell-resized")
	handle = c.find("Resize inner slideover sidebar")
	x = handle.Bounds.X + handle.Bounds.Width/2
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "click", "clicks": 2, "x": x, "y": y}}}, nil)
	if n := c.find("Resize inner slideover sidebar"); math.Abs(n.Bounds.X-initial) > .1 {
		c.t.Fatalf("sidebar double-click reset: %+v", n)
	}
	c.press("Toggle inner slideover sidebar")
}

func (c *client) checkGallerySidebar() {
	if width := c.find("Gallery sidebar").Bounds.Width; math.Abs(width-256) > .1 {
		c.t.Fatalf("Gallery initial width: %v", width)
	}
	handle := c.find("Resize gallery sidebar")
	c.call("act", map[string]any{"id": handle.ID, "action": "focus"}, nil)
	c.key("Right")
	if width := c.find("Gallery sidebar").Bounds.Width; math.Abs(width-266) > .1 {
		c.t.Fatalf("Gallery keyboard resize: %v", width)
	}
	handle = c.find("Resize gallery sidebar")
	x, y := handle.Bounds.X+handle.Bounds.Width/2, handle.Bounds.Y+handle.Bounds.Height/2
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "pointerDown", "x": x, "y": y}}}, nil)
	c.idle()
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "pointerMove", "x": x + 34, "y": y}}}, nil)
	c.idle()
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "pointerUp", "x": x + 34, "y": y}}}, nil)
	if width := c.find("Gallery sidebar").Bounds.Width; math.Abs(width-300) > .1 {
		c.t.Fatalf("Gallery pointer resize: %v", width)
	}
	c.press("Toggle gallery sidebar")
	if width := c.find("Gallery sidebar").Bounds.Width; math.Abs(width-64) > .1 {
		c.t.Fatalf("Gallery compact width: %v", width)
	}
	c.press("Input")
	c.find("Workspace name")
	c.capture("gallery-sidebar-compact")
	c.press("Toggle gallery sidebar")
	if width := c.find("Gallery sidebar").Bounds.Width; math.Abs(width-300) > .1 {
		c.t.Fatalf("Gallery page switch lost width: %v", width)
	}
	c.capture("gallery-sidebar-resized")
	handle = c.find("Resize gallery sidebar")
	x = handle.Bounds.X + handle.Bounds.Width/2
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "click", "clicks": 2, "x": x, "y": y}}}, nil)
	if width := c.find("Gallery sidebar").Bounds.Width; math.Abs(width-256) > .1 {
		c.t.Fatalf("Gallery double-click reset: %v", width)
	}
	c.navigate("button", "Button")
	c.press("App")
	c.key("End")
	c.key("Enter")
	c.find("Task 00001")
	c.press("Task browser")
	c.key("Home")
	c.key("Enter")
	c.find("Light and dark")
	c.navigate("button", "Button")
}

func (c *client) theme(mode string) {
	c.press("Theme")
	c.press(mode)
}

// Exercise the palette menus through native accessibility actions and rendering.
func (c *client) checkThemePalettes() {
	c.press("Theme")
	c.press("Primary palette")
	if n := c.find("Green"); n.Role != "menuItemRadio" || !slices.Contains(n.States, "checked") {
		c.t.Fatalf("default primary selection: %+v", n)
	}
	c.capture("theme-primary-menu-light")
	c.press("Violet")
	c.press("Theme")
	c.press("Neutral palette")
	c.press("Mauve")
	c.theme("Dark")
	c.press("Theme")
	c.press("Neutral palette")
	if n := c.find("Mauve"); n.Role != "menuItemRadio" || !slices.Contains(n.States, "checked") {
		c.t.Fatalf("selected neutral after mode change: %+v", n)
	}
	c.capture("theme-neutral-menu-dark")
	c.press("Mauve")
	c.press("Theme")
	c.press("Primary palette")
	if n := c.find("Violet"); !slices.Contains(n.States, "checked") {
		c.t.Fatalf("primary lost after neutral/mode change: %+v", n)
	}
	c.press("Black")
	c.capture("theme-black-dark")
	c.theme("Light")
	c.capture("theme-black-light")
	// Restore the defaults used by the remaining Gallery probes.
	c.press("Theme")
	c.press("Primary palette")
	c.press("Green")
	c.press("Theme")
	c.press("Neutral palette")
	c.press("Slate")
}

func (c *client) capture(name string) []byte {
	c.t.Helper()
	c.idle()
	var result struct {
		PNG           string
		Width, Height int
	}
	c.call("capture", nil, &result)
	data, err := base64.StdEncoding.DecodeString(result.PNG)
	if err != nil {
		c.t.Fatal(err)
	}
	if result.Width < 100 || result.Height < 100 {
		c.t.Fatal("empty capture")
	}
	if dir := os.Getenv("NUI_NATIVE_ARTIFACTS"); dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			c.t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name+".png"), data, 0644); err != nil {
			c.t.Fatal(err)
		}
		tr, _ := json.MarshalIndent(c.tree(), "", "  ")
		if err := os.WriteFile(filepath.Join(dir, name+".json"), tr, 0644); err != nil {
			c.t.Fatal(err)
		}
	}
	c.t.Logf("%s: %dx%d", name, result.Width, result.Height)
	return data
}

func (c *client) checkListboxStates() {
	c.navigate("listbox", "Listbox")
	empty := c.find("No members")
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "wheel", "x": empty.Bounds.X - 10, "y": 100, "dx": 0, "dy": 2000, "precise": true}}}, nil)
	if slices.Contains(c.find("No members").States, "offscreen") {
		c.t.Fatal("empty member message was not visible")
	}
	c.find("Empty list")
	c.find("Disabled list")
	list := c.find("Disabled members list")
	if !slices.Contains(list.States, "disabled") {
		c.t.Fatal("disabled member list was enabled")
	}
	clicked := false
	for _, n := range c.tree().Nodes {
		if n.Role == "listItem" && n.Label == "Ada Lovelace" && slices.Contains(n.States, "disabled") && !slices.Contains(n.States, "offscreen") && n.Bounds.Y >= list.Bounds.Y && n.Bounds.Y+n.Bounds.Height <= list.Bounds.Y+list.Bounds.Height {
			c.call("input", map[string]any{"events": []any{map[string]any{"type": "click", "x": n.Bounds.X + n.Bounds.Width/2, "y": n.Bounds.Y + n.Bounds.Height/2}}}, nil)
			clicked = true
			break
		}
	}
	if !clicked {
		c.t.Fatal("disabled member row was not visible")
	}
	c.key("Down")
	c.key("Enter")
	c.key("Space")
	c.find("Disabled selected IDs: [lin] · 0 changes")
	c.capture("listbox-empty-disabled")
}

func (c *client) checkCarouselParity() {
	c.navigate("carousel", "Carousel")
	c.edit("Contract one draft", "saved carousel draft")
	c.press("Customize carousel buttons")
	prev, next, viewport := c.find("Previous Contract tour"), c.find("Next Contract tour"), c.find("Contract tour slides")
	center := viewport.Bounds.Y + viewport.Bounds.Height/2
	if math.Abs(prev.Bounds.Y+prev.Bounds.Height/2-center) > .5 || math.Abs(next.Bounds.Y+next.Bounds.Height/2-center) > .5 {
		c.t.Fatal("custom carousel arrows are not centered")
	}
	c.press("Next Contract tour")
	c.find("Carousel index: 1 · Changes: 1 · Paused: false")
	c.press("Remove contract slide two")
	if c.tree().Focus != c.find("Contract tour slides").ID {
		c.t.Fatal("removed slide did not restore viewport focus")
	}
	c.find("Carousel index: 1 · Changes: 1 · Paused: false")
	c.press("Previous Contract tour")
	if c.find("Contract one draft").Value != "saved carousel draft" {
		c.t.Fatal("carousel navigation lost draft")
	}
	c.press("Previous button avatar")
	c.press("RTL contract carousel")
	c.capture("carousel-contract-rtl-custom")
	c.press("Vertical contract carousel")
	prev, next, viewport = c.find("Previous Contract tour"), c.find("Next Contract tour"), c.find("Contract tour slides")
	center = viewport.Bounds.X + viewport.Bounds.Width/2
	if math.Abs(prev.Bounds.X+prev.Bounds.Width/2-center) > .5 || math.Abs(next.Bounds.X+next.Bounds.Width/2-center) > .5 {
		c.t.Fatal("vertical arrows are not centered")
	}
	c.theme("Dark")
	c.call("act", map[string]any{"id": c.find("Contract tour slides").ID, "action": "scrollIntoView"}, nil)
	c.capture("carousel-contract-vertical-dark")
	c.theme("Light")
	c.press("Disable contract carousel")
	if !slices.Contains(c.find("Next Contract tour").States, "disabled") {
		c.t.Fatal("carousel parent disabling did not reach buttons")
	}
	c.press("Disable contract carousel")
	c.press("Empty contract slides")
	c.find("No slides")
	c.press("Restore contract slides")
	c.find("Contract one draft")
}

func (c *client) checkMarqueeSplitterParity() {
	focus := func(label string) {
		n := c.find(label)
		c.call("act", map[string]any{"id": n.ID, "action": "scrollIntoView"}, nil)
		c.idle()
		c.call("act", map[string]any{"id": n.ID, "action": "focus"}, nil)
		c.idle()
	}
	c.navigate("marquee", "Marquee")
	c.press("Marquee copies 0")
	if c.has("Marquee release 0") {
		c.t.Fatal("zero marquee copies retained content")
	}
	c.press("Marquee copies 2")
	focus("Contract marquee")
	c.key("Space")
	c.find("Copies: 2 · Paused: true · Content: 0")
	c.press("Replace marquee content")
	c.find("Marquee release 1")
	c.press("Style marquee content")
	c.call("act", map[string]any{"id": c.find("Contract marquee").ID, "action": "scrollIntoView"}, nil)
	c.capture("marquee-contract-horizontal")
	c.press("Vertical contract marquee")
	c.press("RTL contract marquee")
	c.press("Fade contract marquee")
	c.theme("Dark")
	c.call("act", map[string]any{"id": c.find("Contract marquee").ID, "action": "scrollIntoView"}, nil)
	c.capture("marquee-contract-vertical-dark")
	c.theme("Light")
	c.press("Disable marquee parent")
	if !slices.Contains(c.find("Contract marquee").States, "disabled") {
		c.t.Fatal("marquee ancestor disabling not exposed")
	}
	c.navigate("splitter", "Splitter")
	c.press("Lock contract divider")
	if !slices.Contains(c.find("Contract divider").States, "disabled") {
		c.t.Fatal("splitter divider did not lock")
	}
	c.edit("Contract pane editor", "Native editable pane")
	if c.find("Contract pane editor").Value != "Native editable pane" {
		c.t.Fatal("locked divider blocked editor")
	}
	c.press("Lock contract divider")
	focus("Contract divider")
	c.key("Right")
	c.find("Splitter value: 45 · Changes: 1")
	c.key("End")
	c.key("Right")
	c.find("Splitter value: 70 · Changes: 2")
	c.press("RTL contract splitter")
	focus("Contract divider")
	c.key("Right")
	c.find("Splitter value: 65 · Changes: 3")
	c.press("Style contract divider")
	c.call("act", map[string]any{"id": c.find("Contract panes").ID, "action": "scrollIntoView"}, nil)
	c.capture("splitter-contract-rtl")
	c.press("Vertical contract splitter")
	focus("Contract divider")
	c.key("Down")
	c.find("Splitter value: 70 · Changes: 4")
	c.press("Zero splitter arrow step")
	focus("Contract divider")
	c.key("Up")
	c.find("Splitter value: 70 · Changes: 4")
	if c.find("Contract pane editor").Value != "Native editable pane" {
		c.t.Fatal("splitter layout change lost draft")
	}
	c.theme("Dark")
	c.call("act", map[string]any{"id": c.find("Contract panes").ID, "action": "scrollIntoView"}, nil)
	c.capture("splitter-contract-vertical-dark")
	c.theme("Light")
	c.press("Disable splitter parent")
	if !slices.Contains(c.find("Contract pane editor").States, "disabled") {
		c.t.Fatal("splitter ancestor disabling not inherited")
	}
	c.press("Set splitter to maximum")
	c.find("Splitter value: 70 · Changes: 4")
}

func (c *client) checkLayoutParity() {
	keepFocusClick := func(label string) {
		n := c.find(label)
		x, y := n.Bounds.X+n.Bounds.Width/2, n.Bounds.Y+n.Bounds.Height/2
		for _, kind := range []string{"pointerDown", "pointerUp"} {
			c.call("input", map[string]any{"events": []any{map[string]any{"type": kind, "x": x, "y": y}}}, nil)
			c.idle()
		}
	}
	c.navigate("sidebar", "Sidebar")
	if os.Getenv("NUI_NATIVE_NARROW") == "1" {
		for _, mode := range []string{"slideover", "modal", "drawer"} {
			c.press("Open inner " + mode + " sidebar")
			c.find("Project navigation")
			c.press(mode + " sidebar header action")
			c.edit(mode+" sidebar draft", "Native sidebar draft")
			c.capture("sidebar-" + mode + "-regions")
			c.key("Escape")
			if c.tree().Focus != c.find("Open inner "+mode+" sidebar").ID {
				c.t.Fatal("sidebar overlay did not return focus to its opener")
			}
			c.press("Open inner " + mode + " sidebar")
			if c.find(mode+" sidebar draft").Value != "Native sidebar draft" {
				c.t.Fatal("sidebar reopen lost draft")
			}
			c.key("Escape")
		}
	} else {
		c.press("slideover sidebar header action")
		c.edit("slideover sidebar draft", "Native sidebar draft")
		c.press("Toggle inner slideover sidebar")
		c.press("Toggle inner slideover sidebar")
		if c.find("slideover sidebar draft").Value != "Native sidebar draft" {
			c.t.Fatal("sidebar collapse lost draft")
		}
		c.capture("sidebar-desktop-regions")
	}
	c.navigate("header", "Header")
	c.edit("Layout header draft", "Native header draft")
	editor := c.find("Layout header draft")
	if c.tree().Focus != editor.ID {
		c.t.Fatal("header editor did not receive focus")
	}
	keepFocusClick("Toggle header leading")
	if c.tree().Focus != editor.ID || c.find("Layout header draft").Value != "Native header draft" {
		c.t.Fatal("header leading update lost editor identity")
	}
	c.press("Open composed header menu")
	c.press("Header menu settings")
	c.find("Header page: settings")
	if c.tree().Focus != c.find("Open composed header menu").ID {
		c.t.Fatal("header-composed sidebar did not return focus")
	}
	c.press("Style layout parts")
	c.press("Long layout text")
	c.press("RTL layout parts")
	c.capture("header-parts-rtl")
	c.navigate("footer", "Footer")
	c.press("Style layout parts")
	c.press("Long layout text")
	c.press("RTL layout parts")
	c.press("Contract footer action")
	c.find("Layout actions: 1")
	c.capture("footer-parts-rtl")
	c.navigate("footer-columns", "FooterColumns")
	c.press("Contract examples")
	c.press("Cancel footer navigation")
	c.press("Contract documentation")
	c.press("Contract unavailable")
	c.find("Footer active: examples · Selects: 2 · Navigations: 1")
	c.press("Contract examples")
	link := c.find("Contract examples")
	keepFocusClick("Reverse footer columns")
	if c.tree().Focus != link.ID {
		c.t.Fatal("footer column reorder lost link identity")
	}
	c.press("Style footer links")
	c.press("RTL footer columns")
	c.press("Narrow footer columns")
	if c.find("Product").Bounds.Y <= c.find("Resources").Bounds.Y {
		c.t.Fatal("narrow footer columns did not wrap in data order")
	}
	c.theme("Dark")
	c.call("act", map[string]any{"id": c.find("Contract documentation").ID, "action": "scrollIntoView"}, nil)
	c.capture("footer-columns-narrow-rtl")
	c.theme("Light")
}

func (c *client) checkChartParity() {
	c.navigate("chart", "Chart")
	chart := c.find("Contract chart")
	c.call("act", map[string]any{"id": chart.ID, "action": "focus"}, nil)
	c.key("Home")
	c.key("Enter")
	c.find("Chart focus: 1 · Selections: 1 · Legend changes: 0")
	c.press("Contract series one")
	c.press("Reverse contract series")
	if slices.Contains(c.find("Contract series one").States, "checked") {
		c.t.Fatal("series visibility lost during reorder")
	}
	c.press("Contract series one")
	c.press("Single contract series")
	if c.has("Contract series one") {
		c.t.Fatal("automatic legend shown for one series")
	}
	c.press("Always show chart legend")
	c.find("Contract series one")
	c.press("Single contract series")
	c.press("Show all line points")
	c.press("Custom chart palette")
	c.press("RTL contract chart")
	chart = c.find("Contract chart")
	c.call("act", map[string]any{"id": chart.ID, "action": "scrollIntoView"}, nil)
	c.call("act", map[string]any{"id": chart.ID, "action": "focus"}, nil)
	c.key("Home")
	c.key("Left")
	c.capture("chart-contract-rtl-line")
	c.press("Show chart X axis")
	if c.has("Review month") {
		c.t.Fatal("hidden X axis retained title")
	}
	c.press("Show chart X axis")
	for _, kind := range []string{"Bar", "Area", "Scatter", "Pie", "Donut"} {
		c.press("Contract chart kind")
		c.press("Contract " + kind)
		c.capture("chart-contract-" + strings.ToLower(kind))
	}
	c.press("Show chart center")
	c.find("Review total")
	c.theme("Dark")
	c.capture("chart-contract-center-dark")
	c.theme("Light")
	c.press("Empty contract chart")
	c.find("No chart data")
	c.press("Empty contract chart")
}

func (c *client) checkFileUploadParity() {
	c.navigate("file-upload", "FileUpload")
	c.press("Contract attachments")
	c.find("File changes: 1 · Rejected: 1")
	c.press("Contract attachments")
	c.find("File changes: 1 · Rejected: 1")
	c.press("Choose result: cancel")
	c.press("Contract attachments")
	c.find("File changes: 1 · Rejected: 1")
	c.press("Place files inside")
	c.press("Style contract file parts")
	c.press("Remove notes.txt")
	if c.tree().Focus != c.find("Contract attachments").ID {
		c.t.Fatal("file removal lost picker focus")
	}
	c.find("File changes: 2 · Rejected: 1")
	c.capture("file-upload-inside-light")
	c.press("Disable contract uploader")
	if !slices.Contains(c.find("Contract attachments").States, "disabled") {
		c.t.Fatal("upload area remains enabled")
	}
	c.capture("file-upload-disabled")
	c.press("Disable contract uploader")
	c.press("Interactive contract area")
	c.press("Choose result: error")
	c.press("Choose contract files")
	c.find("sample selection failed")
	c.find("File changes: 2 · Rejected: 1")
	c.press("Contract file grid")
	c.press("Show file avatar")
	c.press("Show upload icon")
	c.theme("Dark")
	c.call("act", map[string]any{"id": c.find("Remove application.exe").ID, "action": "scrollIntoView"}, nil)
	c.capture("file-upload-grid-dark")
	c.theme("Light")
}

func (c *client) checkGalleryStateExamples() {
	c.navigate("select", "Select")
	c.press("Unselected plan")
	c.key("Enter")
	c.find("Empty select")
	c.capture("select-empty-unselected")
	c.navigate("table", "Table")
	message := c.find("No workspace records")
	c.call("act", map[string]any{"id": message.ID, "action": "scrollIntoView"}, nil)
	c.capture("table-empty")
	c.press("Add sample record")
	if c.has("No workspace records") {
		c.t.Fatal("native empty table message survived adding a record")
	}
	c.press("Remove sample record")
	message = c.find("No workspace records")
	c.call("act", map[string]any{"id": message.ID, "action": "scrollIntoView"}, nil)
	c.capture("table-empty-restored")
	c.navigate("header", "Header")
	c.press("constrained layout action")
	c.find("Action completed · 1 actions")
	c.capture("header-constrained")
}

func (c *client) checkCheckboxIcons() {
	c.navigate("checkbox", "Checkbox")
	if got := c.find("Accept terms").Description; got != "Applies to this workspace\nRequired before continuing\nRequired" {
		c.t.Fatalf("checkbox lost its field help, description or required notice: %q", got)
	}
	if n := c.find("Select all files"); n.Role != "checkBox" || !slices.Contains(n.States, "mixed") {
		c.t.Fatalf("tri-state checkbox did not expose mixed semantics: %+v", n)
	}
	c.press("Custom mixed icon")
	if !slices.Contains(c.find("Select all files").States, "mixed") {
		c.t.Fatal("changing the mixed icon changed the selection")
	}
	c.press("Select all files")
	if !slices.Contains(c.find("Select all files").States, "checked") {
		c.t.Fatal("mixed checkbox did not become checked")
	}
	c.press("Reset mixed state")
	if !slices.Contains(c.find("Select all files").States, "mixed") {
		c.t.Fatal("external state reset lost mixed semantics")
	}
	c.capture("checkbox-custom-mixed")
	c.press("Workspace alerts")
	if !slices.Contains(c.find("Workspace alerts").States, "checked") {
		c.t.Fatal("hidden-indicator card did not toggle")
	}
	c.key("Space")
	if slices.Contains(c.find("Workspace alerts").States, "checked") {
		c.t.Fatal("Space did not clear the hidden-indicator card")
	}
	c.press("Icon only choice")
	if !slices.Contains(c.find("Icon only choice").States, "checked") {
		c.t.Fatal("icon-only checkbox lost its accessible control")
	}
	c.capture("checkbox-hidden-icons")
	if !slices.Contains(c.find("Locked checkbox").States, "disabled") {
		c.t.Fatal("disabled Checkbox is interactive")
	}
	c.press("Checkbox ignores Loading")
	c.find("Disabled: true · Shared loading value: true")
	c.navigate("button", "Button")
}

func (c *client) checkSwitchLoading() {
	c.navigate("switch", "Switch")
	if !slices.Contains(c.find("Loading switch").States, "disabled") {
		c.t.Fatal("loading Switch did not expose disabled state")
	}
	c.press("Toggle loading state")
	if slices.Contains(c.find("Loading switch").States, "disabled") {
		c.t.Fatal("completed loading left Switch disabled")
	}
	c.press("Loading switch")
	c.find("Disabled: true · Loading: true")
	c.key("Space")
	c.find("Disabled: true · Loading: false")
	c.find("Busy: false · Changes: 2")
	c.press("Toggle loading state")
	c.find("Busy: true · Changes: 2")
	c.capture("switch-loading-states")
	c.navigate("button", "Button")
}

func (c *client) checkChoiceGroupIcons() {
	for _, radio := range []bool{false, true} {
		id, label := "checkbox-group", "CheckboxGroup"
		if radio {
			id, label = "radio-group", "RadioGroup"
		}
		c.navigate(id, label)
		c.press("Run tests")
		c.key("Right")
		want := "Run tests"
		if radio {
			want = "Build task"
		}
		if c.tree().Focus != c.find(want).ID {
			c.t.Fatalf("%s used the wrong default wrapping policy", label)
		}
		c.press("Loop choices")
		c.press("Run tests")
		c.key("Right")
		want = "Build task"
		if radio {
			want = "Run tests"
		}
		if c.tree().Focus != c.find(want).ID {
			c.t.Fatalf("%s did not apply the explicit wrapping policy", label)
		}
		c.press("Reverse choices")
		if radio {
			c.find("Icon choice: test · Changes: 3")
		} else {
			c.find("Icon choices: [build] · Changes: 2")
		}
		if !slices.Contains(c.find("Deploy task").States, "disabled") {
			c.t.Fatal("disabled choice became interactive after reordering")
		}
		c.capture(id + "-icons-reordered")
	}
	c.navigate("button", "Button")
}

func (c *client) checkContainerPadding() {
	c.navigate("container", "Container")
	outer, inner := c.find("Responsive container"), c.find("Container content boundary")
	want := float64(32)
	if os.Getenv("NUI_NATIVE_NARROW") == "1" {
		want = 24
	}
	if math.Abs(inner.Bounds.X-outer.Bounds.X-want) > 1 {
		c.t.Fatalf("container padding=%g, want %g", inner.Bounds.X-outer.Bounds.X, want)
	}
	c.edit("Container draft", "Preserved container draft")
	c.press("Toggle container padding")
	outer, inner = c.find("Responsive container"), c.find("Container content boundary")
	if math.Abs(inner.Bounds.X-outer.Bounds.X) > 1 || c.find("Container draft").Value != "Preserved container draft" {
		c.t.Fatal("explicit zero padding failed or discarded the draft")
	}
	c.press("Toggle container padding")
	outer, inner = c.find("Responsive container"), c.find("Container content boundary")
	if math.Abs(inner.Bounds.X-outer.Bounds.X-want) > 1 || c.find("Container draft").Value != "Preserved container draft" {
		c.t.Fatal("default padding did not restore independently of the draft")
	}
	c.capture("container-default-padding")
}

func (c *client) checkButtonAvatar() {
	c.navigate("button", "Button")
	c.press("Account with avatar")
	c.key("Space")
	if !c.has("Avatar actions: 2") {
		c.t.Fatal("avatar button mouse/Space count")
	}
	c.press("Load account action")
	if !slices.Contains(c.find("Account with avatar").States, "disabled") {
		c.t.Fatal("avatar button remained enabled while loading")
	}
	c.press("Account with avatar")
	if !c.has("Avatar actions: 2") {
		c.t.Fatal("loading avatar action activated")
	}
	c.press("Load account action")
	c.press("Account avatar")
	if !c.has("Avatar actions: 3") {
		c.t.Fatal("avatar-only action failed")
	}
	c.capture("button-avatar")
	c.navigate("container", "Container")
	c.navigate("button", "Button")
}

func (c *client) checkBadgeAvatar() {
	c.navigate("badge", "Badge")
	if !c.has("Ada Lovelace") {
		c.t.Fatal("badge avatar missing")
	}
	c.press("Override badge avatar with icon")
	if c.has("Ada Lovelace") {
		c.t.Fatal("badge leading icon did not replace avatar")
	}
	c.press("Override badge avatar with icon")
	if !c.has("Ada Lovelace") {
		c.t.Fatal("badge avatar did not return")
	}
	icon := c.find("Icon-only badge")
	// The desktop's text scale changes the icon size, but equal padding keeps
	// a single-icon badge square. Nominal DIP sizes are checked headlessly.
	if icon.Bounds.Width <= 0 || math.Abs(icon.Bounds.Width-icon.Bounds.Height) > 1 {
		c.t.Fatalf("icon-only badge geometry: %+v", icon.Bounds)
	}
	c.call("act", map[string]any{"id": icon.ID, "action": "scrollIntoView"}, nil)
	c.capture("badge-avatar-square")
	c.navigate("button", "Button")
}

func (c *client) checkCardSpacing() {
	c.navigate("card", "Card")
	c.edit("Card team name", "Preserved card draft")
	c.key("Ctrl+A")
	editor := c.find("Card team name")
	for _, show := range []bool{false, true} {
		toggle := c.find("Toggle card title")
		c.call("act", map[string]any{"id": toggle.ID, "action": "scrollIntoView"}, nil)
		toggle = c.find("Toggle card title")
		x, y := toggle.Bounds.X+toggle.Bounds.Width/2, toggle.Bounds.Y+toggle.Bounds.Height/2
		c.call("input", map[string]any{"events": []any{map[string]any{"type": "pointerDown", "x": x, "y": y}, map[string]any{"type": "pointerUp", "x": x, "y": y}}}, nil)
		if c.tree().Focus != editor.ID || c.has("Team settings") != show {
			c.t.Fatal("card title toggle lost body focus or did not update header")
		}
	}
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "Replacement card draft"}}}, nil)
	if c.find("Card team name").Value != "Replacement card draft" {
		c.t.Fatal("card header toggle lost editor selection")
	}
	c.press("Save card")
	if !c.has("Saved: Replacement card draft") {
		c.t.Fatal("card content action failed")
	}
	desc := c.find("Description without title")
	c.call("act", map[string]any{"id": desc.ID, "action": "scrollIntoView"}, nil)
	header, desc := c.find("Description card header"), c.find("Description without title")
	if math.Abs(desc.Bounds.Y-header.Bounds.Y-20) > 1 {
		c.t.Fatalf("description-only top offset=%g", desc.Bounds.Y-header.Bounds.Y)
	}
	c.capture("card-description-spacing")
	c.navigate("button", "Button")
}

func (c *client) checkAlertLayout() {
	c.navigate("alert", "Alert")
	c.edit("Alert action draft", "Persistent draft")
	c.key("Ctrl+A")
	editor := c.find("Alert action draft")
	for _, label := range []string{"Toggle alert avatar", "Toggle alert direction", "Toggle alert avatar", "Toggle alert direction"} {
		toggle := c.find(label)
		c.call("act", map[string]any{"id": toggle.ID, "action": "scrollIntoView"}, nil)
		toggle = c.find(label)
		x, y := toggle.Bounds.X+toggle.Bounds.Width/2, toggle.Bounds.Y+toggle.Bounds.Height/2
		c.call("input", map[string]any{"events": []any{map[string]any{"type": "pointerDown", "x": x, "y": y}, map[string]any{"type": "pointerUp", "x": x, "y": y}}}, nil)
		if c.tree().Focus != editor.ID {
			c.t.Fatalf("%s discarded alert editor focus", label)
		}
	}
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "Replacement alert draft"}}}, nil)
	if c.find("Alert action draft").Value != "Replacement alert draft" {
		c.t.Fatal("alert layout changes discarded editor selection")
	}
	c.press("Close custom notice")
	c.find("Alert close requests: 1")
	c.key("Enter")
	c.find("Alert close requests: 2")
	c.find("Ada Lovelace")
	c.capture("alert-avatar-actions")
	c.navigate("button", "Button")
}

func (c *client) checkAvatarImage() {
	c.navigate("avatar", "Avatar")
	before := c.find("Gradient fixture")
	c.press("Toggle avatar image size")
	root := c.find("Gradient fixture")
	if root.Bounds.Width <= before.Bounds.Width || root.Bounds.Width != root.Bounds.Height {
		c.t.Fatalf("avatar root did not resize: before=%+v after=%+v", before.Bounds, root.Bounds)
	}
	c.capture("avatar-image-resized")
	fallback := c.find("Fallback Avatar")
	c.press("Toggle avatar image")
	c.capture("avatar-fallback-icon")
	c.press("Toggle avatar fallback icon")
	c.capture("avatar-fallback-text")
	c.press("Toggle avatar fallback text")
	if n := c.find("Fallback Avatar"); n.Role != "image" || n.ID != fallback.ID {
		c.t.Fatalf("avatar name lost image semantics: %+v", n)
	}
	c.press("Show avatar chip")
	c.capture("avatar-fallback")
	c.navigate("button", "Button")
}

func (c *client) checkAvatarGroupStyles() {
	c.navigate("avatar-group", "AvatarGroup")
	grace := c.find("Group Grace")
	if c.find("Group Ada").Bounds.Width <= grace.Bounds.Width {
		c.t.Fatal("group size replaced explicit member size")
	}
	c.press("Reverse avatar members")
	c.find("Group Lin")
	if c.has("Group Ada") || c.find("Group Grace").ID != grace.ID {
		c.t.Fatal("group reorder retained wrong members or discarded identity")
	}
	c.press("Toggle avatar member style")
	member := c.find("Group Grace")
	c.call("act", map[string]any{"id": member.ID, "action": "scrollIntoView"}, nil)
	c.capture("avatar-group-item-styles")
	c.navigate("button", "Button")
}

func (c *client) checkChipExamples() {
	c.navigate("chip", "Chip")
	before := c.find("Ada online")
	c.press("Toggle status chip")
	c.find("Chip hidden: true")
	if c.find("Ada online").ID != before.ID {
		c.t.Fatal("hiding chip replaced avatar host")
	}
	c.press("Inset chip markers")
	styled := c.find("Styled standalone chip")
	c.call("act", map[string]any{"id": styled.ID, "action": "scrollIntoView"}, nil)
	c.find("999+")
	c.capture("chip-corners-and-standalone")
	neutral := c.find("Custom neutral chip")
	c.call("act", map[string]any{"id": neutral.ID, "action": "scrollIntoView"}, nil)
	c.capture("chip-neutral-background")
	c.navigate("button", "Button")
}

func (c *client) checkSeparatorContent() {
	c.navigate("separator", "Separator")
	c.find("Separator member")
	c.edit("Separator draft", "Persistent draft")
	c.key("Ctrl+A")
	editor := c.find("Separator draft")
	for round := range 3 {
		for _, label := range []string{"Cycle separator position", "Cycle separator position", "Cycle separator position", "Toggle separator direction", "Toggle separator direction"} {
			toggle := c.find(label)
			c.call("act", map[string]any{"id": toggle.ID, "action": "scrollIntoView"}, nil)
			toggle = c.find(label)
			x, y := toggle.Bounds.X+toggle.Bounds.Width/2, toggle.Bounds.Y+toggle.Bounds.Height/2
			c.call("input", map[string]any{"events": []any{map[string]any{"type": "pointerDown", "x": x, "y": y}, map[string]any{"type": "pointerUp", "x": x, "y": y}}}, nil)
			if state := c.tree(); state.Focus != editor.ID {
				focused := ""
				for _, n := range state.Nodes {
					if n.ID == state.Focus {
						focused = n.Label
					}
				}
				c.capture("separator-focus-failure")
				c.t.Fatalf("round %d %s discarded separator focus: want=%s got=%s (%s), toggle=%+v", round+1, label, editor.ID, state.Focus, focused, toggle.Bounds)
			}
		}
	}
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "Replacement separator draft"}}}, nil)
	if c.find("Separator draft").Value != "Replacement separator draft" {
		c.t.Fatal("separator layout change discarded selection")
	}
	label := c.find("这是一段需要在窄容器中换行的分隔说明文字")
	c.call("act", map[string]any{"id": label.ID, "action": "scrollIntoView"}, nil)
	label = c.find(label.Label)
	if label.Bounds.Height <= 30 || label.Bounds.Width > 101 {
		c.t.Fatalf("vertical label did not wrap: %+v", label.Bounds)
	}
	c.capture("separator-content-and-wrap")
	c.navigate("button", "Button")
}

func (c *client) checkKbdSymbolsAndTheme() {
	c.navigate("kbd", "Kbd")
	c.find("⊞")
	c.find("⌘")
	c.press("Show macOS key labels")
	c.find("⇧")
	c.press("Custom inverted background")
	alert := c.find("Neutral alert")
	c.call("act", map[string]any{"id": alert.ID, "action": "scrollIntoView"}, nil)
	c.capture("kbd-symbols-neutral-theme")
	c.navigate("button", "Button")
}

func (c *client) checkSkeletonShape() {
	c.navigate("skeleton", "Skeleton")
	c.press("Square skeleton corners")
	sample := c.find("Skeleton shape sample")
	if sample.Role != "status" || slices.Contains(sample.Actions, "press") || sample.Bounds.Width != 160 || sample.Bounds.Height != 48 {
		c.t.Fatalf("skeleton role or size: %+v", sample)
	}
	c.call("act", map[string]any{"id": sample.ID, "action": "scrollIntoView"}, nil)
	c.capture("skeleton-square-style")
	c.press("Square skeleton corners")
	c.capture("skeleton-rounded-ui")
	c.navigate("button", "Button")
}

func (c *client) checkProgressParity() {
	c.navigate("progress", "Progress")
	c.press("Advance progress")
	known := c.find("Preparing files")
	if known.Role != "progress" || known.Now == nil || *known.Now != 59 || known.Max == nil || *known.Max != 100 {
		c.t.Fatalf("progress range did not update: %+v", known)
	}
	c.press("Custom progress background")
	c.press("Invert progress")
	known = c.find("Neutral progress")
	if known.Value != "50 / 100 files" {
		c.t.Fatalf("custom progress value text: %+v", known)
	}
	unknown := c.find("Unknown neutral progress")
	if unknown.Now == nil || unknown.Min == nil || *unknown.Now >= *unknown.Min || unknown.Value != "In progress" {
		c.t.Fatalf("unknown progress semantics: %+v", unknown)
	}
	c.call("act", map[string]any{"id": unknown.ID, "action": "scrollIntoView"}, nil)
	c.capture("progress-neutral-inverted-unknown")
	c.navigate("button", "Button")
}

func (c *client) checkProgressGroupItems() {
	c.navigate("progress-group", "ProgressGroup")
	c.press("Custom progress group background")
	c.edit("Segment draft", "Original segment draft")
	c.key("Ctrl+A")
	editor := c.find("Segment draft")
	for _, label := range []string{"Reverse progress segments", "Toggle progress group direction"} {
		// press deliberately focuses its target; use pointer input for KeepFocus.
		toggle := c.find(label)
		c.call("act", map[string]any{"id": toggle.ID, "action": "scrollIntoView"}, nil)
		toggle = c.find(label)
		x, y := toggle.Bounds.X+toggle.Bounds.Width/2, toggle.Bounds.Y+toggle.Bounds.Height/2
		c.call("input", map[string]any{"events": []any{map[string]any{"type": "pointerDown", "x": x, "y": y}, map[string]any{"type": "pointerUp", "x": x, "y": y}}}, nil)
		if state := c.tree(); state.Focus != editor.ID {
			c.capture("progress-group-focus-failure")
			c.t.Fatalf("%s discarded custom legend focus: want=%s got=%s", label, editor.ID, state.Focus)
		}
	}
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "Updated segment draft"}}}, nil)
	if c.find("Segment draft").Value != "Updated segment draft" {
		c.t.Fatal("custom legend lost its selection")
	}
	group := c.find("Styled progress group")
	if group.Now == nil || math.Abs(*group.Now-60) > 1e-9 || group.Bounds.Height != 160 {
		c.t.Fatalf("group aggregate or height: %+v", group)
	}
	c.call("act", map[string]any{"id": group.ID, "action": "scrollIntoView"}, nil)
	c.capture("progress-group-icons-editor")
	c.press("Show progress group icons")
	c.capture("progress-group-dots-editor")
	c.navigate("button", "Button")
}

func (c *client) checkEmptyComposition() {
	c.navigate("empty", "Empty")
	if heading := c.find("No results"); heading.Role != "heading" {
		c.t.Fatalf("empty title role: %s", heading.Role)
	}
	avatar := c.find("Empty example avatar")
	// Native dimensions include system text scaling; exact default sizes are
	// covered by the headless contract test at a controlled scale.
	if avatar.Bounds.Width <= 0 || math.Abs(avatar.Bounds.Width-avatar.Bounds.Height) > .01 {
		c.t.Fatalf("empty avatar size: %+v", avatar.Bounds)
	}
	c.edit("Empty action draft", "Original empty draft")
	c.key("Ctrl+A")
	editor := c.find("Empty action draft")
	toggle := c.find("Toggle empty title")
	c.call("act", map[string]any{"id": toggle.ID, "action": "scrollIntoView"}, nil)
	toggle = c.find("Toggle empty title")
	x, y := toggle.Bounds.X+toggle.Bounds.Width/2, toggle.Bounds.Y+toggle.Bounds.Height/2
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "pointerDown", "x": x, "y": y}, map[string]any{"type": "pointerUp", "x": x, "y": y}}}, nil)
	if c.tree().Focus != editor.ID {
		c.t.Fatal("empty header change discarded action focus")
	}
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "Updated empty draft"}}}, nil)
	if c.find("Empty action draft").Value != "Updated empty draft" {
		c.t.Fatal("empty header change discarded selection")
	}
	c.press("Create from empty action")
	c.find("Empty actions: 1")
	c.press("Load empty content")
	if loadingAvatar := c.find("Empty example avatar"); math.Abs(loadingAvatar.Bounds.Width-avatar.Bounds.Width) > .01 {
		c.t.Fatal("loading changed the image avatar's size")
	}
	c.capture("empty-avatar-loading-actions")
	c.press("Custom empty header")
	c.press("Custom empty body")
	c.find("Custom empty header content")
	c.press("Create from custom body")
	c.find("Empty actions: 2")
	c.capture("empty-custom-header-body")
	c.navigate("button", "Button")
}

func (c *client) checkUserContent() {
	c.navigate("user", "User")
	c.press("Open profile")
	c.key("Enter")
	c.key("Space")
	c.find("Profile opened 3 times")
	c.capture("user-active-profile")
	disabled := c.find("Disabled profile")
	c.call("act", map[string]any{"id": disabled.ID, "action": "scrollIntoView"}, nil)
	disabled = c.find("Disabled profile")
	x, y := disabled.Bounds.X+disabled.Bounds.Width/2, disabled.Bounds.Y+disabled.Bounds.Height/2
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "pointerDown", "x": x, "y": y}, map[string]any{"type": "pointerUp", "x": x, "y": y}}}, nil)
	c.find("Profile opened 3 times")
	c.edit("User draft", "Original user draft")
	c.key("Ctrl+A")
	editor := c.find("User draft")
	for _, label := range []string{"Toggle user avatar", "Toggle user direction"} {
		toggle := c.find(label)
		c.call("act", map[string]any{"id": toggle.ID, "action": "scrollIntoView"}, nil)
		toggle = c.find(label)
		x, y := toggle.Bounds.X+toggle.Bounds.Width/2, toggle.Bounds.Y+toggle.Bounds.Height/2
		c.call("input", map[string]any{"events": []any{map[string]any{"type": "pointerDown", "x": x, "y": y}, map[string]any{"type": "pointerUp", "x": x, "y": y}}}, nil)
		if c.tree().Focus != editor.ID {
			c.t.Fatalf("%s discarded user body focus", label)
		}
	}
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "Updated user draft"}}}, nil)
	if c.find("User draft").Value != "Updated user draft" {
		c.t.Fatal("user body lost its selection")
	}
	c.find("5")
	c.find("User state avatar")
	c.find("Description without a name")
	c.capture("user-chip-content-state")
	c.navigate("button", "Button")
}

func (c *client) checkInputAccessories() {
	c.navigate("input", "Input")
	c.edit("Account draft", "Original account draft")
	c.key("Ctrl+A")
	editor := c.find("Account draft")
	for _, label := range []string{"Toggle input avatar", "Toggle input loading"} {
		toggle := c.find(label)
		c.call("act", map[string]any{"id": toggle.ID, "action": "scrollIntoView"}, nil)
		toggle = c.find(label)
		x, y := toggle.Bounds.X+toggle.Bounds.Width/2, toggle.Bounds.Y+toggle.Bounds.Height/2
		c.call("input", map[string]any{"events": []any{map[string]any{"type": "pointerDown", "x": x, "y": y}, map[string]any{"type": "pointerUp", "x": x, "y": y}}}, nil)
		if c.tree().Focus != editor.ID {
			c.t.Fatalf("%s discarded input focus", label)
		}
	}
	c.find("Account input avatar")
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "Edited while loading"}}}, nil)
	if c.find("Account draft").Value != "Edited while loading" {
		c.t.Fatal("input accessory changes lost selection or loading blocked editing")
	}
	c.key("Enter")
	// edit first clears the existing draft, then inserts the replacement.
	n := c.find("Account: Edited while loading · Changes: 3 · Submits: 1 · Clears: 0")
	c.call("act", map[string]any{"id": n.ID, "action": "scrollIntoView"}, nil)
	c.capture("input-avatar-loading")
	c.press("Clear account draft")
	c.find("Account:  · Changes: 4 · Submits: 1 · Clears: 1")
	if c.tree().Focus != editor.ID {
		c.t.Fatal("clear did not restore input focus")
	}
	c.navigate("button", "Button")
}

func (c *client) checkTextareaAccessories() {
	c.navigate("textarea", "Textarea")
	c.edit("Accessory notes", "First\nSecond\nThird\nFourth")
	c.find("Notes changes: 2")
	c.key("Ctrl+A")
	editor := c.find("Accessory notes")
	for _, label := range []string{"Toggle textarea avatar", "Toggle textarea growth"} {
		toggle := c.find(label)
		c.call("act", map[string]any{"id": toggle.ID, "action": "scrollIntoView"}, nil)
		toggle = c.find(label)
		x, y := toggle.Bounds.X+toggle.Bounds.Width/2, toggle.Bounds.Y+toggle.Bounds.Height/2
		c.call("input", map[string]any{"events": []any{map[string]any{"type": "pointerDown", "x": x, "y": y}, map[string]any{"type": "pointerUp", "x": x, "y": y}}}, nil)
		if c.tree().Focus != editor.ID {
			c.t.Fatalf("%s discarded textarea focus", label)
		}
	}
	if c.find("Accessory notes").Bounds.Height <= editor.Bounds.Height {
		c.t.Fatal("textarea did not grow after enabling autoresize")
	}
	c.find("Textarea account avatar")
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "Replacement notes"}}}, nil)
	c.find("Notes changes: 3") // Flush this edit before injecting the next input event.
	c.key("Enter")
	c.find("Notes changes: 4")
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "第二行"}}}, nil)
	if c.find("Accessory notes").Value != "Replacement notes\n第二行" {
		c.t.Fatal("textarea changed selection or Enter behavior")
	}
	n := c.find("Notes changes: 5")
	c.call("act", map[string]any{"id": n.ID, "action": "scrollIntoView"}, nil)
	c.capture("textarea-avatar-loading")
	c.press("Clear accessory notes")
	c.find("Notes changes: 6")
	if c.tree().Focus != editor.ID {
		c.t.Fatal("textarea clear did not restore focus")
	}
	c.navigate("button", "Button")
}

func (c *client) checkFormFieldLayout() {
	c.navigate("form-field", "FormField")
	c.edit("Field draft", "Original field draft")
	c.key("Ctrl+A")
	editor := c.find("Field draft")
	for _, label := range []string{"Toggle field label", "Toggle field direction", "Toggle field error"} {
		toggle := c.find(label)
		c.call("act", map[string]any{"id": toggle.ID, "action": "scrollIntoView"}, nil)
		toggle = c.find(label)
		x, y := toggle.Bounds.X+toggle.Bounds.Width/2, toggle.Bounds.Y+toggle.Bounds.Height/2
		c.call("input", map[string]any{"events": []any{map[string]any{"type": "pointerDown", "x": x, "y": y}, map[string]any{"type": "pointerUp", "x": x, "y": y}}}, nil)
		if c.tree().Focus != editor.ID {
			c.t.Fatalf("%s discarded field editor focus", label)
		}
	}
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "Updated field draft"}}}, nil)
	if c.find("Field draft").Value != "Updated field draft" {
		c.t.Fatal("field layout changes discarded selection")
	}
	if !strings.Contains(c.find("Field draft").Description, "State field error") || c.has("State field help") {
		c.t.Fatal("field error was not associated or did not replace help")
	}
	label := c.find("State field")
	c.call("act", map[string]any{"id": label.ID, "action": "scrollIntoView"}, nil)
	label = c.find("State field")
	x, y := label.Bounds.X+label.Bounds.Width/2, label.Bounds.Y+label.Bounds.Height/2
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "pointerDown", "x": x, "y": y}, map[string]any{"type": "pointerUp", "x": x, "y": y}}}, nil)
	if c.tree().Focus != editor.ID {
		c.t.Fatal("field label did not focus the editor")
	}
	status := c.find("Field value: Updated field draft")
	c.call("act", map[string]any{"id": status.ID, "action": "scrollIntoView"}, nil)
	c.capture("form-field-regions-horizontal")
	c.navigate("button", "Button")
}

func (c *client) checkFieldGroupState() {
	c.navigate("field-group", "FieldGroup")
	c.edit("Group draft", "Original group draft")
	c.key("Ctrl+A")
	editor := c.find("Group draft")
	for _, label := range []string{"Reorder field group", "Rotate field group"} {
		toggle := c.find(label)
		c.call("act", map[string]any{"id": toggle.ID, "action": "scrollIntoView"}, nil)
		toggle = c.find(label)
		x, y := toggle.Bounds.X+toggle.Bounds.Width/2, toggle.Bounds.Y+toggle.Bounds.Height/2
		c.call("input", map[string]any{"events": []any{map[string]any{"type": "pointerDown", "x": x, "y": y}, map[string]any{"type": "pointerUp", "x": x, "y": y}}}, nil)
		if c.tree().Focus != editor.ID {
			c.t.Fatalf("%s discarded group editor focus", label)
		}
	}
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "Updated group draft"}}}, nil)
	if c.find("Group draft").Value != "Updated group draft" {
		c.t.Fatal("group changes discarded the selection")
	}
	c.press("Check group draft")
	status := c.find("Checked group: Updated group draft")
	c.call("act", map[string]any{"id": status.ID, "action": "scrollIntoView"}, nil)
	c.capture("field-group-reordered-vertical")
	c.navigate("button", "Button")
}

func (c *client) checkSliderParity() {
	c.navigate("slider", "Slider")
	for _, label := range []string{"Decimal slider", "Decimal range minimum"} {
		n := c.find(label)
		c.call("act", map[string]any{"id": n.ID, "action": "focus"}, nil)
		c.key("Right")
		if value := c.find(label).Now; value == nil || *value != .3 {
			c.t.Fatalf("%s decimal keyboard value=%v", label, value)
		}
	}
	c.find("Decimal: 0.3 · Changes: 1 · Commits: 1")
	status := c.find("Decimal range: 0.3 – 0.8 · Changes: 1 · Commits: 1")
	c.call("act", map[string]any{"id": status.ID, "action": "scrollIntoView"}, nil)
	c.capture("slider-neutral-decimal")
	c.press("Mark sliders invalid")
	for _, label := range []string{"Decimal slider", "Decimal range minimum"} {
		if !strings.Contains(c.find(label).Description, "Choose another slider value") {
			c.t.Fatalf("%s missing its field error", label)
		}
	}
	status = c.find("Decimal range: 0.3 – 0.8 · Changes: 1 · Commits: 1")
	c.call("act", map[string]any{"id": status.ID, "action": "scrollIntoView"}, nil)
	c.capture("slider-field-error")
	c.press("Mark sliders invalid")
	c.press("Style slider parts")
	c.press("Pad slider track")
	n := c.find("Decimal slider")
	c.call("act", map[string]any{"id": n.ID, "action": "focus"}, nil)
	status = c.find("Decimal range: 0.3 – 0.8 · Changes: 1 · Commits: 1")
	c.call("act", map[string]any{"id": status.ID, "action": "scrollIntoView"}, nil)
	c.capture("slider-part-colors-padded-tooltip")
	c.press("Disable slider parent")
	n = c.find("Decimal slider")
	x, y := n.Bounds.X+n.Bounds.Width/2, n.Bounds.Y+n.Bounds.Height/2
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "pointerDown", "x": x, "y": y}, map[string]any{"type": "pointerUp", "x": x, "y": y}}}, nil)
	c.find("Decimal: 0.3 · Changes: 1 · Commits: 1")
	c.navigate("button", "Button")
}

func (c *client) checkSelectParity() {
	c.navigate("select", "Select")
	n := c.find("Accessory team")
	c.call("act", map[string]any{"id": n.ID, "action": "scrollIntoView"}, nil)
	c.call("act", map[string]any{"id": n.ID, "action": "focus"}, nil)
	c.idle()
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "b"}}}, nil)
	c.find("Accessory: beta · Changes: 1 · Open: false")
	c.press("Load select choices")
	status := c.find("Accessories: [alpha] · Changes: 0 · Open: false")
	c.call("act", map[string]any{"id": status.ID, "action": "scrollIntoView"}, nil)
	c.press("Accessory team")
	c.find("Accessory: beta · Changes: 1 · Open: true")
	c.capture("select-avatar-trailing-loading")
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "a"}}}, nil)
	c.key("Enter")
	c.find("Accessory: alpha · Changes: 2 · Open: false")
	if c.tree().Focus != c.find("Accessory team").ID {
		c.t.Fatal("single select did not restore trigger focus")
	}
	c.press("Customize select icons")
	status = c.find("Accessories: [alpha] · Changes: 0 · Open: false")
	c.call("act", map[string]any{"id": status.ID, "action": "scrollIntoView"}, nil)
	c.press("Accessory teams")
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "b"}}}, nil)
	c.key("Enter")
	c.find("Accessories: [alpha beta] · Changes: 1 · Open: true")
	c.capture("select-custom-icons-multiple")
	c.key("Escape")
	if c.tree().Focus != c.find("Accessory teams").ID {
		c.t.Fatal("multiple select did not restore trigger focus")
	}
	c.press("Toggle single select")
	c.find("Accessory: alpha · Changes: 2 · Open: true")
	c.key("Escape")
	c.find("Accessory: alpha · Changes: 2 · Open: false")
	c.press("Disable select parent")
	c.press("Accessory teams")
	c.find("Accessories: [alpha beta] · Changes: 1 · Open: false")
	c.navigate("button", "Button")
}

func (c *client) checkInputMenuParity() {
	c.navigate("input-menu", "InputMenu")
	c.find("Policy query:  · ID:  · IDs: [alpha beta]")
	c.press("Clear resets selected IDs")
	c.press("Clear Policy search")
	c.find("Policy query:  · ID:  · IDs: [alpha beta]")
	c.find("Policy events: query 0 · select 0 · values 0 · clear 1 · submit 0")
	c.press("Clear resets selected IDs")
	c.press("Clear Policy search")
	c.find("Policy query:  · ID:  · IDs: []")
	c.find("Policy events: query 0 · select 0 · values 1 · clear 2 · submit 0")
	c.press("InputMenu loading")
	c.edit("Policy search", "Be")
	c.key("Enter")
	c.find("Policy query:  · ID: beta · IDs: [beta]")
	c.find("Policy events: query 1 · select 1 · values 2 · clear 2 · submit 0")
	status := c.find("Policy events: query 1 · select 1 · values 2 · clear 2 · submit 0")
	c.call("act", map[string]any{"id": status.ID, "action": "scrollIntoView"}, nil)
	c.capture("input-menu-empty-draft-clear-loading")
	remove := c.find("Remove Beta")
	c.call("act", map[string]any{"id": remove.ID, "action": "focus"}, nil)
	c.key("Space")
	if c.tree().Focus != c.find("Policy search").ID {
		c.t.Fatal("tag removal lost input focus")
	}
	c.find("Policy events: query 1 · select 1 · values 3 · clear 2 · submit 0")
	c.press("InputMenu autocomplete")
	c.edit("Policy search", "Al")
	c.key("Enter")
	c.find("Policy query: Alpha · ID: alpha · IDs: []")
	c.find("Policy events: query 2 · select 2 · values 3 · clear 2 · submit 0")
	status = c.find("Policy events: query 2 · select 2 · values 3 · clear 2 · submit 0")
	c.call("act", map[string]any{"id": status.ID, "action": "scrollIntoView"}, nil)
	c.capture("input-menu-autocomplete-suggestion")
	c.press("InputMenu read only")
	c.press("Toggle policy search")
	if c.has("Clear Policy search") {
		c.t.Fatal("read-only input showed clear")
	}
	c.find("Policy query: Alpha · ID: alpha · IDs: []")
	c.navigate("button", "Button")
}

func (c *client) checkPopoverParity() {
	c.navigate("popover", "Popover")
	c.press("Style controlled popup")
	c.press("Open controlled popup")
	c.edit("Controlled popup draft", "Selected draft")
	c.key("Ctrl+A")
	n := c.find("Move popup anchor")
	x, y := n.Bounds.X+n.Bounds.Width/2, n.Bounds.Y+n.Bounds.Height/2
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "pointerDown", "x": x, "y": y}, map[string]any{"type": "pointerUp", "x": x, "y": y}}}, nil)
	if c.tree().Focus != c.find("Controlled popup draft").ID {
		c.t.Fatal("reanchoring lost popup editor focus")
	}
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "Moved draft"}}}, nil)
	if c.find("Controlled popup draft").Value != "Moved draft" {
		c.t.Fatal("reanchoring lost popup selection")
	}
	c.capture("popover-reanchored-styled")
	c.press("Close controlled popup")
	if c.tree().Focus != c.find("Alternate popup anchor").ID {
		c.t.Fatal("close did not restore the current popup anchor")
	}
	c.press("Alternate popup anchor")
	c.key("Escape")
	if c.tree().Focus != c.find("Alternate popup anchor").ID {
		c.t.Fatal("Escape did not restore the current anchor")
	}
	c.find("Controlled popup value: Moved draft")
	c.navigate("button", "Button")
}

func (c *client) checkScrollAreaParity() {
	c.navigate("scroll-area", "ScrollArea")
	offsets := func() (float64, float64) {
		for _, n := range c.tree().Nodes {
			if strings.HasPrefix(n.Label, "Scroll position:") {
				var x, y float64
				if _, err := fmt.Sscanf(n.Label, "Scroll position: x=%f y=%f", &x, &y); err != nil {
					c.t.Fatal(err)
				}
				return x, y
			}
		}
		c.t.Fatal("missing scroll position")
		return 0, 0
	}
	n := c.find("State scroll body")
	c.call("act", map[string]any{"id": n.ID, "action": "scrollIntoView"}, nil)
	n = c.find("State scroll body")
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "wheel", "x": n.Bounds.X + 70, "y": n.Bounds.Y + 50, "dx": 60, "dy": 70, "precise": true}}}, nil)
	x, y := offsets()
	if x <= 0 || y <= 0 {
		c.t.Fatalf("both-axis native wheel=(%v,%v)", x, y)
	}
	c.press("Wide scroll thumb")
	if a, b := offsets(); a != x || b != y {
		c.t.Fatal("thumb styling changed native offsets")
	}
	c.press("Scroll to end")
	a, b := offsets()
	if a < x || b < y {
		c.t.Fatal("end action did not reach both ends")
	}
	c.press("Scroll to start")
	c.find("Scroll position: x=0 y=0")
	n = c.find("State scroll body")
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "pointerMove", "x": n.Bounds.X + n.Bounds.Width - 8, "y": n.Bounds.Y + 20}}}, nil)
	c.capture("scroll-area-rounded-stable-state")
	c.navigate("button", "Button")
}

func (c *client) checkTooltipParity() {
	c.navigate("tooltip", "Tooltip")
	c.press("Instant tooltip")
	c.press("Narrow tooltip")
	c.press("Toggle timing tooltip")
	c.find("Timing tooltip open: true")
	anchor := c.find("Timing tooltip anchor")
	c.call("act", map[string]any{"id": anchor.ID, "action": "scrollIntoView"}, nil)
	c.capture("tooltip-narrow-shortcuts")
	c.key("Escape")
	c.find("Timing tooltip open: false")
	c.press("Shortcut keys only")
	n := c.find("Timing tooltip anchor")
	if !strings.Contains(n.Description, "Workspace action") || !strings.Contains(n.Description, "Ctrl K") || strings.Contains(n.Description, "Search workspace") {
		c.t.Fatalf("keys-only tooltip description=%q", n.Description)
	}
	c.call("act", map[string]any{"id": n.ID, "action": "scrollIntoView"}, nil)
	n = c.find("Timing tooltip anchor")
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "pointerMove", "x": n.Bounds.X + 8, "y": n.Bounds.Y + 8}}}, nil)
	c.find("Timing tooltip open: true")
	c.capture("tooltip-keys-only")
	c.press("Disable timing tooltip")
	c.find("Timing tooltip open: false")
	c.press("Disable timing tooltip")
	n = c.find("Timing tooltip anchor")
	c.call("act", map[string]any{"id": n.ID, "action": "focus"}, nil)
	c.find("Timing tooltip open: true")
	c.key("Escape")
	c.find("Timing tooltip open: false")
	c.navigate("button", "Button")
}

func (c *client) checkModalParity() {
	c.navigate("modal", "Modal")
	c.press("Block dialog dismissal")
	c.press("Transparent dialog backdrop")
	c.press("Disable dialog motion")
	c.press("Open audited dialog")
	c.key("Escape")
	c.find("Audited dialog draft")
	c.edit("Audited dialog draft", "Selected native draft")
	c.key("Ctrl+A")
	for _, label := range []string{"Toggle dialog header", "Toggle dialog fullscreen"} {
		n := c.find(label)
		x, y := n.Bounds.X+n.Bounds.Width/2, n.Bounds.Y+n.Bounds.Height/2
		c.call("input", map[string]any{"events": []any{map[string]any{"type": "pointerDown", "x": x, "y": y}, map[string]any{"type": "pointerUp", "x": x, "y": y}}}, nil)
		if c.tree().Focus != c.find("Audited dialog draft").ID {
			c.t.Fatal("dialog layout change lost editor focus")
		}
	}
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "Native dialog replacement"}}}, nil)
	if c.find("Audited dialog draft").Value != "Native dialog replacement" {
		c.t.Fatal("dialog layout change lost selection")
	}
	c.capture("modal-headerless-fullscreen-draft")
	c.press("Toggle dialog header")
	c.press("Toggle dialog fullscreen")
	c.capture("modal-custom-close-scroll")
	c.press("Close audited dialog")
	c.find("Dialog notifications: closed=1 prevented=1")
	c.find("Dialog draft: Native dialog replacement")
	if c.tree().Focus != c.find("Open audited dialog").ID {
		c.t.Fatal("dialog did not return focus to opener")
	}
	c.navigate("button", "Button")
}

func (c *client) checkSlideoverParity() {
	c.navigate("slideover", "Slideover")
	c.press("Disable panel motion")
	c.press("Block panel dismissal")
	c.press("Custom panel backdrop")
	c.press("Open reviewed panel")
	c.key("Escape")
	c.edit("Reviewed panel draft", "Selected panel draft")
	c.key("Ctrl+A")
	for _, label := range []string{"Move panel side", "Toggle panel inset"} {
		n := c.find(label)
		x, y := n.Bounds.X+n.Bounds.Width/2, n.Bounds.Y+n.Bounds.Height/2
		c.call("input", map[string]any{"events": []any{map[string]any{"type": "pointerDown", "x": x, "y": y}, map[string]any{"type": "pointerUp", "x": x, "y": y}}}, nil)
		if c.tree().Focus != c.find("Reviewed panel draft").ID {
			c.t.Fatal("slideover placement change lost focus")
		}
	}
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "Native panel replacement"}}}, nil)
	if c.find("Reviewed panel draft").Value != "Native panel replacement" {
		c.t.Fatal("slideover placement change lost selection")
	}
	c.capture("slideover-inset-left-draft")
	c.press("Close Reviewed panel")
	c.find("Panel notifications: closed=1 prevented=1")
	c.find("Panel draft: Native panel replacement")
	if c.tree().Focus != c.find("Open reviewed panel").ID {
		c.t.Fatal("slideover did not restore its opener")
	}
	c.navigate("button", "Button")
}

func (c *client) checkToastParity() {
	c.navigate("toast", "Toast")
	c.press("Toast avatar")
	c.press("Custom toast close")
	c.press("Toast position")
	c.press("Top right")
	c.press("Show policy toast")
	c.find("Ada Lovelace")
	c.find("Close policy toast")
	c.press("Keep policy toast")
	n := c.find("Update focused toast")
	c.call("act", map[string]any{"id": n.ID, "action": "scrollIntoView"}, nil)
	n = c.find("Update focused toast")
	x, y := n.Bounds.X+n.Bounds.Width/2, n.Bounds.Y+n.Bounds.Height/2
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "pointerDown", "x": x, "y": y}, map[string]any{"type": "pointerUp", "x": x, "y": y}}}, nil)
	if c.tree().Focus != c.find("Keep policy toast").ID {
		c.t.Fatal("toast update lost action identity or focus")
	}
	c.key("Enter")
	c.find("Policy toast actions: 2 · Closed: 0")
	c.capture("toast-avatar-horizontal-stable-action")
	c.press("Done policy toast")
	c.find("Policy toast actions: 3 · Closed: 1")
	c.press("Show policy toast")
	c.press("Close policy toast")
	c.find("Policy toast actions: 3 · Closed: 2")
	c.navigate("button", "Button")
}

func TestNativeGallery(t *testing.T) {
	socket := os.Getenv("NUI_AUTOMATION_SOCKET")
	if socket == "" {
		t.Skip("run task test:native, or python3 scripts/test-native.py")
	}
	var conn net.Conn
	var err error
	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); {
		conn, err = net.DialTimeout("unix", socket, 250*time.Millisecond)
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			t.Errorf("close automation connection: %v", err)
		}
	}()
	c := client{t: t, conn: conn, dec: json.NewDecoder(conn), enc: json.NewEncoder(conn)}
	var windows struct {
		Windows []struct {
			ID          int
			Kind, Title string
		}
	}
	c.call("windows", nil, &windows)
	if len(windows.Windows) != 1 || windows.Windows[0].Kind != "content" {
		t.Fatalf("windows: %+v", windows)
	}
	c.window = windows.Windows[0].ID
	first := c.window
	c.capture("components-light")
	if os.Getenv("NUI_NATIVE_HIDDEN") == "1" {
		t.Log("hidden window: first-frame smoke only; run without the hidden-smoke option for interaction")
		return
	}
	c.checkCheckboxIcons()
	c.checkSwitchLoading()
	c.checkChoiceGroupIcons()
	c.checkContainerPadding()
	c.checkButtonAvatar()
	c.checkBadgeAvatar()
	c.checkCardSpacing()
	c.checkAlertLayout()
	c.checkAvatarImage()
	c.checkAvatarGroupStyles()
	c.checkChipExamples()
	c.checkSeparatorContent()
	c.checkKbdSymbolsAndTheme()
	c.checkSkeletonShape()
	c.checkProgressParity()
	c.checkProgressGroupItems()
	c.checkEmptyComposition()
	c.checkUserContent()
	c.checkInputAccessories()
	c.checkTextareaAccessories()
	c.checkFormFieldLayout()
	c.checkFieldGroupState()
	c.checkSliderParity()
	c.checkSelectParity()
	c.checkInputMenuParity()
	c.checkPopoverParity()
	c.checkScrollAreaParity()
	c.checkTooltipParity()
	c.checkModalParity()
	c.checkSlideoverParity()
	c.checkToastParity()
	c.checkDropdownMenuParity()
	c.checkContextMenuParity()
	c.checkTabsParity()
	c.checkDisclosureParity()
	c.checkNavigationParity()
	c.checkFormParity()
	c.checkAdvancedInputsParity()
	c.checkRatingParity()
	c.checkSelectionParity()
	c.checkCalendarParity()
	c.checkDateTimeInputParity()
	c.checkColorPickerParity()
	c.checkTableParity()
	c.checkTreeParity()
	c.checkNestedSplitters()
	c.checkMarqueeSplitterParity()
	c.checkLayoutParity()
	if os.Getenv("NUI_NATIVE_NARROW") == "1" {
		c.checkNarrowDirectory()
		c.checkListboxStates()
		c.checkGalleryStateExamples()
		return
	}
	c.checkGallerySidebar()
	c.checkThemePalettes()
	c.checkListboxStates()
	c.checkGalleryStateExamples()
	c.navigate("button", "Button")
	c.press("Save changes")
	c.find("Saved 1 times")
	c.idle()
	c.find("Saved 1 times")
	c.press("New window")
	c.call("windows", nil, &windows)
	if len(windows.Windows) != 2 {
		t.Fatal("second native window missing")
	}
	second := windows.Windows[1].ID
	if second == first {
		second = windows.Windows[0].ID
	}
	c.theme("Dark")
	darkPNG := c.capture("components-dark")
	c.window = second
	c.find("Saved 0 times")
	lightPNG := c.capture("second-window-light")
	dark, err := png.Decode(bytes.NewReader(darkPNG))
	if err != nil {
		t.Fatal(err)
	}
	light, err := png.Decode(bytes.NewReader(lightPNG))
	if err != nil {
		t.Fatal(err)
	}
	if dark.At(2, 2) == light.At(2, 2) {
		t.Fatal("window theme leaked")
	}
	c.window = first
	c.navigate("theme", "Theme")
	c.press("Local dialog")
	c.find("Local theme dialog")
	c.capture("local-dialog-light")
	c.window = second
	for _, n := range c.tree().Nodes {
		if n.Role == "dialog" {
			t.Fatal("dialog leaked across windows")
		}
	}
	c.navigate("select", "Select")
	if c.find("Plan").Value != "Personal" {
		t.Fatal("initial plan")
	}
	c.window = first
	c.press("Close local dialog")
	c.navigate("input", "Input")
	n := c.find("Workspace name")
	c.call("act", map[string]any{"id": n.ID, "action": "focus"}, nil)
	c.idle()
	c.key("Ctrl+A")
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "原生中文验证"}}}, nil)
	c.idle()
	if got := c.find("Workspace name").Value; got != "原生中文验证" {
		c.capture("input-failure")
		t.Fatalf("native Unicode text: got %q", got)
	}
	ro := c.find("Workspace ID")
	if !strings.Contains(strings.Join(ro.States, ","), "readOnly") {
		t.Fatal("readonly semantics")
	}
	c.navigate("select", "Select")
	c.press("Plan")
	c.key("End")
	c.key("Enter")
	if got := c.find("Plan").Value; got != "Team" {
		c.capture("select-failure")
		t.Fatalf("native select: got %q", got)
	}
	c.capture("inputs-dark")
	c.window = second
	if c.find("Plan").Value != "Personal" {
		t.Fatal("model leaked across windows")
	}
	c.navigate("input", "Input")
	if c.find("Workspace name").Value != "Nuxt workspace" {
		t.Fatal("input model leaked across windows")
	}
	// Keep a modal open while switching windows and confirm its draft survives.
	c.window = first
	c.navigate("modal", "Modal")
	c.press("Edit workspace")
	c.edit("Dialog workspace name", "原生中文验证")
	closeID := c.find("Close Edit workspace").ID
	c.window = second
	if c.has("Language / 语言") {
		t.Fatal("Gallery should not show a language switch")
	}
	c.window = first
	if c.find("Close Edit workspace").ID != closeID || c.find("Dialog workspace name").Value != "原生中文验证" {
		t.Fatal("switching windows reset a modal or its edited value")
	}
	c.press("Close Edit workspace")
	c.window = first
	c.press("Edit workspace")
	c.press("Dialog plan")
	c.capture("modal-select-dark")
	c.key("Escape")
	c.find("Save nested")
	c.key("Escape")
	tr := c.tree()
	// MyGo omits accessibility focus while a different OS window is active.
	if tr.Focus != "" && tr.Focus != c.find("Edit workspace").ID {
		t.Fatal("native focus points to the wrong trigger")
	}
	c.key("Enter")
	c.find("Save nested") // the restored opener must activate from the keyboard
	c.key("Escape")
	c.navigate("dropdown-menu", "DropdownMenu")
	c.press("Project actions nested")
	c.press("Export")
	c.press("CSV file")
	c.find("CSV export prepared")
	c.press("Rich actions")
	c.press("Export options")
	c.capture("rich-menu-dark")
	c.press("Export rich CSV")
	c.find("Rich CSV exported")
	c.navigate("table", "Table")
	c.capture("table-dark")
	c.navigate("example/workspace", "Workspace")
	c.press("Save workspace")
	c.find("Workspace saved 1 times")
	c.press("Add example document")
	c.press("Foundation component notes")
	c.find("Document opened")
	c.press("Remove document")
	c.capture("workspace-dark")
	c.navigate("avatar-group", "AvatarGroup")
	c.find("2 more people")
	c.navigate("chip", "Chip")
	c.find("Notifications 3")
	c.capture("display-dark")
	c.navigate("progress", "Progress")
	progress := c.find("Preparing files")
	// The early parity probe advanced this persistent scene from 42 to 59.
	if progress.Role != "progress" || progress.Now == nil || *progress.Now != 59 || progress.Max == nil || *progress.Max != 100 {
		t.Fatalf("progress semantics: %+v", progress)
	}
	t.Logf("Progress accessibility range: %g / %g", *progress.Now, *progress.Max)
	if unknown := c.find("Synchronizing"); unknown.Now == nil || unknown.Min == nil || *unknown.Now >= *unknown.Min || unknown.Value != "In progress" {
		t.Fatal("indeterminate progress lost MyGo's unknown-value sentinel")
	}
	c.press("Advance progress")
	c.find("Preparing files · 76%")
	if got := c.find("Preparing files"); got.Now == nil || *got.Now != 76 {
		t.Fatal("progress range did not follow value")
	}
	c.navigate("alert", "Alert")
	c.press("Dismiss Your changes are saved")
	for _, n := range c.tree().Nodes {
		if n.Role == "status" && n.Label == "Your changes are saved" {
			t.Fatal("alert did not close")
		}
	}
	c.capture("feedback-dark")
	c.navigate("theme", "Theme")
	c.press("Choose primary color")
	c.press("Blue")
	c.press("Toggle density")
	c.capture("theme-blue-compact-dark")
	c.window = second
	c.navigate("example/workspace", "Workspace")
	c.press("Save workspace")
	c.find("Workspace saved 1 times")
	c.capture("workspace-second-light")
	c.navigate("example/settings", "Settings")
	c.press("Clear display name")
	c.press("Save changes")
	name := c.find("Display name *")
	if !strings.Contains(strings.Join(name.States, ","), "invalid") || !strings.Contains(name.Description, "Enter your display name") || !strings.Contains(name.Description, "Required") {
		t.Fatalf("field semantics: %+v", name)
	}
	c.call("act", map[string]any{"id": name.ID, "action": "focus"}, nil)
	c.idle()
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "原生表单验证"}}}, nil)
	c.idle()
	c.press("Save changes")
	c.find("Changes saved · 1")
	c.press("Password")
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "sample-secret-123"}}}, nil)
	c.idle()
	if n := c.find("Password"); n.Value != "" || !strings.Contains(strings.Join(n.States, ","), "password") {
		t.Fatalf("password accessibility should expose only its state: %+v", n)
	}
	c.capture("profile-light")
	c.press("Settings: Notifications")
	c.press("Enable notifications")
	if n := c.find("Desktop"); !strings.Contains(strings.Join(n.States, ","), "disabled") {
		t.Fatal("dependent checkbox was enabled")
	}
	c.press("Enable notifications")
	c.press("Desktop")
	if n := c.find("Desktop"); !strings.Contains(strings.Join(n.States, ","), "checked") {
		t.Fatal("checkbox not checked")
	}
	volume := c.find("Notification volume")
	if volume.Role != "slider" || volume.Now == nil || *volume.Now != 40 {
		t.Fatalf("slider semantics: %+v", volume)
	}
	c.call("act", map[string]any{"id": volume.ID, "action": "focus"}, nil)
	c.idle()
	c.key("Right")
	if n := c.find("Notification volume"); n.Now == nil || *n.Now != 45 {
		t.Fatal("native slider step")
	}
	c.capture("notifications-light")
	c.press("Settings: Workspace")
	c.press("Region")
	c.key("End")
	c.key("Enter")
	if got := c.find("Region").Value; got != "Asia Pacific" {
		t.Fatalf("region %q", got)
	}
	owner := c.find("Workspace owner")
	c.call("act", map[string]any{"id": owner.ID, "action": "focus"}, nil)
	c.idle()
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "Taipei"}}}, nil)
	c.idle()
	c.key("Enter")
	if got := c.find("Workspace owner").Value; got != "Alex Chen" {
		t.Fatalf("owner %q", got)
	}
	c.key("Ctrl+A")
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "No match"}}}, nil)
	c.idle()
	c.find("No people match your search")
	c.capture("empty-search-light")
	c.key("Escape")
	c.press("About workspace access")
	c.find("Got it")
	c.capture("popover-light")
	c.key("Escape")
	c.capture("workspace-light")
	c.window = first
	c.navigate("example/settings", "Settings")
	if got := c.find("Display name *").Value; got != "Lin Chen / 林晓" {
		t.Fatal("settings leaked across windows")
	}
	c.capture("profile-first-dark")
	// Roving focus requires the foreground native window; the second window is active.
	c.window = second
	for _, n := range c.tree().Nodes {
		if n.Role == "button" && strings.HasPrefix(n.Label, "Dismiss ") {
			c.call("act", map[string]any{"id": n.ID, "action": "press"}, nil)
			c.idle()
		}
	}
	c.theme("Dark")
	c.press("Actions")
	if got := c.find("Settings actions"); got.Role != "menu" {
		t.Fatalf("menu role: %+v", got)
	}
	c.key("Down")
	c.key("Right")
	c.find("Quiet mode")
	c.capture("settings-submenu-dark")
	c.key("Escape")
	c.key("Escape")
	c.press("Actions")
	c.key("Enter")
	if n := c.find("Edit profile details"); n.Role != "dialog" {
		t.Fatalf("dialog role: %+v", n)
	}
	c.capture("settings-edit-dark")
	c.press("Cancel edit")
	// A menu-opened modal returns to the persistent menu button, not a removed item.
	c.key("Enter")
	c.find("Settings actions")
	c.key("End")
	c.key("Enter")
	c.find("Reset notifications?")
	c.capture("settings-confirm-dark")
	c.press("Reset notifications")
	c.find("Undo reset")
	c.capture("settings-toast-dark")
	c.press("Undo reset")
	c.press("Notification filters")
	c.press("Sound")
	c.find("Filter volume")
	c.capture("settings-slideover-dark")
	c.press("Done filtering")
	c.theme("Light")
	c.window = second
	c.navigate("dropdown-menu", "DropdownMenu")
	c.press("Project actions checks")
	if n := c.find("Pin workspace"); n.Role != "menuItemCheckBox" {
		t.Fatalf("checkbox menu role: %+v", n)
	}
	c.key("Escape")
	c.press("Project actions nested")
	c.key("End")
	c.key("Right")
	c.capture("menu-light")
	c.key("Enter")
	c.find("CSV export prepared")
	c.navigate("context-menu", "ContextMenu")
	contextTarget := c.find("Project context basic")
	c.call("act", map[string]any{"id": contextTarget.ID, "action": "focus"}, nil)
	c.idle()
	c.key("Shift+F10")
	c.find("Project context menu basic")
	c.capture("context-light")
	c.key("Escape")
	c.navigate("modal", "Modal")
	c.press("Edit workspace")
	c.press("Dialog plan")
	c.capture("modal-select-light")
	c.key("Escape")
	c.find("Save nested")
	c.key("Escape")
	c.navigate("slideover", "Slideover")
	c.press("Open bottom panel")
	c.capture("bottom-light")
	c.press("Done with preferences")
	c.navigate("accordion", "Accordion")
	c.press("Sharing permissions")
	c.capture("navigation-light")
	c.navigate("pagination", "Pagination")
	c.press("Next page")
	c.find("Page 6 · Changes: 1")
	c.capture("pagination-light")
	c.navigate("timeline", "Timeline")
	c.capture("timeline-light")
	c.navigate("error", "Error")
	c.press("Restore activity")
	c.find("Activity restored")
	c.navigate("input-number", "InputNumber")
	c.edit("bounded seats", "3")
	c.press("Increase bounded seats")
	if n := c.find("bounded seats"); n.Value != "3.5" {
		t.Fatalf("native number: %+v", n)
	}
	seatsID := c.find("bounded seats").ID
	c.press("Number orientation")
	c.press("vertical")
	if n := c.find("bounded seats"); n.ID != seatsID || n.Value != "3.5" {
		t.Fatalf("orientation reset the native editor: %+v", n)
	}
	c.press("Increase bounded seats")
	if c.find("bounded seats").Value != "4" {
		t.Fatal("vertical number increase failed")
	}
	c.capture("input-number-vertical")
	c.press("Number orientation")
	c.press("horizontal")
	c.press("Decrease bounded seats")
	if n := c.find("bounded seats"); n.ID != seatsID || n.Value != "3.5" {
		t.Fatalf("horizontal number decrease or editor identity failed: %+v", n)
	}
	c.capture("input-number-horizontal")
	c.navigate("pin-input", "PinInput")
	c.press("numeric verification code")
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "123456"}}}, nil)
	c.idle()
	if n := c.find("numeric verification code"); n.Value != "123456" {
		t.Fatalf("native PIN: %+v", n)
	}
	c.capture("advanced-light")
	c.navigate("select-menu", "SelectMenu")
	c.press("single members menu")
	c.press("Search single members menu")
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "Ada"}}}, nil)
	c.idle()
	c.key("Down")
	c.key("Enter")
	c.capture("select-menu-light")
	c.key("Escape")
	c.navigate("input-date", "InputDate")
	c.press("Choose calendar release date")
	c.press("calendar release date calendar 2026-11-15")
	if n := c.find("calendar release date day"); n.Value != "15" {
		t.Fatalf("native date: %+v", n)
	}
	day := c.find("calendar release date day")
	c.call("act", map[string]any{"id": day.ID, "action": "focus"}, nil)
	c.key("Up")
	if n := c.find("calendar release date day"); n.Value != "16" {
		t.Fatalf("native date increment: %+v", n)
	}
	// Region settings can change the segment order without changing language.
	month := c.find("calendar release date month")
	if month.Bounds.X < day.Bounds.X {
		c.key("Left")
	} else {
		c.key("Right")
	}
	if tr := c.tree(); tr.Focus != c.find("calendar release date month").ID {
		t.Fatalf("native segment navigation: focus %s, month %s", tr.Focus, month.ID)
	}
	c.navigate("input-time", "InputTime")
	hour := c.find("24-hour reminder hour")
	c.call("act", map[string]any{"id": hour.ID, "action": "focus"}, nil)
	c.key("Right")
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "45"}}}, nil)
	c.idle()
	if n := c.find("24-hour reminder minute"); n.Value != "45" {
		t.Fatalf("native time typing: %+v", n)
	}
	c.navigate("input-date", "InputDate")
	start := c.find("range release date start month")
	c.call("act", map[string]any{"id": start.ID, "action": "focus"}, nil)
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "2026-11-10"}}}, nil)
	c.idle()
	// Cross the visual boundary from the final start segment to the first end segment.
	lastStart, firstEnd := start, c.find("range release date end month")
	for _, part := range []string{"day", "year"} {
		n := c.find("range release date start " + part)
		if n.Bounds.X > lastStart.Bounds.X {
			lastStart = n
		}
		n = c.find("range release date end " + part)
		if n.Bounds.X < firstEnd.Bounds.X {
			firstEnd = n
		}
	}
	c.call("act", map[string]any{"id": lastStart.ID, "action": "focus"}, nil)
	c.key("Right")
	if tr := c.tree(); tr.Focus != firstEnd.ID {
		t.Fatal("native range navigation")
	}
	c.navigate("color-picker", "ColorPicker")
	c.press("palette accent A")
	c.key("Ctrl+A")
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "128"}}}, nil)
	c.idle()
	c.capture("dates-light")
	c.theme("Dark")
	c.capture("dates-dark")
	c.navigate("example/task-browser", "Task browser")
	c.press("Task 00001")
	c.press("Edit task")
	c.capture("task-edit-dark")
	c.press("Cancel task edit")
	c.press("Filters")
	c.capture("task-filters-dark")
	c.press("Apply filters")
	c.navigate("example/settings", "Settings")
	c.press("Settings: Project setup")
	c.press("Continue")
	c.find("Enter a project name")
	c.capture("wizard-dark")

	c.navigate("carousel", "Carousel")
	c.checkCarouselParity()
	c.capture("interactions-dark")
	c.press("Next horizontal review guide")
	c.find("Slide 2 · 1 changes · paused false")
	c.navigate("command-palette", "CommandPalette")
	c.checkCommandPaletteParity()
	c.press("Open search commands")
	for range 40 {
		found := false
		for _, n := range c.tree().Nodes {
			if n.Label == "Open activity" {
				found = true
			}
		}
		if found {
			break
		}
		time.Sleep(10 * time.Millisecond)
		c.idle()
	}
	c.capture("commands-dark")
	c.press("Appearance")
	for range 40 {
		found := false
		for _, n := range c.tree().Nodes {
			if n.Label == "Use light theme" {
				found = true
			}
		}
		if found {
			break
		}
		time.Sleep(10 * time.Millisecond)
		c.idle()
	}
	c.capture("nested-commands-dark")
	c.press("Use light theme")
	c.navigate("drawer", "Drawer")
	c.checkDrawerParity()
	c.press("Open basic drawer")
	c.press("Resize basic inspector")
	c.key("End")
	c.capture("drawer-light")
	c.key("Escape")
	c.navigate("chart", "Chart")
	c.checkChartParity()
	c.press("Animate charts")
	c.press("Update data")
	c.press("Area")
	c.press("Chart stacking")
	c.press("Normalized")
	c.press("Builds")
	if slices.Contains(c.find("Builds").States, "checked") {
		t.Fatal("chart legend remained checked after hiding")
	}
	c.press("Update data")
	if slices.Contains(c.find("Builds").States, "checked") {
		t.Fatal("chart data update reset visibility")
	}
	c.capture("chart-stacked-hidden")
	// Category labels must stay aligned after data updates and legend changes.
	jan, mar, may, jul := c.find("Jan").Bounds, c.find("Mar").Bounds, c.find("May").Bounds, c.find("Jul").Bounds
	x := func(b struct{ X, Y, Width, Height float64 }) float64 { return b.X + b.Width/2 }
	if d1, d2 := x(may)-x(mar), x(jul)-x(may); d1 < .8*d2 || d1 > 1.2*d2 {
		t.Fatalf("category tick alignment changed: Jan=%+v Mar=%+v May=%+v Jul=%+v", jan, mar, may, jul)
	}
	c.press("Pie")
	c.press("Update data")
	c.press("Donut")
	c.press("Passed")
	if slices.Contains(c.find("Passed").States, "checked") {
		t.Fatal("pie legend remained checked after hiding")
	}
	c.capture("chart-donut-hidden")
	c.press("Reset data")
	c.press("Line")
	c.press("Chart X axis")
	c.press("Time")
	c.capture("chart-time-axis")
	c.press("Animate charts")
	c.navigate("file-upload", "FileUpload")
	c.checkFileUploadParity()
	c.press("Add native sample")
	c.find("review-1.md")
	c.capture("chart-files-light")
	c.navigate("navigation-menu", "NavigationMenu")
	c.checkNavigationMenuParity()
	c.press("Workspace settings")
	c.press("Members")
	c.find("Opened members")
	c.capture("shell-light")
	c.navigate("sidebar", "Sidebar")
	c.press("Toggle inner slideover sidebar")
	c.capture("shell-compact-light")
	c.checkSidebar()

	c.navigate("switch", "Switch")
	if !slices.Contains(c.find("Loading switch").States, "disabled") {
		t.Fatal("loading switch was not disabled")
	}
	c.capture("props-forms-light")
	c.navigate("pin-input", "PinInput")
	c.press("numeric verification code")
	c.capture("pin-empty-caret-light")
	c.key("Ctrl+A")
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "123456"}}}, nil)
	c.key("Home")
	c.key("Right")
	c.key("Right")
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "9"}}}, nil)
	c.idle()
	if n := c.find("numeric verification code"); n.Value != "129456" {
		t.Fatalf("native grouped PIN replacement: %+v", n)
	}
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "8"}}}, nil)
	c.idle()
	if n := c.find("numeric verification code"); n.Value != "129856" {
		t.Fatalf("native sequential PIN replacement shifted the suffix: %+v", n)
	}
	c.key("Ctrl+Z")
	if n := c.find("numeric verification code"); n.Value != "129456" {
		t.Fatalf("native PIN replacement undo: %+v", n)
	}
	c.capture("pin-grouped-focus-light")
	c.key("Tab")
	if tr := c.tree(); tr.Focus == c.find("numeric verification code").ID {
		t.Fatal("grouped PIN retained focus after Tab")
	}
	c.navigate("select-menu", "SelectMenu")
	c.press("single members menu")
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "Ada"}}}, nil)
	c.idle()
	c.press("Ada Lovelace")
	c.press("single members menu")
	c.find("Grace Hopper")
	c.key("Escape")
	c.navigate("input-menu", "InputMenu")
	c.press("Team search")
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "Beta"}}}, nil)
	c.idle()
	c.key("Enter")
	tag, draft := c.find("Remove Beta"), c.find("Team search")
	if tag.Bounds.X+tag.Bounds.Width > draft.Bounds.X {
		t.Fatalf("native InputMenu tag is not before its draft: tag %+v, draft %+v", tag.Bounds, draft.Bounds)
	}
	c.key("Escape")
	c.capture("input-menu-multiple-light")
	c.press("Remove Beta")
	c.key("Escape")
	c.capture("props-selection-light")
	c.navigate("slider", "Slider")
	c.press("Range orientation")
	c.press("vertical")
	minimum := c.find("Workspace range minimum")
	c.call("act", map[string]any{"id": minimum.ID, "action": "focus"}, nil)
	c.key("Up")
	if n := c.find("Workspace range minimum"); n.Now == nil || *n.Now != 25 {
		t.Fatalf("native vertical range: %+v", n)
	}
	c.capture("props-vertical-light")
	c.press("Reverse range")
	minimum = c.find("Workspace range minimum")
	c.call("act", map[string]any{"id": minimum.ID, "action": "focus"}, nil)
	c.key("Up")
	if n := c.find("Workspace range minimum"); n.Now == nil || *n.Now != 20 {
		t.Fatalf("native inverted range: %+v", n)
	}
	c.press("Reverse range")
	c.press("Range orientation")
	c.press("horizontal")
	c.capture("props-horizontal-light")

}

func (c *client) checkDropdownMenuParity() {
	c.navigate("dropdown-menu", "DropdownMenu")
	c.press("Filter menu actions")
	c.press("Open contract menu")
	if c.tree().Focus != c.find("Menu action search").ID {
		c.t.Fatal("native menu filter was not focused")
	}
	c.edit("Menu action search", "copy")
	c.key("Enter")
	c.find("Copied project")
	c.press("Open contract menu")
	c.key("Escape")
	c.press("Pin contract project")
	c.find("Menu actions: 2 · Checked: 1 · Open changes: 3 · Reorders: 0")
	c.press("Reverse menu items")
	c.key("Enter")
	if c.tree().Focus != c.find("Reverse menu items").ID {
		c.t.Fatal("native menu action lost identity during reorder")
	}
	c.find("Menu actions: 2 · Checked: 1 · Open changes: 3 · Reorders: 2")
	c.capture("dropdown-filter-avatar-checked")
	c.press("Contract export")
	c.edit("Format search", "json")
	c.capture("dropdown-submenu-filter")
	c.key("Enter")
	c.find("Contract JSON selected")
	if c.tree().Focus != c.find("Open contract menu").ID {
		c.t.Fatal("native menu selection did not restore trigger focus")
	}
	c.navigate("button", "Button")
}

func (c *client) checkContextMenuParity() {
	c.navigate("context-menu", "ContextMenu")
	c.press("Disable point menu")
	c.press("Point menu target")
	c.key("Shift+F10")
	if c.has("Point context panel") {
		c.t.Fatal("disabled context menu opened")
	}
	c.find("Target clicks: 1 · Menu changes: 0 · Moved: false")
	c.press("Disable point menu")
	target := c.find("Point menu target")
	rightClick := func(x, y float64) {
		c.call("input", map[string]any{"events": []any{map[string]any{"type": "click", "button": "right", "x": x, "y": y}}}, nil)
		c.idle()
	}
	rightClick(target.Bounds.X+target.Bounds.Width-4, target.Bounds.Y+4)
	first := c.find("Point context panel")
	rightClick(target.Bounds.X+4, target.Bounds.Y+target.Bounds.Height-4)
	second := c.find("Point context panel")
	if first.Bounds == second.Bounds {
		c.t.Fatal("native context menu did not reposition")
	}
	c.press("Move context target")
	after := c.find("Point context panel")
	if after.Bounds != second.Bounds {
		c.t.Fatal("native pointer menu moved with its target")
	}
	c.find("Target clicks: 1 · Menu changes: 1 · Moved: true")
	c.capture("context-fixed-point-moved-target")
	c.key("Escape")
	if c.tree().Focus != c.find("Point menu target").ID {
		c.t.Fatal("context Escape did not restore target")
	}
	c.press("Open context at target")
	target = c.find("Point menu target")
	opened := c.find("Point context panel")
	if math.Abs(opened.Bounds.X-target.Bounds.X-target.Bounds.Width-2) > 1 || math.Abs(opened.Bounds.Y-target.Bounds.Y) > 1 {
		c.t.Fatalf("native controlled context anchor: target=%v menu=%v", target.Bounds, opened.Bounds)
	}
	c.capture("context-controlled-target-anchor")
	c.key("Escape")
	c.press("Filter menu actions")
	contract := c.find("Open contract menu")
	c.call("act", map[string]any{"id": contract.ID, "action": "focus"}, nil)
	c.key("Shift+F10")
	mediumAvatar := c.find("Team avatar").Bounds.Width
	c.key("Escape")
	c.press("Menu size")
	c.press("XS")
	contract = c.find("Open contract menu")
	c.call("act", map[string]any{"id": contract.ID, "action": "focus"}, nil)
	c.key("Shift+F10")
	// Menu items are atomic accessibility nodes; the non-interactive heading
	// exposes its avatar separately. Compare sizes without overriding system text scaling.
	if width := c.find("Team avatar").Bounds.Width; math.Abs(width/mediumAvatar-.8) > .01 {
		c.t.Fatalf("native context avatar did not change from MD to XS: %v to %v", mediumAvatar, width)
	}
	c.edit("Menu action search", "copy")
	c.key("Enter")
	c.find("Copied project")
	c.key("Shift+F10")
	c.key("Escape")
	c.press("Pin contract project")
	c.find("Menu actions: 2 · Checked: 1 · Open changes: 5 · Reorders: 0")
	c.capture("context-xs-filter-checked")
	c.key("Escape")
	c.navigate("button", "Button")
}

func (c *client) checkTabsParity() {
	c.navigate("tabs", "Tabs")
	c.press("Manual tab activation")
	c.press("Account tab")
	c.key("Right")
	if c.tree().Focus != c.find("Activity tab").ID {
		c.t.Fatal("manual tabs did not navigate focus")
	}
	c.find("Selected tab: account · Changes: 0")
	c.key("Space")
	c.find("Selected tab: activity · Changes: 1")
	c.key("Down")
	c.find("Selected tab: activity · Changes: 1")
	c.press("Account tab")
	c.edit("Contract tab draft", "Selected draft")
	c.key("Ctrl+A")
	keepFocusClick := func(label string) {
		n := c.find(label)
		c.call("act", map[string]any{"id": n.ID, "action": "scrollIntoView"}, nil)
		n = c.find(label)
		x, y := n.Bounds.X+n.Bounds.Width/2, n.Bounds.Y+n.Bounds.Height/2
		c.call("input", map[string]any{"events": []any{map[string]any{"type": "pointerDown", "x": x, "y": y}, map[string]any{"type": "pointerUp", "x": x, "y": y}}}, nil)
		c.idle()
	}
	keepFocusClick("Reverse contract tabs")
	keepFocusClick("Toggle contract direction")
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "Edited after reorder"}}}, nil)
	c.idle()
	if c.tree().Focus != c.find("Contract tab draft").ID || c.find("Contract tab draft").Value != "Edited after reorder" {
		c.t.Fatal("tab reorder/direction lost editor selection")
	}
	c.press("Custom tab indicator")
	c.press("Account tab")
	c.capture("tabs-avatar-badge-vertical-indicator")
	c.press("Contract tab draft")
	keepFocusClick("Remove account tab")
	c.find("Selected tab: activity · Changes: 2")
	if c.tree().Focus != c.find("Activity tab").ID {
		c.t.Fatal("removed active tab left focus in hidden content")
	}
	c.press("Restore account tab")
	c.press("Account tab")
	c.press("Show tab content")
	if c.has("Contract tab draft") {
		c.t.Fatal("content=false retained native tab panels")
	}
	c.press("Show tab content")
	if c.find("Contract tab draft").Value != "Edited after reorder" {
		c.t.Fatal("tab content recreation lost application data")
	}
	c.capture("tabs-restored-content")
	c.navigate("button", "Button")
}

func (c *client) checkDisclosureParity() {
	for _, sample := range []struct{ id, title string }{{"collapsible", "Collapsible"}, {"accordion", "Accordion"}} {
		c.navigate(sample.id, sample.title)
		c.press("Disable contract heading")
		c.press("Contract body action")
		c.find("Disclosure open: true · Changes: 0 · Actions: 1")
		c.press("Disable contract heading")
		c.press("Reduce disclosure motion")
		c.edit("Contract disclosure draft", "Native retained draft")
		// A pointer click respects KeepFocus and lets external closing restore focus.
		n := c.find("Toggle contract disclosure externally")
		c.call("act", map[string]any{"id": n.ID, "action": "scrollIntoView"}, nil)
		n = c.find("Toggle contract disclosure externally")
		x, y := n.Bounds.X+n.Bounds.Width/2, n.Bounds.Y+n.Bounds.Height/2
		c.call("input", map[string]any{"events": []any{map[string]any{"type": "pointerDown", "x": x, "y": y}, map[string]any{"type": "pointerUp", "x": x, "y": y}}}, nil)
		c.idle()
		if c.has("Contract disclosure draft") || c.tree().Focus != c.find("Contract disclosure heading").ID {
			c.t.Fatalf("%s external close failed to hide content or restore heading focus", sample.id)
		}
		c.find("Disclosure open: false · Changes: 0 · Actions: 1")
		c.key("Space")
		if c.find("Contract disclosure draft").Value != "Native retained draft" {
			c.t.Fatalf("%s lost its draft", sample.id)
		}
		c.find("Disclosure open: true · Changes: 1 · Actions: 1")
		c.capture(sample.id + "-icons-mounted-content")
	}
	c.navigate("button", "Button")
}

func (c *client) checkDrawerParity() {
	c.navigate("drawer", "Drawer")
	c.theme("Light")
	c.press("Allow contract drawer dismissal")
	c.press("Style contract drawer indicator")
	c.press("Open contract drawer")
	c.capture("drawer-contract-opening")
	if c.has("Close Contract drawer") {
		c.t.Fatal("drawer close button should be hidden by default")
	}
	c.edit("Contract drawer draft", "native drawer draft")
	c.press("Resize Contract drawer")
	c.key("End")
	c.key("End")
	c.find("Snap: 2 · Changes: 1 · Closed: 0 · Prevented: 0")
	c.key("Home")
	c.key("Escape")
	drag := func(label string) {
		n := c.find(label)
		x, y := n.Bounds.X+n.Bounds.Width/2, n.Bounds.Y+n.Bounds.Height/2
		for _, event := range []map[string]any{{"type": "pointerDown", "x": x, "y": y}, {"type": "pointerMove", "x": x, "y": y + 200}, {"type": "pointerUp", "x": x, "y": y + 200}} {
			c.call("input", map[string]any{"events": []any{event}}, nil)
			c.idle()
		}
	}
	drag("Resize Contract drawer")
	c.find("Snap: 0 · Changes: 2 · Closed: 0 · Prevented: 2")
	button := c.find("Close drawer from app")
	c.call("act", map[string]any{"id": button.ID, "action": "scrollIntoView"}, nil)
	c.idle()
	button = c.find("Close drawer from app")
	panel := c.find("Contract drawer")
	if button.Bounds.Y < panel.Bounds.Y-.5 || button.Bounds.Y+button.Bounds.Height > panel.Bounds.Y+panel.Bounds.Height+.5 {
		c.capture("drawer-lowest-debug")
		c.t.Fatalf("lowest drawer snap clipped its closing action: button=%+v panel=%+v", button.Bounds, panel.Bounds)
	}
	c.capture("drawer-contract-lowest-footer")
	c.press("Resize Contract drawer")
	c.key("End")
	c.capture("drawer-contract-light")
	c.press("Toggle drawer inset")
	c.press("Toggle drawer title")
	c.find("Resize Drawer")
	if c.find("Contract drawer draft").Value != "native drawer draft" {
		c.t.Fatal("drawer layout change lost its draft")
	}
	c.press("Toggle drawer handle")
	if c.has("Resize Drawer") {
		c.t.Fatal("drawer handle remained after hiding it")
	}
	c.capture("drawer-contract-inset-hidden-handle")
	c.press("Toggle drawer handle")
	c.press("Close drawer from app")
	if c.tree().Focus != c.find("Open contract drawer").ID {
		c.t.Fatal("drawer did not restore explicit return focus")
	}
	c.press("Allow contract drawer dismissal")
	c.theme("Dark")
	c.press("Open contract drawer")
	c.capture("drawer-contract-dark-titleless")
	c.press("Resize Drawer")
	c.key("Home")
	drag("Resize Drawer")
	c.find("Snap: 0 · Changes: 4 · Closed: 1 · Prevented: 2")
	if c.tree().Focus != c.find("Open contract drawer").ID {
		c.t.Fatal("drag dismissal did not restore drawer return focus")
	}
}

func (c *client) checkNavigationMenuParity() {
	c.navigate("navigation-menu", "NavigationMenu")
	c.theme("Light")
	c.find("Workspace members")
	c.capture("navigation-contract-expanded")
	c.press("Protected destination")
	c.find("Navigation prevented")
	c.press("Workspace members")
	c.find("Destination: members · Actions: 1 · Expansion changes: 0")
	c.press("Compact contract navigation")
	c.press("Manage workspace")
	c.press("Protected destination")
	c.find("Workspace members")
	c.capture("navigation-contract-compact")
	c.key("Escape")
	if c.tree().Focus != c.find("Manage workspace").ID {
		c.t.Fatal("compact navigation did not restore the trigger")
	}
	c.press("Vertical contract navigation")
	if c.find("Overview page").Bounds.Width < 70 {
		c.t.Fatal("horizontal Compact hid the navigation label")
	}
	c.press("Manage workspace")
	c.press("Workspace members")
	c.find("Destination: members · Actions: 2 · Expansion changes: 0")
	c.capture("navigation-contract-horizontal")
	c.press("Compact contract navigation")
	c.press("Vertical contract navigation")
	c.press("Single navigation branch")
	c.press("Help center")
	if c.has("Workspace members") {
		c.t.Fatal("single navigation kept a second sibling open")
	}
	c.find("Quickstart guide")
	c.press("Allow navigation collapse")
	c.press("Help center")
	c.find("Quickstart guide")
	c.press("Disable contract navigation")
	c.press("Project library")
	c.find("Destination: members · Actions: 2 · Expansion changes: 1")
	c.press("Disable contract navigation")
	c.press("RTL contract navigation")
	c.press("Toggle navigation color")
	c.press("Toggle navigation variant")
	c.theme("Dark")
	c.capture("navigation-contract-rtl-dark")
}

func (c *client) checkCommandPaletteParity() {
	c.navigate("command-palette", "CommandPalette")
	c.theme("Light")
	waitFor := func(label string) {
		c.t.Helper()
		for deadline := time.Now().Add(2 * time.Second); time.Now().Before(deadline); {
			for _, n := range c.tree().Nodes {
				if n.Label == label {
					return
				}
			}
			time.Sleep(10 * time.Millisecond)
		}
		c.find(label)
	}
	c.press("Open contract commands")
	waitFor("Personal workspace")
	c.capture("commands-contract-light")
	c.press("Search Contract commands")
	c.key("Down")
	c.key("End")
	waitFor("Protected command")
	c.press("Protected command")
	c.find("Activation prevented · Requests: 0")
	c.key("Escape")
	if c.tree().Focus != c.find("Open contract commands").ID {
		c.t.Fatal("command palette did not restore its opener")
	}
	c.press("Keep commands open")
	c.press("Open contract commands")
	waitFor("Service status")
	c.press("Service status")
	c.press("Service status")
	c.find("Command: status · Calls: 2 · Edits: 0")
	c.press("Appearance options")
	waitFor("Light workspace")
	c.capture("commands-contract-nested")
	c.press("Search Contract commands")
	c.key("Backspace")
	waitFor("Appearance options")
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "invoice"}}}, nil)
	waitFor("File Create document Local")
	c.capture("commands-contract-description")
	c.key("Escape")
	c.press("Manual command results")
	c.press("Toggle command delay")
	c.press("RTL command palette")
	c.theme("Dark")
	c.press("Open contract commands")
	c.press("Search Contract commands")
	c.key("Ctrl+A")
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "no match"}}}, nil)
	time.Sleep(175 * time.Millisecond)
	c.idle()
	c.key("Ctrl+A")
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "Local"}}}, nil)
	time.Sleep(175 * time.Millisecond)
	c.idle()
	c.press("Try stale command result")
	c.find("Stale command result rejected · Requests: 3")
	c.press("Complete command search")
	waitFor("File Create document Local")
	c.capture("commands-contract-remote-rtl-dark")
	c.key("Escape")
}

func (c *client) checkTreeParity() {
	c.navigate("tree", "Tree")
	c.press("Custom tree parts")
	c.press("Contract Alpha")
	c.find("Tree selection: [alpha] · Changes: 1 · Expansions: 0")
	// This custom trailing text has no separate native accessibility node.
	// Compare the rendered mixed and unlinked states instead.
	c.capture("tree-mixed-parts")
	c.press("Select tree descendants")
	c.press("Bubble tree selection")
	c.capture("tree-unlinked-parts")
	c.press("Protected tree item")
	c.find("Selection prevented · Loads: 0")
	c.press("Move Alpha last")
	c.find("Tree selection: [alpha] · Changes: 1 · Expansions: 0")
	c.press("Prevent tree toggle")
	c.press("Contract project")
	c.key("Left")
	c.find("Toggle prevented · Loads: 0")
	c.find("Contract Alpha")
	c.press("Prevent tree toggle")
	c.call("act", map[string]any{"id": c.find("Contract tree").ID, "action": "focus"}, nil)
	c.key("Left")
	if c.has("Contract Alpha") {
		c.t.Fatal("tree branch did not collapse")
	}
	c.key("Right")
	c.find("Contract Alpha")
	c.press("RTL contract tree")
	c.call("act", map[string]any{"id": c.find("Contract tree").ID, "action": "focus"}, nil)
	c.key("Right")
	if c.has("Contract Alpha") {
		c.t.Fatal("RTL tree branch did not collapse")
	}
	c.key("Left")
	c.press("Remote records")
	c.key("Left")
	c.find("Loading remote records · Loads: 1")
	c.press("Collapse remote branch")
	c.press("Complete empty tree load")
	c.find("Stale tree completion rejected · Loads: 1")
	c.call("act", map[string]any{"id": c.find("Contract tree").ID, "action": "focus"}, nil)
	c.key("Left")
	c.find("Loading remote records · Loads: 2")
	c.press("Complete empty tree load")
	c.find("Remote directory is empty · Loads: 2")
	c.press("Custom tree parts")
	c.press("Disable contract tree")
	if !slices.Contains(c.find("Contract tree").States, "disabled") {
		c.t.Fatal("tree disabled state missing")
	}
	c.press("Disable contract tree")
	c.call("act", map[string]any{"id": c.find("Remote directory is empty · Loads: 2").ID, "action": "scrollIntoView"}, nil)
	c.capture("tree-rtl-loaded-empty")
	c.theme("Dark")
	c.call("act", map[string]any{"id": c.find("Contract tree").ID, "action": "focus"}, nil)
	c.key("Down")
	c.capture("tree-dark-keyboard")
	c.theme("Light")
}

func (c *client) checkTableParity() {
	c.navigate("table", "Table")
	c.press("Group")
	c.find("First: A2 · Sort events: 1 · Selection events: 0 · Activated: None")
	c.press("Sort group then ID")
	c.find("First: A1 · Sort events: 1 · Selection events: 0 · Activated: None")
	c.press("Restore input order")
	c.press("Manual table sorting")
	c.press("Group")
	c.find("First: B2 · Sort events: 2 · Selection events: 0 · Activated: None")
	cell := c.find("B2")
	c.call("act", map[string]any{"id": cell.ID, "action": "scrollIntoView"}, nil)
	cell = c.find("B2")
	x, y := cell.Bounds.X+cell.Bounds.Width/2, cell.Bounds.Y+cell.Bounds.Height/2
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "pointerDown", "x": x, "y": y}}}, nil)
	c.idle()
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "pointerUp", "x": x, "y": y}}}, nil)
	c.idle()
	c.key("Enter")
	c.find("First: B2 · Sort events: 2 · Selection events: 1 · Activated: B2")
	c.capture("table-contract")
	c.press("Disable contract table")
	if !slices.Contains(c.find("Contract table").States, "disabled") {
		c.t.Fatal("table did not inherit disabled state")
	}
	c.press("Disable contract table")
	c.press("Loading contract table")
	c.press("RTL contract table")
	c.press("Empty contract table")
	loading := c.find("Fetching contract records")
	c.call("act", map[string]any{"id": loading.ID, "action": "scrollIntoView"}, nil)
	loading = c.find("Fetching contract records")
	header := c.find("Group")
	if loading.Bounds.Y < header.Bounds.Y+header.Bounds.Height {
		c.t.Fatal("table loading content precedes its header")
	}
	c.capture("table-loading-rtl")
	c.press("Cycle loading animation")
	c.press("Cycle loading animation")
	c.press("Cycle loading animation")
	c.theme("Dark")
	c.call("act", map[string]any{"id": c.find("Fetching contract records").ID, "action": "scrollIntoView"}, nil)
	c.capture("table-loading-dark")
	c.theme("Light")
	c.press("Loading contract table")
	c.find("No contract records")
	c.capture("table-empty-contract")
}

func (c *client) checkColorPickerParity() {
	c.navigate("color-picker", "ColorPicker")
	area := c.find("Saturation and brightness")
	if math.Abs(area.Bounds.Width-area.Bounds.Height) > 1 {
		c.t.Fatalf("color area is not square: %+v", area.Bounds)
	}
	c.edit("Hex", "#225588AA")
	c.find("Color value: #225588AA · Size: MD")
	c.find("Color notified: #225588AA · Changes: 1")
	c.edit("Contract color R", "40")
	c.find("Color notified: #285588AA · Changes: 2")
	c.edit("Contract color R", "40")
	c.find("Color notified: #285588AA · Changes: 2")
	c.press("Cycle color size")
	c.find("Color value: #285588AA · Size: LG")
	larger := c.find("Saturation and brightness")
	if larger.Bounds.Width <= area.Bounds.Width || math.Abs(larger.Bounds.Width-larger.Bounds.Height) > 1 {
		c.t.Fatal("color size did not resize the square")
	}
	c.press("Style color parts")
	c.call("act", map[string]any{"id": larger.ID, "action": "scrollIntoView"}, nil)
	c.capture("color-picker-panel")
	c.call("act", map[string]any{"id": c.find("Contract color A").ID, "action": "scrollIntoView"}, nil)
	c.capture("color-picker-channels")
	c.press("Disable contract color")
	if !slices.Contains(c.find("Contract color R").States, "disabled") || !slices.Contains(c.find("Saturation and brightness").States, "disabled") {
		c.t.Fatal("color picker ignored inherited disabled")
	}
	c.key("Up")
	c.find("Color notified: #285588AA · Changes: 2")
	c.press("Disable contract color")
	c.press("Set translucent blue")
	c.find("Color value: #11223344 · Size: LG")
	c.find("Color notified: #285588AA · Changes: 2")
	c.press("Use default color throttle")
	c.call("act", map[string]any{"id": c.find("Contract color G").ID, "action": "focus"}, nil)
	c.key("Up")
	c.find("Color value: #11233344 · Size: LG")
	c.press("RTL color form")
	c.call("act", map[string]any{"id": c.find("Opacity").ID, "action": "scrollIntoView"}, nil)
	opacity := c.find("Opacity")
	x, y := opacity.Bounds.X+opacity.Bounds.Width*.75, opacity.Bounds.Y+opacity.Bounds.Height/2
	// The native slider reads its held pointer on a frame. Keep down/up in
	// separate frames, as the other native drag checks do.
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "pointerDown", "x": x, "y": y}}}, nil)
	c.idle()
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "pointerUp", "x": x, "y": y}}}, nil)
	c.idle()
	if current := c.find("Opacity"); current.Now == nil || *current.Now < .6 {
		c.capture("color-picker-rtl-failure")
		c.t.Fatalf("RTL color input: slider=%+v alpha=%+v", current, c.find("Contract color A"))
	}
	c.capture("color-picker-rtl")
	c.theme("Dark")
	c.call("act", map[string]any{"id": c.find("Saturation and brightness").ID, "action": "scrollIntoView"}, nil)
	c.capture("color-picker-dark")
	c.theme("Light")
}

func (c *client) checkDateTimeInputParity() {
	c.navigate("input-date", "InputDate")
	c.press("Choose Contract date")
	c.press("Contract date calendar 2024-02-29")
	c.find("Date input: 2024-02-29 · Changes: 1")
	c.call("act", map[string]any{"id": c.find("Contract date day").ID, "action": "focus"}, nil)
	c.key("Down")
	c.find("Date input: 2024-02-27 · Changes: 2")
	c.press("Clear Contract date range")
	c.find("Date input range:  to  · Changes: 1 · Clears: 1")
	c.press("Choose Contract date range")
	c.press("Contract date range calendar 2024-02-18")
	c.press("Contract date range calendar 2024-02-22")
	c.find("Date input range: 2024-02-18 to 2024-02-22 · Changes: 3 · Clears: 1")
	c.press("Custom date separator")
	c.press("Trailing date loader")
	c.press("Read-only date inputs")
	c.call("act", map[string]any{"id": c.find("Contract date day").ID, "action": "focus"}, nil)
	c.key("Up")
	c.find("Date input: 2024-02-27 · Changes: 2")
	c.capture("input-date-contract")
	c.theme("Dark")
	c.capture("input-date-contract-dark")
	c.theme("Light")
	c.navigate("input-time", "InputTime")
	c.call("act", map[string]any{"id": c.find("Contract time hour").ID, "action": "focus"}, nil)
	c.key("Up")
	c.find("Time input: 12:55:42 · Set: true · Changes: 1 · Clears: 0")
	c.call("act", map[string]any{"id": c.find("Contract time minute").ID, "action": "focus"}, nil)
	c.key("Up")
	c.find("Time input: 12:00:42 · Set: true · Changes: 2 · Clears: 0")
	c.press("Contract time hour only")
	c.find("Time input: 12:00:42 · Set: true · Changes: 2 · Clears: 0")
	c.call("act", map[string]any{"id": c.find("Contract time hour").ID, "action": "focus"}, nil)
	c.key("Down")
	c.find("Time input: 11:00:42 · Set: true · Changes: 3 · Clears: 0")
	c.capture("input-time-hour-precision")
	c.press("Contract time hour only")
	c.call("act", map[string]any{"id": c.find("Contract time minute").ID, "action": "focus"}, nil)
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "08"}}}, nil)
	c.idle()
	c.find("Time input: 11:10:42 · Set: true · Changes: 5 · Clears: 0")
	c.press("Trailing time loader")
	c.capture("input-time-contract")
	c.theme("Dark")
	c.capture("input-time-contract-dark")
	c.theme("Light")
	c.press("Clear Contract time")
	c.find("Time input: 00:00:00 · Set: false · Changes: 6 · Clears: 1")
}

func (c *client) checkCalendarParity() {
	c.navigate("calendar", "Calendar")
	c.press("Contract calendar 2024-02-15")
	c.find("Calendar date:  to  · Changes: 1")
	c.press("Contract calendar 2024-02-29")
	c.key("Right")
	c.key("Enter")
	c.find("Calendar date: 2024-03-01 to  · Changes: 3")
	c.capture("calendar-leap-day-navigation")
	c.press("Read-only contract calendar")
	c.press("Contract calendar 2024-03-15")
	c.key("Enter")
	c.find("Calendar date: 2024-03-01 to  · Changes: 3")
	c.press("Change contract calendar view")
	c.find("Contract calendar Mar 2024")
	c.capture("calendar-month-view")
	c.press("Change contract calendar view")
	c.find("Contract calendar 2025")
	c.capture("calendar-year-view")
	c.press("Contract calendar 2025")
	c.press("Contract calendar Feb 2025")
	c.find("Calendar date: 2024-03-01 to  · Changes: 3")
	c.press("Jump calendar to February")
	c.press("Read-only contract calendar")
	c.press("Clear contract calendar")
	c.press("Contract calendar range")
	c.press("Allow unavailable range dates")
	c.press("Contract calendar 2024-02-18")
	c.press("Contract calendar 2024-02-22")
	c.find("Calendar date: 2024-02-18 to 2024-02-22 · Changes: 5")
	c.press("Show two calendar months")
	c.press("Show calendar week numbers")
	c.press("Variable calendar weeks")
	c.press("Custom calendar parts")
	c.call("act", map[string]any{"id": c.find("Change contract calendar view").ID, "action": "scrollIntoView"}, nil)
	c.capture("calendar-range-multiple-months")
	c.press("Toggle calendar size")
	c.press("Toggle calendar variant")
	c.theme("Dark")
	c.call("act", map[string]any{"id": c.find("Change contract calendar view").ID, "action": "scrollIntoView"}, nil)
	c.capture("calendar-outline-dark")
	c.call("act", map[string]any{"id": c.find("Contract calendar 2024-03-15").ID, "action": "scrollIntoView"}, nil)
	c.capture("calendar-second-month")
	c.theme("Light")
	c.navigate("button", "Button")
}

func (c *client) checkSelectionParity() {
	focused := func(label string) bool {
		tr := c.tree()
		return slices.ContainsFunc(tr.Nodes, func(n node) bool { return n.ID == tr.Focus && n.Label == label })
	}
	c.navigate("listbox", "Listbox")
	c.edit("Contract filter", "cafe")
	c.key("Enter")
	c.find("Selected: [tea] · Changes: 0 · Activations: 0")
	c.capture("listbox-filter-cancelled")
	c.press("Protect cafe selection")
	c.press("Contract filter")
	c.key("Enter")
	c.find("Selected: [exact] · Changes: 1 · Activations: 1")
	c.press("Reorder contract choices")
	c.find("Selected: [exact] · Changes: 1 · Activations: 1")
	c.press("Contract filter")
	toggle := c.find("Toggle contract filter")
	c.call("act", map[string]any{"id": toggle.ID, "action": "scrollIntoView"}, nil)
	toggle = c.find("Toggle contract filter")
	x, y := toggle.Bounds.X+toggle.Bounds.Width/2, toggle.Bounds.Y+toggle.Bounds.Height/2
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "pointerDown", "x": x, "y": y}, map[string]any{"type": "pointerUp", "x": x, "y": y}}}, nil)
	c.idle()
	// Native accessibility focus resolves the list's active descendant.
	c.key("End")
	if !focused("A café") {
		c.t.Fatal("hiding Listbox search did not retain list navigation")
	}
	c.press("Contract choices loading")
	c.call("act", map[string]any{"id": c.find("Contract choices").ID, "action": "focus"}, nil)
	c.key("Enter")
	c.find("Selected: [exact] · Changes: 1 · Activations: 1")
	c.capture("listbox-loading")
	c.press("Contract choices loading")
	c.press("Custom choice parts")
	c.theme("Dark")
	c.capture("listbox-parts-dark")
	c.theme("Light")
	c.navigate("select-menu", "SelectMenu")
	c.press("Contract choices")
	c.edit("Contract filter", "cafe")
	c.key("Enter")
	c.find("Selected: [tea] · Changes: 0 · Activations: 0")
	c.find("Contract filter")
	c.capture("select-menu-filter-cancelled")
	c.key("Escape")
	c.press("Protect cafe selection")
	c.press("Contract choices")
	c.key("Enter")
	c.find("Selected: [exact] · Changes: 1 · Activations: 1")
	c.key("Tab")
	if !focused("Clear selection") {
		c.t.Fatal("SelectMenu clear action is not keyboard reachable")
	}
	c.capture("select-menu-clear-action")
	c.key("Enter")
	c.find("Selected: [] · Changes: 2 · Activations: 1")
	if c.tree().Focus != c.find("Contract choices").ID {
		c.t.Fatal("clearing SelectMenu lost trigger focus")
	}
	c.press("Focus menu search")
	c.press("Contract choices")
	c.key("End")
	if !focused("Tea") {
		c.t.Fatal("FocusSearch=false did not enable list navigation")
	}
	c.capture("select-menu-accessories")
	c.key("Escape")
	c.press("Remote choices")
	c.edit("Search Remote choices", "old")
	c.edit("Search Remote choices", "current")
	c.key("Escape")
	c.press("Complete stale options")
	c.find("Stale options rejected")
	c.press("Complete current options")
	c.find("Current options accepted")
	c.press("Remote choices")
	c.find("Result: current")
	c.capture("select-menu-remote-results")
	c.key("Enter")
	c.find("Remote selected: [current]")
	c.press("Cancel option request")
	c.find("Option request cancelled")
	c.navigate("button", "Button")
}

func (c *client) checkRatingParity() {
	c.navigate("input-rating", "InputRating")
	rating := c.find("Contract rating")
	c.call("act", map[string]any{"id": rating.ID, "action": "focus"}, nil)
	c.key("Right")
	c.find("Contract rating: 4 · Changes: 1")
	c.key("Space")
	c.find("Contract rating: 0 · Changes: 2")
	c.key("Space")
	c.find("Contract rating: 1 · Changes: 3")
	c.capture("rating-neutral-focus")
	c.press("Toggle rating icons")
	c.press("Custom rating layer colors")
	c.press("Preview rating on hover")
	// The native slider exposes one accessible node. Its five square items
	// share the remaining width as gaps; hover the fourth by actual geometry.
	box := c.find("Contract rating").Bounds
	gap := (box.Width - 5*box.Height) / 4
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "pointerMove", "x": box.X + 3*(box.Height+gap) + box.Height/2, "y": box.Y + box.Height/2}}}, nil)
	c.find("Contract rating: 1 · Changes: 3")
	c.capture("rating-icon-layer-hover")
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "pointerMove", "x": 0, "y": 0}}}, nil)
	c.call("act", map[string]any{"id": rating.ID, "action": "focus"}, nil)
	c.key("End")
	c.find("Contract rating: 5 · Changes: 4")
	for _, label := range []string{"Toggle rating orientation", "Shrink rating maximum"} {
		toggle := c.find(label)
		c.call("act", map[string]any{"id": toggle.ID, "action": "scrollIntoView"}, nil)
		toggle = c.find(label)
		x, y := toggle.Bounds.X+toggle.Bounds.Width/2, toggle.Bounds.Y+toggle.Bounds.Height/2
		c.call("input", map[string]any{"events": []any{map[string]any{"type": "pointerDown", "x": x, "y": y}, map[string]any{"type": "pointerUp", "x": x, "y": y}}}, nil)
		if c.find("Contract rating").ID != rating.ID || c.tree().Focus != rating.ID {
			c.t.Fatalf("%s discarded rating identity or focus", label)
		}
		if label == "Toggle rating orientation" {
			c.key("Up")
			c.find("Contract rating: 4 · Changes: 5")
		}
	}
	c.find("Contract rating: 3 · Changes: 5")
	c.capture("rating-vertical-clamped")
	c.press("Read-only contract rating")
	c.call("act", map[string]any{"id": rating.ID, "action": "focus"}, nil)
	c.key("Up")
	c.key("Space")
	box = c.find("Contract rating").Bounds
	x, y := box.X+box.Width/2, box.Y+box.Height/2 // Middle item of the vertical three-item rating.
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "pointerDown", "x": x, "y": y}, map[string]any{"type": "pointerUp", "x": x, "y": y}}}, nil)
	c.find("Contract rating: 3 · Changes: 5")
	c.capture("rating-readonly-layers")
	c.navigate("button", "Button")
}

func (c *client) checkAdvancedInputsParity() {
	c.navigate("input-number", "InputNumber")
	c.press("Increase Contract number")
	c.find("Contract number: {19 true} · Changes: 1")
	if c.tree().Focus != c.find("Contract number").ID {
		c.t.Fatal("number stepping did not focus the editor")
	}
	c.key("PageDown")
	c.find("Contract number: {9 true} · Changes: 2")
	c.press("Use empty number")
	c.press("Increase Contract number")
	c.find("Contract number: {40 true} · Changes: 3")
	editor := c.find("Contract number")
	c.press("Toggle increment control")
	if c.has("Increase Contract number") || c.find("Contract number").ID != editor.ID || c.find("Contract number").Bounds.Width <= editor.Bounds.Width {
		c.t.Fatal("hidden number control did not preserve and widen the editor")
	}
	c.press("Toggle increment control")
	c.press("Use number limit")
	control := c.find("Increase Contract number")
	c.call("act", map[string]any{"id": control.ID, "action": "scrollIntoView"}, nil)
	control = c.find("Increase Contract number")
	x, y := control.Bounds.X+control.Bounds.Width/2, control.Bounds.Y+control.Bounds.Height/2
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "pointerDown", "x": x, "y": y}}}, nil)
	time.Sleep(550 * time.Millisecond)
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "pointerUp", "x": x, "y": y}}}, nil)
	c.find("Contract number: {100 true} · Changes: 5")
	if !slices.Contains(c.find("Increase Contract number").States, "disabled") {
		c.t.Fatal("number hold did not stop at the upper bound")
	}
	c.capture("input-number-start-step-controls")
	c.navigate("input-tags", "InputTags")
	c.call("act", map[string]any{"id": c.find("Contract tags").ID, "action": "focus"}, nil)
	c.key("Backspace")
	c.find("Tag changes: 0 · Added: 0 · Removed: 0 · Rejected: 0")
	c.capture("input-tags-selected-label")
	c.key("Backspace")
	c.find("Tag changes: 1 · Added: 0 · Removed: 1 · Rejected: 0")
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "rust"}}}, nil)
	c.key("Tab")
	c.find("Tag changes: 2 · Added: 1 · Removed: 1 · Rejected: 0")
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "GO,Rust,python,"}}}, nil)
	c.find("Tag changes: 3 · Added: 2 · Removed: 1 · Rejected: 2")
	c.press("Add long tag")
	c.call("act", map[string]any{"id": c.find("Contract tags").ID, "action": "scrollIntoView"}, nil)
	c.capture("input-tags-inline-wrapped")
	c.press("Read-only contract tags")
	c.edit("Contract tags", "blocked")
	if c.find("Contract tags").Value != "" {
		c.t.Fatal("read-only tags accepted a draft")
	}
	c.find("Tag changes: 3 · Added: 2 · Removed: 1 · Rejected: 2")
	c.navigate("pin-input", "PinInput")
	c.press("Contract verification code")
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "123456"}}}, nil)
	c.find("PIN characters: 6 · Changes: 1 · Complete: 1")
	code := c.find("Contract verification code")
	c.press("Highlight all PIN cells")
	c.press("Custom PIN separator spacing")
	c.press("Custom PIN separator spacing")
	c.press("Highlight all PIN cells")
	if current := c.find("Contract verification code"); current.ID != code.ID || current.Value != "123456" {
		c.t.Fatal("PIN styling replaced its editor or value")
	}
	c.find("PIN characters: 6 · Changes: 1 · Complete: 1")
	c.capture("pin-input-highlight-separator-parts")
	c.navigate("button", "Button")
}

func (c *client) checkFormParity() {
	c.navigate("form", "Form")
	c.press("Submit validation")
	c.find("Workspace name is required")
	c.edit("validation workspace name", "Native form")
	c.press("Reject form with a global error")
	c.press("Submit validation")
	c.find("Workspace is unavailable")
	if c.has("Workspace name is required") {
		c.t.Fatal("form kept a field error that the new validation cleared")
	}
	c.find("Requests: 0 · Errors: 2 · Dirty: true")
	c.capture("form-global-validation")
	c.press("Reject form with a global error")
	c.press("Restore form name")
	c.find("Requests: 0 · Errors: 2 · Dirty: false")
	c.press("Submit validation")
	c.find("Workspace submitted · 1 successful submissions")
	c.press("Submit async")
	if !slices.Contains(c.find("async workspace name").States, "disabled") || !slices.Contains(c.find("Submit async").States, "disabled") {
		c.t.Fatal("pending form did not disable its controls")
	}
	c.press("Repeat pending submit")
	c.find("Requests: 1 · Errors: 0 · Dirty: false")
	c.capture("form-pending-disabled")
	c.press("Complete simulated request")
	c.find("Simulated request completed · 1 successful submissions")
	c.press("Disable form while submitting")
	c.press("Submit async")
	c.edit("async workspace name", "Edited while pending")
	c.press("Complete simulated request")
	c.find("Stale request rejected · 1 successful submissions")
	c.find("Requests: 2 · Errors: 0 · Dirty: true")
	c.capture("form-edit-rejects-stale-result")
	c.press("Submit async")
	c.navigate("button", "Button")
	c.navigate("form", "Form")
	c.press("Complete simulated request")
	c.find("Stale request rejected · 1 successful submissions")
	c.navigate("button", "Button")
}

func (c *client) checkNavigationParity() {
	pointerClick := func(label string) {
		n := c.find(label)
		c.call("act", map[string]any{"id": n.ID, "action": "scrollIntoView"}, nil)
		n = c.find(label)
		x, y := n.Bounds.X+n.Bounds.Width/2, n.Bounds.Y+n.Bounds.Height/2
		c.call("input", map[string]any{"events": []any{map[string]any{"type": "pointerDown", "x": x, "y": y}, map[string]any{"type": "pointerUp", "x": x, "y": y}}}, nil)
		c.idle()
	}
	c.navigate("breadcrumb", "Breadcrumb")
	c.press("Contract current")
	c.find("Path: current · Clicks: 1")
	c.press("Disable final path item")
	pointerClick("Contract current")
	c.find("Path: current · Clicks: 1")
	c.press("Disable final path item")
	c.press("Custom path separator")
	c.capture("breadcrumb-avatar-truncated-separators")
	c.navigate("pagination", "Pagination")
	c.press("Contract next page")
	c.find("Contract page: 6 · Changes: 1")
	c.capture("pagination-custom-controls-ellipsis")
	c.press("Shrink contract total")
	c.find("Contract page: 1 · Changes: 1")
	if !slices.Contains(c.find("Contract next page").States, "disabled") {
		c.t.Fatal("pagination controls remained enabled after the total shrank")
	}
	c.navigate("stepper", "Stepper")
	c.edit("Contract step draft", "Selected step draft")
	c.key("Ctrl+A")
	pointerClick("Reverse contract steps")
	pointerClick("Toggle contract step direction")
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "Edited step draft"}}}, nil)
	c.idle()
	if c.tree().Focus != c.find("Contract step draft").ID || c.find("Contract step draft").Value != "Edited step draft" {
		c.t.Fatal("step reorder or orientation lost editor selection")
	}
	c.capture("stepper-neutral-content-reordered")
	c.press("Disable contract steps")
	c.edit("Contract step draft", "Editable disabled step")
	if c.find("Contract step draft").Value != "Editable disabled step" {
		c.t.Fatal("disabled step headings disabled their content")
	}
	c.press("Disable contract steps")
	c.press("Contract step draft")
	pointerClick("Remove contract details")
	if c.has("Contract step draft") || c.tree().Focus != c.find("Contract review").ID {
		c.t.Fatal("removed step content did not restore a usable heading")
	}
	c.key("Space")
	c.find("Contract step: review · Changes: 1")
	c.find("Contract review body")
	c.navigate("timeline", "Timeline")
	c.press("Contract created")
	c.key("Space")
	c.find("Timeline: create · Selections: 2")
	c.press("Reverse contract timeline")
	c.press("Horizontal contract timeline")
	c.capture("timeline-dates-neutral-reverse")
	c.navigate("banner", "Banner")
	pointerClick("Contract banner action")
	c.find("Banner selections: 0 · Actions: 1 · Closed: 0")
	c.press("Contract banner")
	c.press("Contract banner action")
	button := c.find("Contract banner action").ID
	pointerClick("Toggle contract banner icon")
	if c.tree().Focus != button {
		c.t.Fatal("banner icon update lost action focus")
	}
	c.key("Enter")
	c.find("Banner selections: 1 · Actions: 3 · Closed: 0")
	c.press("Custom contract banner close")
	c.capture("banner-neutral-actions-close")
	pointerClick("Close contract banner")
	c.find("Banner selections: 1 · Actions: 3 · Closed: 1")
	if c.has("Contract banner") {
		c.t.Fatal("closed banner remains visible")
	}
	c.press("Show contract banner")
	c.find("Banner selections: 1 · Actions: 3 · Closed: 1")
	c.press("Close local banner")
	c.find("Local banner closes: 1")
	if c.has("Local announcement") {
		c.t.Fatal("unbound banner did not retain its local closed state")
	}
	c.navigate("error", "Error")
	c.press("Contract retry")
	c.press("Custom contract retry")
	c.press("Contract retry")
	c.find("Error retries: 2 · Recoveries: 0")
	c.capture("error-status-icon-custom-retry")
	c.press("Show contract retry")
	c.press("Replace contract error actions")
	c.press("Contract recover")
	c.find("Error retries: 2 · Recoveries: 1")
	if c.has("Contract retry") {
		c.t.Fatal("clear=false retained the default error action")
	}
	c.capture("error-replacement-actions")
	c.navigate("button", "Button")
}
