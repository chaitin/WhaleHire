/**
 * 文件下载工具函数
 */

/**
 * 通过URL下载文件
 * @param url 文件URL
 * @param filename 下载文件名（可选）
 */
export const downloadFileFromUrl = async (
  url: string,
  filename?: string
): Promise<void> => {
  try {
    // 验证URL
    if (!url || typeof url !== 'string') {
      throw new Error('无效的文件URL');
    }

    // 创建临时链接元素
    const link = document.createElement('a');
    link.href = url;

    // 设置下载属性
    if (filename) {
      link.download = filename;
    } else {
      // 从URL中提取文件名
      const urlParts = url.split('/');
      const urlFilename = urlParts[urlParts.length - 1];
      if (urlFilename && urlFilename.includes('.')) {
        link.download = urlFilename;
      }
    }

    // 设置target为_blank以避免页面跳转
    link.target = '_blank';
    link.rel = 'noopener noreferrer';

    // 添加到DOM并触发点击
    document.body.appendChild(link);
    link.click();

    // 清理DOM
    document.body.removeChild(link);
  } catch (error) {
    console.error('文件下载失败:', error);
    throw error;
  }
};

/**
 * 下载简历文件
 * @param resumeFileUrl 简历文件URL
 * @param resumeName 简历所有者姓名
 */
export const downloadResumeFile = async (
  resumeFileUrl: string,
  resumeName: string
): Promise<void> => {
  try {
    if (!resumeFileUrl) {
      throw new Error('简历文件URL不存在');
    }

    // 生成下载文件名
    const timestamp = new Date().toISOString().slice(0, 10); // YYYY-MM-DD
    const filename = `${resumeName}_简历_${timestamp}.pdf`;

    await downloadFileFromUrl(resumeFileUrl, filename);
  } catch (error) {
    console.error('简历下载失败:', error);
    throw error;
  }
};

/**
 * 检查文件URL是否有效
 * @param url 文件URL
 * @returns 是否有效
 */
export const isValidFileUrl = (url?: string): boolean => {
  if (!url || typeof url !== 'string') {
    return false;
  }

  try {
    new URL(url);
    return true;
  } catch {
    return false;
  }
};
