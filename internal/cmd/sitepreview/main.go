// Sitepreview runs one website example on the desktop or in a browser.
package main

import (
	"log"
	"os"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/internal/showcase"
	"github.com/josexy/nuxt-ui-go/internal/showcase/preview"
)

type app struct {
	demo    showcase.Demo
	config  showcase.Config
	example preview.Example
	window  *mygo.Window
	closed  bool
}

func (a *app) view(c *ui.Context) {
	t := nui.App(c, a.config.Theme())
	ui.Column(c.Key("preview")).Fill().Background(t.Background).Padding(24).Gap(16).Children(func() {
		a.example.View(c, t, a.config)
	})
}

func main() {
	id := "button"
	if len(os.Args) > 1 {
		id = os.Args[1]
	}
	demo, ok := showcase.Find(id)
	if !ok {
		log.Fatalf("unknown component %q", id)
	}
	cfg, err := demo.Normalize(showcase.Config{})
	if err != nil {
		log.Fatal(err)
	}
	example, err := preview.New(id)
	if err != nil {
		log.Fatal(err)
	}
	a := &app{demo: demo, config: cfg, example: example}
	mygo.App.SetLanguages(nui.SupportedLocales()...)
	mygo.App.WhenReady(func() {
		mygo.App.SetLocale(a.config.Locale)
		a.window = mygo.NewWindow(mygo.WindowOptions{Title: demo.Name + " · NuxtUI Go", Width: 760, Height: demo.Height, Content: ui.View(a.view)})
		a.example.SetWindow(a.window)
		a.window.OnClosed(func() { a.closed = true; a.example.Dispose() })
		connect(a)
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
