package nui

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
)

// Keep datasets separate: a choice fixture must not retain an unrelated tree,
// menu, and table in its heap. Source generation is outside each timed loop.
func performanceItems(n int) []ListboxItem {
	items := make([]ListboxItem, n)
	for i := range items {
		items[i] = ListboxItem{ID: strconv.Itoa(i), Label: fmt.Sprintf("Item %06d", i)}
	}
	return items
}

func performanceOptions(n int) []Option[string] {
	items := make([]Option[string], n)
	for i := range items {
		items[i] = Option[string]{Value: strconv.Itoa(i), Label: fmt.Sprintf("Item %06d", i)}
	}
	return items
}

type performanceBuilds struct {
	views, rows int
}

func (p *performanceBuilds) item(slot Slot, _ ui.Element) {
	if slot == Item {
		p.rows++
	}
}

func (p *performanceBuilds) steady(b *testing.B, tt *ui.Tester) {
	b.Helper()
	tt.Move(690, 390)
	tt.Frame()
	before := p.views
	tt.Frame()
	if p.views-before != 1 {
		b.Fatalf("settled operation built %d views, want 1", p.views-before)
	}
	p.views, p.rows = 0, 0
	b.ReportAllocs()
	for b.Loop() {
		tt.Frame()
	}
	b.ReportMetric(float64(p.views)/float64(b.N), "builds/op")
	b.ReportMetric(float64(p.rows)/float64(b.N), "rows/op")
	if p.views != b.N {
		b.Fatalf("%d operations built %d views", b.N, p.views)
	}
}

// Selection-count curves supplement the default-row fixtures in
// BenchmarkDataComponents. The Style callback counts constructed rows without
// replacing their content. Compare like-for-like instrumented fixtures only.
func BenchmarkPerformanceSelection(b *testing.B) {
	for _, kind := range []string{"Listbox", "SelectMultiple", "SelectMenu", "InputMenu"} {
		for _, n := range []int{100, 10000, 100000} {
			for _, s := range []int{0, 1, n} {
				for _, open := range []bool{false, true} {
					if kind == "Listbox" && !open {
						continue
					}
					b.Run(fmt.Sprintf("%s/N=%d/S=%d/Open=%v", kind, n, s, open), func(b *testing.B) {
						values := make([]string, s)
						for i := range values {
							values[i] = strconv.Itoa(i)
						}
						var items []ListboxItem
						var options SelectState[string]
						var selected SelectionState[string]
						var search SearchState
						var builds performanceBuilds
						if kind != "InputMenu" {
							selected.Set(values)
						}
						switch kind {
						case "SelectMultiple":
							options.SetOptions(performanceOptions(n))
						case "InputMenu":
							source := performanceOptions(n)
							for i := range source {
								label := source[i].Label
								source[i].Build = func(c *ui.Context, _ Theme) { builds.rows++; ui.Text(c, label) }
							}
							search.SetOptions(source)
							search.SelectedValues.Set(values)
						default:
							items = performanceItems(n)
						}
						var list ListboxState
						menu := SelectMenuState{Open: open}
						switch kind {
						case "Listbox":
							list.SetItems(items)
						case "SelectMenu":
							menu.SetItems(items)
						}
						d := Decoration{Style: builds.item}
						tt := ui.NewTester(func(c *ui.Context) {
							builds.views++
							th := App(c, AppConfig{DisableAnimations: true})
							switch kind {
							case "Listbox":
								Listbox(c.Key("choices"), th, &selected, &list, ListboxOptions{Height: 320, Multiple: true, Decoration: d}).Width(640)
							case "SelectMultiple":
								SelectMultiple(c.Key("choices"), th, &selected, &options, SelectOptions{Open: &open, Decoration: d}).Width(640)
							case "SelectMenu":
								SelectMenu(c.Key("choices"), th, &selected, &menu, SelectMenuOptions{Multiple: true, SearchInput: Set(false), Decoration: d}).Width(640)
							case "InputMenu":
								InputMenu(c.Key("choices"), th, &search, InputMenuOptions{Label: "Choices", Multiple: true, ExternalFilter: true, Open: &open}).Width(640)
							}
						}, 700, 400)
						builds.steady(b, tt)
						if kind == "InputMenu" {
							tags := 0
							for _, text := range tt.Texts() {
								if strings.HasPrefix(text, "Remove Item ") {
									tags++
								}
							}
							b.ReportMetric(float64(tags), "tags/lastbuild")
							if tags > 16 {
								b.Fatalf("built %d tag controls for %d selections", tags, s)
							}
						}
					})
				}
			}
		}
	}
}

