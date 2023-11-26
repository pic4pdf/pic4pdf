package gui

import (
	"path/filepath"
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type FileOverview struct {
	widget.BaseWidget

	moveDown *widget.Button
	moveUp *widget.Button
	moveDownFull *widget.Button
	moveUpFull *widget.Button
	list *widget.List
	obj *fyne.Container

	FileSelector *FileSelector

	paths []string
}

// Sets fileSelector.OnSelected and OnUnselected!
func NewFileOverview(fileSelector *FileSelector) *FileOverview {
	fl := &FileOverview{
		FileSelector: fileSelector,
	}
	fl.ExtendBaseWidget(fl)
	return fl
}

func (fl *FileOverview) ExtendBaseWidget(w fyne.Widget) {
	fl.BaseWidget.ExtendBaseWidget(w)
	fl.list = widget.NewList(
		func() int {
			return len(fl.paths)
		}, func() fyne.CanvasObject {
			item := newFileItem(
				"PLACEHOLDER",
				nil,
				theme.ContentRemoveIcon(),
				nil,
			)
			item.LabelButton.Hide()
			return item
		}, func(id widget.ListItemID, obj fyne.CanvasObject) {
			item := obj.(*FileItem)
			item.Label.SetText(filepath.Base(fl.paths[id]))
			item.LabelIcon.SetResource(theme.FileImageIcon())
			item.IconButton.OnTapped = func() {
				fl.FileSelector.Unselect(fl.paths[id])
			}
		},
	)

	var moveFileItem func(up bool, full bool)
	fl.moveDown = widget.NewButtonWithIcon("", theme.MenuDropDownIcon(), func() {moveFileItem(false, false)})
	fl.moveUp = widget.NewButtonWithIcon("", theme.MenuDropUpIcon(), func() {moveFileItem(true, false)})
	fl.moveDownFull = widget.NewButtonWithIcon("", theme.MoveDownIcon(), func() {moveFileItem(false, true)})
	fl.moveUpFull = widget.NewButtonWithIcon("", theme.MoveUpIcon(), func() {moveFileItem(true, true)})
	selectedFileID := -1

	fl.list.OnSelected = func(id widget.ListItemID) {
		if id == len(fl.paths)-1 {
			fl.moveDown.Disable()
			fl.moveDownFull.Disable()
		} else {
			fl.moveDown.Enable()
			fl.moveDownFull.Enable()
		}
		if id == 0 {
			fl.moveUp.Disable()
			fl.moveUpFull.Disable()
		} else {
			fl.moveUp.Enable()
			fl.moveUpFull.Enable()
		}
		selectedFileID = id
	}

	moveFileItem = func(up bool, full bool) {
		id := selectedFileID
		if id == -1 {
			return
		}
		defer func() {
			fl.list.Select(id)
			fl.list.Refresh()
		}()
		for {
			if up {
				if id == 0 {
					return
				}
				fl.paths[id], fl.paths[id-1] = fl.paths[id-1], fl.paths[id]
				id = id-1
			} else {
				if id == len(fl.paths)-1 {
					return
				}
				fl.paths[id], fl.paths[id+1] = fl.paths[id+1], fl.paths[id]
				id = id+1
			}
			if !full {
				return
			}
		}
	}

	fl.FileSelector.OnSelected = func(path string) {
		if slices.Index[[]string, string](fl.paths, path) == -1 {
			fl.paths = append(fl.paths, path)
		}
		fl.list.Refresh()
	}

	fl.FileSelector.OnUnselected = func(path string) {
		if idx := slices.Index[[]string, string](fl.paths, path); idx != -1 {
			fl.paths = append(fl.paths[:idx], fl.paths[idx+1:]...)
		}
		fl.list.Refresh()
	}

	fl.obj = container.NewBorder(
		container.NewBorder(nil, nil, nil, container.NewHBox(
			fl.moveDown,
			fl.moveUp,
			fl.moveDownFull,
			fl.moveUpFull,
		)),
		nil, nil, nil,
		fl.list,
	)
}

func (fl *FileOverview) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(fl.obj)
}
