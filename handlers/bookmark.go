package handlers

import (
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"nibstash/models"
)

const rootFolderToken = "__root__"

type FolderNode struct {
	Name        string
	Path        string
	EncodedPath string
	PageSizeParam string
	Children    []*FolderNode
	Active      bool
	Open        bool
}

func buildFolderTree(paths []string, activePath string, pageSizeParam string) []*FolderNode {
	var roots []*FolderNode
	for _, path := range paths {
		trimmed := strings.TrimSpace(path)
		if trimmed == "" {
			continue
		}
		parts := strings.Split(trimmed, "/")
		current := &roots
		var currentPathParts []string
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			currentPathParts = append(currentPathParts, part)
			currentPath := strings.Join(currentPathParts, "/")
			node := findOrCreateFolderNode(current, part, currentPath)
			current = &node.Children
		}
	}

	sortFolderNodes(roots)
	markFolderState(roots, activePath, pageSizeParam)
	return roots
}

func findOrCreateFolderNode(nodes *[]*FolderNode, name, path string) *FolderNode {
	for _, node := range *nodes {
		if node.Name == name {
			return node
		}
	}
	node := &FolderNode{
		Name: name,
		Path: path,
	}
	*nodes = append(*nodes, node)
	return node
}

func sortFolderNodes(nodes []*FolderNode) {
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Name < nodes[j].Name
	})
	for _, node := range nodes {
		sortFolderNodes(node.Children)
	}
}

func markFolderState(nodes []*FolderNode, activePath string, pageSizeParam string) {
	for _, node := range nodes {
		node.EncodedPath = url.QueryEscape(node.Path)
		node.PageSizeParam = pageSizeParam
		node.Active = activePath != "" && node.Path == activePath
		if activePath != "" && (node.Active || strings.HasPrefix(activePath, node.Path+"/")) {
			node.Open = true
		}
		markFolderState(node.Children, activePath, pageSizeParam)
	}
}

func decodeFolderParam(param string) string {
	decoded := param
	if strings.Contains(decoded, "%") {
		if unescaped, err := url.QueryUnescape(decoded); err == nil {
			decoded = unescaped
			if strings.Contains(decoded, "%") {
				if unescapedAgain, err := url.QueryUnescape(decoded); err == nil {
					decoded = unescapedAgain
				}
			}
		}
	}
	return decoded
}

// IndexHandler 首页/书签列表
func IndexHandler(tmpl TemplateRenderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}
		pageSize := 20
		sizeParam := r.URL.Query().Get("size")
		if sizeParam != "" {
			if size, err := strconv.Atoi(sizeParam); err == nil {
				if size < 5 {
					size = 5
				} else if size > 200 {
					size = 200
				}
				pageSize = size
			}
		}

		tagIDStr := r.URL.Query().Get("tag")
		var tagID int64
		if tagIDStr != "" {
			tagID, _ = strconv.ParseInt(tagIDStr, 10, 64)
		}

		search := r.URL.Query().Get("q")
		sortBy := r.URL.Query().Get("sort")
		folderParam := decodeFolderParam(r.URL.Query().Get("folder"))
		filterFolder := folderParam != ""
		folderPath := ""
		if filterFolder {
			if folderParam == rootFolderToken {
				folderPath = ""
			} else {
				folderPath = folderParam
			}
		}

		pageSizeParam := ""
		if sizeParam != "" {
			pageSizeParam = strconv.Itoa(pageSize)
		}
		bookmarks, total, _ := models.ListBookmarks(page, pageSize, tagID, search, folderPath, filterFolder, sortBy)
		tags, _ := models.ListTags()
		folderPaths, _ := models.ListFolderPaths()
		folderTree := buildFolderTree(folderPaths, folderPath, pageSizeParam)
		hasUncategorized := models.HasUncategorizedBookmarks()

		totalPages := (total + pageSize - 1) / pageSize

		// 当前选中的标签
		var currentTag *models.Tag
		if tagID > 0 {
			currentTag, _ = models.GetTagByID(tagID)
		}

		folderParamEncoded := ""
		if filterFolder {
			folderParamEncoded = url.QueryEscape(folderParam)
		}
		currentFolderName := ""
		isRootFolder := false
		if filterFolder {
			if folderParam == rootFolderToken {
				currentFolderName = "未分类"
				isRootFolder = true
			} else if folderPath != "" {
				parts := strings.Split(folderPath, "/")
				currentFolderName = parts[len(parts)-1]
			}
		}

		tmpl.Render(w, "index.html", map[string]interface{}{
			"Bookmarks":         bookmarks,
			"Tags":              tags,
			"CurrentTag":        currentTag,
			"Search":            search,
			"SortBy":            sortBy,
			"Page":              page,
			"TotalPages":        totalPages,
			"Total":             total,
			"HasPrev":           page > 1,
			"HasNext":           page < totalPages,
			"PrevPage":          page - 1,
			"NextPage":          page + 1,
			"FolderTree":        folderTree,
			"HasUncategorized":  hasUncategorized,
			"IsAllFolders":      !filterFolder,
			"IsRootFolder":      isRootFolder,
			"FolderParam":       folderParamEncoded,
			"PageSize":          pageSize,
			"PageSizeParam":     pageSizeParam,
			"CurrentFolderName": currentFolderName,
		})
	}
}

