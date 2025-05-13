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
	userRepo            data.UserRepo
	jwt                 *jwt.JWT
}

func NewChatRoute(
	lingChatService *service.LingChatService,
	conversationService *service.ConversationService,
	userRepo data.UserRepo,
	jwt *jwt.JWT,
) *ChatRoute {
	return &ChatRoute{
		lingChatService:     lingChatService,
		conversationService: conversationService,
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

	resp, err := c.lingChatService.LingChat(ctx, req.Message, req.ConversationID, req.PrevMessageID)
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
