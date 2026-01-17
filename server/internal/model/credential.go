package model

import "time"

type Credential struct {
	ID        int64     `json:"id"`
	Domain    string    `json:"domain"`
	Title     string    `json:"title"`
	Username  string    `json:"username"`
	Password  string    `json:"password"` // 解密后的密码
	Notes     string    `json:"notes"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CredentialCreateRequest struct {
	Domain   string `json:"domain" binding:"required"`
	Title    string `json:"title"`
	Username string `json:"username"`
	Password string `json:"password"`
	Notes    string `json:"notes"`
}

type CredentialUpdateRequest struct {
	Title    string `json:"title"`
	Username string `json:"username"`
	Password string `json:"password"`
	Notes    string `json:"notes"`
}
