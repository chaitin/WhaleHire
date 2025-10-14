import {
  MatchingTask,
  MatchingStats,
  MatchingTaskDetail,
  MatchResult,
} from '@/types/matching';

/**
 * 模拟匹配任务数据
 */
export const mockMatchingTasks: MatchingTask[] = [
  {
    id: '1',
    taskId: 'MT20230512001',
    jobPositions: ['高级前端工程师'],
    resumeCount: 128,
    status: 'completed',
    creator: '张经理',
    createdAt: 1683875425, // 2023-05-12 14:30:25
  },
  {
    id: '2',
    taskId: 'MT20230511002',
    jobPositions: ['UI/UX设计师'],
    resumeCount: 86,
    status: 'completed',
    creator: '李主管',
    createdAt: 1683770142, // 2023-05-11 09:15:42
  },
  {
    id: '3',
    taskId: 'MT20230510003',
    jobPositions: ['产品经理'],
    resumeCount: 215,
    status: 'in_progress',
    creator: '王总监',
    createdAt: 1683709338, // 2023-05-10 16:42:18
  },
  {
    id: '4',
    taskId: 'MT20230509004',
    jobPositions: ['后端开发工程师'],
    resumeCount: 156,
    status: 'completed',
    creator: '张经理',
    createdAt: 1683603633, // 2023-05-09 11:20:33
  },
  {
    id: '5',
    taskId: 'MT20230508005',
    jobPositions: ['数据分析师'],
    resumeCount: 78,
    status: 'failed',
    creator: '赵专员',
    createdAt: 1683530712, // 2023-05-08 15:05:12
  },
  {
    id: '6',
    taskId: 'MT20230507006',
    jobPositions: ['全栈开发工程师'],
    resumeCount: 132,
    status: 'completed',
    creator: '张经理',
    createdAt: 1683424245, // 2023-05-07 10:30:45
  },
  {
    id: '7',
    taskId: 'MT20230506007',
    jobPositions: ['测试工程师'],
    resumeCount: 95,
    status: 'completed',
    creator: '李主管',
    createdAt: 1683351756, // 2023-05-06 14:22:36
  },
  {
    id: '8',
    taskId: 'MT20230505008',
    jobPositions: ['DevOps工程师'],
    resumeCount: 42,
    status: 'in_progress',
    creator: '王总监',
    createdAt: 1683249917, // 2023-05-05 09:45:17
  },
  {
    id: '9',
    taskId: 'MT20230504009',
    jobPositions: ['iOS开发工程师', 'Android开发工程师'],
    resumeCount: 118,
    status: 'completed',
    creator: '张经理',
    createdAt: 1683187828, // 2023-05-04 16:10:28
  },
  {
    id: '10',
    taskId: 'MT20230503010',
    jobPositions: ['运维工程师'],
    resumeCount: 65,
    status: 'completed',
    creator: '赵专员',
    createdAt: 1683085549, // 2023-05-03 11:55:49
  },
];

/**
 * 模拟统计数据
 */
export const mockMatchingStats: MatchingStats = {
  total: 128,
  inProgress: 24,
  completed: 104,
};

/**
 * 获取匹配任务列表（分页）
 */
export function getMockMatchingTasks(
  page: number = 1,
  pageSize: number = 10,
  filters?: {
    position?: string;
    status?: string;
    keywords?: string;
  }
): {
  data: MatchingTask[];
  pagination: {
    current: number;
    pageSize: number;
    total: number;
    totalPages: number;
  };
} {
  let filteredTasks = [...mockMatchingTasks];

  // 应用筛选条件
  if (filters?.status && filters.status !== 'all') {
    filteredTasks = filteredTasks.filter(
      (task) => task.status === filters.status
    );
  }

  if (filters?.position && filters.position !== 'all') {
    filteredTasks = filteredTasks.filter((task) =>
      task.jobPositions.some((pos) => pos.includes(filters.position!))
    );
  }

  if (filters?.keywords) {
    const keywords = filters.keywords.toLowerCase();
    filteredTasks = filteredTasks.filter(
      (task) =>
        task.taskId.toLowerCase().includes(keywords) ||
        task.jobPositions.some((pos) => pos.toLowerCase().includes(keywords)) ||
        task.creator.toLowerCase().includes(keywords)
    );
  }

  const total = filteredTasks.length;
  const totalPages = Math.ceil(total / pageSize);
  const start = (page - 1) * pageSize;
  const end = start + pageSize;
  const data = filteredTasks.slice(start, end);

  return {
    data,
    pagination: {
      current: page,
      pageSize,
      total,
      totalPages,
    },
  };
}

/**
 * 获取匹配统计数据
 */
export function getMockMatchingStats(): MatchingStats {
  return mockMatchingStats;
}

/**
 * 模拟匹配结果数据
 */
