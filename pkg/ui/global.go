package ui

import (
	"os"
	"path/filepath"
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
				logrus.WithFields(logrus.Fields{
					"target": event.Name,
					"op":     event.Op.String(),
				}).Info("received fsnotify event")

				// I do not care about these events
				if event.Op&fsnotify.Chmod == fsnotify.Chmod {
					continue
				}

				// rename or delete events remove those from the watcher
				// we do not know if the event trigger thingy was a file or dir
				// guess we just have to call remove on watcher and hope it does not explode
				if event.Op&(fsnotify.Remove|fsnotify.Rename) > 0 {
					watcher.Remove(event.Name)
					continue
				}

				fsInfo, err := os.Stat(event.Name)
				if err != nil {
					logrus.WithError(err).Error("unable to get stats of modified file/dir")
					continue
				}

				// create ( also triggered for renames ) should add the new directory to the watcher
				if fsInfo.IsDir() && event.Op == fsnotify.Create {
					baseName := filepath.Base(event.Name)
					if baseName[0:1] == "." {
						continue
					}
					watcher.Add(event.Name)
					continue
				}

				g.hasValidBuild = false
				g.projectFileChangeHandler()
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logrus.Info("error:", err)
			}
		}
	}()

	watcher.Add(projectDir)
	AddRecusiveAllDirs(projectDir, watcher)

	logrus.WithField("dir", projectDir).Info("watcher started")

	return finishChan
}

func AddRecusiveAllDirs(dir string, watcher *fsnotify.Watcher) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"dir":   dir,
		}).Error("unable to read directory")
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			if entry.Name()[0:1] == "." {
				continue
			}
			fullPath := filepath.Join(dir, entry.Name())
			watcher.Add(fullPath)
			AddRecusiveAllDirs(fullPath, watcher)
		}
	}
}
