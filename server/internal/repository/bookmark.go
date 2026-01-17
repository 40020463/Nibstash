package repository

import (
	"Nibstash_v2_server/database"
	"Nibstash_v2_server/internal/model"
	"strings"
)

type BookmarkRepository struct {
	tagRepo    *TagRepository
	domainRepo *DomainRepository
}

func NewBookmarkRepository() *BookmarkRepository {
	return &BookmarkRepository{
		tagRepo:    NewTagRepository(),
		domainRepo: NewDomainRepository(),
	}
}

func (r *BookmarkRepository) Create(url, title, description, favicon, folderPath string, tagIDs []int64) (*model.Bookmark, error) {
	result, err := database.DB.Exec(`
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
		database.DB.Exec(`INSERT OR IGNORE INTO bookmark_tags (bookmark_id, tag_id) VALUES (?, ?)`, id, tagID)
	}

	// 同步添加域名到 domains 表
	r.domainRepo.AddDomain(url)

	return r.GetByID(id)
}

func (r *BookmarkRepository) GetByID(id int64) (*model.Bookmark, error) {
	bookmark := &model.Bookmark{}
	err := database.DB.QueryRow(`
		SELECT id, url, title, description, folder_path, favicon, created_at, updated_at
		FROM bookmarks WHERE id = ?
	`, id).Scan(&bookmark.ID, &bookmark.URL, &bookmark.Title, &bookmark.Description,
		&bookmark.FolderPath, &bookmark.Favicon, &bookmark.CreatedAt, &bookmark.UpdatedAt)
	if err != nil {
		return nil, err
	}

	bookmark.Tags, _ = r.tagRepo.GetByBookmarkID(id)
	return bookmark, nil
}

func (r *BookmarkRepository) GetByURL(url string) (*model.Bookmark, error) {
	bookmark := &model.Bookmark{}
	err := database.DB.QueryRow(`
		SELECT id, url, title, description, folder_path, favicon, created_at, updated_at
		FROM bookmarks WHERE url = ?
	`, url).Scan(&bookmark.ID, &bookmark.URL, &bookmark.Title, &bookmark.Description,
		&bookmark.FolderPath, &bookmark.Favicon, &bookmark.CreatedAt, &bookmark.UpdatedAt)
	if err != nil {
		return nil, err
	}

	bookmark.Tags, _ = r.tagRepo.GetByBookmarkID(bookmark.ID)
	return bookmark, nil
}

func (r *BookmarkRepository) Update(id int64, url, title, description string, tagIDs []int64) error {
	_, err := database.DB.Exec(`
		UPDATE bookmarks SET url = ?, title = ?, description = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, url, title, description, id)
	if err != nil {
		return err
	}

	// 更新标签关联
	database.DB.Exec(`DELETE FROM bookmark_tags WHERE bookmark_id = ?`, id)
	for _, tagID := range tagIDs {
		database.DB.Exec(`INSERT OR IGNORE INTO bookmark_tags (bookmark_id, tag_id) VALUES (?, ?)`, id, tagID)
	}

	return nil
}

func (r *BookmarkRepository) Delete(id int64) error {
	_, err := database.DB.Exec(`DELETE FROM bookmarks WHERE id = ?`, id)
	return err
}

func (r *BookmarkRepository) DeleteByIDs(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	placeholders := strings.Repeat("?,", len(ids))
	placeholders = placeholders[:len(placeholders)-1]
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	_, err := database.DB.Exec(`DELETE FROM bookmarks WHERE id IN (`+placeholders+`)`, args...)
	return err
}

func (r *BookmarkRepository) List(page, pageSize int, tagID int64, search, folderPath string, filterFolder bool, sortBy string) ([]model.Bookmark, int, error) {
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
	database.DB.QueryRow(countQuery, args...).Scan(&total)

	// 排序
	orderClause := "b.created_at DESC"
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

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var bookmarks []model.Bookmark
	for rows.Next() {
		var b model.Bookmark
		err := rows.Scan(&b.ID, &b.URL, &b.Title, &b.Description, &b.FolderPath, &b.Favicon, &b.CreatedAt, &b.UpdatedAt)
		if err != nil {
			continue
		}
		b.Tags, _ = r.tagRepo.GetByBookmarkID(b.ID)
		bookmarks = append(bookmarks, b)
	}

	return bookmarks, total, nil
}

func (r *BookmarkRepository) Exists(url, folderPath string) bool {
	var count int
	database.DB.QueryRow(`SELECT COUNT(*) FROM bookmarks WHERE url = ? AND folder_path = ?`, url, folderPath).Scan(&count)
	return count > 0
}

func (r *BookmarkRepository) MoveToFolder(ids []int64, targetFolder string) error {
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
	_, err := database.DB.Exec(`UPDATE bookmarks SET folder_path = ?, updated_at = CURRENT_TIMESTAMP WHERE id IN (`+placeholders+`)`, args...)
	return err
}

func (r *BookmarkRepository) UpdateFavicon(id int64, favicon string) error {
	_, err := database.DB.Exec(`UPDATE bookmarks SET favicon = ? WHERE id = ?`, favicon, id)
	return err
}

func (r *BookmarkRepository) GetWithoutFavicon() ([]model.Bookmark, error) {
	rows, err := database.DB.Query(`SELECT id, url FROM bookmarks WHERE (favicon = '' OR favicon IS NULL) AND url NOT LIKE 'nibstash://folder-placeholder/%'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookmarks []model.Bookmark
	for rows.Next() {
		var b model.Bookmark
		if err := rows.Scan(&b.ID, &b.URL); err == nil {
			bookmarks = append(bookmarks, b)
		}
	}
	return bookmarks, nil
}

func (r *BookmarkRepository) GetAllForExport() ([]model.Bookmark, error) {
	rows, err := database.DB.Query(`SELECT url, title, folder_path FROM bookmarks WHERE url NOT LIKE 'nibstash://folder-placeholder/%' ORDER BY folder_path, title`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookmarks []model.Bookmark
	for rows.Next() {
		var b model.Bookmark
		if err := rows.Scan(&b.URL, &b.Title, &b.FolderPath); err == nil {
			bookmarks = append(bookmarks, b)
		}
	}
	return bookmarks, nil
}

func (r *BookmarkRepository) BatchImport(bookmarks []model.ImportBookmark, skipDuplicates bool) (imported, skipped int, err error) {
	tx, err := database.DB.Begin()
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

	// 收集需要添加的域名
	var importedURLs []string

	for _, bm := range bookmarks {
		if bm.URL == "" || bm.Title == "" {
			continue
		}

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
				importedURLs = append(importedURLs, bm.URL)
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return imported, skipped, err
	}

	// 同步添加域名到 domains 表
	for _, url := range importedURLs {
		r.domainRepo.AddDomain(url)
	}

	return imported, skipped, nil
}

// DeleteAll 清空所有书签
func (r *BookmarkRepository) DeleteAll() error {
	_, err := database.DB.Exec(`DELETE FROM bookmarks`)
	return err
}

// DeleteByFolder 删除指定文件夹的书签
func (r *BookmarkRepository) DeleteByFolder(folderPath string) error {
	if folderPath == "" {
		_, err := database.DB.Exec(`DELETE FROM bookmarks WHERE folder_path = ''`)
		return err
	}
	_, err := database.DB.Exec(`DELETE FROM bookmarks WHERE folder_path = ? OR folder_path LIKE ?`, folderPath, folderPath+"/%")
	return err
}
