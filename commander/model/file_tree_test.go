package model

import (
	"fmt"
	"strings"
	"testing"
)

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func AssertDiffType(node *FileNode, expectedDiffType DiffType) error {
	if node.Data.DiffType != expectedDiffType {
		return fmt.Errorf("expecting node at %s to have DiffType %v, but had %v", node.AbsPath(), expectedDiffType, node.Data.DiffType)
	}
	return nil
}

func TestTreePwdString(t *testing.T) {
	tree := NewFileTreeModel()
	_ = tree.Root.AddChild("etc", FileInfo{})
	usr := tree.Root.AddChild("usr", FileInfo{})
	usrBin := usr.AddChild("bin", FileInfo{})
	usrBin.AddChild("gawk", FileInfo{})
	usrBin.AddChild("awk", FileInfo{})

	varDir := tree.Root.AddChild("var", FileInfo{})
	varLog := varDir.AddChild("log", FileInfo{})
	_ = varDir.AddChild("lib", FileInfo{})
	varLog.AddChild("syslog", FileInfo{})
	varLog.AddChild("kern.log", FileInfo{})

	// format: <pwd: expectedOutput>
	fixture := map[string]string{
		"/":         "etc\nusr\nvar\n",
		"/etc":      "..\n",
		"/usr":      "..\nbin\n",
		"/usr/bin":  "..\nawk\ngawk\n",
		"/var/":     "..\nlib\nlog\n",
		"/var/log/": "..\nkern.log\nsyslog\n",
	}

	for pwd, expectedOutput := range fixture {
		tree.SetPwd(pwd)
		actual := tree.String(false)

		if expectedOutput != actual {
			t.Errorf("Expected tree representation for %s:\n--->%s<---\nGot:\n--->%s<---",
				pwd, expectedOutput, actual)
		}
	}
}

func TestStringBetween(t *testing.T) {
	tree := NewFileTreeModel()

	fixturePaths := []string{
		"/var/lib/sudo",
		"/var/lib/systemd",
		"/var/lib/snmp",
		"/var/lib/grub",
		"/var/lib/fprint",
		"/var/lib/apt",
		"/var/lib/alsa",
	}

	for _, element := range fixturePaths {
		_, _, err := tree.AddPath(element, FileInfo{})
		if err != nil {
			t.Errorf("could not setup test: %v", err)
		}
	}

	tree.SetPwd("/var/lib")
	expected := `..
alsa
apt
fprint
grub
snmp
sudo
systemd
`
	actual := tree.String(false)
	if expected != actual {
		t.Errorf("Expected tree representation: \n--->%s<---\nGot:\n--->%s<---", expected, actual)
	}

	expected = `fprint
grub
`
	actual = tree.StringBetween(3, 5, false)
	if expected != actual {
		t.Errorf("Expected tree representation: \n--->%s<---\nGot:\n--->%s<---", expected, actual)
	}

}

func TestRejectRelativePath(t *testing.T) {
	tree := NewFileTreeModel()
	_, _, err := tree.AddPath("./var/lib/systemd", FileInfo{})
	if err != nil {
		t.Errorf("could not setup test: %v", err)
	}
	_, _, err = tree.AddPath("./", FileInfo{})

	if err == nil {
		t.Errorf("expected to reject relative fqfp, but did not")
	}

}

func TestAddRelativePath(t *testing.T) {
	tree := NewFileTreeModel()
	_, _, err := tree.AddPath("./var/lib/systemd", FileInfo{})
	if err != nil {
		t.Errorf("could not setup test: %v", err)
	}

	expected := "var\n"
	actual := tree.String(false)

	if expected != actual {
		t.Errorf("Expected tree string:\n--->%s<---\nGot:\n--->%s<---", expected, actual)
	}
}

