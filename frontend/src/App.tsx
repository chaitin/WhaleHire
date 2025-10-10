import {
  BrowserRouter as Router,
  Routes,
  Route,
  Navigate,
} from 'react-router-dom';
import { MainLayout } from '@/components/layout/main-layout';
import { DashboardPage } from '@/pages/dashboard';
import { ResumeManagementPage } from '@/pages/resume-management';
import { PlaceholderPage } from '@/pages/placeholder-page';
import LoginPage from '@/pages/login';
import RegisterPage from '@/pages/register';
import OAuthCallbackPage from '@/pages/oauth-callback';
import { AuthProvider } from '@/hooks/useAuth';
import { ErrorBoundary } from '@/components/error-boundary';
import '@/styles/globals.css';

function App() {
  return (
    <ErrorBoundary>
      <Router>
        <AuthProvider>
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
