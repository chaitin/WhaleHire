'use client'

import React, { useState, useEffect, useCallback, useRef, forwardRef, useImperativeHandle } from 'react'
import { Button } from '@/components/ui/button'
import { MessageCircle, Settings, Loader2 } from 'lucide-react'
import { listConversations } from '@/lib/api/general-agent'
import type { ListConversationsRequest, Conversation } from '@/lib/api/types'

interface ChatHistoryItem {
  id: string
  title: string
  timestamp: string
}

interface ChatHistoryProps {
  onChatClick?: (chatId: string) => void
  className?: string
}

export interface ChatHistoryRef {
  loadMore: () => void
}

export const ChatHistory = forwardRef<ChatHistoryRef, ChatHistoryProps>(({ onChatClick, className = '' }, ref) => {
  const [chatHistory, setChatHistory] = useState<ChatHistoryItem[]>([])
  const [loading, setLoading] = useState(false)
  const [hasMore, setHasMore] = useState(true)
  const [_page, setPage] = useState(1)
  const loadingRef = useRef<HTMLDivElement>(null)
  const loadChatHistoryRef = useRef<((pageNum: number, reset?: boolean) => Promise<void>) | null>(null)

  // 格式化相对时间
  const formatRelativeTime = useCallback((timestamp: string) => {
    const now = new Date()
    const time = new Date(timestamp)
    const diffInMs = now.getTime() - time.getTime()
    const diffInHours = Math.floor(diffInMs / (1000 * 60 * 60))
    const diffInDays = Math.floor(diffInHours / 24)
    
    if (diffInHours < 1) {
      return '刚刚'
    } else if (diffInHours < 24) {
      return `${diffInHours}小时前`
    } else if (diffInDays < 7) {
      return `${diffInDays}天前`
    } else {
      return time.toLocaleDateString('zh-CN')
    }
  }, [])

  // 加载聊天历史记录
  const loadChatHistory = useCallback(async (pageNum: number, reset: boolean = false) => {
    if (loading) return
    
    setLoading(true)
    try {
      const request: ListConversationsRequest = {
        page: pageNum,
        size: 10 // 每页加载10条记录
      }
      
      const result = await listConversations(request)
      if (result && result.conversations) {
        const newHistory: ChatHistoryItem[] = result.conversations.map((conv: Conversation) => ({
          id: conv.id,
          title: conv.title || '新对话',
          timestamp: formatRelativeTime(conv.updated_at || conv.created_at)
        }))
        
        if (reset) {
          setChatHistory(newHistory)
        } else {
          setChatHistory(prev => {
            // 过滤掉重复的记录，避免key冲突
            const existingIds = new Set(prev.map(item => item.id))
            const uniqueNewHistory = newHistory.filter(item => !existingIds.has(item.id))
            return [...prev, ...uniqueNewHistory]
          })
        }
        
        // 检查是否还有更多数据
        const hasMoreData = result.page_info ? 
          (result.page_info.page * result.page_info.size) < result.page_info.total : 
          newHistory.length === 10
        setHasMore(hasMoreData)
      } else {
        setHasMore(false)
      }
    } catch (error) {
      console.error('加载聊天历史失败:', error)
      setHasMore(false)
    } finally {
      setLoading(false)
    }
  }, [loading, formatRelativeTime])

  // 更新ref引用
  loadChatHistoryRef.current = loadChatHistory

  // 处理聊天点击
  const handleChatClick = useCallback((chatId: string) => {
    onChatClick?.(chatId)
  }, [onChatClick])

  // 暴露加载更多方法给父组件
  const handleLoadMore = useCallback(() => {
    if (hasMore && !loading) {
      setPage(prev => {
        const nextPage = prev + 1
        loadChatHistory(nextPage)
        return nextPage
      })
    }
  }, [hasMore, loading, loadChatHistory])

  // 暴露方法给父组件
  useImperativeHandle(ref, () => ({
    loadMore: handleLoadMore
  }), [handleLoadMore])

  // 初始加载
  useEffect(() => {
    const initialLoad = async () => {
      if (loadChatHistoryRef.current) {
        await loadChatHistoryRef.current(1, true)
      }
    }
    initialLoad()
  }, [])

  // 刷新聊天历史
  const refreshChatHistory = useCallback(() => {
    setPage(1)
    setHasMore(true)
    const refresh = async () => {
      await loadChatHistory(1, true)
    }
    refresh()
  }, [loadChatHistory])

  return (
    <div className={`flex-1 px-4 ${className}`}>
      <div className="flex items-center justify-between mb-3">
        <h3 className="text-xs font-medium text-gray-300">Recent Chats</h3>
        <Button 
          variant="ghost" 
          size="sm" 
          className="h-6 w-6 p-0 text-gray-400 hover:text-white"
          onClick={refreshChatHistory}
          disabled={loading}
        >
          <MessageCircle className="w-3 h-3" />
        </Button>
      </div>
      
      <div className="space-y-1.5">
          {loading && chatHistory.length === 0 && (
            <div className="flex items-center justify-center py-4">
              <div className="flex items-center space-x-2 text-gray-400">
                <Loader2 className="w-3 h-3 animate-spin" />
                <span className="text-xs">加载中...</span>
              </div>
            </div>
          )}
          {chatHistory.map((chat) => (
            <div
              key={chat.id}
              onClick={() => handleChatClick(chat.id)}
              className="p-2.5 rounded-md bg-gray-800 hover:bg-gray-700 cursor-pointer transition-colors group"
            >
              <div className="flex items-start justify-between">
                <div className="flex-1 min-w-0 pr-2">
                  <p className="text-xs text-gray-200 truncate leading-tight mb-1">
                    {chat.title}
                  </p>
                  <p className="text-xs text-gray-400">
                    {chat.timestamp}
                  </p>
                </div>
                <div className="flex items-center space-x-1 opacity-0 group-hover:opacity-100 transition-opacity flex-shrink-0">
                  <Button 
                    variant="ghost" 
                    size="sm" 
                    className="h-5 w-5 p-0 text-gray-400 hover:text-white"
                    onClick={(e) => {
                      e.stopPropagation()
                      // TODO: 实现设置功能
                    }}
                  >
                    <Settings className="w-2.5 h-2.5" />
                  </Button>
                </div>
              </div>
            </div>
          ))}
          
          {/* 加载更多指示器 */}
          {hasMore && (
            <div 
              ref={loadingRef}
              className="flex items-center justify-center py-4"
            >
              {loading && (
                <div className="flex items-center space-x-2 text-gray-400">
                  <Loader2 className="w-3 h-3 animate-spin" />
                  <span className="text-xs">加载中...</span>
                </div>
              )}
            </div>
          )}
          
          {/* 无更多数据提示 */}
          {!hasMore && chatHistory.length > 0 && (
            <div className="text-center py-2">
              <p className="text-xs text-gray-500">没有更多聊天记录了</p>
            </div>
          )}
          
          {/* 空状态 */}
          {!loading && chatHistory.length === 0 && (
            <div className="text-center py-8">
              <MessageCircle className="w-8 h-8 text-gray-500 mx-auto mb-2" />
              <p className="text-xs text-gray-500">暂无聊天记录</p>
            </div>
          )}
        </div>
    </div>
  )
})

ChatHistory.displayName = 'ChatHistory'

export type { ChatHistoryItem }