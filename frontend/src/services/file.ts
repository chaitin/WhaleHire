// 文件管理相关API服务 - 根据swagger.json定义
import { apiGet, apiPost } from '@/lib/api';

// 文件上传响应 - 根据swagger ObjectUploadResp定义
export interface FileUploadResponse {
  key: string;
}

// 文件下载响应 - 根据swagger ObjectDownloadResp定义
export interface FileDownloadResponse {
  url: string;
}

// 上传文件
export const uploadFile = async (
  file: File,
  kb_id?: string
): Promise<FileUploadResponse> => {
  const formData = new FormData();
  formData.append('file', file);
  if (kb_id) {
    formData.append('kb_id', kb_id);
  }

  return apiPost<FileUploadResponse>('/v1/file/upload', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });
};

// 获取文件下载链接
export const getFileDownloadUrl = async (
  key: string
): Promise<FileDownloadResponse> => {
  return apiGet<FileDownloadResponse>(
    `/v1/file/download?key=${encodeURIComponent(key)}`
  );
};

// 下载文件
export const downloadFile = async (
  key: string,
  fileName?: string
): Promise<void> => {
  const { url } = await getFileDownloadUrl(key);

  // 创建临时链接下载文件
  const link = document.createElement('a');
  link.href = url;
  link.download = fileName || key;
  link.target = '_blank';
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
};
