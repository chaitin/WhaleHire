'use client'

import React, { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator } from '@/components/ui/dropdown-menu'
import { ScrollArea } from '@/components/ui/scroll-area'
import { GlassNavigation } from '@/components/custom/glass-navigation'
import { Search, Send, Plus, MessageCircle, FileText, Users, Briefcase, Settings, Mic, ArrowUpRight, ChevronLeft, ChevronRight, User, LogOut } from 'lucide-react'
import Image from 'next/image'
import { useRouter } from 'next/navigation'
import { getUserProfile, getAvatarUrl, getDisplayName, logout } from '@/lib/api'
import type { UserProfile } from '@/lib/api'
import { createConversation } from '@/lib/api/general-agent'
import { UserProfileDialog } from '@/components/business/user-profile-dialog'

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

export default function Dashboard() {
  const router = useRouter()
  const [searchQuery, setSearchQuery] = useState('')
  const [userProfile, setUserProfile] = useState<UserProfile | null>(null)
  const [isLoadingProfile, setIsLoadingProfile] = useState(true)
  const [message, setMessage] = useState('')
  const [isUserProfileDialogOpen, setIsUserProfileDialogOpen] = useState(false)

  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)
  const [chatHistory] = useState<ChatHistory[]>([
    { id: '1', title: 'How can I increase the number of', timestamp: '2小时前' },
    { id: '2', title: "What's the best approach to", timestamp: '1天前' },
    { id: '3', title: "What's the best approach to", timestamp: '2天前' }
  ])

  const [navigationItems, setNavigationItems] = useState([
    { id: 'general', name: 'General', active: true },
    { id: 'jd-agent', name: 'JD Agent', active: false },
    { id: 'screen-agent', name: 'Screen Agent', active: false },
    { id: 'interview-agent', name: 'Interview Agent', active: false }
  ])
  const featureCards: FeatureCard[] = [
    {
      id: '1',
      title: '制定岗位JD',
      description: '智能助手帮您快速制定专业的岗位描述，精准定位人才需求。',
      icon: <FileText className="w-6 h-6" />,
      color: 'bg-white border-gray-200'
    },
    {
      id: '2',
      title: '一键发布JD',
      description: '便捷地将职位发布到多个招聘平台，扩大招聘覆盖范围。',
      icon: <Briefcase className="w-6 h-6" />,
      color: 'bg-white border-gray-200'
    },
    {
      id: '3',
      title: '简历搜索',
      description: '智能匹配算法帮您快速筛选合适的候选人简历。',
      icon: <Search className="w-6 h-6" />,
      color: 'bg-white border-gray-200'
    },
    {
      id: '4',
      title: '候选人评估',
      description: '全方位评估候选人能力，助您做出更好的招聘决策。',
      icon: <Users className="w-6 h-6" />,
      color: 'bg-white border-gray-200'
    }
  ]

  const handleSendMessage = async () => {
    if (!message.trim()) return
    
    const userMessage = message.trim()
    setMessage('')
    
    try {
      const conv = await createConversation({ title: userMessage.slice(0, 30) || '新对话' })
      if (conv && conv.id) {
        router.push(`/chat/${conv.id}?prompt=${encodeURIComponent(userMessage)}`)
      }
    } catch (error) {
      console.error('发送消息失败:', error)
    }
  }

  const handleFeatureClick = (feature: FeatureCard) => {
    console.log('点击功能卡片:', feature.title)
  }

  // 获取用户信息
  useEffect(() => {
    const fetchUserProfile = async () => {
      setIsLoadingProfile(true)
      try {
        const profile = await getUserProfile()
        setUserProfile(profile)
      } catch (error) {
        console.error('获取用户信息失败:', error)
      } finally {
        setIsLoadingProfile(false)
      }
    }

    fetchUserProfile()
  }, [])

  const handleNavigationClick = (itemId: string, _itemName: string) => {
    setNavigationItems(items => 
      items.map(item => ({
        ...item,
        active: item.id === itemId
      }))
    )
  }

  const handleLogout = async () => {
    const success = await logout()
    
    if (success) {
      // 清除本地状态
      setUserProfile(null)
      // 跳转到登录页面
      router.push('/')
    } else {
      console.error('注销失败')
    }
  }

  return (
    <div className="flex h-screen bg-gradient-to-br from-blue-50 to-sky-50">
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
                <h1 className="text-lg font-medium">开始新对话</h1>
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
              <h3 className="text-sm font-medium text-gray-300 mb-3">Recent Chats</h3>
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
                  <AvatarImage src={userProfile ? getAvatarUrl(userProfile.avatar_url) : '/avatar.svg'} />
                  <AvatarFallback className="bg-gray-600 text-white">
                    {userProfile ? getDisplayName(userProfile).charAt(0).toUpperCase() : 'U'}
                  </AvatarFallback>
                </Avatar>
                <div className="flex-1 min-w-0">
                  {isLoadingProfile ? (
                    <p className="text-sm font-medium text-gray-400">加载中...</p>
                  ) : userProfile ? (
                    <>
                      <p className="text-sm font-medium text-gray-200">{getDisplayName(userProfile)}</p>
                      <p className="text-xs text-gray-400 truncate">{userProfile.email}</p>
                    </>
                  ) : (
                    <p className="text-sm font-medium text-gray-400">未登录</p>
                  )}
                </div>
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="ghost" size="sm" className="h-8 w-8 p-0 text-gray-400 hover:text-white">
                      <Settings className="w-4 h-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent side="top">
                     <DropdownMenuItem onClick={() => setIsUserProfileDialogOpen(true)}>
                       <User className="w-4 h-4 mr-2" />
                       用户信息
                     </DropdownMenuItem>
                     <DropdownMenuSeparator />
                     <DropdownMenuItem onClick={handleLogout}>
                       <LogOut className="w-4 h-4 mr-2" />
                       注销登录
                     </DropdownMenuItem>
                   </DropdownMenuContent>
                </DropdownMenu>
              </div>
            </div>
          </>
         )}
         
         {/* 收缩状态下的简化菜单 */}
         {sidebarCollapsed && (
           <div className="flex flex-col items-center space-y-4 py-4">
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
        {/* 液体流动背景效果 */}
        <div className="absolute inset-0">
          {/* 流动的浅蓝色液体形状 */}
          <div className="absolute top-20 left-12 w-64 h-64 bg-gradient-to-br from-blue-200/15 via-sky-200/8 to-cyan-200/12 rounded-full blur-3xl animate-pulse transform rotate-12"></div>
          <div className="absolute top-60 right-16 w-72 h-48 bg-gradient-to-l from-cyan-200/12 via-blue-200/6 to-sky-200/15 rounded-full blur-3xl animate-pulse transform -rotate-6" style={{animationDelay: '2s'}}></div>
          <div className="absolute bottom-32 left-1/3 w-56 h-72 bg-gradient-to-t from-blue-200/10 via-sky-200/12 to-cyan-200/8 rounded-full blur-3xl animate-pulse transform rotate-45" style={{animationDelay: '4s'}}></div>
        </div>
        {/* 顶部导航栏 */}
        <div className="backdrop-blur-sm border-b border-white/20 px-6 py-4 relative z-10">
          <div className="flex items-center justify-between">
            {/* 左侧 WhaleHire 标题 */}
            <div className="flex-shrink-0 w-48">
              <h1 className="text-xl font-bold bg-gradient-to-r from-blue-600 via-blue-500 to-cyan-500 bg-clip-text text-transparent drop-shadow-sm">WhaleHire</h1>
            </div>
            
            {/* 中间导航栏 */}
            <div className="flex-1 flex justify-center">
              <GlassNavigation 
                items={navigationItems}
                onItemClick={handleNavigationClick}
                className="bg-white/30 backdrop-blur-lg border border-white/40 shadow-xl"
              />
            </div>
            
            {/* 右侧占位区域，保持布局平衡 */}
            <div className="flex-shrink-0 w-48 flex justify-end">
              <div className="w-6 h-6"></div>
            </div>
          </div>
        </div>

        {/* 主要内容 */}
        <div className="flex-1 p-8 overflow-auto relative z-10">
          <div className="max-w-4xl mx-auto relative z-10">
            {/* 欢迎信息 */}
            <div className="text-center mb-12">
              <div className="w-16 h-16 rounded-full mx-auto mb-6 flex items-center justify-center relative">
                <Image 
                  src="/logo.svg" 
                  alt="Logo" 
                  width={100} 
                  height={100}
                  className="absolute"
                  style={{
                    position: 'absolute',
                    width: '100px',
                    height: '100px'
                  }}
                />
              </div>
              <h1 className="text-3xl font-bold text-gray-900 mb-4">
                嗨，很高兴见到您！
              </h1>
              <p className="text-lg text-gray-600 max-w-2xl mx-auto leading-relaxed">
                我是您的 AI 招聘伙伴，可以为您提供岗位推荐、简历解析和面试管理等支持。请选择您当前需要的服务，我们将一起高效推进招聘流程。
              </p>
            </div>

            {/* 功能卡片网格 */}
            <div className="grid grid-cols-4 gap-4 mb-12">
              {featureCards.map((feature) => (
                <Card
                  key={feature.id}
                  className={`cursor-pointer transition-all duration-200 hover:shadow-md ${feature.color} shadow-sm relative`}
                  onClick={() => handleFeatureClick(feature)}
                >
                  <CardHeader className="pb-2">
                    <div className="flex items-center space-x-2">
                      <div className="p-1.5 rounded-full bg-gray-100">
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
            <div className="bg-white rounded-2xl p-6 mb-4 shadow_sm">
              <div className="flex items-start space-x-4">
                <div className="p-3 bg-blue-100 rounded-full">
                  <MessageCircle className="w-6 h-6 text-blue-600" />
                </div>
                <div className="flex-1">
                  <h3 className="text-lg font-semibold text-gray-900 mb-2">
                    需要帮助尽管找我
                  </h3>
                  <p className="text-gray-600">
                    如果您有任何招聘相关的问题或需要个性化建议，请随时在下方输入框中告诉我。我会根据您的具体需求提供专业的解决方案。
                  </p>
                </div>
              </div>
            </div>

            {/* 输入框 */}
            <div className="bg-white rounded-2xl p-4 shadow-sm">
              <div className="flex items-end space-x-4">
                <div className="flex-1">
                  <Textarea
                    placeholder="请输入您的问题"
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
                    className="h-10 w-10 p-0 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-300 rounded-full"
                  >
                    <Send className="w-4 h-4" />
                  </Button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      {/* User Profile Dialog */}
      <UserProfileDialog 
        open={isUserProfileDialogOpen} 
        onOpenChange={setIsUserProfileDialogOpen} 
      />
    </div>
  )
}