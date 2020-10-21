package controller

import (
	"fmt"
	"github.com/mushkevych/9ofm/commander/format"
	"github.com/mushkevych/9ofm/utils"
	log "github.com/sirupsen/logrus"
	"strings"

	"github.com/jroimartin/gocui"
)

// StatusController holds the UI objects and data models for populating the bottom-most pane. Specifically the panel
// shows the user a set of possible actions to take in the window and currently selected pane.
type StatusController struct {
	name string
	gui  *gocui.Gui
	view *gocui.View

	selectedView    Keymapper
	requestedHeight int

	keymaps []KeymapDetail
}

// NewStatusController creates a new controller object attached the the global [gocui] screen object.
func NewStatusController(gui *gocui.Gui) (controller *StatusController) {
	controller = new(StatusController)

	// populate main fields
	controller.name = "status"
	controller.gui = gui
	controller.keymaps = make([]KeymapDetail, 0)
	controller.requestedHeight = 1

	return controller
}

func (c *StatusController) SetCurrentView(r Keymapper) {
	c.selectedView = r
}

func (c *StatusController) Name() string {
	return c.name
}

func (c *StatusController) AddKeymap(keys ...KeymapDetail) {
	c.keymaps = append(c.keymaps, keys...)
}

// Setup initializes the UI concerns within the context of a global [gocui] controller object.
func (c *StatusController) Setup(view *gocui.View) error {
	log.Tracef("controller.Setup() %s", c.Name())

	// set controller options
	c.view = view
	c.view.Frame = false

	return c.Render()
}

// Update refreshes the state objects for future rendering (currently does nothing).
func (c *StatusController) Update() error {
	return nil
}

// Render flushes the state objects to the screen.
func (c *StatusController) Render() error {
	log.Tracef("controller.Render() %s", c.Name())

	c.gui.Update(func(g *gocui.Gui) error {
		c.view.Clear()

		var selectedKeymap string
		if c.selectedView != nil {
			selectedKeymap = c.selectedView.Keymap()
		}

		_, err := fmt.Fprintln(c.view, c.Keymap()+selectedKeymap+format.StatusNormal("▏"+strings.Repeat(" ", 1000)))
		if err != nil {
			log.Debug("unable to write to buffer: ", err)
		}

		return err
	})
	return nil
}

// IsVisible indicates if the status controller pane is currently initialized.
func (c *StatusController) IsVisible() bool {
	return c != nil
}

// SetVisible hides or shows the File Panel (currently not used)
func (c *StatusController) SetVisible(visible bool) error {
	return nil
}

// Keymap indicates all the possible global actions a user can take when any pane is selected.
func (c *StatusController) Keymap() string {
	var keymaps string
	for _, keymap := range c.keymaps {
		keymaps += keymap.String()
	}
	return keymaps
}

// OnLayoutChange is called whenever the screen dimensions are changed
func (c *StatusController) OnLayoutChange() error {
	err := c.Update()
	if err != nil {
		return err
	}
	return c.Render()
}

func (c *StatusController) Layout(g *gocui.Gui, minX, minY, maxX, maxY int) error {
	log.Tracef("controller.Layout(minX: %d, minY: %d, maxX: %d, maxY: %d) %s", minX, minY, maxX, maxY, c.Name())

	view, viewErr := g.SetView(c.Name(), minX, minY, maxX, maxY)
	if utils.IsNewView(viewErr) {
		err := c.Setup(view)
		if err != nil {
			log.Error("unable to setup status controller", err)
			return err
		}
	}
	return nil
}

func (c *StatusController) RequestedSize(available int) *int {
	return &c.requestedHeight
}
