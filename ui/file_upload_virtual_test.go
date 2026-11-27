package nui

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
)

func uploadFixtureFiles(n int) []UploadFile {
	files := make([]UploadFile, n)
	for i := range files {
		files[i] = UploadFile{ID: fmt.Sprint(i), Name: fmt.Sprintf("File %06d", i), Path: fmt.Sprintf("/files/%d", i), Size: 1024}
	}
	return files
}

func TestFileUploadBatchRespectsCurrentCollection(t *testing.T) {
	s := FileUploadState{Files: []UploadFile{{ID: "file-1", Path: "old", Status: UploadFailed}, {ID: "file-3", Path: "rejected", Status: UploadRejected}}}
	rules := UploadRules{Accept: []string{".txt"}, MaxFiles: 2}
	s.Add([]UploadFile{{Path: "old", Name: "ignored.txt"}, {Path: "bad", Name: "bad.png"}, {Path: "bad", Name: "duplicate.txt"}, {Path: "good", Name: "good.txt"}, {Path: "extra", Name: "extra.txt"}, {Name: "pathless.txt"}}, rules)
	if len(s.Files) != 6 || s.Files[2].ID != "file-2" || s.Files[3].ID != "file-4" || s.Files[3].Status != UploadReady || s.Files[4].Status != UploadRejected || s.Files[5].Status != UploadRejected {
		t.Fatalf("batch admission: %+v", s.Files)
	}
	s.Files = []UploadFile{{ID: "file-9", Path: "replacement", Status: UploadRejected}}
	s.Add([]UploadFile{{Path: "good", Name: "good.txt"}, {Path: "replacement", Name: "ignored.txt"}, {Name: "one.txt"}, {Name: "two.txt"}}, rules)
	if len(s.Files) != 4 || s.Files[1].Status != UploadReady || s.Files[2].Status != UploadReady || s.Files[3].Status != UploadRejected {
		t.Fatalf("external mutation or pathless admission: %+v", s.Files)
	}
	ctx, ticket := s.Begin(s.Files[1].ID)
	s.Add([]UploadFile{{Name: "later.txt"}}, rules)
	if !s.Report(ticket, .4, false, nil) || ctx.Err() != nil {
		t.Fatal("batch admission changed an existing job identity")
	}
	s.Remove(s.Files[1].ID)
	if ctx.Err() == nil || s.Report(ticket, 1, true, nil) {
		t.Fatal("removed job accepted completion")
	}
}

func TestFileUploadEmptyBatchKeepsCancellationCheap(t *testing.T) {
	s := FileUploadState{Files: uploadFixtureFiles(10000)}
	if allocs := testing.AllocsPerRun(10, func() { s.Add(nil, UploadRules{}) }); allocs != 0 {
		t.Fatalf("empty picker result allocated %g times", allocs)
	}
	if len(s.Files) != 10000 {
		t.Fatal("empty picker result changed the queue")
	}
}

func TestFileUploadVirtualViewports(t *testing.T) {
	const n = 10000
	for _, layout := range []UploadLayout{UploadList, UploadGrid} {
		t.Run(fmt.Sprintf("Layout=%d/N=%d", layout, n), func(t *testing.T) {
			s := FileUploadState{Files: uploadFixtureFiles(n)}
			height, preview := Optional[float32]{}, true
			changes, removed := 0, ""
			tt := ui.NewTester(func(c *ui.Context) {
				FileUpload(c.Key("files"), App(c, AppConfig{DisableAnimations: true}), &s, FileUploadOptions{Label: "Files", Variant: UploadButton, Layout: layout, PreviewHeight: height, Preview: Set(preview), OnChange: func([]UploadFile) { changes++ }, OnRemove: func(id string) { removed = id }}).Width(500)
			}, 550, 500)
			tt.Frame()
			bounded := func() {
				t.Helper()
				count := 0
				for _, label := range tt.Texts() {
					if strings.HasPrefix(label, "Remove File ") {
						count++
					}
				}
				if count < 1 || count > 24 {
					t.Fatalf("constructed %d removal controls for %d files", count, n)
				}
				box, ok := tt.Find("File previews")
				if !ok || box.H != valueOr(height, 320) {
					t.Fatalf("viewport=%v present=%v", box, ok)
				}
			}
			bounded()
			if layout == UploadGrid {
				s.previews.grid.ScrollTo(n-1, ui.End)
			} else {
				s.previews.list.ScrollTo(n-1, ui.End)
			}
			tt.Frame()
			bounded()
			ctx, ticket := s.Begin(s.Files[n-1].ID)
			if !s.Report(ticket, .5, false, errors.New("full failure detail")) || ctx.Err() == nil {
				t.Fatal("upload failure was not published")
			}
			tt.Frame()
			if !tt.HasText("full failure detail") {
				t.Fatal("scrolled file lost its current error")
			}
			label := "Remove " + s.Files[n-1].Name
			if err := tt.Click(label); err != nil {
				t.Fatal(err)
			}
			if removed != fmt.Sprint(n-1) || len(s.Files) != n-1 || changes != 1 || !tt.Focused("Choose files") {
				t.Fatal("visible removal lost stable identity, callback count, or picker focus")
			}
			height = Set(float32(160))
			tt.Frame()
			bounded()
			height = Set(float32(0))
			tt.Frame()
			if tt.HasText("File previews") || tt.HasText("Remove File 000000") {
				t.Fatal("zero height built previews")
			}
			height, preview = Set(float32(250)), false
			tt.Frame()
			if tt.HasText("File previews") || len(s.Files) != n-1 {
				t.Fatal("Preview=false changed data or built a viewport")
			}
		})
	}
}

