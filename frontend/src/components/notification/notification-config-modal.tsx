import { useState, useEffect, useCallback } from 'react';
import { X, Plus, Edit, Trash2, Loader2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { ConfirmDialog } from '@/components/ui/confirm-dialog';
import { AddNotificationModal } from './add-notification-modal';
import {
  getNotificationSettings,
  deleteNotificationSetting,
} from '@/services/notification';

export interface NotificationConfig {
  id: string;
  description: string; // 通知名称
  channel: 'dingtalk' | 'email'; // 通知方式
  dingtalk_config?: {
    webhook_url: string;
    secret: string;
    token?: string;
  };
  enabled: boolean; // 是否启用
  created_at: string;
  updated_at: string;
}

interface NotificationConfigModalProps {
  open: boolean;
  onClose: () => void;
}

export function NotificationConfigModal({
  open,
  onClose,
}: NotificationConfigModalProps) {
  const [notifications, setNotifications] = useState<NotificationConfig[]>([]);
  const [isAddModalOpen, setIsAddModalOpen] = useState(false);
  const [editingNotification, setEditingNotification] =
    useState<NotificationConfig | null>(null);
  const [loading, setLoading] = useState(false);
  const [deleting, setDeleting] = useState<string | null>(null);
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
  const [pendingDeleteId, setPendingDeleteId] = useState<string | null>(null);
  const [message, setMessage] = useState<{
    type: 'success' | 'error';
    text: string;
  } | null>(null);

  // 显示消息提示
  const showMessage = (type: 'success' | 'error', text: string) => {
    setMessage({ type, text });
    setTimeout(() => setMessage(null), 3000);
  };

  // 加载通知配置列表
  const loadNotifications = useCallback(async () => {
    try {
      setLoading(true);
      const data = await getNotificationSettings();
      console.log('📋 通知配置列表数据:', data);
      console.log('📋 数据数量:', data?.length);
      setNotifications(data);
    } catch (error) {
      console.error('❌ 加载通知配置失败:', error);
      showMessage('error', '加载通知配置失败，请稍后重试');
    } finally {
      setLoading(false);
    }
  }, []);

  // 当弹窗打开时加载数据
  useEffect(() => {
    if (open) {
      loadNotifications();
    }
  }, [open, loadNotifications]);

  // 保存后刷新列表
  const handleSaveSuccess = () => {
    setIsAddModalOpen(false);
    setEditingNotification(null);
    loadNotifications();
  };

  // 打开删除确认对话框
  const handleDeleteClick = (id: string) => {
    setPendingDeleteId(id);
    setDeleteConfirmOpen(true);
  };

  // 确认删除通知
  const handleConfirmDelete = async () => {
    if (!pendingDeleteId) return;

    try {
      setDeleting(pendingDeleteId);
      await deleteNotificationSetting(pendingDeleteId);
      showMessage('success', '删除成功');
      loadNotifications();
    } catch (error) {
      console.error('删除通知配置失败:', error);
      showMessage('error', '删除失败，请稍后重试');
    } finally {
      setDeleting(null);
      setPendingDeleteId(null);
    }
  };

  if (!open) return null;

  return (
    <>
      <div
        className="fixed inset-0 flex items-center justify-center bg-black bg-opacity-50 z-[1000]"
        onClick={onClose}
      >
        <div
          className="bg-white rounded-lg shadow-xl w-[1200px] max-h-[90vh] overflow-hidden"
          onClick={(e) => e.stopPropagation()}
        >
          {/* 弹窗头部 */}
          <div className="flex items-center justify-between px-6 py-5 border-b border-gray-100">
            <h2 className="text-lg font-medium text-gray-900">通知配置</h2>
            <button
              onClick={onClose}
              className="text-gray-400 hover:text-gray-600 transition-colors w-6 h-6"
            >
              <X className="w-6 h-6" />
            </button>
          </div>

          {/* 弹窗内容 */}
          <div className="p-6 overflow-y-auto max-h-[calc(90vh-140px)]">
            {/* 消息提示 */}
            {message && (
              <div
                className={`mb-4 p-3 rounded-lg ${
                  message.type === 'success'
                    ? 'bg-green-50 text-green-800 border border-green-200'
                    : 'bg-red-50 text-red-800 border border-red-200'
                }`}
              >
                {message.text}
              </div>
            )}

            {/* 添加通知按钮 */}
            <div className="mb-6">
              <Button className="gap-2" onClick={() => setIsAddModalOpen(true)}>
                <Plus className="w-4 h-4" />
                添加通知
              </Button>
            </div>

            {/* 通知列表表格 */}
            <div className="border border-gray-200 rounded-lg overflow-hidden">
              {/* 表格头部 */}
              <div className="bg-gray-50 border-b border-gray-200">
                <div className="flex items-center px-6 py-3 w-full">
                  <div className="flex items-center w-48 pr-4">
                    <span className="text-sm font-medium text-gray-700">
                      通知名称
                    </span>
                  </div>
                  <div className="flex items-center w-32 pr-4">
                    <span className="text-sm font-medium text-gray-700">
                      通知方式
                    </span>
                  </div>
                  <div className="flex items-center w-28 pr-4">
                    <span className="text-sm font-medium text-gray-700">
                      通知状态
                    </span>
                  </div>
                  <div className="flex items-center flex-1 min-w-0 pr-4">
                    <span className="text-sm font-medium text-gray-700">
                      Webhook地址
                    </span>
                  </div>
                  <div className="flex items-center justify-center w-32">
                    <span className="text-sm font-medium text-gray-700">
                      操作
                    </span>
                  </div>
                </div>
              </div>

              {/* 表格内容 */}
              <div className="bg-white">
                {loading ? (
                  <div className="px-6 py-12 flex items-center justify-center text-gray-500">
                    <Loader2 className="w-5 h-5 animate-spin mr-2" />
                    加载中...
                  </div>
                ) : notifications.length === 0 ? (
                  <div className="px-6 py-12 text-center text-gray-500">
                    暂无通知配置，请点击"添加通知"按钮添加
                  </div>
                ) : (
                  notifications.map((notification) => (
                    <div
                      key={notification.id}
                      className="flex items-center px-6 py-4 border-b border-gray-100 last:border-b-0 hover:bg-gray-50 transition-colors"
                    >
                      <div className="flex items-center w-48 pr-4">
                        <span className="text-sm text-gray-900 truncate">
                          {notification.description}
                        </span>
                      </div>
                      <div className="flex items-center w-32 pr-4">
                        <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                          {notification.channel === 'dingtalk'
                            ? '钉钉通知'
                            : '邮件通知'}
                        </span>
                      </div>
                      <div className="flex items-center w-28 pr-4">
                        <span
                          className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${
                            notification.enabled
                              ? 'bg-green-100 text-green-800'
                              : 'bg-gray-100 text-gray-800'
                          }`}
                        >
                          {notification.enabled ? '已启用' : '已禁用'}
                        </span>
                      </div>
                      <div className="flex items-center flex-1 min-w-0 pr-4">
                        <span
                          className="text-sm text-gray-700 truncate"
                          title={
                            notification.dingtalk_config?.webhook_url || '-'
                          }
                        >
                          {notification.dingtalk_config?.webhook_url || '-'}
                        </span>
                      </div>
                      <div className="flex items-center justify-center gap-2 w-32">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => {
                            setEditingNotification(notification);
                            setIsAddModalOpen(true);
                          }}
                          className="h-8 w-8 p-0"
                          title="编辑"
                          disabled={deleting === notification.id}
                        >
                          <Edit className="h-4 w-4 text-gray-600" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleDeleteClick(notification.id)}
                          className="h-8 w-8 p-0"
                          title="删除"
                          disabled={deleting === notification.id}
                        >
                          {deleting === notification.id ? (
                            <Loader2 className="h-4 w-4 text-gray-600 animate-spin" />
                          ) : (
                            <Trash2 className="h-4 w-4 text-gray-600" />
                          )}
                        </Button>
                      </div>
                    </div>
                  ))
                )}
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* 添加/编辑通知弹窗 */}
      {isAddModalOpen && (
        <AddNotificationModal
          open={isAddModalOpen}
          onClose={() => {
            setIsAddModalOpen(false);
            setEditingNotification(null);
          }}
          onSaveSuccess={handleSaveSuccess}
          onShowMessage={showMessage}
          editingNotification={editingNotification}
        />
      )}

      {/* 删除确认对话框 */}
      <ConfirmDialog
        open={deleteConfirmOpen}
        onOpenChange={setDeleteConfirmOpen}
        title="删除通知配置"
        description="确定要删除此通知配置吗？删除后将无法恢复。"
        confirmText="删除"
        cancelText="取消"
        variant="destructive"
        onConfirm={handleConfirmDelete}
        loading={deleting !== null}
        zIndex="z-[1002]"
      />
    </>
  );
}
