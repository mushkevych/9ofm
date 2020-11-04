package model

// NodeData is the payload for a FileNode
type NodeData struct {
	FileInfo FileInfo
	DiffType DiffType
}

// NewNodeData creates an empty NodeData struct for a FileNode
func NewNodeData() *NodeData {
	return &NodeData{
		FileInfo: FileInfo{},
		DiffType: Unmodified,
	}
}

// Clone duplicates a NodeData
func (data *NodeData) Clone() *NodeData {
	return &NodeData{
		FileInfo: *data.FileInfo.Clone(),
		DiffType: data.DiffType,
	}
}
