// 时间格式化工具
export class DateUtils {
  /**
   * 解析RFC3339格式时间并格式化为友好显示
   * @param {string} rfc3339String - RFC3339格式的时间字符串 (如: "2023-07-04T18:16:25.123Z")
   * @returns {string} 格式化后的时间字符串
   */
  static formatRFC3339(rfc3339String) {
    if (!rfc3339String) {
      return "未知时间";
    }

    try {
      const date = new Date(rfc3339String);
      
      // 检查日期是否有效
      if (isNaN(date.getTime())) {
        return "无效时间";
      }

      return this.formatDate(date);
    } catch (error) {
      console.warn("解析时间失败:", rfc3339String, error);
      return "时间解析失败";
    }
  }

  /**
   * 格式化Date对象为友好显示
   * @param {Date} date - Date对象
   * @returns {string} 格式化后的时间字符串
   */
  static formatDate(date) {
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMinutes = Math.floor(diffMs / (1000 * 60));
    const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

    // 相对时间显示
    if (diffMinutes < 1) {
      return "刚刚";
    } else if (diffMinutes < 60) {
      return `${diffMinutes}分钟前`;
    } else if (diffHours < 24) {
      return `${diffHours}小时前`;
    } else if (diffDays < 7) {
      return `${diffDays}天前`;
    }

    // 绝对时间显示
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    const hours = String(date.getHours()).padStart(2, '0');
    const minutes = String(date.getMinutes()).padStart(2, '0');

    const currentYear = now.getFullYear();
    
    // 如果是今年，不显示年份
    if (year === currentYear) {
      return `${month}-${day} ${hours}:${minutes}`;
    } else {
      return `${year}-${month}-${day} ${hours}:${minutes}`;
    }
  }

  /**
   * 解析多种时间格式的通用方法
   * @param {string} timeString - 时间字符串
   * @returns {string} 格式化后的时间字符串
   */
  static parseAndFormat(timeString) {
    if (!timeString) {
      return "未知时间";
    }

    // 如果已经是格式化后的字符串（包含中文），直接返回
    if (/[\u4e00-\u9fa5]/.test(timeString)) {
      return timeString;
    }

    // 尝试解析RFC3339/ISO8601格式
    if (timeString.includes('T') || timeString.includes('Z')) {
      return this.formatRFC3339(timeString);
    }

    // 尝试解析其他常见格式
    try {
      const date = new Date(timeString);
      if (!isNaN(date.getTime())) {
        return this.formatDate(date);
      }
    } catch (error) {
      console.warn("解析时间失败:", timeString, error);
    }

    return timeString; // 如果无法解析，返回原字符串
  }
} 