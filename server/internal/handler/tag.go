package handler

import (
	"net/http"
	"strconv"

	"Nibstash_v2_server/internal/model"
	"Nibstash_v2_server/internal/repository"

	"github.com/gin-gonic/gin"
)

type TagHandler struct {
	tagRepo *repository.TagRepository
}

func NewTagHandler() *TagHandler {
	return &TagHandler{
		tagRepo: repository.NewTagRepository(),
	}
}

// List 获取标签列表
func (h *TagHandler) List(c *gin.Context) {
	tags, err := h.tagRepo.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取标签列表失败"})
		return
	}

	if tags == nil {
		tags = []model.Tag{}
	}

	c.JSON(http.StatusOK, tags)
}

// Create 创建标签
func (h *TagHandler) Create(c *gin.Context) {
	var req model.TagCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 检查是否已存在
	existing, _ := h.tagRepo.GetByName(req.Name)
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "标签已存在"})
		return
	}

	tag, err := h.tagRepo.Create(req.Name, req.Color)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建标签失败"})
		return
	}

	c.JSON(http.StatusCreated, tag)
}

// Update 更新标签
func (h *TagHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var req model.TagUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 获取现有标签
	existing, err := h.tagRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "标签不存在"})
		return
	}

	// 合并更新
	name := existing.Name
	color := existing.Color
	if req.Name != "" {
		name = req.Name
	}
	if req.Color != "" {
		color = req.Color
	}

	if err := h.tagRepo.Update(id, name, color); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新标签失败"})
		return
	}

	tag, _ := h.tagRepo.GetByID(id)
	c.JSON(http.StatusOK, tag)
}

// Delete 删除标签
func (h *TagHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	if err := h.tagRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除标签失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
