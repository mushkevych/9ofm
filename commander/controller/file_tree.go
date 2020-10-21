package controller

import (
	"fmt"
	"github.com/mushkevych/9ofm/commander/configuration"
	"regexp"

	"github.com/jroimartin/gocui"
	"github.com/mushkevych/9ofm/commander/format"
	"github.com/mushkevych/9ofm/commander/model"
	"github.com/mushkevych/9ofm/commander/view"
	"github.com/mushkevych/9ofm/utils"
	log "github.com/sirupsen/logrus"
)

type ViewOptionChangeListener func() error

// FileTreeController holds the UI objects and data models for populating the File Tree Pane.
type FileTreeController struct {
	name   string
	gui    *gocui.Gui
	view   *gocui.View
	header *gocui.View
	vm     *view.FileTreeView
	title  string

	filterRegex         *regexp.Regexp
	listeners           []ViewOptionChangeListener
	keymaps             []KeymapDetail
	requestedWidthRatio float64
}

// NewFileTreeController creates a new FileTreeController object attached the the global [gocui] screen object.
func NewFileTreeController(gui *gocui.Gui, name string, fileTree *model.FileTreeModel) (controller *FileTreeController, err error) {
	controller = new(FileTreeController)
	controller.listeners = make([]ViewOptionChangeListener, 0)

	// populate main fields
	controller.name = name
	controller.gui = gui
	controller.vm, err = view.NewFileTreeView(fileTree)
	if err != nil {
		return nil, err
	}

	requestedWidthRatio := configuration.Config.GetFloat64("filetree.pane-width")
	if requestedWidthRatio >= 1 || requestedWidthRatio <= 0 {
		log.Errorf("invalid config value: 'filetree.pane-width' should be 0 < value < 1, given '%v'", requestedWidthRatio)
		requestedWidthRatio = 0.5
	}
	controller.requestedWidthRatio = requestedWidthRatio

	return controller, err
}

func (c *FileTreeController) AddViewOptionChangeListener(listener ...ViewOptionChangeListener) {
	c.listeners = append(c.listeners, listener...)
}

func (c *FileTreeController) SetTitle(title string) {
	c.title = title
}

func (c *FileTreeController) SetFilterRegex(filterRegex *regexp.Regexp) {
	c.filterRegex = filterRegex
}

func (c *FileTreeController) Name() string {
	return c.name
}

// Setup initializes the UI concerns within the context of a global [gocui] view object.
func (c *FileTreeController) Setup(view *gocui.View, header *gocui.View) error {
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

	var keymaps = []KeymapDetail{
		{
			KeyboardShortcut: "Space",
			OnAction:         c.toggleCollapse,
			Display:          "Collapse dir",
		},
		{
			KeyboardShortcut: "Ctrl+Space",
			OnAction:         c.toggleCollapseAll,
			Display:          "Collapse all dir",
		},
		{
			KeyboardShortcut: "Ctrl+a",
			OnAction:   func() error { return c.toggleShowDiffType(model.Added) },
			IsSelected: func() bool { return !c.vm.HiddenDiffTypes[model.Added] },
			Display:    "Added",
		},
		{
			KeyboardShortcut: "Ctrl+r",
			OnAction:   func() error { return c.toggleShowDiffType(model.Removed) },
			IsSelected: func() bool { return !c.vm.HiddenDiffTypes[model.Removed] },
			Display:    "Removed",
		},
		{
			KeyboardShortcut: "Ctrl+m",
			OnAction:   func() error { return c.toggleShowDiffType(model.Modified) },
			IsSelected: func() bool { return !c.vm.HiddenDiffTypes[model.Modified] },
			Display:    "Modified",
		},
		{
			KeyboardShortcut: "Ctrl+u",
			OnAction:   func() error { return c.toggleShowDiffType(model.Unmodified) },
			IsSelected: func() bool { return !c.vm.HiddenDiffTypes[model.Unmodified] },
			Display:    "Unmodified",
		},
		{
			KeyboardShortcut: "Ctrl+b",
			OnAction:         c.toggleAttributes,
			IsSelected:       func() bool { return c.vm.ShowAttributes },
			Display:          "Attributes",
		},
		{
			KeyboardShortcut: "Pgup",
			OnAction:         c.PageUp,
		},
		{
			KeyboardShortcut: "Pgdn",
			OnAction:         c.PageDown,
		},
		{
			KeyboardShortcut: "Down",
			OnAction:         c.CursorDown,
		},
		{
			KeyboardShortcut: "Up",
			OnAction:         c.CursorUp,
		},
		{
			KeyboardShortcut: "Left",
			OnAction:         c.CursorLeft,
		},
		{
			KeyboardShortcut: "Right",
			OnAction:         c.CursorRight,
		},
	}

	err := RegisterKeymaps(c.gui, c.name, keymaps)
	if err != nil {
		return err
	}
	c.keymaps = keymaps

	_, height := c.view.Size()
	c.vm.Setup(0, height)
	_ = c.Update()
	_ = c.Render()

	return nil
}

