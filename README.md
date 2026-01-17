# Nibstash v2 (囤囤鼠)

一个现代化的个人书签收藏夹和账号密码管理器，采用前后端分离架构。

## 📋 项目简介

Nibstash v2 是从原有的 Golang + HTML 模板架构迁移到 Vue3 前后端分离架构的全新版本。提供书签管理、标签系统、文件夹组织、域名凭证管理等功能。

## ✨ 核心功能

- **📚 书签管理**
  - 书签的增删改查
  - 批量操作（删除、移动）
  - 文件夹树形结构组织
  - 全文搜索和多维度排序
  - 导入/导出功能（支持浏览器书签格式）
  - Favicon 自动获取

- **🏷️ 标签系统**
  - 多标签关联
  - 自定义标签颜色
  - 按标签筛选书签

- **🔐 域名凭证管理**
  - 按域名分组的账号密码管理
  - 密码 AES-GCM 加密存储
  - 域名自动提取和分类

- **🔖 Bookmarklet**
  - 浏览器快速收藏工具
  - 一键保存当前页面

## 🏗️ 技术架构

### 后端 (server/)

| 技术栈 | 版本 | 说明 |
|--------|------|------|
| Go | 1.24.2 | 编程语言 |
| Gin | 1.11.0 | Web 框架 |
| SQLite | modernc.org/sqlite 1.37.1 | 数据库 |
| JWT | golang-jwt/jwt/v5 5.2.1 | 身份认证 |
| Crypto | golang.org/x/crypto | 密码加密 (AES-GCM) |

**项目结构：**
```
server/
├── main.go                 # 入口文件
├── config/                 # 配置管理
├── database/               # 数据库初始化和迁移
├── internal/
│   ├── handler/           # HTTP 处理器
│   ├── middleware/        # 中间件（认证、CORS）
│   ├── model/             # 数据模型
│   ├── repository/        # 数据访问层
│   └── util/              # 工具函数（加密等）
└── data/                  # 数据库文件目录
```

### 前端 (web/)

| 技术栈 | 版本 | 说明 |
|--------|------|------|
| Vue | 3.5.26 | 前端框架 |
| Vite | 7.3.0 | 构建工具 |
| Element Plus | 2.13.1 | UI 组件库 |
| Pinia | 3.0.4 | 状态管理 |
| Vue Router | 4.6.4 | 路由管理 |
| Axios | 1.13.2 | HTTP 客户端 |

**项目结构：**
```
web/
├── src/
│   ├── components/        # 可复用组件
│   ├── views/             # 页面组件
│   ├── stores/            # Pinia 状态管理
│   ├── router/            # 路由配置
│   ├── styles/            # 全局样式
│   └── main.js            # 入口文件
├── public/                # 静态资源
└── dist/                  # 构建输出目录
```

## 🗄️ 数据库结构

使用单个 SQLite 数据库 (`data/nibstash.db`)，包含以下表：

| 表名 | 说明 | 主要字段 |
|------|------|----------|
| users | 用户表 | id, username, password, created_at, updated_at |
| bookmarks | 书签表 | id, url, title, description, folder_path, favicon, created_at, updated_at |
| tags | 标签表 | id, name, color |
| bookmark_tags | 书签-标签关联表 | bookmark_id, tag_id |
| credentials | 凭证表 | id, domain, title, username, password (加密), notes, created_at, updated_at |
| domains | 域名表 | id, domain, top_domain, created_at |
| settings | 系统配置表 | key, value |

**索引优化：**
- bookmarks: url, created_at, folder_path
- tags: name
- credentials: domain
- domains: domain, top_domain

## 🚀 快速开始

### 环境要求

- **后端**：Go 1.24.2 或更高版本
- **前端**：Node.js 20.19.0+ 或 22.12.0+
- **操作系统**：Windows / Linux / macOS

### 安装步骤

#### 1. 克隆项目

```bash
git clone <repository-url>
cd Nibstash_v2
```

#### 2. 后端设置

```bash
cd server

# 安装依赖
go mod download

# 配置文件（可选，首次运行会自动生成默认配置）
# 编辑 config.json 修改端口、密码等配置

# 启动后端服务
go run main.go

# 或构建可执行文件
go build -o server.exe
./server.exe
```

**默认配置：**
- 端口：8080
- 默认密码：nibstash
- 数据库路径：data/nibstash.db

#### 3. 前端设置

```bash
cd web

# 安装依赖
npm install

# 开发模式（热重载）
npm run dev

# 生产构建
npm run build

# 预览构建结果
npm run preview
```

**开发模式：**
- 前端开发服务器：http://localhost:5173
- API 代理到后端：http://localhost:8080

#### 4. 访问应用

- **开发模式**：http://localhost:5173
- **生产模式**：http://localhost:8080（后端会服务前端构建文件）

**默认登录信息：**
- 用户名：admin
- 密码：nibstash（可在 config.json 中修改）

## 📡 API 接口

### 认证相关
- `POST /api/auth/login` - 用户登录
- `GET /api/auth/me` - 获取当前用户信息
- `PUT /api/auth/password` - 修改密码

