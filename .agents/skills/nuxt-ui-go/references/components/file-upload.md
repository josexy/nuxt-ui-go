<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# FileUpload

Select files and display their status and upload progress.

- Category: Advanced interaction and layout
- Version: nuxt-ui-go v0.0.2 plus unreleased changes (MyGo v0.9.0)
- Functions: `FileUpload`
- Types: `FileUploadOptions`, `UploadStatus`
- Theme key: `nui.ComponentFileUpload` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

FileUploadState holds a queue; the application implements transfers through Begin / Report. Rules checks type, size and count, retaining rejected rows with reasons. OnChange reports user selections or removals that change the queue; OnReject reports newly rejected files.

## Things to know

Choose returns simulated metadata: no native dialog, file content reads or uploads. Go keeps a queue; Dialog.Multiple controls one native dialog selection and Rules.MaxFiles limits accepted files in the queue. Position affects UploadArea / UploadList only. The native grid retains names, status and progress, unlike Nuxt’s image grid. The reading message and State.Error use persistent polite status regions; empty regions take no layout space. Adding or changing an exposed message reads it once. Clearing, unchanged rebuilds and hidden updates are silent. Mounting or exposing an existing message makes it accessible without announcing a change.

## API

```go
func FileUpload(c *ui.Context, t Theme, s *FileUploadState, o FileUploadOptions) ui.Element
```

### FileUploadOptions

```go
type FileUploadOptions struct {
	Label, Description, ChooseLabel            string
	Disabled, Highlight                        bool
	Icon, FileIcon, FileDeleteIcon             *ui.SVG
	ShowIcon, FileImage                        Optional[bool]
	FileDeleteOptions                          *ButtonOptions
	Dropzone, Interactive, FileDelete, Preview Optional[bool]
	Variant                                    UploadVariant
	Layout                                     UploadLayout
	Position                                   UploadPosition
	Color                                      Color
	Size                                       Size
	Rules                                      UploadRules
	Dialog                                     mygo.OpenDialogOptions
	// Choose overrides the native dialog. It executes on a worker and must not mutate UI state.
	Choose             func(context.Context) ([]UploadFile, error)
	OnChange           func([]UploadFile)
	OnReject           func([]UploadFile)
	OnCancel, OnRemove func(string)
	Decoration
}
```

### UploadStatus

```go
type UploadStatus uint8

const (
	UploadReady UploadStatus = iota
	UploadRunning
	UploadDone
	UploadFailed
	UploadCanceled
	UploadRejected
)
```

### Related types

```go
type FileUploadState struct {
	Files []UploadFile
	Error string
	// contains filtered or unexported fields
}

// Add validates metadata and adds rejected entries with their reason. Duplicate paths are ignored.
func (s *FileUploadState) Add(files []UploadFile, r UploadRules)

func (s *FileUploadState) Begin(id string) (context.Context, UploadTicket)

func (s *FileUploadState) Cancel(id string)

// Dispose cancels pending inspections and application upload contexts. Call on hiding/unmount/close.
func (s *FileUploadState) Dispose()

func (s *FileUploadState) Pending() bool

func (s *FileUploadState) Remove(id string)

// Report must run on the UI thread (e.g. Window.Update). Stale, canceled and removed jobs are ignored.
func (s *FileUploadState) Report(ticket UploadTicket, progress float64, done bool, err error) bool
```

```go
type UploadVariant uint8

const (
	UploadArea UploadVariant = iota
	UploadButton
)
```

```go
type UploadLayout uint8

const (
	UploadLayoutDefault UploadLayout = iota
	UploadList
	UploadGrid
)
```

```go
type UploadPosition uint8

const (
	UploadOutside UploadPosition = iota
	UploadInside
)
```

```go
type UploadRules struct {
	Accept   []string
	MaxBytes int64
	MaxFiles int
}

func (r UploadRules) Validate(f UploadFile) error
```

