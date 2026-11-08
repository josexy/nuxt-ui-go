package nui

import (
	"context"
	"fmt"
	"math"
	"mime"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

type UploadStatus uint8

const (
	UploadReady UploadStatus = iota
	UploadRunning
	UploadDone
	UploadFailed
	UploadCanceled
	UploadRejected
)

// formatFileSize is Nuxt's formatFileSize: 1024-based units, whole numbers,
// no space: "0B", "512B", "12KB", "3MB".
func formatFileSize(bytes int64) string {
	if bytes <= 0 {
		return "0B"
	}
	units := []string{"B", "KB", "MB", "GB", "TB"}
	i := min(int(math.Floor(math.Log(float64(bytes))/math.Log(1024))), len(units)-1)
	value := float64(bytes) / math.Pow(1024, float64(i))
	if i == 0 {
		return strconv.FormatInt(bytes, 10) + units[0]
	}
	return strconv.FormatFloat(math.Round(value), 'f', 0, 64) + units[i]
}

// UploadFile owns metadata only. MIME is advisory (extension-derived by InspectUploadFiles).
// The application owns file handles, preview images, upload transports and cleanup.
type UploadFile struct {
	ID, Path, Name, MIME, Error string
	Avatar                      *AvatarOptions
	Size                        int64
	Progress                    float64
	Status                      UploadStatus
	validation                  *localizedError
	validationText              string
}

func (f *UploadFile) setError(err error) {
	f.validation, _ = err.(*localizedError)
	f.validationText = ""
	f.Error = err.Error()
	if f.validation != nil {
		f.validationText = f.Error
	}
}

func (f *UploadFile) localizeError(locale string) {
	if f.validation == nil {
		return
	}
	if f.Error != f.validationText {
		// The caller replaced the error; never translate caller-owned text.
		f.validation, f.validationText = nil, ""
		return
	}
	f.Error = f.validation.in(locale)
	f.validationText = f.Error
}

type UploadRules struct {
	Accept   []string
	MaxBytes int64
	MaxFiles int
}

func (r UploadRules) Validate(f UploadFile) error {
	if f.Error != "" {
		if f.validation != nil && f.Error == f.validationText {
			return f.validation
		}
		return fmt.Errorf("%s", f.Error)
	}
	if f.Size < 0 {
		return &localizedError{key: "unknownFileSize"}
	}
	if r.MaxBytes > 0 && f.Size > r.MaxBytes {
		return &localizedError{key: "fileTooLarge", args: map[string]any{"size": r.MaxBytes}}
	}
	if len(r.Accept) == 0 {
		return nil
	}
	name := strings.ToLower(f.Name)
	kind := strings.ToLower(strings.TrimSpace(strings.Split(f.MIME, ";")[0]))
	filtered := false
	for _, accept := range r.Accept {
		for entry := range strings.SplitSeq(accept, ",") {
			a := strings.ToLower(strings.TrimSpace(entry))
			if a == "" {
				continue
			}
			filtered = true
			if a == "*" || a == "*/*" || (strings.HasPrefix(a, ".") && strings.HasSuffix(name, a)) || (strings.HasSuffix(a, "/*") && strings.HasPrefix(kind, strings.TrimSuffix(a, "*"))) || a == kind {
				return nil
			}
		}
	}
	if !filtered {
		return nil
	}
	return &localizedError{key: "fileTypeNotAccepted"}
}

// InspectUploadFiles runs outside rendering, preferably on a worker. It never opens file content.
func InspectUploadFiles(ctx context.Context, paths []string) ([]UploadFile, error) {
	out := make([]UploadFile, 0, len(paths))
	for _, path := range paths {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		f := UploadFile{Path: filepath.Clean(path), Name: filepath.Base(path)}
		info, err := os.Stat(path)
		switch {
		case err != nil:
			f.Error = err.Error()
		case !info.Mode().IsRegular():
			f.setError(&localizedError{key: "regularFileRequired"})
		default:
			f.Size = info.Size()
			f.MIME = mime.TypeByExtension(strings.ToLower(filepath.Ext(path)))
		}
		out = append(out, f)
	}
	return out, nil
}

// PickUploadFiles uses the OS dialog; cancel returns an empty result. The OS dialog itself
// is not forcibly closed by context cancellation; its eventual result is ignored.
func PickUploadFiles(ctx context.Context, o mygo.OpenDialogOptions) ([]UploadFile, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	paths, err := mygo.Dialog.Open(o)
	if err != nil {
		return nil, err
	}
	return InspectUploadFiles(ctx, paths)
}

type UploadTicket struct {
	owner   *byte
	id      string
	version uint64
}
type uploadJob struct {
	version uint64
	cancel  context.CancelFunc
}
type uploadResult struct {
	ticket QueryTicket
	files  []UploadFile
	err    error
}
type FileUploadState struct {
	Files           []UploadFile
	Error           string
	pending         QueryState[UploadFile]
	mailbox         chan uploadResult
	jobs            map[string]uploadJob
	identity        *byte
	picker          ui.Handle
	serial, version uint64
}

// Add validates metadata and adds rejected entries with their reason. Duplicate paths are ignored.
func (s *FileUploadState) Add(files []UploadFile, r UploadRules) {
	for _, f := range files {
		if f.Path != "" && slices.ContainsFunc(s.Files, func(old UploadFile) bool { return old.Path == f.Path }) {
			continue
		}
		for {
			s.serial++
			f.ID = fmt.Sprintf("file-%d", s.serial)
			if !slices.ContainsFunc(s.Files, func(old UploadFile) bool { return old.ID == f.ID }) {
				break
			}
		}
		f.Progress = 0
		f.Status = UploadReady
		accepted := 0
		for _, old := range s.Files {
			if old.Status != UploadRejected {
				accepted++
			}
		}
		if err := r.Validate(f); err != nil {
			f.Status = UploadRejected
			f.setError(err)
		} else if r.MaxFiles > 0 && accepted >= r.MaxFiles {
			f.Status = UploadRejected
			f.setError(&localizedError{key: "tooManyFiles", args: map[string]any{"count": r.MaxFiles}})
		}
		s.Files = append(s.Files, f)
	}
}
func (s *FileUploadState) Begin(id string) (context.Context, UploadTicket) {
	for i := range s.Files {
		f := &s.Files[i]
		if f.ID != id || f.Status == UploadRejected {
			continue
		}
		s.Cancel(id)
		if s.identity == nil {
			s.identity = new(byte)
		}
		if s.jobs == nil {
			s.jobs = map[string]uploadJob{}
		}
		s.version++
		ctx, cancel := context.WithCancel(context.Background())
		s.jobs[id] = uploadJob{s.version, cancel}
		f.Status = UploadRunning
		f.Progress = 0
		f.Error = ""
		f.validation, f.validationText = nil, ""
		return ctx, UploadTicket{s.identity, id, s.version}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx, UploadTicket{}
}

// Report must run on the UI thread (e.g. Window.Update). Stale, canceled and removed jobs are ignored.
func (s *FileUploadState) Report(ticket UploadTicket, progress float64, done bool, err error) bool {
	if math.IsNaN(progress) || math.IsInf(progress, 0) {
		return false
	}
	j, ok := s.jobs[ticket.id]
	if !ok || ticket.owner != s.identity || ticket.version != j.version {
		return false
	}
	for i := range s.Files {
		f := &s.Files[i]
		if f.ID != ticket.id {
			continue
		}
		f.Progress = max(0, min(1, progress))
		if err != nil {
			f.Status = UploadFailed
			f.setError(err)
		} else if done {
			f.Status = UploadDone
			f.Progress = 1
		}
		if err != nil || done {
			j.cancel()
			delete(s.jobs, ticket.id)
		}
		return true
	}
	return false
}
func (s *FileUploadState) Cancel(id string) {
	if j, ok := s.jobs[id]; ok {
		j.cancel()
		delete(s.jobs, id)
		for i := range s.Files {
			if s.Files[i].ID == id {
				s.Files[i].Status = UploadCanceled
			}
		}
	}
}
func (s *FileUploadState) Remove(id string) {
	s.Cancel(id)
	s.Files = slices.DeleteFunc(s.Files, func(f UploadFile) bool { return f.ID == id })
}

// Dispose cancels pending inspections and application upload contexts. Call on hiding/unmount/close.
func (s *FileUploadState) Dispose() {
	s.pending.Cancel()
	s.mailbox = nil
	for id := range s.jobs {
		s.Cancel(id)
	}
}
func (s *FileUploadState) Pending() bool { return s.pending.Loading }

// load runs job on a worker. Its result waits in the mailbox for the next
// build, which the worker asks for once it is there, rather than the window
// polling for as long as a native dialog stays open.
func (s *FileUploadState) load(services ui.Services, job func(context.Context) ([]UploadFile, error)) {
	ctx, ticket := s.pending.Start(context.Background(), "")
	s.Error = ""
	mailbox := make(chan uploadResult, 1)
	s.mailbox = mailbox
	go func() {
		files, err := job(ctx)
		select {
		case <-ctx.Done():
			return
		case mailbox <- uploadResult{ticket, files, err}:
			services.Invalidate()
		}
	}()
}

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
type UploadVariant uint8

const (
	UploadArea UploadVariant = iota
	UploadButton
)

type UploadLayout uint8

const (
	UploadLayoutDefault UploadLayout = iota
	UploadList
	UploadGrid
)

type UploadPosition uint8

const (
	UploadOutside UploadPosition = iota
	UploadInside
)

func FileUpload(c *ui.Context, t Theme, s *FileUploadState, o FileUploadOptions) ui.Element {
	o.Disabled = o.Disabled || t.field.Disabled
	config := selected(defaults(t, ComponentFileUpload, ComponentTheme{Color: Primary, Size: MD}), ComponentTheme{Color: o.Color, Size: inheritedSize(t, o.Size)})
	o.Color, o.Size = config.Color, config.Size
	if o.Layout == UploadLayoutDefault {
		o.Layout = UploadGrid
	}
	if o.Variant > UploadButton || o.Layout > UploadGrid || o.Position > UploadInside {
		panic("nui: invalid file upload appearance")
	}
	if s == nil {
		panic("nui: FileUpload requires state")
	}
	root := ui.Column(c).MinWidth(0).Disabled(o.Disabled)
	o.Disabled = root.IsDisabled()
	if o.Disabled && s.pending.Loading {
		s.pending.Cancel()
		s.mailbox = nil
	}
	if s.mailbox != nil {
		select {
		case r := <-s.mailbox:
			if s.pending.Complete(r.ticket, r.files, r.err) {
				s.Error = s.pending.Error
				before := len(s.Files)
				if r.err == nil {
					s.Add(r.files, o.Rules)
				}
				if len(s.Files) != before {
					rejected := []UploadFile{}
					for i := before; i < len(s.Files); i++ {
						s.Files[i].localizeError(c.Locale())
						if s.Files[i].Status == UploadRejected {
							rejected = append(rejected, s.Files[i])
						}
					}
					files := slices.Clone(s.Files)
					if len(rejected) > 0 && o.OnReject != nil {
						o.OnReject(rejected)
					}
					if o.OnChange != nil {
						o.OnChange(files)
					}
				}
			}
			s.mailbox = nil
		default:
		}
	}
	services := c.Services()
	choose := func() {
		if o.Disabled || s.Pending() {
			return
		}
		job := o.Choose
		if job == nil {
			dialog := o.Dialog
			job = func(ctx context.Context) ([]UploadFile, error) { return PickUploadFiles(ctx, dialog) }
		}
		s.load(services, job)
	}
	opacity := float32(1)
	if o.Disabled {
		opacity = .75
	}
	finish(t, ComponentFileUpload, Root, root, PartStyle{Gap: Set(t.spacing(8)), Opacity: Set(opacity)}, o.Decoration)
	root.Children(func() {
		chooseLabel := o.ChooseLabel
		if chooseLabel == "" {
			chooseLabel = message(c, "chooseFiles")
		}
		zone := ui.Column(c).MinWidth(0).AlignItems(ui.Center).Role(ui.RoleGroup)
		inside := o.Position == UploadInside && o.Layout == UploadList && o.Variant == UploadArea
		picker := zone
		// Nuxt's interactive area is itself the picker's button.
		area := o.Variant == UploadArea && valueOr(o.Interactive, true)
		configurePicker := func(target ui.Element) {
			name := o.Label
			if name == "" {
				name = chooseLabel
			}
			target.Role(ui.RoleButton).Label(name).Description(o.Description).Focusable().FocusRing(false).Cursor(ui.CursorPointer).
				OnClick(choose).OnShortcut(0, ui.KeyEnter, choose).OnShortcut(0, ui.KeySpace, choose)
			target.Bind(&s.picker)
			bindField(c, target, t)
			if o.Label != "" || o.ChooseLabel != "" {
				target.Label(name)
			}
		}
		if area && !inside {
			configurePicker(zone)
		}
		dropzone := valueOr(o.Dropzone, true) && o.Variant == UploadArea && !o.Disabled
		if dropzone {
			if paths := zone.DroppedFiles(); len(paths) > 0 {
				paths = slices.Clone(paths)
				s.load(services, func(ctx context.Context) ([]UploadFile, error) { return InspectUploadFiles(ctx, paths) })
			}
		}
		zone.Children(func() {
			if inside {
				fileUploadList(c, t, s, o)
			}
			buildPicker := func() {
				if o.Variant == UploadArea {
					if valueOr(o.ShowIcon, true) {
						icon := o.Icon
						if icon == nil {
							icon = lucide.Upload()
						}
						leading := Avatar(c.Key("avatar"), t, AvatarOptions{Icon: icon, Size: o.Size}).Margin(0, 0, t.spacing(4), 0)
						finish(t, ComponentFileUpload, Leading, leading, PartStyle{}, o.Decoration)
					}
					if o.Label != "" {
						label := ui.Text(c.Key("label"), o.Label)
						finish(t, ComponentFileUpload, Label, label, PartStyle{Weight: Set(500), Foreground: Set(t.Text), FontSize: Set(buttonMetrics(t, o.Size).Font)}, o.Decoration)
					}
					if o.Description != "" {
						description := ui.Text(c.Key("description"), o.Description)
						finish(t, ComponentFileUpload, Description, description, PartStyle{Foreground: Set(t.Muted), FontSize: Set(buttonMetrics(t, o.Size).Font)}, o.Decoration)
					}
				}
				if !area {
					icon := o.Icon
					if icon == nil {
						icon = lucide.Upload()
					}
					if !valueOr(o.ShowIcon, true) {
						icon = nil
					}
					button := Button(c.Key("choose"), t, chooseLabel, ButtonOptions{Variant: Outline, Color: o.Color, Size: o.Size, Leading: icon, Disabled: s.Pending()}).OnClick(choose)
					button.Bind(&s.picker)
					bindField(c, button, t)
					button.Label(chooseLabel)
					finish(t, ComponentFileUpload, Trigger, button, PartStyle{}, o.Decoration)
				}
			}
			if inside && area {
				// A button is an accessibility leaf. Keep file actions outside the
				// picker button so native assistive APIs can reach each action.
				picker = ui.Column(c.Key("picker")).MinWidth(0).AlignSelf(ui.Stretch).AlignItems(ui.Center).Gap(t.spacing(8))
				configurePicker(picker)
				finish(t, ComponentFileUpload, Trigger, picker, PartStyle{}, o.Decoration)
				picker.Children(buildPicker)
			} else {
				buildPicker()
			}
		})
		bg := t.Background
		if dropzone && zone.FileDragOver() || area && zone.Hovered() && !o.Disabled {
			bg = t.Elevated.Alpha(.25)
		}
		border := t.Border
		if o.Highlight || area && picker.FocusVisible() {
			border = inputColor(t, o.Color)
		}
		if t.field.Error != "" {
			border = t.Error
		}
		// Nuxt's dashed dropzone. Part borders draw as inset rings, which
		// cannot dash, so the zone draws its own border from the merged parts.
		base := PartStyle{Background: Set(bg), Radius: Set(t.radius(8)), BorderWidth: Set(float32(1)), BorderColor: Set(border), PaddingX: Set(t.spacing(16)), PaddingY: Set(t.spacing(16)), Gap: Set(t.spacing(8))}
		if o.Variant == UploadButton {
			base.BorderWidth, base.PaddingX, base.PaddingY = Set(float32(0)), Set(float32(0)), Set(float32(0))
			zone.AlignItems(ui.Start)
		}
		zoneStyle := mergeStyle(mergeStyle(base, t.components[ComponentFileUpload].UI[Body]), o.UI[Body])
		if w := zoneStyle.BorderWidth.Value; w > 0 {
			zone.BorderWidth(w).BorderColor(zoneStyle.BorderColor.Value)
			if o.Variant == UploadArea && valueOr(o.Dropzone, true) {
				zone.BorderStyle(ui.BorderDashed)
			}
		}
		zoneStyle.BorderWidth = Set(float32(0))
		zoneStyle.apply(zone)
		if area && picker.FocusVisible() {
			focus := inputColor(t, o.Color)
			if t.field.Error != "" {
				focus = t.Error
			}
			zone.Shadow(0, 0, 0, 3, focus.Alpha(.25))
		}
		if o.Style != nil {
			o.Style(Body, zone)
		}
		if s.Pending() {
			ui.Text(c, message(c, "readingFiles")).Role(ui.RoleStatus).TextColor(t.Muted)
		}
		if s.Error != "" {
			ui.Text(c, s.Error).Role(ui.RoleStatus).TextColor(t.Error)
		}
		if !inside {
			fileUploadList(c, t, s, o)
		}
	})
	return root
}

func fileUploadList(c *ui.Context, t Theme, s *FileUploadState, o FileUploadOptions) {
	if !valueOr(o.Preview, true) || len(s.Files) == 0 {
		return
	}
	files := ui.Column(c.Key("files")).MinWidth(0).AlignSelf(ui.Stretch)
	finish(t, ComponentFileUpload, List, files, PartStyle{Gap: Set(t.spacing(8))}, o.Decoration)
	if o.Layout == UploadGrid {
		files.Row().Wrap()
	}
	files.Children(func() {
		for i := range s.Files {
			s.Files[i].localizeError(c.Locale())
			f := s.Files[i]
			row := ui.Row(c.Key(f.ID)).MinWidth(0).Gap(8).AlignItems(ui.Center)
			if o.Layout == UploadGrid {
				row.Column().Width(220).AlignItems(ui.Stretch)
			}
			finish(t, ComponentFileUpload, Item, row, PartStyle{BorderWidth: Set(float32(1)), BorderColor: Set(t.Border), Radius: Set(t.radius(6)), PaddingX: Set(t.spacing(12)), PaddingY: Set(t.spacing(8))}, o.Decoration)
			row.Children(func() {
				avatar := AvatarOptions{Icon: o.FileIcon, Size: o.Size, Alt: f.Name}
				if valueOr(o.FileImage, true) && f.Avatar != nil {
					avatar = *f.Avatar
					if avatar.Size == SizeDefault {
						avatar.Size = o.Size
					}
					if avatar.Alt == "" {
						avatar.Alt = f.Name
					}
				}
				if avatar.Icon == nil && avatar.Text == "" {
					avatar.Icon = o.FileIcon
					if avatar.Icon == nil {
						avatar.Icon = lucide.File()
					}
				}
				leading := Avatar(c.Key("avatar"), t, avatar).Shrink(0)
				finish(t, ComponentFileUpload, ItemLeadingSlot, leading, PartStyle{}, o.Decoration)
				ui.Column(c).Grow(1).MinWidth(0).Gap(4).Children(func() {
					label := ui.Text(c.Key("name"), f.Name).SingleLine().Description(f.Path)
					finish(t, ComponentFileUpload, ItemLabelSlot, label, PartStyle{Foreground: Set(t.Text), FontSize: Set(buttonMetrics(t, o.Size).Font)}, o.Decoration)
					status := ui.Text(c.Key("status"), formatMessage(c, "fileSize", map[string]any{"size": formatFileSize(f.Size), "status": uploadStatusLabel(c, f.Status)}))
					finish(t, ComponentFileUpload, StatusSlot, status, PartStyle{FontSize: Set(t.text(12)), Foreground: Set(t.Muted)}, o.Decoration)
					if f.Error != "" {
						errorText := ui.Text(c.Key("error"), f.Error)
						finish(t, ComponentFileUpload, ErrorSlot, errorText, PartStyle{Foreground: Set(t.Error)}, o.Decoration)
					}
					if f.Status == UploadRunning {
						Progress(c, t, f.Progress*100, ProgressOptions{Label: namedMessage(c, "progressNamed", f.Name)})
					}
				})
				if f.Status == UploadRunning {
					button := Button(c.Key("cancel"), t, message(c, "cancel"), ButtonOptions{AccessibleLabel: namedMessage(c, "cancelNamed", f.Name), Variant: Ghost, Color: Neutral, Size: o.Size}).OnClick(func() {
						s.Cancel(f.ID)
						s.picker.Focus()
						if o.OnCancel != nil {
							o.OnCancel(f.ID)
						}
					})
					finish(t, ComponentFileUpload, ItemTrailingSlot, button, PartStyle{}, o.Decoration)
				} else if valueOr(o.FileDelete, true) {
					deleteOptions := ButtonOptions{Square: true, Variant: LinkVariant, Color: Neutral, Size: o.Size}
					if o.Layout == UploadGrid {
						deleteOptions.Variant, deleteOptions.Size = Solid, XS
					}
					if o.FileDeleteOptions != nil {
						custom := *o.FileDeleteOptions
						if custom.Variant == VariantDefault {
							custom.Variant = deleteOptions.Variant
						}
						if custom.Color == ColorDefault {
							custom.Color = deleteOptions.Color
						}
						if custom.Size == SizeDefault {
							custom.Size = deleteOptions.Size
						}
						deleteOptions = custom
					}
					if deleteOptions.AccessibleLabel == "" {
						deleteOptions.AccessibleLabel = namedMessage(c, "removeNamed", f.Name)
					}
					if deleteOptions.Leading == nil && deleteOptions.Trailing == nil {
						deleteOptions.Trailing = o.FileDeleteIcon
						if deleteOptions.Trailing == nil {
							deleteOptions.Trailing = lucide.X()
						}
					}
					button := Button(c.Key("remove"), t, "", deleteOptions).OnClick(func() {
						s.Remove(f.ID)
						s.picker.Focus()
						if o.OnChange != nil {
							o.OnChange(slices.Clone(s.Files))
						}
						if o.OnRemove != nil {
							o.OnRemove(f.ID)
						}
					})
					finish(t, ComponentFileUpload, ItemTrailingSlot, button, PartStyle{}, o.Decoration)
				}
			})
		}
	})
}
func uploadStatusLabel(c *ui.Context, s UploadStatus) string {
	switch s {
	case UploadRunning:
		return message(c, "uploadRunning")
	case UploadDone:
		return message(c, "uploadDone")
	case UploadFailed:
		return message(c, "uploadFailed")
	case UploadCanceled:
		return message(c, "uploadCanceled")
	case UploadRejected:
		return message(c, "uploadRejected")
	default:
		return message(c, "uploadReady")
	}
}
