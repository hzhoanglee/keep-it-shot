package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func playground() {
	a := app.New()
	w := a.NewWindow("TODO App")

	w.Resize(fyne.NewSize(300, 400))

	w.ShowAndRun()
}
