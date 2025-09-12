'use client'

import React, { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator } from '@/components/ui/dropdown-menu'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Search, Plus, MessageCircle, Settings, ChevronLeft, ChevronRight, User, LogOut } from 'lucide-react'
import Image from 'next/image'
import { useRouter } from 'next/navigation'
import { getUserProfile, getAvatarUrl, getDisplayName, logout } from '@/lib/api'
import type { UserProfile } from '@/lib/api'
import { UserProfileDialog } from '@/components/business/user-profile-dialog'

interface ChatHistory {
  id: string
  title: string
  timestamp: string
}

interface NavigationItem {
  id: string
  name: string
  active: boolean
}

interface SidebarProps {
  navigationItems?: NavigationItem[]
  onNavigationClick?: (itemId: string, itemName: string) => void
  chatHistory?: ChatHistory[]
  onChatClick?: (chatId: string) => void
  className?: string
}

export function Sidebar({ 
  navigationItems = [],
  onNavigationClick,
  chatHistory = [],
  onChatClick,
  className = ''
}: SidebarProps) {
  const router = useRouter()
  const [searchQuery, setSearchQuery] = useState('')
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)
  const [userProfile, setUserProfile] = useState<UserProfile | null>(null)
  const [isLoadingProfile, setIsLoadingProfile] = useState(true)
  const [isUserProfileDialogOpen, setIsUserProfileDialogOpen] = useState(false)

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

  const handleChatClick = (chatId: string) => {
    if (onChatClick) {
      onChatClick(chatId)
    } else {
      // 默认行为：跳转到聊天页面
      router.push(`/chat/${chatId}`)
    }
  }

  return (
    <>
      <div className={`${sidebarCollapsed ? 'w-16' : 'w-64'} bg-gray-900 text-white flex flex-col transition-all duration-300 ${className}`}>
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
              {navigationItems.length > 0 && (
                <div className="space-y-1">
                  {navigationItems.map((item) => (
                    <button
                      key={item.id}
                      onClick={() => onNavigationClick?.(item.id, item.name)}
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
              )}
            </div>

            {/* 聊天历史 */}
            {chatHistory.length > 0 && (
              <div className="flex-1 px-4">
                <h3 className="text-sm font-medium text-gray-300 mb-3">Recent Chats</h3>
                <ScrollArea className="h-full">
                  <div className="space-y-2">
                    {chatHistory.map((chat) => (
                      <div
                        key={chat.id}
                        onClick={() => handleChatClick(chat.id)}
                        className="p-3 rounded-lg bg-gray-800 hover:bg-gray-700 cursor-pointer transition-colors group"
                      >
                        <div className="flex items-start justify-between">
                          <div className="flex-1 min-w-0">
                            <p className="text-sm text-gray-200 truncate">{chat.title}</p>
                            <p className="text-xs text-gray-400 mt-1">{chat.timestamp}</p>
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
            )}

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
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="sm" className="w-8 h-8 p-0 text-gray-400 hover:text-white">
                  <Settings className="w-4 h-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent side="right">
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
        )}
      </div>
      
      {/* User Profile Dialog */}
      <UserProfileDialog 
        open={isUserProfileDialogOpen} 
        onOpenChange={setIsUserProfileDialogOpen} 
      />
    </>
  )
}

export type { NavigationItem, ChatHistory }