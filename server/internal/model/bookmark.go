package model

import "time"

type Bookmark struct {
	ID          int64     `json:"id"`
	URL         string    `json:"url"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	FolderPath  string    `json:"folder_path"`
	Favicon     string    `json:"favicon"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Tags        []Tag     `json:"tags,omitempty"`
}

type BookmarkCreateRequest struct {
	URL         string  `json:"url" binding:"required"`
	Title       string  `json:"title" binding:"required"`
	Description string  `json:"description"`
	FolderPath  string  `json:"folder_path"`
	Favicon     string  `json:"favicon"`
	TagIDs      []int64 `json:"tag_ids"`
}

type BookmarkUpdateRequest struct {
	URL         string  `json:"url"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	FolderPath  string  `json:"folder_path"`
	TagIDs      []int64 `json:"tag_ids"`
}

type BookmarkListRequest struct {
	Page         int    `form:"page" binding:"min=1"`
	PageSize     int    `form:"page_size" binding:"min=1,max=100"`
	TagID        int64  `form:"tag_id"`
	Search       string `form:"search"`
	FolderPath   string `form:"folder_path"`
	FilterFolder bool   `form:"filter_folder"`
	SortBy       string `form:"sort_by"`
}

type BookmarkListResponse struct {
	Bookmarks []Bookmark `json:"bookmarks"`
	Total     int        `json:"total"`
	Page      int        `json:"page"`
	PageSize  int        `json:"page_size"`
}

type BookmarkBatchRequest struct {
	Action string  `json:"action" binding:"required,oneof=delete move"`
	IDs    []int64 `json:"ids" binding:"required"`
	Target string  `json:"target"` // 用于 move 操作的目标文件夹
}

type ImportBookmark struct {
	URL        string
	Title      string
	FolderPath string
	Favicon    string
}
