package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"

	"LingChat/api/routes/middleware"
	"LingChat/api/routes/v1/request"
	"LingChat/api/routes/v1/response"
	"LingChat/internal/data"
	"LingChat/internal/service"
	"LingChat/pkg/jwt"
)

type ChatRoute struct {
	lingChatService     *service.LingChatService
	conversationService *service.ConversationService
	characterService    service.CharacterService
	userRepo            data.UserRepo
	jwt                 *jwt.JWT
}

func NewChatRoute(
	lingChatService *service.LingChatService,
	conversationService *service.ConversationService,
	characterService service.CharacterService,
	userRepo data.UserRepo,
	jwt *jwt.JWT,
) *ChatRoute {
	return &ChatRoute{
		lingChatService:     lingChatService,
		conversationService: conversationService,
		characterService:    characterService,
		userRepo:            userRepo,
		jwt:                 jwt,
	}
}

func (c *ChatRoute) RegisterRoute(r *gin.RouterGroup) {
	rg := r.Group("/v1/chat")
	{
		rg.POST("/completion", middleware.TokenAuth(false, c.jwt, c.userRepo), c.chatCompletion)
		rg.GET("/history/list", middleware.TokenAuth(false, c.jwt, c.userRepo), c.listConversations)
		rg.GET("/history/detail", middleware.TokenAuth(false, c.jwt, c.userRepo), c.getChatHistory)
		rg.PUT("/history", middleware.TokenAuth(false, c.jwt, c.userRepo), c.loadChatHistory)

		// 角色相关路由
		rg.GET("/character/get_all_characters", c.getAllCharacters)
		rg.GET("/character/avatar/:avatar_file", c.getCharacterAvatar)
	}
}

func (c *ChatRoute) chatCompletion(ctx *gin.Context) {
	var req request.ChatCompletionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "请求格式错误: " + err.Error(),
		})
		return
	}

	resp, err := c.lingChatService.LingChat(ctx, req.Message, req.ConversationID, req.PrevMessageID, req.CharacterID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "处理聊天请求失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"data": response.CompletionResponse{
			ConversationID: resp.ConversationID,
			MessageID:      resp.MessageID,
			Messages:       resp.Messages,
		},
	})
}

func (c *ChatRoute) listConversations(ctx *gin.Context) {
	// 获取分页参数
	page := ctx.DefaultQuery("page", "1")
	pageSize := ctx.DefaultQuery("page_size", "10")

	pageNum, err := strconv.Atoi(page)
	if err != nil || pageNum < 1 {
		pageNum = 1
	}

	pageSizeNum, err := strconv.Atoi(pageSize)
	if err != nil || pageSizeNum < 1 {
		pageSizeNum = 10
	}

	// 调用Service层处理业务逻辑
	result, err := c.conversationService.ListConversations(ctx, pageNum, pageSizeNum)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"data": result,
	})
}

func (c *ChatRoute) getChatHistory(ctx *gin.Context) {
	// 获取conversationID参数
	conversationID := ctx.Query("conversation_id")

	// 获取分页参数
	page := ctx.DefaultQuery("page", "1")
	pageSize := ctx.DefaultQuery("page_size", "20")

	pageNum, err := strconv.Atoi(page)
	if err != nil || pageNum < 1 {
		pageNum = 1
	}

	pageSizeNum, err := strconv.Atoi(pageSize)
	if err != nil || pageSizeNum < 1 {
		pageSizeNum = 20
	}

	// 调用Service层获取会话详情
	result, err := c.conversationService.GetConversationDetail(ctx, conversationID, pageNum, pageSizeNum)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"data": result,
	})
}

func (c *ChatRoute) loadChatHistory(ctx *gin.Context) {
	var messages []openai.ChatCompletionMessage
	if err := ctx.ShouldBindJSON(&messages); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的聊天历史记录格式: " + err.Error(),
		})
		return
	}

	resp := c.lingChatService.LoadChatHistory(ctx, messages)

	ctx.JSON(http.StatusOK, resp)
}

// getAllCharacters 获取所有角色列表
func (c *ChatRoute) getAllCharacters(ctx *gin.Context) {
	// 通过服务层获取所有角色
	characters, err := c.characterService.GetAllCharacters(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  "获取角色列表失败: " + err.Error(),
		})
		return
	}

	// 如果没有角色，返回空列表
	if len(characters) == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"data":    []response.CharacterItem{},
			"message": "未找到任何角色",
		})
		return
	}

	// 转换为响应格式
	var characterItems []response.CharacterItem
	for _, char := range characters {
		characterItems = append(characterItems, response.CharacterItem{
			CharacterID: char.CharacterID,
			Title:       char.Title,
			Info:        char.Info,
			AvatarPath:  char.AvatarPath,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"data": characterItems,
	})
}

func (c *ChatRoute) getCharacterAvatar(ctx *gin.Context) {
	// 获取路径参数
	avatarFile := ctx.Param("avatar_file")
	if avatarFile == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "avatar_file parameter is required",
		})
		return
	}

	// 获取查询参数
	characterID := ctx.Query("character_id")
	if characterID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "character_id query parameter is required",
		})
		return
	}

	// 通过service层获取完整的文件路径
	fullPath, err := c.characterService.GetCharacterAvatarPath(ctx, characterID, avatarFile)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"code": http.StatusNotFound,
			"msg":  err.Error(),
		})
		return
	}

	// 返回文件
	ctx.File(fullPath)
}
