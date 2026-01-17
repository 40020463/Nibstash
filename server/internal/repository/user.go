package repository

import (
	"Nibstash_v2_server/config"
	"Nibstash_v2_server/database"
	"Nibstash_v2_server/internal/model"

	"golang.org/x/crypto/bcrypt"
)

type UserRepository struct{}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

// GetByID 根据ID获取用户
func (r *UserRepository) GetByID(id int64) (*model.User, error) {
	user := &model.User{}
	err := database.DB.QueryRow(`
		SELECT id, username, password, created_at, updated_at
		FROM users WHERE id = ?
	`, id).Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetByUsername 根据用户名获取用户
func (r *UserRepository) GetByUsername(username string) (*model.User, error) {
	user := &model.User{}
	err := database.DB.QueryRow(`
		SELECT id, username, password, created_at, updated_at
		FROM users WHERE username = ?
	`, username).Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// EnsureDefaultUser 确保默认用户存在
func (r *UserRepository) EnsureDefaultUser() error {
	var count int
	database.DB.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count)
	if count > 0 {
		return nil
	}

	// 创建默认用户
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(config.App.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = database.DB.Exec(`
		INSERT INTO users (username, password) VALUES (?, ?)
	`, "admin", string(hashedPassword))
	return err
}

// UpdatePassword 更新密码
func (r *UserRepository) UpdatePassword(id int64, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = database.DB.Exec(`
		UPDATE users SET password = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?
	`, string(hashedPassword), id)
	return err
}

// VerifyPassword 验证密码
func (r *UserRepository) VerifyPassword(user *model.User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return err == nil
}
