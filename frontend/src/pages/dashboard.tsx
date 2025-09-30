import { BarChart3, Users, FileText, Calendar } from 'lucide-react';

export function DashboardPage() {
  const stats = [
    {
      title: '总简历数',
      value: '1,234',
      change: '+12%',
      icon: FileText,
      color: 'text-blue-600',
      bgColor: 'bg-blue-100',
    },
    {
      title: '待筛选',
      value: '89',
      change: '+5%',
      icon: Users,
      color: 'text-yellow-600',
      bgColor: 'bg-yellow-100',
    },
    {
      title: '已录用',
      value: '156',
      change: '+18%',
      icon: BarChart3,
      color: 'text-green-600',
      bgColor: 'bg-green-100',
    },
    {
      title: '面试安排',
      value: '23',
      change: '+3%',
      icon: Calendar,
      color: 'text-purple-600',
      bgColor: 'bg-purple-100',
    },
  ];

  return (
    <div className="flex flex-col h-full">
      {/* Page Header */}
      <div className="border-b bg-background px-6 py-4">
        <div>
          <h1 className="text-2xl font-semibold">仪表盘</h1>
          <p className="text-muted-foreground">查看系统概览和关键指标</p>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 space-y-6 p-6">
        {/* Stats Grid */}
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          {stats.map((stat) => {
            const Icon = stat.icon;
            return (
              <div
                key={stat.title}
                className="rounded-lg border bg-card p-6 shadow-sm"
              >
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">
                      {stat.title}
                    </p>
                    <p className="text-2xl font-bold">{stat.value}</p>
                    <p className="text-xs text-green-600">{stat.change}</p>
                  </div>
                  <div className={`rounded-full p-3 ${stat.bgColor}`}>
                    <Icon className={`h-6 w-6 ${stat.color}`} />
                  </div>
                </div>
              </div>
            );
          })}
        </div>

        {/* Placeholder for Charts */}
        <div className="grid gap-6 md:grid-cols-2">
          <div className="rounded-lg border bg-card p-6">
            <h3 className="text-lg font-semibold mb-4">简历投递趋势</h3>
            <div className="h-64 flex items-center justify-center text-muted-foreground">
              图表组件将在这里显示
            </div>
          </div>
          <div className="rounded-lg border bg-card p-6">
            <h3 className="text-lg font-semibold mb-4">岗位分布</h3>
            <div className="h-64 flex items-center justify-center text-muted-foreground">
              图表组件将在这里显示
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