// ResetCursor moves the cursor back to the top of the buffer and translates to the top of the buffer.
func (c *FileTreeController) resetCursor() {
	_ = c.view.SetCursor(0, 0)
	c.vm.ResetCursor()
}

// CursorDown moves the cursor down and renders the controller.
// Note: we cannot use the gocui buffer since any state change requires writing the entire tree to the buffer.
// Instead we are keeping an upper and lower bounds of the tree string to render and only flushing
// this range into the controller buffer. This is much faster when tree sizes are large.
func (c *FileTreeController) CursorDown() error {
	if c.vm.CursorDown() {
		return c.Render()
	}
	return nil
}

// CursorUp moves the cursor up and renders the controller.
// Note: we cannot use the gocui buffer since any state change requires writing the entire tree to the buffer.
// Instead we are keeping an upper and lower bounds of the tree string to render and only flushing
// this range into the controller buffer. This is much faster when tree sizes are large.
func (c *FileTreeController) CursorUp() error {
	if c.vm.CursorUp() {
		return c.Render()
	}
	return nil
}

// CursorLeft moves the cursor up until we reach the Parent Node or top of the tree
func (c *FileTreeController) CursorLeft() error {
	err := c.vm.CursorLeft(c.filterRegex)
	if err != nil {
		return err
	}
	_ = c.Update()
	return c.Render()
}

// CursorRight descends into directory expanding it if needed
func (c *FileTreeController) CursorRight() error {
	err := c.vm.CursorRight(c.filterRegex)
	if err != nil {
		return err
	}
	_ = c.Update()
	return c.Render()
}

// PageDown moves to next page putting the cursor on top
func (c *FileTreeController) PageDown() error {
	err := c.vm.PageDown()
	if err != nil {
		return err
	}
	return c.Render()
}

// PageUp moves to previous page putting the cursor on top
func (c *FileTreeController) PageUp() error {
	err := c.vm.PageUp()
	if err != nil {
		return err
	}
	return c.Render()
}

// getAbsPositionNode determines the selected screen cursor's location in the file tree, returning the selected FileNode.
// func (controller *FileTreeController) getAbsPositionNode() (node *model.FileNode) {
// 	return controller.vm.getAbsPositionNode(filterRegex())
// }

// ToggleCollapse will collapse/expand the selected FileNode.
func (c *FileTreeController) toggleCollapse() error {
	err := c.vm.ToggleCollapse(c.filterRegex)
	if err != nil {
		return err
	}
	_ = c.Update()
	return c.Render()
}

// ToggleCollapseAll will collapse/expand the all directories.
func (c *FileTreeController) toggleCollapseAll() error {
	err := c.vm.ToggleCollapseAll()
	if err != nil {
		return err
	}
	if c.vm.CollapseAll {
		c.resetCursor()
	}
	_ = c.Update()
	return c.Render()
}

func (c *FileTreeController) notifyOnViewOptionChangeListeners() error {
	for _, listener := range c.listeners {
		err := listener()
		if err != nil {
			log.Errorf("notifyOnViewOptionChangeListeners error: %+c", err)
			return err
		}
	}
	return nil
}

// ToggleAttributes will show/hide file attributes
func (c *FileTreeController) toggleAttributes() error {
	err := c.vm.ToggleAttributes()
	if err != nil {
		return err
	}

	err = c.Update()
	if err != nil {
		return err
	}
	err = c.Render()
	if err != nil {
		return err
	}

	// we need to render the changes to the status pane as well (not just this contoller/controller)
	return c.notifyOnViewOptionChangeListeners()
}

