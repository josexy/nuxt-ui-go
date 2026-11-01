package gallery

import (
	"fmt"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

type propsState struct {
	Demo                                 string
	Direction                            nui.Orientation
	Name                                 string
	Rating                               int
	Low, High                            float64
	Inverted                             bool
	Listbox                              nui.ListboxState
	Selected                             []string
	Carousel                             nui.CarouselState
	Arrows, Dots                         bool
	Loading, Enabled                     bool
	Highlight                            bool
	CommittedRange                       string
	Choices                              []string
	ChoiceVariant                        nui.ChoiceVariant
	Indicator                            nui.IndicatorPlacement
	Tags                                 []string
	TagsState                            nui.TagsState
	AddTab, AddBlur, AddPaste, Duplicate bool
	Search, Reset, External              bool
	Menu                                 nui.SelectMenuState
	MenuValue, TypedValue                []string
	InputMenu                            nui.SearchState
	Page, Siblings                       int
	Controls, Edges                      bool
	Manual, Collapsible                  bool
	Tab                                  string
	Accordion                            []string
	Popover, ControlledTip, Dismissible  bool
	Status                               string
	Pin                                  string
	PinState                             nui.PinState
	Preview, Remove, Grid                bool
	Files                                nui.FileUploadState
}

func newProps() propsState {
	return propsState{Demo: "forms", Name: "Nuxt workspace", Rating: 3, Low: 20, High: 75, Selected: []string{"beta"}, Arrows: true, Dots: true, ChoiceVariant: nui.ChoiceCard, Indicator: nui.IndicatorEnd, AddTab: true, AddBlur: true, AddPaste: true, Search: true, Reset: true, Page: 5, Controls: true, Edges: true, Manual: true, Tab: "overview", Accordion: []string{"intro"}, Collapsible: true, Dismissible: true, Status: "Ready", Preview: true, Remove: true, Files: nui.FileUploadState{Files: []nui.UploadFile{{ID: "sample", Name: "review.md", Size: 2048}}}}
}

func propsChecks(c *ui.Context, t nui.Theme, checks ...struct {
	label string
	value *bool
}) {
	ui.Row(c).Wrap().Gap(14).Children(func() {
		for _, check := range checks {
			nui.Checkbox(c.Key(check.label), t, check.value, nui.CheckboxOptions{Label: check.label, Size: nui.SM})
		}
	})
}
func check(label string, value *bool) struct {
	label string
	value *bool
} {
	return struct {
		label string
		value *bool
	}{label, value}
}

func (a *State) commonProps(c *ui.Context, t nui.Theme) {
	p := &a.Props
	heading(c, t, "Props & behavior", "Change a policy, then try the mouse and keyboard.")
	nui.Tabs(c, t, &p.Demo, []nui.TabItem{
		{ID: "forms", Label: "Form controls", Content: func(c *ui.Context, t nui.Theme) { a.formProps(c, t) }},
		{ID: "selection", Label: "Selection policies", Content: func(c *ui.Context, t nui.Theme) { a.selectionProps(c, t) }},
		{ID: "overlays", Label: "Overlays & files", Content: func(c *ui.Context, t nui.Theme) { a.overlayProps(c, t) }},
	}, nui.TabsOptions{Label: "Example category"})
}

func (a *State) formProps(c *ui.Context, t nui.Theme) {
	p := &a.Props
	section(c, t, "Switch & choices", func() {
		propsChecks(c, t, check("Loading switch", &p.Loading), check("Highlight choices", &p.Highlight))
		nui.Switch(c, t, &p.Enabled, nui.SwitchOptions{Label: "Sync workspace", Loading: p.Loading, CheckedIcon: lucide.Check(), UncheckedIcon: lucide.X()})
		ui.Row(c).Wrap().Gap(12).Children(func() {
			nui.Select(c, t, &p.ChoiceVariant, []nui.Option[nui.ChoiceVariant]{{Value: nui.ChoiceList, Label: "list"}, {Value: nui.ChoiceCard, Label: "card"}, {Value: nui.ChoiceTable, Label: "table"}}, nui.SelectOptions{Label: "Choice variant"}).Width(150)
			nui.Select(c, t, &p.Indicator, []nui.Option[nui.IndicatorPlacement]{{Value: nui.IndicatorStart, Label: "start"}, {Value: nui.IndicatorEnd, Label: "end"}, {Value: nui.IndicatorHidden, Label: "hidden"}}, nui.SelectOptions{Label: "Indicator position"}).Width(150)
		})
		nui.CheckboxGroup(c, t, &p.Choices, []nui.Option[string]{{Value: "build", Label: "Build", Description: "Compile projects"}, {Value: "test", Label: "Test", Description: "Run checks"}, {Value: "deploy", Label: "Deploy", Description: "Publish a release"}}, nui.CheckboxGroupOptions{Label: "Pick up to two", Max: nui.Set(2), Horizontal: true, Variant: p.ChoiceVariant, Indicator: p.Indicator, Highlight: p.Highlight})
	})
	section(c, t, "Tags & PIN", func() {
		propsChecks(c, t, check("Add on Tab", &p.AddTab), check("Add on blur", &p.AddBlur), check("Add on paste", &p.AddPaste), check("Allow duplicate tags", &p.Duplicate))
		nui.InputTags(c, t, &p.Tags, &p.TagsState, nui.InputTagsOptions{Label: "Policy tags", Placeholder: "Type a tag (up to 12 characters)", AddOnTab: p.AddTab, AddOnBlur: p.AddBlur, AddOnPaste: nui.Set(p.AddPaste), Duplicate: p.Duplicate, MaxLength: 12, MaxTags: 6})
		ui.Text(c, "Tab adds a tag and keeps focus; an empty draft lets Tab move on.").FontSize(12).TextColor(t.Muted)
		nui.PinInput(c, t, &p.Pin, &p.PinState, nui.PinInputOptions{Numeric: true, Length: 6, Label: "Grouped PIN", Separator: []int{3}, Color: nui.Info, Size: nui.LG})
	})
	section(c, t, "Activation & disclosure", func() {
		propsChecks(c, t, check("Manual tab activation", &p.Manual), check("Allow closing all sections", &p.Collapsible))
		mode := nui.ActivationAutomatic
		if p.Manual {
			mode = nui.ActivationManual
		}
		nui.Tabs(c, t, &p.Tab, []nui.TabItem{{ID: "overview", Label: "Overview", Content: func(c *ui.Context, t nui.Theme) { ui.Text(c, "Overview selected").TextColor(t.Muted) }}, {ID: "settings", Label: "Settings", Content: func(c *ui.Context, t nui.Theme) { ui.Text(c, "Settings selected").TextColor(t.Muted) }}}, nui.TabsOptions{Label: "Activation preview", ActivationMode: mode})
		nui.Accordion(c, t, &p.Accordion, []nui.AccordionItem{{ID: "intro", Label: "Keyboard help", Content: func(c *ui.Context, t nui.Theme) {
			ui.Text(c, "Manual tabs move focus with arrows; press Enter or Space to select.").TextColor(t.Muted)
		}}, {ID: "more", Label: "Disclosure help", Content: func(c *ui.Context, t nui.Theme) {
			ui.Text(c, "Disable closing all sections to keep the active section open.").TextColor(t.Muted)
		}}}, nui.AccordionOptions{Collapsible: nui.Set(p.Collapsible)})
	})
}

var propsItems = []nui.ListboxItem{{ID: "alpha", Label: "Alpha", Description: "Design"}, {ID: "beta", Label: "Beta", Description: "Engineering"}, {ID: "gamma", Label: "Gamma", Description: "Support"}}
var propsOptions = []nui.Option[string]{{Value: "alpha", Label: "Alpha", Description: "Design"}, {Value: "beta", Label: "Beta", Description: "Engineering"}, {Value: "gamma", Label: "Gamma", Description: "Support"}}

func (a *State) selectionProps(c *ui.Context, t nui.Theme) {
	p := &a.Props
	section(c, t, "SelectMenu & typed multiple Select", func() {
		propsChecks(c, t, check("Search input", &p.Search), check("Reset search on select / close", &p.Reset))
		nui.SelectMenu(c, t, &p.MenuValue, &p.Menu, propsItems, nui.SelectMenuOptions{Label: "Team menu", SearchInput: nui.Set(p.Search), ResetSearchTermOnSelect: nui.Set(p.Reset), ResetSearchTermOnBlur: nui.Set(p.Reset), Size: nui.LG})
		nui.SelectMultiple(c, t, &p.TypedValue, propsOptions, nui.SelectOptions{Label: "Multiple teams", Placeholder: "Select teams"})
		ui.Text(c, "Close and reopen the menu, or focus its trigger and press ↑.").TextColor(t.Muted)
	})
	section(c, t, "InputMenu filtering & creation", func() {
		propsChecks(c, t, check("Use external filtering", &p.External))
		nui.InputMenu(c, t, &p.InputMenu, propsOptions, nui.InputMenuOptions{Label: "Team search", Placeholder: "Search teams or create one", Multiple: true, ExternalFilter: p.External, OnCreate: func(v string) { p.Status = "Create requested: " + v }})
		ui.Text(c, "External filtering keeps every supplied result; the app owns the search request.").TextColor(t.Muted)
		ui.Text(c, p.Status).TextColor(t.Primary)
	})
	section(c, t, "Pagination", func() {
		ui.Row(c).Wrap().Gap(12).Children(func() {
			nui.Select(c, t, &p.Siblings, []nui.Option[int]{{Value: 0, Label: "0 neighbors"}, {Value: 1, Label: "1 neighbor"}, {Value: 2, Label: "2 neighbors"}}, nui.SelectOptions{Label: "Sibling count"}).Width(180)
			propsChecks(c, t, check("Page controls", &p.Controls), check("Edge pages", &p.Edges))
		})
		nui.Pagination(c, t, &p.Page, nui.PaginationOptions{Total: 100, PageSize: 10, SiblingCount: nui.Set(p.Siblings), ShowControls: nui.Set(p.Controls), ShowEdges: nui.Set(p.Edges), Size: nui.MD})
	})
}

func (a *State) overlayProps(c *ui.Context, t nui.Theme) {
	p := &a.Props
	section(c, t, "Hover & controlled overlays", func() {
		propsChecks(c, t, check("Dismiss on Escape / outside click", &p.Dismissible), check("Show controlled tooltip", &p.ControlledTip))
		ui.Row(c).Wrap().Gap(16).Children(func() {
			anchor := nui.Button(c, t, "Hover for details", nui.ButtonOptions{Variant: nui.Outline})
			nui.Popover(c, t, anchor, &p.Popover, nui.PopoverOptions{Mode: nui.PopoverHover, OpenDelay: 200 * time.Millisecond, CloseDelay: 150 * time.Millisecond, Dismissible: nui.Set(p.Dismissible)}, func(t nui.Theme) {
				ui.Text(c, "The panel stays open while you hover over it.")
				nui.Button(c, t, "Close details", nui.ButtonOptions{Size: nui.SM}).OnClick(func() { p.Popover = false })
			})
			help := nui.Button(c, t, "Shortcut help", nui.ButtonOptions{Variant: nui.Ghost})
			nui.Tooltip(c, t, help, "Search", nui.TooltipOptions{Open: &p.ControlledTip, Placement: nui.Set(nui.BottomStart), Kbds: []string{"Ctrl", "K"}})
			nui.Button(c, t, "Show action toast", nui.ButtonOptions{}).OnClick(func() {
				nui.Toast(c, t, nui.ToastOptions{Title: "Workspace ready", Description: "Review or dismiss this update.", Actions: []nui.ToastActionOptions{{Label: "Review", Close: nui.Set(false), OnAction: func() { p.Status = "Review opened" }}, {Label: "Done"}}})
			})
		})
		ui.Text(c, p.Status).TextColor(t.Muted)
	})
	section(c, t, "File preview policies", func() {
		propsChecks(c, t, check("Show file previews", &p.Preview), check("Allow file removal", &p.Remove), check("Grid layout", &p.Grid))
		layout := nui.UploadList
		if p.Grid {
			layout = nui.UploadGrid
		}
		nui.FileUpload(c, t, &p.Files, nui.FileUploadOptions{Label: "Review attachments", Description: "Choose local files to inspect their metadata.", Layout: layout, Preview: nui.Set(p.Preview), FileDelete: nui.Set(p.Remove), Dialog: a.FileDialog})
	})
}

func (a *State) orientationProps(c *ui.Context, t nui.Theme) {
	p := &a.Props
	heading(c, t, "Orientation", "Switch the layout while values and focused editors stay in place.")
	ui.Row(c).Wrap().Gap(16).Children(func() {
		nui.Select(c, t, &p.Direction, []nui.Option[nui.Orientation]{{Value: nui.Horizontal, Label: "horizontal"}, {Value: nui.Vertical, Label: "vertical"}}, nui.SelectOptions{Label: "Shared orientation"}).Width(210)
		propsChecks(c, t, check("Carousel arrows", &p.Arrows), check("Carousel dots", &p.Dots))
	})
	section(c, t, "FormField & Alert", func() {
		nui.FormField(c, t, nui.FormFieldOptions{Label: "Workspace name", Description: "Visible to your team", Hint: "Required", Help: "You can rename it later.", Orientation: nui.Set(p.Direction)}, func(t nui.Theme) { nui.Input(c, t, &p.Name, nui.InputOptions{Label: "Orientation name"}) })
		nui.Alert(c, t, nui.AlertOptions{Title: "Workspace ready", Description: "Choose an action to continue.", Variant: nui.Soft, Orientation: nui.Set(p.Direction), Actions: func() {
			nui.Button(c, t, "Preview toast", nui.ButtonOptions{Size: nui.SM}).OnClick(func() {
				nui.Toast(c, t, nui.ToastOptions{Title: "Layout preview", Description: "Actions follow the selected orientation.", Orientation: nui.Set(p.Direction), Actions: []nui.ToastActionOptions{{Label: "Review", Close: nui.Set(false)}, {Label: "Done"}}})
			})
		}})
	})
	section(c, t, "Rating, range & progress", func() {
		propsChecks(c, t, check("Reverse sliders", &p.Inverted))
		ui.Row(c).Wrap().Gap(24).AlignItems(ui.Start).Children(func() {
			ui.Column(c).Width(170).Gap(10).Children(func() {
				ui.Text(c, "Rating").TextColor(t.Muted)
				nui.InputRating(c, t, &p.Rating, nui.InputRatingOptions{Label: "Orientation rating", Orientation: p.Direction, Clearable: true})
				ui.Textf(c, "%d / 5", p.Rating).TextColor(t.Muted)
			})
			ui.Column(c).Width(210).Gap(10).Children(func() {
				ui.Text(c, "Range · 2-step minimum gap").TextColor(t.Muted)
				nui.RangeSlider(c, t, &p.Low, &p.High, nui.RangeSliderOptions{Label: "Orientation range", Orientation: p.Direction, Inverted: p.Inverted, Min: 0, Max: 100, Step: 5, MinStepsBetweenThumbs: 2, Tooltip: true, OnCommit: func(low, high float64) { p.CommittedRange = fmt.Sprintf("Confirmed: %g – %g", low, high) }})
				ui.Text(c, fmt.Sprintf("%g – %g", p.Low, p.High)).TextColor(t.Muted)
				if p.CommittedRange != "" {
					ui.Text(c, p.CommittedRange).TextColor(t.Muted)
				}
			})
			ui.Column(c).Width(200).Gap(10).Children(func() {
				ui.Text(c, "Progress group").TextColor(t.Muted)
				nui.ProgressGroup(c, t, []nui.ProgressSegment{{ID: "done", Label: "Complete", Value: 60, Color: nui.Primary}, {ID: "review", Label: "Review", Value: 25, Color: nui.Info}}, nui.ProgressGroupOptions{Label: "Work progress", Max: nui.Set(100.0), Orientation: p.Direction})
			})
		})
	})
	section(c, t, "Listbox & Carousel", func() {
		nui.Listbox(c, t, &p.Selected, &p.Listbox, propsItems, nui.ListboxOptions{Label: "Orientation list", Orientation: nui.Set(p.Direction), Height: 110, Loop: true})
		nui.Carousel(c, t, &p.Carousel, []nui.CarouselItem{{ID: "start", Label: "Start", Build: func(c *ui.Context, t nui.Theme) {
			ui.Text(c, "1 · Start a workspace").FontSize(20).FontWeight(600)
			ui.Text(c, "Use arrow keys or drag along the selected axis.").TextColor(t.Muted)
		}}, {ID: "review", Label: "Review", Build: func(c *ui.Context, t nui.Theme) { ui.Text(c, "2 · Review your changes").FontSize(20).FontWeight(600) }}, {ID: "share", Label: "Share", Build: func(c *ui.Context, t nui.Theme) { ui.Text(c, "3 · Share with the team").FontSize(20).FontWeight(600) }}}, nui.CarouselOptions{Label: "Orientation guide", Height: 120, Orientation: p.Direction, Arrows: nui.Set(p.Arrows), Dots: p.Dots})
	})
}
