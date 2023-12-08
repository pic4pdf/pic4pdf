package gui

import (
	"image"
	"image/color"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
	p4p "github.com/pic4pdf/lib-p4p"
)

type PDFImageView struct {
	widget.BaseWidget

	pdf *p4p.P4P
	imgOpts p4p.ImageOptions
	minSize fyne.Size
	imgData image.Image
	unit p4p.Unit
	pageSize p4p.PageSize

	img *canvas.Image
	desc *widget.Label
	lock sync.Mutex
}

func (iv *PDFImageView) rerenderImage() {
	iv.lock.Lock()
	if iv.imgData == nil {
		iv.lock.Unlock()
		return
	}
	var img image.Image
	pxBounds := iv.imgData.Bounds()
	if x1, y1, x2, y2, mustCrop := iv.pdf.CalcImageCropCoords(pxBounds.Dx(), pxBounds.Dy(), iv.imgOpts); mustCrop {
		if subImg, ok := iv.imgData.(interface{SubImage(r image.Rectangle) image.Image}); ok {
			img = subImg.SubImage(image.Rect(x1, y1, x2, y2))
		} else {
			panic("image must support SubImage")
		}
	} else {
		img = iv.imgData
	}
	iv.img = canvas.NewImageFromImage(img)
	iv.lock.Unlock()
}

func NewPDFImageView(unit p4p.Unit, pageSize p4p.PageSize) *PDFImageView {
	iv := &PDFImageView{
		pdf: p4p.New(unit, pageSize),
		desc: widget.NewLabel(""),
		unit: unit,
		pageSize: pageSize,
	}
	iv.ExtendBaseWidget(iv)
	return iv
}

func (iv *PDFImageView) SetMinSize(size fyne.Size) {
	iv.lock.Lock()
	iv.minSize = size
	iv.lock.Unlock()
	iv.Refresh()
}

func (iv *PDFImageView) SetParams(unit p4p.Unit, pageSize p4p.PageSize) {
	iv.lock.Lock()
	if unit == iv.unit && pageSize == iv.pageSize {
		iv.lock.Unlock()
		return
	}
	iv.pdf = p4p.New(unit, pageSize)
	iv.unit = unit
	iv.pageSize = pageSize
	iv.lock.Unlock()
	iv.rerenderImage()
	iv.Refresh()
}

func (iv *PDFImageView) SetOptions(opts p4p.ImageOptions) {
	iv.lock.Lock()
	iv.imgOpts = opts
	iv.lock.Unlock()
	iv.rerenderImage()
	iv.Refresh()
}

func (iv *PDFImageView) SetImage(img image.Image) {
	iv.lock.Lock()
	if iv.imgData == img {
		iv.lock.Unlock()
		return
	}
	iv.imgData = img
	iv.lock.Unlock()
	iv.rerenderImage()
	iv.Refresh()
}

func (iv *PDFImageView) SetDescription(text string) {
	iv.lock.Lock()
	iv.desc.SetText(text)
	iv.lock.Unlock()
}

func (iv *PDFImageView) ExtendBaseWidget(w fyne.Widget) {
	iv.BaseWidget.ExtendBaseWidget(w)
}

func (iv *PDFImageView) CreateRenderer() fyne.WidgetRenderer {
	r := &pdfImageViewRenderer{
		iv: iv,
		bg: canvas.NewRectangle(color.RGBA{255, 255, 255, 255}),
		descRect: canvas.NewRectangle(color.RGBA{128, 128, 128, 235}),
	}
	r.descRect.CornerRadius = 4
	return r
}

type pdfImageViewRenderer struct {
	iv *PDFImageView
	bg *canvas.Rectangle
	descRect *canvas.Rectangle

	minSize fyne.Size
}

func (r *pdfImageViewRenderer) refreshMinSize() {
	r.iv.lock.Lock()
	pgW, pgH := r.iv.pdf.PageSize()
	if r.iv.minSize.Width == 0 {
		r.minSize = fyne.NewSize(r.iv.minSize.Height * float32(pgW) / float32(pgH), r.iv.minSize.Height)
	} else if r.iv.minSize.Height == 0 {
		r.minSize = fyne.NewSize(r.iv.minSize.Width, r.iv.minSize.Width * float32(pgH) / float32(pgW))
	} else {
		r.minSize = r.iv.minSize
	}
	r.iv.lock.Unlock()
}

func (r *pdfImageViewRenderer) Destroy() {
}

func (r *pdfImageViewRenderer) Layout(size fyne.Size) {
	r.refreshMinSize()
	r.iv.lock.Lock()
	pgW, pgH := r.iv.pdf.PageSize()
	var effSize fyne.Size
	if float32(pgW/pgH) > size.Width/size.Height {
		effSize = fyne.NewSize(size.Width, size.Width * float32(pgH) / float32(pgW))
	} else {
		effSize = fyne.NewSize(size.Height * float32(pgW) / float32(pgH), size.Height)
	}
	oX, oY := (size.Width-effSize.Width)/2, (size.Height-effSize.Height)/2
	if r.iv.img != nil {
		pxBounds := r.iv.img.Image.Bounds()
		imgX, imgY, imgW, imgH := r.iv.pdf.CalcImageLayout(pxBounds.Dx(), pxBounds.Dy(), r.iv.imgOpts)
		r.iv.img.Move(fyne.NewPos(
			float32(imgX / pgW) * effSize.Width,
			float32(imgY / pgH) * effSize.Height,
		).AddXY(oX, oY))
		r.iv.img.Resize(fyne.NewSize(
			float32(imgW / pgW) * effSize.Width,
			float32(imgH / pgH) * effSize.Height,
		))
	}
	r.iv.desc.Move(fyne.NewPos(5, 5).AddXY(oX, oY))
	r.descRect.Move(fyne.NewPos(5, 5).AddXY(oX, oY))
	r.descRect.Resize(r.iv.desc.MinSize())
	r.iv.lock.Unlock()
	r.bg.Move(fyne.NewPos(oX, oY))
	r.bg.Resize(effSize)
}

func (r *pdfImageViewRenderer) MinSize() fyne.Size {
	r.refreshMinSize()
	return r.minSize
}

func (r *pdfImageViewRenderer) Refresh() {
}

func (r *pdfImageViewRenderer) Objects() []fyne.CanvasObject {
	objs := []fyne.CanvasObject{
		r.bg,
	}
	r.iv.lock.Lock()
	if r.iv.img != nil {
		objs = append(objs, r.iv.img)
	}
	objs = append(objs, r.descRect, r.iv.desc)
	r.iv.lock.Unlock()
	return objs
}
