package handler

import (
	"net/http"
	"strconv"

	"Nibstash_v2_server/internal/repository"

	"github.com/gin-gonic/gin"
)

type FaviconHandler struct {
	bookmarkRepo *repository.BookmarkRepository
}

func NewFaviconHandler() *FaviconHandler {
	return &FaviconHandler{
		bookmarkRepo: repository.NewBookmarkRepository(),
	}
}

// GetPending 获取待处理的 favicon 列表
func (h *FaviconHandler) GetPending(c *gin.Context) {
	bookmarks, err := h.bookmarkRepo.GetWithoutFavicon()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"bookmarks": bookmarks})
}

// Update 更新书签的 favicon
func (h *FaviconHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var req struct {
		Favicon string `json:"favicon" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	if err := h.bookmarkRepo.UpdateFavicon(id, req.Favicon); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}
