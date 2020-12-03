package model

import (
	"fmt"
	"github.com/mushkevych/9ofm/utils"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	newLine = "\n"
)

// FileTreeModel represents a set of files, directories, and their relations.
type FileTreeModel struct {
	// absolute root of the filesystem
	Root *FileNode

	// parent working directory - pointer to a node in the tree
	// File Panel is build from this point
	pwd *FileNode

	// number of files and folders in this Tree
	// NOTE: this number is different from the len(pwd.Children)
	Size int

	Name string
	Id   uuid.UUID
}

// NewFileTreeModel creates an empty FileTreeModel
func NewFileTreeModel() (tree *FileTreeModel) {
	tree = new(FileTreeModel)
	tree.Size = 0
	tree.Root = new(FileNode)
	tree.Root.fqfp = "/"
	tree.Root.Name = "/"
	tree.Root.Tree = tree
	tree.Root.Children = make(map[string]*FileNode)
	tree.pwd = tree.Root
	tree.Id = uuid.New()
	return tree
}

func (tree *FileTreeModel) SetPwd(fqfp string) error {
	if fqfp == "/" {
		tree.pwd = tree.Root
		tree.pwd.fqfp = fqfp
	} else {
		fileNode, err := tree.GetNode(fqfp)
		if err != nil {
			return err
		}
		tree.pwd = fileNode
	}
	return nil
}

func (tree *FileTreeModel) GetPwd() string {
	return tree.pwd.AbsPath()
}

func (tree *FileTreeModel) sortedNamesInPwd() []string {
	var keys []string

	if tree.pwd != tree.Root {
		keys = append(keys, "..")
	}
	for key := range tree.pwd.Children {
		keys = append(keys, key)
	}

	// in-place sorting
	sort.Strings(keys)
	return keys
}

// GetNodeAt returns FileNode representing n-th element in the FileTree by the 0-based index
func (tree *FileTreeModel) GetNodeAt(index int) *FileNode {
	keys := tree.sortedNamesInPwd()

	if index == 0 && tree.pwd != tree.Root {
		// 1st node is "..", which refers to the tree.pwd
		return tree.pwd.Parent
	} else {
		childKey := keys[index]
		return tree.pwd.Children[childKey]
	}
}

// GetNodeByName returns FileNode by name in the subtree represented by PWD; supports ".." name
// returns nil if name can not be found
func (tree *FileTreeModel) GetNodeByName(name string) *FileNode {
	if node, ok := tree.pwd.Children[name]; ok {
		return node
	}
	if name == ".." && tree.pwd != tree.Root {
		return tree.pwd
	}
	return nil
}

func (tree *FileTreeModel) VisibleSize() int {
	if tree.pwd != tree.Root {
		// +1 includes ".." parent reference
		return len(tree.pwd.Children) + 1
	} else {
		return len(tree.pwd.Children)
	}
}

// String returns the entire tree in an ASCII representation.
func (tree *FileTreeModel) String(showAttributes bool) string {
	return tree.StringBetween(0, tree.VisibleSize(), showAttributes)
}

// StringBetween returns a partial tree in an ASCII representation.
// start is inclusive, 0-based index pointer
// stop is exclusive, 0-based index pointer
func (tree *FileTreeModel) StringBetween(start, stop int, showAttributes bool) string {
	// account for use case when list of available files is less than available visual area
	start = utils.MaxOf(start, 0)
	stop = utils.MinOf(stop, tree.VisibleSize())

	singleLine := func(node *FileNode) string {
		var line string
		if showAttributes {
			line += node.MetadataString() + " "
		}
		if node == tree.pwd {
			line += ".." + newLine
		} else {
			line += node.String() + newLine
		}
		return line
	}

	var result string
	keys := tree.sortedNamesInPwd()
	for i := start; i < stop; i++ {
		nodeNames := keys[i]
		node := tree.GetNodeByName(nodeNames)
		result += singleLine(node)
	}

	return result
}

// StringArrayBetween returns a partial tree in an ASCII representation.
// start is inclusive, 0-based index pointer
// stop is exclusive, 0-based index pointer
func (tree *FileTreeModel) StringArrayBetween(start, stop int) ([][]string, []*FileNode) {
	// account for use case when list of available files is less than available visual area
	start = utils.MaxOf(start, 0)
	stop = utils.MinOf(stop, tree.VisibleSize())

	singleLine := func(node *FileNode) []string {
		var line []string
		line = append(line, node.MetadataString())
		if node == tree.pwd {
			line = append(line, "..")
		} else {
			line = append(line, node.String())
		}
		return line
	}

	var fileNodes []*FileNode
	var result [][]string
	keys := tree.sortedNamesInPwd()
	for i := start; i < stop; i++ {
		nodeNames := keys[i]
		node := tree.GetNodeByName(nodeNames)
		result = append(result, singleLine(node))
		fileNodes = append(fileNodes, node)
	}

	return result, fileNodes
}


// Clone returns a copy of the given FileTreeModel
func (tree *FileTreeModel) Clone() *FileTreeModel {
	newTree := NewFileTreeModel()
	newTree.Size = tree.Size
	newTree.Root = tree.Root.Copy(newTree.Root)
	_ = newTree.SetPwd(tree.pwd.fqfp)

	// update the tree pointers
	err := newTree.DepthFirstSearch(func(node *FileNode) error {
		node.Tree = newTree
		return nil
	}, nil)

	if err != nil {
		log.Errorf("unable to propagate tree on copy(): %+v", err)
	}

	return newTree
}

// Visitor is a function that processes, observes, or otherwise transforms the given node
type Visitor func(*FileNode) error

