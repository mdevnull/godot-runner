package flexbox

import (
	"fyne.io/fyne/v2"
	"github.com/sirupsen/logrus"
)

type FlexChild interface {
	Grow() int
}

type fbox struct {
	flexH bool
}

func NewHFlex() fyne.Layout {
	return &fbox{true}
}

func NewVFlex() fyne.Layout {
	return &fbox{false}
}

func (f *fbox) MinSize(objects []fyne.CanvasObject) fyne.Size {
	w, h := float32(0), float32(0)
	for _, o := range objects {
		childSize := o.MinSize()
		if f.flexH {
			if childSize.Height > h {
				h = childSize.Height
			}
			w += childSize.Width
		} else {
			if childSize.Width > w {
				w = childSize.Width
			}
			h += childSize.Height
		}
	}

	return fyne.NewSize(w, h)
}

func (f *fbox) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	var freeSpace float32
	var growthTotal int
	if f.flexH {
		freeSpace = containerSize.Width
	} else {
		freeSpace = containerSize.Height
	}

	for _, o := range objects {
		if f.flexH {
			freeSpace -= o.MinSize().Width
		} else {
			freeSpace -= o.MinSize().Height
		}

		if fc, ok := o.(FlexChild); ok {
			growthTotal += fc.Grow()
		}
	}

	pos := fyne.NewPos(0, 0)
	for _, o := range objects {
		w, h := float32(0), float32(0)
		minSize := o.MinSize()
		w = minSize.Width
		h = minSize.Height
		if fc, ok := o.(FlexChild); ok && fc.Grow() > 0 {
			freeMulti := float32(fc.Grow()) / float32(growthTotal)
			growthAbs := freeSpace * freeMulti
			if f.flexH {
				w += growthAbs
			} else {
				h += growthAbs
			}
			logrus.WithFields(logrus.Fields{
				"growth":      fc.Grow(),
				"growthTotal": growthTotal,
				"freeSpace":   freeSpace,
				"growthAbs":   growthAbs,
				"w":           w,
				"h":           h,
			}).Debug("flexbox resizing")
		}

		o.Resize(fyne.NewSize(w, h))
		o.Move(pos)

		if f.flexH {
			pos = pos.Add(fyne.NewSize(w, 0))
		} else {
			pos = pos.Add(fyne.NewSize(0, h))
		}
	}
}
