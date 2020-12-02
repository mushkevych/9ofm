package controller

import (
	"errors"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"regexp"
	"strings"

	"github.com/mushkevych/9ofm/commander/model"
	"github.com/mushkevych/9ofm/commander/view"
	log "github.com/sirupsen/logrus"
)

type ViewOptionChangeListener func() error

// FileTreeController holds the UI objects and data models for populating the File Tree Pane.
type FileTreeController struct {
	tviewApp       *tview.Application
	name           string
	graphicElement GraphicElement
	ftv            *view.FileTreeView

	filterRegex *regexp.Regexp
	listeners   []ViewOptionChangeListener
}

// NewFileTreeController creates a new FileTreeController object attached the the global [tview] screen object.
func NewFileTreeController(tviewApp *tview.Application, name string, fileTree *model.FileTreeModel) (controller *FileTreeController, err error) {
	controller = new(FileTreeController)
	controller.listeners = make([]ViewOptionChangeListener, 0)

	// populate main fields
	controller.name = name
	controller.ftv, err = view.NewFileTreeView(fileTree)
	if err != nil {
		return nil, err
	}

	controller.graphicElement = tview.NewList()
	controller.graphicElement.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		var err error
		switch event.Key() {
		case tcell.KeyEnter:
			err = controller.navigateTo()
		case tcell.KeyCtrlA:
			err = controller.toggleShowDiffType(model.Added)
		case tcell.KeyCtrlR:
			err = controller.toggleShowDiffType(model.Removed)
		case tcell.KeyCtrlO:
			err = controller.toggleShowDiffType(model.Modified)
		case tcell.KeyCtrlU:
			err = controller.toggleShowDiffType(model.Unmodified)
		case tcell.KeyCtrlB:
			err = controller.toggleAttributes()
		}

		if err != nil {
			log.WithError(err)
		}
		return event
	})
	return controller, err
}

func (c *FileTreeController) SetFilterRegex(filterRegex *regexp.Regexp) {
	c.filterRegex = filterRegex
}

func (c *FileTreeController) Name() string {
	return c.name
}

// navigateTo will enter the directory
func (c *FileTreeController) navigateTo() error {
	fileNode := c.ftv.GetNodeAtCursor()
	if fileNode.IsDir() || fileNode.AbsPath() == "/" {
		fqfp := fileNode.AbsPath()
		fileTree, err := model.ReadFileTree(fqfp)
		if err != nil {
			return err
		}

		c.ftv, err = view.NewFileTreeView(fileTree)
		if err != nil {
			return err
		}
	}

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

// toggleAttributes will show/hide file attributes
func (c *FileTreeController) toggleAttributes() error {
	err := c.ftv.ToggleAttributes()
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

// toggleShowDiffType will show/hide the selected DiffType in the model pane.
func (c *FileTreeController) toggleShowDiffType(diffType model.DiffType) error {
	c.ftv.ToggleShowDiffType(diffType)

	err := c.Render()
	if err != nil {
		return err
	}

	// we need to render the changes to the status pane as well (not just this contoller/controller)
	return c.notifyOnViewOptionChangeListeners()
}

// Render flushes the state objects (file tree) to the pane.
func (c *FileTreeController) Render() error {
	log.Tracef("controller.Render() %s", c.Name())

	list, ok := c.graphicElement.(*tview.List)
	if !ok {
		return errors.New("unable to cast graphicElement to tview.List")
	}

	list.Clear()
	var headerStr string
	if c.ftv.ShowFileAttributes {
		headerStr = fmt.Sprintf(model.AttributeFormat, "Permission", "UID:GID", "Size", "Filetree")
	} else {
		headerStr = fmt.Sprintf(model.AttributeFormat, "Permission", "UID:GID", "Size", "Filetree")
	}
	c.graphicElement.SetTitle(headerStr)

	treeString := c.ftv.ModelTree.String(c.ftv.ShowFileAttributes)
	lines := strings.Split(treeString, "\n")
	for _, line := range lines {
		list.AddItem(line, "", 'a', nil)
	}

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

// GetPrimitive returns graphicElement used by tview framework to render the UI interface
func (c *FileTreeController) GraphicElement() GraphicElement {
	return c.graphicElement
}

//func (c *FileTreeController) AddOnChangeListener(listener ...ViewOptionChangeListener) {
//	c.listeners = append(c.listeners, listener...)
//}
//
//// resetCursor moves the cursor back to the top of the buffer and translates to the top of the buffer.
//func (c *FileTreeController) resetCursor() {
//	c.ftv.ResetCursor()
//}
//
//// CursorDown moves the cursor down and renders the controller.
//func (c *FileTreeController) CursorDown() error {
//	if c.ftv.CursorDown() {
//		return c.Render()
//	}
//	return nil
//}
//
//// CursorUp moves the cursor up and renders the controller.
//func (c *FileTreeController) CursorUp() error {
//	if c.ftv.CursorUp() {
//		return c.Render()
//	}
//	return nil
//}
//
//// PageDown moves to next page putting the cursor on top
//func (c *FileTreeController) PageDown() error {
//	err := c.ftv.PageDown()
//	if err != nil {
//		return err
//	}
//	return c.Render()
//}
//
//// PageUp moves to previous page putting the cursor on top
//func (c *FileTreeController) PageUp() error {
//	err := c.ftv.PageUp()
//	if err != nil {
//		return err
//	}
//	return c.Render()
//}
