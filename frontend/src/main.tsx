import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App.tsx';

// 全局错误处理 - 捕获未被React错误边界捕获的错误
window.addEventListener('error', (event) => {
  // 过滤掉URL构造错误
  if (event.error?.message?.includes('Failed to construct \'URL\'') || 
      event.error?.message?.includes('Invalid URL')) {
    console.warn('全局错误处理器捕获并忽略URL错误:', event.error.message);
    event.preventDefault(); // 阻止错误显示在控制台
    return;
  }
  
  // 其他错误正常处理
  console.error('全局错误:', event.error);
});

// 捕获Promise rejection错误
window.addEventListener('unhandledrejection', (event) => {
  if (event.reason?.message?.includes('Failed to construct \'URL\'') || 
      event.reason?.message?.includes('Invalid URL')) {
    console.warn('全局Promise rejection处理器捕获并忽略URL错误:', event.reason.message);
    event.preventDefault(); // 阻止错误显示
    return;
  }
  
  console.error('未处理的Promise rejection:', event.reason);
});

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
