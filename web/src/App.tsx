import { Navigate, Route, Routes, useLocation } from "react-router-dom";
import { useCurrentUser } from "./hooks/useAuth";
import { LoginPage } from "./pages/LoginPage";
import { StartPage } from "./pages/StartPage";
import { PipelinePage } from "./pages/PipelinePage";
import { AdminTemplatesPage } from "./pages/AdminTemplatesPage";
import { AdminBrandAssetsPage } from "./pages/AdminBrandAssetsPage";
import { AdminAgentSettingsPage } from "./pages/AdminAgentSettingsPage";
import { AppShell } from "./components/common/AppShell";
import { Spinner } from "./components/common/Spinner";

// Layout route: renders AppShell (Sidebar + animated Outlet, via
// RouteTransition) once per auth state, instead of once per matched route —
// Sidebar no longer unmounts/remounts on every navigation.
function RequireAuth({ adminOnly = false }: { adminOnly?: boolean }) {
  const { data: user, isLoading } = useCurrentUser();
  const location = useLocation();

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <Spinner label="Loading..." />
      </div>
    );
  }
  if (!user) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }
  if (adminOnly && user.role !== "admin") {
    return <Navigate to="/" replace />;
  }
  return <AppShell user={user} />;
}

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />

      <Route element={<RequireAuth />}>
        <Route index element={<StartPage />} />
        <Route path="projects/:id" element={<PipelinePage />} />
      </Route>

      <Route element={<RequireAuth adminOnly />}>
        <Route path="admin/templates" element={<AdminTemplatesPage />} />
        <Route path="admin/brand-assets" element={<AdminBrandAssetsPage />} />
        <Route path="admin/agent-settings" element={<AdminAgentSettingsPage />} />
      </Route>

      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
