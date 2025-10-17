import { useState } from 'react';
import { X } from 'lucide-react';
import { EmailConfigTab } from './EmailConfigTab';

interface ResumeCollectionModalProps {
  isOpen: boolean;
  onClose: () => void;
}

type TabType = 'email' | 'crawler' | 'agent';

export function ResumeCollectionModal({
  isOpen,
  onClose,
}: ResumeCollectionModalProps) {
  const [activeTab, setActiveTab] = useState<TabType>('email');

  if (!isOpen) return null;

  const tabs: { id: TabType; label: string; icon: string }[] = [
    { id: 'email', label: '邮箱配置', icon: '📧' },
    { id: 'crawler', label: '爬虫配置', icon: '🕷️' },
    { id: 'agent', label: 'Agent配置', icon: '🤖' },
  ];

  return (
    <div
      className="fixed inset-0 flex items-center justify-center bg-black bg-opacity-50 z-[1000]"
      onClick={onClose}
    >
      <div
        className="bg-white rounded-lg shadow-xl w-[1100px] max-h-[90vh] overflow-hidden flex flex-col"
        onClick={(e) => e.stopPropagation()}
      >
        {/* 弹窗头部 */}
        <div className="flex items-center justify-between px-6 py-5 border-b border-gray-100">
          <h2 className="text-xl font-semibold text-primary">简历收集配置</h2>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600 transition-colors w-6 h-6"
          >
            <X className="w-6 h-6" />
          </button>
        </div>

        {/* 选项卡导航 */}
        <div className="flex items-center border-b border-gray-200 bg-gray-50">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`
                relative px-8 py-4 text-sm font-medium transition-all
                ${
                  activeTab === tab.id
                    ? 'text-primary bg-white border-b-2 border-primary'
                    : 'text-gray-600 hover:text-gray-900 hover:bg-gray-100'
                }
              `}
            >
              <span className="mr-2 text-base">{tab.icon}</span>
              {tab.label}
            </button>
          ))}
        </div>

        {/* 选项卡内容区域 */}
        <div className="flex-1 overflow-y-auto">
          {activeTab === 'email' && <EmailConfigTab />}
          {activeTab === 'crawler' && <CrawlerConfigTabContent />}
          {activeTab === 'agent' && <AgentConfigTabContent />}
        </div>
      </div>
    </div>
  );
}

// 爬虫配置选项卡内容
function CrawlerConfigTabContent() {
  return (
    <div className="flex flex-col items-center justify-center py-20">
      <div className="text-6xl mb-4">🕷️</div>
      <h3 className="text-xl font-medium text-gray-900 mb-2">爬虫配置</h3>
      <p className="text-gray-500">该功能正在开发中...</p>
    </div>
  );
}

// Agent配置选项卡内容
function AgentConfigTabContent() {
  return (
    <div className="flex flex-col items-center justify-center py-20">
      <div className="text-6xl mb-4">🤖</div>
      <h3 className="text-xl font-medium text-gray-900 mb-2">Agent配置</h3>
      <p className="text-gray-500">该功能正在开发中...</p>
    </div>
  );
}
