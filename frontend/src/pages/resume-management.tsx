import { useState, useEffect, useRef } from 'react';
import { Upload, AlertCircle, CheckCircle } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { ResumeFiltersComponent } from '@/components/resume/resume-filters';
import { ResumeTable } from '@/components/resume/resume-table';
import { UploadResumeModal } from '@/components/modals/upload-resume-modal';
import { EditResumeModal } from '@/components/modals/edit-resume-modal';
import { ResumeFilters, Resume } from '@/types/resume';
import { useResumeList } from '@/hooks/useResume';
import { Alert, AlertDescription } from '@/components/ui/alert';

export function ResumeManagementPage() {
  const [filters, setFilters] = useState<ResumeFilters>({
    position: 'all',
    status: 'all',
    keywords: '',
  });
  const [queryFilters, setQueryFilters] = useState<ResumeFilters>({
    position: 'all',
    status: 'all',
    keywords: '',
  });
  const [currentPage, setCurrentPage] = useState(1);
  const [isUploadModalOpen, setIsUploadModalOpen] = useState(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [editingResume, setEditingResume] = useState<Resume | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const [hasProcessingResumes, setHasProcessingResumes] = useState(false);
  const autoRefreshIntervalRef = useRef<NodeJS.Timeout | null>(null);
  const fetchResumesRef = useRef<typeof fetchResumes | null>(null);
  const previousResumesRef = useRef<Resume[]>([]);
  const refreshCountRef = useRef(0);

  // 使用简历管理Hook
  const {
    resumes,
    pagination,
    loading: isLoading,
    error,
    fetchResumes,
    updateResumeItem: updateResumeRecord,
    deleteResumeItem: deleteResumeRecord,
    searchResumeList,
    addResumeItem,
  } = useResumeList();

  // 保存最新的 fetchResumes 函数引用
  useEffect(() => {
    fetchResumesRef.current = fetchResumes;
  }, [fetchResumes]);

  // 清除错误
  const clearError = () => {
    // 这里可以添加清除错误的逻辑
    console.log('清除错误');
  };

  // 清除成功消息
  const clearSuccessMessage = () => {
    setSuccessMessage(null);
  };

  // 当筛选条件改变时，重新获取数据
  useEffect(() => {
    const params = {
      page: currentPage,
      size: pagination.pageSize,
      position: queryFilters.position,
      status: queryFilters.status,
      keywords: queryFilters.keywords?.trim() || undefined,
    };
    fetchResumes(params);
  }, [queryFilters, currentPage, fetchResumes, pagination.pageSize]);

  // 基于简历状态变化的智能自动刷新逻辑
  useEffect(() => {
    // 检查是否有处理中的简历
    const processingResumes = resumes.some(resume => 
      resume.status === 'pending' || resume.status === 'processing'
    );

    // 检测状态变化
    const hasStatusChanged = previousResumesRef.current.length > 0 && 
      resumes.some((resume) => {
        const prevResume = previousResumesRef.current.find(prev => prev.id === resume.id);
        return prevResume && prevResume.status !== resume.status;
      });

    if (hasStatusChanged) {
      console.log('检测到简历状态变化');
      refreshCountRef.current = 0; // 重置刷新计数
    }

    // 更新处理状态
    setHasProcessingResumes(processingResumes);

    // 保存当前简历状态用于下次比较
    previousResumesRef.current = [...resumes];

    // 清除现有的定时器
    if (autoRefreshIntervalRef.current) {
      clearInterval(autoRefreshIntervalRef.current);
      autoRefreshIntervalRef.current = null;
    }

    // 只有当存在处理中的简历时才启动自动刷新
    if (processingResumes && fetchResumesRef.current) {
      console.log('检测到处理中的简历，启动自动刷新监控');
      refreshCountRef.current = 0;
      
      autoRefreshIntervalRef.current = setInterval(() => {
        if (fetchResumesRef.current) {
          refreshCountRef.current += 1;
          
          const params = {
            page: currentPage,
            size: pagination.pageSize,
            position: queryFilters.position !== 'all' ? queryFilters.position : undefined,
            status: queryFilters.status !== 'all' ? queryFilters.status : undefined,
            keywords: queryFilters.keywords?.trim() || undefined,
          };
          
          console.log(`自动刷新简历列表 (第${refreshCountRef.current}次)，检查状态变化`);
          fetchResumesRef.current(params, { force: true }); // 强制刷新以获取最新状态
          
          // 防止无限刷新，最多刷新60次（12分钟）
          if (refreshCountRef.current >= 60) {
            console.log('达到最大刷新次数，停止自动刷新');
            if (autoRefreshIntervalRef.current) {
              clearInterval(autoRefreshIntervalRef.current);
              autoRefreshIntervalRef.current = null;
            }
          }
        }
      }, 12000); // 12秒刷新一次，检查状态变化
    } else if (!processingResumes) {
      console.log('所有简历处理完成，停止自动刷新');
      refreshCountRef.current = 0;
    }

    // 清理函数
    return () => {
      if (autoRefreshIntervalRef.current) {
        clearInterval(autoRefreshIntervalRef.current);
        autoRefreshIntervalRef.current = null;
      }
    };
  }, [resumes, currentPage, pagination.pageSize, queryFilters]);

  const handleFiltersChange = (
    newFilters: ResumeFilters,
    options?: { shouldSearch?: boolean }
  ) => {
    setFilters(newFilters);

    if (options?.shouldSearch) {
      setCurrentPage(1);
      setQueryFilters({
        ...newFilters,
        keywords: newFilters.keywords?.trim() ?? '',
      });
    }
  };

  const handleSearch = async () => {
    setCurrentPage(1);
    const searchKeywords = filters.keywords?.trim();
    
    try {
      if (searchKeywords) {
        // 有搜索关键词时，调用搜索接口
        const searchParams = {
          keywords: searchKeywords,
          page: 1,
          size: pagination.pageSize,
          position: filters.position !== 'all' ? filters.position : undefined,
          status: filters.status !== 'all' ? filters.status : undefined,
        };
        
        // 使用Hook中的搜索方法
        await searchResumeList(searchParams);
        
        // 更新查询条件状态
        setQueryFilters({
          ...filters,
          keywords: searchKeywords,
        });
      } else {
        // 没有搜索关键词时，调用普通列表接口获取所有数据
        const params = {
          page: 1,
          size: pagination.pageSize,
          position: filters.position !== 'all' ? filters.position : undefined,
          status: filters.status !== 'all' ? filters.status : undefined,
        };
        
        await fetchResumes(params);
        
        // 更新查询条件状态
        setQueryFilters({
          ...filters,
          keywords: '',
        });
      }
    } catch (error) {
      console.error('搜索失败:', error);
    }
  };

  const handlePageChange = (page: number) => {
    setCurrentPage(page);
  };

  const handleEdit = (resume: Resume) => {
    setEditingResume(resume);
    setIsEditModalOpen(true);
  };

  const handleSaveEdit = async (updatedResume: Resume) => {
    try {
      await updateResumeRecord(updatedResume.id, updatedResume);
      setIsEditModalOpen(false);
      setEditingResume(null);
      // 编辑成功后强制刷新列表数据
      const params = {
        page: currentPage,
        size: pagination.pageSize,
        position: queryFilters.position,
        status: queryFilters.status,
        keywords: queryFilters.keywords?.trim() || undefined,
      };
      fetchResumes(params, { force: true });
    } catch (error) {
      console.error('更新简历失败:', error);
    }
  };

  // 原来的定时刷新逻辑已移除，现在使用基于状态变化的刷新机制

  const handleUploadSuccess = (uploadedResume?: Resume) => {
    console.log('简历上传成功，返回的数据:', uploadedResume);
    
    // 显示成功提示
    setSuccessMessage('简历上传成功！正在解析中...');
    
    // 如果有返回的简历数据，立即添加到列表顶部
    if (uploadedResume) {
      console.log('将新简历添加到列表顶部:', uploadedResume.name, '状态:', uploadedResume.status);
      
      // 使用 addResumeItem 方法将新简历添加到列表
      addResumeItem(uploadedResume);
      
      // 重置到第一页以确保能看到新上传的简历
      setCurrentPage(1);
      
      // 重置筛选条件到默认状态，确保新简历能被显示
      if (queryFilters.status !== 'all') {
        setQueryFilters(prev => ({ ...prev, status: 'all' }));
        setFilters(prev => ({ ...prev, status: 'all' }));
      }
    } else {
      console.log('未返回简历数据，强制刷新列表');
      
      // 如果没有返回数据，则强制刷新列表
      const params = {
        page: 1, // 重置到第一页
        size: pagination.pageSize,
        position: queryFilters.position !== 'all' ? queryFilters.position : undefined,
        status: undefined, // 重置状态筛选，确保能看到新上传的简历
        keywords: queryFilters.keywords?.trim() || undefined,
      };
      
      // 更新当前页码状态和筛选条件
      setCurrentPage(1);
      setQueryFilters(prev => ({ ...prev, status: 'all' }));
      setFilters(prev => ({ ...prev, status: 'all' }));
      
      // 强制刷新列表
      fetchResumes(params, { force: true });
    }
    
    console.log('上传成功处理完成，自动刷新机制将监控状态变化');
    
    // 3秒后自动隐藏成功提示
    setTimeout(() => {
      setSuccessMessage(null);
    }, 3000);
  };

  const handleDelete = async (id: string) => {
    try {
      await deleteResumeRecord(id);
      // 删除成功后强制刷新列表数据
      const params = {
        page: currentPage,
        size: pagination.pageSize,
        position: queryFilters.position,
        status: queryFilters.status,
        keywords: queryFilters.keywords?.trim() || undefined,
      };
      fetchResumes(params, { force: true });
    } catch (error) {
      console.error('删除简历失败:', error);
    }
  };

  return (
    <div className="flex h-full flex-col gap-6 px-6 pb-6 pt-6">
      {/* 成功提示 */}
      {successMessage && (
        <Alert className="border-green-200 bg-green-50">
          <CheckCircle className="h-4 w-4 text-green-600" />
          <AlertDescription className="text-green-800 flex items-center justify-between">
            <span>{successMessage}</span>
            <Button variant="ghost" size="sm" onClick={clearSuccessMessage} className="text-green-600 hover:text-green-700">
              关闭
            </Button>
          </AlertDescription>
        </Alert>
      )}

      {/* 错误提示 */}
      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription className="flex items-center justify-between">
            <span>{error}</span>
            <Button variant="ghost" size="sm" onClick={clearError} className="text-red-600 hover:text-red-700">
              关闭
            </Button>
          </AlertDescription>
        </Alert>
      )}

      {/* 页面标题卡片 */}
      <div className="flex items-center justify-between rounded-xl border border-[#E5E7EB] bg-white px-6 py-5 shadow-sm">
        <div>
          <h1 className="text-2xl font-semibold text-gray-900">简历管理</h1>
          <div className="flex items-center gap-3">
            <p className="text-sm text-muted-foreground">管理和查看所有投递的简历</p>
            {hasProcessingResumes && (
              <div className="flex items-center gap-1.5 text-xs text-blue-600 bg-blue-50 px-2 py-1 rounded-full">
                <div className="w-1.5 h-1.5 bg-blue-600 rounded-full animate-pulse"></div>
                正在解析简历... ({resumes.filter(r => r.status === 'pending' || r.status === 'processing').length}个)
              </div>
            )}
          </div>
        </div>
        <Button
          className="gap-2 rounded-lg px-5 py-2 shadow-sm"
          onClick={() => setIsUploadModalOpen(true)}
          disabled={isLoading}
        >
          <Upload className="h-4 w-4" />
          上传简历
        </Button>
      </div>

      {/* 筛选和列表组合卡片 */}
      <div className="rounded-xl border border-[#E5E7EB] bg-white shadow-sm overflow-hidden">
        <div className="border-b border-[#E5E7EB] px-6 py-6">
          <ResumeFiltersComponent
            filters={filters}
            onFiltersChange={handleFiltersChange}
            onSearch={handleSearch}
          />
        </div>
        <ResumeTable
          resumes={resumes}
          pagination={pagination}
          onPageChange={handlePageChange}
          onEdit={handleEdit}
          onDelete={handleDelete}
          isLoading={isLoading}
          className="px-6 py-6"
        />
      </div>

      {/* 上传简历弹窗 */}
      <UploadResumeModal
        open={isUploadModalOpen}
        onOpenChange={setIsUploadModalOpen}
        onSuccess={handleUploadSuccess}
      />

      {/* 编辑简历弹窗 */}
      <EditResumeModal
        isOpen={isEditModalOpen}
        onClose={() => setIsEditModalOpen(false)}
        resume={editingResume}
        onSave={handleSaveEdit}
      />
    </div>
  );
}
