package nui

import (
	"bytes"
	"context"
	"fmt"
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestCommandTextMatchesRichTextRendering(t *testing.T) {
	for _, dark := range []bool{false, true} {
		for _, scale := range []float32{1, 1.5} {
			for _, text := range []string{"", "Plain command", "打开 café e\u0301 🧭 command"} {
				for _, marks := range [][]int{nil, {0, 1, 2, 8}, {8, 1, 1, 0, -1, 200}} {
					t.Run(fmt.Sprintf("%v/%g/%q/%v", dark, scale, text, marks), func(t *testing.T) {
						original := slices.Clone(marks)
						legacy := false
						tt := ui.NewTester(func(c *ui.Context) {
							mode := LightMode
							if dark {
								mode = DarkMode
							}
							th := App(c, AppConfig{Mode: mode, DisableAnimations: true})
							th.TextScale = scale
							base := PartStyle{Font: Set(th.Font), FontSize: Set(th.text(14)), LineHeight: Set(float32(1.5)), Foreground: Set(th.Text), PaddingX: Set(float32(3))}
							if !legacy {
								commandText(c.Key("text"), th, text, marks, ItemLabelSlot, base, CommandPaletteOptions{}, CommandAction{}).Label("Caption")
								return
							}
							runes := []rune(text)
							var spans []ui.Span
							for start := 0; start < len(runes); {
								on, end := slices.Contains(marks, start), start+1
								for end < len(runes) && slices.Contains(marks, end) == on {
									end++
								}
								span := ui.Span{Text: string(runes[start:end])}
								if on {
									span.Color, span.Background, span.Radius = th.Primary, th.Primary.Alpha(.15), 2
								}
								spans = append(spans, span)
								start = end
							}
							e := ui.RichText(c.Key("text"), spans...).MinWidth(0).SingleLine().Label("Caption")
							finishCommandPart(th, ItemLabelSlot, e, base, Decoration{}, Decoration{})
						}, 380, 70)
						pixels := slices.Clone(tt.Image().Pix)
						before, _ := tt.Find("Caption")
						legacy = true
						tt.Frame()
						after, _ := tt.Find("Caption")
						if !bytes.Equal(pixels, tt.Image().Pix) || before != after {
							t.Fatalf("plain/ordered highlight rendering differs: %v / %v", before, after)
						}
						if !slices.Equal(marks, original) {
							t.Fatal("rendering mutated immutable result marks")
						}
					})
				}
			}
		}
	}
}

func TestCommandPaletteSubmitsOnlyChangedResultRows(t *testing.T) {
	index := NewCommandIndex([]CommandAction{{ID: "first", Label: "First command"}, {ID: "second", Label: "Second command"}})
	var state CommandPaletteState
	state.SetIndex(index)
	state.Open = true
	defer state.Close()
	options := CommandPaletteOptions{SearchInput: Set(false), Search: func(ctx context.Context, ticket QueryTicket, query, parent string) {
		state.Results.Complete(ticket, index.Search(ctx, query, parent, 100), nil)
	}}
	tt := ui.NewTester(func(c *ui.Context) {
		CommandPalette(c.Key("commands"), App(c, AppConfig{DisableAnimations: true}), &state, options)
	}, 800, 600)
	tt.Frame()
	if !tt.HasText("First command") || state.list.source.revision == 0 {
		t.Fatal("results were not submitted to the visible list")
	}
	// The submission revision measures actual publications, including setters
	// that might submit equal data again; it is not a data-equality assertion.
	revision := state.list.source.revision
	for range 4 {
		tt.Frame()
	}
	state.selected.Set([]string{"second"})
	tt.Frame()
	if state.list.source.revision != revision {
		t.Fatal("unchanged frames or selection resubmitted result rows")
	}
	options.UI = Styles{ItemLabelSlot: {Weight: Set(600)}}
	tt.Frame()
	if state.list.source.revision <= revision {
		t.Fatal("changed row configuration did not publish its descriptors")
	}
	revision = state.list.source.revision
	tt.Frame()
	if state.list.source.revision != revision {
		t.Fatal("changed configuration continued resubmitting after settling")
	}
	index = NewCommandIndex([]CommandAction{{ID: "first", Label: "Renamed command"}})
	state.SetIndex(index)
	tt.Frame()
	if !tt.HasText("Renamed command") || tt.HasText("First command") || state.list.source.revision <= revision {
		t.Fatal("replacement index did not update submitted rows")
	}
}
