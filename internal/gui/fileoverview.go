package gui

import (
	"path/filepath"
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type FileOverview struct {
	widget.BaseWidget

	moveDown     *widget.Button
	moveUp       *widget.Button
	moveDownFull *widget.Button
	moveUpFull   *widget.Button
	list         *widget.List
	obj          *fyne.Container

	OnSelected   func(path string)
	OnUnselected func(path string)
	OnReorder    func()

	FileSelector *FileSelector

	paths []string
}

// Sets fileSelector.OnSelected and OnUnselected!
// Plase use FileOverview.OnSelected and OnUnselected instead.
func NewFileOverview(fileSelector *FileSelector) *FileOverview {
	fl := &FileOverview{
		FileSelector: fileSelector,
	}
	fl.ExtendBaseWidget(fl)
	return fl
}

func (fo *FileOverview) ExtendBaseWidget(w fyne.Widget) {
	fo.BaseWidget.ExtendBaseWidget(w)
	fo.list = widget.NewList(
		func() int {
			return len(fo.paths)
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
			item.Label.SetText(filepath.Base(fo.paths[id]))
			item.LabelIcon.SetResource(theme.FileImageIcon())
			item.IconButton.OnTapped = func() {
				fo.FileSelector.Unselect(fo.paths[id])
			}
		},
	)

	var moveFileItem func(up bool, full bool)
	fo.moveDown = widget.NewButtonWithIcon("", theme.MenuDropDownIcon(), func() { moveFileItem(false, false) })
	fo.moveUp = widget.NewButtonWithIcon("", theme.MenuDropUpIcon(), func() { moveFileItem(true, false) })
	fo.moveDownFull = widget.NewButtonWithIcon("", theme.MoveDownIcon(), func() { moveFileItem(false, true) })
	fo.moveUpFull = widget.NewButtonWithIcon("", theme.MoveUpIcon(), func() { moveFileItem(true, true) })
	selectedFileID := -1

	refreshButtons := func() {
		id := selectedFileID
		if len(fo.paths) == 0 {
			id = -1
		}
		if id == -1 || id == len(fo.paths)-1 {
			fo.moveDown.Disable()
			fo.moveDownFull.Disable()
		} else {
			fo.moveDown.Enable()
			fo.moveDownFull.Enable()
		}
		if id == -1 || id == 0 {
			fo.moveUp.Disable()
			fo.moveUpFull.Disable()
		} else {
			fo.moveUp.Enable()
			fo.moveUpFull.Enable()
		}
	}

	fo.list.OnSelected = func(id widget.ListItemID) {
		selectedFileID = id
		refreshButtons()
	}

	moveFileItem = func(up bool, full bool) {
		id := selectedFileID
		if id == -1 {
			return
		}
		defer func() {
			fo.list.Select(id)
			fo.list.Refresh()
			if fo.OnReorder != nil {
				fo.OnReorder()
			}
		}()
		for {
			if up {
				if id == 0 {
					return
				}
				fo.paths[id], fo.paths[id-1] = fo.paths[id-1], fo.paths[id]
				id = id - 1
			} else {
				if id == len(fo.paths)-1 {
					return
				}
				fo.paths[id], fo.paths[id+1] = fo.paths[id+1], fo.paths[id]
				id = id + 1
			}
			if !full {
				return
			}
		}
	}

	fo.FileSelector.OnSelected = func(path string) {
		if slices.Index[[]string, string](fo.paths, path) == -1 {
			fo.paths = append(fo.paths, path)
		}
		if fo.OnSelected != nil {
			fo.OnSelected(path)
		}
		refreshButtons()
		fo.list.Refresh()
	}

	fo.FileSelector.OnUnselected = func(path string) {
		if idx := slices.Index[[]string, string](fo.paths, path); idx != -1 {
			fo.paths = append(fo.paths[:idx], fo.paths[idx+1:]...)
		}
		if fo.OnUnselected != nil {
			fo.OnUnselected(path)
		}
		refreshButtons()
		fo.list.Refresh()
	}

	fo.obj = container.NewBorder(
		container.NewBorder(nil, nil, nil, container.NewHBox(
			fo.moveDown,
			fo.moveUp,
			fo.moveDownFull,
			fo.moveUpFull,
		)),
		nil, nil, nil,
		fo.list,
	)
}

func (fo *FileOverview) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(fo.obj)
}

func (fo *FileOverview) NumSelected() int {
	return len(fo.paths)
}

func (fo *FileOverview) Selected() []string {
	res := make([]string, len(fo.paths))
	copy(res, fo.paths)
	return res
}
