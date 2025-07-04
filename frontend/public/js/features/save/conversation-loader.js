import { DOM } from "../../ui/dom.js";
import request from "../../core/request.js";
import conversationState from "../../core/conversation-state.js";
import { DateUtils } from "../../utils/date-utils.js";

export class ConversationLoader {
  constructor(containerId, userId, options = {}) {
    this.containerId = containerId;
    this.userId = userId;
    this.options = {
      pageSize: options.pageSize || 10,
      ...options
    };
    this.conversations = [];
    this.onConversationLoaded = null;
    this.onLoadError = null;
    this.bindUploadEvents();
  }

  bindUploadEvents() {
    const uploadBtn = document.getElementById("upload-btn");
    if (uploadBtn) {
      uploadBtn.addEventListener("click", () => this.uploadUserLogFile());
    }
  }

  async loadConversations() {
    try {
      const conversations = await request.historyList(
        this.userId,
        1,
        this.options.pageSize
      );
      this.conversations = conversations;
      this.render();
    } catch (error) {
      console.error("加载对话列表失败:", error);
      this.renderError("加载对话列表失败");
    }
  }

  render() {
    const container = document.getElementById(this.containerId);
    if (!container) return;

    if (this.conversations.length === 0) {
      container.innerHTML = "<p>暂无对话记录</p>";
      return;
    }

    container.innerHTML = this.conversations
      .map(
        (convo) => `
          <div class="conversation-item">
            <div class="conversation-info">
              <div class="conversation-title">${convo.title || "未命名对话"}</div>
              <div class="conversation-date">${DateUtils.parseAndFormat(convo.updated_at)}</div>
            </div>
            <div class="conversation-actions">
              <button class="action-btn load-btn" data-id="${convo.id}">
                选择对话
              </button>
              <button class="action-btn delete-btn" data-id="${convo.id}">
                删除
              </button>
            </div>
          </div>
        `
      )
      .join("");

    // 绑定事件
    this.bindItemEvents();
  }

  bindItemEvents() {
    const container = document.getElementById(this.containerId);
    if (!container) return;

    // 绑定选择对话按钮事件
    container.querySelectorAll(".load-btn").forEach((btn) => {
      btn.addEventListener("click", async (e) => {
        const conversationId = e.target.dataset.id;
        await this.selectConversation(conversationId);
      });
    });

    // 绑定删除按钮事件
    container.querySelectorAll(".delete-btn").forEach((btn) => {
      btn.addEventListener("click", async (e) => {
        const conversationId = e.target.dataset.id;
        if (confirm("确定要删除这个对话吗？")) {
          await this.deleteUserConversation(conversationId);
        }
      });
    });
  }

  async selectConversation(conversationId) {
    try {
      // 仅设置ConversationId到全局状态
      conversationState.setConversationId(conversationId);
      conversationState.setMessageId(''); // 重置messageId
      
      console.log(`已选择对话: ${conversationId}`);
      
      if (this.onConversationLoaded) {
        this.onConversationLoaded([]);
      }
    } catch (error) {
      console.error("选择对话失败:", error);
      if (this.onLoadError) {
        this.onLoadError(error);
      }
    }
  }

  async deleteUserConversation(conversationId) {
    try {
      await request.historyDelete(this.userId, conversationId);
      console.log(`对话 ${conversationId} 删除成功`);
      
      // 刷新对话列表
      await this.loadConversations();
    } catch (error) {
      console.error("删除对话失败:", error);
      alert("删除对话失败: " + error.message);
    }
  }

  renderError(message) {
    const container = document.getElementById(this.containerId);
    if (container) {
      container.innerHTML = `<p style="color: red;">${message}</p>`;
    }
  }

  async uploadUserLogFile() {
    console.log("触发了点击事件");
    const fileInput = document.getElementById("log-upload");
    const statusDiv = document.getElementById("upload-status");

    if (!fileInput.files.length) {
      statusDiv.textContent = "请先选择文件";
      return;
    }

    const file = fileInput.files[0];

    try {
      statusDiv.textContent = "正在处理文件...";

      // 读取文件内容
      const text = await file.text();

      // 发送到后端处理
      const result = await request.histortyInput(this.userId, text, file.name);
      statusDiv.textContent = `成功导入 ${result.processed_count} 条对话记录`;
      // 刷新对话列表
      this.loadConversations();
    } catch (error) {
      statusDiv.textContent = `上传失败: ${error}`;
      console.error("上传错误:", error);
    }
  }
}
