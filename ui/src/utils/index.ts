import { Decimal } from 'decimal.js';
import dayjs from 'dayjs';

/**
 * 格式化时间
 * @param time 时间戳
 * @returns 格式化后的时间
 */

/**
 * 将token数量转换为人民币金额
 * @param tokenCount token数量
 * @returns 转换后的人民币金额(元)，带有金额单位
 */
export const convertTokensToRMB = (
  tokenCount: number | string,
  precision = 6
): string => {
  // 将token数量转为数字类型
  const tokens = Number(tokenCount);
  // 检查是否为有效数字
  if (isNaN(tokens)) {
    return '0元';
  }
  // 1元 = 1M tokens, 所以除以1000000得到元
  const rmb = Decimal(tokens).div(1000000);
  // 保留2位小数并添加单位
  return `${rmb.toFixed(precision)}`;
};

/**
 * 计算本月结束时间
 * @param startTimestamp 开始时间戳
 * @param endTimestamp 结束时间戳
 * @returns 计算后的结束时间戳
 */
export const calculateMonthEndTime = (
  startTimestamp: number,
  endTimestamp: number
): number => {
  // 一个月的秒数 (30天)
  const ONE_MONTH_SECONDS = 2592000;

  // 如果时间差大于一个月
  if (endTimestamp - startTimestamp > ONE_MONTH_SECONDS) {
    return startTimestamp + ONE_MONTH_SECONDS;
  }

  return endTimestamp;
};

// 定义聚合数据的接口
interface AggregatedDataItem {
  created_at: number;
  token_used: number;
  count: number;
  quota: number;
}

interface InputDataItem {
  created_at: number;
  token_used: number;
  count: number;
  quota: number;
}

// 按日期聚合数据
export const aggregatedTime = (
  data: InputDataItem[],
  type: 'd' | 's' = 'd',
  sort: 'asc' | 'desc' = 'desc'
): AggregatedDataItem[] => {
  function getDateFromTimestamp(timestamp: number) {
    return dayjs.unix(timestamp).format('YYYY-MM-DD');
  }
  const aggregatedData = data.reduce((acc: Record<string, AggregatedDataItem>, curr: InputDataItem) => {
    const date =
      type === 'd' ? getDateFromTimestamp(curr.created_at) : curr.created_at;
    if (!acc[date]) {
      acc[date] = {
        created_at: curr.created_at,
        token_used: 0,
        count: 0,
        quota: 0,
      };
    }
    acc[date].token_used += curr.token_used;
    acc[date].count += curr.count;
    acc[date].quota += curr.quota;

    return acc;
  }, {});
  return Object.values(aggregatedData).sort((a: AggregatedDataItem, b: AggregatedDataItem) =>
    sort === 'desc' ? b.created_at - a.created_at : a.created_at - b.created_at
  );
};

export function addCommasToNumber(num: number = 0) {
  return num.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',');
}

export function processText(inputText: string): string {
  function escapeRegExp(str: string): string {
    return str.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  }

  // 提取并移除所有<toolresult>标签
  const toolResults: Array<{ toolId: string; content: string }> = [];
  const toolResultPattern =
    /<toolresult\s+id\s*=\s*"result_([^"]+)"[^>]*>([\s\S]*?)<\/toolresult>/g;

  let modifiedText = inputText.replace(toolResultPattern, (match, toolId) => {
    toolResults.push({ toolId, content: match });
    return ''; // 移除原标签
  });

  // 将toolresult插入对应tool标签
  toolResults.forEach(({ toolId, content }) => {
    const escapedId = escapeRegExp(toolId);

    // 处理自闭合标签 <tool .../>
    // const selfClosingPattern = new RegExp(
    //   `<tool\\s+id\\s*=\\s*"${escapedId}"([^>]*?)\\s*\\/>`,
    //   "g"
    // );
    // modifiedText = modifiedText.replace(selfClosingPattern, (_, attrs) => {
    //   return `<tool id="${toolId}"${attrs}>${content}</tool>`;
    // });

    // 处理正常标签 <tool ...>...</tool>
    const normalPattern = new RegExp(
      `(<tool\\s+id\\s*=\\s*"${escapedId}"([^>]*)>)([\\s\\S]*?)(<\\/tool>)`,
      'g'
    );
    modifiedText = modifiedText.replace(
      normalPattern,
      (_, startTag, attrs, innerContent, endTag) => {
        return `${startTag}${innerContent}${content}\n\n${endTag}`;
      }
    );
  });

  return modifiedText;
}

