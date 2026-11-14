package gallery

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

// The Gallery nests Main, Header and Footer previews inside its own page, so
// each window must still expose at most one of each landmark.
func TestGalleryKeepsOneLandmarkOfEachKind(t *testing.T) {
	a := newStaticGallery()
	t.Cleanup(a.Dispose)
	tt := ui.NewTester(a.View, 1080, 820)
	for _, p := range componentDirectory {
		a.Navigate(p.ID)
		tt.Frame()
		counts := map[ui.Role]int{}
		for _, info := range tt.AccessAll() {
			switch info.Role {
			case ui.RoleMain, ui.RoleBanner, ui.RoleContentInfo:
				counts[info.Role]++
			}
		}
		if counts[ui.RoleMain] > 1 || counts[ui.RoleBanner] > 1 || counts[ui.RoleContentInfo] > 1 {
			t.Errorf("%s exposes main=%d banner=%d contentinfo=%d", p.ID, counts[ui.RoleMain], counts[ui.RoleBanner], counts[ui.RoleContentInfo])
		}
		if p.ID == "example/workspace" && counts[ui.RoleMain] != 1 {
			t.Errorf("the workspace page lost its main landmark: %d", counts[ui.RoleMain])
		}
	}
}
