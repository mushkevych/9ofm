package view

import (
	"bytes"
	"github.com/sergi/go-diff/diffmatchpatch"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/mushkevych/9ofm/commander/model"
)

const allowTestDataCapture = false

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func testCaseDataFilePath(name string) string {
	return filepath.Join("testdata", name+".txt")
}

func helperLoadBytes(t *testing.T) []byte {
	path := testCaseDataFilePath(t.Name())
	theBytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("unable to load test data ('%s'): %+v", t.Name(), err)
	}
	return theBytes
}

func helperCaptureBytes(t *testing.T, data []byte) {
	if !allowTestDataCapture {
		t.Fatalf("cannot capture data in test mode: %s", t.Name())
	}

	path := testCaseDataFilePath(t.Name())
	err := ioutil.WriteFile(path, data, 0644)

	if err != nil {
		t.Fatalf("unable to save test data ('%s'): %+v", t.Name(), err)
	}
}

func helperCheckDiff(t *testing.T, expected, actual []byte) {
	if !bytes.Equal(expected, actual) {
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(string(expected), string(actual), true)
		t.Errorf(dmp.DiffPrettyText(diffs))
		t.Errorf("%s: bytes mismatch", t.Name())
	}
}

func assertTestData(t *testing.T, actualBytes []byte) {
	path := testCaseDataFilePath(t.Name())
	if !fileExists(path) {
		if allowTestDataCapture {
			helperCaptureBytes(t, actualBytes)
		} else {
			t.Fatalf("missing test data: %s", path)
		}
	}
	expectedBytes := helperLoadBytes(t)
	helperCheckDiff(t, expectedBytes, actualBytes)
}

func initializeTestViewModel(t *testing.T) *FileTreeView {
	tree := model.NewFileTreeModel()
	view, err := NewFileTreeView(tree)
	if err != nil {
		t.Errorf("failed to create File Tree View: %v", err)
	}

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

		"/etc/aliases",
		"/etc/bash.bashrc",
		"/etc/cups",
		"/etc/dhcp",
		"/etc/fstab",

		"/var/lib/sudo",
		"/var/lib/systemd",
		"/var/lib/snmp",
		"/var/lib/grub",
		"/var/lib/fprint",
		"/var/lib/apt",
		"/var/lib/alsa",
	}

	for _, element := range fixturePaths {
		_, _, err := tree.AddPath(element, model.FileInfo{})
		if err != nil {
			t.Errorf("could not setup test: %v", err)
		}
	}

	return view
}

func runTestCase(t *testing.T, vm *FileTreeView, width, height int, filterRegex *regexp.Regexp) {
	output := vm.ModelTree.String(true)
	assertTestData(t, []byte(output))
}

func checkError(t *testing.T, err error, message string) {
	if err != nil {
		t.Errorf(message+": %+v", err)
	}
}

func TestFileTreeGoCase(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 1000
	runTestCase(t, vm, width, height, nil)
}

func TestFileTreeNoAttributes(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 1000
	runTestCase(t, vm, width, height, nil)
}

func TestFileTreeRestrictedHeight(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 20
	runTestCase(t, vm, width, height, nil)
}

func TestFileTreeFilterTree(t *testing.T) {
	vm := initializeTestViewModel(t)
	width, height := 100, 1000

	regex, err := regexp.Compile("network")
	if err != nil {
		t.Errorf("could not create filter regex: %+v", err)
	}

	runTestCase(t, vm, width, height, regex)
}

func TestFileTreeHideAddedRemovedModified(t *testing.T) {
	vm := initializeTestViewModel(t)
	width, height := 100, 100

	// hide added files
	vm.ToggleShowDiffType(model.Added)

	// hide modified files
	vm.ToggleShowDiffType(model.Modified)

	// hide removed files
	vm.ToggleShowDiffType(model.Removed)

	runTestCase(t, vm, width, height, nil)
}

func TestFileTreeHideUnmodified(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 100

	// hide unmodified files
	vm.ToggleShowDiffType(model.Unmodified)

	runTestCase(t, vm, width, height, nil)
}

func TestFileTreeHideTypeWithFilter(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 100

	// hide added files
	vm.ToggleShowDiffType(model.Added)

	regex, err := regexp.Compile("saved")
	if err != nil {
		t.Errorf("could not create filter regex: %+v", err)
	}

	runTestCase(t, vm, width, height, regex)
}
