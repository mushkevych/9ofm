package view

import (
	"fmt"
	"github.com/mushkevych/9ofm/commander/system"
	"strings"

	"github.com/mushkevych/9ofm/commander/model"
)

// FileTreeViewModel is a wrapper around FileTreeModel
// TODO: either remove in favour of FileTreeModel, or move more UI code into FileTreeView
type FileTreeView struct {
	// Tree model that holds representation of the filesystem for given path
	ModelTree *model.FileTreeModel

	HiddenDiffTypes []bool
}

// NewFileTreeView creates a new view object attached the the global [gocui] screen object.
func NewFileTreeView(tree *model.FileTreeModel) (treeViewModel *FileTreeView, err error) {
	treeViewModel = new(FileTreeView)

	// populate main fields
	treeViewModel.ModelTree = tree
	treeViewModel.HiddenDiffTypes = make([]bool, 4)

	hiddenTypes := system.Config.GetStringSlice("diff.hide", ",")
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

// ToggleShowDiffType will show/hide the selected DiffType in the FileTree pane.
func (v *FileTreeView) ToggleShowDiffType(diffType model.DiffType) {
	v.HiddenDiffTypes[diffType] = !v.HiddenDiffTypes[diffType]
}
