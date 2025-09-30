// 简历管理相关API服务 - 根据swagger.json定义
import { apiGet, apiPost, apiPut, apiDelete } from '@/lib/api';
import { Resume, ResumeDetail, ResumeStatus } from '@/types/resume';


// 简历列表查询参数 - 根据swagger定义
export interface ResumeListParams {
  page?: number;
  size?: number;
  next_token?: string;
  position?: string;
  status?: string;
  keywords?: string;
}

// 简历搜索参数
export interface ResumeSearchParams extends ResumeListParams {
  keywords: string;
}

// 简历列表响应 - 根据swagger ListResumeResp定义
export interface ResumeListResponse {
  resumes: Resume[];
  total_count: number;
  has_next_page: boolean;
  next_token?: string;
}

// 简历搜索响应 - 根据swagger SearchResumeResp定义
export interface ResumeSearchResponse {
  resumes: Resume[];
  total_count: number;
  has_next_page: boolean;
  next_token?: string;
}

// 简历更新参数 - 根据swagger UpdateResumeReq定义
export interface ResumeUpdateParams {
  id: string;
  name?: string;
  email?: string;
  phone?: string;
  gender?: string;
  birthday?: string;
  current_city?: string;
  highest_education?: string;
  years_experience?: number;
}

// 简历解析进度 - 根据swagger ResumeParseProgress定义
export interface ResumeParseProgress {
  resume_id: string;
  status: ResumeStatus;
  progress: number; // 0-100
  message: string;
  error_message?: string;
  started_at: string;
  completed_at?: string;
}

// 获取简历列表
export const getResumeList = async (params: ResumeListParams = {}): Promise<ResumeListResponse> => {
  const queryParams = new URLSearchParams();

  if (params.page) queryParams.append('page', params.page.toString());
  if (params.size) queryParams.append('size', params.size.toString());
  if (params.next_token) queryParams.append('next_token', params.next_token);
  if (params.position && params.position !== 'all') queryParams.append('position', params.position);
  if (params.status && params.status !== 'all') queryParams.append('status', params.status);
  if (params.keywords) queryParams.append('keywords', params.keywords);

  return apiGet<ResumeListResponse>(`/v1/resume/list?${queryParams.toString()}`);
};

// 搜索简历
export const searchResumes = async (params: ResumeSearchParams): Promise<ResumeSearchResponse> => {
  const queryParams = new URLSearchParams();

  // 搜索关键词是必需的
  queryParams.append('keywords', params.keywords);
  if (params.page) queryParams.append('page', params.page.toString());
  if (params.size) queryParams.append('size', params.size.toString());
  if (params.next_token) queryParams.append('next_token', params.next_token);
  // 添加其他筛选条件
  if (params.position && params.position !== 'all') queryParams.append('position', params.position);
  if (params.status && params.status !== 'all') queryParams.append('status', params.status);

  return apiGet<ResumeSearchResponse>(`/v1/resume/search?${queryParams.toString()}`);
};

// 获取简历详情
export const getResumeDetail = async (id: string): Promise<ResumeDetail> => {
  return apiGet<ResumeDetail>(`/v1/resume/${id}`);
};

// 更新简历
export const updateResume = async (id: string, params: Omit<ResumeUpdateParams, 'id'>): Promise<Resume> => {
  const updateData = { ...params, id };
  return apiPut<Resume>(`/v1/resume/${id}`, updateData);
};

// 删除简历
export const deleteResume = async (id: string): Promise<void> => {
  await apiDelete(`/v1/resume/${id}`);
};

// 上传简历
export const uploadResume = async (file: File, position?: string): Promise<Resume> => {
  const formData = new FormData();
  formData.append('file', file);
  if (position) {
    formData.append('position', position);
  }

  return apiPost<Resume>('/v1/resume/upload', formData);
};

// 获取简历解析进度
export const getResumeProgress = async (id: string): Promise<ResumeParseProgress> => {
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

// 下载简历文件
export const downloadResumeFile = async (resume: Resume, fileName?: string): Promise<void> => {
  try {
    console.log('🔽 开始下载简历:', { resumeId: resume.id, fileName, resume_file_url: resume.resume_file_url });
    
    let response: Response;
    let downloadUrl = '';
    
    // 方法1: 尝试使用 resume_id 参数
    try {
      downloadUrl = `/api/v1/file/download?resume_id=${resume.id}`;
      console.log('🔽 尝试方法1 - resume_id:', downloadUrl);
      response = await fetch(downloadUrl, {
        method: 'GET',
        credentials: 'include', // 使用 Cookie 认证
      });
      
      if (response.ok) {
        console.log('✅ 方法1成功');
      } else {
        throw new Error(`方法1失败: ${response.status}`);
      }
    } catch (error1) {
      console.log('❌ 方法1失败:', error1);
      
      // 方法2: 尝试使用 key 参数（使用 resume.id）
      try {
        downloadUrl = `/api/v1/file/download?key=${resume.id}`;
        console.log('🔽 尝试方法2 - key (resume.id):', downloadUrl);
        response = await fetch(downloadUrl, {
          method: 'GET',
          credentials: 'include',
        });
        
        if (response.ok) {
          console.log('✅ 方法2成功');
        } else {
          throw new Error(`方法2失败: ${response.status}`);
        }
      } catch (error2) {
        console.log('❌ 方法2失败:', error2);
        
        // 方法3: 如果有 resume_file_url，尝试从中提取文件名作为 key
        if (resume.resume_file_url) {
          const fileKey = resume.resume_file_url.split('/').pop() || resume.id;
          downloadUrl = `/api/v1/file/download?key=${fileKey}`;
          console.log('🔽 尝试方法3 - key (from URL):', downloadUrl);
          response = await fetch(downloadUrl, {
            method: 'GET',
            credentials: 'include',
          });
          
          if (response.ok) {
            console.log('✅ 方法3成功');
          } else {
            throw new Error(`方法3失败: ${response.status}`);
          }
        } else {
          throw new Error('所有下载方法都失败了');
        }
      }
    }

    if (!response.ok) {
      throw new Error(`下载失败: ${response.status} ${response.statusText}`);
    }

    // 获取文件名
    const contentDisposition = response.headers.get('Content-Disposition');
    let downloadFileName = fileName || `${resume.name}_简历.pdf`;
    
    if (contentDisposition) {
      const fileNameMatch = contentDisposition.match(/filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/);
      if (fileNameMatch && fileNameMatch[1]) {
        downloadFileName = fileNameMatch[1].replace(/['"]/g, '');
      }
    }

    console.log('🔽 开始下载文件:', downloadFileName);

    // 创建下载链接
    const blob = await response.blob();
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
    throw error;
  }
};