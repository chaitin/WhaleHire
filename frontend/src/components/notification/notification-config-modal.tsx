import { useState } from 'react';
import { X, Plus, Edit, Trash2, Send } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { AddNotificationModal } from './add-notification-modal';

export interface NotificationConfig {
  id: string;
  name: string;
  type: 'webhook' | 'email';
  webhook_url?: string;
  email_recipients?: string[];
  email_subject?: string;
  created_at: number;
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
  const [testingId, setTestingId] = useState<string | null>(null);

  // 添加通知
  const handleAddNotification = (notification: NotificationConfig) => {
    setNotifications([...notifications, notification]);
    setIsAddModalOpen(false);
  };

  // 编辑通知
  const handleEditNotification = (notification: NotificationConfig) => {
    setNotifications(
      notifications.map((n) => (n.id === notification.id ? notification : n))
    );
    setEditingNotification(null);
  };

  // 删除通知
  const handleDeleteNotification = (id: string) => {
    if (confirm('确定要删除此通知配置吗？')) {
      setNotifications(notifications.filter((n) => n.id !== id));
    }
  };

  // 测试通知
  const handleTestNotification = async (id: string) => {
    setTestingId(id);
    try {
      // TODO: 调用测试接口
      await new Promise((resolve) => setTimeout(resolve, 1000));
      alert('测试通知已发送');
    } catch {
      alert('测试失败，请检查配置');
    } finally {
      setTestingId(null);
    }
  };

  // 格式化时间
  const formatDate = (timestamp: number) => {
    return new Date(timestamp).toLocaleString('zh-CN');
  };

  if (!open) return null;

  return (
    <>
      <div
        className="fixed inset-0 flex items-center justify-center bg-black bg-opacity-50 z-[1000]"
        onClick={onClose}
      >
        <div
          className="bg-white rounded-lg shadow-xl w-[800px] max-h-[90vh] overflow-hidden"
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
                  <div className="flex items-center flex-1 min-w-0 pr-4">
                    <span className="text-sm font-medium text-gray-700">
                      通知名称
                    </span>
                  </div>
                  <div className="flex items-center w-32 pr-4">
                    <span className="text-sm font-medium text-gray-700">
                      通知方式
                    </span>
                  </div>
                  <div className="flex items-center w-40 pr-4">
                    <span className="text-sm font-medium text-gray-700">
                      创建时间
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
                {notifications.length === 0 ? (
                  <div className="px-6 py-12 text-center text-gray-500">
                    暂无通知配置，请点击"添加通知"按钮添加
                  </div>
                ) : (
                  notifications.map((notification) => (
                    <div
                      key={notification.id}
                      className="flex items-center px-6 py-4 border-b border-gray-100 last:border-b-0 hover:bg-gray-50 transition-colors"
                    >
                      <div className="flex items-center flex-1 min-w-0 pr-4">
                        <span className="text-sm text-gray-900 truncate">
                          {notification.name}
                        </span>
                      </div>
                      <div className="flex items-center w-32 pr-4">
                        <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                          {notification.type === 'webhook'
                            ? 'Webhook'
                            : '邮件通知'}
                        </span>
                      </div>
                      <div className="flex items-center w-40 pr-4">
                        <span className="text-sm text-gray-500">
                          {formatDate(notification.created_at)}
                        </span>
                      </div>
                      <div className="flex items-center justify-center gap-2 w-32">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() =>
                            handleTestNotification(notification.id)
                          }
                          disabled={testingId === notification.id}
                          className="h-8 w-8 p-0"
                          title="测试通知"
                        >
                          <Send className="h-4 w-4 text-blue-600" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => {
                            setEditingNotification(notification);
                            setIsAddModalOpen(true);
                          }}
                          className="h-8 w-8 p-0"
                          title="编辑"
                        >
                          <Edit className="h-4 w-4 text-gray-600" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() =>
                            handleDeleteNotification(notification.id)
                          }
                          className="h-8 w-8 p-0"
                          title="删除"
                        >
                          <Trash2 className="h-4 w-4 text-red-600" />
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
          onSave={
            editingNotification ? handleEditNotification : handleAddNotification
          }
          editingNotification={editingNotification}
        />
      )}
    </>
  );
}
