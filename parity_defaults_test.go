package nui_test

import (
	"testing"

	"github.com/egoist/mygo/i18n"
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

// Default values aligned with Nuxt UI v4.12 (fix plan section 1, D1–D15).

func TestStepperIsLinearByDefault(t *testing.T) {
	for _, tc := range []struct {
		linear nui.Optional[bool]
		jumps  bool
	}{{nui.Optional[bool]{}, false}, {nui.Set(true), false}, {nui.Set(false), true}} {
		step := "one"
		tt := ui.NewTester(func(c *ui.Context) {
			th := nui.App(c, nui.AppConfig{})
			nui.Stepper(c, th, &step, []nui.StepItem{{ID: "one", Title: "One"}, {ID: "two", Title: "Two"}, {ID: "three", Title: "Three"}}, nui.StepperOptions{Linear: tc.linear})
		}, 600, 200)
		click(t, tt, "Three")
		if (step == "three") != tc.jumps {
			t.Fatalf("Linear=%+v: clicking the third step selected %q", tc.linear, step)
		}
		click(t, tt, "Two")
		if step != "two" {
			t.Fatalf("Linear=%+v: the next step is always reachable, got %q", tc.linear, step)
		}
	}
}

func TestInputTimeFollowsLocaleHourCycle(t *testing.T) {
	clock := nui.Set(nui.ClockTime{Hour: 13, Minute: 5})
	var s nui.TimeInputState
	var twelve nui.Optional[bool]
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.InputTime(c, th, &clock, &s, nui.InputTimeOptions{InputOptions: nui.InputOptions{Label: "Time"}, TwelveHour: twelve})
	}, 400, 120)
	tt.SetLocale("en")
	if s.Text != "01:05 PM" {
		t.Fatalf("en: %q, want 12-hour", s.Text)
	}
	tt.SetLocale("zh-Hans")
	if s.Text != "13:05" {
		t.Fatalf("zh-Hans: %q, want 24-hour", s.Text)
	}
	// The user's regional formats win over the language's default.
	tt.SetRegion(i18n.Region{Locale: "en-US", HourCycle: 12})
	tt.Frame()
	if s.Text != "01:05 PM" {
		t.Fatalf("zh-Hans with 12-hour regional formats: %q", s.Text)
	}
	twelve = nui.Set(false)
	tt.Frame()
	if s.Text != "13:05" {
		t.Fatalf("explicit 24-hour: %q", s.Text)
	}
}
