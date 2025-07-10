import EventBus from "./event-bus.js";
import conversationState from "./conversation-state.js";

export class ChatClient {
  constructor(baseUrl) {
    this.baseUrl = baseUrl;
    this.apiUrl = `${baseUrl}/api/v1/chat/completion`;
    this.isConnected = true;
    this.eventListeners = {
      message: [],
      error: [],
      close: [],
      open: []
    };

    // 模拟连接打开
    setTimeout(() => {
      this.eventListeners.open.forEach((cb) => cb());
      EventBus.emit("connection:open");
    }, 100);
  }

  addEventListener(event, callback) {
    if (this.eventListeners[event]) {
      this.eventListeners[event].push(callback);
    }
  }

  removeEventListener(event, callback) {
    if (this.eventListeners[event]) {
      const index = this.eventListeners[event].indexOf(callback);
      if (index > -1) {
        this.eventListeners[event].splice(index, 1);
      }
    }
  }

  // 向后兼容的方法
  onmessage(callback) {
    this.eventListeners.message.push(callback);
  }

  onopen(callback) {
    this.eventListeners.open.push(callback);
  }

  onclose(callback) {
    this.eventListeners.close.push(callback);
  }

  onerror(callback) {
    this.eventListeners.error.push(callback);
  }

  async send(data) {
    try {
      // 从全局状态获取对话ID和消息ID
      const requestData = {
        message: data.content,
        conversation_id: conversationState.getConversationId(),
        prev_message_id: conversationState.getMessageId()
      };

      const response = await fetch(this.apiUrl, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(requestData)
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const result = await response.json();
      
      // 更新全局状态
      if (result.data) {
        conversationState.setConversationId(result.data.conversation_id);
        conversationState.setMessageId(result.data.message_id);
        
        // AI回复返回，结束思考状态，启用按钮
        EventBus.emit("chat:thinking", false);
        
        // 模拟WebSocket消息事件，为每个消息分别触发
        result.data.messages.forEach(message => {
          const messageEvent = {
            data: JSON.stringify(message)
          };
          this.eventListeners.message.forEach((cb) => cb(messageEvent));
        });
      }

    } catch (error) {
      console.error('发送消息失败:', error);
      this.eventListeners.error.forEach((cb) => cb(error));
      EventBus.emit("connection:error", error);
      throw error;
    }
  }

  // 兼容性方法
  close() {
    this.isConnected = false;
    this.eventListeners.close.forEach((cb) => cb());
  }
}

// 保持向后兼容性，添加ChatSocket别名
export const ChatSocket = ChatClient;

