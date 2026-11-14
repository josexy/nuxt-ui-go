package nui_test

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

// Count the colored pixels across a stroke, including its immediate neighbors.
// Bounds are window DIPs (already zoomed); the image uses device pixels.
func strokePixels(im *image.RGBA, r ui.Rect, scale float32, vertical bool, ink color.RGBA) int {
	x, y := int((r.X+r.W/2)*scale), int((r.Y+r.H/2)*scale)
	lo, hi := int(math.Floor(float64(r.Y*scale)))-3, int(math.Ceil(float64((r.Y+r.H)*scale)))+3
	if vertical {
		lo, hi = int(math.Floor(float64(r.X*scale)))-3, int(math.Ceil(float64((r.X+r.W)*scale)))+3
	}
	n := 0
	for at := lo; at <= hi; at++ {
		px, py := x, at
		if vertical {
			px, py = at, y
		}
		if im.RGBAAt(px, py) == ink {
			n++
		}
	}
	return n
}

func dipScales(t *testing.T, tt *ui.Tester, check func(t *testing.T, scale, zoom float32)) {
	t.Helper()
	for _, scale := range []float32{1, 1.25, 1.5, 2} {
		tt.SetScale(scale)
		for _, zoom := range []float32{1, 1.19} {
			tt.SetZoom(zoom)
			t.Run(fmt.Sprintf("dpi=%g/zoom=%g", scale, zoom), func(t *testing.T) { check(t, scale, zoom) })
		}
	}
}

func TestDIPTableHeader(t *testing.T) {
	for _, dark := range []bool{false, true} {
		for _, font := range []float32{14, 21} {
			th := nui.Light()
			if dark {
				th = nui.Dark()
			}
			th.FontSize, th.Density = font, .8
			var state nui.TableState[string]
			state.SetRows([]string{"row"}, func(v string) string { return v })
			tt := ui.NewTester(func(c *ui.Context) {
				nui.Table(c.Key("table"), th, &state, []nui.TableColumn[string]{{ID: "value", Title: "Value", Value: func(v string) string { return v }}}, nui.TableOptions{Label: "Records", Decoration: nui.Decoration{Style: func(slot nui.Slot, e ui.Element) {
					if slot == nui.SeparatorSlot {
						e.Label("Header line")
					}
				}}}).Absolute().Left(10.3).Top(10.3).Width(330).Height(150)
			}, 400, 230)
			dipScales(t, tt, func(t *testing.T, scale, zoom float32) {
				r := bounds(t, tt, "Header line")
				if got := strokePixels(tt.Image(), r, scale, false, rgba(th.BorderAccented)); got != 1 {
					t.Fatalf("dark=%v font=%g scale=%g zoom=%g: header pixels=%d", dark, font, scale, zoom, got)
				}
			})
		}
	}
}

func TestDIPMenuHairlines(t *testing.T) {
	for _, contextMenu := range []bool{false, true} {
		for _, dark := range []bool{false, true} {
			th := nui.Light()
			if dark {
				th = nui.Dark()
			}
			th.Motion.Reduce = true
			var state nui.MenuState
			items := []nui.MenuItem{{ID: "first", Label: "First"}, {ID: "line1", Kind: nui.MenuSeparator}, {ID: "rich", Label: "Rich", Description: "Unequal row height"}, {ID: "line2", Kind: nui.MenuSeparator}, {ID: "sub", Label: "Submenu", Children: []nui.MenuItem{{ID: "child", Label: "Child"}, {ID: "line3", Kind: nui.MenuSeparator}, {ID: "last", Label: "Last"}}}}
			count := 0
			tt := ui.NewTester(func(c *ui.Context) {
				count = 0
				a := nui.Button(c.Key("anchor"), th, "Actions", nui.ButtonOptions{}).Absolute().Left(20.3).Top(20.3)
				o := nui.MenuOptions{Decoration: nui.Decoration{Style: func(slot nui.Slot, e ui.Element) {
					if slot == nui.SeparatorSlot {
						e.Label(fmt.Sprintf("line %d", count))
						count++
					}
				}}}
				if contextMenu {
					nui.ContextMenu(c, th, a, &state, items, o)
				} else {
					nui.DropdownMenu(c, th, a, &state, items, o)
				}
			}, 650, 430)
			dipScales(t, tt, func(t *testing.T, scale, zoom float32) {
				if !state.Open {
					click(t, tt, "Actions")
					if contextMenu {
						tt.Key(ui.Shift, ui.KeyF10)
					}
				}
				clickOver(t, tt, "Submenu")
				tt.Advance(time.Second)
				if count != 3 {
					t.Fatalf("separators=%d, want main and submenu", count)
				}
				for i := range count {
					r := bounds(t, tt, fmt.Sprintf("line %d", i))
					if got := strokePixels(tt.Image(), r, scale, false, rgba(th.Border)); got != 1 {
						t.Fatalf("context=%v dark=%v scale=%g zoom=%g line=%d: pixels=%d", contextMenu, dark, scale, zoom, i, got)
					}
				}
				tt.Key(0, ui.KeyEscape)
				tt.Key(0, ui.KeyEscape)
			})
		}
	}
}

