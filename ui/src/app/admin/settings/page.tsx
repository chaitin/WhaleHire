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
  Settings, 
  ChevronLeft, 
  ChevronRight,
  Shield
} from 'lucide-react'
import Image from 'next/image'
import { SystemSettings } from '@/components/admin/system-settings'

interface ChatHistory {
  id: string
  title: string
  timestamp: string
  user?: string
  time?: string
  message?: string
}

export default function AdminSettingsPage() {
  const [searchQuery, setSearchQuery] = useState('')
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)
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
    { id: 'general', name: 'General', active: false },
    { id: 'user-management', name: 'User Management', active: false },
    { id: 'system-monitor', name: 'System Monitor', active: false },
    { id: 'data-analytics', name: 'Data Analytics', active: false },
    { id: 'system-settings', name: 'System Settings', active: true }
  ])

  const handleNavigationClick = (itemId: string, _itemName: string) => {
    if (itemId === 'general') {
      window.location.href = '/admin/dashboard'
      return
    }
    if (itemId === 'system-settings') {
      return // 当前页面，不需要跳转
    }
    
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
                    onClick={() => handleNavigationClick(item.id, item.name)}
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
        <SystemSettings />
      </main>

      {/* 管理员资料对话框 */}
      <AdminProfileDialog 
        open={adminProfileOpen} 
        onOpenChange={setAdminProfileOpen} 
      />
    </div>
  )
}