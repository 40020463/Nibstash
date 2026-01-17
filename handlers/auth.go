package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"
)

var (
	sessions     = make(map[string]time.Time)
	sessionMutex sync.RWMutex
)

const sessionCookieName = "nibstash_session"
const sessionDuration = 24 * time.Hour

// GenerateSessionID 生成会话ID
func GenerateSessionID() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// CreateSession 创建会话
func CreateSession(w http.ResponseWriter) {
	sessionID := GenerateSessionID()
	sessionMutex.Lock()
	sessions[sessionID] = time.Now().Add(sessionDuration)
	sessionMutex.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   int(sessionDuration.Seconds()),
	})
}

// ValidateSession 验证会话
func ValidateSession(r *http.Request) bool {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return false
	}

	sessionMutex.RLock()
	expiry, exists := sessions[cookie.Value]
	sessionMutex.RUnlock()

	if !exists || time.Now().After(expiry) {
		return false
	}
	return true
}

// DestroySession 销毁会话
func DestroySession(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err == nil {
		sessionMutex.Lock()
		delete(sessions, cookie.Value)
		sessionMutex.Unlock()
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
}

// AuthMiddleware 认证中间件
func AuthMiddleware(password string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 如果密码为空，不需要认证
		if password == "" {
			next(w, r)
			return
		}

		if !ValidateSession(r) {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next(w, r)
	}
}

// LoginHandler 登录页面
func LoginHandler(tmpl TemplateRenderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if ValidateSession(r) {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		tmpl.Render(w, "login.html", map[string]interface{}{
			"Error": "",
		})
	}
}

// LoginPostHandler 登录验证
func LoginPostHandler(password string, tmpl TemplateRenderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		inputPassword := r.FormValue("password")

		if inputPassword == password {
			CreateSession(w)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		tmpl.Render(w, "login.html", map[string]interface{}{
			"Error": "密码错误",
		})
	}
}

// LogoutHandler 登出
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	DestroySession(w, r)
	http.Redirect(w, r, "/login", http.StatusFound)
}

// TemplateRenderer 模板渲染接口
type TemplateRenderer interface {
	Render(w http.ResponseWriter, name string, data interface{})
}