func TestDIPComponentStrokes(t *testing.T) {
	for _, kind := range []string{"tabs", "navigation", "splitter", "stepper", "timeline", "listbox"} {
		for _, vertical := range []bool{false, true} {
			for _, dark := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/vertical=%v/dark=%v", kind, vertical, dark), func(t *testing.T) {
					th := nui.Light()
					if dark {
						th = nui.Dark()
					}
					th.Motion.Reduce = true
					ink := ui.RGB(255, 0, 255)
					active, split := "b", float32(101.3)
					var nav nui.NavigationState
					var list nui.ListboxState
					var selected []string
					count := 0
					tt := ui.NewTester(func(c *ui.Context) {
						count = 0
						slot := nui.Indicator
						if kind == "stepper" || kind == "timeline" || kind == "listbox" {
							slot = nui.SeparatorSlot
						}
						d := nui.Decoration{UI: nui.Styles{slot: {Background: nui.Set(ink), Radius: nui.Set(float32(0))}}, Style: func(s nui.Slot, e ui.Element) {
							if s == slot {
								e.Label(fmt.Sprintf("stroke %d", count))
								count++
							}
						}}
						ui.Box(c.Key("offset")).Absolute().Left(10.3).Top(10.3).Width(470).Height(360).Children(func() {
							switch kind {
							case "tabs":
								nui.Tabs(c, th, &active, []nui.TabItem{{ID: "a", Label: "Alpha"}, {ID: "b", Label: "Beta"}, {ID: "c", Label: "Gamma"}}, nui.TabsOptions{Underline: true, Vertical: vertical, Content: nui.Set(false), Decoration: d})
							case "navigation":
								items := []nui.NavigationItem{{ID: "a", Label: "Alpha"}, {ID: "b", Label: "Beta"}, {ID: "c", Label: "Gamma"}}
								if vertical {
									items = []nui.NavigationItem{{ID: "parent", Label: "Parent", DefaultOpen: true, Children: items}}
								}
								nui.NavigationMenu(c, th, &active, &nav, items, nui.NavigationOptions{Vertical: vertical, Highlight: true, Decoration: d})
							case "splitter":
								nui.Splitter(c, th, &split, nui.SplitterOptions{Vertical: !vertical, Decoration: d}, func() { ui.Text(c, "First pane") }, func() { ui.Text(c, "Second pane") }).Fill()
							case "stepper":
								nui.Stepper(c, th, &active, []nui.StepItem{{ID: "a", Title: "Alpha"}, {ID: "b", Title: "Beta"}, {ID: "c", Title: "Gamma"}, {ID: "d", Title: "Delta"}}, nui.StepperOptions{Vertical: vertical, Decoration: d}).Fill()
							case "timeline":
								nui.Timeline(c, th, []nui.TimelineItem{{ID: "a", Title: "Alpha", Description: "Details"}, {ID: "b", Title: "Beta"}, {ID: "c", Title: "Gamma"}}, nui.TimelineOptions{Horizontal: !vertical, Decoration: d}).Fill()
							case "listbox":
								nui.Listbox(c, th, &selected, &list, []nui.ListboxItem{{ID: "a", Label: "Alpha"}, {ID: "line1", Kind: nui.ListboxSeparator}, {ID: "b", Label: "Beta", Description: "Details"}, {ID: "line2", Kind: nui.ListboxSeparator}, {ID: "c", Label: "Gamma"}}, nui.ListboxOptions{Decoration: d}).Fill()
							}
						})
					}, 510, 410)
					dipScales(t, tt, func(t *testing.T, scale, zoom float32) {
						for pos := range 3 {
							active = []string{"a", "b", "c"}[pos]
							split = 101.3 + float32(pos)*11.3
							tt.Frame()
							width := float32(1)
							if kind == "stepper" || kind == "timeline" {
								width = 2
							}
							want := int(math.Round(float64(width * scale * zoom)))
							vert := vertical
							if kind == "listbox" {
								want, vert = 1, false
							}
							if count == 0 {
								t.Fatal("no stroke built")
							}
							for i := range count {
								r := bounds(t, tt, fmt.Sprintf("stroke %d", i))
								if got := strokePixels(tt.Image(), r, scale, vert, rgba(ink)); got != want {
									t.Fatalf("scale=%g zoom=%g stroke=%d r=%v pixels=%d want=%d", scale, zoom, i, r, got, want)
								}
							}
						}
					})
				})
			}
		}
	}
}

