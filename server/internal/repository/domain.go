package repository

import (
	"Nibstash_v2_server/database"
	"Nibstash_v2_server/internal/model"
	"sort"
	"strings"
)

type DomainRepository struct{}

func NewDomainRepository() *DomainRepository {
	return &DomainRepository{}
}

// AddDomain 添加域名到 domains 表（添加书签时调用）
func (r *DomainRepository) AddDomain(url string) error {
	domain := ExtractDomain(url)
	if domain == "" {
		return nil
	}

	topDomain := GetTopDomain(domain)
	if topDomain == "" {
		return nil
	}

	// 使用 INSERT OR IGNORE 避免重复
	_, err := database.DB.Exec(`
		INSERT OR IGNORE INTO domains (domain, top_domain) VALUES (?, ?)
	`, domain, topDomain)
	return err
}

// DeleteDomain 删除域名记录（同时删除该域名下的凭证）
func (r *DomainRepository) DeleteDomain(domain string) error {
	// 删除域名记录
	_, err := database.DB.Exec(`DELETE FROM domains WHERE domain = ? OR top_domain = ?`, domain, domain)
	return err
}

// GetAllDomains 从 domains 表获取域名列表，并计算书签数量和凭证信息
func (r *DomainRepository) GetAllDomains() ([]model.DomainGroup, error) {
	// 1. 从 domains 表获取所有域名
	rows, err := database.DB.Query(`SELECT domain, top_domain FROM domains ORDER BY top_domain`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 收集所有域名
	allDomains := make(map[string]string) // domain -> top_domain
	for rows.Next() {
		var domain, topDomain string
		if err := rows.Scan(&domain, &topDomain); err == nil {
			allDomains[domain] = topDomain
		}
	}

	// 2. 计算每个域名的书签数量
	domainCount := make(map[string]int)
	bookmarkRows, err := database.DB.Query(`
		SELECT url FROM bookmarks
		WHERE url NOT LIKE 'nibstash://folder-placeholder/%'
		AND url LIKE 'http%'
	`)
	if err != nil {
		return nil, err
	}
	defer bookmarkRows.Close()

	for bookmarkRows.Next() {
		var url string
		if err := bookmarkRows.Scan(&url); err == nil {
			domain := ExtractDomain(url)
			if domain != "" {
				domainCount[domain]++
			}
		}
	}

	// 3. 获取有凭证的域名
	credRows, err := database.DB.Query(`SELECT DISTINCT domain FROM credentials`)
	if err != nil {
		return nil, err
	}
	defer credRows.Close()

	credentialDomains := make(map[string]bool)
	for credRows.Next() {
		var domain string
		if err := credRows.Scan(&domain); err == nil && domain != "" {
			credentialDomains[domain] = true
		}
	}

	// 4. 按顶级域名分组
	groups := make(map[string]*model.DomainGroup)
	for domain, topDomain := range allDomains {
		if _, exists := groups[topDomain]; !exists {
			groups[topDomain] = &model.DomainGroup{
				TopDomain:      topDomain,
				SubDomains:     []string{},
				BookmarkCount:  0,
				HasCredentials: false,
			}
		}
		groups[topDomain].BookmarkCount += domainCount[domain]
		if domain != topDomain {
			groups[topDomain].SubDomains = append(groups[topDomain].SubDomains, domain)
		}
		// 标记该域名组是否有凭证
		if credentialDomains[domain] || credentialDomains[topDomain] {
			groups[topDomain].HasCredentials = true
		}
	}

	result := make([]model.DomainGroup, 0, len(groups))
	for _, group := range groups {
		result = append(result, *group)
	}

	// 按书签数量排序
	sort.Slice(result, func(i, j int) bool {
		if result[i].BookmarkCount == result[j].BookmarkCount {
			if result[i].HasCredentials != result[j].HasCredentials {
				return result[i].HasCredentials
			}
			return result[i].TopDomain < result[j].TopDomain
		}
		return result[i].BookmarkCount > result[j].BookmarkCount
	})

	return result, nil
}

// SyncDomainsFromBookmarks 从现有书签同步域名到 domains 表（用于初始化或修复）
func (r *DomainRepository) SyncDomainsFromBookmarks() error {
	rows, err := database.DB.Query(`
		SELECT DISTINCT url FROM bookmarks
		WHERE url NOT LIKE 'nibstash://folder-placeholder/%'
		AND url LIKE 'http%'
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err == nil {
			r.AddDomain(url)
		}
	}
	return nil
}

// GetBookmarksByDomain 获取指定域名的所有书签
func (r *DomainRepository) GetBookmarksByDomain(domain string) ([]model.Bookmark, error) {
	rows, err := database.DB.Query(`
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

	var bookmarks []model.Bookmark
	for rows.Next() {
		var b model.Bookmark
		if err := rows.Scan(&b.ID, &b.URL, &b.Title, &b.Description, &b.FolderPath, &b.Favicon, &b.CreatedAt, &b.UpdatedAt); err == nil {
			bookmarks = append(bookmarks, b)
		}
	}
	return bookmarks, nil
}

// ExtractDomain 从URL中提取域名（导出供其他包使用）
func ExtractDomain(url string) string {
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
	if idx := strings.Index(domain, ":"); idx != -1 {
		domain = domain[:idx]
	}

	return domain
}

// GetTopDomain 获取顶级域名（导出供其他包使用）
func GetTopDomain(domain string) string {
	parts := strings.Split(domain, ".")
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

	return parts[len(parts)-2] + "." + parts[len(parts)-1]
}