func TestFileUploadPreviewHeightRejectsInvalidValues(t *testing.T) {
	for _, height := range []float32{-1, float32(math.NaN()), float32(math.Inf(1)), float32(math.Inf(-1))} {
		t.Run(fmt.Sprint(height), func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid preview height accepted")
				}
			}()
			var state FileUploadState
			ui.NewTester(func(c *ui.Context) {
				FileUpload(c, App(c, AppConfig{}), &state, FileUploadOptions{PreviewHeight: Set(height), Preview: Set(false)})
			}, 400, 400)
		})
	}
}

func TestFileUploadLayoutSwitchKeepsActionFocus(t *testing.T) {
	s := FileUploadState{Files: uploadFixtureFiles(1000)}
	layout, custom := UploadList, false
	tt := ui.NewTester(func(c *ui.Context) {
		o := FileUploadOptions{Variant: UploadButton, Layout: layout}
		if custom {
			o.UI = Styles{Item: {PaddingY: Set(float32(30))}}
		}
		FileUpload(c.Key("files"), App(c, AppConfig{DisableAnimations: true}), &s, o).Width(500)
	}, 550, 500)
	s.previews.list.ScrollTo(500, ui.Start)
	tt.Frame()
	for range 24 {
		if tt.Focused("Remove File 000500") {
			break
		}
		tt.Key(0, ui.KeyTab)
	}
	if !tt.Focused("Remove File 000500") {
		t.Fatal("scrolled file action was not reachable by Tab")
	}
	layout = UploadGrid
	tt.Frame()
	// GridView commits its initial column measurement and late-built handles
	// on the following frame, as when a native window resizes the grid.
	tt.Frame()
	if !tt.Focused("Remove File 000500") || !s.previews.gridMode {
		t.Fatalf("list to grid lost focused file action: pending=%q index=%d active=%+v texts=%v", s.previews.pending, s.previews.index, s.previews.active, tt.Texts())
	}
	custom = true
	tt.Frame()
	tt.Frame()
	if !tt.Focused("Remove File 000500") || s.previews.gridMode {
		t.Fatal("custom grid sizing did not use measured rows or lost focus")
	}
}

func TestFileUploadGridPreservesFullErrorAndTextScaling(t *testing.T) {
	for _, scale := range []float32{1, 1.5, 2} {
		t.Run(fmt.Sprint(scale), func(t *testing.T) {
			full := strings.Repeat("A complete error with more detail. ", 20)
			s := FileUploadState{Files: []UploadFile{{ID: "one", Name: strings.Repeat("Long filename ", 8), Error: full, Status: UploadFailed}}}
			tt := ui.NewTester(func(c *ui.Context) {
				th := App(c, AppConfig{DisableAnimations: true})
				th.TextScale = scale
				FileUpload(c, th, &s, FileUploadOptions{Variant: UploadButton, Layout: UploadGrid, PreviewHeight: Set(float32(600))}).Width(500)
			}, 550, 800)
			tt.Frame()
			info, ok := tt.AccessByRole(ui.RoleGroup, s.Files[0].Name)
			if !ok || info.Description != full {
				t.Fatal("grid lost the full accessible error or filename")
			}
			errorInfo, ok := tt.AccessByRole(ui.RoleText, full)
			if !ok || errorInfo.Description != full {
				t.Fatal("bounded error text lost the complete description")
			}
			if box, ok := tt.Find(full); !ok || box.H > 40*scale+.01 {
				t.Fatalf("error excerpt is not bounded to two lines: %v", box)
			}
			if !s.previews.gridMode {
				t.Fatal("default grid unexpectedly used the list fallback")
			}
		})
	}
}