```go
// UploadFile owns metadata only. MIME is advisory (extension-derived by InspectUploadFiles).
// The application owns file handles, preview images, upload transports and cleanup.
type UploadFile struct {
	ID, Path, Name, MIME, Error string
	Avatar                      *AvatarOptions
	Size                        int64
	Progress                    float64
	Status                      UploadStatus
	// contains filtered or unexported fields
}

// InspectUploadFiles runs outside rendering, preferably on a worker. It never opens file content.
func InspectUploadFiles(ctx context.Context, paths []string) ([]UploadFile, error)

// PickUploadFiles uses the OS dialog; cancel returns an empty result. The OS dialog itself
// is not forcibly closed by context cancellation; its eventual result is ignored.
func PickUploadFiles(ctx context.Context, o mygo.OpenDialogOptions) ([]UploadFile, error)
```

```go
type UploadTicket struct {
	// contains filtered or unexported fields
}
```

Types documented in other references: `ButtonOptions` → [button.md](button.md), `Color` → [theme.md](theme.md), `Decoration` → [theme.md](theme.md), `Optional` → [theme.md](theme.md), `Size` → [theme.md](theme.md), `Theme` → [theme.md](theme.md), `Variant` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Advanced interaction, charts, and layout

#### FileUpload and resource ownership

##### The component handles selection and state; your app handles transfer

FileUpload provides file selection, rule checks, a file list, and progress display. Your app implements the network upload; selecting files does not start a transfer.

FileUploadState.Files stores file metadata, not open file handles. UploadFile contains Path, Name, MIME, Size, a status, a Progress from 0 to 1, Error, and an app-supplied Avatar. Add assigns stable IDs that do not collide with existing entries and ignores entries with an identical Path. If you assign Files directly, your app must provide unique IDs. The component does not load or decode preview images itself.

| Option | Purpose |
| --- | --- |
| UploadRules.Accept | Extensions, MIME types, or categories, such as `.png`, `text/plain`, or `image/*`. Supports `*`, `*/*`, comma-separated lists, compound extensions, and case-insensitive matching. An empty or whitespace-only value allows all types |
| MaxBytes / MaxFiles | Maximum size per file and maximum number of accepted files |
| Dialog | The native file picker's parent window, multiple selection, and filters |
| Choose(ctx) | Replaces the default file source |
| Variant / Layout / Position | Default UploadArea / UploadGrid / UploadOutside. UploadInside places only the UploadList inside the selection area |
| Icon / ShowIcon / Highlight | The upload entry icon, its visibility, and border highlighting. Neutral uses the inverted color, and a field error uses the error color first |
| FileIcon / FileImage / UploadFile.Avatar | The fallback file icon and the app-supplied preview accessory. FileImage=false shows only FileIcon |
| FileDelete / FileDeleteOptions / FileDeleteIcon | Whether the delete button is shown, its options, and its icon. Defaults to Neutral / LinkVariant / component size in the list, and Neutral / Solid / XS in the grid |
| Preview | Default true; false hides the file list but keeps the data and tasks. See [Sidebar, navigation, and file selection](sidebar.md) |

Files that exceed a limit or cannot be read are shown as Rejected with the reason, and the user can remove them. A non-positive MaxBytes or MaxFiles means no limit. Rejected files do not count toward the limit, while accepted files that later fail or are canceled still count. The component inherits Disabled from its parent container and FormField, as well as FormField's Size, label, and error message. The default color and size come from ComponentFileUpload.

OnChange fires once, with a copy of the list, only when a selection or removal actually changes the list. Duplicate paths, a canceled selection, a failed selection, and direct Add, Remove, or Report calls from your app do not fire it. OnReject reports the newly rejected files before OnChange and does not clear rejected entries from the queue. After a removal or a canceled transfer, focus returns to the file selection entry point.

Go keeps an upload queue model. Dialog.Multiple only controls whether one native dialog accepts several files; you cannot use it to limit the queue size. Nuxt UI's single-file modelValue replacement, `multiple` forcing the area variant, and the web input reset have no direct equivalent. The grid shows file names, status, and progress in wrapping cards and does not claim to match upstream's image grid exactly.

