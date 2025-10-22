import { useState, useEffect } from 'react';
import { X, Loader2 } from 'lucide-react';
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
import {
  createNotificationSetting,
  updateNotificationSetting,
} from '@/services/notification';

interface AddNotificationModalProps {
  open: boolean;
  onClose: () => void;
  onSaveSuccess: () => void;
  onShowMessage: (type: 'success' | 'error', text: string) => void;
  editingNotification: NotificationConfig | null;
}

export function AddNotificationModal({
  open,
  onClose,
  onSaveSuccess,
  onShowMessage,
  editingNotification,
}: AddNotificationModalProps) {
  const [name, setName] = useState(''); // 通知名称
  const [channel, setChannel] = useState<'dingtalk' | 'email'>('dingtalk'); // 通知方式
  const [webhookUrl, setWebhookUrl] = useState(''); // 钉钉Webhook URL
  const [secret, setSecret] = useState(''); // 加签密钥
  const [enabled, setEnabled] = useState(true); // 状态：默认启用
  const [submitting, setSubmitting] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});

  // 编辑模式：填充表单
  useEffect(() => {
    if (editingNotification) {
      setName(editingNotification.name);
      setChannel(editingNotification.channel);
      setWebhookUrl(editingNotification.dingtalk_config?.webhook_url || '');
      setSecret(editingNotification.dingtalk_config?.secret || '');
      setEnabled(editingNotification.enabled);
    } else {
      resetForm();
    }
  }, [editingNotification]);

  // 重置表单
  const resetForm = () => {
    setName('');
    setChannel('dingtalk');
    setWebhookUrl('');
    setSecret('');
    setEnabled(true);
    setErrors({});
  };

  // 从Webhook URL中提取access_token
  const extractToken = (url: string): string | undefined => {
    const match = url.match(/access_token=([^&]+)/);
    return match ? match[1] : undefined;
  };

  // 验证表单
  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!name.trim()) {
      newErrors.name = '请输入通知名称';
    }

    if (channel === 'dingtalk') {
      if (!webhookUrl.trim()) {
        newErrors.webhookUrl = '请输入钉钉Webhook URL';
      } else if (
        !webhookUrl.startsWith('http://') &&
        !webhookUrl.startsWith('https://')
      ) {
        newErrors.webhookUrl = 'URL必须以http://或https://开头';
      }
      if (!secret.trim()) {
        newErrors.secret = '请输入加签密钥';
      }
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  // 保存通知配置
  const handleSave = async () => {
    if (!validateForm()) return;

    setSubmitting(true);
    try {
      const payload: {
        name: string;
        channel: 'dingtalk' | 'email';
        enabled: boolean;
        dingtalk_config?: {
          webhook_url: string;
          secret: string;
          token?: string;
        };
      } = {
        name: name.trim(),
        channel,
        enabled,
      };

      if (channel === 'dingtalk') {
        const token = extractToken(webhookUrl.trim());
        payload.dingtalk_config = {
          webhook_url: webhookUrl.trim(),
          secret: secret.trim(),
          token, // 从URL中提取的token
        };
      }

      if (editingNotification) {
        // 编辑模式：调用更新接口
        await updateNotificationSetting(editingNotification.id, payload);
        onShowMessage('success', '通知配置更新成功');
      } else {
        // 新建模式：调用创建接口
        await createNotificationSetting(payload);
        onShowMessage('success', '通知配置添加成功');
      }

      resetForm();
      onSaveSuccess();
    } catch (error) {
      console.error('保存通知配置失败:', error);
      onShowMessage(
        'error',
        editingNotification ? '更新失败，请稍后重试' : '添加失败，请稍后重试'
      );
    } finally {
      setSubmitting(false);
    }
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
                value={channel}
                onValueChange={(value: 'dingtalk' | 'email') =>
                  setChannel(value)
                }
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="dingtalk">钉钉通知</SelectItem>
                  <SelectItem value="email" disabled>
                    邮件通知（暂不支持）
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>

            {/* 钉钉配置 */}
            {channel === 'dingtalk' && (
              <>
                <div>
                  <Label className="text-sm font-medium text-gray-700 mb-2">
                    钉钉Webhook URL <span className="text-red-500">*</span>
                  </Label>
                  <Input
                    value={webhookUrl}
                    onChange={(e) => setWebhookUrl(e.target.value)}
                    placeholder="https://oapi.dingtalk.com/robot/send?access_token=xxx"
                    className={errors.webhookUrl ? 'border-red-500' : ''}
                  />
                  {errors.webhookUrl && (
                    <p className="text-xs text-red-500 mt-1">
                      {errors.webhookUrl}
                    </p>
                  )}
                  <p className="text-xs text-gray-500 mt-1">
                    请输入钉钉机器人的Webhook地址，系统会自动提取access_token
                  </p>
                </div>

                <div>
                  <Label className="text-sm font-medium text-gray-700 mb-2">
                    加签密钥 <span className="text-red-500">*</span>
                  </Label>
                  <Input
                    value={secret}
                    onChange={(e) => setSecret(e.target.value)}
                    placeholder="请输入加签密钥"
                    className={errors.secret ? 'border-red-500' : ''}
                  />
                  {errors.secret && (
                    <p className="text-xs text-red-500 mt-1">{errors.secret}</p>
                  )}
                  <p className="text-xs text-gray-500 mt-1">
                    用于验证请求的安全密钥
                  </p>
                </div>
              </>
            )}

            {/* 状态选择 */}
            <div>
              <Label className="text-sm font-medium text-gray-700 mb-2">
                状态 <span className="text-red-500">*</span>
              </Label>
              <div className="flex items-center gap-4">
                <label className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="radio"
                    name="enabled"
                    checked={enabled === true}
                    onChange={() => setEnabled(true)}
                    className="w-4 h-4 text-emerald-600 focus:ring-emerald-500"
                  />
                  <span className="text-sm text-gray-700">启用</span>
                </label>
                <label className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="radio"
                    name="enabled"
                    checked={enabled === false}
                    onChange={() => setEnabled(false)}
                    className="w-4 h-4 text-emerald-600 focus:ring-emerald-500"
                  />
                  <span className="text-sm text-gray-700">禁用</span>
                </label>
              </div>
            </div>
          </div>
        </div>

        {/* 底部操作按钮 */}
        <div className="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-100 bg-gray-50">
          <Button variant="outline" onClick={onClose} disabled={submitting}>
            取消
          </Button>
          <Button onClick={handleSave} disabled={submitting}>
            {submitting ? (
              <>
                <Loader2 className="w-4 h-4 animate-spin mr-2" />
                {editingNotification ? '保存中...' : '添加中...'}
              </>
            ) : (
              <>{editingNotification ? '保存' : '添加'}</>
            )}
          </Button>
        </div>
      </div>
    </div>
  );
}
