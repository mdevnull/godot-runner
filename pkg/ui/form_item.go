package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/devnull-twitch/godot-runner/pkg/flexbox"
	"github.com/sirupsen/logrus"
)

func newFilePickerFormItem(
	name string,
	g *global,
	selectionFilter storage.FileFilter,
	preProcessor func(string) string,
) (*widget.FormItem, binding.String) {
	strBinding := binding.NewString()
	strBinding.Set("")

	entry := flexbox.NewEntry(1)
	entry.Bind(strBinding)
	entry.Disable()

	return widget.NewFormItem(name, container.New(
		flexbox.NewHFlex(),
		entry,
		widget.NewButton("Select", func() {
			projectPath, _ := g.projectPathBinding.Get()
			dia := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
				if uc == nil {
					strBinding.Set("")
					return
				}
				uri := uc.URI()
				path := uri.Path()
				if preProcessor != nil {
					path = preProcessor(path)
				}

				strBinding.Set(path)
			}, g.win)

			if selectionFilter != nil {
				dia.SetFilter(selectionFilter)
			}

			if projectPath != "" {
				uri := storage.NewFileURI(projectPath)
				uriList, err := storage.ListerForURI(uri)
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"error": err,
						"uri":   uri,
					}).Error("unable to create lister for uri")
				}
				if uriList != nil {
					dia.SetLocation(uriList)
				}
			}
			dia.Show()
		}),
	)), strBinding
}

func newFolderPickerFormItem(name string, g *global) (*widget.FormItem, binding.String) {
	strBinding := binding.NewString()
	strBinding.Set("")

	entry := flexbox.NewEntry(1)
	entry.Bind(strBinding)
	entry.Disable()

	return widget.NewFormItem(name, container.New(
		flexbox.NewHFlex(),
		entry,
		widget.NewButton("Select", func() {
			dia := dialog.NewFolderOpen(func(lu fyne.ListableURI, err error) {
				if lu == nil {
					strBinding.Set("")
					return
				}
				strBinding.Set(lu.Path())
			}, g.win)

			dia.Show()
		}),
	)), strBinding
}
