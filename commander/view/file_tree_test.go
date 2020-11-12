package view

import (
	"bytes"
	"github.com/fatih/color"
	"github.com/sergi/go-diff/diffmatchpatch"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/mushkevych/9ofm/commander/format"
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

	format.Selected = color.New(color.ReverseVideo, color.Bold).SprintFunc()
	return view
}

func runTestCase(t *testing.T, vm *FileTreeView, width, height int, filterRegex *regexp.Regexp) {
	err := vm.Update(filterRegex, width, height)
	if err != nil {
		t.Errorf("failed to update view: %v", err)
	}

	err = vm.Render()
	if err != nil {
		t.Errorf("failed to render view: %v", err)
	}

	assertTestData(t, vm.Buffer.Bytes())
}

func checkError(t *testing.T, err error, message string) {
	if err != nil {
		t.Errorf(message+": %+v", err)
	}
}

func TestFileTreeGoCase(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 1000
	vm.Setup(0, height)
	vm.ShowFileAttributes = true

	runTestCase(t, vm, width, height, nil)
}

func TestFileTreeNoAttributes(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 1000
	vm.Setup(0, height)
	vm.ShowFileAttributes = false

	runTestCase(t, vm, width, height, nil)
}

func TestFileTreeRestrictedHeight(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 20
	vm.Setup(0, height)
	vm.ShowFileAttributes = false

	runTestCase(t, vm, width, height, nil)
}

func TestFileTreeDirCursor(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 100
	vm.Setup(0, height)
	vm.ShowFileAttributes = true

	moved := vm.CursorDown()
	if !moved {
		t.Error("unable to cursor down")
	}

	moved = vm.CursorDown()
	if !moved {
		t.Error("unable to cursor down")
	}

	runTestCase(t, vm, width, height, nil)
}

func TestFileTreeSelectLayer(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 100
	vm.Setup(0, height)
	vm.ShowFileAttributes = true

	runTestCase(t, vm, width, height, nil)
}

func TestFileShowAggregateChanges(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 100
	vm.Setup(0, height)
	vm.ShowFileAttributes = true

	runTestCase(t, vm, width, height, nil)
}

func TestFileTreePageDown(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 10
	vm.Setup(0, height)
	vm.ShowFileAttributes = true
	err := vm.Update(nil, width, height)
	checkError(t, err, "unable to update")

	err = vm.PageDown()
	checkError(t, err, "unable to page down")

	err = vm.PageDown()
	checkError(t, err, "unable to page down")

	err = vm.PageDown()
	checkError(t, err, "unable to page down")

	runTestCase(t, vm, width, height, nil)
}

func TestFileTreePageUp(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 10
	vm.Setup(0, height)
	vm.ShowFileAttributes = true

	// these operations have a render step for intermediate results, which require at least one update to be done first
	err := vm.Update(nil, width, height)
	checkError(t, err, "unable to update")

	err = vm.PageDown()
	checkError(t, err, "unable to page down")

	err = vm.PageDown()
	checkError(t, err, "unable to page down")

	err = vm.PageUp()
	checkError(t, err, "unable to page up")

	runTestCase(t, vm, width, height, nil)
}

func TestFileTreeDirNavigateTo(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 100
	vm.Setup(0, height)
	vm.ShowFileAttributes = true

	moved := vm.CursorDown()
	if !moved {
		t.Error("unable to cursor down")
	}

	moved = vm.CursorDown()
	if !moved {
		t.Error("unable to cursor down")
	}

	// expand /etc
	//err := vm.NavigateTo(nil)
	//checkError(t, err, "unable to cursor right")

	runTestCase(t, vm, width, height, nil)
}

func TestFileTreeFilterTree(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 1000
	vm.Setup(0, height)
	vm.ShowFileAttributes = true

	regex, err := regexp.Compile("network")
	if err != nil {
		t.Errorf("could not create filter regex: %+v", err)
	}

	runTestCase(t, vm, width, height, regex)
}

func TestFileTreeHideAddedRemovedModified(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 100
	vm.Setup(0, height)
	vm.ShowFileAttributes = true

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
	vm.Setup(0, height)
	vm.ShowFileAttributes = true

	// hide unmodified files
	vm.ToggleShowDiffType(model.Unmodified)

	runTestCase(t, vm, width, height, nil)
}

func TestFileTreeHideTypeWithFilter(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 100
	vm.Setup(0, height)
	vm.ShowFileAttributes = true

	// hide added files
	vm.ToggleShowDiffType(model.Added)

	regex, err := regexp.Compile("saved")
	if err != nil {
		t.Errorf("could not create filter regex: %+v", err)
	}

	runTestCase(t, vm, width, height, regex)
}
