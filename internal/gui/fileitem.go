package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type FileItem struct {
	widget.BaseWidget

	obj *fyne.Container

	LabelIcon   *widget.Icon
	Label       *widget.Label
	LabelButton *widget.Button
	IconButton  *widget.Button
	Overlay     *canvas.Rectangle
}

func (itm *FileItem) overlayColor() color.Color {
	s := fyne.CurrentApp().Settings()
	return s.Theme().Color(
		theme.ColorNameSelection,
		s.ThemeVariant(),
	)
}

func newFileItem(label string, onTapped func(), icon fyne.Resource, onIconTapped func()) *FileItem {
	itm := &FileItem{
		LabelIcon:   widget.NewIcon(nil),
		Label:       widget.NewLabel(label),
		LabelButton: widget.NewButton("", onTapped),
		IconButton:  widget.NewButtonWithIcon("", icon, onIconTapped),
	}
	itm.ExtendBaseWidget(itm)
	return itm
}

func (itm *FileItem) ExtendBaseWidget(w fyne.Widget) {
	itm.BaseWidget.ExtendBaseWidget(w)
	itm.Label.Truncation = fyne.TextTruncateEllipsis
	itm.Overlay = canvas.NewRectangle(itm.overlayColor())
	itm.Overlay.Hide()
	spaceL := canvas.NewRectangle(color.Transparent)
	spaceL.SetMinSize(fyne.NewSize(2, 0))
	itm.obj = container.NewStack(
		itm.LabelButton,
		container.NewBorder(
			nil, nil, spaceL, nil,
			container.NewBorder(
				nil, nil,
				itm.LabelIcon, itm.IconButton,
				itm.Label,
			),
		),
		itm.Overlay,
	)
}

func (itm *FileItem) Refresh() {
	itm.Overlay.FillColor = itm.overlayColor()
	itm.obj.Refresh()
}

func (itm *FileItem) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(itm.obj)
}
