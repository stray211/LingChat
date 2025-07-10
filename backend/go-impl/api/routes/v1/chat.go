package v1

import (
	"net/http"
	"os"
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
	backgroundService   service.BackgroundService
	userRepo            data.UserRepo
	jwt                 *jwt.JWT
}

func NewChatRoute(
	lingChatService *service.LingChatService,
	conversationService *service.ConversationService,
	characterService service.CharacterService,
	backgroundService service.BackgroundService,
	userRepo data.UserRepo,
	jwt *jwt.JWT,
) *ChatRoute {
	return &ChatRoute{
		lingChatService:     lingChatService,
		conversationService: conversationService,
		characterService:    characterService,
		backgroundService:   backgroundService,
		userRepo:            userRepo,
		jwt:                 jwt,
	}
}

func (cr *ChatRoute) RegisterRoute(r *gin.RouterGroup) {
	rg := r.Group("/v1/chat")
	{
		rg.POST("/completion", middleware.TokenAuth(false, cr.jwt, cr.userRepo), cr.chatCompletion)
		rg.GET("/history/list", middleware.TokenAuth(false, cr.jwt, cr.userRepo), cr.listConversations)
		rg.GET("/history/detail", middleware.TokenAuth(false, cr.jwt, cr.userRepo), cr.getChatHistory)
		rg.PUT("/history", middleware.TokenAuth(false, cr.jwt, cr.userRepo), cr.loadChatHistory)

		// 角色相关路由
		rg.GET("/character/get_all_characters", cr.getAllCharacters)
		rg.GET("/character/avatar/:avatar_file", cr.getCharacterAvatar)
		rg.GET("/character/info", cr.getCharacterInfo)

		// 背景相关路由
		rg.GET("/background/list", cr.listBackgrounds)
		rg.GET("/background/background_file/:background_file", cr.getBackgroundFile)

		// 语音文件状态查询路由
		rg.GET("/voice/status/:filename", middleware.TokenAuth(false, cr.jwt, cr.userRepo), cr.getVoiceFileStatus)
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

// getCharacterInfo 获取角色详细信息
func (c *ChatRoute) getCharacterInfo(ctx *gin.Context) {
	// 获取character_id参数
	characterID := ctx.Query("character_id")
	if characterID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "character_id parameter is required",
		})
		return
	}

	// 通过service层获取角色详细信息
	characterInfo, err := c.characterService.GetCharacterInfo(ctx, characterID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  "获取角色信息失败: " + err.Error(),
		})
		return
	}

	// 返回响应
	ctx.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"data": response.CharacterInfoResponse{
			AIName:          characterInfo.AIName,
			AISubtitle:      characterInfo.AISubtitle,
			UserName:        characterInfo.UserName,
			UserSubtitle:    characterInfo.UserSubtitle,
			CharacterID:     characterInfo.CharacterID,
			ThinkingMessage: characterInfo.ThinkingMessage,
			Scale:           characterInfo.Scale,
			Offset:          characterInfo.Offset,
			BubbleTop:       characterInfo.BubbleTop,
			BubbleLeft:      characterInfo.BubbleLeft,
		},
	})
}

// listBackgrounds 获取背景图片列表
func (c *ChatRoute) listBackgrounds(ctx *gin.Context) {
	// 调用service层获取背景列表
	backgrounds, err := c.backgroundService.ListBackgrounds(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "获取背景列表失败",
		})
		return
	}

	// 如果没有找到背景图片
	if len(backgrounds) == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"data":    []response.BackgroundItem{},
			"message": "背景图片一个都没找到",
		})
		return
	}

	// 转换为响应格式
	var backgroundItems []response.BackgroundItem
	for _, bg := range backgrounds {
		backgroundItems = append(backgroundItems, response.BackgroundItem{
			ImagePath:    bg.ImagePath,
			Title:        bg.Title,
			ModifiedTime: bg.ModifiedTime,
		})
	}

	// 返回背景列表
	ctx.JSON(http.StatusOK, gin.H{
		"data": backgroundItems,
	})
}

func (c *ChatRoute) getBackgroundFile(ctx *gin.Context) {
	// 获取路径参数
	backgroundFile := ctx.Param("background_file")
	if backgroundFile == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "background_file parameter is required",
		})
		return
	}

	// 通过service层获取完整的文件路径
	fullPath, err := c.backgroundService.GetBackgroundFile(ctx, backgroundFile)
	if err != nil {
		if err == os.ErrNotExist {
			ctx.JSON(http.StatusNotFound, gin.H{
				"code":   http.StatusNotFound,
				"detail": "Background not found",
			})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  err.Error(),
			})
		}
		return
	}

	// 返回文件
	ctx.File(fullPath)
}

// getVoiceFileStatus 查询语音文件状态
func (c *ChatRoute) getVoiceFileStatus(ctx *gin.Context) {
	filename := ctx.Param("filename")
	if filename == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "文件名不能为空",
		})
		return
	}

	// 查询文件状态
	status := c.lingChatService.GetVoiceFileStatus(filename)

	ctx.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"data": gin.H{
			"filename":  filename,
			"create_at": status.CreateAt,
			"status":    status.Status,
		},
	})
}
