package handler

import (
	"net/http"
	"strconv"

	"Nibstash_v2_server/internal/model"
	"Nibstash_v2_server/internal/repository"

	"github.com/gin-gonic/gin"
)

type CredentialHandler struct {
	credRepo *repository.CredentialRepository
}

func NewCredentialHandler() *CredentialHandler {
	return &CredentialHandler{
		credRepo: repository.NewCredentialRepository(),
	}
}

// List 获取凭证列表
func (h *CredentialHandler) List(c *gin.Context) {
	creds, err := h.credRepo.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取凭证列表失败"})
		return
	}

	if creds == nil {
		creds = []model.Credential{}
	}

	c.JSON(http.StatusOK, creds)
}

// Create 创建凭证
func (h *CredentialHandler) Create(c *gin.Context) {
	var req model.CredentialCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	cred, err := h.credRepo.Create(req.Domain, req.Title, req.Username, req.Password, req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建凭证失败"})
		return
	}

	c.JSON(http.StatusCreated, cred)
}

// Get 获取凭证详情
func (h *CredentialHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	cred, err := h.credRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "凭证不存在"})
		return
	}

	c.JSON(http.StatusOK, cred)
}

// GetByDomain 根据域名获取凭证
func (h *CredentialHandler) GetByDomain(c *gin.Context) {
	domain := c.Param("domain")
	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "域名不能为空"})
		return
	}

	creds, err := h.credRepo.GetByDomain(domain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取凭证失败"})
		return
	}

	if creds == nil {
		creds = []model.Credential{}
	}

	c.JSON(http.StatusOK, creds)
}

// Update 更新凭证
func (h *CredentialHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var req model.CredentialUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 获取现有凭证
	existing, err := h.credRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "凭证不存在"})
		return
	}

	// 合并更新
	title := existing.Title
	username := existing.Username
	password := existing.Password
	notes := existing.Notes
	if req.Title != "" {
		title = req.Title
	}
	if req.Username != "" {
		username = req.Username
	}
	if req.Password != "" {
		password = req.Password
	}
	if req.Notes != "" {
		notes = req.Notes
	}

	if err := h.credRepo.Update(id, title, username, password, notes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新凭证失败"})
		return
	}

	cred, _ := h.credRepo.GetByID(id)
	c.JSON(http.StatusOK, cred)
}

// Delete 删除凭证
func (h *CredentialHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	if err := h.credRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除凭证失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
