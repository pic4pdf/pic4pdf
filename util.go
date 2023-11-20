package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type tappableCard struct {
	widget.Card
}

func newTappableCard(title, subtitle string, content fyne.CanvasObject) *tappableCard {
	card := &tappableCard{}
	card.ExtendBaseWidget(card)
	card.SetTitle(title)
	card.SetSubTitle(subtitle)
	card.SetContent(content)

	return card
}

func (t *tappableCard) Tapped(_ *fyne.PointEvent) {
	//thing should go here
	fmt.Println("tapped")
	start()
}

func (t *tappableCard) TappedSecondary(_ *fyne.PointEvent) {

}