// BookmarkAddHandler 添加书签页面
func BookmarkAddHandler(tmpl TemplateRenderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tags, _ := models.ListTags()

		// 支持从 URL 参数预填充（用于 Bookmarklet）
		url := r.URL.Query().Get("url")
		title := r.URL.Query().Get("title")

		tmpl.Render(w, "bookmark_form.html", map[string]interface{}{
			"Bookmark":       nil,
			"Tags":           tags,
			"IsEdit":         false,
			"PrefilledURL":   url,
			"PrefilledTitle": title,
			"Error":          "",
		})
	}
}

// BookmarkAddPostHandler 保存书签
func BookmarkAddPostHandler(tmpl TemplateRenderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url := strings.TrimSpace(r.FormValue("url"))
		title := strings.TrimSpace(r.FormValue("title"))
		description := strings.TrimSpace(r.FormValue("description"))
		tagIDsStr := r.Form["tags"]

		// 验证
		if url == "" || title == "" {
			tags, _ := models.ListTags()
			tmpl.Render(w, "bookmark_form.html", map[string]interface{}{
				"Bookmark":       nil,
				"Tags":           tags,
				"IsEdit":         false,
				"PrefilledURL":   url,
				"PrefilledTitle": title,
				"Error":          "URL 和标题不能为空",
			})
			return
		}

		// 检查是否已存在
		if models.BookmarkExists(url, "") {
			tags, _ := models.ListTags()
			tmpl.Render(w, "bookmark_form.html", map[string]interface{}{
				"Bookmark":       nil,
				"Tags":           tags,
				"IsEdit":         false,
				"PrefilledURL":   url,
				"PrefilledTitle": title,
				"Error":          "该 URL 已被收藏",
			})
			return
		}

		// 解析标签ID
		var tagIDs []int64
		for _, idStr := range tagIDsStr {
			if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
				tagIDs = append(tagIDs, id)
			}
		}

		// 获取 favicon
		favicon := getFaviconURL(url)

		_, err := models.CreateBookmark(url, title, description, favicon, tagIDs, "")
		if err != nil {
			tags, _ := models.ListTags()
			tmpl.Render(w, "bookmark_form.html", map[string]interface{}{
				"Bookmark":       nil,
				"Tags":           tags,
				"IsEdit":         false,
				"PrefilledURL":   url,
				"PrefilledTitle": title,
				"Error":          "保存失败: " + err.Error(),
			})
			return
		}

		http.Redirect(w, r, "/", http.StatusFound)
	}
}

// BookmarkEditHandler 编辑书签页面
func BookmarkEditHandler(tmpl TemplateRenderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.URL.Query().Get("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		bookmark, err := models.GetBookmarkByID(id)
		if err != nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		tags, _ := models.ListTags()

		// 标记已选中的标签
		selectedTagIDs := make(map[int64]bool)
		for _, t := range bookmark.Tags {
			selectedTagIDs[t.ID] = true
		}

		tmpl.Render(w, "bookmark_form.html", map[string]interface{}{
			"Bookmark":       bookmark,
			"Tags":           tags,
			"SelectedTagIDs": selectedTagIDs,
			"IsEdit":         true,
			"Error":          "",
		})
	}
}

