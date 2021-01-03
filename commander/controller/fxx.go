package controller

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mushkevych/9ofm/commander/model"
	"github.com/mushkevych/9ofm/commander/system"
	"github.com/mushkevych/9ofm/commander/view"
	log "github.com/sirupsen/logrus"
	tview "gitlab.com/tslocum/cview"
	"io/ioutil"
	"os"
)

// FxxController defines the bottom UI row with F1-F12 functional keys, and related properties and functions
type FxxController struct {
	tviewApp       *tview.Application
	pages          *tview.Pages
	name           string
	graphicElement GraphicElement

	sourceFilePanel *FilePanelController
	targetFilePanel *FilePanelController
}

// NewFxxController creates a new controller object attached the the global [tview] screen object.
func NewFxxController(tviewApp *tview.Application, pages *tview.Pages) (controller *FxxController) {
	controller = new(FxxController)

	// populate main fields
	controller.tviewApp = tviewApp
	controller.pages = pages
	controller.name = "bottom_row"

	// create tview graphicElement
	buttonHeight := 3 // number of rows
	buttonWidth := 10 // number of chars

	buttonFactory := func(label string, handler func()) *tview.Button {
		button := tview.NewButton(label)
		button.SetSelectedFunc(handler)
		button.SetBorder(false)
		button.SetRect(0, 0, buttonWidth, buttonHeight)
		return button
	}

	buttonF2 := buttonFactory("F2: RENAME", nil)
	buttonF3 := buttonFactory("F3: VIEW", nil)
	buttonF4 := buttonFactory("F4: EDIT", nil)
	buttonF5 := buttonFactory("F5: COPY", func() { _ = controller.F5 })
	buttonF6 := buttonFactory("F6: HELP", func() { _ = controller.F6 })
	buttonF7 := buttonFactory("F7: MKDIR", func() { _ = controller.F7 })
	buttonF8 := buttonFactory("F8: RM", func() { _ = controller.F8 })
	buttonF9 := buttonFactory("F9: TERM", nil)
	buttonF10 := buttonFactory("F10: EXIT", func() { _ = controller.F10 })

	flex := tview.NewFlex()
	flex.SetDirection(tview.FlexColumn)
	flex.AddItem(buttonF2, 0, 1, false)
	flex.AddItem(buttonF3, 0, 1, false)
	flex.AddItem(buttonF4, 0, 1, false)
	flex.AddItem(buttonF5, 0, 1, false)
	flex.AddItem(buttonF6, 0, 1, false)
	flex.AddItem(buttonF7, 0, 1, false)
	flex.AddItem(buttonF8, 0, 1, false)
	flex.AddItem(buttonF9, 0, 1, false)
	flex.AddItem(buttonF10, 0, 1, false)
	controller.graphicElement = flex

	controller.graphicElement.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		var err error
		switch event.Key() {
		case tcell.KeyF5:
			err = controller.F5()
		case tcell.KeyF6:
			err = controller.F6()
		case tcell.KeyF7:
			err = controller.F7()
		case tcell.KeyF8:
			err = controller.F8()
		case tcell.KeyF10:
			err = controller.F10()
		}

		if err != nil {
			log.WithError(err)
		}
		return event
	})

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

	formId := "formMkdir"
	label := "New directory name:"
	modalForm := tview.NewModal()
	modalForm.SetBorder(true)
	modalForm.SetTitle("Create Folder")
	modalForm.SetTitleAlign(tview.AlignCenter)
	modalForm.GetForm().AddInputField(label, "", 20, nil, nil)
	modalForm.AddButtons([]string{"OK", "Cancel"})
	modalForm.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		switch buttonLabel {
		case "OK":
			newFolderName := modalForm.GetForm().GetFormItemByLabel(label).(*tview.InputField).GetText()
			fqfp := c.sourceFilePanel.GetPwd() + string(os.PathSeparator) + newFolderName
			err := os.Mkdir(fqfp, 0755)
			if err != nil {
				system.MessageBus.Error(err.Error())
			}

			err = c.refreshFilePanel(c.sourceFilePanel)
			if err != nil {
				system.MessageBus.Error(err.Error())
			}

			c.hideModalForm(formId)
		case "Cancel":
			c.hideModalForm(formId)
		}
	})

	c.showModalForm(formId, modalForm)
	return nil
}

func (c *FxxController) F8() error {
	if c.sourceFilePanel == nil || c.targetFilePanel == nil {
		return nil
	}

	formId := "formRmdir"
	label := "Delete Folder"
	modalForm := tview.NewModal()
	modalForm.SetBorder(true)
	modalForm.SetTitle("Delete Folder")
	modalForm.SetTitleAlign(tview.AlignCenter)

	sourceFileNode := c.sourceFilePanel.GetSelectedFileNode()
	modalForm.GetForm().AddInputField(label, sourceFileNode.AbsPath(), 20, func(text string, ch rune) bool {return false}, nil)
	modalForm.AddButtons([]string{"OK", "Cancel"})
	modalForm.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		switch buttonLabel {
		case "OK":
			err := os.Remove(sourceFileNode.AbsPath())
			if err != nil {
				system.MessageBus.Error(err.Error())
			}

			err = c.refreshFilePanel(c.sourceFilePanel)
			if err != nil {
				system.MessageBus.Error(err.Error())
			}

			c.hideModalForm(formId)
		case "Cancel":
			c.hideModalForm(formId)
		}
	})

	c.showModalForm(formId, modalForm)
	return nil
}

func (c *FxxController) F10() error {
	modalWindow := tview.NewModal()
	modalWindow.SetText("Do you want to quit the application?")
	modalWindow.AddButtons([]string{"Quit", "Cancel"})
	modalWindow.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Quit" {
				_ = c.exit()
			}
		})

	// Display and focus the dialog
	c.tviewApp.SetRoot(modalWindow, false)

	return nil
}

func (c *FxxController) hideModalForm(formId string) {
	c.pages.HidePage(formId)
	c.pages.RemovePage(formId)
	c.tviewApp.SetFocus(c.sourceFilePanel.graphicElement)
}

func (c *FxxController) showModalForm(formId string, form tview.Primitive) {
	c.pages.AddPage(
		formId,
		form,
		false,
		true,
	)
}
