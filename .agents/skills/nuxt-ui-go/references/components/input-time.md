<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# InputTime

Edit hours, minutes and seconds in 12- or 24-hour format.

- Category: Forms and data
- Version: nuxt-ui-go v0.0.2 plus unreleased changes (MyGo v0.9.0)
- Functions: `InputTime`
- Types: `InputTimeOptions`
- Theme key: `nui.ComponentInputTime` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Bind Optional[ClockTime] and TimeInputState. Unset TwelveHour follows the regional hour cycle; explicit false selects 24-hour time. Granularity chooses hours, minutes or seconds; when unset, Seconds determines the precision.

#### Try the controls

Minutes cycle by the configured step. With StepSnapping, leaving an edited segment rounds it to the nearest step. Changing precision preserves hidden units and moves focus to a visible segment without notifying a value change. The optional valid interval is 09:00–17:30.

## Things to know

Out-of-range or unavailable edits remain drafts without replacing the accepted value. Min/Max are inclusive and compare the full time, including hidden units. Hour-only input uses HH or hh AM/PM; other precisions use HH:MM[:SS]. The model represents one time of day; compose two InputTime fields with application validation for a range. Dates, milliseconds and time zones are not included.

## API

```go
func InputTime(c *ui.Context, t Theme, value *Optional[ClockTime], s *TimeInputState, o InputTimeOptions) ui.Element
```

### InputTimeOptions

```go
type InputTimeOptions struct {
	InputOptions
	// TwelveHour defaults to the locale and the user's regional hour cycle.
	TwelveHour Optional[bool]
	Seconds    bool
	// Granularity overrides Seconds when set. Hidden units retain their value.
	Granularity       TimeGranularity
	Min, Max          Optional[ClockTime]
	PlaceholderValue  Optional[ClockTime]
	Step              TimeStep
	StepSnapping      bool
	IsTimeUnavailable func(ClockTime) bool
	OnValueChange     func(Optional[ClockTime])
}
```

### Related types

```go
// ClockTime is a wall-clock value, not an instant. Conversion to a timezone is application-owned.
type ClockTime struct{ Hour, Minute, Second int }

func (v ClockTime) String() string

func (v ClockTime) Valid() bool
```

```go
type TimeInputState struct {
	Text, Error string
	// contains filtered or unexported fields
}
```

```go
type TimeGranularity uint8

const (
	TimeGranularityDefault TimeGranularity = iota
	TimeHour
	TimeMinute
	TimeSecond
)
```

```go
// TimeStep configures segment cycling. Zero uses one; negative steps are invalid.
type TimeStep struct{ Hour, Minute, Second int }
```

Types documented in other references: `InputOptions` → [input.md](input.md), `Optional` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Forms, dates, and data components

#### Lifecycle and themes (excerpt)

InputNumber, InputDate, and InputTime share Input's sizes, five field variants, colors, read-only and disabled states, icons, and error settings. Other components are configured through their own Options at the end of this page.

#### Calendar, InputDate, and InputTime

##### Choose a value type

| Type | Represents |
| --- | --- |
| `Date{Year, Month, Day}` | A Gregorian date in the years 1–9999, without a time zone or time of day |
| `Optional[Date]` | A single date that can be empty |
| `DateSelection{Start, End}` | A single date or a date range; both ends are Optional[Date] |
| `ClockTime{Hour, Minute, Second}` | A time without a date or time zone; Hour is 0–23 |

Date provides Valid, Compare, AddDays, AddMonths, String, and ParseDate. When AddMonths lands on a day that does not exist in the target month, it uses the last day of that month. DateFromTime reads the year, month, and day in the Location of the given time.Time without converting to UTC. To represent a specific instant, your app chooses the time zone and handles daylight saving time.

##### InputTime: type a time

InputTime binds an `*Optional[ClockTime]`, and TimeInputState stores the draft. When TwelveHour is unset, the window locale's clock is used; `nui.Set(true)` and `nui.Set(false)` fix it to 12-hour and 24-hour respectively. Up and Down cycle within the current segment. When the hour crosses noon or midnight, AM/PM updates, and A or P selects the period directly.

Granularity can be TimeHour, TimeMinute, or TimeSecond. When unset, minutes are shown by default, and Seconds=true adds seconds; an explicit Granularity takes precedence. Changing the clock or granularity does not rewrite the business value or fire notifications, and hidden units are kept. If the focused minute or second segment is hidden, focus moves to a visible segment; if the AM/PM segment is hidden, focus moves to the hour.

TimeStep sets the step for hours, minutes, and seconds separately. 0 uses 1, negative values are invalid, and changes do not carry into other segments. StepSnapping defaults to false. When enabled, leaving a non-empty segment rounds it to the nearest valid step counted from 0, rounding up when exactly halfway; in 12-hour mode, the rounding uses the full 24-hour time. No snapping happens when read-only or disabled, for rejected drafts, or when only display options change.

