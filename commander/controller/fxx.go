package controller

import (
	"fmt"
	"github.com/mushkevych/9ofm/commander/format"
	"github.com/mushkevych/9ofm/utils"
	log "github.com/sirupsen/logrus"
	"strings"

	"github.com/jroimartin/gocui"
)

// FxxController defines the bottom UI row with F1-F12 functional keys, and related properties and functions
type FxxController struct {
	name string
	gui  *gocui.Gui
	view *gocui.View

	requestedHeight int

	keymaps []KeymapDetail
}

// NewFxxController creates a new controller object attached the the global [gocui] screen object.
func NewFxxController(gui *gocui.Gui) (controller *FxxController) {
	controller = new(FxxController)

	// populate main fields
	controller.name = "bottom_row"
	controller.gui = gui
	controller.keymaps = make([]KeymapDetail, 0)
	controller.requestedHeight = 1

	return controller
}

func (c *FxxController) Name() string {
	return c.name
}

func (c *FxxController) AddKeymap(keys ...KeymapDetail) {
	c.keymaps = append(c.keymaps, keys...)
}

// Setup initializes the UI concerns within the context of a global [gocui] controller object.
func (c *FxxController) Setup(view *gocui.View) error {
	log.Tracef("controller.Setup() %s", c.Name())

	// set controller options
	c.view = view
	c.view.Editable = false
	c.view.Wrap = false
	c.view.Frame = true

	var keymaps = []KeymapDetail{
		{
			KeyboardShortcut: "F2",
			OnAction:         c.dummy,
			Display:          "Rename",
		},
		{
			KeyboardShortcut: "F3",
			OnAction:         c.dummy,
			Display:          "View",
		},
		{
			KeyboardShortcut: "F4",
			OnAction:         c.dummy,
			Display:          "View",
		},
		{
			KeyboardShortcut: "F5",
			OnAction:         c.dummy,
			Display:          "Copy",
		},
		{
			KeyboardShortcut: "F6",
			OnAction:         c.dummy,
			Display:          "Move",
		},
		{
			KeyboardShortcut: "F7",
			OnAction:         c.dummy,
			Display:          "MkDir",
		},
		{
			KeyboardShortcut: "F8",
			OnAction:         c.dummy,
			Display:          "Delete",
		},
		{
			KeyboardShortcut: "F9",
			OnAction:         c.dummy,
			Display:          "Term",
		},
		{
			KeyboardShortcut: "Ctrl+q",
			OnAction:         c.exit,
		},
		{
			KeyboardShortcut: "F10",
			OnAction:         c.exit,
			Display:          "Exit",
		},
	}

	// NOTE: viewname="" makes the keymapping global
	err := RegisterKeymaps(c.gui, "", keymaps)
	if err != nil {
		return err
	}
	c.keymaps = keymaps
	return c.Render()
}

// OnLayoutChange is called whenever the screen dimensions are changed
func (c *FxxController) OnLayoutChange() error {
	err := c.Update()
	if err != nil {
		return err
	}
	return c.Render()
}

// Update refreshes the state objects for future rendering (currently does nothing).
func (c *FxxController) Update() error {
	return nil
}

// Render flushes the state objects to the screen.
func (c *FxxController) Render() error {
	log.Tracef("controller.Render() %s", c.Name())

	c.gui.Update(func(g *gocui.Gui) error {
		c.view.Clear()

		_, err := fmt.Fprintln(c.view, c.Keymap()+format.StatusNormal("▏"+strings.Repeat(" ", 1000)))
		if err != nil {
			log.Debug("unable to write to buffer: ", err)
		}

		return err
	})
	return nil
}

// IsVisible indicates if the status controller pane is currently initialized.
func (c *FxxController) IsVisible() bool {
	return c != nil
}

// SetVisible is not used for Functional Key Row
func (c *FxxController) SetVisible(visible bool) error {
	return nil
}

// Keymap indicates all the possible global keyboard actions a user can take when any pane is selected.
func (c *FxxController) Keymap() string {
	var keymaps string
	for _, keymap := range c.keymaps {
		keymaps += keymap.String()
	}
	return keymaps
}

func (c *FxxController) Layout(g *gocui.Gui, minX, minY, maxX, maxY int) error {
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

func (c *FxxController) RequestedSize(available int) *int {
	return &c.requestedHeight
}

// *** F1-F12 Functions ***
// quit is the gocui callback invoked when the user hits Ctrl+C
func (c *FxxController) exit() error {
	return gocui.ErrQuit
}

func (c *FxxController) dummy() error {
	return nil
}
