package models

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// DomainDB 独立的域名数据库连接
var DomainDB *sql.DB

// InitDomainDB 初始化域名数据库
func InitDomainDB(dbPath string) error {
	// 确保目录存在
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	var err error
	DomainDB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}

	// 创建表
	if err := createDomainTables(); err != nil {
		return err
	}

	log.Println("域名数据库初始化成功")
	return nil
}

func createDomainTables() error {
	// 域名凭证表（支持多账号，移除 UNIQUE 约束）
	_, err := DomainDB.Exec(`
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
	`)
	if err != nil {
		return err
	}

	// 域名缓存表
	_, err = DomainDB.Exec(`
		CREATE TABLE IF NOT EXISTS domain_cache (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			top_domain TEXT NOT NULL,
			sub_domain TEXT DEFAULT '',
			bookmark_count INTEGER DEFAULT 0,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(top_domain, sub_domain)
		)
	`)
	if err != nil {
		return err
	}

	DomainDB.Exec(`CREATE INDEX IF NOT EXISTS idx_credentials_domain ON credentials(domain)`)
	DomainDB.Exec(`CREATE INDEX IF NOT EXISTS idx_domain_cache_top ON domain_cache(top_domain)`)

	return nil
}

// CloseDomainDB 关闭域名数据库
func CloseDomainDB() {
	if DomainDB != nil {
		DomainDB.Close()
	}
}

