package flexbox

import (
	"fyne.io/fyne/v2/widget"
)

type FlexEntry struct {
	widget.Entry

	growth int
}

func NewEntry(grow int) *FlexEntry {
	entry := &FlexEntry{
		growth: grow,
	}
	entry.ExtendBaseWidget(entry)

	return entry
}

func (fw *FlexEntry) Grow() int {
	return fw.growth
}
