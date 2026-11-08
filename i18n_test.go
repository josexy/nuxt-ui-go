package nui_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

func TestLocaleSwitchPreservesInputFocusAndSelection(t *testing.T) {
	value := "中文 draft"
	changes := 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Input(c, th, &value, nui.InputOptions{Label: "Name", Clearable: true, OnChange: func(string) { changes++ }})
	}, 400, 120)
	click(t, tt, "Name")
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.SetLocale("zh-Hans")
	if !tt.Focused("Name") || !tt.HasText("清空Name") || tt.HasText("Clear Name") {
		t.Fatal("input lost focus or retained its English clear action")
	}
	tt.Command("copy")
	if tt.Clipboard() != value || value != "中文 draft" || changes != 0 {
		t.Fatal("language change modified the value, selection or callbacks")
	}
	tt.SetLocale("en")
	click(t, tt, "Clear Name")
	if value != "" || changes != 1 || !tt.Focused("Name") {
		t.Fatal("clear action stopped working after changing languages")
	}
}

func TestLocaleSwitchPinInputNamePreservesEditing(t *testing.T) {
	for _, numeric := range []bool{false, true} {
		t.Run(fmt.Sprintf("numeric=%v", numeric), func(t *testing.T) {
			value, english, chinese := "中A", "5 characters", "5 个字符"
			if numeric {
				value, english, chinese = "12", "5 digits", "5 位数字"
			}
			original := value
			var state nui.PinState
			changes, completes := 0, 0
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				nui.PinInput(c, th, &value, &state, nui.PinInputOptions{Numeric: numeric,
					OnChange: func(string) { changes++ }, OnComplete: func(string) { completes++ },
				})
			}, 400, 120)
			click(t, tt, english)
			tt.Key(ui.Ctrl, ui.KeyA)
			tt.SetLocale("zh-Hans")
			if !tt.Focused(chinese) || tt.HasText(english) {
				t.Fatalf("PIN name did not switch while keeping focus: %v", tt.Texts())
			}
			tt.Command("copy")
			if tt.Clipboard() != original || value != original || changes != 0 || completes != 0 {
				t.Fatal("PIN locale switch changed selection, value or notifications")
			}
			tt.SetLocale("en")
			if !tt.Focused(english) {
				t.Fatal("PIN name did not switch back")
			}
			tt.Type("3")
			if value != "3" || changes != 1 || completes != 0 {
				t.Fatal("PIN selection/editing changed after locale switch")
			}
		})
	}
}

func TestPinInputLocaleRespectsExplicitAndFieldLabels(t *testing.T) {
	for _, label := range []string{"", "Application PIN"} {
		value := ""
		var state nui.PinState
		tt := ui.NewTester(func(c *ui.Context) {
			th := nui.App(c, nui.AppConfig{})
			nui.FormField(c, th, nui.FormFieldOptions{Label: "Verification", Required: true}, func(th nui.Theme) {
				nui.PinInput(c, th, &value, &state, nui.PinInputOptions{Label: label, Placeholder: "·"})
			})
		}, 400, 120)
		for _, locale := range []string{"en", "zh-Hans"} {
			tt.SetLocale(locale)
			if tt.HasText("5 characters") || tt.HasText("5 个字符") || !tt.HasText("·") {
				t.Fatal("PIN fallback overrode field label or explicit placeholder")
			}
			want := label
			if want == "" {
				want = "Verification *"
			}
			if _, found := tt.Find(want); !found {
				t.Fatalf("missing application-owned label %q", want)
			}
		}
	}
}

