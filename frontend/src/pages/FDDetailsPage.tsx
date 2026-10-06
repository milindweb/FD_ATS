import { useMemo } from "react";
import { Link, useLocation, useNavigate, useParams } from "react-router-dom";
import { GetFD, errorMessage, type FDDetail } from "../lib/api";
import { daysLabel, formatDate, formatMoney, formatPercent, formatTenure, todayISO } from "../lib/format";
import { useAsync } from "../lib/hooks";
import { Alert } from "../components/ui/Alert";
import { StatusBadge } from "../components/ui/Badge";
import { Button } from "../components/ui/Button";
import { Card, CardBody, CardHead } from "../components/ui/Card";
import { DataTable, type Column } from "../components/data/DataTable";
import { ErrorState, Loading } from "../components/ui/States";
import { PageHeader } from "../components/ui/PageHeader";
import type { HistoryEntry } from "../lib/api";

function daysRemaining(maturityDate: string): number {
  const ms = new Date(`${maturityDate}T00:00:00`).getTime() - new Date(`${todayISO()}T00:00:00`).getTime();
  return Math.round(ms / 86400000);
}

function InfoItem({ label, value, mono }: { label: string; value: string; mono?: boolean }) {
  return (
    <div className="hs-info">
      <dt className="hs-info__label">{label}</dt>
      <dd className={`hs-info__value ${mono ? "hs-mono" : ""}`}>{value}</dd>
    </div>
  );
}

