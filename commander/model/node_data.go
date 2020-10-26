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

// Clone duplicates a NodeData
func (data *NodeData) Clone() *NodeData {
	return &NodeData{
		Hidden:   data.Hidden,
		FileInfo: *data.FileInfo.Clone(),
		DiffType: data.DiffType,
	}
}