func TestDIPSeparatorThickness(t *testing.T) {
	for _, dark := range []bool{false, true} {
		for _, vertical := range []bool{false, true} {
			for _, size := range []nui.Size{nui.XS, nui.SM, nui.MD, nui.LG, nui.XL} {
				t.Run(fmt.Sprintf("dark=%v/vertical=%v/size=%v", dark, vertical, size), func(t *testing.T) {
					th := nui.Light()
					if dark {
						th = nui.Dark()
					}
					tt := ui.NewTester(func(c *ui.Context) {
						for i := range 6 {
							nui.Separator(c.Key(i), th, nui.SeparatorOptions{Vertical: vertical, Size: size, Decoration: nui.Decoration{Style: func(slot nui.Slot, e ui.Element) {
								if slot == nui.Indicator {
									e.Label(fmt.Sprint(i))
								}
							}}}).Absolute().Left(10.1 + float32(i)*19.3).Top(10.1 + float32(i)*19.3).Width(100).Height(100)
						}
					}, 260, 260)
					for _, scale := range []float32{1, 1.25, 1.5, 2} {
						tt.SetScale(scale)
						for _, zoom := range []float32{1, 1.19} {
							tt.SetZoom(zoom)
							want := int(math.Round(float64(float32(size-nui.XS+1) * scale * zoom)))
							for i := range 6 {
								r := bounds(t, tt, fmt.Sprint(i))
								if got := strokePixels(tt.Image(), r, scale, vertical, rgba(th.Border)); got != want {
									t.Fatalf("scale=%v zoom=%v line=%d bounds=%v: thickness=%d want=%d", scale, zoom, i, r, got, want)
								}
							}
						}
					}
				})
			}
		}
	}
}

func TestDIPSidebarBorderAlignment(t *testing.T) {
	for _, dark := range []bool{false, true} {
		for _, rtl := range []bool{false, true} {
			for _, side := range []nui.Side{nui.SideLeft, nui.SideRight} {
				th := nui.Light()
				if dark {
					th = nui.Dark()
				}
				var state nui.SidebarState
				collapsed := false
				tt := ui.NewTester(func(c *ui.Context) {
					ui.Row(c).Fill().Padding(10.3).Children(func() {
						nui.Sidebar(c, th, &collapsed, nui.SidebarOptions{Title: "Files", Side: nui.Set(side), State: &state, Resizable: true, Width: nui.Set(float32(220.3))}, nil)
					})
				}, 420, 300)
				if rtl {
					tt.SetDirection(ui.RightToLeft)
				}
				dipScales(t, tt, func(t *testing.T, scale, zoom float32) {
					for pos := range 4 {
						state.Width = nui.Set(float32(220.3) + float32(pos)*10.3)
						tt.Frame()
						r := bounds(t, tt, "Resize sidebar")
						tt.Move(r.X+r.W/2, r.Y+r.H/2)
						want := int(math.Round(float64(scale * zoom)))
						if got := strokePixels(tt.Image(), r, scale, true, rgba(th.Primary)); got != want {
							t.Fatalf("side=%v rtl=%v scale=%g zoom=%g: highlight=%d want=%d", side, rtl, scale, zoom, got, want)
						}
						if got := strokePixels(tt.Image(), r, scale, true, rgba(th.Border)); got != 0 {
							t.Fatalf("side=%v rtl=%v scale=%g zoom=%g: %d uncovered border pixels", side, rtl, scale, zoom, got)
						}
					}
				})
			}
		}
	}
}

