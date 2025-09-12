'use client'

import React, { useState, useEffect, useRef, useCallback } from 'react'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Card } from '@/components/ui/card'
import { Send, Mic, Copy, ThumbsUp, ThumbsDown } from 'lucide-react'
import { useRouter, useParams, useSearchParams } from 'next/navigation'
import { getConversationHistory, generateStream } from '@/lib/api/general-agent'
import type { Conversation, Message, GenerateRequest, StreamChunk } from '@/lib/api/types'
import { Sidebar } from '@/components/custom/sidebar'
import type { NavigationItem } from '@/components/custom/sidebar'




export default function ChatPage() {
  const router = useRouter()
  const params = useParams()
  const searchParams = useSearchParams()
  const conversationId = params.conversationId as string
  
  const [_conversation, setConversation] = useState<Conversation | null>(null)
  const [messages, setMessages] = useState<Message[]>([])
  const [message, setMessage] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [isGenerating, setIsGenerating] = useState(false)
  const [streamingContent, setStreamingContent] = useState('')
  const [displayedContent, setDisplayedContent] = useState('')
  const [streamingMessageId, setStreamingMessageId] = useState<string | null>(null)
  const scrollAreaRef = useRef<HTMLDivElement>(null)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  
  // 侧边栏状态已由Sidebar组件内部处理
  
  const [navigationItems, setNavigationItems] = useState<NavigationItem[]>([
    { id: 'general', name: 'General', active: true },
    { id: 'jd-agent', name: 'JD Agent', active: false },
    { id: 'screen-agent', name: 'Screen Agent', active: false },
    { id: 'interview-agent', name: 'Interview Agent', active: false }
  ])

  // 处理导航点击
  const handleNavigationClick = (itemId: string) => {
    setNavigationItems(items => 
      items.map(item => ({ ...item, active: item.id === itemId }))
    )
    
    // 根据导航项进行路由跳转
    if (itemId === 'dashboard') {
      router.push('/dashboard')
    }
  }

  // 处理聊天历史点击
  const handleChatClick = (chatId: string) => {
    router.push(`/chat/${chatId}`)
  }

  // 滚动到底部
  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }


  // 发送消息
  const handleSendMessage = useCallback(async () => {
    if (!message.trim() || isGenerating) return
    
    const userMessage = message.trim()
    setMessage('')
    
    // 添加用户消息到界面（临时显示）
    const tempUserMessage: Message = {
      id: `temp-user-${Date.now()}`,
      conversation_id: conversationId,
      role: 'user',
      content: userMessage,
      type: 'text',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    }
    
    setMessages(prev => [...prev, tempUserMessage])
    setIsGenerating(true)
    
    try {
      const request: GenerateRequest = {
        prompt: userMessage,
        conversation_id: conversationId,
        history: messages
      }
      
      let assistantContent = ''
      const assistantMessageId = `temp-assistant-${Date.now()}`
      
      // 立即创建一个临时的助手消息
      const tempAssistantMessage: Message = {
        id: assistantMessageId,
        conversation_id: conversationId,
        role: 'assistant',
        content: '',
        type: 'text',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      }
      
      // 立即添加到消息列表
      setMessages(prev => [...prev, tempAssistantMessage])
      
      // 设置流式消息ID和初始状态
      setStreamingMessageId(assistantMessageId)
      setStreamingContent('')
      setDisplayedContent('')
      
      await generateStream(
        request,
        (chunk: StreamChunk) => {
          if (chunk.content && !chunk.done) {
            assistantContent += chunk.content
           // 逐块 append
            setDisplayedContent(prev => prev + chunk.content)
          }
        },
        (error: string) => {
          console.error('流式生成失败:', error)
          
          // 显示错误消息
          const errorMessage: Message = {
            id: `error-${Date.now()}`,
            conversation_id: conversationId,
            role: 'assistant',
            content: '抱歉，生成回复时出现了错误，请稍后重试。',
            type: 'text',
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString()
          }
          
          // 替换失败的消息为错误消息
          setMessages(prev => 
            prev.map(msg => 
              msg.id === assistantMessageId ? errorMessage : msg
            )
          )
          
          // 清理流式状态
          setStreamingMessageId(null)
          setStreamingContent('')
          setDisplayedContent('')
          setIsGenerating(false)
        },
        async () => {
          // 生成完成，确保显示完整内容
          setDisplayedContent(assistantContent)
          
          // 延迟一下再更新最终消息内容和清理状态
          setTimeout(() => {
            setMessages(prev => 
              prev.map(msg => 
                msg.id === assistantMessageId 
                  ? { ...msg, content: assistantContent }
                  : msg
              )
            )
            
            // 清理流式状态
            setStreamingMessageId(null)
            setStreamingContent('')
            setDisplayedContent('')
            setIsGenerating(false)
            
            // 聊天历史会由Sidebar组件自动刷新
          }, 500) // 延迟500ms确保用户能看到完整的流式效果
        }
      )
    } catch (error) {
      console.error('发送消息失败:', error)
      
      // 显示错误提示
      const errorMessage: Message = {
        id: `error-${Date.now()}`,
        conversation_id: conversationId,
        role: 'assistant',
        content: '网络连接出现问题，请检查网络后重试。',
        type: 'text',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      }
      
      setMessages(prev => [...prev, errorMessage])
      setIsGenerating(false)
    }
  }, [conversationId, message, messages, isGenerating])

  // 聊天历史加载已移至Sidebar组件内部

  // 加载对话历史
  useEffect(() => {
    const loadConversation = async () => {
      if (!conversationId) return
      
      // 如果URL中有prompt参数，说明是从dashboard页面发送消息创建的新对话，不加载历史消息
      const promptFromUrl = searchParams.get('prompt')
      if (promptFromUrl) {
        setIsLoading(false)
        return
      }
      
      // 只在初次加载或conversationId变化时显示loading
      // 如果已经有消息数据，则不显示loading避免闪烁
      if (messages.length === 0) {
        setIsLoading(true)
      }
      
      try {
        const result = await getConversationHistory({ conversation_id: conversationId })
        if (result) {
          setConversation(result)
          setMessages(result.messages || [])
        }
      } catch (error) {
        console.error('加载对话历史失败:', error)
      } finally {
        setIsLoading(false)
      }
    }

    loadConversation()
  }, [conversationId, searchParams])

  // 检查 URL 参数中的 prompt 并自动发送
  useEffect(() => {
    const promptFromUrl = searchParams.get('prompt')
    if (promptFromUrl && !isLoading && messages.length === 0 && !isGenerating) {
      setMessage(promptFromUrl)
      // 延迟执行以确保组件完全加载
      setTimeout(() => {
        handleSendMessage()
      }, 100)
    }
  }, [searchParams, isLoading, messages.length, isGenerating, handleSendMessage])

  // 逐字显示效果
  useEffect(() => {
    if (!streamingContent || !streamingMessageId) {
      return
    }
    
    // 如果流式内容比当前显示内容长，继续逐字显示
    if (streamingContent.length > displayedContent.length) {
      const timer = setTimeout(() => {
        setDisplayedContent(streamingContent.slice(0, displayedContent.length + 1))
      }, 20) // 20ms间隔，更快的显示速度
      
      return () => clearTimeout(timer)
    }
  }, [streamingContent, streamingMessageId, displayedContent])

  // 自动滚动到底部
  useEffect(() => {
    scrollToBottom()
  }, [messages, displayedContent])

  // 复制消息内容
  const handleCopyMessage = (content: string) => {
    navigator.clipboard.writeText(content)
  }

  // 格式化时间
  const formatTime = (timestamp: string) => {
    return new Date(timestamp).toLocaleTimeString('zh-CN', {
      hour: '2-digit',
      minute: '2-digit'
    })
  }

  if (isLoading) {
    return (
      <div className="flex h-screen bg-gradient-to-br from-blue-50 to-sky-50">
        {/* 保持侧边栏显示，避免布局闪烁 */}
        <Sidebar
          navigationItems={navigationItems}
          onNavigationClick={handleNavigationClick}
          onChatClick={handleChatClick}
        />
        <div className="flex-1 flex items-center justify-center">
          <div className="text-gray-500">加载中...</div>
        </div>
      </div>
    )
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
      <div className="flex-1 flex flex-col">

        {/* 消息区域 */}
        <div className="flex-1 overflow-hidden">
          <ScrollArea className="h-full" ref={scrollAreaRef}>
            <div className="max-w-4xl mx-auto p-6 space-y-6">
              {messages.map((msg) => (
                <div key={msg.id} className="flex space-x-4">
                  {/* 头像 */}
                  <Avatar className="w-8 h-8 flex-shrink-0">
                    {msg.role === 'user' ? (
                      <AvatarFallback className="bg-blue-100 text-blue-600">
                        U
                      </AvatarFallback>
                    ) : (
                      <AvatarFallback className="bg-green-100 text-green-600">
                        AI
                      </AvatarFallback>
                    )}
                  </Avatar>
                  
                  {/* 消息内容 */}
                  <div className="flex-1 space-y-2">
                    <div className="flex items-center space-x-2">
                      <span className="text-sm font-medium text-gray-900">
                        {msg.role === 'user' ? '您' : 'AI助手'}
                      </span>
                      <span className="text-xs text-gray-500">
                        {formatTime(msg.created_at)}
                      </span>
                    </div>
                    
                    <Card className="p-4 bg-white shadow-sm">
                      <div className="prose prose-sm max-w-none">
                        {/* 如果是正在生成的消息且还没有内容，显示思考状态 */}
                        {msg.id === streamingMessageId && !streamingContent && !displayedContent ? (
                          <div className="flex items-center space-x-2">
                            <div className="w-2 h-2 bg-blue-500 rounded-full animate-bounce"></div>
                            <div className="w-2 h-2 bg-blue-500 rounded-full animate-bounce" style={{animationDelay: '0.1s'}}></div>
                            <div className="w-2 h-2 bg-blue-500 rounded-full animate-bounce" style={{animationDelay: '0.2s'}}></div>
                            <span className="text-sm text-gray-600 ml-2">正在思考您的问题...</span>
                          </div>
                        ) : (
                          <p className="text-gray-800 whitespace-pre-wrap">
                            {msg.id === streamingMessageId ? displayedContent : msg.content}
                            {msg.id === streamingMessageId && streamingContent && displayedContent.length < streamingContent.length && (
                              <span className="inline-block w-2 h-5 bg-blue-500 animate-pulse ml-1"></span>
                            )}
                          </p>
                        )}
                      </div>
                      
                      {/* 消息操作按钮 */}
                      {msg.role === 'assistant' && (
                        <div className="flex items-center space-x-2 mt-3 pt-3 border-t border-gray-100">
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleCopyMessage(msg.content || '')}
                            className="text-gray-500 hover:text-gray-700"
                          >
                            <Copy className="w-3 h-3 mr-1" />
                            复制
                          </Button>
                          {/* 如果是错误消息，显示重试按钮 */}
                          {(msg.content?.includes('出现了错误') || msg.content?.includes('网络连接出现问题')) ? (
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => {
                                // 获取上一条用户消息重新发送
                                const userMessages = messages.filter(m => m.role === 'user')
                                const lastUserMessage = userMessages[userMessages.length - 1]
                                if (lastUserMessage?.content) {
                                  setMessage(lastUserMessage.content)
                                  setTimeout(() => handleSendMessage(), 100)
                                }
                              }}
                              className="text-blue-500 hover:text-blue-700"
                              disabled={isGenerating}
                            >
                              <Send className="w-3 h-3 mr-1" />
                              重试
                            </Button>
                          ) : (
                            <>
                              <Button
                                variant="ghost"
                                size="sm"
                                className="text-gray-500 hover:text-gray-700"
                              >
                                <ThumbsUp className="w-3 h-3 mr-1" />
                                赞
                              </Button>
                              <Button
                                variant="ghost"
                                size="sm"
                                className="text-gray-500 hover:text-gray-700"
                              >
                                <ThumbsDown className="w-3 h-3 mr-1" />
                                踩
                              </Button>
                            </>
                          )}
                        </div>
                      )}
                    </Card>
                  </div>
                </div>
              ))}
              
              <div ref={messagesEndRef} />
            </div>
          </ScrollArea>
        </div>

        {/* 底部输入区域 */}
        <div className="bg-white/80 backdrop-blur-sm border-t border-gray-200 p-6">
          <div className="max-w-4xl mx-auto">
            <div className="bg-white rounded-2xl p-4 shadow-sm border border-gray-200">
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
                    disabled={isGenerating}
                  />
                </div>
                <div className="flex items-center space-x-2">
                  <Button 
                    variant="ghost" 
                    size="sm" 
                    className="h-10 w-10 p-0"
                    disabled={isGenerating}
                  >
                    <Mic className="w-5 h-5 text-gray-500" />
                  </Button>
                  <Button
                    onClick={handleSendMessage}
                    disabled={!message.trim() || isGenerating}
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