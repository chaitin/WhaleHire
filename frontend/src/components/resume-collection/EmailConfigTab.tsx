import { useState, useCallback, useEffect } from 'react';
import {
  X,
  Plus,
  Edit,
  Trash2,
  ChevronLeft,
  ChevronRight,
  Loader2,
  AlertCircle,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { ConfirmDialog } from '@/components/ui/confirm-dialog';
import { cn } from '@/lib/utils';
import { getCurrentUser } from '@/services/auth';
import { listJobProfiles } from '@/services/job-profile';
import type { JobProfileDetail } from '@/types/job-profile';

// 邮箱配置类型定义
interface EmailConfigFormData {
  task_name: string;
  email_address: string;
  auth_code: string;
  server_type: string;
  imap_server: string;
  imap_port: number;
  use_ssl: boolean;
  status: 'enabled' | 'disabled';
  resume_uploader: string;
  uploader_id: string;
  job_profile_id: string;
}

interface EmailConfig {
  id: string;
  task_name: string;
  email_address: string;
  status: 'enabled' | 'disabled';
  last_sync_time: number;
  synced_count: number;
  resume_uploader: string;
  job_profile_name: string;
  server_type: string;
  imap_server: string;
  imap_port: number;
  use_ssl: boolean;
  auth_code: string;
  created_at: number;
}

export function EmailConfigTab() {
  // 邮箱配置列表数据
  const [emailConfigs, setEmailConfigs] = useState<EmailConfig[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [totalCount, setTotalCount] = useState(0);
  const [pageSize] = useState(7);

  // 添加邮箱配置弹窗状态
  const [isAddModalOpen, setIsAddModalOpen] = useState(false);

  // 表单数据
  const [formData, setFormData] = useState<EmailConfigFormData>({
    task_name: '',
    email_address: '',
    auth_code: '',
    server_type: 'IMAP',
    imap_server: '',
    imap_port: 993,
    use_ssl: true,
    status: 'enabled',
    resume_uploader: '',
    uploader_id: '',
    job_profile_id: '',
  });

  // 岗位画像列表
  const [jobProfiles, setJobProfiles] = useState<JobProfileDetail[]>([]);
  const [loadingJobProfiles, setLoadingJobProfiles] = useState(false);

  // 删除确认弹窗状态
  const [isDeleteConfirmOpen, setIsDeleteConfirmOpen] = useState(false);
  const [configToDelete, setConfigToDelete] = useState<{
    id: string;
    task_name: string;
  } | null>(null);

  // 操作状态
  const [submitting, setSubmitting] = useState(false);
  const [deletingId, setDeletingId] = useState<string | null>(null);

  // Mock数据 - 获取邮箱配置列表
  const fetchEmailConfigs = useCallback(
    async (page?: number) => {
      try {
        setLoading(true);
        setError(null);
        const targetPage = page || currentPage;

        // 模拟API延迟
        await new Promise((resolve) => setTimeout(resolve, 300));

        // Mock数据
        const mockData: EmailConfig[] = [
          {
            id: '1',
            task_name: '公司招聘邮箱收集',
            email_address: 'hr@company.com',
            status: 'enabled',
            last_sync_time: Date.now() / 1000 - 3600, // 1小时前
            synced_count: 156,
            resume_uploader: '张三',
            job_profile_name: '高级前端工程师',
            server_type: 'IMAP',
            imap_server: 'imap.company.com',
            imap_port: 993,
            use_ssl: true,
            auth_code: '****',
            created_at: Date.now() / 1000 - 86400 * 30, // 30天前
          },
          {
            id: '2',
            task_name: '备用招聘邮箱',
            email_address: 'recruitment@company.com',
            status: 'disabled',
            last_sync_time: Date.now() / 1000 - 86400, // 1天前
            synced_count: 89,
            resume_uploader: '李四',
            job_profile_name: 'Java开发工程师',
            server_type: 'IMAP',
            imap_server: 'imap.company.com',
            imap_port: 993,
            use_ssl: true,
            auth_code: '****',
            created_at: Date.now() / 1000 - 86400 * 15, // 15天前
          },
        ];

        const total = mockData.length;
        const newTotalPages = Math.ceil(total / pageSize);
        const startIndex = (targetPage - 1) * pageSize;
        const endIndex = startIndex + pageSize;
        const paginatedData = mockData.slice(startIndex, endIndex);

        setEmailConfigs(paginatedData);
        setTotalCount(total);
        setTotalPages(newTotalPages);

        if (targetPage > newTotalPages && newTotalPages > 0) {
          setCurrentPage(newTotalPages);
          await fetchEmailConfigs(newTotalPages);
          return;
        }
      } catch (err) {
        setError('获取邮箱配置失败，请重试');
        console.error('获取邮箱配置失败:', err);
        setEmailConfigs([]);
        setTotalCount(0);
        setTotalPages(0);
      } finally {
        setLoading(false);
      }
    },
    [currentPage, pageSize]
  );

  // 获取当前登录用户信息
  const fetchCurrentUser = useCallback(async () => {
    try {
      const user = await getCurrentUser();
      setFormData((prev) => ({
        ...prev,
        resume_uploader: user.username,
        uploader_id: user.id,
      }));
    } catch (err) {
      console.error('获取当前用户信息失败:', err);
    }
  }, []);

  // 获取岗位画像列表
  const fetchJobProfiles = useCallback(async () => {
    try {
      setLoadingJobProfiles(true);
      const response = await listJobProfiles({ page: 1, page_size: 100 });
      setJobProfiles(response.items || []);
    } catch (err) {
      console.error('获取岗位画像列表失败:', err);
      setJobProfiles([]);
    } finally {
      setLoadingJobProfiles(false);
    }
  }, []);

  // 打开添加弹窗
  const handleOpenAddModal = async () => {
    setIsAddModalOpen(true);
    // 获取当前用户信息和岗位画像列表
    await Promise.all([fetchCurrentUser(), fetchJobProfiles()]);
  };

  // 关闭添加弹窗
  const handleCloseAddModal = () => {
    setIsAddModalOpen(false);
    // 重置表单
    setFormData({
      task_name: '',
      email_address: '',
      auth_code: '',
      server_type: 'IMAP',
      imap_server: '',
      imap_port: 993,
      use_ssl: true,
      status: 'enabled',
      resume_uploader: '',
      uploader_id: '',
      job_profile_id: '',
    });
  };

  // 表单输入变化
  const handleFormChange = (
    field: string,
    value: string | number | boolean | string[]
  ) => {
    setFormData((prev) => ({
      ...prev,
      [field]: value,
    }));
  };

  // 保存邮箱配置
  const handleSaveEmailConfig = async () => {
    if (!formData.task_name.trim()) {
      setError('任务名称不能为空');
      return;
    }

    if (!formData.email_address.trim()) {
      setError('邮箱地址不能为空');
      return;
    }

    if (!formData.auth_code.trim()) {
      setError('邮箱授权码不能为空');
      return;
    }

    if (!formData.imap_server.trim()) {
      setError('服务器地址不能为空');
      return;
    }

    if (!formData.job_profile_id) {
      setError('请选择岗位画像');
      return;
    }

    try {
      setSubmitting(true);
      setError(null);

      // 模拟API调用
      await new Promise((resolve) => setTimeout(resolve, 500));

      console.log('保存邮箱配置:', formData);

      handleCloseAddModal();
      setCurrentPage(1);
      await fetchEmailConfigs(1);
    } catch (err) {
      setError('保存邮箱配置失败，请重试');
      console.error('保存邮箱配置失败:', err);
    } finally {
      setSubmitting(false);
    }
  };

  // 删除邮箱配置
  const handleDeleteConfig = (configId: string, taskName: string) => {
    setConfigToDelete({ id: configId, task_name: taskName });
    setIsDeleteConfirmOpen(true);
  };

  // 确认删除
  const handleConfirmDelete = async () => {
    if (!configToDelete) return;

    try {
      setDeletingId(configToDelete.id);
      setError(null);

      // 模拟API调用
      await new Promise((resolve) => setTimeout(resolve, 500));

      console.log(`删除邮箱配置: ${configToDelete.id}`);

      setIsDeleteConfirmOpen(false);
      setConfigToDelete(null);
      await fetchEmailConfigs(currentPage);
    } catch (err) {
      setError('删除失败，请重试');
      console.error('删除失败:', err);
    } finally {
      setDeletingId(null);
    }
  };

  // 页码变化
  const handlePageChange = (page: number) => {
    if (page < 1 || page > totalPages) return;
    setCurrentPage(page);
  };

  // 生成页码数组
  const generatePageNumbers = (): number[] => {
    const pages: number[] = [];
    const maxPagesToShow = 5;

    if (totalPages <= maxPagesToShow) {
      for (let i = 1; i <= totalPages; i++) {
        pages.push(i);
      }
    } else {
      if (currentPage <= 3) {
        for (let i = 1; i <= 4; i++) {
          pages.push(i);
        }
        pages.push(totalPages);
      } else if (currentPage >= totalPages - 2) {
        pages.push(1);
        for (let i = totalPages - 3; i <= totalPages; i++) {
          pages.push(i);
        }
      } else {
        pages.push(1);
        pages.push(currentPage - 1);
        pages.push(currentPage);
        pages.push(currentPage + 1);
        pages.push(totalPages);
      }
    }

    return pages;
  };

  // 格式化日期时间
  const formatDateTime = (timestamp: number) => {
    const date = new Date(timestamp * 1000);
    return date.toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  };

  // 截断邮箱地址,只显示前10位
  const truncateEmail = (email: string) => {
    if (email.length <= 10) {
      return email;
    }
    return email.substring(0, 10) + '...';
  };

  // 获取状态显示
  const getStatusDisplay = (status: string) => {
    switch (status) {
      case 'enabled':
        return {
          text: '启用',
          bgColor: 'bg-green-50',
          textColor: 'text-green-600',
        };
      case 'disabled':
        return {
          text: '禁用',
          bgColor: 'bg-gray-50',
          textColor: 'text-gray-600',
        };
      default:
        return {
          text: '未知',
          bgColor: 'bg-gray-50',
          textColor: 'text-gray-600',
        };
    }
  };

  // 页面切换时获取数据
  useEffect(() => {
    fetchEmailConfigs();
  }, [currentPage, fetchEmailConfigs]);

  return (
    <>
      {/* 选项卡内容 */}
      <div className="p-6">
        {/* 错误提示 */}
        {error && (
          <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-md">
            <div className="flex items-center">
              <AlertCircle className="w-4 h-4 text-red-500 mr-2" />
              <span className="text-sm text-red-700">{error}</span>
            </div>
          </div>
        )}

        {/* 添加配置按钮 - 左侧显示 */}
        <div className="mb-6">
          <Button
            size="sm"
            className="gap-1.5 rounded-lg px-3 py-1.5 shadow-sm"
            onClick={handleOpenAddModal}
            disabled={submitting}
          >
            <Plus className="w-3.5 h-3.5" />
            添加邮箱配置
          </Button>
        </div>

        {/* 配置列表表格 */}
        <div className="border border-gray-200 rounded-lg overflow-hidden">
          {/* 表格头部 */}
          <div className="bg-gray-50 border-b border-gray-200">
            <div className="flex items-center px-6 py-3 w-full">
              <div className="flex items-center w-40 pr-4">
                <span className="text-sm font-medium text-gray-700">
                  任务名称
                </span>
              </div>
              <div className="flex items-center w-48 pr-4">
                <span className="text-sm font-medium text-gray-700">
                  邮箱地址
                </span>
              </div>
              <div className="flex items-center w-24 pr-4">
                <span className="text-sm font-medium text-gray-700">状态</span>
              </div>
              <div className="flex items-center w-40 pr-4">
                <span className="text-sm font-medium text-gray-700">
                  最后同步时间
                </span>
              </div>
              <div className="flex items-center w-24 pr-4">
                <span className="text-sm font-medium text-gray-700">
                  已同步数
                </span>
              </div>
              <div className="flex items-center w-32 pr-4">
                <span className="text-sm font-medium text-gray-700">
                  简历上传人
                </span>
              </div>
              <div className="flex items-center w-40 pr-4">
                <span className="text-sm font-medium text-gray-700">
                  岗位名称
                </span>
              </div>
              <div className="flex items-center justify-center w-24">
                <span className="text-sm font-medium text-gray-700">操作</span>
              </div>
            </div>
          </div>

          {/* 表格内容 */}
          <div className="bg-white">
            {loading ? (
              <div className="flex items-center justify-center py-8">
                <Loader2 className="w-6 h-6 animate-spin text-gray-400" />
                <span className="ml-2 text-sm text-gray-500">加载中...</span>
              </div>
            ) : emailConfigs.length === 0 ? (
              <div className="flex items-center justify-center py-8">
                <span className="text-sm text-gray-500">暂无配置数据</span>
              </div>
            ) : (
              emailConfigs.map((config, index) => {
                const statusDisplay = getStatusDisplay(config.status);
                return (
                  <div
                    key={config.id}
                    className={`flex items-center px-6 py-4 w-full ${index !== emailConfigs.length - 1 ? 'border-b border-gray-100' : ''}`}
                  >
                    {/* 任务名称 */}
                    <div className="flex items-center w-40 pr-4">
                      <span className="text-sm text-gray-900 truncate">
                        {config.task_name}
                      </span>
                    </div>

                    {/* 邮箱地址 */}
                    <div className="flex items-center w-48 pr-4">
                      <span
                        className="text-sm text-gray-900"
                        title={config.email_address}
                      >
                        {truncateEmail(config.email_address)}
                      </span>
                    </div>

                    {/* 状态 */}
                    <div className="flex items-center w-24 pr-4">
                      <div
                        className={`px-3 py-1 rounded ${statusDisplay.bgColor}`}
                      >
                        <span
                          className={`text-xs font-medium ${statusDisplay.textColor}`}
                        >
                          {statusDisplay.text}
                        </span>
                      </div>
                    </div>

                    {/* 最后同步时间 */}
                    <div className="flex items-center w-40 pr-4">
                      <span className="text-xs text-gray-500">
                        {formatDateTime(config.last_sync_time)}
                      </span>
                    </div>

                    {/* 已同步数 */}
                    <div className="flex items-center w-24 pr-4">
                      <span className="text-sm text-gray-900">
                        {config.synced_count}
                      </span>
                    </div>

                    {/* 简历上传人 */}
                    <div className="flex items-center w-32 pr-4">
                      <span className="text-sm text-gray-500 truncate">
                        {config.resume_uploader}
                      </span>
                    </div>

                    {/* 岗位名称 */}
                    <div className="flex items-center w-40 pr-4">
                      <span className="text-sm text-gray-500 truncate">
                        {config.job_profile_name}
                      </span>
                    </div>

                    {/* 操作 */}
                    <div className="flex items-center justify-center gap-2 w-24">
                      <button
                        className="text-gray-500 hover:text-blue-500 transition-colors"
                        title="编辑"
                        disabled={submitting}
                      >
                        <Edit className="w-4 h-4" />
                      </button>

                      <button
                        className="text-gray-500 hover:text-red-500 transition-colors"
                        title="删除"
                        onClick={() =>
                          handleDeleteConfig(config.id, config.task_name)
                        }
                        disabled={submitting || deletingId === config.id}
                      >
                        {deletingId === config.id ? (
                          <Loader2 className="w-4 h-4 animate-spin" />
                        ) : (
                          <Trash2 className="w-4 h-4" />
                        )}
                      </button>
                    </div>
                  </div>
                );
              })
            )}
          </div>
        </div>

        {/* 分页区域 */}
        {!loading && (
          <div className="flex flex-col gap-6 lg:flex-row lg:items-center lg:justify-between mt-6">
            <div className="flex items-center gap-4">
              <div className="text-sm text-[#6B7280]">
                显示{' '}
                <span className="text-[#6B7280]">
                  {totalCount > 0 ? (currentPage - 1) * pageSize + 1 : 0}
                </span>{' '}
                到{' '}
                <span className="text-[#6B7280]">
                  {totalCount > 0
                    ? Math.min(currentPage * pageSize, totalCount)
                    : 0}
                </span>{' '}
                条，共 <span className="text-[#6B7280]">{totalCount}</span>{' '}
                条结果
              </div>
            </div>

            {/* 分页按钮 */}
            {totalCount > 0 && totalPages > 1 && (
              <div className="flex flex-wrap items-center gap-1">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handlePageChange(currentPage - 1)}
                  disabled={currentPage === 1}
                  className={cn(
                    'h-[34px] w-[34px] rounded border border-[#D1D5DB] bg-white p-0',
                    currentPage === 1 && 'opacity-50'
                  )}
                >
                  <ChevronLeft className="h-4 w-4 text-[#6B7280]" />
                </Button>

                {generatePageNumbers().map((page) => (
                  <Button
                    key={page}
                    variant="outline"
                    size="sm"
                    onClick={() => handlePageChange(page)}
                    className={cn(
                      'h-[34px] w-[34px] rounded border border-[#D1D5DB] bg-white p-0 text-sm font-normal text-[#374151]',
                      page === currentPage &&
                        'border-[#36CFC9] bg-[#36CFC9] text-white'
                    )}
                  >
                    {page}
                  </Button>
                ))}

                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handlePageChange(currentPage + 1)}
                  disabled={currentPage >= totalPages}
                  className={cn(
                    'h-[34px] w-[34px] rounded border border-[#D1D5DB] bg-white p-0',
                    currentPage >= totalPages && 'opacity-50'
                  )}
                >
                  <ChevronRight className="h-4 w-4 text-[#6B7280]" />
                </Button>
              </div>
            )}
          </div>
        )}
      </div>

      {/* 添加邮箱配置弹窗 */}
      {isAddModalOpen && (
        <AddEmailConfigModal
          isOpen={isAddModalOpen}
          onClose={handleCloseAddModal}
          formData={formData}
          onFormChange={handleFormChange}
          onSave={handleSaveEmailConfig}
          submitting={submitting}
          error={error}
          jobProfiles={jobProfiles}
          loadingJobProfiles={loadingJobProfiles}
        />
      )}

      {/* 删除确认弹窗 */}
      <ConfirmDialog
        open={isDeleteConfirmOpen}
        onOpenChange={setIsDeleteConfirmOpen}
        title="确认删除"
        description={
          configToDelete ? (
            <span>
              确定要删除 <strong>{configToDelete.task_name}</strong>{' '}
              配置吗？此操作无法撤销。
            </span>
          ) : (
            ''
          )
        }
        confirmText="删除"
        cancelText="取消"
        onConfirm={handleConfirmDelete}
        variant="destructive"
        loading={configToDelete ? deletingId === configToDelete.id : false}
        zIndex="z-[1002]"
      />
    </>
  );
}

