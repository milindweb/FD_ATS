import { useState } from "react";
import {
  BackupDatabase,
  ChangePassword,
  ChangeUsername,
  GetAuthInfo,
  GetRateSlabs,
  LoadSampleData,
  PickBackupPath,
  PickRestorePath,
  RestoreDatabase,
  SaveSlabs,
  errorMessage,
  type RateSlab,
} from "../lib/api";
import { formatPercent } from "../lib/format";
import { useAsync } from "../lib/hooks";
import { Alert } from "../components/ui/Alert";
import { Button } from "../components/ui/Button";
import { Card, CardBody, CardHead } from "../components/ui/Card";
import { ConfirmDialog } from "../components/ui/Modal";
import { FormField, Input } from "../components/ui/FormField";
import { SegmentedControl, type SegmentedOption } from "../components/ui/SegmentedControl";
import { ErrorState, Loading } from "../components/ui/States";
import { PageHeader } from "../components/ui/PageHeader";

type SettingsTab = "slabs" | "credentials" | "backup";

const SETTINGS_TABS: ReadonlyArray<SegmentedOption<SettingsTab>> = [
  { value: "slabs", label: "Slab Rates" },
  { value: "credentials", label: "Login Credentials" },
  { value: "backup", label: "Backup & Restore" },
];

interface CardMessage {
  tone: "success" | "danger";
  text: string;
  recoveryCode?: string;
}

function CredentialAlert({ msg }: { msg: CardMessage | null }) {
  if (!msg) return null;
  return (
    <Alert tone={msg.tone}>
      {msg.text}
      {msg.recoveryCode && (
        <div
          className="u-mt-2 u-flex-wrap"
          style={{ display: "flex", gap: "var(--space-3)", alignItems: "center" }}
        >
          <span className="hs-mono">{msg.recoveryCode}</span>
          <Button size="sm" onClick={() => void navigator.clipboard.writeText(msg.recoveryCode ?? "")}>
            Copy
          </Button>
        </div>
      )}
    </Alert>
  );
}

