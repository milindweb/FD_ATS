import { useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  DashboardStats,
  UpcomingMaturities,
  type UpcomingFD,
} from "../lib/api";
import { daysLabel, formatDate, formatMoney, formatPercent } from "../lib/format";
import { useAsync } from "../lib/hooks";
import { DataTable, type Column } from "../components/data/DataTable";
import { KpiCard } from "../components/data/KpiCard";
import { MaturityChartCard } from "../components/data/MaturityChart";
import { Card, CardBody, CardHead } from "../components/ui/Card";
import { Button } from "../components/ui/Button";
import { Input } from "../components/ui/FormField";
import { ErrorState, Loading } from "../components/ui/States";
import { PageHeader } from "../components/ui/PageHeader";
import { StatusBadge } from "../components/ui/Badge";
import { Icon } from "../lib/icons";

/** Overview screen: quick search, quick actions, KPIs and maturity data (SRS §9.1). */
export function DashboardPage() {
  const navigate = useNavigate();
  const [query, setQuery] = useState("");

  const stats = useAsync(() => DashboardStats(), []);
  const upcoming = useAsync(() => UpcomingMaturities({ fromDate: "", toDate: "" }), []);

  // Enter hands the quick-search query to the FD Master page (SRS §9.1).
  const openFDMaster = () => {
    const q = query.trim();
    navigate(q ? `/fds?q=${encodeURIComponent(q)}` : "/fds");
  };

  const upcomingColumns = useMemo<Array<Column<UpcomingFD>>>(
    () => [
      { key: "fdNumber", header: "FD Number", mono: true, render: (r) => r.fdNumber },
      { key: "customer", header: "Member Name", render: (r) => r.customerName },
      { key: "principal", header: "Principal", numeric: true, render: (r) => formatMoney(r.principal) },
      { key: "maturityDate", header: "Maturity Date", render: (r) => formatDate(r.maturityDate) },
      { key: "days", header: "Days Remaining", numeric: true, render: (r) => daysLabel(r.daysRemaining) },
      { key: "maturityAmount", header: "Maturity Amount", numeric: true, render: (r) => formatMoney(r.maturityAmount) },
    ],
    [],
  );

  return (
    <>
      <PageHeader title="Dashboard" context="Overview — KPIs, maturity chart, upcoming maturities and quick actions" />

      {/* Quick search sits between the header and the action buttons (SRS §9.1). */}
      <div className="hs-filter-bar">
        <div className="hs-filter-bar__search">
          <div style={{ position: "relative" }}>
            <span
              aria-hidden="true"
              style={{
                position: "absolute",
                left: 10,
                top: "50%",
                transform: "translateY(-50%)",
                color: "var(--color-text-muted)",
                display: "flex",
              }}
            >
              <Icon name="search" size={15} />
            </span>
            <Input
              type="search"
              value={query}
              placeholder="Quick search FDs — press Enter to open results…"
              aria-label="Quick search"
              onChange={(e) => setQuery(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") openFDMaster();
              }}
              style={{ paddingInlineStart: 32 }}
            />
          </div>
        </div>
      </div>

      {/* Quick actions shown side by side, directly under the search (SRS §9.1). */}
      <div className="u-row u-gap-2 u-mb-4">
        <Button variant="primary" icon="plus" onClick={() => navigate("/member/new")}>
          New Member
        </Button>
        <Button variant="secondary" icon="plus" onClick={() => navigate("/new")}>
          Create New FD
        </Button>
        <Button variant="ghost" icon="sheet" onClick={() => navigate("/fds")}>
          View full FD list
        </Button>
      </div>

      {stats.loading ? (
        <Loading />
      ) : stats.error ? (
        <ErrorState message={stats.error} onRetry={stats.reload} />
      ) : (
        <div className="hs-kpi-row">
          <KpiCard
            tone="danger"
            icon="grid"
            label="Active FDs"
            value={String(stats.data?.activeFds ?? 0)}
            meta={`of ${stats.data?.totalFds ?? 0} total`}
          />
          <KpiCard
            tone="success"
            icon="banknote"
            label="Total Deposit"
            value={formatMoney(stats.data?.activePrincipal)}
            meta="held in active FDs"
          />
          <KpiCard
            tone="info"
            icon="percent"
            label="Total Interest"
            value={formatMoney(stats.data?.totalInterest)}
            meta={`across ${stats.data?.activeFds ?? 0} active FDs`}
          />
          <KpiCard
            tone="violet"
            icon="sheet"
            label="Financial Year"
            value={formatMoney(stats.data?.fyDeposits)}
            meta={`${stats.data?.fyLabel || "FY —"} · deposited this FY`}
          />
        </div>
      )}

      <div className="hs-two-col hs-two-col--fixed">
        <MaturityChartCard />
        <Card>
          <CardHead
            title="Upcoming Maturities"
            actions={<span className="u-text-xs u-muted">Next 90 days</span>}
          />
          <CardBody flush>
            <DataTable
              columns={upcomingColumns}
              rows={upcoming.data ?? []}
              rowKey={(r) => r.fdNumber}
              loading={upcoming.loading}
              emptyTitle="No maturities in the next 90 days"
              emptyDescription="FDs maturing within the next 90 days appear here automatically."
            />
          </CardBody>
        </Card>
      </div>
    </>
  );
}
