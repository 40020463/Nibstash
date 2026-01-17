package main

import (
	"fmt"
	"log"
	"path/filepath"

	"Nibstash_v2_server/config"
	"Nibstash_v2_server/database"
	"Nibstash_v2_server/internal/handler"
	"Nibstash_v2_server/internal/middleware"
	"Nibstash_v2_server/internal/repository"
	"Nibstash_v2_server/internal/util"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	if err := config.Load("config.json"); err != nil {
		log.Printf("加载配置失败，使用默认配置: %v", err)
	}

	// 初始化加密模块
	if err := util.InitCrypto(config.App.EncryptKey); err != nil {
		log.Fatalf("初始化加密模块失败: %v", err)
	}

	// 初始化数据库
	if err := database.Init(config.App.DBPath); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer database.Close()

	// 执行数据库迁移
	if err := database.Migrate(); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 确保默认用户存在
	userRepo := repository.NewUserRepository()
	if err := userRepo.EnsureDefaultUser(); err != nil {
		log.Fatalf("创建默认用户失败: %v", err)
	}

	// 同步现有书签的域名到 domains 表
	domainRepo := repository.NewDomainRepository()
	if err := domainRepo.SyncDomainsFromBookmarks(); err != nil {
		log.Printf("同步域名失败: %v", err)
	}

	// 创建 Gin 实例
	r := gin.Default()

	// 中间件
	r.Use(middleware.CORS())

	// 初始化 Handler
	authHandler := handler.NewAuthHandler()
	bookmarkHandler := handler.NewBookmarkHandler()
	folderHandler := handler.NewFolderHandler()
	tagHandler := handler.NewTagHandler()
	domainHandler := handler.NewDomainHandler()
	credentialHandler := handler.NewCredentialHandler()
	faviconHandler := handler.NewFaviconHandler()
	importHandler := handler.NewImportHandler()
	bookmarkletHandler := handler.NewBookmarkletHandler()

	// API 路由
	api := r.Group("/api")
	{
		// 认证相关（无需登录）
		api.POST("/auth/login", authHandler.Login)

		// Bookmarklet（自己处理认证）
		api.GET("/bookmarklet", bookmarkletHandler.Handle)
		api.POST("/bookmarklet", bookmarkletHandler.Save)

		// 需要认证的路由
		auth := api.Group("")
		auth.Use(middleware.Auth())
		{
			// 用户
			auth.GET("/auth/me", authHandler.GetMe)
			auth.PUT("/auth/password", authHandler.ChangePassword)

			// 书签
			auth.GET("/bookmarks", bookmarkHandler.List)
			auth.POST("/bookmarks", bookmarkHandler.Create)
			auth.GET("/bookmarks/:id", bookmarkHandler.Get)
			auth.PUT("/bookmarks/:id", bookmarkHandler.Update)
			auth.DELETE("/bookmarks/:id", bookmarkHandler.Delete)
			auth.POST("/bookmarks/batch", bookmarkHandler.Batch)
			auth.GET("/bookmarks/export", bookmarkHandler.Export)
			auth.POST("/bookmarks/import", importHandler.Import)
			auth.DELETE("/bookmarks/clear", bookmarkHandler.ClearAll)
			auth.POST("/bookmarks/clear-folder", bookmarkHandler.ClearFolder)

			// 文件夹
			auth.GET("/folders", folderHandler.List)
			auth.POST("/folders", folderHandler.Create)
			auth.PUT("/folders/move", folderHandler.Move)
			auth.PUT("/folders/merge", folderHandler.Merge)
			auth.DELETE("/folders", folderHandler.Delete)

			// 标签
			auth.GET("/tags", tagHandler.List)
			auth.POST("/tags", tagHandler.Create)
			auth.PUT("/tags/:id", tagHandler.Update)
			auth.DELETE("/tags/:id", tagHandler.Delete)

			// 域名（实时计算）
			auth.GET("/domains", domainHandler.List)
			auth.GET("/domains/:domain/bookmarks", domainHandler.GetBookmarks)
			auth.DELETE("/domains/:domain", domainHandler.Delete)

			// 凭证
			auth.GET("/credentials", credentialHandler.List)
			auth.POST("/credentials", credentialHandler.Create)
			auth.GET("/credentials/:id", credentialHandler.Get)
			auth.GET("/credentials/domain/:domain", credentialHandler.GetByDomain)
			auth.PUT("/credentials/:id", credentialHandler.Update)
			auth.DELETE("/credentials/:id", credentialHandler.Delete)

			// Favicon
			auth.GET("/favicons/pending", faviconHandler.GetPending)
			auth.PUT("/favicons/:id", faviconHandler.Update)
		}
	}

	// Vue dist 目录路径
	distPath := filepath.Join("..", "web", "dist")

	// 静态资源
	r.Static("/assets", filepath.Join(distPath, "assets"))

	// favicon
	r.StaticFile("/favicon.ico", filepath.Join(distPath, "favicon.ico"))

	// 根路径返回 index.html
	r.GET("/", func(c *gin.Context) {
		c.File(filepath.Join(distPath, "index.html"))
	})

	// SPA 路由兜底（非 API 路由都返回 index.html）
	r.NoRoute(func(c *gin.Context) {
		c.File(filepath.Join(distPath, "index.html"))
	})

	// 启动服务器
	addr := fmt.Sprintf(":%d", config.App.Port)
	log.Printf("服务器启动在 http://localhost%s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}