export function FDDetailsPage() {
  const { fdNumber = "" } = useParams();
  const navigate = useNavigate();
  const location = useLocation();
  const detail = useAsync(() => GetFD(fdNumber), [fdNumber]);

  const created = (location.state as { created?: boolean } | null)?.created === true;
  const fd = detail.data?.fd;

  const historyColumns = useMemo<Array<Column<HistoryEntry>>>(
    () => [
      { key: "eventDate", header: "Date", render: (r) => formatDate(r.eventDate) },
      { key: "eventType", header: "Event", render: (r) => r.eventType },
      { key: "amount", header: "Amount", numeric: true, render: (r) => (r.amount ? formatMoney(r.amount) : "—") },
      { key: "interest", header: "Interest", numeric: true, render: (r) => (r.interest ? formatMoney(r.interest) : "—") },
      { key: "referenceFd", header: "Reference FD", mono: true, render: (r) => r.referenceFd || "—" },
      { key: "remarks", header: "Remarks", render: (r) => r.remarks || "—" },
    ],
    [],
  );

  if (detail.loading) return <Loading />;
  if (detail.error || !fd) return <ErrorState message={detail.error || "Fixed Deposit not found."} onRetry={detail.reload} />;

  const remaining = daysRemaining(fd.maturityDate);
  const isActive = fd.status === "ACTIVE";

  return (
    <>
      <PageHeader
        title={
          <span className="u-row u-gap-2">
            <span className="hs-mono">{fd.fdNumber}</span>
            <StatusBadge status={fd.status} detail={fd.closureType} />
          </span>
        }
        context={`${fd.customerName}${fd.customerNumber ? ` · ${fd.customerNumber}` : ""}`}
        actions={
          <>
            <Button icon="arrow-left" onClick={() => navigate("/")}>
              Back to Dashboard
            </Button>
            {isActive && (
              <>
                <Button variant="primary" icon="refresh" onClick={() => navigate(`/fd/${fd.fdNumber}/renew`)}>
                  Renew
                </Button>
                <Button variant="danger" onClick={() => navigate(`/fd/${fd.fdNumber}/close`)}>
                  Close FD
                </Button>
              </>
            )}
          </>
        }
      />

      {created && <Alert tone="success" title="Fixed Deposit created">Saved as {fd.fdNumber}.</Alert>}

      <div className="hs-kpi-row">
        <div className="hs-kpi hs-kpi--success">
          <div className="hs-kpi__label">Deposit Amount</div>
          <div className="hs-kpi__value">{formatMoney(fd.principal)}</div>
          <div className="hs-kpi__meta">{formatPercent(fd.interestRate)} interest</div>
        </div>
        <div className="hs-kpi hs-kpi--info">
          <div className="hs-kpi__label">Interest</div>
          <div className="hs-kpi__value">{formatMoney(fd.interestAmount)}</div>
          <div className="hs-kpi__meta">over {formatTenure(fd.tenureDays)}</div>
        </div>
        <div className="hs-kpi hs-kpi--violet">
          <div className="hs-kpi__label">Maturity Amount</div>
          <div className="hs-kpi__value">{formatMoney(fd.maturityAmount)}</div>
          <div className="hs-kpi__meta">on {formatDate(fd.maturityDate)}</div>
        </div>
        <div className="hs-kpi hs-kpi--danger">
          <div className="hs-kpi__label">Days Remaining</div>
          <div className="hs-kpi__value">{isActive ? daysLabel(remaining) : "—"}</div>
          <div className="hs-kpi__meta">{isActive ? (remaining < 0 ? "matured" : "until maturity") : fd.status.toLowerCase()}</div>
        </div>
      </div>

      <div className="hs-two-col">
        <Card>
          <CardHead title="Deposit Details" />
          <CardBody>
            <dl className="hs-info-grid">
              <InfoItem label="FD Number" value={fd.fdNumber} mono />
              <InfoItem label="Customer / Member" value={fd.customerName} />
              <InfoItem label="Customer / Member No" value={fd.customerNumber || "—"} mono />
              <InfoItem label="Deposit Amount" value={formatMoney(fd.principal)} />
              <InfoItem label="Start Date" value={formatDate(fd.startDate)} />
              <InfoItem label="Tenure" value={formatTenure(fd.tenureDays)} />
              <InfoItem label="Interest Rate" value={formatPercent(fd.interestRate)} />
              <InfoItem label="Maturity Date" value={formatDate(fd.maturityDate)} />
              <InfoItem label="Interest Amount" value={formatMoney(fd.interestAmount)} />
              <InfoItem label="Maturity Amount" value={formatMoney(fd.maturityAmount)} />
            </dl>
          </CardBody>
        </Card>

        <Card>
          <CardHead title={isActive ? "Status" : "Closure Details"} />
          <CardBody>
            <dl className="hs-info-grid">
              <InfoItem label="Status" value={fd.status} />
              {fd.closureDate && <InfoItem label="Closure Date" value={formatDate(fd.closureDate)} />}
              {fd.closureType && <InfoItem label="Closure Type" value={fd.closureType} />}
              {fd.closureDays !== undefined && fd.closureDays !== null && (
                <InfoItem label="Days Held" value={formatTenure(fd.closureDays)} />
              )}
              {fd.closureRate !== undefined && fd.closureRate !== null && (
                <InfoItem label="Closure Rate" value={formatPercent(fd.closureRate)} />
              )}
              {fd.closureInterest !== undefined && fd.closureInterest !== null && (
                <InfoItem label="Closure Interest" value={formatMoney(fd.closureInterest)} />
              )}
              {fd.closurePayable !== undefined && fd.closurePayable !== null && (
                <InfoItem label="Amount Paid" value={formatMoney(fd.closurePayable)} />
              )}
              {fd.closureRemark && <InfoItem label="Remarks" value={fd.closureRemark} />}
              {fd.renewedFrom && (
                <InfoItem label="Renewed From" value={fd.renewedFrom} mono />
              )}
              {fd.renewedTo && <InfoItem label="Renewed To" value={fd.renewedTo} mono />}
            </dl>
            {(fd.renewedFrom || fd.renewedTo) && (
              <div className="u-mt-3 u-row u-gap-2 u-flex-wrap">
                {fd.renewedFrom && (
                  <Link className="hs-btn hs-btn--secondary hs-btn--sm" to={`/fd/${fd.renewedFrom}`}>
                    View previous FD
                  </Link>
                )}
                {fd.renewedTo && (
                  <Link className="hs-btn hs-btn--secondary hs-btn--sm" to={`/fd/${fd.renewedTo}`}>
                    View renewed FD
                  </Link>
                )}
              </div>
            )}
          </CardBody>
        </Card>
      </div>

      <Card className="u-mt-5">
        <CardHead title="History" actions={<span className="u-text-xs u-muted">{detail.data?.history.length ?? 0} events</span>} />
        <CardBody flush>
          <DataTable
            columns={historyColumns}
            rows={detail.data?.history ?? []}
            rowKey={(r) => String(r.id)}
            emptyTitle="No history yet"
            emptyDescription="Every create, renew and close event is recorded here."
          />
        </CardBody>
      </Card>
    </>
  );
}
