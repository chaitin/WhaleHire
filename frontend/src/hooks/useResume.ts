// 简历管理Hook - 根据swagger.json更新
import { useState, useEffect, useCallback, useRef } from 'react';
import { 
  getResumeList, 
  getResumeDetail, 
  updateResume, 
  deleteResume, 
  batchDeleteResumes,
  uploadResume,
  getResumeProgress,
  reparseResume,
  searchResumes,
  ResumeListParams,
  ResumeUpdateParams,
  ResumeListResponse,
  ResumeSearchParams,
  ResumeSearchResponse,
  ResumeParseProgress
} from '@/services/resume';
import { Resume, ResumeDetail } from '@/types/resume';

const shallowEqualParams = (a: ResumeListParams, b: ResumeListParams): boolean => {
  const keys = new Set<string>([...Object.keys(a), ...Object.keys(b)]);

  for (const key of keys) {
    const typedKey = key as keyof ResumeListParams;
    if (a[typedKey] !== b[typedKey]) {
      return false;
    }
  }

  return true;
};

// 简历列表管理 Hook
export const useResumeList = (initialParams: ResumeListParams = {}) => {
  const initialParamsRef = useRef<ResumeListParams>(initialParams);
  const [resumes, setResumes] = useState<Resume[]>([]);
  const resumesRef = useRef<Resume[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
    hasNextPage: false,
    nextToken: undefined as string | undefined
  });
  const [params, setParamsState] = useState<ResumeListParams>(initialParamsRef.current);
  const lastParamsRef = useRef<ResumeListParams>(initialParamsRef.current);
  const pollingTimerRef = useRef<ReturnType<typeof setInterval> | null>(null);
  type FetchResumesFunction = (
    newParams?: ResumeListParams,
    options?: { force?: boolean }
  ) => Promise<void>;
  const fetchResumesRef = useRef<FetchResumesFunction | null>(null);

  useEffect(() => {
    resumesRef.current = resumes;
  }, [resumes]);

  const stopProgressPolling = useCallback(() => {
    if (pollingTimerRef.current) {
      clearInterval(pollingTimerRef.current);
      pollingTimerRef.current = null;
    }
  }, []);


  const scheduleProgressPolling = useCallback(() => {
    // 不再进行状态轮询，改为依赖列表刷新
    stopProgressPolling();
  }, [stopProgressPolling]);

  const fetchResumes = useCallback(
    async (newParams?: ResumeListParams, options?: { force?: boolean }) => {
      const mergedParams = { ...lastParamsRef.current, ...(newParams ?? {}) };
      const shouldSkip =
        !options?.force && Boolean(newParams) && shallowEqualParams(mergedParams, lastParamsRef.current);

      if (shouldSkip) {
        return;
      }

      lastParamsRef.current = mergedParams;
      setParamsState((prev) => (shallowEqualParams(prev, mergedParams) ? prev : mergedParams));

      setLoading(true);
      setError(null);

      try {
        console.log('获取简历列表，参数:', mergedParams);
        const response: ResumeListResponse = await getResumeList(mergedParams);
        console.log('简历列表响应:', response);

        // 确保响应数据结构正确
        if (!response || typeof response !== 'object') {
          throw new Error('API 响应格式错误');
        }

        // 更新简历列表数据 - 如果 resumes 为空数组，这是正常情况，不应该报错
        setResumes(response.resumes || []);
        
        // 更新分页信息
        setPagination({
          current: mergedParams.page || 1,
          pageSize: mergedParams.size || 10,
          total: response.total_count || 0,
          hasNextPage: response.has_next_page || false,
          nextToken: response.next_token || '',
        });

        // 检查是否有状态变化的简历
        const resumes = response.resumes || [];
        const processingCount = resumes.filter(resume => 
          resume.status === 'pending' || resume.status === 'processing'
        ).length;
        
        if (processingCount > 0) {
          console.log(`检测到 ${processingCount} 个处理中的简历`);
        } else {
          console.log('所有简历处理完成');
        }
        
      } catch (err) {
        console.error('获取简历列表失败:', err);
        setError(err instanceof Error ? err.message : '获取简历列表失败');
      } finally {
        setLoading(false);
        scheduleProgressPolling();
      }
    },
    [scheduleProgressPolling]
  );

  useEffect(() => {
    fetchResumesRef.current = fetchResumes;
  }, [fetchResumes]);

  // 更新简历
  const updateResumeItem = useCallback(async (id: string, data: Omit<ResumeUpdateParams, 'id'>) => {
    setLoading(true);
    setError(null);
    
    try {
      const updatedResume = await updateResume(id, data);
      setResumes(prev => prev.map(resume => 
        resume.id === updatedResume.id ? updatedResume : resume
      ));
      return updatedResume;
    } catch (err) {
      setError(err instanceof Error ? err.message : '更新简历失败');
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  // 删除简历
  const deleteResumeItem = useCallback(async (id: string) => {
    setLoading(true);
    setError(null);
    
    try {
      await deleteResume(id);
      setResumes(prev => prev.filter(resume => resume.id !== id));
    } catch (err) {
      setError(err instanceof Error ? err.message : '删除简历失败');
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  // 批量删除
  const batchDelete = useCallback(async (ids: string[]) => {
    setLoading(true);
    setError(null);
    
    try {
      await batchDeleteResumes(ids);
      setResumes(prev => prev.filter(resume => !ids.includes(resume.id)));
    } catch (err) {
      setError(err instanceof Error ? err.message : '批量删除失败');
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  // 搜索简历
  const searchResumeList = useCallback(async (searchParams: ResumeSearchParams) => {
    setLoading(true);
    setError(null);
    
    try {
      const response: ResumeSearchResponse = await searchResumes(searchParams);
      setResumes(response.resumes);
      setPagination({
        current: searchParams.page || 1,
        pageSize: searchParams.size || 10,
        total: response.total_count,
        hasNextPage: response.has_next_page,
        nextToken: response.next_token
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : '搜索简历失败');
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!shallowEqualParams(initialParamsRef.current, initialParams)) {
      initialParamsRef.current = initialParams;
      lastParamsRef.current = {} as ResumeListParams;
      fetchResumes(initialParams);
    }
  }, [fetchResumes, initialParams]);

  useEffect(() => {
    fetchResumes();
  }, [fetchResumes]);

  useEffect(() => {
    scheduleProgressPolling();
  }, [resumes, scheduleProgressPolling]);

  const hasNewResumeRef = useRef(false);

  const triggerRefreshOnCreate = useCallback(() => {
    if (!hasNewResumeRef.current) {
      return;
    }

    hasNewResumeRef.current = false;
    void fetchResumes(lastParamsRef.current, { force: true });
  }, [fetchResumes]);

  useEffect(() => {
    triggerRefreshOnCreate();
  }, [resumes, triggerRefreshOnCreate]);

  useEffect(() => {
    return () => {
      stopProgressPolling();
    };
  }, [stopProgressPolling]);

  const addResumeItem = useCallback((resume: Resume) => {
    console.log('添加简历到列表:', resume.name, '状态:', resume.status);
    
    setResumes((prev) => {
      const exists = prev.some((item) => item.id === resume.id);

      const nextResumes = exists
        ? prev.map((item) => (item.id === resume.id ? { ...item, ...resume } : item))
        : [resume, ...prev];

      console.log('更新后的简历列表长度:', nextResumes.length);
      console.log('新简历是否在列表顶部:', nextResumes[0]?.id === resume.id);

      resumesRef.current = nextResumes;
      hasNewResumeRef.current = true;
      scheduleProgressPolling();

      // 更新分页信息以反映新增的简历
      setPagination(prevPagination => ({
        ...prevPagination,
        total: exists ? prevPagination.total : prevPagination.total + 1
      }));

      return nextResumes;
    });
  }, [scheduleProgressPolling]);

  return {
    resumes,
    loading,
    error,
    pagination,
    params,
    fetchResumes,
    updateResumeItem,
    deleteResumeItem,
    batchDelete,
    searchResumeList,
    addResumeItem,
    refresh: () => fetchResumes(lastParamsRef.current, { force: true }),
    setParams: (newParams: ResumeListParams) => {
      fetchResumes(newParams);
    }
  };
};

// 简历详情管理 Hook
export const useResumeDetail = (id?: string) => {
  const [resume, setResume] = useState<ResumeDetail | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchResume = useCallback(async (resumeId: string) => {
    setLoading(true);
    setError(null);
    
    try {
      const resumeData = await getResumeDetail(resumeId);
      setResume(resumeData);
    } catch (err) {
      setError(err instanceof Error ? err.message : '获取简历详情失败');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (id) {
      fetchResume(id);
    }
  }, [id, fetchResume]);

  return {
    resume,
    loading,
    error,
    fetchResume,
    refresh: () => id && fetchResume(id)
  };
};

// 简历上传 Hook
export const useResumeUpload = () => {
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [uploadProgress, setUploadProgress] = useState(0);

  const uploadFile = useCallback(async (file: File, position?: string): Promise<Resume> => {
    setUploading(true);
    setError(null);
    setUploadProgress(0);
    
    try {
      // 模拟上传进度
      const progressInterval = setInterval(() => {
        setUploadProgress(prev => {
          if (prev >= 90) {
            clearInterval(progressInterval);
            return prev;
          }
          return prev + 10;
        });
      }, 200);

      const response = await uploadResume(file, position);
      
      clearInterval(progressInterval);
      setUploadProgress(100);
      
      return response;
    } catch (err) {
      setError(err instanceof Error ? err.message : '上传失败');
      throw err;
    } finally {
      setUploading(false);
      setTimeout(() => setUploadProgress(0), 1000);
    }
  }, []);

  return {
    uploading,
    error,
    uploadProgress,
    uploadFile
  };
};

// 简历解析进度 Hook
export const useResumeProgress = (id?: string) => {
  const [progress, setProgress] = useState<ResumeParseProgress | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchProgress = useCallback(async (resumeId: string) => {
    setLoading(true);
    setError(null);
    
    try {
      const progressData = await getResumeProgress(resumeId);
      setProgress(progressData);
    } catch (err) {
      setError(err instanceof Error ? err.message : '获取解析进度失败');
    } finally {
      setLoading(false);
    }
  }, []);

  const reparse = useCallback(async (resumeId: string) => {
    setLoading(true);
    setError(null);
    
    try {
      await reparseResume(resumeId);
      // 重新获取进度
      await fetchProgress(resumeId);
    } catch (err) {
      setError(err instanceof Error ? err.message : '重新解析失败');
      throw err;
    } finally {
      setLoading(false);
    }
  }, [fetchProgress]);

  useEffect(() => {
    if (id) {
      fetchProgress(id);
    }
  }, [id, fetchProgress]);

  return {
    progress,
    loading,
    error,
    fetchProgress,
    reparse,
    refresh: () => id && fetchProgress(id)
  };
};