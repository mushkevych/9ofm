package controller

import (
	"github.com/rivo/tview"
	log "github.com/sirupsen/logrus"
)

type FilterEditListener func(string) error

// FilterController holds the UI objects and logic to allow real-time filtering of the files in File Panels by name
type FilterController struct {
	tviewApp  *tview.Application
	name      string
	graphicElement GraphicElement

	labelStr  string
	maxLength int
	isVisible bool

	filterEditListeners []FilterEditListener
}

// NewFilterController creates a new controller object attached the the global [gocui] screen object.
func NewFilterController(tviewApp *tview.Application) (controller *FilterController) {
	controller = new(FilterController)

	controller.filterEditListeners = make([]FilterEditListener, 0)

	// populate main fields
	controller.tviewApp = tviewApp
	controller.name = "filter"
	controller.labelStr = "AbsPath FilterController: "
	controller.isVisible = false

	return controller
}

func (c *FilterController) AddFilterEditListener(listener ...FilterEditListener) {
	c.filterEditListeners = append(c.filterEditListeners, listener...)
}

func (c *FilterController) Name() string {
	return c.name
}

// Edit intercepts the key press events in the filer controller to update the file controller in real time.
//func (c *FilterController) Edit(view *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
//	if !c.IsVisible() {
//		return
//	}
//
//	cx, _ := view.Cursor()
//	ox, _ := view.Origin()
//	limit := ox+cx+1 > c.maxLength
//	switch {
//	case ch != 0 && mod == 0 && !limit:
//		view.EditWrite(ch)
//	case key == gocui.KeySpace && !limit:
//		view.EditWrite(' ')
//	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
//		view.EditDelete(true)
//	}
//
//	// notify listeners
//	c.notifyFilterEditListeners()
//}

//func (c *FilterController) notifyFilterEditListeners() {
//	currentValue := strings.TrimSpace(c.view.Buffer())
//	for _, listener := range c.filterEditListeners {
//		err := listener(currentValue)
//		if err != nil {
//			// note: cannot propagate error from here since this is from the main gogui thread
//			log.Errorf("notifyFilterEditListeners: %+c", err)
//		}
//	}
//}
//
// Render flushes the state objects to the screen. Currently this is the users path filter input.
func (c *FilterController) Render() error {
	log.Tracef("controller.Render() %s", c.Name())

	//c.gui.Update(func(g *gocui.Gui) error {
	//	_, err := fmt.Fprintln(c.header, format.Header(c.labelStr))
	//	if err != nil {
	//		log.Error("unable to write to buffer: ", err)
	//	}
	//	return err
	//})
	return nil
}

// IsVisible indicates if the filter controller pane is currently initialized
func (c *FilterController) IsVisible() bool {
	if c == nil {
		return false
	}
	return c.isVisible
}

// SetVisible hides or shows the FileTree filter panel
func (c *FilterController) SetVisible(visible bool) error {
//	// delete all user input from the tree controller
//	c.view.Clear()
//
//	if c.isVisible = visible; c.isVisible {
//		_, err := c.gui.SetCurrentView(c.name)
//		if err != nil {
//			log.Error("unable to toggle filter controller: ", err)
//			return err
//		}
//		return nil
//	}
//
//	// reset the cursor for the next time it is visible
//	// Note: there is a subtle gocui behavior here where this cannot be called when the controller
//	// is newly visible. Is this a problem with 9ofm or gocui?
//	return c.view.SetCursor(0, 0)

	return nil
}

// GraphicElement returns UI graphicElement used by tview framework to render the UI interface
func (c *FilterController) GraphicElement() GraphicElement {
	return c.graphicElement
}
