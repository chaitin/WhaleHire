import { useState } from 'react';
import {
  Sparkles,
  Send,
  Trash2,
  Save,
  MapPin,
  Briefcase,
  GraduationCap,
} from 'lucide-react';
import { Button } from '@/ui/button';
import { Textarea } from '@/ui/textarea';
import { JobProfilePreview } from './job-profile-preview';
import type { JobProfileDetail } from '@/types/job-profile';

interface GeneratedJobContent {
  name: string;
  salaryRange: string;
  location: string;
  experience: string;
  education: string;
  responsibilities: string[];
  requirements: string[];
  otherInfo: string[];
}

interface AIJobEditorProps {
  onClose?: () => void;
  onSave?: (content: GeneratedJobContent) => void;
}

// 常用模板数据
const templates = [
  {
    id: 1,
    name: '前端开发工程师',
    prompt:
      '生成一个前端开发工程师的岗位，要求3-5年工作经验、本科以上学历、熟悉React/Vue框架…',
  },
  {
    id: 2,
    name: 'UI/UX设计师',
    prompt:
      '生成一个UI/UX设计师的岗位，要求2-3年设计经验、熟练使用Figma、Sketch等工具…',
  },
  {
    id: 3,
    name: '产品经理',
    prompt:
      '生成一个产品经理的岗位，要求3-5年工作经验、硕士以上学历、熟悉敏捷开发流程…',
  },
  {
    id: 4,
    name: '数据分析师',
    prompt: '生成一个数据分析师的岗位，要求熟悉SQL、Python，有数据可视化经验…',
  },
];

// 模拟生成的岗位数据
const mockGeneratedJob = {
  name: '高级前端开发工程师',
  salaryRange: '30K-50K/月',
  location: '北京市海淀区',
  experience: '5年以上工作经验',
  education: '本科及以上学历',
  responsibilities: [
    '负责前端架构设计与技术选型，制定前端开发规范',
    '带领前端团队完成复杂Web应用的开发与优化',
    '与产品、设计、后端团队紧密协作，推动产品迭代',
    '优化前端性能，提升用户体验',
    '参与技术文档编写与知识分享',
  ],
  requirements: [
    '精通HTML5、CSS3、JavaScript',
    '熟练掌握React、Vue等主流前端框架',
    '熟悉Webpack、Vite等构建工具',
    '具备良好的代码规范和编程习惯',
    '5年以上前端开发经验，2年以上团队管理经验',
    '有大型Web应用开发经验者优先',
    '计算机相关专业本科及以上学历',
    '有电商、金融类产品开发经验者优先',
  ],
  otherInfo: [
    '公司提供五险一金、补充医疗保险、年度体检',
    '弹性工作制度，周末双休，法定节假日正常休息',
    '定期团队建设活动，技术分享交流会',
    '提供专业技能培训和职业发展通道',
  ],
};

