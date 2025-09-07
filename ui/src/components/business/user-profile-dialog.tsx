"use client"

import React, { useState, useEffect, useRef } from 'react'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { User, Upload, Save, Eye, EyeOff } from 'lucide-react'
import { Switch } from '@/components/ui/switch'
import { getUserProfile, updateProfile, getAvatarUrl, getDisplayName } from '@/lib/api/user'
import type { UserProfile, ProfileUpdateRequest } from '@/lib/api/types'

interface UserProfileDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function UserProfileDialog({ open, onOpenChange }: UserProfileDialogProps) {
  const [userProfile, setUserProfile] = useState<UserProfile | null>(null)
  const [loading, setLoading] = useState(false)
  const [showPassword, setShowPassword] = useState(false)
  const [showOldPassword, setShowOldPassword] = useState(false)
  const [enablePasswordChange, setEnablePasswordChange] = useState(false)
  const [formData, setFormData] = useState({
    username: '',
    password: '',
    oldPassword: '',
    avatar: ''
  })
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  // 获取用户信息
  const fetchUserProfile = async () => {
    setLoading(true)
    try {
      const profile = await getUserProfile()
      if (profile) {
        setUserProfile(profile)
        setFormData({
          username: profile.username,
          password: '',
          oldPassword: '',
          avatar: profile.avatar_url || ''
        })
      }
    } catch (error) {
      console.error('获取用户信息失败:', error)
      setMessage({ type: 'error', text: '获取用户信息失败' })
    } finally {
      setLoading(false)
    }
  }