// ToggleShowDiffType will show/hide the selected DiffType in the model pane.
func (c *FileTreeController) toggleShowDiffType(diffType model.DiffType) error {
	c.vm.ToggleShowDiffType(diffType)

	err := c.Update()
	if err != nil {
		return err
	}
	err = c.Render()
	if err != nil {
		return err
	}

	// we need to render the changes to the status pane as well (not just this contoller/controller)
	return c.notifyOnViewOptionChangeListeners()
}

// Update refreshes the state objects for future rendering.
func (c *FileTreeController) Update() error {
	var width, height int

	if c.view != nil {
		width, height = c.view.Size()
	} else {
		// before the TUI is setup there may not be a controller to reference. Use the entire screen as reference.
		width, height = c.gui.Size()
	}
	// height should account for the header
	return c.vm.Update(c.filterRegex, width, height-1)
}

// Render flushes the state objects (file tree) to the pane.
func (c *FileTreeController) Render() error {
	log.Tracef("controller.Render() %s", c.Name())

	title := c.title
	isSelected := c.gui.CurrentView() == c.view

	c.gui.Update(func(g *gocui.Gui) error {
		if c.header == nil {
			// shershen hack
			return nil
		}

		// update the header
		c.header.Clear()
		width, _ := g.Size()
		headerStr := format.FmtHeader(title, width, isSelected)
		if c.vm.ShowAttributes {
			headerStr += fmt.Sprintf(model.AttributeFormat+" %s", "P", "ermission", "UID:GID", "Size", "Filetree")
		}
		_, _ = fmt.Fprintln(c.header, headerStr)

		// update the contents
		c.view.Clear()
		err := c.vm.Render()
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(c.view, c.vm.Buffer.String())

		return err
	})
	return nil
}

// IsVisible indicates if the file tree controller is currently initialized
func (c *FileTreeController) IsVisible() bool {
	return c != nil
}

// SetVisible hides or shows the File Panel (currently not used)
func (c *FileTreeController) SetVisible(visible bool) error {
	return nil
}

// Keymap indicates all the possible actions a user can take while the current pane is selected.
func (c *FileTreeController) Keymap() string {
	var keymaps string
	for _, keymap := range c.keymaps {
		keymaps += keymap.String()
	}
	return keymaps
}

// OnLayoutChange is called by the UI framework to inform the controller-model of the new screen dimensions
func (c *FileTreeController) OnLayoutChange() error {
	err := c.Update()
	if err != nil {
		return err
	}
	return c.Render()
}

func (c *FileTreeController) Layout(g *gocui.Gui, minX, minY, maxX, maxY int) error {
	log.Tracef("controller.Layout(minX: %d, minY: %d, maxX: %d, maxY: %d) %s", minX, minY, maxX, maxY, c.Name())
	attributeRowSize := 0

	// make the layout responsive to the available realestate. Make more room for the main content by hiding auxillary
	// content when there is not enough room
	if maxX-minX < 60 {
		c.vm.ConstrainLayout()
	} else {
		c.vm.ExpandLayout()
	}

	if c.vm.ShowAttributes {
		attributeRowSize = 1
	}

	// header + attribute header
	headerSize := 1 + attributeRowSize
	// note: maxY needs to account for the (invisible) border, thus a +1
	header, headerErr := g.SetView(c.Name()+"header", minX, minY, maxX, minY+headerSize+1)
	// we are going to overlap the controller over the (invisible) border (so minY will be one less than expected).
	// additionally, maxY will be bumped by one to include the border
	guiView, viewErr := g.SetView(c.Name(), minX, minY+headerSize, maxX, maxY+1)
	if utils.IsNewView(viewErr, headerErr) {
		err := c.Setup(guiView, header)
		if err != nil {
			log.Error("unable to setup tree controller", err)
			return err
		}
	}
	return nil
}

func (c *FileTreeController) RequestedSize(available int) *int {
	//var requestedWidth = int(float64(available) * (1.0 - c.requestedWidthRatio))
	//return &requestedWidth
	return nil
}
