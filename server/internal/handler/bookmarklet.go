package handler

import (
	"encoding/json"
	"net/http"

	"Nibstash_v2_server/internal/middleware"
	"Nibstash_v2_server/internal/repository"

	"github.com/gin-gonic/gin"
)

type BookmarkletHandler struct {
	bookmarkRepo *repository.BookmarkRepository
	folderRepo   *repository.FolderRepository
}

func NewBookmarkletHandler() *BookmarkletHandler {
	return &BookmarkletHandler{
		bookmarkRepo: repository.NewBookmarkRepository(),
		folderRepo:   repository.NewFolderRepository(),
	}
}

// Handle 处理 bookmarklet GET 请求，显示收藏表单
func (h *BookmarkletHandler) Handle(c *gin.Context) {
	url := c.Query("url")
	title := c.Query("title")

	// 检查是否已登录
	token, _ := c.Cookie("token")

	// 验证 token
	if token == "" {
		h.renderError(c, "请先登录囤囤鼠")
		return
	}

	_, err := middleware.ParseToken(token)
	if err != nil {
		h.renderError(c, "登录已过期，请重新登录")
		return
	}

	if url == "" {
		h.renderError(c, "URL 不能为空")
		return
	}

	if title == "" {
		title = url
	}

	// 获取文件夹列表
	folders := h.folderRepo.GetAllPaths()
	foldersJSON, _ := json.Marshal(folders)

	// 渲染表单页面
	h.renderForm(c, url, title, string(foldersJSON))
}

// Save 处理 bookmarklet POST 请求，保存书签
func (h *BookmarkletHandler) Save(c *gin.Context) {
	// 检查是否已登录
	token, _ := c.Cookie("token")
	if token == "" {
		h.renderResult(c, "error", "请先登录囤囤鼠", "")
		return
	}

	_, err := middleware.ParseToken(token)
	if err != nil {
		h.renderResult(c, "error", "登录已过期，请重新登录", "")
		return
	}

	url := c.PostForm("url")
	title := c.PostForm("title")
	description := c.PostForm("description")
	folderPath := c.PostForm("folder")

	if url == "" {
		h.renderResult(c, "error", "URL 不能为空", "")
		return
	}

	if title == "" {
		title = url
	}

	// 检查是否已存在
	if h.bookmarkRepo.Exists(url, folderPath) {
		h.renderResult(c, "warning", "该网址已被收藏", url)
		return
	}

	// 创建书签
	_, err = h.bookmarkRepo.Create(url, title, description, "", folderPath, nil)
	if err != nil {
		h.renderResult(c, "error", "收藏失败: "+err.Error(), "")
		return
	}

	h.renderResult(c, "success", "收藏成功！", title)
}

