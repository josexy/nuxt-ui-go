package nui

import "github.com/egoist/mygo/ui"

// statusText builds text as a polite live status (ui.RoleStatus) under a
// fixed key. MyGo tells a status from the last frame's by its element, so
// the key keeps siblings that come and go before it from turning it into
// another status, whose text change would be read out. A status reads its
// whole text when it changes, and nothing when it first shows or clears.
func statusText(c *ui.Context, key, text string) ui.Element {
	return ui.Text(c.Key(key), text).Role(ui.RoleStatus)
}

// shownMessages remembers which transient messages a component instance
// showed in its last build: its progress message and its error.
type shownMessages struct{ progress, err bool }

// announceShown reports whether a transient message shows in this build,
// and announces its text on the build where it starts showing, which its
// status does not read. *shown holds whether it showed in the last build;
// keep it in the component's ui.Local memory so it follows the instance and
// a second build pass of the same frame does not announce again.
func announceShown(c *ui.Context, shown *bool, show bool, text string) bool {
	show = show && text != ""
	if show && !*shown {
		c.Announce(text)
	}
	*shown = show
	return show
}
