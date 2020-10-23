package model

// NodeData is the payload for a FileNode
type NodeData struct {
	FileInfo FileInfo
	DiffType DiffType
	Hidden   bool
}

// NewNodeData creates an empty NodeData struct for a FileNode
func NewNodeData() *NodeData {
	return &NodeData{
		Hidden:   false,
		FileInfo: FileInfo{},
		DiffType: Unmodified,
	}
}

// Copy duplicates a NodeData
func (data *NodeData) Copy() *NodeData {
	return &NodeData{
		Hidden:   data.Hidden,
		FileInfo: *data.FileInfo.Copy(),
		DiffType: data.DiffType,
	}
}
