/**
 * JWT 工具类
 */
class JWTUtils {
  /**
   * 解析JWT token
   * @param {string} token - JWT token (可以包含 Bearer 前缀)
   * @returns {object|null} - 解析后的payload，失败返回null
   */
  static parseToken(token) {
    try {
      if (!token) {
        return null;
      }
      
      // 移除 Bearer 前缀
      const actualToken = token.startsWith('Bearer ') 
        ? token.substring(7) 
        : token;
      
      // JWT token 由三部分组成，用 . 分隔
      const parts = actualToken.split('.');
      if (parts.length !== 3) {
        console.error('Invalid JWT token format');
        return null;
      }
      
      // 解码 payload (第二部分)
      const payload = parts[1];
      // 处理 base64url 编码 (替换字符并添加填充)
      const base64 = payload.replace(/-/g, '+')
                           .replace(/_/g, '/')
                           .padEnd(payload.length + (4 - payload.length % 4) % 4, '=');
      
      const decoded = atob(base64);
      return JSON.parse(decoded);
    } catch (error) {
      console.error('JWT parsing error:', error);
      return null;
    }
  }
  
  /**
   * 检查token是否过期
   * @param {object} payload - JWT payload
   * @returns {boolean} - 是否过期
   */
  static isTokenExpired(payload) {
    if (!payload || !payload.exp) {
      return true;
    }
    
    const currentTime = Math.floor(Date.now() / 1000);
    return payload.exp < currentTime;
  }
  
  /**
   * 从token中获取用户信息
   * @param {string} token - JWT token
   * @returns {object|null} - 用户信息对象
   */
  static getUserFromToken(token) {
    const payload = this.parseToken(token);
    if (!payload || this.isTokenExpired(payload)) {
      return null;
    }
    
    return {
      userId: payload.user_id,
      username: payload.username,
      email: payload.email,
      role: payload.role,
      isLoggedIn: true
    };
  }
  
  /**
   * 从localStorage获取token
   * @returns {string|null} - token或null
   */
  static getStoredToken() {
    return localStorage.getItem('auth_token');
  }
  
  /**
   * 保存token到localStorage
   * @param {string} token - JWT token
   */
  static saveToken(token) {
    localStorage.setItem('auth_token', token);
  }
  
  /**
   * 删除token
   */
  static removeToken() {
    localStorage.removeItem('auth_token');
  }
  
  /**
   * 获取当前登录的用户信息
   * @returns {object|null} - 用户信息或null
   */
  static getCurrentUser() {
    const token = this.getStoredToken();
    return this.getUserFromToken(token);
  }
}

// 导出到全局作用域
window.JWTUtils = JWTUtils; 