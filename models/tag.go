package models

type Tag struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
	Count int    `json:"count"` // 书签数量
}

// CreateTag 创建标签
func CreateTag(name, color string) (*Tag, error) {
	if color == "" {
		color = "#3b82f6"
	}
	result, err := DB.Exec(`INSERT INTO tags (name, color) VALUES (?, ?)`, name, color)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &Tag{ID: id, Name: name, Color: color}, nil
}

// GetTagByID 根据ID获取标签
func GetTagByID(id int64) (*Tag, error) {
	tag := &Tag{}
	err := DB.QueryRow(`SELECT id, name, color FROM tags WHERE id = ?`, id).Scan(&tag.ID, &tag.Name, &tag.Color)
	if err != nil {
		return nil, err
	}
	return tag, nil
}

// UpdateTag 更新标签
func UpdateTag(id int64, name, color string) error {
	_, err := DB.Exec(`UPDATE tags SET name = ?, color = ? WHERE id = ?`, name, color, id)
	return err
}

// DeleteTag 删除标签
func DeleteTag(id int64) error {
	_, err := DB.Exec(`DELETE FROM tags WHERE id = ?`, id)
	return err
}

// ListTags 获取所有标签（带书签数量）
func ListTags() ([]Tag, error) {
	rows, err := DB.Query(`
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

	var tags []Tag
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Color, &t.Count); err == nil {
			tags = append(tags, t)
		}
	}
	return tags, nil
}

// GetTagByName 根据名称获取标签
func GetTagByName(name string) (*Tag, error) {
	tag := &Tag{}
	err := DB.QueryRow(`SELECT id, name, color FROM tags WHERE name = ?`, name).Scan(&tag.ID, &tag.Name, &tag.Color)
	if err != nil {
		return nil, err
	}
	return tag, nil
}
