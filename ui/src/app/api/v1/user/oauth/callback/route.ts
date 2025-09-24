import { NextRequest, NextResponse } from 'next/server';

/**
 * OAuth 回调处理
 * 代理转发到后端 OAuth 回调接口
 */
export async function GET(request: NextRequest) {
  try {
    // 获取查询参数
    const { searchParams } = new URL(request.url);
    
    // 构建后端 API 地址，透传所有查询参数
    const backendHost = process.env.BACKEND_HOST || 'localhost';
    const backendUrl = `http://${backendHost}:8888/api/v1/user/oauth/callback?${searchParams.toString()}`;
    
    // 转发请求到后端
    const response = await fetch(backendUrl, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        // 转发原始请求的认证信息
        'Cookie': request.headers.get('cookie') || '',
        'Authorization': request.headers.get('authorization') || '',
        'User-Agent': request.headers.get('user-agent') || '',
        // 转发客户端 IP
        'X-Forwarded-For': request.headers.get('x-forwarded-for') || request.headers.get('x-real-ip') || '',
        'X-Real-IP': request.headers.get('x-real-ip') || '',
      },
    });

    // 检查后端是否返回重定向
    if (response.status === 302 || response.status === 301) {
      const location = response.headers.get('location');
      if (location) {
        // 返回重定向响应
        return NextResponse.redirect(location, response.status);
      }
    }
    
    // 检查响应状态
    if (!response.ok) {
      const errorText = await response.text();
      console.error('Backend OAuth callback failed:', response.status, errorText);
      
      return NextResponse.json(
        { 
          code: response.status,
          message: `OAuth callback failed: ${response.statusText}`,
          data: null 
        },
        { status: response.status }
      );
    }

    // 处理成功响应
    const contentType = response.headers.get('content-type');
    
    if (contentType && contentType.includes('application/json')) {
      // JSON 响应
      const data = await response.json();
      return NextResponse.json(data, {
        status: response.status,
        headers: {
          'Set-Cookie': response.headers.get('set-cookie') || '',
        },
      });
    } else {
      // 其他类型响应（如 HTML）
      const data = await response.text();
      return new NextResponse(data, {
        status: response.status,
        headers: {
          'Content-Type': contentType || 'text/plain',
          'Set-Cookie': response.headers.get('set-cookie') || '',
        },
      });
    }
    
  } catch (error) {
    console.error('OAuth callback proxy error:', error);
    return NextResponse.json(
      { 
        code: 500,
        message: 'Internal server error during OAuth callback',
        data: null 
      },
      { status: 500 }
    );
  }
}

/**
 * 处理 CORS 预检请求
 */
export async function OPTIONS() {
  return new NextResponse(null, {
    status: 200,
    headers: {
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Methods': 'GET, OPTIONS',
      'Access-Control-Allow-Headers': 'Content-Type, Authorization, Cookie',
      'Access-Control-Max-Age': '86400',
    },
  });
}