// Public submission includes ownership, validation and metadata preparation.
// These are intentionally distinct from the baseline's legacy cache fixtures.
func BenchmarkPerformanceSubmission(b *testing.B) {
	for _, n := range []int{100, 10000, 100000} {
		for _, kind := range []string{"Select", "InputMenu", "Listbox", "SelectMenu", "Menu", "Selection"} {
			b.Run(fmt.Sprintf("%s/N=%d", kind, n), func(b *testing.B) {
				var options []Option[string]
				var items []ListboxItem
				var commands []MenuItem
				var values []string
				switch kind {
				case "Select", "InputMenu":
					options = performanceOptions(n)
				case "Listbox", "SelectMenu":
					items = performanceItems(n)
				case "Menu":
					commands, _ = performanceMenu(n, 1)
				case "Selection":
					values = make([]string, n)
					for i := range values {
						values[i] = strconv.Itoa(i)
					}
				}
				b.ReportAllocs()
				for b.Loop() {
					switch kind {
					case "Select":
						var state SelectState[string]
						state.SetOptions(options)
					case "InputMenu":
						var state SearchState
						state.SetOptions(options)
					case "Listbox":
						var state ListboxState
						state.SetItems(items)
					case "SelectMenu":
						var state SelectMenuState
						state.SetItems(items)
					case "Menu":
						var state MenuState
						state.SetItems(commands)
					case "Selection":
						var state SelectionState[string]
						state.Set(values)
					}
				}
			})
		}
	}
}

func BenchmarkPerformanceQuery(b *testing.B) {
	for _, n := range []int{100, 10000, 100000} {
		for _, kind := range []string{"First", "Changed", "Unchanged"} {
			b.Run(fmt.Sprintf("%s/N=%d", kind, n), func(b *testing.B) {
				var data listboxItems
				data.update(performanceItems(n))
				var filter listboxFilter
				filter.apply(&data, "Item 0", nil)
				queries := [2]string{"Item 0", "Item 00"}
				query := 0
				b.ReportAllocs()
				for b.Loop() {
					switch kind {
					case "First":
						filter = listboxFilter{}
					case "Changed":
						query = 1 - query
					}
					filter.apply(&data, queries[query], nil)
				}
				b.ReportMetric(float64(filter.result.count()), "last-results")
			})
		}
	}
}

func BenchmarkPerformancePopupHeight(b *testing.B) {
	for _, n := range []int{100, 10000, 100000} {
		b.Run(fmt.Sprintf("N=%d", n), func(b *testing.B) {
			var state SelectMenuState
			state.SetItems(performanceItems(n))
			theme, height := Light(), float32(0)
			b.ReportAllocs()
			for b.Loop() {
				height = listboxContentHeight(theme, MD, &state.data, 240)
			}
			b.ReportMetric(float64(height), "height-dip")
		})
	}
}

