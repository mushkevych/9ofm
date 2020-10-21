package controller

import (
	"fmt"

	"github.com/jroimartin/gocui"
	"github.com/mushkevych/9ofm/commander/format"
	"github.com/mushkevych/9ofm/utils"
	log "github.com/sirupsen/logrus"
)

// DebugController debug pane controller
type DebugController struct {
	name   string
	gui    *gocui.Gui
	view   *gocui.View
	header *gocui.View

	selectedView Keymapper
}

// NewDebugController creates a new controller object attached the the global [gocui] screen object.
func NewDebugController(gui *gocui.Gui) (controller *DebugController) {
	controller = new(DebugController)

	// populate main fields
	controller.name = "debug"
	controller.gui = gui

	return controller
}

func (c *DebugController) SetCurrentView(r Keymapper) {
	c.selectedView = r
}

func (c *DebugController) Name() string {
	return c.name
}

// Setup initializes the UI concerns within the context of a global [gocui] controller object.
func (c *DebugController) Setup(view *gocui.View, header *gocui.View) error {
	log.Tracef("controller.Setup() %s", c.Name())

	// set controller options
	c.view = view
	c.view.Editable = false
	c.view.Wrap = false
	c.view.Frame = false

	c.header = header
	c.header.Editable = false
	c.header.Wrap = false
	c.header.Frame = false

	return c.Render()
}

// OnLayoutChange is called whenever the screen dimensions are changed
func (c *DebugController) OnLayoutChange() error {
	err := c.Update()
	if err != nil {
		return err
	}
	return c.Render()
}

// Update refreshes the state objects for future rendering (currently does nothing).
func (c *DebugController) Update() error {
	return nil
}

// Render flushes the state objects to the screen.
func (c *DebugController) Render() error {
	log.Tracef("controller.Render() %s", c.Name())

	c.gui.Update(func(g *gocui.Gui) error {
		if c.header == nil {
			// shershen hack
			return nil
		}

		// update header...
		c.header.Clear()
		width, _ := g.Size()
		headerStr := format.FmtHeader("DebugController", width, false)
		_, _ = fmt.Fprintln(c.header, headerStr)

		// update controller...
		c.view.Clear()
		_, err := fmt.Fprintln(c.view, "blerg")
		if err != nil {
			log.Debug("unable to write to buffer: ", err)
		}

		return nil
	})
	return nil
}

// IsVisible indicates if the debug controller pane is currently initialized.
func (c *DebugController) IsVisible() bool {
	return c != nil
}

// SetVisible hides or shows the Debug Panel (currently not used)
func (c *DebugController) SetVisible(visible bool) error {
	return nil
}

func (c *DebugController) Layout(g *gocui.Gui, minX, minY, maxX, maxY int) error {
	log.Tracef("controller.Layout(minX: %d, minY: %d, maxX: %d, maxY: %d) %s", minX, minY, maxX, maxY, c.Name())

	// header
	headerSize := 1
	// note: maxY needs to account for the (invisible) border, thus a +1
	header, headerErr := g.SetView(c.Name()+"header", minX, minY, maxX, minY+headerSize+1)
	// we are going to overlap the controller over the (invisible) border (so minY will be one less than expected).
	// additionally, maxY will be bumped by one to include the border
	view, viewErr := g.SetView(c.Name(), minX, minY+headerSize, maxX, maxY+1)
	if utils.IsNewView(viewErr, headerErr) {
		err := c.Setup(view, header)
		if err != nil {
			log.Error("unable to setup debug controller", err)
			return err
		}
	}
	return nil
}

func (c *DebugController) RequestedSize(available int) *int {
	return nil
}
