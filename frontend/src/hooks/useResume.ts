// 简历管理Hook - 根据swagger.json更新
import { useState, useEffect, useCallback, useRef } from 'react';
import {
  getResumeList,
  getResumeDetail,
  updateResume,
  deleteResume,
  batchDeleteResumes,
  batchUploadResume,
  getBatchUploadStatus,
  getResumeProgress,
  reparseResume,
  searchResumes,
} from '@/services/resume';
import type { BatchUploadStatus } from '@/services/resume';
import {
  Resume,
  ResumeDetail,
  ResumeListParams,
  ResumeUpdateParams,
  ResumeListResponse,
  ResumeSearchParams,
  ResumeSearchResponse,
  ResumeParseProgress,
} from '@/types/resume';

const shallowEqualParams = (
  a: ResumeListParams,
  b: ResumeListParams
): boolean => {
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
    nextToken: undefined as string | undefined,
  });
  const [params, setParamsState] = useState<ResumeListParams>(
    initialParamsRef.current
  );
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
        !options?.force &&
        Boolean(newParams) &&
        shallowEqualParams(mergedParams, lastParamsRef.current);

      if (shouldSkip) {
        return;
      }

      lastParamsRef.current = mergedParams;
      setParamsState((prev) =>
        shallowEqualParams(prev, mergedParams) ? prev : mergedParams
      );

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
        const processingCount = resumes.filter(
          (resume) =>
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
  const updateResumeItem = useCallback(
    async (id: string, data: Omit<ResumeUpdateParams, 'id'>) => {
      setLoading(true);
      setError(null);

      try {
        const updatedResume = await updateResume(id, data);
        setResumes((prev) =>
          prev.map((resume) =>
            resume.id === updatedResume.id ? updatedResume : resume
          )
        );
        return updatedResume;
      } catch (err) {
        setError(err instanceof Error ? err.message : '更新简历失败');
        throw err;
      } finally {
        setLoading(false);
      }
    },
    []
  );

  // 删除简历
  const deleteResumeItem = useCallback(async (id: string) => {
    setLoading(true);
    setError(null);

    try {
      await deleteResume(id);
      setResumes((prev) => prev.filter((resume) => resume.id !== id));
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
      setResumes((prev) => prev.filter((resume) => !ids.includes(resume.id)));
    } catch (err) {
      setError(err instanceof Error ? err.message : '批量删除失败');
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  // 搜索简历
  const searchResumeList = useCallback(
    async (searchParams: ResumeSearchParams) => {
      setLoading(true);
      setError(null);

      try {
        const response: ResumeSearchResponse =
          await searchResumes(searchParams);
        setResumes(response.resumes);
        setPagination({
          current: searchParams.page || 1,
          pageSize: searchParams.size || 10,
          total: response.total_count,
          hasNextPage: response.has_next_page,
          nextToken: response.next_token,
        });
      } catch (err) {
        setError(err instanceof Error ? err.message : '搜索简历失败');
        throw err;
      } finally {
        setLoading(false);
      }
    },
    []
  );

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

  const addResumeItem = useCallback(
    (resume: Resume) => {
      console.log('添加简历到列表:', resume.name, '状态:', resume.status);

      setResumes((prev) => {
        const exists = prev.some((item) => item.id === resume.id);

        const nextResumes = exists
          ? prev.map((item) =>
              item.id === resume.id ? { ...item, ...resume } : item
            )
          : [resume, ...prev];

        console.log('更新后的简历列表长度:', nextResumes.length);
        console.log('新简历是否在列表顶部:', nextResumes[0]?.id === resume.id);

        resumesRef.current = nextResumes;
        hasNewResumeRef.current = true;
        scheduleProgressPolling();

        // 更新分页信息以反映新增的简历
        setPagination((prevPagination) => ({
          ...prevPagination,
          total: exists ? prevPagination.total : prevPagination.total + 1,
        }));

        return nextResumes;
      });
    },
    [scheduleProgressPolling]
  );

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
    },
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
    refresh: () => id && fetchResume(id),
  };
};

// 简历上传 Hook（使用批量上传接口）
export const useResumeUpload = () => {
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [taskId, setTaskId] = useState<string | null>(null);
  const [uploadStatus, setUploadStatus] = useState<BatchUploadStatus | null>(
    null
  );
  const pollingTimerRef = useRef<ReturnType<typeof setInterval> | null>(null);

  // 停止轮询
  const stopPolling = useCallback(() => {
    if (pollingTimerRef.current) {
      clearInterval(pollingTimerRef.current);
      pollingTimerRef.current = null;
    }
  }, []);

  // 查询上传状态
  const fetchUploadStatus = useCallback(
    async (taskIdToFetch: string, silent = false) => {
      if (!silent) {
        setUploading(true);
      }
      setError(null);

      try {
        const status = await getBatchUploadStatus(taskIdToFetch);
        setUploadStatus(status);

        // 计算进度百分比
        const progressPercent =
          status.total_count > 0
            ? Math.round((status.completed_count / status.total_count) * 100)
            : 0;
        setUploadProgress(progressPercent);

        console.log('📤 批量上传状态:', {
          taskId: status.task_id,
          status: status.status,
          progress: progressPercent,
          completed: status.completed_count,
          total: status.total_count,
          success: status.success_count,
          failed: status.failed_count,
        });

        // 检查是否完成或失败，停止轮询
        if (
          status.status === 'completed' ||
          status.status === 'failed' ||
          status.status === 'cancelled'
        ) {
          stopPolling();
          setUploading(false);

          if (status.status === 'failed') {
            setError('部分文件上传失败');
          }
        }

        return status;
      } catch (err) {
        const errorMessage =
          err instanceof Error ? err.message : '获取上传状态失败';
        setError(errorMessage);
        stopPolling();
        setUploading(false);
        throw err;
      }
    },
    [stopPolling]
  );

  // 开始轮询上传状态
  const startPolling = useCallback(
    (taskIdToPoll: string, pollingInterval = 1000) => {
      stopPolling();

      console.log('🔄 开始轮询上传状态，任务ID:', taskIdToPoll);

      // 立即获取一次状态
      fetchUploadStatus(taskIdToPoll, true);

      // 开始轮询
      pollingTimerRef.current = setInterval(() => {
        fetchUploadStatus(taskIdToPoll, true);
      }, pollingInterval);
    },
    [fetchUploadStatus, stopPolling]
  );

  // 上传文件（支持单个或多个文件）
  const uploadFile = useCallback(
    async (
      files: File | File[],
      jobPositionIds?: string[],
      position?: string
    ): Promise<{ taskId: string; message: string }> => {
      setUploading(true);
      setError(null);
      setUploadProgress(0);
      setTaskId(null);
      setUploadStatus(null);

      try {
        const fileList = Array.isArray(files) ? files : [files];
        const fileNames = fileList.map((f) => f.name).join(', ');
        console.log(`📤 开始批量上传 ${fileList.length} 个简历:`, fileNames);

        // 调用批量上传接口
        const response = await batchUploadResume(
          files,
          jobPositionIds,
          position
        );

        console.log('📤 批量上传响应:', response);

        setTaskId(response.task_id);

        // 开始轮询上传状态
        startPolling(response.task_id);

        return {
          taskId: response.task_id,
          message: response.message,
        };
      } catch (err) {
        setError(err instanceof Error ? err.message : '上传失败');
        setUploading(false);
        throw err;
      }
    },
    [startPolling]
  );

  // 清理轮询
  useEffect(() => {
    return () => {
      stopPolling();
    };
  }, [stopPolling]);

  return {
    uploading,
    error,
    uploadProgress,
    uploadFile,
    taskId,
    uploadStatus,
    fetchUploadStatus,
    stopPolling,
    isPolling: !!pollingTimerRef.current,
  };
};

// 简历解析进度 Hook
export const useResumeProgress = (
  id?: string,
  options?: {
    autoPolling?: boolean;
    pollingInterval?: number;
    onComplete?: (
      progress: ResumeParseProgress,
      resumeDetail?: ResumeDetail
    ) => void;
    onError?: (error: string) => void;
  }
) => {
  const [progress, setProgress] = useState<ResumeParseProgress | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [resumeDetail, setResumeDetail] = useState<ResumeDetail | null>(null);
  const pollingTimerRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const {
    autoPolling = false,
    pollingInterval = 2000,
    onComplete,
    onError,
  } = options || {};

  const stopPolling = useCallback(() => {
    if (pollingTimerRef.current) {
      clearInterval(pollingTimerRef.current);
      pollingTimerRef.current = null;
    }
  }, []);

  const fetchProgress = useCallback(
    async (resumeId: string, silent = false) => {
      if (!silent) {
        setLoading(true);
      }
      setError(null);

      try {
        const progressData = await getResumeProgress(resumeId);
        setProgress(progressData);

        // 检查是否完成或失败，停止轮询
        if (
          progressData.status === 'completed' ||
          progressData.status === 'failed'
        ) {
          stopPolling();

          if (progressData.status === 'completed') {
            // 解析完成后，获取详细数据
            try {
              const detailData = await getResumeDetail(resumeId);
              setResumeDetail(detailData);
              if (onComplete) {
                onComplete(progressData, detailData);
              }
            } catch (detailError) {
              console.error('获取简历详情失败:', detailError);
              if (onComplete) {
                onComplete(progressData);
              }
            }
          } else if (progressData.status === 'failed' && onError) {
            onError(progressData.error_message || '解析失败');
          }
        }

        return progressData;
      } catch (err) {
        const errorMessage =
          err instanceof Error ? err.message : '获取解析进度失败';
        setError(errorMessage);
        if (onError) {
          onError(errorMessage);
        }
        stopPolling();
      } finally {
        if (!silent) {
          setLoading(false);
        }
      }
    },
    [stopPolling, onComplete, onError]
  );

  const startPolling = useCallback(
    (resumeId: string) => {
      stopPolling();

      // 立即获取一次进度
      fetchProgress(resumeId, true);

      // 开始轮询
      pollingTimerRef.current = setInterval(() => {
        fetchProgress(resumeId, true);
      }, pollingInterval);
    },
    [fetchProgress, pollingInterval, stopPolling]
  );

  const reparse = useCallback(
    async (resumeId: string) => {
      setLoading(true);
      setError(null);

      try {
        await reparseResume(resumeId);
        // 重新获取进度并开始轮询
        await fetchProgress(resumeId);
        if (autoPolling) {
          startPolling(resumeId);
        }
      } catch (err) {
        const errorMessage =
          err instanceof Error ? err.message : '重新解析失败';
        setError(errorMessage);
        if (onError) {
          onError(errorMessage);
        }
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [fetchProgress, autoPolling, startPolling, onError]
  );

  // 手动开始轮询
  const startProgressPolling = useCallback(
    (resumeId?: string) => {
      const targetId = resumeId || id;
      if (targetId) {
        startPolling(targetId);
      }
    },
    [id, startPolling]
  );

  // 清理轮询
  useEffect(() => {
    return () => {
      stopPolling();
    };
  }, [stopPolling]);

  // 初始化和自动轮询
  useEffect(() => {
    if (id) {
      fetchProgress(id);

      // 如果启用自动轮询且状态需要轮询
      if (autoPolling) {
        // 延迟启动轮询，等待初始状态获取完成
        const timer = setTimeout(() => {
          if (
            progress &&
            (progress.status === 'pending' || progress.status === 'processing')
          ) {
            startPolling(id);
          }
        }, 1000);

        return () => clearTimeout(timer);
      }
    }
  }, [id, fetchProgress, autoPolling, startPolling, progress]);

  // 监听进度状态变化，决定是否需要轮询
  useEffect(() => {
    if (autoPolling && id && progress) {
      if (progress.status === 'pending' || progress.status === 'processing') {
        if (!pollingTimerRef.current) {
          startPolling(id);
        }
      } else {
        stopPolling();
      }
    }
  }, [progress, autoPolling, id, startPolling, stopPolling]);

  return {
    progress,
    loading,
    error,
    resumeDetail,
    fetchProgress,
    reparse,
    startProgressPolling,
    stopPolling,
    isPolling: !!pollingTimerRef.current,
    refresh: () => id && fetchProgress(id),
  };
};
