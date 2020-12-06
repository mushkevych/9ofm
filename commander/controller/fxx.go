package controller

import (
	"github.com/mushkevych/9ofm/commander/model"
	"github.com/mushkevych/9ofm/commander/system"
	"github.com/mushkevych/9ofm/commander/view"
	"github.com/rivo/tview"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
)

// FxxController defines the bottom UI row with F1-F12 functional keys, and related properties and functions
type FxxController struct {
	tviewApp       *tview.Application
	name           string
	graphicElement GraphicElement

	sourceFilePanel *FilePanelController
	targetFilePanel *FilePanelController
}

// NewFxxController creates a new controller object attached the the global [gocui] screen object.
func NewFxxController(tviewApp *tview.Application) (controller *FxxController) {
	controller = new(FxxController)

	// populate main fields
	controller.tviewApp = tviewApp
	controller.name = "bottom_row"

	// create tview graphicElement
	buttonHeight := 3 // number of rows
	buttonWidth := 10 // number of chars

	buttonF2 := tview.NewButton("F2: RENAME").SetSelectedFunc(nil)
	buttonF2.SetBorder(false).SetRect(0, 0, buttonWidth, buttonHeight)

	buttonF3 := tview.NewButton("F3: VIEW").SetSelectedFunc(nil)
	buttonF3.SetBorder(false).SetRect(0, 0, buttonWidth, buttonHeight)

	buttonF4 := tview.NewButton("F4: EDIT").SetSelectedFunc(nil)
	buttonF4.SetBorder(false).SetRect(0, 0, buttonWidth, buttonHeight)

	buttonF5 := tview.NewButton("F5: COPY").SetSelectedFunc(func() { _ = controller.F5 })
	buttonF5.SetBorder(false).SetRect(0, 0, buttonWidth, buttonHeight)

	buttonF6 := tview.NewButton("F6: HELP").SetSelectedFunc(func() { _ = controller.F6 })
	buttonF6.SetBorder(false).SetRect(0, 0, buttonWidth, buttonHeight)

	buttonF7 := tview.NewButton("F7: MKDIR").SetSelectedFunc(func() { _ = controller.F7 })
	buttonF7.SetBorder(false).SetRect(0, 0, buttonWidth, buttonHeight)

	buttonF8 := tview.NewButton("F8: RM").SetSelectedFunc(func() { _ = controller.F8 })
	buttonF8.SetBorder(false).SetRect(0, 0, buttonWidth, buttonHeight)

	buttonF9 := tview.NewButton("F9: TERM").SetSelectedFunc(nil)
	buttonF9.SetBorder(false).SetRect(0, 0, buttonWidth, buttonHeight)

	buttonF10 := tview.NewButton("F10: EXIT").SetSelectedFunc(func() { _ = controller.F10 })
	buttonF10.SetBorder(false).SetRect(0, 0, buttonWidth, buttonHeight)

	controller.graphicElement = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(buttonF2, 0, 1, false).
		AddItem(buttonF3, 0, 1, false).
		AddItem(buttonF4, 0, 1, false).
		AddItem(buttonF5, 0, 1, false).
		AddItem(buttonF6, 0, 1, false).
		AddItem(buttonF7, 0, 1, false).
		AddItem(buttonF8, 0, 1, false).
		AddItem(buttonF9, 0, 1, false).
		AddItem(buttonF10, 0, 1, false)

	return controller
}

func (c *FxxController) Name() string {
	return c.name
}

// SetFilePanels sets active and inactive File Panels
func (c *FxxController) SetFilePanels(activeFilePanel, targetFilePanel *FilePanelController) {
	c.sourceFilePanel = activeFilePanel
	c.targetFilePanel = targetFilePanel
}

// Update refreshes the state objects for future rendering (currently does nothing).
func (c *FxxController) Update() error {
	return nil
}

// Render flushes the state objects to the screen.
func (c *FxxController) Render() error {
	log.Tracef("controller.Render() %s", c.Name())

	return nil
}

