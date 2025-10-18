# WhaleHire 动效使用指南

本指南介绍如何在 WhaleHire 项目中使用各种动效。

## 📋 目录

1. [页面切换动画](#页面切换动画)
2. [按钮动效](#按钮动效)
3. [卡片悬停效果](#卡片悬停效果)
4. [列表项动画](#列表项动画)
5. [表格行动画](#表格行动画)
6. [模态框动画](#模态框动画)
7. [自定义 Hooks](#自定义-hooks)

---

## 页面切换动画

### 基础页面淡入

```tsx
function MyPage() {
  return (
    <div className="page-enter">
      <div className="page-content">
        {/* 页面内容 */}
      </div>
    </div>
  );
}
```

### 使用 Hook

```tsx
import { usePageAnimation } from '@/hooks/useAnimation';

function MyPage() {
  const isVisible = usePageAnimation();
  
  return (
    <div className={`transition-opacity duration-300 ${isVisible ? 'opacity-100' : 'opacity-0'}`}>
      {/* 页面内容 */}
    </div>
  );
}
```

---

## 按钮动效

### 主要按钮（带缩放和阴影）

```tsx
<Button className="btn-primary">
  提交
</Button>
```

### 次要按钮（轻微缩放）

```tsx
<Button className="btn-secondary">
  取消
</Button>
```

### 操作按钮（用于表格操作等）

```tsx
<Button className="action-btn">
  <Edit className="h-4 w-4" />
</Button>
```

### 自定义按钮动效

```tsx
<button className="transition-all duration-200 hover:scale-105 hover:shadow-md active:scale-95">
  自定义按钮
</button>
```

---

## 卡片悬停效果

### 标准卡片悬停

```tsx
<div className="card-hover rounded-lg border p-6">
  <h3>卡片标题</h3>
  <p>卡片内容</p>
</div>
```

### 自定义卡片动效

```tsx
<div className="transition-all duration-300 hover:shadow-xl hover:-translate-y-2 hover:border-primary">
  {/* 卡片内容 */}
</div>
```

---

## 列表项动画

### 基础列表项

```tsx
<ul>
  {items.map((item) => (
    <li key={item.id} className="list-item">
      {item.name}
    </li>
  ))}
</ul>
```

### 交错动画列表

```tsx
import { useStaggerAnimation } from '@/hooks/useAnimation';

function MyList({ items }) {
  const visibleItems = useStaggerAnimation(items.length, 50);
  
  return (
    <ul>
      {items.map((item, index) => (
        <li
          key={item.id}
          className={`transform transition-all duration-300 ${
            visibleItems.includes(index)
              ? 'opacity-100 translate-x-0'
              : 'opacity-0 -translate-x-4'
          }`}
        >
          {item.name}
        </li>
      ))}
    </ul>
  );
}
```

---

## 表格行动画

### 表格行悬停

```tsx
<table>
  <tbody>
    {data.map((row) => (
      <tr key={row.id} className="table-row-hover">
        <td>{row.name}</td>
        <td>{row.value}</td>
      </tr>
    ))}
  </tbody>
</table>
```

### 选中效果

```tsx
<tr className={selected ? 'selected-item' : 'table-row-hover'}>
  {/* 表格单元格 */}
</tr>
```

---

## 模态框动画

### 标准模态框

```tsx
<Dialog open={open} onOpenChange={setOpen}>
  <DialogContent className="modal-enter">
    {/* 模态框内容 */}
  </DialogContent>
</Dialog>
```

### 带遮罩层动画

```tsx
{isOpen && (
  <>
    <div className="modal-overlay fixed inset-0 bg-black/50" />
    <div className="modal-enter fixed inset-0 flex items-center justify-center">
      {/* 模态框内容 */}
    </div>
  </>
)}
```

---

## 输入框动效

### 带焦点动画的输入框

```tsx
<input
  type="text"
  className="input-focus w-full rounded-lg border px-4 py-2"
  placeholder="请输入..."
/>
```

### 搜索框展开效果

```tsx
<input
  type="search"
  className="search-input w-48 focus:w-64 rounded-lg border px-4 py-2"
  placeholder="搜索..."
/>
```

---

## 下拉菜单动画

```tsx
<DropdownMenu>
  <DropdownMenuTrigger>打开菜单</DropdownMenuTrigger>
  <DropdownMenuContent className="dropdown-enter">
    <DropdownMenuItem>选项 1</DropdownMenuItem>
    <DropdownMenuItem>选项 2</DropdownMenuItem>
  </DropdownMenuContent>
</DropdownMenu>
```

---

## 标签页动画

```tsx
<div className="flex border-b">
  {tabs.map((tab) => (
    <button
      key={tab.id}
      className={`tab-item px-4 py-2 ${
        activeTab === tab.id ? 'active' : ''
      }`}
      onClick={() => setActiveTab(tab.id)}
    >
      {tab.label}
    </button>
  ))}
</div>
```

---

## 加载动画

### 骨架屏

```tsx
<div className="skeleton h-20 w-full rounded-lg" />
```

### 闪烁效果

```tsx
<div className="skeleton-shimmer h-20 w-full rounded-lg" />
```

### 使用 Hook

```tsx
import { useLoadingAnimation } from '@/hooks/useAnimation';

function MyComponent() {
  const [loading, setLoading] = useState(true);
  const showLoading = useLoadingAnimation(loading, 300);
  
  return showLoading ? <Skeleton /> : <Content />;
}
```

---

## 通知动画

```tsx
<div className="notification-enter fixed top-4 right-4 bg-white rounded-lg shadow-lg p-4">
  通知内容
</div>
```

---

## 图标动画

### 旋转图标

```tsx
<RefreshCw className="icon-spin h-5 w-5" />
```

### 自定义图标动画

```tsx
<ChevronDown
  className={`h-4 w-4 transition-transform duration-200 ${
    isExpanded ? 'rotate-180' : 'rotate-0'
  }`}
/>
```

---

## 进度条动画

```tsx
<div className="h-2 bg-gray-200 rounded-full overflow-hidden">
  <div
    className="progress-bar h-full bg-primary"
    style={{ width: `${progress}%` }}
  />
</div>
```

---

## 滚动显示动画

```tsx
import { useScrollAnimation } from '@/hooks/useAnimation';

function ScrollRevealComponent() {
  const { ref, isVisible } = useScrollAnimation(0.1);
  
  return (
    <div
      ref={ref}
      className={`transition-all duration-500 ${
        isVisible
          ? 'opacity-100 translate-y-0'
          : 'opacity-0 translate-y-10'
      }`}
    >
      {/* 内容 */}
    </div>
  );
}
```

---

## 数字滚动动画

```tsx
import { useCountAnimation } from '@/hooks/useAnimation';

function CounterComponent({ target }: { target: number }) {
  const count = useCountAnimation(target, 1000);
  
  return <div className="text-4xl font-bold">{count}</div>;
}
```

---

## 打字机效果

```tsx
import { useTypewriterEffect } from '@/hooks/useAnimation';

function TypewriterText({ text }: { text: string }) {
  const { displayedText } = useTypewriterEffect(text, 50);
  
  return <p>{displayedText}</p>;
}
```

---

## 波纹效果

```tsx
import { useRippleEffect } from '@/hooks/useAnimation';

function RippleButton() {
  const { ripples, addRipple } = useRippleEffect();
  
  return (
    <button
      className="relative overflow-hidden"
      onClick={addRipple}
    >
      点击我
      {ripples.map((ripple) => (
        <span
          key={ripple.id}
          className="absolute animate-ping bg-white/30 rounded-full"
          style={{
            left: ripple.x,
            top: ripple.y,
            width: 10,
            height: 10,
            transform: 'translate(-50%, -50%)',
          }}
        />
      ))}
    </button>
  );
}
```

---

## 自定义动画类

### 快速过渡

```tsx
<div className="transition-quick">
  {/* 150ms 过渡 */}
</div>
```

### 平滑过渡

```tsx
<div className="transition-smooth">
  {/* 300ms 过渡 */}
</div>
```

### 慢速过渡

```tsx
<div className="transition-slow">
  {/* 500ms 过渡 */}
</div>
```

---

## 最佳实践

### 1. 性能优化

- 使用 `will-change` 属性提示浏览器
```css
.will-animate {
  will-change: transform, opacity;
}
```

- 优先使用 `transform` 和 `opacity`（GPU 加速）
- 避免在大量元素上同时使用复杂动画

### 2. 动画时长建议

- **快速交互**: 150-200ms（按钮点击、悬停）
- **标准过渡**: 300ms（页面切换、模态框）
- **慢速展示**: 500-600ms（大型元素、复杂动画）

### 3. 缓动函数

- **ease-out**: 适合元素进入（快速开始，缓慢结束）
- **ease-in**: 适合元素退出（缓慢开始，快速结束）
- **ease-in-out**: 适合状态切换（平滑的开始和结束）

### 4. 辅助功能

对于有运动敏感的用户，可以添加：

```css
@media (prefers-reduced-motion: reduce) {
  * {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}
```

---

## 常用动画组合

### 卡片进入效果

```tsx
<div className="animate-scale-in">
  {/* 卡片内容 */}
</div>
```

### 列表项滑入

```tsx
<div className="animate-slide-in-left">
  {/* 列表项内容 */}
</div>
```

### 弹出效果

```tsx
<div className="animate-bounce-in">
  {/* 弹出内容 */}
</div>
```

### 闪烁提示

```tsx
<div className="animate-pulse-subtle">
  {/* 提示内容 */}
</div>
```

---

## 调试技巧

### 1. 慢动作播放

在浏览器开发者工具中，可以降低动画速度：

```css
* {
  animation-duration: 3s !important;
  transition-duration: 3s !important;
}
```

### 2. 显示动画边界

```css
* {
  outline: 1px solid red !important;
}
```

---

## 总结

- 所有动画类都在 `globals.css` 中定义
- 自定义 Hooks 在 `hooks/useAnimation.ts` 中
- Tailwind 配置扩展了额外的动画关键帧
- 优先使用 CSS 动画，必要时使用 JS 动画
- 保持动画简洁、快速、有意义

---

## 更多资源

- [Tailwind CSS 动画文档](https://tailwindcss.com/docs/animation)
- [Framer Motion（高级动画库）](https://www.framer.com/motion/)
- [CSS Transition 指南](https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_Transitions)