func TestLocaleSwitchRefreshesExistingFormErrors(t *testing.T) {
	for _, kind := range []string{"number", "date", "time"} {
		t.Run(kind, func(t *testing.T) {
			var form nui.FormState
			var ns nui.NumberState
			var ds nui.DateInputState
			var ts nui.TimeInputState
			number := nui.Set(2.0)
			date := nui.Set(nui.Date{Year: 2026, Month: time.November, Day: 1})
			clock := nui.Set(nui.ClockTime{Hour: 10, Minute: 30})
			changes := 0
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				nui.Form(c, th, &form, nui.FormOptions{LoadingAuto: nui.Set(false)}, func(ft nui.Theme) {
					nui.FormField(c, ft, nui.FormFieldOptions{Name: "value", Label: "Field", Value: fmt.Sprint(number, date, clock)}, func(it nui.Theme) {
						o := nui.InputOptions{Label: "Editor", OnChange: func(string) { changes++ }}
						switch kind {
						case "number":
							nui.InputNumber(c, it, &number, &ns, nui.InputNumberOptions{InputOptions: o})
						case "date":
							nui.InputDate(c, it, &date, &ds, nui.InputDateOptions{InputOptions: o})
						case "time":
							nui.InputTime(c, it, &clock, &ts, nui.InputTimeOptions{InputOptions: o})
						}
					})
				})
			}, 600, 240)
			editorLabel := "Editor"
			if kind == "date" {
				editorLabel += " month"
			}
			if kind == "time" {
				editorLabel += " hour"
			}
			click(t, tt, editorLabel)
			tt.Key(ui.Ctrl, ui.KeyA)
			tt.Type("-")
			tt.Frame()
			if form.Validate(nil) {
				t.Fatal("invalid draft was accepted")
			}
			tt.Frame()
			english := form.Field("value").Error
			want := map[string]string{"number": "请输入允许范围内的数字", "date": "请输入可用日期", "time": "请输入有效时间"}[kind]
			before := changes
			form.BeginAsync()
			tt.SetLocale("zh-Hans")
			if form.Field("value").Error != want || !tt.HasText(want) || tt.HasText(english) {
				t.Fatalf("stored error did not translate: %q, text=%v", form.Field("value").Error, tt.Texts())
			}
			if kind == "date" {
				editorLabel = "Editor 月"
			}
			if kind == "time" {
				editorLabel = "Editor 时"
			}
			if !tt.Focused(editorLabel) || !form.Submitting || changes != before || number.Value != 2 || date.Value.Day != 1 || clock.Value.Hour != 10 {
				t.Fatal("language change altered focus, form ticket, callbacks or model")
			}
			draft := map[string]string{"number": ns.Text, "date": ds.Text, "time": ts.Text}[kind]
			if draft != "-" {
				t.Fatalf("lost draft: %q", draft)
			}
			tt.SetLocale("en")
			if form.Field("value").Error != english {
				t.Fatal("error did not switch back to English")
			}
			form.Field("value").Error = "Server validation"
			tt.SetLocale("zh-Hans")
			if form.Field("value").Error != "Server validation" || !tt.HasText("Server validation") {
				t.Fatal("application-owned error was replaced")
			}
		})
	}
}

func TestLocaleSwitchInsideSelectMenu(t *testing.T) {
	state, plain := nui.SelectMenuState{}, nui.SelectMenuState{Open: true, Query: "draft"}
	selected, none := []string{"go"}, []string(nil)
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Gap(8).Children(func() {
			nui.SelectMenu(c, th, &selected, &state, []nui.ListboxItem{{ID: "go", Label: "Go"}}, nui.SelectMenuOptions{Label: "Choices", Clearable: true, OnCreate: func(string) {}})
			// Without OnCreate, an unmatched query shows the empty state.
			nui.SelectMenu(c, th, &none, &plain, []nui.ListboxItem{{ID: "go", Label: "Go"}}, nui.SelectMenuOptions{Label: "Plain", Popup: nui.PopoverOptions{Placement: nui.TopStart}})
		})
	}, 500, 460)
	click(t, tt, "Choices")
	tt.Type("draft")
	tt.SetLocale("zh-Hans")
	if !state.Open || !tt.Focused("搜索Choices") || state.Query != "draft" || !slices.Equal(selected, []string{"go"}) {
		t.Fatal("open search lost state or focus during language change")
	}
	// Clicking the first menu dismissed the second, which reset its search.
	plain.Open, plain.Query = true, "draft"
	tt.Frame()
	for _, text := range []string{"没有匹配的数据", "创建“draft”", "清空选择"} {
		if !tt.HasText(text) {
			t.Fatalf("missing %q: %v", text, tt.Texts())
		}
	}
	click(t, tt, "清空选择")
	if len(selected) != 0 || state.Query != "" {
		t.Fatal("translated clear action stopped working")
	}
	state.Open, state.Query = true, "draft"
	tt.Frame()
	tt.SetLocale("en")
	if !tt.HasText("Create “draft”") {
		t.Fatal("popup did not switch back to English")
	}
}

func TestLocaleSwitchCalendarAndPagination(t *testing.T) {
	selection := nui.DateSelection{Start: nui.Set(nui.Date{Year: 2026, Month: time.November, Day: 1})}
	state := nui.CalendarState{}
	page := 2
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Children(func() {
			nui.Calendar(c, th, &selection, &state, nui.CalendarOptions{Label: "Date"})
			nui.Pagination(c, th, &page, nui.PaginationOptions{Total: 100, PageSize: 10})
		})
	}, 640, 560)
	click(t, tt, "Next month")
	tt.SetLocale("zh-Hans")
	for _, text := range []string{"2026年12月", "日", "下个月", "上一页", "上一年", "下一年"} {
		if !tt.HasText(text) {
			t.Fatalf("missing %q: %v", text, tt.Texts())
		}
	}
	if state.Month.Month != time.December || selection.Start.Value.Month != time.November || page != 2 {
		t.Fatal("language change reset browsed month, selection or page")
	}
	click(t, tt, "下一页")
	if page != 3 {
		t.Fatal("translated pagination stopped working")
	}
}

