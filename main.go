package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"github.com/pic4pdf/pic4pdf/internal/gui"
)

func main() {
	a := app.NewWithID("com.pic4pdf")
	w := a.NewWindow("pic4pdf")
	w.Resize(fyne.NewSize(800, 600))

	fileSel := gui.NewFileSelectorPersistent("Main")
	closeWatcher := fileSel.CreateSimpleWatcher()
	defer closeWatcher()

	split := container.NewHSplit(
		fileSel,
		gui.NewFileOverview(fileSel),
	)
	split.Offset = 0.6

	w.SetContent(split)
	w.ShowAndRun()
}
