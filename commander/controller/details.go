package controller

import (
	"fmt"
	"strings"

	"github.com/mushkevych/9ofm/commander/format"
	log "github.com/sirupsen/logrus"

	"github.com/jroimartin/gocui"
)

// TODO: use it
// DetailsController holds the UI objects and data models for populating the opposite pane with details of the chosen
// object: file, directory, etc
type DetailsController struct {
	name   string
	gui    *gocui.Gui
	view   *gocui.View
	header *gocui.View
}

// NewDetailsController creates a new controller object attached the the global [gocui] screen object.
func NewDetailsController(gui *gocui.Gui) (controller *DetailsController) {
	controller = new(DetailsController)

	// populate main fields
	controller.name = "details"
	controller.gui = gui
	return controller
}

func (c *DetailsController) Name() string {
	return c.name
}

// Setup initializes the UI concerns within the context of a global [gocui] controller object.
func (c *DetailsController) Setup(view *gocui.View, header *gocui.View) error {
	log.Tracef("controller.Setup() %s", c.Name())

	// set controller options
	c.view = view
	c.view.Editable = false
	c.view.Wrap = true
	c.view.Highlight = false
	c.view.Frame = false

	c.header = header
	c.header.Editable = false
	c.header.Wrap = false
	c.header.Frame = false

	var keymaps = []KeymapDetail{
		{
			KeyboardShortcut: "Down",
			OnAction:         c.CursorDown,
		},
		{
			KeyboardShortcut: "Up",
			OnAction:         c.CursorUp,
		},
	}

	err := RegisterKeymaps(c.gui, c.name, keymaps)
	if err != nil {
		return err
	}

	return c.Render()
}

// CursorDown moves the cursor down in the details pane (currently indicates nothing).
func (c *DetailsController) CursorDown() error {
	return CursorDown(c.gui, c.view)
}

// CursorUp moves the cursor up in the details pane (currently indicates nothing).
func (c *DetailsController) CursorUp() error {
	return CursorUp(c.gui, c.view)
}

// OnLayoutChange is called whenever the screen dimensions are changed
func (c *DetailsController) OnLayoutChange() error {
	err := c.Update()
	if err != nil {
		return err
	}
	return c.Render()
}

// Update refreshes the state objects for future rendering.
func (c *DetailsController) Update() error {
	return nil
}

// Render flushes the state objects to the screen. The details pane reports:
func (c *DetailsController) Render() error {
	log.Tracef("controller.Render() %s", c.Name())

	template := "%5s  %12s  %-s\n"
	inefficiencyReport := fmt.Sprintf(format.Header(template), "Count", "Total Space", "Path")

	c.gui.Update(func(g *gocui.Gui) error {
		if c.header == nil {
			// shershen hack
			return nil
		}

		// update header
		c.header.Clear()
		width, _ := c.view.Size()

		headerStr := format.FmtHeader("Header DetailsController", width, false)

		_, err := fmt.Fprintln(c.header, headerStr)
		if err != nil {
			return err
		}

		// update contents
		c.view.Clear()

		var lines = make([]string, 0)
		lines = append(lines, format.Header("Tags:   ")+"(none)")
		lines = append(lines, format.Header("Command:"))
		lines = append(lines, inefficiencyReport)

		_, err = fmt.Fprintln(c.view, strings.Join(lines, "\n"))
		if err != nil {
			log.Debug("unable to write to buffer: ", err)
		}
		return err
	})
	return nil
}

// IsVisible indicates if the details controller pane is currently initialized.
func (c *DetailsController) IsVisible() bool {
	return c != nil
}

// SetVisible hides or shows the Details Panel (currently not used)
func (c *DetailsController) SetVisible(visible bool) error {
	return nil
}

func (c *DetailsController) Keymap() string {
	return "TBD"
}
