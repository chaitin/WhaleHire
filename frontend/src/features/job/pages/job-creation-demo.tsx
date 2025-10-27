import { useState } from 'react';
import { Button } from '@/ui/button';
import { Plus } from 'lucide-react';
import { CreateJobModal } from '../components/create-job-modal';

/**
 * 岗位创建演示页面
 * 用于展示新的岗位创建界面
 */
export function JobCreationDemo() {
  const [isModalOpen, setIsModalOpen] = useState(false);

  return (
    <div className="min-h-screen bg-gray-50 p-8">
      <div className="max-w-4xl mx-auto">
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-8">
          <h1 className="text-2xl font-bold text-gray-900 mb-4">
            岗位画像 - 新建岗位
          </h1>
          <p className="text-gray-600 mb-6">
            点击下方按钮打开岗位创建界面,体验全新的UI设计
          </p>

          <Button
            onClick={() => setIsModalOpen(true)}
            className="flex items-center gap-2 bg-blue-600 hover:bg-blue-700 text-white"
          >
            <Plus className="h-5 w-5" />
            新建岗位
          </Button>

          <div className="mt-8 p-4 bg-blue-50 border border-blue-200 rounded-lg">
            <h3 className="text-sm font-semibold text-blue-900 mb-2">
              设计特点:
            </h3>
            <ul className="text-sm text-blue-800 space-y-1 list-disc list-inside">
              <li>清晰的模式选择:支持手动编辑和AI编辑两种模式</li>
              <li>图标装饰:每个字段都配有相应的图标,提升视觉效果</li>
              <li>优雅的表单布局:合理的间距和分组,提升用户体验</li>
              <li>统一的色彩方案:使用蓝色作为主色调,保持一致性</li>
              <li>圆角设计:所有输入框和按钮都采用圆角设计,更加现代</li>
              <li>聚焦效果:输入框聚焦时有蓝色边框和阴影效果</li>
            </ul>
          </div>
        </div>
      </div>

      {/* 岗位创建弹窗 */}
      <CreateJobModal open={isModalOpen} onOpenChange={setIsModalOpen} />
    </div>
  );
}