func TestAddPath(t *testing.T) {
	tree := NewFileTreeModel()

	fixturePaths := []string{
		"/bin/cat",
		"/bin/chmod",
		"/bin/chown",
		"/bin/cp",
		"/bin/date",
		"/bin/dd",
		"/bin/df",
		"/bin/dmesg",
		"/bin/echo",
	}

	for _, element := range fixturePaths {
		_, _, err := tree.AddPath(element, FileInfo{})
		if err != nil {
			t.Errorf("could not setup test: %v", err)
		}
	}

	tree.SetPwd("/bin")
	expected := `..
cat
chmod
chown
cp
date
dd
df
dmesg
echo
`
	actual := tree.String(false)
	if expected != actual {
		t.Errorf("Expected tree representation: \n--->%s<---\nGot:\n--->%s<---", expected, actual)
	}
}

func TestRemovePath(t *testing.T) {
	tree := NewFileTreeModel()
	fixturePaths := []string{
		"/bin/cat",
		"/bin/chmod",
		"/bin/chown",
		"/bin/cp",
		"/bin/date",
		"/bin/dd",
		"/bin/df",
		"/bin/dmesg",
		"/bin/echo",
		"/var/lib/sudo",
		"/var/lib/systemd",
		"/var/lib/snmp",
		"/var/lib/grub",
		"/var/lib/fprint",
		"/var/lib/apt",
		"/var/lib/alsa",
	}

	for _, element := range fixturePaths {
		_, _, err := tree.AddPath(element, FileInfo{})
		if err != nil {
			t.Errorf("could not setup test: %v", err)
		}
	}

	fixtureDeletes := []string{
		"/bin/cat",
		"/bin/chmod",
		"/var/lib/apt",
		"/var/lib/alsa",
	}
	for _, element := range fixtureDeletes {
		err := tree.RemovePath(element)
		if err != nil {
			t.Errorf("could not setup test: %v", err)
		}
	}

	// format: <pwd: expectedOutput>
	fixtureOutput := map[string]string{
		"/bin":         `..
chown
cp
date
dd
df
dmesg
echo
`,
		"/var/lib":      `..
fprint
grub
snmp
sudo
systemd
`,
	}

	for pwd, expectedOutput := range fixtureOutput {
		tree.SetPwd(pwd)
		actual := tree.String(false)

		if expectedOutput != actual {
			t.Errorf("Expected tree representation for %s:\n--->%s<---\nGot:\n--->%s<---",
				pwd, expectedOutput, actual)
		}
	}
}

func TestStack(t *testing.T) {
	treeA := NewFileTreeModel()
	treeB := NewFileTreeModel()

	fixturePathsA := []string{
		"/bin/chown",
		"/bin/cp",
		"/bin/date",
		"/bin/dd",
		"/bin/df",
		"/bin/dmesg",
		"/bin/echo",
	}

	fixturePathsB := []string{
		"/var/lib/sudo",
		"/var/lib/systemd",
		"/var/lib/snmp",
		"/var/lib/grub",
		"/var/lib/fprint",
	}

	for _, element := range fixturePathsA {
		_, _, err := treeA.AddPath(element, FileInfo{})
		if err != nil {
			t.Errorf("could not setup test: %v", err)
		}
	}

	for _, element := range fixturePathsB {
		_, _, err := treeB.AddPath(element, FileInfo{})
		if err != nil {
			t.Errorf("could not setup test: %v", err)
		}
	}

	expectedOutputA := `..
chown
cp
date
dd
df
dmesg
echo
`
	treeA.SetPwd("/bin")
	actual := treeA.String(false)
	if expectedOutputA != actual {
		t.Errorf("Expected tree representation for:\n--->%s<---\nGot:\n--->%s<---", expectedOutputA, actual)
	}

 expectedOutputB := `..
fprint
grub
snmp
sudo
systemd
`
	treeB.SetPwd("/var/lib")
	actual = treeB.String(false)
	if expectedOutputB != actual {
		t.Errorf("Expected tree representation for:\n--->%s<---\nGot:\n--->%s<---", expectedOutputB, actual)
	}

	failedPaths, err := treeA.Stack(treeB)
	if err != nil {
		t.Errorf("Could not stack refTrees: %v", err)
	}
	if len(failedPaths) > 0 {
		t.Errorf("expected no filepath errors, got %d", len(failedPaths))
	}

	//treeA.SetPwd("/bin")
	actual = treeA.String(false)
	if expectedOutputA != actual {
		t.Errorf("Expected tree representation for:\n--->%s<---\nGot:\n--->%s<---", expectedOutputA, actual)
	}

	treeA.SetPwd("/var/lib")
	actual = treeA.String(false)
	if expectedOutputB != actual {
		t.Errorf("Expected tree representation for:\n--->%s<---\nGot:\n--->%s<---", expectedOutputA, actual)
	}
}

