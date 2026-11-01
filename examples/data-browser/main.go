package main

import (
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/internal/databrowser"
)

func main() {
	mygo.App.WhenReady(func() {
		state := databrowser.New(10000)
		mygo.NewWindow(mygo.WindowOptions{Title: "NuxtUI Go data browser", Width: 1100, Height: 800, Content: ui.View(state.View)})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
