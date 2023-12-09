package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/pic4pdf/lib-p4p"

	"github.com/pic4pdf/pic4pdf/internal/gui"
)

func main() {
	a := app.NewWithID("com.pic4pdf")
	w := a.NewWindow("pic4pdf")
	w.Resize(fyne.NewSize(800, 600))

	fileSel := gui.NewFileSelectorPersistent("Main")
	closeWatcher := fileSel.CreateSimpleWatcher()
	defer closeWatcher()

	fileOw := gui.NewFileOverview(fileSel)

	imgs := make(map[string]image.Image)

	l := widget.NewList(
		func() int {
			return fileOw.NumSelected()
		},
		func() fyne.CanvasObject {
			iv := gui.NewPDFImageView(p4p.Millimeter, p4p.A4())
			iv.SetMinSize(fyne.NewSize(0, 300))
			return iv
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			sel := fileOw.Selected()
			iv := obj.(*gui.PDFImageView)
			iv.SetDescription(fmt.Sprintf("%v/%v", id+1, len(sel)))
			if id < len(sel) {
				if img, ok := imgs[sel[id]]; ok {
					mode := p4p.Fit
					scale := float64(1)
					iv.SetOptions(p4p.ImageOptions{
						Mode: mode,
						Scale: scale,
					})
					iv.SetImage(img)
				}
			}
		},
	)
	l.OnSelected = func(widget.ListItemID) {
		l.UnselectAll()
	}

	fileOw.OnSelected = func(path string) {
		f, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		img, _, err := image.Decode(f)
		if err != nil {
			panic(err)
		}
		imgs[path] = img
		l.Refresh()
	}
	fileOw.OnUnselected = func(path string) {
		delete(imgs, path)
		l.Refresh()
	}
	fileOw.OnReorder = func() {
		l.Refresh()
	}

	split := container.NewHSplit(
		container.NewHSplit(
			fileSel,
			fileOw,
		),
		l,
	)
	split.Offset = 0.6

	w.SetContent(split)
	w.ShowAndRun()
}
