package ui

import (
	"log"
	"os/exec"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
)

type global struct {
	exeBinding               binding.String
	projectPathBinding       binding.String
	errBind                  binding.String
	win                      fyne.Window
	projectPathChangeHandler func(string)
	projectFileChangeHandler func()
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
	projectPathBinding.AddListener(binding.NewDataListener(func() {
		str, _ := projectPathBinding.Get()
		g.projectPathChangeHandler(str)
	}))

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

func (g *global) Watcher(projectDir string) chan<- bool {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logrus.WithError(err).Error("unable to get watcher")
		return nil
	}

	finishChan := make(chan bool)

	go func() {
		run := true
		for run {
			select {
			case <-finishChan:
				run = false
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&(fsnotify.Create|fsnotify.Write) > 0 {
					log.Println("modified file:", event.Name)
					g.projectFileChangeHandler()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logrus.Info("error:", err)
			}
		}
	}()

	watcher.Add(projectDir)
	logrus.WithField("dir", projectDir).Info("watcher started")

	return finishChan
}
