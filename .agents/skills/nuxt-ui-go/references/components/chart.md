<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Chart

Display data as lines, bars, areas or slices.

- Category: Advanced interaction and layout
- Version: nuxt-ui-go v0.0.2 plus unreleased changes (MyGo v0.9.0)
- Functions: `Chart`
- Types: `ChartOptions`, `ChartSelection`
- Theme key: `nui.ComponentChart` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Keep ChartState persistent and update it with SetCategories, SetData, SetTimeData or SetSlices. Automatic legends appear only for multiple series or slices. Height defaults to 300; AspectRatio applies when height is unset.

Grid lines paint one device pixel; data series LineWidth remains in DIPs.

## Things to know

Based on Nuxt UI 4.12.0 with TanStack Charts; Go uses MyGo and retains scatter/pie charts, ordered numeric/time axes and interactive legends. ChartSelection identifies an X row or slice, not raw Web engine point/group objects. YLabel is horizontal above the plot. Explicit DonutInnerRadius overrides Thickness. Height covers the plot and axes; legends, titles and the description line use additional space. The description line is plain text, not a live region: keyboard selection is announced once, and pointer hover is not read aloud.

## API

```go
func Chart(c *ui.Context, t Theme, s *ChartState, o ChartOptions) ui.Element
```

### ChartOptions

```go
type ChartOptions struct {
	Label, Description, EmptyText string
	XLabel, YLabel                string
	Kind                          ChartKind
	Curve                         ChartCurve
	// Stack applies to bars and areas; other types ignore it.
	Stack                           ChartStackMode
	Height, Width, AspectRatio      float32
	XAxis, YAxis, Tooltip           Optional[bool]
	TooltipOptions                  *TooltipOptions
	Points                          bool
	Colors                          []ui.Color
	Thickness                       Optional[float32]
	Legend, InteractiveLegend, Grid Optional[bool]
	FormatX, FormatY                func(float64) string
	FormatTime                      func(time.Time) string
	// Animate defaults to true and respects theme/system reduced motion.
	Animate                 Optional[bool]
	AnimationDuration       time.Duration     // default 350ms
	BarPadding              Optional[float32] // category band inner padding, default .25
	DonutInnerRadius        Optional[float64] // Explicit fraction overrides Thickness.
	SlicePadding            Optional[float64] // angle in radians, default .02
	LineWidth, PointSize    Optional[float32]
	OnLegendChange          func(id string, visible bool)
	OnFocusChange, OnSelect func(ChartSelection)
	Center                  func(*ui.Context, Theme)
	Decoration
}
```

### ChartSelection

```go
// ChartSelection identifies the focused X row or radial slice. Index is -1 when
// nothing is focused; SliceID is empty for Cartesian charts.
type ChartSelection struct {
	Index   int
	SliceID string
}
```

### Related types

```go
// ChartState owns copied data and per-window selection, visibility and animation.
// SetData, SetCategories and SetTimeData replace Cartesian data; SetSlices replaces
// radial data independently. Publish worker-prepared state through Window.Update;
// never mutate a state while a view is using it, or copy an already rendered state.
type ChartState struct {
	Selected      int
	SelectedSlice string
	// contains filtered or unexported fields
}

// Prepare prewarms the default line chart cache. Other options are prepared by Chart.
func (s *ChartState) Prepare(width int)

func (s *ChartState) SampleCount() int

func (s *ChartState) SeriesVisible(id string) bool

// SetCategories sets a band axis with unique, nonempty category labels.
func (s *ChartState) SetCategories(labels []string, series []ChartSeries) error

// SetData sets a shared numeric X axis. NaN/Inf Y values are gaps, not zeros.
// An error leaves the previous data and visibility unchanged.
func (s *ChartState) SetData(xs []float64, series []ChartSeries) error

// SetSeriesVisible returns whether a known series changed visibility.
// Visibility follows the ID across data updates and reordering.
func (s *ChartState) SetSeriesVisible(id string, visible bool) bool

func (s *ChartState) SetSliceVisible(id string, visible bool) bool

func (s *ChartState) SetSlices(values []ChartSlice) error

// SetTimeData sets a strictly increasing time axis; ticks use the first time's location.
func (s *ChartState) SetTimeData(times []time.Time, series []ChartSeries) error

func (s *ChartState) SliceVisible(id string) bool
```

```go
type ChartKind uint8

const (
	ChartLine ChartKind = iota
	ChartBar
	ChartArea
	ChartScatter
	ChartPie
	ChartDonut
)
```