export function SettingsPage() {
  const [tab, setTab] = useState<SettingsTab>("slabs");

  const slabs = useAsync(() => GetRateSlabs(), []);
  const [draft, setDraft] = useState<RateSlab[] | null>(null);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<{ tone: "success" | "danger"; text: string } | null>(null);

  const rows = draft ?? slabs.data ?? [];

  const authInfo = useAsync(() => GetAuthInfo(), []);

  const [userCurrentPassword, setUserCurrentPassword] = useState("");
  const [newUsername, setNewUsername] = useState("");
  const [usernameBusy, setUsernameBusy] = useState(false);
  const [usernameMsg, setUsernameMsg] = useState<CardMessage | null>(null);

  const [passCurrentPassword, setPassCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [passwordBusy, setPasswordBusy] = useState(false);
  const [passwordMsg, setPasswordMsg] = useState<CardMessage | null>(null);

  const [restorePath, setRestorePath] = useState("");
  const [confirmRestore, setConfirmRestore] = useState(false);
  const [confirmSample, setConfirmSample] = useState(false);
  const [dataBusy, setDataBusy] = useState(false);

  const switchTab = (next: SettingsTab) => {
    setMessage(null);
    setTab(next);
  };

  const changeUsername = async () => {
    setUsernameMsg(null);
    if (!userCurrentPassword) {
      setUsernameMsg({ tone: "danger", text: "Enter your current password." });
      return;
    }
    if (!newUsername.trim()) {
      setUsernameMsg({ tone: "danger", text: "Enter a new username." });
      return;
    }
    setUsernameBusy(true);
    try {
      const result = await ChangeUsername({
        currentPassword: userCurrentPassword,
        newUsername: newUsername.trim(),
      });
      setUserCurrentPassword("");
      setNewUsername("");
      authInfo.reload();
      setUsernameMsg({
        tone: "success",
        text: "Username updated. A fresh recovery code was generated — the old recovery code no longer works.",
        recoveryCode: result.recoveryCode,
      });
    } catch (err: unknown) {
      setUsernameMsg({ tone: "danger", text: errorMessage(err) });
    } finally {
      setUsernameBusy(false);
    }
  };

  const clearUsernameForm = () => {
    setUserCurrentPassword("");
    setNewUsername("");
    setUsernameMsg(null);
  };

  const changePassword = async () => {
    setPasswordMsg(null);
    if (!passCurrentPassword) {
      setPasswordMsg({ tone: "danger", text: "Enter your current password." });
      return;
    }
    if (newPassword.length < 4) {
      setPasswordMsg({ tone: "danger", text: "Password must be at least 4 characters." });
      return;
    }
    if (newPassword !== confirmPassword) {
      setPasswordMsg({ tone: "danger", text: "Passwords do not match." });
      return;
    }
    setPasswordBusy(true);
    try {
      const result = await ChangePassword({
        currentPassword: passCurrentPassword,
        newPassword,
      });
      setPassCurrentPassword("");
      setNewPassword("");
      setConfirmPassword("");
      authInfo.reload();
      setPasswordMsg({
        tone: "success",
        text: "Password updated. A fresh recovery code was generated — the old recovery code no longer works.",
        recoveryCode: result.recoveryCode,
      });
    } catch (err: unknown) {
      setPasswordMsg({ tone: "danger", text: errorMessage(err) });
    } finally {
      setPasswordBusy(false);
    }
  };

  const clearPasswordForm = () => {
    setPassCurrentPassword("");
    setNewPassword("");
    setConfirmPassword("");
    setPasswordMsg(null);
  };

  const backup = async () => {
    setMessage(null);
    try {
      const path = await PickBackupPath();
      if (!path) return;
      const saved = await BackupDatabase(path);
      setMessage({ tone: "success", text: `Backup saved to ${saved}` });
    } catch (err: unknown) {
      setMessage({ tone: "danger", text: errorMessage(err) });
    }
  };

  const pickRestore = async () => {
    setMessage(null);
    try {
      const path = await PickRestorePath();
      if (!path) return;
      setRestorePath(path);
      setConfirmRestore(true);
    } catch (err: unknown) {
      setMessage({ tone: "danger", text: errorMessage(err) });
    }
  };

  const restore = async () => {
    setConfirmRestore(false);
    setDataBusy(true);
    try {
      await RestoreDatabase(restorePath);
      setRestorePath("");
      setDraft(null);
      slabs.reload();
      setMessage({ tone: "success", text: "Backup restored. The application now shows the restored data." });
    } catch (err: unknown) {
      setMessage({ tone: "danger", text: errorMessage(err) });
    } finally {
      setDataBusy(false);
    }
  };

  const loadSample = async () => {
    setConfirmSample(false);
    setDataBusy(true);
    try {
      const count = await LoadSampleData();
      setMessage({ tone: "success", text: `Sample data ready — ${count} FDs stored.` });
    } catch (err: unknown) {
      setMessage({ tone: "danger", text: errorMessage(err) });
    } finally {
      setDataBusy(false);
    }
  };

  const updateRow = (index: number, patch: Partial<RateSlab>) => {
    setMessage(null);
    setDraft(rows.map((row, i) => (i === index ? { ...row, ...patch } : row)));
  };

  const save = async () => {
    setMessage(null);
    if (!draft) return;
    for (const row of draft) {
      if (row.minDays < 0 || (row.maxDays > 0 && row.maxDays < row.minDays)) {
        setMessage({ tone: "danger", text: "Each slab needs a valid day range (max ≥ min, or 0 for open-ended)." });
        return;
      }
      if (row.ratePercent < 0 || row.ratePercent > 100) {
        setMessage({ tone: "danger", text: "Interest rates must be between 0 and 100 percent." });
        return;
      }
    }
    setSaving(true);
    try {
      await SaveSlabs(draft.map((row, i) => ({ ...row, sortOrder: i + 1 })));
      setDraft(null);
      slabs.reload();
      setMessage({ tone: "success", text: "Interest rate slabs saved." });
    } catch (err: unknown) {
      setMessage({ tone: "danger", text: errorMessage(err) });
    } finally {
      setSaving(false);
    }
  };

  if (slabs.loading) return <Loading />;
  if (slabs.error) return <ErrorState message={slabs.error} onRetry={slabs.reload} />;

  return (
    <>
      <PageHeader
        title="Settings"
        context="Interest rate slabs, database backup and demo data"
        actions={
          tab === "slabs" ? (
            <>
              {draft && (
                <Button onClick={() => setDraft(null)} disabled={saving}>
                  Discard changes
                </Button>
              )}
              <Button variant="primary" onClick={save} loading={saving} disabled={!draft}>
                Save Changes
              </Button>
            </>
          ) : null
        }
      />

      {message && <Alert tone={message.tone}>{message.text}</Alert>}

      <div className="u-mb-4">
        <SegmentedControl label="Settings sections" options={SETTINGS_TABS} value={tab} onChange={switchTab} />
      </div>

      {tab === "slabs" && (
        <Card>
          <CardHead title="Rate Slabs" actions={<span className="u-text-xs u-muted">0 for max days means open-ended</span>} />
          <CardBody>
            <div className="hs-table-wrap is-responsive">
              <table className="hs-table">
                <thead>
                  <tr>
                    <th scope="col">Label</th>
                    <th scope="col">Min Days</th>
                    <th scope="col">Max Days</th>
                    <th scope="col">Rate (%)</th>
                  </tr>
                </thead>
                <tbody>
                  {rows.map((row, index) => (
                    <tr key={row.id || index}>
                      <td data-label="Label">
                        <FormField label="Label" htmlFor={`slab-label-${index}`}>
                          <Input
                            id={`slab-label-${index}`}
                            value={row.label}
                            onChange={(e) => updateRow(index, { label: e.target.value })}
                          />
                        </FormField>
                      </td>
                      <td data-label="Min Days" className="hs-num">
                        <FormField label="Min Days" htmlFor={`slab-min-${index}`}>
                          <Input
                            id={`slab-min-${index}`}
                            type="number"
                            min={0}
                            value={row.minDays}
                            onChange={(e) => updateRow(index, { minDays: Number(e.target.value) })}
                          />
                        </FormField>
                      </td>
                      <td data-label="Max Days" className="hs-num">
                        <FormField label="Max Days" htmlFor={`slab-max-${index}`}>
                          <Input
                            id={`slab-max-${index}`}
                            type="number"
                            min={0}
                            value={row.maxDays}
                            onChange={(e) => updateRow(index, { maxDays: Number(e.target.value) })}
                          />
                        </FormField>
                      </td>
                      <td data-label="Rate (%)" className="hs-num">
                        <FormField label="Rate (%)" htmlFor={`slab-rate-${index}`}>
                          <Input
                            id={`slab-rate-${index}`}
                            type="number"
                            min={0}
                            max={100}
                            step={0.05}
                            value={row.ratePercent}
                            onChange={(e) => updateRow(index, { ratePercent: Number(e.target.value) })}
                          />
                        </FormField>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            <div className="u-mt-3">
              <Button
                size="sm"
                onClick={() => {
                  setMessage(null);
                  setDraft([
                    ...rows,
                    {
                      id: 0,
                      sortOrder: rows.length + 1,
                      minDays: rows.length ? rows[rows.length - 1].maxDays + 1 : 0,
                      maxDays: 0,
                      ratePercent: 0,
                      label: "New slab",
                    },
                  ]);
                }}
              >
                Add slab
              </Button>
            </div>

            <p className="u-mt-3 u-text-xs u-muted">
              Current selection rules: {rows.map((r) => `${r.minDays}–${r.maxDays || "∞"} days = ${formatPercent(r.ratePercent)}`).join(" · ")}
            </p>
          </CardBody>
        </Card>
      )}

      {tab === "credentials" && (
        <>
          <Card>
            <CardHead title="Current Login" actions={<span className="u-text-xs u-muted">Signed-in account</span>} />
            <CardBody>
              <div className="hs-auth-info">
                <div className="hs-auth-info__item">
                  <div className="hs-auth-info__label">Signed-in username</div>
                  <p className="hs-auth-info__value">{authInfo.data?.username || "—"}</p>
                </div>
                <div className="hs-auth-info__item">
                  <div className="hs-auth-info__label">Recovery code</div>
                  <p className="hs-auth-info__value hs-mono">{authInfo.data?.recoveryCode || "—"}</p>
                </div>
              </div>
              <p className="u-muted">
                Keep the recovery code safe — it is also written to FD_ATS_RECOVERY.txt in the data folder
                and lets you reset a forgotten login on the sign-in screen.
              </p>
            </CardBody>
          </Card>

          <div className="hs-two-col u-mt-5">
            <Card>
              <CardHead
                title="Change Username"
                actions={<span className="u-text-xs u-muted">Current password required</span>}
              />
              <CardBody>
                <CredentialAlert msg={usernameMsg} />
                <div className="hs-field-grid">
                  <FormField label="Current password" htmlFor="user-current" required>
                    <Input
                      id="user-current"
                      type="password"
                      autoComplete="current-password"
                      value={userCurrentPassword}
                      onChange={(e) => setUserCurrentPassword(e.target.value)}
                    />
                  </FormField>
                  <FormField label="New username" htmlFor="user-name" required>
                    <Input
                      id="user-name"
                      autoComplete="username"
                      value={newUsername}
                      onChange={(e) => setNewUsername(e.target.value)}
                    />
                  </FormField>
                </div>
                <div className="u-mt-4" style={{ display: "flex", gap: "var(--space-3)" }}>
                  <Button variant="primary" onClick={changeUsername} loading={usernameBusy}>
                    Change username
                  </Button>
                  <Button onClick={clearUsernameForm} disabled={usernameBusy}>
                    Clear
                  </Button>
                </div>
              </CardBody>
            </Card>

            <Card>
              <CardHead
                title="Change Password"
                actions={<span className="u-text-xs u-muted">Minimum 4 characters</span>}
              />
              <CardBody>
                <CredentialAlert msg={passwordMsg} />
                <div className="hs-field-grid">
                  <FormField label="Current password" htmlFor="pass-current" required>
                    <Input
                      id="pass-current"
                      type="password"
                      autoComplete="current-password"
                      value={passCurrentPassword}
                      onChange={(e) => setPassCurrentPassword(e.target.value)}
                    />
                  </FormField>
                  <FormField label="New password" htmlFor="pass-new" required hint="Minimum 4 characters">
                    <Input
                      id="pass-new"
                      type="password"
                      autoComplete="new-password"
                      value={newPassword}
                      onChange={(e) => setNewPassword(e.target.value)}
                    />
                  </FormField>
                  <FormField label="Confirm new password" htmlFor="pass-confirm" required hint="Must match the new password">
                    <Input
                      id="pass-confirm"
                      type="password"
                      autoComplete="new-password"
                      value={confirmPassword}
                      onChange={(e) => setConfirmPassword(e.target.value)}
                    />
                  </FormField>
                </div>
                <div className="u-mt-4" style={{ display: "flex", gap: "var(--space-3)" }}>
                  <Button variant="primary" onClick={changePassword} loading={passwordBusy}>
                    Change password
                  </Button>
                  <Button onClick={clearPasswordForm} disabled={passwordBusy}>
                    Clear
                  </Button>
                </div>
              </CardBody>
            </Card>
          </div>
        </>
      )}

      {tab === "backup" && (
        <div className="hs-two-col">
          <Card>
            <CardHead title="Backup & Restore" actions={<span className="u-text-xs u-muted">Database file</span>} />
            <CardBody>
              <p className="u-muted">
                Backups capture every FD, closure and rate setting. Restore replaces the current data with the
                chosen backup file.
              </p>
              <div className="u-mt-3 u-flex-wrap" style={{ display: "flex", gap: "var(--space-3)" }}>
                <Button icon="download" onClick={backup} loading={dataBusy}>
                  Backup database
                </Button>
                <Button icon="refresh" onClick={pickRestore} loading={dataBusy} variant="danger">
                  Restore from backup
                </Button>
              </div>
            </CardBody>
          </Card>

          <Card>
            <CardHead
              title="Sample Data"
              actions={<span className="u-text-xs u-muted">Temporary</span>}
            />
            <CardBody>
              <p className="u-muted">
                Loads a few demo FDs so the screens can be explored. Available only while the database is empty —
                real data will never be overwritten.
              </p>
              <div className="u-mt-3">
                <Button onClick={() => { setMessage(null); setConfirmSample(true); }} loading={dataBusy}>
                  Load sample data
                </Button>
              </div>
            </CardBody>
          </Card>
        </div>
      )}

      <ConfirmDialog
        open={confirmRestore}
        title="Restore backup?"
        message={
          <>
            The current database will be replaced with <strong>{restorePath}</strong>. A safety copy of the
            current data is kept until the restore succeeds.
          </>
        }
        confirmLabel="Restore"
        danger
        busy={dataBusy}
        onConfirm={restore}
        onCancel={() => setConfirmRestore(false)}
      />

      <ConfirmDialog
        open={confirmSample}
        title="Load sample data?"
        message="Demo FDs will be created in the empty database. They can be closed or replaced once real data arrives."
        confirmLabel="Load sample data"
        busy={dataBusy}
        onConfirm={loadSample}
        onCancel={() => setConfirmSample(false)}
      />
    </>
  );
}
