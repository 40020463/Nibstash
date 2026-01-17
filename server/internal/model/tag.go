package model

type Tag struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
	Count int    `json:"count,omitempty"`
}

type TagCreateRequest struct {
	Name  string `json:"name" binding:"required"`
	Color string `json:"color"`
}

type TagUpdateRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}
