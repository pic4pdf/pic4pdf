package main

import (
	"path/filepath"
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/pic4pdf/pic4pdf/internal/gui"
)

func main() {
	a := app.NewWithID("com.pic4pdf")
	w := a.NewWindow("pic4pdf")

	fileSel := gui.NewFileSelectorPersistent("Main")
	closeWatcher := fileSel.CreateSimpleWatcher()
	defer closeWatcher()

	var files []string
	fileList := widget.NewList(
		func() int {
			return len(files)
		}, func() fyne.CanvasObject {
			item := gui.NewFileItem(
				"PLACEHOLDER",
				nil,
				theme.ContentRemoveIcon(),
				nil,
			)
			item.LabelButton.Hide()
			return item
		}, func(id widget.ListItemID, obj fyne.CanvasObject) {
			item := obj.(*gui.FileItem)
			item.Label.SetText(filepath.Base(files[id]))
			item.LabelIcon.SetResource(theme.FileImageIcon())
			item.IconButton.OnTapped = func() {
				fileSel.Unselect(files[id])
			}
		},
	)

	fileSel.OnSelected = func(path string) {
		if slices.Index[[]string, string](files, path) == -1 {
			files = append(files, path)
		}
		fileList.Refresh()
	}

	fileSel.OnUnselected = func(path string) {
		if idx := slices.Index[[]string, string](files, path); idx != -1 {
			files = append(files[:idx], files[idx+1:]...)
		}
		fileList.Refresh()
	}

	w.SetContent(container.NewHSplit(
		fileSel,
		fileList,
	))
	w.ShowAndRun()
}
