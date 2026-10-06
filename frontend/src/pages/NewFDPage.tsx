import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  CreateFD,
  PreviewFD,
  errorMessage,
  type Calculation,
  type PreviewRequest,
} from "../lib/api";
import { formatDate, formatMoney, formatPercent, formatTenure, todayISO } from "../lib/format";
import { Alert } from "../components/ui/Alert";
import { Button } from "../components/ui/Button";
import { Card, CardBody, CardHead } from "../components/ui/Card";
import { FormField, Input } from "../components/ui/FormField";
import { PageHeader } from "../components/ui/PageHeader";

interface FormState {
  customerName: string;
  customerNumber: string;
  principal: string;
  startDate: string;
  tenureDays: string;
}

const initialForm = (): FormState => ({
  customerName: "",
  customerNumber: "",
  principal: "",
  startDate: todayISO(),
  tenureDays: "1095",
});

export function NewFDPage() {
  const navigate = useNavigate();
  const [form, setForm] = useState<FormState>(initialForm);
  const [preview, setPreview] = useState<Calculation | null>(null);
  const [previewError, setPreviewError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [fieldErrors, setFieldErrors] = useState<Partial<Record<keyof FormState, string>>>({});

  const request: PreviewRequest = useMemo(
    () => ({
      customerName: form.customerName.trim(),
      customerNumber: form.customerNumber.trim(),
      principal: Math.round(Number(form.principal) || 0),
      startDate: form.startDate,
      tenureDays: Number(form.tenureDays) || 0,
    }),
    [form],
  );

  const valid =
    request.customerName !== "" &&
    request.principal > 0 &&
    request.startDate !== "" &&
    request.tenureDays > 0;

  useEffect(() => {
    if (!valid) {
      setPreview(null);
      setPreviewError(null);
      return;
    }
    let cancelled = false;
    const id = setTimeout(() => {
      PreviewFD(request)
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
  }, [request, valid]);

  const set = (key: keyof FormState) => (e: { target: { value: string } }) => {
    setForm((f) => ({ ...f, [key]: e.target.value }));
    setFieldErrors((fe) => ({ ...fe, [key]: undefined }));
  };

  const validate = (): boolean => {
    const errors: Partial<Record<keyof FormState, string>> = {};
    if (!form.customerName.trim()) errors.customerName = "Customer name is required.";
    const principal = Number(form.principal);
    if (!form.principal || Number.isNaN(principal)) errors.principal = "Deposit amount is required.";
    else if (principal <= 0) errors.principal = "Deposit amount must be greater than zero.";
    if (!form.startDate) errors.startDate = "Start date is required.";
    const tenure = Number(form.tenureDays);
    if (!form.tenureDays || Number.isNaN(tenure)) errors.tenureDays = "Tenure is required.";
    else if (tenure <= 0) errors.tenureDays = "Tenure must be at least 1 day.";
    setFieldErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const submit = async () => {
    setSubmitError(null);
    if (!validate()) return;
    setSaving(true);
    try {
      const fd = await CreateFD(request);
      navigate(`/fd/${fd.fdNumber}`, { state: { created: true } });
    } catch (err: unknown) {
      setSubmitError(errorMessage(err));
      setSaving(false);
    }
  };

  return (
    <>
      <PageHeader
        title="New Fixed Deposit"
        context="Enter customer and deposit details — the calculation updates as you type"
        actions={<Button onClick={() => navigate("/")}>Cancel</Button>}
      />

      {submitError && (
        <Alert tone="danger" title="Could not save the Fixed Deposit">
          {submitError}
        </Alert>
      )}

      <div className="hs-two-col">
        <Card>
          <CardHead title="Deposit Details" />
          <CardBody>
            <div className="hs-field-grid">
              <FormField label="Customer / Member Name" required error={fieldErrors.customerName} htmlFor="fd-customer">
                <Input
                  id="fd-customer"
                  value={form.customerName}
                  onChange={set("customerName")}
                  invalid={!!fieldErrors.customerName}
                  placeholder="Full name"
                />
              </FormField>

              <FormField label="Customer / Member Number" hint="Optional — e.g. membership or account number" htmlFor="fd-number">
                <Input id="fd-number" value={form.customerNumber} onChange={set("customerNumber")} placeholder="e.g. M-1042" />
              </FormField>

              <FormField label="Deposit Amount (₹)" required error={fieldErrors.principal} htmlFor="fd-principal">
                <Input
                  id="fd-principal"
                  type="number"
                  min={1}
                  step={1}
                  large
                  value={form.principal}
                  onChange={set("principal")}
                  invalid={!!fieldErrors.principal}
                  placeholder="50000"
                />
              </FormField>

              <FormField label="Start Date" required error={fieldErrors.startDate} htmlFor="fd-start">
                <Input
                  id="fd-start"
                  type="date"
                  value={form.startDate}
                  onChange={set("startDate")}
                  invalid={!!fieldErrors.startDate}
                />
              </FormField>

              <FormField label="Tenure (days)" required error={fieldErrors.tenureDays} htmlFor="fd-tenure" hint="Common tenures: 365, 730, 1095 days">
                <Input
                  id="fd-tenure"
                  type="number"
                  min={1}
                  step={1}
                  value={form.tenureDays}
                  onChange={set("tenureDays")}
                  invalid={!!fieldErrors.tenureDays}
                />
              </FormField>

              <div className="hs-tenure-presets" role="group" aria-label="Tenure presets">
                {[180, 365, 730, 1095].map((d) => (
                  <Button
                    key={d}
                    size="sm"
                    variant={form.tenureDays === String(d) ? "primary" : "secondary"}
                    onClick={() => setForm((f) => ({ ...f, tenureDays: String(d) }))}
                  >
                    {d === 180 ? "6 months" : d === 365 ? "1 year" : d === 730 ? "2 years" : "3 years"}
                  </Button>
                ))}
              </div>
            </div>
          </CardBody>
        </Card>

        <Card>
          <CardHead title="Calculation Preview" />
          <CardBody>
            {previewError ? (
              <Alert tone="warning">{previewError}</Alert>
            ) : preview ? (
              <dl className="hs-preview">
                <div className="hs-preview__row">
                  <dt>Deposit Amount</dt>
                  <dd>{formatMoney(preview.principal)}</dd>
                </div>
                <div className="hs-preview__row">
                  <dt>Interest Rate</dt>
                  <dd>{formatPercent(preview.ratePercent)}</dd>
                </div>
                <div className="hs-preview__row">
                  <dt>Tenure</dt>
                  <dd>
                    {formatTenure(preview.tenureDays)} ({formatDate(preview.startDate)} → {formatDate(preview.maturityDate)})
                  </dd>
                </div>
                <div className="hs-preview__row">
                  <dt>Interest</dt>
                  <dd>{formatMoney(preview.interest)}</dd>
                </div>
                <div className="hs-preview__row hs-preview__row--total">
                  <dt>Maturity Amount</dt>
                  <dd>{formatMoney(preview.maturityAmount)}</dd>
                </div>
              </dl>
            ) : (
              <p className="u-muted">Fill the deposit details to see the interest and maturity calculation.</p>
            )}

            <div className="u-mt-4">
              <Button variant="primary" size="lg" block onClick={submit} loading={saving} disabled={!valid}>
                Save Fixed Deposit
              </Button>
            </div>
            <p className="u-mt-2 u-text-xs u-muted">
              Interest = Principal × Rate × Days ÷ 365 ÷ 100, rounded to the nearest rupee.
            </p>
          </CardBody>
        </Card>
      </div>
    </>
  );
}