##### How files are checked

By default, the selection action calls `mygo.Dialog.Open` in the background and then checks for regular files with `os.Stat`. External DroppedFiles go through the same checks and rules. You can also call InspectUploadFiles and PickUploadFiles directly.

MIME types are derived only from the extension; file contents are not read. Accept checks selected and dropped files but is not converted into native Dialog.Filters automatically. If your business needs a trustworthy file type, validate the content yourself. Dropzone=false registers no drop target, and Interactive=false keeps a separate selection button. When Choose returns an empty result, the selection was canceled; when it returns an error, State.Error is set and the file queue is unchanged.

The "reading file information" message and State.Error use persistent polite status regions. Regions without text take no layout space. While the component is exposed to assistive technology, adding or changing a message reads it once; clearing it and unchanged rebuilds read nothing. Hidden parents, including inactive kept-mounted tab panels and AccessHidden containers, suppress announcements. Mounting or exposing a region with an existing message makes it accessible without announcing it as a change. Per-file errors in the list are not live regions.

##### Start, cancel, and retry uploads

```go
// After creating the window, set the parent window and register cleanup.
options.Dialog = mygo.OpenDialogOptions{Parent: window, Multiple: true}
window.OnClosed(state.Files.Dispose)

// Start the task in your app's upload action; your app implements upload.
ctx, ticket := state.Files.Begin(fileID)
go func() {
    err := upload(ctx, path)
    window.Update(func() {
        state.Files.Report(ticket, 1, true, err)
    })
}()
```

Begin starts a task, and your app publishes progress, completion, or errors with Report in Window.Update. Background code must not change Files directly. The UI shows Ready, Uploading, Complete, Failed, Canceled, and Rejected. After a failure, you can call Begin again; the library has no additional retry protocol.

Cancel(id) cancels the task's context, and Remove(id) cancels and removes the file. Old tickets from before a cancel, removal, or retry, or from a task that has already completed, can no longer write back. OnCancel and OnRemove can perform extra cleanup. Your app also closes any file streams, temporary files, and network connections it opened.

##### Cleaning up on hide and close

Call Dispose when the component is hidden or unmounted, or when the window closes, to cancel file checks and upload tasks. Disabling the field also cancels pending file checks and discards late results.

Canceling the context cannot force an open native file dialog to close, but its result is discarded. For platform verification of native selection, multiple selection, cancellation, drag and drop, and filters. Your app manages macOS sandbox permissions and bookmarks.

#### State and part index (excerpt)

| Component | Persistent state or data | Parts |
| --- | --- | --- |
| FileUpload | FileUploadState | Root, Body, Leading, Label, Description, Trigger, List, Item, ItemLeadingSlot, ItemLabelSlot, StatusSlot, ErrorSlot, ItemTrailingSlot |

### Options and defaults

#### Unset, zero, and default values › Frequently checked defaults (excerpt)

| Component | Defaults |
| --- | --- |
| Carousel, FileUpload | Carousel: Arrows=false. FileUpload: Layout=UploadGrid |

FileUpload's `Dropzone`, `Interactive`, `FileDelete`, and `Preview` all default to true. The zero value of `UploadLayout` is `UploadLayoutDefault`; use the enum names in options rather than relying on the underlying numbers.

#### Common options by component › Sidebar, navigation, and file selection (excerpt)

| Component | Common options and behavior |
| --- | --- |
| FileUpload | Variant is UploadArea or UploadButton. Layout defaults to UploadGrid. Position affects only the list inside the area. ShowIcon, FileImage, FileDelete, and Preview default to true. Supports icon and delete button configuration and Highlight |

