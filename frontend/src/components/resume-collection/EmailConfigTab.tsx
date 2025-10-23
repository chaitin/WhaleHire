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
  RefreshCw,
  Eye,
  EyeOff,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { ConfirmDialog } from '@/components/ui/confirm-dialog';
import { MultiSelect, Option } from '@/components/ui/multi-select';
import { cn } from '@/lib/utils';
import { getCurrentUser } from '@/services/auth';
import { listJobProfiles } from '@/services/job-profile';
import { useAuth } from '@/hooks/useAuth';
import {
  getResumeMailboxSettings,
  getResumeMailboxSetting,
  createResumeMailboxSetting,
  updateResumeMailboxSetting,
  deleteResumeMailboxSetting,
  syncResumeMailboxNow,
  getMailboxStatisticsSummary,
  type ResumeMailboxSetting,
  type CreateResumeMailboxSettingReq,
} from '@/services/resume-mailbox';

// 邮箱配置类型定义
interface EmailConfigFormData {
  task_name: string;
  email_address: string;
  auth_code: string;
  server_type: string;
  imap_server: string;
  imap_port: number;
  use_ssl: boolean;
  sync_frequency: string;
  status: 'enabled' | 'disabled';
  resume_uploader: string;
  uploader_id: string;
  job_profile_ids: string[]; // 改为数组支持多选
}

// 使用从服务导入的类型
type EmailConfig = ResumeMailboxSetting;

