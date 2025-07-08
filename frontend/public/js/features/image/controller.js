import { DOM } from "../../ui/dom.js";
import { DomUtils } from "../../utils/dom-utils.js";
import request from "../../core/request.js";
import { KousanManager } from "./kuosan-manager.js";

export class ImageController {
  constructor() {
    this.backgroundList = document.getElementById("background-list");
    this.domUtils = DomUtils;
    this.currentSelectedCard = null; // 跟踪当前选中的卡片
    this.kousanManager = new KousanManager();
    
    // 设置默认背景数据
    this.defaultBackgrounds = [
      {
        title: "默认背景",
        url: "../pictures/backgrounds/default.png",
        time: new Date().toISOString()
      }
    ];
    
    this.init();
  }

  async init() {
    try {
      await this.refreshBackgroundWithCustom();
      this.bindEvents();

      // 从本地存储恢复选中的背景
      const savedBg = localStorage.getItem("selectedBackground");
      if (savedBg) {
        const card = this.findCardByUrl(savedBg);
        if (card) {
          this.selectBackground(card, savedBg);
        }
      } else {
        const randomCard = this.getRandomCard();
        if (randomCard) {
          this.selectBackground(
            randomCard,
            randomCard.querySelector(".background-select-btn").dataset.backgroundUrl
          );
          console.log("已选随机背景");
        }
      }
    } catch (error) {
      console.error("加载背景图片失败", error);
      // 失败时使用默认背景和自定义背景
      const customBackgrounds = JSON.parse(localStorage.getItem('customBackgrounds') || '[]');
      const allBackgrounds = [...customBackgrounds, ...this.defaultBackgrounds];
      this.renderBackgrounds(allBackgrounds);
      this.bindEvents();
    }
  }

  async refreshBackground() {
    const backgrounds = await this.fetchBackgrounds();
    this.renderBackgrounds(backgrounds);
  }

  getRandomCard() {
    const cards = this.backgroundList.querySelectorAll(".background-card");
    if (cards.length === 0) {
      return null; // 如果没有卡片，返回null
    }
    const randomIndex = Math.floor(Math.random() * cards.length);
    return cards[randomIndex];
  }

  findCardByUrl(url) {
    const cards = this.backgroundList.querySelectorAll(".background-card");
    for (const card of cards) {
      const btn = card.querySelector(".background-select-btn");
      if (btn && btn.dataset.backgroundUrl === url) {
        return card;
      }
    }
    return null;
  }

  bindEvents() {
    // 图片菜单点击事件
    DOM.menuImage?.addEventListener("click", () => this.showImagePanel());

    // 添加背景选择事件委托
    this.backgroundList.addEventListener("click", (e) => {
      const selectBtn = e.target.closest(".background-select-btn");
      if (selectBtn) {
        const card = selectBtn.closest(".background-card");
        const bgUrl = selectBtn.dataset.backgroundUrl;
        this.selectBackground(card, bgUrl);
        return;
      }

      // 添加删除按钮事件处理
      const deleteBtn = e.target.closest(".background-delete-btn");
      if (deleteBtn) {
        const bgUrl = deleteBtn.dataset.backgroundUrl;
        this.deleteCustomBackground(bgUrl);
        return;
      }
    });

    // 背景上传处理
    DOM.image.uploadBgBtn?.addEventListener("click", () =>
      DOM.image.uploadBgInput?.click()
    );

    // 自定义背景上传处理
    DOM.image.uploadBgInput?.addEventListener("change", (e) => {
      const file = e.target.files[0];
      if (!file) return;

      // 上传背景
      this.uploadBackground(file);
      
      // 重置input的值，确保可以重复选择同一文件
      e.target.value = '';
    });
  }

  showImagePanel() {
    this.domUtils.showElements([DOM.menuImage, DOM.imagePage]);
    this.domUtils.hideElements(
      this.domUtils.getOtherPanelElements([DOM.menuImage, DOM.imagePage])
    );
  }

