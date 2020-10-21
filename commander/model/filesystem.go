package model

import (
	"os"
	"path/filepath"
)

func ReadFileTree(startPath string) (*FileTreeModel, error) {
	fileTree := NewFileTreeModel()
	fileTree.Root.Name = "/"

	err := filepath.Walk(startPath, func(path string, info os.FileInfo, err error) error {
		fileInfo := NewFileInfo(path, info, err)
		fileTree.AddPath(path, fileInfo)

		if fileInfo.IsDir() && path != startPath {
			return filepath.SkipDir // skip walking the directory and its contents.
		}
		return nil
	})

	return fileTree, err
}


