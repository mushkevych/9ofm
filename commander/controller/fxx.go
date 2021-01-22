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
	"path/filepath"
	"strings"
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
	buttonF6 := buttonFactory("F6: MOVE", func() { _ = controller.F6 })
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
		case tcell.KeyF2:
			err = controller.F2()
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

func copy(source, destination string) error {
	var err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		var relPath string = strings.Replace(path, source, "", 1)
		if relPath == "" {
			return nil
		}
		if info.IsDir() {
			return os.Mkdir(filepath.Join(destination, relPath), 0755)
		} else {
			var data, err1 = ioutil.ReadFile(filepath.Join(source, relPath))
			if err1 != nil {
				return err1
			}
			return ioutil.WriteFile(filepath.Join(destination, relPath), data, 0777)
		}
	})
	return err
}

// ignoreInput is used as a validator function
func ignoreInput(text string, ch rune) bool {
	return false
}

func (c *FxxController) F2() error {
	if c.sourceFilePanel == nil || c.targetFilePanel == nil {
		return nil
	}

	formId := "formRename"
	label := "Rename :"
	currentFileName := c.sourceFilePanel.GetSelectedFileNode().Name
	modalForm := tview.NewModal()
	modalForm.SetBorder(true)
	modalForm.SetTitle("Rename: " + currentFileName)
	modalForm.SetTitleAlign(tview.AlignCenter)
	modalForm.GetForm().AddInputField(label, currentFileName, 20, nil, nil)
	modalForm.AddButtons([]string{"OK", "Cancel"})
	modalForm.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		switch buttonLabel {
		case "OK":
			sourceFileNode := c.sourceFilePanel.GetSelectedFileNode()
			targetFileName := modalForm.GetForm().GetFormItemByLabel(label).(*tview.InputField).GetText()
			if !strings.HasPrefix(targetFileName, "/") {
				// assume the target directory as current one
				pwd := c.sourceFilePanel.ftv.ModelTree.GetPwd()
				targetFileName = pwd + string(os.PathSeparator) + targetFileName
			}

			err := os.Rename(sourceFileNode.AbsPath(), targetFileName)
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

func (c *FxxController) F5() error {
	if c.sourceFilePanel == nil || c.targetFilePanel == nil {
		return nil
	}

	formId := "formCopy"
	label := "Copy :"

	modalForm := tview.NewModal()
	modalForm.SetBorder(true)
	modalForm.SetTitle("Copy: " + c.sourceFilePanel.GetSelectedFileNode().Name)
	modalForm.SetTitleAlign(tview.AlignCenter)

	defaultTargetFolder := c.targetFilePanel.ftv.ModelTree.GetPwd()
	if !strings.HasSuffix(defaultTargetFolder, "/"){
		defaultTargetFolder += string(os.PathSeparator)
	}
	modalForm.GetForm().AddInputField(label, defaultTargetFolder, 20, ignoreInput, nil)

	modalForm.AddButtons([]string{"OK", "Cancel"})
	modalForm.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		switch buttonLabel {
		case "OK":
			sourceFileNode := c.sourceFilePanel.GetSelectedFileNode()
			targetFolder := c.targetFilePanel.ftv.ModelTree.GetPwd()
			if !strings.HasSuffix(targetFolder, "/") {
				targetFolder += string(os.PathSeparator)
			}

			targetFileFqfp := targetFolder + string(os.PathSeparator) + sourceFileNode.Name
			err := copy(sourceFileNode.AbsPath(), targetFileFqfp)
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

func (c *FxxController) F6() error {
	if c.sourceFilePanel == nil || c.targetFilePanel == nil {
		return nil
	}

	formId := "formMove"
	label := "Move :"

	modalForm := tview.NewModal()
	modalForm.SetBorder(true)
	modalForm.SetTitle("Move: " + c.sourceFilePanel.GetSelectedFileNode().AbsPath())
	modalForm.SetTitleAlign(tview.AlignCenter)

	sourceFileNode := c.sourceFilePanel.GetSelectedFileNode()
	defaultTargetFolder := c.targetFilePanel.ftv.ModelTree.GetPwd()
	defaultTargetFileFqfp := defaultTargetFolder + string(os.PathSeparator) + sourceFileNode.Name
	modalForm.GetForm().AddInputField(label, defaultTargetFileFqfp, 20, nil, nil)

	modalForm.AddButtons([]string{"OK", "Cancel"})
	modalForm.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		switch buttonLabel {
		case "OK":
			targetFileName := modalForm.GetForm().GetFormItemByLabel(label).(*tview.InputField).GetText()
			err := os.Rename(sourceFileNode.AbsPath(), targetFileName)
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
	modalForm.GetForm().AddInputField(label, sourceFileNode.AbsPath(), 20, ignoreInput, nil)
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
