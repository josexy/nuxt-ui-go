package gallery

import (
	"github.com/josexy/nuxt-ui-go/icons/lucide"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

type advancedState struct {
	ChartAxisOptions         nui.SelectState[string]
	ChartCurveOptions        nui.SelectState[nui.ChartCurve]
	ChartStackOptions        nui.SelectState[nui.ChartStackMode]
	Chart                    nui.ChartState
	ChartKind                nui.ChartKind
	ChartCurve               nui.ChartCurve
	ChartStack               nui.ChartStackMode
	ChartAxis                string
	ChartRevision            int
	ChartNegative, ChartGaps bool
}

func newAdvanced() advancedState {
	s := advancedState{ChartAxis: "category", ChartCurve: nui.ChartCurveMonotone}
	s.ChartAxisOptions.SetOptions([]nui.Option[string]{{Value: "numeric", Label: "Numeric"}, {Value: "category", Label: "Categories"}, {Value: "time", Label: "Time"}})
	s.ChartCurveOptions.SetOptions([]nui.Option[nui.ChartCurve]{
		{Value: nui.ChartCurveLinear, Label: "Linear"}, {Value: nui.ChartCurveMonotone, Label: "Monotone"},
		{Value: nui.ChartCurveStepAfter, Label: "Step after"}, {Value: nui.ChartCurveStepBefore, Label: "Step before"}, {Value: nui.ChartCurveStep, Label: "Step middle"},
	})
	s.ChartStackOptions.SetOptions([]nui.Option[nui.ChartStackMode]{
		{Value: nui.ChartStackNone, Label: "None"}, {Value: nui.ChartStackStandard, Label: "Standard"},
		{Value: nui.ChartStackNormalized, Label: "Normalized"}, {Value: nui.ChartStackCenter, Label: "Centered"},
	})
	s.updateChartData()
	return s
}

var shellNavigation = []nui.NavigationItem{
	{ID: "activity", Label: "Activity", Group: "Workspace", Icon: lucide.Search()},
	{ID: "attachments", Label: "Attachments", Icon: lucide.FolderOpen()},
	{ID: "settings", Label: "Workspace settings", Group: "Manage", Icon: lucide.Settings(), Children: []nui.NavigationItem{{ID: "members", Label: "Members"}, {ID: "permissions", Label: "Access permissions and sharing"}}},
}

// Dispose ends window-owned work and is safe to repeat.
func (a *State) Dispose() {
	if a.disposed {
		return
	}
	a.Deactivate()
	for _, p := range a.pages {
		p.stop()
	}
	a.disposed = true
}
