/**
 * 动效工具 Hook
 * 提供页面切换、组件显示等动画效果
 */

import { useEffect, useState, useRef } from 'react';

/**
 * 页面进入动画 Hook
 * 用于页面加载时的淡入效果
 */
export function usePageAnimation() {
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    // 延迟一帧执行，确保初始状态被应用
    requestAnimationFrame(() => {
      setIsVisible(true);
    });
  }, []);

  return isVisible;
}

/**
 * 列表项交错动画 Hook
 * 用于列表项依次出现的效果
 */
export function useStaggerAnimation(itemCount: number, delay = 50) {
  const [visibleItems, setVisibleItems] = useState<number[]>([]);

  useEffect(() => {
    const timers: NodeJS.Timeout[] = [];

    for (let i = 0; i < itemCount; i++) {
      const timer = setTimeout(() => {
        setVisibleItems((prev) => [...prev, i]);
      }, i * delay);
      timers.push(timer);
    }

    return () => {
      timers.forEach((timer) => clearTimeout(timer));
    };
  }, [itemCount, delay]);

  return visibleItems;
}

/**
 * 滚动显示动画 Hook
 * 元素进入视口时触发动画
 */
export function useScrollAnimation(threshold = 0.1) {
  const elementRef = useRef<HTMLDivElement>(null);
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    const element = elementRef.current;
    if (!element) return;

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setIsVisible(true);
          observer.unobserve(element);
        }
      },
      { threshold }
    );

    observer.observe(element);

    return () => {
      if (element) {
        observer.unobserve(element);
      }
    };
  }, [threshold]);

  return { ref: elementRef, isVisible };
}

/**
 * Hover 动画 Hook
 * 提供悬停状态管理
 */
export function useHoverAnimation() {
  const [isHovered, setIsHovered] = useState(false);

  const hoverProps = {
    onMouseEnter: () => setIsHovered(true),
    onMouseLeave: () => setIsHovered(false),
  };

  return { isHovered, hoverProps };
}

/**
 * 点击波纹效果 Hook
 */
export function useRippleEffect() {
  const [ripples, setRipples] = useState<
    Array<{ x: number; y: number; id: number }>
  >([]);

  const addRipple = (event: React.MouseEvent<HTMLElement>) => {
    const rect = event.currentTarget.getBoundingClientRect();
    const x = event.clientX - rect.left;
    const y = event.clientY - rect.top;
    const id = Date.now();

    setRipples((prev) => [...prev, { x, y, id }]);

    setTimeout(() => {
      setRipples((prev) => prev.filter((ripple) => ripple.id !== id));
    }, 600);
  };

  return { ripples, addRipple };
}

/**
 * 加载动画 Hook
 * 用于显示加载状态
 */
export function useLoadingAnimation(loading: boolean, minDuration = 300) {
  const [showLoading, setShowLoading] = useState(loading);

  useEffect(() => {
    if (loading) {
      setShowLoading(true);
    } else {
      const timer = setTimeout(() => {
        setShowLoading(false);
      }, minDuration);

      return () => clearTimeout(timer);
    }
  }, [loading, minDuration]);

  return showLoading;
}

/**
 * 计数动画 Hook
 * 数字滚动效果
 */
export function useCountAnimation(target: number, duration = 1000, start = 0) {
  const [count, setCount] = useState(start);

  useEffect(() => {
    const startTime = Date.now();
    const difference = target - start;

    const animate = () => {
      const now = Date.now();
      const progress = Math.min((now - startTime) / duration, 1);
      const easeProgress = 1 - Math.pow(1 - progress, 3); // easeOut cubic

      setCount(Math.floor(start + difference * easeProgress));

      if (progress < 1) {
        requestAnimationFrame(animate);
      }
    };

    requestAnimationFrame(animate);
  }, [target, duration, start]);

  return count;
}

/**
 * 打字机效果 Hook
 */
export function useTypewriterEffect(text: string, speed = 50) {
  const [displayedText, setDisplayedText] = useState('');
  const [isComplete, setIsComplete] = useState(false);

  useEffect(() => {
    setDisplayedText('');
    setIsComplete(false);
    let currentIndex = 0;

    const timer = setInterval(() => {
      if (currentIndex < text.length) {
        setDisplayedText((prev) => prev + text[currentIndex]);
        currentIndex++;
      } else {
        setIsComplete(true);
        clearInterval(timer);
      }
    }, speed);

    return () => clearInterval(timer);
  }, [text, speed]);

  return { displayedText, isComplete };
}

/**
 * 淡入淡出切换 Hook
 */
export function useFadeTransition(value: boolean, duration = 300) {
  const [internalValue, setInternalValue] = useState(value);
  const [isTransitioning, setIsTransitioning] = useState(false);

  useEffect(() => {
    if (value !== internalValue) {
      setIsTransitioning(true);

      const timer = setTimeout(() => {
        setInternalValue(value);
        setIsTransitioning(false);
      }, duration);

      return () => clearTimeout(timer);
    }
  }, [value, internalValue, duration]);

  return { value: internalValue, isTransitioning };
}
