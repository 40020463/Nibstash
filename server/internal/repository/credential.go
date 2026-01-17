package repository

import (
	"Nibstash_v2_server/database"
	"Nibstash_v2_server/internal/model"
	"Nibstash_v2_server/internal/util"
)

type CredentialRepository struct{}

func NewCredentialRepository() *CredentialRepository {
	return &CredentialRepository{}
}

func (r *CredentialRepository) Create(domain, title, username, password, notes string) (*model.Credential, error) {
	// 加密密码
	encryptedPassword, err := util.Encrypt(password)
	if err != nil {
		return nil, err
	}

	result, err := database.DB.Exec(`
		INSERT INTO credentials (domain, title, username, password, notes, updated_at)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, domain, title, username, encryptedPassword, notes)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return r.GetByID(id)
}

func (r *CredentialRepository) GetByID(id int64) (*model.Credential, error) {
	cred := &model.Credential{}
	var encryptedPassword string
	err := database.DB.QueryRow(`
		SELECT id, domain, title, username, password, notes, created_at, updated_at
		FROM credentials WHERE id = ?
	`, id).Scan(&cred.ID, &cred.Domain, &cred.Title, &cred.Username, &encryptedPassword, &cred.Notes, &cred.CreatedAt, &cred.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// 解密密码
	if encryptedPassword != "" {
		decrypted, err := util.Decrypt(encryptedPassword)
		if err == nil {
			cred.Password = decrypted
		} else {
			// 如果解密失败，可能是旧数据（未加密），直接使用
			cred.Password = encryptedPassword
		}
	}

	return cred, nil
}

func (r *CredentialRepository) GetByDomain(domain string) ([]model.Credential, error) {
	rows, err := database.DB.Query(`
		SELECT id, domain, title, username, password, notes, created_at, updated_at
		FROM credentials WHERE domain = ? ORDER BY id ASC
	`, domain)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var creds []model.Credential
	for rows.Next() {
		var cred model.Credential
		var encryptedPassword string
		if err := rows.Scan(&cred.ID, &cred.Domain, &cred.Title, &cred.Username, &encryptedPassword, &cred.Notes, &cred.CreatedAt, &cred.UpdatedAt); err == nil {
			// 解密密码
			if encryptedPassword != "" {
				decrypted, err := util.Decrypt(encryptedPassword)
				if err == nil {
					cred.Password = decrypted
				} else {
					cred.Password = encryptedPassword
				}
			}
			creds = append(creds, cred)
		}
	}
	return creds, nil
}

func (r *CredentialRepository) Update(id int64, title, username, password, notes string) error {
	// 加密密码
	encryptedPassword, err := util.Encrypt(password)
	if err != nil {
		return err
	}

	_, err = database.DB.Exec(`
		UPDATE credentials SET title = ?, username = ?, password = ?, notes = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, title, username, encryptedPassword, notes, id)
	return err
}

func (r *CredentialRepository) Delete(id int64) error {
	_, err := database.DB.Exec(`DELETE FROM credentials WHERE id = ?`, id)
	return err
}

func (r *CredentialRepository) DeleteByDomain(domain string) error {
	_, err := database.DB.Exec(`DELETE FROM credentials WHERE domain = ?`, domain)
	return err
}

func (r *CredentialRepository) List() ([]model.Credential, error) {
	rows, err := database.DB.Query(`
		SELECT id, domain, title, username, password, notes, created_at, updated_at
		FROM credentials ORDER BY domain ASC, id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var creds []model.Credential
	for rows.Next() {
		var cred model.Credential
		var encryptedPassword string
		if err := rows.Scan(&cred.ID, &cred.Domain, &cred.Title, &cred.Username, &encryptedPassword, &cred.Notes, &cred.CreatedAt, &cred.UpdatedAt); err == nil {
			if encryptedPassword != "" {
				decrypted, err := util.Decrypt(encryptedPassword)
				if err == nil {
					cred.Password = decrypted
				} else {
					cred.Password = encryptedPassword
				}
			}
			creds = append(creds, cred)
		}
	}
	return creds, nil
}