### 书签管理
- `GET /api/bookmarks` - 获取书签列表（支持分页、搜索、排序）
- `POST /api/bookmarks` - 创建书签
- `GET /api/bookmarks/:id` - 获取单个书签
- `PUT /api/bookmarks/:id` - 更新书签
- `DELETE /api/bookmarks/:id` - 删除书签
- `POST /api/bookmarks/batch` - 批量操作（删除、移动）
- `GET /api/bookmarks/export` - 导出书签
- `POST /api/bookmarks/import` - 导入书签
- `DELETE /api/bookmarks/clear` - 清空所有书签
- `POST /api/bookmarks/clear-folder` - 清空文件夹

### 文件夹管理
- `GET /api/folders` - 获取文件夹树
- `POST /api/folders` - 创建文件夹
- `PUT /api/folders/move` - 移动文件夹
- `PUT /api/folders/merge` - 合并文件夹
- `DELETE /api/folders` - 删除文件夹

### 标签管理
- `GET /api/tags` - 获取所有标签
- `POST /api/tags` - 创建标签
- `PUT /api/tags/:id` - 更新标签
- `DELETE /api/tags/:id` - 删除标签

### 域名管理
- `GET /api/domains` - 获取域名列表（实时计算）
- `GET /api/domains/:domain/bookmarks` - 获取域名下的书签
- `DELETE /api/domains/:domain` - 删除域名及其书签

### 凭证管理
- `GET /api/credentials` - 获取凭证列表
- `POST /api/credentials` - 创建凭证
- `GET /api/credentials/:id` - 获取单个凭证
- `GET /api/credentials/domain/:domain` - 按域名获取凭证
- `PUT /api/credentials/:id` - 更新凭证
- `DELETE /api/credentials/:id` - 删除凭证

### Favicon 管理
- `GET /api/favicons/pending` - 获取待更新的 Favicon
- `PUT /api/favicons/:id` - 更新 Favicon

### Bookmarklet
- `GET /api/bookmarklet` - Bookmarklet 页面
- `POST /api/bookmarklet` - 保存书签（支持 token 认证）

## 🔒 安全特性

- **JWT 认证**：基于 Token 的无状态认证
- **密码加密**：凭证密码使用 AES-GCM 加密存储
- **CORS 配置**：跨域请求安全控制
- **SQL 注入防护**：使用参数化查询
- **XSS 防护**：前端输入验证和转义

## 🛠️ 开发指南

### 后端开发

```bash
cd server

# 运行测试
go test ./...

# 代码格式化
go fmt ./...

# 构建
go build -o server.exe
```

### 前端开发

```bash
cd web

# 启动开发服务器（热重载）
npm run dev

# 代码检查
npm run lint

# 构建生产版本
npm run build
```

### 数据库迁移

数据库迁移在应用启动时自动执行（`database/migration.go`）。如需手动迁移：

```go
// 在 main.go 中已包含
database.Migrate()
```

## 📝 配置说明

### 后端配置 (server/config.json)

```json
{
  "port": 8080,                                    // 服务端口
  "password": "nibstash",                          // 默认密码
  "jwt_secret": "nibstash-jwt-secret-change-me-32bytes!",  // JWT 密钥（生产环境请修改）
  "db_path": "data/nibstash.db",                   // 数据库路径
  "base_url": "http://localhost:8080",             // 基础 URL
  "app_name": "囤囤鼠",                             // 应用名称
  "encrypt_key": "nibstash-encrypt-key-32-bytes!!!" // AES 加密密钥（32字节，生产环境请修改）
}
```

### 前端配置 (web/vite.config.js)

```javascript
export default defineConfig({
  server: {
    port: 5173,                    // 开发服务器端口
    proxy: {
      '/api': {
        target: 'http://localhost:8080',  // 后端 API 地址
        changeOrigin: true
      }
    }
  }
})
```

## 🚧 已知问题

- ~~域名管理栏无法实时刷新数据~~ (已修复：改为实时计算)
- ~~数据库结构不合理~~ (已优化：统一使用单数据库)

## 📦 部署

### 生产环境部署

1. **构建前端**
```bash
cd web
npm run build
# 构建产物在 web/dist/ 目录
```

2. **构建后端**
```bash
cd server
go build -o server.exe
```

3. **配置**
- 修改 `server/config.json` 中的敏感配置（jwt_secret, encrypt_key, password）
- 确保 `encrypt_key` 为 32 字节

4. **运行**
```bash
cd server
./server.exe
```

后端会自动服务前端构建文件，访问 http://localhost:8080 即可。

### Docker 部署（可选）

```dockerfile
# 待补充 Dockerfile
```

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request！

1. Fork 本项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 📄 许可证

[待补充]

## 🙏 致谢

- [Gin](https://github.com/gin-gonic/gin) - Go Web 框架
- [Vue.js](https://vuejs.org/) - 渐进式 JavaScript 框架
- [Element Plus](https://element-plus.org/) - Vue 3 组件库
- [modernc.org/sqlite](https://gitlab.com/cznic/sqlite) - 纯 Go SQLite 驱动

## 📮 联系方式

[待补充]

---

**注意**：本项目仅供个人学习和使用，请勿用于商业用途。生产环境部署前请务必修改默认密码和密钥配置。
