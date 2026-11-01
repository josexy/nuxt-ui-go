package nui

import (
	"context"
	"fmt"
	"math"
	"mime"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

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

// UploadFile owns metadata only. MIME is advisory (extension-derived by InspectUploadFiles).
// The application owns file handles, content validation, upload transports and cleanup.
type UploadFile struct {
	ID, Path, Name, MIME, Error string
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
	ext := strings.ToLower(filepath.Ext(f.Name))
	kind := strings.ToLower(strings.Split(f.MIME, ";")[0])
	for _, accept := range r.Accept {
		a := strings.ToLower(strings.TrimSpace(accept))
		if a == "" {
			continue
		}
		if a == "*/*" || (strings.HasPrefix(a, ".") && a == ext) || (strings.HasSuffix(a, "/*") && strings.HasPrefix(kind, strings.TrimSuffix(a, "*"))) || a == kind {
			return nil
		}
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
	serial, version uint64
}

// Add validates metadata and adds rejected entries with their reason. Duplicate paths are ignored.
func (s *FileUploadState) Add(files []UploadFile, r UploadRules) {
	for _, f := range files {
		if f.Path != "" && slices.ContainsFunc(s.Files, func(old UploadFile) bool { return old.Path == f.Path }) {
			continue
		}
		s.serial++
		f.ID = fmt.Sprintf("file-%d", s.serial)
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
func (s *FileUploadState) load(job func(context.Context) ([]UploadFile, error)) {
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
		}
	}()
}

type FileUploadOptions struct {
	Label, Description, ChooseLabel            string
	Disabled                                   bool
	Dropzone, Interactive, FileDelete, Preview Optional[bool]
	Variant                                    UploadVariant
	Layout                                     UploadLayout
	Color                                      Color
	Size                                       Size
	Rules                                      UploadRules
	Dialog                                     mygo.OpenDialogOptions
	// Choose overrides the native dialog. It executes on a worker and must not mutate UI state.
	Choose             func(context.Context) ([]UploadFile, error)
	OnChange           func([]UploadFile)
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

func FileUpload(c *ui.Context, t Theme, s *FileUploadState, o FileUploadOptions) ui.Element {
	o.Disabled = o.Disabled || t.field.Disabled
	config := selected(defaults(t, ComponentFileUpload, ComponentTheme{Color: Primary, Size: MD}), ComponentTheme{Color: o.Color, Size: inheritedSize(t, o.Size)})
	o.Color, o.Size = config.Color, config.Size
	if o.Layout == UploadLayoutDefault {
		o.Layout = UploadGrid
	}
	if o.Variant > UploadButton || o.Layout > UploadGrid {
		panic("nui: invalid file upload appearance")
	}
	if s == nil {
		panic("nui: FileUpload requires state")
	}
	if o.Disabled && s.pending.Loading {
		s.pending.Cancel()
		s.mailbox = nil
	}
	if s.mailbox != nil {
		select {
		case r := <-s.mailbox:
			if s.pending.Complete(r.ticket, r.files, r.err) {
				s.Error = s.pending.Error
				s.Add(r.files, o.Rules)
				if o.OnChange != nil {
					o.OnChange(slices.Clone(s.Files))
				}
			}
			s.mailbox = nil
		default:
		}
	}
	if s.mailbox != nil {
		c.After(20 * time.Millisecond)
	}
	root := ui.Column(c).MinWidth(0).Disabled(o.Disabled)
	choose := func() {
		if o.Disabled || s.Pending() {
			return
		}
		job := o.Choose
		if job == nil {
			dialog := o.Dialog
			job = func(ctx context.Context) ([]UploadFile, error) { return PickUploadFiles(ctx, dialog) }
		}
		s.load(job)
	}
	finish(t, ComponentFileUpload, Root, root, PartStyle{Gap: Set(t.spacing(8))}, o.Decoration)
	root.Children(func() {
		zone := ui.Column(c).MinWidth(0).AlignItems(ui.Center).Label(o.Label).Role(ui.RoleGroup)
		if valueOr(o.Interactive, true) && o.Variant == UploadArea {
			zone.OnClick(choose)
		}
		bg := t.Background
		if zone.FileDragOver() {
			bg = t.Elevated
		}
		finish(t, ComponentFileUpload, Body, zone, PartStyle{Background: Set(bg), Radius: Set(t.radius(8)), BorderWidth: Set(float32(1)), BorderColor: Set(t.Border), PaddingX: Set(t.spacing(16)), PaddingY: Set(t.spacing(20)), Gap: Set(t.spacing(8))}, o.Decoration)
		if paths := zone.DroppedFiles(); valueOr(o.Dropzone, true) && o.Variant == UploadArea && !o.Disabled && len(paths) > 0 {
			paths = slices.Clone(paths)
			s.load(func(ctx context.Context) ([]UploadFile, error) { return InspectUploadFiles(ctx, paths) })
		}
		zone.Children(func() {
			if o.Variant == UploadArea {
				Icon(c, t, lucide.FolderOpen(), IconOptions{})
				ui.Text(c, o.Label).FontWeight(600)
				ui.Text(c, o.Description).TextColor(t.Muted)
			}
			label := o.ChooseLabel
			if label == "" {
				label = message(c, "chooseFiles")
			}
			Button(c.Key("choose"), t, label, ButtonOptions{Variant: Outline, Color: o.Color, Size: o.Size, Disabled: s.Pending()}).OnClick(choose)
		})
		if o.Variant == UploadButton {
			zone.BorderWidth(0).Padding(0).AlignItems(ui.Start)
		}
		if s.Pending() {
			ui.Text(c, message(c, "readingFiles")).Role(ui.RoleStatus).TextColor(t.Muted)
		}
		if s.Error != "" {
			ui.Text(c, s.Error).Role(ui.RoleStatus).TextColor(t.Error)
		}
		if !valueOr(o.Preview, true) {
			return
		}
		files := ui.Column(c.Key("files")).MinWidth(0).Gap(t.spacing(8))
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
					ui.Column(c).Grow(1).MinWidth(0).Gap(4).Children(func() {
						ui.Text(c, f.Name).SingleLine().Description(f.Path)
						ui.Text(c, formatMessage(c, "fileSummary", map[string]any{"size": f.Size, "status": uploadStatusLabel(c, f.Status)})).FontSize(t.text(12)).TextColor(t.Muted)
						if f.Error != "" {
							ui.Text(c, f.Error).TextColor(t.Error)
						}
						if f.Status == UploadRunning {
							Progress(c, t, f.Progress*100, ProgressOptions{Label: namedMessage(c, "progressNamed", f.Name)})
						}
					})
					if f.Status == UploadRunning {
						Button(c.Key("cancel"), t, message(c, "cancel"), ButtonOptions{AccessibleLabel: namedMessage(c, "cancelNamed", f.Name), Variant: Ghost, Color: Neutral}).OnClick(func() {
							s.Cancel(f.ID)
							if o.OnCancel != nil {
								o.OnCancel(f.ID)
							}
						})
					} else if valueOr(o.FileDelete, true) {
						Button(c.Key("remove"), t, "", ButtonOptions{Leading: lucide.X(), AccessibleLabel: namedMessage(c, "removeNamed", f.Name), Square: true, Variant: Ghost, Color: Neutral}).OnClick(func() {
							s.Remove(f.ID)
							if o.OnChange != nil {
								o.OnChange(slices.Clone(s.Files))
							}
							if o.OnRemove != nil {
								o.OnRemove(f.ID)
							}
						})
					}
				})
			}
		})
	})
	return root
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
