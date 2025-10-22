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
  Play,
  Pause,
  TestTube,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { ConfirmDialog } from '@/components/ui/confirm-dialog';
import { cn } from '@/lib/utils';

// 邮箱配置类型定义
interface EmailConfigFormData {
  display_name: string;
  email_address: string;
  email_type: string;
  imap_server: string;
  imap_port: number;
  auth_code: string;
  use_ssl: boolean;
  monitor_folder: string;
  sync_frequency: string;
  keyword_filter: string;
  attachment_types: string[];
}

interface EmailConfig {
  id: string;
  display_name: string;
  email_address: string;
  email_type: string;
  status: 'running' | 'stopped' | 'error';
  sync_frequency: string;
  last_sync: number;
  synced_count: number;
  imap_server: string;
  imap_port: number;
  use_ssl: boolean;
  auth_code: string;
  monitor_folder: string;
  keyword_filter?: string;
  attachment_types: string[];
  created_at: number;
}

interface EmailConfigModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export function EmailConfigModal({ isOpen, onClose }: EmailConfigModalProps) {
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
    display_name: '',
    email_address: '',
    email_type: '',
    imap_server: '',
    imap_port: 993,
    auth_code: '',
    use_ssl: true,
    monitor_folder: 'INBOX',
    sync_frequency: '实时同步',
    keyword_filter: '',
    attachment_types: ['PDF', 'Word'],
  });

  // 删除确认弹窗状态
  const [isDeleteConfirmOpen, setIsDeleteConfirmOpen] = useState(false);
  const [configToDelete, setConfigToDelete] = useState<{
    id: string;
    display_name: string;
  } | null>(null);

  // 操作状态
  const [submitting, setSubmitting] = useState(false);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [statusChangingId, setStatusChangingId] = useState<string | null>(null);
  const [testingId, setTestingId] = useState<string | null>(null);

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
            display_name: '公司招聘邮箱',
            email_address: 'hr@company.com',
            email_type: '企业邮箱',
            status: 'running',
            sync_frequency: '每15分钟',
            last_sync: Date.now() / 1000 - 3600, // 1小时前
            synced_count: 156,
            imap_server: 'imap.company.com',
            imap_port: 993,
            use_ssl: true,
            auth_code: '****',
            monitor_folder: 'INBOX',
            keyword_filter: '简历,应聘,求职',
            attachment_types: ['PDF', 'Word'],
            created_at: Date.now() / 1000 - 86400 * 30, // 30天前
          },
          {
            id: '2',
            display_name: '备用招聘邮箱',
            email_address: 'recruitment@company.com',
            email_type: '企业邮箱',
            status: 'stopped',
            sync_frequency: '实时同步',
            last_sync: Date.now() / 1000 - 86400, // 1天前
            synced_count: 89,
            imap_server: 'imap.company.com',
            imap_port: 993,
            use_ssl: true,
            auth_code: '****',
            monitor_folder: 'INBOX',
            keyword_filter: '简历',
            attachment_types: ['PDF', 'Word', '图片'],
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

  // 打开添加弹窗
  const handleOpenAddModal = () => {
    setIsAddModalOpen(true);
  };

  // 关闭添加弹窗
  const handleCloseAddModal = () => {
    setIsAddModalOpen(false);
    // 重置表单
    setFormData({
      display_name: '',
      email_address: '',
      email_type: '',
      imap_server: '',
      imap_port: 993,
      auth_code: '',
      use_ssl: true,
      monitor_folder: 'INBOX',
      sync_frequency: '实时同步',
      keyword_filter: '',
      attachment_types: ['PDF', 'Word'],
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
    if (!formData.email_address.trim()) {
      setError('邮箱地址不能为空');
      return;
    }

    if (!formData.email_type) {
      setError('请选择邮箱类型');
      return;
    }

    if (!formData.imap_server.trim()) {
      setError('IMAP服务器不能为空');
      return;
    }

    if (!formData.auth_code.trim()) {
      setError('授权码不能为空');
      return;
    }

    if (!formData.sync_frequency) {
      setError('请选择同步频率');
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

  // 测试连接
  const handleTestConnection = async () => {
    if (
      !formData.email_address.trim() ||
      !formData.imap_server.trim() ||
      !formData.auth_code.trim()
    ) {
      setError('请填写完整的连接信息');
      return;
    }

    try {
      setTestingId('test');
      setError(null);

      // 模拟API调用
      await new Promise((resolve) => setTimeout(resolve, 1000));

      alert('连接测试成功！');
    } catch (err) {
      setError('连接测试失败，请检查配置');
      console.error('连接测试失败:', err);
    } finally {
      setTestingId(null);
    }
  };

  // 切换邮箱状态
  const handleToggleStatus = async (
    configId: string,
    currentStatus: string
  ) => {
    try {
      setStatusChangingId(configId);
      setError(null);

      // 模拟API调用
      await new Promise((resolve) => setTimeout(resolve, 500));

      console.log(`切换邮箱状态: ${configId}, 当前状态: ${currentStatus}`);

      await fetchEmailConfigs(currentPage);
    } catch (err) {
      setError('切换状态失败，请重试');
      console.error('切换状态失败:', err);
    } finally {
      setStatusChangingId(null);
    }
  };

  // 删除邮箱配置
  const handleDeleteConfig = (configId: string, displayName: string) => {
    setConfigToDelete({ id: configId, display_name: displayName });
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

  // 获取状态显示
  const getStatusDisplay = (status: string) => {
    switch (status) {
      case 'running':
        return {
          icon: '✓',
          text: '运行中',
          bgColor: 'bg-green-50',
          textColor: 'text-green-600',
        };
      case 'stopped':
        return {
          icon: '⏸',
          text: '已停用',
          bgColor: 'bg-gray-50',
          textColor: 'text-gray-600',
        };
      case 'error':
        return {
          icon: '✗',
          text: '错误',
          bgColor: 'bg-red-50',
          textColor: 'text-red-600',
        };
      default:
        return {
          icon: '?',
          text: '未知',
          bgColor: 'bg-gray-50',
          textColor: 'text-gray-600',
        };
    }
  };

  // 页面切换时获取数据
  useEffect(() => {
    if (isOpen) {
      fetchEmailConfigs();
    }
  }, [currentPage, fetchEmailConfigs, isOpen]);

  if (!isOpen) return null;

  return (
    <>
      {/* 主弹窗 */}
      <div
        className="fixed inset-0 flex items-center justify-center bg-black bg-opacity-50 z-[1000]"
        onClick={onClose}
      >
        <div
          className="bg-white rounded-lg shadow-xl w-[1000px] max-h-[90vh] overflow-hidden"
          onClick={(e) => e.stopPropagation()}
        >
          {/* 弹窗头部 */}
          <div className="flex items-center justify-between px-6 py-5 border-b border-gray-100">
            <h2 className="text-lg font-medium text-gray-900">邮箱配置</h2>
            <button
              onClick={onClose}
              className="text-gray-400 hover:text-gray-600 transition-colors w-6 h-6"
            >
              <X className="w-6 h-6" />
            </button>
          </div>

          {/* 弹窗内容 */}
          <div className="p-6 overflow-y-auto max-h-[calc(90vh-140px)]">
            {/* 错误提示 */}
            {error && (
              <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-md">
                <div className="flex items-center">
                  <AlertCircle className="w-4 h-4 text-red-500 mr-2" />
                  <span className="text-sm text-red-700">{error}</span>
                </div>
              </div>
            )}

            {/* 添加配置按钮 */}
            <div className="mb-6 flex justify-end">
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
                  <div className="flex items-center w-48 pr-4">
                    <span className="text-sm font-medium text-gray-700">
                      显示名称
                    </span>
                  </div>
                  <div className="flex items-center w-52 pr-4">
                    <span className="text-sm font-medium text-gray-700">
                      邮箱地址
                    </span>
                  </div>
                  <div className="flex items-center w-28 pr-4">
                    <span className="text-sm font-medium text-gray-700">
                      邮箱类型
                    </span>
                  </div>
                  <div className="flex items-center w-28 pr-4">
                    <span className="text-sm font-medium text-gray-700">
                      状态
                    </span>
                  </div>
                  <div className="flex items-center w-28 pr-4">
                    <span className="text-sm font-medium text-gray-700">
                      同步频率
                    </span>
                  </div>
                  <div className="flex items-center w-40 pr-4">
                    <span className="text-sm font-medium text-gray-700">
                      最后同步
                    </span>
                  </div>
                  <div className="flex items-center w-24 pr-4">
                    <span className="text-sm font-medium text-gray-700">
                      已同步数
                    </span>
                  </div>
                  <div className="flex items-center justify-center w-40">
                    <span className="text-sm font-medium text-gray-700">
                      操作
                    </span>
                  </div>
                </div>
              </div>

              {/* 表格内容 */}
              <div className="bg-white">
                {loading ? (
                  <div className="flex items-center justify-center py-8">
                    <Loader2 className="w-6 h-6 animate-spin text-gray-400" />
                    <span className="ml-2 text-sm text-gray-500">
                      加载中...
                    </span>
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
                        {/* 显示名称 */}
                        <div className="flex items-center w-48 pr-4">
                          <span className="text-sm text-gray-900 truncate">
                            {config.display_name}
                          </span>
                        </div>

                        {/* 邮箱地址 */}
                        <div className="flex items-center w-52 pr-4">
                          <span className="text-sm text-gray-900 truncate">
                            {config.email_address}
                          </span>
                        </div>

                        {/* 邮箱类型 */}
                        <div className="flex items-center w-28 pr-4">
                          <span className="text-sm text-gray-500 truncate">
                            {config.email_type}
                          </span>
                        </div>

                        {/* 状态 */}
                        <div className="flex items-center w-28 pr-4">
                          <div
                            className={`px-3 py-1 rounded ${statusDisplay.bgColor}`}
                          >
                            <span
                              className={`text-xs font-medium ${statusDisplay.textColor}`}
                            >
                              {statusDisplay.icon} {statusDisplay.text}
                            </span>
                          </div>
                        </div>

                        {/* 同步频率 */}
                        <div className="flex items-center w-28 pr-4">
                          <span className="text-sm text-gray-500 truncate">
                            {config.sync_frequency}
                          </span>
                        </div>

                        {/* 最后同步 */}
                        <div className="flex items-center w-40 pr-4">
                          <span className="text-xs text-gray-500">
                            {formatDateTime(config.last_sync)}
                          </span>
                        </div>

                        {/* 已同步数 */}
                        <div className="flex items-center w-24 pr-4">
                          <span className="text-sm text-gray-900">
                            {config.synced_count} 份
                          </span>
                        </div>

                        {/* 操作 */}
                        <div className="flex items-center justify-center gap-2 w-40">
                          <button
                            className="text-gray-500 hover:text-blue-500 transition-colors"
                            title="编辑"
                            disabled={submitting}
                          >
                            <Edit className="w-4 h-4" />
                          </button>

                          <button
                            className={cn(
                              'transition-colors',
                              config.status === 'running'
                                ? 'text-gray-500 hover:text-orange-500'
                                : 'text-gray-500 hover:text-green-500'
                            )}
                            title={
                              config.status === 'running' ? '停用' : '启用'
                            }
                            onClick={() =>
                              handleToggleStatus(config.id, config.status)
                            }
                            disabled={
                              submitting || statusChangingId === config.id
                            }
                          >
                            {statusChangingId === config.id ? (
                              <Loader2 className="w-4 h-4 animate-spin" />
                            ) : config.status === 'running' ? (
                              <Pause className="w-4 h-4" />
                            ) : (
                              <Play className="w-4 h-4" />
                            )}
                          </button>

                          <button
                            className="text-gray-500 hover:text-purple-500 transition-colors"
                            title="测试连接"
                            disabled={submitting}
                          >
                            <TestTube className="w-4 h-4" />
                          </button>

                          <button
                            className="text-gray-500 hover:text-red-500 transition-colors"
                            title="删除"
                            onClick={() =>
                              handleDeleteConfig(config.id, config.display_name)
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
        </div>
      </div>

      {/* 添加邮箱配置弹窗 - 将在下一个文件继续 */}
      {isAddModalOpen && (
        <AddEmailConfigModal
          isOpen={isAddModalOpen}
          onClose={handleCloseAddModal}
          formData={formData}
          onFormChange={handleFormChange}
          onSave={handleSaveEmailConfig}
          onTest={handleTestConnection}
          submitting={submitting}
          testingId={testingId}
          error={error}
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
              确定要删除 <strong>{configToDelete.display_name}</strong>{' '}
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
  onTest: () => void;
  submitting: boolean;
  testingId: string | null;
  error: string | null;
}

function AddEmailConfigModal({
  isOpen,
  onClose,
  formData,
  onFormChange,
  onSave,
  onTest,
  submitting,
  testingId,
  error: _error,
}: AddEmailConfigModalProps) {
  if (!isOpen) return null;

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-2xl">
        <DialogHeader className="flex flex-row items-center justify-between space-y-0 pb-6 border-b">
          <DialogTitle className="text-lg font-medium text-gray-900">
            添加邮箱配置
          </DialogTitle>
          <Button
            variant="ghost"
            size="sm"
            onClick={onClose}
            className="h-6 w-6 p-0"
          >
            <X className="h-4 w-4" />
          </Button>
        </DialogHeader>

        <div className="py-6">
          <form className="space-y-6">
            {/* 基本信息部分 */}
            <div className="space-y-4">
              <h3 className="text-base font-semibold text-gray-900 pb-3 border-b border-gray-200">
                基本信息
              </h3>

              {/* 邮箱地址 */}
              <div className="space-y-2">
                <Label
                  htmlFor="email_address"
                  className="text-sm font-medium text-gray-700"
                >
                  邮箱地址 <span className="text-red-500">*</span>
                </Label>
                <Input
                  id="email_address"
                  type="email"
                  value={formData.email_address}
                  onChange={(e) =>
                    onFormChange('email_address', e.target.value)
                  }
                  placeholder="example@company.com"
                  className="h-10"
                />
              </div>

              {/* 邮箱类型 */}
              <div className="space-y-2">
                <Label
                  htmlFor="email_type"
                  className="text-sm font-medium text-gray-700"
                >
                  邮箱类型 <span className="text-red-500">*</span>
                </Label>
                <select
                  id="email_type"
                  value={formData.email_type}
                  onChange={(e) => onFormChange('email_type', e.target.value)}
                  className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                >
                  <option value="">请选择</option>
                  <option value="企业邮箱">企业邮箱</option>
                  <option value="个人邮箱">个人邮箱</option>
                </select>
              </div>

              {/* 显示名称 */}
              <div className="space-y-2">
                <Label
                  htmlFor="display_name"
                  className="text-sm font-medium text-gray-700"
                >
                  显示名称
                </Label>
                <Input
                  id="display_name"
                  type="text"
                  value={formData.display_name}
                  onChange={(e) => onFormChange('display_name', e.target.value)}
                  placeholder="用于区分多个邮箱"
                  className="h-10"
                />
              </div>
            </div>

            {/* 授权信息部分 */}
            <div className="space-y-4">
              <h3 className="text-base font-semibold text-gray-900 pb-3 border-b border-gray-200">
                授权信息
              </h3>

              {/* IMAP服务器 */}
              <div className="space-y-2">
                <Label
                  htmlFor="imap_server"
                  className="text-sm font-medium text-gray-700"
                >
                  IMAP服务器 <span className="text-red-500">*</span>
                </Label>
                <Input
                  id="imap_server"
                  type="text"
                  value={formData.imap_server}
                  onChange={(e) => onFormChange('imap_server', e.target.value)}
                  placeholder="imap.example.com"
                  className="h-10"
                />
              </div>

              {/* IMAP端口 */}
              <div className="space-y-2">
                <Label
                  htmlFor="imap_port"
                  className="text-sm font-medium text-gray-700"
                >
                  IMAP端口 <span className="text-red-500">*</span>
                </Label>
                <Input
                  id="imap_port"
                  type="number"
                  value={formData.imap_port}
                  onChange={(e) =>
                    onFormChange('imap_port', parseInt(e.target.value) || 993)
                  }
                  placeholder="993"
                  className="h-10"
                />
              </div>

              {/* 授权码/密码 */}
              <div className="space-y-2">
                <Label
                  htmlFor="auth_code"
                  className="text-sm font-medium text-gray-700"
                >
                  授权码/密码 <span className="text-red-500">*</span>
                </Label>
                <Input
                  id="auth_code"
                  type="password"
                  value={formData.auth_code}
                  onChange={(e) => onFormChange('auth_code', e.target.value)}
                  placeholder="请输入邮箱授权码"
                  className="h-10"
                />
                <p className="text-xs text-gray-500">
                  请到邮箱设置中开启IMAP服务并获取授权码
                </p>
              </div>

              {/* 使用SSL/TLS加密 */}
              <div className="flex items-center">
                <input
                  type="checkbox"
                  id="use_ssl"
                  checked={formData.use_ssl}
                  onChange={(e) => onFormChange('use_ssl', e.target.checked)}
                  className="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-2 focus:ring-primary"
                />
                <label
                  htmlFor="use_ssl"
                  className="ml-2 text-sm font-medium text-gray-900"
                >
                  使用SSL/TLS加密
                </label>
              </div>
            </div>

            {/* 同步规则部分 */}
            <div className="space-y-4">
              <h3 className="text-base font-semibold text-gray-900 pb-3 border-b border-gray-200">
                同步规则
              </h3>

              {/* 监控文件夹 */}
              <div className="space-y-2">
                <Label
                  htmlFor="monitor_folder"
                  className="text-sm font-medium text-gray-700"
                >
                  监控文件夹
                </Label>
                <Input
                  id="monitor_folder"
                  type="text"
                  value={formData.monitor_folder}
                  onChange={(e) =>
                    onFormChange('monitor_folder', e.target.value)
                  }
                  placeholder="INBOX"
                  className="h-10"
                />
              </div>

              {/* 同步频率 */}
              <div className="space-y-2">
                <Label
                  htmlFor="sync_frequency"
                  className="text-sm font-medium text-gray-700"
                >
                  同步频率 <span className="text-red-500">*</span>
                </Label>
                <select
                  id="sync_frequency"
                  value={formData.sync_frequency}
                  onChange={(e) =>
                    onFormChange('sync_frequency', e.target.value)
                  }
                  className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                >
                  <option value="实时同步">实时同步</option>
                  <option value="每5分钟">每5分钟</option>
                  <option value="每15分钟">每15分钟</option>
                  <option value="每30分钟">每30分钟</option>
                  <option value="每小时">每小时</option>
                </select>
              </div>

              {/* 关键词过滤 */}
              <div className="space-y-2">
                <Label
                  htmlFor="keyword_filter"
                  className="text-sm font-medium text-gray-700"
                >
                  关键词过滤
                </Label>
                <Input
                  id="keyword_filter"
                  type="text"
                  value={formData.keyword_filter}
                  onChange={(e) =>
                    onFormChange('keyword_filter', e.target.value)
                  }
                  placeholder="简历,应聘,求职（用逗号分隔）"
                  className="h-10"
                />
                <p className="text-xs text-gray-500">
                  邮件主题或正文包含这些关键词时才会同步
                </p>
              </div>

              {/* 附件类型 */}
              <div className="space-y-2">
                <Label className="text-sm font-medium text-gray-700">
                  附件类型
                </Label>
                <div className="flex items-center gap-5">
                  <label className="flex items-center">
                    <input
                      type="checkbox"
                      checked={formData.attachment_types.includes('PDF')}
                      onChange={(e) => {
                        const types = formData.attachment_types;
                        if (e.target.checked) {
                          onFormChange('attachment_types', [...types, 'PDF']);
                        } else {
                          onFormChange(
                            'attachment_types',
                            types.filter((t: string) => t !== 'PDF')
                          );
                        }
                      }}
                      className="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-2 focus:ring-primary"
                    />
                    <span className="ml-2 text-sm font-medium text-gray-900">
                      PDF
                    </span>
                  </label>

                  <label className="flex items-center">
                    <input
                      type="checkbox"
                      checked={formData.attachment_types.includes('Word')}
                      onChange={(e) => {
                        const types = formData.attachment_types;
                        if (e.target.checked) {
                          onFormChange('attachment_types', [...types, 'Word']);
                        } else {
                          onFormChange(
                            'attachment_types',
                            types.filter((t: string) => t !== 'Word')
                          );
                        }
                      }}
                      className="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-2 focus:ring-primary"
                    />
                    <span className="ml-2 text-sm font-medium text-gray-900">
                      Word
                    </span>
                  </label>

                  <label className="flex items-center">
                    <input
                      type="checkbox"
                      checked={formData.attachment_types.includes('图片')}
                      onChange={(e) => {
                        const types = formData.attachment_types;
                        if (e.target.checked) {
                          onFormChange('attachment_types', [...types, '图片']);
                        } else {
                          onFormChange(
                            'attachment_types',
                            types.filter((t: string) => t !== '图片')
                          );
                        }
                      }}
                      className="w-4 h-4 text-gray-900 border-gray-300 rounded focus:ring-2 focus:ring-primary"
                    />
                    <span className="ml-2 text-sm font-medium text-gray-900">
                      图片
                    </span>
                  </label>
                </div>
              </div>
            </div>

            {/* 按钮组 */}
            <div className="flex justify-end space-x-3 pt-4">
              <Button
                type="button"
                variant="outline"
                onClick={onTest}
                disabled={submitting || testingId !== null}
                className="px-4 py-2"
              >
                {testingId ? (
                  <>
                    <Loader2 className="w-4 h-4 animate-spin mr-2" />
                    测试中...
                  </>
                ) : (
                  '测试连接'
                )}
              </Button>
              <Button
                type="button"
                variant="outline"
                onClick={onClose}
                disabled={submitting}
                className="px-4 py-2"
              >
                取消
              </Button>
              <Button
                type="button"
                className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white"
                onClick={onSave}
                disabled={submitting || testingId !== null}
              >
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
          </form>
        </div>
      </DialogContent>
    </Dialog>
  );
}
