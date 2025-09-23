"use client";

import { useEffect, useState, Suspense } from "react";
import { useSearchParams } from "next/navigation";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Loader2 } from "lucide-react";
import { oauthCallback } from "@/lib/api";

function AuthCallbackContent() {
  const [status, setStatus] = useState<'loading' | 'success' | 'error'>('loading');
  const [message, setMessage] = useState('');
  const searchParams = useSearchParams();

  useEffect(() => {
    const handleCallback = async () => {
      const code = searchParams.get('code');
      const state = searchParams.get('state');
      const error = searchParams.get('error');

      if (error) {
        setStatus('error');
        setMessage('OAuth授权被取消或失败');
        return;
      }

      if (!code || !state) {
        setStatus('error');
        setMessage('缺少必要的授权参数');
        return;
      }

      try {
        const result = await oauthCallback({ code, state });
        
        if (result.success) {
          setStatus('success');
          setMessage('登录成功！正在跳转...');
          
          // 延迟跳转到dashboard
          setTimeout(() => {
            window.location.href = '/dashboard';
          }, 2000);
        } else {
          setStatus('error');
          setMessage(result.message || 'OAuth登录失败');
        }
      } catch (error) {
        console.error('OAuth回调处理错误:', error);
        setStatus('error');
        setMessage('登录处理失败，请重试');
      }
    };

    handleCallback();
  }, [searchParams]);

  const handleReturnToLogin = () => {
    window.location.href = '/';
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        <Card className="shadow-xl">
          <CardHeader className="text-center">
            <CardTitle className="text-2xl">
              {status === 'loading' && '正在处理登录...'}
              {status === 'success' && '登录成功'}
              {status === 'error' && '登录失败'}
            </CardTitle>
            <CardDescription>
              {status === 'loading' && '请稍候，正在验证您的身份'}
              {status === 'success' && '欢迎使用 WhaleHire'}
              {status === 'error' && '登录过程中出现了问题'}
            </CardDescription>
          </CardHeader>
          <CardContent className="text-center space-y-4">
            {status === 'loading' && (
              <div className="flex justify-center">
                <Loader2 className="h-8 w-8 animate-spin text-blue-600" />
              </div>
            )}
            
            {status === 'success' && (
              <div className="space-y-4">
                <div className="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mx-auto">
                  <svg className="w-8 h-8 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                  </svg>
                </div>
                <p className="text-green-700">{message}</p>
              </div>
            )}
            
            {status === 'error' && (
              <div className="space-y-4">
                <div className="w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mx-auto">
                  <svg className="w-8 h-8 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </div>
                <p className="text-red-700">{message}</p>
                <button
                  onClick={handleReturnToLogin}
                  className="w-full bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded-md transition-colors"
                >
                  返回登录页面
                </button>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

export default function AuthCallbackPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4">
        <div className="w-full max-w-md">
          <Card className="shadow-xl">
            <CardHeader className="text-center">
              <CardTitle className="text-2xl">正在处理登录...</CardTitle>
              <CardDescription>请稍候，正在验证您的身份</CardDescription>
            </CardHeader>
            <CardContent className="text-center space-y-4">
              <div className="flex justify-center">
                <Loader2 className="h-8 w-8 animate-spin text-blue-600" />
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    }>
      <AuthCallbackContent />
    </Suspense>
  );
}