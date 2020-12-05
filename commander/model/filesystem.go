package model

import (
	"os"
	"path/filepath"
)

// ReadFileTree reads directory specified by the fqfp (Fully Qualified File AbsPath)
func ReadFileTree(fqfp string) (*FileTreeModel, error) {
	fileTree := NewFileTreeModel()

	err := filepath.Walk(fqfp, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		fileInfo := NewFileInfo(path, info, err)
		_, _, err = fileTree.AddPath(path, fileInfo)

		if fileInfo.IsDir() && path != fqfp {
			return filepath.SkipDir // skip walking nested directories and its content.
		}
		return nil
	})

	// Add parent directory reference ".."
	if err := fileTree.SetPwd(fqfp); err != nil {
		return nil, err
	}

	return fileTree, err
}
