<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Calendar

Choose a date or a date range from a calendar.

- Category: Forms and data
- Version: nuxt-ui-go v0.0.2 plus unreleased changes (MyGo v0.9.0)
- Functions: `Calendar`
- Types: `CalendarOptions`
- Theme key: `nui.ComponentCalendar` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Calendar binds DateSelection. CalendarState keeps the first displayed month, active date, navigation view and persistent focus. Selecting a date again clears it by default; PreventDeselect disables this. ReadOnly retains browsing and focus without changing the value.

## Things to know

This example fixes today and its initial value at February 15, 2024 and starts weeks on Monday. The component retains its native Sunday WeekStart default; week numbers follow the explicit setting. Unavailable dates can be browsed but not selected as endpoints. AllowNonContiguousRanges permits unavailable interior dates, not disabled dates. Month/year views navigate; values remain Gregorian dates or ranges.

## API

```go
func Calendar(c *ui.Context, t Theme, value *DateSelection, s *CalendarState, o CalendarOptions) ui.Element
```

### CalendarOptions

```go
type CalendarOptions struct {
	Label                                                            string
	Range, Disabled                                                  bool
	ReadOnly                                                         bool
	Color                                                            Color
	Variant                                                          Variant
	Size                                                             Size
	Min, Max                                                         Optional[Date]
	IsDateDisabled                                                   func(Date) bool
	IsDateUnavailable                                                func(Date) bool
	IsDateHighlightable                                              func(Date) bool
	AllowNonContiguousRanges                                         bool
	PreventDeselect                                                  bool
	MaximumDays                                                      int // Inclusive range length; 0 is unlimited.
	FixedDate                                                        CalendarFixedDate
	WeekStart                                                        time.Weekday
	WeekdayFormat                                                    CalendarWeekdayFormat
	WeekNumbers                                                      bool
	FixedWeeks                                                       Optional[bool]
	NumberOfMonths                                                   int // 0 uses one month; at most 12.
	PagedNavigation                                                  bool
	PreviousPage, NextPage                                           func(Date) Date
	DisableOutsideDays                                               bool
	MonthControls, YearControls                                      Optional[bool]
	ViewControl                                                      Optional[bool]
	ViewControlOptions                                               *ButtonOptions
	YearsPerPage                                                     int // 0 uses 12; at most 120.
	IsMonthDisabled, IsMonthUnavailable                              func(Date) bool
	IsYearDisabled, IsYearUnavailable                                func(Date) bool
	PreviousMonth, NextMonth, PreviousYear, NextYear                 *ButtonOptions
	PreviousMonthIcon, NextMonthIcon, PreviousYearIcon, NextYearIcon *ui.SVG
	Today                                                            Optional[Date]
	OnChange                                                         func(DateSelection)
	OnMonthChange                                                    func(Date)
	OnViewChange                                                     func(CalendarView)
	BuildDay, BuildMonth, BuildYear                                  func(*ui.Context, Theme, Date)
	BuildWeekday                                                     func(*ui.Context, Theme, time.Weekday)
	BuildHeading                                                     func(*ui.Context, Theme, Date, CalendarView)
	Decoration
}
```

### Related types

```go
type DateSelection struct{ Start, End Optional[Date] }

// SelectDate rejects ranges crossing a disabled day; reverse selection is normalized.
func SelectDate(value DateSelection, date Date, o CalendarOptions) (DateSelection, bool)
```

```go
type CalendarState struct {
	Month, Active Date
	View          CalendarView
	Handle        ui.Handle
	// contains filtered or unexported fields
}
```

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
type CalendarFixedDate uint8

const (
	CalendarNoFixedDate CalendarFixedDate = iota
	CalendarFixedStart
	CalendarFixedEnd
)
```

```go
type CalendarWeekdayFormat uint8

const (
	CalendarWeekdayNarrow CalendarWeekdayFormat = iota
	CalendarWeekdayShort
	CalendarWeekdayLong
)
```

```go
type CalendarView uint8

