package model

import (
	"testing"
)

func checkError(t *testing.T, err error, message string) {
	if err != nil {
		t.Errorf(message+": %+v", err)
	}
}

func TestAddChild(t *testing.T) {
	var expected, actual int
	tree := NewFileTreeModel()

	one := tree.Root.AddChild("one.txt", FileInfo{
		Fqfp: "/one.txt",
	})

	two := tree.Root.AddChild("folder", FileInfo{})

	tree.Root.AddChild("three.txt!", FileInfo{})
	two.AddChild("sub-one", FileInfo{})
	two.AddChild("sub-two", FileInfo{})
	two.AddChild("sub-three", FileInfo{})

	expected, actual = 6, tree.Size
	if expected != actual {
		t.Errorf("Expected a tree size of %d vs %d.", expected, actual)
	}

	expected, actual = 3, len(two.Children)
	if expected != actual {
		t.Errorf("Expected number of children %d vs %d.", expected, actual)
	}

	expected, actual = 3, len(tree.Root.Children)
	if expected != actual {
		t.Errorf("Expected Root's number of children to be %d got %d.", expected, actual)
	}

	expectedFqfp, actualFqfp := "/one.txt", one.Data.FileInfo.Fqfp
	if expected != actual {
		t.Errorf("Expected fully qualified filepath to be %+v got %+v.", expectedFqfp, actualFqfp)
	}

}

func TestRemoveChild(t *testing.T) {
	var expected, actual int

	tree := NewFileTreeModel()
	tree.Root.AddChild("first", FileInfo{})
	two := tree.Root.AddChild("nil", FileInfo{})
	tree.Root.AddChild("third", FileInfo{})
	forth := two.AddChild("forth", FileInfo{})
	two.AddChild("fifth", FileInfo{})

	err := forth.Remove()
	checkError(t, err, "unable to setup test")

	expected, actual = 4, tree.Size
	if expected != actual {
		t.Errorf("Expected a tree size of %d got %d.", expected, actual)
	}

	if tree.Root.Children["forth"] != nil {
		t.Errorf("Expected 'forth' node to be deleted.")
	}

	err = two.Remove()
	checkError(t, err, "unable to setup test")

	expected, actual = 2, tree.Size
	if expected != actual {
		t.Errorf("Expected a tree size of %d got %d.", expected, actual)
	}

	if tree.Root.Children["nil"] != nil {
		t.Errorf("Expected 'nil' node to be deleted.")
	}

}

func TestPath(t *testing.T) {
	expectedFqfp := "/etc/systemd/system.conf"
	tree := NewFileTreeModel()
	node, _, _ := tree.AddPath(expectedFqfp, FileInfo{Fqfp: expectedFqfp})

	actualFqfp := node.AbsPath()
	if expectedFqfp != actualFqfp {
		t.Errorf("Expected fqfp '%s' got '%s'", expectedFqfp, actualFqfp)
	}

	if tree.Root.Parent != nil {
		t.Errorf("Expected Root's parent to be NIL")
	}

	if tree.Root.Name != tree.Root.AbsPath() || tree.Root.AbsPath() != "/" {
		t.Errorf("Expected Root's Fqfp to be /, got '%s'", tree.Root.AbsPath())
	}

	expected, actual := 3, tree.Size
	if expected != actual {
		t.Errorf("Expected a tree size of %d got %d.", expected, actual)
	}
}

func TestDiffTypeFromAddedChildren(t *testing.T) {
	tree := NewFileTreeModel()
	node, _, _ := tree.AddPath("/usr", *BlankFileChangeInfo("/usr"))
	node.Data.DiffType = Unmodified

	node, _, _ = tree.AddPath("/usr/bin", *BlankFileChangeInfo("/usr/bin"))
	node.Data.DiffType = Added

	node, _, _ = tree.AddPath("/usr/bin2", *BlankFileChangeInfo("/usr/bin2"))
	node.Data.DiffType = Removed

	err := tree.Root.Children["usr"].deriveDiffType(Unmodified)
	checkError(t, err, "unable to setup test")

	if tree.Root.Children["usr"].Data.DiffType != Modified {
		t.Errorf("Expected Modified but got %v", tree.Root.Children["usr"].Data.DiffType)
	}
}
func TestDiffTypeFromRemovedChildren(t *testing.T) {
	tree := NewFileTreeModel()
	_, _, _ = tree.AddPath("/usr", *BlankFileChangeInfo("/usr"))

	info1 := BlankFileChangeInfo("/usr/bin/awk")
	node, _, _ := tree.AddPath("/usr/bin/awk", *info1)
	node.Data.DiffType = Removed

	info2 := BlankFileChangeInfo("/usr/bin/gawk")
	node, _, _ = tree.AddPath("/usr/bin/gawk", *info2)
	node.Data.DiffType = Removed

	err := tree.Root.Children["usr"].Children["bin"].deriveDiffType(Unmodified)
	checkError(t, err, "unable to setup test")

	if tree.Root.Children["usr"].Children["bin"].Data.DiffType != Modified {
		t.Errorf("Expected Modified but got %v", tree.Root.Children["usr"].Data.DiffType)
	}
}

func TestDirSize(t *testing.T) {
	tree1 := NewFileTreeModel()

	_, _, err := tree1.AddPath("/usr/bin/awk", FileInfo{Size: 100})
	checkError(t, err, "unable to setup test")
	_, _, err = tree1.AddPath("/usr/bin/gawk", FileInfo{Size: 200})
	checkError(t, err, "unable to setup test")
	_, _, err = tree1.AddPath("/usr/bin/basename", FileInfo{Size: 300})
	checkError(t, err, "unable to setup test")

	node, _ := tree1.GetNode("/usr/bin")
	expected, actual := "0 1 2", node.MetadataString()
	if expected != actual {
		t.Errorf("Expected metadata '%s' got '%s'", expected, actual)
	}
}
