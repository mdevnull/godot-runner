package ui

import (
	"strings"

	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

type envFormBindings struct {
	name         binding.String
	args         binding.String
	sceneBinding binding.String
}

func (e *env) createEnvFormItems() ([]*widget.FormItem, *envFormBindings) {
	sceneItem, innerBidning := newFilePickerFormItem(
		"Select scene",
		e.global,
		storage.NewExtensionFileFilter([]string{".tscn"}),
		func(path string) string {
			projectPath, _ := e.global.projectPathBinding.Get()

			if len(projectPath) > 0 {
				path = strings.Replace(path, projectPath, "", 1)
				path = strings.TrimLeft(path, "/")
			}

			return path
		},
	)
	if e.scene != "" {
		innerBidning.Set(e.scene)
	}

	argsBinding := binding.NewString()
	if e.args != "" {
		argsBinding.Set(e.args)
	}
	argumentsEntry := widget.NewEntryWithData(argsBinding)
	argumentsEntry.MultiLine = true

	nameBinding := binding.NewString()
	if e.name != "" {
		nameBinding.Set(e.name)
	}

	return []*widget.FormItem{
		widget.NewFormItem("Name", widget.NewEntryWithData(nameBinding)),
		sceneItem,
		widget.NewFormItem("Arguments", argumentsEntry),
		widget.NewFormItem("Hide Window", widget.NewCheck("Flag", func(b bool) {
			e.noWindow = b
		})),
		widget.NewFormItem("Debug Collisions", widget.NewCheck("Flag", func(b bool) {
			e.debugCollisions = b
		})),
		widget.NewFormItem("Debug Navigation", widget.NewCheck("Flag", func(b bool) {
			e.debugNavigation = b
		})),
	}, &envFormBindings{nameBinding, argsBinding, innerBidning}
}
