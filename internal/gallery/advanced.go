package gallery

import (
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

type advancedState struct {
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
