package main

import (
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"path/filepath"
	"strconv"
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

	validFilename := func(name string) bool {
		ext := strings.ToLower(filepath.Ext(name))
		return ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".webp"
	}

	fileSel := gui.NewFileSelectorPersistent("Main")
	fileSel.SetValidFilename(validFilename)
	closeWatcher := fileSel.CreateSimpleWatcher()
	defer closeWatcher()

	fileOw := gui.NewFileOverview(fileSel)

	pv := gui.NewPDFPreview(fileOw, p4p.Millimeter, p4p.A4())
	pv.OnError = func(err error) {
		dialog.ShowError(err, w)
	}

	w.SetOnDropped(func(_ fyne.Position, uris []fyne.URI) {
		added := 0
		for _, uri := range uris {
			path := uri.Path()
			if !validFilename(path) {
				continue
			}
			added++
			fileSel.Select(path)
		}
		if added != len(uris) {
			n := len(uris) - added
			var e error
			if n == 1 {
				e = fmt.Errorf("could not add file with unsupported format")
			} else {
				e = fmt.Errorf("could not add %v files with unsupported format", n)
			}
			dialog.ShowError(e, w)
		}
	})

	var options *widget.Accordion
	{
		var scaleSld *widget.Slider
		layoutModeSel := widget.NewSelect(
			[]string{"Center", "Fill", "Fit"},
			func(s string) {
				switch s {
				case "Center":
					pv.SetLayout(p4p.Center)
				case "Fill":
					pv.SetLayout(p4p.Fill)
				case "Fit":
					pv.SetLayout(p4p.Fit)
				}
				scaleSld.SetValue(1)
			},
		)
		layoutModeSel.Selected = "Fit"
		scaleLabel := widget.NewLabel(fmt.Sprintf("%.1f", pv.Scale))
		var scaleReset *widget.Button
		scaleSld = widget.NewSlider(0.2, 4)
		scaleSld.Value = pv.Scale
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
			pv.SetScale(v)
		}
		scaleReset = widget.NewButtonWithIcon("", theme.ContentUndoIcon(), func() {
			scaleSld.SetValue(1)
		})
		scaleReset.Disable()
		var pageSizeSel *widget.Select
		pageSizeW := widget.NewEntry()
		pageSizeW.Scroll = container.ScrollNone
		pageSizeW.Wrapping = fyne.TextWrapOff
		pageSizeW.OnChanged = func(s string) {
			pageSizeSel.SetSelected("Custom")
			v, _ := strconv.ParseFloat(s, 64)
			ps := pv.PageSize.Convert(pv.Unit)
			ps.W = v
			pv.SetPageSize(ps)
		}
		pageSizeH := widget.NewEntry()
		pageSizeH.Scroll = container.ScrollNone
		pageSizeH.Wrapping = fyne.TextWrapOff
		pageSizeH.OnChanged = func(s string) {
			pageSizeSel.SetSelected("Custom")
			v, _ := strconv.ParseFloat(s, 64)
			ps := pv.PageSize.Convert(pv.Unit)
			ps.H = v
			pv.SetPageSize(ps)
		}
		updatePageSize := func() {
			ps := pv.PageSize.Convert(pv.Unit)
			pageSizeW.Text = strconv.FormatFloat(math.Round(ps.W*100)/100, 'f', -1, 64)
			pageSizeW.Refresh()
			pageSizeH.Text = strconv.FormatFloat(math.Round(ps.H*100)/100, 'f', -1, 64)
			pageSizeH.Refresh()
		}
		pageSizeRotate := widget.NewButtonWithIcon("", theme.MediaReplayIcon(), func() {
			pv.SetPageSize(pv.PageSize.Rotate())
			updatePageSize()
		})
		pageSizeUnitSel := widget.NewSelect(
			[]string{"pt", "mm", "cm", "in"},
			func(s string) {
				switch s {
				case "pt":
					pv.SetUnit(p4p.Point)
				case "mm":
					pv.SetUnit(p4p.Millimeter)
				case "cm":
					pv.SetUnit(p4p.Centimeter)
				case "in":
					pv.SetUnit(p4p.Inch)
				}
				updatePageSize()
			},
		)
		pageSizeUnitSel.Selected = "mm"
		pageSizeSel = widget.NewSelect(
			[]string{"A4", "A5", "A6", "Letter", "Legal", "Tabloid", "A3", "A2", "A1", "Custom"},
			func(s string) {
				switch s {
				case "A4":
					pv.SetPageSize(p4p.A4())
				case "A5":
					pv.SetPageSize(p4p.A5())
				case "A6":
					pv.SetPageSize(p4p.A6())
				case "Letter":
					pv.SetPageSize(p4p.Letter())
				case "Legal":
					pv.SetPageSize(p4p.Legal())
				case "Tabloid":
					pv.SetPageSize(p4p.Tabloid())
				case "A3":
					pv.SetPageSize(p4p.A3())
				case "A2":
					pv.SetPageSize(p4p.A2())
				case "A1":
					pv.SetPageSize(p4p.A1())
				case "Custom":
					return
				}
				updatePageSize()
			},
		)
		pageSizeSel.SetSelected("A4")
		pageSizeCustomize := container.NewBorder(
			nil, nil,
			container.NewHBox(pageSizeW, widget.NewLabel("x"), pageSizeH),
			nil,
			container.NewBorder(nil, nil, nil, pageSizeRotate, pageSizeUnitSel),
		)
		form := widget.NewForm(
			widget.NewFormItem("Page", container.NewVBox(pageSizeSel, pageSizeCustomize)),
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
			nil, options, nil, nil, pv,
		),
	)
	split.Offset = 0.6

	w.SetContent(split)
	w.ShowAndRun()
}
