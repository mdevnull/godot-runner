package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

type (
	statusColor color.Color
	statusRect  struct {
		canvasObject fyne.CanvasObject
		colorRect    *canvas.Rectangle
		labelBinding binding.String
	}
)

var (
	statusColorWaiting statusColor = color.RGBA{73, 73, 73, 255}
	statusColorOk      statusColor = color.RGBA{39, 86, 19, 255}
	statusColorWarn    statusColor = color.RGBA{196, 186, 0, 255}
	statusColorError   statusColor = color.RGBA{135, 12, 24, 255}
)

func newStatusRect() *statusRect {
	sr := &statusRect{
		colorRect:    canvas.NewRectangle(statusColorWaiting),
		labelBinding: binding.NewString(),
	}

	sr.canvasObject = container.NewMax(
		sr.colorRect,
		widget.NewLabelWithData(sr.labelBinding),
	)

	return sr
}

func (cr *statusRect) SetStatus(name string, color statusColor) {
	cr.labelBinding.Set(name)
	cr.colorRect.FillColor = color
	cr.colorRect.Refresh()
}
