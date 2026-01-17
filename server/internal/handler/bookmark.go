package handler

import (
	"net/http"
	"strconv"

	"Nibstash_v2_server/internal/model"
	"Nibstash_v2_server/internal/repository"

	"github.com/gin-gonic/gin"
)

type BookmarkHandler struct {
	bookmarkRepo *repository.BookmarkRepository
}

func NewBookmarkHandler() *BookmarkHandler {
	return &BookmarkHandler{
		bookmarkRepo: repository.NewBookmarkRepository(),
	}
}

// List 获取书签列表
func (h *BookmarkHandler) List(c *gin.Context) {
	var req model.BookmarkListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	bookmarks, total, err := h.bookmarkRepo.List(
		req.Page, req.PageSize, req.TagID, req.Search,
		req.FolderPath, req.FilterFolder, req.SortBy,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取书签列表失败"})
		return
	}

	if bookmarks == nil {
		bookmarks = []model.Bookmark{}
	}

	c.JSON(http.StatusOK, model.BookmarkListResponse{
		Bookmarks: bookmarks,
		Total:     total,
		Page:      req.Page,
		PageSize:  req.PageSize,
	})
}

// Create 创建书签
func (h *BookmarkHandler) Create(c *gin.Context) {
	var req model.BookmarkCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 检查是否已存在
	if h.bookmarkRepo.Exists(req.URL, req.FolderPath) {
		c.JSON(http.StatusConflict, gin.H{"error": "书签已存在"})
		return
	}

	bookmark, err := h.bookmarkRepo.Create(
		req.URL, req.Title, req.Description, req.Favicon, req.FolderPath, req.TagIDs,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建书签失败"})
		return
	}

	c.JSON(http.StatusCreated, bookmark)
}

// Get 获取书签详情
func (h *BookmarkHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	bookmark, err := h.bookmarkRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "书签不存在"})
		return
	}

	c.JSON(http.StatusOK, bookmark)
}

// Update 更新书签
func (h *BookmarkHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var req model.BookmarkUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 获取现有书签
	existing, err := h.bookmarkRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "书签不存在"})
		return
	}

	// 合并更新
	url := existing.URL
	title := existing.Title
	description := existing.Description
	if req.URL != "" {
		url = req.URL
	}
	if req.Title != "" {
		title = req.Title
	}
	if req.Description != "" {
		description = req.Description
	}

	if err := h.bookmarkRepo.Update(id, url, title, description, req.TagIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新书签失败"})
		return
	}

	bookmark, _ := h.bookmarkRepo.GetByID(id)
	c.JSON(http.StatusOK, bookmark)
}

// Delete 删除书签
func (h *BookmarkHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	if err := h.bookmarkRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除书签失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// Batch 批量操作
func (h *BookmarkHandler) Batch(c *gin.Context) {
	var req model.BookmarkBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	switch req.Action {
	case "delete":
		if err := h.bookmarkRepo.DeleteByIDs(req.IDs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "批量删除失败"})
			return
		}
	case "move":
		if err := h.bookmarkRepo.MoveToFolder(req.IDs, req.Target); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "批量移动失败"})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的操作"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "操作成功"})
}

// Export 导出书签
func (h *BookmarkHandler) Export(c *gin.Context) {
	bookmarks, err := h.bookmarkRepo.GetAllForExport()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "导出失败"})
		return
	}

	// 生成 HTML 格式的书签文件
	html := generateBookmarkHTML(bookmarks)

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=bookmarks.html")
	c.String(http.StatusOK, html)
}

func generateBookmarkHTML(bookmarks []model.Bookmark) string {
	html := `<!DOCTYPE NETSCAPE-Bookmark-file-1>
<META HTTP-EQUIV="Content-Type" CONTENT="text/html; charset=UTF-8">
<TITLE>Bookmarks</TITLE>
<H1>Bookmarks</H1>
<DL><p>
`
	currentFolder := ""
	folderDepth := 0

	for _, b := range bookmarks {
		// 处理文件夹变化
		if b.FolderPath != currentFolder {
			// 关闭之前的文件夹
			for i := 0; i < folderDepth; i++ {
				html += "</DL><p>\n"
			}
			folderDepth = 0

			// 打开新文件夹
			if b.FolderPath != "" {
				parts := splitPath(b.FolderPath)
				for _, part := range parts {
					html += "<DT><H3>" + escapeHTML(part) + "</H3>\n<DL><p>\n"
					folderDepth++
				}
			}
			currentFolder = b.FolderPath
		}

		html += "<DT><A HREF=\"" + escapeHTML(b.URL) + "\">" + escapeHTML(b.Title) + "</A>\n"
	}

	// 关闭所有文件夹
	for i := 0; i < folderDepth; i++ {
		html += "</DL><p>\n"
	}

	html += "</DL><p>\n"
	return html
}

func splitPath(path string) []string {
	if path == "" {
		return nil
	}
	var parts []string
	start := 0
	for i := 0; i <= len(path); i++ {
		if i == len(path) || path[i] == '/' {
			if i > start {
				parts = append(parts, path[start:i])
			}
			start = i + 1
		}
	}
	return parts
}

func escapeHTML(s string) string {
	result := ""
	for _, c := range s {
		switch c {
		case '<':
			result += "&lt;"
		case '>':
			result += "&gt;"
		case '&':
			result += "&amp;"
		case '"':
			result += "&quot;"
		default:
			result += string(c)
		}
	}
	return result
}

// ClearAll 清空所有书签
func (h *BookmarkHandler) ClearAll(c *gin.Context) {
	if err := h.bookmarkRepo.DeleteAll(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "清空失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "清空成功"})
}

// ClearFolder 清空指定文件夹的书签
func (h *BookmarkHandler) ClearFolder(c *gin.Context) {
	var req struct {
		FolderPath string `json:"folder_path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	if err := h.bookmarkRepo.DeleteByFolder(req.FolderPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "清空失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "清空成功"})
}
