package gallery

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func TestChartsGalleryControls(t *testing.T) {
	a := New()
	defer a.Dispose()
	if !a.Animations {
		t.Fatal("normal Gallery must demonstrate motion")
	}
	a.Advanced = newAdvanced()
	tt := ui.NewTester(func(c *ui.Context) {
		theme := nui.App(c, a.themeConfig())
		ui.Scroll(c).Fill().Padding(20).Gap(12).Children(func() { a.chartDemo(c, theme) })
	}, 1080, 820)
	choose := func(control, value string) {
		t.Helper()
		press(t, tt, control)
		press(t, tt, value)
	}
	press(t, tt, "Update data")
	tt.Advance(80 * time.Millisecond)
	if a.Advanced.ChartRevision != 1 || !tt.FramePending() {
		t.Fatal("update did not animate")
	}
	press(t, tt, "Area")
	choose("Chart stacking", "Normalized")
	choose("Chart curve", "Step before")
	choose("Chart X axis", "Time")
	if a.Advanced.ChartStack != nui.ChartStackNormalized || a.Advanced.ChartCurve != nui.ChartCurveStepBefore || a.Advanced.ChartAxis != "time" {
		t.Fatalf("chart option bindings: stack=%v curve=%v axis=%v", a.Advanced.ChartStack, a.Advanced.ChartCurve, a.Advanced.ChartAxis)
	}
	press(t, tt, "Negative values")
	press(t, tt, "Data gaps")
	press(t, tt, "Builds")
	if !a.Advanced.ChartNegative || !a.Advanced.ChartGaps || a.Advanced.Chart.SeriesVisible("build") {
		t.Fatal("data or legend bindings")
	}
	press(t, tt, "Update data")
	if a.Advanced.Chart.SeriesVisible("build") {
		t.Fatal("update reset visibility")
	}
	press(t, tt, "Donut")
	press(t, tt, "Passed")
	if a.Advanced.Chart.SliceVisible("passed") {
		t.Fatal("slice legend")
	}
	press(t, tt, "Reset data")
	if a.Advanced.ChartRevision != 0 || a.Advanced.ChartNegative || a.Advanced.ChartGaps || !a.Advanced.Chart.SliceVisible("passed") || !a.Advanced.Chart.SeriesVisible("build") {
		t.Fatal("reset")
	}
	press(t, tt, "Animate charts")
	press(t, tt, "Update data")
	tt.Advance(time.Second)
	if a.Animations || tt.FramePending() {
		t.Fatal("motion toggle")
	}
	// Removing and returning to both Cartesian and radial controls preserves identity.
	for _, kind := range []string{"Scatter", "Pie", "Line", "Bars", "Donut"} {
		press(t, tt, kind)
		tt.Frame()
		tt.Frame()
	}
}

func TestChartGalleryVisuals(t *testing.T) {
	dir := os.Getenv("NUI_CHART_ARTIFACTS")
	if dir == "" {
		t.Skip("NUI_CHART_ARTIFACTS not set")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	scenes := []struct {
		name           string
		kind           nui.ChartKind
		curve          nui.ChartCurve
		stack          nui.ChartStackMode
		axis           string
		negative, gaps bool
	}{
		{"line", nui.ChartLine, nui.ChartCurveMonotone, nui.ChartStackNone, "category", false, false},
		{"step-gaps", nui.ChartLine, nui.ChartCurveStep, nui.ChartStackNone, "numeric", true, true},
		{"bars", nui.ChartBar, nui.ChartCurveLinear, nui.ChartStackNormalized, "category", false, false},
		{"area-time", nui.ChartArea, nui.ChartCurveMonotone, nui.ChartStackStandard, "time", false, true},
		{"area-center", nui.ChartArea, nui.ChartCurveLinear, nui.ChartStackCenter, "category", false, false},
		{"scatter", nui.ChartScatter, nui.ChartCurveLinear, nui.ChartStackNone, "numeric", true, false},
		{"pie", nui.ChartPie, nui.ChartCurveLinear, nui.ChartStackNone, "category", false, false},
		{"donut", nui.ChartDonut, nui.ChartCurveLinear, nui.ChartStackNone, "category", false, false},
	}
	for _, dark := range []bool{false, true} {
		for _, scene := range scenes {
			a := newStaticGallery()
			a.Advanced, a.Dark = newAdvanced(), dark
			a.Advanced.ChartKind, a.Advanced.ChartCurve, a.Advanced.ChartStack, a.Advanced.ChartAxis = scene.kind, scene.curve, scene.stack, scene.axis
			a.Advanced.ChartNegative, a.Advanced.ChartGaps = scene.negative, scene.gaps
			a.Advanced.updateChartData()
			tt := ui.NewTester(func(c *ui.Context) {
				theme := nui.App(c, a.themeConfig())
				ui.Scroll(c).Fill().Padding(20).Gap(12).Children(func() { a.chartDemo(c, theme) })
			}, 1080, 920)
			tt.Move(0, 0)
			path := filepath.Join(dir, fmt.Sprintf("%s-dark-%t.png", scene.name, dark))
			f, err := os.Create(path)
			if err != nil {
				t.Fatal(err)
			}
			err = png.Encode(f, tt.Image())
			closeErr := f.Close()
			if err != nil {
				t.Fatal(err)
			}
			if closeErr != nil {
				t.Fatal(closeErr)
			}
			a.Dispose()
		}
	}
}