// IsVisible indicates if the status controller pane is currently initialized.
func (c *FxxController) IsVisible() bool {
	return c != nil
}

// SetVisible is not used for Functional Rune Row
func (c *FxxController) SetVisible(visible bool) error {
	return nil
}

// GetPrimitive returns graphicElement used by tview framework to render the UI interface
func (c *FxxController) GraphicElement() GraphicElement {
	return c.graphicElement
}

// *** F1-F12 Functions ***
// quit is the tview callback invoked when the user hits Ctrl+C
func (c *FxxController) exit() error {
	c.tviewApp.Stop()
	return nil
}

func (c *FxxController) dummy() error {
	return nil
}

func (c *FxxController) refreshFilePanel(fpc *FilePanelController) error {
	if fpc == nil {
		return nil
	}

	fileTree, err := model.ReadFileTree(fpc.ftv.ModelTree.GetPwd())
	if err != nil {
		return err
	}

	fpc.ftv, err = view.NewFileTreeView(fileTree)
	if err != nil {
		return err
	}

	return fpc.Render()
}

func (c *FxxController) F5() error {
	if c.sourceFilePanel == nil || c.targetFilePanel == nil {
		return nil
	}

	sourceFileNode := c.sourceFilePanel.GetSelectedFileNode()
	targetFolder := c.targetFilePanel.ftv.ModelTree.GetPwd()
	targetFileName := targetFolder + string(os.PathSeparator) + sourceFileNode.Name

	input, err := ioutil.ReadFile(sourceFileNode.AbsPath())
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(targetFileName, input, sourceFileNode.Data.FileInfo.Mode)
	if err != nil {
		return err
	}

	err = c.refreshFilePanel(c.targetFilePanel)
	if err != nil {
		return err
	}

	return nil
}

func (c *FxxController) F6() error {
	if c.sourceFilePanel == nil || c.targetFilePanel == nil {
		return nil
	}

	err := c.F5()
	if err != nil {
		return err
	}

	err = c.F8()
	if err != nil {
		return err
	}

	err = c.refreshFilePanel(c.sourceFilePanel)
	if err != nil {
		return err
	}

	err = c.refreshFilePanel(c.targetFilePanel)
	if err != nil {
		return err
	}

	return nil
}

func (c *FxxController) F7() error {
	if c.sourceFilePanel == nil || c.targetFilePanel == nil {
		return nil
	}

	label := "New directory name:"
	form := tview.NewForm()
	form.AddInputField(label, "", 20, nil, nil).
		AddButton("OK", func() {
			newFolderName := form.GetFormItemByLabel(label).(*tview.InputField).GetText()
			err := os.Mkdir(newFolderName, 0755)
			if err != nil {
				system.MessageBus.Error(err.Error())
			}
		}).
		AddButton("Cancel", func() {
			c.tviewApp.SetFocus(c.sourceFilePanel.GraphicElement())
			err := c.refreshFilePanel(c.sourceFilePanel)
			if err != nil {
				system.MessageBus.Error(err.Error())
			}
		},
		)
	form.SetBorder(true).SetTitle("Create Folder").SetTitleAlign(tview.AlignCenter)
	c.tviewApp.SetFocus(form)

	err := c.refreshFilePanel(c.sourceFilePanel)
	if err != nil {
		return err
	}

	return nil
}

func (c *FxxController) F8() error {
	if c.sourceFilePanel == nil || c.targetFilePanel == nil {
		return nil
	}

	// TODO: add panel popup
	sourceFileNode := c.sourceFilePanel.GetSelectedFileNode()
	err := os.Remove(sourceFileNode.AbsPath())
	if err != nil {
		return err
	}

	err = c.refreshFilePanel(c.sourceFilePanel)
	if err != nil {
		return err
	}
	return nil
}

func (c *FxxController) F10() error {
	modalWindow := tview.NewModal().
		SetText("Do you want to quit the application?").
		AddButtons([]string{"Quit", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Quit" {
				_ = c.exit()
			}
		})

	// Display and focus the dialog
	c.tviewApp.SetRoot(modalWindow, false)

	return nil
}
