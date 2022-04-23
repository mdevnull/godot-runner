package ui

import (
	"os"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
)

type env struct {
	name            string
	args            string
	scene           string
	lastLogFile     *os.File
	global          *global
	listeners       []func()
	isRunning       bool
	currentProcess  *exec.Cmd
	noWindow        bool
	debugCollisions bool
	debugNavigation bool
	restartOnChange bool
	restartNext     bool
}

func createEnv(g *global) *env {
	return &env{
		global:    g,
		listeners: make([]func(), 0),
	}
}

func (e *env) start(completeFn func(), errorFn func(), restart func(bool)) {
	execPath, err := e.global.exeBinding.Get()
	if err != nil {
		e.global.errBind.Set("Missing path to executable")
		return
	}
	projectPath, err := e.global.projectPathBinding.Get()
	if err != nil {
		e.global.errBind.Set("Missing path to project")
		return
	}

	customArgs := strings.Split(e.args, " ")

	args := []string{"--path", projectPath, e.scene}
	if e.noWindow {
		args = append(args, "--no-window")
	}
	if e.debugCollisions {
		args = append(args, "--debug-collisions")
	}
	if e.debugNavigation {
		args = append(args, "--debug-navigation")
	}
	args = append(args, customArgs...)

	// create new empty binding for this run
	e.currentProcess = exec.Command(execPath, args...)

	tmpFile, err := os.CreateTemp(os.TempDir(), "runnerlog")
	if err != nil {
		e.global.errBind.Set("Unable to create tmp logfile")
		return
	}
	logrus.WithField("tmp_file", tmpFile.Name()).Info("created tmp file")
	e.lastLogFile = tmpFile

	e.currentProcess.Stdout = tmpFile
	e.currentProcess.Stderr = tmpFile

	if err := e.currentProcess.Run(); err != nil {
		if e.restartNext {
			e.restartNext = false
			restart(false)
		} else {
			errorFn()
		}
		return
	}
	if e.restartNext {
		e.restartNext = false
		restart(true)
	} else {
		completeFn()
	}

}

func (e *env) AddListener(listener func()) {
	e.listeners = append(e.listeners, listener)
}

func (e *env) refresh() {
	for _, l := range e.listeners {
		l()
	}
}
