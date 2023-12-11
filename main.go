package main

import (
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	p4p "github.com/pic4pdf/lib-p4p"

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
	layoutMode := p4p.Fit
	scale := float64(1)

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
			if id < len(sel) {
				iv.SetDescription(fmt.Sprintf("%v/%v (%v)", id+1, len(sel), filepath.Base(sel[id])))
				if img, ok := imgs[sel[id]]; ok {
					iv.SetOptions(p4p.ImageOptions{
						Mode:  layoutMode,
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

	w.SetOnDropped(func(p fyne.Position, u []fyne.URI) {
		for _, files := range u {
			err := ValidateFile(files.Path())
			if err != nil {
				dialog.ShowError(fmt.Errorf("%v isn't a valid file.\nReason: %v", files.Path(), err), w)
				return
			}
			fileSel.Select(files.Path())
		}
	})

	fileOw.OnUnselected = func(path string) {
		delete(imgs, path)
		l.Refresh()
	}
	fileOw.OnReorder = func() {
		l.Refresh()
	}

	var options *widget.Accordion
	{
		var scaleSld *widget.Slider
		layoutModeSel := widget.NewSelect(
			[]string{"Center", "Fill", "Fit"},
			func(s string) {
				switch s {
				case "Center":
					layoutMode = p4p.Center
				case "Fill":
					layoutMode = p4p.Fill
				case "Fit":
					layoutMode = p4p.Fit
				}
				scaleSld.SetValue(1)
				l.Refresh()
			},
		)
		layoutModeSel.Selected = "Fit"
		scaleLabel := widget.NewLabel(fmt.Sprintf("%.1f", scale))
		var scaleReset *widget.Button
		scaleSld = widget.NewSlider(0.2, 4)
		scaleSld.Value = scale
		scaleSld.Step = 0.1
		scaleSld.OnChanged = func(v float64) {
			if v == 1 {
				scaleReset.Disable()
			} else {
				scaleReset.Enable()
			}
			scaleLabel.SetText(fmt.Sprintf("%.1f", v))
		}
		scaleSld.OnChangeEnded = func(v float64) {
			scale = v
			l.Refresh()
		}
		scaleReset = widget.NewButtonWithIcon("", theme.ContentUndoIcon(), func() {
			scaleSld.SetValue(1)
		})
		scaleReset.Disable()
		form := widget.NewForm(
			widget.NewFormItem("Layout Mode", layoutModeSel),
			widget.NewFormItem("Scale", container.NewBorder(nil, nil, scaleLabel, scaleReset, scaleSld)),
		)
		optsItem := widget.NewAccordionItem("Options", form)
		optsItem.Open = true
		options = widget.NewAccordion(optsItem)
	}

	split := container.NewHSplit(
		container.NewHSplit(
			fileSel,
			fileOw,
		),
		container.NewBorder(
			nil, options, nil, nil, l,
		),
	)
	split.Offset = 0.6

	w.SetContent(split)
	w.ShowAndRun()
}

func ValidateFile(path string) error {
	//Check if the file is valid, and then adds it to the selector
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	zeroBytes, _ := f.Stat()
	if zeroBytes.Size() <= 1 {
		return errors.New("file has 0 bytes")
	}

	/// Valid extensions: png, jpg, webp, jpeg
	validExt := []string{".png", ".jpg", ".jpeg", ".webp"}

	// Extract and normalize the file extension
	fileExt := strings.ToLower(filepath.Ext(path))

	// Check if the file extension is in the valid list
	valid := false
	for _, v := range validExt {
		if fileExt == strings.ToLower(v) {
			valid = true
			break
		}
	}

	if !valid {
		return errors.New("invalid file type")
	}
	f.Close()
	return nil
}