const (
	CalendarDayView CalendarView = iota
	CalendarMonthView
	CalendarYearView
)
```

Types documented in other references: `ButtonOptions` → [button.md](button.md), `Color` → [theme.md](theme.md), `Decoration` → [theme.md](theme.md), `Optional` → [theme.md](theme.md), `Size` → [theme.md](theme.md), `Theme` → [theme.md](theme.md), `Variant` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Forms, dates, and data components

#### Calendar, InputDate, and InputTime

##### Choose a value type

| Type | Represents |
| --- | --- |
| `Date{Year, Month, Day}` | A Gregorian date in the years 1–9999, without a time zone or time of day |
| `Optional[Date]` | A single date that can be empty |
| `DateSelection{Start, End}` | A single date or a date range; both ends are Optional[Date] |
| `ClockTime{Hour, Minute, Second}` | A time without a date or time zone; Hour is 0–23 |

Date provides Valid, Compare, AddDays, AddMonths, String, and ParseDate. When AddMonths lands on a day that does not exist in the target month, it uses the last day of that month. DateFromTime reads the year, month, and day in the Location of the given time.Time without converting to UTC. To represent a specific instant, your app chooses the time zone and handles daylight saving time.

##### Calendar: pick a date by clicking

Calendar binds a `*DateSelection`. CalendarState stores the first displayed month (Month), the keyboard's active date (Active), the navigation view (View), and a persistent Handle. With Range=false, only Start is used. With Range=true, the user picks the start and then the end, and a reversed selection is put in order. By default, selecting the same single date, or the start of an incomplete range, again deselects it; PreventDeselect=true keeps the value instead. Enable it too if users need to pick a range that starts and ends on the same day.

ReadOnly allows focusing, browsing, and paging but prevents the user from changing the value. OnChange fires only when the value actually changes, not for outside assignments or rebuilds caused by option changes. OnMonthChange and OnViewChange report changes to the first month and the navigation view caused by the user; changing the state directly from outside is silent.

| Option | Behavior |
| --- | --- |
| Min / Max / IsDateDisabled | Restrict dates. Arrow keys skip disabled dates, and range ends must be selectable |
| IsDateUnavailable | Shows a strikethrough. The keyboard can move onto it, but it cannot be chosen as a range end |
| AllowNonContiguousRanges | Default false; true allows a range to span unavailable dates |
| IsDateHighlightable | Explicitly allows a range to span the given inner dates. Does not lift the restriction on range ends |
| MaximumDays | Maximum range length in days, counting both ends; 0 means no limit. Candidates beyond the limit are disabled |
| FixedDate | Fixes the start or end of a complete range. Moving past the fixed end updates that end and keeps the dates in order |
| Color / Variant / Size | Default Primary / Solid / MD. Supports Solid, Outline, Soft, and Subtle. The size is inherited from the field or the component theme |
| WeekStart / WeekdayFormat | WeekStart keeps the native default of Sunday. Weekdays use narrow names by default, and abbreviated or full names are also available |
| FixedWeeks | Default true, which always shows six weeks; false shows four to six weeks depending on the month |
| NumberOfMonths / PagedNavigation | One month by default, up to 12, wrapping when space runs out. PagedNavigation makes the month buttons page by the number of months shown |
| MonthControls / YearControls | Default true, which shows the month and year buttons in the day view. Buttons and icons can be configured individually |
| ViewControl / ViewControlOptions | The heading button is on by default and switches between the day, month, and year views. Month and year cells only navigate; they do not commit a date |
| YearsPerPage | The year view shows 12 years per page by default, from 1 to 120. IsMonthDisabled/Unavailable and IsYearDisabled/Unavailable configure the matching views |
| PreviousPage / NextPage | Custom target dates for paging back and forward. Used by the month buttons and the paging buttons of the month and year views; does not change the separate previous/next year buttons or the keyboard rules |
| Today | Pins today's date, which is useful for examples and tests. When unset, the current local date is used |

| Key | Behavior in the day view |
| --- | --- |
| ←→ / ↑↓ | Move by day / by week. Left and Right follow RTL; disabled dates are skipped and bounds are respected |
| PageUp / PageDown | Previous / next month, keeping the day of the month where possible |
| Shift+PageUp / PageDown | Previous / next year |
| Home / End | First / last available date of the current week |
| Enter / Space | Select or deselect the current date. Read-only, disabled, or unavailable dates do not change the value |

The month and year views use a four-column grid. Arrow keys move by column or row, and PageUp/PageDown move by year or by page of years. Choosing a cell returns to the next view down without changing the date value. This navigation also works in read-only mode.

The day, month, and year views are exposed to assistive technology as grids (`ui.RoleGrid`), as Reka UI's `role="grid"`, named by Label. Rows hold grid cells named by Label and date, and the day view has weekday column headers. The active date is the grid's active descendant, and the chosen dates report selected. With ReadOnly, the grid and its cells report read-only while arrow keys still move the active date. What real screen readers say for this grid, and whether they switch to focus mode on it, has not been verified.

WeekNumbers are computed from the explicit WeekStart. With Sunday, week 1 is the week containing January 1; with any other start day, it is the week containing January 4, so Monday gives ISO week numbers. This differs from Nuxt UI, which derives the week start and week numbers from the locale. Changing the language updates the heading, weekdays, and button names but does not change an explicit week setting.

Month is normalized to the first of the month. When the selected start or end changes from outside, the view follows the valid date; clearing the selection keeps the current view. Changing the number of months keeps the active date visible, and at the end of the year 9999 only the remaining months are shown. Multiple months share one native keyboard focus and wrap to fit narrow containers.

BuildDay, BuildMonth, and BuildYear build cell content synchronously, and BuildWeekday configures the weekday content. The container still provides the default names, selection, disabled state, and focus. Keep cell content non-interactive; the calendar determines its size. BuildHeading replaces the default heading and receives the current month and view; custom controls manage their state through the persistent CalendarState. Builders and date predicates may be called many times, so never perform I/O or submit anything from them.

Dates and ranges are always complete Gregorian dates. There is no discrete multi-date model and no separate whole-month or whole-year value model. The month and year navigation views do not implement Nuxt UI's `type="month"` or `type="year"`; manage those values with app controls such as Select or InputNumber. Other calendars, time zones, and browser DOM slots are outside this native model.

##### Date and time styles and limits

The date and time inputs share Input's five sizes, five variants, colors, avatar, icons, loading state, disabled and read-only states, and Clearable. A leading icon or leading spinner takes precedence over Avatar. LoadingTrailing controls the spinner position, and an explicit Leading keeps a leading spinner. With Neutral, focus and Highlight use the inverted background color, and Highlight also works with None. Shadow overrides on Root from the theme or the instance are kept, and Style runs last.

Each date or time segment is a stepper for assistive technology, named by the field label and the unit, such as "Due month", with its displayed text as the value. With ReadOnly, every segment reports read-only, so screen readers neither set nor step it; editable segments do not.

Root styles the frame, Editor sets the base style of the segment text, Segment styles placeholders, the focus background, and corner radius, and Leading and Trailing style the accessories. Style's editor is the first focusable segment in the current order and can bind a Handle. It is not a TextInputBase, so text selection APIs do not apply. Placeholder, Password, and MaxLength do not apply to segmented input; placeholder text comes from the language catalog.

Date represents a complete Gregorian date and ClockTime a single time of day with second precision; neither carries another calendar, milliseconds, or a time zone. Dates and times are edited separately. For a time range, combine two InputTimes and validate them in your app. Segmented input has no undo/redo, and the full screen reader experience has not been verified. Calendar's month and year views are for navigation and do not change the date value type.

#### Style part index (excerpt)

| Component | Parts |
| --- | --- |
| Calendar | Root, Header, HeadingSlot, Body, GridSlot, GridRowSlot, WeekDaySlot, WeekNumberSlot, Item |

### Options and defaults

#### Unset, zero, and default values › Frequently checked defaults (excerpt)

| Component | Defaults |
| --- | --- |
| Calendar | Color=Primary, Variant=Solid, Size=MD. FixedWeeks, MonthControls, YearControls, and ViewControl=true. NumberOfMonths=1, YearsPerPage=12 |

#### Limitations (excerpt)

Go Options fields and types follow this library's API; you cannot copy Nuxt Vue props directly. For date and time operations and migration, see Calendar, InputDate, and InputTime. For platform, IME, and screen reader verification.

### Compatibility and known limitations

#### Feature scope › Data and files (excerpt)

| Area | Supported | Not provided, or left to the app |
| --- | --- | --- |
| Calendar | Gregorian single dates and ranges, day/month/year navigation, wrapping multi-month layout, range limits | No discrete multi-date selection and no separate year-month value model. WeekStart defaults to Sunday. Week numbers are computed from the explicit week start; full Intl locale week rules are not guaranteed |

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
	nui "github.com/josexy/nuxt-ui-go/ui"
)

type example struct {
	Value                  nui.DateSelection
	Calendar               nui.CalendarState
	Changes, Months, Views int
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	o := nui.CalendarOptions{Label: "Date", WeekStart: time.Monday, NumberOfMonths: 1, Today: nui.Set(nui.Date{Year: 2024, Month: time.February, Day: 15}), OnChange: func(nui.DateSelection) { s.Changes++ }, OnMonthChange: func(nui.Date) { s.Months++ }, OnViewChange: func(nui.CalendarView) { s.Views++ }}

	o.IsDateUnavailable = func(date nui.Date) bool { return date.Day == 20 }

	nui.Calendar(c.Key("calendar"), t, &s.Value, &s.Calendar, o)
	ui.Row(c).Wrap().Gap(8).Children(func() {
		nui.Button(c.Key("reset"), t, "Reset example date", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() {
			s.Value = nui.DateSelection{Start: nui.Set(nui.Date{Year: 2024, Month: time.February, Day: 15})}
			s.Calendar.Month = nui.Date{Year: 2024, Month: time.February, Day: 1}
			s.Calendar.View = nui.CalendarDayView
		})
		nui.Button(c.Key("clear"), t, "Clear example date", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { s.Value = nui.DateSelection{} })
	})
	ui.Textf(c, "Date: %s to %s · Changes: %d", s.Value.Start.Value.String(), s.Value.End.Value.String(), s.Changes)
	ui.Textf(c, "Month: %s · Month changes: %d · View changes: %d", s.Calendar.Month, s.Months, s.Views)
}

func main() {
	s := &example{}

	s.Value.Start = nui.Set(nui.Date{Year: 2024, Month: time.February, Day: 15})
	s.Calendar.Month = nui.Date{Year: 2024, Month: time.February, Day: 1}
	palettes := nui.DefaultPalettes()
	palettes.Primary, _ = nui.PaletteByName("green")
	palettes.Neutral, _ = nui.PaletteByName("slate")
	config := nui.AppConfig{Mode: nui.LightMode, Light: nui.NewTheme(false, palettes), Dark: nui.NewTheme(true, palettes)}
	mygo.App.SetLanguages(nui.SupportedLocales()...)
	mygo.App.WhenReady(func() {
		mygo.App.SetLocale("en")
		mygo.NewWindow(mygo.WindowOptions{Title: "Calendar · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
