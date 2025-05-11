package ws

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"

	"LingChat/api/routes/ws/types"
)

// WebSocketHandler 管理 WebSocket 连接
type WebSocketHandler struct {
	handler types.MessageHandler
}

// NewWebSocketHandler 创建新的 WebSocket 服务器
func NewWebSocketHandler(handler types.MessageHandler) *WebSocketHandler {
	return &WebSocketHandler{
		handler: handler,
	}
}

// ServeHTTP 实现 http.Handler 接口
func (s *WebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.HandleWebSocket(w, r)
}

func (s *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 将 HTTP 连接升级为 WebSocket
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true }, // 允许所有来源
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket升级错误: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("新的WebSocket连接已建立: %s", r.RemoteAddr)

	for {
		// 读取消息
		_, rawMessage, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket连接异常关闭: %v", err)
			} else {
				log.Println("读取错误:", err)
			}
			break
		}

		// 处理消息
		rawResp, err := s.handler(rawMessage)
		if err != nil {
			log.Printf("消息处理错误: %v", err)
			errorResp := types.Response{
				Type:  "error",
				Error: err.Error(),
			}
			errorJSON, _ := json.Marshal(errorResp)
			if err := conn.WriteMessage(websocket.TextMessage, errorJSON); err != nil {
				log.Printf("发送错误响应失败: %v", err)
				break
			}
			continue
		}

		// 发送响应
		for _, msg := range rawResp {
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Printf("发送响应失败: %v", err)
				break
			}
		}
	}

	log.Printf("WebSocket连接已关闭: %s", r.RemoteAddr)
}
