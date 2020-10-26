package view

import (
	"bytes"
	"fmt"
	"github.com/mushkevych/9ofm/commander/configuration"
	"regexp"
	"strings"

	"github.com/mushkevych/9ofm/commander/format"
	"github.com/mushkevych/9ofm/commander/model"

	"github.com/lunixbochs/vtclean"
	log "github.com/sirupsen/logrus"
)

// FileTreeViewModel holds the UI objects and data models for populating the right pane. Specifically the pane that
// shows selected layer or aggregate file ASCII tree.
type FileTreeView struct {
	ModelTree *model.FileTreeModel
	ViewTree  *model.FileTreeModel

	constrainedRealEstate bool

	ShowAttributes              bool
	unconstrainedShowAttributes bool
	HiddenDiffTypes             []bool
	TreeIndex                   int
	bufferIndex                 int
	bufferIndexLowerBound       int

	refHeight int
	refWidth  int

	Buffer bytes.Buffer
}

// Setup initializes the UI concerns within the context of a global [gocui] controller object.
func (v *FileTreeView) Setup(lowerBound, height int) {
	v.bufferIndexLowerBound = lowerBound
	v.refHeight = height
}

// height returns the current height and considers the header
func (v *FileTreeView) height() int {
	if v.ShowAttributes {
		return v.refHeight - 1
	}
	return v.refHeight
}

// bufferIndexUpperBound returns the current upper bounds for the view
func (v *FileTreeView) bufferIndexUpperBound() int {
	return v.bufferIndexLowerBound + v.height()
}

// IsVisible indicates if the file tree view pane is currently initialized
func (v *FileTreeView) IsVisible() bool {
	return v != nil
}

// ResetCursor moves the cursor back to the top of the buffer and translates to the top of the buffer.
func (v *FileTreeView) ResetCursor() {
	v.TreeIndex = 0
	v.bufferIndex = 0
	v.bufferIndexLowerBound = 0
}

// CursorUp performs the internal view's buffer adjustments on cursor up. Note: this is independent of the gocui buffer.
func (v *FileTreeView) CursorUp() bool {
	if v.TreeIndex <= 0 {
		return false
	}
	v.TreeIndex--
	if v.TreeIndex < v.bufferIndexLowerBound {
		v.bufferIndexLowerBound--
	}
	if v.bufferIndex > 0 {
		v.bufferIndex--
	}
	return true
}

// CursorDown performs the internal view's buffer adjustments on cursor down. Note: this is independent of the gocui buffer.
func (v *FileTreeView) CursorDown() bool {
	if v.TreeIndex >= v.ModelTree.VisibleSize() {
		return false
	}
	v.TreeIndex++
	if v.TreeIndex > v.bufferIndexUpperBound() {
		v.bufferIndexLowerBound++
	}
	v.bufferIndex++
	if v.bufferIndex > v.height() {
		v.bufferIndex = v.height()
	}
	return true
}

// CursorLeft moves the cursor up until we reach the Parent Node or top of the tree
func (v *FileTreeView) CursorLeft(filterRegex *regexp.Regexp) error {
	var visitor func(*model.FileNode) error
	var evaluator func(*model.FileNode) bool
	var dfsCounter, newIndex int
	oldIndex := v.TreeIndex
	currentNode := v.GetAbsPositionNode(filterRegex)

	if currentNode == nil {
		return nil
	}
	parentPath := currentNode.Parent.AbsPath()

	visitor = func(curNode *model.FileNode) error {
		if strings.Compare(parentPath, curNode.AbsPath()) == 0 {
			newIndex = dfsCounter
		}
		dfsCounter++
		return nil
	}

	evaluator = func(curNode *model.FileNode) bool {
		regexMatch := true
		if filterRegex != nil {
			match := filterRegex.Find([]byte(curNode.AbsPath()))
			regexMatch = match != nil
		}
		return !curNode.Data.Hidden && regexMatch
	}

	err := v.ModelTree.VisitDepthParentFirst(visitor, evaluator)
	if err != nil {
		log.Errorf("could not propagate tree on cursorLeft: %+v", err)
		return err
	}

	v.TreeIndex = newIndex
	moveIndex := oldIndex - newIndex
	if newIndex < v.bufferIndexLowerBound {
		v.bufferIndexLowerBound = v.TreeIndex
	}

	if v.bufferIndex > moveIndex {
		v.bufferIndex -= moveIndex
	} else {
		v.bufferIndex = 0
	}

	return nil
}

