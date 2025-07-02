// auth-check.js - Script to check authentication status using JWT tokens

function performAuthCheck() {
  // Skip login check for the login page
  if (window.location.pathname.includes("login")) {
    return;
  }

  // Check if JWTUtils is available
  if (typeof JWTUtils === 'undefined') {
    window.location.href = "/login";
    return;
  }

  // Use JWT utils to check user login status
  const currentUser = JWTUtils.getCurrentUser();

  if (!currentUser) {
    // Clean up any invalid data
    JWTUtils.removeToken();
    window.location.href = "/login";
    return;
  }

  // Set global user info for other scripts to use
  window.currentUser = currentUser;
}

// 立即执行一次认证检查（适用于脚本动态加载的情况）
if (document.readyState === 'loading') {
  document.addEventListener("DOMContentLoaded", performAuthCheck);
} else {
  // 如果DOM已经加载完成，立即执行
  performAuthCheck();
}

// Function to check login status
function checkLoginStatus() {
  const currentUser = JWTUtils.getCurrentUser();
  return currentUser !== null;
}

// Function to check if user is admin
function isAdminUser() {
  const currentUser = JWTUtils.getCurrentUser();
  
  if (currentUser) {
    return currentUser.role === "admin";
  }
  
  return false;
}

// Function to get the current user from JWT
function getCurrentUserFromAuth() {
  return JWTUtils.getCurrentUser();
}

// Function to logout
function logoutUser() {
  JWTUtils.removeToken();
  clearAllUserData(); // Clear all user data
  window.location.href = "/login";
}

// 将函数挂载到全局作用域
window.checkLoginStatus = checkLoginStatus;
window.isAdminUser = isAdminUser;
window.getCurrentUserFromAuth = getCurrentUserFromAuth;
window.logoutUser = logoutUser;