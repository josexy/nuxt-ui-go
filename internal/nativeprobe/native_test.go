// Package nativeprobe exercises the built gallery through MyGo's opt-in socket.
package nativeprobe

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
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
	c.edit("Search components", id)
	search := c.find("Search components")
	for _, n := range c.tree().Nodes {
		if n.Label == label && n.Role == "button" && n.Bounds.Y > search.Bounds.Y && n.Bounds.X < search.Bounds.X+search.Bounds.Width {
			c.call("act", map[string]any{"id": n.ID, "action": "scrollIntoView"}, nil)
			c.idle()
			c.call("act", map[string]any{"id": n.ID, "action": "focus"}, nil)
			c.idle()
			c.call("act", map[string]any{"id": n.ID, "action": "press"}, nil)
			c.idle()
			if c.has("Search components") {
				c.edit("Search components", "")
			}
			return
		}
	}
	c.t.Fatalf("directory is missing %s (%s)", label, id)
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
		c.t.Fatal("native outer Splitter did not resize independently")
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

func TestNativeGallery(t *testing.T) {
	socket := os.Getenv("NUI_AUTOMATION_SOCKET")
	if socket == "" {
		t.Skip("run scripts/test-native.ps1 on Windows or bash scripts/test-native.sh on Linux")
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
	if os.Getenv("NUI_NATIVE_NARROW") == "1" {
		c.checkNarrowDirectory()
		c.checkListboxStates()
		c.checkGalleryStateExamples()
		return
	}
	c.checkGallerySidebar()
	c.checkThemePalettes()
	c.checkNestedSplitters()
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
	if progress.Role != "progress" || progress.Now == nil || *progress.Now != 42 || progress.Max == nil || *progress.Max != 100 {
		t.Fatalf("progress semantics: %+v", progress)
	}
	t.Logf("Progress accessibility range: %g / %g", *progress.Now, *progress.Max)
	if unknown := c.find("Synchronizing"); unknown.Now == nil || unknown.Min == nil || *unknown.Now >= *unknown.Min || unknown.Value != "In progress" {
		t.Fatal("indeterminate progress lost MyGo's unknown-value sentinel")
	}
	c.press("Advance progress")
	c.find("Preparing files · 59%")
	if got := c.find("Preparing files"); got.Now == nil || *got.Now != 59 {
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
	c.capture("interactions-dark")
	c.press("Next horizontal review guide")
	c.find("Slide 2 · 1 changes · paused false")
	c.navigate("command-palette", "CommandPalette")
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
	c.press("Open basic drawer")
	c.press("Resize basic inspector")
	c.key("End")
	c.capture("drawer-light")
	c.key("Escape")
	c.navigate("chart", "Chart")
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
	c.press("Add native sample")
	c.find("review-1.md")
	c.capture("chart-files-light")
	c.navigate("navigation-menu", "NavigationMenu")
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
