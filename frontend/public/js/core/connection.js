import EventBus from "./event-bus.js";

export class ChatClient {
  constructor(baseUrl) {
    this.baseUrl = baseUrl || '';
    this.apiUrl = `${this.baseUrl}/api/v1/chat/completion`;
    this.conversationId = null;
    this.messageId = null;
    this.isConnected = true;

    // 保持与WebSocket相同的事件监听器接口
    this.eventListeners = {
      message: [],
      open: [],
      close: [],
      error: [],
    };

    // 模拟连接打开
    setTimeout(() => {
      this.eventListeners.open.forEach((cb) => cb());
      EventBus.emit("connection:open");
    }, 100);
  }

  async send(data) {
    try {
      // 转换消息格式：WebSocket格式 -> HTTP格式
      const requestData = {
        message: data.content,
        conversation_id: this.conversationId || "",
        prev_message_id: this.messageId || ""
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
      
      // 更新会话状态
      if (result.data) {
        this.conversationId = result.data.conversation_id;
        this.messageId = result.data.message_id;
        
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

  // 保持与WebSocket相同的事件监听接口
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

  // 兼容性方法
  close() {
    this.isConnected = false;
    this.eventListeners.close.forEach((cb) => cb());
  }
}

// 为了向后兼容，保留旧的类名作为别名
export const ChatSocket = ChatClient;

