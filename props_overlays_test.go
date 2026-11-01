package nui_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

func TestPropsHoverPopoverAndControlledTooltip(t *testing.T) {
	open, tip, dismiss := false, false, true
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(30).Gap(12).Children(func() {
			a := nui.Button(c, th, "Hover", nui.ButtonOptions{})
			nui.Popover(c, th, a, &open, nui.PopoverOptions{Label: "Hover panel", Mode: nui.PopoverHover, OpenDelay: 200 * time.Millisecond, CloseDelay: 100 * time.Millisecond, Dismissible: nui.Set(dismiss)}, func(th nui.Theme) { nui.Button(c, th, "Inside", nui.ButtonOptions{}) })
			a = nui.Button(c, th, "Help", nui.ButtonOptions{})
			nui.Tooltip(c, th, a, "Help text", nui.TooltipOptions{Open: &tip, Placement: nui.Set(nui.RightStart), Kbds: []string{"Ctrl", "K"}})
		})
	}, 600, 400)
	r := bounds(t, tt, "Hover")
	tt.Move(r.X+4, r.Y+4)
	tt.Advance(199 * time.Millisecond)
	if open {
		t.Fatal("hover delay ignored")
	}
	tt.Advance(2 * time.Millisecond)
	if !open || !tt.HasText("Inside") {
		t.Fatal("hover did not open")
	}
	p := bounds(t, tt, "Hover panel")
	tt.Move(r.X+4, (r.Y+r.H+p.Y)/2)
	tt.Advance(150 * time.Millisecond)
	if !open {
		t.Fatal("anchor gap closed popup")
	}
	tt.Move(p.X+8, p.Y+8)
	tt.Advance(150 * time.Millisecond)
	if !open {
		t.Fatal("hover panel did not keep popup open")
	}
	tt.Move(590, 390)
	tt.Advance(99 * time.Millisecond)
	if !open {
		t.Fatal("close delay ignored")
	}
	tt.Advance(2 * time.Millisecond)
	if open {
		t.Fatal("hover did not close")
	}
	tt.Move(r.X+4, r.Y+4)
	tt.Advance(201 * time.Millisecond)
	dismiss = false
	tt.Frame()
	tt.Key(0, ui.KeyEscape)
	if !open {
		t.Fatal("non-dismissible Escape")
	}
	dismiss = true
	tt.Frame()
	tt.Key(0, ui.KeyEscape)
	tt.Advance(time.Second)
	if open {
		t.Fatal("dismissed hover reopened without pointer leaving")
	}
	tip = true
	tt.Frame()
	if !tt.HasText("Help text") || !tt.HasText("K") {
		t.Fatal("controlled tooltip/keys")
	}
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if tip || tt.HasText("Help text") {
		t.Fatal("controlled tooltip Escape")
	}
}

func TestPropsOrientationPreservesFieldAndCarousel(t *testing.T) {
	direction := nui.Vertical
	text := "original"
	var carousel nui.CarouselState
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(20).Gap(20).Children(func() {
			nui.FormField(c, th, nui.FormFieldOptions{Label: "Name", Orientation: nui.Set(direction)}, func(th nui.Theme) { nui.Input(c, th, &text, nui.InputOptions{Label: "Editor"}) })
			nui.Carousel(c, th, &carousel, []nui.CarouselItem{{ID: "a", Label: "Slide A"}, {ID: "b", Label: "Slide B"}, {ID: "c", Label: "Slide C"}}, nui.CarouselOptions{Label: "Guide", Orientation: direction, Arrows: nui.Set(false), Dots: true}).Width(400)
		})
	}, 600, 500)
	click(t, tt, "Name")
	if !tt.Focused("Editor") {
		t.Fatal("field label association")
	}
	tt.Key(ui.Ctrl, ui.KeyA)
	direction = nui.Horizontal
	tt.Frame()
	tt.Type("edited")
	if text != "edited" || !tt.Focused("Editor") {
		t.Fatal("orientation change lost editor identity/selection")
	}
	direction = nui.Vertical
	tt.Frame()
	if tt.HasText("Next Guide") {
		t.Fatal("arrows=false")
	}
	click(t, tt, "Guide slides")
	tt.Key(0, ui.KeyDown)
	if carousel.Index != 1 {
		t.Fatal("vertical carousel keyboard")
	}
	click(t, tt, "Go to slide 3")
	if carousel.Index != 2 {
		t.Fatal("dots navigation")
	}
	click(t, tt, "Guide slides")
	tt.Key(0, ui.KeyHome)
	r := bounds(t, tt, "Guide slides")
	tt.Press(r.X+40, r.Y+140)
	tt.Move(r.X+40, r.Y+20)
	tt.Advance(time.Second)
	tt.Release(r.X+40, r.Y+20)
	if carousel.Index != 1 {
		t.Fatalf("vertical drag %d", carousel.Index)
	}
}

