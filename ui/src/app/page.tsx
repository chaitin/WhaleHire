"use client";

import { useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { PasswordInput } from "@/components/custom/password-input";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import Image from "next/image";
import { userLogin, adminLogin, userRegister } from "@/lib/api";

// 登录表单验证模式
const loginSchema = z.object({
  username: z.string()
    .min(1, "用户名不能为空")
    .max(50, "用户名不能超过50个字符")
    .regex(/^[a-zA-Z0-9_\u4e00-\u9fa5]+$/, "用户名只能包含字母、数字、下划线和中文"),
  password: z.string()
    .min(6, "密码至少6位")
    .max(128, "密码不能超过128个字符"),
});

// 注册表单验证模式
const registerSchema = z.object({
  username: z.string()
    .min(3, "用户名至少3位")
    .max(20, "用户名不能超过20个字符")
    .regex(/^[a-zA-Z0-9_\u4e00-\u9fa5]+$/, "用户名只能包含字母、数字、下划线和中文")
    .refine((val) => !/^\d+$/.test(val), "用户名不能全为数字"),
  email: z.string()
    .email("请输入有效的邮箱地址")
    .max(100, "邮箱地址不能超过100个字符"),
  password: z.string()
    .min(8, "密码至少8位")
    .max(128, "密码不能超过128个字符")
    .regex(/^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/, "密码必须包含大小写字母和数字"),
  confirmPassword: z.string().min(8, "确认密码至少8位"),
}).refine((data) => data.password === data.confirmPassword, {
  message: "两次输入的密码不一致",
  path: ["confirmPassword"],
});

type LoginFormData = z.infer<typeof loginSchema>;
type RegisterFormData = z.infer<typeof registerSchema>;

export default function AuthPage() {
  const [isLoading, setIsLoading] = useState(false);
  const [activeTab, setActiveTab] = useState("user-login");
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  // 用户登录表单
  const userLoginForm = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      username: "",
      password: "",
    },
  });

  // 管理员登录表单
  const adminLoginForm = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      username: "",
      password: "",
    },
  });

  // 注册表单
  const registerForm = useForm<RegisterFormData>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      username: "",
      email: "",
      password: "",
      confirmPassword: "",
    },
  });

  // 用户登录处理
  const handleUserLogin = async (data: LoginFormData) => {
    setIsLoading(true);
    setMessage(null);
    
    const result = await userLogin({
      username: data.username,
      password: data.password,
      source: 'browser',
    });
    
    if (result.success) {
      setMessage({ type: 'success', text: '登录成功！正在跳转...' });
      // 跳转到dashboard页面
      setTimeout(() => {
        window.location.href = '/dashboard';
      }, 1500);
      console.log('用户登录成功:', result.data);
    } else {
      setMessage({ type: 'error', text: result.message || '登录失败，请检查用户名和密码' });
    }
    
    setIsLoading(false);
  };

  // 管理员登录处理
  const handleAdminLogin = async (data: LoginFormData) => {
    setIsLoading(true);
    setMessage(null);
    
    const result = await adminLogin({
      username: data.username,
      password: data.password,
      source: 'browser',
    });
    
    if (result.success) {
      setMessage({ type: 'success', text: '管理员登录成功！正在跳转...' });
      // 处理登录成功逻辑，比如跳转到管理后台
      setTimeout(() => {
        window.location.href = result.data?.redirect_url || '/admin/dashboard';
      }, 1500);
      console.log('管理员登录成功，跳转到:', result.data?.redirect_url || '/admin/dashboard');
    } else {
      setMessage({ type: 'error', text: result.message || '登录失败，请检查管理员账号和密码' });
    }
    
    setIsLoading(false);
  };

  // 注册处理
  const handleRegister = async (data: RegisterFormData) => {
    setIsLoading(true);
    setMessage(null);
    
    const result = await userRegister({
      username: data.username,
      email: data.email,
      password: data.password,
    });
    
    if (result.success) {
      setMessage({ type: 'success', text: '注册成功！请使用新账户登录' });
      // 重置注册表单
      registerForm.reset();
      // 延迟切换到登录页面
      setTimeout(() => {
        setActiveTab('user-login');
        setMessage(null);
      }, 2000);
      console.log('用户注册成功:', result.data);
    } else {
      setMessage({ type: 'error', text: result.message || '注册失败，请检查输入信息' });
    }
    
    setIsLoading(false);
  };

  // 清除消息提示
  const clearMessage = () => {
    if (message) {
      setMessage(null);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        {/* Logo和标题 */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-16 h-16 bg-blue-600 rounded-full mb-4">
            <Image 
              src="/logo.svg" 
              alt="WhaleHire Logo" 
              width={32} 
              height={32}
              className="brightness-0 invert"
            />
          </div>
          <h1 className="text-3xl font-bold text-gray-900 mb-2">WhaleHire</h1>
          <p className="text-gray-600">AI智能招聘平台</p>
        </div>

        <Card className="shadow-xl">
          <CardHeader className="space-y-1">
            <CardTitle className="text-2xl text-center">欢迎使用</CardTitle>
            <CardDescription className="text-center">
              请选择登录方式或创建新账户
            </CardDescription>
          </CardHeader>
           <CardContent>
             {/* 消息提示 */}
             {message && (
               <div className={`mb-4 p-3 rounded-md text-sm ${
                 message.type === 'success' 
                   ? 'bg-green-50 text-green-700 border border-green-200' 
                   : 'bg-red-50 text-red-700 border border-red-200'
               }`}>
                 {message.text}
               </div>
             )}
             
             <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
              <TabsList className="grid w-full grid-cols-3">
                <TabsTrigger value="user-login">用户登录</TabsTrigger>
                <TabsTrigger value="admin-login">管理员</TabsTrigger>
                <TabsTrigger value="register">注册</TabsTrigger>
              </TabsList>
              
              {/* 用户登录 */}
              <TabsContent value="user-login" className="space-y-4">
                <Form {...userLoginForm}>
                  <form onSubmit={userLoginForm.handleSubmit(handleUserLogin)} className="space-y-4">
                    <FormField
                      control={userLoginForm.control}
                      name="username"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>用户名</FormLabel>
                          <FormControl>
                             <Input placeholder="请输入用户名" {...field} onFocus={clearMessage} />
                           </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={userLoginForm.control}
                      name="password"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>密码</FormLabel>
                          <FormControl>
                             <PasswordInput placeholder="请输入密码" {...field} onFocus={clearMessage} />
                           </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <Button type="submit" className="w-full" disabled={isLoading}>
                      {isLoading ? "登录中..." : "登录"}
                    </Button>
                  </form>
                </Form>
              </TabsContent>

              {/* 管理员登录 */}
              <TabsContent value="admin-login" className="space-y-4">
                <Form {...adminLoginForm}>
                  <form onSubmit={adminLoginForm.handleSubmit(handleAdminLogin)} className="space-y-4">
                    <FormField
                      control={adminLoginForm.control}
                      name="username"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>管理员用户名</FormLabel>
                          <FormControl>
                             <Input placeholder="请输入管理员用户名" {...field} onFocus={clearMessage} />
                           </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={adminLoginForm.control}
                      name="password"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>密码</FormLabel>
                          <FormControl>
                             <PasswordInput placeholder="请输入密码" {...field} onFocus={clearMessage} />
                           </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <Button type="submit" className="w-full bg-orange-600 hover:bg-orange-700" disabled={isLoading}>
                      {isLoading ? "登录中..." : "管理员登录"}
                    </Button>
                  </form>
                </Form>
              </TabsContent>

              {/* 注册 */}
              <TabsContent value="register" className="space-y-4">
                <Form {...registerForm}>
                  <form onSubmit={registerForm.handleSubmit(handleRegister)} className="space-y-4">
                    <FormField
                      control={registerForm.control}
                      name="username"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>用户名</FormLabel>
                          <FormControl>
                             <Input placeholder="请输入用户名" {...field} onFocus={clearMessage} />
                           </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={registerForm.control}
                      name="email"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>邮箱</FormLabel>
                          <FormControl>
                             <Input type="email" placeholder="请输入邮箱地址" {...field} onFocus={clearMessage} />
                           </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={registerForm.control}
                      name="password"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>密码</FormLabel>
                          <FormControl>
                             <PasswordInput placeholder="请输入密码" {...field} onFocus={clearMessage} />
                           </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={registerForm.control}
                      name="confirmPassword"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>确认密码</FormLabel>
                          <FormControl>
                             <PasswordInput placeholder="请再次输入密码" {...field} onFocus={clearMessage} />
                           </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <Button type="submit" className="w-full bg-green-600 hover:bg-green-700" disabled={isLoading}>
                      {isLoading ? "注册中..." : "注册账户"}
                    </Button>
                  </form>
                </Form>
              </TabsContent>
            </Tabs>
          </CardContent>
        </Card>

        {/* 底部信息 */}
        <div className="text-center mt-8 text-sm text-gray-500">
          <p>© 2024 WhaleHire. 专业的AI招聘解决方案</p>
        </div>
      </div>
    </div>
  );
}
