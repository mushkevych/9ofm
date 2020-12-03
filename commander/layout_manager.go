package commander

import (
	"github.com/rivo/tview"
)

type GraphicElement interface {

}

type LayoutManager struct {
	tviewApp    *tview.Application
	flexLayout  *tview.Flex
	application *Application
}

func NewManager(tviewApp *tview.Application, application *Application) *LayoutManager {
	return &LayoutManager{
		tviewApp:    tviewApp,
		flexLayout:  tview.NewFlex(),
		application: application,
	}
}

func (lm *LayoutManager) renderAllPrimitives() error {
	for _, element := range lm.application.Renderers() {
		err := element.Render()
		if err != nil {
			return err
		}
	}
	return nil
}

func (lm *LayoutManager) BuildLayout() error {
	//lm.flexLayout.
	//AddItem(tview.NewBox().SetBorder(true).SetTitle("Header"), 1, 1, false).
	//AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
	//	AddItem(tview.NewBox().SetBorder(true).SetTitle("Left Panel"), 0, 1, false).
	//	AddItem(tview.NewBox().SetBorder(true).SetTitle("Right Panel"), 0, 1, false),
	//	0, 1, false).
	//AddItem(tview.NewBox().SetBorder(true).SetTitle("Fxx (1 row)"), 1, 1, false)

	lm.flexLayout.
		AddItem(tview.NewBox().SetBorder(true).SetTitle("Header"), 1, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(lm.application.AlphaPanel.GraphicElement(), 0, 1, false).
			AddItem(lm.application.BetaPanel.GraphicElement(), 0, 1, false),
			0, 1, false).
		AddItem(lm.application.BottomRow.GraphicElement(), 1, 1, false)

	lm.tviewApp.SetRoot(lm.flexLayout, true).SetFocus(lm.flexLayout)
	return nil
}
