import React, { useState, useEffect } from 'react';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import { Eye, EyeOff, User, Lock, Github } from 'lucide-react';
import { cn } from '@/lib/utils';
import { useAuth } from '@/hooks/useAuth';
import { login as loginApi, LoginFormData, getOAuthUrl } from '@/services/auth';
import { ApiError } from '@/lib/api';

export default function LoginPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const { login, isAuthenticated } = useAuth();
  const [showPassword, setShowPassword] = useState(false);
  const [loginError, setLoginError] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [isOAuthLoading, setIsOAuthLoading] = useState(false);
  const [formData, setFormData] = useState({
    username: '',
    password: '',
  });

  // 如果已经登录，重定向到目标页面或首页
  useEffect(() => {
    console.log('登录页面 useEffect 触发，isAuthenticated:', isAuthenticated);
    if (isAuthenticated) {
      const from = location.state?.from?.pathname || '/resume-management';
      console.log('用户已认证，准备跳转到:', from);
      navigate(from, { replace: true });
    }
  }, [isAuthenticated, navigate, location]);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: value,
    }));
    // 清除错误信息
    if (loginError) {
      setLoginError('');
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setLoginError('');

    // 输入验证
    if (!formData.username.trim()) {
      setLoginError('请输入账号');
      setIsLoading(false);
      return;
    }

    if (!formData.password.trim()) {
      setLoginError('请输入密码');
      setIsLoading(false);
      return;
    }

    try {
      const loginData: LoginFormData = {
        username: formData.username.trim(),
        password: formData.password,
      };

      console.log('开始登录，发送数据:', loginData);

      // 调用登录API
      const response = await loginApi(loginData);
      console.log('登录API响应:', response);

      // 更新认证状态
      login(response.user);
      console.log('已更新认证状态，用户信息:', response.user);

      // 跳转到目标页面
      const from = location.state?.from?.pathname || '/resume-management';
      console.log('准备跳转到:', from);
      navigate(from, { replace: true });
    } catch (error) {
      // 处理错误，显示错误信息但不跳转页面
      if (error instanceof ApiError) {
        // 根据不同的错误代码显示不同的错误信息
        if (error.code === 401) {
          setLoginError('您输入的信息有误，请重试');
        } else if (error.code === 400) {
          setLoginError('您输入的信息有误，请重试');
        } else if (error.code >= 500) {
          setLoginError('服务器暂时无法响应，请稍后重试');
        } else {
          setLoginError('您输入的信息有误，请重试');
        }
      } else {
        setLoginError('您输入的信息有误，请重试');
      }
      console.error('登录失败:', error);
      // 确保不跳转页面，只显示错误信息
    } finally {
      setIsLoading(false);
    }
  };

  const handleOAuthLogin = async () => {
    setIsOAuthLoading(true);
    setLoginError('');

    try {
      console.log('开始OAuth登录流程...');

      // 调用后端API获取认证地址
      const authUrl = await getOAuthUrl();
      console.log('获取到认证地址:', authUrl);

      // 跳转到认证页面
      window.location.assign(authUrl);
    } catch (error) {
      // 处理OAuth错误，显示错误信息但不跳转页面
      if (error instanceof ApiError) {
        if (error.code >= 500) {
          setLoginError('服务器暂时无法响应，请稍后重试');
        } else {
          setLoginError('您输入的信息有误，请重试');
        }
      } else {
        setLoginError('您输入的信息有误，请重试');
      }
      console.error('OAuth登录失败:', error);
      setIsOAuthLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex">
      {/* Left Section - Hero */}
      <section className="flex-1 relative overflow-hidden">
        {/* Background Gradient */}
        <div className="absolute inset-0 bg-gradient-to-br from-blue-50 via-blue-50 to-purple-50"></div>

        {/* Content */}
        <div className="relative z-10 flex flex-col justify-center h-full px-16">
          <div className="max-w-lg space-y-6">
            {/* Logo/Title */}
            <h1 className="text-5xl font-bold text-gray-900 mb-2 text-left">
              WhaleHire
            </h1>

            {/* Subtitle */}
            <h2 className="text-3xl font-semibold text-gray-900 text-left">
              智能招聘
              <span
                className="ml-3 bg-clip-text text-transparent"
                style={{
                  backgroundImage:
                    'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                }}
              >
                AI驱动未来
              </span>
            </h2>

            {/* Description */}
            <p className="text-gray-600 text-lg leading-relaxed text-left">
              让AI助手帮助您快速撰写专业JD，智能筛选简历，打造高效招聘流程，轻松找到理想人才。
            </p>

            {/* Tech Recruitment Illustration */}
            <div className="mt-12 relative">
              <div className="bg-white/80 backdrop-blur-sm rounded-xl p-8 shadow-xl border border-white/20">
                <div className="w-full h-64 bg-gradient-to-br from-blue-50 via-indigo-50 to-purple-50 rounded-lg relative overflow-hidden">
                  {/* Tech Background Pattern */}
                  <div className="absolute inset-0 opacity-10">
                    <div className="absolute top-4 left-4 w-8 h-8 border-2 border-blue-400 rounded rotate-45"></div>
                    <div
                      className="absolute top-8 right-8 w-6 h-6 rounded-full"
                      style={{
                        background:
                          'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                      }}
                    ></div>
                    <div className="absolute bottom-8 left-8 w-4 h-4 bg-purple-400 rounded"></div>
                    <div className="absolute bottom-4 right-12 w-10 h-10 border-2 border-indigo-400 rounded-full"></div>
                  </div>

                  {/* Main Content */}
                  <div className="relative z-10 flex items-center justify-center h-full">
                    <div className="text-center space-y-4">
                      {/* AI Brain Icon */}
                      <div className="relative mx-auto w-20 h-20">
                        <div className="w-20 h-20 bg-gradient-to-r from-blue-500 to-purple-500 rounded-full flex items-center justify-center shadow-lg">
                          <svg
                            className="w-10 h-10 text-white"
                            fill="currentColor"
                            viewBox="0 0 24 24"
                          >
                            <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z" />
                          </svg>
                        </div>
                        {/* Floating particles */}
                        <div
                          className="absolute -top-2 -right-2 w-4 h-4 rounded-full animate-pulse"
                          style={{
                            background:
                              'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                          }}
                        ></div>
                        <div className="absolute -bottom-2 -left-2 w-3 h-3 bg-blue-400 rounded-full animate-pulse delay-300"></div>
                      </div>

                      {/* Text */}
                      <div className="space-y-2">
                        <h3 className="text-gray-800 font-bold text-lg">
                          AI智能筛选
                        </h3>
                        <p className="text-gray-600 text-sm">
                          精准匹配 • 高效招聘 • 智能推荐
                        </p>
                      </div>

                      {/* Tech Elements */}
                      <div className="flex justify-center space-x-4 mt-4">
                        <div className="w-8 h-8 bg-blue-100 rounded-lg flex items-center justify-center">
                          <div className="w-4 h-4 bg-blue-500 rounded"></div>
                        </div>
                        <div
                          className="w-8 h-8 rounded-lg flex items-center justify-center"
                          style={{
                            background:
                              'linear-gradient(135deg, rgba(123, 184, 255, 0.2) 0%, rgba(63, 54, 99, 0.2) 100%)',
                          }}
                        >
                          <div
                            className="w-4 h-4 rounded-full"
                            style={{
                              background:
                                'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                            }}
                          ></div>
                        </div>
                        <div className="w-8 h-8 bg-purple-100 rounded-lg flex items-center justify-center">
                          <div className="w-4 h-4 bg-purple-500 rounded rotate-45"></div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Right Section - Login Form */}
      <section className="flex-1 bg-white flex items-center justify-center px-8">
        <div className="w-full max-w-md">
          <div className="bg-white rounded-2xl shadow-lg border border-gray-100 p-8">
            {/* Header */}
            <div className="text-center mb-8">
              <h2 className="text-2xl font-semibold text-gray-900">欢迎登录</h2>
            </div>

            {/* Error Message */}
            {loginError && (
              <div className="bg-red-50 border border-red-200 text-red-600 px-4 py-3 rounded-lg text-sm mb-4">
                <div className="flex items-center">
                  <svg
                    className="w-4 h-4 mr-2 flex-shrink-0"
                    fill="currentColor"
                    viewBox="0 0 20 20"
                  >
                    <path
                      fillRule="evenodd"
                      d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                      clipRule="evenodd"
                    />
                  </svg>
                  <span className="font-medium">{loginError}</span>
                </div>
              </div>
            )}

            {/* Login Form */}
            <form onSubmit={handleSubmit} className="space-y-6">
              {/* Username Field */}
              <div className="space-y-2">
                <label
                  htmlFor="username"
                  className="block text-sm font-medium text-gray-900"
                >
                  账号
                </label>
                <div className="relative">
                  <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                    <User className="h-4 w-4 text-gray-400" />
                  </div>
                  <input
                    id="username"
                    name="username"
                    type="text"
                    value={formData.username}
                    onChange={handleInputChange}
                    className="block w-full pl-10 pr-3 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-[#7bb8ff] focus:border-[#7bb8ff] outline-none transition-colors"
                    placeholder="请输入账号"
                  />
                </div>
              </div>

              {/* Password Field */}
              <div className="space-y-2">
                <label
                  htmlFor="password"
                  className="block text-sm font-medium text-gray-900"
                >
                  密码
                </label>
                <div className="relative">
                  <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                    <Lock className="h-4 w-4 text-gray-400" />
                  </div>
                  <input
                    id="password"
                    name="password"
                    type={showPassword ? 'text' : 'password'}
                    value={formData.password}
                    onChange={handleInputChange}
                    className="block w-full pl-10 pr-12 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-[#7bb8ff] focus:border-[#7bb8ff] outline-none transition-colors"
                    placeholder="请输入密码"
                  />
                  <button
                    type="button"
                    className="absolute inset-y-0 right-0 pr-3 flex items-center"
                    onClick={() => setShowPassword(!showPassword)}
                  >
                    {showPassword ? (
                      <Eye className="h-4 w-4 text-gray-400 hover:text-gray-600" />
                    ) : (
                      <EyeOff className="h-4 w-4 text-gray-400 hover:text-gray-600" />
                    )}
                  </button>
                </div>
              </div>

              {/* Forgot Password */}
              <div className="flex justify-end">
                <button
                  type="button"
                  disabled
                  className="text-sm text-gray-400 cursor-not-allowed"
                >
                  忘记密码？
                </button>
              </div>

              {/* Login Button */}
              <button
                type="submit"
                disabled={isLoading}
                className={cn(
                  'w-full font-medium py-3 px-4 rounded-lg transition-colors focus:outline-none focus:ring-2 focus:ring-[#7bb8ff] focus:ring-offset-2',
                  isLoading
                    ? 'bg-gray-400 cursor-not-allowed'
                    : 'hover:bg-[#5aa3e6] text-white'
                )}
                style={
                  !isLoading
                    ? {
                        background:
                          'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                      }
                    : undefined
                }
              >
                {isLoading ? (
                  <div className="flex items-center justify-center">
                    <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin mr-2"></div>
                    登录中...
                  </div>
                ) : (
                  '登录'
                )}
              </button>

              {/* Divider */}
              <div className="relative">
                <div className="absolute inset-0 flex items-center">
                  <div className="w-full border-t border-gray-200" />
                </div>
                <div className="relative flex justify-center text-sm">
                  <span className="px-4 bg-white text-gray-500">
                    其他登录方式
                  </span>
                </div>
              </div>

              {/* OAuth Login */}
              <div className="flex justify-center">
                <button
                  type="button"
                  disabled={isOAuthLoading || isLoading}
                  className={cn(
                    'w-12 h-12 border border-gray-200 rounded-full flex items-center justify-center transition-colors',
                    isOAuthLoading || isLoading
                      ? 'cursor-not-allowed bg-gray-100 border-gray-100'
                      : 'hover:border-gray-300 hover:bg-gray-50'
                  )}
                  onClick={handleOAuthLogin}
                >
                  {isOAuthLoading ? (
                    <div className="w-5 h-5 border-2 border-gray-400 border-t-transparent rounded-full animate-spin" />
                  ) : (
                    <Github className="w-5 h-5 text-gray-600" />
                  )}
                </button>
              </div>

              {/* Sign Up Link */}
              <div className="text-center">
                <p className="text-gray-600">
                  还没有账号？
                  <Link
                    to="/register"
                    className="font-medium ml-1 bg-clip-text text-transparent hover:opacity-80"
                    style={{
                      backgroundImage:
                        'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                    }}
                  >
                    立即注册
                  </Link>
                </p>
              </div>
            </form>
          </div>
        </div>
      </section>
    </div>
  );
}
