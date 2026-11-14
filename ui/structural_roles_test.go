package nui_test

import (
	"os"
	"os/exec"
	"slices"
	"testing"
	"testing/fstest"

	"github.com/egoist/mygo/i18n"
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func inside(inner, outer ui.Rect) bool {
	return inner.W > 0 && inner.H > 0 && inner.X >= outer.X && inner.Y >= outer.Y && inner.X+inner.W <= outer.X+outer.W && inner.Y+inner.H <= outer.Y+outer.H
}

func TestShellComponentsAreLandmarks(t *testing.T) {
	optOut := false
	var header, main, footer ui.Handle
	var bounds [3]ui.Rect
	r := newReaderTester(t, func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Fill().Children(func() {
			h := nui.HeaderLayout(c.Key("header"), th, nui.HeaderOptions{Title: "Project", Description: "Overview"}).Bind(&header)
			m := nui.Main(c.Key("main"), th, func() { ui.Text(c, "Quarterly report") }).Bind(&main)
			f := nui.FooterLayout(c.Key("footer"), th, nui.FooterOptions{Text: "All changes saved"}).Bind(&footer)
			if optOut {
				h.Role(ui.RoleGroup)
				m.Role(ui.RoleGroup)
				f.Role(ui.RoleGroup)
			}
			bounds = [3]ui.Rect{header.Bounds(c), main.Bounds(c), footer.Bounds(c)}
		})
	}, 800, 600)
	r.Frame()
	before := bounds

	for _, role := range []ui.Role{ui.RoleBanner, ui.RoleMain, ui.RoleContentInfo} {
		if got := r.withRole(role); len(got) != 1 {
			t.Fatalf("%s landmarks: %v", roleName(role), accessSummary(r.AccessAll()))
		}
	}
	banner := r.role(ui.RoleBanner, "")
	heading := r.role(ui.RoleHeading, "Project")
	if heading.Level != 1 || !inside(heading.Bounds, banner.Bounds) {
		t.Fatalf("title %+v is not a level-1 heading inside the banner %+v", heading, banner.Bounds)
	}
	mainInfo := r.role(ui.RoleMain, "")
	body := r.role(ui.RoleText, "Quarterly report")
	if !inside(body.Bounds, mainInfo.Bounds) {
		t.Fatalf("page content %v is outside the main landmark %v", body.Bounds, mainInfo.Bounds)
	}
	info := r.role(ui.RoleContentInfo, "")
	saved := r.role(ui.RoleText, "All changes saved")
	if !inside(saved.Bounds, info.Bounds) {
		t.Fatalf("footer text %v is outside the content-info landmark %v", saved.Bounds, info.Bounds)
	}

	optOut = true
	r.Frame()
	r.Frame()
	for _, role := range []ui.Role{ui.RoleBanner, ui.RoleMain, ui.RoleContentInfo} {
		if got := r.withRole(role); len(got) != 0 {
			t.Fatalf("opted-out %s landmark is still reported: %v", roleName(role), accessSummary(got))
		}
	}
	if bounds != before {
		t.Fatalf("opting out moved the shell: before %v after %v", before, bounds)
	}
	if heading := r.role(ui.RoleHeading, "Project"); heading.Level != 1 {
		t.Fatalf("opted-out header lost its heading: %+v", heading)
	}
	r.expectSilent("landmark changes")
}

