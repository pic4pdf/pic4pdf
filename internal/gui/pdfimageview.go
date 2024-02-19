package gui

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	p4p "github.com/pic4pdf/lib-p4p"
	"golang.org/x/image/draw"
)

type PDFImageView struct {
	widget.BaseWidget

	imgOpts  p4p.ImageOptions
	minSize  fyne.Size
	imgData  image.Image
	unit     p4p.Unit
	pageSize p4p.PageSize
	// Max image size in pixels, for rendering optimization
	maxImgW int
	maxImgH int
	// Image layout in PDF units
	imgX float64
	imgY float64
	imgW float64
	imgH float64

	img      *canvas.Image
	desc     *widget.Label
	descRect *canvas.Rectangle
	lock     sync.Mutex
}

// Requires iv.lock to be locked!
func (iv *PDFImageView) getConvPageSize() (w, h float64) {
	s := iv.pageSize.Convert(iv.unit)
	return s.W, s.H
}

func (iv *PDFImageView) rerenderImage() {
	fmt.Printf("Rerender (rand ID: %02x)\n", rand.Intn(0xFF))
	iv.lock.Lock()
	if iv.imgData == nil {
		iv.lock.Unlock()
		return
	}
	img := iv.imgData
	pxBounds := img.Bounds()
	x, y, w, h, cropX1, cropY1, cropX2, cropY2, crop := p4p.Render(iv.pageSize, iv.unit, pxBounds.Dx(), pxBounds.Dy(), iv.imgOpts)
	// Crop image.
	if crop {
		if subImg, ok := img.(interface {
			SubImage(r image.Rectangle) image.Image
		}); ok {
			img = subImg.SubImage(image.Rect(cropX1, cropY1, cropX2+1, cropY2+1))
		} else {
			panic("image must support SubImage")
		}
	}
	// Calculate image coords.
	{
		pgW, pgH := iv.getConvPageSize()
		iv.imgX = math.Max(0, x)
		iv.imgY = math.Max(0, y)
		iv.imgW = math.Min(pgW, w)
		iv.imgH = math.Min(pgH, h)
	}
	// Downscale image.
	{
		src := img
		srcSz := src.Bounds().Size()
		dstSz := srcSz
		if dstSz.X > iv.maxImgW {
			dstSz.X = iv.maxImgW
			dstSz.Y = int(float64(srcSz.Y) * (float64(dstSz.X) / float64(srcSz.X)))
		}
		if dstSz.Y > iv.maxImgH {
			dstSz.Y = iv.maxImgH
			dstSz.X = int(float64(srcSz.X) * (float64(dstSz.Y) / float64(srcSz.Y)))
		}
		dst := image.NewRGBA(image.Rect(0, 0, dstSz.X, dstSz.Y))
		draw.NearestNeighbor.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)
		img = dst
	}
	iv.img = canvas.NewImageFromImage(img)
	iv.lock.Unlock()
}

func NewPDFImageView(unit p4p.Unit, pageSize p4p.PageSize) *PDFImageView {
	iv := &PDFImageView{
		desc:     widget.NewLabel(""),
		unit:     unit,
		pageSize: pageSize,
		maxImgW:  600,
		maxImgH:  600,
	}
	iv.ExtendBaseWidget(iv)
	return iv
}

func (iv *PDFImageView) SetMaxImageRenderSize(w, h int) {
	iv.lock.Lock()
	iv.maxImgW = w
	iv.maxImgH = h
	iv.lock.Unlock()
	iv.Refresh()
}

func (iv *PDFImageView) SetMinSize(size fyne.Size) {
	iv.lock.Lock()
	iv.minSize = size
	iv.lock.Unlock()
	iv.Refresh()
}

// Will update only if the parameters differ from the previous ones.
func (iv *PDFImageView) SetParams(unit p4p.Unit, pageSize p4p.PageSize) {
	iv.lock.Lock()
	if unit == iv.unit && pageSize == iv.pageSize {
		iv.lock.Unlock()
		return
	}
	iv.unit = unit
	iv.pageSize = pageSize
	iv.lock.Unlock()
	iv.rerenderImage()
	iv.Refresh()
}

// Will update only if opts differ from the previous options.
func (iv *PDFImageView) SetOptions(opts p4p.ImageOptions) {
	iv.lock.Lock()
	if iv.imgOpts == opts {
		iv.lock.Unlock()
		return
	}
	iv.imgOpts = opts
	iv.lock.Unlock()
	iv.rerenderImage()
	iv.Refresh()
}

// Will only update if img differs from the previous image.
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
	iv.descRect = canvas.NewRectangle(color.RGBA{128, 128, 128, 235})
	iv.descRect.CornerRadius = 4
}

func (iv *PDFImageView) CreateRenderer() fyne.WidgetRenderer {
	r := &pdfImageViewRenderer{
		iv: iv,
		bg: canvas.NewRectangle(color.RGBA{255, 255, 255, 255}),
	}
	return r
}

func (iv *PDFImageView) MouseIn(e *desktop.MouseEvent) {
}

func (iv *PDFImageView) MouseMoved(e *desktop.MouseEvent) {
	if e.Position.Y < 50 {
		iv.desc.Hide()
		iv.descRect.Hide()
	} else {
		iv.desc.Show()
		iv.descRect.Show()
	}
}

func (iv *PDFImageView) MouseOut() {
	iv.desc.Show()
	iv.descRect.Show()
}

type pdfImageViewRenderer struct {
	iv       *PDFImageView
	bg       *canvas.Rectangle
	descRect *canvas.Rectangle

	minSize fyne.Size
}

func (r *pdfImageViewRenderer) refreshMinSize() {
	r.iv.lock.Lock()
	pgW, pgH := r.iv.getConvPageSize()
	if r.iv.minSize.Width == 0 {
		r.minSize = fyne.NewSize(r.iv.minSize.Height*float32(pgW)/float32(pgH), r.iv.minSize.Height)
	} else if r.iv.minSize.Height == 0 {
		r.minSize = fyne.NewSize(r.iv.minSize.Width, r.iv.minSize.Width*float32(pgH)/float32(pgW))
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
	pgW, pgH := r.iv.getConvPageSize()
	var effSize fyne.Size
	if float32(pgW/pgH) > size.Width/size.Height {
		effSize = fyne.NewSize(size.Width, size.Width*float32(pgH)/float32(pgW))
	} else {
		effSize = fyne.NewSize(size.Height*float32(pgW)/float32(pgH), size.Height)
	}
	oX, oY := (size.Width-effSize.Width)/2, (size.Height-effSize.Height)/2
	if r.iv.img != nil {
		r.iv.img.Move(fyne.NewPos(
			float32(r.iv.imgX/pgW)*effSize.Width,
			float32(r.iv.imgY/pgH)*effSize.Height,
		).AddXY(oX, oY))
		r.iv.img.Resize(fyne.NewSize(
			float32(r.iv.imgW/pgW)*effSize.Width,
			float32(r.iv.imgH/pgH)*effSize.Height,
		))
	}
	r.iv.desc.Move(fyne.NewPos(5, 5).AddXY(oX, oY))
	r.iv.descRect.Move(fyne.NewPos(5, 5).AddXY(oX, oY))
	r.iv.descRect.Resize(r.iv.desc.MinSize())
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
	objs = append(objs, r.iv.descRect, r.iv.desc)
	r.iv.lock.Unlock()
	return objs
}
