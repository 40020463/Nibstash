package handlers

import (
	"encoding/json"
	"net/http"
	"nibstash/models"
	"strconv"
)

// APIDomainListHandler 获取域名列表（使用缓存）
func APIDomainListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// 检查是否需要刷新缓存
		refresh := r.URL.Query().Get("refresh") == "1"
		if refresh || models.IsDomainCacheEmpty() {
			models.RefreshDomainCache()
		}

		groups, err := models.GetAllDomainsFromCache()
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"domains": groups,
		})
	}
}

// APIDomainRefreshHandler 刷新域名缓存
func APIDomainRefreshHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		err := models.RefreshDomainCache()
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})
	}
}

// APIDomainCredentialHandler 获取域名凭证（支持多账号）
func APIDomainCredentialHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		domain := r.URL.Query().Get("domain")
		if domain == "" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "域名不能为空",
			})
			return
		}

		// 获取所有凭证
		creds, err := models.GetCredentialsByDomain(domain)
		if err != nil {
			creds = []models.DomainCredential{}
		}

		// 检查是否需要加载书签
		noBookmarks := r.URL.Query().Get("no_bookmarks") == "1"
		var bookmarks []models.Bookmark
		if !noBookmarks {
			bookmarks, _ = models.GetBookmarksByDomain(domain)
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":     true,
			"credentials": creds,
			"bookmarks":   bookmarks,
		})
	}
}

// APIDomainCredentialSaveHandler 保存域名凭证（新增）
func APIDomainCredentialSaveHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != "POST" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "方法不允许",
			})
			return
		}

		domain := r.FormValue("domain")
		title := r.FormValue("title")
		username := r.FormValue("username")
		password := r.FormValue("password")
		notes := r.FormValue("notes")

		if domain == "" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "域名不能为空",
			})
			return
		}

		err := models.SaveCredential(domain, title, username, password, notes)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})
	}
}

// APIDomainCredentialUpdateHandler 更新域名凭证
func APIDomainCredentialUpdateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != "POST" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "方法不允许",
			})
			return
		}

		idStr := r.FormValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "无效的凭证ID",
			})
			return
		}

		title := r.FormValue("title")
		username := r.FormValue("username")
		password := r.FormValue("password")
		notes := r.FormValue("notes")

		err = models.UpdateCredential(id, title, username, password, notes)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})
	}
}

// APIDomainCredentialDeleteHandler 删除域名凭证
func APIDomainCredentialDeleteHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != "POST" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "方法不允许",
			})
			return
		}

		idStr := r.FormValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "无效的凭证ID",
			})
			return
		}

		err = models.DeleteCredential(id)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})
	}
}

// APIDomainDeleteHandler 删除域名数据（凭证和缓存，不影响收藏夹）
func APIDomainDeleteHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != "POST" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "方法不允许",
			})
			return
		}

		domain := r.FormValue("domain")
		if domain == "" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "域名不能为空",
			})
			return
		}

		err := models.DeleteDomainData(domain)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})
	}
}
