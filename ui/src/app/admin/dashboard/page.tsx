'use client'

import React, { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { ScrollArea } from '@/components/ui/scroll-area'
import { GlassNavigation } from '@/components/custom/glass-navigation'
import { 
  Search, 
  Send, 
  Plus, 
  MessageCircle, 
  Users, 
  Settings, 
  Mic, 
  ArrowUpRight, 
  ChevronLeft, 
  ChevronRight,
  Shield,
  UserCheck,
  BarChart3,
  Database
} from 'lucide-react'
import Image from 'next/image'

interface ChatHistory {
  id: string
  title: string
  timestamp: string
}

interface FeatureCard {
  id: string
  title: string
  description: string
  icon: React.ReactNode
  color: string
}

export default function AdminDashboard() {
  const [searchQuery, setSearchQuery] = useState('')
  const [message, setMessage] = useState('')
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)
  
  const [chatHistory] = useState<ChatHistory[]>([
    { id: '1', title: '用户权限管理相关问题', timestamp: '2小时前' },
    { id: '2', title: '系统数据统计分析', timestamp: '1天前' },
    { id: '3', title: '招聘流程优化建议', timestamp: '2天前' }
  ])

  const [navigationItems, setNavigationItems] = useState([
    { id: 'general', name: 'General', active: true },
    { id: 'user-management', name: 'User Management', active: false },
    { id: 'system-monitor', name: 'System Monitor', active: false },
    { id: 'data-analytics', name: 'Data Analytics', active: false }
  ])

  const featureCards: FeatureCard[] = [
    {
      id: '1',
      title: '用户管理',
      description: '管理系统用户，包括普通用户和管理员的权限分配与账户状态。',
      icon: <Users className="w-6 h-6" />,
      color: 'bg-white border-gray-200'
    },
    {
      id: '2',
      title: '权限控制',
      description: '配置系统权限，管理角色分配和访问控制策略。',
      icon: <Shield className="w-6 h-6" />,
      color: 'bg-white border-gray-200'
    },
    {
      id: '3',
      title: '数据统计',
      description: '查看系统使用统计，分析用户行为和平台运营数据。',
      icon: <BarChart3 className="w-6 h-6" />,
      color: 'bg-white border-gray-200'
    },
    {
      id: '4',
      title: '系统监控',
      description: '监控系统运行状态，管理数据库和服务器性能指标。',
      icon: <Database className="w-6 h-6" />,
      color: 'bg-white border-gray-200'
    }
  ]

  const handleSendMessage = () => {
    if (message.trim()) {
      console.log('发送管理员消息:', message)
      setMessage('')
    }
  }

  const handleFeatureClick = (feature: FeatureCard) => {
    console.log('点击管理员功能卡片:', feature.title)
  }

  const handleNavigationClick = (itemId: string, _itemName: string) => {
    setNavigationItems(items => 
      items.map(item => ({
        ...item,
        active: item.id === itemId
      }))
    )
  }

  return (
    <div className="flex h-screen bg-gradient-to-br from-purple-50 to-indigo-50">
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
              <div className="flex items-center space-x-3">
                <Avatar className="h-8 w-8">
                  <AvatarImage src="/avatar.svg" />
                  <AvatarFallback className="bg-orange-600 text-white">A</AvatarFallback>
                </Avatar>
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium text-gray-200">Admin User</p>
                  <p className="text-xs text-gray-400">管理员</p>
                </div>
                <Button variant="ghost" size="sm" className="h-8 w-8 p-0 text-gray-400 hover:text-white">
                  <Settings className="w-4 h-4" />
                </Button>
              </div>
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
             <Button variant="ghost" size="sm" className="w-8 h-8 p-0 text-gray-400 hover:text-white">
               <Settings className="w-4 h-4" />
             </Button>
           </div>
         )}
      </div>

      {/* 主要内容区域 */}
      <div className="flex-1 flex flex-col relative">
        {/* 液体流动背景效果 - 管理员主题色 */}
        <div className="absolute inset-0">
          {/* 流动的紫色液体形状 */}
          <div className="absolute top-20 left-12 w-64 h-64 bg-gradient-to-br from-purple-200/15 via-indigo-200/8 to-violet-200/12 rounded-full blur-3xl animate-pulse transform rotate-12"></div>
          <div className="absolute top-60 right-16 w-72 h-48 bg-gradient-to-l from-violet-200/12 via-purple-200/6 to-indigo-200/15 rounded-full blur-3xl animate-pulse transform -rotate-6" style={{animationDelay: '2s'}}></div>
          <div className="absolute bottom-32 left-1/3 w-56 h-72 bg-gradient-to-t from-purple-200/10 via-indigo-200/12 to-violet-200/8 rounded-full blur-3xl animate-pulse transform rotate-45" style={{animationDelay: '4s'}}></div>
        </div>
        
        {/* 顶部导航栏 */}
        <div className="backdrop-blur-sm border-b border-white/20 px-6 py-4 relative z-10">
          <div className="flex items-center justify-between">
            {/* 左侧 WhaleHire Admin 标题 */}
            <div className="flex-shrink-0">
              <h1 className="text-xl font-bold bg-gradient-to-r from-purple-600 via-indigo-500 to-violet-500 bg-clip-text text-transparent drop-shadow-sm">
                WhaleHire Admin
              </h1>
            </div>
            
            {/* 中间导航栏 */}
            <div className="flex-1 flex justify-center">
              <GlassNavigation 
                items={navigationItems}
                onItemClick={handleNavigationClick}
                className="bg-white/30 backdrop-blur-lg border border-white/40 shadow-xl"
              />
            </div>
          </div>
        </div>

        {/* 主要内容 */}
        <div className="flex-1 p-8 overflow-auto relative z-10">
          <div className="max-w-4xl mx-auto relative z-10">
            {/* 欢迎信息 */}
            <div className="text-center mb-12">
              <div className="w-16 h-16 rounded-full mx-auto mb-6 flex items-center justify-center relative">
                <div className="absolute inset-0 bg-gradient-to-r from-purple-500 to-indigo-500 rounded-full opacity-20"></div>
                <Shield className="w-8 h-8 text-purple-600 relative z-10" />
              </div>
              <h1 className="text-3xl font-bold text-gray-900 mb-4">
                欢迎来到管理员控制台
              </h1>
              <p className="text-lg text-gray-600 max-w-2xl mx-auto leading-relaxed">
                您正在使用管理员权限访问系统。在这里您可以管理用户账户、配置系统权限、查看数据统计和监控系统运行状态。请谨慎操作，确保系统安全稳定运行。
              </p>
            </div>

            {/* 功能卡片网格 */}
            <div className="grid grid-cols-4 gap-4 mb-12">
              {featureCards.map((feature) => (
                <Card
                  key={feature.id}
                  className={`cursor-pointer transition-all duration-200 hover:shadow-md ${feature.color} shadow-sm relative hover:shadow-purple-100`}
                  onClick={() => handleFeatureClick(feature)}
                >
                  <CardHeader className="pb-2">
                    <div className="flex items-center space-x-2">
                      <div className="p-1.5 rounded-full bg-purple-100">
                        {feature.icon}
                      </div>
                      <CardTitle className="text-base font-semibold text-gray-900">
                        {feature.title}
                      </CardTitle>
                    </div>
                  </CardHeader>
                  <CardContent className="pt-0">
                    <p className="text-sm text-gray-600">{feature.description}</p>
                  </CardContent>
                  <div className="absolute top-4 right-4">
                    <ArrowUpRight className="w-4 h-4 text-gray-400" />
                  </div>
                </Card>
              ))}
            </div>
          </div>
        </div>

        {/* 底部输入区域 */}
        <div className="backdrop-blur-sm p-6 relative z-10">
          <div className="max-w-4xl mx-auto relative z-10">
            {/* AI助手提示卡片 */}
            <div className="bg-white rounded-2xl p-6 mb-4 shadow-sm border border-purple-100">
              <div className="flex items-start space-x-4">
                <div className="p-3 bg-purple-100 rounded-full">
                  <UserCheck className="w-6 h-6 text-purple-600" />
                </div>
                <div className="flex-1">
                  <h3 className="text-lg font-semibold text-gray-900 mb-2">
                    管理员AI助手
                  </h3>
                  <p className="text-gray-600">
                    作为管理员，您可以询问关于用户管理、权限配置、系统监控等相关问题。我会为您提供专业的管理建议和操作指导，帮助您更好地管理WhaleHire平台。
                  </p>
                </div>
              </div>
            </div>

            {/* 输入框 */}
            <div className="bg-white rounded-2xl p-4 shadow-sm border border-purple-100">
              <div className="flex items-end space-x-4">
                <div className="flex-1">
                  <Textarea
                    placeholder="请输入您的管理问题或需求"
                    value={message}
                    onChange={(e) => setMessage(e.target.value)}
                    className="min-h-[50px] max-h-32 resize-none border-0 focus:ring-0 bg-transparent text-gray-900 placeholder-gray-500"
                    onKeyDown={(e) => {
                      if (e.key === 'Enter' && !e.shiftKey) {
                        e.preventDefault()
                        handleSendMessage()
                      }
                    }}
                  />
                </div>
                <div className="flex items-center space-x-2">
                  <Button variant="ghost" size="sm" className="h-10 w-10 p-0">
                    <Mic className="w-5 h-5 text-gray-500" />
                  </Button>
                  <Button
                    onClick={handleSendMessage}
                    disabled={!message.trim()}
                    className="h-10 w-10 p-0 bg-purple-600 hover:bg-purple-700 disabled:bg-gray-300 rounded-full"
                  >
                    <Send className="w-4 h-4" />
                  </Button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}