// CursorRight descends into directory expanding it if needed
func (v *FileTreeView) CursorRight(filterRegex *regexp.Regexp) error {
	node := v.GetAbsPositionNode(filterRegex)
	if node == nil {
		return nil
	}

	if !node.Data.FileInfo.IsDir() {
		return nil
	}

	if len(node.Children) == 0 {
		return nil
	}

	v.TreeIndex++
	if v.TreeIndex > v.bufferIndexUpperBound() {
		v.bufferIndexLowerBound++
	}

	v.bufferIndex++
	if v.bufferIndex > v.height() {
		v.bufferIndex = v.height()
	}

	return nil
}

// PageDown moves to next page putting the cursor on top
func (v *FileTreeView) PageDown() error {
	nextBufferIndexLowerBound := v.bufferIndexLowerBound + v.height()
	nextBufferIndexUpperBound := nextBufferIndexLowerBound + v.height()

	// todo: this work should be saved or passed to render...
	treeString := v.ViewTree.StringBetween(nextBufferIndexLowerBound, nextBufferIndexUpperBound, v.ShowAttributes)
	lines := strings.Split(treeString, "\n")

	newLines := len(lines) - 1
	if v.height() >= newLines {
		nextBufferIndexLowerBound = v.bufferIndexLowerBound + newLines
	}

	v.bufferIndexLowerBound = nextBufferIndexLowerBound

	if v.TreeIndex < nextBufferIndexLowerBound {
		v.bufferIndex = 0
		v.TreeIndex = nextBufferIndexLowerBound
	} else {
		v.bufferIndex -= newLines
	}

	return nil
}

// PageUp moves to previous page putting the cursor on top
func (v *FileTreeView) PageUp() error {
	nextBufferIndexLowerBound := v.bufferIndexLowerBound - v.height()
	nextBufferIndexUpperBound := nextBufferIndexLowerBound + v.height()

	// todo: this work should be saved or passed to render...
	treeString := v.ViewTree.StringBetween(nextBufferIndexLowerBound, nextBufferIndexUpperBound, v.ShowAttributes)
	lines := strings.Split(treeString, "\n")

	newLines := len(lines) - 2
	if v.height() >= newLines {
		nextBufferIndexLowerBound = v.bufferIndexLowerBound - newLines
	}

	v.bufferIndexLowerBound = nextBufferIndexLowerBound

	if v.TreeIndex > (nextBufferIndexUpperBound - 1) {
		v.bufferIndex = 0
		v.TreeIndex = nextBufferIndexLowerBound
	} else {
		v.bufferIndex += newLines
	}
	return nil
}

// GetAbsPositionNode determines the selected screen cursor's location in the file tree,
// returning the selected FileNode.
func (v *FileTreeView) GetAbsPositionNode(filterRegex *regexp.Regexp) (node *model.FileNode) {
	var visitor func(*model.FileNode) error
	var evaluator func(*model.FileNode) bool
	var dfsCounter int

	visitor = func(curNode *model.FileNode) error {
		if dfsCounter == v.TreeIndex {
			node = curNode
		}
		dfsCounter++
		return nil
	}

	evaluator = func(curNode *model.FileNode) bool {
		if curNode.Parent == nil{
			return true
		}
		regexMatch := true
		if filterRegex != nil {
			match := filterRegex.Find([]byte(curNode.AbsPath()))
			regexMatch = match != nil
		}
		return !curNode.Data.Hidden && regexMatch
	}

	err := v.ModelTree.VisitDepthParentFirst(visitor, evaluator)
	if err != nil {
		log.Errorf("unable to get node position: %+v", err)
	}

	return node
}

// ToggleCollapse will collapse/expand the selected FileNode.
func (v *FileTreeView) ToggleCollapse(filterRegex *regexp.Regexp) error {
	node := v.GetAbsPositionNode(filterRegex)
	if node != nil && node.Data.FileInfo.IsDir() {
		//node.Data.ViewInfo.Collapsed = !node.Data.ViewInfo.Collapsed
	}
	return nil
}

func (v *FileTreeView) ConstrainLayout() {
	if !v.constrainedRealEstate {
		log.Debugf("constraining model layout")
		v.constrainedRealEstate = true
		v.unconstrainedShowAttributes = v.ShowAttributes
		v.ShowAttributes = false
	}
}

func (v *FileTreeView) ExpandLayout() {
	if v.constrainedRealEstate {
		log.Debugf("expanding model layout")
		v.ShowAttributes = v.unconstrainedShowAttributes
		v.constrainedRealEstate = false
	}
}