```go
type ChartCurve uint8

const (
	ChartCurveLinear ChartCurve = iota
	ChartCurveMonotone
	ChartCurveStepAfter
	ChartCurveStepBefore
	ChartCurveStep
)
```

```go
type ChartStackMode uint8

const (
	ChartStackNone ChartStackMode = iota
	ChartStackStandard
	ChartStackNormalized
	ChartStackCenter
)
```

```go
type ChartSeries struct {
	ID, Label string
	Values    []float64
	Color     Color
}
```

```go
// ChartSlice is one independently named and colored pie/donut segment.
// Values must be finite and nonnegative. Zero values remain in the legend.
type ChartSlice struct {
	ID, Label string
	Value     float64
	Color     Color
}
```

Types documented in other references: `Decoration` → [theme.md](theme.md), `Optional` → [theme.md](theme.md), `Theme` → [theme.md](theme.md), `TooltipOptions` → [tooltip.md](tooltip.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Advanced interaction, charts, and layout

#### Chart

##### Prepare the data, then draw the chart

Chart draws with MyGo's `ui/chart` and adds theming, interactive legends, and animation. A ChartState should be owned by a single chart instance in a single window. Do not copy a rendered state or share mutable data with background code.

Call the setter methods when the data is initialized or changes, and check the error:

```go
err := state.Activity.SetCategories([]string{"Jan", "Feb", "Mar"}, []nui.ChartSeries{
    {ID: "build", Label: "Builds", Values: []float64{30, 45, 38}},
    {ID: "review", Label: "Reviews", Values: []float64{12, 19, 16}},
})
if err != nil {
    return err
}
```

Put this fragment in a data initialization function that returns an error. Write background results to the window state through Window.Update. Create the chart in the view function:

```go
nui.Chart(c, theme, &state.Activity, nui.ChartOptions{
    Label: "Activity",
    Kind: nui.ChartArea,
    Curve: nui.ChartCurveMonotone,
    Stack: nui.ChartStackStandard,
    Height: 240,
    AnimationDuration: 400 * time.Millisecond,
})
```

##### Chart types and data APIs

| Need | API |
| --- | --- |
| Line, bar, area, scatter, pie, and donut charts | ChartLine, ChartBar, ChartArea, ChartScatter, ChartPie, ChartDonut |
| Numeric / category / time X axis | SetData, SetCategories, SetTimeData |
| Pie and donut charts | SetSlices([]ChartSlice) |
| Curves | ChartCurveLinear, ChartCurveMonotone, ChartCurveStepAfter, ChartCurveStepBefore, ChartCurveStep |
| Stacking | ChartStackNone, ChartStackStandard, ChartStackNormalized, ChartStackCenter; for bar and area charts |

Every setter validates before copying, so an error leaves the existing data unchanged. The data must meet these requirements:

| Method | Data requirements |
| --- | --- |
| SetData(xs, series) | X values are finite and strictly increasing; each series has the same length as X; series IDs are unique and non-empty. A Y value of NaN or Inf marks a missing value |
| SetCategories(labels, series) | Category labels are unique and non-empty; uses a band category axis |
| SetTimeData(times, series) | Times are strictly increasing; ticks use the time zone of the first entry |
| SetSlices(slices) | Slice IDs are unique and non-empty; values are finite and non-negative, and the total is finite. A value of 0 keeps its legend entry but draws no slice |

Scatter charts also use shared, ordered X data. ChartSlice contains ID, Label, Value, and Color. The same state can hold both Cartesian data and slice data, and Kind chooses which one to draw.

##### Axes, colors, and stacking

The Y axis always includes 0. Standard stacks positive values upward and negative values downward separately. Normalized divides by the sum of the absolute visible values at each X, so all-positive values add up to 100%, and mixed positive and negative values share that denominator. Center centers the algebraic sum on 0.

The default series colors are, in order, Primary, blue, violet, amber, rose, and cyan. After the first round, each additional round mixes the colors with the background by a factor of 0.65. Light mode uses shade 500 and dark mode uses shade 400. Colors can supply a list of native colors, and a series' explicit Color takes precedence; Neutral uses the inverted color. Mixing uses MyGo's native color interpolation and does not claim to match CSS oklab mixing exactly.

FormatX, FormatY, and FormatTime set the formatting. Numbers are formatted with Go by default, not with browser Intl number formatting. XAxis and YAxis default to true and control each axis's ticks and title; Grid independently controls the horizontal grid lines. XLabel is centered below the chart, and YLabel is shown horizontally above it. In RTL, the X axis data and ticks are reversed, the Y axis moves to the right, and physical arrow keys and pointer positions are mapped accordingly.

Height defaults to 300 DIP, and a positive value takes precedence; otherwise, a positive AspectRatio computes the height from the actual width. A positive Width fixes the width; otherwise, the chart follows its parent and uses 640 before the first measurement. Height applies to the plot area; the legend, axis titles, and the description line take extra space. All sizes must be finite. BarPadding, SlicePadding, LineWidth, and PointSize remain as native geometry options. Points defaults to false; when on, line charts show the sampled data points. PointSize does not affect the separate focus marker.

A donut's Thickness defaults to 24 DIP. An explicit DonutInnerRadius can still be given as a ratio and takes precedence over Thickness; it is clamped to 0–0.95. Thickness is clamped to non-negative values, and 0 makes the inner and outer edges meet. Center synchronously builds content in the center of the plot area without intercepting the pointer; the callback must not keep Elements or Contexts.

##### Legends and data selection

When Legend is unset, the legend appears only if there is more than one series or slice in total, counting hidden and zero-value items; an explicit true or false takes precedence. The legend is clickable by default, and focused entries can toggle a series with Space or Enter. OnLegendChange(id, visible) reports only user actions. Your app can query and change visibility with SeriesVisible/SetSeriesVisible and SliceVisible/SetSliceVisible. An empty Label defaults to the corresponding ID.

InteractiveLegend=false only displays the legend, and Legend=false hides it entirely. Hiding a series recomputes the axes and stacking, and hiding a slice recomputes the visible proportions. Visibility is kept by ID across updates and reordering, and is cleared when an ID is removed. Even when everything is hidden, the legend stays so the user can restore items.

Cartesian charts support pointing at the nearest X with the mouse, and Left, Right, Home, End, and Escape; Home and End go to the logical first and last data points. The description line at the bottom and the Tooltip show the raw values, and Tooltip=false hides only the tooltip. TooltipOptions configures settings such as position; the default is TopCenter with a 12 DIP offset and no delay. Pie and donut charts support pointing at slices and arrow keys, and show percentages of the visible total. The donut's center does not respond to slice pointing. Selection on dense data always looks up the raw values.

OnFocusChange fires when the pointer or keyboard changes the focused row or slice; repeated keys at the bounds do not fire it again. Clicking or pressing Enter/Space fires OnSelect, and activating again can fire it again. ChartSelection.Index is the raw X row or slice index, or -1 for no focus; SliceID is used only by radial charts. Go works in whole X rows and has no separate per-series point, focusGroupChange object, or DOM event from the web engine. The chart uses focusable group semantics, and native assistive technology can still read the center text and the empty state.

The plot area reports the focused row or slice as its value. A selection moved with the keyboard is announced once, with the row's values or the slice's label, value, and percentage. Selection that follows the pointer is not announced, and the description line is plain text rather than a live region, so hovering never reads aloud. The empty-state message is a status that is not read when it first appears.

##### Animation

Animate is on by default, AnimationDuration defaults to 350 ms, and both theme and system reduced motion settings are respected. The first display appears directly; only later data or option changes animate.

When the chart type, series, and missing-value structure stay the same, coordinates are interpolated, and switching between Pie and Donut also animates the inner radius. When the chart type, series, point count, visibility, or missing-value structure changes, the chart cross-fades. An update in the middle of an animation continues from the current frame, keeping at most 3 exiting frames and 1 target frame. Axis labels and tooltips use the target data.

When the animation finishes, or the chart is invisible or unmounted, it stops requesting frames; there is no background timer. `Animate: nui.Set(false)` shows the target values immediately.

##### Large data and limits

Prepare(pixelWidth) can prepare the default line cache in advance; Chart prepares everything else. The sampling width is capped at 2,048 and is recomputed only after the data, visibility, size, or geometry options change. Drawing and interpolation work on the sampled shapes.

Unstacked data uses min-max sampling. When there are too many gaps in the data, fixed buckets are used instead, and buckets containing missing values stay disconnected. Dense stacked data shares bucket endpoints and overall envelope extremes so the X boundaries stay aligned, and details of a single series within one pixel may be dropped. Dense bar charts are not summed or averaged automatically; aggregate your business data first if you need that.

The chart is modeled on the TanStack Charts used by Nuxt UI 4.12.0, but Go actually uses MyGo's ui/chart and depends on no web charting engine. Native capabilities such as scatter and pie charts and numeric and time axes are kept. Arbitrary TanStack definitions, dual Y axes, logarithmic axes, brush zoom, and touch gestures are not provided. For low-level custom marks and layouts, compose ui/chart directly.

#### State and part index (excerpt)

| Component | Persistent state or data | Parts |
| --- | --- | --- |
| Chart | ChartState | Root, Body, Item (legend), Label, Description, EmptySlot, List, Indicator, XAxisSlot, YAxisSlot, CenterSlot |

### Compatibility and known limitations

#### Feature scope › Data and files (excerpt)

| Area | Supported | Not provided, or left to the app |
| --- | --- | --- |
| Chart | Six native chart types; category, value, and time axes; stacking; interactive legends; axis and tooltip options; RTL; row and sector selection notifications | Modeled on TanStack Charts. No arbitrary definitions, standalone web point/group objects, dual Y axes, logarithmic axes, or brush zoom. See Chart |

#### Accessibility and motion (excerpt)

Since MyGo v0.9.0, `ui.RoleStatus` is a polite, atomic live region. The library's own progress, error, empty, and loading messages keep stable identities: such a status is read once when its text changes and is silent when it first appears, clears, or rebuilds unchanged. Alert and Banner take their identity from the key the app gives them. CommandPalette also announces its progress and error messages once when they appear. FileUpload keeps empty status regions mounted, so adding a message to an exposed region is read once; hidden updates are silent, and mounting or exposing an existing message does not announce it as a change. Chart announces keyboard selection once and never reads pointer hover. Alert and Banner read their new title and description once when either changes. These rules are checked headlessly with `Tester.UseScreenReader`; what VoiceOver, NVDA, Narrator, and Orca actually say has not been verified.

## Complete example

A runnable program. App state lives in `example`, outside the view, and the view only builds UI.

```go
package main

import (
	"fmt"
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

type example struct {
	Chart                        nui.ChartState
	Slices                       nui.ChartState
	Focuses, Selections, Legends int
	Last                         string
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	kind := nui.ChartLine
	chart := &s.Chart
	if kind == nui.ChartPie || kind == nui.ChartDonut {
		chart = &s.Slices
	}
	o := nui.ChartOptions{Label: "Weekly activity", Description: "Daily views and visits", Kind: kind, XAxis: nui.Set(true), YAxis: nui.Set(true), XLabel: "Weekday", YLabel: "Events", Tooltip: nui.Set(true), Thickness: nui.Set(float32(24)), OnFocusChange: func(nui.ChartSelection) { s.Focuses++ }, OnSelect: func(point nui.ChartSelection) {
		s.Selections++
		s.Last = fmt.Sprintf("Activated index %d / slice %s", point.Index, point.SliceID)
	}, OnLegendChange: func(string, bool) { s.Legends++ }}

	direction := ui.LeftToRight

	ui.Column(c.Key("direction")).Direction(direction).Children(func() { nui.Chart(c.Key("chart"), t, chart, o) })
	ui.Textf(c, "Focus changes: %d · Selections: %d · Legend changes: %d", s.Focuses, s.Selections, s.Legends)
	ui.Text(c, s.Last)
}

func main() {
	s := &example{}

	if err := s.Chart.SetCategories([]string{"Mon", "Tue", "Wed", "Thu", "Fri"}, []nui.ChartSeries{{ID: "views", Label: "Views", Values: []float64{25, 42, 35, 65, 80}}, {ID: "visits", Label: "Visits", Values: []float64{15, 24, 20, 45, 55}}}); err != nil {
		panic(err)
	}
	if err := s.Slices.SetSlices([]nui.ChartSlice{{ID: "go", Label: "Go", Value: 55}, {ID: "web", Label: "Web", Value: 30}, {ID: "docs", Label: "Docs", Value: 15}}); err != nil {
		panic(err)
	}
	s.Last = "No activation"
	palettes := nui.DefaultPalettes()
	palettes.Primary, _ = nui.PaletteByName("green")
	palettes.Neutral, _ = nui.PaletteByName("slate")
	config := nui.AppConfig{Mode: nui.LightMode, Light: nui.NewTheme(false, palettes), Dark: nui.NewTheme(true, palettes)}
	mygo.App.SetLanguages(nui.SupportedLocales()...)
	mygo.App.WhenReady(func() {
		mygo.App.SetLocale("en")
		mygo.NewWindow(mygo.WindowOptions{Title: "Chart · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
