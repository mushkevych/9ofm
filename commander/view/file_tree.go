package view

import (
	"fmt"
	"github.com/mushkevych/9ofm/commander/configuration"
	"strings"

	"github.com/mushkevych/9ofm/commander/model"
)

// FileTreeViewModel holds the UI objects and data models for populating the right pane. Specifically the pane that
// shows selected layer or aggregate file ASCII tree.
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

// IsVisible indicates if the file tree view pane is currently initialized
func (v *FileTreeView) IsVisible() bool {
	return v != nil
}

// ToggleShowDiffType will show/hide the selected DiffType in the FileTree pane.
func (v *FileTreeView) ToggleShowDiffType(diffType model.DiffType) {
	v.HiddenDiffTypes[diffType] = !v.HiddenDiffTypes[diffType]
}