// 添加邮箱配置弹窗组件
interface AddEmailConfigModalProps {
  isOpen: boolean;
  onClose: () => void;
  formData: EmailConfigFormData;
  onFormChange: (
    field: string,
    value: string | number | boolean | string[]
  ) => void;
  onSave: () => void;
  submitting: boolean;
  error: string | null;
  jobProfiles: JobProfileDetail[];
  loadingJobProfiles: boolean;
}

function AddEmailConfigModal({
  isOpen,
  onClose,
  formData,
  onFormChange,
  onSave,
  submitting,
  error: _error,
  jobProfiles,
  loadingJobProfiles,
}: AddEmailConfigModalProps) {
  if (!isOpen) return null;

  return (
    <div
      className="fixed inset-0 flex items-center justify-center bg-black bg-opacity-50 z-[1001]"
      onClick={onClose}
    >
      <div
        className="bg-white rounded-lg shadow-xl w-[700px] max-h-[90vh] overflow-hidden"
        onClick={(e) => e.stopPropagation()}
      >
        {/* 弹窗头部 */}
        <div className="flex items-center justify-between px-6 py-5 border-b border-gray-100">
          <h2 className="text-lg font-medium text-gray-900">添加邮箱配置</h2>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600 transition-colors w-6 h-6"
          >
            <X className="w-6 h-6" />
          </button>
        </div>

        {/* 表单内容 - 滚动区域 */}
        <div className="p-6 overflow-y-auto max-h-[calc(90vh-200px)] space-y-6">
          {/* 任务名称 */}
          <div className="mb-6">
            <label className="block text-sm font-medium text-gray-700 mb-2">
              任务名称 <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              value={formData.task_name}
              onChange={(e) => onFormChange('task_name', e.target.value)}
              placeholder="请输入任务名称"
              className="w-full h-10 px-3 py-2 border border-gray-200 rounded-md focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent"
            />
          </div>

          {/* 邮箱地址 */}
          <div className="mb-6">
            <label className="block text-sm font-medium text-gray-700 mb-2">
              邮箱地址 <span className="text-red-500">*</span>
            </label>
            <input
              type="email"
              value={formData.email_address}
              onChange={(e) => onFormChange('email_address', e.target.value)}
              placeholder="example@company.com"
              className="w-full h-10 px-3 py-2 border border-gray-200 rounded-md focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent"
            />
          </div>

          {/* 邮箱授权码 */}
          <div className="mb-6">
            <label className="block text-sm font-medium text-gray-700 mb-2">
              邮箱授权码 <span className="text-red-500">*</span>
            </label>
            <input
              type="password"
              value={formData.auth_code}
              onChange={(e) => onFormChange('auth_code', e.target.value)}
              placeholder="请输入邮箱授权码"
              className="w-full h-10 px-3 py-2 border border-gray-200 rounded-md focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent"
            />
            <p className="mt-1 text-xs text-gray-500">
              请到邮箱设置中开启IMAP服务并获取授权码
            </p>
          </div>

          {/* 服务器类型 */}
          <div className="mb-6">
            <label className="block text-sm font-medium text-gray-700 mb-2">
              服务器类型 <span className="text-red-500">*</span>
            </label>
            <select
              value={formData.server_type}
              onChange={(e) => onFormChange('server_type', e.target.value)}
              className="w-full h-10 px-3 py-2 border border-gray-200 rounded-md focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent"
            >
              <option value="IMAP">IMAP</option>
            </select>
          </div>

          {/* 服务器地址、端口 */}
          <div className="mb-6 grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                服务器地址 <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                value={formData.imap_server}
                onChange={(e) => onFormChange('imap_server', e.target.value)}
                placeholder="imap.example.com"
                className="w-full h-10 px-3 py-2 border border-gray-200 rounded-md focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                端口 <span className="text-red-500">*</span>
              </label>
              <input
                type="number"
                value={formData.imap_port}
                onChange={(e) =>
                  onFormChange('imap_port', parseInt(e.target.value) || 993)
                }
                placeholder="993"
                className="w-full h-10 px-3 py-2 border border-gray-200 rounded-md focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent"
              />
            </div>
          </div>

          {/* SSL/TLS加密传输 */}
          <div className="mb-6 flex items-center">
            <input
              type="checkbox"
              id="use_ssl"
              checked={formData.use_ssl}
              onChange={(e) => onFormChange('use_ssl', e.target.checked)}
              className="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-2 focus:ring-emerald-500"
            />
            <label
              htmlFor="use_ssl"
              className="ml-2 text-sm font-medium text-gray-700"
            >
              SSL/TLS加密传输
            </label>
          </div>

          {/* 状态 */}
          <div className="mb-6">
            <label className="block text-sm font-medium text-gray-700 mb-2">
              状态 <span className="text-red-500">*</span>
            </label>
            <div className="flex items-center gap-6">
              <label className="flex items-center">
                <input
                  type="radio"
                  value="enabled"
                  checked={formData.status === 'enabled'}
                  onChange={(e) => onFormChange('status', e.target.value)}
                  className="w-4 h-4 text-emerald-600 border-gray-300 focus:ring-2 focus:ring-emerald-500"
                />
                <span className="ml-2 text-sm font-medium text-gray-700">
                  启用
                </span>
              </label>

              <label className="flex items-center">
                <input
                  type="radio"
                  value="disabled"
                  checked={formData.status === 'disabled'}
                  onChange={(e) => onFormChange('status', e.target.value)}
                  className="w-4 h-4 text-emerald-600 border-gray-300 focus:ring-2 focus:ring-emerald-500"
                />
                <span className="ml-2 text-sm font-medium text-gray-700">
                  禁用
                </span>
              </label>
            </div>
          </div>

          {/* 简历创建人 */}
          <div className="mb-6">
            <label className="block text-sm font-medium text-gray-700 mb-2">
              简历创建人 <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              value={formData.resume_uploader}
              onChange={(e) => onFormChange('resume_uploader', e.target.value)}
              placeholder="默认为当前登录人"
              className="w-full h-10 px-3 py-2 border border-gray-200 rounded-md focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent"
            />
          </div>

          {/* 岗位画像 */}
          <div className="mb-6">
            <label className="block text-sm font-medium text-gray-700 mb-2">
              岗位画像 <span className="text-red-500">*</span>
            </label>
            <select
              value={formData.job_profile_id}
              onChange={(e) => onFormChange('job_profile_id', e.target.value)}
              disabled={loadingJobProfiles}
              className="w-full h-10 px-3 py-2 border border-gray-200 rounded-md focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent"
            >
              <option value="">
                {loadingJobProfiles ? '加载岗位中...' : '请选择岗位画像'}
              </option>
              {jobProfiles.map((job) => (
                <option key={job.id} value={job.id}>
                  {job.name}
                </option>
              ))}
            </select>
          </div>
        </div>

        {/* 弹窗底部 */}
        <div className="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-100">
          <Button variant="outline" className="px-6" onClick={onClose}>
            取消
          </Button>
          <Button className="px-6" onClick={onSave} disabled={submitting}>
            {submitting ? (
              <>
                <Loader2 className="w-4 h-4 animate-spin mr-2" />
                保存中...
              </>
            ) : (
              '保存配置'
            )}
          </Button>
        </div>
      </div>
    </div>
  );
}