// A held worker measures UI query dispatch independently of worker throughput.
// Every fixture explicitly uses Background, including N=100, for comparison.
func BenchmarkPerformanceSearchDispatch(b *testing.B) {
	for _, n := range []int{100, 10000, 100000} {
		b.Run(fmt.Sprintf("Background/N=%d", n), func(b *testing.B) {
			var state ListboxState
			var selected SelectionState[string]
			state.SetItems(performanceItems(n))
			state.Query = "Item"
			release, started := make(chan struct{}), make(chan struct{}, 1)
			worker := newChoiceSearchWorker()
			worker.work = func(job choiceSearchJob, index *choiceSearchIndex) choiceSearchResult {
				started <- struct{}{}
				<-release
				return runChoiceSearch(job, index)
			}
			state.filter.search.worker = worker
			defer func() {
				state.Close()
				close(release)
				waitChoiceWorker(b, worker)
			}()
			builds := 0
			tt := ui.NewTester(func(c *ui.Context) {
				builds++
				Listbox(c.Key("choices"), App(c, AppConfig{DisableAnimations: true}), &selected, &state,
					ListboxOptions{Height: 320, SearchExecution: SearchBackground}).Width(640)
			}, 700, 400)
			<-started
			tt.Frame()
			builds = 0
			queries, next := [2]string{"Item 01", "Item 02"}, 0
			b.ReportAllocs()
			for b.Loop() {
				next = 1 - next
				state.Query = queries[next]
				tt.Frame()
			}
			b.ReportMetric(float64(builds)/float64(b.N), "builds/op")
			if builds != b.N {
				b.Fatalf("%d query updates built %d views", b.N, builds)
			}
		})
	}
}

func BenchmarkPerformanceBulkSelection(b *testing.B) {
	for _, n := range []int{100, 10000, 100000} {
		b.Run(fmt.Sprintf("ListboxSelectAllClear/N=%d", n), func(b *testing.B) {
			items := performanceItems(n)
			var values SelectionState[string]
			var state ListboxState
			state.SetItems(items)
			builds := 0
			tt := ui.NewTester(func(c *ui.Context) {
				builds++
				Listbox(c.Key("choices"), App(c, AppConfig{DisableAnimations: true}), &values, &state,
					ListboxOptions{Multiple: true, Height: 320}).Width(640)
			}, 700, 400)
			tt.Key(0, ui.KeyTab)
			tt.Move(690, 390)
			builds = 0
			b.ReportAllocs()
			for b.Loop() {
				tt.Key(ui.Ctrl, ui.KeyA)
				if values.Len() != n {
					b.Fatalf("Select All selected %d of %d", values.Len(), n)
				}
				values.Clear()
				tt.Frame()
			}
			b.ReportMetric(float64(builds)/float64(b.N), "builds/op")
		})
	}
}

func BenchmarkPerformanceTree(b *testing.B) {
	for _, n := range []int{100, 10000, 100000} {
		for _, shape := range []string{"Folded", "Flat", "OpenBranches"} {
			for _, selected := range []bool{false, true} {
				b.Run(fmt.Sprintf("%s/N=%d/Selected=%v", shape, n, selected), func(b *testing.B) {
					nodes := make([]TreeNode, n)
					var state TreeState
					for i := range nodes {
						id := strconv.Itoa(i)
						nodes[i] = TreeNode{ID: id, Label: "Node " + id}
						if shape == "OpenBranches" {
							nodes[i].Children = []TreeNode{{ID: "leaf-" + id, Label: "Leaf"}}
							state.Outline.Open.Add(id)
						}
						if selected {
							state.Selection.Add(id)
						}
					}
					if shape == "Folded" {
						nodes = []TreeNode{{ID: "root", Label: "Root", Children: nodes}}
					}
					state.SetNodes(nodes)
					var builds performanceBuilds
					tt := ui.NewTester(func(c *ui.Context) {
						builds.views++
						Tree(c.Key("tree"), App(c, AppConfig{DisableAnimations: true}), &state,
							TreeOptions{Height: 320, Multiple: true, BubbleSelect: selected}).Width(640)
						builds.rows += state.RowBuilds
					}, 700, 400)
					builds.steady(b, tt)
					b.ReportMetric(float64(state.Outline.Open.Len()), "open-branches")
					b.ReportMetric(float64(state.Selection.Len()), "selected")
				})
			}
		}
	}
}

