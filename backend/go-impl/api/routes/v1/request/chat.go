package request

type ChatCompletionRequest struct {
	Message        string `json:"message" binding:"required"`
	ConversationID string `json:"conversation_id"`
	PrevMessageID  string `json:"prev_message_id"`
	CharacterID    string `json:"character_id"`
}
