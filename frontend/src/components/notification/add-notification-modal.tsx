import { useState, useEffect } from 'react';
import { X, Send } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import type { NotificationConfig } from './notification-config-modal';

interface AddNotificationModalProps {
  open: boolean;
  onClose: () => void;
  onSave: (notification: NotificationConfig) => void;
  editingNotification: NotificationConfig | null;
}

export function AddNotificationModal({
  open,
  onClose,
  onSave,
  editingNotification,
}: AddNotificationModalProps) {
  const [name, setName] = useState('');
  const [type, setType] = useState<'webhook' | 'email'>('webhook');
  const [webhookUrl, setWebhookUrl] = useState('');
  const [emailRecipients, setEmailRecipients] = useState('');
  const [emailSubject, setEmailSubject] = useState('');
  const [testing, setTesting] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});

  // 编辑模式：填充表单
  useEffect(() => {
    if (editingNotification) {
      setName(editingNotification.name);
      setType(editingNotification.type);
      setWebhookUrl(editingNotification.webhook_url || '');
      setEmailRecipients(
        editingNotification.email_recipients?.join(', ') || ''
      );
      setEmailSubject(editingNotification.email_subject || '');
    } else {
      resetForm();
    }
  }, [editingNotification]);

  // 重置表单
  const resetForm = () => {
    setName('');
    setType('webhook');
    setWebhookUrl('');
    setEmailRecipients('');
    setEmailSubject('');
    setErrors({});
  };

  // 验证表单
  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!name.trim()) {
      newErrors.name = '请输入通知名称';
    }

    if (type === 'webhook') {
      if (!webhookUrl.trim()) {
        newErrors.webhookUrl = '请输入Webhook URL';
      } else if (
        !webhookUrl.startsWith('http://') &&
        !webhookUrl.startsWith('https://')
      ) {
        newErrors.webhookUrl = 'URL必须以http://或https://开头';
      }
    } else if (type === 'email') {
      if (!emailRecipients.trim()) {
        newErrors.emailRecipients = '请输入收件人邮箱';
      } else {
        const emails = emailRecipients.split(',').map((e) => e.trim());
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        if (!emails.every((e) => emailRegex.test(e))) {
          newErrors.emailRecipients =
            '请输入有效的邮箱地址，多个邮箱用逗号分隔';
        }
      }
      if (!emailSubject.trim()) {
        newErrors.emailSubject = '请输入邮件主题';
      }
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  // 测试通知
  const handleTest = async () => {
    if (!validateForm()) return;

    setTesting(true);
    try {
      // TODO: 调用测试接口
      await new Promise((resolve) => setTimeout(resolve, 1000));
      alert('测试通知已发送');
    } catch {
      alert('测试失败，请检查配置');
    } finally {
      setTesting(false);
    }
  };

  // 保存通知配置
  const handleSave = () => {
    if (!validateForm()) return;

    const notification: NotificationConfig = {
      id: editingNotification?.id || `notif_${Date.now()}`,
      name: name.trim(),
      type,
      created_at: editingNotification?.created_at || Date.now(),
    };

    if (type === 'webhook') {
      notification.webhook_url = webhookUrl.trim();
    } else {
      notification.email_recipients = emailRecipients
        .split(',')
        .map((e) => e.trim())
        .filter(Boolean);
      notification.email_subject = emailSubject.trim();
    }

    onSave(notification);
    resetForm();
  };

  if (!open) return null;

  return (
    <div
      className="fixed inset-0 flex items-center justify-center bg-black bg-opacity-50 z-[1001]"
      onClick={onClose}
    >
      <div
        className="bg-white rounded-lg shadow-xl w-[600px] max-h-[90vh] overflow-hidden"
        onClick={(e) => e.stopPropagation()}
      >
        {/* 弹窗头部 */}
        <div className="flex items-center justify-between px-6 py-5 border-b border-gray-100">
          <h2 className="text-lg font-medium text-gray-900">
            {editingNotification ? '编辑通知配置' : '添加通知配置'}
          </h2>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600 transition-colors w-6 h-6"
          >
            <X className="w-6 h-6" />
          </button>
        </div>

        {/* 弹窗内容 */}
        <div className="p-6 overflow-y-auto max-h-[calc(90vh-200px)]">
          <div className="space-y-4">
            {/* 通知名称 */}
            <div>
              <Label className="text-sm font-medium text-gray-700 mb-2">
                通知名称 <span className="text-red-500">*</span>
              </Label>
              <Input
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="请输入通知名称"
                className={errors.name ? 'border-red-500' : ''}
              />
              {errors.name && (
                <p className="text-xs text-red-500 mt-1">{errors.name}</p>
              )}
            </div>

            {/* 通知方式 */}
            <div>
              <Label className="text-sm font-medium text-gray-700 mb-2">
                通知方式 <span className="text-red-500">*</span>
              </Label>
              <Select
                value={type}
                onValueChange={(value: 'webhook' | 'email') => setType(value)}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="webhook">Webhook通知配置</SelectItem>
                  <SelectItem value="email">邮件通知</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {/* Webhook配置 */}
            {type === 'webhook' && (
              <div>
                <Label className="text-sm font-medium text-gray-700 mb-2">
                  Webhook URL <span className="text-red-500">*</span>
                </Label>
                <Input
                  value={webhookUrl}
                  onChange={(e) => setWebhookUrl(e.target.value)}
                  placeholder="https://example.com/webhook"
                  className={errors.webhookUrl ? 'border-red-500' : ''}
                />
                {errors.webhookUrl && (
                  <p className="text-xs text-red-500 mt-1">
                    {errors.webhookUrl}
                  </p>
                )}
                <p className="text-xs text-gray-500 mt-1">
                  接收通知的Webhook地址
                </p>
              </div>
            )}

            {/* 邮件配置 */}
            {type === 'email' && (
              <>
                <div>
                  <Label className="text-sm font-medium text-gray-700 mb-2">
                    收件人邮箱 <span className="text-red-500">*</span>
                  </Label>
                  <Input
                    value={emailRecipients}
                    onChange={(e) => setEmailRecipients(e.target.value)}
                    placeholder="example@email.com, another@email.com"
                    className={errors.emailRecipients ? 'border-red-500' : ''}
                  />
                  {errors.emailRecipients && (
                    <p className="text-xs text-red-500 mt-1">
                      {errors.emailRecipients}
                    </p>
                  )}
                  <p className="text-xs text-gray-500 mt-1">
                    多个邮箱地址用逗号分隔
                  </p>
                </div>

                <div>
                  <Label className="text-sm font-medium text-gray-700 mb-2">
                    邮件主题 <span className="text-red-500">*</span>
                  </Label>
                  <Input
                    value={emailSubject}
                    onChange={(e) => setEmailSubject(e.target.value)}
                    placeholder="请输入邮件主题"
                    className={errors.emailSubject ? 'border-red-500' : ''}
                  />
                  {errors.emailSubject && (
                    <p className="text-xs text-red-500 mt-1">
                      {errors.emailSubject}
                    </p>
                  )}
                </div>
              </>
            )}
          </div>
        </div>

        {/* 底部操作按钮 */}
        <div className="flex items-center justify-between px-6 py-4 border-t border-gray-100 bg-gray-50">
          <Button
            variant="outline"
            onClick={handleTest}
            disabled={testing}
            className="gap-2"
          >
            <Send className="w-4 h-4" />
            {testing ? '测试中...' : '测试'}
          </Button>
          <div className="flex gap-3">
            <Button variant="outline" onClick={onClose}>
              取消
            </Button>
            <Button onClick={handleSave}>
              {editingNotification ? '保存' : '添加'}
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
