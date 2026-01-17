package model

// DomainGroup 域名分组（用于展示）
type DomainGroup struct {
	TopDomain      string   `json:"top_domain"`
	SubDomains     []string `json:"sub_domains"`
	BookmarkCount  int      `json:"bookmark_count"`
	HasCredentials bool     `json:"has_credentials"`
}

// DomainListResponse 域名列表响应
type DomainListResponse struct {
	Domains []DomainGroup `json:"domains"`
}