// Changing selection invalidates derived summaries even with every descendant
// folded. Keep this separate from unchanged builds and many-open-branch costs.
func BenchmarkPerformanceTreeFoldedSelection(b *testing.B) {
	for _, n := range []int{100, 10000, 100000} {
		for _, custom := range []bool{false, true} {
			b.Run(fmt.Sprintf("N=%d/Custom=%v", n, custom), func(b *testing.B) {
				children := make([]TreeNode, n)
				for i := range children {
					children[i] = TreeNode{ID: strconv.Itoa(i), Label: "Leaf"}
				}
				var state TreeState
				state.SetNodes([]TreeNode{{ID: "root", Label: "Root", Children: children}})
				o := TreeOptions{Height: 320}
				if custom {
					o.Build = func(c *ui.Context, _ Theme, item TreeItemState) { ui.Text(c.Key("text"), item.Label) }
				}
				builds := 0
				tt := ui.NewTester(func(c *ui.Context) {
					builds++
					Tree(c.Key("tree"), App(c, AppConfig{DisableAnimations: true}), &state, o).Width(640)
				}, 700, 400)
				tt.Move(690, 390)
				tt.Frame()
				builds = 0
				b.ReportAllocs()
				for b.Loop() {
					if state.Selection.Has("root") {
						state.Selection.Remove("root")
					} else {
						state.Selection.Add("root")
					}
					tt.Frame()
				}
				b.ReportMetric(float64(builds)/float64(b.N), "builds/op")
				b.ReportMetric(float64(state.RowBuilds), "rows/build")
				if builds != b.N {
					b.Fatalf("%d operations built %d views", b.N, builds)
				}
			})
		}
	}
}

func BenchmarkPerformanceTable(b *testing.B) {
	for _, columns := range []int{5, 50, 200} {
		for _, selected := range []bool{false, true} {
			for _, callback := range []bool{false, true} {
				b.Run(fmt.Sprintf("N=100000/Columns=%d/Selected=%v/Callback=%v", columns, selected, callback), func(b *testing.B) {
					rows := make([]int, 100000)
					for i := range rows {
						rows[i] = i
					}
					var state KeyedTableState[int, int]
					state.SetBorrowedRows(rows, func(v int) int { return v })
					if selected {
						for _, id := range rows {
							state.Selection.Add(id)
						}
					}
					cols := make([]TableColumn[int], columns)
					for i := range cols {
						cols[i] = TableColumn[int]{ID: strconv.Itoa(i), Title: "Column", Width: 120, Value: strconv.Itoa}
					}
					o := KeyedTableOptions[int]{Height: 320}
					if callback {
						o.OnChange = func() {}
					}
					var builds performanceBuilds
					tt := ui.NewTester(func(c *ui.Context) {
						builds.views++
						Table(c.Key("table"), App(c, AppConfig{DisableAnimations: true}), &state, cols, o).Width(640)
						builds.rows += state.CellBuilds / columns
					}, 700, 400)
					builds.steady(b, tt)
					b.ReportMetric(float64(state.CellBuilds), "cells/build")
				})
			}
		}
	}
}

func performanceMenu(n, levels int) ([]MenuItem, []string) {
	var items []MenuItem
	var path []string
	for level := levels - 1; level >= 0; level-- {
		count := n / levels
		if level == 0 {
			count += n % levels
		}
		next := make([]MenuItem, count)
		for i := range next {
			id := fmt.Sprintf("level-%d-item-%d", level, i)
			next[i] = MenuItem{ID: id, Label: id}
		}
		if items != nil {
			next[0].Children = items
			path = append([]string{next[0].ID}, path...)
		}
		items = next
	}
	return items, path
}

func BenchmarkPerformanceMenu(b *testing.B) {
	for _, n := range []int{100, 10000, 100000} {
		for _, levels := range []int{1, 3} {
			for _, open := range []bool{false, true} {
				b.Run(fmt.Sprintf("N=%d/Levels=%d/Open=%v", n, levels, open), func(b *testing.B) {
					items, path := performanceMenu(n, levels)
					state := MenuState{Open: open}
					state.SetItems(items)
					var builds performanceBuilds
					tt := ui.NewTester(func(c *ui.Context) {
						builds.views++
						th := App(c, AppConfig{DisableAnimations: true})
						anchor := ui.Button(c.Key("anchor"), "Menu")
						DropdownMenu(c.Key("menu"), th, anchor, &state,
							MenuOptions{MaxHeight: Set(float32(320)), Decoration: Decoration{Style: builds.item}})
					}, 700, 400)
					if open && len(path) > 0 {
						state.focus = nil
						state.path = path
						state.handles[path[len(path)-1]].Focus()
						tt.Frame()
						if len(state.path) != len(path) {
							b.Fatalf("opened %d of %d submenu levels", len(state.path), len(path))
						}
					}
					builds.steady(b, tt)
				})
			}
		}
	}
}

