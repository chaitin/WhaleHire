'use client'

import React, { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

import { ScrollArea } from '@/components/ui/scroll-area'
import { AdminProfileDialog } from '@/components/business/admin-profile-dialog'
import { AdminDropdownMenu } from '@/components/business/admin-dropdown-menu'
import { 
  Search, 
  Plus, 
  MessageCircle, 
  Users, 
  Settings, 
  ChevronLeft, 
  ChevronRight,
  Shield,
  BarChart3,
  Database,
  Cog,
  User
} from 'lucide-react'
import Image from 'next/image'

interface ChatHistory {
  id: string
  title: string
  timestamp: string
  user?: string
  time?: string
  message?: string
}

interface FeatureCard {
  id: string
  title: string
  description: string
  icon: React.ReactNode
  color: string
  value?: string
}

export default function AdminDashboard() {
  const [searchQuery, setSearchQuery] = useState('')
  const [message, setMessage] = useState('')
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)
  const [_currentPage, _setCurrentPage] = useState('general')
  const [adminProfileOpen, setAdminProfileOpen] = useState(false)
  
  const [chatHistory] = useState<ChatHistory[]>([
    { 
      id: '1', 
      title: '用户权限管理相关问题', 
      timestamp: '2小时前',
      user: '管理员',
      time: '2小时前',
      message: '用户权限管理相关问题'
    },
    { 
      id: '2', 
      title: '系统数据统计分析', 
      timestamp: '1天前',
      user: '系统',
      time: '1天前',
      message: '系统数据统计分析'
    },
    { 
      id: '3', 
      title: '招聘流程优化建议', 
      timestamp: '2天前',
      user: '用户',
      time: '2天前',
      message: '招聘流程优化建议'
    }
  ])

  const [navigationItems, setNavigationItems] = useState([
    { id: 'general', name: 'General', active: true },
    { id: 'user-management', name: 'User Management', active: false },
    { id: 'system-monitor', name: 'System Monitor', active: false },
    { id: 'data-analytics', name: 'Data Analytics', active: false },
    { id: 'system-settings', name: 'System Settings', active: false }
  ])

  const featureCards: FeatureCard[] = [
    {
      id: '1',
      title: '用户管理',
      description: '管理系统用户，包括普通用户和管理员的权限分配与账户状态。',
      icon: <Users className="w-6 h-6" />,
      color: 'bg-white border-gray-200',
      value: '1,234'
    },
    {
      id: '2',
      title: '权限控制',
      description: '配置系统权限，管理角色分配和访问控制策略。',
      icon: <Shield className="w-6 h-6" />,
      color: 'bg-white border-gray-200',
      value: '56'
    },
    {
      id: '3',
      title: '数据统计',
      description: '查看系统使用统计，分析用户行为和平台运营数据。',
      icon: <BarChart3 className="w-6 h-6" />,
      color: 'bg-white border-gray-200',
      value: '89%'
    },
    {
      id: '4',
      title: '系统监控',
      description: '监控系统运行状态，管理数据库和服务器性能指标。',
      icon: <Database className="w-6 h-6" />,
      color: 'bg-white border-gray-200',
      value: '正常'
    },
    {
      id: '5',
      title: '系统设置',
      description: '配置系统参数，包括OAuth设置、安全策略和基础配置。',
      icon: <Cog className="w-6 h-6" />,
      color: 'bg-white border-gray-200',
      value: '已配置'
    }
  ]

  const _handleSendMessage = () => {
    if (message.trim()) {
      console.log('发送管理员消息:', message)
      setMessage('')
    }
  }

  // 修复handleFeatureClick函数
  const _handleFeatureClick = (_title: string) => {
    // 功能卡片的处理逻辑可以在这里添加
  }

  const _handleNavigationClick = (itemId: string, _itemName: string) => {
    if (itemId === 'system-settings') {
      // 导航到系统设置页面
      window.location.href = '/admin/settings'
      return
    }
    
    _setCurrentPage(itemId)
    setNavigationItems(items => 
      items.map(item => ({
        ...item,
        active: item.id === itemId
      }))
    )
  }

  return (
    <div className="flex h-screen bg-white">
      {/* 左侧边栏 */}
      <div className={`${sidebarCollapsed ? 'w-16' : 'w-64'} bg-gray-900 text-white flex flex-col transition-all duration-300`}>
        {/* 收缩按钮 */}
        <div className="p-2 border-b border-gray-800 flex justify-between items-center">
          <div className="flex items-center">
            <Image 
              src="/logo.svg" 
              alt="Logo" 
              width={24} 
              height={24}
              className="brightness-0 invert"
            />
            {!sidebarCollapsed && (
              <span className="ml-2 text-sm font-medium text-orange-400">Admin</span>
            )}
          </div>
          <Button 
            variant="ghost" 
            size="sm" 
            className="text-white hover:bg-gray-800 p-1"
            onClick={() => setSidebarCollapsed(!sidebarCollapsed)}
          >
            {sidebarCollapsed ? <ChevronRight className="w-4 h-4" /> : <ChevronLeft className="w-4 h-4" />}
          </Button>
        </div>
        
        {!sidebarCollapsed && (
          <>
            {/* 顶部标题区域 */}
            <div className="p-4">
              <div className="flex items-center justify-between mb-6">
                <h1 className="text-lg font-medium">管理员控制台</h1>
                <Button variant="ghost" size="sm" className="text-white hover:bg-gray-800">
                  <Plus className="w-4 h-4" />
                </Button>
              </div>
          
              {/* 搜索框 */}
              <div className="relative mb-6">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4" />
                <Input
                  placeholder="Search"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-10 bg-gray-800 border-gray-700 text-white placeholder-gray-400"
                />
              </div>

              {/* 导航菜单 */}
              <div className="space-y-1">
                {navigationItems.map((item) => (
                  <button
                    key={item.id}
                    onClick={() => _handleNavigationClick(item.id, item.name)}
                    className={`w-full text-left px-3 py-2 rounded-lg text-sm transition-colors ${
                      item.active 
                        ? 'bg-gray-800 text-white' 
                        : 'text-gray-300 hover:bg-gray-800 hover:text-white'
                    }`}
                  >
                    {item.name}
                  </button>
                ))}
              </div>
            </div>

            {/* 聊天历史 */}
            <div className="flex-1 px-4">
              <h3 className="text-sm font-medium text-gray-300 mb-3">Recent Admin Chats</h3>
              <ScrollArea className="h-full">
                <div className="space-y-2">
                  {chatHistory.map((chat) => (
                    <div
                      key={chat.id}
                      className="p-3 rounded-lg bg-gray-800 hover:bg-gray-700 cursor-pointer transition-colors group"
                    >
                      <div className="flex items-start justify-between">
                        <div className="flex-1 min-w-0">
                          <p className="text-sm text-gray-200 truncate">{chat.title}</p>
                        </div>
                        <div className="flex items-center space-x-1 opacity-0 group-hover:opacity-100 transition-opacity">
                          <Button variant="ghost" size="sm" className="h-6 w-6 p-0 text-gray-400 hover:text-white">
                            <MessageCircle className="w-3 h-3" />
                          </Button>
                          <Button variant="ghost" size="sm" className="h-6 w-6 p-0 text-gray-400 hover:text-white">
                            <Settings className="w-3 h-3" />
                          </Button>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </ScrollArea>
            </div>

            {/* 底部用户信息 */}
            <div className="p-4 border-t border-gray-700">
              <AdminDropdownMenu />
            </div>
          </>
         )}
         
         {/* 收缩状态下的简化菜单 */}
         {sidebarCollapsed && (
           <div className="flex flex-col items-center space-y-4 py-4">
             <Button variant="ghost" size="sm" className="w-8 h-8 p-0 text-orange-400 hover:text-white">
               <Shield className="w-4 h-4" />
             </Button>
             <Button variant="ghost" size="sm" className="w-8 h-8 p-0 text-gray-400 hover:text-white">
               <Plus className="w-4 h-4" />
             </Button>
             <Button variant="ghost" size="sm" className="w-8 h-8 p-0 text-gray-400 hover:text-white">
               <Search className="w-4 h-4" />
             </Button>
             <Button variant="ghost" size="sm" className="w-8 h-8 p-0 text-gray-400 hover:text-white">
               <MessageCircle className="w-4 h-4" />
             </Button>
             <div className="flex-1"></div>
             <AdminDropdownMenu collapsed={true} />
           </div>
         )}
      </div>

      {/* 主内容区域 */}
      <main className="flex-1 p-6 overflow-auto">
        <>
          {/* 搜索栏 */}
          <div className="mb-6">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-4 w-4" />
              <input
                  type="text"
                  placeholder="搜索用户、消息或设置..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                />
              </div>
            </div>

            {/* 功能卡片网格 */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-8">
              {featureCards.map((card, index) => (
                <div
                  key={index}
                  onClick={() => _handleFeatureClick(card.title)}
                  className="bg-white rounded-lg shadow-md p-6 hover:shadow-lg transition-shadow cursor-pointer border border-gray-200"
                >
                  <div className="flex items-center mb-4">
                    <div className="p-2 bg-purple-100 rounded-lg mr-3">
                      {card.icon}
                    </div>
                    <h3 className="text-lg font-semibold text-gray-900">{card.title}</h3>
                  </div>
                  <p className="text-gray-600 mb-4">{card.description}</p>
                  <div className="text-2xl font-bold text-purple-600">{card.value}</div>
                </div>
              ))}
            </div>

            {/* 聊天历史 */}
            <div className="bg-white rounded-lg shadow-md">
              <div className="p-6 border-b border-gray-200">
                <h2 className="text-xl font-semibold text-gray-900">最近活动</h2>
              </div>
              <div className="p-6">
                <div className="space-y-4">
                  {chatHistory.map((chat, index) => (
                    <div key={index} className="flex items-start space-x-3 p-3 hover:bg-gray-50 rounded-lg">
                      <div className="flex-shrink-0">
                        <div className="w-8 h-8 bg-purple-100 rounded-full flex items-center justify-center">
                          <User className="h-4 w-4 text-purple-600" />
                        </div>
                      </div>
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center justify-between">
                          <p className="text-sm font-medium text-gray-900">{chat.user}</p>
                          <p className="text-xs text-gray-500">{chat.time}</p>
                        </div>
                        <p className="text-sm text-gray-600 truncate">{chat.message}</p>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </>
      </main>

      {/* 管理员资料对话框 */}
      <AdminProfileDialog 
        open={adminProfileOpen} 
        onOpenChange={setAdminProfileOpen} 
      />
    </div>
  )
}