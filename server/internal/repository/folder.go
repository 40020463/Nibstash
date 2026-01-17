package repository

import (
	"Nibstash_v2_server/database"
	"Nibstash_v2_server/internal/model"
	"sort"
	"strings"
)

type FolderRepository struct{}

func NewFolderRepository() *FolderRepository {
	return &FolderRepository{}
}

// GetFolderTree 获取文件夹树结构
func (r *FolderRepository) GetFolderTree() ([]model.FolderNode, error) {
	// 获取所有文件夹路径及其书签数量
	rows, err := database.DB.Query(`
		SELECT folder_path, COUNT(*) as count
		FROM bookmarks
		WHERE url NOT LIKE 'nibstash://folder-placeholder/%'
		GROUP BY folder_path
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pathCounts := make(map[string]int)
	for rows.Next() {
		var path string
		var count int
		if err := rows.Scan(&path, &count); err == nil {
			pathCounts[path] = count
		}
	}

	// 构建所有节点（使用指针）
	nodeMap := make(map[string]*model.FolderNode)

	for path := range pathCounts {
		if path == "" {
			continue
		}

		parts := strings.Split(path, "/")
		currentPath := ""

		for i, part := range parts {
			if i > 0 {
				currentPath += "/"
			}
			currentPath += part

			if _, exists := nodeMap[currentPath]; !exists {
				nodeMap[currentPath] = &model.FolderNode{
					Name:     part,
					Path:     currentPath,
					Count:    0,
					Children: []model.FolderNode{},
				}
			}
		}
	}

	// 设置每个节点的书签数量
	for path, count := range pathCounts {
		if node, exists := nodeMap[path]; exists {
			node.Count = count
		}
	}

	// 构建父子关系 - 先收集所有根节点和子节点关系
	rootNodes := []*model.FolderNode{}
	childrenMap := make(map[string][]*model.FolderNode)

	for path, node := range nodeMap {
		parentPath := getParentPath(path)
		if parentPath == "" {
			rootNodes = append(rootNodes, node)
		} else {
			childrenMap[parentPath] = append(childrenMap[parentPath], node)
		}
	}

	// 递归构建树
	var buildTree func(node *model.FolderNode)
	buildTree = func(node *model.FolderNode) {
		children := childrenMap[node.Path]
		for _, child := range children {
			buildTree(child)
			node.Children = append(node.Children, *child)
		}
		// 排序子节点
		sort.Slice(node.Children, func(i, j int) bool {
			return node.Children[i].Name < node.Children[j].Name
		})
	}

	// 构建每个根节点的子树
	result := []model.FolderNode{}
	for _, root := range rootNodes {
		buildTree(root)
		result = append(result, *root)
	}

	// 排序根节点
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result, nil
}

func getParentPath(path string) string {
	idx := strings.LastIndex(path, "/")
	if idx == -1 {
		return ""
	}
	return path[:idx]
}

func sortFolderNodes(nodes []model.FolderNode) {
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Name < nodes[j].Name
	})
	for i := range nodes {
		sortFolderNodes(nodes[i].Children)
	}
}

// ListPaths 获取所有文件夹路径
func (r *FolderRepository) ListPaths() ([]string, error) {
	rows, err := database.DB.Query(`SELECT DISTINCT folder_path FROM bookmarks WHERE folder_path != '' ORDER BY folder_path`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var paths []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err == nil {
			paths = append(paths, path)
		}
	}
	return paths, nil
}

// Create 创建文件夹（通过创建占位书签）
func (r *FolderRepository) Create(path string) error {
	// 检查是否已存在
	var count int
	database.DB.QueryRow(`SELECT COUNT(*) FROM bookmarks WHERE folder_path = ?`, path).Scan(&count)
	if count > 0 {
		return nil // 已存在
	}

	// 创建占位书签
	_, err := database.DB.Exec(`
		INSERT INTO bookmarks (url, title, description, folder_path, favicon)
		VALUES (?, ?, '', ?, '')
	`, "nibstash://folder-placeholder/"+path, path, path)
	return err
}

// Move 移动文件夹
func (r *FolderRepository) Move(sourceFolder, targetFolder string) error {
	if sourceFolder == "" {
		return nil
	}

	parts := strings.Split(sourceFolder, "/")
	folderName := parts[len(parts)-1]

	var newFolderBase string
	if targetFolder == "" {
		newFolderBase = folderName
	} else {
		newFolderBase = targetFolder + "/" + folderName
	}

	if sourceFolder == newFolderBase {
		return nil
	}

	if strings.HasPrefix(newFolderBase, sourceFolder+"/") {
		return nil
	}

	// 更新精确匹配源文件夹的书签
	_, err := database.DB.Exec(`UPDATE OR IGNORE bookmarks SET folder_path = ?, updated_at = CURRENT_TIMESTAMP WHERE folder_path = ?`,
		newFolderBase, sourceFolder)
	if err != nil {
		return err
	}

	// 更新源文件夹子目录下的书签
	oldPrefix := sourceFolder + "/"
	newPrefix := newFolderBase + "/"

	_, err = database.DB.Exec(`UPDATE OR IGNORE bookmarks SET folder_path = ? || SUBSTR(folder_path, ?), updated_at = CURRENT_TIMESTAMP WHERE folder_path LIKE ?`,
		newPrefix, len(oldPrefix)+1, oldPrefix+"%")

	return err
}

// Merge 合并文件夹
func (r *FolderRepository) Merge(sourceFolder, targetFolder string) error {
	if sourceFolder == "" {
		return nil
	}

	if sourceFolder == targetFolder {
		return nil
	}

	if targetFolder != "" && strings.HasPrefix(targetFolder, sourceFolder+"/") {
		return nil
	}

	// 先尝试更新，冲突的会被跳过
	_, err := database.DB.Exec(`UPDATE OR IGNORE bookmarks SET folder_path = ?, updated_at = CURRENT_TIMESTAMP WHERE folder_path = ?`,
		targetFolder, sourceFolder)
	if err != nil {
		return err
	}

	// 删除源文件夹中剩余的书签
	_, err = database.DB.Exec(`DELETE FROM bookmarks WHERE folder_path = ?`, sourceFolder)
	if err != nil {
		return err
	}

	// 处理子文件夹
	oldPrefix := sourceFolder + "/"
	var newPrefix string
	if targetFolder == "" {
		newPrefix = ""
	} else {
		newPrefix = targetFolder + "/"
	}

	rows, err := database.DB.Query(`SELECT DISTINCT folder_path FROM bookmarks WHERE folder_path LIKE ?`, oldPrefix+"%")
	if err != nil {
		return err
	}

	var subFolders []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err == nil {
			subFolders = append(subFolders, path)
		}
	}
	rows.Close()

	for _, oldPath := range subFolders {
		var newPath string
		if newPrefix == "" {
			newPath = strings.TrimPrefix(oldPath, oldPrefix)
		} else {
			newPath = newPrefix + strings.TrimPrefix(oldPath, oldPrefix)
		}

		_, err = database.DB.Exec(`UPDATE OR IGNORE bookmarks SET folder_path = ?, updated_at = CURRENT_TIMESTAMP WHERE folder_path = ?`,
			newPath, oldPath)
		if err != nil {
			return err
		}

		_, err = database.DB.Exec(`DELETE FROM bookmarks WHERE folder_path = ?`, oldPath)
		if err != nil {
			return err
		}
	}

	return nil
}

// Delete 删除文件夹及其所有书签
func (r *FolderRepository) Delete(folderPath string) error {
	if folderPath == "" {
		_, err := database.DB.Exec(`DELETE FROM bookmarks WHERE folder_path = ''`)
		return err
	}
	_, err := database.DB.Exec(`DELETE FROM bookmarks WHERE folder_path = ? OR folder_path LIKE ?`, folderPath, folderPath+"/%")
	return err
}

// HasUncategorized 检查是否有未分类书签
func (r *FolderRepository) HasUncategorized() bool {
	var count int
	database.DB.QueryRow(`SELECT COUNT(*) FROM bookmarks WHERE folder_path = '' AND url NOT LIKE 'nibstash://folder-placeholder/%'`).Scan(&count)
	return count > 0
}

// GetAllPaths 获取所有文件夹路径（用于 bookmarklet）
func (r *FolderRepository) GetAllPaths() []string {
	rows, err := database.DB.Query(`SELECT DISTINCT folder_path FROM bookmarks WHERE folder_path != '' ORDER BY folder_path`)
	if err != nil {
		return []string{}
	}
	defer rows.Close()

	var paths []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err == nil {
			paths = append(paths, path)
		}
	}
	return paths
}
