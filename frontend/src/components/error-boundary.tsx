import { Component, ErrorInfo, ReactNode } from 'react';

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
}

interface State {
  hasError: boolean;
  error?: Error;
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error): State {
    // 对于URL构造错误，不设置错误状态
    if (
      error.message.includes("Failed to construct 'URL'") ||
      error.message.includes('Invalid URL')
    ) {
      console.warn('URL构造错误已被错误边界捕获并忽略:', error.message);
      return { hasError: false };
    }

    // 对于其他错误，设置错误状态
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    // 对于URL错误，只记录警告
    if (
      error.message.includes("Failed to construct 'URL'") ||
      error.message.includes('Invalid URL')
    ) {
      console.warn('URL构造错误已被错误边界捕获并忽略:', error.message);
      return;
    }

    // 记录其他错误到控制台
    console.error('错误边界捕获到错误:', error, errorInfo);
  }

  render() {
    if (this.state.hasError && this.state.error) {
      // 显示fallback UI
      return (
        this.props.fallback || (
          <div className="flex items-center justify-center min-h-screen">
            <div className="text-center">
              <h2 className="text-lg font-semibold text-gray-900 mb-2">
                出现了一些问题
              </h2>
              <p className="text-gray-600 mb-4">请刷新页面重试</p>
              <button
                onClick={() => window.location.reload()}
                className="px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600"
              >
                刷新页面
              </button>
            </div>
          </div>
        )
      );
    }

    return this.props.children;
  }
}