// DomainCredential 域名凭证
type DomainCredential struct {
	ID        int64     `json:"id"`
	Domain    string    `json:"domain"`
	Title     string    `json:"title"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	Notes     string    `json:"notes"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DomainGroup 域名分组（用于展示）
type DomainGroup struct {
	TopDomain     string   `json:"top_domain"`
	SubDomains    []string `json:"sub_domains"`
	BookmarkCount int      `json:"bookmark_count"`
}

// GetCredentialByDomain 根据域名获取凭证（返回第一个，兼容旧代码）
func GetCredentialByDomain(domain string) (*DomainCredential, error) {
	cred := &DomainCredential{}
	err := DomainDB.QueryRow(`
		SELECT id, domain, title, username, password, notes, created_at, updated_at
		FROM credentials WHERE domain = ? ORDER BY id ASC LIMIT 1
	`, domain).Scan(&cred.ID, &cred.Domain, &cred.Title, &cred.Username, &cred.Password, &cred.Notes, &cred.CreatedAt, &cred.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return cred, nil
}

// GetCredentialsByDomain 根据域名获取所有凭证（多账号）
func GetCredentialsByDomain(domain string) ([]DomainCredential, error) {
	rows, err := DomainDB.Query(`
		SELECT id, domain, title, username, password, notes, created_at, updated_at
		FROM credentials WHERE domain = ? ORDER BY id ASC
	`, domain)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var creds []DomainCredential
	for rows.Next() {
		var cred DomainCredential
		if err := rows.Scan(&cred.ID, &cred.Domain, &cred.Title, &cred.Username, &cred.Password, &cred.Notes, &cred.CreatedAt, &cred.UpdatedAt); err == nil {
			creds = append(creds, cred)
		}
	}
	return creds, nil
}

// GetCredentialByID 根据 ID 获取凭证
func GetCredentialByID(id int64) (*DomainCredential, error) {
	cred := &DomainCredential{}
	err := DomainDB.QueryRow(`
		SELECT id, domain, title, username, password, notes, created_at, updated_at
		FROM credentials WHERE id = ?
	`, id).Scan(&cred.ID, &cred.Domain, &cred.Title, &cred.Username, &cred.Password, &cred.Notes, &cred.CreatedAt, &cred.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return cred, nil
}

// SaveCredential 保存域名凭证（新增或更新）
func SaveCredential(domain, title, username, password, notes string) error {
	_, err := DomainDB.Exec(`
		INSERT INTO credentials (domain, title, username, password, notes, updated_at)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, domain, title, username, password, notes)
	return err
}

// UpdateCredential 更新指定 ID 的凭证
func UpdateCredential(id int64, title, username, password, notes string) error {
	_, err := DomainDB.Exec(`
		UPDATE credentials SET title = ?, username = ?, password = ?, notes = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, title, username, password, notes, id)
	return err
}

// DeleteCredential 删除指定 ID 的凭证
func DeleteCredential(id int64) error {
	_, err := DomainDB.Exec(`DELETE FROM credentials WHERE id = ?`, id)
	return err
}

// DeleteDomainData 删除指定域名的所有数据（凭证和缓存）
func DeleteDomainData(domain string) error {
	// 删除该域名的所有凭证
	_, err := DomainDB.Exec(`DELETE FROM credentials WHERE domain = ?`, domain)
	if err != nil {
		return err
	}

	// 从缓存中删除（可能是顶级域名或子域名）
	_, err = DomainDB.Exec(`DELETE FROM domain_cache WHERE top_domain = ? OR sub_domain = ?`, domain, domain)
	return err
}

// GetAllDomainsFromCache 从缓存获取域名列表（快速）
func GetAllDomainsFromCache() ([]DomainGroup, error) {
	rows, err := DomainDB.Query(`
		SELECT top_domain, sub_domain, bookmark_count
		FROM domain_cache
		ORDER BY bookmark_count DESC, top_domain ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := make(map[string]*DomainGroup)
	for rows.Next() {
		var topDomain, subDomain string
		var count int
		if err := rows.Scan(&topDomain, &subDomain, &count); err == nil {
			if _, exists := groups[topDomain]; !exists {
				groups[topDomain] = &DomainGroup{
					TopDomain:     topDomain,
					SubDomains:    []string{},
					BookmarkCount: 0,
				}
			}
			groups[topDomain].BookmarkCount += count
			if subDomain != "" && subDomain != topDomain {
				groups[topDomain].SubDomains = append(groups[topDomain].SubDomains, subDomain)
			}
		}
	}

	result := make([]DomainGroup, 0, len(groups))
	for _, group := range groups {
		result = append(result, *group)
	}

	sortDomainGroups(result)
	return result, nil
}

// IsDomainCacheEmpty 检查缓存是否为空
func IsDomainCacheEmpty() bool {
	var count int
	DomainDB.QueryRow(`SELECT COUNT(*) FROM domain_cache`).Scan(&count)
	return count == 0
}

// RefreshDomainCache 刷新域名缓存（从主数据库读取书签）
func RefreshDomainCache() error {
	// 清空缓存
	DomainDB.Exec(`DELETE FROM domain_cache`)

	// 从主数据库的书签中提取所有域名
	rows, err := DB.Query(`
		SELECT url FROM bookmarks
		WHERE url NOT LIKE 'nibstash://folder-placeholder/%'
		AND url LIKE 'http%'
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	// 收集所有域名及其书签数量
	domainCount := make(map[string]int)
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err == nil {
			domain := extractDomain(url)
			if domain != "" {
				domainCount[domain]++
			}
		}
	}

	// 按顶级域名分组并插入缓存
	for domain, count := range domainCount {
		topDomain := getTopDomain(domain)
		if topDomain == "" {
			continue
		}

		subDomain := ""
		if domain != topDomain {
			subDomain = domain
		}

		DomainDB.Exec(`
			INSERT INTO domain_cache (top_domain, sub_domain, bookmark_count, updated_at)
			VALUES (?, ?, ?, CURRENT_TIMESTAMP)
			ON CONFLICT(top_domain, sub_domain) DO UPDATE SET
				bookmark_count = excluded.bookmark_count,
				updated_at = CURRENT_TIMESTAMP
		`, topDomain, subDomain, count)
	}

	return nil
}

// extractDomain 从URL中提取域名
func extractDomain(url string) string {
	var domain string
	if len(url) > 8 && url[:8] == "https://" {
		domain = url[8:]
	} else if len(url) > 7 && url[:7] == "http://" {
		domain = url[7:]
	} else {
		return ""
	}

	// 找到第一个 / 或字符串结束
	for i := 0; i < len(domain); i++ {
		if domain[i] == '/' || domain[i] == '?' || domain[i] == '#' {
			domain = domain[:i]
			break
		}
	}

	// 移除端口号
	if idx := findChar(domain, ':'); idx != -1 {
		domain = domain[:idx]
	}

	return domain
}

// 辅助函数：查找字符位置
func findChar(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

// 辅助函数：获取顶级域名（简化版）
func getTopDomain(domain string) string {
	parts := splitDomain(domain)
	if len(parts) < 2 {
		return domain
	}

	// 常见的二级顶级域名
	secondLevelTLDs := map[string]bool{
		"com.cn": true, "net.cn": true, "org.cn": true, "gov.cn": true,
		"co.uk": true, "co.jp": true, "co.kr": true,
		"com.hk": true, "com.tw": true,
	}

	if len(parts) >= 3 {
		possibleSecondLevel := parts[len(parts)-2] + "." + parts[len(parts)-1]
		if secondLevelTLDs[possibleSecondLevel] {
			if len(parts) >= 3 {
				return parts[len(parts)-3] + "." + possibleSecondLevel
			}
			return possibleSecondLevel
		}
	}

	// 返回最后两部分作为顶级域名
	return parts[len(parts)-2] + "." + parts[len(parts)-1]
}

// 辅助函数：分割域名
func splitDomain(domain string) []string {
	var parts []string
	start := 0
	for i := 0; i <= len(domain); i++ {
		if i == len(domain) || domain[i] == '.' {
			if i > start {
				parts = append(parts, domain[start:i])
			}
			start = i + 1
		}
	}
	return parts
}

// 辅助函数：按书签数量排序
func sortDomainGroups(groups []DomainGroup) {
	for i := 0; i < len(groups)-1; i++ {
		for j := i + 1; j < len(groups); j++ {
			if groups[j].BookmarkCount > groups[i].BookmarkCount {
				groups[i], groups[j] = groups[j], groups[i]
			}
		}
	}
}

// GetBookmarksByDomain 获取指定域名的所有书签（从主数据库）
func GetBookmarksByDomain(domain string) ([]Bookmark, error) {
	rows, err := DB.Query(`
		SELECT id, url, title, description, folder_path, favicon, created_at, updated_at
		FROM bookmarks
		WHERE (url LIKE ? OR url LIKE ?)
		AND url NOT LIKE 'nibstash://folder-placeholder/%'
		ORDER BY created_at DESC
	`, "http://%"+domain+"%", "https://%"+domain+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookmarks []Bookmark
	for rows.Next() {
		var b Bookmark
		if err := rows.Scan(&b.ID, &b.URL, &b.Title, &b.Description, &b.FolderPath, &b.Favicon, &b.CreatedAt, &b.UpdatedAt); err == nil {
			bookmarks = append(bookmarks, b)
		}
	}
	return bookmarks, nil
}
