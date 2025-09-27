"use client"

import React, { useState, useEffect } from 'react'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Shield, Save, Eye, EyeOff } from 'lucide-react'
import { Switch } from '@/components/ui/switch'
import { getAdminProfile, updateAdminProfile } from '@/lib/api/admin'
import type { AdminProfile, AdminProfileUpdateRequest } from '@/lib/api/admin'

interface AdminProfileDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function AdminProfileDialog({ open, onOpenChange }: AdminProfileDialogProps) {
  const [adminProfile, setAdminProfile] = useState<AdminProfile | null>(null)
  const [loading, setLoading] = useState(false)
  const [showPassword, setShowPassword] = useState(false)
  const [showOldPassword, setShowOldPassword] = useState(false)
  const [enablePasswordChange, setEnablePasswordChange] = useState(false)
  const [formData, setFormData] = useState({
    username: '',
    password: '',
    oldPassword: ''
  })
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null)

  // 获取管理员信息
  const fetchAdminProfile = async () => {
    setLoading(true)
    try {
      const profile = await getAdminProfile()
      if (profile) {
        setAdminProfile(profile)
        setFormData({
          username: profile.username,
          password: '',
          oldPassword: ''
        })
      }
    } catch (error) {
      console.error('获取管理员信息失败:', error)
      setMessage({ type: 'error', text: '获取管理员信息失败' })
    } finally {
      setLoading(false)
    }
  }

  // 更新管理员资料
  const handleUpdateProfile = async () => {
    if (!adminProfile) return

    setLoading(true)
    setMessage(null)
    
    try {
      const updateParams: AdminProfileUpdateRequest = {
        username: adminProfile.username,
        password: '',
        old_password: ''
      }
      
      // 只有当字段有变化时才包含在更新参数中
      if (formData.username !== adminProfile.username) {
        updateParams.username = formData.username
      }
      
      if (enablePasswordChange && formData.password.trim() && formData.oldPassword.trim()) {
        updateParams.password = formData.password
        updateParams.old_password = formData.oldPassword
      }

      // 如果没有任何变化，直接返回
      if (Object.keys(updateParams).length === 0) {
        setMessage({ type: 'error', text: '没有检测到任何变化' })
        setLoading(false)
        return
      }

      const updatedProfile = await updateAdminProfile(updateParams)
      if (updatedProfile) {
        setAdminProfile(updatedProfile)
        setMessage({ type: 'success', text: '管理员资料更新成功' })
        setFormData({ ...formData, password: '', oldPassword: '' }) // 清空密码字段
        setEnablePasswordChange(false) // 关闭密码修改开关
        
        // 如果修改了密码，需要重新登录
        if (updateParams.password) {
          setTimeout(() => {
            onOpenChange(false) // 关闭弹窗
            window.location.href = '/' // 返回管理员登录页面
          }, 1000) // 延迟1秒让用户看到成功消息
        } else {
          // 如果只是修改了其他信息，直接关闭弹窗
          setTimeout(() => {
            onOpenChange(false)
          }, 1000)
        }
      } else {
        setMessage({ type: 'error', text: '更新管理员资料失败' })
      }
    } catch (error) {
      console.error('更新管理员资料失败:', error)
      setMessage({ type: 'error', text: '更新管理员资料失败' })
    } finally {
      setLoading(false)
    }
  }

  // 当弹窗打开时获取管理员信息
  useEffect(() => {
    if (open) {
      fetchAdminProfile()
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
            <Shield className="w-5 h-5" />
            管理员资料
          </DialogTitle>
          <DialogDescription>
            编辑管理员资料信息
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
        ) : adminProfile ? (
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">管理员信息</CardTitle>
            </CardHeader>
            <CardContent className="space-y-6">
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

              {/* 角色信息展示 */}
              <div className="space-y-2">
                <Label className="text-sm font-medium text-gray-600">角色</Label>
                <div className="p-3 bg-gray-50 rounded-lg">
                  <p className="text-sm text-gray-900">
                    {adminProfile.role?.name || '未分配角色'}
                  </p>
                  {adminProfile.role?.description && (
                    <p className="text-xs text-gray-500 mt-1">
                      {adminProfile.role.description}
                    </p>
                  )}
                </div>
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

              {/* 操作按钮 */}
              <div className="flex justify-end pt-4">
                {/* 确认修改和取消按钮 */}
                <div className="flex space-x-2">
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
              </div>
            </CardContent>
          </Card>
        ) : (
          <div className="text-center py-8">
            <p className="text-gray-500">无法获取管理员信息</p>
            <Button 
              variant="outline" 
              onClick={fetchAdminProfile}
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