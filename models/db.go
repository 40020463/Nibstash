package models

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB(dbPath string) error {
	// 确保目录存在
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}

	// 创建表
	if err := createTables(); err != nil {
		return err
	}

	log.Println("数据库初始化成功")
	return nil
}

func createTables() error {
	// 书签表 - url + folder_path 组合唯一
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS bookmarks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			url TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT DEFAULT '',
			folder_path TEXT DEFAULT '',
			favicon TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(url, folder_path)
		)
	`)
	if err != nil {
		return err
	}

	// 兼容旧库：补充 folder_path 字段
	if err := ensureBookmarkFolderColumn(); err != nil {
		return err
	}

	// 标签表
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS tags (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			color TEXT DEFAULT '#3b82f6'
		)
	`)
	if err != nil {
		return err
	}

	// 书签-标签关联表
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS bookmark_tags (
			bookmark_id INTEGER NOT NULL,
			tag_id INTEGER NOT NULL,
			PRIMARY KEY (bookmark_id, tag_id),
			FOREIGN KEY (bookmark_id) REFERENCES bookmarks(id) ON DELETE CASCADE,
			FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	// 创建索引
	DB.Exec(`CREATE INDEX IF NOT EXISTS idx_bookmarks_url ON bookmarks(url)`)
	DB.Exec(`CREATE INDEX IF NOT EXISTS idx_bookmarks_created ON bookmarks(created_at DESC)`)
	DB.Exec(`CREATE INDEX IF NOT EXISTS idx_bookmarks_folder ON bookmarks(folder_path)`)
	DB.Exec(`CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name)`)

	return nil
}

func ensureBookmarkFolderColumn() error {
	_, err := DB.Exec(`ALTER TABLE bookmarks ADD COLUMN folder_path TEXT DEFAULT ''`)
	if err != nil {
		// 已存在字段时会报错，这里忽略即可
		if strings.Contains(err.Error(), "duplicate column name") {
			return nil
		}
		return err
	}
	return nil
}

func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}