Clicking an interactive UploadArea, or pressing Enter or Space on it, opens the file picker; its accessible name comes from Label first. In the built-in list, file actions are separate from the title/icon entry point so native assistive technology can reach each action. The Choose button appears only with UploadButton or `Interactive: nui.Set(false)`. Dropzone=false removes the drop target. Preview=false hides the whole list, and FileImage=false hides only the app-supplied preview accessories; the Files data is kept in both cases. OnChange fires when a selection or removal changes the queue, and OnReject reports newly rejected files. After a removal, focus returns to the selection entry point. For full defaults and responsibilities, see FileUpload. For Sidebar's region builders, sizes, and keyboard controls, see [Sidebar state and size](sidebar.md).

### Compatibility and known limitations

#### Feature scope › Data and files (excerpt)

| Area | Supported | Not provided, or left to the app |
| --- | --- | --- |

#### Accessibility and motion (excerpt)

Since MyGo v0.9.0, `ui.RoleStatus` is a polite, atomic live region. The library's own progress, error, empty, and loading messages keep stable identities: such a status is read once when its text changes and is silent when it first appears, clears, or rebuilds unchanged. Alert and Banner take their identity from the key the app gives them. CommandPalette also announces its progress and error messages once when they appear. FileUpload keeps empty status regions mounted, so adding a message to an exposed region is read once; hidden updates are silent, and mounting or exposing an existing message does not announce it as a change. Chart announces keyboard selection once and never reads pointer hover. Alert and Banner read their new title and description once when either changes. These rules are checked headlessly with `Tester.UseScreenReader`; what VoiceOver, NVDA, Narrator, and Orca actually say has not been verified.

## Complete example

A runnable program. App state lives in `example`, outside the view, and the view only builds UI.

```go
package main

import (
	"context"
	"errors"
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

type example struct {
	Files             nui.FileUploadState
	Changes, Rejected int
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	outcome := "files"
	o := nui.FileUploadOptions{Label: "Project files", Description: "Sample metadata; no file reads or network upload", Interactive: nui.Set(true), Dropzone: nui.Set(true), Preview: nui.Set(true), FileImage: nui.Set(true), ShowIcon: nui.Set(true), Variant: nui.UploadArea, Layout: nui.UploadList, Position: nui.UploadOutside, Rules: nui.UploadRules{Accept: []string{".md, .txt, .png"}, MaxFiles: 3, MaxBytes: 2 << 20}, Choose: func(context.Context) ([]nui.UploadFile, error) {
		if outcome == "cancel" {
			return nil, nil
		}
		if outcome == "error" {
			return nil, errors.New("sample selection failed")
		}
		return []nui.UploadFile{{Path: "sample/notes.txt", Name: "notes.txt", Size: 512}, {Path: "sample/application.exe", Name: "application.exe", Size: 1024}}, nil
	}, OnChange: func([]nui.UploadFile) { s.Changes++ }, OnReject: func(files []nui.UploadFile) { s.Rejected += len(files) }}

	nui.FileUpload(c.Key("files"), t, &s.Files, o)
	ui.Textf(c, "Changes: %d · Rejected: %d", s.Changes, s.Rejected)
}

func main() {
	s := &example{}

	s.Files.Files = []nui.UploadFile{{ID: "readme", Path: "sample/README.md", Name: "README.md", Size: 2048, Status: nui.UploadDone, Avatar: &nui.AvatarOptions{Text: "MD", Color: nui.Info}}, {ID: "design", Path: "sample/design.png", Name: "design.png", Size: 184320, Status: nui.UploadReady}}
	defer s.Files.Dispose()
	palettes := nui.DefaultPalettes()
	palettes.Primary, _ = nui.PaletteByName("green")
	palettes.Neutral, _ = nui.PaletteByName("slate")
	config := nui.AppConfig{Mode: nui.LightMode, Light: nui.NewTheme(false, palettes), Dark: nui.NewTheme(true, palettes)}
	mygo.App.SetLanguages(nui.SupportedLocales()...)
	mygo.App.WhenReady(func() {
		mygo.App.SetLocale("en")
		mygo.NewWindow(mygo.WindowOptions{Title: "FileUpload · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
