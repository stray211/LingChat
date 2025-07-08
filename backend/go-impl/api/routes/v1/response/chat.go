package response

import (
	"time"

	"LingChat/api/routes/ws/types"
)

type CompletionResponse struct {
	ConversationID string           `json:"conversation_id"`
	MessageID      string           `json:"message_id"`
	Messages       []types.Response `json:"messages"`
}

type ConversationItem struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	LatestMessageID string    `json:"latest_message_id,omitempty"`
}

type ConversationListResponse struct {
	Conversations []ConversationItem `json:"conversations"`
	Total         int                `json:"total"`
}

type ConversationMessageItem struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	Role           string    `json:"role"`
	Content        string    `json:"content"`
	Model          string    `json:"model,omitempty"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

type ConversationDetailResponse struct {
	Conversation ConversationItem          `json:"conversation"`
	Messages     []ConversationMessageItem `json:"messages"`
	Total        int                       `json:"total"`
}

// CharacterItem 角色列表项
type CharacterItem struct {
	CharacterID string `json:"character_id"`
	Title       string `json:"title"`
	Info        string `json:"info"`
	AvatarPath  string `json:"avatar_path"`
}

// CharacterListResponse 角色列表响应
type CharacterListResponse struct {
	Characters []CharacterItem `json:"data"`
	Message    string          `json:"message,omitempty"`
}

// CharacterInfoResponse 角色信息响应
type CharacterInfoResponse struct {
	AIName          string `json:"ai_name"`
	AISubtitle      string `json:"ai_subtitle"`
	UserName        string `json:"user_name"`
	UserSubtitle    string `json:"user_subtitle"`
	CharacterID     string `json:"character_id"`
	ThinkingMessage string `json:"thinking_message"`
	Scale           string `json:"scale"`
	Offset          string `json:"offset"`
	BubbleTop       string `json:"bubble_top"`
	BubbleLeft      string `json:"bubble_left"`
}

// BackgroundItem 背景项
type BackgroundItem struct {
	ImagePath    string  `json:"image_path"`
	Title        string  `json:"title"`
	ModifiedTime float64 `json:"modified_time"`
}

// BackgroundListResponse 背景列表响应
type BackgroundListResponse struct {
	Data    []BackgroundItem `json:"data"`
	Message string           `json:"message,omitempty"`
}
