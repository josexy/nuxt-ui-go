<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# InputDate

Edit date segments or choose a date from a calendar.

- Category: Forms and data
- Version: nuxt-ui-go v0.0.2 (MyGo v0.8.3)
- Functions: `InputDate`, `InputDateRange`
- Types: `InputDateOptions`
- Theme key: `nui.ComponentInputDate` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Bind an Optional[Date] and DateInputState for one date, or DateSelection and DateRangeInputState for a range. Incomplete or rejected drafts retain the last accepted value. Use OnValueChange for one date and OnRangeChange for a range; clearing both endpoints sends one range notification.

#### Try the controls

Change the day step and use Up/Down, or enable the calendar. Bounds restrict values to February–March 2024; the 20th can be unavailable. Try the custom separator, avatar, loader position and highlight.

## Things to know

Read-only fields still allow navigation and copying. ShowCalendar defaults to false and HideCalendar overrides it. Calendar range constraints also apply to segment edits. Calendar.OnChange remains a legacy range notification; for one date it reports calendar actions only. PlaceholderValue seeds empty segments; string Placeholder, Password and MaxLength do not apply. Values are full Gregorian dates without a time or time zone.

## API

```go
func InputDate(c *ui.Context, t Theme, value *Optional[Date], s *DateInputState, o InputDateOptions) ui.Element

func InputDateRange(c *ui.Context, t Theme, value *DateSelection, s *DateRangeInputState, o InputDateOptions) ui.Element
```

### InputDateOptions

```go
type InputDateOptions struct {
	InputOptions
	Step             DateStep
	PlaceholderValue Optional[Date]
	SeparatorIcon    *ui.SVG
	BuildSeparator   func(*ui.Context, Theme)
	// ShowCalendar adds a calendar button and popup. The default is a date field.
	ShowCalendar bool
	// HideCalendar overrides ShowCalendar; retained for existing callers.
	HideCalendar  bool
	Calendar      CalendarOptions
	OnValueChange func(Optional[Date])
	OnRangeChange func(DateSelection)
}
```

### Related types

```go
// Date is a civil Gregorian date; it has no timezone or time of day.
type Date struct {
	Year  int
	Month time.Month
	Day   int
}

func DateFromTime(v time.Time) Date

func (d Date) AddDays(n int) Date

func (d Date) AddMonths(n int) Date

func (d Date) Compare(other Date) int

func (d Date) String() string

func (d Date) Valid() bool
```

```go
type DateInputState struct {
	// Text is the canonical ISO draft (or rejected pasted text), independent
	// of the locale's visual segment order. Incomplete drafts retain the value.
	Text, Error string
	Calendar    CalendarState
	Open        bool
}
```

```go
type DateSelection struct{ Start, End Optional[Date] }

// SelectDate rejects ranges crossing a disabled day; reverse selection is normalized.
func SelectDate(value DateSelection, date Date, o CalendarOptions) (DateSelection, bool)
```

```go
// DateRangeInputState owns two segmented endpoints and an optional shared calendar.
type DateRangeInputState struct {
	Start, End DateInputState
	Calendar   CalendarState
	Open       bool
}
```

```go
// DateStep configures segment cycling. Zero uses one; negative steps are invalid.
type DateStep struct{ Year, Month, Day int }
```

```go
type CalendarState struct {
	Month, Active Date
	View          CalendarView
	Handle        ui.Handle
	// contains filtered or unexported fields
}
```

Types documented in other references: `CalendarOptions` → [calendar.md](calendar.md), `InputOptions` → [input.md](input.md), `Optional` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

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

##### InputDate: type a date

InputDate binds an `*Optional[Date]`, and DateInputState stores the segments, the draft, and the optional calendar state. Dates are Gregorian, and the segment order and separators follow the window's locale. PlaceholderValue provides a reference date for the first adjustment of an empty segment; it does not set the business value and does not replace the placeholder text.

| Action | Behavior |
| --- | --- |
| ←→, Tab / Shift+Tab | Move between segments. Arrow keys follow RTL; Tab enters or leaves the field |
| Typing digits | Advances to the next segment once the digit count or the maximum value is reached. Years take four digits |
| ↑↓, Home / End | Cycle by DateStep, or go to the segment's minimum or maximum. A step of 0 uses 1, and changes do not carry into neighboring segments |
| Backspace / Delete | Clear the current segment. After Ctrl/Cmd+A, deleting clears the current date endpoint |
| Ctrl/Cmd+C/X/V | Copy, cut, or paste the current endpoint; system edit commands also work |

Changing the year or month clamps a day past the end of the month to the last day of the target month. Pasting accepts ISO dates or numeric dates in the current locale's order. Calendar.Min/Max and IsDateDisabled/IsDateUnavailable restrict both typing and the calendar, and the bounds are inclusive. Outside assignments and option changes do not rewrite the model; if the current value violates a constraint, an error is shown.

DateInputState.Text stores an ISO draft, with `-` for unfilled segments. An invalid paste keeps the pasted text and the Error, while the segments on screen and the last valid business value stay unchanged. The business value is cleared only when every date segment is empty. OnChange reports the draft; OnValueChange fires once only when the valid single date, or the empty value, actually changes. Outside assignments are silent. Changing the language keeps the draft and focus. Read-only mode allows navigation and copying but not editing.

