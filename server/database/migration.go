package database

import (
	"log"
)

// Migrate 执行数据库迁移
func Migrate() error {
	// 用户表
	if _, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE DEFAULT 'admin',
			password TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		return err
	}

	// 书签表
	if _, err := DB.Exec(`
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
	`); err != nil {
		return err
	}

	// 标签表
	if _, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS tags (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			color TEXT DEFAULT '#3b82f6'
		)
	`); err != nil {
		return err
	}

	// 书签-标签关联表
	if _, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS bookmark_tags (
			bookmark_id INTEGER NOT NULL,
			tag_id INTEGER NOT NULL,
			PRIMARY KEY (bookmark_id, tag_id),
			FOREIGN KEY (bookmark_id) REFERENCES bookmarks(id) ON DELETE CASCADE,
			FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
		)
	`); err != nil {
		return err
	}

	// 凭证表（密码加密存储）
	if _, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS credentials (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			domain TEXT NOT NULL,
			title TEXT DEFAULT '',
			username TEXT DEFAULT '',
			password TEXT DEFAULT '',
			notes TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		return err
	}

	// 系统配置表
	if _, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT DEFAULT ''
		)
	`); err != nil {
		return err
	}

	// 域名表（持久化存储域名，不随书签删除而消失）
	if _, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS domains (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			domain TEXT NOT NULL UNIQUE,
			top_domain TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		return err
	}

	// 创建索引
	DB.Exec(`CREATE INDEX IF NOT EXISTS idx_bookmarks_url ON bookmarks(url)`)
	DB.Exec(`CREATE INDEX IF NOT EXISTS idx_bookmarks_created ON bookmarks(created_at DESC)`)
	DB.Exec(`CREATE INDEX IF NOT EXISTS idx_bookmarks_folder ON bookmarks(folder_path)`)
	DB.Exec(`CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name)`)
	DB.Exec(`CREATE INDEX IF NOT EXISTS idx_credentials_domain ON credentials(domain)`)
	DB.Exec(`CREATE INDEX IF NOT EXISTS idx_domains_domain ON domains(domain)`)
	DB.Exec(`CREATE INDEX IF NOT EXISTS idx_domains_top_domain ON domains(top_domain)`)

	log.Println("数据库迁移完成")
	return nil
}
