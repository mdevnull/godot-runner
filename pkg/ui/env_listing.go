package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/devnull-twitch/godot-runner/pkg/flexbox"
)

func createListing(e *env) (fyne.CanvasObject, binding.String) {
	nameBinding := binding.NewString()
	nameLabel := flexbox.NewLabel(1)
	nameLabel.Bind(nameBinding)
	nameBinding.Set(e.name)

	runningStatus := newStatusRect()
	runningStatus.SetStatus("off", statusColorWaiting)

	return container.New(
		flexbox.NewHFlex(),
		nameLabel,
		runningStatus.canvasObject,
		widget.NewButton("Logs", func() {
			e.showLogsWin()
		}),
		widget.NewButton("Start/Stop", func() {
			if e.isRunning {
				e.currentProcess.Process.Kill()
				return
			}

			runningStatus.SetStatus("starting", statusColorWarn)
			e.isRunning = true
			go e.global.BuildSolution(func() {
				runningStatus.SetStatus("running", statusColorOk)
				e.start(func() {
					e.isRunning = false
					runningStatus.SetStatus("complete", statusColorWaiting)
				}, func() {
					e.isRunning = false
					runningStatus.SetStatus("runtime error", statusColorError)
				})
			}, func() {
				e.isRunning = false
				runningStatus.SetStatus("build error", statusColorError)
			})
		}),
		widget.NewButton("Edit", func() {
			formItems, editBindings := e.createEnvFormItems()
			formDia := dialog.NewForm("Update new environment", "Save", "Cancel", formItems, func(b bool) {
				if b {
					e.name, _ = editBindings.name.Get()
					e.scene, _ = editBindings.sceneBinding.Get()
					e.args, _ = editBindings.args.Get()

					e.refresh()
				}
			}, e.global.win)
			s := formDia.MinSize()
			if s.Width < e.global.win.Canvas().Size().Width*0.75 {
				s.Width = e.global.win.Canvas().Size().Width * 0.75
			}
			formDia.Resize(s)
			formDia.Show()
		}),
	), nameBinding
}
