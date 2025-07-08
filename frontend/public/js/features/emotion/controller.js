import { EMOTION_CONFIG } from "./config.js";
import { DOM } from "../../ui/dom.js";
import EventBus from "../../core/event-bus.js";
import conversationState from "../../core/conversation-state.js";

export class EmotionController {
  constructor(uiController) {
    this.uiController = uiController;
    this.currentEmotion = "normal";
    this.animationEndHandler = this._handleAnimationEnd.bind(this);
    DOM.avatar.container.addEventListener(
      "animationend",
      this.animationEndHandler
    );
  }

  /**
   * 切换表情
   * @param {string} emotion - 表情标识符
   * @param {object} [options] - 额外选项
   * @param {boolean} [options.force=false] - 是否强制重新播放
   */
  setEmotion(emotion, { force = false } = {}) {
    if (!EMOTION_CONFIG[emotion]) {
      console.log(`未知表情: ${emotion}`);
      return;
    }

    if (this.currentEmotion === emotion && !force) return;

    const config = EMOTION_CONFIG[emotion];
    this._clearCurrentEffects();

    // 更新头像图片（除了AI思考状态）
    if (emotion !== "AI思考") {
      try {
        // 显示加载状态
        DOM.avatar.img.classList.add("loading");

        // 创建新的Image对象预加载
        const img = new Image();
        img.onload = () => {
          // 加载完成后替换src
          DOM.avatar.img.src = img.src;
          DOM.avatar.img.classList.remove("loading");
        };
        img.onerror = () => {
          console.log(`加载头像失败: ${emotion}.png`);
          DOM.avatar.img.classList.remove("loading");
          // 设置默认头像
          const characterId = conversationState.getCharacterId();
          DOM.avatar.img.src = `/api/v1/chat/character/avatar/正常.png?character_id=${characterId}`;
        };

        // 构建新的avatar URL
        const characterId = conversationState.getCharacterId();
        const avatarFile = `${emotion}.png`;
        const timestamp = Date.now();
        img.src = `/api/v1/chat/character/avatar/${avatarFile}?character_id=${characterId}&t=${timestamp}`;
      } catch (error) {
        console.error("设置表情头像时出错:", error);
        DOM.avatar.img.classList.remove("loading");
      }
    }

    // 设置动画
    if (config.animation !== "none") {
      DOM.avatar.container.classList.remove(this.currentEmotion);
      DOM.avatar.container.classList.add(config.animation);
    }

    // 触发气泡效果
    if (config.bubbleImage !== "none") {
      this._showBubbleEffect(config);
    }

    // 播放音效
    if (config.audio !== "none") {
      EventBus.emit("audio:play-effect", {
        type: "emotion",
        src: config.audio,
      });
      // 设置音频源并播放
      DOM.bubbleAudio.src = config.audio;
      DOM.bubbleAudio.load(); // 重新加载音频
      DOM.bubbleAudio.play();
    }

    this.currentEmotion = emotion;
    EventBus.emit("emotion:changed", emotion);
  }

  _clearCurrentEffects() {
    // 移除所有动画类
    Object.values(EMOTION_CONFIG).forEach(({ animation }) => {
      if (animation !== "none") {
        DOM.avatar.container.classList.remove(animation);
      }
    });
  }

  _showBubbleEffect(config) {
    const version = Date.now();

    if (config.bubbleImage === "none") return;

    // 重置动画
    DOM.avatar.bubble.classList.remove("show");
    void DOM.avatar.bubble.offsetWidth;

    // 设置带时间戳的新源
    DOM.avatar.bubble.style.backgroundImage = `url(${config.bubbleImage}??t=${version}#t=0.1)`;

    // 添加对应的表情类
    DOM.avatar.bubble.classList.add(config.bubbleClass);

    // 触发显示
    DOM.avatar.bubble.classList.add("show");

    // 动画结束后隐藏
    setTimeout(() => {
      DOM.avatar.bubble.classList.remove("show");
      DOM.avatar.bubble.classList.remove(config.bubbleClass);
    }, 2000); // 2秒后执行
  }

  _handleAnimationEnd() {
    this._clearCurrentEffects();
    DOM.avatar.container.classList.add("normal");
  }

  destroy() {
    DOM.avatar.container.removeEventListener(
      "animationend",
      this.animationEndHandler
    );
  }
}
