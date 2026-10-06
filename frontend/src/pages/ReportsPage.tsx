import { useState } from "react";
import { ExportReport, PickReportPath, PreviewReport, errorMessage, reportKinds } from "../lib/api";
import { todayISO } from "../lib/format";
import { useAsync } from "../lib/hooks";
import { DataTable, type Column } from "../components/data/DataTable";
import { Alert } from "../components/ui/Alert";
import { Button } from "../components/ui/Button";
import { Card, CardBody, CardHead } from "../components/ui/Card";
import { ErrorState, Loading } from "../components/ui/States";
import { FormField, Input } from "../components/ui/FormField";
import { PageHeader } from "../components/ui/PageHeader";

interface PreviewRow {
  index: number;
  cells: string[];
}

const NUMERIC_HEADER = /amount|rate|days|tenure|interest/i;

export function ReportsPage() {
  const [kind, setKind] = useState<(typeof reportKinds)[number]["kind"]>("REGISTER");
  const [fromDate, setFromDate] = useState("");
  const [toDate, setToDate] = useState("");
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState<{ tone: "success" | "danger"; text: string } | null>(null);

  const selected = reportKinds.find((k) => k.kind === kind)!;
  const dateFiltered = kind === "REGISTER" || kind === "MATURITY";

  const preview = useAsync(
    () => PreviewReport({ kind, fromDate, toDate, path: "" }),
    [kind, fromDate, toDate],
  );

  const previewRows: PreviewRow[] = (preview.data?.rows ?? []).map((cells, index) => ({ index, cells }));
  const previewColumns: Array<Column<PreviewRow>> = (preview.data?.headers ?? []).map((header, idx) => ({
    key: String(idx),
    header,
    numeric: NUMERIC_HEADER.test(header),
    render: (row) => row.cells[idx] ?? "",
  }));

  const exportReport = async () => {
    setMessage(null);
    setBusy(true);
    try {
      const path = await PickReportPath(`${selected.file}_${todayISO()}.xlsx`);
      if (!path) {
        setBusy(false);
        return;
      }
      const result = await ExportReport({ kind, fromDate, toDate, path });
      setMessage({ tone: "success", text: `Report saved with ${result.rowCount} rows → ${result.path}` });
    } catch (err: unknown) {
      setMessage({ tone: "danger", text: errorMessage(err) });
    } finally {
      setBusy(false);
    }
  };

  return (
    <>
      <PageHeader title="Reports" context="Generate Excel reports for the full register, maturities or status groups" />

      {message && <Alert tone={message.tone}>{message.text}</Alert>}

      <div className="hs-two-col">
        <Card>
          <CardHead title="Choose a Report" />
          <CardBody>
            <div className="hs-report-list" role="radiogroup" aria-label="Report type">
              {reportKinds.map((k) => (
                <button
                  key={k.kind}
                  type="button"
                  role="radio"
                  aria-checked={kind === k.kind}
                  className={`hs-report-option ${kind === k.kind ? "is-active" : ""}`}
                  onClick={() => setKind(k.kind)}
                >
                  <span className="hs-report-option__title">{k.label}</span>
                  <span className="hs-report-option__file hs-mono">{k.file}.xlsx</span>
                </button>
              ))}
            </div>
          </CardBody>
        </Card>

        <Card>
          <CardHead title="Date Range" />
          <CardBody>
            {dateFiltered ? (
              <div className="hs-field-grid">
                <FormField label="From Date" hint="Leave empty for no lower bound" htmlFor="rep-from">
                  <Input id="rep-from" type="date" value={fromDate} onChange={(e) => setFromDate(e.target.value)} />
                </FormField>
                <FormField label="To Date" hint="Leave empty for no upper bound" htmlFor="rep-to">
                  <Input id="rep-to" type="date" value={toDate} onChange={(e) => setToDate(e.target.value)} />
                </FormField>
              </div>
            ) : (
              <p className="u-muted">
                The {selected.label} covers all {kind === "ACTIVE" ? "active" : "closed"} FDs — no date filter
                applies.
              </p>
            )}

            <div className="u-mt-4">
              <Button variant="primary" size="lg" block icon="download" onClick={exportReport} loading={busy}>
                Export {selected.label}
              </Button>
            </div>
            <p className="u-mt-2 u-text-xs u-muted">
              Excel columns: FD Number, Customer, Principal, Rate, Start, Maturity Date, Interest, Maturity
              Amount, Status and Closure details — with filters enabled on the header row.
            </p>
          </CardBody>
        </Card>
      </div>

      <Card className="u-mt-5">
        <CardHead
          title={`Preview — ${selected.label}`}
          actions={<span className="u-text-xs u-muted">Exactly what will be exported</span>}
        />
        <CardBody flush>
          {preview.loading ? (
            <Loading />
          ) : preview.error ? (
            <ErrorState message={preview.error} onRetry={preview.reload} />
          ) : (
            <>
              <DataTable
                columns={previewColumns}
                rows={previewRows}
                rowKey={(row) => String(row.index)}
                emptyTitle="Nothing to show"
                emptyDescription="No rows match the selected report and date range."
              />
              {(preview.data?.total ?? 0) > previewRows.length && (
                <p className="u-mb-3 u-text-xs u-muted" style={{ padding: "0 var(--space-4)" }}>
                  Showing the first {previewRows.length} of {preview.data?.total} rows — the exported file
                  contains all of them.
                </p>
              )}
            </>
          )}
        </CardBody>
      </Card>
    </>
  );
}
