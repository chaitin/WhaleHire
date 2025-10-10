// 简历管理相关API服务 - 根据swagger.json定义
import { apiGet, apiPost, apiPut, apiDelete } from '@/lib/api';
import {
  Resume,
  ResumeDetail,
  ResumeListParams,
  ResumeSearchParams,
  ResumeListResponse,
  ResumeSearchResponse,
  ResumeUpdateParams,
  ResumeParseProgress,
} from '@/types/resume';

// 获取简历列表
export const getResumeList = async (
  params: ResumeListParams = {}
): Promise<ResumeListResponse> => {
  const queryParams = new URLSearchParams();

  if (params.page) queryParams.append('page', params.page.toString());
  if (params.size) queryParams.append('size', params.size.toString());
  if (params.next_token) queryParams.append('next_token', params.next_token);
  if (params.position && params.position !== 'all')
    queryParams.append('position', params.position);
  if (params.status && params.status !== 'all')
    queryParams.append('status', params.status);
  if (params.keywords) queryParams.append('keywords', params.keywords);

  return apiGet<ResumeListResponse>(
    `/v1/resume/list?${queryParams.toString()}`
  );
};

// 搜索简历
export const searchResumes = async (
  params: ResumeSearchParams
): Promise<ResumeSearchResponse> => {
  const queryParams = new URLSearchParams();

  // 搜索关键词是必需的
  queryParams.append('keywords', params.keywords);
  if (params.page) queryParams.append('page', params.page.toString());
  if (params.size) queryParams.append('size', params.size.toString());
  if (params.next_token) queryParams.append('next_token', params.next_token);
  // 添加其他筛选条件
  if (params.position && params.position !== 'all')
    queryParams.append('position', params.position);
  if (params.status && params.status !== 'all')
    queryParams.append('status', params.status);

  return apiGet<ResumeSearchResponse>(
    `/v1/resume/search?${queryParams.toString()}`
  );
};

// 获取简历详情
export const getResumeDetail = async (id: string): Promise<ResumeDetail> => {
  return apiGet<ResumeDetail>(`/v1/resume/${id}`);
};

// 更新简历
export const updateResume = async (
  id: string,
  params: ResumeUpdateParams
): Promise<Resume> => {
  return apiPut<Resume>(`/v1/resume/${id}`, params as Record<string, unknown>);
};

// 删除简历
export const deleteResume = async (id: string): Promise<void> => {
  await apiDelete(`/v1/resume/${id}`);
};

// 上传简历
export const uploadResume = async (
  file: File,
  position?: string
): Promise<Resume> => {
  const formData = new FormData();
  formData.append('file', file);
  if (position) {
    formData.append('position', position);
  }

  return apiPost<Resume>('/v1/resume/upload', formData);
};

// 获取简历解析进度
export const getResumeProgress = async (
  id: string
): Promise<ResumeParseProgress> => {
  return apiGet<ResumeParseProgress>(`/v1/resume/${id}/progress`);
};

// 重新解析简历
export const reparseResume = async (id: string): Promise<void> => {
  await apiPost(`/v1/resume/${id}/reparse`);
};

// 批量删除简历（前端实现）
export const batchDeleteResumes = async (ids: string[]): Promise<void> => {
  const promises = ids.map((id) => deleteResume(id));
  await Promise.all(promises);
};

// 验证和处理文件URL
const processFileUrl = (url: string): string => {
  if (!url) {
    throw new Error('文件URL为空');
  }

  // 如果是完整的URL，直接返回
  if (url.startsWith('http://') || url.startsWith('https://')) {
    return url;
  }

  // 如果是相对路径，转换为完整URL
  if (url.startsWith('/')) {
    // 获取当前域名和端口
    const baseUrl = window.location.origin;
    return `${baseUrl}${url}`;
  }

  // 如果不是以/开头的相对路径，添加/前缀
  const baseUrl = window.location.origin;
  return `${baseUrl}/${url}`;
};

// 下载简历文件 - 修改为通过 /api/v1/resume/{id} 接口获取 resume_file_url 进行下载
export const downloadResumeFile = async (
  resume: Resume,
  fileName?: string
): Promise<void> => {
  try {
    console.log('🔽 开始下载简历:', { resumeId: resume.id, fileName });

    // 首先调用 /api/v1/resume/{id} 接口获取最新的简历详情，确保获取到正确的 resume_file_url
    const resumeDetail = await getResumeDetail(resume.id);
    console.log('🔽 获取到简历详情:', {
      id: resumeDetail.id,
      name: resumeDetail.name,
      resume_file_url: resumeDetail.resume_file_url,
    });

    if (!resumeDetail.resume_file_url) {
      console.error('❌ 简历文件URL不存在');
      throw new Error('简历文件URL不存在，请联系管理员检查文件是否已上传');
    }

    // 处理和验证文件URL
    let fileUrl: string;
    try {
      fileUrl = processFileUrl(resumeDetail.resume_file_url);
      console.log('🔽 处理后的文件URL:', fileUrl);
    } catch (urlError) {
      console.error('❌ URL处理失败:', urlError);
      throw new Error('文件URL格式无效，无法下载');
    }

    // 使用处理后的URL下载文件
    console.log('🔽 开始请求文件:', fileUrl);
    const response = await fetch(fileUrl, {
      method: 'GET',
      credentials: 'include', // 使用 Cookie 认证
    });

    console.log('🔽 文件请求响应:', {
      status: response.status,
      statusText: response.statusText,
      headers: Object.fromEntries(response.headers.entries()),
    });

    if (!response.ok) {
      if (response.status === 404) {
        throw new Error('文件不存在，可能已被删除或移动');
      } else if (response.status === 403) {
        throw new Error('没有权限访问该文件');
      } else if (response.status === 401) {
        throw new Error('请先登录后再下载');
      } else {
        throw new Error(`下载失败: ${response.status} ${response.statusText}`);
      }
    }

    // 检查响应内容类型
    const contentType = response.headers.get('Content-Type');
    console.log('🔽 文件内容类型:', contentType);

    // 获取文件名
    const contentDisposition = response.headers.get('Content-Disposition');
    let downloadFileName =
      fileName || `${resumeDetail.name || resume.name}_简历.pdf`;

    if (contentDisposition) {
      const fileNameMatch = contentDisposition.match(
        /filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/
      );
      if (fileNameMatch && fileNameMatch[1]) {
        downloadFileName = fileNameMatch[1].replace(/['"]/g, '');
        // 解码URL编码的文件名
        try {
          downloadFileName = decodeURIComponent(downloadFileName);
        } catch {
          console.warn('文件名解码失败，使用原始文件名');
        }
      }
    }

    console.log('🔽 最终下载文件名:', downloadFileName);

    // 创建下载链接
    const blob = await response.blob();
    console.log('🔽 文件大小:', blob.size, 'bytes');

    if (blob.size === 0) {
      throw new Error('文件为空，无法下载');
    }

    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = downloadFileName;
    document.body.appendChild(link);
    link.click();

    // 清理
    document.body.removeChild(link);
    window.URL.revokeObjectURL(url);

    console.log('✅ 文件下载完成');
  } catch (error) {
    console.error('❌ 下载简历失败:', error);
    // 重新抛出错误，保持原始错误信息
    throw error instanceof Error ? error : new Error('下载失败，请稍后重试');
  }
};