// BookmarkEditPostHandler 更新书签
func BookmarkEditPostHandler(tmpl TemplateRenderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.FormValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		url := strings.TrimSpace(r.FormValue("url"))
		title := strings.TrimSpace(r.FormValue("title"))
		description := strings.TrimSpace(r.FormValue("description"))
		tagIDsStr := r.Form["tags"]

		// 验证
		if url == "" || title == "" {
			bookmark, _ := models.GetBookmarkByID(id)
			tags, _ := models.ListTags()
			tmpl.Render(w, "bookmark_form.html", map[string]interface{}{
				"Bookmark": bookmark,
				"Tags":     tags,
				"IsEdit":   true,
				"Error":    "URL 和标题不能为空",
			})
			return
		}

		// 解析标签ID
		var tagIDs []int64
		for _, idStr := range tagIDsStr {
			if tid, err := strconv.ParseInt(idStr, 10, 64); err == nil {
				tagIDs = append(tagIDs, tid)
			}
		}

		err = models.UpdateBookmark(id, url, title, description, tagIDs)
		if err != nil {
			bookmark, _ := models.GetBookmarkByID(id)
			tags, _ := models.ListTags()
			tmpl.Render(w, "bookmark_form.html", map[string]interface{}{
				"Bookmark": bookmark,
				"Tags":     tags,
				"IsEdit":   true,
				"Error":    "更新失败: " + err.Error(),
			})
			return
		}

		http.Redirect(w, r, "/", http.StatusFound)
	}
}

// BookmarkDeleteHandler 删除书签
func BookmarkDeleteHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	models.DeleteBookmark(id)
	http.Redirect(w, r, "/", http.StatusFound)
}

