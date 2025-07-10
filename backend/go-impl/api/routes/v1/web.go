package v1

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type WebRoute struct {
	staticDir    string
	tempVoiceDir string
}

func NewWebRoute(staticDir, tempVoiceDir string) *WebRoute {
	return &WebRoute{
		staticDir:    staticDir,
		tempVoiceDir: tempVoiceDir,
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

	// 为临时音频文件提供自定义静态文件服务，支持从odd/even子目录中查找
	engine.GET("/audio/*filepath", w.audioFileHandler)

	// 页面路由
	engine.GET("/", w.indexPage)
	engine.GET("/login", w.loginPage)
	engine.GET("/about", w.aboutPage)
}

// audioFileHandler 自定义音频文件处理器，支持从odd/even子目录中查找文件
func (w *WebRoute) audioFileHandler(c *gin.Context) {
	requestPath := c.Param("filepath")
	if requestPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File path is required"})
		return
	}

	// 去掉路径开头的斜杠
	if requestPath[0] == '/' {
		requestPath = requestPath[1:]
	}

	// 尝试在odd和even目录中查找文件
	subDirs := []string{"odd", "even"}

	for _, subDir := range subDirs {
		fullPath := filepath.Join(w.tempVoiceDir, subDir, requestPath)

		// 检查文件是否存在
		if _, err := os.Stat(fullPath); err == nil {
			c.File(fullPath)
			return
		}
	}

	// 如果在子目录中都没找到，尝试直接在根目录查找（向后兼容）
	directPath := filepath.Join(w.tempVoiceDir, requestPath)
	if _, err := os.Stat(directPath); err == nil {
		c.File(directPath)
		return
	}

	// 文件未找到
	c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
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