// renderError 渲染错误页面
func (h *BookmarkletHandler) renderError(c *gin.Context, message string) {
	html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>囤囤鼠 - 错误</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        .card {
            background: white;
            border-radius: 16px;
            padding: 32px;
            text-align: center;
            box-shadow: 0 10px 40px rgba(0,0,0,0.2);
            max-width: 320px;
            width: 100%;
        }
        .icon {
            width: 64px;
            height: 64px;
            border-radius: 50%;
            background: #f56c6c;
            color: white;
            font-size: 32px;
            display: flex;
            align-items: center;
            justify-content: center;
            margin: 0 auto 20px;
        }
        .message {
            font-size: 16px;
            color: #303133;
            margin-bottom: 20px;
        }
        .close-btn {
            background: #f56c6c;
            color: white;
            border: none;
            padding: 10px 24px;
            border-radius: 8px;
            font-size: 14px;
            cursor: pointer;
        }
        .close-btn:hover { opacity: 0.9; }
    </style>
</head>
<body>
    <div class="card">
        <div class="icon">✗</div>
        <div class="message">` + message + `</div>
        <button class="close-btn" onclick="window.close()">关闭窗口</button>
    </div>
</body>
</html>`

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// renderForm 渲染收藏表单页面
func (h *BookmarkletHandler) renderForm(c *gin.Context, url, title, foldersJSON string) {
	html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>囤囤鼠 - 收藏</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
        }
        .card {
            background: white;
            border-radius: 12px;
            padding: 20px;
            box-shadow: 0 10px 40px rgba(0,0,0,0.2);
            max-width: 400px;
            margin: 0 auto;
        }
        .header {
            display: flex;
            align-items: center;
            margin-bottom: 20px;
            padding-bottom: 15px;
            border-bottom: 1px solid #eee;
        }
        .header-icon {
            font-size: 24px;
            margin-right: 10px;
        }
        .header-title {
            font-size: 18px;
            font-weight: 600;
            color: #303133;
        }
        .form-group {
            margin-bottom: 16px;
        }
        .form-label {
            display: block;
            font-size: 13px;
            font-weight: 500;
            color: #606266;
            margin-bottom: 6px;
        }
        .form-input, .form-textarea, .form-select {
            width: 100%;
            padding: 10px 12px;
            border: 1px solid #dcdfe6;
            border-radius: 6px;
            font-size: 14px;
            transition: border-color 0.2s;
        }
        .form-input:focus, .form-textarea:focus, .form-select:focus {
            outline: none;
            border-color: #667eea;
        }
        .form-textarea {
            resize: vertical;
            min-height: 60px;
        }
        .folder-wrapper {
            display: flex;
            gap: 8px;
        }
        .folder-wrapper .form-select {
            flex: 1;
        }
        .new-folder-btn {
            padding: 10px 12px;
            background: #f5f7fa;
            border: 1px solid #dcdfe6;
            border-radius: 6px;
            cursor: pointer;
            font-size: 14px;
        }
        .new-folder-btn:hover {
            background: #e4e7ed;
        }
        .new-folder-input {
            display: none;
            margin-top: 8px;
        }
        .new-folder-input.show {
            display: block;
        }
        .btn-group {
            display: flex;
            gap: 10px;
            margin-top: 20px;
        }
        .btn {
            flex: 1;
            padding: 12px;
            border: none;
            border-radius: 8px;
            font-size: 14px;
            font-weight: 500;
            cursor: pointer;
            transition: opacity 0.2s;
        }
        .btn:hover { opacity: 0.9; }
        .btn-cancel {
            background: #f5f7fa;
            color: #606266;
        }
        .btn-submit {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
        }
    </style>
</head>
<body>
    <div class="card">
        <div class="header">
            <span class="header-icon">🐿️</span>
            <span class="header-title">收藏到囤囤鼠</span>
        </div>
        <form method="POST" action="/api/bookmarklet">
            <div class="form-group">
                <label class="form-label">标题</label>
                <input type="text" name="title" class="form-input" value="` + escapeHTML(title) + `" required>
            </div>
            <div class="form-group">
                <label class="form-label">网址</label>
                <input type="url" name="url" class="form-input" value="` + escapeHTML(url) + `" required>
            </div>
            <div class="form-group">
                <label class="form-label">描述（可选）</label>
                <textarea name="description" class="form-textarea" placeholder="添加一些描述..."></textarea>
            </div>
            <div class="form-group">
                <label class="form-label">文件夹</label>
                <div class="folder-wrapper">
                    <select name="folder" id="folderSelect" class="form-select">
                        <option value="新收藏">新收藏</option>
                    </select>
                    <button type="button" class="new-folder-btn" onclick="toggleNewFolder()">+</button>
                </div>
                <input type="text" id="newFolderInput" class="form-input new-folder-input" placeholder="输入新文件夹名称，按回车确认">
            </div>
            <div class="btn-group">
                <button type="button" class="btn btn-cancel" onclick="window.close()">取消</button>
                <button type="submit" class="btn btn-submit">收藏</button>
            </div>
        </form>
    </div>
    <script>
        var folders = ` + foldersJSON + `;
        var select = document.getElementById('folderSelect');

        // 添加现有文件夹选项
        folders.forEach(function(folder) {
            if (folder && folder !== '新收藏') {
                var option = document.createElement('option');
                option.value = folder;
                option.textContent = folder;
                select.appendChild(option);
            }
        });

        function toggleNewFolder() {
            var input = document.getElementById('newFolderInput');
            input.classList.toggle('show');
            if (input.classList.contains('show')) {
                input.focus();
            }
        }

        document.getElementById('newFolderInput').addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                e.preventDefault();
                var newFolder = this.value.trim();
                if (newFolder) {
                    // 检查是否已存在
                    var exists = false;
                    for (var i = 0; i < select.options.length; i++) {
                        if (select.options[i].value === newFolder) {
                            exists = true;
                            select.value = newFolder;
                            break;
                        }
                    }
                    if (!exists) {
                        var option = document.createElement('option');
                        option.value = newFolder;
                        option.textContent = newFolder;
                        select.appendChild(option);
                        select.value = newFolder;
                    }
                    this.value = '';
                    this.classList.remove('show');
                }
            }
        });
    </script>
</body>
</html>`

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// renderResult 渲染结果页面
func (h *BookmarkletHandler) renderResult(c *gin.Context, status, message, detail string) {
	var bgColor, icon string
	switch status {
	case "success":
		bgColor = "#67c23a"
		icon = "✓"
	case "warning":
		bgColor = "#e6a23c"
		icon = "!"
	case "error":
		bgColor = "#f56c6c"
		icon = "✗"
	}

	html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>囤囤鼠 - 收藏结果</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        .card {
            background: white;
            border-radius: 16px;
            padding: 32px;
            text-align: center;
            box-shadow: 0 10px 40px rgba(0,0,0,0.2);
            max-width: 320px;
            width: 100%;
        }
        .icon {
            width: 64px;
            height: 64px;
            border-radius: 50%;
            background: ` + bgColor + `;
            color: white;
            font-size: 32px;
            display: flex;
            align-items: center;
            justify-content: center;
            margin: 0 auto 20px;
        }
        .message {
            font-size: 18px;
            font-weight: 600;
            color: #303133;
            margin-bottom: 12px;
        }
        .detail {
            font-size: 14px;
            color: #909399;
            word-break: break-all;
            margin-bottom: 20px;
        }
        .close-btn {
            background: ` + bgColor + `;
            color: white;
            border: none;
            padding: 10px 24px;
            border-radius: 8px;
            font-size: 14px;
            cursor: pointer;
            transition: opacity 0.2s;
        }
        .close-btn:hover { opacity: 0.9; }
    </style>
</head>
<body>
    <div class="card">
        <div class="icon">` + icon + `</div>
        <div class="message">` + escapeHTML(message) + `</div>
        <div class="detail">` + escapeHTML(detail) + `</div>
        <button class="close-btn" onclick="window.close()">关闭窗口</button>
    </div>
    <script>
        if ('` + status + `' === 'success') {
            setTimeout(function() { window.close(); }, 2000);
        }
    </script>
</body>
</html>`

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}
