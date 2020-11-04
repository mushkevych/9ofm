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

	// parent working directory - File Panel is build from this point
	pwd *FileNode

	// number of files in this Tree
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
	tree.Root.Tree = tree
	tree.Root.Children = make(map[string]*FileNode)
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

func (tree *FileTreeModel) getSortedChildrenKeys() []string {
	var keys []string
	for key := range tree.pwd.Children {
		keys = append(keys, key)
	}

	// in-place sorting
	sort.Strings(keys)
	return keys
}

// GetNodeAt returns FileNode representing n-th element in the FileTree by the 0-based index
func (tree *FileTreeModel) GetNodeAt(index int) *FileNode {
	keys := tree.getSortedChildrenKeys()

	if tree.pwd != tree.Root {
		if index == 0 {
			// 1st node is "..", which refers to the tree.pwd
			return tree.pwd.Parent
		} else {
			childKey := keys[index-1]
			return tree.pwd.Children[childKey]
		}
	} else {
		childKey := keys[index]
		return tree.Root.Children[childKey]
	}
}

func (tree *FileTreeModel) VisibleSize() int {
	if tree.pwd != tree.Root {
		// +1 includes ".." parent reference
		return len(tree.pwd.Children) + 1
	} else {
		return len(tree.Root.Children)
	}
}

// String returns the entire tree in an ASCII representation.
func (tree *FileTreeModel) String(showAttributes bool) string {
	return tree.StringBetween(0, tree.VisibleSize(), showAttributes)
}

// StringBetween returns a partial tree in an ASCII representation.
func (tree *FileTreeModel) StringBetween(start, stop int, showAttributes bool) string {
	// account for use case when list of available files is less than available visual area
	stop = utils.MinOf(stop, tree.VisibleSize())

	singleLine := func(node *FileNode) string {
		var line string
		if showAttributes {
			line += node.MetadataString() + " "
		}
		line += node.String() + newLine
		return line
	}

	var result string
	if start == 0 {
		if tree.pwd != tree.Root {
			if showAttributes {
				result += tree.pwd.Parent.MetadataString() + " "
			}
			result += ".." + newLine

			stop -= 1
		}
	}

	keys := tree.getSortedChildrenKeys()
	for i := start; i < stop; i++ {
		childKey := keys[i]
		childNode := tree.pwd.Children[childKey]
		result += singleLine(childNode)
	}

	return result
}

// Clone returns a copy of the given FileTreeModel
func (tree *FileTreeModel) Clone() *FileTreeModel {
	newTree := NewFileTreeModel()
	newTree.Size = tree.Size
	newTree.Root = tree.Root.Copy(newTree.Root)
	newTree.SetPwd(tree.pwd.fqfp)

	// update the tree pointers
	err := newTree.VisitDepthChildFirst(func(node *FileNode) error {
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

// VisitEvaluator is a function that indicates whether the given node should be visited by a Visitor.
type VisitEvaluator func(*FileNode) bool

// VisitDepthChildFirst iterates the given tree depth-first, evaluating the deepest depths first (visit on bubble up)
func (tree *FileTreeModel) VisitDepthChildFirst(visitor Visitor, evaluator VisitEvaluator) error {
	return tree.Root.VisitDepthChildFirst(visitor, evaluator)
}

// VisitDepthParentFirst iterates the given tree depth-first, evaluating the shallowest depths first (visit while sinking down)
func (tree *FileTreeModel) VisitDepthParentFirst(visitor Visitor, evaluator VisitEvaluator) error {
	return tree.Root.VisitDepthParentFirst(visitor, evaluator)
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

type compareMark struct {
	lowerNode *FileNode
	upperNode *FileNode
	tentative DiffType
	final     DiffType
}

// CompareAndMark marks the FileNodes in the owning (lower) tree with DiffType annotations when compared to the given (upper) tree.
func (tree *FileTreeModel) CompareAndMark(upper *FileTreeModel) ([]PathError, error) {
	// always compare relative to the original, unaltered tree.
	originalTree := tree

	modifications := make([]compareMark, 0)
	failed := make([]PathError, 0)

	graft := func(upperNode *FileNode) error {
		// note: since we are not comparing against the original tree (copying the tree is expensive) we may mark the parent
		// of an added node incorrectly as modified. This will be corrected later.
		originalLowerNode, _ := originalTree.GetNode(upperNode.AbsPath())

		if originalLowerNode == nil {
			_, newNodes, err := tree.AddPath(upperNode.AbsPath(), upperNode.Data.FileInfo)
			if err != nil {
				failed = append(failed, NewPathError(upperNode.AbsPath(), ActionAdd, err))
				return nil
			}
			for idx := len(newNodes) - 1; idx >= 0; idx-- {
				newNode := newNodes[idx]
				modifications = append(modifications, compareMark{lowerNode: newNode, upperNode: upperNode, tentative: -1, final: Added})
			}
			return nil
		}

		// the file exists in the lower layer
		lowerNode, _ := tree.GetNode(upperNode.AbsPath())
		diffType := lowerNode.compare(upperNode)
		modifications = append(modifications, compareMark{lowerNode: lowerNode, upperNode: upperNode, tentative: diffType, final: -1})

		return nil
	}
	// we must visit from the leaves upwards to ensure that diff types can be derived from and assigned to children
	err := upper.VisitDepthChildFirst(graft, nil)
	if err != nil {
		return failed, err
	}

	// take note of the comparison results on each note in the owning tree.
	for _, pair := range modifications {
		if pair.final > 0 {
			err = pair.lowerNode.AssignDiffType(pair.final)
			if err != nil {
				return failed, err
			}
		} else if pair.lowerNode.Data.DiffType == Unmodified {
			err = pair.lowerNode.deriveDiffType(pair.tentative)
			if err != nil {
				return failed, err
			}
		}

		// persist the upper's payload on the owning tree
		pair.lowerNode.Data.FileInfo = *pair.upperNode.Data.FileInfo.Clone()
	}
	return failed, nil
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
	stackErr = upper.VisitDepthChildFirst(graft, nil)
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
