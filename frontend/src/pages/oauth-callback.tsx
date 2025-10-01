import { useEffect, useState, useRef } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useAuth } from '@/hooks/useAuth';
import { handleOAuthCallback} from '@/services/auth';
import { ApiError } from '@/lib/api';

export default function OAuthCallbackPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { login } = useAuth();
  const [error, setError] = useState<string>('');
  const [isProcessing, setIsProcessing] = useState(true);
  
  // 使用 ref 防止 React.StrictMode 导致的重复执行
  const hasProcessedRef = useRef(false);

  useEffect(() => {
    // 如果已经处理过，直接返回
    if (hasProcessedRef.current) {
      console.log('OAuth回调已处理过，跳过重复执行');
      return;
    }

    const processCallback = async () => {
      try {
        // 标记为已处理，防止重复执行
        hasProcessedRef.current = true;
        
        // 从URL参数中获取code和state
        const code = searchParams.get('code');
        const state = searchParams.get('state');
        const errorParam = searchParams.get('error');

        // 检查是否有错误参数
        if (errorParam) {
          setError('OAuth认证被取消或失败');
          setIsProcessing(false);
          return;
        }

        // 检查必需的参数
        if (!code || !state) {
          setError('OAuth回调参数缺失');
          setIsProcessing(false);
          return;
        }

        console.log('处理OAuth回调，code:', code, 'state:', state);

        // 调用后端API处理回调 - 后端只返回 redirect_url
        const response = await handleOAuthCallback(code, state);
        console.log('OAuth回调处理成功:', response);

        // 更新认证状态
        login(response.user);
        console.log('已更新认证状态，用户信息:', response.user);

        // 跳转到简历管理页面
        navigate('/resume-management', { replace: true });

      } catch (error) {
        console.error('OAuth回调处理失败:', error);
        // 处理失败时重置标记，允许重试
        hasProcessedRef.current = false;
        
        if (error instanceof ApiError) {
          setError(error.message);
        } else {
          setError('OAuth登录失败，请稍后重试');
        }
        setIsProcessing(false);
      }
    };

    processCallback();
  }, [searchParams, login, navigate]);

  // 如果处理失败，5秒后自动跳转到登录页
  useEffect(() => {
    if (error) {
      const timer = setTimeout(() => {
        navigate('/login', { replace: true });
      }, 5000);

      return () => clearTimeout(timer);
    }
  }, [error, navigate]);

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-8 text-center">
        {isProcessing ? (
          <>
            {/* 加载状态 */}
            <div className="mb-6">
              <div className="w-16 h-16 border-4 border-green-500 border-t-transparent rounded-full animate-spin mx-auto"></div>
            </div>
            <h2 className="text-xl font-semibold text-gray-900 mb-2">
              正在处理登录...
            </h2>
            <p className="text-gray-600">
              请稍候，我们正在验证您的身份信息
            </p>
          </>
        ) : (
          <>
            {/* 错误状态 */}
            <div className="mb-6">
              <div className="w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mx-auto">
                <svg className="w-8 h-8 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </div>
            </div>
            <h2 className="text-xl font-semibold text-gray-900 mb-2">
              登录失败
            </h2>
            <p className="text-gray-600 mb-6">
              {error}
            </p>
            <div className="space-y-3">
              <button
                onClick={() => navigate('/login', { replace: true })}
                className="w-full bg-green-500 hover:bg-green-600 text-white font-medium py-2 px-4 rounded-lg transition-colors"
              >
                返回登录页面
              </button>
              <p className="text-sm text-gray-500">
                5秒后将自动跳转到登录页面
              </p>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