func TestPropsToastActionsAndProgress(t *testing.T) {
	actions, closed := 0, 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Button(c, th, "Notify", nui.ButtonOptions{}).OnClick(func() {
			nui.Toast(c, th, nui.ToastOptions{Title: "Ready", Description: "Two actions", Timeout: time.Second, Orientation: nui.Set(nui.Horizontal), Close: nui.Set(false), OnClose: func() { closed++ }, Actions: []nui.ToastActionOptions{{Label: "Keep", Close: nui.Set(false), OnAction: func() { actions++ }}, {Label: "Done", OnAction: func() { actions++ }}}})
		})
		nui.ToastViewport(c, th, nui.ToastViewportOptions{})
	}, 650, 350)
	click(t, tt, "Notify")
	if tt.HasText("Dismiss Ready") {
		t.Fatal("close=false")
	}
	click(t, tt, "Keep")
	tt.Advance(2 * time.Second)
	if !tt.HasText("Ready") || actions != 1 || closed != 0 {
		t.Fatal("action retention / focused toast pause")
	}
	click(t, tt, "Done")
	tt.Advance(time.Second)
	if tt.HasText("Ready") || actions != 2 || closed != 1 {
		t.Fatalf("toast close actions=%d close=%d", actions, closed)
	}
}

func TestPropsRatingClearableAndPinSeparator(t *testing.T) {
	rating, pin := 3, "123456"
	var state nui.PinState
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Row(c).Gap(24).Children(func() {
			star := 0
			nui.InputRating(c, th, &rating, nui.InputRatingOptions{Label: "Stars", Orientation: nui.Vertical, Size: nui.LG, Clearable: true, Style: func(slot nui.Slot, e ui.Element) {
				if slot == nui.Item {
					star++
					e.Label(fmt.Sprintf("Star %d", star))
				}
			}})
			i := 0
			nui.PinInput(c, th, &pin, &state, nui.PinInputOptions{Numeric: true, Length: 6, Label: "PIN", Size: nui.LG, Separator: []int{3}, Style: func(slot nui.Slot, e ui.Element) {
				if slot == nui.Item {
					i++
					e.Label(fmt.Sprintf("Slot %d", i))
				}
			}})
		})
	}, 600, 350)
	r := bounds(t, tt, "Stars")
	tt.Move(r.X+r.W/2, r.Y+3)
	if rating != 3 {
		t.Fatal("hover changed value")
	}
	click(t, tt, "Star 3")
	if rating != 0 {
		t.Fatalf("clearable rating %d", rating)
	}
	click(t, tt, "Slot 4")
	tt.Type("9")
	if pin != "123956" {
		t.Fatalf("separator hit mapping %s", pin)
	}
}

func TestPropsFileUploadPresentationAndInteraction(t *testing.T) {
	var state nui.FileUploadState
	var calls atomic.Int32
	interactive, preview, remove := false, true, false
	choose := func(context.Context) ([]nui.UploadFile, error) {
		calls.Add(1)
		return []nui.UploadFile{{ID: "one", Name: "one.txt", Size: 12}}, nil
	}
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.FileUpload(c, th, &state, nui.FileUploadOptions{Label: "Files", Interactive: nui.Set(interactive), Preview: nui.Set(preview), FileDelete: nui.Set(remove), Dropzone: nui.Set(false), Layout: nui.UploadGrid, Choose: choose}).Width(450)
	}, 500, 400)
	t.Cleanup(state.Dispose)
	r := bounds(t, tt, "Files")
	tt.Press(r.X+10, r.Y+10)
	tt.Release(r.X+10, r.Y+10)
	if calls.Load() != 0 || state.Pending() {
		t.Fatal("interactive=false opened picker")
	}
	click(t, tt, "Choose files")
	until := time.Now().Add(time.Second)
	for state.Pending() && time.Now().Before(until) {
		time.Sleep(time.Millisecond)
		tt.Frame()
	}
	if calls.Load() != 1 || !tt.HasText("one.txt") || tt.HasText("Remove one.txt") {
		t.Fatal("choose/fileDelete behavior")
	}
	preview = false
	tt.Frame()
	if tt.HasText("one.txt") || len(state.Files) != 1 {
		t.Fatal("preview=false changed file data")
	}
	preview, remove = true, true
	tt.Frame()
	click(t, tt, "Remove one.txt")
	if len(state.Files) != 0 {
		t.Fatal("remove option")
	}
}
