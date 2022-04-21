package ui

import (
	"fmt"
	"io/ioutil"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func (e *env) showLogsWin() {
	logWin := fyne.CurrentApp().NewWindow(fmt.Sprintf("%s - Logs", e.name))
	logWin.Resize(fyne.NewSize(900, 600))

	logEntry := widget.NewEntry()
	logEntry.MultiLine = true
	logWin.SetContent(container.NewVScroll(logEntry))

	logBytes, err := ioutil.ReadFile(e.lastLogFile.Name())
	if err != nil {
		panic(err)
	}
	logEntry.SetText(string(logBytes))

	logWin.Show()
}
