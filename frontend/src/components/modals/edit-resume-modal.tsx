import { useState, useEffect } from 'react';
import { X } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Resume } from '@/types/resume';

interface EditResumeModalProps {
  isOpen: boolean;
  onClose: () => void;
  resume: Resume | null;
  onSave: (updatedResume: Resume) => void;
}

const statusOptions = [
  { value: 'pending', label: '待筛选' },
  { value: 'interview', label: '已邀请面试' },
  { value: 'rejected', label: '已拒绝' },
  { value: 'hired', label: '已录用' },
];

export function EditResumeModal({
  isOpen,
  onClose,
  resume,
  onSave,
}: EditResumeModalProps) {
  const [formData, setFormData] = useState({
    name: '',
    phone: '',
    position: '',
    status: 'pending' as Resume['status'],
  });

  useEffect(() => {
    if (resume) {
      setFormData({
        name: resume.name,
        phone: resume.phone,
        position: '', // 应聘岗位字段是禁用的，不从resume中读取
        status: resume.status,
      });
    }
  }, [resume]);

  const handleSave = () => {
    if (!resume) return;

    const updatedResume: Resume = {
      ...resume,
      name: formData.name,
      phone: formData.phone,
    };

    onSave(updatedResume);
    onClose();
  };

  const handleCancel = () => {
    onClose();
  };

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-2xl">
        <DialogHeader className="flex flex-row items-center justify-between space-y-0 pb-6 border-b">
          <DialogTitle className="text-lg font-medium text-gray-900">
            编辑简历信息
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
            <div className="grid grid-cols-2 gap-6">
              <div className="space-y-2">
                <Label
                  htmlFor="name"
                  className="text-sm font-medium text-gray-700"
                >
                  姓名
                </Label>
                <Input
                  id="name"
                  value={formData.name}
                  onChange={(e) =>
                    setFormData({ ...formData, name: e.target.value })
                  }
                  className="h-10"
                />
              </div>

              <div className="space-y-2">
                <Label
                  htmlFor="position"
                  className="text-sm font-medium text-gray-400"
                >
                  应聘岗位
                </Label>
                <Input
                  id="position"
                  value=""
                  onChange={() => {}}
                  placeholder=""
                  className="h-10 text-gray-400 placeholder:text-gray-400 bg-gray-100 cursor-not-allowed"
                  disabled
                />
              </div>

              <div className="space-y-2">
                <Label
                  htmlFor="phone"
                  className="text-sm font-medium text-gray-700"
                >
                  电话
                </Label>
                <Input
                  id="phone"
                  value={formData.phone}
                  onChange={(e) =>
                    setFormData({ ...formData, phone: e.target.value })
                  }
                  className="h-10"
                />
              </div>

              <div className="space-y-2">
                <Label
                  htmlFor="status"
                  className="text-sm font-medium text-gray-400"
                >
                  简历状态
                </Label>
                <Select value="pending" onValueChange={() => {}} disabled>
                  <SelectTrigger className="h-10 text-gray-400 bg-gray-100 cursor-not-allowed">
                    <SelectValue placeholder="待筛选" />
                  </SelectTrigger>
                  <SelectContent>
                    {statusOptions.map((option) => (
                      <SelectItem key={option.value} value={option.value}>
                        {option.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
          </form>
        </div>

        <div className="flex justify-end gap-3 pt-6 border-t bg-gray-50 -mx-6 -mb-6 px-6 py-4">
          <Button variant="outline" onClick={handleCancel} className="px-6">
            取消
          </Button>
          <Button
            onClick={handleSave}
            className="px-6 bg-green-500 hover:bg-green-600"
          >
            保存修改
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
