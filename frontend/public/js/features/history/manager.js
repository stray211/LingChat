import { DOM } from "../../ui/dom.js";
import { DomUtils } from "../../utils/dom-utils.js";
import EventBus from "../../core/event-bus.js";
import conversationState from "../../core/conversation-state.js";
import request from "../../core/request.js";
import { DateUtils } from "../../utils/date-utils.js";

export class HistoryManager {
  constructor() {
    this.ai_name = "默认";
    this.domUtils = DomUtils;
    this.init();
  }

  init() {
    this.bindEvents();
  }

  bindEvents() {
    if (!DOM.history.clearBtn) return;

    DOM.history.toggle.addEventListener(
      "click",
      this.toggleHistoryPanel.bind(this)
    );

    DOM.history.clearBtn.addEventListener(
      "click",
      this.startNewConversation.bind(this)
    );

    EventBus.on("ui:name-updated", (names) => {
      this.ai_name = names.ai_name;
    });
  }

  startNewConversation() {
    conversationState.startNewConversation();
    this.render([]);
    DOM.history.list.innerHTML = "<p>已开启新对话，请开始聊天</p>";
  }

  async toggleHistoryPanel() {
    this.domUtils.showElements([
      DOM.menuContent,
      DOM.history.toggle,
      DOM.history.content,
      DOM.history.clearBtn,
    ]);

    this.domUtils.hideElements(
      this.domUtils.getOtherPanelElements([
        DOM.history.toggle,
        DOM.history.content,
        DOM.history.clearBtn,
      ])
    );

    await this.loadHistoryDetail();
  }

  async loadHistoryDetail() {
    const currentConversationId = conversationState.getConversationId();
    
    if (!currentConversationId) {
      this.render([]);
      return;
    }

    try {
      const historyData = await request.historyDetail(currentConversationId);
      this.render(historyData.messages || []);
    } catch (error) {
      console.error('加载历史记录失败:', error);
      DOM.history.list.innerHTML = "<p>加载历史记录失败</p>";
    }
  }

  render(messages = []) {
    if (!DOM.history.list) return;

    if (messages.length === 0) {
      DOM.history.list.innerHTML = "<p>暂无历史记录</p>";
      return;
    }

    DOM.history.list.innerHTML = messages
      .map(
        (message) => `
          <div class="history-item">
            <p><strong>${message.sender || '用户'}:</strong> ${message.content}</p>
            <p><small>时间: ${DateUtils.parseAndFormat(message.created_at)}</small></p>
          </div>
        `
      )
      .join("");
  }

  addMessage(userMessage, aiResponse, isFinal) {
    if (isFinal) {
      EventBus.emit("history:message-added", {
        userMessage,
        aiResponse,
        timestamp: new Date().toISOString()
      });
    }
  }
}
