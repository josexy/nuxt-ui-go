package gallery

import (
	"context"
	"errors"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func addFileUploadContract(p *componentPage) {
	state := nui.FileUploadState{}
	state.Add([]nui.UploadFile{{Path: "sample/readme.md", Name: "README.md", Size: 2048, Avatar: &nui.AvatarOptions{Text: "MD", Color: nui.Info}}}, nui.UploadRules{})
	p.Deactivate = append(p.Deactivate, state.Dispose)
	disabled, inside, grid, highlight, showIcon, fileImage := false, false, false, true, true, true
	dropzone, interactive, custom := true, true, false
	changes, rejected := 0, 0
	outcome := "files"
	p.Scenes = append(p.Scenes, scene("contract", "File selection and notifications", "This scene uses a simulated chooser. Repeated results and cancellation leave Changes unchanged; invalid files remain visible and increment Rejected. Removal returns focus to the picker. The following native scenes use the OS dialog.", func(c *ui.Context, t nui.Theme) {
		ui.Row(c).Wrap().Gap(8).Children(func() {
			nui.Checkbox(c, t, &disabled, nui.CheckboxOptions{Label: "Disable contract uploader"})
			nui.Checkbox(c, t, &inside, nui.CheckboxOptions{Label: "Place files inside"})
			nui.Checkbox(c, t, &grid, nui.CheckboxOptions{Label: "Contract file grid"})
			nui.Checkbox(c, t, &highlight, nui.CheckboxOptions{Label: "Highlight upload area"})
			nui.Checkbox(c, t, &showIcon, nui.CheckboxOptions{Label: "Show upload icon"})
			nui.Checkbox(c, t, &fileImage, nui.CheckboxOptions{Label: "Show file avatar"})
			nui.Checkbox(c, t, &dropzone, nui.CheckboxOptions{Label: "Allow contract file drops"})
			nui.Checkbox(c, t, &interactive, nui.CheckboxOptions{Label: "Interactive contract area"})
			nui.Checkbox(c, t, &custom, nui.CheckboxOptions{Label: "Style contract file parts"})
		})
		ui.Row(c).Wrap().Gap(8).Children(func() {
			for _, mode := range []string{"files", "cancel", "error"} {
				nui.Button(c.Key(mode), t, "Choose result: "+mode, nui.ButtonOptions{Variant: nui.Outline, Active: outcome == mode}).OnClick(func() { outcome = mode })
			}
		})
		mode := outcome
		o := nui.FileUploadOptions{
			Label: "Contract attachments", Description: "Choose sample metadata; .txt / .md, at most two accepted files.", ChooseLabel: "Choose contract files",
			Disabled: disabled, Highlight: highlight, ShowIcon: nui.Set(showIcon), FileImage: nui.Set(fileImage), Dropzone: nui.Set(dropzone), Interactive: nui.Set(interactive), Layout: nui.UploadList,
			Rules: nui.UploadRules{Accept: []string{".txt, .md"}, MaxFiles: 2}, Icon: lucide.FolderUp(),
			Choose: func(context.Context) ([]nui.UploadFile, error) {
				if mode == "cancel" {
					return nil, nil
				}
				if mode == "error" {
					return nil, errors.New("sample selection failed")
				}
				return []nui.UploadFile{{Path: "sample/notes.txt", Name: "notes.txt", Size: 512}, {Path: "sample/application.exe", Name: "application.exe", Size: 1024}}, nil
			},
			OnChange: func([]nui.UploadFile) { changes++ }, OnReject: func(files []nui.UploadFile) { rejected += len(files) },
		}
		if inside {
			o.Position = nui.UploadInside
		}
		if grid {
			o.Layout = nui.UploadGrid
		}
		if custom {
			o.FileDeleteOptions = &nui.ButtonOptions{Color: nui.Error, Variant: nui.Outline}
			o.FileDeleteIcon = lucide.Trash()
			o.UI = nui.Styles{nui.ItemLabelSlot: {Foreground: nui.Set(t.Info)}, nui.Body: {BorderColor: nui.Set(t.Info)}}
		}
		nui.FileUpload(c.Key("contract-files"), t, &state, o)
		ui.Textf(c, "File changes: %d · Rejected: %d", changes, rejected)
	}))
}
