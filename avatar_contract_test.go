package nui_test

import (
	"image"
	"image/color"
	"image/draw"
	"testing"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

func TestAvatarImageFillsStyledRootAndExposesPart(t *testing.T) {
	pixels := image.NewRGBA(image.Rect(0, 0, 20, 30))
	draw.Draw(pixels, pixels.Bounds(), image.NewUniform(color.RGBA{R: 210, G: 40, B: 50, A: 255}), image.Point{}, draw.Src)
	bitmap := ui.NewBitmap(pixels)
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Avatar(c, th, nui.AvatarOptions{Image: bitmap, Alt: "Photo", Style: func(slot nui.Slot, e ui.Element) {
			switch slot {
			case nui.Root:
				e.Size(48, 48)
			case nui.Body:
				e.Label("Avatar image")
			}
		}})
	}, 100, 100)
	root, img := bounds(t, tt, "Photo"), bounds(t, tt, "Avatar image")
	if img != root {
		t.Fatalf("image did not fill customized root: image=%+v root=%+v", img, root)
	}
}

func TestAvatarFallbackPriorityAndNames(t *testing.T) {
	bitmap := ui.NewBitmap(image.NewRGBA(image.Rect(0, 0, 20, 30)))
	opts := nui.AvatarOptions{Image: bitmap, Icon: lucide.User(), Text: "+12", Alt: "Ada Lovelace"}
	parts := map[nui.Slot]bool{}
	opts.Style = func(slot nui.Slot, _ ui.Element) { parts[slot] = true }
	tt := ui.NewTester(func(c *ui.Context) {
		clear(parts)
		nui.Avatar(c, nui.App(c, nui.AppConfig{}), opts)
	}, 100, 100)
	check := func(part nui.Slot, text string) {
		t.Helper()
		if !parts[part] || len(parts) != 2 || (text != "" && !tt.HasText(text)) {
			t.Fatalf("fallback priority: parts=%v, text=%q", parts, text)
		}
		bounds(t, tt, opts.Alt)
	}
	check(nui.Body, "")
	opts.Image = nil
	tt.Frame()
	check(nui.Leading, "")
	opts.Icon = nil
	tt.Frame()
	check(nui.Label, "+12")
	opts.Text = ""
	tt.Frame()
	check(nui.Label, "AL")
	opts.Alt = "林晓"
	tt.Frame()
	check(nui.Label, "林")
	opts.Alt, opts.Text = "", "+5"
	tt.Frame()
	bounds(t, tt, "+5")
	if bounds(t, tt, "+5").W != 32 {
		t.Fatal("literal Text was not used as the image name when Alt was empty")
	}
}
