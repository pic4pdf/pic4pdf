package main

import (
	"fmt"
	"image"
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

	f, err := os.Open("../lib-p4p/gophers/gopher1.jpg")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}

	l := widget.NewList(
		func() int {
			return 10
		},
		func() fyne.CanvasObject {
			iv := gui.NewPDFImageView(p4p.Millimeter, p4p.A4())
			iv.SetMinSize(fyne.NewSize(0, 400))
			return iv
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			iv := obj.(*gui.PDFImageView)
			iv.SetDescription(fmt.Sprintf("%v/%v", id+1, 10))
			iv.SetOptions(p4p.ImageOptions{
				Mode: p4p.Fill,
			})
			iv.SetImage(img)
		},
	)
	l.OnSelected = func(widget.ListItemID) {
		l.UnselectAll()
	}

	split := container.NewHSplit(
		container.NewHSplit(
			fileSel,
			gui.NewFileOverview(fileSel),
		),
		l,
	)
	split.Offset = 0.6

	w.SetContent(split)
	w.ShowAndRun()
}