// These fixtures deliberately retain each component's nonvirtual layout.
// Their curves measure bookkeeping plus all children, not constant viewport
// rendering. Command text has its own rune/highlight curve below.
func BenchmarkPerformanceBookkeeping(b *testing.B) {
	for _, kind := range []string{"InputTags", "Accordion", "Tabs", "CheckboxGroup", "Form"} {
		for _, n := range []int{100, 1000, 10000} {
			b.Run(fmt.Sprintf("%s/N=%d", kind, n), func(b *testing.B) {
				ids := make([]string, n)
				for i := range ids {
					ids[i] = strconv.Itoa(i)
				}
				var tabs []TabItem
				var accordion []AccordionItem
				var options []Option[string]
				if kind == "Tabs" {
					tabs = make([]TabItem, n)
					for i := range tabs {
						tabs[i] = TabItem{ID: ids[i], Label: ids[i]}
					}
				}
				if kind == "Accordion" {
					accordion = make([]AccordionItem, n)
					for i := range accordion {
						accordion[i] = AccordionItem{ID: ids[i], Label: ids[i]}
					}
				}
				if kind == "CheckboxGroup" {
					options = make([]Option[string], n)
					for i := range options {
						options[i] = Option[string]{Value: ids[i], Label: ids[i]}
					}
				}
				var tags TagsState
				var form FormState
				active := ids[0]
				var builds performanceBuilds
				tt := ui.NewTester(func(c *ui.Context) {
					builds.views++
					th := App(c, AppConfig{DisableAnimations: true})
					switch kind {
					case "InputTags":
						InputTags(c, th, &ids, &tags, InputTagsOptions{})
					case "Accordion":
						Accordion(c, th, &ids, accordion, AccordionOptions{Multiple: true})
					case "Tabs":
						Tabs(c, th, &active, tabs, TabsOptions{Content: Set(false)})
					case "CheckboxGroup":
						CheckboxGroup(c, th, &ids, options, CheckboxGroupOptions{})
					case "Form":
						Form(c, th, &form, FormOptions{}, func(inner Theme) {
							for _, id := range ids {
								FormField(c.Key(id), inner, FormFieldOptions{Name: id, Value: id}, nil)
							}
						})
					}
				}, 700, 400)
				builds.steady(b, tt)
				// Each nonvirtual fixture constructs all n tags/headings/options
				// or fields. Do not report the unused Style counter as zero rows.
				b.ReportMetric(float64(n), "rows/op")
			})
		}
	}
}

func BenchmarkPerformanceTagBatch(b *testing.B) {
	for _, n := range []int{100, 1000, 10000} {
		b.Run(fmt.Sprintf("N=%d", n), func(b *testing.B) {
			var text strings.Builder
			for i := range n {
				fmt.Fprintf(&text, "%d,%d,", i, i)
			}
			batch := text.String()
			b.ReportAllocs()
			for b.Loop() {
				if got := AddTags(nil, batch, ",", 0); len(got) != n {
					b.Fatal("incorrect batch fixture")
				}
			}
		})
	}
}

func BenchmarkPerformanceCommandText(b *testing.B) {
	for _, n := range []int{100, 1000, 10000} {
		for _, highlight := range []bool{false, true} {
			b.Run(fmt.Sprintf("N=%d/Highlight=%v", n, highlight), func(b *testing.B) {
				text := strings.Repeat("命令", n/2)
				var marks []int
				if highlight {
					for i := 0; i < n; i += 2 {
						marks = append(marks, i)
					}
				}
				var builds performanceBuilds
				tt := ui.NewTester(func(c *ui.Context) {
					builds.views++
					th := App(c, AppConfig{DisableAnimations: true})
					commandText(c, th, text, marks, ItemLabelSlot, PartStyle{}, CommandPaletteOptions{}, CommandAction{})
				}, 700, 400)
				builds.steady(b, tt)
			})
		}
	}
}

