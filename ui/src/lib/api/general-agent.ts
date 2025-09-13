// General Agent 相关 API

import type {
  ApiResponse,
  Conversation,
  CreateConversationRequest,
  GenerateRequest,
  GenerateResponse,
  ListConversationsRequest,
  ListConversationsResponse,
  GetConversationHistoryRequest,
  AddMessageToConversationRequest,
  StreamChunk,
  StreamMetadata,
} from './types';

/**
 * 生成AI回复
 */
export async function generate(request: GenerateRequest): Promise<GenerateResponse | null> {
  try {
    const response = await fetch('/api/v1/general-agent/generate', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      console.error('生成AI回复失败:', response.status, response.statusText);
      return null;
    }

    const result: ApiResponse<GenerateResponse> = await response.json();
    
    if (result.code === 0 && result.data) {
      return result.data;
    } else {
      console.error('API返回错误:', result.message);
      return null;
    }
  } catch (error) {
    console.error('生成AI回复请求失败:', error);
    return null;
  }
}

/**
 * 流式生成AI回复 (SSE 版本)
 */
export async function generateStream(
  request: GenerateRequest,
  onChunk: (chunk: StreamChunk) => void,
  onError?: (error: string) => void,
  onComplete?: () => void,
  onMetadata?: (metadata: StreamMetadata) => void
): Promise<void> {
  try {
    const response = await fetch('/api/v1/general-agent/generate-stream', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'text/event-stream',
        'Cache-Control': 'no-cache',
      },
      credentials: 'include',
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      const errorMsg = `流式生成失败: ${response.status} ${response.statusText}`;
      console.error(errorMsg);
      onError?.(errorMsg);
      return;
    }

    const reader = response.body?.getReader();
    if (!reader) {
      const errorMsg = '无法获取响应流';
      console.error(errorMsg);
      onError?.(errorMsg);
      return;
    }

    const decoder = new TextDecoder();
    let buffer = '';

    try {
      while (true) {
        const { done, value } = await reader.read();

        if (done) {
          // 读完了，触发完成
          onComplete?.();
          break;
        }

        buffer += decoder.decode(value, { stream: true });

        // SSE 事件之间用 \n\n 分隔
        const events = buffer.split('\n\n');
        buffer = events.pop() || ''; // 留下可能不完整的一条

        for (const eventBlock of events) {
          const lines = eventBlock.split('\n');
          let eventType = '';
          let eventData = '';

          for (const line of lines) {
            if (line.startsWith('event: ')) {
              eventType = line.slice(7).trim();
            } else if (line.startsWith('data: ')) {
              eventData += line.slice(6).trim();
            }
          }

          try {
            if (eventType === 'data') {
              const data = JSON.parse(eventData);

              // 判断是否是 metadata
              if (data.version && data.conversation_id) {
                onMetadata?.(data as StreamMetadata);
              } else {
                onChunk(data as StreamChunk);
              }
            } else if (eventType === 'error') {
              onError?.(eventData || '流式生成出错');
              return;
            } else if (eventType === 'done') {
              onComplete?.();
              return;
            }
          } catch (err) {
            console.error('解析 SSE 数据失败:', err, eventData);
          }
        }
      }
    } finally {
      reader.releaseLock();
    }
  } catch (error) {
    const errorMsg = `流式生成请求失败: ${error}`;
    console.error(errorMsg);
    onError?.(errorMsg);
  }
}


/**
 * 创建新对话会话
 */
export async function createConversation(request: CreateConversationRequest): Promise<Conversation | null> {
  try {
    const response = await fetch('/api/v1/general-agent/conversations', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      console.error('创建对话失败:', response.status, response.statusText);
      return null;
    }

    const result: ApiResponse<Conversation> = await response.json();
    
    if (result.code === 0 && result.data) {
      return result.data;
    } else {
      console.error('API返回错误:', result.message);
      return null;
    }
  } catch (error) {
    console.error('创建对话请求失败:', error);
    return null;
  }
}

/**
 * 分页获取用户对话列表
 */
export async function listConversations(request?: ListConversationsRequest): Promise<ListConversationsResponse | null> {
  try {
    const params = new URLSearchParams();
    if (request?.page) params.append('page', request.page.toString());
    if (request?.size) params.append('size', request.size.toString());
    if (request?.search) params.append('search', request.search);

    const url = `/api/v1/general-agent/conversations${params.toString() ? '?' + params.toString() : ''}`;
    
    const response = await fetch(url, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
    });

    if (!response.ok) {
      console.error('获取对话列表失败:', response.status, response.statusText);
      return null;
    }

    const result: ApiResponse<ListConversationsResponse> = await response.json();
    
    console.log('获取对话列表响应:', result.data);
    if (result.code === 0 && result.data) {
      return result.data;
    } else {
      console.error('API返回错误:', result.message);
      return null;
    }
  } catch (error) {
    console.error('获取对话列表请求失败:', error);
    return null;
  }
}

/**
 * 获取对话历史记录
 */
export async function getConversationHistory(request: GetConversationHistoryRequest): Promise<Conversation | null> {
  try {
    const response = await fetch('/api/v1/general-agent/conversations/history', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify({
        conversation_id: request.conversation_id,
      }),
    });

    if (!response.ok) {
      console.error('获取对话历史失败:', response.status, response.statusText);
      return null;
    }

    const result: ApiResponse<Conversation> = await response.json();
    
    console.log('获取对话历史响应:', result.data);
    if (result.code === 0 && result.data) {
      return result.data;
    } else {
      console.error('API返回错误:', result.message);
      return null;
    }
  } catch (error) {
    console.error('获取对话历史请求失败:', error);
    return null;
  }
}

/**
 * 删除指定对话
 */
export async function deleteConversation(conversationId: string): Promise<boolean> {
  try {
    const response = await fetch(`/api/v1/general-agent/conversations?id=${conversationId}`, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
    });

    if (!response.ok) {
      console.error('删除对话失败:', response.status, response.statusText);
      return false;
    }

    const result: ApiResponse = await response.json();
    
    if (result.code === 0) {
      return true;
    } else {
      console.error('API返回错误:', result.message);
      return false;
    }
  } catch (error) {
    console.error('删除对话请求失败:', error);
    return false;
  }
}

/**
 * 向对话添加新消息
 */
export async function addMessageToConversation(conversationId: string, request: AddMessageToConversationRequest): Promise<boolean> {
  try {
    const response = await fetch(`/api/v1/general-agent/conversations/${conversationId}/addmessage`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      console.error('添加消息失败:', response.status, response.statusText);
      return false;
    }

    const result: ApiResponse = await response.json();
    
    if (result.code === 0) {
      return true;
    } else {
      console.error('API返回错误:', result.message);
      return false;
    }
  } catch (error) {
    console.error('添加消息请求失败:', error);
    return false;
  }
}