func TestClone(t *testing.T) {
	tree := NewFileTreeModel()
	fixturePaths := []string{
		"/bin/chown",
		"/bin/cp",
		"/bin/date",
		"/bin/dd",
		"/bin/df",
		"/bin/dmesg",
		"/bin/echo",
	}

	for _, element := range fixturePaths {
		_, _, err := tree.AddPath(element, FileInfo{})
		if err != nil {
			t.Errorf("could not setup test: %v", err)
		}
	}

	expectedOutput := `..
chown
cp
date
dd
df
dmesg
echo
`
	tree.SetPwd("/bin")
	actual := tree.String(false)
	if expectedOutput != actual {
		t.Errorf("Expected tree representation for:\n--->%s<---\nGot:\n--->%s<---", expectedOutput, actual)
	}

	actual = tree.Clone().String(false)
	if expectedOutput != actual {
		t.Errorf("Expected tree string:\n--->%s<---\nGot:\n--->%s<---", expectedOutput, actual)
	}
}

func TestCompareWithNoChanges(t *testing.T) {
	treeA := NewFileTreeModel()
	treeB := NewFileTreeModel()
	fixturePaths := [...]string{"/etc", "/etc/sudoers", "/etc/hosts", "/usr/bin", "/usr/bin/bash", "/usr"}

	for _, value := range fixturePaths {
		dummyInfo := FileInfo{
			Fqfp: value,
			Mode: 0, // regular file
			hash: 123,
		}
		_, _, err := treeA.AddPath(value, dummyInfo)
		if err != nil {
			t.Errorf("could not setup test: %v", err)
		}
		_, _, err = treeB.AddPath(value, dummyInfo)
		if err != nil {
			t.Errorf("could not setup test: %v", err)
		}
	}

	err := CompareAndMark(treeA, treeB)
	if err != nil {
		t.Errorf("could not setup test: %v", err)
	}

	visitor := func(n *FileNode) error {
		if n.AbsPath() == "/" {
			return nil
		}
		if (n.Data.DiffType) != Unmodified {
			t.Errorf("Expecting node at %s to have DiffType unchanged, but had %v", n.AbsPath(), n.Data.DiffType)
		}
		return nil
	}
	err = treeA.DepthFirstSearch(visitor, nil)
	if err != nil {
		t.Error(err)
	}
	err = treeB.DepthFirstSearch(visitor, nil)
	if err != nil {
		t.Error(err)
	}
}

