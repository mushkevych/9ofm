package controller

import (
	"fmt"
	"strings"

	"github.com/jroimartin/gocui"
	"github.com/mushkevych/9ofm/commander/format"
	"github.com/mushkevych/9ofm/utils"
	log "github.com/sirupsen/logrus"
)

type FilterEditListener func(string) error

// FilterController holds the UI objects and logic to allow real-time filtering of the files in File Panels by name
type FilterController struct {
	name            string
	gui             *gocui.Gui
	view            *gocui.View
	header          *gocui.View
	labelStr        string
	maxLength       int
	isVisible       bool
	requestedHeight int

	filterEditListeners []FilterEditListener
}

// NewFilterController creates a new controller object attached the the global [gocui] screen object.
func NewFilterController(gui *gocui.Gui) (controller *FilterController) {
	controller = new(FilterController)

	controller.filterEditListeners = make([]FilterEditListener, 0)

	// populate main fields
	controller.name = "filter"
	controller.gui = gui
	controller.labelStr = "AbsPath FilterController: "
	controller.isVisible = false

	controller.requestedHeight = 1

	return controller
}

func (c *FilterController) AddFilterEditListener(listener ...FilterEditListener) {
	c.filterEditListeners = append(c.filterEditListeners, listener...)
}

func (c *FilterController) Name() string {
	return c.name
}

// Setup initializes the UI concerns within the context of a global [gocui] controller object.
func (c *FilterController) Setup(view *gocui.View, header *gocui.View) error {
	log.Tracef("controller.Setup() %s", c.Name())

	// set controller options
	c.view = view
	c.maxLength = 200
	c.view.Frame = false
	c.view.BgColor = gocui.AttrReverse
	c.view.Editable = true
	c.view.Editor = c

	c.header = header
	c.header.BgColor = gocui.AttrReverse
	c.header.Editable = false
	c.header.Wrap = false
	c.header.Frame = false

	return c.Render()
}

// Edit intercepts the key press events in the filer controller to update the file controller in real time.
func (c *FilterController) Edit(view *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	if !c.IsVisible() {
		return
	}

	cx, _ := view.Cursor()
	ox, _ := view.Origin()
	limit := ox+cx+1 > c.maxLength
	switch {
	case ch != 0 && mod == 0 && !limit:
		view.EditWrite(ch)
	case key == gocui.KeySpace && !limit:
		view.EditWrite(' ')
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		view.EditDelete(true)
	}

	// notify listeners
	c.notifyFilterEditListeners()
}

func (c *FilterController) notifyFilterEditListeners() {
	currentValue := strings.TrimSpace(c.view.Buffer())
	for _, listener := range c.filterEditListeners {
		err := listener(currentValue)
		if err != nil {
			// note: cannot propagate error from here since this is from the main gogui thread
			log.Errorf("notifyFilterEditListeners: %+c", err)
		}
	}
}

// Update refreshes the state objects for future rendering (currently does nothing).
func (c *FilterController) Update() error {
	return nil
}

// Render flushes the state objects to the screen. Currently this is the users path filter input.
func (c *FilterController) Render() error {
	log.Tracef("controller.Render() %s", c.Name())

	c.gui.Update(func(g *gocui.Gui) error {
		_, err := fmt.Fprintln(c.header, format.Header(c.labelStr))
		if err != nil {
			log.Error("unable to write to buffer: ", err)
		}
		return err
	})
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
	// delete all user input from the tree controller
	c.view.Clear()

	if c.isVisible = visible; c.isVisible {
		_, err := c.gui.SetCurrentView(c.name)
		if err != nil {
			log.Error("unable to toggle filter controller: ", err)
			return err
		}
		return nil
	}

	// reset the cursor for the next time it is visible
	// Note: there is a subtle gocui behavior here where this cannot be called when the controller
	// is newly visible. Is this a problem with 9ofm or gocui?
	return c.view.SetCursor(0, 0)}

// Keymap indicates all the possible actions a user can take while the current pane is selected.
func (c *FilterController) Keymap() string {
	return format.StatusControlNormal("▏Type to filter the file tree ")
}

// OnLayoutChange is called whenever the screen dimensions are changed
func (c *FilterController) OnLayoutChange() error {
	err := c.Update()
	if err != nil {
		return err
	}
	return c.Render()
}

func (c *FilterController) Layout(g *gocui.Gui, minX, minY, maxX, maxY int) error {
	log.Tracef("controller.Layout(minX: %d, minY: %d, maxX: %d, maxY: %d) %s", minX, minY, maxX, maxY, c.Name())

	label, labelErr := g.SetView(c.Name()+"label", minX, minY, len(c.labelStr), maxY)
	view, viewErr := g.SetView(c.Name(), minX+(len(c.labelStr)-1), minY, maxX, maxY)

	if utils.IsNewView(viewErr, labelErr) {
		err := c.Setup(view, label)
		if err != nil {
			log.Error("unable to setup status controller", err)
			return err
		}
	}
	return nil
}

func (c *FilterController) RequestedSize(available int) *int {
	return &c.requestedHeight
}