// BookmarkClearHandler 清空全部书签
func BookmarkClearHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	if err := models.DeleteAllBookmarks(); err != nil {
		http.Error(w, "清空失败", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func BookmarkClearFolderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	folderParam := decodeFolderParam(r.FormValue("folder"))
	if folderParam == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	folderPath := folderParam
	if folderParam == rootFolderToken {
		folderPath = ""
	}
	if err := models.DeleteBookmarksByFolder(folderPath); err != nil {
		http.Error(w, "删除失败", http.StatusInternalServerError)
		return
	}
	if referer := r.Referer(); referer != "" {
		http.Redirect(w, r, referer, http.StatusFound)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func BookmarkDeleteBatchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "删除失败", http.StatusBadRequest)
		return
	}
	idValues := r.Form["ids"]
	var ids []int64
	for _, idStr := range idValues {
		if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
			ids = append(ids, id)
		}
	}
	if err := models.DeleteBookmarksByIDs(ids); err != nil {
		http.Error(w, "删除失败", http.StatusInternalServerError)
		return
	}
	if referer := r.Referer(); referer != "" {
		http.Redirect(w, r, referer, http.StatusFound)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

// getFaviconURL 获取网站图标URL
func getFaviconURL(url string) string {
	// 使用 Google 的 favicon 服务
	// 提取域名
	domain := url
	if strings.HasPrefix(domain, "http://") {
		domain = domain[7:]
	} else if strings.HasPrefix(domain, "https://") {
		domain = domain[8:]
	}
	if idx := strings.Index(domain, "/"); idx > 0 {
		domain = domain[:idx]
	}
	return "https://www.google.com/s2/favicons?domain=" + domain + "&sz=32"
}

// BookmarkExportHandler 导出书签为HTML格式
func BookmarkExportHandler(w http.ResponseWriter, r *http.Request) {
	bookmarks, err := models.GetAllBookmarksForExport()
	if err != nil {
		http.Error(w, "导出失败", http.StatusInternalServerError)
		return
	}

	// 生成浏览器标准书签HTML格式
	html := generateBookmarkHTML(bookmarks)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=bookmarks_export.html")
	w.Write([]byte(html))
}

// generateBookmarkHTML 生成浏览器标准书签HTML
func generateBookmarkHTML(bookmarks []models.ExportBookmark) string {
	var sb strings.Builder
	sb.WriteString(`<!DOCTYPE NETSCAPE-Bookmark-file-1>
<!-- This is an automatically generated file.
     It will be read and overwritten.
     DO NOT EDIT! -->
<META HTTP-EQUIV="Content-Type" CONTENT="text/html; charset=UTF-8">
<TITLE>Bookmarks</TITLE>
<H1>Bookmarks</H1>
<DL><p>
`)

	// 按文件夹组织书签
	folderMap := make(map[string][]models.ExportBookmark)
	for _, bm := range bookmarks {
		folderMap[bm.FolderPath] = append(folderMap[bm.FolderPath], bm)
	}

	// 获取所有文件夹路径并排序
	var folders []string
	for folder := range folderMap {
		folders = append(folders, folder)
	}
	sort.Strings(folders)

	// 跟踪当前打开的文件夹层级
	var currentPath []string

	for _, folder := range folders {
		bms := folderMap[folder]

		if folder == "" {
			// 根目录书签
			for _, bm := range bms {
				sb.WriteString("    <DT><A HREF=\"" + escapeHTML(bm.URL) + "\">" + escapeHTML(bm.Title) + "</A>\n")
			}
		} else {
			// 处理文件夹层级
			parts := strings.Split(folder, "/")

			// 关闭不再需要的文件夹
			for len(currentPath) > 0 && (len(currentPath) > len(parts) || !hasPrefix(parts, currentPath)) {
				indent := strings.Repeat("    ", len(currentPath))
				sb.WriteString(indent + "</DL><p>\n")
				currentPath = currentPath[:len(currentPath)-1]
			}

			// 打开新文件夹
			for i := len(currentPath); i < len(parts); i++ {
				indent := strings.Repeat("    ", i+1)
				sb.WriteString(indent + "<DT><H3>" + escapeHTML(parts[i]) + "</H3>\n")
				sb.WriteString(indent + "<DL><p>\n")
				currentPath = append(currentPath, parts[i])
			}

			// 写入书签
			indent := strings.Repeat("    ", len(parts)+1)
			for _, bm := range bms {
				sb.WriteString(indent + "<DT><A HREF=\"" + escapeHTML(bm.URL) + "\">" + escapeHTML(bm.Title) + "</A>\n")
			}
		}
	}

	// 关闭所有打开的文件夹
	for len(currentPath) > 0 {
		indent := strings.Repeat("    ", len(currentPath))
		sb.WriteString(indent + "</DL><p>\n")
		currentPath = currentPath[:len(currentPath)-1]
	}

	sb.WriteString("</DL><p>\n")
	return sb.String()
}

func hasPrefix(full, prefix []string) bool {
	if len(prefix) > len(full) {
		return false
	}
	for i, p := range prefix {
		if full[i] != p {
			return false
		}
	}
	return true
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}

// BookmarkMoveHandler 批量移动书签到指定文件夹
func BookmarkMoveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "移动失败", http.StatusBadRequest)
		return
	}

	idValues := r.Form["ids"]
	targetFolder := r.FormValue("target_folder")

	var ids []int64
	for _, idStr := range idValues {
		if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
			ids = append(ids, id)
		}
	}

	if err := models.MoveBookmarksToFolder(ids, targetFolder); err != nil {
		http.Error(w, "移动失败", http.StatusInternalServerError)
		return
	}

	if referer := r.Referer(); referer != "" {
		http.Redirect(w, r, referer, http.StatusFound)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

// APIFaviconPendingHandler 获取缺少favicon的书签列表
func APIFaviconPendingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 使用 session 验证
		if !ValidateSession(r) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"unauthorized"}`))
			return
		}

		bookmarks, err := models.GetBookmarksWithoutFavicon()
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"` + err.Error() + `"}`))
			return
		}

		// 构建JSON响应
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"total":` + strconv.Itoa(len(bookmarks)) + `,"bookmarks":[`))
		for i, bm := range bookmarks {
			if i > 0 {
				w.Write([]byte(","))
			}
			w.Write([]byte(`{"id":` + strconv.FormatInt(bm.ID, 10) + `,"url":"` + escapeJSONString(bm.URL) + `"}`))
		}
		w.Write([]byte(`]}`))
	}
}

// APIFaviconUpdateHandler 更新单个书签的favicon
func APIFaviconUpdateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error":"method not allowed"}`))
			return
		}

		// 使用 session 验证
		if !ValidateSession(r) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"unauthorized"}`))
			return
		}

		idStr := r.FormValue("id")
		favicon := r.FormValue("favicon")

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"invalid id"}`))
			return
		}

		if err := models.UpdateBookmarkFavicon(id, favicon); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"` + err.Error() + `"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"success":true}`))
	}
}

func escapeJSONString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}

// APIMoveFolderHandler 移动文件夹到另一个文件夹
func APIMoveFolderHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error":"method not allowed"}`))
			return
		}

		// 使用 session 验证
		if !ValidateSession(r) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"unauthorized"}`))
			return
		}

		sourceFolder := r.FormValue("source")
		targetFolder := r.FormValue("target")

		// 解码路径
		sourceFolder = decodeFolderParam(sourceFolder)
		targetFolder = decodeFolderParam(targetFolder)

		// __root__ 表示根目录
		if targetFolder == rootFolderToken {
			targetFolder = ""
		}

		if err := models.MoveFolderToFolder(sourceFolder, targetFolder); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"` + escapeJSONString(err.Error()) + `"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"success":true}`))
	}
}

