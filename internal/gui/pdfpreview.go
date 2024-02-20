package gui

import (
	"fmt"
	"image"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	p4p "github.com/pic4pdf/lib-p4p"
)

type PDFPreview struct {
	widget.BaseWidget

	OnError  func(error)
	Layout   p4p.Mode
	Scale    float64
	Unit     p4p.Unit
	PageSize p4p.PageSize

	Overview *FileOverview

	imgs map[string]image.Image

	list *widget.List
}

// Sets ow.OnSelected, OnUnselected and OnReorder!
func NewPDFPreview(ow *FileOverview, unit p4p.Unit, pageSize p4p.PageSize) *PDFPreview {
	il := &PDFPreview{
		Layout:   p4p.Fit,
		Scale:    1,
		Overview: ow,
		Unit:     unit,
		PageSize: pageSize,
	}
	il.ExtendBaseWidget(il)
	return il
}

func (il *PDFPreview) SetLayout(mode p4p.Mode) {
	il.Layout = mode
	il.Refresh()
}

func (il *PDFPreview) SetScale(sc float64) {
	il.Scale = sc
	il.Refresh()
}

func (il *PDFPreview) SetUnit(u p4p.Unit) {
	il.Unit = u
	il.Refresh()
}

func (il *PDFPreview) SetPageSize(s p4p.PageSize) {
	il.PageSize = s
	il.Refresh()
}

func (il *PDFPreview) ExtendBaseWidget(w fyne.Widget) {
	il.BaseWidget.ExtendBaseWidget(w)
	il.imgs = make(map[string]image.Image)
	il.list = widget.NewList(
		func() int {
			return il.Overview.NumSelected()
		},
		func() fyne.CanvasObject {
			iv := NewPDFImageView(il.Unit, il.PageSize)
			iv.SetMinSize(fyne.NewSize(0, 300))
			return iv
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			sel := il.Overview.Selected()
			iv := obj.(*PDFImageView)
			if id < len(sel) {
				iv.SetDescription(fmt.Sprintf("%v/%v (%v)", id+1, len(sel), filepath.Base(sel[id])))
				if img, ok := il.imgs[sel[id]]; ok {
					iv.SetOptions(p4p.ImageOptions{
						Mode:  il.Layout,
						Scale: il.Scale,
					})
					iv.SetImage(img)
					iv.SetParams(il.Unit, il.PageSize)
				}
			}
		},
	)
	il.list.OnSelected = func(widget.ListItemID) {
		il.list.UnselectAll()
	}

	il.Overview.OnSelected = func(path string) {
		f, err := os.Open(path)
		if err != nil {
			if il.OnError != nil {
				il.OnError(err)
			}
			return
		}
		defer f.Close()
		img, _, err := image.Decode(f)
		if err != nil {
			if il.OnError != nil {
				il.OnError(fmt.Errorf("invalid image '%v': %w", filepath.Base(path), err))
			}
			return
		}
		il.imgs[path] = img
		il.list.Refresh()
	}
	il.Overview.OnUnselected = func(path string) {
		delete(il.imgs, path)
		il.list.Refresh()
	}
	il.Overview.OnReorder = func() {
		il.list.Refresh()
	}
}

func (il *PDFPreview) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(il.list)
}