Min and Max are inclusive bounds on the complete time and include the hidden minutes and seconds in the comparison. Reversed bounds for a range that crosses midnight are not supported. IsTimeUnavailable can reject specific times. Invalid input keeps the draft and the last valid value; outside assignments are silent, and an out-of-range outside value shows an error. PlaceholderValue provides a reference value for the first adjustment of an empty segment, and at hour granularity, a new value's hidden units can also come from it.

The Text and paste formats are `HH` and `HH:MM[:SS]`, with ` AM` or ` PM` appended in 12-hour mode. The displayed text and segment order use the locale format. Currently, only ASCII digits and the canonical time text above are accepted; localized AM/PM words are not parsed.

##### Date and time styles and limits

The date and time inputs share Input's five sizes, five variants, colors, avatar, icons, loading state, disabled and read-only states, and Clearable. A leading icon or leading spinner takes precedence over Avatar. LoadingTrailing controls the spinner position, and an explicit Leading keeps a leading spinner. With Neutral, focus and Highlight use the inverted background color, and Highlight also works with None. Shadow overrides on Root from the theme or the instance are kept, and Style runs last.

Each date or time segment is a stepper for assistive technology, named by the field label and the unit, such as "Due month", with its displayed text as the value. With ReadOnly, every segment reports read-only, so screen readers neither set nor step it; editable segments do not.

Root styles the frame, Editor sets the base style of the segment text, Segment styles placeholders, the focus background, and corner radius, and Leading and Trailing style the accessories. Style's editor is the first focusable segment in the current order and can bind a Handle. It is not a TextInputBase, so text selection APIs do not apply. Placeholder, Password, and MaxLength do not apply to segmented input; placeholder text comes from the language catalog.

Date represents a complete Gregorian date and ClockTime a single time of day with second precision; neither carries another calendar, milliseconds, or a time zone. Dates and times are edited separately. For a time range, combine two InputTimes and validate them in your app. Segmented input has no undo/redo, and the full screen reader experience has not been verified. Calendar's month and year views are for navigation and do not change the date value type.

#### Style part index (excerpt)

| Component | Parts |
| --- | --- |
| InputDate / InputDateRange / InputTime | Root, Editor, Segment, Leading, Trailing. The date range also has SeparatorSlot |

### Options and defaults

#### Unset, zero, and default values › Frequently checked defaults (excerpt)

| Component | Defaults |
| --- | --- |
| InputTime | An unset TwelveHour follows the locale's clock. An unset Granularity is derived from Seconds and defaults to minutes. Each TimeStep field treats 0 as 1. StepSnapping=false |

#### Limitations (excerpt)

Go Options fields and types follow this library's API; you cannot copy Nuxt Vue props directly. For date and time operations and migration, see Calendar, InputDate, and InputTime. For platform, IME, and screen reader verification.

### Compatibility and known limitations

#### Canceling IME composition (excerpt)

`PinInputOptions{Numeric: true}`, InputNumber, InputDate, and InputTime accept only numeric or date/time formats and do not enable IME. Other input methods still need testing on each target platform.

### Internationalization

#### Formats and language coverage › Dates and times (excerpt)

Calendar uses MyGo's date formatting for months, years, and weekdays. InputDate and InputTime use MyGo's CLDR locale formats to order the date or time segments, choose separators, and show AM/PM text in the current language. Digits are always ASCII, for both input and display.

## Complete example

A runnable program. App state lives in `example`, outside the view, and the view only builds UI.

```go
package main

import (
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

type example struct {
	Value           nui.Optional[nui.ClockTime]
	Time            nui.TimeInputState
	Changes, Clears int
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	var avatar *nui.AvatarOptions

	avatar = &nui.AvatarOptions{Text: "Ada"}

	var minimum, maximum nui.Optional[nui.ClockTime]

	o := nui.InputTimeOptions{Min: minimum, Max: maximum, InputOptions: nui.InputOptions{Avatar: avatar, Label: "Time", LoadingTrailing: true, Clearable: true, OnClear: func() { s.Clears++ }}, Granularity: nui.TimeGranularityDefault, Step: nui.TimeStep{Minute: 5}, StepSnapping: true, OnValueChange: func(nui.Optional[nui.ClockTime]) { s.Changes++ }}
	nui.InputTime(c.Key("time"), t, &s.Value, &s.Time, o)
	ui.Textf(c, "Time: %s · Set: %t · Changes: %d · Clears: %d", s.Value.Value.String(), s.Value.Set, s.Changes, s.Clears)
	ui.Textf(c, "Draft: %s · Error: %s", s.Time.Text, s.Time.Error)
}

func main() {
	s := &example{}

	s.Value = nui.Set(nui.ClockTime{Hour: 11, Minute: 55, Second: 42})
	palettes := nui.DefaultPalettes()
	palettes.Primary, _ = nui.PaletteByName("green")
	palettes.Neutral, _ = nui.PaletteByName("slate")
	config := nui.AppConfig{Mode: nui.LightMode, Light: nui.NewTheme(false, palettes), Dark: nui.NewTheme(true, palettes)}
	mygo.App.SetLanguages(nui.SupportedLocales()...)
	mygo.App.WhenReady(func() {
		mygo.App.SetLocale("en")
		mygo.NewWindow(mygo.WindowOptions{Title: "InputTime · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
