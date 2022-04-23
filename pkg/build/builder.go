package build

import (
	"os/exec"

	"github.com/sirupsen/logrus"
)

type (
	invalidRequest struct {
		execPath    string
		projectPath string
	}
	RunnerChannels struct {
		stopChan     chan<- bool
		invalidChan  chan<- invalidRequest
		buildResChan <-chan bool
	}
)

func (r *RunnerChannels) Stop() {
	r.stopChan <- true
}

func (r *RunnerChannels) Invalid(execPath, projectPath string) {
	r.invalidChan <- invalidRequest{execPath, projectPath}
}

func (r *RunnerChannels) BuildCompleteChan() <-chan bool {
	return r.buildResChan
}

func Runner() *RunnerChannels {
	stopChan := make(chan bool)
	invalidChan := make(chan invalidRequest)
	buildResChan := make(chan bool)
	channels := &RunnerChannels{
		stopChan:     stopChan,
		invalidChan:  invalidChan,
		buildResChan: buildResChan,
	}

	go func() {
		defer func() {
			close(buildResChan)
		}()

		redoAfter := false
		inProgress := false
		var cmd exec.Cmd

		finishChan := make(chan bool)
		var (
			newestExecPath    string
			newestProjectPath string
		)

		for {
			select {
			case <-stopChan:
				return

			case req := <-invalidChan:
				newestExecPath = req.execPath
				newestProjectPath = req.projectPath

				if inProgress {
					redoAfter = true
					cmd.Process.Kill()
					continue
				}

				inProgress = true
				doBuild(req.execPath, req.projectPath, finishChan)

			case success := <-finishChan:
				logrus.WithFields(logrus.Fields{
					"success": success,
				}).Info("build solution finished")

				if redoAfter {
					doBuild(newestExecPath, newestProjectPath, finishChan)
					continue
				}

				inProgress = false
				buildResChan <- success

			}
		}
	}()

	return channels
}

func doBuild(execPath, projectPath string, finishChan chan<- bool) (*exec.Cmd, error) {
	logrus.WithFields(logrus.Fields{
		"godot_executable": execPath,
		"project_path":     projectPath,
	}).Info("start build solution")

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
		return nil, err
	}

	go func() {
		err := cmd.Wait()
		if err != nil {
			logrus.WithError(err).Error("error build solution")
		}
		finishChan <- err == nil
	}()

	return cmd, nil
}
