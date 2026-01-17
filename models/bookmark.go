package models

import (
	"strings"
	"time"
)

type Bookmark struct {
	ID          int64     `json:"id"`
	URL         string    `json:"url"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	FolderPath  string    `json:"folder_path"`
	Favicon     string    `json:"favicon"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Tags        []Tag     `json:"tags"`
}

// CreateBookmark 创建书签
func CreateBookmark(url, title, description, favicon string, tagIDs []int64, folderPath string) (*Bookmark, error) {
	result, err := DB.Exec(`
		INSERT INTO bookmarks (url, title, description, folder_path, favicon)
		VALUES (?, ?, ?, ?, ?)
	`, url, title, description, folderPath, favicon)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// 关联标签
	for _, tagID := range tagIDs {
		DB.Exec(`INSERT OR IGNORE INTO bookmark_tags (bookmark_id, tag_id) VALUES (?, ?)`, id, tagID)
	}

	return GetBookmarkByID(id)
}

// GetBookmarkByID 根据ID获取书签
func GetBookmarkByID(id int64) (*Bookmark, error) {
	bookmark := &Bookmark{}
	err := DB.QueryRow(`
		SELECT id, url, title, description, folder_path, favicon, created_at, updated_at
		FROM bookmarks WHERE id = ?
	`, id).Scan(&bookmark.ID, &bookmark.URL, &bookmark.Title, &bookmark.Description,
		&bookmark.FolderPath, &bookmark.Favicon, &bookmark.CreatedAt, &bookmark.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// 获取关联的标签
	bookmark.Tags, _ = GetTagsByBookmarkID(id)
	return bookmark, nil
}

// GetBookmarkByURL 根据URL获取书签
func GetBookmarkByURL(url string) (*Bookmark, error) {
	bookmark := &Bookmark{}
	err := DB.QueryRow(`
		SELECT id, url, title, description, folder_path, favicon, created_at, updated_at
		FROM bookmarks WHERE url = ?
	`, url).Scan(&bookmark.ID, &bookmark.URL, &bookmark.Title, &bookmark.Description,
		&bookmark.FolderPath, &bookmark.Favicon, &bookmark.CreatedAt, &bookmark.UpdatedAt)
	if err != nil {
		return nil, err
	}

	bookmark.Tags, _ = GetTagsByBookmarkID(bookmark.ID)
	return bookmark, nil
}

// UpdateBookmark 更新书签
func UpdateBookmark(id int64, url, title, description string, tagIDs []int64) error {
	_, err := DB.Exec(`
		UPDATE bookmarks SET url = ?, title = ?, description = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, url, title, description, id)
	if err != nil {
		return err
	}

	// 更新标签关联
	DB.Exec(`DELETE FROM bookmark_tags WHERE bookmark_id = ?`, id)
	for _, tagID := range tagIDs {
		DB.Exec(`INSERT OR IGNORE INTO bookmark_tags (bookmark_id, tag_id) VALUES (?, ?)`, id, tagID)
	}

	return nil
}

// DeleteBookmark 删除书签
func DeleteBookmark(id int64) error {
	_, err := DB.Exec(`DELETE FROM bookmarks WHERE id = ?`, id)
	return err
}

// ListBookmarks 获取书签列表
func ListBookmarks(page, pageSize int, tagID int64, search string, folderPath string, filterFolder bool, sortBy string) ([]Bookmark, int, error) {
	offset := (page - 1) * pageSize

	var args []interface{}
	whereClause := "1=1 AND b.url NOT LIKE 'nibstash://folder-placeholder/%'"

	if tagID > 0 {
		whereClause += " AND b.id IN (SELECT bookmark_id FROM bookmark_tags WHERE tag_id = ?)"
		args = append(args, tagID)
	}

	if search != "" {
		whereClause += " AND (b.title LIKE ? OR b.url LIKE ? OR b.description LIKE ?)"
		searchPattern := "%" + search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern)
	}

	if filterFolder {
		if folderPath == "" {
			whereClause += " AND b.folder_path = ?"
			args = append(args, "")
		} else {
			whereClause += " AND (b.folder_path = ? OR b.folder_path LIKE ?)"
			args = append(args, folderPath, folderPath+"/%")
		}
	}

	// 获取总数
	var total int
	countQuery := "SELECT COUNT(*) FROM bookmarks b WHERE " + whereClause
	DB.QueryRow(countQuery, args...).Scan(&total)

	// 排序
	orderClause := "b.created_at DESC" // 默认按创建时间倒序
	switch sortBy {
	case "title_asc":
		orderClause = "b.title ASC"
	case "title_desc":
		orderClause = "b.title DESC"
	case "time_asc":
		orderClause = "b.created_at ASC"
	case "time_desc":
		orderClause = "b.created_at DESC"
	case "url_asc":
		orderClause = "b.url ASC"
	case "url_desc":
		orderClause = "b.url DESC"
	}

	// 获取列表
	query := `
		SELECT b.id, b.url, b.title, b.description, b.folder_path, b.favicon, b.created_at, b.updated_at
		FROM bookmarks b
		WHERE ` + whereClause + `
		ORDER BY ` + orderClause + `
		LIMIT ? OFFSET ?
	`
	args = append(args, pageSize, offset)

	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var bookmarks []Bookmark
	for rows.Next() {
		var b Bookmark
		err := rows.Scan(&b.ID, &b.URL, &b.Title, &b.Description, &b.FolderPath, &b.Favicon, &b.CreatedAt, &b.UpdatedAt)
		if err != nil {
			continue
		}
		b.Tags, _ = GetTagsByBookmarkID(b.ID)
		bookmarks = append(bookmarks, b)
	}

	return bookmarks, total, nil
}

// GetTagsByBookmarkID 获取书签的标签
func GetTagsByBookmarkID(bookmarkID int64) ([]Tag, error) {
	rows, err := DB.Query(`
		SELECT t.id, t.name, t.color
		FROM tags t
		JOIN bookmark_tags bt ON t.id = bt.tag_id
		WHERE bt.bookmark_id = ?
	`, bookmarkID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Color); err == nil {
			tags = append(tags, t)
		}
	}
	return tags, nil
}

// BookmarkExists 检查URL在指定文件夹是否已存在
func BookmarkExists(url string, folderPath string) bool {
	var count int
	DB.QueryRow(`SELECT COUNT(*) FROM bookmarks WHERE url = ? AND folder_path = ?`, url, folderPath).Scan(&count)
	return count > 0
}

func DeleteAllBookmarks() error {
	_, err := DB.Exec(`DELETE FROM bookmarks`)
	return err
}

func DeleteBookmarksByIDs(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	placeholders := strings.Repeat("?,", len(ids))
	placeholders = placeholders[:len(placeholders)-1]
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	_, err := DB.Exec(`DELETE FROM bookmarks WHERE id IN (`+placeholders+`)`, args...)
	return err
}

func DeleteBookmarksByFolder(folderPath string) error {
	if folderPath == "" {
		_, err := DB.Exec(`DELETE FROM bookmarks WHERE folder_path = ''`)
		return err
	}
	_, err := DB.Exec(`DELETE FROM bookmarks WHERE folder_path = ? OR folder_path LIKE ?`, folderPath, folderPath+"/%")
	return err
}

func ListFolderPaths() ([]string, error) {
	rows, err := DB.Query(`SELECT DISTINCT folder_path FROM bookmarks WHERE folder_path != '' ORDER BY folder_path`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var paths []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err == nil {
			paths = append(paths, path)
		}
	}
	return paths, nil
}

func HasUncategorizedBookmarks() bool {
	var count int
	DB.QueryRow(`SELECT COUNT(*) FROM bookmarks WHERE folder_path = ''`).Scan(&count)
	return count > 0
}

// ParseTagIDs 解析标签ID字符串
func ParseTagIDs(tagIDsStr string) []int64 {
	var tagIDs []int64
	if tagIDsStr == "" {
		return tagIDs
	}
	parts := strings.Split(tagIDsStr, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			var id int64
			for _, c := range p {
				if c >= '0' && c <= '9' {
					id = id*10 + int64(c-'0')
				}
			}
			if id > 0 {
				tagIDs = append(tagIDs, id)
			}
		}
	}
	return tagIDs
}

// ImportBookmark 导入单个书签的数据结构
type ImportBookmark struct {
	URL        string
	Title      string
	FolderPath string
	Favicon    string
}

// BatchImportBookmarks 批量导入书签（使用事务）
func BatchImportBookmarks(bookmarks []ImportBookmark, skipDuplicates bool) (imported, skipped int, err error) {
	tx, err := DB.Begin()
	if err != nil {
		return 0, 0, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	stmt, err := tx.Prepare(`INSERT OR IGNORE INTO bookmarks (url, title, description, folder_path, favicon) VALUES (?, ?, '', ?, ?)`)
	if err != nil {
		return 0, 0, err
	}
	defer stmt.Close()

	checkStmt, err := tx.Prepare(`SELECT COUNT(*) FROM bookmarks WHERE url = ? AND folder_path = ?`)
	if err != nil {
		return 0, 0, err
	}
	defer checkStmt.Close()

	for _, bm := range bookmarks {
		if bm.URL == "" || bm.Title == "" {
			continue
		}

		// 检查是否已存在
		var count int
		checkStmt.QueryRow(bm.URL, bm.FolderPath).Scan(&count)
		if count > 0 {
			if skipDuplicates {
				skipped++
				continue
			}
		}

		result, err := stmt.Exec(bm.URL, bm.Title, bm.FolderPath, bm.Favicon)
		if err == nil {
			if rows, _ := result.RowsAffected(); rows > 0 {
				imported++
			}
		}
	}

	err = tx.Commit()
	return imported, skipped, err
}

// MoveBookmarksToFolder 批量移动书签到指定文件夹
func MoveBookmarksToFolder(ids []int64, targetFolder string) error {
	if len(ids) == 0 {
		return nil
	}
	placeholders := strings.Repeat("?,", len(ids))
	placeholders = placeholders[:len(placeholders)-1]
	args := make([]interface{}, len(ids)+1)
	args[0] = targetFolder
	for i, id := range ids {
		args[i+1] = id
	}
	_, err := DB.Exec(`UPDATE bookmarks SET folder_path = ?, updated_at = CURRENT_TIMESTAMP WHERE id IN (`+placeholders+`)`, args...)
	return err
}

// ExportBookmark 导出用的书签结构
type ExportBookmark struct {
	URL        string
	Title      string
	FolderPath string
}

// GetAllBookmarksForExport 获取所有书签用于导出
func GetAllBookmarksForExport() ([]ExportBookmark, error) {
	rows, err := DB.Query(`SELECT url, title, folder_path FROM bookmarks WHERE url NOT LIKE 'nibstash://folder-placeholder/%' ORDER BY folder_path, title`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookmarks []ExportBookmark
	for rows.Next() {
		var bm ExportBookmark
		if err := rows.Scan(&bm.URL, &bm.Title, &bm.FolderPath); err == nil {
			bookmarks = append(bookmarks, bm)
		}
	}
	return bookmarks, nil
}

// GetBookmarksWithoutFavicon 获取没有favicon的书签ID和URL
func GetBookmarksWithoutFavicon() ([]struct{ ID int64; URL string }, error) {
	rows, err := DB.Query(`SELECT id, url FROM bookmarks WHERE (favicon = '' OR favicon IS NULL) AND url NOT LIKE 'nibstash://folder-placeholder/%'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []struct{ ID int64; URL string }
	for rows.Next() {
		var item struct{ ID int64; URL string }
		if err := rows.Scan(&item.ID, &item.URL); err == nil {
			result = append(result, item)
		}
	}
	return result, nil
}

// UpdateBookmarkFavicon 更新书签的favicon
func UpdateBookmarkFavicon(id int64, favicon string) error {
	_, err := DB.Exec(`UPDATE bookmarks SET favicon = ? WHERE id = ?`, favicon, id)
	return err
}

// MoveFolderToFolder 移动文件夹到另一个文件夹
// sourceFolder: 源文件夹路径 (如 "A/B")
// targetFolder: 目标文件夹路径 (如 "C/D"，空字符串表示根目录)
// 结果: "A/B" 下的内容会移动到 "C/D/B" 下
func MoveFolderToFolder(sourceFolder, targetFolder string) error {
	if sourceFolder == "" {
		return nil // 不能移动根目录
	}

	// 获取源文件夹名称
	parts := strings.Split(sourceFolder, "/")
	folderName := parts[len(parts)-1]

	// 计算新的文件夹路径
	var newFolderBase string
	if targetFolder == "" {
		newFolderBase = folderName
	} else {
		newFolderBase = targetFolder + "/" + folderName
	}

	// 如果源和目标相同，不需要移动
	if sourceFolder == newFolderBase {
		return nil
	}

	// 防止移动到自己的子文件夹
	if strings.HasPrefix(newFolderBase, sourceFolder+"/") {
		return nil
	}

	// 更新精确匹配源文件夹的书签（使用 OR IGNORE 跳过冲突）
	_, err := DB.Exec(`UPDATE OR IGNORE bookmarks SET folder_path = ?, updated_at = CURRENT_TIMESTAMP WHERE folder_path = ?`,
		newFolderBase, sourceFolder)
	if err != nil {
		return err
	}

	// 更新源文件夹子目录下的书签
	// 使用 REPLACE 函数直接在 SQL 中替换路径前缀
	oldPrefix := sourceFolder + "/"
	newPrefix := newFolderBase + "/"

	_, err = DB.Exec(`UPDATE OR IGNORE bookmarks SET folder_path = ? || SUBSTR(folder_path, ?), updated_at = CURRENT_TIMESTAMP WHERE folder_path LIKE ?`,
		newPrefix, len(oldPrefix)+1, oldPrefix+"%")

	return err
}

// MergeFolderToFolder 合并文件夹到另一个文件夹
// sourceFolder: 源文件夹路径 (如 "A/B")
// targetFolder: 目标文件夹路径 (如 "C/D"，空字符串表示根目录)
// 结果: "A/B" 下的内容直接移动到 "C/D" 下（不保留 B 文件夹名）
// 如果有冲突（同URL同文件夹），删除源文件夹中的重复书签
func MergeFolderToFolder(sourceFolder, targetFolder string) error {
	if sourceFolder == "" {
		return nil // 不能合并根目录
	}

	// 如果源和目标相同，不需要合并
	if sourceFolder == targetFolder {
		return nil
	}

	// 防止合并到自己的子文件夹
	if targetFolder != "" && strings.HasPrefix(targetFolder, sourceFolder+"/") {
		return nil
	}

	// 先尝试更新，冲突的会被跳过
	_, err := DB.Exec(`UPDATE OR IGNORE bookmarks SET folder_path = ?, updated_at = CURRENT_TIMESTAMP WHERE folder_path = ?`,
		targetFolder, sourceFolder)
	if err != nil {
		return err
	}

	// 删除源文件夹中剩余的书签（这些是因为冲突而没有移动的）
	_, err = DB.Exec(`DELETE FROM bookmarks WHERE folder_path = ?`, sourceFolder)
	if err != nil {
		return err
	}

	// 更新源文件夹子目录下的书签
	oldPrefix := sourceFolder + "/"
	var newPrefix string
	if targetFolder == "" {
		newPrefix = ""
	} else {
		newPrefix = targetFolder + "/"
	}

	// 获取所有需要处理的子文件夹路径
	rows, err := DB.Query(`SELECT DISTINCT folder_path FROM bookmarks WHERE folder_path LIKE ?`, oldPrefix+"%")
	if err != nil {
		return err
	}

	var subFolders []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err == nil {
			subFolders = append(subFolders, path)
		}
	}
	rows.Close()

	// 逐个处理子文件夹
	for _, oldPath := range subFolders {
		var newPath string
		if newPrefix == "" {
			newPath = strings.TrimPrefix(oldPath, oldPrefix)
		} else {
			newPath = newPrefix + strings.TrimPrefix(oldPath, oldPrefix)
		}

		// 尝试更新，冲突的跳过
		_, err = DB.Exec(`UPDATE OR IGNORE bookmarks SET folder_path = ?, updated_at = CURRENT_TIMESTAMP WHERE folder_path = ?`,
			newPath, oldPath)
		if err != nil {
			return err
		}

		// 删除该子文件夹中剩余的书签（冲突的）
		_, err = DB.Exec(`DELETE FROM bookmarks WHERE folder_path = ?`, oldPath)
		if err != nil {
			return err
		}
	}

	return nil
}

// QuickCreateBookmark 快速创建书签
func QuickCreateBookmark(url, title, favicon, folderPath string) (*Bookmark, error) {
	result, err := DB.Exec(`
		INSERT INTO bookmarks (url, title, description, folder_path, favicon)
		VALUES (?, ?, '', ?, ?)
	`, url, title, folderPath, favicon)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return GetBookmarkByID(id)
}
