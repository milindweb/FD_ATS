import { useEffect, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { GetFD, PreviewFD, RenewFD, errorMessage, type Calculation } from "../lib/api";
import { formatDate, formatMoney, formatPercent, formatTenure, todayISO } from "../lib/format";
import { TENURE_PRESETS, activeTenurePreset, tenurePresetDays } from "../lib/tenure";
import { useAsync } from "../lib/hooks";
import { Alert } from "../components/ui/Alert";
import { Button } from "../components/ui/Button";
import { Card, CardBody, CardHead } from "../components/ui/Card";
import { FormField, Input, Textarea } from "../components/ui/FormField";
import { ErrorState, Loading } from "../components/ui/States";
import { PageHeader } from "../components/ui/PageHeader";

const MODES = [
  { value: "PRINCIPAL_ONLY", label: "Principal only", hint: "Renew with the deposit amount; paid-out interest goes to the customer." },
  { value: "PRINCIPAL_PLUS_INTEREST", label: "Principal + Interest", hint: "Interest earned is added to the renewed deposit." },
] as const;

type Mode = (typeof MODES)[number]["value"];

export function RenewPage() {
  const { fdNumber = "" } = useParams();
  const navigate = useNavigate();
  const detail = useAsync(() => GetFD(fdNumber), [fdNumber]);

  const [mode, setMode] = useState<Mode>("PRINCIPAL_PLUS_INTEREST");
  const [startDate, setStartDate] = useState(todayISO());
  const [tenureDays, setTenureDays] = useState(() => String(tenurePresetDays(todayISO(), 36)));
  const [remark, setRemark] = useState("");
  const [preview, setPreview] = useState<Calculation | null>(null);
  const [previewError, setPreviewError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fd = detail.data?.fd;
  const memberId = fd?.memberId ?? 0;
  const basePrincipal = useMemo(() => {
    if (!fd) return 0;
    return mode === "PRINCIPAL_PLUS_INTEREST" ? fd.maturityAmount : fd.principal;
  }, [fd, mode]);

  useEffect(() => {
    if (!fd) return;
    // A missed renewal must start where the previous FD ended. Defaulting to
    // today would silently skip the period between maturity and today.
    setStartDate(fd.maturityDate);
    setTenureDays(String(tenurePresetDays(fd.maturityDate, 36)));
  }, [fd]);

  // Preview requires a linked member; legacy FDs without one skip it entirely.
  useEffect(() => {
    if (!fd || memberId <= 0) {
      setPreview(null);
      setPreviewError(null);
      return;
    }
    let cancelled = false;
    const id = setTimeout(() => {
      PreviewFD({
        memberId,
        fdFormNo: "",
        principal: basePrincipal,
        startDate,
        tenureDays: Number(tenureDays) || 0,
      })
        .then((calc) => {
          if (!cancelled) {
            setPreview(calc);
            setPreviewError(null);
          }
        })
        .catch((err: unknown) => {
          if (!cancelled) {
            setPreview(null);
            setPreviewError(errorMessage(err));
          }
        });
    }, 250);
    return () => {
      cancelled = true;
      clearTimeout(id);
    };
  }, [fd, memberId, basePrincipal, startDate, tenureDays]);

  if (detail.loading) return <Loading />;
  if (detail.error || !fd) return <ErrorState message={detail.error || "Fixed Deposit not found."} onRetry={detail.reload} />;
  if (fd.status !== "ACTIVE") {
    return (
      <>
        <PageHeader title="Renew Fixed Deposit" context={fd.fdNumber} />
        <Alert tone="warning" title="This FD cannot be renewed">
          Only active Fixed Deposits can be renewed. This one is {fd.status.toLowerCase()}.
        </Alert>
        <div className="u-mt-3">
          <Button onClick={() => navigate(`/fd/${fd.fdNumber}`)}>Back to FD</Button>
        </div>
      </>
    );
  }

  const tenure = Number(tenureDays) || 0;
  const notMatured = todayISO() < fd.maturityDate;
  const beforeMaturity = startDate !== "" && startDate < fd.maturityDate;
  const valid = basePrincipal > 0 && startDate !== "" && tenure > 0 && !beforeMaturity && !notMatured;

  // Keep a chosen year/month preset in sync when the start date changes, so
  // "3 years" still lands on the anniversary (1,095 or 1,096 days).
  const changeStartDate = (value: string) => {
    const preset = activeTenurePreset(startDate, tenureDays);
    setStartDate(value);
    if (preset) setTenureDays(String(tenurePresetDays(value, preset.months)));
  };

  const submit = async () => {
    setError(null);
    if (!valid) {
      setError("Enter a valid start date and tenure.");
      return;
    }
    setSaving(true);
    try {
      const result = await RenewFD({ fdNumber: fd.fdNumber, mode, startDate, tenureDays: tenure, remark });
      navigate(`/fd/${result.newFd.fdNumber}`, { state: { created: true } });
    } catch (err: unknown) {
      setError(errorMessage(err));
      setSaving(false);
    }
  };

  return (
    <>
      <PageHeader
        title="Renew Fixed Deposit"
        context={`${fd.fdNumber} → new FD number`}
        actions={<Button onClick={() => navigate(`/fd/${fd.fdNumber}`)}>Cancel</Button>}
      />

      {error && <Alert tone="danger" title="Could not renew the Fixed Deposit">{error}</Alert>}

      {notMatured && (
        <Alert tone="warning" title="Renewal not yet available">
          This FD matures on {formatDate(fd.maturityDate)}. Renewal becomes available on or after that date.
        </Alert>
      )}

      <div className="hs-two-col">
        <Card>
          <CardHead title="Renewal Terms" />
          <CardBody>
            <div className="hs-field-grid">
              <FormField label="Renewal Mode" required htmlFor="renew-mode" hint={MODES.find((m) => m.value === mode)?.hint}>
                <div className="hs-segmented hs-segmented--stack" role="group" aria-label="Renewal mode" id="renew-mode">
                  {MODES.map((m) => (
                    <button
                      key={m.value}
                      type="button"
                      className={`hs-segmented__btn ${mode === m.value ? "is-active" : ""}`}
                      aria-pressed={mode === m.value}
                      onClick={() => setMode(m.value)}
                    >
                      {m.label}
                    </button>
                  ))}
                </div>
              </FormField>

              <FormField
                label="New Start Date"
                required
                htmlFor="renew-start"
                error={beforeMaturity ? "Renewal cannot start before the FD maturity date." : null}
              >
                <Input
                  id="renew-start"
                  type="date"
                  value={startDate}
                  min={fd.maturityDate}
                  invalid={beforeMaturity}
                  onChange={(e) => changeStartDate(e.target.value)}
                />
              </FormField>

              <FormField label="New Tenure (days)" required htmlFor="renew-tenure">
                <Input
                  id="renew-tenure"
                  type="number"
                  min={1}
                  step={1}
                  value={tenureDays}
                  onChange={(e) => setTenureDays(e.target.value)}
                />
              </FormField>

              <FormField label="Remark" hint="Optional note stored in the history" htmlFor="renew-remark">
                <Textarea id="renew-remark" rows={3} value={remark} onChange={(e) => setRemark(e.target.value)} />
              </FormField>

              <div className="hs-tenure-presets" role="group" aria-label="Tenure presets">
                {TENURE_PRESETS.map((preset) => {
                  const days = tenurePresetDays(startDate, preset.months);
                  return (
                    <Button
                      key={preset.months}
                      size="sm"
                      variant={tenureDays === String(days) ? "primary" : "secondary"}
                      onClick={() => setTenureDays(String(days))}
                    >
                      {preset.label}
                    </Button>
                  );
                })}
              </div>
            </div>
          </CardBody>
        </Card>

        <Card>
          <CardHead title="Renewal Preview" />
          <CardBody>
            <dl className="hs-preview">
              <div className="hs-preview__row">
                <dt>Previous FD</dt>
                <dd className="hs-mono">{fd.fdNumber}</dd>
              </div>
              <div className="hs-preview__row">
                <dt>Renewing From</dt>
                <dd>{formatMoney(basePrincipal)}</dd>
              </div>
              <div className="hs-preview__row">
                <dt>Paid to customer</dt>
                <dd>{formatMoney(mode === "PRINCIPAL_PLUS_INTEREST" ? 0 : fd.interestAmount)}</dd>
              </div>
            </dl>

            {memberId <= 0 ? (
              <p className="u-mt-3 u-muted">Calculation preview unavailable — FD is not linked to a member</p>
            ) : previewError ? (
              <Alert tone="warning">{previewError}</Alert>
            ) : preview ? (
              <dl className="hs-preview u-mt-3">
                <div className="hs-preview__row">
                  <dt>Interest Rate</dt>
                  <dd>{formatPercent(preview.ratePercent)}</dd>
                </div>
                <div className="hs-preview__row">
                  <dt>Maturity Date</dt>
                  <dd>
                    {formatDate(preview.maturityDate)} · {formatTenure(preview.tenureDays)}
                  </dd>
                </div>
                <div className="hs-preview__row">
                  <dt>Interest</dt>
                  <dd>{formatMoney(preview.interest)}</dd>
                </div>
                <div className="hs-preview__row hs-preview__row--total">
                  <dt>New Maturity Amount</dt>
                  <dd>{formatMoney(preview.maturityAmount)}</dd>
                </div>
              </dl>
            ) : (
              <p className="u-mt-3 u-muted">Set the terms to preview the new deposit.</p>
            )}

            <div className="u-mt-4">
              <Button variant="primary" size="lg" block onClick={submit} loading={saving} disabled={!valid || previewError !== null}>
                Confirm Renewal
              </Button>
            </div>
            <p className="u-mt-2 u-text-xs u-muted">
              The current FD is closed as RENEWED and a new FD number is issued.
            </p>
          </CardBody>
        </Card>
      </div>
    </>
  );
}
