import { useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  DashboardStats,
  ListFDs,
  UpcomingMaturities,
  fdFilters,
  type FD,
  type UpcomingFD,
} from "../lib/api";
import { daysLabel, formatDate, formatMoney, formatPercent } from "../lib/format";
import { useAsync, useDebounced } from "../lib/hooks";
import { DataTable, type Column } from "../components/data/DataTable";
import { FilterBar } from "../components/data/FilterBar";
import { KpiCard } from "../components/data/KpiCard";
import { MaturityChartCard } from "../components/data/MaturityChart";
import { PaginationBar } from "../components/data/PaginationBar";
import { Card, CardBody, CardHead } from "../components/ui/Card";
import { Button } from "../components/ui/Button";
import { ErrorState, Loading } from "../components/ui/States";
import { PageHeader } from "../components/ui/PageHeader";
import { StatusBadge } from "../components/ui/Badge";

const PAGE_SIZE = 25;

export function DashboardPage() {
  const navigate = useNavigate();
  const [search, setSearch] = useState("");
  const [filter, setFilter] = useState<(typeof fdFilters)[number]["value"]>("ALL");
  const [page, setPage] = useState(1);
  const debouncedSearch = useDebounced(search, 300);

  const stats = useAsync(() => DashboardStats(), []);
  const upcoming = useAsync(() => UpcomingMaturities({ fromDate: "", toDate: "" }), []);
  const list = useAsync(
    () => ListFDs({ search: debouncedSearch, filter, page, pageSize: PAGE_SIZE }),
    [debouncedSearch, filter, page],
  );

  const columns = useMemo<Array<Column<FD>>>(
    () => [
      { key: "fdNumber", header: "FD Number", mono: true, render: (r) => r.fdNumber },
      { key: "customer", header: "Customer/Member", render: (r) => r.customerName },
      { key: "principal", header: "Deposit Amount", numeric: true, render: (r) => formatMoney(r.principal) },
      { key: "rate", header: "Interest Rate", numeric: true, render: (r) => formatPercent(r.interestRate) },
      { key: "start", header: "Start Date", render: (r) => formatDate(r.startDate) },
      { key: "maturity", header: "Maturity Date", render: (r) => formatDate(r.maturityDate) },
      { key: "maturityAmount", header: "Maturity Amount", numeric: true, render: (r) => formatMoney(r.maturityAmount) },
      { key: "status", header: "Status", render: (r) => <StatusBadge status={r.status} detail={r.closureType} /> },
    ],
    [],
  );

  const upcomingColumns = useMemo<Array<Column<UpcomingFD>>>(
    () => [
      { key: "fdNumber", header: "FD Number", mono: true, render: (r) => r.fdNumber },
      { key: "customer", header: "Customer/Member", render: (r) => r.customerName },
      { key: "principal", header: "Principal", numeric: true, render: (r) => formatMoney(r.principal) },
      { key: "maturityDate", header: "Maturity Date", render: (r) => formatDate(r.maturityDate) },
      { key: "days", header: "Days Remaining", numeric: true, render: (r) => daysLabel(r.daysRemaining) },
      { key: "maturityAmount", header: "Maturity Amount", numeric: true, render: (r) => formatMoney(r.maturityAmount) },
    ],
    [],
  );

  return (
    <>
      <PageHeader
        title="Dashboard"
        context="All Fixed Deposits — search, filter and manage"
        actions={
          <Button variant="primary" icon="plus" onClick={() => navigate("/new")}>
            Create New FD
          </Button>
        }
      />

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

      <div className="u-mt-5">
        <FilterBar
          search={search}
          onSearchChange={(v) => {
            setSearch(v);
            setPage(1);
          }}
          filters={fdFilters}
          activeFilter={filter}
          onFilterChange={(v) => {
            setFilter(v);
            setPage(1);
          }}
        />
      </div>

      <DataTable
        columns={columns}
        rows={list.data?.items ?? []}
        rowKey={(r) => r.fdNumber}
        loading={list.loading}
        onRowClick={(r) => navigate(`/fd/${r.fdNumber}`)}
        emptyTitle={search || filter !== "ALL" ? "No FDs match your search" : "No Fixed Deposits yet"}
        emptyDescription={
          search || filter !== "ALL"
            ? "Try a different search term or filter."
            : "Create your first Fixed Deposit to get started."
        }
      />
      {!list.loading && !list.error && (list.data?.total ?? 0) > 0 && (
        <PaginationBar page={list.data?.page ?? 1} pageSize={PAGE_SIZE} total={list.data?.total ?? 0} onPageChange={setPage} />
      )}
      {list.error && <ErrorState message={list.error} onRetry={list.reload} />}
    </>
  );
}
