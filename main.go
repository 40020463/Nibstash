package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"nibstash/handlers"
	"nibstash/models"
)

// TemplateRenderer 模板渲染器
type TemplateRenderer struct {
	layoutPath string
	tmplDir    string
	appName    string
}

// Render 渲染模板
func (t *TemplateRenderer) Render(w http.ResponseWriter, name string, data interface{}) {
	// 将数据转换为 map 以便添加通用数据
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		dataMap = make(map[string]interface{})
	}

	// 添加通用数据
	if name != "login.html" {
		dataMap["ShowNav"] = true
	}
	dataMap["AppName"] = t.appName

	// 设置活动导航
	switch name {
	case "index.html":
		dataMap["ActiveNav"] = "home"
	case "tags.html":
		dataMap["ActiveNav"] = "tags"
	case "bookmarklet.html":
		dataMap["ActiveNav"] = "bookmarklet"
	case "import.html":
		dataMap["ActiveNav"] = "import"
	}

	// 为每个页面单独解析模板（layout + 具体页面）
	tmpl, err := template.ParseFiles(t.layoutPath, filepath.Join(t.tmplDir, name))
	if err != nil {
		log.Printf("模板解析错误: %v", err)
		http.Error(w, "内部错误", http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "layout", dataMap)
	if err != nil {
		log.Printf("模板渲染错误: %v", err)
		http.Error(w, "内部错误", http.StatusInternalServerError)
	}
}

func main() {
	// 加载配置
	if err := LoadConfig("config.json"); err != nil {
		log.Printf("配置加载失败，使用默认配置: %v", err)
	}

	// 初始化数据库
	if err := models.InitDB(AppConfig.DBPath); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer models.CloseDB()

	// 初始化域名数据库（独立存储）
	domainDBPath := "data/domain.db"
	if err := models.InitDomainDB(domainDBPath); err != nil {
		log.Fatalf("域名数据库初始化失败: %v", err)
	}
	defer models.CloseDomainDB()

	// 模板渲染器
	renderer := &TemplateRenderer{
		layoutPath: filepath.Join("templates", "layout.html"),
		tmplDir:    "templates",
		appName:    AppConfig.AppName,
	}

	// 静态文件服务
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// 认证相关路由
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			handlers.LoginPostHandler(AppConfig.Password, renderer)(w, r)
		} else {
			handlers.LoginHandler(renderer)(w, r)
		}
	})
	http.HandleFunc("/logout", handlers.LogoutHandler)

	// 需要认证的路由
	authMiddleware := func(h http.HandlerFunc) http.HandlerFunc {
		return handlers.AuthMiddleware(AppConfig.Password, h)
	}

	// 首页
	http.HandleFunc("/", authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		handlers.IndexHandler(renderer)(w, r)
	}))

	// 书签相关
	http.HandleFunc("/bookmark/add", authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			handlers.BookmarkAddPostHandler(renderer)(w, r)
		} else {
			handlers.BookmarkAddHandler(renderer)(w, r)
		}
	}))

	http.HandleFunc("/bookmark/edit", authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			handlers.BookmarkEditPostHandler(renderer)(w, r)
		} else {
			handlers.BookmarkEditHandler(renderer)(w, r)
		}
	}))

	http.HandleFunc("/bookmark/delete", authMiddleware(handlers.BookmarkDeleteHandler))
	http.HandleFunc("/bookmark/clear", authMiddleware(handlers.BookmarkClearHandler))
	http.HandleFunc("/bookmark/delete-selected", authMiddleware(handlers.BookmarkDeleteBatchHandler))
	http.HandleFunc("/bookmark/clear-folder", authMiddleware(handlers.BookmarkClearFolderHandler))
	http.HandleFunc("/bookmark/export", authMiddleware(handlers.BookmarkExportHandler))
	http.HandleFunc("/bookmark/move", authMiddleware(handlers.BookmarkMoveHandler))

	// 标签相关
	http.HandleFunc("/tags", authMiddleware(handlers.TagsHandler(renderer)))
	http.HandleFunc("/tag/add", authMiddleware(handlers.TagAddHandler(renderer)))
	http.HandleFunc("/tag/delete", authMiddleware(handlers.TagDeleteHandler))

	// Bookmarklet
	http.HandleFunc("/bookmarklet", authMiddleware(handlers.BookmarkletHandler(AppConfig.BaseURL, renderer)))

	// 导入书签
	http.HandleFunc("/import", authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			handlers.ImportPostHandler(renderer)(w, r)
		} else {
			handlers.ImportHandler(renderer)(w, r)
		}
	}))

	// API (用于 Bookmarklet)
	http.HandleFunc("/api/add", handlers.APIAddBookmarkHandler(AppConfig.Password))

	// Favicon API
	http.HandleFunc("/api/favicon/pending", handlers.APIFaviconPendingHandler())
	http.HandleFunc("/api/favicon/update", handlers.APIFaviconUpdateHandler())

	// Folder API
	http.HandleFunc("/api/folder/move", handlers.APIMoveFolderHandler())
	http.HandleFunc("/api/folder/merge", handlers.APIMergeFolderHandler())
	http.HandleFunc("/api/folder/create", handlers.APICreateFolderHandler())

	// Quick Add API
	http.HandleFunc("/api/bookmark/quick-add", handlers.APIQuickAddBookmarkHandler())

	// Domain Credential API
	http.HandleFunc("/api/domain/list", authMiddleware(handlers.APIDomainListHandler()))
	http.HandleFunc("/api/domain/refresh", authMiddleware(handlers.APIDomainRefreshHandler()))
	http.HandleFunc("/api/domain/credential", authMiddleware(handlers.APIDomainCredentialHandler()))
	http.HandleFunc("/api/domain/credential/save", authMiddleware(handlers.APIDomainCredentialSaveHandler()))
	http.HandleFunc("/api/domain/credential/update", authMiddleware(handlers.APIDomainCredentialUpdateHandler()))
	http.HandleFunc("/api/domain/credential/delete", authMiddleware(handlers.APIDomainCredentialDeleteHandler()))
	http.HandleFunc("/api/domain/delete", authMiddleware(handlers.APIDomainDeleteHandler()))

	// 启动服务器
	addr := fmt.Sprintf(":%d", AppConfig.Port)
	log.Printf("%s启动成功！访问 http://localhost%s", AppConfig.AppName, addr)
	log.Printf("默认密码: %s", AppConfig.Password)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
