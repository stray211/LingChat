package response

import (
	"LingChat/api/routes/ws"
)

type CompletionResponse struct {
	ConversationID string        `json:"conversation_id"`
	MessageID      string        `json:"message_id"`
	Messages       []ws.Response `json:"messages"`
}
