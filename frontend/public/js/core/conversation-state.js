// 全局对话状态管理器
class ConversationState {
  constructor() {
    this.conversationId = '';
    this.messageId = '';
    this.characterId = 'noiqingling'; // 默认角色ID
    this.listeners = [];
  }

  // 设置对话ID
  setConversationId(id) {
    this.conversationId = id;
    this.notifyListeners();
  }

  // 设置消息ID
  setMessageId(id) {
    this.messageId = id;
    this.notifyListeners();
  }

  // 设置角色ID
  setCharacterId(id) {
    this.characterId = id;
    this.notifyListeners();
  }

  // 开始新对话
  startNewConversation() {
    this.conversationId = '';
    this.messageId = '';
    // 注意：角色ID通常不会在新对话时重置，保持当前选择的角色
    this.notifyListeners();
  }

  // 获取当前对话ID
  getConversationId() {
    return this.conversationId;
  }

  // 获取当前消息ID
  getMessageId() {
    return this.messageId;
  }

  // 获取当前角色ID
  getCharacterId() {
    return this.characterId;
  }

  // 添加状态监听器
  addListener(callback) {
    this.listeners.push(callback);
  }

  // 移除状态监听器
  removeListener(callback) {
    const index = this.listeners.indexOf(callback);
    if (index > -1) {
      this.listeners.splice(index, 1);
    }
  }

  // 通知所有监听器
  notifyListeners() {
    this.listeners.forEach(callback => {
      callback({
        conversationId: this.conversationId,
        messageId: this.messageId,
        characterId: this.characterId
      });
    });
  }
}

// 创建全局实例
const conversationState = new ConversationState();

export default conversationState; 