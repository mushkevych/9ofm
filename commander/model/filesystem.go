package model

import (
	"os"
	"path/filepath"
)

// ReadFileTree reads directory specified by the fqfp (Fully Qualified File Path)
func ReadFileTree(fqfp string) (*FileTreeModel, error) {
	fileTree := NewFileTreeModel()
	fileTree.Root.Name = fqfp

	err := filepath.Walk(fqfp, func(path string, info os.FileInfo, err error) error {
		fileInfo := NewFileInfo(path, info, err)
		fileTree.AddPath(path, fileInfo)

		if fileInfo.IsDir() && path != fqfp {
			return filepath.SkipDir // skip walking the directory and its contents.
		}
		return nil
	})

	return fileTree, err
}