const mockMatchResults: MatchResult[] = [
  {
    id: '1',
    resume: {
      id: 'r1',
      name: '李明',
      age: 32,
      education: '本科',
      experience: '5年经验',
    },
    job: {
      id: 'j1',
      title: '高级前端工程师',
      department: '技术部',
      jobId: 'JOB20230415',
    },
    matchScore: 92,
  },
  {
    id: '2',
    resume: {
      id: 'r2',
      name: '王芳',
      age: 28,
      education: '本科',
      experience: '4年经验',
    },
    job: {
      id: 'j1',
      title: '高级前端工程师',
      department: '技术部',
      jobId: 'JOB20230415',
    },
    matchScore: 87,
  },
  {
    id: '3',
    resume: {
      id: 'r3',
      name: '张伟',
      age: 35,
      education: '硕士',
      experience: '8年经验',
    },
    job: {
      id: 'j1',
      title: '高级前端工程师',
      department: '技术部',
      jobId: 'JOB20230415',
    },
    matchScore: 78,
  },
  {
    id: '4',
    resume: {
      id: 'r4',
      name: '刘洋',
      age: 29,
      education: '本科',
      experience: '5年经验',
    },
    job: {
      id: 'j1',
      title: '高级前端工程师',
      department: '技术部',
      jobId: 'JOB20230415',
    },
    matchScore: 85,
  },
  {
    id: '5',
    resume: {
      id: 'r5',
      name: '陈静',
      age: 27,
      education: '本科',
      experience: '3年经验',
    },
    job: {
      id: 'j1',
      title: '高级前端工程师',
      department: '技术部',
      jobId: 'JOB20230415',
    },
    matchScore: 82,
  },
  {
    id: '6',
    resume: {
      id: 'r6',
      name: '赵强',
      age: 31,
      education: '本科',
      experience: '6年经验',
    },
    job: {
      id: 'j1',
      title: '高级前端工程师',
      department: '技术部',
      jobId: 'JOB20230415',
    },
    matchScore: 89,
  },
  {
    id: '7',
    resume: {
      id: 'r7',
      name: '孙丽',
      age: 26,
      education: '本科',
      experience: '3年经验',
    },
    job: {
      id: 'j1',
      title: '高级前端工程师',
      department: '技术部',
      jobId: 'JOB20230415',
    },
    matchScore: 80,
  },
  {
    id: '8',
    resume: {
      id: 'r8',
      name: '周敏',
      age: 30,
      education: '硕士',
      experience: '5年经验',
    },
    job: {
      id: 'j1',
      title: '高级前端工程师',
      department: '技术部',
      jobId: 'JOB20230415',
    },
    matchScore: 91,
  },
  {
    id: '9',
    resume: {
      id: 'r9',
      name: '吴磊',
      age: 33,
      education: '本科',
      experience: '7年经验',
    },
    job: {
      id: 'j1',
      title: '高级前端工程师',
      department: '技术部',
      jobId: 'JOB20230415',
    },
    matchScore: 88,
  },
  {
    id: '10',
    resume: {
      id: 'r10',
      name: '郑娜',
      age: 28,
      education: '本科',
      experience: '4年经验',
    },
    job: {
      id: 'j1',
      title: '高级前端工程师',
      department: '技术部',
      jobId: 'JOB20230415',
    },
    matchScore: 84,
  },
  {
    id: '11',
    resume: {
      id: 'r11',
      name: '王涛',
      age: 29,
      education: '本科',
      experience: '5年经验',
    },
    job: {
      id: 'j1',
      title: '高级前端工程师',
      department: '技术部',
      jobId: 'JOB20230415',
    },
    matchScore: 86,
  },
  {
    id: '12',
    resume: {
      id: 'r12',
      name: '林芳',
      age: 27,
      education: '本科',
      experience: '3年经验',
    },
    job: {
      id: 'j1',
      title: '高级前端工程师',
      department: '技术部',
      jobId: 'JOB20230415',
    },
    matchScore: 81,
  },
  {
    id: '13',
    resume: {
      id: 'r13',
      name: '黄强',
      age: 34,
      education: '硕士',
      experience: '8年经验',
    },
    job: {
      id: 'j1',
      title: '高级前端工程师',
      department: '技术部',
      jobId: 'JOB20230415',
    },
    matchScore: 90,
  },
  {
    id: '14',
    resume: {
      id: 'r14',
      name: '徐丽',
      age: 26,
      education: '本科',
      experience: '2年经验',
    },
    job: {
      id: 'j1',
      title: '高级前端工程师',
      department: '技术部',
      jobId: 'JOB20230415',
    },
    matchScore: 75,
  },
  {
    id: '15',
    resume: {
      id: 'r15',
      name: '何敏',
      age: 31,
      education: '本科',
      experience: '6年经验',
    },
    job: {
      id: 'j1',
      title: '高级前端工程师',
      department: '技术部',
      jobId: 'JOB20230415',
    },
    matchScore: 87,
  },
];

/**
 * 获取匹配任务详情
 */
export function getMockMatchingTaskDetail(
  taskId: string,
  page: number = 1,
  pageSize: number = 10
): MatchingTaskDetail | null {
  const task = mockMatchingTasks.find((t) => t.id === taskId);
  if (!task) {
    return null;
  }

  const total = mockMatchResults.length;
  const totalPages = Math.ceil(total / pageSize);
  const start = (page - 1) * pageSize;
  const end = start + pageSize;
  const results = mockMatchResults.slice(start, end);

  return {
    ...task,
    results,
    resultsPagination: {
      current: page,
      pageSize,
      total,
      totalPages,
    },
  };
}