func BenchmarkPerformanceUploadAdmission(b *testing.B) {
	for _, n := range []int{1000, 2000, 10000, 20000} {
		b.Run(fmt.Sprintf("Existing=%d/Incoming=%d", n, n), func(b *testing.B) {
			existing := uploadFixtureFiles(n)
			incoming := uploadFixtureFiles(n)
			for i := range incoming {
				if i%5 != 0 {
					incoming[i].Path += "-new"
				}
				if i%7 == 0 {
					incoming[i].Error = "Rejected sample"
				}
			}
			b.ReportAllocs()
			for b.Loop() {
				s := FileUploadState{Files: make([]UploadFile, n, 2*n)}
				copy(s.Files, existing)
				s.Add(incoming, UploadRules{MaxFiles: n + n/2})
				if len(s.Files) != n+n*4/5 {
					b.Fatal("batch fixture did not admit its unique paths")
				}
			}
		})
	}
}

func BenchmarkPerformanceUploadPreviews(b *testing.B) {
	for _, layout := range []UploadLayout{UploadList, UploadGrid} {
		for _, n := range []int{100, 10000, 100000} {
			b.Run(fmt.Sprintf("Layout=%d/N=%d", layout, n), func(b *testing.B) {
				s := FileUploadState{Files: uploadFixtureFiles(n)}
				var builds performanceBuilds
				tt := ui.NewTester(func(c *ui.Context) {
					builds.views++
					FileUpload(c.Key("files"), App(c, AppConfig{DisableAnimations: true}), &s, FileUploadOptions{Variant: UploadButton, Layout: layout}).Width(500)
				}, 700, 400)
				tt.Frame()
				controls := 0
				for _, label := range tt.Texts() {
					if strings.HasPrefix(label, "Remove File ") {
						controls++
					}
				}
				if controls == 0 || controls > 24 {
					b.Fatalf("built %d preview actions", controls)
				}
				builds.steady(b, tt)
				b.ReportMetric(float64(controls), "rows/op")
			})
		}
	}
}

// Unlike an unmounted SetRows fixture, this binds the wrapper-owned native
// cursor and moves it to the final record before measuring unchanged-ID updates.
func BenchmarkPerformanceTableCursor(b *testing.B) {
	for _, n := range []int{100, 10000, 100000} {
		for _, work := range []string{"Restore", "OwnedRebind", "BorrowedRebind"} {
			b.Run(fmt.Sprintf("%s/N=%d", work, n), func(b *testing.B) {
				rows := make([]int, n)
				for i := range rows {
					rows[i] = i
				}
				key := func(v int) int { return v }
				var state KeyedTableState[int, int]
				if work == "BorrowedRebind" {
					state.SetBorrowedRows(rows, key)
				} else {
					state.SetRows(rows, key)
				}
				tt := ui.NewTester(func(c *ui.Context) {
					Table(c.Key("table"), App(c, AppConfig{DisableAnimations: true}), &state,
						[]TableColumn[int]{{ID: "value", Title: "Value", Value: strconv.Itoa}},
						KeyedTableOptions[int]{OnActivate: func(int) {}}).Width(300)
				}, 400, 350)
				tt.Key(0, ui.KeyTab)
				tt.Key(0, ui.KeyEnd)
				if id, ok := state.activeRowID(); !ok || id != n-1 {
					b.Fatalf("cursor is not bound to the final record: %d, %v", id, ok)
				}
				b.ReportAllocs()
				for b.Loop() {
					switch work {
					case "Restore":
						state.restoreActiveRow(n-1, true)
					case "OwnedRebind":
						state.SetRows(rows, key)
					case "BorrowedRebind":
						state.SetBorrowedRows(rows, key)
					}
				}
				runtime.KeepAlive(tt)
			})
		}
	}
}
