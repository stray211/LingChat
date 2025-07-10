import EventBus from "../../core/event-bus.js";
import { DOM } from "../../ui/dom.js";
import { EMOTION_CONFIG } from "../emotion/config.js";

export class ChatManager {
  constructor({ connection, historyManager }) {
    this.connection = connection;
    this.messageQueue = [];
    this.currentMessagePart = null;
    this.isProcessing = false;
    this.isWaitingForResponse = false;
    this.historyManager = historyManager;
    this.setupSocketHandlers();
    this.setupEventListeners();
  }

  /**
   * 验证并修复emotion字段
   * @param {string} emotion - 原始emotion值
   * @param {string} originalTag - 原始标签作为fallback
   * @returns {string} - 修复后的emotion值
   */
  validateEmotion(emotion, originalTag) {
    // 检查emotion是否为空、undefined、null或者"unknown"
    if (!emotion || emotion === "unknown" || emotion.trim() === "") {
      console.warn(`收到无效的emotion值: "${emotion}", 使用originalTag作为fallback: "${originalTag}"`);
      return originalTag || "正常";
    }

    // 检查emotion是否在配置中存在
    if (!EMOTION_CONFIG[emotion]) {
      console.warn(`收到未知的emotion值: "${emotion}", 使用originalTag作为fallback: "${originalTag}"`);
      return originalTag || "正常";
    }

    return emotion;
  }

  setupSocketHandlers() {
    this.connection.onmessage((event) => {
      try {
        const data = JSON.parse(event.data);

        if (data.type === "reply") {
          if (data.isMultiPart) {
            this.handleMultiPartMessage(data);
          } else {
            this.handleSingleMessage(data);
          }
        }
      } catch (e) {
        console.error("消息解析错误:", e, "原始数据:", event.data);
        EventBus.emit("chat:error", { error: e, rawData: event.data });
      }
    });
  }

  setupEventListeners() {
    EventBus.on("ui:send-message", (text) => {
      this.sendOrContinue(text);
    });

    EventBus.on("ui:continue", () => {
      this.handleContinue();
    });

    EventBus.on("chat:enable-input", () => {
      this.enableInput();
    });
  }

  handleMultiPartMessage(data) {
    this.messageQueue.push(data);

    if (data.partIndex === 0 && !this.isProcessing) {
      this.processNextMessage();
    }
  }

  handleSingleMessage(data) {
    console.log("【处理单条消息】", data);
    
    // 验证并修复emotion字段
    const validatedEmotion = this.validateEmotion(data.emotion, data.originalTag);
    
    EventBus.emit("chat:message", {
      content: data.message || data.content,
      emotion: validatedEmotion,
      originalTag: data.originalTag,
      audioFile: data.audioFile,
      isFinal: true,
    });
  }

  processNextMessage() {
    if (this.messageQueue.length === 0) {
      this.isProcessing = false;
      return;
    }

    this.isProcessing = true;
    this.currentMessagePart = this.messageQueue.shift();

    let displayText = "";
    if (!this.currentMessagePart) {
      displayText = ""; // 处理 null 情况
    } else {
      displayText =
        this.currentMessagePart.motionText &&
        this.currentMessagePart.motionText !== ""
          ? `${this.currentMessagePart.message}（${this.currentMessagePart.motionText}）`
          : this.currentMessagePart.message;
    }

    this.historyManager.addMessage(
      null,
      displayText,
      this.currentMessagePart.partIndex ===
        this.currentMessagePart.totalParts - 1
    );

    // 验证并修复emotion字段
    const validatedEmotion = this.validateEmotion(
      this.currentMessagePart.emotion, 
      this.currentMessagePart.originalTag
    );

    EventBus.emit("chat:message", {
      content: this.currentMessagePart.message,
      emotion: validatedEmotion,
      originalTag: this.currentMessagePart.originalTag,
      audioFile: this.currentMessagePart.audioFile,
      isFinal:
        this.currentMessagePart.partIndex ===
        this.currentMessagePart.totalParts - 1,
      motionText: this.currentMessagePart.motionText,
    });
  }

  sendOrContinue(text) {
    if (this.isProcessing) {
      this.handleContinue();
    } else if (text) {
      this.sendMessage(text);
    }
  }

  sendMessage(text) {
    if (!text.trim()) return;

    const message = {
      type: "message",
      content: text,
    };

    EventBus.emit("chat:thinking", true);

    // 先确保音频开始播放再发送消息
    setTimeout(() => {
      this.connection.send(message);
    }, 600); // 300ms延迟确保音频已启动
    this.historyManager.addMessage(text, null, false);

    // 更新状态
    this.isWaitingForResponse = true;
    this.messageQueue = [];
    this.isProcessing = false;
  }

  handleContinue() {
    // 检查是否是最后一条消息
    if (
      this.currentMessagePart &&
      this.currentMessagePart.partIndex ===
        this.currentMessagePart.totalParts - 1
    ) {
      this.resetConversationState();
      EventBus.emit("chat:conversation-end");
    } else {
      this.processNextMessage();
    }
  }

  enableInput() {
    this.isWaitingForResponse = false;
    EventBus.emit("chat:input-enabled");
  }

  resetConversationState() {
    this.currentMessagePart = null;
    this.messageQueue = [];
    this.isProcessing = false;
    this.isWaitingForResponse = false;

    EventBus.emit("chat:reset");
    EventBus.emit("chat:enable-input");
  }
}
