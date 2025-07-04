import { DOM } from "../../ui/dom.js";
import { DomUtils } from "../../utils/dom-utils.js";

export class AccountController {
  constructor() {
    this.processing = false;
    this.domUtils = DomUtils;
    this.init();
  }

  init() {
    this.bindEvents();
  }

  bindEvents() {
    if (!DOM.menuAccount) return;

    DOM.menuAccount.addEventListener("click", () => this.showAccountPanel());
  }

  showAccountPanel() {
    if (this.processing) return;
    this.processing = true;

    requestAnimationFrame(() => {
      // 显示账户相关元素
      this.domUtils.showElements([DOM.menuAccount, DOM.accountPage]);

      // 隐藏其他面板元素
      this.domUtils.hideElements(
        this.domUtils.getOtherPanelElements([DOM.menuAccount, DOM.accountPage])
      );

      // 更新账户信息
      this.updateAccountInfo();

      setTimeout(() => {
        this.processing = false;
      }, 300);
    });
  }

  updateAccountInfo() {
    const accountInfo = DOM.account.info;
    const accountActions = DOM.account.actions;

    if (!accountInfo || !accountActions) return;

    // 检查是否已登录
    const currentUser = this.getCurrentUser();

    if (currentUser) {
      // 已登录，显示用户信息
      accountInfo.innerHTML = `
        <div class="account-info-card">
          <div class="account-field">
            <span class="field-label">用户ID:</span>
            <span class="field-value">${currentUser.userId || 'N/A'}</span>
          </div>
          <div class="account-field">
            <span class="field-label">用户名:</span>
            <span class="field-value">${currentUser.username || 'N/A'}</span>
          </div>
          <div class="account-field">
            <span class="field-label">邮箱:</span>
            <span class="field-value">${currentUser.email || 'N/A'}</span>
          </div>
          <div class="account-field">
            <span class="field-label">角色:</span>
            <span class="field-value">${currentUser.role || 'user'}</span>
          </div>
          <div class="account-field">
            <span class="field-label">登录状态:</span>
            <span class="field-value status-logged-in">已登录</span>
          </div>
        </div>
      `;

      accountActions.innerHTML = `
        <button class="big-button logout-btn" id="logout-btn">
          🚪 登出
        </button>
      `;

      // 绑定登出按钮事件
      const logoutBtn = document.getElementById("logout-btn");
      if (logoutBtn) {
        logoutBtn.addEventListener("click", () => this.handleLogout());
      }
    } else {
      // 未登录，显示提示
      accountInfo.innerHTML = `
        <div class="account-info-card">
          <div class="account-field">
            <span class="field-label">登录状态:</span>
            <span class="field-value status-not-logged-in">尚未登录</span>
          </div>
          <div class="account-message">
            <p>您当前未登录，请先登录以查看账户信息</p>
          </div>
        </div>
      `;

      accountActions.innerHTML = `
        <button class="big-button login-btn" id="go-login-btn">
          🔐 前往登录
        </button>
      `;

      // 绑定登录按钮事件
      const loginBtn = document.getElementById("go-login-btn");
      if (loginBtn) {
        loginBtn.addEventListener("click", () => this.handleGoToLogin());
      }
    }
  }

  getCurrentUser() {
    // 使用全局的JWTUtils来获取当前用户信息
    if (typeof JWTUtils !== 'undefined') {
      return JWTUtils.getCurrentUser();
    }
    return null;
  }

  handleLogout() {
    if (confirm("确定要登出吗？")) {
      try {
        // 使用全局的清理用户数据函数
        if (typeof clearAllUserData === 'function') {
          clearAllUserData();
        } else {
          // 手动清理数据
          if (typeof JWTUtils !== 'undefined') {
            JWTUtils.removeToken();
          }
          localStorage.clear();
        }

        // 跳转到登录页面
        window.location.href = "/login";
      } catch (error) {
        console.error("登出时发生错误:", error);
        alert("登出时发生错误，请刷新页面后重试");
      }
    }
  }

  handleGoToLogin() {
    window.location.href = "/login";
  }
} 