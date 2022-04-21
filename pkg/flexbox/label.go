package flexbox

import (
	"fyne.io/fyne/v2/widget"
)

type FlexLabel struct {
	widget.Label

	growth int
}

func NewLabel(grow int) *FlexLabel {
	label := &FlexLabel{
		growth: grow,
	}
	label.ExtendBaseWidget(label)

	return label
}

func (fw *FlexLabel) Grow() int {
	return fw.growth
}
