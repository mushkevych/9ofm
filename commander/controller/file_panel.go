package controller

import (
	"errors"
	"github.com/gdamore/tcell/v2"
	"github.com/mushkevych/9ofm/commander/model"
	"github.com/mushkevych/9ofm/commander/view"
	"github.com/rivo/tview"
	log "github.com/sirupsen/logrus"
	"regexp"
)

type ViewOptionChangeListener func() error

// FilePanelController holds the UI objects and data models for populating the File Tree Panel.
type FilePanelController struct {
	tviewApp       *tview.Application
	name           string
	graphicElement GraphicElement
	ftv            *view.FileTreeView

	filterRegex *regexp.Regexp
	listeners   []ViewOptionChangeListener
}

// NewFilePanelController creates a new FilePanelController object attached the the global [tview] screen object.
func NewFilePanelController(tviewApp *tview.Application, name string, fileTree *model.FileTreeModel) (controller *FilePanelController, err error) {
	controller = new(FilePanelController)
	controller.listeners = make([]ViewOptionChangeListener, 0)

	controller.tviewApp = tviewApp
	controller.name = name
	controller.ftv, err = view.NewFileTreeView(fileTree)
	if err != nil {
		return nil, err
	}

	table := tview.NewTable().SetBorders(false).SetSeparator('│')
	table.SetTitle(name+"Table")
	table.SetFixed(1, 0).SetSelectable(true, false).
		SetSelectedFunc(func(row int, column int) {
			fileNode := table.GetCell(row, column).Reference
			err = controller.navigateTo(fileNode.(*model.FileNode))
		})

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		var err error
		switch event.Key() {
		//case tcell.KeyEnter:
		//	err = controller.navigateTo()
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

	controller.graphicElement = table
	return controller, err
}

func (c *FilePanelController) SetFilterRegex(filterRegex *regexp.Regexp) {
	c.filterRegex = filterRegex
}

func (c *FilePanelController) Name() string {
	return c.name
}

// navigateTo will enter the directory
func (c *FilePanelController) navigateTo(fileNode *model.FileNode) error {
	if fileNode == nil {
		// old pre-tview logic
		fileNode = c.ftv.GetNodeAtCursor()
	}

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

func (c *FilePanelController) notifyOnViewOptionChangeListeners() error {
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
func (c *FilePanelController) toggleAttributes() error {
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
func (c *FilePanelController) toggleShowDiffType(diffType model.DiffType) error {
	c.ftv.ToggleShowDiffType(diffType)

	err := c.Render()
	if err != nil {
		return err
	}

	// we need to render the changes to the status pane as well (not just this contoller/controller)
	return c.notifyOnViewOptionChangeListeners()
}

// Render flushes the state objects (file tree) to the pane.
func (c *FilePanelController) Render() error {
	log.Tracef("controller.Render() %s", c.Name())

	table, ok := c.graphicElement.(*tview.Table)
	if !ok {
		return errors.New("unable to cast graphicElement to tview.Table")
	}

	table.Clear()
	headerColumns := []string{"Permission", "UID:GID", "Size", "Name"}
	for idx, columnName := range headerColumns {
		table.SetCell(0, idx,
			tview.NewTableCell(columnName).
				SetTextColor(tcell.ColorYellow).
				SetAlign(tview.AlignCenter),
		)
	}

	rows, fileNodes := c.ftv.ModelTree.StringArrayBetween(0, c.ftv.ModelTree.VisibleSize())

	for idxRow, row := range rows {
		for idxCol, col := range row {
			table.SetCell(idxRow+1, idxCol,
				tview.NewTableCell(col).
					SetTextColor(tcell.ColorWhite).
					SetAlign(tview.AlignLeft).
					SetReference(fileNodes[idxRow]))
		}
	}
	return nil
}

// IsVisible indicates if the file tree controller is currently initialized
func (c *FilePanelController) IsVisible() bool {
	return c != nil
}

// SetVisible hides or shows the File Panel (currently not used)
func (c *FilePanelController) SetVisible(visible bool) error {
	return nil
}

// GetPrimitive returns graphicElement used by tview framework to render the UI interface
func (c *FilePanelController) GraphicElement() GraphicElement {
	return c.graphicElement
}

//func (c *FilePanelController) AddOnChangeListener(listener ...ViewOptionChangeListener) {
//	c.listeners = append(c.listeners, listener...)
//}
//
//// resetCursor moves the cursor back to the top of the buffer and translates to the top of the buffer.
//func (c *FilePanelController) resetCursor() {
//	c.ftv.ResetCursor()
//}
//
//// CursorDown moves the cursor down and renders the controller.
//func (c *FilePanelController) CursorDown() error {
//	if c.ftv.CursorDown() {
//		return c.Render()
//	}
//	return nil
//}
//
//// CursorUp moves the cursor up and renders the controller.
//func (c *FilePanelController) CursorUp() error {
//	if c.ftv.CursorUp() {
//		return c.Render()
//	}
//	return nil
//}
//
//// PageDown moves to next page putting the cursor on top
//func (c *FilePanelController) PageDown() error {
//	err := c.ftv.PageDown()
//	if err != nil {
//		return err
//	}
//	return c.Render()
//}
//
//// PageUp moves to previous page putting the cursor on top
//func (c *FilePanelController) PageUp() error {
//	err := c.ftv.PageUp()
//	if err != nil {
//		return err
//	}
//	return c.Render()
//}