ShowCalendar defaults to false. When enabled, it adds a calendar button and a Popover that closes after a date is chosen. HideCalendar takes precedence over ShowCalendar. The Calendar's appearance, navigation, and part options apply to the popup, while Range is determined by which input function you call; a custom Label is kept. For a single date, Calendar.OnChange reports only calendar actions; segment editing is reported through OnValueChange. The popup closes if the input, its field, or an ancestor becomes disabled.

##### InputDateRange: type a start and end date

InputDateRange binds a `*DateSelection` and keeps a DateRangeInputState. Both ends share one frame, navigation can cross between them, and they wrap when space runs out. SeparatorIcon defaults to a minus sign. BuildSeparator can replace it synchronously with non-interactive content, and SeparatorSlot styles it.

Moving the start after the end clears the end, and the end cannot be earlier than the start. Range rules such as MaximumDays, unavailable dates, and AllowNonContiguousRanges also apply to segment input. If the start value set from outside is past the calendar's upper bound, an error is shown instead of producing conflicting calendar bounds.

OnRangeChange reports actual range changes caused by segment editing or the calendar. Clearable clears both ends at once with a single notification. Calendar.OnChange remains as a compatible range notification and is also called only once per change. Ranges do not call the single-date OnValueChange. OnChange still reports each endpoint's draft, so clearing or a calendar change to both ends can fire it once per end; OnClear reports the clear action as a whole. The popup closes after the second calendar endpoint is chosen.

Inside a single FormField, an error focuses the first segment of the start date. To register errors and focus for each end separately, combine two InputDates.

##### Date and time styles and limits

The date and time inputs share Input's five sizes, five variants, colors, avatar, icons, loading state, disabled and read-only states, and Clearable. A leading icon or leading spinner takes precedence over Avatar. LoadingTrailing controls the spinner position, and an explicit Leading keeps a leading spinner. With Neutral, focus and Highlight use the inverted background color, and Highlight also works with None. Shadow overrides on Root from the theme or the instance are kept, and Style runs last.

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
| InputDate, InputDateRange | ShowCalendar=false, and HideCalendar wins when both are set. Each DateStep field treats 0 as 1. The range separator defaults to a minus sign |

#### Limitations (excerpt)

Go Options fields and types follow this library's API; you cannot copy Nuxt Vue props directly. For date and time operations and migration, see Calendar, InputDate, and InputTime. For platform, IME, and screen reader verification.

### Inputs, selection, and fields

#### FormField and FieldGroup › FieldGroup: join several controls (excerpt)

Disabled is passed down to the items. Size comes from the instance first, then the ComponentFieldGroup default, then the enclosing FormField; if none is set, each child uses its own default size. Popups opened by child controls keep their own corner radius, such as SelectMenu's search field and InputDate's calendar.

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
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

type example struct {
	Value                         nui.Optional[nui.Date]
	Date                          nui.DateInputState
	Range                         nui.DateSelection
	RangeInput                    nui.DateRangeInputState
	Changes, RangeChanges, Clears int
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	var avatar *nui.AvatarOptions

	avatar = &nui.AvatarOptions{Text: "Ada"}

	calendar := nui.CalendarOptions{}

	o := nui.InputDateOptions{InputOptions: nui.InputOptions{Avatar: avatar, Label: "Date", LoadingTrailing: true, Clearable: true, OnClear: func() { s.Clears++ }}, Calendar: calendar, Step: nui.DateStep{Day: 1}, OnValueChange: func(nui.Optional[nui.Date]) { s.Changes++ }}
	r := nui.InputDateOptions{InputOptions: nui.InputOptions{Avatar: avatar, Label: "Date range", LoadingTrailing: true, Clearable: true, OnClear: func() { s.Clears++ }}, Calendar: calendar, Step: nui.DateStep{Day: 1}, OnRangeChange: func(nui.DateSelection) { s.RangeChanges++ }}
	nui.InputDate(c.Key("date"), t, &s.Value, &s.Date, o)
	nui.InputDateRange(c.Key("range"), t, &s.Range, &s.RangeInput, r)
	ui.Textf(c, "Date: %s · Changes: %d", s.Value.Value.String(), s.Changes)
	ui.Textf(c, "Range: %s to %s · Changes: %d · Clears: %d", s.Range.Start.Value.String(), s.Range.End.Value.String(), s.RangeChanges, s.Clears)
}

func main() {
	s := &example{}

	s.Value = nui.Set(nui.Date{Year: 2024, Month: time.February, Day: 15})
	palettes := nui.DefaultPalettes()
	palettes.Primary, _ = nui.PaletteByName("green")
	palettes.Neutral, _ = nui.PaletteByName("slate")
	config := nui.AppConfig{Mode: nui.LightMode, Light: nui.NewTheme(false, palettes), Dark: nui.NewTheme(true, palettes)}
	mygo.App.SetLanguages(nui.SupportedLocales()...)
	mygo.App.WhenReady(func() {
		mygo.App.SetLocale("en")
		mygo.NewWindow(mygo.WindowOptions{Title: "InputDate · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
