package nui_test

import (
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
)

// readerTester is a ui.Tester with a screen reader attached from its first
// frame, for tests of what assistive technology reads and sees.
type readerTester struct {
	*ui.Tester
	t *testing.T
}

// newReaderTester starts view with a screen reader attached and forgets
// what the first frame showed, as a real screen reader attaching to an open
// window reads nothing of it.
func newReaderTester(t *testing.T, view func(*ui.Context), width, height int) *readerTester {
	t.Helper()
	tt := ui.NewTester(view, width, height)
	tt.UseScreenReader()
	tt.Frame()
	r := &readerTester{Tester: tt, t: t}
	r.Announcements()
	r.Alerts()
	return r
}

// expectRead checks that the polite announcements since the last check are
// exactly want, and that nothing was read assertively.
func (r *readerTester) expectRead(step string, want ...string) {
	r.t.Helper()
	got, alerts := r.Announcements(), r.Alerts()
	if !slices.Equal(got, want) || len(alerts) != 0 {
		r.t.Fatalf("%s: read %q and alerted %q, want %q", step, got, alerts, want)
	}
}

// expectSilent checks that nothing was read since the last check.
func (r *readerTester) expectSilent(step string) {
	r.t.Helper()
	r.expectRead(step)
}

// role returns the first element with role and accessible name, failing the
// test when there is none.
func (r *readerTester) role(role ui.Role, name string) ui.AccessInfo {
	r.t.Helper()
	info, ok := r.AccessByRole(role, name)
	if !ok {
		r.t.Fatalf("no element with role %d named %q among %+v", role, name, accessSummary(r.AccessAll()))
	}
	return info
}

// withRole returns every element of the last frame with role, in reading
// order.
func (r *readerTester) withRole(role ui.Role) []ui.AccessInfo {
	var out []ui.AccessInfo
	for _, info := range r.AccessAll() {
		if info.Role == role {
			out = append(out, info)
		}
	}
	return out
}

// accessSummary is a short role/name list for failure messages.
func accessSummary(all []ui.AccessInfo) []string {
	out := make([]string, 0, len(all))
	for _, info := range all {
		out = append(out, roleName(info.Role)+":"+info.Name)
	}
	return out
}

func roleName(role ui.Role) string {
	switch role {
	case ui.RoleStatus:
		return "status"
	case ui.RoleAlert:
		return "alert"
	case ui.RoleMain:
		return "main"
	case ui.RoleBanner:
		return "banner"
	case ui.RoleContentInfo:
		return "contentinfo"
	case ui.RoleTabPanel:
		return "tabpanel"
	case ui.RoleTab:
		return "tab"
	case ui.RoleGrid:
		return "grid"
	case ui.RoleTable:
		return "table"
	case ui.RoleCell:
		return "cell"
	case ui.RoleGroup:
		return "group"
	case ui.RoleText:
		return "text"
	case ui.RoleButton:
		return "button"
	}
	return "role"
}

func TestReaderTesterReadsAKeyedStatusOnce(t *testing.T) {
	text, before := "Draft", false
	r := newReaderTester(t, func(c *ui.Context) {
		ui.Column(c).Children(func() {
			if before {
				ui.Text(c, "Sibling")
			}
			ui.Text(c.Key("status"), text).Role(ui.RoleStatus)
		})
	}, 300, 200)
	if info := r.role(ui.RoleStatus, "Draft"); info.Name != "Draft" {
		t.Fatalf("status %+v", info)
	}
	r.Frame()
	r.expectSilent("unchanged rebuild")
	text = "Saved"
	r.Frame()
	r.expectRead("changed status", "Saved")
	r.Frame()
	r.Frame()
	r.expectSilent("rebuilds after the change")
	before = true
	r.Frame()
	r.expectSilent("sibling inserted before the keyed status")
	if got := r.withRole(ui.RoleStatus); len(got) != 1 {
		t.Fatalf("statuses %v", accessSummary(got))
	}
}