// Scan a row/column, returning lengths of separated runs of the target color.
func dipRuns(im *image.RGBA, r ui.Rect, scale float32, vertical bool, matches func(color.RGBA) bool) []int {
	lo, hi := int(math.Round(float64(r.X*scale))), int(math.Round(float64((r.X+r.W)*scale)))
	fixed := int((r.Y + r.H/2) * scale)
	if vertical {
		lo, hi = int(math.Round(float64(r.Y*scale))), int(math.Round(float64((r.Y+r.H)*scale)))
		fixed = int((r.X + r.W/2) * scale)
	}
	var runs []int
	n := 0
	for at := lo; at < hi; at++ {
		x, y := at, fixed
		if vertical {
			x, y = fixed, at
		}
		if matches(im.RGBAAt(x, y)) {
			n++
			continue
		}
		if n > 0 {
			runs = append(runs, n)
			n = 0
		}
	}
	if n > 0 {
		runs = append(runs, n)
	}
	return runs
}

func TestDIPSeparatorPatternsAndLabel(t *testing.T) {
	for _, vertical := range []bool{false, true} {
		for _, pattern := range []nui.SeparatorType{nui.SeparatorSolid, nui.SeparatorDashed, nui.SeparatorDotted} {
			for _, size := range []nui.Size{nui.XS, nui.MD, nui.XL} {
				ink := ui.RGB(255, 0, 255)
				count := 0
				tt := ui.NewTester(func(c *ui.Context) {
					count = 0
					nui.Separator(c, nui.Light(), nui.SeparatorOptions{Label: "Continue", Vertical: vertical, Type: pattern, Size: size, Decoration: nui.Decoration{UI: nui.Styles{nui.Indicator: {Background: nui.Set(ink)}}, Style: func(slot nui.Slot, e ui.Element) {
						if slot == nui.Indicator {
							e.Label(fmt.Sprintf("segment %d", count))
							count++
						}
					}}}).Absolute().Top(10.3).Left(10.3).Width(380).Height(380)
				}, 410, 410)
				dipScales(t, tt, func(t *testing.T, scale, zoom float32) {
					if count != 2 {
						t.Fatal("label must have two line segments")
					}
					width := int(math.Round(float64(float32(size-nui.XS+1) * scale * zoom)))
					for i := range count {
						r := bounds(t, tt, fmt.Sprintf("segment %d", i))
						if pattern == nui.SeparatorSolid {
							if got := strokePixels(tt.Image(), r, scale, vertical, rgba(ink)); got != width {
								t.Fatalf("labeled solid width=%d want=%d", got, width)
							}
							continue
						}
						// Sample the snapped center, including antialiased dot edges.
						runs := dipRuns(tt.Image(), r, scale, vertical, func(c color.RGBA) bool { return int(c.R) > int(c.G)+30 && int(c.B) > int(c.G)+30 })
						want := width
						if pattern == nui.SeparatorDashed {
							want *= 3
						}
						if len(runs) < 2 {
							t.Fatalf("pattern=%v scale=%g zoom=%g: missing dashes: %v", pattern, scale, zoom, runs)
						}
						for _, got := range runs[:len(runs)-1] {
							if got != want {
								t.Fatalf("pattern=%v size=%v scale=%g zoom=%g: runs=%v want=%d", pattern, size, scale, zoom, runs, want)
							}
						}
					}
				})
			}
		}
	}
}

func TestDIPCardAndChartHairlines(t *testing.T) {
	for _, dark := range []bool{false, true} {
		th := nui.Light()
		if dark {
			th = nui.Dark()
		}
		th.Motion.Reduce = true
		ink := ui.RGB(255, 0, 255)
		th.Border, th.BorderMuted = ink, ink
		var chart nui.ChartState
		if err := chart.SetData([]float64{0, 1, 2}, []nui.ChartSeries{{ID: "values", Values: []float64{0, 5, 10}}}); err != nil {
			t.Fatal(err)
		}
		for _, kind := range []string{"card", "chart"} {
			tt := ui.NewTester(func(c *ui.Context) {
				if kind == "card" {
					nui.Card(c, th, nui.CardOptions{Variant: nui.Soft, Header: func() { ui.Box(c).Height(20.3) }, Footer: func() { ui.Box(c).Height(16.7) }}, func() { ui.Box(c).Height(50.3) }).Absolute().Top(10.3).Left(10.3).Width(300).Label("area")
				} else {
					nui.Chart(c, th, &chart, nui.ChartOptions{Height: 217.3, Decoration: nui.Decoration{Style: func(slot nui.Slot, e ui.Element) {
						if slot == nui.Body {
							e.Label("area")
						}
					}}}).Absolute().Top(10.3).Left(10.3).Width(300)
				}
			}, 340, 300)
			dipScales(t, tt, func(t *testing.T, scale, zoom float32) {
				r := bounds(t, tt, "area")
				if kind == "chart" {
					r.X += r.W * .9
					r.W = 1
				}
				runs := dipRuns(tt.Image(), r, scale, true, func(c color.RGBA) bool { return c == rgba(ink) })
				if len(runs) < 2 {
					t.Fatalf("%s missing lines: %v", kind, runs)
				}
				for _, n := range runs {
					if n != 1 {
						t.Fatalf("%s scale=%g zoom=%g: hairlines=%v", kind, scale, zoom, runs)
					}
				}
				if scale == 1.5 && zoom == 1.19 {
					parityScreenshot(t, tt, fmt.Sprintf("dip-%s-dark-%v", kind, dark))
				}
			})
		}
	}
}

