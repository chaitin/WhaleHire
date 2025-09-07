"use client"

import React, { useState, useEffect } from 'react'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { User, Edit3, Save, X, Eye, EyeOff } from 'lucide-react'
import { getUserProfile, updateUser, getAvatarUrl, getDisplayName } from '@/lib/api/user'
import type { UserProfile, UpdateUserRequest } from '@/lib/api/types'

interface UserProfileDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function UserProfileDialog({ open, onOpenChange }: UserProfileDialogProps) {
  const [userProfile, setUserProfile] = useState<UserProfile | null>(null)
  const [loading, setLoading] = useState(false)
  const [isEditing, setIsEditing] = useState(false)
  const [showPassword, setShowPassword] = useState(false)
  const [formData, setFormData] = useState({
    password: '',
    status: 'active' as 'active' | 'locked' | 'inactive'
  })
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null)

  // 获取用户信息
  const fetchUserProfile = async () => {
    setLoading(true)
    try {
      const profile = await getUserProfile()
      if (profile) {
        setUserProfile(profile)
        setFormData({
          password: '',
          status: profile.status as 'active' | 'locked' | 'inactive'
        })
      }
    } catch (error) {
      console.error('获取用户信息失败:', error)
      setMessage({ type: 'error', text: '获取用户信息失败' })
    } finally {
      setLoading(false)
    }
  }

  // 更新用户信息
  const handleUpdateUser = async () => {
    if (!userProfile) return

    setLoading(true)
    setMessage(null)
    
    try {
      const updateParams: UpdateUserRequest = {
        id: userProfile.id,
        status: formData.status
      }
      
      // 只有当密码不为空时才包含密码字段
      if (formData.password.trim()) {
        updateParams.password = formData.password
      }

      const updatedProfile = await updateUser(updateParams)
      if (updatedProfile) {
        setUserProfile(updatedProfile)
        setMessage({ type: 'success', text: '用户信息更新成功' })
        setIsEditing(false)
        setFormData({ ...formData, password: '' })
      } else {
        setMessage({ type: 'error', text: '更新用户信息失败' })
      }
    } catch (error) {
      console.error('更新用户信息失败:', error)
      setMessage({ type: 'error', text: '更新用户信息失败' })
    } finally {
      setLoading(false)
    }
  }

  // 格式化时间戳
  const formatTimestamp = (timestamp: number) => {
    return new Date(timestamp * 1000).toLocaleString('zh-CN')
  }

  // 获取状态显示文本和颜色
  const getStatusInfo = (status: string) => {
    switch (status) {
      case 'active':
        return { text: '正常', color: 'bg-green-100 text-green-800' }
      case 'locked':
        return { text: '锁定', color: 'bg-red-100 text-red-800' }
      case 'inactive':
        return { text: '禁用', color: 'bg-gray-100 text-gray-800' }
      default:
        return { text: '未知', color: 'bg-gray-100 text-gray-800' }
    }
  }

  // 当弹窗打开时获取用户信息
  useEffect(() => {
    if (open) {
      fetchUserProfile()
      setIsEditing(false)
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
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <User className="w-5 h-5" />
            用户信息
          </DialogTitle>
          <DialogDescription>
            查看和编辑用户基本信息
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
          <Tabs defaultValue="profile" className="w-full">
            <TabsList className="grid w-full grid-cols-2">
              <TabsTrigger value="profile">基本信息</TabsTrigger>
              <TabsTrigger value="settings">设置</TabsTrigger>
            </TabsList>
            
            <TabsContent value="profile" className="space-y-4">
              <Card>
                <CardHeader>
                  <div className="flex items-center space-x-4">
                    <Avatar className="h-16 w-16">
                      <AvatarImage src={getAvatarUrl(userProfile.avatar_url)} />
                      <AvatarFallback className="bg-blue-100 text-blue-600 text-lg">
                        {getDisplayName(userProfile).charAt(0).toUpperCase()}
                      </AvatarFallback>
                    </Avatar>
                    <div className="flex-1">
                      <CardTitle className="text-xl">{getDisplayName(userProfile)}</CardTitle>
                      <CardDescription className="flex items-center gap-2 mt-1">
                        <Badge className={getStatusInfo(userProfile.status).color}>
                          {getStatusInfo(userProfile.status).text}
                        </Badge>
                      </CardDescription>
                    </div>
                  </div>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <Label className="text-sm font-medium text-gray-500">用户ID</Label>
                      <p className="text-sm font-mono bg-gray-50 p-2 rounded">{userProfile.id}</p>
                    </div>
                    <div>
                      <Label className="text-sm font-medium text-gray-500">用户名</Label>
                      <p className="text-sm">{userProfile.username}</p>
                    </div>
                    <div>
                      <Label className="text-sm font-medium text-gray-500">邮箱</Label>
                      <p className="text-sm">{userProfile.email}</p>
                    </div>
                    <div>
                      <Label className="text-sm font-medium text-gray-500">状态</Label>
                      <Badge className={getStatusInfo(userProfile.status).color}>
                        {getStatusInfo(userProfile.status).text}
                      </Badge>
                    </div>
                    <div>
                      <Label className="text-sm font-medium text-gray-500">创建时间</Label>
                      <p className="text-sm">{formatTimestamp(userProfile.created_at)}</p>
                    </div>
                    <div>
                      <Label className="text-sm font-medium text-gray-500">最后活跃</Label>
                      <p className="text-sm">{formatTimestamp(userProfile.last_active_at)}</p>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </TabsContent>
            
            <TabsContent value="settings" className="space-y-4">
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center justify-between">
                    编辑用户信息
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setIsEditing(!isEditing)}
                    >
                      {isEditing ? (
                        <><X className="w-4 h-4 mr-1" /> 取消</>
                      ) : (
                        <><Edit3 className="w-4 h-4 mr-1" /> 编辑</>
                      )}
                    </Button>
                  </CardTitle>
                  <CardDescription>
                    修改用户的基本设置信息
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="space-y-4">
                    <div>
                      <Label htmlFor="status">用户状态</Label>
                      <Select
                        value={formData.status}
                        onValueChange={(value: string) => 
                          setFormData({ ...formData, status: value as 'active' | 'locked' | 'inactive' })
                        }
                        disabled={!isEditing}
                      >
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="active">正常</SelectItem>
                          <SelectItem value="locked">锁定</SelectItem>
                          <SelectItem value="inactive">禁用</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                    
                    <div>
                      <Label htmlFor="password">重置密码</Label>
                      <div className="relative">
                        <Input
                          id="password"
                          type={showPassword ? 'text' : 'password'}
                          value={formData.password}
                          onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                          placeholder="留空则不修改密码"
                          disabled={!isEditing}
                        />
                        <Button
                          type="button"
                          variant="ghost"
                          size="sm"
                          className="absolute right-0 top-0 h-full px-3 py-2 hover:bg-transparent"
                          onClick={() => setShowPassword(!showPassword)}
                          disabled={!isEditing}
                        >
                          {showPassword ? (
                            <EyeOff className="h-4 w-4" />
                          ) : (
                            <Eye className="h-4 w-4" />
                          )}
                        </Button>
                      </div>
                      <p className="text-xs text-gray-500 mt-1">
                        密码长度至少6位，留空则不修改当前密码
                      </p>
                    </div>
                  </div>
                  
                  {isEditing && (
                    <div className="flex justify-end space-x-2 pt-4">
                      <Button
                        variant="outline"
                        onClick={() => {
                          setIsEditing(false)
                          setFormData({
                            password: '',
                            status: userProfile.status as 'active' | 'locked' | 'inactive'
                          })
                        }}
                      >
                        取消
                      </Button>
                      <Button
                        onClick={handleUpdateUser}
                        disabled={loading}
                      >
                        {loading ? (
                          <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                        ) : (
                          <Save className="w-4 h-4 mr-2" />
                        )}
                        保存更改
                      </Button>
                    </div>
                  )}
                </CardContent>
              </Card>
            </TabsContent>
          </Tabs>
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