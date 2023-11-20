package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/ncruces/zenity"
)

func main() {
	pic4pdfApp := app.NewWithID("com.pic4pdf")
	pic4pdfWin := pic4pdfApp.NewWindow("Pic4Pdf")
	pic4pdfWin.Resize(fyne.Size{Width: 800, Height: 600})
	pic4pdfWin.CenterOnScreen()

	infoTextTitle := widget.NewRichTextFromMarkdown("# Pic4Pdf")
	infoExit := widget.NewButtonWithIcon("", theme.CancelIcon(), nil)
	infoExit.Importance = widget.DangerImportance
	infoHeader := container.NewBorder(nil, nil, infoTextTitle, nil)
	infoText := widget.NewLabelWithStyle("Create PDFs from images, no bullshit involved.", fyne.TextAlignCenter, widget.RichTextStyleEmphasis.TextStyle)
	newFileIcon := canvas.NewImageFromResource(theme.FileApplicationIcon())
	newFileIcon.FillMode = canvas.ImageFillContain
	newFileIcon.SetMinSize(fyne.Size{Width: 128, Height: 128})
	labelNewFile := widget.NewLabel("New File")
	labelNewFile.Alignment = fyne.TextAlignCenter
	cardNewFile := newTappableCard("", "", labelNewFile)
	cardNewFile.SetImage(newFileIcon)

	info := widget.NewModalPopUp(container.NewBorder(infoHeader, nil, nil, nil, container.NewVBox(infoText, cardNewFile)), pic4pdfWin.Canvas())
	infoExit.OnTapped = func() { info.Hide() }
	info.Show()

	showPop := widget.NewHyperlink("show", nil)
	showPop.OnTapped = func() { info.Show() }
	pic4pdfWin.SetContent(container.NewCenter(showPop))

	pic4pdfWin.Show()
	pic4pdfApp.Run()
}

func start() {
	selectMultipleImages, err := zenity.SelectFileMultiple(zenity.Title("Select your image(s)"), zenity.DisallowEmpty(), zenity.FileFilters{{"Images", []string{"*.png", "*.jpg", "*.jpeg"}, true}})
	if err != nil && err != zenity.ErrCanceled {
		panic("Something went very wrong")
	}
	fmt.Println(selectMultipleImages)
}
