package model

// FolderNode 文件夹树节点
type FolderNode struct {
	Name     string       `json:"name"`
	Path     string       `json:"path"`
	Count    int          `json:"count"`
	Children []FolderNode `json:"children,omitempty"`
}

type FolderCreateRequest struct {
	Path string `json:"path" binding:"required"`
}

type FolderMoveRequest struct {
	SourcePath string `json:"source_path" binding:"required"`
	TargetPath string `json:"target_path"`
}

type FolderMergeRequest struct {
	SourcePath string `json:"source_path" binding:"required"`
	TargetPath string `json:"target_path"`
}
