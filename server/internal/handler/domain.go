package handler

import (
	"net/http"

	"Nibstash_v2_server/internal/repository"

	"github.com/gin-gonic/gin"
)

type DomainHandler struct {
	domainRepo     *repository.DomainRepository
	credentialRepo *repository.CredentialRepository
}

func NewDomainHandler() *DomainHandler {
	return &DomainHandler{
		domainRepo:     repository.NewDomainRepository(),
		credentialRepo: repository.NewCredentialRepository(),
	}
}

// List 获取域名列表（实时计算，解决刷新问题）
func (h *DomainHandler) List(c *gin.Context) {
	domains, err := h.domainRepo.GetAllDomains()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取域名列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"domains": domains})
}

// GetBookmarks 获取指定域名的书签
func (h *DomainHandler) GetBookmarks(c *gin.Context) {
	domain := c.Param("domain")
	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "域名不能为空"})
		return
	}

	bookmarks, err := h.domainRepo.GetBookmarksByDomain(domain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取书签失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"bookmarks": bookmarks})
}

// Delete 删除域名记录和凭证数据（不删除书签）
func (h *DomainHandler) Delete(c *gin.Context) {
	domain := c.Param("domain")
	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "域名不能为空"})
		return
	}

	// 删除该域名的凭证
	if err := h.credentialRepo.DeleteByDomain(domain); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除凭证失败"})
		return
	}

	// 删除域名记录
	if err := h.domainRepo.DeleteDomain(domain); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除域名失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
