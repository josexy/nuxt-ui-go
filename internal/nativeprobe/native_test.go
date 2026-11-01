// Package nativeprobe exercises the built gallery through MyGo's opt-in socket.
package nativeprobe

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image/png"
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
	c.t.Fatalf("window %d: missing %q", c.window, label)
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
	c.checkThemePalettes()
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
	c.press("Local dialog")
	c.find("Local theme dialog")
	c.capture("local-dialog-light")
	c.window = second
	for _, n := range c.tree().Nodes {
		if n.Role == "dialog" {
			t.Fatal("dialog leaked across windows")
		}
	}
	c.press("Inputs & overlays")
	if c.find("Plan").Value != "Team" {
		t.Fatal("initial plan")
	}
	c.window = first
	c.press("Close local dialog")
	c.press("Inputs & overlays")
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
	c.press("Plan")
	c.key("Home")
	c.key("Enter")
	if got := c.find("Plan").Value; got != "Personal" {
		c.capture("select-failure")
		t.Fatalf("native select: got %q", got)
	}
	c.capture("inputs-dark")
	c.window = second
	if c.find("Plan").Value != "Team" || c.find("Workspace name").Value != "Nuxt workspace" {
		t.Fatal("model leaked across windows")
	}
	// The application locale must update every existing window. Keep a modal
	// open in the first window while changing the language from the second.
	c.window = first
	c.press("Edit workspace")
	closeID := c.find("Close Edit workspace").ID
	c.window = second
	c.press("Language / 语言")
	if c.find("Language / 语言").Description != "简体中文" {
		t.Fatal("language switch did not update the second window")
	}
	c.capture("i18n-second-zh")
	c.window = first
	if c.find("关闭Edit workspace").ID != closeID || c.find("Dialog workspace name").Value != "原生中文验证" {
		t.Fatal("global locale change reset a modal or its edited value")
	}
	c.capture("i18n-modal-zh")
	c.window = second
	c.press("Language / 语言")
	if c.find("Language / 语言").Description != "English" {
		t.Fatal("language switch did not restore English")
	}
	c.window = first
	if c.find("Close Edit workspace").ID != closeID {
		t.Fatal("locale change replaced the modal close control")
	}
	c.press("Close Edit workspace")
	c.window = first
	c.press("Edit workspace")
	c.press("Dialog plan")
	c.capture("modal-select-dark")
	c.key("Escape")
	c.find("Save workspace")
	c.key("Escape")
	tr := c.tree()
	// MyGo omits accessibility focus while a different OS window is active.
	if tr.Focus != "" && tr.Focus != c.find("Edit workspace").ID {
		t.Fatal("native focus points to the wrong trigger")
	}
	c.key("Enter")
	c.find("Save workspace") // the restored opener must activate from the keyboard
	c.key("Escape")
	c.press("Data & menus")
	c.press("Workspace actions")
	c.press("Export")
	c.press("CSV file")
	c.find("Exported CSV")
	c.press("Rich actions")
	c.press("Export options")
	c.capture("rich-menu-dark")
	c.press("Export rich CSV")
	c.find("Rich CSV exported")
	c.capture("table-dark")
	c.press("Workspace")
	c.press("Save workspace")
	c.find("Workspace saved 2 times")
	c.press("Add example document")
	c.press("Foundation component notes")
	c.find("Document opened")
	c.press("Remove document")
	c.capture("workspace-dark")
	c.press("Display")
	c.find("1 more person")
	c.press("Notifications")
	c.find("3 unread notifications")
	c.capture("display-dark")
	c.press("Feedback")
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
	c.press("Dismiss Your changes are saved")
	for _, n := range c.tree().Nodes {
		if n.Role == "status" && n.Label == "Your changes are saved" {
			t.Fatal("alert did not close")
		}
	}
	c.capture("feedback-dark")
	c.press("Theme & buttons")
	c.press("Choose primary color")
	c.press("Blue")
	c.press("Toggle density")
	c.capture("theme-blue-compact-dark")
	c.window = second
	c.press("Workspace")
	c.press("Save workspace")
	c.find("Workspace saved 1 times")
	c.capture("workspace-second-light")
	c.press("Forms & settings")
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
	c.press("Forms & settings")
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
	c.press("Overlays & menus")
	c.press("Project actions")
	if n := c.find("Pin workspace"); n.Role != "menuItemCheckBox" {
		t.Fatalf("checkbox menu role: %+v", n)
	}
	c.key("End")
	c.key("Up")
	c.key("Right")
	c.capture("menu-light")
	c.key("Enter")
	c.find("CSV export prepared")
	contextTarget := c.find("Project context area")
	c.call("act", map[string]any{"id": contextTarget.ID, "action": "focus"}, nil)
	c.idle()
	c.key("Shift+F10")
	c.find("Project context menu")
	c.capture("context-light")
	c.key("Escape")
	c.press("Review changes")
	c.press("Review plan")
	c.capture("modal-select-light")
	c.key("Escape")
	c.find("Confirm changes")
	c.key("Escape")
	c.press("Bottom panel")
	c.capture("bottom-light")
	c.press("Done with preferences")
	c.press("Navigation & disclosure")
	c.press("Sharing permissions")
	c.capture("navigation-light")
	c.press("Activity")
	c.press("Page Next")
	c.find("Activity page 2")
	c.capture("timeline-light")
	c.press("Try again")
	c.find("Activity restored")
	c.press("Advanced forms")
	c.press("Increase Seats")
	if n := c.find("Seats"); n.Value != "4" {
		t.Fatalf("native number: %+v", n)
	}
	seatsID := c.find("Seats").ID
	c.press("Number orientation")
	c.press("vertical")
	if n := c.find("Seats"); n.ID != seatsID || n.Value != "4" {
		t.Fatalf("orientation reset the native editor: %+v", n)
	}
	c.press("Increase Seats")
	if c.find("Seats").Value != "5" {
		t.Fatal("vertical number increase failed")
	}
	c.capture("input-number-vertical")
	c.press("Number orientation")
	c.press("horizontal")
	c.press("Decrease Seats")
	if n := c.find("Seats"); n.ID != seatsID || n.Value != "4" {
		t.Fatalf("horizontal number decrease or editor identity failed: %+v", n)
	}
	c.capture("input-number-horizontal")
	c.press("Verification code")
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "123456"}}}, nil)
	c.idle()
	if n := c.find("Verification code"); n.Value != "123456" {
		t.Fatalf("native PIN: %+v", n)
	}
	c.capture("advanced-light")
	c.press("Project members")
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "Ada"}}}, nil)
	c.idle()
	c.key("Down")
	c.key("Enter")
	c.capture("select-menu-light")
	c.key("Escape")
	c.press("Dates & color")
	c.press("Choose Release date")
	c.press("Release date calendar 2026-11-15")
	if n := c.find("Release date"); n.Value != "2026-11-15" {
		t.Fatalf("native date: %+v", n)
	}
	c.press("Accent color A")
	c.key("Ctrl+A")
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "128"}}}, nil)
	c.idle()
	c.capture("dates-light")
	c.theme("Dark")
	c.capture("dates-dark")
	c.press("Task browser")
	c.press("Task 00001")
	c.press("Edit task")
	c.capture("task-edit-dark")
	c.press("Cancel task edit")
	c.press("Filters")
	c.capture("task-filters-dark")
	c.press("Apply filters")
	c.press("Forms & settings")
	c.press("Settings: Project setup")
	c.press("Continue")
	c.find("Enter a project name")
	c.capture("wizard-dark")

	c.press("Advanced interactions")
	c.capture("interactions-dark")
	c.press("Next Review guide")
	c.find("Inspect recent activity")
	c.press("Search commands")
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
	c.press("Open inspector")
	c.press("Resize Activity inspector")
	c.key("End")
	c.capture("drawer-light")
	c.key("Escape")
	c.press("Activity & attachments")
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
	c.press("Add sample attachment")
	c.find("review.md")
	c.capture("chart-files-light")
	c.press("Application layout")
	c.press("Workspace settings")
	c.press("Members")
	c.find("Opened members")
	c.capture("shell-light")
	c.press("Toggle sidebar")
	c.capture("shell-compact-light")

	c.press("Props & behavior")
	c.press("Loading switch")
	if !slices.Contains(c.find("Sync workspace").States, "disabled") {
		t.Fatal("loading switch was not disabled")
	}
	c.press("Loading switch")
	c.capture("props-forms-light")
	c.press("Selection policies")
	c.press("Team menu")
	c.call("input", map[string]any{"events": []any{map[string]any{"type": "text", "text": "Alpha"}}}, nil)
	c.idle()
	c.press("Alpha")
	c.press("Team menu")
	c.find("Beta")
	c.key("Escape")
	c.capture("props-selection-light")
	c.press("Orientation")
	editorID := c.find("Orientation name").ID
	c.press("Shared orientation")
	c.press("vertical")
	if c.find("Orientation name").ID != editorID {
		t.Fatal("shared orientation replaced editor")
	}
	minimum := c.find("Orientation range minimum")
	c.call("act", map[string]any{"id": minimum.ID, "action": "focus"}, nil)
	c.key("Up")
	if n := c.find("Orientation range minimum"); n.Now == nil || *n.Now != 25 {
		t.Fatalf("native vertical range: %+v", n)
	}
	c.capture("props-vertical-light")
	c.press("Reverse sliders")
	minimum = c.find("Orientation range minimum")
	c.call("act", map[string]any{"id": minimum.ID, "action": "focus"}, nil)
	c.key("Up")
	if n := c.find("Orientation range minimum"); n.Now == nil || *n.Now != 20 {
		t.Fatalf("native inverted range: %+v", n)
	}
	c.press("Reverse sliders")
	c.press("Shared orientation")
	c.press("horizontal")
	c.capture("props-horizontal-light")

}
