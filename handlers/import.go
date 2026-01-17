package handlers

import (
	"io"
	"net/http"
	"regexp"
	"strings"
	"nibstash/models"
)

// ImportHandler 导入页面
func ImportHandler(tmpl TemplateRenderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.Render(w, "import.html", map[string]interface{}{
			"Error":   "",
			"Success": "",
		})
	}
}

// ImportPostHandler 处理导入
func ImportPostHandler(tmpl TemplateRenderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 限制上传大小 100MB
		r.ParseMultipartForm(100 << 20)

		file, _, err := r.FormFile("bookmarks_file")
		if err != nil {
			tmpl.Render(w, "import.html", map[string]interface{}{
				"Error":   "请选择文件",
				"Success": "",
			})
			return
		}
		defer file.Close()

		// 读取文件内容
		content, err := io.ReadAll(file)
		if err != nil {
			tmpl.Render(w, "import.html", map[string]interface{}{
				"Error":   "读取文件失败",
				"Success": "",
			})
			return
		}

		skipDuplicates := r.FormValue("skip_duplicates") == "1"

		// 解析书签
		parsedBookmarks := parseBookmarksHTML(string(content))

		// 转换为导入格式（不获取favicon，后续异步补充）
		importBookmarks := make([]models.ImportBookmark, 0, len(parsedBookmarks))
		for _, bm := range parsedBookmarks {
			if bm.URL != "" && bm.Title != "" {
				importBookmarks = append(importBookmarks, models.ImportBookmark{
					URL:        bm.URL,
					Title:      bm.Title,
					FolderPath: bm.FolderPath,
					Favicon:    "", // 暂不获取favicon
				})
			}
		}

		// 批量导入
		imported, skipped, err := models.BatchImportBookmarks(importBookmarks, skipDuplicates)
		if err != nil {
			tmpl.Render(w, "import.html", map[string]interface{}{
				"Error":   "导入失败: " + err.Error(),
				"Success": "",
			})
			return
		}

		tmpl.Render(w, "import.html", map[string]interface{}{
			"Error":   "",
			"Success": formatImportResult(imported, skipped),
		})
	}
}

type parsedBookmark struct {
	URL   string
	Title string
	FolderPath string
}

// parseBookmarksHTML 解析浏览器导出的书签 HTML
func parseBookmarksHTML(html string) []parsedBookmark {
	var bookmarks []parsedBookmark

	var folderStack []string

	// 匹配文件夹、书签与层级结束标记
	tokenRe := regexp.MustCompile(`(?is)<H3[^>]*>([^<]+)</H3>|</DL>|<A\s+[^>]*HREF="([^"]+)"[^>]*>([^<]+)</A>`)
	matches := tokenRe.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		switch {
		case match[1] != "":
			folder := strings.TrimSpace(match[1])
			if folder != "" {
				folderStack = append(folderStack, folder)
			}
		case match[2] != "":
			url := strings.TrimSpace(match[2])
			title := strings.TrimSpace(match[3])

			// 过滤非 http/https 链接
			if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
				bookmarks = append(bookmarks, parsedBookmark{
					URL:        url,
					Title:      title,
					FolderPath: strings.Join(folderStack, "/"),
				})
			}
		default:
			if len(folderStack) > 0 {
				folderStack = folderStack[:len(folderStack)-1]
			}
		}
	}

	return bookmarks
}

func formatImportResult(imported, skipped int) string {
	result := ""
	if imported > 0 {
		result += "成功导入 " + itoa(imported) + " 个书签"
	}
	if skipped > 0 {
		if result != "" {
			result += "，"
		}
		result += "跳过 " + itoa(skipped) + " 个重复书签"
	}
	if result == "" {
		result = "没有找到可导入的书签"
	}
	return result
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var result []byte
	for n > 0 {
		result = append([]byte{byte('0' + n%10)}, result...)
		n /= 10
	}
	return string(result)
}
