import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { CloseFD, GetFD, PreviewClosure, errorMessage, type ClosurePreview } from "../lib/api";
import { formatDate, formatMoney, formatPercent, formatTenure, todayISO } from "../lib/format";
import { useAsync } from "../lib/hooks";
import { Alert } from "../components/ui/Alert";
import { Badge } from "../components/ui/Badge";
import { Button } from "../components/ui/Button";
import { Card, CardBody, CardHead } from "../components/ui/Card";
import { FormField, Input, Textarea } from "../components/ui/FormField";
import { ConfirmDialog } from "../components/ui/Modal";
import { ErrorState, Loading } from "../components/ui/States";
import { PageHeader } from "../components/ui/PageHeader";

export function ClosePage() {
  const { fdNumber = "" } = useParams();
  const navigate = useNavigate();
  const detail = useAsync(() => GetFD(fdNumber), [fdNumber]);

  const [closureDate, setClosureDate] = useState(todayISO());
  const [remark, setRemark] = useState("");
  const [preview, setPreview] = useState<ClosurePreview | null>(null);
  const [previewError, setPreviewError] = useState<string | null>(null);
  const [confirming, setConfirming] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fd = detail.data?.fd;

  useEffect(() => {
    if (!fd || fd.status !== "ACTIVE" || !closureDate) return;
    let cancelled = false;
    const id = setTimeout(() => {
      PreviewClosure({ fdNumber: fd.fdNumber, closureDate, remark })
        .then((p) => {
          if (!cancelled) {
            setPreview(p);
            setPreviewError(null);
          }
        })
        .catch((err: unknown) => {
          if (!cancelled) {
            setPreview(null);
            setPreviewError(errorMessage(err));
          }
        });
    }, 200);
    return () => {
      cancelled = true;
      clearTimeout(id);
    };
  }, [fd, closureDate]);

  if (detail.loading) return <Loading />;
  if (detail.error || !fd) return <ErrorState message={detail.error || "Fixed Deposit not found."} onRetry={detail.reload} />;
  if (fd.status !== "ACTIVE") {
    return (
      <>
        <PageHeader title="Close Fixed Deposit" context={fd.fdNumber} />
        <Alert tone="warning" title="This FD is already closed">
          Only active Fixed Deposits can be closed. This one is {fd.status.toLowerCase()}.
        </Alert>
        <div className="u-mt-3">
          <Button onClick={() => navigate(`/fd/${fd.fdNumber}`)}>Back to FD</Button>
        </div>
      </>
    );
  }

  const confirmClose = async () => {
    setConfirming(false);
    setSaving(true);
    setError(null);
    try {
      await CloseFD({ fdNumber: fd.fdNumber, closureDate, remark });
      navigate(`/fd/${fd.fdNumber}`, { state: { closed: true } });
    } catch (err: unknown) {
      setError(errorMessage(err));
      setSaving(false);
    }
  };

  return (
    <>
      <PageHeader
        title="Close Fixed Deposit"
        context={fd.fdNumber}
        actions={<Button onClick={() => navigate(`/fd/${fd.fdNumber}`)}>Cancel</Button>}
      />

      {error && <Alert tone="danger" title="Could not close the Fixed Deposit">{error}</Alert>}

      <div className="hs-two-col">
        <Card>
          <CardHead title="Closure Details" />
          <CardBody>
            <dl className="hs-info-grid">
              <div className="hs-info">
                <dt className="hs-info__label">Deposit Amount</dt>
                <dd className="hs-info__value">{formatMoney(fd.principal)}</dd>
              </div>
              <div className="hs-info">
                <dt className="hs-info__label">Maturity Date</dt>
                <dd className="hs-info__value">{formatDate(fd.maturityDate)}</dd>
              </div>
              <div className="hs-info">
                <dt className="hs-info__label">Interest Rate</dt>
                <dd className="hs-info__value">{formatPercent(fd.interestRate)}</dd>
              </div>
              <div className="hs-info">
                <dt className="hs-info__label">Maturity Amount</dt>
                <dd className="hs-info__value">{formatMoney(fd.maturityAmount)}</dd>
              </div>
            </dl>

            <div className="hs-field-grid u-mt-4">
              <FormField label="Closure Date" required htmlFor="close-date" hint="Today by default; must not be before the start date.">
                <Input
                  id="close-date"
                  type="date"
                  value={closureDate}
                  min={fd.startDate}
                  onChange={(e) => setClosureDate(e.target.value)}
                />
              </FormField>

              <FormField label="Remark" hint="Optional reason recorded in the history" htmlFor="close-remark">
                <Textarea id="close-remark" rows={3} value={remark} onChange={(e) => setRemark(e.target.value)} />
              </FormField>
            </div>
          </CardBody>
        </Card>

        <Card>
          <CardHead title="Payable Preview" />
          <CardBody>
            {previewError ? (
              <Alert tone="warning">{previewError}</Alert>
            ) : preview ? (
              <>
                <div className="u-mb-3">
                  <Badge tone={preview.isPremature ? "warning" : "success"}>
                    {preview.isPremature ? "Premature closure" : "On maturity"}
                  </Badge>
                </div>
                <dl className="hs-preview">
                  <div className="hs-preview__row">
                    <dt>Days Held</dt>
                    <dd>{formatTenure(preview.daysHeld)}</dd>
                  </div>
                  <div className="hs-preview__row">
                    <dt>Interest Rate Applied</dt>
                    <dd>{formatPercent(preview.ratePercent)}</dd>
                  </div>
                  <div className="hs-preview__row">
                    <dt>Interest</dt>
                    <dd>{formatMoney(preview.interest)}</dd>
                  </div>
                  <div className="hs-preview__row hs-preview__row--total">
                    <dt>Amount Payable</dt>
                    <dd>{formatMoney(preview.payable)}</dd>
                  </div>
                </dl>
                {preview.isPremature && (
                  <p className="u-mt-2 u-text-xs u-muted">
                    Premature closures earn interest for the actual days held at the applicable slab rate.
                  </p>
                )}
                <div className="u-mt-4">
                  <Button variant="danger" size="lg" block onClick={() => setConfirming(true)} loading={saving}>
                    Close FD
                  </Button>
                </div>
              </>
            ) : (
              <p className="u-muted">Set the closure date to see the payable amount.</p>
            )}
          </CardBody>
        </Card>
      </div>

      <ConfirmDialog
        open={confirming}
        danger
        busy={saving}
        title="Confirm closure"
        confirmLabel="Confirm Closure"
        message={
          <>
            Close {fd.fdNumber} and pay out{" "}
            <strong>{formatMoney(preview?.payable ?? 0)}</strong>? This action is recorded in the
            FD history and cannot be undone.
          </>
        }
        onConfirm={confirmClose}
        onCancel={() => setConfirming(false)}
      />
    </>
  );
}
