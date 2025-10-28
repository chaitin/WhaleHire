import {
  BrowserRouter as Router,
  Routes,
  Route,
  Navigate,
} from 'react-router-dom';
import { MainLayout } from '@/components/layout/main-layout';
import { DashboardPage } from '@/app/dashboard';
import { ResumeManagementPage } from '@/features/resume/pages/resume-management';
import JobProfilePage from '@/features/job/pages/job-profile';
import PlatformConfig from '@/features/notification/pages/platform-config';
import { IntelligentMatchingPage } from '@/features/matching/pages/intelligent-matching';
import AuditLogPage from '@/features/audit/pages/audit-log';
import { PlaceholderPage } from '@/components/common/placeholder-page';
import LoginPage from '@/features/auth/pages/login';
import RegisterPage from '@/features/auth/pages/register';
import OAuthCallbackPage from '@/features/auth/pages/oauth-callback';
import { AuthProvider } from '@/hooks/useAuth';
import { ErrorBoundary } from '@/components/common/error-boundary';
import { ToastContainer } from '@/ui/toast';
import '@/styles/globals.css';

function App() {
  return (
    <ErrorBoundary>
      <Router>
        <AuthProvider>
          <ToastContainer />
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/register" element={<RegisterPage />} />
            <Route path="/oauth/callback" element={<OAuthCallbackPage />} />
            <Route path="/" element={<MainLayout />}>
              <Route
                index
                element={<Navigate to="/resume-management" replace />}
              />
              <Route path="dashboard" element={<DashboardPage />} />
              <Route
                path="resume-management"
                element={<ResumeManagementPage />}
              />
              <Route path="job-profile" element={<JobProfilePage />} />
              <Route path="platform-config" element={<PlatformConfig />} />
              <Route
                path="intelligent-matching"
                element={<IntelligentMatchingPage />}
              />
              <Route
                path="position-management"
                element={
                  <PlaceholderPage
                    title="岗位管理"
                    description="管理和发布招聘岗位"
                  />
                }
              />
              <Route
                path="candidate"
                element={
                  <PlaceholderPage
                    title="候选人"
                    description="管理候选人信息"
                  />
                }
              />
              <Route
                path="interview-schedule"
                element={
                  <PlaceholderPage
                    title="面试安排"
                    description="安排和管理面试时间"
                  />
                }
              />
              <Route
                path="data-analysis"
                element={
                  <PlaceholderPage
                    title="数据分析"
                    description="查看招聘数据分析报告"
                  />
                }
              />
              <Route
                path="system-settings"
                element={
                  <PlaceholderPage
                    title="系统设置"
                    description="配置系统参数"
                  />
                }
              />
              <Route
                path="user-management"
                element={
                  <PlaceholderPage
                    title="用户管理"
                    description="管理系统用户"
                  />
                }
              />
              <Route path="operation-log" element={<AuditLogPage />} />
              <Route
                path="permission-management"
                element={
                  <PlaceholderPage
                    title="权限管理"
                    description="配置用户权限"
                  />
                }
              />
            </Route>
          </Routes>
        </AuthProvider>
      </Router>
    </ErrorBoundary>
  );
}

export default App;
