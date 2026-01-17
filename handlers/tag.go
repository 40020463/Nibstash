package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"nibstash/models"
)

// TagsHandler 标签管理页面
func TagsHandler(tmpl TemplateRenderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tags, _ := models.ListTags()
		tmpl.Render(w, "tags.html", map[string]interface{}{
			"Tags":  tags,
			"Error": "",
		})
	}
}

// TagAddHandler 添加标签
func TagAddHandler(tmpl TemplateRenderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimSpace(r.FormValue("name"))
		color := strings.TrimSpace(r.FormValue("color"))

		if name == "" {
			tags, _ := models.ListTags()
			tmpl.Render(w, "tags.html", map[string]interface{}{
				"Tags":  tags,
				"Error": "标签名不能为空",
			})
			return
		}

		if color == "" {
			color = "#3b82f6"
		}

		_, err := models.CreateTag(name, color)
		if err != nil {
			tags, _ := models.ListTags()
			tmpl.Render(w, "tags.html", map[string]interface{}{
				"Tags":  tags,
				"Error": "创建失败，标签可能已存在",
			})
			return
		}

		http.Redirect(w, r, "/tags", http.StatusFound)
	}
}

// TagDeleteHandler 删除标签
func TagDeleteHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Redirect(w, r, "/tags", http.StatusFound)
		return
	}

	models.DeleteTag(id)
	http.Redirect(w, r, "/tags", http.StatusFound)
}

// BookmarkletHandler Bookmarklet 页面
func BookmarkletHandler(baseURL string, tmpl TemplateRenderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.Render(w, "bookmarklet.html", map[string]interface{}{
			"BaseURL": baseURL,
		})
	}
}

// APIAddBookmarkHandler Bookmarklet API (支持 JSONP)
func APIAddBookmarkHandler(password string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 设置 CORS 头
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			return
		}

		// 验证会话（如果需要密码）
		if password != "" && !ValidateSession(r) {
			callback := r.URL.Query().Get("callback")
			response := map[string]interface{}{
				"success": false,
				"error":   "未登录，请先登录",
			}
			sendJSONP(w, callback, response)
			return
		}

		url := r.URL.Query().Get("url")
		title := r.URL.Query().Get("title")

		if url == "" {
			callback := r.URL.Query().Get("callback")
			response := map[string]interface{}{
				"success": false,
				"error":   "URL 不能为空",
			}
			sendJSONP(w, callback, response)
			return
		}

		if title == "" {
			title = url
		}

		// 检查是否已存在
		if models.BookmarkExists(url, "") {
			callback := r.URL.Query().Get("callback")
			response := map[string]interface{}{
				"success": false,
				"error":   "该 URL 已被收藏",
			}
			sendJSONP(w, callback, response)
			return
		}

		favicon := getFaviconURL(url)
	_, err := models.CreateBookmark(url, title, "", favicon, nil, "")

		callback := r.URL.Query().Get("callback")
		if err != nil {
			response := map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			}
			sendJSONP(w, callback, response)
			return
		}

		response := map[string]interface{}{
			"success": true,
			"message": "收藏成功",
		}
		sendJSONP(w, callback, response)
	}
}

func sendJSONP(w http.ResponseWriter, callback string, data interface{}) {
	jsonData, _ := json.Marshal(data)
	if callback != "" {
		w.Header().Set("Content-Type", "application/javascript")
		w.Write([]byte(callback + "("))
		w.Write(jsonData)
		w.Write([]byte(");"))
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	}
}
