"use client";

import { useEffect, useState, Suspense } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { oauthCallback } from "@/lib/api";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Loader2, CheckCircle, XCircle } from "lucide-react";

function OAuthCallbackContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [status, setStatus] = useState<'loading' | 'success' | 'error'>('loading');
  const [message, setMessage] = useState<string>('正在处理OAuth登录...');

  useEffect(() => {
    const handleCallback = async () => {
      try {
        const code = searchParams.get('code');
        const state = searchParams.get('state');
        
        if (!code || !state) {
          setStatus('error');
          setMessage('OAuth回调参数缺失');
          return;
        }

        console.log('OAuth回调处理开始:', { code, state });
        
        const result = await oauthCallback({ code, state });
        
        if (result.success) {
          setStatus('success');
          setMessage('OAuth登录成功！正在跳转...');
          
          // 延迟跳转，让用户看到成功消息
          setTimeout(() => {
            if (result.data?.redirect_url) {
              window.location.href = result.data.redirect_url;
            } else {
              router.push('/dashboard');
            }
          }, 1500);
        } else {
          setStatus('error');
          setMessage(result.message || 'OAuth登录失败');
          
          // 3秒后跳转回登录页面
          setTimeout(() => {
            router.push('/');
          }, 3000);
        }
      } catch (error) {
        console.error('OAuth回调处理错误:', error);
        setStatus('error');
        setMessage('登录处理失败，请重试');
        
        // 3秒后跳转回登录页面
        setTimeout(() => {
          router.push('/');
        }, 3000);
      }
    };

    handleCallback();
  }, [searchParams, router]);

  const getIcon = () => {
    switch (status) {
      case 'loading':
        return <Loader2 className="h-8 w-8 animate-spin text-blue-500" />;
      case 'success':
        return <CheckCircle className="h-8 w-8 text-green-500" />;
      case 'error':
        return <XCircle className="h-8 w-8 text-red-500" />;
    }
  };

  const getStatusColor = () => {
    switch (status) {
      case 'loading':
        return 'text-blue-600';
      case 'success':
        return 'text-green-600';
      case 'error':
        return 'text-red-600';
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <div className="flex justify-center mb-4">
            {getIcon()}
          </div>
          <CardTitle className={getStatusColor()}>
            {status === 'loading' && 'OAuth登录处理中'}
            {status === 'success' && 'OAuth登录成功'}
            {status === 'error' && 'OAuth登录失败'}
          </CardTitle>
          <CardDescription>
            {message}
          </CardDescription>
        </CardHeader>
        <CardContent className="text-center text-sm text-gray-500">
          {status === 'loading' && '请稍候，正在验证您的身份...'}
          {status === 'success' && '即将跳转到系统主页面'}
          {status === 'error' && '将在3秒后返回登录页面'}
        </CardContent>
      </Card>
    </div>
  );
}

export default function OAuthCallbackPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <Card className="w-full max-w-md">
          <CardHeader className="text-center">
            <div className="flex justify-center mb-4">
              <Loader2 className="h-8 w-8 animate-spin text-blue-500" />
            </div>
            <CardTitle className="text-blue-600">
              OAuth登录处理中
            </CardTitle>
            <CardDescription>
              正在处理OAuth登录...
            </CardDescription>
          </CardHeader>
          <CardContent className="text-center text-sm text-gray-500">
            请稍候，正在验证您的身份...
          </CardContent>
        </Card>
      </div>
    }>
      <OAuthCallbackContent />
    </Suspense>
  );
}