  async fetchBackgrounds() {
    try {
      const { data } = await request.backgroundList();
      return data.map((background) => ({
        title: background.title,
        // 使用新的文件获取接口
        url: background.image_path
          ? `/api/v1/chat/background/background_file/${encodeURIComponent(
              background.image_path
            )}`
          : "../pictures/background/default.png",
        time: background.modified_time,
      }));
    } catch (error) {
      console.error("获取背景列表失败:", error);
      return [];
    }
  }

  renderBackgrounds(backgrounds) {
    this.backgroundList.innerHTML = backgrounds
      .map((background) => this.createBackgroundCard(background))
      .join("");
  }

  createBackgroundCard(background) {
    const deleteButton = background.isCustom 
      ? `<button class="background-delete-btn" data-background-url="${background.url}" title="删除自定义背景">×</button>`
      : '';
    
    return `
      <div class="background-card">
        <div class="background-image-container">
          <img src="${background.url}" alt="${background.title}" class="background-image">
          ${deleteButton}
        </div>
        <div class="background-title" data-title="${background.title}">
          <button class="background-select-btn" data-background-url="${background.url}">选择</button>
        </div>
      </div>
    `;
  }

  selectBackground(card, bgUrl) {
    // 移除之前选中卡片的样式
    if (this.currentSelectedCard) {
      this.currentSelectedCard.classList.remove("selected");
      const prevBtn = this.currentSelectedCard.querySelector(
        ".background-select-btn"
      );
      if (prevBtn) {
        prevBtn.textContent = "选择";
        prevBtn.style.backgroundColor = "#4f46e5";
      }
    }

    // 设置新选中卡片的样式
    card.classList.add("selected");
    const selectBtn = card.querySelector(".background-select-btn");
    if (selectBtn) {
      selectBtn.textContent = "已选中";
      selectBtn.style.backgroundColor = "#10b981"; // 绿色表示已选中
    }

    // 更新背景图片
    document.body.style.backgroundImage = `url(${bgUrl})`;
    document.body.style.backgroundSize = "cover";
    document.body.style.backgroundPosition = "center";
    document.body.style.backgroundAttachment = "fixed";
    document.body.style.backgroundRepeat = "no-repeat";

    // 让预览区也更新
    this.kousanManager.updatePreviewBackground(bgUrl);

    // 保存当前选中的卡片
    this.currentSelectedCard = card;

    // 这里可以添加保存到本地存储的逻辑
    localStorage.setItem("selectedBackground", bgUrl);
  }

