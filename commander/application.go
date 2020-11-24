package commander

import (
	"github.com/mushkevych/9ofm/commander/configuration"
	"github.com/mushkevych/9ofm/commander/layout"
	"regexp"

	"github.com/mushkevych/9ofm/commander/controller"
	"github.com/mushkevych/9ofm/commander/model"
	"github.com/rivo/tview"
	log "github.com/sirupsen/logrus"
)

type Application struct {
	tviewApp     *tview.Application
	AlphaTree       *controller.FileTreeController
	BetaTree        *controller.FileTreeController
	activeFilePanel *controller.FileTreeController
	Status          *controller.StatusController
	BottomRow       *controller.FxxController
	Filter          *controller.FilterController
	Debug           *controller.DebugController
	LayoutManager   *layout.Manager
}

func (app *Application) Renderers() []controller.Renderer {
	return []controller.Renderer{
		app.AlphaTree,
		app.BetaTree,
		app.Status,
		app.BottomRow,
		app.Filter,
	}
}

func NewApplication(tviewApp *tview.Application) (*Application, error) {
	application, err := buildControllers(tviewApp)
	if err != nil {
		return nil, err
	}
	application, err = buildLayout(application)
	if err != nil {
		return nil, err
	}
	err = application.registerGlobalKeymaps()
	if err != nil {
		return nil, err
	}
	return application, nil
}

func buildControllers(app *tview.Application) (*Application, error) {
	alphaFileTree, err := model.ReadFileTree("/")
	if err != nil {
		return nil, err
	}
	AlphaTree, err := controller.NewFileTreeController(gui, "alphaFileTree", alphaFileTree)
	if err != nil {
		return nil, err
	}

	betaFileTree, err := model.ReadFileTree("/")
	if err != nil {
		return nil, err
	}
	BetaTree, err := controller.NewFileTreeController(gui, "betaFileTree", betaFileTree)
	if err != nil {
		return nil, err
	}

	application := &Application{
		gui:             gui,
		AlphaTree:       AlphaTree,
		BetaTree:        BetaTree,
		activeFilePanel: AlphaTree,
		Status:          controller.NewStatusController(gui),
		BottomRow:       controller.NewFxxController(gui),
		Filter:          controller.NewFilterController(gui),
		Debug:           controller.NewDebugController(gui),
	}

	// update the status pane when a model option is changed by the user
	application.AlphaTree.AddViewOptionChangeListener(application.onFileTreeViewOptionChange)
	application.BetaTree.AddViewOptionChangeListener(application.onFileTreeViewOptionChange)

	// update the tree application while the user types into the filter application
	application.Filter.AddFilterEditListener(application.onFilterEdit)

	return application, nil
}

func buildLayout(application *Application) (*Application, error) {
	application.LayoutManager = layout.NewManager()
	application.LayoutManager.Add(application.BottomRow, layout.LocationFooter)
	application.LayoutManager.Add(application.Status, layout.LocationFooter)
	application.LayoutManager.Add(application.Filter, layout.LocationFooter)
	application.LayoutManager.Add(application.BetaTree, layout.LocationColumn)
	application.LayoutManager.Add(application.AlphaTree, layout.LocationColumn)

	if configuration.Config.GetBoolOrDefault("debug", false) {
		application.LayoutManager.Add(application.Debug, layout.LocationColumn)
	}
	application.gui.Cursor = false
	//application.gui.Mouse = true
	application.gui.SetManagerFunc(application.LayoutManager.Layout)

	return application, nil
}

func (app *Application) registerGlobalKeymaps() error {
	var keymaps = []controller.KeymapDetail{
		{
			KeyboardShortcut: "Tab",
			OnAction:         app.ToggleActiveFilePanel,
			Display:          "Switch panes",
		},
		{
			KeyboardShortcut: "Ctrl+f",
			OnAction:         app.ShowFilterView,
			//IsSelected:       app.Filter.IsVisible,
			Display: "Filter",
		},
		{
			KeyboardShortcut: "Esc",
			OnAction:         app.HideFilterView,
			//IsSelected:       app.Filter.IsVisible,
			Display: "Filter",
		},
	}

	err := controller.RegisterKeymaps(app.gui, "", keymaps)
	if err != nil {
		return err
	}

	app.Status.AddKeymap(keymaps...)
	return nil
}

func (app *Application) onFileTreeViewOptionChange() error {
	err := app.Status.Update()
	if err != nil {
		return err
	}
	return app.Status.Render()
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

	err = app.activeFilePanel.Update()
	if err != nil {
		return err
	}

	return app.activeFilePanel.Render()
}

func (app *Application) UpdateAndRender() error {
	err := app.Update()
	if err != nil {
		log.Debug("failed update: ", err)
		return err
	}

	err = app.Render()
	if err != nil {
		log.Debug("failed render: ", err)
		return err
	}

	return nil
}

// Update refreshes the state objects for future rendering.
func (app *Application) Update() error {
	for _, renderer := range app.Renderers() {
		err := renderer.Update()
		if err != nil {
			log.Debug("unable to update appController: ")
			return err
		}
	}
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

// ToggleActiveFilePanel switches between the two file panels
func (app *Application) ToggleActiveFilePanel() (err error) {
	v := app.gui.CurrentView()
	if v == nil || v.Name() == app.AlphaTree.Name() {
		app.activeFilePanel = app.BetaTree
		app.BottomRow.SetFilePanels(app.BetaTree, app.AlphaTree)
	} else {
		app.activeFilePanel = app.AlphaTree
		app.BottomRow.SetFilePanels(app.AlphaTree, app.BetaTree)
	}

	app.Status.SetCurrentView(app.activeFilePanel)
	_, err = app.gui.SetCurrentView(app.activeFilePanel.Name())
	if err != nil {
		log.Error("unable to select Pane: ", app.activeFilePanel.Name())
		return err
	}
	return app.UpdateAndRender()
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

	return app.UpdateAndRender()
}

// HideFilterView shows the FilePanel filter UI components
func (app *Application) ShowFilterView() error {
	// delete all user input from the tree appController
	err := app.Filter.SetVisible(true)
	if err != nil {
		log.Error("unable to show Filter: ", err)
		return err
	}
	return app.UpdateAndRender()
}
