// Gallery displays components and application examples. Run: go run ./examples/gallery
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/internal/gallery"
)

func main() {
	hidden := flag.Bool("hidden", false, "create hidden native windows for automation")
	static := flag.Bool("static", false, "pause motion for deterministic screenshots and automation")
	width := flag.Int("width", 1080, "initial window width in DIP")
	height := flag.Int("height", 820, "initial window height in DIP")
	flag.Parse()
	mygo.App.SetLanguages("en")
	count := 0
	var open func()
	open = func() {
		count++
		state := gallery.New()
		state.Animations = !*static
		state.NewWindow = open
		window := mygo.NewWindow(mygo.WindowOptions{Title: fmt.Sprintf("NuxtUI Go Gallery %d", count), Width: *width, Height: *height, Hidden: true, Content: ui.View(state.View)})
		state.FileDialog.Parent = window
		window.OnClosed(state.Dispose)
		window.OnHide(func() { window.Update(state.Deactivate) })
		window.OnShow(func() { window.Update(func() { state.WindowInactive = false }) })
		if !*hidden {
			window.OnReadyToShow(func() {
				window.Show()
				// Windows may apply a launcher's SW_HIDE to the first Show call.
				if !window.IsVisible() {
					window.Show()
				}
			})
		}
	}
	mygo.App.WhenReady(open)
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
