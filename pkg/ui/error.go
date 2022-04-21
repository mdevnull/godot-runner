package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

func errorTuple() (fyne.CanvasObject, binding.String) {
	errBind := binding.NewString()
	errContainer := container.NewCenter(widget.NewLabelWithData(errBind))

	errContainer.Hide()
	errBind.AddListener(binding.NewDataListener(func() {
		v, _ := errBind.Get()
		if v == "" {
			errContainer.Hide()
		} else {
			errContainer.Show()
		}
	}))

	return errContainer, errBind
}
