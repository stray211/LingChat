import { DOM } from "../../ui/dom.js";
import { DomUtils } from "../../utils/dom-utils.js";
import request from "../../core/request.js";

export class CharacterController {
  constructor(ui_controller) {
    this.ui_controller = ui_controller;
    this.characterList = document.getElementById("character-list");
    this.domUtils = DomUtils;
    this.userId = 1;
    
    // 设置默认角色数据
    this.defaultCharacters = [
      {
        id: 1,
        title: "默认角色",
        info: "系统默认角色，后端服务连接后将显示更多角色",
        avatar: "../pictures/characters/default.png"
      }
    ];
    
    this.init();
  }

  init() {
    // 使用默认角色数据渲染界面
    this.renderCharacters(this.defaultCharacters);
    this.bindEvents();
    this.updateSelectedStatus(this.ui_controller.character_id);
  }

  updateSelectedStatus(characterId) {
    document.querySelectorAll(".character-select-btn").forEach((btn) => {
      btn.classList.remove("selected");
      btn.textContent = "选择"; // 重置按钮文字

      if (btn.dataset.characterId === String(characterId)) {
        btn.classList.add("selected");
        btn.textContent = "✓ 已选择"; // 更新选中按钮文字
      }
    });
  }

  bindEvents() {
    // 图片菜单点击事件
    DOM.menuCharacter?.addEventListener("click", () =>
      this.showCharacterPanel()
    );

    DOM.character.refreshCharactersBtn?.addEventListener("click", () =>
      this.refreshCharacters()
    );

    DOM.character.openWebBtn?.addEventListener("click", () =>
      this.openCreativeWeb()
    );

    document.querySelectorAll(".character-select-btn").forEach((btn) => {
      btn.addEventListener("click", () => {
        this.selectCharacter(btn.dataset.characterId);
      });
    });
  }

  showCharacterPanel() {
    this.domUtils.showElements([DOM.menuCharacter, DOM.characterPage]);
    this.domUtils.hideElements(
      this.domUtils.getOtherPanelElements([
        DOM.menuCharacter,
        DOM.characterPage,
      ])
    );
  }

  async openCreativeWeb() {
    try {
      const response = await fetch("/api/v1/chat/character/open_web");
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();
      return data;
    } catch (error) {
      alert("启动失败，请手动去lingchat的discussion网页");
      console.error("刷新失败:", error);
      throw error;
    }
  }

  async refreshCharacters() {
    try {
      const response = await fetch(
        "/api/v1/chat/character/refresh_characters",
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
        }
      );

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();
      alert("刷新成功");
      
      // 刷新成功后重新加载角色列表
      this.loadCharacters();
      
      return data;
    } catch (error) {
      alert("刷新失败");
      console.error("刷新失败:", error);
      throw error;
    }
  }

  // 新增方法：手动加载角色列表（用于刷新后）
  async loadCharacters() {
    try {
      const characters = await this.fetchCharacters();
      this.renderCharacters(characters);
      this.bindEvents();
      this.updateSelectedStatus(this.ui_controller.character_id);
    } catch (error) {
      console.error("加载角色列表失败:", error);
      // 失败时继续使用默认角色
      this.renderCharacters(this.defaultCharacters);
      this.bindEvents();
    }
  }

  async fetchCharacters() {
    try {
      const list = await request.characterGetAll();
      return list.map((char) => ({
        id: char.character_id,
        title: char.title,
        info: char.info || "暂无角色描述",
        // 使用新的文件获取接口
        avatar: char.avatar_path
          ? `/api/v1/chat/character/character_file/${encodeURIComponent(
              char.avatar_path
            )}`
          : "../pictures/characters/default.png",
      }));
    } catch (error) {
      console.error("获取角色列表失败:", error);
      return this.defaultCharacters;
    }
  }

  renderCharacters(characters) {
    this.characterList.innerHTML = characters
      .map((char) => this.createCharacterCard(char))
      .join("");
  }

  createCharacterCard(character) {
    return `
      <div class="character-card">
        <div class="character-avatar-container">
          <img src="${character.avatar}" alt="${character.title}" class="character-avatar">
        </div>
        <div class="character-content">
          <div class="character-title">${character.title}</div>
          <div class="character-description">${character.info}</div>
          <button class="character-select-btn" data-character-id="${character.id}">选择</button>
        </div>
      </div>
    `;
  }

  async selectCharacter(characterId) {
    return request
      .characterSelect(this.userId, characterId)
      .then((character) => {
        this.updateSelectedStatus(characterId);
        return this.ui_controller.getAndApplyAIInfo();
      })
      .catch((error) => {
        console.error("切换角色失败:", error);
      });
  }
}
