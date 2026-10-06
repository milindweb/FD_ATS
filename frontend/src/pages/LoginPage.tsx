import { useEffect, useState, type FormEvent } from "react";
import { Login, ResetCredentials, SystemStatus, errorMessage } from "../lib/api";
import { brand } from "../lib/brandInfo";
import { Alert } from "../components/ui/Alert";
import { Button } from "../components/ui/Button";
import { Card, CardBody } from "../components/ui/Card";
import { FormField, Input } from "../components/ui/FormField";
import { ErrorState, Loading } from "../components/ui/States";
import aartiLogo from "../assets/images/logo-ats.svg";

// Mirrors the backend's exact messages for client-side pre-validation.
const MSG_PASSWORD_SHORT = "Password must be at least 4 characters.";
const MSG_PASSWORDS_DIFFER = "Passwords do not match.";
const MSG_NOTHING_TO_CHANGE = "Enter a new username or password.";

export interface LoginPageProps {
  onSuccess: () => void;
}

/** Full-screen sign-in (design.md §6 auth pattern). No app shell around it. */
export function LoginPage({ onSuccess }: LoginPageProps) {
  const [status, setStatus] = useState<{ ready: boolean; error: string } | null>(null);

  const [mode, setMode] = useState<"login" | "reset">("login");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);

  const [code, setCode] = useState("");
  const [newUser, setNewUser] = useState("");
  const [newPass, setNewPass] = useState("");
  const [confirmPass, setConfirmPass] = useState("");

  useEffect(() => {
    let cancelled = false;
    const check = () =>
      SystemStatus()
        .then((s) => {
          if (!cancelled) {
            setStatus({ ready: s.ready, error: s.error });
            if (!s.ready && !s.error) setTimeout(check, 700);
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

  const signIn = async (event: FormEvent) => {
    event.preventDefault();
    setError(null);
    setBusy(true);
    try {
      await Login({ username, password });
      onSuccess();
    } catch (err: unknown) {
      setError(errorMessage(err));
    } finally {
      setBusy(false);
    }
  };

  const reset = async (event: FormEvent) => {
    event.preventDefault();
    setError(null);
    if (newPass.length < 4) {
      setError(MSG_PASSWORD_SHORT);
      return;
    }
    if (newPass !== confirmPass) {
      setError(MSG_PASSWORDS_DIFFER);
      return;
    }
    setBusy(true);
    try {
      const info = await ResetCredentials({
        recoveryCode: code,
        newUsername: newUser,
        newPassword: newPass,
      });
      setMode("login");
      setUsername(info.username);
      setPassword("");
      setCode("");
      setNewUser("");
      setNewPass("");
      setConfirmPass("");
      setNotice("Credentials reset. Sign in with your new username and password.");
    } catch (err: unknown) {
      setError(errorMessage(err));
    } finally {
      setBusy(false);
    }
  };

  if (status === null || (!status.ready && !status.error)) {
    return (
      <div className="hs-auth">
        <Loading label="Starting the application…" />
      </div>
    );
  }
  if (!status.ready) {
    return (
      <div className="hs-auth">
        <ErrorState message={status.error} onRetry={() => window.location.reload()} />
      </div>
    );
  }

  return (
    <div className="hs-auth">
      <Card className="hs-auth__card">
        <CardBody>
          <img className="hs-auth__logo" src={aartiLogo} alt="Aarti Tech Services" />
          <h1 className="hs-auth__title">{brand.product}</h1>
          <p className="hs-auth__sub">
            {mode === "login" ? "Sign in to continue" : "Reset your username and password"}
          </p>

          {error && <Alert tone="danger">{error}</Alert>}
          {notice && mode === "login" && <Alert tone="success">{notice}</Alert>}

          {mode === "login" ? (
            <form onSubmit={signIn}>
              <FormField label="Username" htmlFor="auth-username">
                <Input
                  id="auth-username"
                  value={username}
                  autoComplete="username"
                  autoFocus
                  onChange={(e) => setUsername(e.target.value)}
                />
              </FormField>
              <FormField label="Password" htmlFor="auth-password">
                <Input
                  id="auth-password"
                  type="password"
                  value={password}
                  autoComplete="current-password"
                  onChange={(e) => setPassword(e.target.value)}
                />
              </FormField>
              <Button type="submit" variant="primary" size="lg" block loading={busy}>
                Sign in
              </Button>
            </form>
          ) : (
            <form onSubmit={reset}>
              <FormField label="Recovery code" htmlFor="auth-code" hint="From FD_ATS_RECOVERY.txt">
                <Input
                  id="auth-code"
                  value={code}
                  placeholder="FD-XXXX-XXXX"
                  autoFocus
                  onChange={(e) => setCode(e.target.value)}
                />
              </FormField>
              <FormField label="New username" htmlFor="auth-newuser">
                <Input
                  id="auth-newuser"
                  value={newUser}
                  autoComplete="username"
                  onChange={(e) => setNewUser(e.target.value)}
                />
              </FormField>
              <FormField label="New password" htmlFor="auth-newpass" hint="Minimum 4 characters">
                <Input
                  id="auth-newpass"
                  type="password"
                  value={newPass}
                  autoComplete="new-password"
                  onChange={(e) => setNewPass(e.target.value)}
                />
              </FormField>
              <FormField label="Confirm password" htmlFor="auth-confirm">
                <Input
                  id="auth-confirm"
                  type="password"
                  value={confirmPass}
                  autoComplete="new-password"
                  onChange={(e) => setConfirmPass(e.target.value)}
                />
              </FormField>
              <Button type="submit" variant="primary" size="lg" block loading={busy}>
                Reset credentials
              </Button>
            </form>
          )}

          <p className="hs-auth__foot">
            {mode === "login" ? (
              <button
                type="button"
                className="hs-linklike"
                onClick={() => {
                  setMode("reset");
                  setError(null);
                  setNotice(null);
                }}
              >
                Forgot password?
              </button>
            ) : (
              <button
                type="button"
                className="hs-linklike"
                onClick={() => {
                  setMode("login");
                  setError(null);
                }}
              >
                Back to sign in
              </button>
            )}
          </p>
        </CardBody>
      </Card>
    </div>
  );
}
