import { NextRequest, NextResponse } from 'next/server';

/**
 * 透明代理流式响应处理
 * 避免 Next.js rewrites 机制对 SSE 流的缓冲问题
 */
export async function POST(request: NextRequest) {
  try {
    // 获取请求体
    const body = await request.text();
    
    // 构建后端 API 地址
    const backendUrl = 'http://localhost:8888/api/v1/general-agent/generate-stream';
    
    // 转发请求到后端
    const response = await fetch(backendUrl, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'text/event-stream',
        'Cache-Control': 'no-cache',
        // 转发原始请求的认证信息
        'Cookie': request.headers.get('cookie') || '',
        'Authorization': request.headers.get('authorization') || '',
      },
      body: body,
    });

    // 检查响应状态
    if (!response.ok) {
      return NextResponse.json(
        { error: `Backend request failed: ${response.status} ${response.statusText}` },
        { status: response.status }
      );
    }

    // 检查是否为流式响应
    if (!response.body) {
      return NextResponse.json(
        { error: 'No response body from backend' },
        { status: 500 }
      );
    }

    // 创建流式响应
    const stream = new ReadableStream({
      start(controller) {
        const reader = response.body!.getReader();
        
        function pump(): Promise<void> {
          return reader.read().then(({ done, value }) => {
            if (done) {
              controller.close();
              return;
            }
            
            // 直接转发数据块，不进行任何缓冲
            controller.enqueue(value);
            return pump();
          }).catch((error) => {
            console.error('Stream error:', error);
            controller.error(error);
          });
        }
        
        return pump();
      },
    });

    // 返回流式响应，设置正确的头部
    return new NextResponse(stream, {
      status: 200,
      headers: {
        'Content-Type': 'text/event-stream',
        'Cache-Control': 'no-cache, no-store, must-revalidate',
        'Connection': 'keep-alive',
        'Access-Control-Allow-Origin': '*',
        'Access-Control-Allow-Methods': 'POST, OPTIONS',
        'Access-Control-Allow-Headers': 'Content-Type, Authorization, Cookie',
      },
    });
  } catch (error) {
    console.error('Proxy error:', error);
    return NextResponse.json(
      { error: 'Internal proxy error' },
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
      'Access-Control-Allow-Methods': 'POST, OPTIONS',
      'Access-Control-Allow-Headers': 'Content-Type, Authorization, Cookie',
      'Access-Control-Max-Age': '86400',
    },
  });
}