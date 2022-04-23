package ui

import (
	"log"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/devnull-twitch/godot-runner/pkg/build"
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
)

type global struct {
	exeBinding               binding.String
	projectPathBinding       binding.String
	errBind                  binding.String
	win                      fyne.Window
	hasValidBuild            bool
	buildCom                 *build.RunnerChannels
	projectPathChangeHandler func(string)
	projectFileChangeHandler func()
	listerLock               *sync.Mutex
	buildCompleteListener    []chan<- bool
}

func globalForm(errBind binding.String, win fyne.Window) (fyne.Widget, *global) {
	buildCom := build.Runner()
	g := &global{
		errBind:               errBind,
		win:                   win,
		buildCom:              buildCom,
		buildCompleteListener: make([]chan<- bool, 0),
		listerLock:            &sync.Mutex{},
	}

	go func() {
		for success := range buildCom.BuildCompleteChan() {
			g.emitBuildComplete(success)
		}
	}()

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

func (g *global) emitBuildComplete(success bool) {
	g.hasValidBuild = true
	for _, c := range g.buildCompleteListener {
		select {
		case c <- success:
		case <-time.After(time.Millisecond * 10):
			logrus.WithField("timeout", "10ms").Warn("skipped build complete listener")
		}
	}
}

func (g *global) NextBuildComplete(complete func(bool)) {
	c := make(chan bool)
	go func() {
		success := <-c
		complete(success)

		g.listerLock.Lock()
		defer g.listerLock.Unlock()
		filtered := make([]chan<- bool, 0)
		for _, tester := range g.buildCompleteListener {
			if tester != c {
				filtered = append(filtered, tester)
			}
		}
		g.buildCompleteListener = filtered
	}()

	g.listerLock.Lock()
	defer g.listerLock.Unlock()
	g.buildCompleteListener = append(g.buildCompleteListener, c)
}

func (g *global) BuildSolution() {
	if g.hasValidBuild {
		g.emitBuildComplete(true)
		return
	}

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

	g.buildCom.Invalid(execPath, projectPath)
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
					g.hasValidBuild = false
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
