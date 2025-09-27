'use client'

import React, { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator } from '@/components/ui/dropdown-menu'
import { Settings, User, LogOut } from 'lucide-react'
import { getAdminProfile, adminLogout } from '@/lib/api/admin'
import type { AdminProfile } from '@/lib/api/admin'
import { AdminProfileDialog } from '@/components/business/admin-profile-dialog'

interface AdminDropdownMenuProps {
  className?: string
  collapsed?: boolean
}

export function AdminDropdownMenu({ className = '', collapsed = false }: AdminDropdownMenuProps) {
  const [adminProfile, setAdminProfile] = useState<AdminProfile | null>(null)
  const [isLoadingProfile, setIsLoadingProfile] = useState(true)
  const [isAdminProfileDialogOpen, setIsAdminProfileDialogOpen] = useState(false)

  // 获取管理员资料
  useEffect(() => {
    const fetchAdminProfile = async () => {
      try {
        setIsLoadingProfile(true)
        const profile = await getAdminProfile()
        if (profile) {
          setAdminProfile(profile)
        } else {
          console.error('获取管理员资料失败')
        }
      } catch (error) {
        console.error('获取管理员资料失败:', error)
      } finally {
        setIsLoadingProfile(false)
      }
    }

    fetchAdminProfile()
  }, [])

  // 处理管理员退出登录
  const handleLogout = async () => {
    try {
      const result = await adminLogout()
      if (result.success) {
        // 清除本地状态
        setAdminProfile(null)
        // 跳转到首页
        window.location.href = '/'
      } else {
        console.error('注销失败:', result.message)
      }
    } catch (error) {
      console.error('注销失败:', error)
    }
  }

  // 收缩状态下的简化菜单
  if (collapsed) {
    return (
      <>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="sm" className="w-8 h-8 p-0 text-gray-400 hover:text-white">
              <Settings className="w-4 h-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent side="right">
            <DropdownMenuItem onClick={() => setIsAdminProfileDialogOpen(true)}>
              <User className="w-4 h-4 mr-2" />
              管理员信息
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={handleLogout}>
              <LogOut className="w-4 h-4 mr-2" />
              注销登录
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>

        {/* Admin Profile Dialog */}
        <AdminProfileDialog 
          open={isAdminProfileDialogOpen} 
          onOpenChange={setIsAdminProfileDialogOpen} 
        />
      </>
    )
  }

  // 完整状态下的用户信息和菜单
  return (
    <>
      <div className={`flex items-center space-x-3 ${className}`}>
        <div className="flex-1 min-w-0">
          {isLoadingProfile ? (
            <p className="text-sm font-medium text-gray-400">加载中...</p>
          ) : adminProfile ? (
            <>
              <p className="text-sm font-medium text-gray-200">{adminProfile.username || '管理员'}</p>
              <p className="text-xs text-gray-400">管理员</p>
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
            <DropdownMenuItem onClick={() => setIsAdminProfileDialogOpen(true)}>
              <User className="w-4 h-4 mr-2" />
              管理员信息
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={handleLogout}>
              <LogOut className="w-4 h-4 mr-2" />
              注销登录
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>

      {/* Admin Profile Dialog */}
      <AdminProfileDialog 
        open={isAdminProfileDialogOpen} 
        onOpenChange={setIsAdminProfileDialogOpen} 
      />
    </>
  )
}