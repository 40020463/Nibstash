package handler

import (
	"net/http"

	"Nibstash_v2_server/internal/model"
	"Nibstash_v2_server/internal/repository"

	"github.com/gin-gonic/gin"
)

type FolderHandler struct {
	folderRepo *repository.FolderRepository
}

func NewFolderHandler() *FolderHandler {
	return &FolderHandler{
		folderRepo: repository.NewFolderRepository(),
	}
}

// List 获取文件夹树
func (h *FolderHandler) List(c *gin.Context) {
	tree, err := h.folderRepo.GetFolderTree()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取文件夹列表失败"})
		return
	}

	if tree == nil {
		tree = []model.FolderNode{}
	}

	// 添加未分类节点
	hasUncategorized := h.folderRepo.HasUncategorized()

	c.JSON(http.StatusOK, gin.H{
		"folders":          tree,
		"has_uncategorized": hasUncategorized,
	})
}

// Create 创建文件夹
func (h *FolderHandler) Create(c *gin.Context) {
	var req model.FolderCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	if err := h.folderRepo.Create(req.Path); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建文件夹失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "创建成功", "path": req.Path})
}

// Move 移动文件夹
func (h *FolderHandler) Move(c *gin.Context) {
	var req model.FolderMoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	if err := h.folderRepo.Move(req.SourcePath, req.TargetPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "移动文件夹失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "移动成功"})
}

// Merge 合并文件夹
func (h *FolderHandler) Merge(c *gin.Context) {
	var req model.FolderMergeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	if err := h.folderRepo.Merge(req.SourcePath, req.TargetPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "合并文件夹失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "合并成功"})
}

// Delete 删除文件夹
func (h *FolderHandler) Delete(c *gin.Context) {
	path := c.Query("path")

	if err := h.folderRepo.Delete(path); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除文件夹失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