func TestCompareWithAddons(t *testing.T) {
	treeA := NewFileTreeModel()
	treeB := NewFileTreeModel()
	fixturePathsA := [...]string{"/etc", "/etc/sudoers", "/usr", "/etc/hosts", "/usr/bin"}
	fixturePathsB := [...]string{"/etc", "/etc/sudoers", "/usr", "/etc/hosts", "/usr/bin", "/usr/bin/bash", "/etc/fstab"}

	for _, fqfp := range fixturePathsA {
		_, _, err := treeA.AddPath(fqfp, FileInfo{
			Fqfp: fqfp,
			Mode: 0, // regular file
			hash: 123,
		})
		if err != nil {
			t.Errorf("could not setup test: %v", err)
		}
	}

	for _, fqfp := range fixturePathsB {
		_, _, err := treeB.AddPath(fqfp, FileInfo{
			Fqfp: fqfp,
			Mode: 0, // regular file
			hash: 123,
		})
		if err != nil {
			t.Errorf("could not setup test: %v", err)
		}
	}

	var failedAssertions []error
	err := CompareAndMark(treeA, treeB)
	if err != nil {
		t.Errorf("Expected tree compare to have no errors, got: %v", err)
	}

	visitorA := func(n *FileNode) error {
		if n.AbsPath() == "/" {
			return nil
		}
		if (n.Data.DiffType) != Unmodified {
			t.Errorf("Expecting node at %s to have DiffType unchanged, but had %v", n.AbsPath(), n.Data.DiffType)
		}
		return nil
	}

	err = treeA.DepthFirstSearch(visitorA, nil)
	if err != nil {
		t.Errorf("Expected no errors when visiting nodes, got: %+v", err)
	}

	visitorB := func(n *FileNode) error {
		p := n.AbsPath()
		if p == "/" {
			return nil
		} else if stringInSlice(p, []string{"/usr/bin/bash", "/etc/fstab"}) {
			// two files were "added" in treeB
			if err := AssertDiffType(n, Added); err != nil {
				failedAssertions = append(failedAssertions, err)
			}
		} else if stringInSlice(p, []string{"/usr", "/usr/bin", "/etc"}) {
			// three folders contains differences and as results are considered "modified" in treeB
			if err := AssertDiffType(n, Modified); err != nil {
				failedAssertions = append(failedAssertions, err)
			}
		} else {
			if err := AssertDiffType(n, Unmodified); err != nil {
				failedAssertions = append(failedAssertions, err)
			}
		}
		return nil
	}
	err = treeB.DepthFirstSearch(visitorB, nil)
	if err != nil {
		t.Errorf("Expected no errors when visiting nodes, got: %+v", err)
	}

	if len(failedAssertions) > 0 {
		str := "\n"
		for _, value := range failedAssertions {
			str += fmt.Sprintf("  - %s\n", value.Error())
		}
		t.Errorf("Expected no errors when evaluating nodes, got: %s", str)
	}
}

func TestCompareWithChanges(t *testing.T) {
	treeA := NewFileTreeModel()
	treeB := NewFileTreeModel()
	fixturePaths := []string{"/etc", "/usr", "/etc/hosts", "/etc/sudoers", "/usr/bin"}

	// use-case different hash
	for _, fqfp := range fixturePaths {
		_, _, err := treeA.AddPath(fqfp, FileInfo{
			Fqfp: fqfp,
			Mode: 0, // regular file
			hash: 123,
		})
		if err != nil {
			t.Errorf("could not setup test: %v", err)
		}

		_, _, err = treeB.AddPath(fqfp, FileInfo{
			Fqfp: fqfp,
			Mode: 0, // regular file
			hash: 456,
		})
		if err != nil {
			t.Errorf("could not setup test: %v", err)
		}
	}

	// use-case different file permissions
	chmodPath := "/etc/fstab"
	_, _, err := treeA.AddPath(chmodPath, FileInfo{
		Fqfp: chmodPath,
		Mode: 0777,
		hash: 123,
	})
	if err != nil {
		t.Errorf("could not setup test: %v", err)
	}

	_, _, err = treeB.AddPath(chmodPath, FileInfo{
		Fqfp: chmodPath,
		Mode: 0,
		hash: 123,
	})
	if err != nil {
		t.Errorf("could not setup test: %v", err)
	}
	fixturePaths = append(fixturePaths, chmodPath)

	// use-case different owner
	chownPath := "/etc/gshadow"
	_, _, err = treeA.AddPath(chmodPath, FileInfo{
		Fqfp: chownPath,
		Mode: 0, // regular file,
		Gid:  0,
		Uid:  0,
	})
	if err != nil {
		t.Errorf("could not setup test: %v", err)
	}

	_, _, err = treeB.AddPath(chmodPath, FileInfo{
		Fqfp: chownPath,
		Mode: 0, // regular file
		Gid:  12,
		Uid:  12,
	})
	if err != nil {
		t.Errorf("could not setup test: %v", err)
	}

	fixturePaths = append(fixturePaths, chownPath)
	err = CompareAndMark(treeA, treeB)
	if err != nil {
		t.Errorf("unable to compare and mark: %+v", err)
	}

	var failedAssertions []error
	visitor := func(n *FileNode) error {
		p := n.AbsPath()
		if p == "/" {
			return nil
		} else if stringInSlice(p, fixturePaths) {
			if err := AssertDiffType(n, Modified); err != nil {
				failedAssertions = append(failedAssertions, err)
			}
		} else {
			if err := AssertDiffType(n, Unmodified); err != nil {
				failedAssertions = append(failedAssertions, err)
			}
		}
		return nil
	}
	err = treeA.DepthFirstSearch(visitor, nil)
	if err != nil {
		t.Errorf("Expected no errors when visiting nodes, got: %+v", err)
	}

	if len(failedAssertions) > 0 {
		str := "\n"
		for _, value := range failedAssertions {
			str += fmt.Sprintf("  - %s\n", value.Error())
		}
		t.Errorf("Expected no errors when evaluating nodes, got: %s", str)
	}
}