// ToggleAttributes will show/hide file attributes
func (v *FileTreeView) ToggleAttributes() error {
	// ignore any attempt to show the attributes when the layout is constrained
	if v.constrainedRealEstate {
		return nil
	}
	v.ShowAttributes = !v.ShowAttributes
	return nil
}

// ToggleShowDiffType will show/hide the selected DiffType in the FileTree pane.
func (v *FileTreeView) ToggleShowDiffType(diffType model.DiffType) {
	v.HiddenDiffTypes[diffType] = !v.HiddenDiffTypes[diffType]
}

// Update refreshes the state objects for future rendering.
func (v *FileTreeView) Update(filterRegex *regexp.Regexp, width, height int) error {
	v.refWidth = width
	v.refHeight = height

	// keep the v selection in parity with the current DiffType selection
	err := v.ModelTree.VisitDepthChildFirst(func(node *model.FileNode) error {
		node.Data.Hidden = v.HiddenDiffTypes[node.Data.DiffType]
		visibleChild := false
		for _, child := range node.Children {
			if !child.Data.Hidden {
				visibleChild = true
				node.Data.Hidden = false
			}
		}
		// hide nodes that do not match the current file filter regex (also don't unhide nodes that are already hidden)
		if filterRegex != nil && !visibleChild && !node.Data.Hidden {
			match := filterRegex.FindString(node.AbsPath())
			node.Data.Hidden = len(match) == 0
		}
		return nil
	}, nil)

	if err != nil {
		log.Errorf("unable to propagate v model tree: %+v", err)
		return err
	}

	// make a new tree with only visible nodes
	v.ViewTree = v.ModelTree.Clone()
	err = v.ViewTree.VisitDepthParentFirst(func(node *model.FileNode) error {
		if node.Data.Hidden {
			err1 := v.ViewTree.RemovePath(node.AbsPath())
			if err1 != nil {
				return err1
			}
		}
		return nil
	}, nil)

	if err != nil {
		log.Errorf("unable to propagate v view tree: %+v", err)
		return err
	}

	return nil
}

// Render flushes the state objects (file tree) to the pane.
func (v *FileTreeView) Render() error {
	treeString := v.ViewTree.StringBetween(v.bufferIndexLowerBound, v.bufferIndexUpperBound(), v.ShowAttributes)
	lines := strings.Split(treeString, "\n")
	// lines := [...]string{
	// 	"[0mdrwxr-xr-x         0:0     1.2 MB [0m ├── [0mbin[0m",
	// 	"[0mdrwxr-xr-x         0:0     1.3 MB [0m ├── [0mbin[0m",
	// }

	// update the contents
	v.Buffer.Reset()
	for idx, line := range lines {
		if idx == v.bufferIndex {
			_, err := fmt.Fprintln(&v.Buffer, format.Selected(vtclean.Clean(line, false)))
			if err != nil {
				log.Debug("unable to write to buffer: ", err)
				return err
			}
		} else {
			_, err := fmt.Fprintln(&v.Buffer, line)
			if err != nil {
				log.Debug("unable to write to buffer: ", err)
				return err
			}
		}
	}
	return nil
}

// NewFileTreeView creates a new view object attached the the global [gocui] screen object.
func NewFileTreeView(tree *model.FileTreeModel) (treeViewModel *FileTreeView, err error) {
	treeViewModel = new(FileTreeView)

	// populate main fields
	treeViewModel.ShowAttributes = configuration.Config.GetBool("filetree.show-attributes")
	treeViewModel.unconstrainedShowAttributes = treeViewModel.ShowAttributes
	treeViewModel.ModelTree = tree
	treeViewModel.HiddenDiffTypes = make([]bool, 4)

	hiddenTypes := configuration.Config.GetStringSlice("diff.hide", ",")
	for _, hType := range hiddenTypes {
		switch t := strings.ToLower(hType); t {
		case "added":
			treeViewModel.HiddenDiffTypes[model.Added] = true
		case "removed":
			treeViewModel.HiddenDiffTypes[model.Removed] = true
		case "modified":
			treeViewModel.HiddenDiffTypes[model.Modified] = true
		case "unmodified":
			treeViewModel.HiddenDiffTypes[model.Unmodified] = true
		default:
			return nil, fmt.Errorf("unknown diff.hide value: %s", t)
		}
	}

	return treeViewModel, nil
}
