package controller

import (
	"errors"
	"github.com/gdamore/tcell/v2"
	"github.com/mushkevych/9ofm/commander/model"
	"github.com/mushkevych/9ofm/commander/view"
	tview "gitlab.com/tslocum/cview"

	log "github.com/sirupsen/logrus"
	"regexp"
)

type ViewOptionChangeListener func() error

var diffTypeColor = map[model.DiffType]tcell.Color{
	model.Added:      tcell.ColorGreen,
	model.Removed:    tcell.ColorRed,
	model.Modified:   tcell.ColorYellow,
	model.Unmodified: tcell.ColorWhite,
}

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

	table := tview.NewTable()
	table.SetBorders(false)
	table.SetSeparator('│')
	table.SetTitle(name + "Table")
	table.SetFixed(1, 0)
	table.SetSelectable(true, false)

	// SetSelectedFunc function is handler for tcell.KeyEnter
	table.SetSelectedFunc(func(row int, column int) {
		fileNode, ok := table.GetCell(row, column).Reference.(*model.FileNode)
		if !ok {
			log.Errorf("unable to cast cell.Reference to model.FileNode")
		}

		FILE_NAME_COLUMN_INDEX := 3
		selectedFileName := ".."
		cellText := string(table.GetCell(row, FILE_NAME_COLUMN_INDEX).Text)
		if cellText == ".." {
			selectedFileName = fileNode.Name
			fileNode = fileNode.Parent
		}
		err = controller.navigateTo(fileNode)
		if err != nil {
			log.Errorf("error in table.SetSelectedFunc->navigateTo(%v)", fileNode)
		}

		newSelectedRow := 1 // row that should be selected after navigateTo function is complete
		for rowIdx := 1; rowIdx < table.GetRowCount(); rowIdx++ {
			cellText = string(table.GetCell(rowIdx, FILE_NAME_COLUMN_INDEX).Text)
			if cellText == selectedFileName {
				newSelectedRow = rowIdx
				break
			}
		}

		table.SetOffset(newSelectedRow, 0)
		table.Select(newSelectedRow, 0)
	})

	table.SetSelectionChangedFunc(func(row, column int) {
		if row, _ := table.GetSelection(); row == 0 {
			// select top-most row, instead of a header
			table.Select(1, 0)
		}
	})

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		var err error
		switch event.Key() {
		case tcell.KeyCtrlA:
			err = controller.toggleShowDiffType(model.Added)
		case tcell.KeyCtrlR:
			err = controller.toggleShowDiffType(model.Removed)
		case tcell.KeyCtrlO:
			err = controller.toggleShowDiffType(model.Modified)
		case tcell.KeyCtrlU:
			err = controller.toggleShowDiffType(model.Unmodified)
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

func (c *FilePanelController) GetPwd() string {
	return c.ftv.ModelTree.GetPwd()
}

// navigateTo will enter the directory
func (c *FilePanelController) navigateTo(fileNode *model.FileNode) error {
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
		tableCell := tview.NewTableCell(columnName)
		tableCell.SetTextColor(tcell.ColorYellow)
		tableCell.SetAlign(tview.AlignCenter)
		table.SetCell(0, idx, tableCell)
	}

	rows, fileNodes := c.ftv.ModelTree.StringArrayBetween(0, c.ftv.ModelTree.VisibleSize())

	for idxRow, row := range rows {
		for idxCol, col := range row {
			fileNode := fileNodes[idxRow]
			cellTextColor := diffTypeColor[fileNode.Data.DiffType]

			tableCell := tview.NewTableCell(col)
			tableCell.SetTextColor(cellTextColor)
			tableCell.SetAlign(tview.AlignLeft)
			tableCell.SetReference(fileNode)

			table.SetCell(idxRow+1, idxCol, tableCell)
		}
	}

	if row, _ := table.GetSelection(); row == 0 {
		// select top-most row, instead of a header
		table.Select(1, 0)
	}
	return nil
}

// IsVisible indicates if the file tree controller is currently initialized
func (c *FilePanelController) GetSelectedFileNode() *model.FileNode {
	table := c.graphicElement.(*tview.Table)
	row, column := table.GetSelection()
	fileNode := table.GetCell(row, column).Reference
	return fileNode.(*model.FileNode)
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
