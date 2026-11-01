package nui

import "github.com/egoist/mygo/ui"

// Window applies a value snapshot to this window for the current build.
// Pass the returned theme and the live Context explicitly to constructors.
func Window(c *ui.Context, theme Theme) Theme {
	c.SetTheme(theme.Native())
	return theme
}

// Themed builds a column with a local theme. It intentionally has layout and a
// background, just like ui.Themed. The callback executes in the current build.
func Themed(c *ui.Context, theme Theme, children func(Theme)) ui.Element {
	return ui.Themed(c, theme.Native()).Children(func() { children(theme) })
}