export const isValidUrl = (url: string) => {
  const regex = /^(https?):\/\/[^\s/$.?#].[^\s]*$/i;
  return regex.test(url);
};

export const getRedirectUrl = (source: 'user' | 'admin' = 'admin') => {
  const searchParams = new URLSearchParams(location.search);
  const redirect =
    searchParams.get('redirect') ||
    `${source === 'admin' ? '/admin' : ''}/dashboard`;
  let redirectUrl: URL | null = null;
  try {
    redirectUrl = redirect ? new URL(decodeURIComponent(redirect)) : null;
  } catch (_) {
    redirectUrl = redirect
      ? new URL(location.origin + decodeURIComponent(redirect))
      : null;
  }

  redirectUrl = isValidUrl(redirectUrl?.href || '')
    ? redirectUrl
    : new URL(
          `${source === 'admin' ? '/admin' : ''}/dashboard`,
          location.origin
        );
  return redirectUrl as URL;
};

export const getRecentDaysData = (
  data: Record<string, number>[] = [],
  label: { keyLabel?: string; valueLabel?: string } = {},
  durationDay: number = 90
) => {
  const { keyLabel = 'timestamp', valueLabel = 'tokens' } = label;
  const xData: string[] = [];
  const yData: number[] = [];
  const dateMap: Record<string, number> = {};
  data.forEach((item) => {
    // 保留原始时间戳和tokens
    dateMap[dayjs.unix(item[keyLabel]!).format('YYYY-MM-DD')] =
      item[valueLabel]!;
  });

  for (let i = 0; i < durationDay; i++) {
    const time = dayjs().startOf('day').subtract(i, 'day').format('YYYY-MM-DD');
    if (dateMap[time]) {
      xData.unshift(time);
      yData.unshift(dateMap[time]);
    } else {
      xData.unshift(time);
      yData.unshift(0);
    }
  }
  return { xData, yData };
};

export const getRecent60MinutesData = (
  data: Record<string, number>[] = [],
  label: { keyLabel?: string; valueLabel?: string } = {}
) => {
  const { keyLabel = 'timestamp', valueLabel = 'tokens' } = label;
  const xData: string[] = [];
  const yData: number[] = [];
  const dateMap: Record<string, number> = {};
  data.forEach((item) => {
    // 保留原始时间戳和tokens
    dateMap[dayjs.unix(item[keyLabel]!).format('YYYY-MM-DD HH:mm')] =
      item[valueLabel]!;
  });

  for (let i = 0; i < 60; i++) {
    const time = dayjs().subtract(i, 'minute').format('YYYY-MM-DD HH:mm');
    if (dateMap[time]) {
      xData.unshift(time);
      yData.unshift(dateMap[time]);
    } else {
      xData.unshift(time);
      yData.unshift(0);
    }
  }
  return { xData, yData };
};

export const formatNumber = (num: number) => {
  if (num >= 1e11) {
    // 100B (100,000,000,000)
    return (num / 1e9).toFixed(2).replace(/\B(?=(\d{3})+(?!\d))/g, ',') + 'B'; // 直接取整，避免小数
  } else if (num >= 1e8) {
    // 100M (100,000,000)
    return (num / 1e6).toFixed(2).replace(/\B(?=(\d{3})+(?!\d))/g, ',') + 'M';
  } else if (num >= 1e5) {
    // 100K (100,000)
    return (num / 1e3).toFixed(2).replace(/\B(?=(\d{3})+(?!\d))/g, ',') + 'K';
  } else {
    // 添加千位逗号分隔符（如 99999 → 99,999）
    return num.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',');
  }
};

export const getRecent24HoursData = (
  data: Record<string, number>[] = [],
  label: { keyLabel?: string; valueLabel?: string } = {}
) => {
  const { keyLabel = 'timestamp', valueLabel = 'tokens' } = label;
  const xData: string[] = [];
  const yData: number[] = [];
  const dateMap: Record<string, number> = {};

  data.forEach((item) => {
    // 转为整点小时
    const hour = dayjs
      .unix(item[keyLabel]!)
      .startOf('hour')
      .format('YYYY-MM-DD HH:00');
    dateMap[hour] = item[valueLabel]!;
  });

  // 当前整点
  const now = dayjs().startOf('hour');
  for (let i = 23; i >= 0; i--) {
    const time = now.subtract(i, 'hour').format('YYYY-MM-DD HH:00');
    xData.push(time);
    yData.push(dateMap[time] || 0);
  }
  return { xData, yData };
};

export const getBaseLanguageId = (languageId: string): string => {
  const map: Record<string, string> = {
    typescriptreact: 'typescript',
    javascriptreact: 'javascript',
    tailwindcss: 'css',
    shellscript: 'shell',
    'vue-html': 'vue',
    tsx: 'typescript',
    jsx: 'javascript',
    py: 'python',
  };
  return map[languageId] || languageId;
};


export const getRangeData = (
  timeDuration: { start_at: number; end_at: number },
  data: { timestamp?: number; value?: number }[],
  precision: 'month' | 'day' | 'hour',
  label: { keyLabel?: string; valueLabel?: string } = { keyLabel: 'timestamp', valueLabel: 'value' }
) => {
  const { keyLabel = 'timestamp', valueLabel = 'value' } = label;
  const xData: (string | number)[] = [];
  const yData: number[] = [];
  const dateMap: Record<string, number> = {};

  // 将原始数据转换为时间映射
  data.forEach((item) => {
    const timestampValue = item[keyLabel as keyof typeof item] as number;
    const dataValue = item[valueLabel as keyof typeof item] as number;

    if (timestampValue && dataValue !== undefined) {
      // API返回的timestamp是秒，直接使用dayjs.unix()
      const timeKey = precision === 'day'
        ? dayjs.unix(timestampValue).format('YYYY-MM-DD')
        : dayjs.unix(timestampValue).format('YYYY-MM-DD HH:00');
      dateMap[timeKey] = dataValue;
    }
  });

  // timeDuration中的时间戳是毫秒，需要除以1000
  const startTime = dayjs.unix(timeDuration.start_at);
  const endTime = dayjs.unix(timeDuration.end_at);

  // 根据 precision 对齐时间点
  const startPoint = precision === 'day' ? startTime.startOf('day') : startTime.startOf('hour');
  const endPoint = precision === 'day' ? endTime.startOf('day') : endTime.startOf('hour');

  // 从开始时间到结束时间，逐个补齐数据点
  let currentTime = startPoint;
  while (currentTime.unix() <= endPoint.unix()) {
    const timeKey = precision === 'day'
      ? currentTime.format('YYYY-MM-DD')
      : currentTime.format('YYYY-MM-DD HH:00');

    xData.push(timeKey);
    yData.push(dateMap[timeKey] || 0);

    currentTime = currentTime.add(1, precision);
  }

  return { xData, yData };
};