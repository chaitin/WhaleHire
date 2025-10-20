// 全局简历轮询管理器
import { useRef, useCallback, useEffect } from 'react';
import { getResumeProgress } from '@/services/resume';
import { ResumeParseProgress } from '@/types/resume';

interface PollingItem {
  id: string;
  callbacks: Set<(progress: ResumeParseProgress) => void>;
  lastProgress?: ResumeParseProgress;
}

class ResumePollingManager {
  private pollingItems = new Map<string, PollingItem>();
  private pollingTimer: ReturnType<typeof setInterval> | null = null;
  private pollingInterval = 10000; // 10秒轮询间隔

  // 添加轮询项
  addPolling(
    resumeId: string,
    callback: (progress: ResumeParseProgress) => void
  ) {
    if (!this.pollingItems.has(resumeId)) {
      this.pollingItems.set(resumeId, {
        id: resumeId,
        callbacks: new Set(),
      });
    }

    const item = this.pollingItems.get(resumeId)!;
    item.callbacks.add(callback);

    // 如果是第一个轮询项，启动轮询
    if (this.pollingItems.size === 1 && !this.pollingTimer) {
      this.startPolling();
    }

    console.log(
      `➕ 添加轮询: ${resumeId}, 当前轮询数量: ${this.pollingItems.size}`
    );
  }

  // 移除轮询项
  removePolling(
    resumeId: string,
    callback: (progress: ResumeParseProgress) => void
  ) {
    const item = this.pollingItems.get(resumeId);
    if (!item) return;

    item.callbacks.delete(callback);

    // 如果没有回调了，移除整个轮询项
    if (item.callbacks.size === 0) {
      this.pollingItems.delete(resumeId);
      console.log(
        `➖ 移除轮询: ${resumeId}, 剩余轮询数量: ${this.pollingItems.size}`
      );
    }

    // 如果没有轮询项了，停止轮询
    if (this.pollingItems.size === 0) {
      this.stopPolling();
    }
  }

  // 启动轮询
  private startPolling() {
    if (this.pollingTimer) return;

    console.log(
      `🔄 启动全局简历进度轮询 (间隔: ${this.pollingInterval / 1000}秒)`
    );
    this.pollingTimer = setInterval(() => {
      this.pollAllItems();
    }, this.pollingInterval);

    // 立即执行一次
    this.pollAllItems();
  }

  // 停止轮询
  private stopPolling() {
    if (this.pollingTimer) {
      clearInterval(this.pollingTimer);
      this.pollingTimer = null;
      console.log('⏹️ 停止全局简历进度轮询');
    }
  }

  // 轮询所有项目
  private async pollAllItems() {
    const promises = Array.from(this.pollingItems.entries()).map(
      async ([resumeId, item]) => {
        try {
          const progress = await getResumeProgress(resumeId);

          // 检查状态是否需要继续轮询
          if (progress.status === 'completed' || progress.status === 'failed') {
            // 通知所有回调后移除轮询
            item.callbacks.forEach((callback) => callback(progress));

            // 移除所有回调，这会自动清理轮询项
            const callbacksToRemove = Array.from(item.callbacks);
            callbacksToRemove.forEach((callback) => {
              this.removePolling(resumeId, callback);
            });

            return;
          }

          // 只有在进度发生变化时才通知回调
          if (
            !item.lastProgress ||
            item.lastProgress.progress !== progress.progress ||
            item.lastProgress.status !== progress.status
          ) {
            item.lastProgress = progress;
            item.callbacks.forEach((callback) => callback(progress));
          }
        } catch (error) {
          console.error(`轮询简历 ${resumeId} 进度失败:`, error);
          // 发生错误时也移除轮询
          const callbacksToRemove = Array.from(item.callbacks);
          callbacksToRemove.forEach((callback) => {
            this.removePolling(resumeId, callback);
          });
        }
      }
    );

    await Promise.allSettled(promises);
  }

  // 获取当前轮询状态
  getPollingStatus() {
    return {
      isPolling: !!this.pollingTimer,
      pollingCount: this.pollingItems.size,
      pollingIds: Array.from(this.pollingItems.keys()),
    };
  }

  // 清理所有轮询
  cleanup() {
    this.stopPolling();
    this.pollingItems.clear();
  }
}

// 全局单例
const globalPollingManager = new ResumePollingManager();

// Hook 接口
export const useResumePollingManager = () => {
  const callbackRef = useRef<((progress: ResumeParseProgress) => void) | null>(
    null
  );
  const currentResumeIdRef = useRef<string | null>(null);

  const startPolling = useCallback(
    (resumeId: string, callback: (progress: ResumeParseProgress) => void) => {
      // 清理之前的轮询
      if (currentResumeIdRef.current && callbackRef.current) {
        globalPollingManager.removePolling(
          currentResumeIdRef.current,
          callbackRef.current
        );
      }

      // 设置新的轮询
      callbackRef.current = callback;
      currentResumeIdRef.current = resumeId;
      globalPollingManager.addPolling(resumeId, callback);
    },
    []
  );

  const stopPolling = useCallback(() => {
    if (currentResumeIdRef.current && callbackRef.current) {
      globalPollingManager.removePolling(
        currentResumeIdRef.current,
        callbackRef.current
      );
      callbackRef.current = null;
      currentResumeIdRef.current = null;
    }
  }, []);

  const getStatus = useCallback(() => {
    return globalPollingManager.getPollingStatus();
  }, []);

  // 清理
  useEffect(() => {
    return () => {
      stopPolling();
    };
  }, [stopPolling]);

  return {
    startPolling,
    stopPolling,
    getStatus,
  };
};

export default globalPollingManager;
