package model

import (
	"testing"
)

func BlankFileChangeInfo(fqfp string) (f *FileInfo) {
	result := FileInfo{
		Fqfp: fqfp,
		Mode: 0, // regular file
		hash: 123,
	}
	return &result
}

func TestAssignDiffType(t *testing.T) {
	tree := NewFileTreeModel()
	node, _, err := tree.AddPath("/usr", *BlankFileChangeInfo("/usr"))
	if err != nil {
		t.Errorf("Expected no error from fetching fqfp. got: %v", err)
	}
	node.Data.DiffType = Modified
	if tree.Root.Children["usr"].Data.DiffType != Modified {
		t.Fail()
	}
}

func TestMergeDiffTypes(t *testing.T) {
	a := Unmodified
	b := Unmodified
	merged := a.merge(b)
	if merged != Unmodified {
		t.Errorf("Expected Unchaged (0) but got %v", merged)
	}
	a = Modified
	b = Unmodified
	merged = a.merge(b)
	if merged != Modified {
		t.Errorf("Expected Unchaged (0) but got %v", merged)
	}
}
