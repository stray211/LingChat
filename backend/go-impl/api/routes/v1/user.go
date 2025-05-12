package v1

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"LingChat/api/routes/common"
	"LingChat/api/routes/middleware"
	"LingChat/api/routes/v1/request"
	"LingChat/internal/data"
	"LingChat/internal/data/ent/ent"
	"LingChat/internal/service"
	"LingChat/pkg/jwt"
)

type UserRoute struct {
	userService service.UserService
	jwt         *jwt.JWT
	userRepo    data.UserRepo
}

func NewUserRoute(userService service.UserService, userRepo data.UserRepo, jwt *jwt.JWT) *UserRoute {
	return &UserRoute{
		userService: userService,
		jwt:         jwt,
		userRepo:    userRepo,
	}
}

func (u *UserRoute) RegisterRoute(r *gin.RouterGroup) {
	rg := r.Group("/v1/user")
	{
		rg.POST("/register", u.register)
		rg.POST("/login", u.login)
		rg.POST("/password", middleware.TokenAuth(true, u.jwt, u.userRepo), u.changePassword)
	}
}

// register 用户注册处理
func (u *UserRoute) register(c *gin.Context) {
	var req request.UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "Invalid request: " + err.Error(),
		})
		return
	}

	user, token, err := u.userService.Register(c.Request.Context(), &data.User{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  err.Error(),
		})
		return
	}

	// 设置Cookie
	c.SetCookie("token", token, 86400, "/", "", false, true) // 1天有效期

	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"msg":  "Registration successful",
		"data": gin.H{
			"user_id":  user.ID,
			"username": user.Username,
			"email":    user.Email,
			"token":    token,
		},
	})
}

// login 用户登录处理
func (u *UserRoute) login(c *gin.Context) {
	var req request.UserRequest
	var user *ent.User
	var token string
	var err error
	if c.Request.ContentLength != 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  "Invalid request: " + err.Error(),
			})
			return
		}

		user, token, err = u.userService.Login(c.Request.Context(), &data.User{
			Username: req.Username,
			Password: req.Password,
		})
	} else {
		user, token, err = u.userService.Login(c.Request.Context(), nil)
	}
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": http.StatusUnauthorized,
			"msg":  err.Error(),
		})
		return
	}

	// 设置Cookie
	c.SetCookie("token", token, int(7*24*time.Hour), "/", "", false, true) // 7天有效期

	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"msg":  "Login successful",
		"data": gin.H{
			"user_id":  user.ID,
			"username": user.Username,
			"email":    user.Email,
			"token":    token,
		},
	})
}

// changePassword 修改密码处理
func (u *UserRoute) changePassword(c *gin.Context) {
	// 获取用户请求参数
	var req request.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "无效的请求参数: " + err.Error(),
		})
		return
	}

	// 从上下文中获取当前登录用户
	user, exists := c.Get(common.CurrentUserInfoKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": http.StatusUnauthorized,
			"msg":  "用户未登录",
		})
		return
	}
	currentUser := user.(*ent.User)

	// 调用服务修改密码
	err := u.userService.ChangePassword(c.Request.Context(), currentUser.ID, req.OldPassword, req.NewPassword)
	if err != nil {
		var statusCode int
		if err == service.ErrInvalidPassword {
			statusCode = http.StatusBadRequest
		} else {
			statusCode = http.StatusInternalServerError
		}
		c.JSON(statusCode, gin.H{
			"code": statusCode,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"msg":  "密码修改成功",
	})
}
