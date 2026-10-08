import { useEffect, useState } from "react";
import { HashRouter, Navigate, Route, Routes } from "react-router-dom";
import { AppShell } from "./components/layout/AppShell";
import { AboutPage } from "./pages/AboutPage";
import { ClosePage } from "./pages/ClosePage";
import { DashboardPage } from "./pages/DashboardPage";
import { EditFDPage } from "./pages/EditFDPage";
import { FDDetailsPage } from "./pages/FDDetailsPage";
import { FDMasterPage } from "./pages/FDMasterPage";
import { LoginPage } from "./pages/LoginPage";
import { MemberFormPage } from "./pages/MemberFormPage";
import { MemberListPage } from "./pages/MemberListPage";
import { MemberProfilePage } from "./pages/MemberProfilePage";
import { NewFDPage } from "./pages/NewFDPage";
import { PrivacyPage } from "./pages/PrivacyPage";
import { RenewPage } from "./pages/RenewPage";
import { ReportsPage } from "./pages/ReportsPage";
import { SettingsPage } from "./pages/SettingsPage";
import { Loading } from "./components/ui/States";
import { Logout, SystemStatus } from "./lib/api";

function App() {
  const [authed, setAuthed] = useState(false);
  const [checking, setChecking] = useState(true);

  // A frontend reload keeps the Go session alive — ask the backend first so a
  // refresh does not force a new login. Fresh launches still start signed out.
  useEffect(() => {
    SystemStatus()
      .then((s) => setAuthed(Boolean(s.ready && s.authed)))
      .catch(() => setAuthed(false))
      .finally(() => setChecking(false));
  }, []);

  if (checking) {
    return (
      <div className="hs-auth">
        <Loading label="Starting the application…" />
      </div>
    );
  }

  if (!authed) {
    return <LoginPage onSuccess={() => setAuthed(true)} />;
  }

  const handleLogout = () => {
    Logout();
    setAuthed(false);
  };

  return (
    <HashRouter>
      <AppShell onLogout={handleLogout}>
        <Routes>
          <Route path="/" element={<DashboardPage />} />
          <Route path="/fds" element={<FDMasterPage />} />
          <Route path="/member" element={<MemberListPage />} />
          <Route path="/member/new" element={<MemberFormPage mode="create" />} />
          <Route path="/member/:memberId/edit" element={<MemberFormPage mode="edit" />} />
          <Route path="/member/:memberId" element={<MemberProfilePage />} />
          <Route path="/new" element={<NewFDPage />} />
          <Route path="/fd/:fdNumber" element={<FDDetailsPage />} />
          <Route path="/fd/:fdNumber/edit" element={<EditFDPage />} />
          <Route path="/fd/:fdNumber/renew" element={<RenewPage />} />
          <Route path="/fd/:fdNumber/close" element={<ClosePage />} />
          <Route path="/reports" element={<ReportsPage />} />
          <Route path="/settings" element={<SettingsPage />} />
          <Route path="/about" element={<AboutPage />} />
          <Route path="/privacy" element={<PrivacyPage />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </AppShell>
    </HashRouter>
  );
}

export default App;