// VisitEvaluator is a function that returns True if the given node should be visited by a Visitor.
type VisitEvaluator func(*FileNode) bool

// DepthFirstSearch starts at the tree root explores as far as possible along each branch before backtracking
func (tree *FileTreeModel) DepthFirstSearch(visitor Visitor, evaluator VisitEvaluator) error {
	return tree.Root.DepthFirstSearch(visitor, evaluator)
}

// GetNode fetches a single node when given a fully qualified file path - slash-delimited string
// from the root ('/') to the desired node (e.g. '/a/node/path.txt')
func (tree *FileTreeModel) GetNode(fqfp string) (*FileNode, error) {
	nodeNames := strings.Split(strings.Trim(fqfp, "/"), "/")
	node := tree.Root
	for _, name := range nodeNames {
		if name == "" {
			continue
		}
		if node.Children[name] == nil {
			return nil, fmt.Errorf("absolute path does not exist: %s", fqfp)
		}
		node = node.Children[name]
	}
	return node, nil
}

// AddPath adds a new node to the tree when given fully qualified file path and optional payload
func (tree *FileTreeModel) AddPath(fqfp string, info FileInfo) (*FileNode, []*FileNode, error) {
	fqfp = path.Clean(fqfp)
	if fqfp == "." {
		return nil, nil, fmt.Errorf("cannot add relative path '%s'", fqfp)
	}
	node := tree.Root
	addedNodes := make([]*FileNode, 0)

	currentFqfp := ""
	nodeNames := strings.Split(strings.Trim(fqfp, "/"), "/")
	for idx, name := range nodeNames {
		if name == "" {
			continue
		}

		currentFqfp += string(os.PathSeparator) + name
		if node.Children[name] != nil {
			node = node.Children[name]
		} else {
			if idx == len(nodeNames)-1 {
				// this is the actual child that was intended to be added
				node = node.AddChild(name, info)
			} else {
				// this is an intermediary node in the absolute path
				// for instance: /a/b/intermediary/the_child
				osFileInfo, err := os.Stat(currentFqfp)
				fileInfo := NewFileInfo(currentFqfp, osFileInfo, err)
				node = node.AddChild(name, fileInfo)
			}

			addedNodes = append(addedNodes, node)
			if node == nil {
				// the child could not be added
				return node, addedNodes, fmt.Errorf(fmt.Sprintf("could not add child node: '%s' (path: '%s')", name, fqfp))
			}
		}
	}
	return node, addedNodes, nil
}

// RemovePath removes a node from the tree given its fully qualified file path - slash-delimited string
// from the root ('/') to the desired node (e.g. '/a/node/path.txt').
func (tree *FileTreeModel) RemovePath(fqfp string) error {
	node, err := tree.GetNode(fqfp)
	if err != nil {
		return err
	}
	return node.Remove()
}

// CompareAndMark compares this treeA and treeB by performing two passages:
// 1: iterate over nodes in A and if they are absent in B - mark them as "deleted"
// 2: iterate over nodes in B and if they are absent in A - mark them as "added"
// *: rest of the nodes are compared for being "modified"
// NOTE: for every "added", "deleted" or "modified" node -  all their parents are marked as well as "modified"
func CompareAndMark(treeA, treeB *FileTreeModel) error {
	comparator := func(left, right *FileTreeModel) error {
		visitor := func(rightNode *FileNode) error {
			leftNode, _ := left.GetNode(rightNode.AbsPath())
			diffType := leftNode.compare(rightNode)
			if diffType != Unmodified {
				rightNode.Data.DiffType = diffType
				err := markParentsModified(rightNode)
				if err != nil {
					return err
				}
			}
			return nil
		}

		err := right.DepthFirstSearch(visitor, nil)
		if err != nil {
			return err
		}
		return nil
	}

	err := comparator(treeA, treeB)
	if err != nil {
		return err
	}

	err = comparator(treeB, treeA)
	if err != nil {
		return err
	}
	return nil
}

// markParentsModified traverses all parents and mark them as "modified"
func markParentsModified(node *FileNode) error {
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		err := parent.AssignDiffType(Modified)
		if err != nil {
			return err
		}
	}
	return nil
}

// markRemoved annotates the FileNode at the given fully qualified file path as Removed.
func (tree *FileTreeModel) markRemoved(fqfp string) error {
	node, err := tree.GetNode(fqfp)
	if err != nil {
		return err
	}
	return node.AssignDiffType(Removed)
}

// Stack takes two trees and combines them together. This is done by "stacking" the given tree on top of the owning tree.
func (tree *FileTreeModel) Stack(upper *FileTreeModel) (failed []PathError, stackErr error) {
	graft := func(node *FileNode) error {
		_, _, err := tree.AddPath(node.AbsPath(), node.Data.FileInfo)
		if err != nil {
			failed = append(failed, NewPathError(node.AbsPath(), ActionRemove, err))
		}
		return nil
	}
	stackErr = upper.DepthFirstSearch(graft, nil)
	return failed, stackErr
}

// StackTreeRange combines an array of trees into a single tree
func StackTreeRange(trees []*FileTreeModel, start, stop int) (*FileTreeModel, []PathError, error) {
	errors := make([]PathError, 0)
	tree := trees[0].Clone()
	for idx := start; idx <= stop; idx++ {
		failedPaths, err := tree.Stack(trees[idx])
		if len(failedPaths) > 0 {
			errors = append(errors, failedPaths...)
		}
		if err != nil {
			log.Errorf("could not stack tree range: %v", err)
			return nil, nil, err
		}
	}
	return tree, errors, nil
}
