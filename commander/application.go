package commander

import (
	"github.com/gdamore/tcell/v2"
	"regexp"

	"github.com/mushkevych/9ofm/commander/controller"
	"github.com/mushkevych/9ofm/commander/model"
	"github.com/rivo/tview"
	log "github.com/sirupsen/logrus"
)

type Application struct {
	tviewApp        *tview.Application
	AlphaTree       *controller.FileTreeController
	BetaTree        *controller.FileTreeController
	activeFilePanel *controller.FileTreeController
	BottomRow       *controller.FxxController
	Filter          *controller.FilterController
	LayoutManager   *LayoutManager
}

func (app *Application) Renderers() []controller.Renderer {
	return []controller.Renderer{
		app.AlphaTree,
		app.BetaTree,
		app.BottomRow,
		app.Filter,
	}
}

func NewApplication(tviewApp *tview.Application) (*Application, error) {
	application, err := buildControllers(tviewApp)
	if err != nil {
		return nil, err
	}
	application, err = buildLayout(tviewApp, application)
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
	AlphaTree, err := controller.NewFileTreeController(tviewApp, "alphaFileTree", alphaFileTree)
	if err != nil {
		return nil, err
	}

	betaFileTree, err := model.ReadFileTree("/")
	if err != nil {
		return nil, err
	}
	BetaTree, err := controller.NewFileTreeController(tviewApp, "betaFileTree", betaFileTree)
	if err != nil {
		return nil, err
	}

	application := &Application{
		tviewApp:        tviewApp,
		AlphaTree:       AlphaTree,
		BetaTree:        BetaTree,
		activeFilePanel: AlphaTree,
		BottomRow:       controller.NewFxxController(tviewApp),
		//Filter:          controller.NewFilterController(tviewApp),
	}

	// TODO: see if this can be removed, as well as AddOnChangeListener & onChangeListener
	// update the status pane when a model option is changed by the user
	//application.AlphaTree.AddOnChangeListener(application.onChangeListener)
	//application.BetaTree.AddOnChangeListener(application.onChangeListener)
	//
	// update the tree application while the user types into the filter application
	//application.Filter.AddFilterEditListener(application.onFilterEdit)

	return application, nil
}

func buildLayout(tviewApp *tview.Application, application *Application) (*Application, error) {
	application.LayoutManager = NewManager(tviewApp, application)
	application.LayoutManager.BuildLayout()

	return application, nil
}

func (app *Application) registerGlobalKeymaps() error {
	app.tviewApp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		var err error
		switch event.Key() {
		case tcell.KeyTab:
			err = app.ToggleActiveFilePanel()
		case tcell.KeyCtrlF:
			err = app.ShowFilterView()
		case tcell.KeyESC:
			err = app.HideFilterView()
		}

		if err != nil {
			log.WithError(err)
		}
		return event
	})
	return nil
}

func (app *Application) onChangeListener() error {
	return nil
}

func (app *Application) onFilterEdit(filter string) error {
	var filterRegex *regexp.Regexp
	var err error

	if len(filter) > 0 {
		filterRegex, err = regexp.Compile(filter)
		if err != nil {
			return err
		}
	}

	app.activeFilePanel.SetFilterRegex(filterRegex)
	return app.activeFilePanel.Render()
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

// ToggleActiveFilePanel switches between the two file panels
func (app *Application) ToggleActiveFilePanel() (err error) {
	v := app.tviewApp.GetFocus()
	if v == nil || v == app.AlphaTree.GraphicElement() {
		app.activeFilePanel = app.BetaTree
		app.BottomRow.SetFilePanels(app.BetaTree, app.AlphaTree)
	} else {
		app.activeFilePanel = app.AlphaTree
		app.BottomRow.SetFilePanels(app.AlphaTree, app.BetaTree)
	}

	return app.Render()
}

// HideFilterView hides the FilePanel filter UI components
func (app *Application) HideFilterView() error {
	// delete all user input from the tree appController
	err := app.Filter.SetVisible(false)
	if err != nil {
		log.Error("unable to hide Filter: ", err)
		return err
	}

	// ...remove any filter from the tree
	app.activeFilePanel.SetFilterRegex(nil)

	// ...adjust focus to a valid (visible) appController
	err = app.ToggleActiveFilePanel()
	if err != nil {
		log.Error("unable to toggle filter appController (back): ", err)
		return err
	}

	return app.Render()
}

// HideFilterView shows the FilePanel filter UI components
func (app *Application) ShowFilterView() error {
	// delete all user input from the tree appController
	err := app.Filter.SetVisible(true)
	if err != nil {
		log.Error("unable to show Filter: ", err)
		return err
	}
	return app.Render()
}
