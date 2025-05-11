package response

import (
	"LingChat/api/routes/ws/types"
)

type CompletionResponse struct {
	ConversationID string           `json:"conversation_id"`
	MessageID      string           `json:"message_id"`
	Messages       []types.Response `json:"messages"`
}