func TestCompareWithRemoves(t *testing.T) {
	treeA := NewFileTreeModel()
	treeB := NewFileTreeModel()
	fixturePathsA := [...]string{"/etc", "/etc/sudoers", "/usr", "/etc/hosts", "/usr/bin", "/usr/bin/bash", "/etc/fstab"}
	fixturePathsB := [...]string{"/etc", "/etc/sudoers", "/usr", "/etc/hosts", "/usr/bin"}

	for _, fqfp := range fixturePathsA {
		_, _, err := treeA.AddPath(fqfp, FileInfo{
			Fqfp: fqfp,
			Mode: 0, // regular file
			hash: 123,
		})
		if err != nil {
			t.Errorf("could not setup test: %v", err)
		}
	}

	for _, fqfp := range fixturePathsB {
		_, _, err := treeB.AddPath(fqfp, FileInfo{
			Fqfp: fqfp,
			Mode: 0, // regular file
			hash: 123,
		})
		if err != nil {
			t.Errorf("could not setup test: %v", err)
		}
	}

	err := CompareAndMark(treeA, treeB)
	if err != nil {
		t.Errorf("could not setup test: %v", err)
	}

	var failedAssertions []error
	visitorA := func(n *FileNode) error {
		p := n.AbsPath()
		if p == "/" {
			return nil
		} else if stringInSlice(p, []string{"/usr/bin/bash", "/etc/fstab"}) {
			if err := AssertDiffType(n, Added); err != nil {
				failedAssertions = append(failedAssertions, err)
			}
		} else if stringInSlice(p, []string{"/usr", "/usr/bin", "/etc"}) {
			if err := AssertDiffType(n, Modified); err != nil {
				failedAssertions = append(failedAssertions, err)
			}
		} else {
			if err := AssertDiffType(n, Unmodified); err != nil {
				failedAssertions = append(failedAssertions, err)
			}
		}
		return nil
	}
	err = treeA.DepthFirstSearch(visitorA, nil)
	if err != nil {
		t.Errorf("Expected no errors when visiting nodes, got: %+v", err)
	}

	visitorB := func(n *FileNode) error {
		if n.AbsPath() == "/" {
			return nil
		}
		if (n.Data.DiffType) != Unmodified {
			t.Errorf("Expecting node at %s to have DiffType unchanged, but had %v", n.AbsPath(), n.Data.DiffType)
		}
		return nil
	}

	err = treeB.DepthFirstSearch(visitorB, nil)
	if err != nil {
		t.Errorf("Expected no errors when visiting nodes, got: %+v", err)
	}

	if len(failedAssertions) > 0 {
		str := "\n"
		for _, value := range failedAssertions {
			str += fmt.Sprintf("  - %s\n", value.Error())
		}
		t.Errorf("Expected no errors when evaluating nodes, got: %s", str)
	}
}

