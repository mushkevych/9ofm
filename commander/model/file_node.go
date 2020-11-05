package model

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"sort"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/phayes/permbits" // plan9 incompatible
)

const (
	AttributeFormat = "%s%s %11s %10s "
)

var diffTypeColor = map[DiffType]*color.Color{
	Added:      color.New(color.FgGreen),
	Removed:    color.New(color.FgRed),
	Modified:   color.New(color.FgYellow),
	Unmodified: color.New(color.Reset),
}

// FileNode represents a single file, its relation to files beneath it, the tree it exists in, and the metadata of the given file.
type FileNode struct {
	Tree     *FileTreeModel
	Parent   *FileNode
	Children map[string]*FileNode
	Data     NodeData

	Name string
	fqfp string
}

// NewFileNode creates a new FileNode relative to the given parent node with a payload.
func NewFileNode(parent *FileNode, fqfp string, name string, info FileInfo) (node *FileNode) {
	if fqfp == "" && name == ".." {
		log.Errorf("absolute path must be provided for .. reference")
		return nil
	}

	node = new(FileNode)
	node.Name = name
	node.Data = *NewNodeData()
	node.Data.FileInfo = *info.Clone()

	node.Children = make(map[string]*FileNode)
	node.Parent = parent
	if parent != nil {
		node.Tree = parent.Tree
	}

	node.fqfp = fqfp
	if node.fqfp == "" {
		node.fqfp = node.buildAbsPath()
	}
	return node
}

// Copy duplicates the existing node relative to a new parent node.
func (node *FileNode) Copy(parent *FileNode) *FileNode {
	newNode := NewFileNode(parent, "", node.Name, node.Data.FileInfo)
	newNode.Data.DiffType = node.Data.DiffType
	for name, child := range node.Children {
		newNode.Children[name] = child.Copy(newNode)
		child.Parent = newNode
	}
	return newNode
}

// AddChild creates a new node relative to the current FileNode.
func (node *FileNode) AddChild(name string, info FileInfo) *FileNode {
	if node.Children[name] != nil {
		// tree node already exists, replace the payload, keep the children
		node.Children[name].Data.FileInfo = *info.Clone()
	} else {
		node.Children[name] = NewFileNode(node, info.Fqfp, name, info)
		node.Tree.Size++
	}

	return node.Children[name]
}

// Remove deletes the current FileNode from it's parent FileNode's relations.
func (node *FileNode) Remove() error {
	if node == node.Tree.Root {
		return fmt.Errorf("cannot remove the tree root")
	}
	for _, child := range node.Children {
		err := child.Remove()
		if err != nil {
			return err
		}
	}
	delete(node.Parent.Children, node.Name)
	node.Tree.Size--
	return nil
}

// String shows the filename formatted into the proper color (by DiffType), additionally indicating if it is a symlink.
func (node *FileNode) String() string {
	var display string
	if node == nil {
		return ""
	}

	display = node.Name
	if node.Data.FileInfo.Mode == os.ModeSymlink {
		display += " → " + node.Data.FileInfo.Linkname
	}
	return diffTypeColor[node.Data.DiffType].Sprint(display)
}

// MetadataString returns the FileNode metadata in a columnar string.
func (node *FileNode) MetadataString() string {
	if node == nil {
		return ""
	}

	dir := "-"
	if node.Data.FileInfo.IsDir() {
		dir = "d"
	}
	user := node.Data.FileInfo.Uid
	group := node.Data.FileInfo.Gid
	userGroup := fmt.Sprintf("%d:%d", user, group)

	var sizeBytes int64

	if node.IsLeaf() {
		sizeBytes = node.Data.FileInfo.Size
	} else {
		sizer := func(curNode *FileNode) error {
			// don't include file sizes of children that have been removed (unless the node in question is a removed dir,
			// then show the accumulated size of removed files)
			if curNode.Data.DiffType != Removed || node.Data.DiffType == Removed {
				sizeBytes += curNode.Data.FileInfo.Size
			}
			return nil
		}

		err := node.VisitDepthChildFirst(sizer, nil)
		if err != nil {
			log.Errorf("unable to propagate node for metadata: %+v", err)
		}
	}

	size := humanize.Bytes(uint64(sizeBytes))

	// FIXME: this line is plan9 incompatible
	fileMode := permbits.FileMode(node.Data.FileInfo.Mode).String()
	return diffTypeColor[node.Data.DiffType].Sprint(fmt.Sprintf(AttributeFormat, dir, fileMode, userGroup, size))
}

