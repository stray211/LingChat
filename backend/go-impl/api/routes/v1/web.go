package v1

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type WebRoute struct {
	staticDir string
}

func NewWebRoute(staticDir string) *WebRoute {
	return &WebRoute{
		staticDir: staticDir,
	}
}

// RegisterRoute 实现Route接口（空实现，因为Web路由在Engine级别注册）
func (w *WebRoute) RegisterRoute(r *gin.RouterGroup) {
	// Web路由由HttpEngine直接在Engine级别注册，这里不需要实现
}

// RegisterWebRoute 在Engine级别注册Web路由
func (w *WebRoute) RegisterWebRoute(engine *gin.Engine) {
	// 静态文件服务
	engine.Static("/css", filepath.Join(w.staticDir, "css"))
	engine.Static("/js", filepath.Join(w.staticDir, "js"))
	engine.Static("/pictures", filepath.Join(w.staticDir, "pictures"))
	engine.Static("/audio_effects", filepath.Join(w.staticDir, "audio_effects"))

	// 页面路由
	engine.GET("/", w.indexPage)
	engine.GET("/login", w.loginPage)
	engine.GET("/about", w.aboutPage)
}

// indexPage 处理根路径，返回index.html
func (w *WebRoute) indexPage(c *gin.Context) {
	c.File(filepath.Join(w.staticDir, "pages", "index.html"))
}

// loginPage 处理登录路径，返回login.html
func (w *WebRoute) loginPage(c *gin.Context) {
	c.File(filepath.Join(w.staticDir, "pages", "login.html"))
}

// aboutPage 处理/about路径，返回about.html
func (w *WebRoute) aboutPage(c *gin.Context) {
	aboutPath := filepath.Join(w.staticDir, "pages", "about.html")
	c.File(aboutPath)
}

// notFound 处理404错误
func (w *WebRoute) notFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{
		"error": "Page not found",
		"code":  404,
	})
}