  // 处理头像上传
  const handleAvatarUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (file) {
      // 检查文件类型
      if (!file.type.startsWith('image/')) {
        setMessage({ type: 'error', text: '请选择图片文件' })
        return
      }
      
      // 检查文件大小 (5MB)
      if (file.size > 5 * 1024 * 1024) {
        setMessage({ type: 'error', text: '图片大小不能超过5MB' })
        return
      }

      // 创建预览URL
      const reader = new FileReader()
      reader.onload = (e) => {
        const result = e.target?.result as string
        setFormData({ ...formData, avatar: result })
      }
      reader.readAsDataURL(file)
    }
  }

  // 更新用户资料
  const handleUpdateProfile = async () => {
    if (!userProfile) return

    setLoading(true)
    setMessage(null)
    
    try {
      const updateParams: ProfileUpdateRequest = {}
      
      // 只有当字段有变化时才包含在更新参数中
      if (formData.username !== userProfile.username) {
        updateParams.username = formData.username
      }
      
      if (enablePasswordChange && formData.password.trim() && formData.oldPassword.trim()) {
        updateParams.password = formData.password
        updateParams.old_password = formData.oldPassword
      }
      
      if (formData.avatar !== (userProfile.avatar_url || '')) {
        updateParams.avatar = formData.avatar
      }

      // 如果没有任何变化，直接返回
      if (Object.keys(updateParams).length === 0) {
        setMessage({ type: 'error', text: '没有检测到任何变化' })
        setLoading(false)
        return
      }

      const updatedProfile = await updateProfile(updateParams)
      if (updatedProfile) {
        setUserProfile(updatedProfile)
        setMessage({ type: 'success', text: '用户资料更新成功' })
        setFormData({ ...formData, password: '', oldPassword: '' }) // 清空密码字段
        setEnablePasswordChange(false) // 关闭密码修改开关
        
        // 如果修改了密码，需要重新登录
        if (updateParams.password) {
          setTimeout(() => {
            onOpenChange(false) // 关闭弹窗
            // window.location.href = '/' // 返回根页面重新登录
          }, 1000) // 延迟1秒让用户看到成功消息
        } else {
          // 如果只是修改了其他信息，直接关闭弹窗
          setTimeout(() => {
            onOpenChange(false)
          }, 1000)
        }
      } else {
        setMessage({ type: 'error', text: '更新用户资料失败' })
      }
    } catch (error) {
      console.error('更新用户资料失败:', error)
      setMessage({ type: 'error', text: '更新用户资料失败' })
    } finally {
      setLoading(false)
    }
  }

  // 当弹窗打开时获取用户信息
  useEffect(() => {
    if (open) {
      fetchUserProfile()
      setMessage(null)
    }
  }, [open])

  // 清除消息
  useEffect(() => {
    if (message) {
      const timer = setTimeout(() => {
        setMessage(null)
      }, 3000)
      return () => clearTimeout(timer)
    }
  }, [message])

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <User className="w-5 h-5" />
            用户资料
          </DialogTitle>
          <DialogDescription>
            编辑您的个人资料信息
          </DialogDescription>
        </DialogHeader>

        {/* 消息提示 */}
        {message && (
          <div className={`p-3 rounded-md text-sm ${
            message.type === 'success' 
              ? 'bg-green-50 text-green-700 border border-green-200' 
              : 'bg-red-50 text-red-700 border border-red-200'
          }`}>
            {message.text}
          </div>
        )}

        {loading ? (
          <div className="flex items-center justify-center py-8">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          </div>
        ) : userProfile ? (
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">个人信息</CardTitle>
            </CardHeader>
            <CardContent className="space-y-6">
              {/* 头像编辑 */}
              <div className="flex flex-col items-center space-y-4">
                <div className="relative">
                  <Avatar className="h-20 w-20">
                    <AvatarImage src={getAvatarUrl(formData.avatar)} />
                    <AvatarFallback className="bg-blue-100 text-blue-600 text-xl">
                      {getDisplayName(userProfile).charAt(0).toUpperCase()}
                    </AvatarFallback>
                  </Avatar>
                  <Button
                    variant="outline"
                    size="sm"
                    className="absolute -bottom-2 -right-2 h-8 w-8 rounded-full p-0"
                    onClick={() => fileInputRef.current?.click()}
                  >
                    <Upload className="h-4 w-4" />
                  </Button>
                  <input
                    ref={fileInputRef}
                    type="file"
                    accept="image/*"
                    onChange={handleAvatarUpload}
                    className="hidden"
                  />
                </div>
                <p className="text-xs text-gray-500 text-center">
                  点击上传按钮更换头像<br />
                  支持 JPG、PNG 格式，大小不超过 5MB
                </p>
              </div>

              {/* 用户名编辑 */}
              <div className="space-y-2">
                <Label htmlFor="username">用户名</Label>
                <Input
                  id="username"
                  value={formData.username}
                  onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                  placeholder="请输入用户名"
                />
              </div>

              {/* 邮箱显示（只读） */}
              <div className="space-y-2">
                <Label className="text-sm font-medium text-gray-500">邮箱</Label>
                <p className="text-sm bg-gray-50 p-2 rounded">{userProfile.email}</p>
              </div>

              {/* 密码修改开关 */}
              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label className="text-sm font-medium">修改密码</Label>
                  <p className="text-xs text-gray-500">开启后可以修改登录密码</p>
                </div>
                <Switch
                   checked={enablePasswordChange}
                   onCheckedChange={(checked: boolean) => {
                     setEnablePasswordChange(checked)
                     if (!checked) {
                       setFormData({ ...formData, password: '', oldPassword: '' })
                     }
                   }}
                 />
              </div>

              {/* 密码修改表单 */}
              {enablePasswordChange && (
                <div className="space-y-4 p-4 bg-gray-50 rounded-lg">
                  {/* 旧密码 */}
                  <div className="space-y-2">
                    <Label htmlFor="oldPassword">当前密码</Label>
                    <div className="relative">
                      <Input
                        id="oldPassword"
                        type={showOldPassword ? 'text' : 'password'}
                        value={formData.oldPassword}
                        onChange={(e) => setFormData({ ...formData, oldPassword: e.target.value })}
                        placeholder="请输入当前密码"
                        required
                      />
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        className="absolute right-0 top-0 h-full px-3 py-2 hover:bg-transparent"
                        onClick={() => setShowOldPassword(!showOldPassword)}
                      >
                        {showOldPassword ? (
                          <EyeOff className="h-4 w-4" />
                        ) : (
                          <Eye className="h-4 w-4" />
                        )}
                      </Button>
                    </div>
                  </div>

                  {/* 新密码 */}
                  <div className="space-y-2">
                    <Label htmlFor="password">新密码</Label>
                    <div className="relative">
                      <Input
                        id="password"
                        type={showPassword ? 'text' : 'password'}
                        value={formData.password}
                        onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                        placeholder="请输入新密码"
                        required
                      />
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        className="absolute right-0 top-0 h-full px-3 py-2 hover:bg-transparent"
                        onClick={() => setShowPassword(!showPassword)}
                      >
                        {showPassword ? (
                          <EyeOff className="h-4 w-4" />
                        ) : (
                          <Eye className="h-4 w-4" />
                        )}
                      </Button>
                    </div>
                    <p className="text-xs text-gray-500">
                      密码长度至少6位，包含字母和数字
                    </p>
                  </div>
                </div>
              )}

              {/* 确认修改按钮 */}
              <div className="flex justify-end space-x-2 pt-4">
                <Button
                  variant="outline"
                  onClick={() => onOpenChange(false)}
                >
                  取消
                </Button>
                <Button
                  onClick={handleUpdateProfile}
                  disabled={loading}
                >
                  {loading ? (
                    <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                  ) : (
                    <Save className="w-4 h-4 mr-2" />
                  )}
                  确认修改
                </Button>
              </div>
            </CardContent>
          </Card>
        ) : (
          <div className="text-center py-8">
            <p className="text-gray-500">无法获取用户信息</p>
            <Button 
              variant="outline" 
              onClick={fetchUserProfile}
              className="mt-4"
            >
              重试
            </Button>
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}