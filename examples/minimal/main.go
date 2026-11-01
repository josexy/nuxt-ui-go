// Minimal demonstrates creating a window, applying a theme, and handling a button click.
package main

import (
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

func main() {
	config := nui.DefaultApp()
	count := 0
	mygo.App.SetLanguages(nui.SupportedLocales()...)
	mygo.App.WhenReady(func() {
		mygo.NewWindow(mygo.WindowOptions{
			Title: "NuxtUI Go minimal", Width: 480, Height: 240,
			Content: ui.View(func(c *ui.Context) {
				theme := nui.App(c, config)
				nui.Container(c, theme, nui.ContainerOptions{}, func() {
					ui.Textf(c, "Saved %d times", count)
					nui.Button(c.Key("save"), theme, "Save", nui.ButtonOptions{}).
						OnClick(func() { count++ })
				}).PaddingY(24)
			}),
		})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
