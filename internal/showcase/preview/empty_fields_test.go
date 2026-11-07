package preview

import (
	"testing"

	nui "github.com/josexy/nuxt-ui-go"
)

func TestEmptyPreviewBuildersAvatarAndActions(t *testing.T) {
	tester, cfg := exampleTester(t, "empty")
	clickExample(t, tester, "Create project")
	if !tester.HasText("Created: 1") {
		t.Fatal("default action did not update the footer")
	}
	cfg.Values["customHeader"], cfg.Values["customBody"] = true, true
	tester.Frame()
	if !tester.HasText("Custom header content") || tester.HasText("No projects yet") || tester.HasText("Create project") {
		t.Fatal("custom header/body controls did not replace default content")
	}
	clickExample(t, tester, "Complete custom body")
	if !tester.HasText("Created: 2") {
		t.Fatal("custom body action lost shared state")
	}
	cfg.Values["customHeader"], cfg.Values["customBody"] = false, false
	cfg.Values["avatar"], cfg.Values["size"] = true, "xl"
	tester.Frame()
	if r, ok := tester.Find("Empty preview avatar"); !ok || r.W != 48 || r.H != 48 || !tester.HasText("NU") {
		t.Fatalf("avatar control or size ignored: %+v", r)
	}
	cfg.Values["loading"], cfg.Values["loadingColor"] = true, true
	tester.Frame()
	if tester.HasText("NU") || !tester.HasText("Created: 2") {
		t.Fatal("loading priority or preserved footer state")
	}
	want := nui.Light().Error
	pixels, count := tester.Image().Pix, 0
	for i := 0; i < len(pixels); i += 4 {
		if pixels[i] == want.R && pixels[i+1] == want.G && pixels[i+2] == want.B && pixels[i+3] == 255 {
			count++
		}
	}
	if count == 0 {
		t.Fatal("loading color control did not change the glyph")
	}
}
