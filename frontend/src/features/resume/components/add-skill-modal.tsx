import { useState } from 'react';
import { Button } from '@/ui/button';
import { Input } from '@/ui/input';
import { Label } from '@/ui/label';
import { Textarea } from '@/ui/textarea';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/ui/dialog';
import { createJobSkillMeta } from '@/services/job-profile';
import type { JobSkillMeta } from '@/types/job-profile';

interface AddSkillModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSkillAdded: (skill: JobSkillMeta) => void;
}

export function AddSkillModal({
  isOpen,
  onClose,
  onSkillAdded,
}: AddSkillModalProps) {
  const [skillName, setSkillName] = useState('');
  const [skillDescription, setSkillDescription] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // 重置表单
  const resetForm = () => {
    setSkillName('');
    setSkillDescription('');
    setError(null);
  };

  // 关闭弹窗
  const handleClose = () => {
    resetForm();
    onClose();
  };

  // 提交表单
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    // 表单验证
    if (!skillName.trim()) {
      setError('请输入技能名称');
      return;
    }

    try {
      setLoading(true);
      setError(null);

      // 调用API创建技能
      const newSkill = await createJobSkillMeta({
        name: skillName.trim(),
      });

      // 通知父组件技能已添加
      onSkillAdded(newSkill);

      // 关闭弹窗
      handleClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : '添加技能失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={handleClose}>
      <DialogContent
        className="sm:max-w-[425px]"
        onInteractOutside={(e) => e.preventDefault()}
      >
        <DialogHeader>
          <DialogTitle className="text-lg font-semibold text-gray-900">
            添加新技能
          </DialogTitle>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          {/* 错误提示 */}
          {error && (
            <div className="p-3 text-sm text-red-600 bg-red-50 border border-red-200 rounded-md">
              {error}
            </div>
          )}

          {/* 技能名称 */}
          <div className="space-y-2">
            <Label
              htmlFor="skillName"
              className="text-sm font-medium text-gray-700"
            >
              技能名称 <span className="text-red-500">*</span>
            </Label>
            <Input
              id="skillName"
              type="text"
              value={skillName}
              onChange={(e) => setSkillName(e.target.value)}
              placeholder="请输入技能名称，如：JavaScript、React等"
              className="w-full"
              disabled={loading}
            />
          </div>

          {/* 技能描述（可选） */}
          <div className="space-y-2">
            <Label
              htmlFor="skillDescription"
              className="text-sm font-medium text-gray-700"
            >
              技能描述（可选）
            </Label>
            <Textarea
              id="skillDescription"
              value={skillDescription}
              onChange={(e) => setSkillDescription(e.target.value)}
              placeholder="请简要描述该技能的相关信息..."
              className="w-full min-h-[80px]"
              disabled={loading}
            />
          </div>

          {/* 按钮组 */}
          <div className="flex justify-end space-x-3 pt-4">
            <Button
              type="button"
              variant="outline"
              onClick={handleClose}
              disabled={loading}
              className="px-4 py-2"
            >
              取消
            </Button>
            <Button
              type="submit"
              disabled={loading || !skillName.trim()}
              className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white"
            >
              {loading ? '添加中...' : '确认添加'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
