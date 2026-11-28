package nui

import "github.com/egoist/mygo/ui"

type uploadAction struct {
	handle ui.Handle
	id     string
}

type uploadPreviews struct {
	list      ui.ListState
	grid      ui.GridState
	gridMode  bool
	modeKnown bool
	active    *uploadAction
	pending   string
	index     int
}

// A mode change is an explicit layout operation. Resolve its one focused ID
// once; unchanged builds never scan the freely mutable Files collection.
func (p *uploadPreviews) prepare(c *ui.Context, grid bool, files []UploadFile) {
	p.index = -1
	if p.modeKnown && p.gridMode != grid && p.active != nil && p.active.handle.Focused(c) {
		for i := range files {
			if files[i].ID == p.active.id {
				p.pending, p.index = p.active.id, i
				break
			}
		}
	}
	p.modeKnown, p.gridMode = true, grid
}

func (p *uploadPreviews) bindAction(c *ui.Context, button ui.Element, id string) {
	action := ui.Local(button, "nui-upload-action", func() uploadAction { return uploadAction{id: id} })
	button.Bind(&action.handle)
	if p.pending == id {
		action.handle.Focus()
		button.Focus()
		p.pending = ""
		p.active = action
	}
	if button.Focused() {
		p.active = action
	}
}

func uploadLayoutStyle(style PartStyle) bool {
	return style.Font.Set || style.FontSize.Set || style.LineHeight.Set || style.Weight.Set ||
		style.PaddingX.Set || style.PaddingY.Set || style.Gap.Set
}

// Arbitrary sizing callbacks/custom buttons need measured rows. Color-only
// part overrides preserve the fixed grid contract. App-supplied Avatar
// decorations in a grid must fit their declared avatar size.
func uploadGridFits(t Theme, o FileUploadOptions) bool {
	if o.Style != nil || o.FileDeleteOptions != nil {
		return false
	}
	for _, slot := range []Slot{Item, ItemLeadingSlot, ItemLabelSlot, StatusSlot, ErrorSlot, ItemTrailingSlot} {
		if uploadLayoutStyle(t.components[ComponentFileUpload].UI[slot]) || uploadLayoutStyle(o.UI[slot]) {
			return false
		}
	}
	for _, component := range []Component{ComponentAvatar, ComponentButton, ComponentProgress} {
		for _, style := range t.components[component].UI {
			if uploadLayoutStyle(style) {
				return false
			}
		}
	}
	return true
}

// Reserve the largest standard avatar, a single-line title/status, two error
// lines, a progress bar, and one action, including all gaps and padding.
func uploadGridHeight(t Theme, size Size) float32 {
	m := buttonMetrics(t, size)
	return avatarSize(t, XL) + 4*m.Line + 2*m.Y + t.text(16) + t.spacing(60)
}
