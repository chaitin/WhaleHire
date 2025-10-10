import { type ClassValue, clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatDate(
  date?: string | number | Date,
  format: 'date' | 'datetime' = 'date'
): string {
  if (!date && date !== 0) {
    return '-';
  }

  let parsedDate: Date;

  if (date instanceof Date) {
    parsedDate = date;
  } else if (typeof date === 'number') {
    // 处理秒或毫秒时间戳
    const timestamp = date < 1e12 ? date * 1000 : date;
    parsedDate = new Date(timestamp);
  } else {
    const normalizedDate = date.replace(' ', 'T');
    parsedDate = new Date(normalizedDate);
  }

  if (Number.isNaN(parsedDate.getTime())) {
    return '-';
  }

  if (format === 'datetime') {
    return parsedDate.toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    });
  }

  return parsedDate.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  });
}

export function formatPhone(phone?: string | null): string {
  if (!phone) {
    return '--';
  }

  const normalized = phone.replace(/\s+/g, '');

  // 格式化手机号，隐藏中间4位
  if (normalized.length === 11) {
    return `${normalized.slice(0, 3)}****${normalized.slice(7)}`;
  }

  return normalized;
}
