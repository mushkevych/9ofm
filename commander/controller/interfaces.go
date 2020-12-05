package controller

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Controller defines a renderable terminal screen pane.
type Renderer interface {
	Name() string
	Render() error
	IsVisible() bool
	SetVisible(bool) error
	GraphicElement() GraphicElement
}

type GraphicElement interface {
	tview.Primitive
	SetInputCapture(capture func(event *tcell.EventKey) *tcell.EventKey) *tview.Box
	GetInputCapture() func(event *tcell.EventKey) *tcell.EventKey
	SetTitle(title string) *tview.Box
	GetTitle() string
}