func TestLocaleSwitchUploadErrorsAndActiveJob(t *testing.T) {
	var state nui.FileUploadState
	state.Add([]nui.UploadFile{{Name: "large.txt", Size: 20}, {Name: "custom.txt", Size: 1, Error: "Storage denied"}, {Name: "ok.txt", Size: 1}}, nui.UploadRules{MaxBytes: 10})
	ctx, ticket := state.Begin(state.Files[2].ID)
	defer state.Dispose()
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		// The button variant shows the application's ChooseLabel.
		nui.FileUpload(c, th, &state, nui.FileUploadOptions{Label: "Attachments", ChooseLabel: "Browse my files", Variant: nui.UploadButton})
	}, 560, 600)
	tt.SetLocale("zh-Hans")
	for _, text := range []string{"文件大小超过 10 字节", "Storage denied", "Browse my files", "上传中", "取消ok.txt"} {
		if !tt.HasText(text) {
			t.Fatalf("missing %q: %v", text, tt.Texts())
		}
	}
	if ctx.Err() != nil || !state.Report(ticket, .5, false, nil) {
		t.Fatal("language change invalidated the upload ticket")
	}
	tt.SetLocale("en")
	if !tt.HasText("File exceeds 10 bytes") || !tt.HasText("Cancel ok.txt") {
		t.Fatal("existing rejected and running uploads did not switch back")
	}
	state.Files[0].Error = "Custom rejection"
	state.Report(ticket, .5, false, errors.New("Server error"))
	tt.SetLocale("zh-Hans")
	if !tt.HasText("Custom rejection") || !tt.HasText("Server error") {
		t.Fatal("application errors were translated or discarded")
	}
}

func TestLocaleSwitchExistingToast(t *testing.T) {
	closed := 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Button(c, th, "Notify", nui.ButtonOptions{}).OnClick(func() {
			nui.Toast(c, th, nui.ToastOptions{Title: "Saved", Timeout: time.Hour, OnClose: func() { closed++ }})
		})
		nui.ToastViewport(c, th, nui.ToastViewportOptions{})
	}, 500, 350)
	click(t, tt, "Notify")
	tt.SetLocale("zh-Hans")
	if !tt.HasText("Saved") || !tt.HasText("关闭Saved") {
		t.Fatal("existing toast lost its title or retained the English close action")
	}
	click(t, tt, "关闭Saved")
	if closed != 1 {
		t.Fatal("translated toast could not be dismissed")
	}
}

func TestLocaleSwitchCommandPalettePendingSearch(t *testing.T) {
	state := nui.CommandPaletteState{Open: true, Query: "draft"}
	searches := 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.CommandPalette(c, th, &state, nui.CommandPaletteOptions{Search: func(context.Context, nui.QueryTicket, string, string) { searches++ }})
	}, 700, 620)
	tt.SetLocale("zh-Hans")
	for _, text := range []string{"命令面板", "搜索中…", "关闭命令面板", "命令面板结果"} {
		if !tt.HasText(text) {
			t.Fatalf("missing %q: %v", text, tt.Texts())
		}
	}
	if !tt.Focused("搜索命令面板") || state.Query != "draft" || searches != 1 || !state.Results.Loading {
		t.Fatal("language change disrupted the pending search or editor")
	}
	click(t, tt, "关闭命令面板")
	if state.Open || state.Results.Loading {
		t.Fatal("translated close action did not close/cancel the palette")
	}
}

func TestLocaleSwitchMyGoColorPicker(t *testing.T) {
	value := ui.RGB(30, 60, 90)
	var state nui.ColorPickerState
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.ColorPicker(c, th, &value, &state, nui.ColorPickerOptions{Label: "Brand"})
	}, 400, 620)
	tt.SetLocale("zh-Hans")
	if !tt.HasText("饱和度与亮度") || !tt.HasText("十六进制") || !tt.HasText("增加Brand R") {
		t.Fatalf("native and library controls should follow the same locale: %v", tt.Texts())
	}
	if value != ui.RGB(30, 60, 90) {
		t.Fatal("language change altered the selected color")
	}
}
