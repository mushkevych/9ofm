package commander

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mushkevych/9ofm/commander/controller"
	"github.com/mushkevych/9ofm/commander/model"
	"github.com/rivo/tview"
	log "github.com/sirupsen/logrus"
)

type Application struct {
	tviewApp   *tview.Application
	AlphaPanel *controller.FilePanelController
	BetaPanel  *controller.FilePanelController
	BottomRow  *controller.FxxController
	flexLayout *tview.Flex
}

func (app *Application) Renderers() []controller.Renderer {
	return []controller.Renderer{
		app.AlphaPanel,
		app.BetaPanel,
		app.BottomRow,
	}
}

func NewApplication(tviewApp *tview.Application) (*Application, error) {
	application, err := buildControllers(tviewApp)
	if err != nil {
		return nil, err
	}

	err = application.buildLayout()
	if err != nil {
		return nil, err
	}

	err = application.registerGlobalKeymaps()
	if err != nil {
		return nil, err
	}
	return application, nil
}

func buildControllers(tviewApp *tview.Application) (*Application, error) {
	alphaFileTree, err := model.ReadFileTree("/")
	if err != nil {
		return nil, err
	}
	AlphaPanel, err := controller.NewFilePanelController(tviewApp, "alphaFilePanel", alphaFileTree)
	if err != nil {
		return nil, err
	}

	betaFileTree, err := model.ReadFileTree("/")
	if err != nil {
		return nil, err
	}
	BetaPanel, err := controller.NewFilePanelController(tviewApp, "betaFilePanel", betaFileTree)
	if err != nil {
		return nil, err
	}

	application := &Application{
		tviewApp:   tviewApp,
		AlphaPanel: AlphaPanel,
		BetaPanel:  BetaPanel,
		BottomRow:  controller.NewFxxController(tviewApp),
		flexLayout: tview.NewFlex(),
	}

	return application, nil
}

func (app *Application) buildLayout() error {
	app.flexLayout.
		AddItem(tview.NewBox().SetBorder(true).SetTitle("Header"), 1, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(app.AlphaPanel.GraphicElement(), 0, 1, false).
			AddItem(app.BetaPanel.GraphicElement(), 0, 1, false),
			0, 1, false).
		AddItem(app.BottomRow.GraphicElement(), 2, 1, false)

	app.tviewApp.SetRoot(app.flexLayout, true).SetFocus(app.AlphaPanel.GraphicElement())
	return nil
}

// Render flushes the state objects to the screen.
func (app *Application) Render() error {
	for _, renderer := range app.Renderers() {
		if renderer.IsVisible() {
			err := renderer.Render()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (app *Application) registerGlobalKeymaps() error {
	app.tviewApp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		var err error
		switch event.Key() {
		case tcell.KeyTab:
			err = app.ToggleActiveFilePanel()
		}

		if err != nil {
			log.WithError(err)
		}
		return event
	})
	return nil
}

// ToggleActiveFilePanel switches between the two file panels
func (app *Application) ToggleActiveFilePanel() (err error) {
	v := app.tviewApp.GetFocus()
	if v == nil || v == app.AlphaPanel.GraphicElement() {
		app.BottomRow.SetFilePanels(app.BetaPanel, app.AlphaPanel)
		app.tviewApp.SetFocus(app.BetaPanel.GraphicElement())
	} else {
		app.BottomRow.SetFilePanels(app.AlphaPanel, app.BetaPanel)
		app.tviewApp.SetFocus(app.AlphaPanel.GraphicElement())
	}

	return app.Render()
}
