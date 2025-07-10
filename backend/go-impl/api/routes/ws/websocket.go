package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"LingChat/api/routes/common"
	"LingChat/api/routes/ws/types"
	"LingChat/internal/service"
)

// WebSocketEngine 管理多个 WebSocket Handler
type WebSocketEngine struct {
	LingChatService *service.LingChatService
}

// NewWebSocketEngine 创建新的 WebSocket 服务器
func NewWebSocketEngine(lingChatService *service.LingChatService) *WebSocketEngine {
	return &WebSocketEngine{
		LingChatService: lingChatService,
	}
}

func (s *WebSocketEngine) LingChatHandler(rawMsg []byte) ([]types.RawResponse, error) {
	var msg types.Message
	err := json.Unmarshal(rawMsg, &msg)
	if err != nil {
		err = fmt.Errorf("JSON 解析错误: %w", err)
		log.Println(err)
		return nil, err
	}

	switch msg.Type {
	case "message":
	case "handshake":
		log.Printf("handshake with message:\"%s\"\n", msg.Content)
		return nil, nil
	case "ping":
		log.Println("Ping received, Pong")
		return nil, nil
	default:
		return nil, fmt.Errorf("invalid type \"%s\" with message: \"%s\"", msg.Type, msg.Content)
	}

	ctx, cancel := context.WithTimeout(
		context.WithValue(
			context.Background(),
			common.UseLegacyTempChatContextKey, true,
		),
		2*time.Minute,
	)
	defer cancel()

	// WebSocket固定使用默认character_id
	resp, err := s.LingChatService.LingChat(ctx, msg.Content, "", "", "noiqingling")
	if err != nil {
		err = fmt.Errorf("LingChat error: %w", err)
		log.Println(err)
		return nil, err
	}

	var respSentences []types.RawResponse
	for _, message := range resp.Messages {
		msgJSON, err := json.Marshal(message)
		if err != nil {
			err = fmt.Errorf("JSON 序列化错误: %w", err)
			log.Println(err)
		}
		respSentences = append(respSentences, msgJSON)
	}

	return respSentences, nil
}

func (s *WebSocketEngine) TestHandler(rawMsg []byte) ([]types.RawResponse, error) {
	var msg types.Message
	err := json.Unmarshal(rawMsg, &msg)
	if err != nil {
		err = fmt.Errorf("JSON 解析错误: %w", err)
		log.Println(err)
		return nil, err
	}

	// // 使用注入的处理器处理消息
	// response := []byte{}
	// if err != nil {
	// 	err = fmt.Errorf("处理错误: %w", err)
	// 	log.Println(err)
	// 	return nil, err
	// }
	fmt.Println(msg)

	responseJSON, err := json.Marshal(msg)
	if err != nil {
		err = fmt.Errorf("JSON 序列化错误: %w", err)
		log.Println(err)
		return nil, err
	}
	return []types.RawResponse{responseJSON}, nil
}
