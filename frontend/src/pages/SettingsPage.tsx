import { useState } from "react";
import {
  BackupDatabase,
  ChangeCredentials,
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
import { ErrorState, Loading } from "../components/ui/States";
import { PageHeader } from "../components/ui/PageHeader";

export function SettingsPage() {
  const slabs = useAsync(() => GetRateSlabs(), []);
  const [draft, setDraft] = useState<RateSlab[] | null>(null);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<{ tone: "success" | "danger"; text: string } | null>(null);

  const rows = draft ?? slabs.data ?? [];

  const authInfo = useAsync(() => GetAuthInfo(), []);
  const [curPass, setCurPass] = useState("");
  const [credUser, setCredUser] = useState("");
  const [credPass, setCredPass] = useState("");
  const [credPass2, setCredPass2] = useState("");
  const [credBusy, setCredBusy] = useState(false);

  const changeCredentials = async () => {
    setMessage(null);
    if (!credPass && !credUser.trim()) {
      setMessage({ tone: "danger", text: "Enter a new username or password." });
      return;
    }
    if (credPass && credPass.length < 4) {
      setMessage({ tone: "danger", text: "Password must be at least 4 characters." });
      return;
    }
    if (credPass !== credPass2) {
      setMessage({ tone: "danger", text: "Passwords do not match." });
      return;
    }
    setCredBusy(true);
    try {
      await ChangeCredentials({
        currentPassword: curPass,
        newUsername: credUser.trim(),
        newPassword: credPass,
      });
      setCurPass("");
      setCredUser("");
      setCredPass("");
      setCredPass2("");
      authInfo.reload();
      setMessage({ tone: "success", text: "Login credentials updated. The recovery file has been refreshed." });
    } catch (err: unknown) {
      setMessage({ tone: "danger", text: errorMessage(err) });
    } finally {
      setCredBusy(false);
    }
  };

  const clearCredForm = () => {
    setCurPass("");
    setCredUser("");
    setCredPass("");
    setCredPass2("");
    setMessage(null);
  };

  const [restorePath, setRestorePath] = useState("");
  const [confirmRestore, setConfirmRestore] = useState(false);
  const [confirmSample, setConfirmSample] = useState(false);
  const [dataBusy, setDataBusy] = useState(false);

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
        }
      />

      {message && <Alert tone={message.tone}>{message.text}</Alert>}

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

      <Card className="u-mt-5">
        <CardHead
          title="Login Credentials"
          actions={<span className="u-text-xs u-muted">Minimum 4 characters</span>}
        />
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

          <div className="hs-field-grid u-mt-4">
            <FormField label="Current password" htmlFor="cred-current" required>
              <Input
                id="cred-current"
                type="password"
                autoComplete="current-password"
                value={curPass}
                onChange={(e) => setCurPass(e.target.value)}
              />
            </FormField>
            <FormField label="New username" htmlFor="cred-user" hint="Leave empty to keep">
              <Input
                id="cred-user"
                autoComplete="username"
                value={credUser}
                onChange={(e) => setCredUser(e.target.value)}
              />
            </FormField>
            <FormField label="New password" htmlFor="cred-pass" hint="Leave empty to keep · min 4 characters">
              <Input
                id="cred-pass"
                type="password"
                autoComplete="new-password"
                value={credPass}
                onChange={(e) => setCredPass(e.target.value)}
              />
            </FormField>
            <FormField label="Confirm new password" htmlFor="cred-pass2">
              <Input
                id="cred-pass2"
                type="password"
                autoComplete="new-password"
                value={credPass2}
                onChange={(e) => setCredPass2(e.target.value)}
              />
            </FormField>
          </div>

          <div className="u-mt-4" style={{ display: "flex", gap: "var(--space-3)" }}>
            <Button variant="primary" onClick={changeCredentials} loading={credBusy}>
              Update credentials
            </Button>
            <Button onClick={clearCredForm} disabled={credBusy}>
              Clear
            </Button>
          </div>
        </CardBody>
      </Card>

      <div className="hs-two-col u-mt-5">
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