export function EmailConfigTab() {
  // 获取当前用户信息
  const { user } = useAuth();

  // 邮箱配置列表数据
  const [emailConfigs, setEmailConfigs] = useState<EmailConfig[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [totalCount, setTotalCount] = useState(0);
  const [pageSize] = useState(7);

  // 添加/编辑邮箱配置弹窗状态
  const [isAddModalOpen, setIsAddModalOpen] = useState(false);
  const [editingConfigId, setEditingConfigId] = useState<string | null>(null);

  // 表单数据
  const [formData, setFormData] = useState<EmailConfigFormData>({
    task_name: '',
    email_address: '',
    auth_code: '',
    server_type: 'imap',
    imap_server: '',
    imap_port: 993,
    use_ssl: true,
    sync_frequency: '15',
    status: 'enabled',
    resume_uploader: '',
    uploader_id: '',
    job_profile_ids: [], // 改为空数组
  });

  // 岗位画像列表
  const [loadingJobProfiles, setLoadingJobProfiles] = useState(false);
  const [jobOptions, setJobOptions] = useState<Option[]>([]);

  // 删除确认弹窗状态
  const [isDeleteConfirmOpen, setIsDeleteConfirmOpen] = useState(false);
  const [configToDelete, setConfigToDelete] = useState<{
    id: string;
    task_name: string;
  } | null>(null);

  // 操作状态
  const [submitting, setSubmitting] = useState(false);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [syncingId, setSyncingId] = useState<string | null>(null);

  // 获取邮箱配置列表
  const fetchEmailConfigs = useCallback(
    async (page?: number) => {
      try {
        setLoading(true);
        setError(null);
        const targetPage = page || currentPage;

        // 调用真实API
        const data = await getResumeMailboxSettings();

        // 为每个邮箱配置获取统计数据
        const dataWithStats = await Promise.all(
          data.map(async (config) => {
            try {
              const stats = await getMailboxStatisticsSummary(config.id);
              return {
                ...config,
                synced_count: stats.total_synced_emails, // 使用统计接口的已同步邮件数
              };
            } catch (error) {
              console.error(`获取邮箱 ${config.id} 统计数据失败:`, error);
              return config; // 如果获取统计失败，返回原始数据
            }
          })
        );

        const total = dataWithStats.length;
        const newTotalPages = Math.ceil(total / pageSize);
        const startIndex = (targetPage - 1) * pageSize;
        const endIndex = startIndex + pageSize;
        const paginatedData = dataWithStats.slice(startIndex, endIndex);

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
      const profiles = response.items || [];

      // 转换为 MultiSelect 所需的 Option 格式
      const options: Option[] = profiles.map((profile) => ({
        value: profile.id,
        label: profile.name,
      }));
      setJobOptions(options);
    } catch (err) {
      console.error('获取岗位画像列表失败:', err);
      setJobOptions([]);
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
    setEditingConfigId(null);
    // 重置表单
    setFormData({
      task_name: '',
      email_address: '',
      auth_code: '',
      server_type: 'imap',
      imap_server: '',
      imap_port: 993,
      use_ssl: true,
      sync_frequency: '15',
      status: 'enabled',
      resume_uploader: '',
      uploader_id: '',
      job_profile_ids: [], // 改为空数组
    });
  };

  // 编辑邮箱配置
  const handleEditConfig = async (configId: string) => {
    try {
      setError(null);
      // 调用API获取配置详情
      const config = await getResumeMailboxSetting(configId);

      // 填充表单数据（映射后端字段到表单字段）
      setFormData({
        task_name: config.name,
        email_address: config.email_address,
        auth_code: config.encrypted_credential?.password || '', // 从对象中提取password
        server_type: config.protocol,
        imap_server: config.host,
        imap_port: config.port,
        use_ssl: config.use_ssl,
        sync_frequency: String(config.sync_interval_minutes || 15),
        status: config.status,
        resume_uploader: config.uploader_name || '',
        uploader_id: config.uploader_id,
        job_profile_ids: config.job_profile_ids || [],
      });

      setEditingConfigId(configId);
      setIsAddModalOpen(true);
    } catch (err) {
      setError('获取配置详情失败，请重试');
      console.error('获取配置详情失败:', err);
    }
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

    if (!formData.job_profile_ids || formData.job_profile_ids.length === 0) {
      setError('请选择岗位画像');
      return;
    }

    // 验证同步频率是否为有效数字
    const syncFreq = parseInt(formData.sync_frequency);
    if (isNaN(syncFreq) || syncFreq < 5 || syncFreq > 1440) {
      setError('同步频率必须是5-1440之间的数字');
      return;
    }

    try {
      setSubmitting(true);
      setError(null);

      const authType = 'password';
      // 构建 encrypted_credential 为 JSON 字符串格式
      const encryptedCredential = JSON.stringify({
        [authType]: formData.auth_code,
      });

      console.log('表单数据:', formData);
      console.log('加密凭证:', encryptedCredential);

      if (editingConfigId) {
        // 编辑模式：调用更新API
        const updateData = {
          name: formData.task_name,
          email_address: formData.email_address,
          auth_type: authType,
          encrypted_credential: { [authType]: formData.auth_code }, // 发送对象而不是字符串
          protocol: formData.server_type,
          host: formData.imap_server,
          port: formData.imap_port,
          use_ssl: true,
          sync_interval_minutes: parseInt(formData.sync_frequency), // 修正字段名
          job_profile_ids: formData.job_profile_ids, // 发送数组
          status: formData.status, // 添加状态字段
        };

        console.log('更新的请求数据:', updateData);
        await updateResumeMailboxSetting(editingConfigId, updateData);
      } else {
        // 验证用户ID
        if (!user?.id) {
          setError('用户未登录，无法创建邮箱配置');
          return;
        }

        // 新增模式：调用创建API
        const requestData: CreateResumeMailboxSettingReq = {
          name: formData.task_name,
          email_address: formData.email_address,
          auth_type: authType,
          encrypted_credential: { [authType]: formData.auth_code }, // 发送对象而不是字符串
          protocol: formData.server_type,
          host: formData.imap_server,
          port: formData.imap_port,
          use_ssl: true,
          sync_interval_minutes: parseInt(formData.sync_frequency), // 修正字段名
          job_profile_ids: formData.job_profile_ids, // 发送数组
          uploader_id: user.id, // 添加必需的用户ID
          status: formData.status, // 添加状态字段
        };

        console.log('===== 调试信息 =====');
        console.log('发送的请求数据:', requestData);
        console.log(
          'encrypted_credential 类型:',
          typeof requestData.encrypted_credential
        );
        console.log(
          'encrypted_credential 值:',
          requestData.encrypted_credential
        );
        console.log('JSON.stringify 后:', JSON.stringify(requestData, null, 2));
        console.log('==================');

        await createResumeMailboxSetting(requestData);
      }

      handleCloseAddModal();
      setCurrentPage(1);
      await fetchEmailConfigs(1);
    } catch (err) {
      setError(
        editingConfigId
          ? '更新邮箱配置失败，请重试'
          : '保存邮箱配置失败，请重试'
      );
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

      // 调用真实API
      await deleteResumeMailboxSetting(configToDelete.id);

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

  // 手动触发同步
  const handleManualSync = async (configId: string) => {
    try {
      setSyncingId(configId);
      setError(null);

      // 调用真实API
      await syncResumeMailboxNow(configId);

      console.log(`手动触发同步邮箱简历: ${configId}`);

      // 刷新列表
      await fetchEmailConfigs(currentPage);
    } catch (err) {
      setError('同步失败，请重试');
      console.error('同步失败:', err);
    } finally {
      setSyncingId(null);
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
  // 格式化日期时间为年月日时分秒
  const formatDateTime = (dateString: string | null | undefined) => {
    if (!dateString) return 'Invalid Date';

    try {
      const date = new Date(dateString);
      if (isNaN(date.getTime())) return 'Invalid Date';

      const year = date.getFullYear();
      const month = String(date.getMonth() + 1).padStart(2, '0');
      const day = String(date.getDate()).padStart(2, '0');
      const hours = String(date.getHours()).padStart(2, '0');
      const minutes = String(date.getMinutes()).padStart(2, '0');
      const seconds = String(date.getSeconds()).padStart(2, '0');

      return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;
    } catch {
      return 'Invalid Date';
    }
  };

  // 截断邮箱地址,只显示前10位
  const truncateEmail = (email: string) => {
    if (email.length <= 10) {
      return email;
    }
    return email.substring(0, 10) + '...';
  };

  // 格式化岗位名称显示
  const formatJobProfileNames = (names: string[] | undefined) => {
    if (!names || names.length === 0) return '-';
    if (names.length === 1) return names[0];
    return `${names[0]}等${names.length}个岗位`;
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
                  同步频率
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
                        {config.name}
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
                        {formatDateTime(config.last_synced_at)}
                      </span>
                    </div>

                    {/* 同步频率 */}
                    <div className="flex items-center w-24 pr-4">
                      <span className="text-sm text-gray-500">
                        {config.sync_interval_minutes
                          ? `${config.sync_interval_minutes}分钟`
                          : '15分钟'}
                      </span>
                    </div>

                    {/* 已同步数 */}
                    <div className="flex items-center w-24 pr-4">
                      <span className="text-sm text-gray-900">
                        {config.synced_count || 0}
                      </span>
                    </div>

                    {/* 简历上传人 */}
                    <div className="flex items-center w-32 pr-4">
                      <span className="text-sm text-gray-500 truncate">
                        {config.uploader_name || '-'}
                      </span>
                    </div>

                    {/* 岗位名称 */}
                    <div className="flex items-center w-40 pr-4">
                      <span
                        className="text-sm text-gray-500 truncate cursor-default"
                        title={config.job_profile_names?.join('、') || '-'}
                      >
                        {formatJobProfileNames(config.job_profile_names)}
                      </span>
                    </div>

                    {/* 操作 */}
                    <div className="flex items-center justify-center gap-2 w-24">
                      <button
                        className="text-gray-500 hover:text-green-500 transition-colors disabled:opacity-50"
                        title="手动同步"
                        onClick={() => handleManualSync(config.id)}
                        disabled={syncingId === config.id || submitting}
                      >
                        {syncingId === config.id ? (
                          <Loader2 className="w-4 h-4 animate-spin" />
                        ) : (
                          <RefreshCw className="w-4 h-4" />
                        )}
                      </button>

                      <button
                        className="text-gray-500 hover:text-blue-500 transition-colors"
                        title="编辑"
                        onClick={() => handleEditConfig(config.id)}
                        disabled={submitting}
                      >
                        <Edit className="w-4 h-4" />
                      </button>

                      <button
                        className="text-gray-500 hover:text-red-500 transition-colors"
                        title="删除"
                        onClick={() =>
                          handleDeleteConfig(config.id, config.name)
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
                        'border-[#7bb8ff] bg-[#7bb8ff] text-white'
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
          loadingJobProfiles={loadingJobProfiles}
          jobOptions={jobOptions}
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
  loadingJobProfiles: boolean;
  jobOptions: Option[];
}

function AddEmailConfigModal({
  isOpen,
  onClose,
  formData,
  onFormChange,
  onSave,
  submitting,
  error: _error,
  loadingJobProfiles,
  jobOptions,
}: AddEmailConfigModalProps) {
  // 控制授权码显示/隐藏
  const [showAuthCode, setShowAuthCode] = useState(false);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 flex items-center justify-center bg-black bg-opacity-50 z-[1001]">
      <div className="bg-white rounded-lg shadow-xl w-[700px] max-h-[90vh] overflow-hidden">
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
            <div className="relative">
              <input
                type={showAuthCode ? 'text' : 'password'}
                value={formData.auth_code}
                onChange={(e) => onFormChange('auth_code', e.target.value)}
                placeholder="请输入邮箱授权码"
                className="w-full h-10 px-3 py-2 pr-10 border border-gray-200 rounded-md focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent"
              />
              <button
                type="button"
                onClick={() => setShowAuthCode(!showAuthCode)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 focus:outline-none"
              >
                {showAuthCode ? (
                  <Eye className="w-4 h-4" />
                ) : (
                  <EyeOff className="w-4 h-4" />
                )}
              </button>
            </div>
            <p className="mt-1 text-xs text-gray-500">
              请到邮箱设置中开启IMAP服务并获取授权码
            </p>
          </div>

          {/* 同步频率 */}
          <div className="mb-6">
            <label className="block text-sm font-medium text-gray-700 mb-2">
              同步频率 <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              value={formData.sync_frequency}
              onChange={(e) => onFormChange('sync_frequency', e.target.value)}
              placeholder="请输入同步频率，最小5分钟，最大1440分钟"
              className="w-full h-10 px-3 py-2 border border-gray-200 rounded-md focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent placeholder:text-gray-400"
            />
            <p className="mt-1 text-xs text-gray-500">
              设置邮箱自动同步简历的频率（5-1440分钟）
            </p>
          </div>

          {/* 服务器类型 */}
          <div className="mb-6">
            <label className="block text-sm font-medium text-gray-700 mb-2">
              服务器类型 <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              value="imap"
              disabled
              className="w-full h-10 px-3 py-2 border border-gray-200 rounded-md bg-gray-50 text-gray-500 cursor-not-allowed"
            />
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
              disabled
              placeholder="默认为当前登录人"
              className="w-full h-10 px-3 py-2 border border-gray-200 rounded-md bg-gray-50 text-gray-500 cursor-not-allowed"
            />
          </div>

          {/* 岗位画像 */}
          <div className="mb-6">
            <label className="block text-sm font-medium text-gray-700 mb-2">
              岗位画像 <span className="text-red-500">*</span>
            </label>
            <MultiSelect
              options={jobOptions}
              selected={formData.job_profile_ids}
              onChange={(selectedIds) => {
                onFormChange('job_profile_ids', selectedIds);
              }}
              placeholder={
                loadingJobProfiles ? '加载岗位中...' : '请选择岗位画像'
              }
              multiple={true}
              searchPlaceholder="搜索岗位名称..."
              disabled={loadingJobProfiles}
              selectCountLabel="岗位"
            />
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
