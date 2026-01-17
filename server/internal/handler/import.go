package handler

import (
	"io"
	"net/http"
	"strings"

	"Nibstash_v2_server/internal/model"
	"Nibstash_v2_server/internal/repository"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/html"
)

type ImportHandler struct {
	bookmarkRepo *repository.BookmarkRepository
}

func NewImportHandler() *ImportHandler {
	return &ImportHandler{
		bookmarkRepo: repository.NewBookmarkRepository(),
	}
}

// Import 导入书签
func (h *ImportHandler) Import(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请上传文件"})
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件失败"})
		return
	}

	bookmarks := parseBookmarkHTML(string(content))
	if len(bookmarks) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未找到有效的书签"})
		return
	}

	imported, skipped, err := h.bookmarkRepo.BatchImport(bookmarks, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "导入失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "导入完成",
		"imported": imported,
		"skipped":  skipped,
		"total":    len(bookmarks),
	})
}

// parseBookmarkHTML 解析书签 HTML 文件
func parseBookmarkHTML(content string) []model.ImportBookmark {
	var bookmarks []model.ImportBookmark
	var folderStack []string

	doc, err := html.Parse(strings.NewReader(content))
	if err != nil {
		return bookmarks
	}

	var parse func(*html.Node)
	parse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "dt" {
			// 检查这个 DT 是文件夹还是书签
			var h3Node, aNode, dlNode *html.Node
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode {
					switch c.Data {
					case "h3":
						h3Node = c
					case "a":
						aNode = c
					case "dl":
						dlNode = c
					}
				}
			}

			if h3Node != nil {
				// 这是一个文件夹
				folderName := getTextContent(h3Node)
				if folderName != "" {
					folderStack = append(folderStack, folderName)
				}
				// 处理子 DL
				if dlNode != nil {
					parse(dlNode)
				}
				// 退出文件夹
				if folderName != "" && len(folderStack) > 0 {
					folderStack = folderStack[:len(folderStack)-1]
				}
				return
			}

			if aNode != nil {
				// 这是一个书签
				var url string
				for _, attr := range aNode.Attr {
					if attr.Key == "href" {
						url = attr.Val
						break
					}
				}
				title := getTextContent(aNode)
				if url != "" && title != "" && (strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")) {
					folderPath := strings.Join(folderStack, "/")
					bookmarks = append(bookmarks, model.ImportBookmark{
						URL:        url,
						Title:      title,
						FolderPath: folderPath,
					})
				}
				return
			}
		}

		// 递归处理子节点
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			parse(c)
		}
	}

	parse(doc)
	return bookmarks
}

// getTextContent 获取节点的文本内容
func getTextContent(n *html.Node) string {
	if n == nil {
		return ""
	}
	var text string
	var extract func(*html.Node)
	extract = func(node *html.Node) {
		if node.Type == html.TextNode {
			text += node.Data
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			extract(c)
		}
	}
	extract(n)
	return strings.TrimSpace(text)
}
