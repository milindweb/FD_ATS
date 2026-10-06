import { useEffect, useState, type ReactNode } from "react";
import { AppFooter } from "./AppFooter";
import { AppHeader } from "./AppHeader";
import { MobileNav } from "./MobileNav";
import { SystemStatus, errorMessage } from "../../lib/api";
import { ErrorState, Loading } from "../ui/States";

export interface AppShellProps {
  children: ReactNode;
  onLogout?: () => void;
}

/** Common header / main / footer layout for every screen (SRS §8). */
export function AppShell({ children, onLogout }: AppShellProps) {
  const [menuOpen, setMenuOpen] = useState(false);
  const [status, setStatus] = useState<{ ready: boolean; error: string } | null>(null);

  useEffect(() => {
    let cancelled = false;
    const check = () =>
      SystemStatus()
        .then((s) => {
          if (!cancelled) {
            setStatus({ ready: s.ready, error: s.error });
            if (!s.ready) setTimeout(check, 700);
          }
        })
        .catch((err: unknown) => {
          if (!cancelled) setStatus({ ready: false, error: errorMessage(err) });
        });
    check();
    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <div className="hs-shell">
      <AppHeader onMenuOpen={() => setMenuOpen(true)} onLogout={onLogout} />
      <MobileNav open={menuOpen} onClose={() => setMenuOpen(false)} />

      <main className="hs-main">
        {status === null || (!status.ready && !status.error) ? (
          <Loading label="Starting the application…" />
        ) : !status.ready ? (
          <ErrorState message={status.error} onRetry={() => window.location.reload()} />
        ) : (
          children
        )}
      </main>

      <AppFooter />
    </div>
  );
}
