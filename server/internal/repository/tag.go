package repository

import (
	"Nibstash_v2_server/database"
	"Nibstash_v2_server/internal/model"
)

type TagRepository struct{}

func NewTagRepository() *TagRepository {
	return &TagRepository{}
}

func (r *TagRepository) Create(name, color string) (*model.Tag, error) {
	if color == "" {
		color = "#3b82f6"
	}
	result, err := database.DB.Exec(`INSERT INTO tags (name, color) VALUES (?, ?)`, name, color)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &model.Tag{ID: id, Name: name, Color: color}, nil
}

func (r *TagRepository) GetByID(id int64) (*model.Tag, error) {
	tag := &model.Tag{}
	err := database.DB.QueryRow(`SELECT id, name, color FROM tags WHERE id = ?`, id).Scan(&tag.ID, &tag.Name, &tag.Color)
	if err != nil {
		return nil, err
	}
	return tag, nil
}

func (r *TagRepository) GetByName(name string) (*model.Tag, error) {
	tag := &model.Tag{}
	err := database.DB.QueryRow(`SELECT id, name, color FROM tags WHERE name = ?`, name).Scan(&tag.ID, &tag.Name, &tag.Color)
	if err != nil {
		return nil, err
	}
	return tag, nil
}

func (r *TagRepository) Update(id int64, name, color string) error {
	_, err := database.DB.Exec(`UPDATE tags SET name = ?, color = ? WHERE id = ?`, name, color, id)
	return err
}

func (r *TagRepository) Delete(id int64) error {
	_, err := database.DB.Exec(`DELETE FROM tags WHERE id = ?`, id)
	return err
}

func (r *TagRepository) List() ([]model.Tag, error) {
	rows, err := database.DB.Query(`
		SELECT t.id, t.name, t.color, COUNT(bt.bookmark_id) as count
		FROM tags t
		LEFT JOIN bookmark_tags bt ON t.id = bt.tag_id
		GROUP BY t.id
		ORDER BY count DESC, t.name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []model.Tag
	for rows.Next() {
		var t model.Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Color, &t.Count); err == nil {
			tags = append(tags, t)
		}
	}
	return tags, nil
}

func (r *TagRepository) GetByBookmarkID(bookmarkID int64) ([]model.Tag, error) {
	rows, err := database.DB.Query(`
		SELECT t.id, t.name, t.color
		FROM tags t
		JOIN bookmark_tags bt ON t.id = bt.tag_id
		WHERE bt.bookmark_id = ?
	`, bookmarkID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []model.Tag
	for rows.Next() {
		var t model.Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Color); err == nil {
			tags = append(tags, t)
		}
	}
	return tags, nil
}
