package main

import (
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/internal/settings"
)

func main() {
	mygo.App.WhenReady(func() {
		state := settings.New()
		mygo.NewWindow(mygo.WindowOptions{Title: "NuxtUI Go settings", Width: 960, Height: 900, Content: ui.View(state.View)})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
