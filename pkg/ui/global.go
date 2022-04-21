package ui

import (
	"os/exec"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/sirupsen/logrus"
)

type global struct {
	exeBinding         binding.String
	projectPathBinding binding.String
	errBind            binding.String
	win                fyne.Window
}

func globalForm(errBind binding.String, win fyne.Window) (fyne.Widget, *global) {
	g := &global{
		errBind: errBind,
		win:     win,
	}

	execFormItem, exeBinding := newFilePickerFormItem("Godot editor", g, nil, nil)
	g.exeBinding = exeBinding

	projectPathItem, projectPathBinding := newFolderPickerFormItem("Project path", g)
	g.projectPathBinding = projectPathBinding

	return widget.NewForm(execFormItem, projectPathItem), g
}

func (g *global) BuildSolution(complete func(), errFn func()) {
	execPath, err := g.exeBinding.Get()
	if err != nil {
		g.errBind.Set("Missing path to executable")
		return
	}
	projectPath, err := g.projectPathBinding.Get()
	if err != nil {
		g.errBind.Set("Missing path to project")
		return
	}

	cmd := exec.Command(
		execPath,
		"--build-solutions",
		"--path",
		projectPath,
		"--no-window",
		"-v",
		"-q",
	)
	if err := cmd.Start(); err != nil {
		logrus.WithError(err).Error("error starting build")
		g.errBind.Set("Unable to start build")
		errFn()
		return
	}

	go func() {
		err := cmd.Wait()
		if err != nil {
			logrus.WithError(err).Error("error in build")
			g.errBind.Set("Missing path to project")
			errFn()
			return
		}

		complete()
	}()
}