export function AIJobEditor({ onSave }: AIJobEditorProps) {
  const [prompt, setPrompt] = useState('');
  const [generatedContent, setGeneratedContent] =
    useState<GeneratedJobContent | null>(null);

  // 预览弹窗状态
  const [isPreviewOpen, setIsPreviewOpen] = useState(false);
  const [previewJob, setPreviewJob] = useState<JobProfileDetail | null>(null);

  // 处理模板点击
  const handleTemplateClick = (templatePrompt: string) => {
    setPrompt(templatePrompt);
  };

  // 处理生成内容
  const handleGenerate = async () => {
    if (!prompt.trim()) return;

    // 模拟生成过程
    setTimeout(() => {
      setGeneratedContent(mockGeneratedJob);
    }, 2000);
  };

  // 处理清空输入
  const handleClearInput = () => {
    setPrompt('');
  };

  // 处理保存内容
  const handleSaveContent = () => {
    if (generatedContent && onSave) {
      onSave(generatedContent);
    }
  };

  // 处理查看按钮点击（保存概览）
  const handleViewClick = (job: GeneratedJobContent) => {
    // 转换为 JobProfileDetail 格式
    const jobDetail: JobProfileDetail = {
      id: '1',
      name: job.name,
      description: job.otherInfo?.join('\n') || '',
      department: '',
      department_id: '',
      location: job.location || '',
      salary_min: parseInt(job.salaryRange?.split('-')[0]) || 0,
      salary_max: parseInt(job.salaryRange?.split('K')[0].split('-')[1]) || 0,
      work_type: 'full_time',
      status: 'published',
      created_at: Date.now(),
      updated_at: Date.now(),
      responsibilities:
        job.responsibilities?.map((r: string, index: number) => ({
          id: String(index + 1),
          job_id: '1',
          responsibility: r,
          sort_order: index + 1,
        })) || [],
      skills: [],
      education_requirements: [],
      experience_requirements: [],
      industry_requirements: [],
    };
    setPreviewJob(jobDetail);
    setIsPreviewOpen(true);
  };

  return (
    <>
      <div className="flex flex-col gap-6">
        {/* 区域1：AI交互输入区域 */}
        <div className="bg-white rounded-lg shadow-sm border border-[#E5E7EB] overflow-hidden">
          <div className="px-6 py-4 border-b border-[#E5E7EB]">
            <h3 className="text-[18px] font-semibold text-[#000000]">
              AI交互输入
            </h3>
          </div>

          <div className="p-6 space-y-4">
            {/* 输入框 */}
            <Textarea
              placeholder="你可以对我说生成一个产品经理的岗位，要求3-5年工作经验、硕士以上学历、熟悉敏捷开发流程…"
              value={prompt}
              onChange={(e) => setPrompt(e.target.value)}
              className="min-h-[128px] resize-none border-[#D1D5DB] text-[14px] placeholder:text-[#9CA3AF]"
              disabled={true}
            />

            {/* 操作按钮 */}
            <div className="flex items-center gap-3">
              <Button
                onClick={handleGenerate}
                disabled={true}
                className="gap-2 bg-[#2563EB] hover:bg-[#1D4ED8] text-white shadow-sm"
              >
                <Send className="h-4 w-4" />
                生成内容
              </Button>

              <Button
                onClick={handleClearInput}
                disabled={true}
                variant="outline"
                className="gap-2 border-[#D1D5DB] text-[#374151] hover:bg-[#F9FAFB]"
              >
                <Trash2 className="h-4 w-4" />
                清空输入
              </Button>
            </div>

            {/* 常用模板 */}
            <div className="space-y-3">
              <span className="text-[14px] text-[#6B7280]">常用模板:</span>
              <div className="flex flex-wrap gap-2">
                {templates.map((template) => (
                  <button
                    key={template.id}
                    onClick={() => handleTemplateClick(template.prompt)}
                    disabled={true}
                    className="px-3 py-1.5 bg-[#EFF6FF] text-[#2563EB] rounded text-[14px] hover:bg-[#DBEAFE] transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {template.name}
                  </button>
                ))}
              </div>
            </div>
          </div>
        </div>

        {/* 区域2：保存概览和生成内容预览（左右布局） */}
        <div className="flex gap-6 min-h-[500px]">
          {/* 左侧：保存概览 */}
          <div className="w-[380px] bg-white rounded-lg shadow-sm border border-[#E5E7EB] overflow-hidden flex flex-col">
            <div className="px-6 py-4 border-b border-[#E5E7EB] flex items-center justify-between flex-shrink-0">
              <h3 className="text-[18px] font-semibold text-[#000000]">
                保存概览
              </h3>
            </div>

            <div className="flex-1 overflow-y-auto p-6">
              <div className="space-y-4">
                {/* 示例历史记录卡片1 */}
                <div className="border border-[#E5E7EB] rounded-lg p-4 hover:shadow-md transition-shadow">
                  <div className="flex items-center justify-between mb-2">
                    <h4 className="text-[16px] font-medium text-[#1D4ED8]">
                      高级前端开发工程师画像
                    </h4>
                    <span className="px-2 py-1 bg-[#DCFCE7] text-[#166534] text-[12px] rounded-full">
                      已保存
                    </span>
                  </div>
                  <p className="text-[14px] text-[#4B5563] mb-3 line-clamp-2">
                    负责前端架构设计与实现，带领团队完成复杂Web应用开发，要求5年以上前端开发经验...
                  </p>
                  <div className="flex items-center justify-between text-[12px]">
                    <span className="text-[#6B7280]">2023-11-15 14:30</span>
                    <div className="flex gap-2">
                      <button
                        onClick={() => handleViewClick(mockGeneratedJob)}
                        disabled={true}
                        className="text-[#2563EB] hover:underline disabled:opacity-50 disabled:cursor-not-allowed"
                      >
                        查看
                      </button>
                      <button
                        disabled={true}
                        className="text-[#4B5563] hover:underline disabled:opacity-50 disabled:cursor-not-allowed"
                      >
                        编辑
                      </button>
                    </div>
                  </div>
                </div>

                {/* 示例历史记录卡片2 */}
                <div className="border border-[#E5E7EB] rounded-lg p-4 hover:shadow-md transition-shadow">
                  <div className="flex items-center justify-between mb-2">
                    <h4 className="text-[16px] font-medium text-[#1D4ED8]">
                      UI/UX设计师岗位画像
                    </h4>
                    <span className="px-2 py-1 bg-[#DCFCE7] text-[#166534] text-[12px] rounded-full">
                      已保存
                    </span>
                  </div>
                  <p className="text-[14px] text-[#4B5563] mb-3 line-clamp-2">
                    负责产品界面设计与用户体验优化，要求精通Figma、Sketch等设计工具...
                  </p>
                  <div className="flex items-center justify-between text-[12px]">
                    <span className="text-[#6B7280]">2023-11-12 09:15</span>
                    <div className="flex gap-2">
                      <button
                        onClick={() => handleViewClick(mockGeneratedJob)}
                        disabled={true}
                        className="text-[#2563EB] hover:underline disabled:opacity-50 disabled:cursor-not-allowed"
                      >
                        查看
                      </button>
                      <button
                        disabled={true}
                        className="text-[#4B5563] hover:underline disabled:opacity-50 disabled:cursor-not-allowed"
                      >
                        编辑
                      </button>
                    </div>
                  </div>
                </div>

                {/* 示例历史记录卡片3 */}
                <div className="border border-[#E5E7EB] rounded-lg p-4 hover:shadow-md transition-shadow">
                  <div className="flex items-center justify-between mb-2">
                    <h4 className="text-[16px] font-medium text-[#1D4ED8]">
                      数据分析师岗位画像
                    </h4>
                    <span className="px-2 py-1 bg-[#DCFCE7] text-[#166534] text-[12px] rounded-full">
                      已保存
                    </span>
                  </div>
                  <p className="text-[14px] text-[#4B5563] mb-3 line-clamp-2">
                    负责业务数据收集、清洗与分析，制作数据报表，要求熟悉SQL、Python等工具...
                  </p>
                  <div className="flex items-center justify-between text-[12px]">
                    <span className="text-[#6B7280]">2023-11-10 16:45</span>
                    <div className="flex gap-2">
                      <button
                        onClick={() => handleViewClick(mockGeneratedJob)}
                        disabled={true}
                        className="text-[#2563EB] hover:underline disabled:opacity-50 disabled:cursor-not-allowed"
                      >
                        查看
                      </button>
                      <button
                        disabled={true}
                        className="text-[#4B5563] hover:underline disabled:opacity-50 disabled:cursor-not-allowed"
                      >
                        编辑
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* 右侧：生成内容概览 */}
          <div className="flex-1 bg-white rounded-lg shadow-sm border border-[#E5E7EB] overflow-hidden flex flex-col">
            <div className="px-6 py-4 border-b border-[#E5E7EB] flex items-center justify-between flex-shrink-0">
              <h3 className="text-[18px] font-semibold text-[#000000]">
                生成内容概览
              </h3>
              {generatedContent && (
                <Button
                  onClick={handleSaveContent}
                  disabled={true}
                  className="gap-2 bg-[#16A34A] hover:bg-[#15803D] text-white shadow-sm"
                >
                  <Save className="h-4 w-4" />
                  保存内容
                </Button>
              )}
            </div>

            <div className="flex-1 overflow-y-auto max-h-[700px]">
              {generatedContent ? (
                <div className="p-6 border border-[#E5E7EB] rounded-lg m-6 space-y-6">
                  {/* 岗位标题和薪资 */}
                  <div className="flex items-center gap-6">
                    <h2 className="text-[24px] font-semibold text-[#1F2937]">
                      {generatedContent.name}
                    </h2>
                    <span className="text-[18px] text-[#4B5563]">
                      {generatedContent.salaryRange}
                    </span>
                  </div>

                  {/* 标签信息 */}
                  <div className="bg-[#F9FAFB] rounded-lg p-4 flex items-center gap-6">
                    <div className="flex items-center gap-2 text-[16px] text-[#4B5563]">
                      <MapPin className="h-4 w-4" />
                      <span>{generatedContent.location}</span>
                    </div>
                    <div className="flex items-center gap-2 text-[16px] text-[#4B5563]">
                      <Briefcase className="h-4 w-4" />
                      <span>{generatedContent.experience}</span>
                    </div>
                    <div className="flex items-center gap-2 text-[16px] text-[#4B5563]">
                      <GraduationCap className="h-5 w-5" />
                      <span>{generatedContent.education}</span>
                    </div>
                  </div>

                  {/* 岗位职责 */}
                  <div className="space-y-3">
                    <h3 className="text-[16px] font-semibold text-[#374151]">
                      岗位职责
                    </h3>
                    <ul className="space-y-2">
                      {generatedContent.responsibilities.map(
                        (item: string, index: number) => (
                          <li
                            key={index}
                            className="text-[16px] text-[#4B5563] pl-5 relative before:content-['•'] before:absolute before:left-0"
                          >
                            {item}
                          </li>
                        )
                      )}
                    </ul>
                  </div>

                  {/* 任职要求 */}
                  <div className="space-y-3">
                    <h3 className="text-[16px] font-semibold text-[#374151]">
                      任职要求
                    </h3>
                    <ul className="space-y-2">
                      {generatedContent.requirements.map(
                        (item: string, index: number) => (
                          <li
                            key={index}
                            className="text-[16px] text-[#4B5563] pl-5 relative before:content-['•'] before:absolute before:left-0"
                          >
                            {item}
                          </li>
                        )
                      )}
                    </ul>
                  </div>

                  {/* 其他信息 */}
                  <div className="space-y-3">
                    <h3 className="text-[16px] font-semibold text-[#374151]">
                      其他信息
                    </h3>
                    <ul className="space-y-2">
                      {generatedContent.otherInfo.map(
                        (item: string, index: number) => (
                          <li
                            key={index}
                            className="text-[16px] text-[#4B5563] pl-5 relative before:content-['•'] before:absolute before:left-0"
                          >
                            {item}
                          </li>
                        )
                      )}
                    </ul>
                  </div>
                </div>
              ) : (
                <div className="flex items-center justify-center h-full">
                  <div className="text-center space-y-3">
                    <Sparkles className="h-12 w-12 text-[#D1D5DB] mx-auto" />
                    <p className="text-[14px] text-[#9CA3AF]">
                      输入您的需求，AI将为您生成岗位内容
                    </p>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* 预览弹窗 */}
      {previewJob && (
        <JobProfilePreview
          open={isPreviewOpen}
          onOpenChange={setIsPreviewOpen}
          jobProfile={previewJob}
        />
      )}
    </>
  );
}
