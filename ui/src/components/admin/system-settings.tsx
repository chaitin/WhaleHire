'use client'

import React, { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { Badge } from '@/components/ui/badge'
import { Loader2, Save, RefreshCw, CheckCircle, AlertCircle } from 'lucide-react'
import { getSetting, updateSetting } from '@/lib/api/setting'
import type { Setting, UpdateSettingRequest } from '@/lib/api/types'

interface SystemSettingsProps {
  className?: string
}

interface MessageState {
  type: 'success' | 'error'
  text: string
}

export function SystemSettings({ className }: SystemSettingsProps) {
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [formData, setFormData] = useState<Partial<Setting>>({})
  const [message, setMessage] = useState<MessageState | null>(null)

  // 加载系统设置
  const loadSettings = async () => {
    try {
      setLoading(true)
      const response = await getSetting()
      
      if (response.success && response.data) {
        setFormData({
          enable_sso: response.data.enable_sso,
          force_two_factor_auth: response.data.force_two_factor_auth,
          disable_password_login: response.data.disable_password_login,
          enable_auto_login: response.data.enable_auto_login,
          base_url: response.data.base_url,
          dingtalk_oauth: response.data.dingtalk_oauth,
          custom_oauth: response.data.custom_oauth
        })
        setMessage({
          type: 'success',
          text: '系统设置加载成功'
        })
      } else {
        setMessage({
          type: 'error',
          text: response.message || '加载系统设置失败'
        })
      }
    } catch (_error) {
      const errorMessage = _error instanceof Error ? _error.message : '加载系统设置时发生错误'
      setMessage({ type: 'error', text: errorMessage })
    } finally {
      setLoading(false)
    }
  }

  // 保存系统设置
  const handleSave = async () => {
    try {
      setSaving(true)
      const response = await updateSetting(formData as UpdateSettingRequest)
      
      if (response.success) {
        setMessage({
          type: 'success',
          text: '系统设置保存成功'
        })
        // 重新加载设置以确保数据同步
        await loadSettings()
      } else {
        setMessage({
          type: 'error',
          text: response.message || '保存系统设置失败'
        })
      }
    } catch (_error) {
      setMessage({
        type: 'error',
        text: '保存系统设置时发生错误'
      })
    } finally {
      setSaving(false)
    }
  }

  // 更新表单数据
  const updateFormData = (path: string, value: unknown) => {
    setFormData(prev => {
      const keys = path.split('.')
      const newData = { ...prev }
      let current: Record<string, unknown> = newData
      
      for (let i = 0; i < keys.length - 1; i++) {
        if (!current[keys[i]]) {
          current[keys[i]] = {}
        }
        current = current[keys[i]] as Record<string, unknown>
      }
      
      current[keys[keys.length - 1]] = value
      return newData
    })
  }

  // 清除消息
  const clearMessage = () => {
    setTimeout(() => {
      setMessage(null)
    }, 3000)
  }

  useEffect(() => {
    loadSettings()
  }, [])

  useEffect(() => {
    if (message) {
      clearMessage()
    }
  }, [message])

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <Loader2 className="h-8 w-8 animate-spin" />
        <span className="ml-2">加载系统设置中...</span>
      </div>
    )
  }

  return (
    <div className={`space-y-6 ${className}`}>
      {/* 页面标题 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">系统设置</h1>
          <p className="text-muted-foreground">配置系统的基础设置和认证选项</p>
        </div>
        <div className="flex items-center space-x-2">
          <Button
            variant="outline"
            onClick={loadSettings}
            disabled={loading}
          >
            <RefreshCw className="h-4 w-4 mr-2" />
            刷新
          </Button>
          <Button
            onClick={handleSave}
            disabled={saving}
          >
            {saving ? (
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
            ) : (
              <Save className="h-4 w-4 mr-2" />
            )}
            保存设置
          </Button>
        </div>
      </div>

      {/* 消息提示 */}
      {message && (
        <div className={`p-4 rounded-lg border ${
          message.type === 'success' 
            ? 'bg-green-50 border-green-200 text-green-800' 
            : 'bg-red-50 border-red-200 text-red-800'
        }`}>
          <div className="flex items-center">
            {message.type === 'success' ? (
              <CheckCircle className="h-4 w-4 mr-2" />
            ) : (
              <AlertCircle className="h-4 w-4 mr-2" />
            )}
            {message.text}
          </div>
        </div>
      )}

      {/* 基础设置 */}
      <Card>
        <CardHeader>
          <CardTitle>基础设置</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="base_url">系统基础URL</Label>
              <Input
                id="base_url"
                value={formData.base_url || ''}
                onChange={(e) => updateFormData('base_url', e.target.value)}
                placeholder="https://example.com"
              />
            </div>
          </div>
        </CardContent>
      </Card>

      {/* 安全设置 */}
      <Card>
        <CardHeader>
          <CardTitle>安全设置</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label>启用SSO单点登录</Label>
              <p className="text-sm text-muted-foreground">
                允许用户通过第三方身份提供商登录
              </p>
            </div>
            <Switch
              checked={formData.enable_sso || false}
              onCheckedChange={(checked) => updateFormData('enable_sso', checked)}
            />
          </div>
          
          <Separator />
          
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label>强制双因子认证</Label>
              <p className="text-sm text-muted-foreground">
                要求所有用户启用双因子认证
              </p>
            </div>
            <Switch
              checked={formData.force_two_factor_auth || false}
              onCheckedChange={(checked) => updateFormData('force_two_factor_auth', checked)}
            />
          </div>

          <Separator />
          
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label>禁用密码登录</Label>
              <p className="text-sm text-muted-foreground">
                禁用传统的用户名密码登录方式
              </p>
            </div>
            <Switch
              checked={formData.disable_password_login || false}
              onCheckedChange={(checked) => updateFormData('disable_password_login', checked)}
            />
          </div>

          <Separator />
          
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label>启用自动登录</Label>
              <p className="text-sm text-muted-foreground">
                记住用户登录状态，自动登录
              </p>
            </div>
            <Switch
              checked={formData.enable_auto_login || false}
              onCheckedChange={(checked) => updateFormData('enable_auto_login', checked)}
            />
          </div>
        </CardContent>
      </Card>

      {/* 钉钉OAuth设置 */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>钉钉OAuth设置</CardTitle>
            <Badge variant={formData.dingtalk_oauth?.enable ? "default" : "secondary"}>
              {formData.dingtalk_oauth?.enable ? "已启用" : "已禁用"}
            </Badge>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label>启用钉钉OAuth</Label>
              <p className="text-sm text-muted-foreground">
                允许用户通过钉钉账号登录
              </p>
            </div>
            <Switch
              checked={formData.dingtalk_oauth?.enable || false}
              onCheckedChange={(checked) => updateFormData('dingtalk_oauth.enable', checked)}
            />
          </div>
          
          {formData.dingtalk_oauth?.enable && (
            <>
              <Separator />
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="dingtalk_client_id">Client ID</Label>
                  <Input
                    id="dingtalk_client_id"
                    value={formData.dingtalk_oauth?.client_id || ''}
                    onChange={(e) => updateFormData('dingtalk_oauth.client_id', e.target.value)}
                    placeholder="钉钉应用的Client ID"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="dingtalk_client_secret">Client Secret</Label>
                  <Input
                    id="dingtalk_client_secret"
                    type="password"
                    value={formData.dingtalk_oauth?.client_secret || ''}
                    onChange={(e) => updateFormData('dingtalk_oauth.client_secret', e.target.value)}
                    placeholder="钉钉应用的Client Secret"
                  />
                </div>
              </div>
            </>
          )}
        </CardContent>
      </Card>

      {/* 自定义OAuth设置 */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>自定义OAuth设置</CardTitle>
            <Badge variant={formData.custom_oauth?.enable ? "default" : "secondary"}>
              {formData.custom_oauth?.enable ? "已启用" : "已禁用"}
            </Badge>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label>启用自定义OAuth</Label>
              <p className="text-sm text-muted-foreground">
                配置自定义的OAuth2.0身份提供商
              </p>
            </div>
            <Switch
              checked={formData.custom_oauth?.enable || false}
              onCheckedChange={(checked) => updateFormData('custom_oauth.enable', checked)}
            />
          </div>
          
          {formData.custom_oauth?.enable && (
            <>
              <Separator />
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="custom_client_id">Client ID</Label>
                  <Input
                    id="custom_client_id"
                    value={formData.custom_oauth?.client_id || ''}
                    onChange={(e) => updateFormData('custom_oauth.client_id', e.target.value)}
                    placeholder="OAuth应用的Client ID"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="custom_client_secret">Client Secret</Label>
                  <Input
                    id="custom_client_secret"
                    type="password"
                    value={formData.custom_oauth?.client_secret || ''}
                    onChange={(e) => updateFormData('custom_oauth.client_secret', e.target.value)}
                    placeholder="OAuth应用的Client Secret"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="authorize_url">授权URL</Label>
                  <Input
                    id="authorize_url"
                    value={formData.custom_oauth?.authorize_url || ''}
                    onChange={(e) => updateFormData('custom_oauth.authorize_url', e.target.value)}
                    placeholder="https://example.com/oauth/authorize"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="access_token_url">Token URL</Label>
                  <Input
                    id="access_token_url"
                    value={formData.custom_oauth?.access_token_url || ''}
                    onChange={(e) => updateFormData('custom_oauth.access_token_url', e.target.value)}
                    placeholder="https://example.com/oauth/token"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="userinfo_url">用户信息URL</Label>
                  <Input
                    id="userinfo_url"
                    value={formData.custom_oauth?.userinfo_url || ''}
                    onChange={(e) => updateFormData('custom_oauth.userinfo_url', e.target.value)}
                    placeholder="https://example.com/api/user"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="scopes">权限范围</Label>
                  <Input
                    id="scopes"
                    value={formData.custom_oauth?.scopes?.join(' ') || ''}
                    onChange={(e) => updateFormData('custom_oauth.scopes', e.target.value.split(' '))}
                    placeholder="read write profile"
                  />
                </div>
              </div>
              
              <Separator />
              
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="id_field">用户ID字段</Label>
                  <Input
                    id="id_field"
                    value={formData.custom_oauth?.id_field || ''}
                    onChange={(e) => updateFormData('custom_oauth.id_field', e.target.value)}
                    placeholder="id"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="name_field">用户名字段</Label>
                  <Input
                    id="name_field"
                    value={formData.custom_oauth?.name_field || ''}
                    onChange={(e) => updateFormData('custom_oauth.name_field', e.target.value)}
                    placeholder="name"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="email_field">邮箱字段</Label>
                  <Input
                    id="email_field"
                    value={formData.custom_oauth?.email_field || ''}
                    onChange={(e) => updateFormData('custom_oauth.email_field', e.target.value)}
                    placeholder="email"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="avatar_field">头像字段</Label>
                  <Input
                    id="avatar_field"
                    value={formData.custom_oauth?.avatar_field || ''}
                    onChange={(e) => updateFormData('custom_oauth.avatar_field', e.target.value)}
                    placeholder="avatar"
                  />
                </div>
              </div>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  )
}