// APIMergeFolderHandler 合并文件夹到另一个文件夹
func APIMergeFolderHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error":"method not allowed"}`))
			return
		}

		// 使用 session 验证
		if !ValidateSession(r) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"unauthorized"}`))
			return
		}

		sourceFolder := r.FormValue("source")
		targetFolder := r.FormValue("target")

		// 解码路径
		sourceFolder = decodeFolderParam(sourceFolder)
		targetFolder = decodeFolderParam(targetFolder)

		// __root__ 表示根目录
		if targetFolder == rootFolderToken {
			targetFolder = ""
		}

		if err := models.MergeFolderToFolder(sourceFolder, targetFolder); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"` + escapeJSONString(err.Error()) + `"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"success":true}`))
	}
}

// APIQuickAddBookmarkHandler 快速添加书签
func APIQuickAddBookmarkHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error":"method not allowed"}`))
			return
		}

		if !ValidateSession(r) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"unauthorized"}`))
			return
		}

		bookmarkURL := r.FormValue("url")
		title := r.FormValue("title")
		folderPath := r.FormValue("folder")

		if bookmarkURL == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"URL不能为空"}`))
			return
		}

		// 解码文件夹路径
		folderPath = decodeFolderParam(folderPath)
		if folderPath == rootFolderToken {
			folderPath = ""
		}

		// 检查是否已存在
		if models.BookmarkExists(bookmarkURL, folderPath) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"该URL在此文件夹已存在"}`))
			return
		}

		// 获取 favicon
		favicon := getFaviconURL(bookmarkURL)

		_, err := models.QuickCreateBookmark(bookmarkURL, title, favicon, folderPath)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"` + escapeJSONString(err.Error()) + `"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"success":true}`))
	}
}

// APICreateFolderHandler 创建文件夹（通过创建占位书签）
func APICreateFolderHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error":"method not allowed"}`))
			return
		}

		if !ValidateSession(r) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"unauthorized"}`))
			return
		}

		folderName := r.FormValue("name")
		parentFolder := r.FormValue("parent")

		if folderName == "" {
			folderName = "新建文件夹"
		}

		// 解码父文件夹路径
		parentFolder = decodeFolderParam(parentFolder)
		if parentFolder == rootFolderToken {
			parentFolder = ""
		}

		// 构建完整路径
		var fullPath string
		if parentFolder == "" {
			fullPath = folderName
		} else {
			fullPath = parentFolder + "/" + folderName
		}

		// 创建一个占位书签来"创建"文件夹
		// 使用特殊的占位URL
		placeholderURL := "nibstash://folder-placeholder/" + fullPath
		_, err := models.QuickCreateBookmark(placeholderURL, "[文件夹占位]", "", fullPath)
		if err != nil {
			// 如果已存在，说明文件夹已存在
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"success":true,"message":"文件夹已存在"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"success":true}`))
	}
}