// VisitDepthChildFirst iterates a tree depth-first (starting at this FileNode), evaluating the deepest depths first (visit on bubble up)
func (node *FileNode) VisitDepthChildFirst(visitor Visitor, evaluator VisitEvaluator) error {
	var keys []string
	for key := range node.Children {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, name := range keys {
		child := node.Children[name]
		err := child.VisitDepthChildFirst(visitor, evaluator)
		if err != nil {
			return err
		}
	}
	// never visit the root node
	if node == node.Tree.Root {
		return nil
	} else if evaluator != nil && evaluator(node) || evaluator == nil {
		return visitor(node)
	}

	return nil
}

// VisitDepthParentFirst iterates a tree depth-first (starting at this FileNode), evaluating the shallowest depths first (visit while sinking down)
func (node *FileNode) VisitDepthParentFirst(visitor Visitor, evaluator VisitEvaluator) error {
	var err error

	doVisit := evaluator != nil && evaluator(node) || evaluator == nil

	if !doVisit {
		return nil
	}

	// never visit the root node
	if node != node.Tree.Root {
		err = visitor(node)
		if err != nil {
			return err
		}
	}

	var keys []string
	for key := range node.Children {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, name := range keys {
		child := node.Children[name]
		err = child.VisitDepthParentFirst(visitor, evaluator)
		if err != nil {
			return err
		}
	}
	return err
}

// IsLeaf returns true is the current node has no child nodes.
func (node *FileNode) IsLeaf() bool {
	return len(node.Children) == 0
}

// IsDir returns true is the current node is a Directory
func (node *FileNode) IsDir() bool {
	return node.Data.FileInfo.IsDir()
}

// AbsPath returns a slash-delimited absolute fqfp of the given file
func (node *FileNode) AbsPath() string {
	return node.fqfp
}

// buildAbsPath rebuilds slash-delimited absolute path of the given file by iterating over its Parental nodes
func (node *FileNode) buildAbsPath() string {
	var fqfpParent string
	if node.Parent != nil {
		// ask Parent for its AbsPath
		fqfpParent = node.Parent.AbsPath()
		if node.Parent.AbsPath() != "/" {
			fqfpParent += string(os.PathSeparator)
		}
	}
	return fqfpParent + node.Name
}

// deriveDiffType determines a DiffType to the current FileNode. Note: the DiffType of a node is always the DiffType of
// its attributes and its contents. The contents are the bytes of the file of the children of a directory.
func (node *FileNode) deriveDiffType(diffType DiffType) error {
	if node.IsLeaf() {
		return node.AssignDiffType(diffType)
	}

	myDiffType := diffType
	for _, v := range node.Children {
		myDiffType = myDiffType.merge(v.Data.DiffType)
	}

	return node.AssignDiffType(myDiffType)
}

// AssignDiffType will assign the given DiffType to this node, possibly affecting child nodes.
func (node *FileNode) AssignDiffType(diffType DiffType) error {
	var err error

	node.Data.DiffType = diffType

	if diffType == Removed {
		// if we've removed this node, then all children have been removed as well
		for _, child := range node.Children {
			err = child.AssignDiffType(diffType)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// compare the current node against the given node, returning a definitive DiffType.
func (node *FileNode) compare(other *FileNode) DiffType {
	if node == nil && other == nil {
		return Unmodified
	}

	if node == nil && other != nil {
		return Added
	}

	if node != nil && other == nil {
		return Removed
	}

	if node.Name != other.Name {
		panic("comparing mismatched nodes")
	}

	return node.Data.FileInfo.Compare(other.Data.FileInfo)
}