func TestTabPanelsAreNamedAndControlledByTheirTabs(t *testing.T) {
	value, keep, content := "account", false, true
	withContent := true
	r := newReaderTester(t, func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		items := []nui.TabItem{{ID: "account", Label: "Account"}, {ID: "password", Label: "Password"}}
		if withContent {
			items[0].Content = func(c *ui.Context, t nui.Theme) { ui.Text(c, "Account settings") }
			items[1].Content = func(c *ui.Context, t nui.Theme) { ui.Text(c, "Password settings") }
		}
		nui.Tabs(c.Key("tabs"), th, &value, items, nui.TabsOptions{Label: "Settings", KeepMounted: keep, Content: nui.Set(content)})
	}, 600, 400)

	panel := r.role(ui.RoleTabPanel, "Account")
	if !inside(r.role(ui.RoleText, "Account settings").Bounds, panel.Bounds) {
		t.Fatalf("panel %+v does not hold its content", panel)
	}
	if tab := r.role(ui.RoleTab, "Account"); !slices.Equal(tab.Controls, []string{"Account"}) || !tab.Selected {
		t.Fatalf("Account tab %+v does not control its panel", tab)
	}
	if tab := r.role(ui.RoleTab, "Password"); len(tab.Controls) != 0 {
		t.Fatalf("Password tab controls %q while its panel is not built", tab.Controls)
	}
	if got := r.withRole(ui.RoleTabPanel); len(got) != 1 {
		t.Fatalf("tab panels %v", accessSummary(got))
	}

	keep = true
	r.Frame()
	if got := r.withRole(ui.RoleTabPanel); len(got) != 1 || got[0].Name != "Account" {
		t.Fatalf("a hidden kept-mounted panel is reported: %v", accessSummary(got))
	}
	if _, ok := r.AccessByRole(ui.RoleText, "Password settings"); ok {
		t.Fatal("the hidden panel's content is reported")
	}
	r.Key(0, ui.KeyTab)
	r.Key(0, ui.KeyRight)
	if value != "password" {
		t.Fatalf("arrow key selected %q", value)
	}
	if got := r.withRole(ui.RoleTabPanel); len(got) != 1 || got[0].Name != "Password" {
		t.Fatalf("panels after switching: %v", accessSummary(got))
	}
	if tab := r.role(ui.RoleTab, "Password"); !slices.Equal(tab.Controls, []string{"Password"}) {
		t.Fatalf("Password tab %+v does not control its panel", tab)
	}

	for _, step := range []struct {
		name               string
		withContent, shown bool
	}{{"Content=false", true, false}, {"items without content", false, true}} {
		withContent, content = step.withContent, step.shown
		r.Frame()
		if got := r.withRole(ui.RoleTabPanel); len(got) != 0 {
			t.Fatalf("%s: panels %v", step.name, accessSummary(got))
		}
		for _, name := range []string{"Account", "Password"} {
			if tab := r.role(ui.RoleTab, name); len(tab.Controls) != 0 {
				t.Fatalf("%s: %s tab controls %q", step.name, name, tab.Controls)
			}
		}
	}
	r.expectSilent("tab panel changes")
}

func TestTabPanelKeepsAnOverriddenLegacyName(t *testing.T) {
	if os.Getenv("NUI_TAB_PANEL_OVERRIDE_CHILD") != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestTabPanelKeepsAnOverriddenLegacyName$", "-test.count=1")
		cmd.Env = append(os.Environ(), "NUI_TAB_PANEL_OVERRIDE_CHILD=1", "CGO_ENABLED=0")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("child: %v\n%s", err, out)
		}
		return
	}
	app, err := i18n.NewCatalog(fstest.MapFS{"en.json": {Data: []byte(`{"@@locale": "en", "nuxtui.contentNamed": "{name} pane"}`)}}, "en")
	if err != nil {
		t.Fatal(err)
	}
	i18n.Register(app)
	value := "account"
	r := newReaderTester(t, func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Tabs(c, th, &value, []nui.TabItem{{ID: "account", Label: "Account", Content: func(c *ui.Context, t nui.Theme) { ui.Text(c, "Account settings") }}}, nui.TabsOptions{})
	}, 600, 400)
	r.SetLocale("en")
	r.role(ui.RoleTabPanel, "Account pane")
	if tab := r.role(ui.RoleTab, "Account"); !slices.Equal(tab.Controls, []string{"Account pane"}) {
		t.Fatalf("tab %+v does not control the renamed panel", tab)
	}
}
