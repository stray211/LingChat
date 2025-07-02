/**
 * 清理所有用户认证数据
 */
function clearAllUserData() {
  console.log("正在清理所有用户数据...");
  
  // 清理JWT token
  JWTUtils.removeToken();
  
  // 清理记住登录的数据
  localStorage.removeItem("remembered_username");
  localStorage.removeItem("remembered_password");
  localStorage.removeItem("remember_me");
  
  // 清理旧的用户数据结构（兼容性）
  localStorage.removeItem("user");
  localStorage.removeItem("remembered_email");
  
  // 清理聊天相关数据
  localStorage.removeItem("currentConversationId");
  localStorage.removeItem("lastAIResponse");
  
  // 清理UI状态
  localStorage.removeItem("selectedAI");
  localStorage.removeItem("selectedModel");
  localStorage.removeItem("backgroundImage");
  localStorage.removeItem("bgmEnabled");
  localStorage.removeItem("soundEnabled");
  
  console.log("用户数据清理完成");
}

/**
 * 用户登出
 */
function logout() {
  clearAllUserData();
  
  // 跳转到登录页面
  window.location.href = "/login";
}

/**
 * 检查用户登录状态并获取用户信息
 * @returns {object|null} 用户信息或null
 */
function getCurrentUserInfo() {
  return JWTUtils.getCurrentUser();
}

// 导出到全局作用域
window.clearAllUserData = clearAllUserData;
window.logout = logout;
window.getCurrentUserInfo = getCurrentUserInfo;