func TestDIPTreeGuides(t *testing.T) {
	for _, rtl := range []bool{false, true} {
		for _, dark := range []bool{false, true} {
			th := nui.Light()
			if dark {
				th = nui.Dark()
			}
			ink := ui.RGB(255, 0, 255)
			var state nui.TreeState
			state.SetNodes([]nui.TreeNode{{ID: "root", Label: "Root", DefaultExpanded: true, Children: []nui.TreeNode{{ID: "branch", Label: "Branch", DefaultExpanded: true, Children: []nui.TreeNode{{ID: "leaf", Label: "Leaf"}}}, {ID: "sibling", Label: "Sibling"}}}})
			width := float32(1)
			tt := ui.NewTester(func(c *ui.Context) {
				nui.Tree(c, th, &state, nui.TreeOptions{Decoration: nui.Decoration{UI: nui.Styles{nui.Indicator: {BorderWidth: nui.Set(width), BorderColor: nui.Set(ink)}}, Style: func(slot nui.Slot, e ui.Element) {
					if slot == nui.Root {
						e.Label("Tree")
					}
				}}}).Absolute().Left(10.3).Top(10.3).Width(330)
			}, 460, 260)
			if rtl {
				tt.SetDirection(ui.RightToLeft)
			}
			dipScales(t, tt, func(t *testing.T, scale, zoom float32) {
				for _, w := range []float32{1, 2, 0} {
					width = w
					tt.Frame()
					for name, depth := range map[string]int{"Leaf": 2, "Sibling": 1} {
						r := bounds(t, tt, name)
						r.X, r.W = 0, 460
						runs := dipRuns(tt.Image(), r, scale, false, func(c color.RGBA) bool { return c == rgba(ink) })
						if w == 0 {
							depth = 0
						}
						if len(runs) != depth {
							t.Fatalf("rtl=%v row=%s guides=%v want=%d", rtl, name, runs, depth)
						}
						want := int(math.Round(float64(w * scale * zoom)))
						for _, n := range runs {
							if n != want {
								t.Fatalf("tree scale=%g zoom=%g: guides=%v want=%d", scale, zoom, runs, want)
							}
						}
					}
				}
			})
		}
	}
}

func TestDIPPinCaret(t *testing.T) {
	for _, dark := range []bool{false, true} {
		th := nui.Light()
		if dark {
			th = nui.Dark()
		}
		th.Motion.Reduce = true
		ink := ui.RGB(255, 0, 255)
		th.Highlighted = ink
		value := "123456"
		var state nui.PinState
		tt := ui.NewTester(func(c *ui.Context) {
			index := 0
			nui.PinInput(c, th, &value, &state, nui.PinInputOptions{Label: "Code", Length: 6, Decoration: nui.Decoration{UI: nui.Styles{nui.Label: {Foreground: nui.Set(th.Text)}}, Style: func(slot nui.Slot, e ui.Element) {
				if slot == nui.Item {
					e.Label(fmt.Sprintf("digit %d", index))
					index++
				}
			}}}).Absolute().Left(10.3).Top(10.3)
		}, 380, 100)
		dipScales(t, tt, func(t *testing.T, scale, zoom float32) {
			for i := range 6 {
				name := fmt.Sprintf("digit %d", i)
				click(t, tt, name)
				r := bounds(t, tt, name)
				// Skip the glyph and sample near the caret's top, inside the border.
				r.Y += 7 * zoom
				r.H = 0
				runs := dipRuns(tt.Image(), r, scale, false, func(c color.RGBA) bool { return c == rgba(ink) })
				want := int(math.Round(float64(scale * zoom)))
				if len(runs) != 1 || runs[0] != want {
					t.Fatalf("caret %d scale=%g zoom=%g: runs=%v want=%d", i, scale, zoom, runs, want)
				}
			}
		})
	}
}