  async uploadBackground(file) {
    const fileName = file.name;
    const fileExt = fileName.slice(fileName.lastIndexOf(".")).toLowerCase();

    // 验证文件类型
    const allowedExts = [
      ".jpg",
      ".png",
      ".webp",
      ".bmp",
      ".svg",
      ".tif",
      ".gif",
    ];
    if (!allowedExts.includes(fileExt)) {
      alert("请上传支持的图片格式: " + allowedExts.join(", "));
      return;
    }

    // 验证文件大小（限制为5MB）
    const maxSize = 5 * 1024 * 1024; // 5MB
    if (file.size > maxSize) {
      alert("文件大小不能超过5MB");
      return;
    }

    try {
      // 显示上传状态
      const uploadStatus = document.getElementById("upload-status");
      if (uploadStatus) {
        uploadStatus.textContent = "正在处理图片...";
        uploadStatus.style.color = "#4f46e5";
      }

      // 使用FileReader读取文件
      const reader = new FileReader();
      reader.onload = (e) => {
        const imageUrl = e.target.result;
        
        // 创建自定义背景对象
        const customBackground = {
          title: `自定义背景 - ${fileName}`,
          url: imageUrl,
          time: new Date().toISOString(),
          isCustom: true
        };

        // 将自定义背景保存到localStorage
        const customBackgrounds = JSON.parse(localStorage.getItem('customBackgrounds') || '[]');
        
        // 检查是否已存在相同的图片（基于文件名）
        const existingIndex = customBackgrounds.findIndex(bg => bg.title === customBackground.title);
        if (existingIndex > -1) {
          // 替换现有的同名背景
          customBackgrounds[existingIndex] = customBackground;
        } else {
          // 添加新背景
          customBackgrounds.push(customBackground);
        }
        
        localStorage.setItem('customBackgrounds', JSON.stringify(customBackgrounds));

        // 重新渲染背景列表（包含自定义背景）
        this.refreshBackgroundWithCustom();
        
        // 直接选择这个新上传的背景
        setTimeout(() => {
          const newCard = this.findCardByUrl(imageUrl);
          if (newCard) {
            this.selectBackground(newCard, imageUrl);
          }
        }, 100);

        // 显示成功状态
        if (uploadStatus) {
          uploadStatus.textContent = "背景上传成功！";
          uploadStatus.style.color = "#10b981";
          setTimeout(() => {
            uploadStatus.textContent = "";
          }, 3000);
        }

        console.log("自定义背景上传成功");
      };

      reader.onerror = () => {
        console.error("文件读取失败");
        const uploadStatus = document.getElementById("upload-status");
        if (uploadStatus) {
          uploadStatus.textContent = "文件读取失败，请重试";
          uploadStatus.style.color = "#ef4444";
          setTimeout(() => {
            uploadStatus.textContent = "";
          }, 3000);
        }
      };

      // 读取文件为Data URL
      reader.readAsDataURL(file);
    } catch (error) {
      console.error("上传背景失败:", error);
      const uploadStatus = document.getElementById("upload-status");
      if (uploadStatus) {
        uploadStatus.textContent = "上传失败，请重试";
        uploadStatus.style.color = "#ef4444";
        setTimeout(() => {
          uploadStatus.textContent = "";
        }, 3000);
      }
    }
  }

  // 新增方法：刷新背景列表，包含自定义背景
  async refreshBackgroundWithCustom() {
    try {
      // 获取服务器背景
      const serverBackgrounds = await this.fetchBackgrounds();
      
      // 获取本地自定义背景
      const customBackgrounds = JSON.parse(localStorage.getItem('customBackgrounds') || '[]');
      
      // 合并背景列表（自定义背景排在前面）
      const allBackgrounds = [...customBackgrounds, ...serverBackgrounds];
      
      this.renderBackgrounds(allBackgrounds);
    } catch (error) {
      console.error("刷新背景列表失败:", error);
      // 失败时至少显示自定义背景
      const customBackgrounds = JSON.parse(localStorage.getItem('customBackgrounds') || '[]');
      const allBackgrounds = [...customBackgrounds, ...this.defaultBackgrounds];
      this.renderBackgrounds(allBackgrounds);
    }
  }

  // 删除自定义背景
  deleteCustomBackground(bgUrl) {
    if (confirm("确定要删除这个自定义背景吗？")) {
      // 从localStorage中删除
      const customBackgrounds = JSON.parse(localStorage.getItem('customBackgrounds') || '[]');
      const updatedBackgrounds = customBackgrounds.filter(bg => bg.url !== bgUrl);
      localStorage.setItem('customBackgrounds', JSON.stringify(updatedBackgrounds));

      // 如果当前选中的是要删除的背景，切换到默认背景
      const currentBg = localStorage.getItem("selectedBackground");
      if (currentBg === bgUrl) {
        localStorage.removeItem("selectedBackground");
        // 选择第一个可用的背景
        setTimeout(() => {
          const firstCard = this.backgroundList.querySelector(".background-card");
          if (firstCard) {
            const firstBgUrl = firstCard.querySelector(".background-select-btn").dataset.backgroundUrl;
            this.selectBackground(firstCard, firstBgUrl);
          }
        }, 100);
      }

      // 重新渲染背景列表
      this.refreshBackgroundWithCustom();
      console.log("自定义背景删除成功");
    }
  }
}
