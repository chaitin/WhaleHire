import { Resume, ResumeStatus } from '@/types/resume';

export const mockResumes: Resume[] = [
  {
    id: '1',
    name: '张三',
    email: 'zhangsan@example.com',
    phone: '13888885678',
    gender: '男',
    current_city: '北京',
    highest_education: '本科',
    years_experience: 3,
    status: ResumeStatus.PENDING,
    resume_file_url: '/files/张三_前端开发工程师_简历.pdf',
    user_id: 'user1',
    created_at: 1689408600000,
    updated_at: 1689408600000,
  },
  {
    id: '2',
    name: '李四',
    email: 'lisi@example.com',
    phone: '13999998765',
    gender: '女',
    current_city: '上海',
    highest_education: '硕士',
    years_experience: 5,
    status: ResumeStatus.PROCESSING,
    resume_file_url: '/files/李四_产品经理_简历.pdf',
    user_id: 'user2',
    created_at: 1689322200000,
    updated_at: 1689322200000,
  },
  {
    id: '3',
    name: '王五',
    email: 'wangwu@example.com',
    phone: '13777774321',
    gender: '男',
    current_city: '深圳',
    highest_education: '本科',
    years_experience: 4,
    status: ResumeStatus.FAILED,
    resume_file_url: '/files/王五_后端开发工程师_简历.pdf',
    error_message: '简历格式不支持',
    user_id: 'user3',
    created_at: 1689149100000,
    updated_at: 1689149100000,
  },
  {
    id: '4',
    name: '赵六',
    email: 'zhaoliu@example.com',
    phone: '13666669876',
    gender: '女',
    current_city: '杭州',
    highest_education: '本科',
    years_experience: 2,
    status: ResumeStatus.PROCESSING,
    resume_file_url: '/files/赵六_UI设计师_简历.pdf',
    user_id: 'user4',
    created_at: 1688976000000,
    updated_at: 1688976000000,
  },
  {
    id: '5',
    name: '孙七',
    email: 'sunqi@example.com',
    phone: '13555552345',
    gender: '男',
    current_city: '广州',
    highest_education: '硕士',
    years_experience: 6,
    status: ResumeStatus.COMPLETED,
    resume_file_url: '/files/孙七_数据分析师_简历.pdf',
    parsed_at: '2023-07-08 15:30:00',
    user_id: 'user5',
    created_at: 1688802600000,
    updated_at: 1688802600000,
  },
];

export const statusLabels: Record<string, string> = {
  [ResumeStatus.PENDING]: '待解析',
  [ResumeStatus.PROCESSING]: '解析中',
  [ResumeStatus.COMPLETED]: '待筛选', // 解析成功翻译为待筛选
  [ResumeStatus.FAILED]: '解析失败',
  [ResumeStatus.ARCHIVED]: '已归档',
};

export const positionOptions = [
  { value: 'all', label: '所有岗位' },
  { value: '前端开发工程师', label: '前端开发工程师' },
  { value: '后端开发工程师', label: '后端开发工程师' },
  { value: '产品经理', label: '产品经理' },
  { value: 'UI/UX设计师', label: 'UI/UX设计师' },
  { value: '数据分析师', label: '数据分析师' },
];

export const statusOptions = [
  { value: 'all', label: '所有状态' },
  { value: ResumeStatus.PROCESSING, label: '解析中' },
  { value: ResumeStatus.COMPLETED, label: '待筛选' }, // 解析成功翻译为待筛选
];