func TestStackRange(t *testing.T) {
	treeA := NewFileTreeModel()
	treeB := NewFileTreeModel()
	treeC := NewFileTreeModel()

	fixturePathsA := []string{
		"/bin/chown",
		"/bin/cp",
		"/bin/date",
		"/bin/dd",
		"/bin/df",
	}

	fixturePathsB := []string{
		"/var/lib/sudo",
		"/var/lib/systemd",
		"/var/lib/snmp",
		"/var/lib/grub",
		"/var/lib/fprint",
	}

	fixturePathsC := []string{
		"/etc/aliases",
		"/etc/bash.bashrc",
		"/etc/cups",
		"/etc/dhcp",
		"/etc/fstab",
	}

	mapping := map[*FileTreeModel] []string{
		treeA: fixturePathsA,
		treeB: fixturePathsB,
		treeC: fixturePathsC,
	}

	for tree, fixturePaths := range mapping {
		for _, element := range fixturePaths {
			_, _, err := tree.AddPath(element, FileInfo{})
			if err != nil {
				t.Errorf("could not setup test: %v", err)
			}
		}
	}

	trees := []*FileTreeModel{treeA, treeB, treeC}
	tree, failedPaths, err := StackTreeRange(trees, 0, 2)
	if len(failedPaths) > 0 {
		t.Errorf("expected no filepath errors, got %d", len(failedPaths))
	}
	if err != nil {
		t.Fatal(err)
	}

	// format: <pwd: expectedOutput>
	fixtureOutput := map[string]string{
		"/bin":         `..
chown
cp
date
dd
df
`,
		"/var/lib":      `..
fprint
grub
snmp
sudo
systemd
`,
		"/etc":      `..
aliases
bash.bashrc
cups
dhcp
fstab
`,
	}

	for pwd, expectedOutput := range fixtureOutput {
		tree.SetPwd(pwd)
		actual := tree.String(false)

		if expectedOutput != actual {
			t.Errorf("Expected tree representation for %s:\n--->%s<---\nGot:\n--->%s<---",
				pwd, expectedOutput, actual)
		}
	}
}

func TestRemoveOnIterate(t *testing.T) {
	tree := NewFileTreeModel()
	fixturePaths := []string{
		"/bin/cat",
		"/bin/chown",
		"/bin/cp",
		"/bin/date",
		"/bin/dd",
		"/bin/df",
		"/bin/dmesg",
		"/bin/echo",
	}

	for _, element := range fixturePaths {
		_, _, err := tree.AddPath(element, FileInfo{})
		if err != nil {
			t.Errorf("could not setup test: %v", err)
		}
	}

	// Visitor is a function that processes, observes, or otherwise transforms the given node
	visitor := func(node *FileNode) error {
		err := tree.RemovePath(node.AbsPath())
		if err != nil {
			t.Errorf("could not setup test: %v", err)
		}
		return nil
	}

	// VisitEvaluator is a function that returns True if the given node should be visited by a Visitor.
	evaluator := func(node *FileNode) bool {
		return strings.Contains(node.AbsPath(), "dmesg") || strings.Contains(node.AbsPath(), "echo")
	}

	// remove /bin/echo and /bin/dmesg
	err := tree.DepthFirstSearch(visitor, evaluator)
	if err != nil {
		t.Errorf("could not setup test: %v", err)
	}

	expected := `..
cat
chown
cp
date
dd
df
`
	tree.SetPwd("/bin/")
	actual := tree.String(false)
	if expected != actual {
		t.Errorf("Expected tree string:\n--->%s<---\nGot:\n--->%s<---", expected, actual)
	}
}
