'use client'

import React, { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { GlassNavigation } from '@/components/custom/glass-navigation'
import { Sidebar, type NavigationItem } from '@/components/custom/sidebar'
import { Send, FileText, Users, Briefcase, Mic, ArrowUpRight, MessageCircle, Search } from 'lucide-react'
import Image from 'next/image'
import { useRouter } from 'next/navigation'
import { createConversation } from '@/lib/api/general-agent'

interface FeatureCard {
  id: string
  title: string
  description: string
  icon: React.ReactNode
  color: string
}

export default function Dashboard() {
  const router = useRouter()
  const [message, setMessage] = useState('')



  const [navigationItems, setNavigationItems] = useState<NavigationItem[]>([
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
    
    // 获取当前选中的 agent
    const activeAgent = navigationItems.find(item => item.active)
    const agentName = activeAgent?.id || 'general'
    
    try {
      const conv = await createConversation({ 
        title: userMessage.slice(0, 30) || '新对话',
        agent_name: agentName
      })
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



  const handleNavigationClick = (itemId: string, _itemName: string) => {
    setNavigationItems(items => 
      items.map(item => ({
        ...item,
        active: item.id === itemId
      }))
    )
  }

  // 处理聊天历史点击
  const handleChatClick = (chatId: string) => {
    router.push(`/chat/${chatId}`)
  }

  return (
    <div className="flex h-screen bg-gradient-to-br from-blue-50 to-sky-50">
      {/* 左侧边栏 */}
      <Sidebar 
        navigationItems={navigationItems}
        onNavigationClick={handleNavigationClick}
        onChatClick={handleChatClick}
      />

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
      

    </div>
  )
}